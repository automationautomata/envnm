package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"envmn/config"
	pb "envmn/pkg/api/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type gRPCClient struct {
	Management pb.ManagementServiceClient
	ClientSvc  pb.ClientServiceClient
	conn       *grpc.ClientConn
}

func newGRPCClient(cfg config.CLIClientConfig) (*gRPCClient, error) {
	address := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)

	cert, err := tls.LoadX509KeyPair(cfg.Certificate.CertPath, cfg.Certificate.KeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	ca, err := os.ReadFile(cfg.Certificate.CACertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caPool := x509.NewCertPool()
	if !caPool.AppendCertsFromPEM(ca) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caPool,
		ServerName:   "localhost",
		MinVersion:   tls.VersionTLS13,
	}
	creds := credentials.NewTLS(tlsConfig)

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(creds))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	return &gRPCClient{
		Management: pb.NewManagementServiceClient(conn),
		ClientSvc:  pb.NewClientServiceClient(conn),
		conn:       conn,
	}, nil
}

func (c *gRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
