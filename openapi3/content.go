package openapi3

import (
	"context"
	"encoding/json"
	"fmt"
	"iter"
	"strings"

	"github.com/go-openapi/jsonpointer"
)

// Content is specified by OpenAPI/Swagger 3.0 standard.
type Content struct {
	m *OrderedMap[string, *MediaType]
}

func NewContent() Content {
	return Content{m: NewOrderedMap[string, *MediaType]()}
}

func NewContentWithCapacity(cap int) Content {
	return Content{m: NewOrderedMapWithCapacity[string, *MediaType](cap)}
}

func ContentFromMap(m map[string]*MediaType) Content {
	content := NewContentWithCapacity(len(m))
	for k, v := range m {
		content.Set(k, v)
	}
	return content
}

func NewContentWithSchema(schema *Schema, consumes []string) Content {
	if len(consumes) == 0 {
		content := NewContentWithCapacity(1)
		content.Set("*/*", NewMediaType().WithSchema(schema))
		return content
	}
	content := NewContentWithCapacity(len(consumes))
	for _, mediaType := range consumes {
		content.Set(mediaType, NewMediaType().WithSchema(schema))
	}
	return content
}

func NewContentWithSchemaRef(schema *SchemaRef, consumes []string) Content {
	if len(consumes) == 0 {
		content := NewContentWithCapacity(1)
		content.Set("*/*", NewMediaType().WithSchemaRef(schema))
		return content
	}
	content := NewContentWithCapacity(len(consumes))
	for _, mediaType := range consumes {
		content.Set(mediaType, NewMediaType().WithSchemaRef(schema))
	}
	return content
}

func NewContentWithJSONSchema(schema *Schema) Content {
	content := NewContentWithCapacity(1)
	content.Set("application/json", NewMediaType().WithSchema(schema))
	return content
}

func NewContentWithJSONSchemaRef(schema *SchemaRef) Content {
	content := NewContentWithCapacity(1)
	content.Set("application/json", NewMediaType().WithSchemaRef(schema))
	return content
}

func NewContentWithFormDataSchema(schema *Schema) Content {
	content := NewContentWithCapacity(1)
	content.Set("multipart/form-data", NewMediaType().WithSchema(schema))
	return content
}

func NewContentWithFormDataSchemaRef(schema *SchemaRef) Content {
	content := NewContentWithCapacity(1)
	content.Set("multipart/form-data", NewMediaType().WithSchemaRef(schema))
	return content
}

// Len returns the number of media types in content.
func (content Content) Len() int {
	if content.m == nil {
		return 0
	}
	return content.m.Len()
}

// Set adds or replaces a media type keyed by mime.
func (content *Content) Set(mime string, value *MediaType) {
	if content.m == nil {
		content.m = NewOrderedMap[string, *MediaType]()
	}
	content.m.Set(mime, value)
}

// Delete removes the media type keyed by mime.
func (content Content) Delete(mime string) {
	if content.m != nil {
		content.m.Delete(mime)
	}
}

// Reorder rearranges existing entries to match order. See OrderedMap.Reorder.
func (content Content) Reorder(order []string) {
	if content.m != nil {
		content.m.Reorder(order)
	}
}

// Value returns the media type keyed by the exact mime given, or nil.
func (content Content) Value(mime string) *MediaType {
	if content.m == nil {
		return nil
	}
	return content.m.Value(mime)
}

// Keys returns the content's mime types in insertion order.
func (content Content) Keys() []string {
	if content.m == nil {
		return nil
	}
	return content.m.Keys()
}

// Iter returns an iterator over the content's mime types in insertion order.
func (content Content) Iter() iter.Seq2[string, *MediaType] {
	if content.m == nil {
		return func(yield func(string, *MediaType) bool) {}
	}
	return content.m.Iter()
}

// Map returns content as a 'map'.
// Note: iteration on Go maps is not ordered.
func (content Content) Map() map[string]*MediaType {
	if content.m == nil {
		return make(map[string]*MediaType)
	}
	return content.m.Map()
}

// Get resolves the media type most applicable to mime, falling back through
// the x/y, x/*, and */* wildcard patterns as per the OpenAPI spec.
func (content Content) Get(mime string) *MediaType {
	// If the mime is empty then short-circuit to the wildcard.
	// We do this here so that we catch only the specific case of
	// and empty mime rather than a present, but invalid, mime type.
	if mime == "" {
		return content.Value("*/*")
	}
	// Start by making the most specific match possible
	// by using the mime type in full.
	if v := content.Value(mime); v != nil {
		return v
	}
	// If an exact match is not found then we strip all
	// metadata from the mime type and only use the x/y
	// portion.
	i := strings.IndexByte(mime, ';')
	if i < 0 {
		// If there is no metadata then preserve the full mime type
		// string for later wildcard searches.
		i = len(mime)
	}
	mime = mime[:i]
	if v := content.Value(mime); v != nil {
		return v
	}
	// If the x/y pattern has no specific match then we
	// try the x/* pattern.
	i = strings.IndexByte(mime, '/')
	if i < 0 {
		// In the case that the given mime type is not valid because it is
		// missing the subtype we return nil so that this does not accidentally
		// resolve with the wildcard.
		return nil
	}
	mime = mime[:i] + "/*"
	if v := content.Value(mime); v != nil {
		return v
	}
	// Finally, the most generic match of */* is returned
	// as a catch-all.
	return content.Value("*/*")
}

// Validate returns an error if Content does not comply with the OpenAPI spec.
func (content Content) Validate(ctx context.Context, opts ...ValidationOption) error {
	ctx = WithValidationOptions(ctx, opts...)

	for _, k := range content.Keys() {
		if err := content.Value(k).Validate(ctx); err != nil {
			return err
		}
	}
	return nil
}

// MarshalJSON returns the JSON encoding of Content, preserving key order.
func (content Content) MarshalJSON() ([]byte, error) {
	if content.m == nil {
		return []byte("{}"), nil
	}
	return content.m.MarshalJSON()
}

// MarshalYAML returns the YAML encoding of Content.
func (content Content) MarshalYAML() (any, error) {
	m := make(map[string]any, content.Len())
	for k, v := range content.Iter() {
		m[k] = v
	}
	return m, nil
}

// UnmarshalJSON sets Content to a copy of data.
func (content *Content) UnmarshalJSON(data []byte) error {
	if content.m == nil {
		content.m = NewOrderedMap[string, *MediaType]()
	}
	return unmarshalJSONWithOrder(data, func(k string, v json.RawMessage) error {
		var vv MediaType
		if err := json.Unmarshal(v, &vv); err != nil {
			return err
		}
		content.m.Set(k, &vv)
		return nil
	})
}

var _ jsonpointer.JSONPointable = (*Content)(nil)

// JSONLookup implements https://pkg.go.dev/github.com/go-openapi/jsonpointer#JSONPointable
func (content Content) JSONLookup(token string) (any, error) {
	v := content.Value(token)
	if v == nil {
		return nil, fmt.Errorf("no content for mime %q", token)
	}
	return v, nil
}
