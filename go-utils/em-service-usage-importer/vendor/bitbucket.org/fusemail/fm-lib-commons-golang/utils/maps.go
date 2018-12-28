package utils

import (
	"encoding/json"

	"golang.org/x/sync/syncmap"
)

// Provides map utilities.

/*
OpenMap provides a general purpose map wrapper, with string keys and interface values.
Does NOT support concurrent access.
For concurrency, use SyncMap instead.
*/
type OpenMap map[string]interface{}

// SyncMap provides concurrent map @ golang.org/x/sync/syncmap.
type SyncMap struct {
	*syncmap.Map
}

// NewSyncMap constructs SyncMap instance for concurrent maps.
func NewSyncMap() *SyncMap {
	return &SyncMap{
		Map: &syncmap.Map{},
	}
}

// MarshalJSON generates json with concurrent protection.
func (m *SyncMap) MarshalJSON() ([]byte, error) {
	o := OpenMap{}
	m.Range(func(k, v interface{}) bool {
		o[k.(string)] = v
		return true
	})
	return json.Marshal(o)
}
