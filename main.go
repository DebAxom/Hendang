package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path"
	"strings"
	"time"
)

type FileChunk struct {
	Dir    string `json:"dir"`
	Chunks int    `json:"chunks"`
	Ext    string `json:"ext"`
}

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func RandString() string {
	seededRand := rand.New(rand.NewSource(time.Now().UnixNano())) // Seed the random number generator
	b := make([]byte, 5)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

func main() {

	nArgs := len(os.Args)

	if nArgs == 0 || nArgs == 1 {
		fmt.Println("Usage: hendang <command> [arguments]")
		return
	}

	cmd := os.Args[1]

	if cmd == "break" {
		if nArgs < 3 {
			fmt.Println("Usage: hendang break <file> <output_dirname (oprional) >")
			return
		}

		fileName := os.Args[2]
		outpurDirName := fileName + ".chd"

		if nArgs > 3 {
			outpurDirName = strings.TrimSuffix(os.Args[3], ".chd") + ".chd"
		}

		totalChunks, err := BreakFile(fileName, outpurDirName)
		if err != nil {
			fmt.Println(err.Error())
		}

		fmt.Println(fileName, "has been broken down into", totalChunks, "chunks !")
		return
	}

	if cmd == "merge" {

		if nArgs < 3 {
			fmt.Println("Usage: hendang merge <dirname.chd>")
			return
		}

		dirname := os.Args[2]
		outputFilePath := dirname[:len(dirname)-4]

		if !strings.HasSuffix(dirname, ".chd") {
			fmt.Println("The directory name must end with .chd")
			return
		}

		if nArgs > 3 {
			outputFilePath = os.Args[3]
		}

		var metadata struct {
			Ext string `json:"ext"`
		}

		metadataFile, _ := os.ReadFile(path.Join(dirname, "metadata.json"))
		json.Unmarshal(metadataFile, &metadata)

		outputFilePath = strings.TrimSuffix(outputFilePath, path.Ext(outputFilePath)) + metadata.Ext

		err := MergeChunks(dirname, outputFilePath)

		if err != nil {
			fmt.Println(err.Error())
		}

		return
	}

	if cmd == "download" {

		if nArgs < 4 {
			fmt.Println("Usage: hendang download <url> <filename>")
			return
		}

		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("Error getting user config directory:", err)
			return
		}

		cliDir := path.Join(homeDir, ".hendang")
		dataFile := path.Join(cliDir, "data.json")

		if _, err := os.Stat(cliDir); os.IsNotExist(err) {
			os.Mkdir(cliDir, 0755)                    // Create the directory
			os.Mkdir(path.Join(cliDir, ".tmp"), 0755) // Create temp directory
			file, err := os.Create(dataFile)          // Create the data file
			if err != nil {
				fmt.Println("An error occured :", err)
				return
			}
			file.Write([]byte("{}"))
			file.Close()
		}

		data, _ := os.ReadFile(dataFile)

		files := make(map[string]FileChunk)
		json.Unmarshal(data, &files)

		URL := strings.TrimSuffix(os.Args[2], "/")
		outputFilePath := os.Args[3]

		if path.Ext(URL) != ".chd" {
			fmt.Println("Invalid folder : The folder name should end with .chd")
			return
		}

		tmpDir := path.Join(cliDir, ".tmp/"+RandString())

		if val, ok := files[URL]; ok {
			fmt.Println("Resuming download for", URL)
			tmpDir = val.Dir

			missingFiles, err := missingChunks(val.Chunks, val.Dir)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = ContinueDownload(URL, missingFiles, tmpDir)
			if err != nil {
				fmt.Println(err)
				return
			}

		} else {
			fmt.Println("Starting new download for", URL)
			metaData, err := GetMetadata(URL + "/metadata.json")

			os.Mkdir(tmpDir, 0755)

			if err != nil {
				fmt.Println("An error occured while fetching metadata :", err)
				return
			}

			files[URL] = FileChunk{
				Dir:    tmpDir,
				Chunks: metaData.chunk,
				Ext:    metaData.ext,
			}

			data, _ = json.Marshal(files)
			os.WriteFile(dataFile, data, 0644)

			err = DownloadFiles(URL, metaData.chunk, tmpDir)
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		outputFilePath = strings.TrimSuffix(outputFilePath, path.Ext(outputFilePath)) + files[URL].Ext

		err = MergeChunks(tmpDir, outputFilePath)

		if err != nil {
			fmt.Println("An error occured :", err)
		}

		os.RemoveAll(tmpDir)
		delete(files, URL)

		data, _ = json.Marshal(files)
		os.WriteFile(dataFile, data, 0644)

		fmt.Println("Download Complete !")
		return
	}

	if cmd == "reset" {
		homeDir, _ := os.UserHomeDir()
		cliDir := path.Join(homeDir, ".hendang")
		tmpDir := path.Join(cliDir, ".tmp")
		dataFile := path.Join(cliDir, "data.json")

		os.Remove(dataFile)  // Remove the data file
		os.RemoveAll(tmpDir) // Remove the temp directory

		file, err := os.Create(dataFile)
		if err != nil {
			fmt.Println("An error occured :", err)
			return
		}
		file.Write([]byte("{}"))
		file.Close()
		os.Mkdir(tmpDir, 0755) // Recreate the temp directory
		return
	}

	fmt.Println(cmd, "is not a valid command !")

}
