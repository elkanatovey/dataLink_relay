package relay

//This file contains entities that are managed within the relay

import (
	"context"
	"errors"
	"sync"
)

// ImporterData contains a message for an exporter and channel for communicating back to the respective importer
type ImporterData struct {
	msg                  ConnectionRequest                  // message for exporter
	resultNotificationCh chan ForwardingSuccessNotification // report to importer routine if message was passed to exporter socket successfully @todo should this be bool?
}

func InitImporterData(cr ConnectionRequest) *ImporterData {
	importer := &ImporterData{
		msg:                  cr,
		resultNotificationCh: make(chan ForwardingSuccessNotification),
	}
	return importer
}

// Exporter is created every time a server reaches out to create a persistent connection. Messages are passed to the
// handler maintaining the persistent connection via the channel below
type Exporter struct {
	ctx                    context.Context    // context is used to tell whether server connection is still open
	exporterNotificationCh chan *ImporterData // messages passed to this channel are to be forwarded to the exporter
}

func InitExporter(freshCTX context.Context) *Exporter {
	exporter := &Exporter{
		ctx:                    freshCTX,
		exporterNotificationCh: make(chan *ImporterData),
	}
	return exporter
}

type ExporterDB struct {
	exporters map[string]*Exporter //map to store the exporters
	mx        sync.RWMutex         //RWMutex to protect the map
}

func InitExporterDB() *ExporterDB {
	db := &ExporterDB{
		exporters: make(map[string]*Exporter),
		mx:        sync.RWMutex{},
	}
	return db
}

// AddExporter is called when a new exporter wishes to advertise via the relay
func (db *ExporterDB) AddExporter(id string, exp *Exporter) {
	db.mx.Lock()
	db.exporters[id] = exp
	db.mx.Unlock()
}

// RemoveExporter is used for cleanup of exporters no longer advertising via the relay
func (db *ExporterDB) RemoveExporter(id string) {
	db.mx.Lock()
	delete(db.exporters, id)
	db.mx.Unlock()
}

// NotifyExporter return an error if the server to access does not exist in the db nil otherwise
func (db *ExporterDB) NotifyExporter(id string, msg *ImporterData) error {
	db.mx.RLock()
	defer db.mx.RUnlock()
	if exporter, ok := db.exporters[id]; ok {
		exporter.exporterNotificationCh <- msg
		return nil
	}
	var ErrNotFound = errors.New("Server was not found")
	return ErrNotFound
}
