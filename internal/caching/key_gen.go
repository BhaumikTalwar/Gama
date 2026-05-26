package caching

import (
	"fmt"
	"sort"
	"strings"
)

type KeyGen struct {
	namespace string
}

func NewKeyGen(namespace string) *KeyGen {
	return &KeyGen{
		namespace: namespace,
	}
}

func (k *KeyGen) WithVersion(entity string, version int64, suffix string) string {
	return fmt.Sprintf("%s:%s:v%d:%s", k.namespace, entity, version, suffix)
}

func (k *KeyGen) WithVersions(versions map[string]map[string]int64, suffix string) string {
	var parts []string
	entities := make([]string, 0, len(versions))
	for entity := range versions {
		entities = append(entities, entity)
	}
	sort.Strings(entities)

	for _, entity := range entities {
		ids := make([]string, 0, len(versions[entity]))
		for id := range versions[entity] {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		for _, id := range ids {
			parts = append(parts, fmt.Sprintf("%s:%s:v%d", entity, id, versions[entity][id]))
		}
	}

	return fmt.Sprintf("%s:%s:%s", k.namespace, strings.Join(parts, ","), suffix)
}

func (k *KeyGen) Simple(entity string, id string, suffix string) string {
	return fmt.Sprintf("%s:%s:%s:%s", k.namespace, entity, id, suffix)
}

func (k *KeyGen) WithParams(key string, params ...interface{}) string {
	if len(params) == 0 {
		return key
	}
	var parts []string
	for i := 0; i < len(params); i += 2 {
		if i+1 < len(params) {
			parts = append(parts, fmt.Sprintf("%v:%v", params[i], params[i+1]))
		}
	}
	return key + ":" + strings.Join(parts, ":")
}

func (k *KeyGen) WithParamMap(key string, params map[string]interface{}) string {
	if len(params) == 0 {
		return key
	}
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%v", k, params[k]))
	}
	return key + ":" + strings.Join(parts, ":")
}