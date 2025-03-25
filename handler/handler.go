package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"

	"github.com/epos-eu/converter-service/connection"
)

func Handler(body string) (string, error) {
	var message Message

	if err := json.Unmarshal([]byte(body), &message); err != nil {
		return "", fmt.Errorf("error converting payload: %v", err)
	}

	// validate the message
	if message.Payload == "" {
		return "", fmt.Errorf("error getting the payload: the payload is nil")
	}
	// either the operationId or the pluginId must be set to convert the message
	if message.Parameters.OperationId == "" && message.Parameters.PluginId == "" {
		return "", fmt.Errorf("error: both the operationId and the pluginId are nil")
	}
	// if the pluginId is not given, then the responseContentType is needed to convert the message
	if message.Parameters.PluginId == "" && message.Parameters.ResponseFormat == "" {
		return "", fmt.Errorf("error: both the pluginId and the ResponseFormat are nil")
	}

	if message.Parameters.PluginId == "" {
		pluginId, err := guessPluginId(message.Parameters)
		if err != nil {
			return "", err
		}
		message.Parameters.PluginId = pluginId
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

// Try to guess the plugin id for a conversion using the given formats (input/output)
func guessPluginIdUsingOriginalFormats(params Parameters) (string, error) {
	pluginId := ""

	pluginRelations, err := connection.GetPluginRelationsByOperationId(params.OperationId)
	if err != nil {
		return "", fmt.Errorf("error getting plugins relations: %v", err)
	}
	// filter the relations using the correct request and response format
	for _, pluginRelation := range pluginRelations {
		if params.RequestFormat == "" {
			if pluginRelation.OutputFormat == params.ResponseFormat {
				pluginId = pluginRelation.PluginID
				break
			}
		} else {
			if pluginRelation.InputFormat == params.RequestFormat && pluginRelation.OutputFormat == params.ResponseFormat {
				pluginId = pluginRelation.PluginID
				break
			}
		}
	}

	if pluginId == "" {
		return "", fmt.Errorf("could not guess pluginId using the given formats: \n\tInput format: %s\n\tOutput format: %s", params.RequestFormat, params.ResponseFormat)
	}
	return pluginId, nil
}

// Try to guess the plugin id for a conversion using the parsed format of the payload
func guessPluginIdUsingPayloadFormat(params Parameters) (string, error) {
	originalRequestFormat := params.RequestFormat
	parsed, err := StringToContentType(originalRequestFormat)
	if err != nil {
		return "", err
	}
	params.RequestFormat = string(parsed)
	pluginId, err := guessPluginIdUsingOriginalFormats(params)
	if err != nil {
		return "", err
	}

	if pluginId == "" {
		return "", fmt.Errorf("cannot infer the pluginId from the operationId and format:\nOperationId: %s\nOriginalRequestFormat: %s\nParsedRequestFormat: %s\nResponseFormat: %s", params.OperationId, originalRequestFormat, params.RequestFormat, params.ResponseFormat)
	}
	return pluginId, nil
}

// Try to guess the plugin for a conversion
func guessPluginId(parameters Parameters) (string, error) {
	pluginId := ""
	pluginId, err := guessPluginIdUsingOriginalFormats(parameters)
	if err != nil {
		log.Printf("could not guess the puling id (#1): %v", err)

		// try to guess by parsing the format of the payload
		pluginId, err = guessPluginIdUsingPayloadFormat(parameters)
		if err != nil {
			log.Printf("could not guess the puling id (#2): %v", err)

			pluginId, err = guessPluginIdFromOutputFormat(parameters)
			if err != nil {
				log.Printf("could not guess the puling id (#3): %v", err)

				// try to use the first plugin connected with this operation id anyway (method #4)
				pluginRelations, err := connection.GetPluginRelationsByOperationId(parameters.OperationId)
				if err != nil {
					return "", fmt.Errorf("error getting plugins relations: %v", err)
				}
				plugin, err := connection.GetPluginById(pluginRelations[0].PluginID)
				if err != nil {
					return "", fmt.Errorf("error getting plugins: %v", err)
				}
				pluginId = plugin.ID
			}
		}
	}

	if pluginId == "" {
		return "", fmt.Errorf("could not infer the pluginId for the conversion")
	}
	return pluginId, nil
}

func guessPluginIdFromOutputFormat(params Parameters) (string, error) {
	pluginId := ""

	pluginRelations, err := connection.GetPluginRelationsByOperationId(params.OperationId)
	if err != nil {
		return "", fmt.Errorf("error getting plugins relations: %v", err)
	}
	// find the first plugin associated with this operationId that has as output format the requested output format
	for _, pluginRelation := range pluginRelations {
		if pluginRelation.OutputFormat == params.ResponseFormat {
			pluginId = pluginRelation.PluginID
			break
		}
	}

	if pluginId == "" {
		return "", fmt.Errorf("could not guess pluginId from the output format: Output format: %s", params.ResponseFormat)
	}
	return pluginId, nil
}
