package main

import (
	"fmt"
	"io"
	"math"
	"os"
	"path"
	"strconv"
	"strings"
)

func makeChunks(filePath string, dirname string, chunkSize int) error {
	file, err := os.OpenFile(filePath, os.O_RDONLY, 0644)

	if err != nil {
		return nil
	}

	defer file.Close()

	buf := make([]byte, chunkSize)

	count := 0

	for {
		count++
		numBytes, err := file.Read(buf)

		if err != nil && err != io.EOF {
			return nil
		}

		if numBytes == 0 || err == io.EOF {
			break
		}

		chunk := make([]byte, numBytes)
		copy(chunk, buf[:numBytes])
		os.WriteFile(path.Join(dirname, strconv.Itoa((count))+".chunk"), chunk, 0644)

	}

	return nil
}

func MergeChunks(dirPath string, outputFilePath string) error {

	files, err := findChunkFiles(dirPath)
	if err != nil {
		return err
	}

	outputFile, _ := os.Create(outputFilePath)
	outputFile.Close()

	outputFile, err = os.OpenFile(outputFilePath, os.O_WRONLY|os.O_APPEND, 0644)

	for index := range files {
		data, _ := os.ReadFile(path.Join(dirPath, strconv.Itoa(index+1)+".chunk"))
		_, err := outputFile.Write(data)
		if err != nil {
			return err
		}
	}

	return nil
}

func BreakFile(name string, dirname string) (int, error) {

	fileInfo, err := os.Stat(name)

	if err != nil {
		fmt.Println("Error:", err.Error())
		return 0, err
	}

	totalFiles := int(math.Ceil(float64(fileInfo.Size()) / (1024 * 3)))

	if _, err := os.Stat(dirname); os.IsNotExist(err) {
		os.Mkdir(dirname, 0755) // Create new directory
	} else {
		os.RemoveAll(dirname)   // Remove existing directory
		os.Mkdir(dirname, 0755) // Create new directory
	}

	metadataFIle, _ := os.Create(path.Join(dirname, "metadata.json"))
	metadataFIle.WriteString("{ \"chunks\" : " + strconv.Itoa(totalFiles) + ", \"ext\" : \"" + path.Ext(name) + "\" }")
	metadataFIle.Close()

	makeChunks(name, dirname, 3*1024) // 3 KB chunks

	return totalFiles, nil
}

func findChunkFiles(dir string) ([]string, error) {
	var chunkFiles []string

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".chunk") {
			chunkFiles = append(chunkFiles, entry.Name())
		}
	}

	return chunkFiles, nil
}
