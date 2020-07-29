package virtualenv

import (
	"ceylon/cli/config"
	"context"
	"fmt"
	"github.com/go-cmd/cmd"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
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
	Context     context.Context
	env         *VirtualEnvService
	Config      *config.DeployConfig
	ProjectPath string
}

func (dp *VEnvDeployManager) readConfig() (*config.DeployConfig, error) {
	path, err := os.Getwd()
	if err != nil {
		panic(err)
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
	projectLocation, err := dp.env.initiateLocation()
	if err != nil {
		panic(err)
		return err
	}
	dp.ProjectPath = projectLocation

	return nil
}

func (dp *VEnvDeployManager) Prepare() error {
	runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "create.bat"), dp.ProjectPath)
	runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "venv/Scripts/python.exe"), "--version")
	runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "venv/Scripts/pip.exe"), "install", "-r", filepath.Join(dp.ProjectPath, "mgt\\images\\core\\requirements.txt"))
	runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "venv/Scripts/pip.exe"), "install", "-r", filepath.Join(dp.ProjectPath, "requirements.txt"))

	return nil
}

func runCommand(out io.Writer, command string, args ...string) {

	exeCmd := cmd.NewCmd(command, args...)
	statusChan := exeCmd.Start()
	go func() {
		status := exeCmd.Status()
		for _, logVal := range status.Stdout {
			log.Println("LOG :: ", logVal)
		}
		//io.Copy(out, status.Stdout)
	}()
	// Check if command is done
	select {
	case finalStatus := <-statusChan:
		log.Println(finalStatus)
	default:
		log.Println(statusChan)
	}

	// Block waiting for command to exit, be stopped, or be killed
	finalStatus := <-statusChan
	log.Println(finalStatus)
	//for {
	//	statusS := <-status
	//	log.Println(statusS.Stdout)
	//}
	//// Print each line of STDOUT from Cmd
	//for _, line := range status.Stdout {
	//	fmt.Println(line)
	//}
	////log.Println("Executing...", command, args)
	////cmd := exec.Command(command, args...)
	////stdout, _ := cmd.StdoutPipe()
	////stderr, _ := cmd.StderrPipe()
	////cmd.Start()
	////io.Copy(out, stdout)
	////io.Copy(out, stderr)
	////cmd.Wait()
}

func (dp *VEnvDeployManager) agentWorker(wg *sync.WaitGroup, agent config.Agent, out io.Writer) {
	defer wg.Done()

	agentArgs := []string{filepath.Join(dp.ProjectPath, "ceylon/source/run.py")}
	agentArgs = append(agentArgs, "--source", agent.Source)
	agentArgs = append(agentArgs, "--agent", agent.Name)

	//--path=/hello --expose=8080 --type=ws
	if agent.Expose != "" {
		agentArgs = append(agentArgs, "--expose", agent.Expose)
	}
	if agent.Type != "" {
		agentArgs = append(agentArgs, "--type", agent.Type)
	}
	if agent.Path != "" {
		agentArgs = append(agentArgs, "--path", agent.Path)
	}

	prepareFilePath := filepath.Join(dp.ProjectPath, "venv/Scripts/python.exe")
	runCommand(out, prepareFilePath, agentArgs...)
}

func (dp *VEnvDeployManager) Run() error {
	var wg sync.WaitGroup
	//var out io.Reader
	for _, agent := range dp.Config.Stack.Agents {
		wg.Add(1)
		go dp.agentWorker(&wg, agent, os.Stdout)
	}
	fmt.Println("Main: Waiting for workers to finish")
	wg.Wait()
	fmt.Println("Main: Completed")

	return nil
}
