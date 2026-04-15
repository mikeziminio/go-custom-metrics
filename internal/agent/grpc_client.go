package agent

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/mikeziminio/go-custom-metrics/internal/model"
	unpb "github.com/mikeziminio/go-custom-metrics/internal/proto/metrics"
)

// GRPCClient represents a gRPC client for sending metrics.
type GRPCClient struct {
	client  unpb.MetricsClient
	conn    *grpc.ClientConn
	localIP string
	logger  *zap.Logger
}

// NewGRPCClient creates a new gRPC client.
func NewGRPCClient(address string, localIP string, logger *zap.Logger) (*GRPCClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	return &GRPCClient{
		client:  unpb.NewMetricsClient(conn),
		conn:    conn,
		localIP: localIP,
		logger:  logger,
	}, nil
}

// SendMetrics sends a batch of metrics to the gRPC server.
func (c *GRPCClient) SendMetrics(ctx context.Context, metrics []model.Metric) error {
	req := &unpb.UpdateMetricsRequest{
		Metrics: make([]*unpb.Metric, 0, len(metrics)),
	}

	for _, m := range metrics {
		req.Metrics = append(req.Metrics, modelToProtoMetric(m))
	}

	md := metadata.Pairs("x-real-ip", c.localIP)
	ctx = metadata.NewOutgoingContext(ctx, md)

	_, err := c.client.UpdateMetrics(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send metrics: %w", err)
	}

	return nil
}

// Close closes the gRPC connection.
func (c *GRPCClient) Close() error {
	return c.conn.Close()
}

// modelToProtoMetric converts a model.Metric to proto.Metric.
func modelToProtoMetric(m model.Metric) *unpb.Metric {
	metric := &unpb.Metric{
		Id:   m.ID,
		Type: unpb.Metric_MType(0),
	}

	switch m.MType {
	case model.Gauge:
		metric.Type = unpb.Metric_GAUGE
		if m.Value != nil {
			metric.Value = *m.Value
		}
	case model.Counter:
		metric.Type = unpb.Metric_COUNTER
		if m.Delta != nil {
			metric.Delta = *m.Delta
		}
	}

	return metric
}
