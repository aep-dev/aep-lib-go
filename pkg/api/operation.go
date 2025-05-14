package api

import (
	"encoding/json"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

var additionalPropertiesTrue = json.RawMessage([]byte("true"))

func OperationSchema() openapi.Schema {
	// jsonString := `true`
	// additionalPropertiesTrue := json.RawMessage([]byte(jsonString))
	return openapi.Schema{
		Type:                 "object",
		XAEPProtoMessageName: "aep.api.Operation",
		Required: []string{
			"name",
			"done",
		},
		Properties: map[string]openapi.Schema{
			"path": {
				Type:        "string",
				Description: "The server-assigned path of the operation, which is unique within the service.",
			},
			"metadata": {
				Type:                 "object",
				Description:          "Service-specific metadata associated with the operation.",
				AdditionalProperties: additionalPropertiesTrue,
			},
			"done": {
				Type:        "boolean",
				Description: "If the value is false, it means the operation is still in progress. If true, the operation is completed.",
			},
			"error": {
				Ref: "https://aep.dev/json-schema/type/problems.json",
			},
			"response": {
				Type:                 "object",
				Description:          "The normal response of the operation in case of success.",
				AdditionalProperties: additionalPropertiesTrue,
			},
		},
	}
}

func OperationResourceWithDefaults() *Resource {
	return OperationResource(Methods{
		Get: &GetMethod{},
	}, []*CustomMethod{})
}

// Return a longrunningoperation resource
func OperationResource(m Methods, cm []*CustomMethod) *Resource {
	schema := OperationSchema()
	return &Resource{
		Singular:      "operation",
		Plural:        "operations",
		Schema:        &schema,
		Methods:       m,
		CustomMethods: cm,
	}
}
