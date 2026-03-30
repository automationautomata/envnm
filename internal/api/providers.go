package api

import (
	grpcapi "envmn/internal/api/grpc"
	"envmn/internal/service"
)

func ProvideGRPCServer(client *service.Client, mgmt *service.Management, settings grpcapi.ServerSettigns) (*grpcapi.Server, error) {
	return grpcapi.NewServer(client, mgmt, settings)
}
