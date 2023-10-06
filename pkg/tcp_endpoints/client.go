/**********************************************************/
/* client : This dials services behind firewalls that can only service outgoing connections via a relay.
/**********************************************************/
// Workflow of client usage
// After relay starts up
//    1) Initialise dialer
//    2) Call dial
//    alternatively call DialTCP directly

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

	logger := logrus.WithField("component", "importingclient")
	logger.Infof("Starting TCP Connect Request to server id %v via relay ip %v", address, r.relayIP)
	url := api.TCP + r.relayIP + api.Dial

	jsonData, err := json.Marshal(api.ConnectionRequest{ClientID: r.clientID, ServerID: address})
	if err != nil {
		logger.Errorln(err)
		return nil, err
	}

	conn, resp := httputils.Connect(r.relayIP, url, string(jsonData))
	if resp == nil {
		logger.Infof("Successfully Connected")
		return conn, nil
	}
	logger.Errorf("connect Request Failed")
	return nil, fmt.Errorf("connect Request Failed")
}

// DialTCP dials a server via the relay at the given ip via RelayDialer.Dial
func DialTCP(network, address string, relayIP string, clientName string) (net.Conn, error) {
	return RelayDialer{relayIP: relayIP, clientID: clientName}.Dial(network, address)
}
