package main

import (
	"context"
	"encoding/base64"
	grpcinc "envmn/internal/api/grpc/interceptors"
	"fmt"
)

type basicAuth struct {
	password string
}

func (b *basicAuth) GetRequestMetadata(ctx context.Context, _ ...string) (map[string]string, error) {
	enc := base64.StdEncoding.EncodeToString([]byte(b.password))
	return map[string]string{
		grpcinc.AuthorizationName: fmt.Sprintf("Basic %s", enc),
	}, nil
}

func (b *basicAuth) RequireTransportSecurity() bool {
	return true
}
