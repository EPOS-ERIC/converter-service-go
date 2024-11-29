package handler

import (
	"encoding/json"
	"fmt"
	"github.com/epos-eu/converter-service/loggers"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
)

func executeCommand(payload string, cmd *exec.Cmd) (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting current directory: %w", err)
	}

	// Generate random unique names for the temp input and output files
	inputFile, outputFile, err := createTempFiles(currentDir, payload)
	if err != nil {
		return "", err
	}
	defer cleanupTempFiles(inputFile, outputFile)

	cmd.Args = append(cmd.Args, inputFile, outputFile)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("error executing the plugin: %w", err)
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

func createTempFiles(dir, payload string) (string, string, error) {
	fileName, err := getUniqueFileName(dir)
	if err != nil {
		return "", "", err
	}
	inputFile := filepath.Join(dir, "payload_"+fileName+".input")
	outputFile := filepath.Join(dir, "payload_"+fileName+".output")

	if err := os.WriteFile(inputFile, []byte(payload), 0644); err != nil {
		return "", "", fmt.Errorf("error writing to temp input file: %w", err)
	}

	return inputFile, outputFile, nil
}

func cleanupTempFiles(files ...string) {
	for _, file := range files {
		if err := os.Remove(file); err != nil {
			loggers.EA_LOGGER.Printf("error removing temp file %s: %v\n", file, err)
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
