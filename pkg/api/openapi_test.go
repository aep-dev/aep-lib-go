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
		ListMethod:   &ListMethod{},
		GetMethod:    &GetMethod{},
		CreateMethod: &CreateMethod{},
		UpdateMethod: &UpdateMethod{},
		DeleteMethod: &DeleteMethod{},
	}
	publisher.Children = append(publisher.Children, book)
	exampleAPI := &API{
		Name:      "Test API",
		ServerURL: "https://api.example.com",
		Resources: map[string]*Resource{
			"book":      book,
			"publisher": publisher,
		},
	}

	tests := []struct {
		name          string
		api           *API
		expectedPaths []string
		wantErr       bool
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
			wantErr: false,
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
			wantCollection: "tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/databases/{database}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "database",
							Required: true,
							Type:     "string",
						},
					},
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
			wantCollection: "tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/databases/{database}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "database",
							Required: true,
							Type:     "string",
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
			wantCollection: "tables",
			wantPathParams: &[]PathWithParams{
				{
					Pattern: "/accounts/{account}/databases/{database}",
					Params: []openapi.Parameter{
						{
							In:       "path",
							Name:     "account",
							Required: true,
							Type:     "string",
						},
						{
							In:       "path",
							Name:     "database",
							Required: true,
							Type:     "string",
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
						gotParam.Type != wantParam.Type {
						t.Errorf("param[%d][%d] = %+v, want %+v", i, j, gotParam, wantParam)
					}
				}
			}
		})
	}
}