package interceptors

import (
	"context"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func IPWhitelist(allowed ...string) grpc.UnaryServerInterceptor {
	whitelist := make(map[string]struct{})
	for _, ip := range allowed {
		whitelist[ip] = struct{}{}
	}
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		p, ok := peer.FromContext(ctx)
		if !ok || p.Addr == nil {
			return nil, status.Error(codes.PermissionDenied, "cannot get peer info")
		}

		clientIP := getIPFromAddr(p.Addr.String())

		if _, ok := whitelist[clientIP]; ok {
			return handler(ctx, req)
		}
		return nil, status.Errorf(codes.PermissionDenied, "ip %s is not whitelisted", clientIP)
	}
}

func getIPFromAddr(addr string) string {
	if idx := strings.LastIndex(addr, ":"); idx != -1 {
		return addr[:idx]
	}
	return addr
}
