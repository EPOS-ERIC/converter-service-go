package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/epos-eu/converter-service/connection"
	"github.com/epos-eu/converter-service/loggers"
)

func Handler(body string) (string, error) {
	var message Message

	loggers.EA_LOGGER.Debug("Handling message", "message", body)

	if err := json.Unmarshal([]byte(body), &message); err != nil {
		return "", fmt.Errorf("error converting payload: %v", err)
	}

	// validate the message
	if message.Payload == "" {
		return "", fmt.Errorf("error getting the payload: the payload is empty")
	}
	// both the distributionId and the pluginId must be specified
	if (message.Parameters.DistributionId == "" || message.Parameters.PluginId == "") {
		return "", fmt.Errorf("error: both the distributionId and the pluginId must be specified. distributionId: %s. pluginId: %s", message.Parameters.DistributionId, message.Parameters.PluginId)
	}

	plugin, err := connection.GetPluginById(message.Parameters.PluginId)
	if err != nil {
		return "", fmt.Errorf("error getting plugins: %v", err)
	}

	log.Printf("Executing plugin: %+v", plugin)

	switch plugin.Runtime {
	case "java":

		cmd := exec.Command("java",
			// Options needed for the EPOS-GEO-JSON library
			"--add-opens=java.base/java.util=ALL-UNNAMED",
			"--add-opens=java.base/sun.reflect.annotation=ALL-UNNAMED",

			"-cp",
			"./plugins/"+plugin.ID+"/"+plugin.Executable,
			plugin.Arguments)

		return executeCommand(message.Payload, cmd)
	case "python":
		cmd := exec.Command("venv/bin/python", plugin.Executable)
		cmd.Dir = filepath.Join("./plugins", plugin.ID)

		return executeCommand(message.Payload, cmd)
	case "go", "binary":
		cmd := exec.Command("./plugins/" + plugin.ID + "/" + plugin.Executable)

		return executeCommand(message.Payload, cmd)
	default:
		log.Printf("error: unknown runtime: %v", plugin.Runtime)
		response, err := json.Marshal("{}")
		if err != nil {
			return "", fmt.Errorf("error on creating json: %v", err)
		}
		return string(response), nil
	}
}
