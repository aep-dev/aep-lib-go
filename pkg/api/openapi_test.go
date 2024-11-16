package api

import (
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
			openAPI, err := convertToOpenAPI(tt.api)
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
