# Documentation for datalink_relay
## Introduction


datalink_relay is a library written in go for the purpose of allowing servers behind a firewall to listen for connections on an untrusted relay server. 
The library exports the net Listener and Dialer interfaces for convenience of use, for servers and clients.

## Features
* TCP support
* MTLS support
* Standard Net/TLS Listener and Dialer interface support
* Optional sealing of routing metadata to the relay's key


## API Calls

### TCP Methods

#### Client connection

| Method                                                                                                | Description                                                             |
|-------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `tcp_endpoints.DialTCP(network, address string, relayIP string, clientName string) (net.Conn, error)` | dials server listening on relay via args                                |
| `RelayDialer.Dial(network, address string) (net.Conn, error)`                                         | dials server listening on relay. Dialer initialised with server address |

#### Server connection

| Method                                                                                      | Description                                                                                                       |
|---------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| `tcp_endpoints.ListenRelay(network, address string, relayURL string) (net.Listener, error)` | listen on relay via args                                                                                          |
| `tcp_endpoints.NewRelayListener(relayURL string) RelayListener`                             | create `RelayListener`. This implements the ``Listener interface                                                  |
| `tcp_endpoints.RelayListener.Listen(network, address string) (net.Listener, error)`         | Listen on `RelayListener`. The returned listener listens on the relay and implements the `net.Listener` interface |

#### Relay operations
The relay is run by calling the `net.http` library's `ListenAndServe` methods

### MTLS Methods

#### Client connection

| Method                                                                                                                       | Description                                                             |
|------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `mtls_endpoint.DialMTLS(network, address string, , config *tls.Config, relayIP string, clientName string) (net.Conn, error)` | dials server listening on relay via args                                |
| `RelayMTLSDialer.Dial(network, address string, config *tls.Config) (net.Conn, error)`                                        | dials server listening on relay. Dialer initialised with server address |

#### Server connection

| Method                                                                                                         | Description                                                                                                                                                                      |
|----------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `mtls_endpoint.ListenMTLS(network, address string, config *tls.Config, relayURL string) (net.Listener, error)` | listen on relay via args                                                                                                                                                         |
| `mtls_endpoint.MTLSRelayListener.Listen(network, address string, config *tls.Config) (net.Listener, error)`    | Listen on `MTLSRelayListener`. The returned listener listens on the relay and implements the `net.Listener` interface. `MTLSRelayListener`must be initialised with the relay url |


## Implementation Details
A `tcp_endpoints.RelayListener` works by receiving connection requests via [SSE](https://en.wikipedia.org/wiki/Server-sent_events#:~:text=Server%2DSent%20Events%20(SSE),client%20connection%20has%20been%20established.) received over a persistent http connection with the relay.
Connection requests are accepted by dialing back to the relay where the relay matches the callback with the original connection request. The relay exposes an http api,
while the client and server use the underlying socket to communicate.

**Important:** the MTLS handshake and data are end-to-end between the client and server; the relay only ever sees ciphertext. The control messages (the connection/listen/callback requests) are, by default, sent to the relay in plaintext HTTP.

## Sealing routing metadata
The routing metadata carried in the control messages — the client and server IDs — can optionally be **sealed to the relay's public key** (X25519 / `nacl` anonymous sealed box) so that on-path network observers cannot see who is connecting to whom. The relay unseals it with its private key in order to route, so this hides the metadata from the *network*, not from the relay itself.

Usage:
* Start the relay with one or more keys: `relay.NewRelay(keyPair)`. Rotate at runtime with `Relay.SetRoutingKeys(...)`; the relay trial-decrypts across its keyring, so rotation does not drop in-flight clients.
* Give the dialer/listener the relay's **public** key via the `tcp_endpoints.WithRelayKey(pub)` option, e.g. `tcp_endpoints.DialTCP("tcp", server, relayIP, client, tcp_endpoints.WithRelayKey(pub))` or `mtls_endpoint.ListenMTLS(..., tcp_endpoints.WithRelayKey(pub))`.

Without the option, routing metadata is sent in plaintext exactly as before. Sealing to a known static key needs no handshake and adds no round trips; it does **not** provide forward secrecy, so rotate the key to bound exposure.