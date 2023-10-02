package api

// ConnectionRequest is sent by an impoerter to an exporter
type ConnectionRequest struct {
	Data     string `json:"Data"`
	ClientID string `json:"ClientID"`
	ServerID string `json:"ServerID"`
}

// ConnectionAccept is sent by an exporter to an importer
type ConnectionAccept struct {
	Data     string `json:"Data"`
	ClientID string `json:"ClientID"`
	ServerID string `json:"ServerID"`
}

// ListenRequest is sent by an exporter opening a persistent connection to a relay
type ListenRequest struct {
	Data     string `json:"Data"`
	ServerID string `json:"ServerID"`
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

// address prefixes
const (
	MTLS string = "https://"
	TCP  string = "http://"
)
