package relay

import (
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"io"
	"mbg-relay/relayconn/api"
	"net"
	"net/http"
)

// getWaitingImporterId calculates the id of a waiting request based on relevant importer/exporter ids
func getWaitingImporterId(cr api.ConnectionRequest) string {
	return cr.ImporterID + cr.ExporterID
}

// getCallingExporterId calculates the id of a callback response based on relevant importer/exporter ids
func getCallingExporterId(ca api.ConnectionAccept) string {
	return ca.ImporterID + ca.ExporterID
}

// Hijack the HTTP connection and use the TCP session
func hijackConn(w http.ResponseWriter) net.Conn {
	// Check if we can hijack connection
	hj, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "server doesn't support hijacking", http.StatusInternalServerError)
		return nil
	}
	w.WriteHeader(http.StatusOK) //should this be here?
	// Hijack the connection
	conn, _, _ := hj.Hijack()
	return conn
}

// uniteConnections glues an importer connection with an exporter connection
func uniteConnections(importerConn net.Conn, exporterConn net.Conn) error {

	var eg errgroup.Group

	eg.Go(func() error {
		defer importerConn.Close()
		defer exporterConn.Close()

		_, err := io.Copy(exporterConn, importerConn)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("exporter->importer: %w", err)
		}

		return nil
	})

	eg.Go(func() error {
		defer importerConn.Close()
		defer exporterConn.Close()

		_, err := io.Copy(importerConn, exporterConn)
		if err != nil && !errors.Is(err, net.ErrClosed) {
			return fmt.Errorf("importer->exporter: %w", err)
		}

		return nil
	})

	return eg.Wait()
}
