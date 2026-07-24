package relay

//This file contains the importerDB a lookup table for connectingClients to pass connection requests to connectingClients waiting for a callback
//on the relay. The format of these messages is also defined here

import (
	"context"
	"errors"
	"net"
	"sync"
)

type ServerConn struct {
	conn net.Conn
	err  error
}

// ConnectingClient is created every time a client dials a server and is waiting for a socket that connects to server at relay.
type ConnectingClient struct {
	ctx        context.Context  // context is used to tell whether client connection is still open
	sockPassCh chan *ServerConn // channel is for passing socket that connects to server
}

func InitConnectingClient(freshCTX context.Context) *ConnectingClient {
	importer := &ConnectingClient{
		ctx:        freshCTX, //@todo when to timeout on request?
		sockPassCh: make(chan *ServerConn, 1),
	}
	return importer
}

type connectingClientDB struct {
	connectingClients map[string]*ConnectingClient //map to store the connectingClients
	mx                sync.RWMutex                 //RWMutex to protect the map
}

func initConnectingClientDB() *connectingClientDB {
	db := &connectingClientDB{
		connectingClients: make(map[string]*ConnectingClient),
		mx:                sync.RWMutex{},
	}
	return db
}

// AddConnectingClient is called when a ConnectingClient is waiting on a connection to finish connecting
func (db *connectingClientDB) AddConnectingClient(id string, imp *ConnectingClient) {
	db.mx.Lock()
	db.connectingClients[id] = imp
	db.mx.Unlock()
}

// RemoveConnectingClient is used for cleanup of a ConnectingClient that is no longer is waiting on a connection via the relay
func (db *connectingClientDB) RemoveConnectingClient(id string) {
	db.mx.Lock()
	delete(db.connectingClients, id)
	db.mx.Unlock()
}

// NotifyConnectingClient hands a callback socket to a waiting client. It returns an error if the client is not
// registered or already has a pending socket; in that case the caller still owns connection and must close it.
func (db *connectingClientDB) NotifyConnectingClient(id string, connection *ServerConn) error {
	db.mx.RLock()
	defer db.mx.RUnlock()
	if importer, ok := db.connectingClients[id]; ok {
		select {
		case importer.sockPassCh <- connection:
			return nil
		default:
			return errors.New("client: " + id + " already has a pending connection")
		}
	}
	return errors.New("server: " + id + " was not found")
}

// removeAndDrainConnectingClient removes a client that stopped waiting and closes any callback socket that was
// delivered but never consumed, so a late callback cannot leak the underlying connection.
func (db *connectingClientDB) removeAndDrainConnectingClient(id string, imp *ConnectingClient) {
	db.RemoveConnectingClient(id)
	select {
	case sc := <-imp.sockPassCh:
		if sc != nil && sc.conn != nil {
			sc.conn.Close()
		}
	default:
	}
}
