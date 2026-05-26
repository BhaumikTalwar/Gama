package postgres

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func generateTestCert(t *testing.T, dir string) (certPath, keyPath string) {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
	}

	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	require.NoError(t, err)

	certPath = filepath.Join(dir, "cert.pem")
	certFile, err := os.Create(certPath)
	require.NoError(t, err)
	defer certFile.Close()
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	require.NoError(t, err)

	keyPath = filepath.Join(dir, "key.pem")
	keyFile, err := os.Create(keyPath)
	require.NoError(t, err)
	defer keyFile.Close()
	err = pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})
	require.NoError(t, err)

	return certPath, keyPath
}

func generateCACert(t *testing.T, dir string) string {
	t.Helper()
	certPath, _ := generateTestCert(t, dir)
	return certPath
}

func TestGetTLSConfig_NoTLSConfig(t *testing.T) {
	cfg := &config.PostgresConfig{Host: "localhost"}
	result, err := getTLSConfig(cfg)
	require.NoError(t, err)
	assert.Equal(t, "localhost", result.ServerName)
	assert.Nil(t, result.RootCAs)
	assert.Empty(t, result.Certificates)
}

func TestGetTLSConfig_WithCA(t *testing.T) {
	dir := t.TempDir()
	caPath := generateCACert(t, dir)

	cfg := &config.PostgresConfig{Host: "db.example.com", SSLCAPath: caPath}
	result, err := getTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, result.RootCAs)
}

func TestGetTLSConfig_CANotFound(t *testing.T) {
	cfg := &config.PostgresConfig{Host: "localhost", SSLCAPath: "/nonexistent/ca.pem"}
	_, err := getTLSConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to read CA cert")
}

func TestGetTLSConfig_InvalidCAPEM(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.pem")
	os.WriteFile(badPath, []byte("not-a-valid-pem"), 0o644)

	cfg := &config.PostgresConfig{Host: "localhost", SSLCAPath: badPath}
	_, err := getTLSConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to append CA cert")
}

func TestGetTLSConfig_WithClientCert(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := generateTestCert(t, dir)

	cfg := &config.PostgresConfig{Host: "localhost", SSLCertPath: certPath, SSLKeyPath: keyPath}
	result, err := getTLSConfig(cfg)
	require.NoError(t, err)
	assert.Len(t, result.Certificates, 1)
}

func TestGetTLSConfig_ClientCertMissingKey(t *testing.T) {
	dir := t.TempDir()
	certPath, _ := generateTestCert(t, dir)

	cfg := &config.PostgresConfig{Host: "localhost", SSLCertPath: certPath, SSLKeyPath: ""}
	result, err := getTLSConfig(cfg)
	require.NoError(t, err)
	assert.Empty(t, result.Certificates)
}

func TestGetTLSConfig_ClientCertNotFound(t *testing.T) {
	cfg := &config.PostgresConfig{
		Host: "localhost", SSLCertPath: "/nonexistent/cert.pem", SSLKeyPath: "/nonexistent/key.pem",
	}
	_, err := getTLSConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load client cert")
}

func TestGetTLSConfig_CAAndClientCert(t *testing.T) {
	dir := t.TempDir()
	caPath := generateCACert(t, dir)
	certPath, keyPath := generateTestCert(t, dir)

	cfg := &config.PostgresConfig{
		Host: "pg.example.com", SSLCAPath: caPath,
		SSLCertPath: certPath, SSLKeyPath: keyPath,
	}
	result, err := getTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, result.RootCAs)
	assert.Len(t, result.Certificates, 1)
}
