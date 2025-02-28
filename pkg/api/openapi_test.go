package api

import (
	"fmt"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/constants"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/stretchr/testify/assert"
)

func TestToOpenAPI(t *testing.T) {
	// Common example API used across tests
	publisher := &Resource{
		Singular: "publisher",
		Plural:   "publishers",
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"title": {Type: "string"},
				"id":    {Type: "string"},
			},
		},
		ListMethod: &ListMethod{},
		GetMethod:  &GetMethod{},
		CreateMethod: &CreateMethod{
			SupportsUserSettableCreate: true,
		},
	}
	book := &Resource{
		Singular: "book",
		Plural:   "books",
		Parents:  []*Resource{publisher},
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"name": {Type: "string"},
				"id":   {Type: "string"},
			},
		},
		ListMethod: &ListMethod{
			HasUnreachableResources: true,
			SupportsFilter:          true,
			SupportsSkip:            true,
		},
		GetMethod: &GetMethod{},
		CreateMethod: &CreateMethod{
			SupportsUserSettableCreate: true,
		},
		UpdateMethod: &UpdateMethod{},
		DeleteMethod: &DeleteMethod{},
		CustomMethods: []*CustomMethod{
			{
				Name:   "archive",
				Method: "POST",
				Request: &openapi.Schema{
					Type:       "object",
					Properties: map[string]openapi.Schema{},
				},
				Response: &openapi.Schema{
					Type: "object",
					Properties: map[string]openapi.Schema{
						"archived": {Type: "boolean"},
					},
				},
			},
		},
	}
	publisher.Children = append(publisher.Children, book)
	bookEdition := &Resource{
		Singular: "book-edition",
		Plural:   "book-editions",
		Parents:  []*Resource{book},
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"date": {Type: "string"},
			},
		},
		ListMethod: &ListMethod{},
		GetMethod:  &GetMethod{},
	}
	book.Children = append(book.Children, bookEdition)
	exampleAPI := &API{
		Name:      "Test API",
		ServerURL: "https://api.example.com",
		Contact: &Contact{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			URL:   "https://example.com",
		},
		Schemas: map[string]*openapi.Schema{
			"account": {
				Type: "object",
			},
		},
		Resources: map[string]*Resource{
			"book":         book,
			"book-edition": bookEdition,
			"publisher":    publisher,
		},
	}
	tests := []struct {
		name                string
		api                 *API
		expectedPaths       []string
		expectedSchemas     []string
		expectedOperations  map[string]openapi.PathItem
		expectedListSchemas map[string]*openapi.Schema
		wantErr             bool
	}{
		{
			name: "Basic resource paths",
			api:  exampleAPI,
			expectedPaths: []string{
				"/publishers",
				"/publishers/{publisher}",
				"/publishers/{publisher}/books",
				"/publishers/{publisher}/books/{book}",
			},
			expectedSchemas: []string{
				"account",
			},
			expectedOperations: map[string]openapi.PathItem{
				"/publishers": {
					Get: &openapi.Operation{
						OperationID: "ListPublisher",
					},
					Post: &openapi.Operation{
						OperationID: "CreatePublisher",
					},
				},
				"/publishers/{publisher}": {
					Get: &openapi.Operation{
						OperationID: "GetPublisher",
					},
				},
				"/publishers/{publisher}/books": {
					Get: &openapi.Operation{
						OperationID: "ListBook",
						Parameters: []openapi.Parameter{
							{
								Name:     "skip",
								In:       "query",
								Required: false,
								Schema: &openapi.Schema{
									Type: "integer",
								},
							},
							{
								Name:     "filter",
								In:       "query",
								Required: false,
								Schema: &openapi.Schema{
									Type: "string",
								},
							},
						},
						Responses: map[string]openapi.Response{
							"200": {
								Description: "Successful response",
								Content: map[string]openapi.MediaType{
									"application/json": {
										Schema: &openapi.Schema{
											Type: "object",
											Properties: map[string]openapi.Schema{
												constants.FIELD_NEXT_PAGE_TOKEN_NAME: {
													Type: "string",
												},
												constants.FIELD_UNREACHABLE_NAME: {
													Type: "array",
													Items: &openapi.Schema{
														Type: "string",
													},
												},
												constants.FIELD_RESULTS_NAME: {
													Type: "array",
													Items: &openapi.Schema{
														Ref: "#/components/schemas/book",
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
						OperationID: "CreateBook",
						Parameters: []openapi.Parameter{
							{
								Name:     "id",
								In:       "query",
								Required: false,
								Schema: &openapi.Schema{
									Type: "string",
								},
							},
						},
					},
				},
				"/publishers/{publisher}/books/{book}": {
					Get: &openapi.Operation{
						OperationID: "GetBook",
					},
					Patch: &openapi.Operation{
						OperationID: "UpdateBook",
						RequestBody: &openapi.RequestBody{
							Required: true,
							Content: map[string]openapi.MediaType{
								"application/merge-patch+json": {
									Schema: &openapi.Schema{
										Ref: "#/components/schemas/book",
									},
								},
							},
						},
						Responses: map[string]openapi.Response{
							"200": {
								Description: "Successful response",
								Content: map[string]openapi.MediaType{
									"application/merge-patch+json": {
										Schema: &openapi.Schema{
											Ref: "#/components/schemas/book",
										},
									},
								},
							},
						},
					},
					Delete: &openapi.Operation{
						OperationID: "DeleteBook",
					},
				},
				"/publishers/{publisher}/books/{book}:archive": {
					Post: &openapi.Operation{
						OperationID: ":ArchiveBook",
						RequestBody: &openapi.RequestBody{
							Required: true,
							Content: map[string]openapi.MediaType{
								"application/json": {
									Schema: &openapi.Schema{
										Type:       "object",
										Properties: map[string]openapi.Schema{},
									},
								},
							},
						},
						Responses: map[string]openapi.Response{
							"200": {
								Description: "Successful response",
								Content: map[string]openapi.MediaType{
									"application/json": {
										Schema: &openapi.Schema{
											Type: "object",
											Properties: map[string]openapi.Schema{
												"archived": {Type: "boolean"},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedListSchemas: map[string]*openapi.Schema{
				"/publishers/{publisher}/books": {
					Type: "object",
					Properties: map[string]openapi.Schema{
						"unreachable": {
							Type: "array",
							Items: &openapi.Schema{
								Type: "string",
							},
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "book edition",
			api:  exampleAPI,
			expectedPaths: []string{
				"/publishers/{publisher}/books/{book}/editions",
				"/publishers/{publisher}/books/{book}/editions/{book-edition}",
			},
			expectedSchemas: []string{
				"account",
			},
			expectedOperations: map[string]openapi.PathItem{
				"/publishers/{publisher}/books/{book}/editions": {
					Get: &openapi.Operation{
						OperationID: "ListBookEdition",
					},
				},
				"/publishers/{publisher}/books/{book}/editions/{book-edition}": {
					Get: &openapi.Operation{
						OperationID: "GetBookEdition",
					},
				},
			},
			expectedListSchemas: map[string]*openapi.Schema{},
			wantErr:             false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			openAPI, err := ConvertToOpenAPI(tt.api)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.NotNil(t, openAPI)

			// Verify basic OpenAPI structure
			assert.Equal(t, "3.1.0", openAPI.OpenAPI)
			assert.Equal(t, tt.api.Name, openAPI.Info.Title)
			assert.Equal(t, tt.api.ServerURL, openAPI.Servers[0].URL)

			// Verify Contact information
			if tt.api.Contact != nil {
				assert.Equal(t, tt.api.Contact.Name, openAPI.Info.Contact.Name)
				assert.Equal(t, tt.api.Contact.Email, openAPI.Info.Contact.Email)
				assert.Equal(t, tt.api.Contact.URL, openAPI.Info.Contact.URL)
			}

			// Verify paths exist
			fmt.Println(openAPI.Paths)
			for _, expectedPath := range tt.expectedPaths {
				_, exists := openAPI.Paths[expectedPath]
				assert.True(t, exists, "Expected path %s not found", expectedPath)
			}

			// Verify schemas exist
			for _, resource := range tt.api.Resources {
				schema, exists := openAPI.Components.Schemas[resource.Singular]
				assert.True(t, exists, "Expected schema %s not found", resource.Singular)
				assert.Equal(t, resource.Schema.Type, schema.Type)
				assert.Equal(t, resource.Schema.XAEPResource.Singular, resource.Singular)
			}
			for _, schema := range tt.expectedSchemas {
				_, exists := openAPI.Components.Schemas[schema]
				assert.True(t, exists, "Expected schema %s not found", schema)
			}

			// Verify operations exist and have correct operationIds
			for path, operations := range tt.expectedOperations {
				pathItem, exists := openAPI.Paths[path]
				assert.True(t, exists, "Expected path %s not found", path)

				assertOperationsMatch(t, path, operations.Get, pathItem.Get)
				assertOperationsMatch(t, path, operations.Post, pathItem.Post)
				assertOperationsMatch(t, path, operations.Put, pathItem.Put)
				assertOperationsMatch(t, path, operations.Patch, pathItem.Patch)
				assertOperationsMatch(t, path, operations.Delete, pathItem.Delete)
			}

			// Add new verification for List response schemas
			for path, expectedSchema := range tt.expectedListSchemas {
				pathItem, exists := openAPI.Paths[path]
				assert.True(t, exists, "Expected path %s not found", path)

				// Verify List operation response schema
				listResponse := pathItem.Get.Responses["200"]
				if expectedSchema != nil {
					assert.NotNil(t, listResponse.Content["application/json"].Schema.Properties["unreachable"],
						"Expected unreachable array in List response schema for path %s", path)
					s := listResponse.Content["application/json"].Schema
					for name, prop := range expectedSchema.Properties {
						assert.Equal(t, prop, s.Properties[name])
					}
				}
			}
		})
	}
}

func TestGenerateParentPatternsWithParams(t *testing.T) {
	tests := []struct {
		name           string
		resource       *Resource
		wantCollection string
		wantPathParams *[]PathWithParams
	}{
		{
			name: "with pattern elements",
			resource: &Resource{
				PatternElems: []string{"databases", "{database}", "tables", "{table}"},
				Singular:     "table",
			},
			wantCollection: "/tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/databases/{database}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "database",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPResourceRef: &openapi.XAEPResourceRef{
								Resource: "database",
							},
						},
					},
				},
			},
		},
		{
			name: "with pattern elements no nesting",
			resource: &Resource{
				PatternElems: []string{"databases", "{database}"},
				Singular:     "database",
			},
			wantCollection: "/databases",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "",
					Params:  []openapi.Parameter{},
				},
			},
		},

		{
			name: "without pattern elements",
			resource: &Resource{
				Singular: "table",
				Plural:   "tables",
				Parents: []*Resource{
					{
						Singular: "database",
						Plural:   "databases",
					},
				},
			},
			wantCollection: "/tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/databases/{database}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "database",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPResourceRef: &openapi.XAEPResourceRef{
								Resource: "database",
							},
						},
					},
				},
			},
		},
		{
			name: "without pattern elements, nested parent",
			resource: &Resource{
				Singular: "table",
				Plural:   "tables",
				Parents: []*Resource{
					{
						Singular: "database",
						Plural:   "databases",
						Parents: []*Resource{
							{
								Singular: "account",
								Plural:   "accounts",
							},
						},
					},
				},
			},
			wantCollection: "/tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/accounts/{account}/databases/{database}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "account",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPResourceRef: &openapi.XAEPResourceRef{
								Resource: "account",
							},
						},
						{
							In:       "path",
							Name:     "database",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPResourceRef: &openapi.XAEPResourceRef{
								Resource: "database",
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCollection, gotPathParams := generateParentPatternsWithParams(tt.resource)

			if gotCollection != tt.wantCollection {
				t.Errorf("collection = %v, want %v", gotCollection, tt.wantCollection)
			}

			if len(*gotPathParams) != len(*tt.wantPathParams) {
				t.Errorf("pathParams length = %v, want %v", len(*gotPathParams), len(*tt.wantPathParams))
			}

			for i, got := range *gotPathParams {
				want := (*tt.wantPathParams)[i]
				if got.Pattern != want.Pattern {
					t.Errorf("pattern[%d] = %v, want %v", i, got.Pattern, want.Pattern)
				}

				if len(got.Params) != len(want.Params) {
					t.Errorf("params[%d] length = %v, want %v", i, len(got.Params), len(want.Params))
				}

				for j, gotParam := range got.Params {
					wantParam := want.Params[j]
					if gotParam.Name != wantParam.Name ||
						gotParam.In != wantParam.In ||
						gotParam.Required != wantParam.Required ||
						gotParam.Schema.Type != wantParam.Schema.Type {
						t.Errorf("param[%d][%d] = %+v, want %+v", i, j, gotParam, wantParam)
					}
				}
			}
		})
	}
}

// assertOperationsMatch compares two OpenAPI operations and verifies they match the expected configuration
func assertOperationsMatch(t *testing.T, path string, expected, actual *openapi.Operation) {
	if expected == nil {
		assert.Nil(t, actual, "unexpected operation for path %s", path)
		return
	}

	assert.NotNil(t, actual, "expected operation for path %s", path)

	// Compare OperationID if specified
	if expected.OperationID != "" {
		assert.Equal(t, expected.OperationID, actual.OperationID,
			"expected matching operationId for path %s", path)
	}

	// Compare Parameters if specified
	for _, expectedParam := range expected.Parameters {
		assert.Contains(t, actual.Parameters, expectedParam,
			"expected parameter %s for path %s", expectedParam.Name, path)
	}

	// Compare RequestBody if specified
	if expected.RequestBody != nil {
		assert.Equal(t, expected.RequestBody, actual.RequestBody,
			"expected matching request body for path %s", path)
	}

	// Compare Responses if specified
	for status, expectedResponse := range expected.Responses {
		actualResponse, exists := actual.Responses[status]
		assert.True(t, exists, "expected response %s for path %s", status, path)
		assert.Equal(t, expectedResponse, actualResponse,
			"expected matching response for status %s path %s", status, path)
	}
}
