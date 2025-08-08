package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os/exec"
	"path/filepath"

	"github.com/epos-eu/converter-service/db"
	"github.com/epos-eu/converter-service/logging"
)

var log = logging.Get("default")

func ExternalAccessHandler(bytes []byte) ([]byte, error) {
	body := string(bytes)

	var message Message

	log.Debug("Handling message", "message", body)

	if err := json.Unmarshal([]byte(body), &message); err != nil {
		return nil, fmt.Errorf("error converting payload: %v", err)
	}

	// validate the message
	if message.Payload == "" {
		return nil, fmt.Errorf("error getting the payload: the payload is empty")
	}
	// both the distributionId and the pluginId must be specified
	if message.Parameters.DistributionID == "" || message.Parameters.PluginID == "" {
		return nil, fmt.Errorf("error: both the distributionId and the pluginId must be specified. distributionId: %s. pluginId: %s", message.Parameters.DistributionID, message.Parameters.PluginID)
	}

	plugin, err := db.GetPluginById(message.Parameters.PluginID)
	if err != nil {
		return nil, fmt.Errorf("error getting plugins: %v", err)
	}

	log.Info("executing plugin",
		slog.Group("plugin",
			"id", plugin.ID,
			"name", plugin.Name,
			"version", plugin.Version,
			"version type", plugin.VersionType,
			"runtime", plugin.Runtime,
			"arguments", plugin.Arguments))

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
		log.Error("unknown runtime", "plugin runtime", plugin.Runtime)
		response, err := json.Marshal("{}")
		if err != nil {
			return nil, fmt.Errorf("error on creating json: %v", err)
		}
		return response, nil
	}
}

type relation struct {
	PluginID     string `json:"pluginId"`
	InputFormat  string `json:"inputFormat"`
	OutputFormat string `json:"outputFormat"`
}

type plugin struct {
	DistributionID string     `json:"distributionId"`
	Relations      []relation `json:"relations"`
}

type resourcesMsg struct {
	Plugins string `json:"plugins"`
}

func ResourcesServiceHandler(bytes []byte) ([]byte, error) {
	var resourcesMsg resourcesMsg
	err := json.Unmarshal(bytes, &resourcesMsg)
	if err != nil || resourcesMsg.Plugins != "all" {
		return nil, fmt.Errorf("failed to process the message: %w", err)
	}

	// get all plugin relations
	relations, err := db.GetPluginRelationForEnabledPlugins()
	if err != nil {
		return nil, fmt.Errorf("failed to get plugin relations: %w", err)
	}
	// group them by OperationID
	operations := make(map[string][]relation)
	for _, r := range relations {
		operations[r.RelationID] = append(operations[r.RelationID], relation{
			PluginID:     r.PluginID,
			InputFormat:  r.InputFormat,
			OutputFormat: r.OutputFormat,
		})
	}

	responseStr := make([]plugin, 0)
	for k, v := range operations {
		responseStr = append(responseStr, plugin{
			DistributionID: k,
			Relations:      v,
		})
	}

	return json.Marshal(responseStr)
}
