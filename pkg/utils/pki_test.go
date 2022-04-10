package utils

import (
	"crypto/x509"
	"encoding/pem"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GenerateECKey(t *testing.T) {
	key, pem, err := GenerateECKey(256)
	assert.NoError(t, err)

	assert.Equal(t, "P-256", key.Params().Name)
	assert.Equal(t, 256, key.Params().BitSize)

	assert.Contains(t, string(pem), "BEGIN EC PRIVATE KEY")
	assert.Contains(t, string(pem), "END EC PRIVATE KEY")
}

func Test_GenerateCertificate(t *testing.T) {
	key, _, err := GenerateECKey(256)
	assert.NoError(t, err)

	certPem, err := GenerateCertificate(1, &key.PublicKey, key, 1, 0, 0)
	assert.NoError(t, err)

	assert.Contains(t, string(certPem), "BEGIN CERTIFICATE")
	assert.Contains(t, string(certPem), "END CERTIFICATE")

	block, _ := pem.Decode(certPem)
	if block == nil {
		t.Error("failed to parse certificate PEM")
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	assert.NoError(t, err)

	assert.Equal(t, false, cert.IsCA)
	assert.Equal(t, []string{"SANS Institute"}, cert.Issuer.Organization)
	assert.Equal(t, true, cert.BasicConstraintsValid)
	assert.Equal(t, x509.KeyUsageDigitalSignature|x509.KeyUsageCertSign, cert.KeyUsage)
	assert.Equal(t, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth}, cert.ExtKeyUsage)
}
