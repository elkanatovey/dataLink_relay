package server

// This code deals with reading ConnectionRequests at the server from a buffered stream

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"mbg-relay/relayconn/api"
)

// EventStreamReader scans an io.Reader looking for EventStream messages.
type EventStreamReader struct {
	scanner *bufio.Scanner
}

// NewEventStreamReader creates an instance of EventStreamReader.
func NewEventStreamReader(eventStream io.Reader, maxBufferSize int) *EventStreamReader {
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

	return &EventStreamReader{
		scanner: scanner,
	}
}

// ReadEvent scans the EventStream for events.
func (e *EventStreamReader) ReadEvent() (*api.ConnectionRequest, error) {
	if e.scanner.Scan() {
		event := e.scanner.Bytes()
		unmarshalled, err := api.UnmarshalFromSSEEvent(string(event[:]))
		return unmarshalled, err
	}
	if err := e.scanner.Err(); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil, io.EOF
		}
		return nil, err
	}
	return nil, io.EOF
}
