package cluster

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/AnubisWatch/anubiswatch/internal/core"
	"github.com/AnubisWatch/anubiswatch/internal/raft"
	"github.com/AnubisWatch/anubiswatch/internal/storage"
)

// buildTLSPeerConfig builds a tls.Config for peer-to-peer TLS from Raft config.
// If no TLS cert/key are configured, it returns nil (plaintext fallback with a warning).
// If RequireClientCert is set, mutual TLS (mTLS) is enabled for peer verification.
func buildTLSPeerConfig(cfg *core.TLSPeerConfig) (*tls.Config, error) {
	if cfg == nil || cfg.CertFile == "" || cfg.KeyFile == "" {
		return nil, nil
	}

	cert, err := tls.LoadX509KeyPair(cfg.CertFile, cfg.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load peer TLS certificate: %w", err)
	}

	tlsCfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}

	// Optionally verify peer certificates with a custom CA
	if cfg.CAFile != "" {
		caBytes, err := os.ReadFile(cfg.CAFile)
		if err != nil {
			return nil, fmt.Errorf("failed to read TLS CA file: %w", err)
		}
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caBytes) {
			return nil, fmt.Errorf("failed to parse TLS CA certificates")
		}
		tlsCfg.RootCAs = caPool

		if cfg.VerifyPeers {
			tlsCfg.ClientAuth = tls.RequireAndVerifyClientCert
			tlsCfg.ClientCAs = caPool
		} else if cfg.RequireClientCert {
			tlsCfg.ClientAuth = tls.RequireAnyClientCert
		}
	} else if cfg.RequireClientCert || cfg.VerifyPeers {
		// No CA provided but mTLS required — require client certs without verification
		tlsCfg.ClientAuth = tls.RequireAnyClientCert
	}

	return tlsCfg, nil
}

// Manager handles cluster coordination
type Manager struct {
	mu            sync.RWMutex
	necroConfig   core.NecropolisConfig
	config        core.RaftConfig
	node          *raft.Node
	db            *storage.CobaltDB
	logStore      *storage.CobaltDBLogStore
	snapshotStore *storage.CobaltDBSnapshotStore
	stableStore   *storage.CobaltDBStableStore
	fsm           *raft.StorageFSM
	logger        *slog.Logger
	isClustered   bool

	// Distribution
	distributor *Distributor

	// Discovery
	discovery *raft.Discovery
}

// NewManager creates a new cluster manager
func NewManager(cfg core.NecropolisConfig, db *storage.CobaltDB, logger *slog.Logger) (*Manager, error) {
	m := &Manager{
		necroConfig: cfg,
		config:      cfg.Raft,
		db:          db,
		logger:      logger.With("component", "cluster"),
		isClustered: cfg.Raft.Bootstrap || len(cfg.Raft.Peers) > 0,
	}

	// Create Raft storage components
	m.logStore = storage.NewCobaltDBLogStore(db)
	m.snapshotStore = storage.NewCobaltDBSnapshotStore(db)
	m.stableStore = storage.NewCobaltDBStableStore(db)

	// Create FSM backed by storage
	m.fsm = raft.NewStorageFSM(db)

	return m, nil
}

// Start initializes and starts the Raft node
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isClustered {
		m.logger.Info("running in standalone mode")
		return nil
	}

	m.logger.Info("starting Raft node", "node_id", m.config.NodeID, "bind_addr", m.config.BindAddr)

	// Build TLS config for peer-to-peer communication (nil if TLS not configured)
	tlsCfg, err := buildTLSPeerConfig(m.config.TLS)
	if err != nil {
		return fmt.Errorf("failed to build TLS config: %w", err)
	}
	if tlsCfg != nil {
		m.logger.Info("Raft peer TLS enabled", "min_version", "TLS 1.2", "verify_peers", m.config.TLS.VerifyPeers)
	} else if m.necroConfig.ClusterSecret != "" {
		m.logger.Warn("Raft inter-node auth uses only pre-shared key (no TLS); specify tls.cert_file to enable mTLS")
	}

	// Create TCP transport with optional TLS
	transport, err := raft.NewTCPTransport(m.config.BindAddr, m.config.AdvertiseAddr, tlsCfg, m.logger)
	if err != nil {
		return fmt.Errorf("failed to create transport: %w", err)
	}

	// Create Raft node
	node, err := raft.NewNode(m.config, m.logStore, m.snapshotStore, m.fsm, m.logger)
	if err != nil {
		return fmt.Errorf("failed to create Raft node: %w", err)
	}

	// Set the transport on the node
	node.SetTransport(transport)
	m.node = node

	// Start Raft node (this also starts the transport)
	if err := node.Start(); err != nil {
		return fmt.Errorf("failed to start Raft node: %w", err)
	}

	m.logger.Info("Raft node started")

	// Initialize and start discovery service
	if cfg := m.necroConfig.Discovery; cfg.Mode != "" && cfg.Mode != "manual" {
		disc, err := raft.NewDiscovery(m.config, m.logger)
		if err != nil {
			m.logger.Warn("failed to create discovery service", "err", err)
		} else {
			// Wire peer discovery callbacks to Raft node
			disc.RegisterPeerCallback(
				func(peer core.RaftPeer) {
					m.logger.Info("peer discovered", "id", peer.ID, "addr", peer.Address)
					if m.node != nil {
						m.node.AddPeer(peer)
					}
				},
				func(nodeID string) {
					m.logger.Info("peer lost", "id", nodeID)
					if m.node != nil {
						m.node.RemovePeer(nodeID)
					}
				},
			)

			if err := disc.Start(); err != nil {
				m.logger.Warn("failed to start discovery service", "err", err)
			} else {
				m.discovery = disc
				m.logger.Info("auto-discovery started", "mode", cfg.Mode)
			}
		}
	}

	// Initialize distributor
	strategy := StrategyLoadBased
	m.distributor = NewDistributor(m.config.NodeID, m.config.Region, strategy, m.logger)

	// Register self
	m.distributor.RegisterNode(m.config.NodeID, m.config.Region)

	// Register peers
	for _, peer := range m.config.Peers {
		m.distributor.RegisterNode(peer.ID, peer.Region)
	}

	// Start distributor
	m.distributor.Start()

	m.logger.Info("Cluster distributor started", "strategy", strategy.String())
	return nil
}

// Stop gracefully shuts down the Raft node
func (m *Manager) Stop(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.distributor != nil {
		m.logger.Info("stopping distributor")
		m.distributor.Stop()
	}

	if m.discovery != nil {
		m.logger.Info("stopping discovery")
		m.discovery.Stop()
	}

	if m.node != nil {
		m.logger.Info("stopping Raft node")
		m.node.Stop()
	}

	return nil
}

// IsLeader returns true if this node is the Raft leader
func (m *Manager) IsLeader() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.node == nil {
		return false
	}

	return m.node.State() == core.StateLeader
}

// Leader returns the current leader's ID
func (m *Manager) Leader() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.node == nil {
		return ""
	}

	return m.node.Leader()
}

// IsClustered returns true if cluster mode is enabled
func (m *Manager) IsClustered() bool {
	return m.isClustered
}

// GetStatus returns cluster status information
func (m *Manager) GetStatus() *ClusterStatus {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := &ClusterStatus{
		IsClustered: m.isClustered,
		NodeID:      m.config.NodeID,
	}

	if m.node != nil {
		status.State = m.node.State().String()
		status.Leader = m.node.Leader()
		status.Term = m.node.CurrentTerm()
		status.PeerCount = len(m.node.Peers())
		status.CommitIndex = m.node.CommitIndex()
	}

	return status
}

// GetDiscoveredPeers returns peers discovered via mDNS/gossip
func (m *Manager) GetDiscoveredPeers() []raft.DiscoveredPeer {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.discovery == nil {
		return nil
	}
	return m.discovery.GetPeers()
}

// ClusterStatus contains cluster status information
type ClusterStatus struct {
	IsClustered bool   `json:"is_clustered"`
	NodeID      string `json:"node_id"`
	State       string `json:"state,omitempty"`
	Leader      string `json:"leader,omitempty"`
	Term        uint64 `json:"term,omitempty"`
	PeerCount   int    `json:"peer_count,omitempty"`
	CommitIndex uint64 `json:"commit_index,omitempty"`
}
