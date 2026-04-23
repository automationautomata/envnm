package main

import (
	"context"
	pb "envmn/pkg/api/proto"
	"fmt"

	"github.com/spf13/cobra"
)

const (
	readPremission  string = "read"
	writePremission string = "write"
)

type policyPremission string

func (m *policyPremission) String() string { return string(*m) }
func (m *policyPremission) Type() string   { return "policy_premission" }
func (m *policyPremission) Set(v string) error {
	switch v {
	case readPremission, writePremission:
		*m = policyPremission(v)
		return nil
	default:
		return fmt.Errorf("must be '%q' or '%q'", readPremission, writePremission)
	}
}

func policyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Access policy operations",
	}
	{
		cmd.AddCommand(newCreatePolicyCmd())
		cmd.AddCommand(newRemovePolicyCmd())
		cmd.AddCommand(newAddPolicyToEnvCmd())
		cmd.AddCommand(newRemovePolicyFromEnvCmd())
	}
	return cmd
}

func newCreatePolicyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "create [policy name]",
		Short: "Create access policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]

			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.CreateAccessPolicyRequest{Name: name}
			id, err := client.Management.CreateAccessPolicy(context.Background(), req)
			if err != nil {
				return err
			}
			cmd.Println(id)
			return nil
		},
	}
}

func newFindPolicyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "find [policy name]",
		Short: "find policy by policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.GetPolicyByNameRequest{Name: name}
			resp, err := client.Management.GetPolicyByName(context.Background(), req)
			if err != nil {
				return err
			}
			cmd.Printf("Id=%s,\n Name=%s,\n Key=%s", resp.Id, resp.Name, resp.Key)
			return nil
		},
	}
}

func newGetAllEnvironmentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list [policy id]",
		Short: "get all policy environments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idStr := args[0]
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.ListPolicyEnvironmentsRequest{PolicyId: idStr}
			resp, err := client.Management.ListPolicyEnvironments(context.Background(), req)
			if err != nil {
				return err
			}

			maxNameLen := 5
			for _, item := range resp.Items {
				maxNameLen = max(len(item.EnvironmentName), maxNameLen)
			}

			cmd.Printf("|%-*s| Premission\n", maxNameLen, "Name")
			for _, item := range resp.Items {
				prem := readPremission
				if item.ChangesPermission {
					prem = writePremission
				}
				cmd.Printf("|%-*s| %s\n", maxNameLen, item.EnvironmentName, prem)
			}
			return nil
		},
	}
}

func newRemovePolicyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [policy id]",
		Short: "Remove access policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			idStr := args[0]
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.RemovePolicyRequest{Id: idStr}
			_, err = client.Management.RemovePolicy(context.Background(), req)
			return err
		},
	}
}

func newAddPolicyToEnvCmd() *cobra.Command {
	var premission policyPremission

	cmd := &cobra.Command{
		Use:   "attach [environment name] [policy id]",
		Short: "Add policy to environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			envName, idStr := args[0], args[1]
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			canChange := premission.String() == writePremission

			req := &pb.AddPolicyToEnvironmentRequest{
				EnvironmentName:   envName,
				PolicyId:          idStr,
				ChangesPermission: canChange,
			}
			_, err = client.Management.AddPolicyToEnvironment(context.Background(), req)
			return err
		},
	}

	cmd.Flags().Var(&premission, "premission", "policy premission")
	return cmd
}

func newRemovePolicyFromEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "detach [environment name] [policy id]",
		Short: "Remove policy from environment",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			envName, idStr := args[0], args[1]
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.RemovePolicyFromEnvironmentRequest{
				EnvironmentName: envName,
				PolicyId:        idStr,
			}
			_, err = client.Management.RemovePolicyFromEnvironment(context.Background(), req)
			return err
		},
	}
}
