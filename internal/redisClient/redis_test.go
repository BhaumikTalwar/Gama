package redisClient

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
	pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	keyPath = filepath.Join(dir, "key.pem")
	keyFile, err := os.Create(keyPath)
	require.NoError(t, err)
	defer keyFile.Close()
	pem.Encode(keyFile, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	return certPath, keyPath
}

func TestGetRedisTLSConfig_NoTLS(t *testing.T) {
	cfg := &config.RedisConfig{RedisHost: "localhost"}
	result, err := getRedisTLSConfig(cfg)
	require.NoError(t, err)
	assert.Equal(t, "localhost", result.ServerName)
	assert.Nil(t, result.RootCAs)
	assert.Empty(t, result.Certificates)
}

func TestGetRedisTLSConfig_WithCA(t *testing.T) {
	dir := t.TempDir()
	caPath, _ := generateTestCert(t, dir)

	cfg := &config.RedisConfig{RedisHost: "r.example.com", RedisTLSCAFile: caPath}
	result, err := getRedisTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, result.RootCAs)
}

func TestGetRedisTLSConfig_CANotFound(t *testing.T) {
	cfg := &config.RedisConfig{RedisHost: "localhost", RedisTLSCAFile: "/nonexistent/ca.pem"}
	_, err := getRedisTLSConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "reading redis CA file")
}

func TestGetRedisTLSConfig_InvalidCAPEM(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.pem")
	os.WriteFile(badPath, []byte("not-a-valid-cert"), 0o644)

	cfg := &config.RedisConfig{RedisHost: "localhost", RedisTLSCAFile: badPath}
	_, err := getRedisTLSConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to append redis CA cert")
}

func TestGetRedisTLSConfig_WithClientCert(t *testing.T) {
	dir := t.TempDir()
	certPath, keyPath := generateTestCert(t, dir)

	cfg := &config.RedisConfig{
		RedisHost: "localhost", RedisTLSCertFile: certPath, RedisTLSKeyFile: keyPath,
	}
	result, err := getRedisTLSConfig(cfg)
	require.NoError(t, err)
	assert.Len(t, result.Certificates, 1)
}

func TestGetRedisTLSConfig_ClientCertMissingKey(t *testing.T) {
	dir := t.TempDir()
	certPath, _ := generateTestCert(t, dir)

	cfg := &config.RedisConfig{
		RedisHost: "localhost", RedisTLSCertFile: certPath, RedisTLSKeyFile: "",
	}
	result, err := getRedisTLSConfig(cfg)
	require.NoError(t, err)
	assert.Empty(t, result.Certificates)
}

func TestGetRedisTLSConfig_ClientCertNotFound(t *testing.T) {
	cfg := &config.RedisConfig{
		RedisHost: "localhost", RedisTLSCertFile: "/nonexistent/cert.pem", RedisTLSKeyFile: "/nonexistent/key.pem",
	}
	_, err := getRedisTLSConfig(cfg)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "loading redis client cert/key")
}

func TestGetRedisTLSConfig_CAAndClientCert(t *testing.T) {
	dir := t.TempDir()
	caPath, _ := generateTestCert(t, dir)
	certPath, keyPath := generateTestCert(t, dir)

	cfg := &config.RedisConfig{
		RedisHost: "r.example.com", RedisTLSCAFile: caPath,
		RedisTLSCertFile: certPath, RedisTLSKeyFile: keyPath,
	}
	result, err := getRedisTLSConfig(cfg)
	require.NoError(t, err)
	assert.NotNil(t, result.RootCAs)
	assert.Len(t, result.Certificates, 1)
}
