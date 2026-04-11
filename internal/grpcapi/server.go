// Package grpcapi provides a gRPC API server for AnubisWatch
package grpcapi

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"

	v1 "github.com/AnubisWatch/anubiswatch/internal/grpcapi/v1"
)

// Server implements the gRPC AnubisWatchService
type Server struct {
	v1.UnimplementedAnubisWatchServiceServer
	mu       sync.RWMutex
	grpc     *grpc.Server
	listener net.Listener
	addr     string
	logger   *slog.Logger

	store Store
	probe ProbeEngine
}

// Store defines the storage operations available to the gRPC server
type Store interface {
	GetSoulNoCtx(id string) (interface{}, error)
	ListSoulsNoCtx(workspace string, offset, limit int) ([]interface{}, error)
	SaveSoulNoCtx(soul interface{}) error
	DeleteSoulNoCtx(id string) error

	ListJudgmentsNoCtx(soulID string, start, end time.Time, limit int) ([]interface{}, error)

	GetChannelNoCtx(id string) (interface{}, error)
	ListChannelsNoCtx(workspace string) ([]interface{}, error)
	SaveChannelNoCtx(ch interface{}) error
	DeleteChannelNoCtx(id string) error

	GetRuleNoCtx(id string) (interface{}, error)
	ListRulesNoCtx(workspace string) ([]interface{}, error)
	SaveRuleNoCtx(rule interface{}) error
	DeleteRuleNoCtx(id string) error

	GetJourneyNoCtx(id string) (interface{}, error)
	ListJourneysNoCtx(workspace string, offset, limit int) ([]interface{}, error)
	SaveJourneyNoCtx(j interface{}) error
	DeleteJourneyNoCtx(id string) error
}

// ProbeEngine interface for probe operations
type ProbeEngine interface {
	ForceCheck(soulID string) (interface{}, error)
}

// AlertManager interface (reserved for future use)
type AlertManager interface{}

// NewServer creates a new gRPC server
func NewServer(addr string, store Store, probe ProbeEngine, logger *slog.Logger) *Server {
	s := &Server{
		addr:   addr,
		logger: logger,
		store:  store,
		probe:  probe,
	}
	s.grpc = grpc.NewServer()
	v1.RegisterAnubisWatchServiceServer(s.grpc, s)
	reflection.Register(s.grpc)
	return s
}

// Start starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", s.addr, err)
	}
	s.listener = lis
	if s.logger != nil {
		s.logger.Info("gRPC server starting", "addr", s.addr)
	}
	go func() {
		if err := s.grpc.Serve(lis); err != nil {
			if s.logger != nil {
				s.logger.Error("gRPC server error", "err", err)
			}
		}
	}()
	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpc != nil {
		s.grpc.GracefulStop()
	}
}

// Helper: convert time to timestamppb
func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

// soulToPB converts a core.Soul to protobuf Soul
func soulToPB(s interface{}) *v1.Soul {
	// Type assertion to core.Soul would go here
	// For now, the concrete implementation lives in the adapter
	return nil
}

// judgmentToPB converts a core.Judgment to protobuf Judgment
func judgmentToPB(j interface{}) *v1.Judgment {
	return nil
}

// channelToPB converts a core.AlertChannel to protobuf Channel
func channelToPB(c interface{}) *v1.Channel {
	return nil
}

// ruleToPB converts a core.AlertRule to protobuf Rule
func ruleToPB(r interface{}) *v1.Rule {
	return nil
}

// journeyToPB converts a core.JourneyConfig to protobuf Journey
func journeyToPB(j interface{}) *v1.Journey {
	return nil
}

// --- Soul RPCs ---

func (s *Server) ListSouls(ctx context.Context, req *v1.ListSoulsRequest) (*v1.ListSoulsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workspace := "default"
	if req.Workspace != nil {
		workspace = *req.Workspace
	}

	limit := int(req.Limit)
	if limit == 0 {
		limit = 20
	}

	souls, err := s.store.ListSoulsNoCtx(workspace, int(req.Offset), limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list souls: %v", err)
	}

	pbSouls := make([]*v1.Soul, 0, len(souls))
	for _, soul := range souls {
		if pb := soulToPB(soul); pb != nil {
			pbSouls = append(pbSouls, pb)
		}
	}

	hasMore := len(pbSouls) >= limit
	var nextOffset *int32
	if hasMore {
		off := int32(int(req.Offset) + limit)
		nextOffset = &off
	}

	return &v1.ListSoulsResponse{
		Souls: pbSouls,
		Pagination: &v1.Pagination{
			Total:      int32(len(souls)),
			Offset:     req.Offset,
			Limit:      int32(limit),
			HasMore:    hasMore,
			NextOffset: nextOffset,
		},
	}, nil
}

func (s *Server) GetSoul(ctx context.Context, req *v1.GetSoulRequest) (*v1.Soul, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	soul, err := s.store.GetSoulNoCtx(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "soul not found: %s", req.Id)
	}
	if pb := soulToPB(soul); pb != nil {
		return pb, nil
	}
	return nil, status.Errorf(codes.Internal, "failed to convert soul")
}

func (s *Server) CreateSoul(ctx context.Context, req *v1.CreateSoulRequest) (*v1.Soul, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) UpdateSoul(ctx context.Context, req *v1.UpdateSoulRequest) (*v1.Soul, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) DeleteSoul(ctx context.Context, req *v1.DeleteSoulRequest) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.store.DeleteSoulNoCtx(req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete soul: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Judgment RPCs ---

func (s *Server) ListJudgments(ctx context.Context, req *v1.ListJudgmentsRequest) (*v1.ListJudgmentsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	limit := int(req.Limit)
	if limit == 0 {
		limit = 20
	}

	var start, end time.Time
	if req.Since != nil {
		start = req.Since.AsTime()
	}
	if req.Until != nil {
		end = req.Until.AsTime()
	}

	soulID := ""
	if req.SoulId != nil {
		soulID = *req.SoulId
	}

	judgments, err := s.store.ListJudgmentsNoCtx(soulID, start, end, limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list judgments: %v", err)
	}

	pbJudgments := make([]*v1.Judgment, 0, len(judgments))
	for _, j := range judgments {
		if pb := judgmentToPB(j); pb != nil {
			pbJudgments = append(pbJudgments, pb)
		}
	}

	return &v1.ListJudgmentsResponse{
		Judgments: pbJudgments,
		Pagination: &v1.Pagination{
			Total:   int32(len(judgments)),
			Offset:  req.Offset,
			Limit:   int32(limit),
			HasMore: len(judgments) >= limit,
		},
	}, nil
}

func (s *Server) GetSoulJudgments(ctx context.Context, req *v1.GetSoulJudgmentsRequest) (*v1.ListJudgmentsResponse, error) {
	return s.ListJudgments(ctx, &v1.ListJudgmentsRequest{
		Offset: req.Offset,
		Limit:  req.Limit,
		SoulId: &req.SoulId,
	})
}

func (s *Server) JudgeSoul(ctx context.Context, req *v1.JudgeSoulRequest) (*v1.Judgment, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result, err := s.probe.ForceCheck(req.SoulId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to judge soul: %v", err)
	}
	if pb := judgmentToPB(result); pb != nil {
		return pb, nil
	}
	return nil, status.Errorf(codes.Internal, "failed to convert judgment")
}

// --- Verdict RPCs ---

func (s *Server) ListVerdicts(ctx context.Context, req *v1.ListVerdictsRequest) (*v1.ListVerdictsResponse, error) {
	return &v1.ListVerdictsResponse{
		Pagination: &v1.Pagination{},
	}, status.Errorf(codes.Unimplemented, "verdict listing not yet available via gRPC")
}

// --- Channel RPCs ---

func (s *Server) ListChannels(ctx context.Context, req *v1.ListChannelsRequest) (*v1.ListChannelsResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workspace := ""
	if req.Workspace != nil {
		workspace = *req.Workspace
	}

	channels, err := s.store.ListChannelsNoCtx(workspace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list channels: %v", err)
	}

	pbChannels := make([]*v1.Channel, 0, len(channels))
	for _, ch := range channels {
		if pb := channelToPB(ch); pb != nil {
			pbChannels = append(pbChannels, pb)
		}
	}

	return &v1.ListChannelsResponse{
		Channels:   pbChannels,
		Pagination: &v1.Pagination{Total: int32(len(channels)), Limit: int32(len(channels))},
	}, nil
}

func (s *Server) GetChannel(ctx context.Context, req *v1.GetChannelRequest) (*v1.Channel, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ch, err := s.store.GetChannelNoCtx(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "channel not found: %s", req.Id)
	}
	if pb := channelToPB(ch); pb != nil {
		return pb, nil
	}
	return nil, status.Errorf(codes.Internal, "failed to convert channel")
}

func (s *Server) CreateChannel(ctx context.Context, req *v1.CreateChannelRequest) (*v1.Channel, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) UpdateChannel(ctx context.Context, req *v1.UpdateChannelRequest) (*v1.Channel, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) DeleteChannel(ctx context.Context, req *v1.DeleteChannelRequest) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.store.DeleteChannelNoCtx(req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete channel: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Rule RPCs ---

func (s *Server) ListRules(ctx context.Context, req *v1.ListRulesRequest) (*v1.ListRulesResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workspace := ""
	if req.Workspace != nil {
		workspace = *req.Workspace
	}

	rules, err := s.store.ListRulesNoCtx(workspace)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list rules: %v", err)
	}

	pbRules := make([]*v1.Rule, 0, len(rules))
	for _, r := range rules {
		if pb := ruleToPB(r); pb != nil {
			pbRules = append(pbRules, pb)
		}
	}

	return &v1.ListRulesResponse{
		Rules:      pbRules,
		Pagination: &v1.Pagination{Total: int32(len(rules)), Limit: int32(len(rules))},
	}, nil
}

func (s *Server) GetRule(ctx context.Context, req *v1.GetRuleRequest) (*v1.Rule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, err := s.store.GetRuleNoCtx(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "rule not found: %s", req.Id)
	}
	if pb := ruleToPB(r); pb != nil {
		return pb, nil
	}
	return nil, status.Errorf(codes.Internal, "failed to convert rule")
}

func (s *Server) CreateRule(ctx context.Context, req *v1.CreateRuleRequest) (*v1.Rule, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) UpdateRule(ctx context.Context, req *v1.UpdateRuleRequest) (*v1.Rule, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) DeleteRule(ctx context.Context, req *v1.DeleteRuleRequest) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.store.DeleteRuleNoCtx(req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete rule: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Journey RPCs ---

func (s *Server) ListJourneys(ctx context.Context, req *v1.ListJourneysRequest) (*v1.ListJourneysResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	workspace := ""
	if req.Workspace != nil {
		workspace = *req.Workspace
	}

	limit := int(req.Limit)
	if limit == 0 {
		limit = 20
	}

	journeys, err := s.store.ListJourneysNoCtx(workspace, int(req.Offset), limit)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list journeys: %v", err)
	}

	pbJourneys := make([]*v1.Journey, 0, len(journeys))
	for _, j := range journeys {
		if pb := journeyToPB(j); pb != nil {
			pbJourneys = append(pbJourneys, pb)
		}
	}

	return &v1.ListJourneysResponse{
		Journeys:   pbJourneys,
		Pagination: &v1.Pagination{Total: int32(len(journeys)), Limit: int32(limit)},
	}, nil
}

func (s *Server) GetJourney(ctx context.Context, req *v1.GetJourneyRequest) (*v1.Journey, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	j, err := s.store.GetJourneyNoCtx(req.Id)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "journey not found: %s", req.Id)
	}
	if pb := journeyToPB(j); pb != nil {
		return pb, nil
	}
	return nil, status.Errorf(codes.Internal, "failed to convert journey")
}

func (s *Server) CreateJourney(ctx context.Context, req *v1.CreateJourneyRequest) (*v1.Journey, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) UpdateJourney(ctx context.Context, req *v1.UpdateJourneyRequest) (*v1.Journey, error) {
	return nil, status.Errorf(codes.Unimplemented, "use REST API for mutations")
}

func (s *Server) DeleteJourney(ctx context.Context, req *v1.DeleteJourneyRequest) (*emptypb.Empty, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.store.DeleteJourneyNoCtx(req.Id); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete journey: %v", err)
	}
	return &emptypb.Empty{}, nil
}

// --- Cluster RPCs ---

func (s *Server) GetClusterStatus(ctx context.Context, req *emptypb.Empty) (*v1.ClusterStatus, error) {
	return &v1.ClusterStatus{
		Clustered: false,
		IsLeader:  true,
		NodeId:    "single-node",
		NodeCount: 1,
	}, nil
}

// --- Streaming RPCs ---

func (s *Server) StreamJudgments(req *v1.StreamRequest, stream v1.AnubisWatchService_StreamJudgmentsServer) error {
	return status.Errorf(codes.Unimplemented, "judgment streaming not yet implemented")
}

func (s *Server) StreamVerdicts(req *v1.StreamRequest, stream v1.AnubisWatchService_StreamVerdictsServer) error {
	return status.Errorf(codes.Unimplemented, "verdict streaming not yet implemented")
}
