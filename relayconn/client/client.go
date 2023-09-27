package client

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"mbg-relay/relayconn/api"
	"mbg-relay/relayconn/utils/httputils"
	"net"
	"net/http"
)

// ImportingClient imports a service
// server exports services via relay
type ImportingClient struct {
	Connection    *http.Client
	RelayURL      string // address of relay + port
	ImporterID    string
	maxBufferSize int
	logger        *logrus.Entry
}

// NewImportingClient creates a new ImportingClient
func NewImportingClient(url string, id string, opts ...func(c *ImportingClient)) *ImportingClient {
	s := &ImportingClient{
		RelayURL:      url,
		ImporterID:    id,
		Connection:    &http.Client{},
		maxBufferSize: 1 << 16,
		logger:        logrus.WithField("component", "importingclient"),
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// TCPConnectReq dials an exporter via the relay at the given ip
func (c *ImportingClient) TCPConnectReq(relayIP string, exporterName string) (net.Conn, error) {
	c.logger.Infof("Starting TCP Connect Request to exporter id %v via relay ip %v", exporterName, relayIP)
	url := api.TCP + relayIP + api.Dial

	jsonData, err := json.Marshal(api.ConnectionRequest{ImporterID: c.ImporterID, ExporterID: exporterName})
	if err != nil {
		c.logger.Errorln(err)
		return nil, err
	}

	conn, resp := httputils.Connect(relayIP, url, string(jsonData))
	if resp == nil {
		c.logger.Infof("Successfully Connected")
		return conn, nil
	}

	c.logger.Errorf("connect Request Failed")
	return nil, fmt.Errorf("connect Request Failed")
}
