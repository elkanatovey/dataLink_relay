# Datalink Relay implementation
datalink_relay is a library written in go for the purpose of allowing servers behind a firewall to listen for connections on an untrusted relay server. The library exports the [net.Listener](https://pkg.go.dev/net#Listener) and [net.Dialer](https://pkg.go.dev/net#Dialer) interfaces for convenience of use, for servers and clients.

## Workflow
1. Relay starts listening for connection/listen requests
2. Server registers a listen request with Relay and maintains persistent connection
3. Client registers connect request at Relay and waits on request
4. Relay forwards Client's connection request to Server over persistent connection
5. Server dials back to Relay
6. Relay completes connection and starts forwarding data

**Note that in current implementation MTLS support is only enabled for connection at step 6, i.e. the connection request and persistent connection send messages in the clear over http**

## Demo Usage:

1. From the project root directory run:

    ```go build -o bin/ ./...```   

2. The compiled executables will be in the bin/ directory

3. Now from 3 separate terminals run in order:

   ```sh
   ./relay
   ./server
   ./client
   ```

4. The client will echo single words back via the terminal

To run a basic demo in a single executable run ```./all```.  It runs all three entities at once and has a few clients print basic messages at the server. Instructions for the MTLS versions are similar. Note that all demos run on localhost with hardcoded values that can be found [here](example/utils).

Documentation of the public facing API can be found [here](docs/DOCUMENTATION.md). 
