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
	Message Notification
	Error   error
}

type Notification string

const (
	NotePassed         Notification = "connection request passed to server" //success
	NoteServerConnLost Notification = "connection request failed server disconnected"
	NoteServerNoExist  Notification = "server requested not registered with relay"
	NoteFail           Notification = "connection request failed" // generic fail
)
