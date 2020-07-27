package virtualenv

import (
	"ceylon/cli/config"
	"context"
	"fmt"
	"log"
	"os"
)

type CreateSettings struct {
	ForceCreate bool
}

type Deploy interface {
	Init() error
	Create() error
	readConfig() error
}
type VEnvDeployManager struct {
	Context context.Context
	env     *VirtualEnvService
	Config  *config.DeployConfig
}

func (dp *VEnvDeployManager) readConfig() (*config.DeployConfig, error) {
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

func CreateInstance(ctx context.Context) *VEnvDeployManager {
	vent := &VEnvDeployManager{Context: ctx}
	vent.Init()
	return vent
}

func (dp *VEnvDeployManager) Init() {
	baseLocation, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	deployConfig, err := dp.readConfig()
	if err != nil {
		panic(err)
	}
	dp.Config = deployConfig
	dp.env = &VirtualEnvService{
		Context:      dp.Context,
		BaseLocation: &baseLocation,
		DeployConfig: deployConfig}
}

func (dp *VEnvDeployManager) Create(config *CreateSettings) (err error) {
	err = dp.env.initiateLocation()
	if err != nil {
		log.Fatal(err)
		return err
	}

	return nil
}
