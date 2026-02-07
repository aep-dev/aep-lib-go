package proto

import (
	"strings"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
	"github.com/stretchr/testify/assert"
)

func TestResourcePathFieldReadOnly(t *testing.T) {
	// Setup a simple API with a resource that has a 'path' field
	a := &api.API{
		Name: "example.library.v1",
		Resources: map[string]*api.Resource{
			"shelf": {
				Singular: "shelf",
				Plural:   "shelves",
				Schema: &openapi.Schema{
					Type: "object",
					Properties: map[string]openapi.Schema{
						"path": {
							Type:     "string",
							ReadOnly: true,
							XAEPField: &openapi.XAEPField{
								FieldNumber: 1,
							},
						},
						"display_name": {
							Type: "string",
							XAEPField: &openapi.XAEPField{
								FieldNumber: 2,
							},
						},
					},
					Required: []string{},
				},
			},
		},
		Schemas: map[string]*openapi.Schema{},
	}

	// Generate Proto
	protoBytes, err := APIToProtoString(a, "example/library/v1")
	assert.NoError(t, err)
	protoContent := string(protoBytes)

	// Check Protobuf output for OUTPUT_ONLY on 'path'
	// We expect something like:
	// string path = 1 [
	//   (google.api.field_behavior) = OUTPUT_ONLY
	// ];
	// Note: field number might vary, so we just check for the presence of the annotation near "path"
	// or generally in the message.

	// Since checking multiline regex on generated code is brittle, we will check if the line defining 'path'
	// is followed by the annotation or contains it.
	// However, protoprint might format it differently.
	// Let's print it to see if it fails (it will fail initially).

	// Assert OUTPUT_ONLY
	// The exact formatting depends on protoprint, but usually annotations appear inside keys or after field type.
	// For `jhump/protoreflect`, it often puts options in brackets `[...]`.

	if !strings.Contains(protoContent, "OUTPUT_ONLY") {
		// failing initially is expected
		t.Errorf("OUTPUT_ONLY annotation not found (found: %s)", protoContent)
	}
}
