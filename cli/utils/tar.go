package utils

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
)

type File struct {
	name string
}

type ArchiveFile struct {
	FilePath        string
	ArchiveFilePath string
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
func FileObjToTar(tw *tar.Writer, fileReader io.Reader, fileName string) error {
	// Read the actual Dockerfile
	readDockerFile, err := ioutil.ReadAll(fileReader)
	if err != nil {
		return err
	}

	// Make a TAR header for the file
	tarHeader := &tar.Header{
		Name: fileName,
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
	return nil
}

func FileToTar(sourceFile string, prefix string, tw *tar.Writer) (string, error) {

	_, fileName := filepath.Split(sourceFile)
	fileName = fmt.Sprintf("%s%s", prefix, fileName)
	// Create a filereader
	sourceFileReader, err := os.Open(sourceFile)
	if err != nil {
		return "", err
	}

	err = FileObjToTar(tw, sourceFileReader, fileName)
	if err != nil {
		return "", err
	}

	return fileName, err
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
		log.Println(fileName)
		header.Name = fileName
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

func CreateAcceptFileList(sourceDir string) (fileList []ArchiveFile, err error) {
	fileList = make([]ArchiveFile, 0)
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
	filepath.Walk(sourceDir, func(filePath string, fi os.FileInfo, err error) error {
		file := strings.Replace(strings.Replace(filePath, sourceDir, "", -1), string("\\"), "/", -1)
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

		fileList = append(fileList, ArchiveFile{
			FilePath:        filePath,
			ArchiveFilePath: fileName,
		})

		return nil
	})
	return
}

func CreateProjectTar(configFiles []string, configDirs []string, sourceDir string, tw *tar.Writer) error {

	_, baseFilePath, _, _ := runtime.Caller(1)

	projectFiles := make([]ArchiveFile, 0)

	for _, configFile := range configFiles {
		configFilePath := path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../../%s", configFile))
		projectFiles = append(projectFiles, ArchiveFile{
			FilePath:        configFilePath,
			ArchiveFilePath: configFile,
		})
	}

	// Add TarFile Dirs
	for _, configDir := range configDirs {
		configDir = path.Join(path.Dir(baseFilePath), fmt.Sprintf("../../../%s", configDir))
		fileList, err := CreateAcceptFileList(configDir)
		if err != nil {
			return err
		}
		projectFiles = append(projectFiles, fileList...)
	}

	fileList, err := CreateAcceptFileList(sourceDir)
	if err != nil {
		return err
	}
	projectFiles = append(projectFiles, fileList...)

	for _, prFile := range projectFiles {
		err = addFile(tw, prFile.FilePath, prFile.ArchiveFilePath)
		if err != nil {
			log.Println("Error ", err.Error())
		}
	}

	return nil
}
func addFile(tw *tar.Writer, path string, archivePath string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	if stat, err := file.Stat(); err == nil {
		// now lets create the header as needed for this file within the tarball
		header := new(tar.Header)
		header.Name = archivePath
		header.Size = stat.Size()
		header.Mode = int64(stat.Mode())
		header.ModTime = stat.ModTime()
		// write the header to the tarball archive
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		// copy the file data to the tarball
		if _, err := io.Copy(tw, file); err != nil {
			return err
		}
	}
	return nil
}

func ExtractTarGz(source string, target string) error {
	file, err := os.Open(source)

	archive, err := gzip.NewReader(file)

	if err != nil {
		fmt.Println("There is a problem with os.Open")
	}
	tr := tar.NewReader(archive)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fileTarget := filepath.Join(target, hdr.Name)
		fmt.Printf("Contents of %s %s \n", hdr.Name, fileTarget)

		//Using a bytes buffer is an important part to print the values as a string

		bud := new(bytes.Buffer)
		bud.ReadFrom(tr)

		os.MkdirAll(filepath.Dir(fileTarget), 0777)
		targetFile, _ := os.Create(fileTarget)
		io.Copy(targetFile, bud)
		targetFile.Close()
	}
	return err
}
