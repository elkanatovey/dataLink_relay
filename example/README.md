# Datalink relay demo
The executables in the below subdirectories run a demo echo server/MTLS echo server.
### General workflow:


1.  compile the respective binaries by running ```go build``` in the respective binaries directory. Then in different terminals:
2. run ```relay```
3. run ```server```
4. run ```client```

### MTLS demo
The mTLS demo needs the throwaway PKI: run ```go run ./example/utils/gencerts``` from the repo root first.

The relay also serves an mTLS control listener on port 3443, which the mTLS server uses for its
registration connection. Start the relay with ```relay -require-control-tls``` to drop registration
from the plaintext listener entirely, so it can only be done with a certificate. The plain TCP demo
above registers in plaintext, so it needs the relay started without that flag.

