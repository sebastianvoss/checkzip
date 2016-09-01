package main

import (
	"archive/zip"
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"
	"strings"
)

func main() {

	directory := flag.String("directory", ".", "Directory to scan for zip files")
	bannedExtensions := flag.String("extensions", ".exe", "Comma separated list of banned extensions")
	maxLevel := flag.Int("max", 20, "Maximum number of nested zip files to check")
	flag.Parse()

	foundFiles := scanDirectory(*directory, *bannedExtensions, *maxLevel)

	printFoundFiles(foundFiles)
}

func scanDirectory(directory, bannedExtensions string, maxLevel int) []string {
	var foundFiles []string

	files, err := ioutil.ReadDir(directory)
	check(err)

	for _, f := range files {
		extension := filepath.Ext(f.Name())
		if extension == ".zip" {
			filePath := path.Join(directory, f.Name())
			b := readFile(filePath)
			files := scanZip(b, bannedExtensions, 0, maxLevel, nil)
			foundFiles = append(foundFiles, files...)
		}
	}

	return foundFiles
}

func scanZip(b []byte, extensions string, level, maxLevel int, foundFiles []string) []string {
	if level > maxLevel {
		return foundFiles
	}

	readerAt := bytes.NewReader(b)
	zipReader, err := zip.NewReader(readerAt, readerAt.Size())
	if err != nil {
		fmt.Println("Error opening zip file")
		return nil
	}

	for _, f := range zipReader.File {
		extension := filepath.Ext(f.Name)
		if extension == ".zip" {
			b := readZip(f)
			files := scanZip(b, extensions, level+1, maxLevel, foundFiles)
			foundFiles = append(foundFiles, files...)
		} else {
			if checkExtension(extension, extensions) {
				foundFiles = append(foundFiles, f.Name)
			}
		}
	}

	return foundFiles

}

func checkExtension(extension, bannedExtensions string) bool {
	return extension != "" && strings.Contains(bannedExtensions, extension[1:])
}

func readFile(filePath string) []byte {
	b, err := ioutil.ReadFile(filePath)
	check(err)
	return b
}

func readZip(f *zip.File) []byte {
	readCloser, err := f.Open()
	check(err)
	defer readCloser.Close()

	b, err := ioutil.ReadAll(readCloser)
	check(err)
	return b
}

func printFoundFiles(foundFiles []string) {
	if len(foundFiles) > 0 {
		fmt.Print("Found banned file extensions '")
		for i, f := range foundFiles {
			if i > 0 {
				fmt.Print(",")
			}
			fmt.Print(f)
		}
		fmt.Println("'")
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
