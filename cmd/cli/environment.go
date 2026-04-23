package main

import (
	"context"

	pb "envmn/pkg/api/proto"

	"github.com/spf13/cobra"
)

func envCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "env",
		Short: "Environment management",
	}
	{
		cmd.AddCommand(newCreateEnvCmd())
		cmd.AddCommand(newListEnvCmd())
		cmd.AddCommand(newDeleteEnvCmd())
		cmd.AddCommand(newUpdateEnvCmd())
	}
	return cmd
}

func newCreateEnvCmd() *cobra.Command {
	var name, desc, file string
	var vars map[string]string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create new environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			if file != "" {
				if loaded, err := loadEnvFile(file); err == nil {
					for k, v := range loaded {
						vars[k] = v
					}
				}
			}

			req := &pb.CreateEnvironmentRequest{
				Name:        name,
				Description: optionalString(desc),
				Variables:   vars,
			}
			resp, err := client.Management.CreateEnvironment(context.Background(), req)
			if err != nil {
				return err
			}
			cmd.Println("Created environment ID:", resp.Id)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Environment name")
	cmd.Flags().StringVar(&desc, "desc", "", "Description")
	cmd.Flags().StringToStringVar(&vars, "var", nil, "Variables (KEY=VALUE)")
	cmd.Flags().StringVar(&file, "file", "", ".env file")
	markFlagsRequired(cmd, "name")
	return cmd
}

func newListEnvCmd() *cobra.Command {
	var reserved bool
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all environments",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.GetAllEnvironmentsRequest{
				Reserved: reserved,
			}
			resp, err := client.Management.GetAllEnvironments(context.Background(), req)
			if err != nil {
				return err
			}

			for _, e := range resp.Environments {
				cmd.Printf("%-20s reserved=%v  %s\n", e.Name, e.Reserved, e.Description)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&reserved, "reserved", false, "Show only reserved environments")
	return cmd
}

func newDeleteEnvCmd() *cobra.Command {
	var name string
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete environment",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.DeleteEnvironmentRequest{Name: name}
			_, err = client.Management.DeleteEnvironment(context.Background(), req)
			return err
		},
	}
	cmd.Flags().StringVar(&name, "name", "", "Environment name")
	markFlagsRequired(cmd, "name")
	return cmd
}

func newUpdateEnvCmd() *cobra.Command {
	var oldName, newName, desc string
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update environment info",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.UpdateEnvironmentInfoRequest{
				OldName:     oldName,
				NewName:     optionalString(newName),
				Description: optionalString(desc),
			}
			_, err = client.Management.UpdateEnvironmentInfo(context.Background(), req)
			return err
		},
	}

	cmd.Flags().StringVar(&oldName, "old", "", "Old name")
	cmd.Flags().StringVar(&newName, "new", "", "New name")
	cmd.Flags().StringVar(&desc, "desc", "", "New description")
	markFlagsRequired(cmd, "old")
	return cmd
}
