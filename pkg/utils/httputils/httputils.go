package httputils

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"
	//"time"

	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("component", "httpHandler")

const (
	RESPOK   string = "Success"
	RESPFAIL string = "Fail"
)

// Get is a convenience function to issue a GET request
func Get(url string, cl http.Client) ([]byte, error) {
	resp, err := cl.Get(url)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return []byte(RESPFAIL), err
	}
	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(RESPFAIL), err
	}
	// Convert the body to type string
	return body, nil
}

// Post is a convenience function to issue a POST request
func Post(url string, jsonData []byte, cl http.Client) ([]byte, error) {
	resp, err := cl.Post(url, "application/json",
		bytes.NewBuffer(jsonData))

	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return []byte(RESPFAIL), err
	}

	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(RESPFAIL), err
	}

	return body, nil
}

// Delete is a convenience function to issue a DELETE request
func Delete(url string, jsonData []byte, cl http.Client) ([]byte, error) {
	req, err := http.NewRequest(http.MethodDelete, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return []byte(RESPFAIL), err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	resp, err := cl.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return []byte(RESPFAIL), err
	}

	// We Read the response body on the line below.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []byte(RESPFAIL), err
	}

	return body, nil
}

// tunnelConn is the connection handed back once a CONNECT has been accepted. Reads go through the
// buffered reader that parsed the response, so any bytes it read ahead of the response are still
// delivered to the caller. Everything else is the raw connection.
type tunnelConn struct {
	net.Conn
	reader io.Reader
}

func (c *tunnelConn) Read(p []byte) (int, error) {
	return c.reader.Read(p)
}

// Connect is a convenience function to issue a CONNECT request
func Connect(address, url string, jsonData string) (net.Conn, error) {
	c, err := dial(address)
	if err != nil {
		return nil, err
	}

	log.Infof("Send Connect request to url: %v", url)
	req, err := http.NewRequest(http.MethodConnect, url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		c.Close()
		return nil, err
	}

	// Write the request straight to the connection rather than going through an http.Transport. A
	// Transport keeps ownership of the socket and reads it in the background, so tunnel bytes that
	// arrive after the response can end up in its buffer where the caller can never see them.
	if err := req.Write(c); err != nil {
		log.Errorln(err)
		c.Close()
		return nil, err
	}

	br := bufio.NewReader(c)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		log.Errorln(err)
		c.Close()
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		c.Close()
		return nil, fmt.Errorf("connect response code: %v", resp.StatusCode)
	}

	return &tunnelConn{Conn: c, reader: br}, nil
}

func dial(addr string) (net.Conn, error) {
	log.Infof("Start dial to address: %v\n", addr)
	c, err := net.Dial("tcp", addr)

	if err != nil {
		return nil, err
	}
	log.Infof("Finish dial to address: %v\n", addr)

	return c, err
}
