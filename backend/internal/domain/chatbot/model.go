package chatbot

import "time"

type AskRequest struct {
	Message string `json:"message"`
	Year    int    `json:"year,omitempty"`
}

type ChatResponse struct {
	Intent          string           `json:"intent"`
	Title           string           `json:"title"`
	Answer          string           `json:"answer"`
	Metrics         []Metric         `json:"metrics"`
	Recommendations []string         `json:"recommendations"`
	Chart           *ChartData       `json:"chart,omitempty"`
	Actions         []Action         `json:"actions"`
	GeneratedAt     time.Time        `json:"generated_at"`
}

type Metric struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type ChartData struct {
	Type   string        `json:"type"`
	Title  string        `json:"title"`
	Labels []string      `json:"labels"`
	Values []interface{} `json:"values"`
	Format string        `json:"format"`
}

type Action struct {
	Label string `json:"label"`
	Url   string `json:"url"`
}
