package interceptors

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

const auth = "authorization"

func PasswordAuth(passwordEnvVarName string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return ctx, status.Error(codes.Unauthenticated, "missing metadata")
		}

		auth := md.Get("authorization")
		authHeader := auth[0]
		if !strings.HasPrefix(authHeader, "Basic ") {
			return ctx, status.Error(codes.Unauthenticated, "invalid authorization scheme")
		}

		password, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(authHeader, "Basic "))
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid auth payload")
		}

		if os.Getenv(passwordEnvVarName) != string(password) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return handler(ctx, req)
	}
}

func IPWhitelist(allowed ...string) grpc.UnaryServerInterceptor {
	whitelist := make(map[string]struct{})
	for _, ip := range allowed {
		whitelist[ip] = struct{}{}
	}
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
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
