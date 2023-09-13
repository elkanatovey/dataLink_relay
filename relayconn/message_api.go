package relayconn

// ConnectionRequest is sent by an impoerter to an exporter
type ConnectionRequest struct {
	Data       string `json:"Data"`
	ImporterID string `json:"ImporterID"`
	ExporterID string `json:"ExporterID"`
}

// ConnectionAccept is sent by an exporter to an importer
type ConnectionAccept struct {
	Data       string `json:"Data"`
	ImporterID string `json:"ImporterID"`
	ExporterID string `json:"ExporterID"`
}

// ExporterAnnouncement is sent by an exporter opening a persistent connection to a relay
type ExporterAnnouncement struct {
	Data       string `json:"Data"`
	ExporterID string `json:"ExporterID"`
}

// ForwardingSuccessNotification informs whether a ConnectionRequest was passed on successfully to the listening exporter by the relay
type ForwardingSuccessNotification struct {
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

// handles for our api
const (
	Dial   string = "/clientconn"
	Listen string = "/serverconn"
	Accept string = "/servercallback"
)
