package handler

type Message struct {
	Parameters Parameters `json:"parameters"`
	Payload    string     `json:"content"`
}

type Parameters struct {
	PluginID       string `json:"pluginId,omitempty"`
	DistributionID string `json:"distributionId"`
	RequestFormat  string `json:"requestContentType,omitempty"`
	ResponseFormat string `json:"responseContentType,omitempty"`
}

type Response struct {
	Payload map[string]any `json:"content"`
}
