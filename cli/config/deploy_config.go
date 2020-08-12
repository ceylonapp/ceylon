package config

import (
	"bytes"
	"gopkg.in/yaml.v2"
	"io"
	"log"
	"os"
	"sort"
)

type Agent struct {
	Source     string            `yaml:"source"`
	Name       string            `yaml:"name"`
	Expose     string            `yaml:"expose"`
	Path       string            `yaml:"path"`
	Type       string            `yaml:"type"`
	Version    string            `yaml:"version"`
	Order      int               `yaml:"order"`
	InitParams map[string]string `yaml:"init_params,flow"`
}
type AgentMap map[string]Agent
type Kv struct {
	Key   string
	Value Agent
}
type Stack struct {
	Ports  []string `yaml:"ports,flow"`
	Agents AgentMap `yaml:"agents,flow"`
}
type DeployConfig struct {
	Name   string   `yaml:"name"`
	Stack  Stack    `yaml:"stack"`
	Envars []string `yaml:"envars"`
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
			Ports: []string{"4554", "4545"},
			Agents: map[string]Agent{
				"python_app": {
					Source: "sample",
					Name:   "sample_name",
					Expose: "4545",
				},
				"python_app2": {
					Source: "sample1",
					Name:   "sample1_name",
				},
			},
		},
		Envars: []string{
			"envars1=testval1",
			"envars2=testval2",
		},
	})

	log.Println("Example file")
	log.Println(string(b))

	log.Println("Read config")
	log.Println(config)

	return config, nil
}

func (amap AgentMap) Order() (agentList []Kv) {
	for k, v := range amap {
		agentList = append(agentList, Kv{
			Key:   k,
			Value: v,
		})
	}
	sort.Slice(agentList, func(i, j int) bool {
		return agentList[i].Value.Order > agentList[j].Value.Order
	})

	return

}
