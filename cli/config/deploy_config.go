package config

import (
	"bytes"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
)

type Agent struct {
	Source string `yaml:"source"`
	Name   string `yaml:"name"`
}
type Stack struct {
	Agents map[string]Agent `yaml:"agents,flow"`
}
type DeployConfig struct {
	Name  string `yaml:"name"`
	Stack Stack  `yaml:"stack"`
}

// NewConfig returns a new decoded Config struct
func NewConfig(configPath string) (*DeployConfig, error) {
	// Create config structure
	config := &DeployConfig{}

	// Open config file
	file, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	buf := bytes.NewBuffer(nil)
	io.Copy(buf, file) // Error handling elided for brevity.

	yaml.Unmarshal(buf.Bytes(), &config)

	b, _ := yaml.Marshal(DeployConfig{
		Name: "ABB",
		Stack: Stack{
			Agents: map[string]Agent{
				"python_app": {
					Source: "sample",
					Name:   "sample_name",
				},
				"python_app2": {
					Source: "sample1",
					Name:   "sample1_name",
				},
			},
		},
	})

	log.Println(string(b))

	//// Init new YAML decode
	//d := yaml.NewDecoder(file)
	//
	//// Start YAML decoding from file
	//if err := d.Decode(&config); err != nil {
	//	return nil, err
	//}

	return config, nil
}
func validateConfigPath(path string) error {
	s, err := os.Stat(path)
	if err != nil {
		return err
	}
	if s.IsDir() {
		return fmt.Errorf("'%s' is a directory, not a normal file", path)
	}
	return nil
}
