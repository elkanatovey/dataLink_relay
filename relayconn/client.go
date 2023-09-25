package relayconn

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

// ImportingClient imports a service
// server exports services via relay
type ImportingClient struct {
	Connection    *http.Client
	RelayURL      string // address of relay
	ImporterID    string
	maxBufferSize int
}

// NewImportingClient creates a new ImportingClient
func NewImportingClient(url string, id string, opts ...func(c *ImportingClient)) *ImportingClient {
	s := &ImportingClient{
		RelayURL:      url,
		ImporterID:    id,
		Connection:    &http.Client{},
		maxBufferSize: 1 << 16,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

// createConnectRequest builds the request to open the connection with the server
func (c *ImportingClient) createConnectRequest(ctx context.Context, exporterName string) (*http.Request, error) {
	reqBody := ConnectionRequest{ImporterID: c.ImporterID, ExporterID: exporterName}
	reqBodyBytes, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", c.RelayURL+Dial, bytes.NewReader(reqBodyBytes)) //@todo should we cancel context in case of error?
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return req, nil
}

//// TCP connection request to other peer
//func (c *ImportingClient) TCPConnectReq(svcID, svcIDDest, svcPolicy, mbgIP string) (net.Conn, error) {
//	clog.Printf("Starting TCP Connect Request to peer at %v for service %v", mbgIP, svcIDDest)
//	url := d.Store.GetProtocolPrefix() + mbgIP + "/exports/serviceEndpoint"
//
//	jsonData, err := json.Marshal(apiObject.ConnectRequest{ID: svcID, IDDest: svcIDDest, Policy: svcPolicy, MbgID: d.Store.GetMyID()})
//	if err != nil {
//		clog.Error(err)
//		return nil, err
//	}
//
//	c, resp := httputils.Connect(mbgIP, url, string(jsonData))
//	if resp == nil {
//		clog.Printf("Successfully Connected")
//		return c, nil
//	}
//
//	return nil, fmt.Errorf("connect Request Failed")
//}
