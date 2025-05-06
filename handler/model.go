package handler

import (
	"fmt"
	"strings"
)

type Message struct {
	Parameters Parameters `json:"parameters"`
	Payload    string     `json:"content"`
}

type Parameters struct {
	PluginId       string `json:"pluginId,omitempty"`
	DistributionId string `json:"distributionId"`
	RequestFormat  string `json:"requestContentType,omitempty"`
	ResponseFormat string `json:"responseContentType,omitempty"`
}

type Response struct {
	Payload map[string]any `json:"content"`
}

type ContentType string

const (
	plainJson = "application/json"
	covJson   = "application/covjson"
	geoJson   = "geojson"
	plainXml  = "application/xml"
)

func StringToContentType(s string) (ContentType, error) {
	s = strings.ToLower(s)
	switch {
	case strings.Contains(s, "xml"):
		return plainXml, nil
	case strings.Contains(s, "epos.geo+json"):
		return geoJson, nil
	case strings.Contains(s, "covjson"):
		return covJson, nil
	case strings.Contains(s, "json"):
		return plainJson, nil
	default:
		return "", fmt.Errorf("error converting string '%s' to ContentType", s)
	}
}
