package main

import (
	"fmt"

	"envmn/config"
	grpcauth "envmn/internal/api/grpc/auth"
	pb "envmn/pkg/api/proto"

	"google.golang.org/grpc"
)

type gRPCClient struct {
	Management pb.ManagementServiceClient
	conn       *grpc.ClientConn
}

func newGRPCClient(cfg config.CLIConfig) (*gRPCClient, error) {
	addr := cfg.Server.Addr()

	certConf := cfg.Certificate
	creds, err := grpcauth.NewMTLSCredentials(certConf.CACertPath, certConf.CertPath, certConf.KeyPath)
	if err != nil {
		return nil, err
	}

	basicAuth := &basicAuth{password: cfg.Auth.Password}

	conn, err := grpc.NewClient(addr,
		grpc.WithTransportCredentials(creds),
		grpc.WithPerRPCCredentials(basicAuth),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", addr, err)
	}

	return &gRPCClient{
		Management: pb.NewManagementServiceClient(conn),
		conn:       conn,
	}, nil
}

func (c *gRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
