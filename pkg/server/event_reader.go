package server

// This code deals with reading ConnectionRequests at the server from a buffered stream

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/sirupsen/logrus"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/api"
	"io"
)

// eventStreamReader scans an io.Reader looking for EventStream messages.
type eventStreamReader struct {
	scanner *bufio.Scanner
	logger  *logrus.Entry
}

// newEventStreamReader creates an instance of eventStreamReader.
func newEventStreamReader(eventStream io.Reader, maxBufferSize int) *eventStreamReader {
	scanner := bufio.NewScanner(eventStream)
	scanner.Buffer(make([]byte, maxBufferSize), maxBufferSize)

	// split returns min all the data available or a single event
	split := func(data []byte, atEOF bool) (int, []byte, error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		// We have a full event payload to parse.
		if i := bytes.Index(data, []byte("\n\n")); i >= 0 {
			return i + 2, data[0:i], nil
		}
		// If we're at EOF, we have all the data.
		if atEOF {
			return len(data), data, nil
		}
		// Request more data.
		return 0, nil, nil
	}
	// Set the split function for the scanning operation.
	scanner.Split(split)

	return &eventStreamReader{
		scanner: scanner,
		logger:  logrus.WithField("component", "eventstreamreader"),
	}
}

// readEvent scans the EventStream for events.
func (e *eventStreamReader) readEvent() (*api.ConnectionRequest, error) {
	if e.scanner.Scan() {
		event := e.scanner.Bytes()
		unmarshalled, err := api.UnmarshalFromSSEEvent(string(event[:]))
		return unmarshalled, err
	}

	if err := e.scanner.Err(); err != nil {
		//client closed connection
		if errors.Is(err, context.Canceled) {
			e.logger.Infof("reader closed due to context cancellation")
			return nil, context.Canceled
		}

		e.logger.Errorln(err, "reader closed unexpectedly")
		// general error
		return nil, err
	}

	//server closed connection
	e.logger.Infof("reader closed due to server closing connection")
	return nil, io.EOF
}
