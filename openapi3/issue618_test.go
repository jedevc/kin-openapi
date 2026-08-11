package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestIssue618(t *testing.T) {
	spec := `
openapi: 3.0.0
info:
  title: foo
  version: 0.0.0
paths:
  /foo:
    get:
      responses:
        '200':
          description: Some description value text
          content:
            application/json:
              schema:
                $ref: ./testdata/schema618.yml#/components/schemas/JournalEntry
`[1:]

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	ctx := loader.Context

	doc, err := loader.LoadFromData([]byte(spec))
	require.NoError(t, err)

	doc.InternalizeRefs(ctx, nil)

	_, ok1 := doc.Components.Schemas.Get("testdata_schema618_JournalEntry")
	require.True(t, ok1, "expected schema testdata_schema618_JournalEntry to exist")
	_, ok2 := doc.Components.Schemas.Get("testdata_schema618_Record")
	require.True(t, ok2, "expected schema testdata_schema618_Record to exist")
	_, ok3 := doc.Components.Schemas.Get("testdata_schema618_Account")
	require.True(t, ok3, "expected schema testdata_schema618_Account to exist")
}
