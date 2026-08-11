package openapi3

import (
	"encoding/json"
	"fmt"
	"iter"
	"reflect"

	"github.com/go-openapi/jsonpointer"
)

// ComponentMap is an ordered map of *E, keyed by component name. Iteration
// follows the order the entries were read from (or added to) the document.
type ComponentMap[E any] struct {
	m *OrderedMap[string, *E]
}

// Schemas is an ordered map of SchemaRef.
type Schemas = ComponentMap[SchemaRef]

// ParametersMap is an ordered map of ParameterRef.
type ParametersMap = ComponentMap[ParameterRef]

// Headers is an ordered map of HeaderRef.
type Headers = ComponentMap[HeaderRef]

// RequestBodies is an ordered map of RequestBodyRef.
type RequestBodies = ComponentMap[RequestBodyRef]

// ResponseBodies is an ordered map of ResponseRef (for components/responses).
type ResponseBodies = ComponentMap[ResponseRef]

// SecuritySchemes is an ordered map of SecuritySchemeRef.
type SecuritySchemes = ComponentMap[SecuritySchemeRef]

// Examples is an ordered map of ExampleRef.
type Examples = ComponentMap[ExampleRef]

// Links is an ordered map of LinkRef.
type Links = ComponentMap[LinkRef]

// Callbacks is an ordered map of CallbackRef.
type Callbacks = ComponentMap[CallbackRef]

// Encodings is an ordered map of Encoding, keyed by field name.
type Encodings = ComponentMap[Encoding]

// ServerVariables is an ordered map of ServerVariable, keyed by variable name.
type ServerVariables = ComponentMap[ServerVariable]

// Webhooks is an ordered map of PathItem, keyed by webhook name.
type Webhooks = ComponentMap[PathItem]

// componentNouns names each element type for JSONLookup error messages.
var componentNouns = map[reflect.Type]string{
	reflect.TypeFor[SchemaRef]():         "schema",
	reflect.TypeFor[ParameterRef]():      "parameter",
	reflect.TypeFor[HeaderRef]():         "header",
	reflect.TypeFor[RequestBodyRef]():    "request body",
	reflect.TypeFor[ResponseRef]():       "response",
	reflect.TypeFor[SecuritySchemeRef](): "security scheme",
	reflect.TypeFor[ExampleRef]():        "example",
	reflect.TypeFor[LinkRef]():           "link",
	reflect.TypeFor[CallbackRef]():       "callback",
	reflect.TypeFor[Encoding]():          "encoding",
	reflect.TypeFor[ServerVariable]():    "server variable",
	reflect.TypeFor[PathItem]():          "webhook",
}

// refWrapper is implemented by the generated *Ref types (see refs.tmpl).
type refWrapper interface {
	RefString() string
	refValue() any
}

func NewComponentMap[E any]() *ComponentMap[E] {
	return &ComponentMap[E]{m: NewOrderedMap[string, *E]()}
}

func NewComponentMapWithCapacity[E any](cap int) *ComponentMap[E] {
	return &ComponentMap[E]{m: NewOrderedMapWithCapacity[string, *E](cap)}
}

func ComponentMapFromMap[E any](m map[string]*E) ComponentMap[E] {
	cm := NewComponentMapWithCapacity[E](len(m))
	for k, v := range m {
		cm.Set(k, v)
	}
	return *cm
}

func (cm *ComponentMap[E]) Len() int {
	if cm == nil || cm.m == nil {
		return 0
	}
	return cm.m.Len()
}

func (cm *ComponentMap[E]) Get(key string) (*E, bool) {
	if cm == nil || cm.m == nil {
		return nil, false
	}
	return cm.m.Get(key)
}

func (cm *ComponentMap[E]) Value(key string) *E {
	if cm == nil || cm.m == nil {
		return nil
	}
	return cm.m.Value(key)
}

func (cm *ComponentMap[E]) Set(key string, value *E) {
	if cm.m == nil {
		cm.m = NewOrderedMap[string, *E]()
	}
	cm.m.Set(key, value)
}

func (cm *ComponentMap[E]) Delete(key string) {
	if cm != nil && cm.m != nil {
		cm.m.Delete(key)
	}
}

func (cm *ComponentMap[E]) Map() map[string]*E {
	if cm == nil || cm.m == nil {
		return make(map[string]*E)
	}
	return cm.m.Map()
}

func (cm *ComponentMap[E]) Iter() iter.Seq2[string, *E] {
	if cm == nil || cm.m == nil {
		return func(yield func(string, *E) bool) {}
	}
	return cm.m.Iter()
}

func (cm *ComponentMap[E]) Keys() []string {
	if cm == nil || cm.m == nil {
		return nil
	}
	return cm.m.Keys()
}

func (cm ComponentMap[E]) MarshalJSON() ([]byte, error) {
	if cm.m == nil {
		return []byte("{}"), nil
	}
	return cm.m.MarshalJSON()
}

func (cm ComponentMap[E]) MarshalYAML() (any, error) {
	m := make(map[string]any, cm.Len())
	for k, v := range cm.Iter() {
		m[k] = v
	}
	return m, nil
}

func (cm *ComponentMap[E]) UnmarshalJSON(data []byte) error {
	if cm.m == nil {
		cm.m = NewOrderedMap[string, *E]()
	}
	return unmarshalJSONWithOrder(data, func(k string, v json.RawMessage) error {
		vv := new(E)
		if err := json.Unmarshal(v, vv); err != nil {
			return err
		}
		cm.m.Set(k, vv)
		return nil
	})
}

var _ jsonpointer.JSONPointable = (*ComponentMap[SchemaRef])(nil)

func (cm *ComponentMap[E]) JSONLookup(token string) (any, error) {
	v, ok := cm.Get(token)
	if !ok || v == nil {
		noun, found := componentNouns[reflect.TypeFor[E]()]
		if !found {
			noun = "component"
		}
		return nil, fmt.Errorf("no %s %q", noun, token)
	}
	if ref, ok := any(v).(refWrapper); ok {
		if s := ref.RefString(); s != "" {
			return &Ref{Ref: s}, nil
		}
		return ref.refValue(), nil
	}
	return v, nil
}

func NewSchemas() *Schemas { return NewComponentMap[SchemaRef]() }

func NewSchemasWithCapacity(cap int) *Schemas { return NewComponentMapWithCapacity[SchemaRef](cap) }

func SchemasFromMap(m map[string]*SchemaRef) Schemas { return ComponentMapFromMap(m) }

func NewParametersMap() *ParametersMap { return NewComponentMap[ParameterRef]() }

func NewParametersMapWithCapacity(cap int) *ParametersMap {
	return NewComponentMapWithCapacity[ParameterRef](cap)
}

func ParametersMapFromMap(m map[string]*ParameterRef) ParametersMap { return ComponentMapFromMap(m) }

func NewHeaders() *Headers { return NewComponentMap[HeaderRef]() }

func NewHeadersWithCapacity(cap int) *Headers { return NewComponentMapWithCapacity[HeaderRef](cap) }

func HeadersFromMap(m map[string]*HeaderRef) Headers { return ComponentMapFromMap(m) }

func NewRequestBodies() *RequestBodies { return NewComponentMap[RequestBodyRef]() }

func NewRequestBodiesWithCapacity(cap int) *RequestBodies {
	return NewComponentMapWithCapacity[RequestBodyRef](cap)
}

func RequestBodiesFromMap(m map[string]*RequestBodyRef) RequestBodies { return ComponentMapFromMap(m) }

func NewResponseBodies() *ResponseBodies { return NewComponentMap[ResponseRef]() }

func NewResponseBodiesWithCapacity(cap int) *ResponseBodies {
	return NewComponentMapWithCapacity[ResponseRef](cap)
}

func ResponseBodiesFromMap(m map[string]*ResponseRef) ResponseBodies { return ComponentMapFromMap(m) }

func NewSecuritySchemes() *SecuritySchemes { return NewComponentMap[SecuritySchemeRef]() }

func NewSecuritySchemesWithCapacity(cap int) *SecuritySchemes {
	return NewComponentMapWithCapacity[SecuritySchemeRef](cap)
}

func SecuritySchemesFromMap(m map[string]*SecuritySchemeRef) SecuritySchemes {
	return ComponentMapFromMap(m)
}

func NewExamples() *Examples { return NewComponentMap[ExampleRef]() }

func NewExamplesWithCapacity(cap int) *Examples { return NewComponentMapWithCapacity[ExampleRef](cap) }

func ExamplesFromMap(m map[string]*ExampleRef) Examples { return ComponentMapFromMap(m) }

func NewLinks() *Links { return NewComponentMap[LinkRef]() }

func NewLinksWithCapacity(cap int) *Links { return NewComponentMapWithCapacity[LinkRef](cap) }

func LinksFromMap(m map[string]*LinkRef) Links { return ComponentMapFromMap(m) }

func NewCallbacks() *Callbacks { return NewComponentMap[CallbackRef]() }

func NewCallbacksWithCapacity(cap int) *Callbacks {
	return NewComponentMapWithCapacity[CallbackRef](cap)
}

func CallbacksFromMap(m map[string]*CallbackRef) Callbacks { return ComponentMapFromMap(m) }

func NewEncodings() *Encodings { return NewComponentMap[Encoding]() }

func NewEncodingsWithCapacity(cap int) *Encodings { return NewComponentMapWithCapacity[Encoding](cap) }

func EncodingsFromMap(m map[string]*Encoding) Encodings { return ComponentMapFromMap(m) }

func NewServerVariables() *ServerVariables { return NewComponentMap[ServerVariable]() }

func NewServerVariablesWithCapacity(cap int) *ServerVariables {
	return NewComponentMapWithCapacity[ServerVariable](cap)
}

func ServerVariablesFromMap(m map[string]*ServerVariable) ServerVariables {
	return ComponentMapFromMap(m)
}

func NewWebhooks() *Webhooks { return NewComponentMap[PathItem]() }

func NewWebhooksWithCapacity(cap int) *Webhooks { return NewComponentMapWithCapacity[PathItem](cap) }

func WebhooksFromMap(m map[string]*PathItem) Webhooks { return ComponentMapFromMap(m) }
