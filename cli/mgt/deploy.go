package mgt

import (
	"fmt"
	"github.com/docker/docker/client"
	"log"
	"os"
)

type DeployConfig struct {
}

type DeployManager struct {
	Config DeployConfig
}

func (dp *DeployManager) Deploy() error {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	packageFileDir := "./"
	imageName := "hello_world_agents"

	client, err := client.NewEnvClient()
	if err != nil {
		return err
	}

	// Client, imagename and Dockerfile location
	tags := []string{imageName}
	dockerFile := "mgt/images/core/Dockerfile"
	configFiles := []string{"mgt/images/core/requirements.txt"}
	fileDirs := []string{"mgt/bases/core"}
	inputEnv := []string{
		fmt.Sprintf("LISTENINGPORT=%s", "8080"),
		"REDIS_HOST=192.168.8.100",
		"REDIS_PORT=6379",
		"REDIS_DB=0",
	}
	err = buildImage(client, tags, dockerFile, packageFileDir, configFiles, fileDirs)
	err = runContainer(client, imageName, "test_1", "8080", inputEnv)
	if err != nil {
		return err
	}
	return nil
}