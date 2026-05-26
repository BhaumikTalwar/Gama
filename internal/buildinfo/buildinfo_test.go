package buildinfo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVersionStr_Default(t *testing.T) {
	v := VersionStr()
	assert.Equal(t, "dev", v)
}

func TestCommitStr_Default(t *testing.T) {
	c := CommitStr()
	assert.Equal(t, "none", c)
}

func TestBuildInfo_ContainsFields(t *testing.T) {
	info := BuildInfo()
	assert.Contains(t, info, "Version")
	assert.Contains(t, info, "Commit")
	assert.Contains(t, info, "Build Time")
}

func TestBuildInfo_FormatsCorrectly(t *testing.T) {
	info := BuildInfo()
	assert.Contains(t, info, "Gama")
	assert.Contains(t, info, Version)
	assert.Contains(t, info, Commit)
}
