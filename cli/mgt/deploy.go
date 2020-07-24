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
	Client  *client.Client
}

func (dp *DeployManager) Init() error {
	cl, err := client.NewEnvClient()
	dp.Client = cl
	if err != nil {
		return err
	}
	return nil
}

func (dp *DeployManager) ReadConfig() (*config.DeployConfig, error) {
	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}

	configPath := fmt.Sprintf("%s/%s", path, "ceylon.yaml")

	deployConfig, err := config.NewConfig(configPath)
	if err != nil {
		panic(err)
	}
	return deployConfig, nil
}

func (dp *DeployManager) Deploy() error {
	err := dp.Init()
	if err != nil {
		panic(err)
	}

	deployConfig, err := dp.ReadConfig()
	if err != nil {
		panic(err)
	}

	packageFileDir := "./"
	imageName := deployConfig.Name

	// Client, imagename and Dockerfile location
	tags := []string{imageName}
	dockerFile := "mgt/images/core/Dockerfile"
	configFiles := []string{"mgt/images/core/requirements.txt"}
	fileDirs := []string{"mgt/bases/core"}

	dockerFile, err = buildDockerImage(dockerFile, deployConfig.Stack.Ports)
	if err != nil {
		log.Fatal(err)
		return err
	}
	err = buildImage(dp.Context, dp.Client, tags, dockerFile, packageFileDir, configFiles, fileDirs)
	if err != nil {
		log.Fatal(err)
		return err
	}
	os.Remove(dockerFile)

	netWorkName := fmt.Sprintf("%s_network", imageName)

	//err = rmNetwork(dp.Context, dp.Client, netWorkName)
	//if err != nil {
	//	log.Println(err.Error())
	//	return err
	//}

	_, err = createNetwork(dp.Context, dp.Client, netWorkName)
	if err != nil {
		log.Fatal(err)
		return err
	}

	redisServerName := fmt.Sprintf("%s_redis", imageName)

	err = rmContainer(dp.Context, dp.Client, redisServerName)
	if err != nil {
		log.Println(err)
	}
	runRedisServer(dp.Context, dp.Client, redisServerName)

	err = attachToNetwork(dp.Context, dp.Client, netWorkName, redisServerName, []string{})
	if err != nil {
		log.Fatal(err)
		return err
	}
	//log.Println("Redis server", id)

	log.Println(deployConfig.Stack)
	for agentName, agent := range deployConfig.Stack.Agents {
		//log.Println(agent)
		inputEnv := []string{
			fmt.Sprintf("CEYLON_SOURCE=%s", agent.Source),
			fmt.Sprintf("CEYLON_AGENT=%s", agent.Name),
			"REDIS_HOST=redis_host",
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

		err = rmContainer(dp.Context, dp.Client, containerName)
		if err != nil {
			log.Println(err)
		}
		id, err := runContainer(dp.Context, dp.Client, imageName, containerName, inputEnv, agent.Expose)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Println("Create container id ", id)

		err = attachToNetwork(dp.Context, dp.Client, netWorkName, containerName, []string{
			fmt.Sprintf("%s:redis_host", redisServerName),
		})
		if err != nil {
			log.Println("Connect to network ", containerName, netWorkName)
			return err
		}
	}
	return nil
}

func (dp *DeployManager) Destroy(isPrune bool) error {
	err := dp.Init()
	if err != nil {
		panic(err)
	}
	deployConfig, err := dp.ReadConfig()
	if err != nil {
		panic(err)
	}
	for agentName, _ := range deployConfig.Stack.Agents {
		containerName := fmt.Sprintf("%s_agent", agentName)
		err = rmContainer(dp.Context, dp.Client, containerName)
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Println("Container removed ", containerName)
		}
	}

	imageName := deployConfig.Name
	netWorkName := fmt.Sprintf("%s_network", imageName)
	println(netWorkName)
	err = rmNetwork(dp.Context, dp.Client, netWorkName)
	if err != nil {
		log.Println(err.Error())
		return err
	}

	err = rmImage(dp.Context, dp.Client, imageName, isPrune)
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}
