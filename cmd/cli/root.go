package main

import (
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

func newRootCmd() *cobra.Command {
	var envFilePath string
	root := &cobra.Command{
		Use:   "envmn",
		Short: "Environment Manager CLI",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := godotenv.Load(envFilePath); err != nil {
				return err
			}
			return nil
		},
	}
	root.Flags().StringVar(&envFilePath, "env-file", "", "path to .env")
	{
		root.AddCommand(envCmd())
		root.AddCommand(policyCmd())
		root.AddCommand(varsCmd())
	}
	return root
}
