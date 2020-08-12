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
	"runtime"
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
	projectLocation, err := dp.env.initiateLocation(config.ForceCreate)
	if err != nil {
		panic(err)
		return err
	}
	dp.ProjectPath = projectLocation

	return nil
}

func (dp *VEnvDeployManager) Prepare() error {
	if runtime.GOOS == "windows" {
		runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "create.bat"), dp.ProjectPath)
		runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "venv/Scripts/python.exe"), "--version")
		runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "venv/Scripts/pip.exe"), "install", "-r", filepath.Join(dp.ProjectPath, "mgt\\images\\core\\requirements.txt"))
		runCommand(os.Stdout, filepath.Join(dp.ProjectPath, "venv/Scripts/pip.exe"), "install", "-r", filepath.Join(dp.ProjectPath, "requirements.txt"))
	} else {
		log.Fatal(fmt.Sprintf("Not yet support for %s", runtime.GOOS))
	}

	return nil
}

func runCommand(out io.Writer, command string, args ...string) {
	commandStr := ""
	for _, arg := range args {
		commandStr = commandStr + " " + arg
	}

	log.Println(fmt.Sprintf("%s %s\n", command, commandStr))
	cmdOptions := cmd.Options{
		Buffered:  false,
		Streaming: true,
	}

	// Create Cmd with options
	envCmd := cmd.NewCmdOptions(cmdOptions, command, args...)

	// Print STDOUT and STDERR lines streaming from Cmd
	doneChan := make(chan struct{})
	go func() {
		defer close(doneChan)
		// Done when both channels have been closed
		// https://dave.cheney.net/2013/04/30/curious-channels
		for envCmd.Stdout != nil || envCmd.Stderr != nil {
			select {
			case line, open := <-envCmd.Stdout:
				if !open {
					envCmd.Stdout = nil
					continue
				}
				fmt.Println(line)
			case line, open := <-envCmd.Stderr:
				if !open {
					envCmd.Stderr = nil
					continue
				}
				fmt.Fprintln(os.Stderr, line)
			}
		}
	}()
	// Run and wait for Cmd to return, discard Status
	<-envCmd.Start()

	// Wait for goroutine to print everything
	<-doneChan

}

func (dp *VEnvDeployManager) agentWorker(wg *sync.WaitGroup, agent config.Agent, out io.Writer) {
	//defer

	agentArgs := []string{filepath.Join(dp.ProjectPath, "ceylon/source/run.py")}
	agentArgs = append(agentArgs, "--stack", dp.Config.Name)
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

	if agent.InitParams != nil {
		for k, v := range agent.InitParams {
			agentArgs = append(agentArgs, fmt.Sprintf("--init-params %s %s", k, v))
		}
	}

	if runtime.GOOS == "windows" {
		prepareFilePath := filepath.Join(dp.ProjectPath, "venv/Scripts/python.exe")
		runCommand(out, prepareFilePath, agentArgs...)
	} else {
		log.Fatal(fmt.Sprintf("Not yet support for %s", runtime.GOOS))
	}
	fmt.Println("Agent process done", agent.Name)
	wg.Done()
}

func (dp *VEnvDeployManager) Run() error {
	var wg sync.WaitGroup

	wg.Add(len(dp.Config.Stack.Agents))
	//var out io.Reader
	for _, kv := range dp.Config.Stack.Agents.Order() {
		agent := kv.Value
		agentName := kv.Key
		log.Print("Agent ", agentName, "Starting...")
		go dp.agentWorker(&wg, agent, os.Stdout)
	}
	fmt.Println("Start agents")
	wg.Wait()
	return nil
}
