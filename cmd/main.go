package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"ceylon/cli/utils"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"log"
	"os"
)

func buildImage(client *client.Client, tags []string, dockerFile string, configFiles []string, sourceDirs []string) error {

	ctx := context.Background()

	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := utils.FileToTar(dockerFile, tw)
	if err != nil {
		return err
	}

	for _, configFile := range configFiles {
		err = utils.FileToTar(configFile, tw)
		if err != nil {
			return err
		}
	}

	// Add File Dirs
	for _, sourceDir := range sourceDirs {
		err = utils.DirToTar(sourceDir, tw)
		if err != nil {
			return err
		}
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
		Context:    dockerFileTarReader,
		Dockerfile: dockerFile,
		Remove:     true,
		Tags:       tags,
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

func main() {

	packageFileDir := "example/hello_ceylon"

	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Unable to create docker client: %s", err)
	}

	// Client, imagename and Dockerfile location
	tags := []string{"hello_world_agents"}
	dockerFile := "mgt/images/core/Dockerfile"
	configFiles := []string{"mgt/images/core/requirements.txt"}
	fileDirs := []string{"mgt/bases/core", packageFileDir}
	err = buildImage(client, tags, dockerFile, configFiles, fileDirs)
	if err != nil {
		panic(err)
		//log.Fatal(err.Error())
	}
}
