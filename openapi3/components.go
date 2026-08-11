package openapi3

import (
	"context"
	"encoding/json"
	"iter"
	"maps"
)

// Components is specified by OpenAPI/Swagger standard version 3.
// See https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#components-object
type Components struct {
	Extensions map[string]any `json:"-" yaml:"-"`
	Origin     *Origin        `json:"-" yaml:"-"`

	Schemas         *Schemas         `json:"schemas,omitempty" yaml:"schemas,omitempty"`
	Parameters      *ParametersMap   `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Headers         *Headers         `json:"headers,omitempty" yaml:"headers,omitempty"`
	RequestBodies   *RequestBodies   `json:"requestBodies,omitempty" yaml:"requestBodies,omitempty"`
	Responses       *ResponseBodies  `json:"responses,omitempty" yaml:"responses,omitempty"`
	SecuritySchemes *SecuritySchemes `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`
	Examples        *Examples        `json:"examples,omitempty" yaml:"examples,omitempty"`
	Links           *Links           `json:"links,omitempty" yaml:"links,omitempty"`
	Callbacks       *Callbacks       `json:"callbacks,omitempty" yaml:"callbacks,omitempty"`
}

func NewComponents() Components {
	return Components{}
}

// MarshalJSON returns the JSON encoding of Components.
func (components Components) MarshalJSON() ([]byte, error) {
	x, err := components.MarshalYAML()
	if err != nil {
		return nil, err
	}
	return json.Marshal(x)
}

// MarshalYAML returns the YAML encoding of Components.
func (components Components) MarshalYAML() (any, error) {
	m := make(map[string]any, 9+len(components.Extensions))
	maps.Copy(m, components.Extensions)
	if x := components.Schemas; x.Len() != 0 {
		m["schemas"] = x
	}
	if x := components.Parameters; x.Len() != 0 {
		m["parameters"] = x
	}
	if x := components.Headers; x.Len() != 0 {
		m["headers"] = x
	}
	if x := components.RequestBodies; x.Len() != 0 {
		m["requestBodies"] = x
	}
	if x := components.Responses; x.Len() != 0 {
		m["responses"] = x
	}
	if x := components.SecuritySchemes; x.Len() != 0 {
		m["securitySchemes"] = x
	}
	if x := components.Examples; x.Len() != 0 {
		m["examples"] = x
	}
	if x := components.Links; x.Len() != 0 {
		m["links"] = x
	}
	if x := components.Callbacks; x.Len() != 0 {
		m["callbacks"] = x
	}
	return m, nil
}

// UnmarshalJSON sets Components to a copy of data.
func (components *Components) UnmarshalJSON(data []byte) error {
	type ComponentsBis Components
	var x ComponentsBis
	if err := json.Unmarshal(data, &x); err != nil {
		return unmarshalError(err)
	}
	_ = json.Unmarshal(data, &x.Extensions)
	delete(x.Extensions, "schemas")
	delete(x.Extensions, "parameters")
	delete(x.Extensions, "headers")
	delete(x.Extensions, "requestBodies")
	delete(x.Extensions, "responses")
	delete(x.Extensions, "securitySchemes")
	delete(x.Extensions, "examples")
	delete(x.Extensions, "links")
	delete(x.Extensions, "callbacks")
	if len(x.Extensions) == 0 {
		x.Extensions = nil
	}
	*components = Components(x)
	return nil
}

// Validate returns an error if Components does not comply with the OpenAPI spec.
func (components *Components) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)
	me := newErrCollector(ctx)

	validateMap := func(label string, names []string, validate func(k string) error) error {
		for _, k := range names {
			if idErr := ValidateIdentifier(k); idErr != nil {
				if err := me.emit(&ComponentValidationError{Section: label, Name: k, Cause: idErr}); err != nil {
					return err
				}
				// Skip validating the component's value when its name is
				// invalid: any leaf error from validate(k) would surface as
				// "<bad-name>: <leaf-error>" and has no resolution path
				// until the name is fixed. The continue keeps the noise
				// per component bounded to a single, actionable finding.
				continue
			}
			wrap := func(e error) error { return &ComponentValidationError{Section: label, Name: k, Cause: e} }
			if err := me.emitWrapped(wrap, validate(k)); err != nil {
				return err
			}
		}
		return nil
	}

	if err := validateMap("schema", components.Schemas.Keys(), func(k string) error {
		return components.Schemas.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("parameter", components.Parameters.Keys(), func(k string) error {
		return components.Parameters.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("request body", components.RequestBodies.Keys(), func(k string) error {
		return components.RequestBodies.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("response", components.Responses.Keys(), func(k string) error {
		return components.Responses.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("header", components.Headers.Keys(), func(k string) error {
		return components.Headers.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("security scheme", components.SecuritySchemes.Keys(), func(k string) error {
		return components.SecuritySchemes.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("example", components.Examples.Keys(), func(k string) error {
		return components.Examples.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("link", components.Links.Keys(), func(k string) error {
		return components.Links.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	if err := validateMap("callback", components.Callbacks.Keys(), func(k string) error {
		return components.Callbacks.Value(k).Validate(ctx)
	}); err != nil {
		return err
	}

	return me.finalize(validateExtensions(ctx, components.Extensions, components.Origin))
}

// refsForCollection iterates the named ComponentRef collection in document
// order. It yields nothing when the name is not one of the collections
// reachable through a $ref (see ComponentRef.CollectionName).
func (components *Components) refsForCollection(name string) iter.Seq2[string, ComponentRef] {
	if components == nil {
		return func(yield func(string, ComponentRef) bool) {}
	}

	switch name {
	case "schemas":
		return asComponentRefs(components.Schemas.Iter())
	case "parameters":
		return asComponentRefs(components.Parameters.Iter())
	case "headers":
		return asComponentRefs(components.Headers.Iter())
	case "requestBodies":
		return asComponentRefs(components.RequestBodies.Iter())
	case "responses":
		return asComponentRefs(components.Responses.Iter())
	case "securitySchemes":
		return asComponentRefs(components.SecuritySchemes.Iter())
	case "examples":
		return asComponentRefs(components.Examples.Iter())
	case "links":
		return asComponentRefs(components.Links.Iter())
	case "callbacks":
		return asComponentRefs(components.Callbacks.Iter())
	}
	return func(yield func(string, ComponentRef) bool) {}
}

func asComponentRefs[V ComponentRef](seq iter.Seq2[string, V]) iter.Seq2[string, ComponentRef] {
	return func(yield func(string, ComponentRef) bool) {
		for k, v := range seq {
			if !yield(k, v) {
				return
			}
		}
	}
}
