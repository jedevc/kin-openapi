#!/bin/bash -eux
set -o pipefail

maplike=./openapi3/maplike.go
maplike_test=./openapi3/maplike_test.go

types=()
types+=('*Responses')
types+=('*Callback')
types+=('*Paths')

value_types=()
value_types+=('*ResponseRef')
value_types+=('*PathItem')
value_types+=('*PathItem')

deref_vs=()
deref_vs+=('v.Value')
deref_vs+=('v')
deref_vs+=('v')

names=()
names+=('responses')
names+=('callback')
names+=('paths')

[[ "${#types[@]}" = "${#value_types[@]}" ]]
[[ "${#types[@]}" = "${#deref_vs[@]}" ]]
[[ "${#types[@]}" = "${#names[@]}" ]]
# Each of these 3 types wraps an OrderedMap field declared in its own file
# (response.go, callback.go, paths.go); this counts those declarations
# rather than grepping loader.go, which no longer touches the field directly
# (it goes through the public Map() accessor there instead).
[[ "${#types[@]}" = "$(git grep -InE '^\s+m \*OrderedMap\[string, ' -- openapi3/response.go openapi3/callback.go openapi3/paths.go | wc -l)" ]]


maplike_header() {
	cat <<EOF >"$maplike"
package openapi3

import (
	"encoding/json"
	"iter"
	"maps"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

EOF
}


test_header() {
	cat <<EOF >"$maplike_test"
package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestMaplikeMethods(t *testing.T) {
	t.Parallel()

EOF
}


test_footer() {
	echo "}" >>"$maplike_test"
}


maplike_NewWithCapa() {
	cat <<EOF >>"$maplike"
// New${type#'*'}WithCapacity builds a ${name} object of the given capacity.
func New${type#'*'}WithCapacity(cap int) ${type} {
	return &${type#'*'}{m: NewOrderedMapWithCapacity[string, ${value_type}](cap)}
}

EOF
}


maplike_KeysValueSetLenDeleteMapIter() {
	cat <<EOF >>"$maplike"
// Keys returns the ${name} keys in insertion order.
func (${name} ${type}) Keys() []string {
	if ${name} == nil || ${name}.m == nil {
		return nil
	}
	return ${name}.m.Keys()
}

// Value returns the ${name} for key or nil
func (${name} ${type}) Value(key string) ${value_type} {
	if ${name} == nil || ${name}.m == nil {
		return nil
	}
	return ${name}.m.Value(key)
}

// Set adds or replaces key 'key' of '${name}' with 'value'.
// Note: '${name}' MUST be non-nil
func (${name} ${type}) Set(key string, value ${value_type}) {
	if ${name}.m == nil {
		${name}.m = NewOrderedMap[string, ${value_type}]()
	}
	${name}.m.Set(key, value)
}

// Len returns the amount of keys in ${name} excluding ${name}.Extensions.
func (${name} ${type}) Len() int {
	if ${name} == nil || ${name}.m == nil {
		return 0
	}
	return ${name}.m.Len()
}

// Delete removes the entry associated with key 'key' from '${name}'.
func (${name} ${type}) Delete(key string) {
	if ${name} != nil && ${name}.m != nil {
		${name}.m.Delete(key)
	}
}

// Map returns ${name} as a 'map'.
// Note: iteration on Go maps is not ordered.
func (${name} ${type}) Map() map[string]${value_type} {
	if ${name} == nil || ${name}.m == nil {
		return make(map[string]${value_type})
	}
	return ${name}.m.Map()
}

// Iter returns an iterator over ${name} in insertion order.
func (${name} ${type}) Iter() iter.Seq2[string, ${value_type}] {
	if ${name} == nil || ${name}.m == nil {
		return func(yield func(string, ${value_type}) bool) {}
	}
	return ${name}.m.Iter()
}

EOF
}


maplike_Pointable() {
	cat <<EOF >>"$maplike"
var _ jsonpointer.JSONPointable = (${type})(nil)

// JSONLookup implements https://github.com/go-openapi/jsonpointer#JSONPointable
func (${name} ${type#'*'}) JSONLookup(token string) (any, error) {
	if v := ${name}.Value(token); v == nil {
		vv, _, err := jsonpointer.GetForToken(${name}.Extensions, token)
		return vv, err
	} else if ref := v.Ref; ref != "" {
		return &Ref{Ref: ref}, nil
	} else {
		return ${deref_v}, nil
	}
}

EOF
}


maplike_UnMarsh() {
	if [[ "$type" != '*'* ]]; then
		echo "TODO: impl non-pointer receiver YAML Marshaler"
		exit 2
	fi
	local nil_condition="${name} == nil"
	if [[ "$type" == '*Responses' ]]; then
		nil_condition+=" || ${name}.isExplicitlyNull()"
	fi
	cat <<EOF >>"$maplike"
// MarshalYAML returns the YAML encoding of ${type#'*'}.
func (${name} ${type}) MarshalYAML() (any, error) {
	if ${nil_condition} {
		return nil, nil
	}
	m := make(map[string]any, ${name}.Len()+len(${name}.Extensions))
	maps.Copy(m, ${name}.Extensions)
	for k, v := range ${name}.Iter() {
		m[k] = v
	}
	return m, nil
}

// MarshalJSON returns the JSON encoding of ${type#'*'}.
func (${name} ${type}) MarshalJSON() ([]byte, error) {
	if ${nil_condition} {
		return []byte("null"), nil
	}
	m := NewOrderedMap[string, any]()
	for k, v := range ${name}.Iter() {
		m.Set(k, v)
	}
	for k, v := range ${name}.Extensions {
		m.Set(k, v)
	}
	return m.MarshalJSON()
}

// UnmarshalJSON sets ${type#'*'} to a copy of data.
func (${name} ${type}) UnmarshalJSON(data []byte) error {
	x := &${type#'*'}{
		Extensions: make(map[string]any),
		m:          NewOrderedMap[string, ${value_type}](),
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

		var vv ${value_type#'*'}
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
	*${name} = *x
	return nil
}
EOF
}


test_body() {
	cat <<EOF >>"$maplike_test"
	t.Run("${type}", func(t *testing.T) {
		t.Parallel()
		t.Run("nil", func(t *testing.T) {
			x := (${type})(nil)
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]${value_type}{}, x.Map())
			require.Equal(t, (${value_type})(nil), x.Value("key"))
			require.Panics(t, func() { x.Set("key", &${value_type#'*'}{}) })
			require.NotPanics(t, func() { x.Delete("key") })
		})
		t.Run("nonnil", func(t *testing.T) {
			x := &${type#'*'}{}
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]${value_type}{}, x.Map())
			require.Equal(t, (${value_type})(nil), x.Value("key"))
			x.Set("key", &${value_type#'*'}{})
			require.Equal(t, 1, x.Len())
			require.Equal(t, map[string]${value_type}{"key": {}}, x.Map())
			require.Equal(t, &${value_type#'*'}{}, x.Value("key"))
			x.Delete("key")
			require.Equal(t, 0, x.Len())
			require.Equal(t, map[string]${value_type}{}, x.Map())
			require.Equal(t, (${value_type})(nil), x.Value("key"))
			require.NotPanics(t, func() { x.Delete("key") })
		})
	})

EOF
}



maplike_header
test_header

for i in "${!types[@]}"; do
	type=${types[$i]}
	value_type=${value_types[$i]}
	deref_v=${deref_vs[$i]}
	name=${names[$i]}

	type="$type" name="$name" value_type="$value_type" maplike_NewWithCapa
	type="$type" name="$name" value_type="$value_type" maplike_KeysValueSetLenDeleteMapIter
	type="$type" name="$name"    deref_v="$deref_v"    maplike_Pointable
	type="$type" name="$name" value_type="$value_type" maplike_UnMarsh
	[[ $((i+1)) != "${#types[@]}" ]] && echo >>"$maplike"

	type="${type/'*'/*openapi3.}" value_type="${value_type/'*'/*openapi3.}" test_body


done

test_footer
