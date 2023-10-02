package client

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"mbg-relay/relayconn/api"
	"mbg-relay/relayconn/utils/httputils"
	"net"
)

// DialTCP dials an exporter via the relay at the given ip
func DialTCP(relayIP string, clientName string, serverName string) (net.Conn, error) {
	logger := logrus.WithField("component", "importingclient")
	logger.Infof("Starting TCP Connect Request to exporter id %v via relay ip %v", serverName, relayIP)
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
