package openapi3_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/getkin/kin-openapi/openapi3"
)

func TestWebhooksField(t *testing.T) {
	t.Run("serialize webhooks in OpenAPI 3.1", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: openapi3.WebhooksFromMap(map[string]*openapi3.PathItem{
				"newPet": {
					Post: &openapi3.Operation{
						Summary: "New pet webhook",
						Responses: openapi3.NewResponses(
							openapi3.WithStatus(200, &openapi3.ResponseRef{
								Value: &openapi3.Response{
									Description: openapi3.Ptr("Success"),
								},
							}),
						),
					},
				},
			}),
		}

		data, err := json.Marshal(doc)
		require.NoError(t, err)

		// Should contain webhooks
		require.Contains(t, string(data), `"webhooks"`)
		require.Contains(t, string(data), `"newPet"`)
	})

	t.Run("deserialize webhooks from OpenAPI 3.1", func(t *testing.T) {
		jsonData := []byte(`{
			"openapi": "3.1.0",
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			},
			"paths": {},
			"webhooks": {
				"newPet": {
					"post": {
						"summary": "New pet webhook",
						"responses": {
							"200": {
								"description": "Success"
							}
						}
					}
				}
			}
		}`)

		var doc openapi3.T
		err := json.Unmarshal(jsonData, &doc)
		require.NoError(t, err)

		require.True(t, doc.IsOpenAPI31OrLater())
		require.NotZero(t, doc.Webhooks.Len())
		require.NotNil(t, doc.Webhooks.Value("newPet"))
		require.NotNil(t, doc.Webhooks.Value("newPet").Post)
		require.Equal(t, "New pet webhook", doc.Webhooks.Value("newPet").Post.Summary)
	})

	t.Run("OpenAPI 3.0 without webhooks", func(t *testing.T) {
		jsonData := []byte(`{
			"openapi": "3.0.3",
			"info": {
				"title": "Test API",
				"version": "1.0.0"
			},
			"paths": {}
		}`)

		var doc openapi3.T
		err := json.Unmarshal(jsonData, &doc)
		require.NoError(t, err)

		require.True(t, doc.IsOpenAPI30())
		require.Zero(t, doc.Webhooks.Len())
	})

	t.Run("validate webhooks", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: openapi3.WebhooksFromMap(map[string]*openapi3.PathItem{
				"validWebhook": {
					Post: &openapi3.Operation{
						Responses: openapi3.NewResponses(
							openapi3.WithStatus(200, &openapi3.ResponseRef{
								Value: &openapi3.Response{
									Description: openapi3.Ptr("Success"),
								},
							}),
						),
					},
				},
			}),
		}

		// Should validate successfully
		err := doc.Validate(t.Context())
		require.NoError(t, err)
	})

	t.Run("validate fails with nil webhook", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: openapi3.WebhooksFromMap(map[string]*openapi3.PathItem{
				"invalidWebhook": nil,
			}),
		}

		err := doc.Validate(t.Context())
		require.Error(t, err)
		require.ErrorContains(t, err, "webhook")
		require.ErrorContains(t, err, "invalidWebhook")
	})
}

func TestJSONLookupWithWebhooks(t *testing.T) {
	doc := &openapi3.T{
		OpenAPI: "3.1.0",
		Info: &openapi3.Info{
			Title:   "Test API",
			Version: "1.0.0",
		},
		Paths: openapi3.NewPaths(),
		Webhooks: openapi3.WebhooksFromMap(map[string]*openapi3.PathItem{
			"test": {
				Post: &openapi3.Operation{
					Summary: "Test webhook",
				},
			},
		}),
	}

	result, err := doc.JSONLookup("webhooks")
	require.NoError(t, err)
	require.NotNil(t, result)

	webhooks, ok := result.(openapi3.Webhooks)
	require.True(t, ok)
	require.NotNil(t, webhooks.Value("test"))
}

func TestVersionBasedBehavior(t *testing.T) {
	t.Run("detect and handle OpenAPI 3.0", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.0.3",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
		}

		if doc.IsOpenAPI30() {
			// OpenAPI 3.0 specific logic
			require.Zero(t, doc.Webhooks.Len())
		}
	})

	t.Run("detect and handle OpenAPI 3.1", func(t *testing.T) {
		doc := &openapi3.T{
			OpenAPI: "3.1.0",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
			Webhooks: openapi3.WebhooksFromMap(map[string]*openapi3.PathItem{
				"test": {
					Post: &openapi3.Operation{
						Summary: "Test",
						Responses: openapi3.NewResponses(
							openapi3.WithStatus(200, &openapi3.ResponseRef{
								Value: &openapi3.Response{
									Description: openapi3.Ptr("OK"),
								},
							}),
						),
					},
				},
			}),
		}

		if doc.IsOpenAPI31OrLater() {
			// OpenAPI 3.1 specific logic
			require.NotZero(t, doc.Webhooks.Len())
			require.NotNil(t, doc.Webhooks.Value("test"))
		}
	})
}

func TestMigrationScenario(t *testing.T) {
	t.Run("upgrade document from 3.0 to 3.1", func(t *testing.T) {
		// Start with 3.0 document
		doc := &openapi3.T{
			OpenAPI: "3.0.3",
			Info: &openapi3.Info{
				Title:   "Test API",
				Version: "1.0.0",
			},
			Paths: openapi3.NewPaths(),
		}

		require.True(t, doc.IsOpenAPI30())
		require.Zero(t, doc.Webhooks.Len())

		// Upgrade to 3.1
		doc.OpenAPI = "3.1.0"

		// Add 3.1 features
		doc.Webhooks = openapi3.WebhooksFromMap(map[string]*openapi3.PathItem{
			"newEvent": {
				Post: &openapi3.Operation{
					Summary: "New event notification",
					Responses: openapi3.NewResponses(
						openapi3.WithStatus(200, &openapi3.ResponseRef{
							Value: &openapi3.Response{
								Description: openapi3.Ptr("Processed"),
							},
						}),
					),
				},
			},
		})

		require.True(t, doc.IsOpenAPI31OrLater())
		require.NotZero(t, doc.Webhooks.Len())

		// Validate the upgraded document
		err := doc.Validate(t.Context())
		require.NoError(t, err)
	})
}
