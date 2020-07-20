package main

import (
	"archive/tar"
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func checkerror(err error) {
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}
}

func writeFileToTar(sourceFile string, tw *tar.Writer) error {
	// Create a filereader
	sourceFileReader, err := os.Open(sourceFile)
	if err != nil {
		return err
	}

	// Read the actual Dockerfile
	readDockerFile, err := ioutil.ReadAll(sourceFileReader)
	if err != nil {
		return err
	}

	// Make a TAR header for the file
	tarHeader := &tar.Header{
		Name: sourceFile,
		Size: int64(len(readDockerFile)),
	}

	//Writes the header described for the TAR file
	err = tw.WriteHeader(tarHeader)
	if err != nil {
		return err
	}

	// Writes the dockerfile data to the TAR file
	_, err = tw.Write(readDockerFile)
	if err != nil {
		return err
	}

	return err
}

func writeDirToTar(sourceDir string,
	tw *tar.Writer) error {

	dir, err := os.Open(sourceDir)
	checkerror(err)
	defer dir.Close()

	// get list of files
	files, err := dir.Readdir(0)
	checkerror(err)

	defer tw.Close()

	log.Println("Number of files ", len(files))
	// walk path
	return filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {

		// return on any error
		if err != nil {
			return err
		}

		// return on non-regular files (thanks to [kumo](https://medium.com/@komuw/just-like-you-did-fbdd7df829d3) for this suggested update)
		if !fi.Mode().IsRegular() {
			return nil
		}

		// create a new dir/file header
		header, err := tar.FileInfoHeader(fi, fi.Name())
		if err != nil {
			return err
		}

		// update the name to correctly reflect the desired destination when untaring

		header.Name = strings.Replace(strings.Replace(file, sourceDir, "", -1), string("\\"), "/", -1)
		log.Println(header.Name)
		// write the header
		if err := tw.WriteHeader(header); err != nil {
			return err
		}

		// open files for taring
		f, err := os.Open(file)
		if err != nil {
			return err
		}

		// copy file data into tar writer
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}

		// manually close here after each file operation; defering would cause each file close
		// to wait until all operations have completed.
		f.Close()

		return nil
	})

	return err
}

func buildImage(client *client.Client, tags []string, dockerFile string, configFiles []string, sourceDirs []string) error {

	ctx := context.Background()

	// Create a buffer
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	err := writeFileToTar(dockerFile, tw)
	if err != nil {
		return err
	}

	for _, configFile := range configFiles {
		err = writeFileToTar(configFile, tw)
		if err != nil {
			return err
		}
	}

	// Add File Dirs
	for _, sourceDir := range sourceDirs {
		err = writeDirToTar(sourceDir, tw)
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
	client, err := client.NewEnvClient()
	if err != nil {
		log.Fatalf("Unable to create docker client: %s", err)
	}

	// Client, imagename and Dockerfile location
	tags := []string{"this_is_a_imagename"}
	dockerFile := "mgt/images/core/Dockerfile"
	configFiles := []string{"mgt/images/core/requirements.txt"}
	fileDirs := []string{"mgt/bases/core"}
	err = buildImage(client, tags, dockerFile, configFiles, fileDirs)
	if err != nil {
		panic(err)
		//log.Fatal(err.Error())
	}
}
