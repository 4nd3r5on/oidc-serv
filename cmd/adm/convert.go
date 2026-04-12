package main

import "sort"

// mapKeysArr returns all keys of m as a slice in unspecified order.
func mapKeysArr[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// sortedKeys returns the string keys of m sorted alphabetically.
func sortedKeys[V any](m map[string]V) []string {
	keys := mapKeysArr(m)
	sort.Strings(keys)
	return keys
}
