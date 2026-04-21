package main

import (
	"bufio"
	"envmn/config"
	"fmt"
	"os"
	"strings"
)

func getGRPCClient() (*gRPCClient, error) {
	cfg, err := config.Load[config.CLIConfig]()
	if err != nil {
		return nil, err
	}
	return newGRPCClient(cfg)
}

func loadEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	res := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && !strings.HasPrefix(line, "#") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) < 2 {
				return nil, fmt.Errorf("variable %q must have a value", parts[0])
			}

			key := strings.TrimSpace(parts[0])
			val := strings.TrimSpace(parts[1])

			res[key] = val
		}

	}
	return res, scanner.Err()
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
