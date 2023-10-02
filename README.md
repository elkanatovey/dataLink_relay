# Clusterlink relay implementation
The Clusterlink relay allows two clusterlink gateways to communicate over mtls even when both the gateways are in their private networks behind a firewall.

### General workflow:


1. Relay starts listening for importers/exporters
2. server registers and maintains persistent connection
3. client requests to connect
4. Relay calls back to server over persistent connection with request for new connection
5. server dials back
6. relay completes connection and starts forwarding
