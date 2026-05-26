package utils

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFileExistsInPath_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(tmpFile, []byte("hello"), 0o644)
	assert.NoError(t, err)

	assert.True(t, FileExistsInPath(tmpFile))
}

func TestFileExistsInPath_NotExists(t *testing.T) {
	assert.False(t, FileExistsInPath("/nonexistent/path/file.txt"))
}

func TestIsDir_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	assert.True(t, IsDir(tmpDir))
}

func TestIsDir_File(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	os.WriteFile(tmpFile, []byte("data"), 0o644)
	assert.False(t, IsDir(tmpFile))
}

func TestIsDir_NotExists(t *testing.T) {
	assert.False(t, IsDir("/nonexistent"))
}

func TestIsRegularFile_File(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "regular.txt")
	os.WriteFile(tmpFile, []byte("data"), 0o644)
	assert.True(t, IsRegularFile(tmpFile))
}

func TestIsRegularFile_Directory(t *testing.T) {
	tmpDir := t.TempDir()
	assert.False(t, IsRegularFile(tmpDir))
}

func TestIsRegularFile_NotExists(t *testing.T) {
	assert.False(t, IsRegularFile("/nonexistent"))
}

func TestIsSymlink_Symlink(t *testing.T) {
	tmpDir := t.TempDir()
	target := filepath.Join(tmpDir, "target")
	link := filepath.Join(tmpDir, "link")
	os.WriteFile(target, []byte("data"), 0o644)
	err := os.Symlink(target, link)
	if err == nil {
		assert.True(t, IsSymlink(link))
	}
}

func TestIsSymlink_RegularFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "regular.txt")
	os.WriteFile(tmpFile, []byte("data"), 0o644)
	assert.False(t, IsSymlink(tmpFile))
}

func TestIsSymlink_NotExists(t *testing.T) {
	assert.False(t, IsSymlink("/nonexistent"))
}
