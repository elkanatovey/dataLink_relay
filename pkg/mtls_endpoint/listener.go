/**********************************************************/
/* mtls_listener : This listener accepts a connection with a client after only if the client verifies with mtls.
/* The initial listen is not necessarily encrypted
*/
/**********************************************************/
// Workflow of listener usage
// After relay starts up
//    1) Connect to relay (possibly unencrypted to be fixed)
//    2) Client wishing to connect connects via  a call to HandleClientConnection
//    3) Relay initiates a callback via persistent connection
//    4) listener calls back, sockets are connected at relay proceed with mtls handshake

package mtls_endpoint

import (
	"crypto/tls"
	"errors"
	"github.com/elkanatovey/dataLink_relay/pkg/tcp_endpoints"
	"net"
)

// MTLSRelayListener is a backing struct for MTLSRelayListener.Listen to maintain the tls.Listener api
type MTLSRelayListener struct {
	relayURL string
	opts     []Option
}

func (r MTLSRelayListener) Listen(network, address string, config *tls.Config) (net.Listener, error) {
	if config == nil || len(config.Certificates) == 0 &&
		config.GetCertificate == nil && config.GetConfigForClient == nil {
		return nil, errors.New("tls: neither Certificates, GetCertificate, nor GetConfigForClient set in Config")
	}
	l := tcp_endpoints.NewRelayListener(r.relayURL, r.opts...)
	listener, err := l.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(listener, config), nil

}

// NewMTLSRelayListener creates an MTLSRelayListener that maintains the tls.Listener api. To run it
// call MTLSRelayListener.Listen. relayURL is the address of the relay via which we listen
func NewMTLSRelayListener(relayURL string, opts ...Option) MTLSRelayListener {
	return MTLSRelayListener{relayURL: relayURL, opts: opts}
}

// ListenMTLS is a version of Listen without the backing struct
func ListenMTLS(network, address string, config *tls.Config, relayURL string, opts ...Option) (net.Listener, error) {
	return NewMTLSRelayListener(relayURL, opts...).Listen(network, address, config)
}
