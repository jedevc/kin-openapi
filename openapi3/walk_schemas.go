package openapi3

import (
	"errors"
	"slices"
	"strconv"
	"strings"
)

// SkipSubtree, when returned by a WalkSchemasFunc, tells WalkSchemas not to
// descend into the current schema's sub-schemas. The schema itself has already
// been visited. Any other non-nil error stops the walk and is returned by
// WalkSchemas. It mirrors filepath.SkipDir.
var SkipSubtree = errors.New("skip schema subtree")

// WalkSchemasFunc is called once for each schema visited by WalkSchemas.
//
// jsonPointer is the RFC 6901 JSON Pointer of the schema within the document,
// e.g. "/components/schemas/Pet/properties/tag" or
// "/paths/~1pets/get/responses/200/content/application~1json/schema". It is
// file-agnostic: once the loader has resolved external $refs the document is a
// single tree, so the pointer addresses a position in that tree and never
// encodes a source file. schema is non-nil and schema.Value is non-nil; the
// callback may modify schema.Value in place, so WalkSchemas serves transformers
// and not only read-only inspection.
//
// Returning SkipSubtree skips this schema's sub-schemas; returning any other
// error aborts the walk and is returned by WalkSchemas.
type WalkSchemasFunc func(jsonPointer string, schema *SchemaRef) error

// WalkSchemas visits every schema reachable from the document exactly once,
// invoking fn for each. It follows resolved $ref targets (schema.Value) and
// guards against reference cycles, so each distinct *Schema is visited a single
// time regardless of how many references point at it. Maps are visited in
// sorted key order, so the traversal is deterministic.
//
// It covers schemas under components (schemas, parameters, headers, request
// bodies, responses, callbacks), the paths and their operations (parameters,
// request bodies, responses, headers, callbacks), and webhooks, then recurses
// through every sub-schema keyword: properties, items, itemSchema,
// allOf/anyOf/oneOf, not, additionalProperties, prefixItems, contains, patternProperties,
// dependentSchemas, propertyNames, if/then/else, and $defs.
//
// It is useful for validation, code generation, schema transformation,
// $ref/dependency analysis, and documentation: any consumer that needs to act
// on every schema in a document without re-deriving the (easy to get wrong)
// traversal itself.
func (doc *T) WalkSchemas(fn WalkSchemasFunc) error {
	if doc == nil {
		return nil
	}
	w := schemaWalker{fn: fn, seen: make(map[*Schema]struct{})}
	return w.document(doc)
}

type schemaWalker struct {
	fn   WalkSchemasFunc
	seen map[*Schema]struct{}
}

// escapeRefString escapes a single JSON Pointer reference token per RFC 6901:
// '~' becomes '~0' and '/' becomes '~1'. It is the inverse of unescapeRefString.
func escapeRefString(s string) string {
	if !strings.ContainsAny(s, "~/") {
		return s
	}
	return strings.NewReplacer("~", "~0", "/", "~1").Replace(s)
}

// sortedKeys returns keys in sorted order, for the ordered-map types.
func sortedKeys(keys []string) []string {
	return slices.Sorted(slices.Values(keys))
}

func (w *schemaWalker) document(doc *T) error {
	if c := doc.Components; c != nil {
		for _, name := range sortedKeys(c.Schemas.Keys()) {
			if err := w.schemaRef("/components/schemas/"+escapeRefString(name), c.Schemas.Value(name)); err != nil {
				return err
			}
		}
		for _, name := range sortedKeys(c.Parameters.Keys()) {
			if err := w.parameter("/components/parameters/"+escapeRefString(name), c.Parameters.Value(name)); err != nil {
				return err
			}
		}
		for _, name := range sortedKeys(c.Headers.Keys()) {
			if err := w.header("/components/headers/"+escapeRefString(name), c.Headers.Value(name)); err != nil {
				return err
			}
		}
		for _, name := range sortedKeys(c.RequestBodies.Keys()) {
			if rbr := c.RequestBodies.Value(name); rbr != nil && rbr.Value != nil {
				if err := w.content("/components/requestBodies/"+escapeRefString(name)+"/content", rbr.Value.Content); err != nil {
					return err
				}
			}
		}
		for _, name := range sortedKeys(c.Responses.Keys()) {
			if err := w.response("/components/responses/"+escapeRefString(name), c.Responses.Value(name)); err != nil {
				return err
			}
		}
		for _, name := range sortedKeys(c.Callbacks.Keys()) {
			if cbr := c.Callbacks.Value(name); cbr != nil && cbr.Value != nil {
				if err := w.callback("/components/callbacks/"+escapeRefString(name), cbr.Value); err != nil {
					return err
				}
			}
		}
	}
	if doc.Paths != nil {
		for _, path := range sortedKeys(doc.Paths.Keys()) {
			if err := w.pathItem("/paths/"+escapeRefString(path), doc.Paths.Value(path)); err != nil {
				return err
			}
		}
	}
	for _, name := range sortedKeys(doc.Webhooks.Keys()) {
		if err := w.pathItem("/webhooks/"+escapeRefString(name), doc.Webhooks.Value(name)); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) pathItem(ptr string, item *PathItem) error {
	if item == nil {
		return nil
	}
	for i, pr := range item.Parameters {
		if err := w.parameter(ptr+"/parameters/"+strconv.Itoa(i), pr); err != nil {
			return err
		}
	}
	for _, entry := range item.operationEntries() {
		if err := w.operation(ptr+entry.pointerSuffix, entry.operation); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) operation(ptr string, op *Operation) error {
	if op == nil {
		return nil
	}
	for i, pr := range op.Parameters {
		if err := w.parameter(ptr+"/parameters/"+strconv.Itoa(i), pr); err != nil {
			return err
		}
	}
	if op.RequestBody != nil && op.RequestBody.Value != nil {
		if err := w.content(ptr+"/requestBody/content", op.RequestBody.Value.Content); err != nil {
			return err
		}
	}
	if op.Responses != nil {
		for _, code := range sortedKeys(op.Responses.Keys()) {
			if err := w.response(ptr+"/responses/"+escapeRefString(code), op.Responses.Value(code)); err != nil {
				return err
			}
		}
	}
	for _, name := range sortedKeys(op.Callbacks.Keys()) {
		if cbr := op.Callbacks.Value(name); cbr != nil && cbr.Value != nil {
			if err := w.callback(ptr+"/callbacks/"+escapeRefString(name), cbr.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

func (w *schemaWalker) callback(ptr string, cb *Callback) error {
	for _, expr := range sortedKeys(cb.Keys()) {
		if err := w.pathItem(ptr+"/"+escapeRefString(expr), cb.Value(expr)); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) parameter(ptr string, pr *ParameterRef) error {
	if pr == nil || pr.Value == nil {
		return nil
	}
	if err := w.schemaRef(ptr+"/schema", pr.Value.Schema); err != nil {
		return err
	}
	return w.content(ptr+"/content", pr.Value.Content)
}

func (w *schemaWalker) header(ptr string, hr *HeaderRef) error {
	if hr == nil || hr.Value == nil {
		return nil
	}
	if err := w.schemaRef(ptr+"/schema", hr.Value.Schema); err != nil {
		return err
	}
	return w.content(ptr+"/content", hr.Value.Content)
}

func (w *schemaWalker) response(ptr string, rr *ResponseRef) error {
	if rr == nil || rr.Value == nil {
		return nil
	}
	for _, name := range sortedKeys(rr.Value.Headers.Keys()) {
		if err := w.header(ptr+"/headers/"+escapeRefString(name), rr.Value.Headers.Value(name)); err != nil {
			return err
		}
	}
	return w.content(ptr+"/content", rr.Value.Content)
}

func (w *schemaWalker) content(ptr string, content Content) error {
	for _, mediaType := range sortedKeys(content.Keys()) {
		media := content.Value(mediaType)
		if media == nil {
			continue
		}
		if err := w.schemaRef(ptr+"/"+escapeRefString(mediaType)+"/schema", media.Schema); err != nil {
			return err
		}
		if err := w.schemaRef(ptr+"/"+escapeRefString(mediaType)+"/itemSchema", media.ItemSchema); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) schemaRefs(ptr string, refs SchemaRefs) error {
	for i, sub := range refs {
		if err := w.schemaRef(ptr+"/"+strconv.Itoa(i), sub); err != nil {
			return err
		}
	}
	return nil
}

func (w *schemaWalker) schemaRef(ptr string, sr *SchemaRef) error {
	if sr == nil || sr.Value == nil {
		return nil
	}
	s := sr.Value
	if _, ok := w.seen[s]; ok {
		// Already visited (shared $ref target or reference cycle).
		return nil
	}
	w.seen[s] = struct{}{}

	if err := w.fn(ptr, sr); err != nil {
		if errors.Is(err, SkipSubtree) {
			return nil
		}
		return err
	}

	for _, name := range sortedKeys(s.Properties.Keys()) {
		if err := w.schemaRef(ptr+"/properties/"+escapeRefString(name), s.Properties.Value(name)); err != nil {
			return err
		}
	}
	if err := w.schemaRef(ptr+"/items", s.Items); err != nil {
		return err
	}
	if s.AdditionalProperties.Schema != nil {
		if err := w.schemaRef(ptr+"/additionalProperties", s.AdditionalProperties.Schema); err != nil {
			return err
		}
	}
	if err := w.schemaRefs(ptr+"/allOf", s.AllOf); err != nil {
		return err
	}
	if err := w.schemaRefs(ptr+"/anyOf", s.AnyOf); err != nil {
		return err
	}
	if err := w.schemaRefs(ptr+"/oneOf", s.OneOf); err != nil {
		return err
	}
	if err := w.schemaRef(ptr+"/not", s.Not); err != nil {
		return err
	}
	if err := w.schemaRefs(ptr+"/prefixItems", s.PrefixItems); err != nil {
		return err
	}
	if err := w.schemaRef(ptr+"/contains", s.Contains); err != nil {
		return err
	}
	for _, name := range sortedKeys(s.PatternProperties.Keys()) {
		if err := w.schemaRef(ptr+"/patternProperties/"+escapeRefString(name), s.PatternProperties.Value(name)); err != nil {
			return err
		}
	}
	for _, name := range sortedKeys(s.DependentSchemas.Keys()) {
		if err := w.schemaRef(ptr+"/dependentSchemas/"+escapeRefString(name), s.DependentSchemas.Value(name)); err != nil {
			return err
		}
	}
	if err := w.schemaRef(ptr+"/propertyNames", s.PropertyNames); err != nil {
		return err
	}
	if err := w.schemaRef(ptr+"/if", s.If); err != nil {
		return err
	}
	if err := w.schemaRef(ptr+"/then", s.Then); err != nil {
		return err
	}
	if err := w.schemaRef(ptr+"/else", s.Else); err != nil {
		return err
	}
	for _, name := range sortedKeys(s.Defs.Keys()) {
		if err := w.schemaRef(ptr+"/$defs/"+escapeRefString(name), s.Defs.Value(name)); err != nil {
			return err
		}
	}
	return nil
}
