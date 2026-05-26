package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/BhaumikTalwar/Gama/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func captureOutput(f func() error) (string, error) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	return buf.String(), err
}

func TestVersionCmd_PrintsVersion(t *testing.T) {
	output, err := captureOutput(func() error {
		return versionCmd.RunE(versionCmd, nil)
	})
	require.NoError(t, err)
	assert.NotEmpty(t, output)
}

func TestInfoCmd_PrintsBuildInfo(t *testing.T) {
	output, err := captureOutput(func() error {
		return infoCmd.RunE(infoCmd, nil)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Version")
	assert.Contains(t, output, "Commit")
}

func TestCommands_Registered(t *testing.T) {
	cmds := rootCmd.Commands()
	cmdNames := make(map[string]bool)
	for _, c := range cmds {
		cmdNames[c.Name()] = true
	}

	assert.True(t, cmdNames["server"])
	assert.True(t, cmdNames["version"])
	assert.True(t, cmdNames["info"])
	assert.True(t, cmdNames["config"])

	configCmd := rootCmd.Commands()[0]
	for _, c := range rootCmd.Commands() {
		if c.Name() == "config" {
			configCmd = c
			break
		}
	}
	subCmdNames := make(map[string]bool)
	for _, c := range configCmd.Commands() {
		subCmdNames[c.Name()] = true
	}
	assert.True(t, subCmdNames["verify"])
	assert.True(t, subCmdNames["init"])
}

func TestSrvDumpCmd_ContainsConfigKeys(t *testing.T) {
	config.SetTestAppConfig(&config.AppConfig{
		AppName:                  "TestApp",
		Env:                      "test",
		CorsAddresses:            []string{"http://localhost:3000"},
		AppSecret:                "test-app-secret-32-bytes-long!!!",
		JWTKey:                   "test-jwt-key-32-bytes-long-for-test!",
		AESKey:                   "test-aes-key-32-bytes-for-tests!",
		AccessTokenDuration:      config.GetAppConfig().AccessTokenDuration,
		RefreshTokenDuration:     config.GetAppConfig().RefreshTokenDuration,
		RefreshRotationThreshold: config.GetAppConfig().RefreshRotationThreshold,
	})

	output, err := captureOutput(func() error {
		return srvDump.RunE(srvDump, nil)
	})
	require.NoError(t, err)
	assert.Contains(t, output, "App")
	assert.Contains(t, output, "<REDACTED>")
}

func TestConfigVerifyCmd_InvalidPath(t *testing.T) {
	setupConfigLoad("")
	cfgFile = "/nonexistent/config.yaml"

	var err error
	output, captureErr := captureOutput(func() error {
		err = configVerifyCmd.RunE(configVerifyCmd, nil)
		return err
	})
	_ = output
	assert.Error(t, captureErr)
}

func setupConfigLoad(configContent string) {
	cfgFile = ""
	if configContent != "" {
		tmpDir, _ := os.MkdirTemp("", "gama-test-*")
		configPath := filepath.Join(tmpDir, "config.yaml")
		os.WriteFile(configPath, []byte(configContent), 0o644)
		cfgFile = configPath
	}
}

func TestRootCmd_Help(t *testing.T) {
	output, err := captureOutput(func() error {
		rootCmd.SetArgs([]string{"--help"})
		return rootCmd.Execute()
	})
	require.NoError(t, err)
	assert.Contains(t, output, "Gama")
}

func TestServeCmd_Registered(t *testing.T) {
	for _, c := range rootCmd.Commands() {
		if c.Name() == "server" {
			assert.NotNil(t, c.Run)
			return
		}
	}
	t.Fatal("server command not found")
}
