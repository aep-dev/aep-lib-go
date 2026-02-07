package api

import (
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddImplicitFieldsAndValidate(t *testing.T) {
	tests := []struct {
		name        string
		api         *API
		expectedErr string
		validate    func(*testing.T, *API)
	}{
		{
			name: "adds path field with readOnly to resource",
			api: &API{
				Resources: map[string]*Resource{
					"book": {
						Singular: "book",
						Plural:   "books",
						Schema: &openapi.Schema{
							Type: "object",
						},
					},
				},
			},
			validate: func(t *testing.T, api *API) {
				book, ok := api.Resources["book"]
				require.True(t, ok)
				pathProp, ok := book.Schema.Properties["path"]
				assert.True(t, ok, "path property should exist")
				assert.Equal(t, "string", pathProp.Type)
				assert.True(t, pathProp.ReadOnly, "path property should be readOnly")
				assert.NotNil(t, pathProp.XAEPField)
				assert.Equal(t, 10018, pathProp.XAEPField.FieldNumber)
			},
		},
		{
			name: "validates resource naming",
			api: &API{
				Resources: map[string]*Resource{
					"Book": { // Invalid: should be lower case
						Singular: "book",
						Plural:   "books",
						Schema:   &openapi.Schema{},
					},
				},
			},
			expectedErr: "resource name Book does not match the regex",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := AddImplicitFieldsAndValidate(tt.api)
			if tt.expectedErr != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedErr)
			} else {
				assert.NoError(t, err)
				if tt.validate != nil {
					tt.validate(t, tt.api)
				}
			}
		})
	}
}
