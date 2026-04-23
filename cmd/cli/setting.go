package main

import (
	"context"
	"envmn/config"
	pb "envmn/pkg/api/proto"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newSettingsCmd() *cobra.Command {
	root := &cobra.Command{
		Use:        "setting",
		Short:      "setting server from file",
		Args:       cobra.ExactArgs(1),
		ArgAliases: []string{"path to settings file"},
		RunE: func(cmd *cobra.Command, args []string) error {
			settings, err := config.LoadSettings(args[0])
			if err != nil {
				return err
			}

			client, err := getGRPCClient()
			if err != nil {
				return err
			}
			defer client.Close()

			if err := validateSettings(context.Background(), settings, client.Management); err != nil {
				return err
			}

			cleanup, err := settingServer(context.Background(), settings, client.Management)
			if err != nil {
				cleanup()
				return err
			}
			return nil
		},
	}
	{
		root.AddCommand(envCmd())
		root.AddCommand(policyCmd())
		root.AddCommand(varsCmd())
	}
	return root
}

func settingServer(ctx context.Context, settings config.Settings, client pb.ManagementServiceClient) (cleanupFn func(), err error) {
	envs := make([]string, 0, len(settings.Environments))
	policies := make([]string, 0, len(settings.Policies))

	envPrefix := ""
	if settings.EnvironmentsPrefix != nil {
		envPrefix = *settings.EnvironmentsPrefix
	}
	for name, env := range settings.Environments {
		vars := make(map[string]string)
		if env.File != nil {
			if loaded, err := loadEnvFile(*env.File); err == nil {
				for k, v := range loaded {
					vars[k] = v
				}
			}
		}
		for k, v := range env.Variables {
			vars[k] = v
		}

		name = fmt.Sprint(envPrefix, name)
		req := &pb.CreateEnvironmentRequest{
			Name:        name,
			Description: optionalString(""),
			Variables:   vars,
		}
		_, err := client.CreateEnvironment(ctx, req)
		if err != nil {
			return cleanup(policies, envs, client), err
		}

		envs = append(envs, name)
	}

	policyPrefix := ""
	if settings.PoliciesPrefix != nil {
		policyPrefix = *settings.PoliciesPrefix
	}
	for policyName, policyEnvs := range settings.Policies {
		name := fmt.Sprint(policyPrefix, policyName)
		req := &pb.CreateAccessPolicyRequest{Name: name}
		_, err := client.CreateAccessPolicy(ctx, req)

		if err != nil {
			return cleanup(policies, envs, client), err
		}

		policies = append(policies, name)

		for envName := range policyEnvs {
			req := &pb.AddPolicyToEnvironmentRequest{
				PolicyId:        fmt.Sprint(policyPrefix, policyName),
				EnvironmentName: fmt.Sprint(envPrefix, envName),
			}
			_, err := client.AddPolicyToEnvironment(ctx, req)

			if err != nil {
				return cleanup(policies, envs, client), err
			}
		}
	}
	return nil, nil
}

func validateSettings(ctx context.Context, settings config.Settings, client pb.ManagementServiceClient) error {
	prefix := ""
	if settings.EnvironmentsPrefix != nil {
		prefix = *settings.EnvironmentsPrefix
	}

	envs := make(map[string]struct{})
	for envName, env := range settings.Environments {
		name := fmt.Sprint(prefix, envName)

		if env.File != nil && !fileExists(*env.File) {
			return fmt.Errorf("invalid env file for environment %s (%s): file not exists", envName, name)
		}

		_, err := client.GetEnvironment(ctx, &pb.GetEnvironmentRequest{Name: name})
		if err == nil {
			return fmt.Errorf("environment %s (%s) already exists", envName, name)
		}
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("environment error %s (%s): %w", envName, name, err)
		}

		envs[envName] = struct{}{}
	}

	prefix = ""
	if settings.PoliciesPrefix != nil {
		prefix = *settings.PoliciesPrefix
	}
	for policyName, envs := range settings.Policies {
		for envName := range envs {
			if _, ok := envs[envName]; !ok {
				return fmt.Errorf("invalid env file for policy %s: environment %s is not defined", policyName, envName)
			}
		}

		name := fmt.Sprint(prefix, policyName)
		_, err := client.GetPolicyByName(ctx, &pb.GetPolicyByNameRequest{Name: name})
		if err == nil {
			return fmt.Errorf("policy %s (%s) already exists", policyName, name)
		}
		if status.Code(err) != codes.NotFound {
			return fmt.Errorf("policy error %s (%s): %w", policyName, name, err)
		}
	}
	return nil
}

func cleanup(policies, envs []string, client pb.ManagementServiceClient) func() {
	return func() {
		for _, policy := range policies {
			_, _ = client.RemovePolicy(context.Background(), &pb.RemovePolicyRequest{Id: policy})
		}
		for _, env := range envs {
			_, _ = client.DeleteEnvironment(context.Background(), &pb.DeleteEnvironmentRequest{Name: env})
		}
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
