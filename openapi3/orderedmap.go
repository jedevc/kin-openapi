package openapi3

import (
	"bytes"
	"encoding/json"
	"iter"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// OrderedMap is a generic map that preserves insertion order.
// It wraps github.com/wk8/go-ordered-map/v2 with a simpler API.
type OrderedMap[K comparable, V any] struct {
	m *orderedmap.OrderedMap[K, V]
}

// NewOrderedMap creates a new OrderedMap with optional initial capacity.
func NewOrderedMap[K comparable, V any]() *OrderedMap[K, V] {
	return &OrderedMap[K, V]{m: orderedmap.New[K, V]()}
}

// NewOrderedMapWithCapacity creates a new OrderedMap with the given capacity.
func NewOrderedMapWithCapacity[K comparable, V any](cap int) *OrderedMap[K, V] {
	if cap <= 0 {
		return &OrderedMap[K, V]{m: orderedmap.New[K, V]()}
	}
	return &OrderedMap[K, V]{m: orderedmap.New[K, V](orderedmap.WithCapacity[K, V](cap))}
}

// Reorder rearranges existing entries to match order, moving each key to the
// back in turn. Keys in order that aren't present in the map are skipped;
// existing keys not mentioned in order keep their current relative position,
// trailing after the ones that were placed. It does not add or remove
// entries. Used to recover true document order after a YAML decode, whose
// intermediate representation does not preserve map key order (see
// unmarshal in marsh.go).
func (om *OrderedMap[K, V]) Reorder(order []K) {
	if om == nil || om.m == nil {
		return
	}
	for _, k := range order {
		_ = om.m.MoveToBack(k)
	}
}

// Len returns the number of entries in the map.
func (om *OrderedMap[K, V]) Len() int {
	if om == nil || om.m == nil {
		return 0
	}
	return om.m.Len()
}

// Get returns the value for the given key and whether it was found.
func (om *OrderedMap[K, V]) Get(key K) (V, bool) {
	if om == nil || om.m == nil {
		var zero V
		return zero, false
	}
	return om.m.Get(key)
}

// Value returns the value for the given key, or the zero value if not found.
func (om *OrderedMap[K, V]) Value(key K) V {
	v, _ := om.Get(key)
	return v
}

// Set adds or updates a key-value pair in the map.
func (om *OrderedMap[K, V]) Set(key K, value V) {
	if om.m == nil {
		om.m = orderedmap.New[K, V]()
	}
	om.m.Set(key, value)
}

// Delete removes a key from the map.
func (om *OrderedMap[K, V]) Delete(key K) {
	if om == nil || om.m == nil {
		return
	}
	om.m.Delete(key)
}

// Iter returns an iterator over key-value pairs in insertion order.
func (om *OrderedMap[K, V]) Iter() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if om == nil || om.m == nil {
			return
		}
		for pair := om.m.Oldest(); pair != nil; pair = pair.Next() {
			if !yield(pair.Key, pair.Value) {
				return
			}
		}
	}
}

// Keys returns a slice of all keys in insertion order.
func (om *OrderedMap[K, V]) Keys() []K {
	if om == nil || om.m == nil {
		return nil
	}
	keys := make([]K, 0, om.m.Len())
	for pair := om.m.Oldest(); pair != nil; pair = pair.Next() {
		keys = append(keys, pair.Key)
	}
	return keys
}

// Map returns a regular map copy of the ordered map.
// Note: iteration on Go maps is not ordered.
func (om *OrderedMap[K, V]) Map() map[K]V {
	if om == nil || om.m == nil {
		return make(map[K]V)
	}
	m := make(map[K]V, om.m.Len())
	for pair := om.m.Oldest(); pair != nil; pair = pair.Next() {
		m[pair.Key] = pair.Value
	}
	return m
}

// Clone returns a shallow copy of the ordered map.
func (om *OrderedMap[K, V]) Clone() *OrderedMap[K, V] {
	if om == nil || om.m == nil {
		return NewOrderedMap[K, V]()
	}
	clone := NewOrderedMapWithCapacity[K, V](om.m.Len())
	for pair := om.m.Oldest(); pair != nil; pair = pair.Next() {
		clone.Set(pair.Key, pair.Value)
	}
	return clone
}

// MarshalJSON implements json.Marshaler, preserving key order.
func (om *OrderedMap[K, V]) MarshalJSON() ([]byte, error) {
	if om == nil || om.m == nil {
		return []byte("null"), nil
	}
	return om.m.MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler, preserving key order.
// Note: This only works for OrderedMap[string, V] where V implements json.Unmarshaler
// or is a basic JSON type.
func (om *OrderedMap[K, V]) UnmarshalJSON(data []byte) error {
	if om.m == nil {
		om.m = orderedmap.New[K, V]()
	}
	return om.m.UnmarshalJSON(data)
}

// unmarshalJSONWithOrder unmarshals JSON while preserving key order using a streaming decoder.
// The decode function is called for each key-value pair with the key and raw JSON value.
func unmarshalJSONWithOrder(data []byte, decode func(key string, value json.RawMessage) error) error {
	dec := json.NewDecoder(bytes.NewReader(data))

	// Read opening brace
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return &json.UnmarshalTypeError{Value: "non-object", Type: nil}
	}

	// Read key-value pairs in order
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return err
		}
		key := tok.(string)

		var rawValue json.RawMessage
		if err := dec.Decode(&rawValue); err != nil {
			return err
		}

		if err := decode(key, rawValue); err != nil {
			return err
		}
	}

	return nil
}
