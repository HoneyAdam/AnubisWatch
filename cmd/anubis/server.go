package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/AnubisWatch/anubiswatch/internal/alert"
	"github.com/AnubisWatch/anubiswatch/internal/api"
	"github.com/AnubisWatch/anubiswatch/internal/auth"
	"github.com/AnubisWatch/anubiswatch/internal/cluster"
	"github.com/AnubisWatch/anubiswatch/internal/core"
	"github.com/AnubisWatch/anubiswatch/internal/dashboard"
	"github.com/AnubisWatch/anubiswatch/internal/grpcapi"
	"github.com/AnubisWatch/anubiswatch/internal/journey"
	"github.com/AnubisWatch/anubiswatch/internal/probe"
	"github.com/AnubisWatch/anubiswatch/internal/statuspage"
	"github.com/AnubisWatch/anubiswatch/internal/storage"
)

// ServerDependencies holds all dependencies for the server
type ServerDependencies struct {
	Config            *core.Config
	Logger            *slog.Logger
	Store             *storage.CobaltDB
	Authenticator     api.Authenticator
	AlertManager      *alert.Manager
	ProbeEngine       *probe.Engine
	JourneyExecutor   *journey.Executor
	ClusterManager    *cluster.Manager
	RESTServer        *api.RESTServer
	GRPCServer        *grpcapi.Server
	StatusPageRepo    *statusPageRepository
	ACMEManager       interface{}
	DashboardHandler  http.Handler
	StatusPageHandler http.Handler
	MCPServer         *api.MCPServer
}

// Server represents the AnubisWatch server
type Server struct {
	deps   *ServerDependencies
	logger *slog.Logger
}

// NewServer creates a new Server instance
func NewServer(deps *ServerDependencies) *Server {
	return &Server{
		deps:   deps,
		logger: deps.Logger,
	}
}

// Start initializes and starts all server components
func (s *Server) Start(ctx context.Context) error {
	cfg := s.deps.Config
	logger := s.logger

	// Start alert manager
	if s.deps.AlertManager != nil {
		if err := s.deps.AlertManager.Start(); err != nil {
			logger.Warn("failed to start alert manager", "err", err)
		} else {
			logger.Info("alert manager started")
		}
	}

	// Assign souls from config
	if len(cfg.Souls) > 0 {
		soulPtrs := make([]*core.Soul, len(cfg.Souls))
		for i := range cfg.Souls {
			soulPtrs[i] = &cfg.Souls[i]
		}
		s.deps.ProbeEngine.AssignSouls(soulPtrs)
		logger.Info("souls assigned", "count", len(cfg.Souls))
	}

	// Start journey executors
	for _, j := range cfg.Journeys {
		if j.Enabled {
			j.WorkspaceID = "default"
			if err := s.deps.Store.SaveJourney(ctx, &j); err != nil {
				logger.Warn("failed to save journey", "journey", j.Name, "err", err)
			}
			if err := s.deps.JourneyExecutor.Start(ctx, &j); err != nil {
				logger.Warn("failed to start journey", "journey", j.Name, "err", err)
			}
		}
	}

	// Start cluster manager
	if s.deps.ClusterManager != nil {
		if err := s.deps.ClusterManager.Start(ctx); err != nil {
			logger.Warn("failed to start cluster manager", "err", err)
		} else {
			logger.Info("cluster manager initialized", "clustered", s.deps.ClusterManager.IsClustered())
		}
	}

	// Start REST server
	if s.deps.RESTServer != nil {
		go func() {
			if err := s.deps.RESTServer.Start(); err != nil {
				logger.Error("REST server failed", "err", err)
			}
		}()
		logger.Info("REST API server initialized", "addr", fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port))
	}

	// Start gRPC server
	if s.deps.GRPCServer != nil {
		if err := s.deps.GRPCServer.Start(); err != nil {
			logger.Warn("failed to start gRPC server", "err", err)
		} else {
			logger.Info("gRPC API server initialized", "addr", cfg.Server.GRPCPort)
		}
	}

	logger.Info("AnubisWatch is ready. The judgment begins.")
	return nil
}

// Stop gracefully shuts down the server
func (s *Server) Stop(ctx context.Context) error {
	logger := s.logger

	logger.Info("shutting down...")

	// Stop REST server
	if s.deps.RESTServer != nil {
		s.deps.RESTServer.Stop(ctx)
	}

	// Stop gRPC server
	if s.deps.GRPCServer != nil {
		s.deps.GRPCServer.Stop()
	}

	// Stop journey executors
	if s.deps.JourneyExecutor != nil {
		s.deps.JourneyExecutor.StopAll()
	}

	// Stop alert manager
	if s.deps.AlertManager != nil {
		s.deps.AlertManager.Stop()
	}

	// Stop cluster manager
	if s.deps.ClusterManager != nil {
		s.deps.ClusterManager.Stop(ctx)
	}

	// Stop probe engine
	if s.deps.ProbeEngine != nil {
		s.deps.ProbeEngine.Stop()
	}

	// Shutdown authenticator
	if s.deps.Authenticator != nil {
		s.deps.Authenticator.Shutdown()
	}

	// Close storage
	if s.deps.Store != nil {
		s.deps.Store.Close()
	}

	logger.Info("⚖️  AnubisWatch stopped. The judgment rests.")
	return nil
}

// WaitForShutdown blocks until shutdown signal is received
func (s *Server) WaitForShutdown() {
	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	<-shutdownCtx.Done()
}

// ServerOptions holds options for building server dependencies
type ServerOptions struct {
	ConfigPath string
	Logger     *slog.Logger
}

// grpcProbeAdapter adapts probe.Engine to grpcapi.ProbeEngine
type grpcProbeAdapter struct {
	engine *probe.Engine
}

func (a *grpcProbeAdapter) ForceCheck(soulID string) (interface{}, error) {
	return a.engine.ForceCheck(soulID)
}

// grpcStorageAdapter wraps restStorageAdapter to return interface{} for gRPC compatibility
type grpcStorageAdapter struct {
	inner *restStorageAdapter
}

func (a *grpcStorageAdapter) GetSoulNoCtx(id string) (interface{}, error)     { return a.inner.GetSoulNoCtx(id) }
func (a *grpcStorageAdapter) ListSoulsNoCtx(ws string, o, l int) ([]interface{}, error) {
	souls, err := a.inner.ListSoulsNoCtx(ws, o, l)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(souls))
	for i, s := range souls {
		result[i] = s
	}
	return result, nil
}
func (a *grpcStorageAdapter) SaveSoulNoCtx(s interface{}) error    { return nil }
func (a *grpcStorageAdapter) DeleteSoulNoCtx(id string) error      { return a.inner.DeleteSoulNoCtx(id) }
func (a *grpcStorageAdapter) ListJudgmentsNoCtx(soulID string, start, end time.Time, limit int) ([]interface{}, error) {
	judgments, err := a.inner.ListJudgmentsNoCtx(soulID, start, end, limit)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(judgments))
	for i, j := range judgments {
		result[i] = j
	}
	return result, nil
}
func (a *grpcStorageAdapter) GetChannelNoCtx(id string) (interface{}, error) { return a.inner.GetChannelNoCtx(id) }
func (a *grpcStorageAdapter) ListChannelsNoCtx(ws string) ([]interface{}, error) {
	channels, err := a.inner.ListChannelsNoCtx(ws)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(channels))
	for i, c := range channels {
		result[i] = c
	}
	return result, nil
}
func (a *grpcStorageAdapter) SaveChannelNoCtx(ch interface{}) error { return nil }
func (a *grpcStorageAdapter) DeleteChannelNoCtx(id string) error    { return a.inner.DeleteChannelNoCtx(id) }
func (a *grpcStorageAdapter) GetRuleNoCtx(id string) (interface{}, error) { return a.inner.GetRuleNoCtx(id) }
func (a *grpcStorageAdapter) ListRulesNoCtx(ws string) ([]interface{}, error) {
	rules, err := a.inner.ListRulesNoCtx(ws)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(rules))
	for i, r := range rules {
		result[i] = r
	}
	return result, nil
}
func (a *grpcStorageAdapter) SaveRuleNoCtx(rule interface{}) error { return nil }
func (a *grpcStorageAdapter) DeleteRuleNoCtx(id string) error      { return a.inner.DeleteRuleNoCtx(id) }
func (a *grpcStorageAdapter) GetJourneyNoCtx(id string) (interface{}, error) { return a.inner.GetJourneyNoCtx(id) }
func (a *grpcStorageAdapter) ListJourneysNoCtx(ws string, o, l int) ([]interface{}, error) {
	journeys, err := a.inner.ListJourneysNoCtx(ws, o, l)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(journeys))
	for i, j := range journeys {
		result[i] = j
	}
	return result, nil
}
func (a *grpcStorageAdapter) SaveJourneyNoCtx(j interface{}) error { return nil }
func (a *grpcStorageAdapter) DeleteJourneyNoCtx(id string) error   { return a.inner.DeleteJourneyNoCtx(id) }
func (a *grpcStorageAdapter) ListEvents(soulID string, limit int) ([]interface{}, error) {
	events, err := a.inner.store.ListAlertEvents(soulID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]interface{}, len(events))
	for i, e := range events {
		result[i] = e
	}
	return result, nil
}
// BuildServerDependencies builds all server dependencies
func BuildServerDependencies(opts ServerOptions) (*ServerDependencies, error) {
	logger := opts.Logger
	if logger == nil {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, nil))
	}

	// Load config
	configPath := opts.ConfigPath
	if configPath == "" {
		configPath = os.Getenv("ANUBIS_CONFIG")
		if configPath == "" {
			configPath = "anubis.json"
		}
	}

	var cfg *core.Config
	if _, statErr := os.Stat(configPath); statErr == nil {
		var err error
		cfg, err = core.LoadConfig(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load config: %w", err)
		}
		logger.Info("config loaded", "path", configPath)
	} else {
		logger.Info("no config file found, using defaults", "path", configPath)
		cfg = core.GenerateDefaultConfig()
	}

	// Create data directory
	if err := os.MkdirAll(cfg.Storage.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Initialize storage
	store, err := storage.NewEngine(cfg.Storage, logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize auth
	sessionPath := filepath.Join(cfg.Storage.Path, "sessions.json")
	adminEmail := "admin@anubis.watch"
	adminPassword := "admin"
	if cfg.Auth.Local.AdminEmail != "" {
		adminEmail = cfg.Auth.Local.AdminEmail
	}
	if cfg.Auth.Local.AdminPassword != "" {
		adminPassword = cfg.Auth.Local.AdminPassword
	}

	var authenticator api.Authenticator
	switch cfg.Auth.Type {
	case "oidc":
		authenticator = auth.NewOIDCAuthenticator(cfg.Auth.OIDC, sessionPath, adminEmail, adminPassword)
		logger.Info("OIDC authentication initialized", "issuer", cfg.Auth.OIDC.Issuer)
	case "ldap":
		authenticator = auth.NewLDAPAuthenticator(cfg.Auth.LDAP, sessionPath, adminEmail, adminPassword)
		logger.Info("LDAP authentication initialized", "url", cfg.Auth.LDAP.URL)
	default:
		authenticator = auth.NewLocalAuthenticator(sessionPath, adminEmail, adminPassword)
	}

	// Initialize alert manager
	alertStorage := &alertStorageAdapter{store: store}
	alertMgr := alert.NewManager(alertStorage, logger)

	// Initialize probe engine
	registry := probe.NewCheckerRegistry()
	probeEngine := probe.NewEngine(probe.EngineOptions{
		Registry: registry,
		Store:    &probeStorageAdapter{store: store},
		Alerter:  alertMgr,
		NodeID:   cfg.Necropolis.NodeName,
		Region:   cfg.Necropolis.Region,
		Logger:   logger,
	})

	// Initialize journey executor
	journeyExec := journey.NewExecutor(store, logger)

	// Initialize cluster manager
	clusterMgr, err := cluster.NewManager(cfg.Necropolis, store, logger)
	if err != nil {
		logger.Warn("failed to initialize cluster manager", "err", err)
		clusterMgr = nil
	}

	// Initialize REST server dependencies
	restStore := &restStorageAdapter{store: store}
	clusterAdapt := &clusterAdapter{mgr: clusterMgr}

	// Initialize dashboard handler
	var dashboardHandler http.Handler
	if cfg.Dashboard.Enabled {
		dh, err := dashboard.NewHandler()
		if err != nil {
			logger.Warn("failed to initialize dashboard", "err", err)
		} else {
			dashboardHandler = dh
			logger.Info("dashboard handler initialized")
		}
	}

	// Initialize status page handler
	statusPageRepo := &statusPageRepository{store: store}
	acmeMgr := initACMEManager(cfg, store, logger)
	statusPageHandler := statuspage.NewHandler(statusPageRepo, acmeMgr)

	// Initialize MCP server
	mcpServer := api.NewMCPServer(restStore, probeEngine, alertMgr, logger)

	restServer := api.NewRESTServer(cfg.Server, cfg.Auth, restStore, probeEngine, alertMgr, authenticator, clusterAdapt, dashboardHandler, statusPageHandler, mcpServer, logger)

	// Wire up WebSocket broadcast for real-time judgment updates
	probeEngine.SetOnJudgment(restServer.OnJudgmentCallback())

	// Initialize gRPC server
	var grpcServer *grpcapi.Server
	if cfg.Server.GRPCPort > 0 {
		grpcStore := &grpcStorageAdapter{inner: restStore}
		grpcServer = grpcapi.NewServer(
			fmt.Sprintf(":%d", cfg.Server.GRPCPort),
			grpcStore,
			&grpcProbeAdapter{engine: probeEngine},
			logger,
		)
	}

	return &ServerDependencies{
		Config:            cfg,
		Logger:            logger,
		Store:             store,
		Authenticator:     authenticator,
		AlertManager:      alertMgr,
		ProbeEngine:       probeEngine,
		JourneyExecutor:   journeyExec,
		ClusterManager:    clusterMgr,
		RESTServer:        restServer,
		GRPCServer:        grpcServer,
		StatusPageRepo:    statusPageRepo,
		ACMEManager:       acmeMgr,
		DashboardHandler:  dashboardHandler,
		StatusPageHandler: statusPageHandler,
		MCPServer:         mcpServer,
	}, nil
}
