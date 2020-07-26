package docker

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
	"runtime"
	"text/template"
	"time"
)

func attachToNetwork(ctx context.Context, client *client.Client, networkName string, containerId string, links []string) error {
	err := client.NetworkConnect(ctx, networkName, containerId, &network.EndpointSettings{
		Links: links,
	})
	return err
}
func deAttachToNetwork(ctx context.Context, client *client.Client, networkName string, containerId string) error {
	err := client.NetworkDisconnect(ctx, networkName, containerId, true)
	return err
}
func stopContainer(ctx context.Context, client *client.Client, containerId string) error {
	timeDuration := time.Second * 10
	err := client.ContainerStop(ctx, containerId, &timeDuration)
	return err
}
func rmNetwork(ctx context.Context, client *client.Client, networkName string) error {
	err := client.NetworkRemove(ctx, networkName)
	if err != nil {
		log.Println(err.Error())
		return err
	}
	return nil
}
func isNetworkExist(ctx context.Context, cl *client.Client, networkName string) (bool, error) {
	_, err := cl.NetworkInspect(ctx, networkName)
	if err != nil {
		return false, err
	}
	return true, nil
}

func createNetwork(ctx context.Context, client *client.Client, networkName string) (string, error) {

	res, err := client.NetworkCreate(ctx, networkName, types.NetworkCreate{
		Internal:   false,
		Attachable: true,
	})
	if err != nil {
		return "", err
	}
	return res.ID, nil
}

func rmImage(ctx context.Context, client *client.Client, imageName string, isPruneChildren bool) error {

	deleted, err := client.ImageRemove(ctx, imageName, types.ImageRemoveOptions{
		Force:         true,
		PruneChildren: isPruneChildren,
	})

	for _, imageDelete := range deleted {
		log.Println("Image deleted", imageDelete.Deleted)
		log.Println("Image untaged", imageDelete.Untagged)
	}

	if err != nil {
		return err
	}

	return nil
}

func buildDockerImage(dockerFile string, expose []string) (string, error) {
	_, baseFilePath, _, _ := runtime.Caller(1)
	dockerFilePath := path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../../%s", dockerFile))
	content, err := ioutil.ReadFile(dockerFilePath)
	if err != nil {
		println(err.Error())
		return "", err
	}

	tmpl, err := template.New("Dockerfile").Parse(string(content))
	if err != nil {
		println(err.Error())
		return "", err
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
		return "", err
	}

	tmpDockerFIle, err := ioutil.TempFile("./", "Dockerfile")
	if err != nil {
		println(err.Error())
		return "", err
	}
	defer os.Remove(tmpDockerFIle.Name())
	err = ioutil.WriteFile(tmpDockerFIle.Name(), tpl.Bytes(), 0644)

	if err != nil {
		println(err.Error())
		return "", err
	}

	tmpDockerFIle.Close()
	if fileExists("Dockerfile") {
		err = os.Remove("Dockerfile")
		if err != nil {
			println(err.Error())
		}
	}

	err = os.Rename(tmpDockerFIle.Name(), "Dockerfile")
	if err != nil {
		println(err.Error())
		return "", err
	}
	return "Dockerfile", nil
}
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
func buildImage(ctx context.Context, client *client.Client, tags []string, dockerFile string, sourceDir string, configFiles []string, configDirs []string) error {
	_, baseFilePath, _, _ := runtime.Caller(1)

	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	//_, fileName := filepath.Split(dockerFilePath)
	//fileName = fmt.Sprintf("%s%s", "config/", fileName)

	//err = utils.FileObjToTar(tw, bytes.NewReader(tpl.Bytes()), fileName)
	//if err != nil {
	//	return err
	//}

	for _, configFile := range configFiles {
		configFile = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../../%s", configFile))
		_, err := utils.FileToTar(configFile, "config/", tw)
		if err != nil {
			return err
		}
	}

	// Add File Dirs
	for _, configDir := range configDirs {
		configDir = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../../%s", configDir))
		err := utils.DirToTar(configDir, tw)
		if err != nil {
			return err
		}
	}

	err := utils.DirToTar(sourceDir, tw)
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
		NoCache:    false,
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

	err = tw.Close()
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

func runRedisServer(ctx context.Context, client *client.Client, containername string) {

	reader, err := client.ImagePull(ctx, "docker.io/library/redis", types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	portBinding := natting.PortMap{}
	cont, err := client.ContainerCreate(
		context.Background(),
		&container.Config{
			Image: "redis",
		},
		&container.HostConfig{
			PortBindings: portBinding,
		}, nil, containername)
	if err != nil {
		panic(err)
	}

	client.ContainerStart(context.Background(), cont.ID, types.ContainerStartOptions{})
	fmt.Printf("Container %s is started", cont.ID)

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
		nil,
		containername,
	)

	if err != nil {
		log.Println(err)
		return "", err
	}

	return cont.ID, nil
}

func startContainer(ctx context.Context, client *client.Client, containername string) error {
	// Run the actual container
	err := client.ContainerStart(ctx, containername, types.ContainerStartOptions{})

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func isContainerExist(ctx context.Context, cli *client.Client, containerName string) (bool, error) {
	_, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return false, err
	}

	return true, nil
}

func updateContainerName(ctx context.Context, client *client.Client, containerName string, containerNewName string) error {
	// Run the actual container
	err := client.ContainerRename(ctx, containerName, containerNewName)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}
