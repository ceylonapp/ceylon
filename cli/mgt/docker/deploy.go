package docker

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

func (dp *DeployManager) Deploy(forceCreate bool) error {
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

	redisServerName := fmt.Sprintf("%s_redis", imageName)

	if forceCreate {
		err = rmContainer(dp.Context, dp.Client, redisServerName)
		if err != nil {
			log.Println(err)
		}
	}
	isRedisServiceExits, _ := isContainerExist(dp.Context, dp.Client, redisServerName)
	if !isRedisServiceExits {
		runRedisServer(dp.Context, dp.Client, redisServerName)
	}

	if forceCreate {
		for agentName, _ := range deployConfig.Stack.Agents {

			containerName := fmt.Sprintf("%s_agent", agentName)
			println(agentName, containerName)

			err = rmContainer(dp.Context, dp.Client, containerName)
			if err != nil {
				log.Println(err)
			}
		}

	}

	netWorkName := fmt.Sprintf("%s_network", imageName)
	if forceCreate {
		err = rmNetwork(dp.Context, dp.Client, netWorkName)
		if err != nil {
			log.Println(err.Error())
		}
	}

	isNetworkExists, _ := isNetworkExist(dp.Context, dp.Client, netWorkName)
	if !isNetworkExists {
		_, err = createNetwork(dp.Context, dp.Client, netWorkName)
		if err != nil {
			log.Fatal(err)
			return err
		}
	}

	if !isNetworkExists || !isRedisServiceExits {
		err = attachToNetwork(dp.Context, dp.Client, netWorkName, redisServerName, []string{})
	}

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
		containerTempName := fmt.Sprintf("%s_agent_temp", agentName)
		println(agentName, containerName)

		isContainerExists, _ := isContainerExist(dp.Context, dp.Client, containerTempName)
		if isContainerExists {
			err = rmContainer(dp.Context, dp.Client, containerTempName)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		id, err := runContainer(dp.Context, dp.Client, imageName, containerTempName, inputEnv, agent.Expose)
		if err != nil {
			log.Fatal(err)
			return err
		}
		log.Println("Create container id ", id)

		err = attachToNetwork(dp.Context, dp.Client, netWorkName, containerTempName, []string{
			fmt.Sprintf("%s:redis_host", redisServerName),
		})
		if err != nil {
			log.Fatal(err)
			return err
		}

		isContainerExists, _ = isContainerExist(dp.Context, dp.Client, containerName)
		if isContainerExists {
			err = rmContainer(dp.Context, dp.Client, containerName)
			if err != nil {
				log.Println(err)
				return err
			}
		}

		err = updateContainerName(dp.Context, dp.Client, containerTempName, containerName)

		err = startContainer(dp.Context, dp.Client, containerName)
		if err != nil {
			log.Fatal(err)
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

	imageName := deployConfig.Name
	netWorkName := fmt.Sprintf("%s_network", imageName)

	redisServerName := fmt.Sprintf("%s_redis", imageName)

	err = rmContainer(dp.Context, dp.Client, redisServerName)
	if err != nil {
		log.Println(err)
	}

	for agentName, _ := range deployConfig.Stack.Agents {
		containerName := fmt.Sprintf("%s_agent", agentName)
		err = rmContainer(dp.Context, dp.Client, containerName)
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Println("Container removed ", containerName)
		}
		err = deAttachToNetwork(dp.Context, dp.Client, containerName, netWorkName)
		if err != nil {
			log.Println(err.Error())
		} else {
			log.Println("Container removed ", containerName)
		}
	}

	println(netWorkName)
	err = rmNetwork(dp.Context, dp.Client, netWorkName)
	if err != nil {
		log.Println(err.Error())
	}

	err = rmImage(dp.Context, dp.Client, imageName, isPrune)
	if err != nil {
		log.Println(err.Error())
	}

	return nil
}
