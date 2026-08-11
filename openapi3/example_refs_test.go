package openapi3_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestParameterExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	param := doc.Paths.Value("/test").Post.Parameters[0].Value
	require.NotNil(t, param, "Parameter should not be nil")
	require.NotNil(t, param.Examples.Value("Test").Value, "Parameter example should not be nil")
	require.Equal(t, "An example", param.Examples.Value("Test").Value.Summary)
}

func TestParameterExampleWithContentRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	param := doc.Paths.Value("/test").Post.Parameters[1].Value
	require.NotNil(t, param, "Parameter should not be nil")
	require.NotNil(t, param.Content.Value("application/json").Examples.Value("Test").Value, "Parameter example should not be nil")
	require.Equal(t, "An example", param.Content.Value("application/json").Examples.Value("Test").Value.Summary)
}

func TestRequestBodyExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	requestBody := doc.Paths.Value("/test").Post.RequestBody.Value
	require.NotNil(t, requestBody, "Request body should not be nil")
	require.NotNil(t, requestBody.Content.Value("application/json").Examples.Value("Test").Value, "Request body example should not be nil")
	require.Equal(t, "An example", requestBody.Content.Value("application/json").Examples.Value("Test").Value.Summary)
}

func TestResponseExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	response := doc.Paths.Value("/test").Post.Responses.Value("200").Value
	require.NotNil(t, response, "Response should not be nil")
	require.NotNil(t, response.Content.Value("application/json").Examples.Value("Test").Value, "Response example should not be nil")
	require.Equal(t, "An example", response.Content.Value("application/json").Examples.Value("Test").Value.Summary)
}

func TestHeaderExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	response := doc.Paths.Value("/test").Post.Responses.Value("200").Value
	header := response.Headers.Value("X-Test-Header").Value
	require.NotNil(t, header, "Header should not be nil")
	require.NotNil(t, header.Examples.Value("Test").Value, "Header example should not be nil")
	require.Equal(t, "An example", header.Examples.Value("Test").Value.Summary)
}

func TestComponentExampleRef(t *testing.T) {
	loader := openapi3.NewLoader()
	doc, err := loader.LoadFromFile("testdata/example_refs.yml")
	require.NoError(t, err)
	err = doc.Validate(loader.Context)
	require.NoError(t, err)
	example := doc.Components.Examples.Value("RefExample").Value
	require.NotNil(t, example, "Example should not be nil")
	require.Equal(t, "An example", example.Summary)
}
