package main

import (
	"context"
	pb "envmn/pkg/api/proto"

	"github.com/spf13/cobra"
)

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
	cmd.MarkFlagRequired("name")
	if err := cmd.MarkFlagRequired("name"); err != nil {
		panic(err)
	}
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
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(err)
	}
	return cmd
}

func newAddPolicyToEnvCmd() *cobra.Command {
	var envName, idStr string

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Add policy to environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.AddPolicyToEnvironmentRequest{
				EnvironmentName: envName,
				PolicyId:        idStr,
			}
			_, err = client.Management.AddPolicyToEnvironment(context.Background(), req)
			return err
		},
	}

	cmd.Flags().StringVar(&envName, "env", "", "Environment name")
	cmd.Flags().StringVar(&idStr, "id", "", "Policy ID")
	if err := cmd.MarkFlagRequired("env"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(err)
	}
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
	if err := cmd.MarkFlagRequired("env"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("id"); err != nil {
		panic(err)
	}
	return cmd
}
