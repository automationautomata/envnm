package main

import (
	"context"
	pb "envmn/pkg/api/proto"

	"github.com/spf13/cobra"
)

func varsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vars",
		Short: "Environment variables",
	}
	{
		cmd.AddCommand(newUpdateVarsCmd())
		cmd.AddCommand(newRemoveVarCmd())
	}
	return cmd
}

func newUpdateVarsCmd() *cobra.Command {
	var envName, file string
	var vars map[string]string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set variables (flags or file)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if file != "" {
				loaded, err := loadEnvFile(file)
				if err != nil {
					return err
				}
				for k, v := range loaded {
					vars[k] = v
				}
			}

			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.UpdateEnvironmentVariablesRequest{
				EnvironmentName: envName,
				Variables:       vars,
			}
			_, err = client.Management.UpdateEnvironmentVariables(context.Background(), req)
			return err
		},
	}

	vars = make(map[string]string)

	cmd.Flags().StringVar(&envName, "env", "", "Environment name")
	cmd.Flags().StringToStringVar(&vars, "var", map[string]string{}, "Variables KEY=VALUE")
	cmd.Flags().StringVar(&file, "file", "", ".env file path")
	if err := cmd.MarkFlagRequired("env"); err != nil {
		panic(err)
	}
	return cmd
}

func newRemoveVarCmd() *cobra.Command {
	var envName, key string

	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove variable",
		RunE: func(cmd *cobra.Command, args []string) error {
			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.RemoveVariableFromEnvironmentRequest{
				EnvironmentName: envName,
				VariableKey:     key,
			}
			_, err = client.Management.RemoveVariableFromEnvironment(context.Background(), req)
			return err
		},
	}

	cmd.Flags().StringVar(&envName, "env", "", "Environment name")
	cmd.Flags().StringVar(&key, "key", "", "Variable key")
	if err := cmd.MarkFlagRequired("env"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("key"); err != nil {
		panic(err)
	}
	return cmd
}
