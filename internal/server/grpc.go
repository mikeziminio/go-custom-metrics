package server

import (
	"context"
	"fmt"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	unpb "github.com/mikeziminio/go-custom-metrics/internal/proto/metrics"
	"github.com/mikeziminio/go-custom-metrics/internal/trustedsubnet"
)

// GRPCServer represents the gRPC server for metrics.
type GRPCServer struct {
	unpb.UnimplementedMetricsServer
	storage Storage
	logger  *zap.Logger
	network *net.IPNet
	address string
}

// NewGRPCServer creates a new gRPC server instance.
func NewGRPCServer(storage Storage, logger *zap.Logger, address, trustedSubnet string) *GRPCServer {
	var network *net.IPNet
	if trustedSubnet != "" {
		var err error
		network, err = trustedsubnet.ParseCIDR(trustedSubnet)
		if err != nil {
			logger.Warn("Invalid trusted subnet CIDR, subnet will be ignored",
				zap.String("subnet", trustedSubnet), zap.Error(err))
		}
	}

	return &GRPCServer{
		storage: storage,
		logger:  logger,
		network: network,
		address: address,
	}
}

// UpdateMetrics implements the Metrics service RPC method.
func (s *GRPCServer) UpdateMetrics(ctx context.Context, req *unpb.UpdateMetricsRequest) (*unpb.UpdateMetricsResponse, error) {
	metrics := make([]model.Metric, 0, len(req.Metrics))
	for _, m := range req.Metrics {
		metrics = append(metrics, protoToModelMetric(m))
	}

	if err := s.storage.Updates(ctx, metrics); err != nil {
		s.logger.Error("failed to update metrics", zap.Error(err))
		return nil, err
	}

	return &unpb.UpdateMetricsResponse{}, nil
}

// Run starts the gRPC server.
func (s *GRPCServer) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", s.address)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	s.logger.Info("gRPC server started", zap.String("address", s.address))

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(trustedsubnet.GRPCMiddlewareHandler(s.network, s.logger)),
	)
	unpb.RegisterMetricsServer(srv, s)

	go func() {
		if err := srv.Serve(lis); err != nil {
			s.logger.Error("gRPC server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	srv.GracefulStop()
	s.logger.Info("gRPC server stopped")
	return nil
}

// protoToModelMetric converts a proto Metric to model.Metric.
func protoToModelMetric(m *unpb.Metric) model.Metric {
	metric := model.Metric{
		ID:    m.Id,
		MType: model.MetricType(m.Type.String()),
	}

	switch m.Type {
	case unpb.Metric_GAUGE:
		val := m.GetValue()
		metric.Value = &val
	case unpb.Metric_COUNTER:
		metric.Delta = &m.Delta
	}

	return metric
}
