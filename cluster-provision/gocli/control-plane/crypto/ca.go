package crypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math"
	"math/big"
	"time"

	"k8s.io/client-go/util/cert"
	"k8s.io/client-go/util/keyutil"
)

func WriteKeyAndCertToFile(certificate *x509.Certificate, key crypto.Signer, certPath, keyPath string) error {
	encoded, err := keyutil.MarshalPrivateKeyToPEM(key)
	if err != nil {
		return err
	}
	if err := keyutil.WriteKey(keyPath, encoded); err != nil {
		return err
	}

	pemBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certificate.Raw,
	}
	certEncoded := pem.EncodeToMemory(&pemBlock)

	return cert.WriteCert(certPath, certEncoded)
}

func GenerateKubernetesCAKeyPair() (*x509.Certificate, crypto.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	cert, err := cert.NewSelfSignedCACert(cert.Config{CommonName: "ca", Organization: []string{""}}, key)
	if err != nil {
		return nil, nil, err
	}

	return cert, key, nil
}

func GenerateCertKeyPairWithCA(cfg cert.Config, ca *x509.Certificate, caKey crypto.Signer) (*x509.Certificate, crypto.Signer, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)

	serial, err := rand.Int(rand.Reader, new(big.Int).SetInt64(math.MaxInt64-1))
	if err != nil {
		return nil, nil, err
	}
	serial = new(big.Int).Add(serial, big.NewInt(1))
	if len(cfg.CommonName) == 0 {
		return nil, nil, err
	}

	keyUsage := x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature

	notAfter := time.Now().Add(time.Hour * 24 * 365).UTC()

	certTmpl := x509.Certificate{
		Subject: pkix.Name{
			CommonName:   cfg.CommonName,
			Organization: cfg.Organization,
		},
		DNSNames:              cfg.AltNames.DNSNames,
		IPAddresses:           cfg.AltNames.IPs,
		SerialNumber:          serial,
		NotBefore:             ca.NotBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           cfg.Usages,
		BasicConstraintsValid: true,
		IsCA:                  false,
	}
	certDERBytes, err := x509.CreateCertificate(rand.Reader, &certTmpl, ca, key.Public(), caKey)
	if err != nil {
		return nil, nil, err
	}
	finalCert, err := x509.ParseCertificate(certDERBytes)
	return finalCert, key, err
}
