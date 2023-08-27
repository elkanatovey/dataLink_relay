package relay

import (
	"time"
)

type ConnectionRequest struct {
	timestamp  time.Time `json:"time"`
	Data       string    `json:"Data"`
	ImporterID string    `json:"ImporterID"`
	ExporterID string    `json:"ExporterID"`
}

type ExporterAnnouncement struct {
	timestamp  time.Time `json:"time"`
	Data       string    `json:"Data"`
	ExporterID string    `json:"ExporterID"`
}
