package grpc

import (
	"context"
	inc "envmn/internal/api/grpc/interceptors"
	"envmn/internal/service"
	"envmn/logs"
	pb "envmn/pkg/api/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	*grpc.Server
}

type ServerSettigns struct {
	Logger               logs.Logger
	Credentials          credentials.TransportCredentials
	ManagementAllowedIPs []string
}

func NewServer(client *service.Client, management *service.Management, settings ServerSettigns) (*Server, error) {
	managementServiceName := pb.ManagementService_ServiceDesc.ServiceName

	s := &Server{
		Server: grpc.NewServer(
			grpc.Creds(settings.Credentials),
			grpc.ChainUnaryInterceptor(
				selector.UnaryServerInterceptor(
					inc.IPWhitelist(settings.ManagementAllowedIPs...),
					selector.MatchFunc(serviceMatcher(managementServiceName)),
				),
			),
		),
	}

	pb.RegisterClientServiceServer(s, newClientServiceServer(client, settings.Logger))
	pb.RegisterManagementServiceServer(s, newManagementServiceServer(management, settings.Logger))
	return s, nil
}

func serviceMatcher(name string) func(ctx context.Context, callMeta interceptors.CallMeta) bool {
	return func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		return callMeta.Service == name
	}
}
