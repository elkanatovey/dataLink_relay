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
//@todo support mtls

package mtls_endpoint

import (
	"crypto/tls"
	"errors"
	"github.ibm.com/mcnet-research/mbg_relay/pkg/tcp_endpoints"
	"net"
)

type MTLSRelayListener struct {
	tcp_endpoints.RelayListener
	config *tls.Config
}

func (l MTLSRelayListener) Accept() (net.Conn, error) {
	c, err := l.RelayListener.Accept()
	if err != nil {
		return nil, err
	}
	return tls.Server(c, l.config), nil
}

func Listen(relayURL string, listenerID string, config *tls.Config) (net.Listener, error) {
	if config == nil || len(config.Certificates) == 0 &&
		config.GetCertificate == nil && config.GetConfigForClient == nil {
		return nil, errors.New("tls: neither Certificates, GetCertificate, nor GetConfigForClient set in Config")
	}
	l, err := tcp_endpoints.Listen(relayURL, listenerID)
	if err != nil {
		return nil, err
	}
	return tls.NewListener(l, config), nil
}

func NewMTLSRelayListener(inner tcp_endpoints.RelayListener, config *tls.Config) net.Listener {
	return tls.NewListener(inner, config)
}
