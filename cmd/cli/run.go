package main

import (
	"context"
	pb "envmn/pkg/api/proto"
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
)

func runCmd() *cobra.Command {
	var envName string
	cmd := &cobra.Command{
		Use:        "run",
		Short:      "Run app with specified environment (bypassing access policies)",
		ArgAliases: []string{"app args"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 || envName == "" {
				return nil
			}

			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			req := &pb.GetEnvironmentRequest{
				Name: envName,
			}
			resp, err := client.Management.GetEnvironment(context.Background(), req)
			if err != nil {
				return fmt.Errorf("cannot get environment: %s", err)
			}
			env := resp.Environment

			runApp := exec.Command("", args...)
			for k, v := range env.Variables {
				envVar := fmt.Sprintf("%s=%s", k, v)
				runApp.Env = append(runApp.Env, envVar)
			}

			if err = runApp.Run(); err != nil {
				return err
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&envName, "env", "", "environment")
	markFlagsRequired(cmd, "env")
	return cmd
}
