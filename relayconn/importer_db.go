package relayconn

//This file contains the importerDB a lookup table for importers to pass connection requests to importers waiting for a callback
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

// Importer is created every time a client dials a server and is waiting for a socket that connects to server at relay.
type Importer struct {
	ctx        context.Context  // context is used to tell whether client connection is still open
	sockPassCh chan *ServerConn // channel is for passing socket that connects to server
}

func InitImporter(freshCTX context.Context) *Importer {
	importer := &Importer{
		ctx:        freshCTX, //@todo when to timeout on request?
		sockPassCh: make(chan *ServerConn),
	}
	return importer
}

type ImporterDB struct {
	importers map[string]*Importer //map to store the importers
	mx        sync.RWMutex         //RWMutex to protect the map
}

func InitImporterDB() *ImporterDB {
	db := &ImporterDB{
		importers: make(map[string]*Importer),
		mx:        sync.RWMutex{},
	}
	return db
}

// AddImporter is called when an importer is waiting on a connection to finish connecting
func (db *ImporterDB) AddImporter(id string, imp *Importer) {
	db.mx.Lock()
	db.importers[id] = imp
	db.mx.Unlock()
}

// RemoveImporter is used for cleanup of importers that no longer are waiting on a connection via the relay
func (db *ImporterDB) RemoveImporter(id string) {
	db.mx.Lock()
	delete(db.importers, id)
	db.mx.Unlock()
}

// NotifyImporter return an error if the server to access does not exist in the db nil otherwise
func (db *ImporterDB) NotifyImporter(id string, connection *ServerConn) error {
	db.mx.RLock()
	defer db.mx.RUnlock()
	if importer, ok := db.importers[id]; ok {
		importer.sockPassCh <- connection
		return nil
	}
	var ErrNotFound = errors.New("server was not found")
	return ErrNotFound
}
