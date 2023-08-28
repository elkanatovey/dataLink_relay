package relay

import (
	"context"
	"net/http"
)

// server exports services via relay
type ExportingServer struct {
	Connection *http.Client
	URL        string
}

func (s *ExportingServer) AdvertiseService() error {
	return nil
}

// request opens the connection to the relay
func (s *ExportingServer) request(ctx context.Context) (*http.Response, error) {
	req, err := http.NewRequest("GET", s.URL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx) //todo integrate with exporterid

	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Connection", "keep-alive")
	return s.Connection.Do(req)
}

func (s *ExportingServer) runServer() error {

	return nil
}
