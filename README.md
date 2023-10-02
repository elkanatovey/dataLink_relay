# Clusterlink relay implementation
The Clusterlink relay allows two clusterlink gateways to communicate over mtls even when both the gateways are in their private networks behind a firewall.

###Project Workflow
1. Relay starts listening for importers/exporters
2. server registers and maintains persistent connection
3. client requests to connect
4. Relay calls back to server over persistent connection with request for new connection
5. server dials back
6. relay completes connection and starts forwarding

### Demo Usage:

1. From the project root directory run:

    ```go build -o bin/ ./...```   
2. You will find the compiled executables in the bin/ directory

3. Now from 3 separate terminals run in order:

```./relay```

```./server```

```./client```

4. The terminal from client will echo single words back via the terminal

 If you wish to run a basic demo in a single executable run ```./mbg-relay```.  It runs all three entities at once and has a few clients print basic messages at the server.
