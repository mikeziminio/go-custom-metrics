package trustedsubnet

import (
	"context"
	"net"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GRPCMiddlewareHandler creates a gRPC UnaryServerInterceptor that checks
// if the client IP is within the trusted subnet.
//
// The middleware reads the IP address from the "x-real-ip" metadata key.
// If network is nil, all requests are allowed.
// If the client IP is not in the trusted subnet, returns codes.PermissionDenied.
func GRPCMiddlewareHandler(network *net.IPNet, logger *zap.Logger) grpc.UnaryServerInterceptor {
	if network == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			logger.Warn("no metadata in request")
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip header")
		}

		ips := md.Get("x-real-ip")
		if len(ips) == 0 {
			logger.Warn("no x-real-ip in metadata")
			return nil, status.Error(codes.PermissionDenied, "missing x-real-ip")
		}

		ipStr := ips[0]
		ip := net.ParseIP(ipStr)
		if ip == nil {
			logger.Warn("invalid x-real-ip", zap.String("ip", ipStr))
			return nil, status.Error(codes.PermissionDenied, "invalid x-real-ip")
		}

		if !network.Contains(ip) {
			logger.Warn("client IP not in trusted subnet", zap.String("ip", ipStr))
			return nil, status.Error(codes.PermissionDenied, "permission denied")
		}

		return handler(ctx, req)
	}
}
