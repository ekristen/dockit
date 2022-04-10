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
	"math/big"
	"time"

	"github.com/bwmarrin/snowflake"
)

func GenerateCertificate(pub, priv interface{}, years, months, days int) (certPem []byte, err error) {
	node, err := snowflake.NewNode(1)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{
		SerialNumber: big.NewInt(node.Generate().Int64()),
		Subject: pkix.Name{
			Organization: []string{"ekristen"},
			Country:      []string{"US"},
			Province:     []string{"dev"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(years, months, days),
		IsCA:                  false,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	certDer, err := x509.CreateCertificate(
		rand.Reader, &template, &template, pub, priv,
	)
	if err != nil {
		return nil, err
	}

	certBlock := pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certDer,
	}

	buf := bytes.NewBuffer(certPem)

	if err := pem.Encode(buf, &certBlock); err != nil {
		return nil, err
	}

	return buf.Bytes(), err
}

func GenerateECKey() (key *ecdsa.PrivateKey, pemData []byte, err error) {
	key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
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
