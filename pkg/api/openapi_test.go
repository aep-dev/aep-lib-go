package api

import (
	"fmt"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/constants"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
)

func TestToOpenAPI(t *testing.T) {
	exampleAPI := ExampleAPI()
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
				"/publishers/{publisher_id}",
				"/publishers/{publisher_id}/books",
				"/publishers/{publisher_id}/books/{book_id}",
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
				"/publishers/{publisher_id}": {
					Get: &openapi.Operation{
						OperationID: "GetPublisher",
					},
				},
				"/publishers/{publisher_id}/books": {
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
				"/publishers/{publisher_id}/books/{book_id}": {
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
				"/publishers/{publisher_id}/books/{book_id}:archive": {
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
												"archived": {
													Type: "boolean",
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
			expectedListSchemas: map[string]*openapi.Schema{
				"/publishers/{publisher_id}/books": {
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
				"/publishers/{publisher_id}/books/{book_id}/editions",
				"/publishers/{publisher_id}/books/{book_id}/editions/{book_edition_id}",
			},
			expectedSchemas: []string{
				"account",
			},
			expectedOperations: map[string]openapi.PathItem{
				"/publishers/{publisher_id}/books/{book_id}/editions": {
					Get: &openapi.Operation{
						OperationID: "ListBookEdition",
					},
				},
				"/publishers/{publisher_id}/books/{book_id}/editions/{book_edition_id}": {
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
				assert.Equal(t, resource.Schema.XAEPResource.Type, fmt.Sprintf("%s/%s", tt.api.Name, resource.Singular))
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
				patternElems: []string{"databases", "{database_id}", "tables", "{table_id}"},
				Singular:     "table",
			},
			wantCollection: "/tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/databases/{database_id}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "database_id",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPField: &openapi.XAEPField{
								ResourceReference: []string{"database"},
							},
						},
					},
				},
			},
		},
		{
			name: "with pattern elements no nesting",
			resource: &Resource{
				patternElems: []string{"databases", "{database_id}"},
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
				parentResources: []*Resource{
					{
						Singular: "database",
						Plural:   "databases",
					},
				},
			},
			wantCollection: "/tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/databases/{database_id}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "database_id",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPField: &openapi.XAEPField{
								ResourceReference: []string{"database"},
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
				parentResources: []*Resource{
					{
						Singular: "database",
						Plural:   "databases",
						parentResources: []*Resource{
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
					Pattern: "/accounts/{account_id}/databases/{database_id}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "account_id",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPField: &openapi.XAEPField{
								ResourceReference: []string{"account"},
							},
						},
						{
							In:       "path",
							Name:     "database_id",
							Required: true,
							Schema: &openapi.Schema{
								Type: "string",
							},
							XAEPField: &openapi.XAEPField{
								ResourceReference: []string{"database"},
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

func TestLongRunningOperation(t *testing.T) {
	customMethod := &CustomMethod{
		Name:          "longRunningTest",
		Method:        "POST",
		IsLongRunning: true,
		Request: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"input": {Type: "string"},
			},
		},
		Response: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"output": {Type: "string"},
			},
		},
	}

	resource := &Resource{
		Singular: "test_resource",
		Plural:   "test_resources",
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"id":   {Type: "string"},
				"name": {Type: "string"},
			},
		},
		CustomMethods: []*CustomMethod{customMethod},
	}

	exampleAPI := &API{
		Name:      "Test API",
		ServerURL: "https://api.example.com",
		Resources: map[string]*Resource{
			"test_resource": resource,
		},
	}

	openAPI, err := ConvertToOpenAPI(exampleAPI)
	assert.NoError(t, err)
	assert.NotNil(t, openAPI)

	path := "/test-resources/{test_resource_id}:longRunningTest"
	operation := openAPI.Paths[path].Post
	assert.NotNil(t, operation, "Expected POST operation for long-running test")
	assert.Equal(t, AEP_OPERATION_REF,
		operation.Responses["200"].Content["application/json"].Schema.Ref,
		"Expected response schema to reference aep.api.Operation")
	assert.NotNil(t, operation.XAEPLongRunningOperation,
		"Expected XAEPLongRunningOperation to be set for long-running operation")
}

func TestLongRunningMethods(t *testing.T) {
	resource := &Resource{
		Singular: "test_resource",
		Plural:   "test_resources",
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"id":   {Type: "string"},
				"name": {Type: "string"},
			},
		},
		Methods: Methods{
			Create: &CreateMethod{
				IsLongRunning: true,
			},
			Apply: &ApplyMethod{
				IsLongRunning: true,
			},
			Update: &UpdateMethod{
				IsLongRunning: true,
			},
			Delete: &DeleteMethod{
				IsLongRunning: true,
			},
		},
	}

	exampleAPI := &API{
		Name:      "Test API",
		ServerURL: "https://api.example.com",
		Resources: map[string]*Resource{
			"test_resource": resource,
		},
	}

	openAPI, err := ConvertToOpenAPI(exampleAPI)
	assert.NoError(t, err)
	assert.NotNil(t, openAPI)

	// Validate CreateMethod
	createPath := "/test-resources"
	createOperation := openAPI.Paths[createPath].Post
	assert.NotNil(t, createOperation, "Expected POST operation for CreateMethod")
	assert.Equal(t, AEP_OPERATION_REF,
		createOperation.Responses["200"].Content["application/json"].Schema.Ref,
		"Expected response schema to reference aep.api.Operation for CreateMethod")
	assert.NotNil(t, createOperation.XAEPLongRunningOperation,
		"Expected XAEPLongRunningOperation to be set for CreateMethod")

	// Validate ApplyMethod
	applyPath := "/test-resources/{test_resource_id}"
	applyOperation := openAPI.Paths[applyPath].Put
	assert.NotNil(t, applyOperation, "Expected PUT operation for ApplyMethod")
	assert.Equal(t, AEP_OPERATION_REF,
		applyOperation.Responses["200"].Content["application/json"].Schema.Ref,
		"Expected response schema to reference aep.api.Operation for ApplyMethod")
	assert.NotNil(t, applyOperation.XAEPLongRunningOperation,
		"Expected XAEPLongRunningOperation to be set for ApplyMethod")

	// Validate UpdateMethod
	updateOperation := openAPI.Paths[applyPath].Patch
	assert.NotNil(t, updateOperation, "Expected PATCH operation for UpdateMethod")
	assert.Equal(t, AEP_OPERATION_REF,
		updateOperation.Responses["200"].Content["application/json"].Schema.Ref,
		"Expected response schema to reference aep.api.Operation for UpdateMethod")
	assert.NotNil(t, updateOperation.XAEPLongRunningOperation,
		"Expected XAEPLongRunningOperation to be set for UpdateMethod")

	// Validate DeleteMethod
	deleteOperation := openAPI.Paths[applyPath].Delete
	assert.NotNil(t, deleteOperation, "Expected DELETE operation for DeleteMethod")
	assert.Equal(t, AEP_OPERATION_REF,
		deleteOperation.Responses["200"].Content["application/json"].Schema.Ref,
		"Expected response schema to reference aep.api.Operation for DeleteMethod")
	assert.NotNil(t, deleteOperation.XAEPLongRunningOperation,
		"Expected XAEPLongRunningOperation to be set for DeleteMethod")
}

func TestDereferenceSchemaWithExternalReference(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	// Register a mock responder for the external schema
	httpmock.RegisterResponder("GET", "https://localhost:8080/mock-schema.json",
		httpmock.NewStringResponder(200, `{
			"type": "object",
			"properties": {
				"name": {"type": "string"},
				"age": {"type": "integer"}
			}
		}`))

	// Test case
	externalSchemaRef := "https://localhost:8080/mock-schema.json"
	schema := openapi.Schema{Ref: externalSchemaRef}
	openAPI := &openapi.OpenAPI{}

	resolvedSchema, err := openAPI.DereferenceSchema(schema)
	assert.NoError(t, err)
	assert.NotNil(t, resolvedSchema)

	// Additional assertions for debugging
	assert.Equal(t, "object", resolvedSchema.Type, "Expected schema type to be 'object'")
	assert.NotNil(t, resolvedSchema.Properties, "Expected schema properties to be populated")
	assert.Contains(t, resolvedSchema.Properties, "name", "Expected 'name' property in schema")
	assert.Contains(t, resolvedSchema.Properties, "age", "Expected 'age' property in schema")
}

func TestXAEPFieldNumberIsZeroed(t *testing.T) {
	// Create a resource with field_number set in XAEPField
	resource := &Resource{
		Singular: "test_resource",
		Plural:   "test_resources",
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"name": {
					Type: "string",
					XAEPField: &openapi.XAEPField{
						FieldNumber: 1,
					},
				},
				"id": {
					Type: "string",
					XAEPField: &openapi.XAEPField{
						FieldNumber: 2,
					},
				},
				"nested": {
					Type: "object",
					Properties: map[string]openapi.Schema{
						"value": {
							Type: "string",
							XAEPField: &openapi.XAEPField{
								FieldNumber: 3,
							},
						},
					},
					XAEPField: &openapi.XAEPField{
						FieldNumber: 4,
					},
				},
			},
			XAEPField: &openapi.XAEPField{
				FieldNumber: 5,
			},
		},
		Methods: Methods{
			Get: &GetMethod{},
		},
		CustomMethods: []*CustomMethod{
			{
				Name:   "custom",
				Method: "POST",
				Request: &openapi.Schema{
					Type: "object",
					Properties: map[string]openapi.Schema{
						"input": {
							Type: "string",
							XAEPField: &openapi.XAEPField{
								FieldNumber: 6,
							},
						},
					},
					XAEPField: &openapi.XAEPField{
						FieldNumber: 7,
					},
				},
				Response: &openapi.Schema{
					Type: "object",
					Properties: map[string]openapi.Schema{
						"output": {
							Type: "string",
							XAEPField: &openapi.XAEPField{
								FieldNumber: 8,
							},
						},
					},
					XAEPField: &openapi.XAEPField{
						FieldNumber: 9,
					},
				},
			},
		},
	}

	// Create additional schemas with field_number in XAEPField
	additionalSchema := &openapi.Schema{
		Type: "object",
		Properties: map[string]openapi.Schema{
			"extra": {
				Type: "string",
				XAEPField: &openapi.XAEPField{
					FieldNumber: 10,
				},
			},
		},
		XAEPField: &openapi.XAEPField{
			FieldNumber: 11,
		},
	}

	exampleAPI := &API{
		Name:      "Test API",
		ServerURL: "https://api.example.com",
		Resources: map[string]*Resource{
			"test_resource": resource,
		},
		Schemas: map[string]*openapi.Schema{
			"additional": additionalSchema,
		},
	}

	openAPI, err := ConvertToOpenAPI(exampleAPI)
	assert.NoError(t, err)
	assert.NotNil(t, openAPI)

	// Check that the main resource schema has field_number zeroed in XAEPField
	resourceSchema, exists := openAPI.Components.Schemas["test_resource"]
	assert.True(t, exists, "Expected test_resource schema to exist")
	assert.Nil(t, resourceSchema.XAEPField, "Main resource schema XAEPField should be nil after field_number removal")

	// Check that all properties in the main resource schema have field_number zeroed
	nameSchema := resourceSchema.Properties["name"]
	assert.Nil(t, nameSchema.XAEPField, "name property XAEPField should be nil after field_number removal")

	idSchema := resourceSchema.Properties["id"]
	assert.Nil(t, idSchema.XAEPField, "id property XAEPField should be nil after field_number removal")

	// Check nested object properties
	nestedSchema := resourceSchema.Properties["nested"]
	assert.Nil(t, nestedSchema.XAEPField, "nested object XAEPField should be nil after field_number removal")
	assert.Nil(t, nestedSchema.Properties["value"].XAEPField, "nested value property XAEPField should be nil after field_number removal")

	// Check that additional schemas have field_number zeroed
	additionalSchemaResult, exists := openAPI.Components.Schemas["additional"]
	assert.True(t, exists, "Expected additional schema to exist")
	assert.Nil(t, additionalSchemaResult.XAEPField, "Additional schema XAEPField should be nil after field_number removal")
	assert.Nil(t, additionalSchemaResult.Properties["extra"].XAEPField, "Additional schema extra property XAEPField should be nil after field_number removal")

	// Check that custom method request/response schemas have field_number zeroed
	// This would be in the paths, but since custom methods don't directly expose their schemas in components,
	// we verify that the removeXAEPFieldNumber function is called on all schemas during conversion
	// The function should recursively zero all field_number values in XAEPField structures
}
