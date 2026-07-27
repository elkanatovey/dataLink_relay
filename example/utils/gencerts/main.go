// Command gencerts generates throwaway demo PKIs for the mTLS example: one for the end-to-end
// client/server session, and a separate one for the relay control plane. The generated material is
// for local demos only. Run it from the repo root:
//
//	go run ./example/utils/gencerts
package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"github.com/elkanatovey/dataLink_relay/pkg/api"
	"log"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"
)

const outDir = "example/utils/certs"

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Printf("wrote demo PKI to %s", outDir)
}

func run() error {
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return err
	}

	// End-to-end PKI: authenticates the client and the server to each other. The relay is not part
	// of this trust domain and never holds a credential from it.
	caCert, caKey, err := newCA("datalink-relay demo CA", "ca-cert.pem", "ca-key.pem")
	if err != nil {
		return err
	}

	// server leaf: SAN must contain the name the client sets as tls.Config.ServerName
	if err := writeLeaf("server", caCert, caKey, x509.ExtKeyUsageServerAuth,
		[]string{"foo", "localhost"}, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return err
	}
	// client leaf: used for mTLS client authentication
	if err := writeLeaf("client", caCert, caKey, x509.ExtKeyUsageClientAuth, []string{"bar"}, nil); err != nil {
		return err
	}

	// Control-plane PKI: authenticates a listening server to the relay. Deliberately a separate CA,
	// so a client certificate from the end-to-end PKI cannot be used to register a server id.
	ctrlCert, ctrlKey, err := newCA("datalink-relay demo control CA", "control-ca-cert.pem", "control-ca-key.pem")
	if err != nil {
		return err
	}

	// relay leaf: presented by the relay on its mTLS control listener
	if err := writeLeaf("relay-control", ctrlCert, ctrlKey, x509.ExtKeyUsageServerAuth,
		[]string{"localhost"}, []net.IP{net.ParseIP("127.0.0.1")}); err != nil {
		return err
	}
	// exporter leaf: SAN must contain the server id being registered, which the relay enforces
	if err := writeLeaf("exporter-control", ctrlCert, ctrlKey, x509.ExtKeyUsageClientAuth,
		[]string{"foo"}, nil); err != nil {
		return err
	}

	// relay X25519 keypair used to seal routing metadata to the relay
	kp, err := api.GenerateRelayKeyPair()
	if err != nil {
		return err
	}
	priv := kp.PrivateKey()
	if err := os.WriteFile(filepath.Join(outDir, "relay.key"), priv[:], 0o600); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(outDir, "relay.pub"), kp.Public[:], 0o644)
}

// newCA generates a self signed CA and writes its certificate and key.
func newCA(commonName, certFile, keyFile string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return nil, nil, err
	}
	tmpl := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: commonName},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		return nil, nil, err
	}
	cert, err := x509.ParseCertificate(der)
	if err != nil {
		return nil, nil, err
	}
	if err := writeCert(certFile, der); err != nil {
		return nil, nil, err
	}
	if err := writeKey(keyFile, key); err != nil {
		return nil, nil, err
	}
	return cert, key, nil
}

func writeLeaf(name string, caCert *x509.Certificate, caKey *ecdsa.PrivateKey, eku x509.ExtKeyUsage, dns []string, ips []net.IP) error {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return err
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		return err
	}
	tmpl := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: name},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{eku},
		DNSNames:     dns,
		IPAddresses:  ips,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, caCert, &key.PublicKey, caKey)
	if err != nil {
		return err
	}
	if err := writeCert(name+"-cert.pem", der); err != nil {
		return err
	}
	return writeKey(name+"-key.pem", key)
}

func writeCert(file string, der []byte) error {
	return writePEM(file, &pem.Block{Type: "CERTIFICATE", Bytes: der})
}

func writeKey(file string, key *ecdsa.PrivateKey) error {
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	return writePEM(file, &pem.Block{Type: "EC PRIVATE KEY", Bytes: der})
}

func writePEM(file string, block *pem.Block) error {
	f, err := os.Create(filepath.Join(outDir, file))
	if err != nil {
		return err
	}
	defer f.Close()
	return pem.Encode(f, block)
}
