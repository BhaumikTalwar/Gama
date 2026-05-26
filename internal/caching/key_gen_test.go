package caching

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestKeyGen_WithVersion(t *testing.T) {
	kg := NewKeyGen("app")

	key := kg.WithVersion("user", 5, "detail")
	assert.Equal(t, "app:user:v5:detail", key)
}

func TestKeyGen_WithVersions(t *testing.T) {
	kg := NewKeyGen("app")

	versions := map[string]map[string]int64{
		"user": {"global": 3},
		"role": {"admin": 7},
	}

	key := kg.WithVersions(versions, "search")
	assert.Equal(t, "app:role:admin:v7,user:global:v3:search", key)
}

func TestKeyGen_Simple(t *testing.T) {
	kg := NewKeyGen("app")

	key := kg.Simple("user", "global", "list")
	assert.Equal(t, "app:user:global:list", key)
}

func TestKeyGen_WithParams(t *testing.T) {
	kg := NewKeyGen("app")

	key := kg.WithParams(kg.Simple("user", "global", "search"), "q", "john", "page", "1")
	assert.Equal(t, "app:user:global:search:q:john:page:1", key)
}

func TestKeyGen_WithParamMap(t *testing.T) {
	kg := NewKeyGen("app")

	key := kg.WithParamMap("search", map[string]any{
		"q":      "john",
		"limit":  10,
		"offset": 20,
	})
	assert.Equal(t, "search:limit:10:offset:20:q:john", key)
}

func TestKeyGen_WithParamMap_Empty(t *testing.T) {
	kg := NewKeyGen("app")

	key := kg.WithParamMap("search", map[string]any{})
	assert.Equal(t, "search", key)
}

func TestKeyGen_WithParams_EmptyParams(t *testing.T) {
	kg := NewKeyGen("app")

	key := kg.WithParams("search")
	assert.Equal(t, "search", key)
}

func TestKeyGen_VersionSorting(t *testing.T) {
	kg := NewKeyGen("app")

	versions := map[string]map[string]int64{
		"zebra": {"1": 1},
		"apple": {"2": 2},
		"banana": {"3": 3},
	}

	key := kg.WithVersions(versions, "search")
	assert.Equal(t, "app:apple:2:v2,banana:3:v3,zebra:1:v1:search", key)
}

func TestKeyGen_IDSorting(t *testing.T) {
	kg := NewKeyGen("app")

	versions := map[string]map[string]int64{
		"user": {"z": 1, "a": 2, "m": 3},
	}

	key := kg.WithVersions(versions, "search")
	assert.Equal(t, "app:user:a:v2,user:m:v3,user:z:v1:search", key)
}