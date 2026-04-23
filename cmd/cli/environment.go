package main

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

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

			if name == "" && file == "" {
				return fmt.Errorf("name is required")
			}
			if name == "" {
				abspath, _ := filepath.Abs(file)
				name = fmt.Sprintf("environment:%s", abspath)
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

			breakline := strings.Repeat("-", 25)
			for _, e := range resp.Environments {
				cmd.Printf(
					"name: %s\ndescription: %s\nreserved:%t\nvariables:\n",
					e.Name, e.Description, e.Reserved,
				)
				for k, v := range e.Variables {
					cmd.Printf("\t%s: %s\n", k, v)
				}
				cmd.Printf("\n\n%s\n\n", breakline)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&reserved, "reserved", false, "Show only reserved environments")
	return cmd
}

func newDeleteEnvCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [environment name]",
		Short: "Delete environment",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
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
