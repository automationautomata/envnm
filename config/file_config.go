package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Premissions map[string]bool

type Envinronment struct {
	File      *string           `yaml:"file"`
	Variables map[string]string `yaml:"variables"`
}

type Settings struct {
	EnvironmentsPrefix *string                 `yaml:"environments-prefix"`
	PoliciesPrefix     *string                 `yaml:"policies-prefix"`
	Environments       map[string]Envinronment `yaml:"environments"`
	Policies           map[string]Premissions  `yaml:"policies"`
}

func LoadSettings(path string) (Settings, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Settings{}, err
	}

	var settings Settings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return Settings{}, err
	}
	return settings, nil
}
