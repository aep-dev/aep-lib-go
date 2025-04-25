package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

var basicOpenAPI = &openapi.OpenAPI{
	OpenAPI: "3.1.0",
	Info: openapi.Info{
		Contact: openapi.Contact{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			URL:   "https://example.com",
		},
	},
	Servers: []openapi.Server{{URL: "https://api.example.com"}},
	Paths: map[string]*openapi.PathItem{
		"/widgets": {
			Get: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Properties: map[string]openapi.Schema{
										"results": {
											Type: "array",
											Items: &openapi.Schema{
												Ref: "#/components/schemas/Widget",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Post: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
		},
		"/widgets/{widget}": {
			Get: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
			Delete: &openapi.Operation{},
			Patch: &openapi.Operation{
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
		},
		"/widgets/{widget}:start": {
			Post: &openapi.Operation{
				RequestBody: &openapi.RequestBody{
					Content: map[string]openapi.MediaType{
						"application/json": {
							Schema: &openapi.Schema{
								Type: "object",
								Properties: map[string]openapi.Schema{
									"foo": {Type: "string"},
									"bar": {Type: "integer"},
									"baz": {
										Type: "array",
										Items: &openapi.Schema{
											Type: "boolean",
										},
									},
								},
							},
						},
					},
				},
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
		},
		"/widgets/{widget}:stop": {
			Post: &openapi.Operation{
				RequestBody: &openapi.RequestBody{
					Content: map[string]openapi.MediaType{
						"application/json": {
							Schema: &openapi.Schema{
								Ref: "#/components/schemas/Widget",
							},
						},
					},
				},
				Responses: map[string]openapi.Response{
					"200": {
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{
									Ref: "#/components/schemas/Widget",
								},
							},
						},
					},
				},
			},
		},
	},
	Components: openapi.Components{
		Schemas: map[string]openapi.Schema{
			"Widget": {
				Type: "object",
				Properties: map[string]openapi.Schema{
					"name": {Type: "string"},
				},
			},
			"Account": {
				Type: "object",
				Properties: map[string]openapi.Schema{
					"title": {Type: "string"},
				},
			},
		},
	},
}

func TestGetAPI(t *testing.T) {
	tests := []struct {
		name           string
		api            *openapi.OpenAPI
		serverURL      string
		expectedError  string
		validateResult func(*testing.T, *API)
	}{
		{
			name: "basic resource with CRUD operations",
			api:  basicOpenAPI,
			validateResult: func(t *testing.T, sd *API) {
				assert.Equal(t, "https://api.example.com", sd.ServerURL)

				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.Equal(t, widget.PatternElems, []string{"widgets", "{widget}"})
				assert.Equal(t, sd.ServerURL, "https://api.example.com")
				assert.NotNil(t, widget.Methods.Get, "should have GET method")
				assert.NotNil(t, widget.Methods.List, "should have LIST method")
				assert.NotNil(t, widget.Methods.Create, "should have CREATE method")
				if widget.Methods.Create != nil {
					assert.False(t, widget.Methods.Create.SupportsUserSettableCreate, "should not support user-settable create")
				}
				assert.NotNil(t, widget.Methods.Update, "should have UPDATE method")
				assert.NotNil(t, widget.Methods.Delete, "should have DELETE method")
			},
		},
		{
			name: "non-resource schemas",
			api:  basicOpenAPI,
			validateResult: func(t *testing.T, sd *API) {
				assert.Contains(t, sd.Schemas, "Account", "should have Account schema")
			},
		},

		{
			name:      "empty openapi with server url override",
			api:       basicOpenAPI,
			serverURL: "https://override.example.com",
			validateResult: func(t *testing.T, sd *API) {
				assert.Equal(t, "https://override.example.com", sd.ServerURL)
			},
		},
		{
			name: "resource with x-aep-resource annotation",
			api: &openapi.OpenAPI{
				OpenAPI: "3.1.0",
				Paths: map[string]*openapi.PathItem{
					"/widgets/{widget}": {
						Get: &openapi.Operation{
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "#/components/schemas/widget",
											},
										},
									},
								},
							},
						},
					},
				},
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Components: openapi.Components{
					Schemas: map[string]openapi.Schema{
						"widget": {
							Type: "object",
							Properties: map[string]openapi.Schema{
								"name": {Type: "string"},
							},
							XAEPResource: &openapi.XAEPResource{
								Singular: "widget",
								Plural:   "widgets",
								Patterns: []string{"/widgets/{widget}"},
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *API) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.Equal(t, "widget", widget.Singular)
				assert.Equal(t, "widgets", widget.Plural)
				assert.Equal(t, []string{"widgets", "{widget}"}, widget.PatternElems)
			},
		},
		{
			name: "missing server URL",
			api: &openapi.OpenAPI{
				OpenAPI: "3.1.0",
				Servers: []openapi.Server{},
			},
			expectedError: "no server URL found in openapi, and none was provided",
		},
		{
			name: "resource with user-settable create ID",
			api: &openapi.OpenAPI{
				OpenAPI: "3.1.0",
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Paths: map[string]*openapi.PathItem{
					"/widgets": {
						Post: &openapi.Operation{
							Parameters: []openapi.Parameter{
								{Name: "id"},
							},
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "#/components/schemas/Widget",
											},
										},
									},
								},
							},
						},
					},
				},
				Components: openapi.Components{
					Schemas: map[string]openapi.Schema{
						"Widget": {
							Type: "object",
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *API) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.True(t, widget.Methods.Create.SupportsUserSettableCreate,
					"should support user-settable create")
			},
		},
		{
			name: "OAS 2.0 style schema in response",
			api: &openapi.OpenAPI{
				Swagger: "2.0",
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Paths: map[string]*openapi.PathItem{
					"/widgets/{widget}": {
						Get: &openapi.Operation{
							Responses: map[string]openapi.Response{
								"200": {
									Schema: &openapi.Schema{
										Ref: "#/definitions/Widget",
									},
								},
							},
						},
					},
				},
				Definitions: map[string]openapi.Schema{
					"Widget": {
						Type: "object",
						Properties: map[string]openapi.Schema{
							"name": {Type: "string"},
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *API) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.NotNil(t, widget.Methods.Get, "should have GET method")
				assert.Equal(t, []string{"widgets", "{widget}"}, widget.PatternElems)
			},
		},
		{
			name: "resource with custom methods",
			api:  basicOpenAPI,
			validateResult: func(t *testing.T, sd *API) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")

				assert.Len(t, widget.CustomMethods, 2, "should have 2 custom methods")
				for _, m := range widget.CustomMethods {
					assert.Contains(t, []string{"start", "stop"}, m.Name)
					assert.Equal(t, "POST", m.Method)
					if m.Name == "start" {
						assert.Equal(t, "object", m.Request.Type)
					}
				}
			},
		},
		{
			name: "list method with skip and unreachable flags",
			api: &openapi.OpenAPI{
				OpenAPI: "3.1.0",
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Paths: map[string]*openapi.PathItem{
					"/widgets": {
						Get: &openapi.Operation{
							Parameters: []openapi.Parameter{
								{Name: "skip"},
								{Name: "unreachable", Schema: &openapi.Schema{Type: "boolean"}},
							},
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Properties: map[string]openapi.Schema{
													"results": {
														Type: "array",
														Items: &openapi.Schema{
															Ref: "#/components/schemas/Widget",
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Components: openapi.Components{
					Schemas: map[string]openapi.Schema{
						"Widget": {
							Type: "object",
							Properties: map[string]openapi.Schema{
								"name": {Type: "string"},
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *API) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")
				assert.NotNil(t, widget.Methods.List, "should have LIST method")
				assert.True(t, widget.Methods.List.SupportsSkip, "should support skip parameter")
				assert.True(t, widget.Methods.List.HasUnreachableResources, "should support unreachable parameter")
			},
		},
		{
			name: "contact information",
			api:  basicOpenAPI,
			validateResult: func(t *testing.T, sd *API) {
				assert.Equal(t, "John Doe", sd.Contact.Name)
				assert.Equal(t, "john.doe@example.com", sd.Contact.Email)
				assert.Equal(t, "https://example.com", sd.Contact.URL)
			},
		},
		{
			name: "resource with long running operations",
			api: &openapi.OpenAPI{
				OpenAPI: "3.1.0",
				Servers: []openapi.Server{{URL: "https://api.example.com"}},
				Paths: map[string]*openapi.PathItem{
					"/widgets": {
						Post: &openapi.Operation{
							XAEPLongRunningOperation: &openapi.XAEPLongRunningOperation{
								Response: openapi.XAEPLongRunningOperationResponse{
									Schema: &openapi.Schema{
										Ref: "#/components/schemas/Widget",
									},
								},
							},
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "//aep.dev/json-schema/type/operation.json",
											},
										},
									},
								},
							},
						},
					},
					"/widgets/{widget}": {
						Get: &openapi.Operation{
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "#/components/schemas/Widget",
											},
										},
									},
								},
							},
						},
						Delete: &openapi.Operation{
							XAEPLongRunningOperation: &openapi.XAEPLongRunningOperation{
								Response: openapi.XAEPLongRunningOperationResponse{
									Schema: &openapi.Schema{
										Ref: "#/components/schemas/Widget",
									},
								},
							},
						},
						Patch: &openapi.Operation{
							XAEPLongRunningOperation: &openapi.XAEPLongRunningOperation{
								Response: openapi.XAEPLongRunningOperationResponse{
									Schema: &openapi.Schema{
										Ref: "#/components/schemas/Widget",
									},
								},
							},
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "//aep.dev/json-schema/type/operation.json",
											},
										},
									},
								},
							},
						},
					},
					"/widgets/{widget}:customOp": {
						Post: &openapi.Operation{
							XAEPLongRunningOperation: &openapi.XAEPLongRunningOperation{
								Response: openapi.XAEPLongRunningOperationResponse{
									Schema: &openapi.Schema{
										Ref: "#/components/schemas/Widget",
									},
								},
							},
							RequestBody: &openapi.RequestBody{
								Content: map[string]openapi.MediaType{
									"application/json": {
										Schema: &openapi.Schema{
											Type: "object",
										},
									},
								},
							},
							Responses: map[string]openapi.Response{
								"200": {
									Content: map[string]openapi.MediaType{
										"application/json": {
											Schema: &openapi.Schema{
												Ref: "//aep.dev/json-schema/type/operation.json",
											},
										},
									},
								},
							},
						},
					},
				},
				Components: openapi.Components{
					Schemas: map[string]openapi.Schema{
						"Widget": {
							Type: "object",
							Properties: map[string]openapi.Schema{
								"name": {Type: "string"},
							},
						},
					},
				},
			},
			validateResult: func(t *testing.T, sd *API) {
				widget, ok := sd.Resources["widget"]
				assert.True(t, ok, "widget resource should exist")

				// Check create method is marked as long running
				assert.NotNil(t, widget.Methods.Create, "should have CREATE method")
				assert.True(t, widget.Methods.Create.IsLongRunning, "CREATE method should be marked as long running")

				// Check update method is marked as long running
				assert.NotNil(t, widget.Methods.Update, "should have UPDATE method")
				assert.True(t, widget.Methods.Update.IsLongRunning, "UPDATE method should be marked as long running")

				// Check delete method is marked as long running
				assert.NotNil(t, widget.Methods.Delete, "should have DELETE method")
				assert.True(t, widget.Methods.Delete.IsLongRunning, "DELETE method should be marked as long running")

				// Check custom method is marked as long running
				assert.NotEmpty(t, widget.CustomMethods, "should have custom methods")
				customOp := widget.CustomMethods[0]
				assert.Equal(t, "customOp", customOp.Name, "should have customOp method")
				assert.True(t, customOp.IsLongRunning, "customOp method should be marked as long running")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetAPI(tt.api, tt.serverURL, "")

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, result)

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestParseBookstoreYAMLDirectly(t *testing.T) {
	// Construct the path relative to the test file's location
	yamlPath := filepath.Join("..", "..", "examples", "resource-definitions", "bookstore.yaml")

	// Read the YAML file
	yamlData, err := os.ReadFile(yamlPath)
	require.NoError(t, err, "Failed to read bookstore.yaml")
	require.NotEmpty(t, yamlData, "bookstore.yaml is empty")

	// Unmarshal YAML into a generic interface{}
	var genericYamlData interface{}
	err = yaml.Unmarshal(yamlData, &genericYamlData)
	require.NoError(t, err, "Failed to unmarshal YAML into generic interface")

	// Marshal the generic interface{} to JSON
	jsonData, err := json.Marshal(genericYamlData)
	require.NoError(t, err, "Failed to marshal generic interface to JSON")
	require.NotEmpty(t, jsonData, "Resulting JSON data is empty")

	// Attempt to unmarshal the JSON directly into api.API
	var apiResult API
	err = json.Unmarshal(jsonData, &apiResult)
	require.NoError(t, err, "Failed to unmarshal JSON into api.API struct")

	// Assert basic fields that might match (like Name, ServerURL, Contact)
	assert.Equal(t, "bookstore.example.com", apiResult.Name, "API Name should be populated if field names match")
	assert.Equal(t, "http://localhost:8081", apiResult.ServerURL, "API ServerURL should be populated based on json tag")
	if assert.NotNil(t, apiResult.Contact, "Contact might be populated if fields match") {
		assert.Equal(t, "API support", apiResult.Contact.Name)
		assert.Equal(t, "aepsupport@aep.dev", apiResult.Contact.Email)
	}

	// Assert that Resources map IS populated correctly
	assert.NotEmpty(t, apiResult.Resources, "Resources map should be populated")
	assert.Contains(t, apiResult.Resources, "publisher", "Resources map should contain 'publisher'")
	assert.Contains(t, apiResult.Resources, "book", "Resources map should contain 'book'")
	assert.Contains(t, apiResult.Resources, "book-edition", "Resources map should contain 'book-edition'")
	assert.Contains(t, apiResult.Resources, "isbn", "Resources map should contain 'isbn'")

	// Check some details of a resource
	publisherResource := apiResult.Resources["publisher"]
	assert.NotNil(t, publisherResource, "'publisher' resource should not be nil")
	assert.Equal(t, "publisher", publisherResource.Singular)
	assert.Equal(t, "publishers", publisherResource.Plural)
	assert.NotNil(t, publisherResource.Schema, "'publisher' resource schema should not be nil")
	assert.Equal(t, "object", publisherResource.Schema.Type)
	assert.Contains(t, publisherResource.Schema.Properties, "description")
	assert.Equal(t, "string", publisherResource.Schema.Properties["description"].Type)
	assert.NotNil(t, publisherResource.Methods.List, "'publisher' should have List method")
	assert.True(t, publisherResource.Methods.List.SupportsFilter)

	// Check book resource details, including custom method
	bookResource := apiResult.Resources["book"]
	assert.NotNil(t, bookResource, "'book' resource should not be nil")
	assert.Equal(t, "book", bookResource.Singular)
	assert.Equal(t, "books", bookResource.Plural)
	assert.NotNil(t, bookResource.Schema, "'book' resource schema should not be nil")
	assert.Contains(t, bookResource.Schema.Properties, "isbn")
	assert.Equal(t, "array", bookResource.Schema.Properties["isbn"].Type)
	assert.NotNil(t, bookResource.Methods.List, "'book' should have List method")
	assert.True(t, bookResource.Methods.List.HasUnreachableResources)
	assert.Len(t, bookResource.CustomMethods, 1, "'book' should have 1 custom method")
	assert.Equal(t, "archive", bookResource.CustomMethods[0].Name)
}
