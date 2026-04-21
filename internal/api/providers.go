package api

import (
	"envmn/internal/api/grpc"
	"envmn/internal/service"
)

func ProvideGRPCServer(
	distr *service.DistributionServices,
	mng *service.ManagementServices,
	settings grpc.Settigns,
) *grpc.Server {
	return grpc.NewServer(distr, mng, settings)
}
