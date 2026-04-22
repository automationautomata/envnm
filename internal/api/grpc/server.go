package grpc

import (
	"context"
	inc "envmn/internal/api/grpc/interceptors"
	"envmn/internal/service"
	"envmn/logs"
	pb "envmn/pkg/api/proto"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/selector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type Server struct {
	*grpc.Server
}

type Settigns struct {
	PasswordEnvVarName          string
	Logger                      logs.Logger
	Credentials                 credentials.TransportCredentials
	ManagementServiceAllowedIPs []string
}

func NewServer(distr *service.DistributionServices, mng *service.ManagementServices, settings Settigns) *Server {
	managementServiceName := pb.ManagementService_ServiceDesc.ServiceName

	interceptors := []grpc.UnaryServerInterceptor{
		recovery.UnaryServerInterceptor(),
		logging.UnaryServerInterceptor(&logDecorator{settings.Logger}),
		selector.UnaryServerInterceptor(
			inc.PasswordAuth(settings.PasswordEnvVarName),
			selector.MatchFunc(serviceMatcher(managementServiceName)),
		),
	}

	if len(settings.ManagementServiceAllowedIPs) != 0 {
		interceptors = append(interceptors,
			selector.UnaryServerInterceptor(
				inc.IPWhitelist(settings.ManagementServiceAllowedIPs...),
				selector.MatchFunc(serviceMatcher(managementServiceName)),
			),
		)
	}

	s := grpc.NewServer(
		grpc.Creds(settings.Credentials),
		grpc.ChainUnaryInterceptor(interceptors...),
	)
	pb.RegisterManagementServiceServer(s, newManagementServiceServer(mng, settings.Logger))
	pb.RegisterDistributionServiceServer(s, newDistributionServiceServer(distr, settings.Logger))
	return &Server{Server: s}
}

func serviceMatcher(name string) func(ctx context.Context, callMeta interceptors.CallMeta) bool {
	return func(ctx context.Context, callMeta interceptors.CallMeta) bool {
		return callMeta.Service == name
	}
}
