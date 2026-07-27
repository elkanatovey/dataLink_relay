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
	"fmt"
	"github.com/elkanatovey/dataLink_relay/pkg/api"
	"github.com/elkanatovey/dataLink_relay/pkg/utils/httputils"
	"github.com/sirupsen/logrus"
	"net"
)

// RelayDialer connects to a server via a relay
type RelayDialer struct {
	relayIP  string
	clientID string
	relayPub *[32]byte
}

// Dial fulfills the same api as net.Dial but does it's dial via a Relay who's IP is in the backing struct
func (r RelayDialer) Dial(network, address string) (net.Conn, error) {
	if network != "tcp" {
		return nil, fmt.Errorf("only tcp supported, got %q", network)
	}

	logger := logrus.WithField("component", "importingclient")
	logger.Infof("Starting TCP Connect Request to server id %v via relay ip %v", address, r.relayIP)
	url := api.TCP + r.relayIP + api.Dial

	jsonData, err := api.EncodeRouting(api.ConnectionRequest{ClientID: r.clientID, ServerID: address}, r.relayPub)
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

// NewRelayDialer creates a RelayDialer that implements the net.Dialer api. To dial call
// RelayDialer.Dial. relayIP is the address of the relay via which we dial, clientName is the id
// this client presents to the relay
func NewRelayDialer(relayIP string, clientName string, opts ...Option) RelayDialer {
	o := applyOptions(opts)
	return RelayDialer{relayIP: relayIP, clientID: clientName, relayPub: o.relayPub}
}

// DialTCP dials a server via the relay at the given ip via RelayDialer.Dial
func DialTCP(network, address string, relayIP string, clientName string, opts ...Option) (net.Conn, error) {
	return NewRelayDialer(relayIP, clientName, opts...).Dial(network, address)
}
