# Clusterlink relay demo
The executables in this file run a demo where the client writes a message to the server via the relay
### General workflow:


1.  compile the respective binaries by running ```go build``` in the respective binaries directory
2. run ```relay```
3. run ```server```
4. run ```client```

### Current bug:

if either ```server``` or ```client``` is compiled with lines 100-102 of file
```relayconn/utils/httputils/httputils.go``` not commented out read/write calls to the returned socket fail with 
```Error sending message: write tcp 127.0.0.1:58658->127.0.0.1:3333: use of closed network connection```

For debugging, I recommend starting up the relay, and then compiling and starting he server with the relevant lines commented out and only debugging the client.

The server will accept multiple client requests in that setting

