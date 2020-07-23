package mgt

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
	natting "github.com/docker/go-connections/nat"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"text/template"
)

func buildImage(ctx context.Context, client *client.Client, tags []string, dockerFile string, sourceDir string, configFiles []string, configDirs []string, expose []string) error {

	_, baseFilePath, _, _ := runtime.Caller(1)
	dockerFilePath := path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../%s", dockerFile))
	content, err := ioutil.ReadFile(dockerFilePath)
	if err != nil {
		println(err.Error())
	}

	tmpl, err := template.New("Dockerfile").Parse(string(content))
	if err != nil {
		println(err.Error())
	}

	//out := bufio.
	var tpl bytes.Buffer
	err = tmpl.Execute(&tpl, struct {
		Expose []string
	}{
		Expose: expose,
	})
	if err != nil {
		println(err.Error())
	}

	tmpDockerFIle, err := ioutil.TempFile("./", "Dockerfile")
	if err != nil {
		log.Fatal(err)
	}
	defer os.Remove(tmpDockerFIle.Name())
	err = ioutil.WriteFile(tmpDockerFIle.Name(), tpl.Bytes(), 0644)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(tmpDockerFIle.Name())

	_, fileName := filepath.Split(dockerFilePath)
	fileName = fmt.Sprintf("%s%s", "config/", fileName)

	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err = utils.FileObjToTar(tw, bytes.NewReader(tpl.Bytes()), fileName)
	if err != nil {
		return err
	}

	for _, configFile := range configFiles {
		configFile = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../%s", configFile))
		configFile, err = utils.FileToTar(configFile, "config/", tw)
		if err != nil {
			return err
		}
	}

	// Add File Dirs
	for _, configDir := range configDirs {
		configDir = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../%s", configDir))
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
		Dockerfile: tmpDockerFIle.Name(),
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

func rmContainer(ctx context.Context, client *client.Client, containername string) error {
	err := client.ContainerRemove(ctx, containername, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		RemoveLinks:   false,
		Force:         true,
	})
	return err
}

func runContainer(ctx context.Context, client *client.Client, imagename string, containername string, inputEnv []string, expose string) (string, error) {
	// Define a PORT opening
	portBindings := natting.PortMap{}
	exposedPorts := map[natting.Port]struct{}{}

	if expose != "" {
		exposeList := make([]string, 0)
		exposeList = append(exposeList, expose)

		for _, port := range exposeList {
			newport, err := natting.NewPort("tcp", port)
			if err != nil {
				fmt.Println("Unable to create docker port")
				return "", err
			}

			portBindings[newport] = []natting.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: port,
				},
			}

			exposedPorts[newport] = struct{}{}
		}
	}

	// Configured hostConfig:
	// https://godoc.org/github.com/docker/docker/api/types/container#HostConfig
	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
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
		return "", err
	}
	// Run the actual container
	client.ContainerStart(ctx, cont.ID, types.ContainerStartOptions{})
	return cont.ID, nil
}
