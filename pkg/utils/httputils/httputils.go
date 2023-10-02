package httputils

import (
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

type myConn struct {
	net.Conn
	r *http.Response
}

func (a myConn) Close() error {
	a.r.Body.Close()
	return a.Conn.Close()
}

// Connect is a convenience function to issue a CONNECT request
func Connect(address, url string, jsonData string) (net.Conn, error) {
	c, err := dial(address)
	if err != nil {
		return nil, err
	}

	log.Infof("Send Connect request to url: %v", url)
	client := http.Client{Transport: &http.Transport{Dial: connDialer{c}.Dial}}
	req, err := http.NewRequest(http.MethodConnect, url, bytes.NewBuffer([]byte(jsonData)))
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)

	if err != nil {
		log.Errorln(err)
		return nil, err
	}
	mc := myConn{c, resp}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("connect response code: %v", resp.StatusCode)
	}

	return mc, nil

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

type connDialer struct {
	c net.Conn
}

// Dial (network , addr)fakes a connect to an existing connection
func (cd connDialer) Dial(_, _ string) (net.Conn, error) {
	return cd.c, nil
}
