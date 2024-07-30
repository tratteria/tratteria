package utils

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func CanonicalizeJSON(data interface{}) (string, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		var pairs []string

		keys := make([]string, 0, len(v))

		for k := range v {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, k := range keys {
			val, err := CanonicalizeJSON(v[k])
			if err != nil {
				return "", err
			}

			pairs = append(pairs, fmt.Sprintf("%q:%s", k, val))
		}

		return "{" + strings.Join(pairs, ",") + "}", nil
	case []interface{}:
		var items []string

		for _, item := range v {
			val, err := CanonicalizeJSON(item)
			if err != nil {
				return "", err
			}

			items = append(items, val)
		}

		return "[" + strings.Join(items, ",") + "]", nil
	default:
		jsonBytes, err := json.Marshal(v)
		if err != nil {
			return "", err
		}

		return string(jsonBytes), nil
	}
}
