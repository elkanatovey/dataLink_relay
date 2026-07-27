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
| `tcp_endpoints.DialTCP(network, address string, relayIP string, clientName string, opts ...Option) (net.Conn, error)` | dials server listening on relay via args                                |
| `tcp_endpoints.NewRelayDialer(relayIP string, clientName string, opts ...Option) RelayDialer`         | create `RelayDialer`. This implements the `net.Dialer` api               |
| `RelayDialer.Dial(network, address string) (net.Conn, error)`                                         | dials server listening on relay. Dialer initialised with server address |

#### Server connection

| Method                                                                                      | Description                                                                                                       |
|---------------------------------------------------------------------------------------------|-------------------------------------------------------------------------------------------------------------------|
| `tcp_endpoints.ListenRelay(network, address string, relayURL string, opts ...Option) (net.Listener, error)` | listen on relay via args                                                                                          |
| `tcp_endpoints.NewRelayListener(relayURL string, opts ...Option) RelayListener`             | create `RelayListener`. This implements the ``Listener interface                                                  |
| `tcp_endpoints.RelayListener.Listen(network, address string) (net.Listener, error)`         | Listen on `RelayListener`. The returned listener listens on the relay and implements the `net.Listener` interface |

#### Relay operations
The relay is run by calling the `net.http` library's `ListenAndServe` methods

### MTLS Methods

#### Client connection

| Method                                                                                                                       | Description                                                             |
|------------------------------------------------------------------------------------------------------------------------------|-------------------------------------------------------------------------|
| `mtls_endpoint.DialMTLS(network, address string, config *tls.Config, relayIP string, clientName string, opts ...Option) (net.Conn, error)` | dials server listening on relay via args                                |
| `mtls_endpoint.NewRelayMTLSDialer(relayIP string, clientName string, opts ...Option) RelayMTLSDialer`                        | create `RelayMTLSDialer`. This maintains the `tls.Dial` api              |
| `RelayMTLSDialer.Dial(network, address string, config *tls.Config) (net.Conn, error)`                                        | dials server listening on relay. Dialer initialised with server address |

#### Server connection

| Method                                                                                                         | Description                                                                                                                                                                      |
|----------------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `mtls_endpoint.ListenMTLS(network, address string, config *tls.Config, relayURL string, opts ...Option) (net.Listener, error)` | listen on relay via args                                                                                                                                                         |
| `mtls_endpoint.NewMTLSRelayListener(relayURL string, opts ...Option) MTLSRelayListener`                        | create `MTLSRelayListener`. This maintains the `tls.Listener` api                                                                                                                |
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
* Give the dialer/listener the relay's **public** key via the `WithRelayKey(pub)` option, e.g. `tcp_endpoints.DialTCP("tcp", server, relayIP, client, tcp_endpoints.WithRelayKey(pub))` or `mtls_endpoint.ListenMTLS(..., mtls_endpoint.WithRelayKey(pub))`. Both packages expose the option and it is the same type, so either spelling works.

Without the option, routing metadata is sent in plaintext exactly as before. Sealing to a known static key needs no handshake and adds no round trips; it does **not** provide forward secrecy, so rotate the key to bound exposure.

## mTLS on the registration connection
A listening server holds one long-lived connection to the relay, over which the relay pushes connection requests. By default that connection is plain HTTP: the stream is readable and injectable by anyone on-path, and the relay cannot tell who registered a given server id.

It can optionally be moved to an **mTLS control endpoint** on the relay. Only this connection is affected. The client dial and server callback hops get hijacked into the end-to-end encrypted tunnel, so putting TLS on them would nest encryption inside encryption and add a handshake to every dial.

Relay side — serve the muxes on separate listeners:

| Handler | Routes | Listener |
|---|---|---|
| `Relay.Mux` | all three | plaintext; the historical single-listener deployment |
| `Relay.DataMux` | `/clientconn`, `/servercallback` | plaintext |
| `Relay.ControlMux` | `/serverconn` | TLS with `tls.RequireAndVerifyClientCert` |

Serving `DataMux` instead of `Mux` is what makes the control plane mandatory: registration is then not routable at all without a certificate. `ControlMux` additionally refuses any request that did not arrive over TLS with a verified peer certificate, so it fails closed if it is misconfigured onto a plaintext listener.

Listener side — pass the option:

```go
listener, err := mtls_endpoint.ListenMTLS("tcp", "foo", tlsConfig, relayAddr,
    mtls_endpoint.WithRelayControlTLS(controlAddr, controlTLS))
```

The relay checks that the server id being registered is covered by a SAN in the presented certificate, so a certificate cannot be used to claim someone else's id.

**Use a separate PKI for this.** The `tls.Config` passed to `Listen` authenticates the server to a *client*; the one passed to `WithRelayControlTLS` authenticates it to the *relay*. If both trust the same CA then every client certificate that CA issued is also a valid registration certificate, and any client can register as your server — the exact attack this is meant to prevent. The demo mints two CAs for that reason. Reusing a single leaf does not work in any case: a `serverAuth` certificate is rejected when presented as a client certificate.

Without the option the registration connection stays plaintext, exactly as before.