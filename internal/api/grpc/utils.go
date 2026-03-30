package grpc

import (
	"envmn/internal/service/dto"
	pb "envmn/pkg/api/proto"
)

func environmentsToProto(envs []*dto.EnvironmentDTO) []*pb.Environment {
	res := make([]*pb.Environment, len(envs))

	for i, e := range envs {
		res[i] = &pb.Environment{
			Name:        e.Name,
			Description: e.Description,
			Variables:   e.Variables,
			Reserved:    e.Reserved,
		}
	}

	return res
}
