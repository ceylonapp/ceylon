package config

type DeployConfig struct {
	Name string `yaml:"name"`

	Stack struct {
		Agents []struct {
			Source string `yaml:"name"`
			Name   string `yaml:"name"`
		}
	}
}
