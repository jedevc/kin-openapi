package openapi3

import (
	"encoding/json"
	"iter"
	"maps"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

// NewResponsesWithCapacity builds a responses object of the given capacity.
func NewResponsesWithCapacity(cap int) *Responses {
	return &Responses{m: NewOrderedMapWithCapacity[string, *ResponseRef](cap)}
}

// Keys returns the responses keys in insertion order.
func (responses *Responses) Keys() []string {
	if responses == nil || responses.m == nil {
		return nil
	}
	return responses.m.Keys()
}

// Value returns the responses for key or nil
func (responses *Responses) Value(key string) *ResponseRef {
	if responses == nil || responses.m == nil {
		return nil
	}
	return responses.m.Value(key)
}

// Set adds or replaces key 'key' of 'responses' with 'value'.
// Note: 'responses' MUST be non-nil
func (responses *Responses) Set(key string, value *ResponseRef) {
	if responses.m == nil {
		responses.m = NewOrderedMap[string, *ResponseRef]()
	}
	responses.m.Set(key, value)
}

// Len returns the amount of keys in responses excluding responses.Extensions.
func (responses *Responses) Len() int {
	if responses == nil || responses.m == nil {
		return 0
	}
	return responses.m.Len()
}

// Delete removes the entry associated with key 'key' from 'responses'.
func (responses *Responses) Delete(key string) {
	if responses != nil && responses.m != nil {
		responses.m.Delete(key)
	}
}

// Map returns responses as a 'map'.
// Note: iteration on Go maps is not ordered.
func (responses *Responses) Map() map[string]*ResponseRef {
	if responses == nil || responses.m == nil {
		return make(map[string]*ResponseRef)
	}
	return responses.m.Map()
}

// Iter returns an iterator over responses in insertion order.
func (responses *Responses) Iter() iter.Seq2[string, *ResponseRef] {
	if responses == nil || responses.m == nil {
		return func(yield func(string, *ResponseRef) bool) {}
	}
	return responses.m.Iter()
}

var _ jsonpointer.JSONPointable = (*Responses)(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (responses Responses) JSONLookup(token string) (any, error) {
	if v := responses.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(responses.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		return v.Value, nil
	}
}

// MarshalYAML returns the YAML encoding of Responses.
func (responses *Responses) MarshalYAML() (any, error) {
	if responses == nil || responses.isExplicitlyNull() {
		return nil, nil
	}
	m := make(map[string]any, responses.Len()+len(responses.Extensions))
	maps.Copy(m, responses.Extensions)
	for k, v := range responses.Iter() {
		m[k] = v
	}
	return m, nil
}

// MarshalJSON returns the JSON encoding of Responses.
func (responses *Responses) MarshalJSON() ([]byte, error) {
	if responses == nil || responses.isExplicitlyNull() {
		return []byte("null"), nil
	}
	m := NewOrderedMap[string, any]()
	for k, v := range responses.Iter() {
		m.Set(k, v)
	}
	for k, v := range responses.Extensions {
		m.Set(k, v)
	}
	return m.MarshalJSON()
}

// UnmarshalJSON sets Responses to a copy of data.
func (responses *Responses) UnmarshalJSON(data []byte) error {
	x := &Responses{
		Extensions: make(map[string]any),
		m:          NewOrderedMap[string, *ResponseRef](),
	}

	if err := unmarshalJSONWithOrder(data, func(k string, v json.RawMessage) error {
		if strings.HasPrefix(k, "x-") {
			var ext any
			if err := json.Unmarshal(v, &ext); err != nil {
				return err
			}
			x.Extensions[k] = ext
			return nil
		}

		var vv ResponseRef
		if err := vv.UnmarshalJSON(v); err != nil {
			return err
		}
		x.m.Set(k, &vv)
		return nil
	}); err != nil {
		return err
	}

	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*responses = *x
	return nil
}

// NewCallbackWithCapacity builds a callback object of the given capacity.
func NewCallbackWithCapacity(cap int) *Callback {
	return &Callback{m: NewOrderedMapWithCapacity[string, *PathItem](cap)}
}

// Keys returns the callback keys in insertion order.
func (callback *Callback) Keys() []string {
	if callback == nil || callback.m == nil {
		return nil
	}
	return callback.m.Keys()
}

// Value returns the callback for key or nil
func (callback *Callback) Value(key string) *PathItem {
	if callback == nil || callback.m == nil {
		return nil
	}
	return callback.m.Value(key)
}

// Set adds or replaces key 'key' of 'callback' with 'value'.
// Note: 'callback' MUST be non-nil
func (callback *Callback) Set(key string, value *PathItem) {
	if callback.m == nil {
		callback.m = NewOrderedMap[string, *PathItem]()
	}
	callback.m.Set(key, value)
}

// Len returns the amount of keys in callback excluding callback.Extensions.
func (callback *Callback) Len() int {
	if callback == nil || callback.m == nil {
		return 0
	}
	return callback.m.Len()
}

// Delete removes the entry associated with key 'key' from 'callback'.
func (callback *Callback) Delete(key string) {
	if callback != nil && callback.m != nil {
		callback.m.Delete(key)
	}
}

// Map returns callback as a 'map'.
// Note: iteration on Go maps is not ordered.
func (callback *Callback) Map() map[string]*PathItem {
	if callback == nil || callback.m == nil {
		return make(map[string]*PathItem)
	}
	return callback.m.Map()
}

// Iter returns an iterator over callback in insertion order.
func (callback *Callback) Iter() iter.Seq2[string, *PathItem] {
	if callback == nil || callback.m == nil {
		return func(yield func(string, *PathItem) bool) {}
	}
	return callback.m.Iter()
}

var _ jsonpointer.JSONPointable = (*Callback)(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (callback Callback) JSONLookup(token string) (any, error) {
	if v := callback.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(callback.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		return v, nil
	}
}

// MarshalYAML returns the YAML encoding of Callback.
func (callback *Callback) MarshalYAML() (any, error) {
	if callback == nil {
		return nil, nil
	}
	m := make(map[string]any, callback.Len()+len(callback.Extensions))
	maps.Copy(m, callback.Extensions)
	for k, v := range callback.Iter() {
		m[k] = v
	}
	return m, nil
}

// MarshalJSON returns the JSON encoding of Callback.
func (callback *Callback) MarshalJSON() ([]byte, error) {
	if callback == nil {
		return []byte("null"), nil
	}
	m := NewOrderedMap[string, any]()
	for k, v := range callback.Iter() {
		m.Set(k, v)
	}
	for k, v := range callback.Extensions {
		m.Set(k, v)
	}
	return m.MarshalJSON()
}

// UnmarshalJSON sets Callback to a copy of data.
func (callback *Callback) UnmarshalJSON(data []byte) error {
	x := &Callback{
		Extensions: make(map[string]any),
		m:          NewOrderedMap[string, *PathItem](),
	}

	if err := unmarshalJSONWithOrder(data, func(k string, v json.RawMessage) error {
		if strings.HasPrefix(k, "x-") {
			var ext any
			if err := json.Unmarshal(v, &ext); err != nil {
				return err
			}
			x.Extensions[k] = ext
			return nil
		}

		var vv PathItem
		if err := vv.UnmarshalJSON(v); err != nil {
			return err
		}
		x.m.Set(k, &vv)
		return nil
	}); err != nil {
		return err
	}

	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*callback = *x
	return nil
}

// NewPathsWithCapacity builds a paths object of the given capacity.
func NewPathsWithCapacity(cap int) *Paths {
	return &Paths{m: NewOrderedMapWithCapacity[string, *PathItem](cap)}
}

// Keys returns the paths keys in insertion order.
func (paths *Paths) Keys() []string {
	if paths == nil || paths.m == nil {
		return nil
	}
	return paths.m.Keys()
}

// Value returns the paths for key or nil
func (paths *Paths) Value(key string) *PathItem {
	if paths == nil || paths.m == nil {
		return nil
	}
	return paths.m.Value(key)
}

// Set adds or replaces key 'key' of 'paths' with 'value'.
// Note: 'paths' MUST be non-nil
func (paths *Paths) Set(key string, value *PathItem) {
	if paths.m == nil {
		paths.m = NewOrderedMap[string, *PathItem]()
	}
	paths.m.Set(key, value)
}

// Len returns the amount of keys in paths excluding paths.Extensions.
func (paths *Paths) Len() int {
	if paths == nil || paths.m == nil {
		return 0
	}
	return paths.m.Len()
}

// Delete removes the entry associated with key 'key' from 'paths'.
func (paths *Paths) Delete(key string) {
	if paths != nil && paths.m != nil {
		paths.m.Delete(key)
	}
}

// Map returns paths as a 'map'.
// Note: iteration on Go maps is not ordered.
func (paths *Paths) Map() map[string]*PathItem {
	if paths == nil || paths.m == nil {
		return make(map[string]*PathItem)
	}
	return paths.m.Map()
}

// Iter returns an iterator over paths in insertion order.
func (paths *Paths) Iter() iter.Seq2[string, *PathItem] {
	if paths == nil || paths.m == nil {
		return func(yield func(string, *PathItem) bool) {}
	}
	return paths.m.Iter()
}

var _ jsonpointer.JSONPointable = (*Paths)(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (paths Paths) JSONLookup(token string) (any, error) {
	if v := paths.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(paths.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		return v, nil
	}
}

// MarshalYAML returns the YAML encoding of Paths.
func (paths *Paths) MarshalYAML() (any, error) {
	if paths == nil {
		return nil, nil
	}
	m := make(map[string]any, paths.Len()+len(paths.Extensions))
	maps.Copy(m, paths.Extensions)
	for k, v := range paths.Iter() {
		m[k] = v
	}
	return m, nil
}

// MarshalJSON returns the JSON encoding of Paths.
func (paths *Paths) MarshalJSON() ([]byte, error) {
	if paths == nil {
		return []byte("null"), nil
	}
	m := NewOrderedMap[string, any]()
	for k, v := range paths.Iter() {
		m.Set(k, v)
	}
	for k, v := range paths.Extensions {
		m.Set(k, v)
	}
	return m.MarshalJSON()
}

// UnmarshalJSON sets Paths to a copy of data.
func (paths *Paths) UnmarshalJSON(data []byte) error {
	x := &Paths{
		Extensions: make(map[string]any),
		m:          NewOrderedMap[string, *PathItem](),
	}

	if err := unmarshalJSONWithOrder(data, func(k string, v json.RawMessage) error {
		if strings.HasPrefix(k, "x-") {
			var ext any
			if err := json.Unmarshal(v, &ext); err != nil {
				return err
			}
			x.Extensions[k] = ext
			return nil
		}

		var vv PathItem
		if err := vv.UnmarshalJSON(v); err != nil {
			return err
		}
		x.m.Set(k, &vv)
		return nil
	}); err != nil {
		return err
	}

	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*paths = *x
	return nil
}
