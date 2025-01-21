package api

import (
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/stretchr/testify/assert"
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
				assert.NotNil(t, widget.GetMethod, "should have GET method")
				assert.NotNil(t, widget.ListMethod, "should have LIST method")
				assert.NotNil(t, widget.CreateMethod, "should have CREATE method")
				if widget.CreateMethod != nil {
					assert.False(t, widget.CreateMethod.SupportsUserSettableCreate, "should not support user-settable create")
				}
				assert.NotNil(t, widget.UpdateMethod, "should have UPDATE method")
				assert.NotNil(t, widget.DeleteMethod, "should have DELETE method")
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
				assert.True(t, widget.CreateMethod.SupportsUserSettableCreate,
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
				assert.NotNil(t, widget.GetMethod, "should have GET method")
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
				assert.NotNil(t, widget.ListMethod, "should have LIST method")
				assert.True(t, widget.ListMethod.SupportsSkip, "should support skip parameter")
				assert.True(t, widget.ListMethod.HasUnreachableResources, "should support unreachable parameter")
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
