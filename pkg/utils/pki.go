package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"
)

func GenerateCertificate(id int64, pub, priv interface{}, years, months, days int) (cert *x509.Certificate, certPem []byte, err error) {
	template := &x509.Certificate{
		SerialNumber: big.NewInt(id),
		Subject: pkix.Name{
			OrganizationalUnit: []string{"dockit"},
			Organization:       []string{"ekristen"},
			Country:            []string{"US"},
			Province:           []string{"dev"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(years, months, days),
		IsCA:                  false,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	certDer, err := x509.CreateCertificate(
		rand.Reader, template, template, pub, priv,
	)
	if err != nil {
		return nil, nil, err
	}

	certBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	}

	buf := bytes.NewBuffer(certPem)

	if err := pem.Encode(buf, &certBlock); err != nil {
		return nil, nil, err
	}

	return template, buf.Bytes(), err
}

func GenerateECKey(curveSize int) (key *ecdsa.PrivateKey, pemData []byte, err error) {
	var curve elliptic.Curve

	switch curveSize {
	case 224:
		curve = elliptic.P224()
	case 256:
		curve = elliptic.P256()
	case 384:
		curve = elliptic.P384()
	default:
		return nil, nil, fmt.Errorf("unsupported curve size: %d", curveSize)
	}

	key, err = ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		return nil, nil, err
	}

	keyDer, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return nil, nil, err
	}

	keyBlock := pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: keyDer,
	}

	buf := bytes.NewBuffer(pemData)

	if err := pem.Encode(buf, &keyBlock); err != nil {
		return nil, nil, err
	}

	return key, buf.Bytes(), nil
}

func GenerateRSAKey(bits int) (key *rsa.PrivateKey, pemData []byte, err error) {
	if bits != 2048 && bits != 4096 && bits != 3072 && bits != 7680 {
		return nil, nil, fmt.Errorf("unsupported bit size: %d", bits)
	}

	key, err = rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, nil, err
	}

	keyDer := x509.MarshalPKCS1PrivateKey(key)

	keyBlock := pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: keyDer,
	}

	buf := bytes.NewBuffer(pemData)

	if err := pem.Encode(buf, &keyBlock); err != nil {
		return nil, nil, err
	}

	return key, buf.Bytes(), nil
}
