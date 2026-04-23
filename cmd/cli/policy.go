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
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create access policy",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&name, "name", "", "Policy name")

	markFlagsRequired(cmd, "name")
	return cmd
}

func newFindPolicyCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "find",
		Short: "find policy by policy",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&name, "name", "", "Policy name")
	markFlagsRequired(cmd, "name")
	return cmd
}

func newGetAllEnvironmentsCmd() *cobra.Command {
	var idStr string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "get all policy environments",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&idStr, "id", "", "Policy id")
	markFlagsRequired(cmd, "id")
	return cmd
}

func newRemovePolicyCmd() *cobra.Command {
	var idStr string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove access policy",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&idStr, "id", "", "Policy ID")
	markFlagsRequired(cmd, "id")
	return cmd
}

func newAddPolicyToEnvCmd() *cobra.Command {
	var (
		envName, idStr string
		premission     policyPremission
	)
	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Add policy to environment",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&envName, "env", "", "Environment name")
	cmd.Flags().StringVar(&idStr, "id", "", "Policy ID")
	cmd.Flags().Var(&premission, "premission", "policy premission")
	markFlagsRequired(cmd, "env", "id", "premission")
	return cmd
}

func newRemovePolicyFromEnvCmd() *cobra.Command {
	var envName, idStr string

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Remove policy from environment",
		RunE: func(cmd *cobra.Command, args []string) error {
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

	cmd.Flags().StringVar(&envName, "env", "", "Environment name")
	cmd.Flags().StringVar(&idStr, "id", "", "Policy ID")
	markFlagsRequired(cmd, "env", "id")
	return cmd
}
