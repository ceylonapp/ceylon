package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"ceylon/cli/utils"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"path"
	"runtime"

	natting "github.com/docker/go-connections/nat"
	"io"
	"log"
	"os"
)

func buildImage(client *client.Client, tags []string, dockerFile string, sourceDir string, configFiles []string, configDirs []string) error {

	_, baseFilePath, _, _ := runtime.Caller(1)
	dockerFile = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../%s", dockerFile))

	ctx := context.Background()

	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	dockerFile, err := utils.FileToTar(dockerFile, "config/", tw)
	if err != nil {
		return err
	}

	for _, configFile := range configFiles {
		configFile = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../%s", configFile))
		configFile, err = utils.FileToTar(configFile, "config/", tw)
		if err != nil {
			return err
		}
	}

	// Add File Dirs
	for _, configDir := range configDirs {
		configDir = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../%s", configDir))
		err = utils.DirToTar(configDir, tw)
		if err != nil {
			return err
		}
	}

	err = utils.DirToTar(sourceDir, tw)
	if err != nil {
		return err
	}

	dockerFileTarReader := bytes.NewReader(buf.Bytes())

	tr := tar.NewReader(bufio.NewReader(buf))
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Contents of %s:\n", hdr.Name)

		fmt.Println()
	}

	// Define the build options to use for the file
	// https://godoc.org/github.com/docker/docker/api/types#ImageBuildOptions
	buildOptions := types.ImageBuildOptions{
		Tags:       tags,
		NoCache:    true,
		Remove:     true,
		Dockerfile: dockerFile,
		Context:    dockerFileTarReader,
	}

	// Build the actual image
	imageBuildResponse, err := client.ImageBuild(
		ctx,
		dockerFileTarReader,
		buildOptions,
	)

	if err != nil {
		return err
	}

	// Read the STDOUT from the build process
	defer imageBuildResponse.Body.Close()
	_, err = io.Copy(os.Stdout, imageBuildResponse.Body)
	if err != nil {
		return err
	}

	return nil
}

func runContainer(client *client.Client, imagename string, containername string, port string, inputEnv []string) error {
	// Define a PORT opening
	newport, err := natting.NewPort("tcp", port)
	if err != nil {
		fmt.Println("Unable to create docker port")
		return err
	}

	// Configured hostConfig:
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		PortBindings: natting.PortMap{
			newport: []natting.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: "always",
		},
		LogConfig: container.LogConfig{
			Type:   "json-file",
			Config: map[string]string{},
		},
	}

	// Define Network config (why isn't PORT in here...?:
	// https://godoc.org/github.com/docker/docker/api/types/network#NetworkingConfig
	networkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{},
	}
	gatewayConfig := &network.EndpointSettings{
		Gateway: "gatewayname",
	}
	networkConfig.EndpointsConfig["bridge"] = gatewayConfig

	// Define ports to be exposed (has to be same as hostconfig.portbindings.newport)
	exposedPorts := map[natting.Port]struct{}{
		newport: struct{}{},
	}

	// Configuration
	// https://godoc.org/github.com/docker/docker/api/types/container#Config
	config := &container.Config{
		Image:        imagename,
		Env:          inputEnv,
		ExposedPorts: exposedPorts,
		Hostname:     fmt.Sprintf("%s-hostnameexample", imagename),
	}

	// Creating the actual container. This is "nil,nil,nil" in every example.
	cont, err := client.ContainerCreate(
		context.Background(),
		config,
		hostConfig,
		networkConfig,
		containername,
	)

	if err != nil {
		log.Println(err)
		return err
	}

	// Run the actual container
	client.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	log.Printf("Container %s is created", cont.ID)

	return nil
}

func main() {

	path, err := os.Getwd()
	if err != nil {
		log.Println(err)
	}
	fmt.Println(path)

	packageFileDir := "./"
	imageName := "hello_world_agents"

	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Unable to create docker client: %s", err)
	}

	// Client, imagename and Dockerfile location
	tags := []string{imageName}
	dockerFile := "mgt/images/core/Dockerfile"
	configFiles := []string{"mgt/images/core/requirements.txt"}
	fileDirs := []string{"mgt/bases/core"}
	err = buildImage(client, tags, dockerFile, packageFileDir, configFiles, fileDirs)
	if err != nil {
		panic(err)
		//log.Fatal(err.Error())
	}
	inputEnv := []string{fmt.Sprintf("LISTENINGPORT=%s", "8080")}
	err = runContainer(client, imageName, "test_1", "8080", inputEnv)
	if err != nil {
		log.Println(err)
	}
}
