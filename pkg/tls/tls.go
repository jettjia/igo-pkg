package tls

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"time"
)

// GenerateKeyTls generate rsa private keys and return content and error messages in pem format
func GenerateKeyTls() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", err
	}

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	var privateKeyPemContent bytes.Buffer
	err = pem.Encode(&privateKeyPemContent, block)
	if err != nil {
		return "", err
	}

	return privateKeyPemContent.String(), nil
}

// GenerateCert generate a self signed certificate from the private key string and return the certificate content and error information in pem format
func GenerateCertTls(privateKeyStr string, commonName string, org []string) (string, error) {
	privateKey, err := parsePrivateKeyFromString(privateKeyStr)
	if err != nil {
		return "", err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			CommonName:   commonName,
			Organization: org,
		},
		NotBefore: time.Now(),
		NotAfter:  time.Now().AddDate(1, 0, 0),

		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,

		DNSNames: []string{"x-data.tech", "www.x-data.tech"},
	}

	// generate a public key
	publicKey := &privateKey.PublicKey

	// generate a certificate
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, publicKey, privateKey)
	if err != nil {
		return "", err
	}

	certBlock := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: derBytes,
	}

	var certPemContent bytes.Buffer
	err = pem.Encode(&certPemContent, certBlock)
	if err != nil {
		return "", err
	}

	return certPemContent.String(), nil
}

// parsePrivateKeyFromString
func parsePrivateKeyFromString(privateKeyStr string) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode([]byte(privateKeyStr))
	if block == nil {
		return nil, errors.New("failed to decode PEM block containing the key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return privateKey, nil
}
