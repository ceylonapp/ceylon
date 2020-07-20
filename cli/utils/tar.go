package utils

import (
	"archive/tar"
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	name string
}

type IgnoreList struct {
	ignoreList  []File
	excludeList []File
}

func CreateIgnore(file *os.File) IgnoreList {

	ignoreFiles := make([]File, 0)
	excludeFiles := make([]File, 0)

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.Trim(scanner.Text(), " ")
		if len(line) == 0 || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, `\#`) {
			line = strings.TrimPrefix(line, `\`)
		}

		if strings.HasPrefix(line, "!") {
			excludeFiles = append(excludeFiles, File{
				name: strings.TrimPrefix(line, "!"),
			})
		} else {
			ignoreFiles = append(ignoreFiles, File{
				name: line,
			})

		}
	}

	ig := IgnoreList{
		ignoreList:  ignoreFiles,
		excludeList: excludeFiles,
	}

	return ig
}

func (ig *IgnoreList) match(path string) bool {

	for _, file := range ig.ignoreList {
		if strings.HasPrefix(path, file.name) {
			return true
		}
	}

	return false
}

func FileToTar(sourceFile string, tw *tar.Writer) error {
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

func DirToTar(sourceDir string,
	tw *tar.Writer) error {
	ignoreFilePath := fmt.Sprintf("%s/.ceylonignore", sourceDir)
	ignoreFile, err := os.Open(ignoreFilePath)
	CheckError(err)
	var ignore IgnoreList
	if ignoreFile != nil {
		ignore = CreateIgnore(ignoreFile)
	}

	dir, err := os.Open(sourceDir)
	CheckError(err)
	defer dir.Close()

	// get list of files
	files, err := dir.Readdir(0)
	CheckError(err)

	log.Println("Number of files ", len(files))
	// walk path
	return filepath.Walk(sourceDir, func(file string, fi os.FileInfo, err error) error {
		file = strings.Replace(strings.Replace(file, sourceDir, "", -1), string("\\"), "/", -1)

		fileName := strings.Replace(file, sourceDir, "", 1)
		if strings.HasPrefix(fileName, "/") {
			fileName = strings.Replace(fileName, "/", "", 1)
		}

		if &ignore != nil {
			isIgnore := ignore.match(fileName)
			if isIgnore {
				return nil
			}

		}

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

		header.Name = file
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
