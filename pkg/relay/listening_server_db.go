package relay

//This file contains the ListeningServerDB a lookup table for connectingClients to pass connection requests to listeningServers listening
//on the relay. The format of these messages is also defined here

import (
	"context"
	"errors"
	"mbg-relay/pkg/api"
	"sync"
)

// ClientData contains a message for an exporter and channel for communicating back to the respective importer
type ClientData struct {
	msg                  api.ConnectionRequest                  // message for exporter
	resultNotificationCh chan api.ForwardingSuccessNotification // report to importer routine if message was passed to exporter socket successfully @todo should this be bool?
}

func InitClientData(cr api.ConnectionRequest) *ClientData {
	importer := &ClientData{
		msg:                  cr,
		resultNotificationCh: make(chan api.ForwardingSuccessNotification, 1),
	}
	return importer
}

// ListeningServer is a relay side representation of a server listening for connections via the relay.
// It is created every time a server reaches out to create a persistent connection. Messages are passed to the
// handler maintaining the persistent connection via the channel below
type ListeningServer struct {
	ctx                  context.Context  // context is used to tell whether server connection is still open
	serverNotificationCh chan *ClientData // messages passed to this channel are to be forwarded to the exporter
	//@todo  maybe have the connection channel for the client stored here in map instead of own db?
}

func InitListeningServer(freshCTX context.Context) *ListeningServer {
	server := &ListeningServer{
		ctx:                  freshCTX,
		serverNotificationCh: make(chan *ClientData, 100),
	}
	return server
}

type ListeningServerDB struct {
	listeningServers map[string]*ListeningServer //map to store the listeningServers
	mx               sync.RWMutex                //RWMutex to protect the map
}

func InitListeningServerDB() *ListeningServerDB {
	db := &ListeningServerDB{
		listeningServers: make(map[string]*ListeningServer),
		mx:               sync.RWMutex{},
	}
	return db
}

// AddListeningServer is called when a new server wishes to advertise via the relay
func (db *ListeningServerDB) AddListeningServer(id string, exp *ListeningServer) {
	db.mx.Lock()
	db.listeningServers[id] = exp
	db.mx.Unlock()
}

// RemoveListeningServer is used for cleanup of listeningServers no longer advertising via the relay
func (db *ListeningServerDB) RemoveListeningServer(id string) {
	db.mx.Lock()
	delete(db.listeningServers, id)
	db.mx.Unlock()
}

// NotifyListeningServer return an error if the server to access does not exist in the db nil otherwise
func (db *ListeningServerDB) NotifyListeningServer(id string, msg *ClientData) error {
	db.mx.RLock()
	defer db.mx.RUnlock()
	if listeningServer, ok := db.listeningServers[id]; ok {
		listeningServer.serverNotificationCh <- msg
		return nil
	}
	var ErrNotFound = errors.New("listeningServer: " + id + " was not found")
	return ErrNotFound
}
