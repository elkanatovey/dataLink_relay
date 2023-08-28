package relay

type ConnectionRequest struct {
	Data       string `json:"Data"`
	ImporterID string `json:"ImporterID"`
	ExporterID string `json:"ExporterID"`
}

type ExporterAnnouncement struct {
	Data       string `json:"Data"`
	ExporterID string `json:"ExporterID"`
}

// ExporterResponse informs whether a ConnectionRequest was passed on successfully
type ExporterResponse struct {
	Message string
	Error   error
}
