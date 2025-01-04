package api

import (
	"fmt"
	"testing"

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
		ListMethod:   &ListMethod{},
		GetMethod:    &GetMethod{},
		CreateMethod: &CreateMethod{},
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
		},
		GetMethod:    &GetMethod{},
		CreateMethod: &CreateMethod{},
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
					},
					Post: &openapi.Operation{
						OperationID: "CreateBook",
					},
				},
				"/publishers/{publisher}/books/{book}": {
					Get: &openapi.Operation{
						OperationID: "GetBook",
					},
					Patch: &openapi.Operation{
						OperationID: "UpdateBook",
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
									Schema: &openapi.Schema{},
								},
							},
						},
						Responses: map[string]openapi.Response{
							"200": {
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
				if operations.Get != nil {
					assert.NotNil(t, pathItem.Get, "expected get operation for path %s", path)
					if operations.Get.OperationID != "" {
						assert.Equal(t, operations.Get.OperationID, pathItem.Get.OperationID,
							"expected matching operationId for GET %s", path)
					}
				}
				if operations.Patch != nil {
					assert.NotNil(t, pathItem.Patch, "expected patch operation for path %s", path)
					if operations.Patch.OperationID != "" {
						assert.Equal(t, operations.Patch.OperationID, pathItem.Patch.OperationID,
							"expected matching operationId for PATCH %s", path)
					}
				}
				if operations.Post != nil {
					assert.NotNil(t, pathItem.Post, "expected post operation for path %s", path)
					if operations.Post.OperationID != "" {
						assert.Equal(t, operations.Post.OperationID, pathItem.Post.OperationID,
							"expected matching operationId for POST %s", path)
					}
				}
				if operations.Put != nil {
					assert.NotNil(t, pathItem.Put, "expected put operation for path %s", path)
					if operations.Put.OperationID != "" {
						assert.Equal(t, operations.Put.OperationID, pathItem.Put.OperationID,
							"expected matching operationId for PUT %s", path)
					}
				}
				if operations.Delete != nil {
					assert.NotNil(t, pathItem.Delete, "expected delete operation for path %s", path)
					if operations.Delete.OperationID != "" {
						assert.Equal(t, operations.Delete.OperationID, pathItem.Delete.OperationID,
							"expected matching operationId for DELETE %s", path)
					}
				}
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
