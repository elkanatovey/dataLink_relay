package tcp_endpoints

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/api"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/utils/httputils"
	"net"
)

// RelayDialer connects to a server via a relay
type RelayDialer struct {
	relayIP  string
	clientID string
}

// Dial fulfills the same api as net.Dial but does it's dial via a Relay who's IP is in the backing struct
func (r RelayDialer) Dial(network, address string) (net.Conn, error) {
	if network != "tcp" {
		panic(" only tcp supported")
	}
	return DialTCP(r.relayIP, r.clientID, address)
}

// DialTCP dials a server via the relay at the given ip
func DialTCP(relayIP string, clientName string, serverName string) (net.Conn, error) {
	logger := logrus.WithField("component", "importingclient")
	logger.Infof("Starting TCP Connect Request to server id %v via relay ip %v", serverName, relayIP)
	url := api.TCP + relayIP + api.Dial

	jsonData, err := json.Marshal(api.ConnectionRequest{ClientID: clientName, ServerID: serverName})
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}

	conn, resp := httputils.Connect(relayIP, url, string(jsonData))
	if resp == nil {
		logger.Infof("Successfully Connected")
		return conn, nil
	}

	logger.Errorf("connect Request Failed")
	return nil, fmt.Errorf("connect Request Failed")
}
