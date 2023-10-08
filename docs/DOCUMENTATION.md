# Documentation for clusterlink_relay
## Introduction


clusterlink_relay is a library written in go for the purpose of allowing servers behind a firewall to listen for connections on an untrusted relay server. 
The library exports the net Listener and Dialer interfaces for convenience of use, for servers and clients.

## Features
* TCP support
* MTLS support
* Standard Net/TLS Listener and Dialer interface support


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

## Implementation Details
A `tcp_endpoints.RelayListener` works by receiving connection requests via [SSE](https://en.wikipedia.org/wiki/Server-sent_events#:~:text=Server%2DSent%20Events%20(SSE),client%20connection%20has%20been%20established.) received over a persistent http connection with the relay.
Connection requests are accepted by dialing back to the relay where the relay matches the callback with the original connection request. The relay exposes an http api,
while the client and server use the underlying socket to communicate.

**Important:** even in MTLS mode communication with the relay is **not** encrypted. The MTLS handshake is only between the client and server.