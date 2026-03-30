package grpc

import (
	"context"
	"envmn/internal/service"
	"envmn/internal/service/dto"
	"envmn/logs"
	pb "envmn/pkg/api/proto"

	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/emptypb"
)

type managementServiceServer struct {
	pb.UnimplementedManagementServiceServer
	svc *service.Management
	log logs.Logger
}

func newManagementServiceServer(svc *service.Management, log logs.Logger) *managementServiceServer {
	return &managementServiceServer{svc: svc, log: log}
}

// ------------------------------- Environment -------------------------------

func (s *managementServiceServer) CreateEnvironment(ctx context.Context, req *pb.CreateEnvironmentRequest) (*pb.CreateEnvironmentResponse, error) {
	id, err := s.svc.CreateEnvironment(ctx, dto.CreateEnvironmentInput{
		Name:        req.Name,
		Description: req.Description,
		Variables:   req.Variables,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot create environment", logs.Args{"error": err})
		}
		return nil, handlerError
	}

	return &pb.CreateEnvironmentResponse{
		Id: id.String(),
	}, nil
}

func (s *managementServiceServer) GetAllEnvironments(ctx context.Context, req *pb.GetAllEnvironmentsRequest) (*pb.GetAllEnvironmentsResponse, error) {
	envs, err := s.svc.GetAllEnvironments(ctx, dto.GetAllEnvironmentsInput{
		Reserved: req.Reserved,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot get all environments", logs.Args{"error": err})
		}
		return nil, handlerError
	}

	return &pb.GetAllEnvironmentsResponse{
		Environments: environmentsToProto(envs),
	}, nil
}

func (s *managementServiceServer) DeleteEnvironment(ctx context.Context, req *pb.DeleteEnvironmentRequest) (*emptypb.Empty, error) {
	err := s.svc.DeleteEnvironment(ctx, dto.DeleteEnvironmentInput{
		Name: req.Name,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot delete environment", logs.Args{"environment": req.Name, "error": err})
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, nil
}

func (s *managementServiceServer) UpdateEnvironmentInfo(ctx context.Context, req *pb.UpdateEnvironmentInfoRequest) (*emptypb.Empty, error) {
	err := s.svc.UpdateEnvironmentInfo(ctx, dto.UpdateEnvironmentInfoInput{
		OldName:     req.OldName,
		NewName:     req.NewName,
		Description: req.Description,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot update environment info", logs.Args{"environment": req.OldName, "error": err})
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, nil
}

// ------------------------------- Access Policy -------------------------------

func (s *managementServiceServer) CreateAccessPolicy(ctx context.Context, req *pb.CreateAccessPolicyRequest) (*pb.CreateAccessPolicyResponse, error) {
	id, err := s.svc.CreateAccessPolicy(ctx, dto.CreateAccessPolicyInput{
		Name:           req.Name,
		Key:            req.Key,
		ChangesAllowed: req.ChangesAllowed,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot create access policy", logs.Args{"name": req.Name, "error": err})
		}
		return nil, handlerError
	}

	return &pb.CreateAccessPolicyResponse{
		Id: id.String(),
	}, nil
}

func (s *managementServiceServer) RemovePolicy(ctx context.Context, req *pb.RemovePolicyRequest) (*emptypb.Empty, error) {
	policyID, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, err
	}

	err = s.svc.RemovePolicy(ctx, dto.RemovePolicyInput{
		ID: policyID,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot remove policy", logs.Args{"policy_id": req.Id, "error": err})
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, nil
}

func (s *managementServiceServer) AddPolicyToEnvironment(ctx context.Context, req *pb.AddPolicyToEnvironmentRequest) (*emptypb.Empty, error) {
	policyID, err := uuid.Parse(req.PolicyId)
	if err != nil {
		return nil, err
	}

	err = s.svc.AddPolicyToEnvironment(ctx, dto.AddPolicyToEnvironmentInput{
		EnvironmentName: req.EnvironmentName,
		PolicyID:        policyID,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot add policy to environment",
				logs.Args{"environment": req.EnvironmentName, "policy_id": req.PolicyId, "error": err})
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, nil
}

func (s *managementServiceServer) RemovePolicyFromEnvironment(ctx context.Context, req *pb.RemovePolicyFromEnvironmentRequest) (*emptypb.Empty, error) {
	policyID, err := uuid.Parse(req.PolicyId)
	if err != nil {
		return nil, err
	}

	err = s.svc.RemovePolicyFromEnvironment(ctx, dto.RemovePolicyFromEnvironmentInput{
		EnvironmentName: req.EnvironmentName,
		PolicyID:        policyID,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot remove policy from environment",
				logs.Args{"environment": req.EnvironmentName, "policy_id": req.PolicyId, "error": err})
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, nil
}

// ------------------------------- Variables -------------------------------

func (s *managementServiceServer) UpdateEnvironmentVariables(ctx context.Context, req *pb.UpdateEnvironmentVariablesRequest) (*emptypb.Empty, error) {
	err := s.svc.UpdateEnvironmentVariables(ctx, dto.UpdateEnvironmentVariablesInput{
		EnvironmentName: req.EnvironmentName,
		Variables:       req.Variables,
		AccessKey:       req.AccessKey,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot update environment variables",
				logs.Args{"environment": req.EnvironmentName, "error": err})
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, nil
}

func (s *managementServiceServer) RemoveVariableFromEnvironment(ctx context.Context, req *pb.RemoveVariableFromEnvironmentRequest) (*emptypb.Empty, error) {
	err := s.svc.RemoveVariableFromEnvironment(ctx, dto.RemoveVariableFromEnvironmentInput{
		EnvironmentName: req.EnvironmentName,
		VariableKey:     req.VariableKey,
	})
	if handlerError, isInternal := toGRPCError(err); handlerError != nil {
		if isInternal {
			s.log.Error("cannot remove variable from environment",
				logs.Args{"environment": req.EnvironmentName, "key": req.VariableKey, "error": err})
		}
		return nil, handlerError
	}
	return &emptypb.Empty{}, nil
}
