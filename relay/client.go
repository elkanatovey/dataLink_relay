package relay

import "net/http"

// ImportingClient imports a service
// server exports services via relay
type ImportingClient struct {
	Connection    *http.Client
	RelayURL      string // address of relay
	ExporterID    string
	maxBufferSize int
}
