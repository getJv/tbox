package di

import (
	"strings"
)

// tracked keeps the dependency build order for cycle detection messages.
type tracked map[string]int

// add returns a copy of the current tracked list including the new dependency.
func (s tracked) add(info depInfo) tracked {
	newList := make(tracked, len(s))

	for k, v := range s {
		newList[k] = v
	}
	newList[info.key] = len(newList)

	return newList
}

// ordered returns dependency keys ordered by insertion index.
func (s tracked) ordered() []string {
	keys := make([]string, len(s))

	for key, i := range s {
		keys[i] = key
	}

	return keys
}

// String returns the tracked dependency chain as a comma-separated list.
func (s tracked) String() string {
	return strings.Join(s.ordered(), ",")
}
