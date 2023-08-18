# Multi-cloud Border Gateway(MBG) relay implementation
The mcb relay allows importers and exporters of services to communicate over mtls even when both the expoerter and importer of a service are behind a firewall.

### General workflow:


1. Relay starts listening for importers/exporters
2. exporter registers and maintains persistent connection
3. importer requests to connect
4. Relay calls back to exporter over persistent connection with request for new connection
5. exporter dials back
6. relay completes connection and starts forwarding
