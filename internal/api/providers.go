package api

import (
	"envmn/internal/api/grpc"
	grpcapi "envmn/internal/api/grpc"
	"envmn/internal/service"
)

func ProvideGRPCServer(
	distr *service.DistributionServices,
	mng *service.ManagementServices,
	settings grpc.Settigns,
) *grpcapi.Server {
	return grpc.NewServer(distr, mng, settings)
}
