package grpc

import (
	"context"
	"envmn/internal/service"
	"envmn/internal/service/dto"
	"envmn/logs"
	pb "envmn/pkg/api/proto"
	"errors"

	"google.golang.org/grpc/peer"
	"google.golang.org/protobuf/types/known/emptypb"
)

var errInvalidAccessKey = errors.New("invalid access key")

type clientServiceServer struct {
	pb.UnimplementedClientServiceServer
	svc *service.Client
	log logs.Logger
}

func newClientServiceServer(svc *service.Client, log logs.Logger) *clientServiceServer {
	return &clientServiceServer{svc: svc, log: log}
}

func (s *clientServiceServer) GetClientVariables(ctx context.Context, req *pb.GetClientVariablesRequest) (*pb.GetClientVariablesResponse, error) {
	vars, err := s.svc.GetClientVariables(ctx, dto.GetClientVariablesInput{
		EnvironmentName: req.EnvironmentName,
		AccessKey:       req.AccessKey,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.logClientError(ctx, req.EnvironmentName, err)
		}
		return nil, handlerError
	}

	return &pb.GetClientVariablesResponse{Variables: vars}, nil
}

func (s *clientServiceServer) UpdateVariables(ctx context.Context, req *pb.UpdateVariablesRequest) (*emptypb.Empty, error) {
	err := s.svc.UpdateVariablesByClient(ctx, dto.UpdateVariablesByClientInput{
		EnvironmentName: req.EnvironmentName,
		AccessKey:       req.AccessKey,
		Variables:       req.Variables,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.logClientError(ctx, req.EnvironmentName, err)
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, err
}

// func (s *clientServiceServer) SubscribeOnUpdates(req *pb.SubscribeOnUpdatesRequest, stream pb.ClientService_SubscribeOnUpdatesServer) error {
// 	key, err := s.svc.SubscribeOnUpdates(stream.Context(), dto.SubscribeOnUpdatesInput{
// 		EnvironmentName: req.EnvironmentName,
// 		AccessKey:       req.AccessKey,
// 	})

// 	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
// 		if isInternal {
// 			s.logClientError(req.AccessKey, req.EnvironmentName, err)
// 		}
// 		return handlerError
// 	}

// 	return stream.Send(&pb.SubscribeOnUpdatesResponse{Key: key})
// }

func (s *clientServiceServer) logClientError(ctx context.Context, envName string, err error) {
	p, ok := peer.FromContext(ctx)
	client := "unknown"
	if ok {
		client = p.Addr.String()
	}
	s.log.Error(
		"cannot update client variables",
		logs.Args{"environment": envName, "error": err, "client": client},
	)
}
