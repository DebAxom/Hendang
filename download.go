package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
)

func DownloadFile(url string, dest string, wg *sync.WaitGroup) error {

	defer wg.Done()

	resp, err := http.Get(url)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status code: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return nil
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func DownloadFiles(url string, chunks int, dest string) error {

	var wg sync.WaitGroup

	for i := range chunks {
		chunkURL := url + "/" + strconv.Itoa(i+1) + ".chunk"
		wg.Add(1)
		go DownloadFile(chunkURL, path.Join(dest, strconv.Itoa(i+1)+".chunk"), &wg)
	}

	wg.Wait()

	return nil
}

func ContinueDownload(url string, files []int, dest string) error {

	var wg sync.WaitGroup

	for _, file := range files {
		chunkURL := url + "/" + strconv.Itoa(file) + ".chunk"
		wg.Add(1)
		go DownloadFile(chunkURL, path.Join(dest, strconv.Itoa(file)+".chunk"), &wg)
	}

	wg.Wait()

	return nil
}

type metaData struct {
	chunk int
	ext   string
}

func GetMetadata(url string) (metaData, error) {
	resp, err := http.Get(url)
	if err != nil {
		return metaData{}, err
	}

	defer resp.Body.Close()

	var data struct {
		Chunks int    `json:"chunks"`
		Ext    string `json:"ext"`
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return metaData{}, err
	}

	return metaData{chunk: data.Chunks, ext: data.Ext}, nil
}

func missingChunks(max int, folder string) ([]int, error) {
	// Step 1: Scan folder
	files, err := os.ReadDir(folder)
	if err != nil {
		return nil, err
	}

	// Step 2: Store found chunk numbers in a map
	exists := make(map[int]bool)
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".chunk") {
			numStr := strings.TrimSuffix(name, ".chunk")
			if num, err := strconv.Atoi(numStr); err == nil {
				exists[num] = true
			}
		}
	}

	// Step 3: Build result excluding existing chunks
	var result []int
	for i := 1; i <= max; i++ {
		if !exists[i] {
			result = append(result, i)
		}
	}

	return result, nil
}
