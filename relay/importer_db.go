package relay

//This file contains the importerDB a lookup table for importers to pass connection requests to importers waiting for a callback
//on the relay. The format of these messages is also defined here

import (
	"context"
	//"errors"
	"net"
	"sync"
)

type ServerConn struct {
	conn net.Conn
	err  error
}

// Importer is created every time a client dials a server and is waiting for a socket that connects to server at relay.
type Importer struct {
	ctx            context.Context // context is used to tell whether client connection is still open
	exporterConnCh chan ServerConn // channel is for passing socket that connects to server
}

func InitImporter(freshCTX context.Context) *Importer {
	importer := &Importer{
		ctx:            freshCTX, //@todo when to timeout on request?
		exporterConnCh: make(chan ServerConn),
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
