package mgt

import (
	"ceylon/cli/config"
	"context"
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"os"
)

type DeployManager struct {
	Config  config.DeployConfig
	Context context.Context
}

func (dp *DeployManager) Deploy() error {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	configPath := fmt.Sprintf("%s/%s", path, "ceylon.yaml")

	deployConfig, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}

	packageFileDir := "./"
	imageName := deployConfig.Name

	client, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// Client, imagename and Dockerfile location
	tags := []string{imageName}
	dockerFile := "mgt/images/core/Dockerfile"
	configFiles := []string{"mgt/images/core/requirements.txt"}
	fileDirs := []string{"mgt/bases/core"}
	err = buildImage(dp.Context, client, tags, dockerFile, packageFileDir, configFiles, fileDirs, deployConfig.Stack.Ports)
	if err != nil {
		log.Fatal(err)
		return err
	}

	log.Println(deployConfig.Stack)
	for agentName, agent := range deployConfig.Stack.Agents {
		//log.Println(agent)
		inputEnv := []string{
			fmt.Sprintf("CEYLON_SOURCE=%s", agent.Source),
			fmt.Sprintf("CEYLON_AGENT=%s", agent.Name),
			"REDIS_HOST=192.168.8.100",
			"REDIS_PORT=6379",
			"REDIS_DB=0",
		}

		if agent.Path != "" {
			inputEnv = append(inputEnv, fmt.Sprintf("CEYLON_PATH=%s", agent.Path))
		}
		if agent.Expose != "" {
			inputEnv = append(inputEnv, fmt.Sprintf("CEYLON_EXPOSE=%s", agent.Expose))
		}
		if agent.Type != "" {
			inputEnv = append(inputEnv, fmt.Sprintf("CEYLON_TYPE=%s", agent.Type))
		}
		if agent.Version != "" {
			inputEnv = append(inputEnv, fmt.Sprintf("CEYLON_VERSION=%s", agent.Version))
		}

		containerName := fmt.Sprintf("%s_agent", agentName)
		println(agentName, containerName)

		err = rmContainer(dp.Context, client, containerName)
		if err != nil {
			log.Println(err)
			return err
		}
		id, err := runContainer(dp.Context, client, imageName, containerName, inputEnv, agent.Expose)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Println("Create container id ", id)
	}
	return nil
}
