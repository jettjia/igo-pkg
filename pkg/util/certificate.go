package util

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/sirupsen/logrus"
)

// ParseCertificates parse the content of the certificate
func ParseCertificates(content []byte) (*x509.Certificate, error) {
	var err error
	if len(content) == 0 {
		err = fmt.Errorf("certificate content is empty")
		return nil, err
	}

	// directly parse certificates in pkcs7 format
	p7Block, rest := pem.Decode(content)
	if p7Block == nil {
		err = fmt.Errorf("failed to decode PKCS7 block")
		return nil, err
	}
	if len(rest) > 0 {
		logrus.Warnf("Extra data found after PKCS7 block")
	}

	var certs []*x509.Certificate
	if certs, err = x509.ParseCertificates(p7Block.Bytes); err != nil {
		return nil, err
	}
	if len(certs) == 0 {
		err = fmt.Errorf("no certificates found in the PKCS7 block")
		return nil, err
	}

	return certs[0], nil
}
