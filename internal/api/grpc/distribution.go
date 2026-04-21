package grpc

import (
	"context"
	grpcerrs "envmn/internal/api/grpc/errors"
	"envmn/internal/service"
	"envmn/internal/service/dto"
	"envmn/logs"
	pb "envmn/pkg/api/proto"
	"errors"

	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

var errInvalidAccessKey = errors.New("invalid access key")

type distributionServiceServer struct {
	pb.UnimplementedDistributionServiceServer
	svc *service.DistributionServices
	log logs.Logger
}

func newDistributionServiceServer(svc *service.DistributionServices, log logs.Logger) *distributionServiceServer {
	return &distributionServiceServer{svc: svc, log: log}
}

func (s *distributionServiceServer) GetClientVariables(ctx context.Context, req *pb.GetClientVariablesRequest) (*pb.GetClientVariablesResponse, error) {
	vars, err := s.svc.GetClientVariables(ctx, dto.GetClientVariablesInput{
		EnvironmentName: req.EnvironmentName,
		AccessKey:       req.AccessKey,
	})
	if handlerError, isInternal := grpcerrs.ToGRPCError(err); handlerError != nil {
		if isInternal {
			s.logClientError(ctx, req.EnvironmentName, err)
		}
		return nil, handlerError
	}

	return &pb.GetClientVariablesResponse{Variables: vars}, nil
}

func (s *distributionServiceServer) UpdateVariables(ctx context.Context, req *pb.UpdateVariablesRequest) (*emptypb.Empty, error) {
	err := s.svc.UpdateVariablesByClient(ctx, dto.UpdateVariablesByClientInput{
		EnvironmentName: req.EnvironmentName,
		AccessKey:       req.AccessKey,
		Variables:       req.Variables,
	})
	if handlerError, isInternal := grpcerrs.ToGRPCError(err); handlerError != nil {
		if isInternal {
			s.logClientError(ctx, req.EnvironmentName, err)
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, err
}

// func (s *distributionServiceServer) SubscribeOnUpdates(req *pb.SubscribeOnUpdatesRequest, stream pb.ClientService_SubscribeOnUpdatesServer) error {
// 	key, err := s.svc.SubscribeOnUpdates(stream.Context(), dto.SubscribeOnUpdatesInput{
// 		EnvironmentName: req.EnvironmentName,
// 		AccessKey:       req.AccessKey,
// 	})

// 	if handlerError, isInternal := grpcerrs.ToGRPCError(err); handlerError != nil {
// 		if isInternal {
// 			s.logClientError(req.AccessKey, req.EnvironmentName, err)
// 		}
// 		return handlerError
// 	}

// 	return stream.Send(&pb.SubscribeOnUpdatesResponse{Key: key})
// }

func (s *distributionServiceServer) logClientError(ctx context.Context, envName string, err error) {
	p, ok := peer.FromContext(ctx)
	distribution := "unknown"
	if ok {
		distribution = p.Addr.String()
	}
	s.log.Error(
		"cannot update distribution variables",
		logs.Args{"environment": envName, "error": err, "distribution": distribution},
	)
}
