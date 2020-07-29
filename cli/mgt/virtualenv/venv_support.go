package virtualenv

import (
	"archive/tar"
	"ceylon/cli/config"
	"ceylon/cli/utils"
	"ceylon/cli/utils/fileutil"
	"compress/gzip"
	"context"
	"fmt"
	"log"
	"os"
	"path"
)

type VirtualEnvService struct {
	Context      context.Context
	BaseLocation *string
	DeployConfig *config.DeployConfig
}

func (s *VirtualEnvService) initiateLocation() (string, error) {
	projectsRuntimePath := "J:\\BotFramework\\ceylon\\tmp"
	projectDir := path.Join(projectsRuntimePath, fmt.Sprintf("%s", s.DeployConfig.Name))

	// Create project path
	_, err := os.Stat(projectDir)

	if os.IsNotExist(err) {
		err := os.MkdirAll(projectDir, 0777)
		if err != nil {
			panic(err)
		}
	} else {
		err := fileutil.RemoveContents(projectDir)
		if err != nil {
			log.Println(err.Error())
			print(err)
		}
	}

	// set up the output file
	projectArchivePath := path.Join(projectDir, "project.tar.gz")
	projectArchive, err := os.Create(projectArchivePath)
	if err != nil {
		panic(err)
	}
	defer projectArchive.Close()
	// set up the gzip writer
	gw := gzip.NewWriter(projectArchive)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()

	configFiles := []string{"mgt/images/core/requirements.txt"}
	fileDirs := []string{"mgt/bases/core"}

	err = utils.CreateProjectTar(configFiles, fileDirs, *s.BaseLocation, tw)
	if err != nil {
		panic(err)
	}
	tw.Close()
	gw.Close()
	projectArchive.Close()

	err = utils.ExtractTarArchive(projectArchivePath, projectDir, true)
	if err != nil {
		panic(err)
	}
	//err = utils.ExtractTarArchive(filepath.Join(projectDir, "mgt/libs/windows/venv.tar.gz"), projectDir, false)
	//if err != nil {
	//	log.Fatal(err)
	//}

	return projectDir, nil
}
