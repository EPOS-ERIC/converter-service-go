package handler

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/epos-eu/converter-service/loggers"
)

func executeCommand(payload string, cmd *exec.Cmd) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	// Generate random unique names for the temp input and output files
	tmpDir, inputFile, outputFile, err := createTempFiles(currentDir, payload)
	if err != nil {
		return "", err
	}
	defer cleanupTempFiles(tmpDir)

	cmd.Args = append(cmd.Args, inputFile, outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		// log the head of the payload that could not be converted for debugging purposes
		return "", fmt.Errorf("error executing the plugin: %w\nHead of payload:\n%v", err, getHead(payload, 100))
	}

	output, err := os.ReadFile(outputFile)
	if err != nil {
		return "", fmt.Errorf("error reading output file: %w", err)
	}

	var outputMap map[string]any
	if err := json.Unmarshal(output, &outputMap); err != nil {
		return "", fmt.Errorf("error parsing output json: %w", err)
	}

	response := Response{outputMap}

	jsonStr, err := json.Marshal(response)
	if err != nil {
		return "", fmt.Errorf("error converting output to json: %w", err)
	}

	return string(jsonStr), nil
}

func createTempFiles(dir, payload string) (string, string, string, error) {
	// get a name that is not used in the current dir
	tmpDir, err := getUniqueFileName(dir)
	if err != nil {
		return "", "", "", err
	}
	// create a temp dir with the unique name
	err = os.Mkdir(tmpDir, os.ModeTemporary)
	if err != nil {
		return "", "", "", err
	}
	// the input and output file will be in the temp dir
	inputFile := filepath.Join(dir, tmpDir, "input")
	outputFile := filepath.Join(dir, tmpDir, "output")
	// create the input file and put the payload in it
	if err := os.WriteFile(inputFile, []byte(payload), 0644); err != nil {
		return "", "", "", fmt.Errorf("error writing to temp input file: %w", err)
	}

	return tmpDir, inputFile, outputFile, nil
}

func cleanupTempFiles(files ...string) {
	for _, file := range files {
		if err := os.RemoveAll(file); err != nil {
			loggers.EA_LOGGER.Printf("error removing temp dir: %v\n", err)
		}
	}
}

func getUniqueFileName(path string) (string, error) {
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		name := randomString(10)
		fullPath := filepath.Join(path, name)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			return name, nil
		}
	}
	return "", fmt.Errorf("could not generate a unique file name in this directory: %s", path)
}

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func randomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// getHead returns the first `length` characters of the `str` string
func getHead(str string, length int) string {
	if len(str) <= length {
		return str
	}
	return str[:length]
}
