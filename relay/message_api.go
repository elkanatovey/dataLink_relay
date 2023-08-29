package relay

// ConnectionRequest is sent by an impoerter to an exporter
type ConnectionRequest struct {
	Data       string `json:"Data"`
	ImporterID string `json:"ImporterID"`
	ExporterID string `json:"ExporterID"`
}

// ExporterAnnouncement is sent by an exporter opening a persistent connection to a relay
type ExporterAnnouncement struct {
	Data       string `json:"Data"`
	ExporterID string `json:"ExporterID"`
}

// ExporterResponse informs whether a ConnectionRequest was passed on successfully
type ExporterResponse struct {
	Message string
	Error   error
}
