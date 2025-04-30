package api

import (
	"encoding/json"
	"fmt"

	"github.com/aep-dev/aep-lib-go/pkg/constants"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

func LoadAPIFromJson(data []byte) (*API, error) {
	api := &API{}
	err := json.Unmarshal(data, api)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling API: %v", err)
	}
	err = addImplicitFieldsAndValidate(api)
	if err != nil {
		return nil, fmt.Errorf("error adding defaults to API: %v", err)
	}
	return api, nil
}

// addImplicitFieldsAndValidate adds implicit fields to the API object,
// such as the "path" variable in the resource.
func addImplicitFieldsAndValidate(api *API) error {
	// add the path variable to the resource
	for _, r := range api.Resources {
		r.API = api
		if r.Schema.Properties == nil {
			r.Schema.Properties = make(map[string]openapi.Schema)
		}
		r.Schema.Properties[constants.FIELD_PATH_NAME] = openapi.Schema{
			Type:            "string",
			Description:     "The server-assigned path of the resource, which is unique within the service.",
			XAEPFieldNumber: constants.FIELD_PATH_NUMBER,
		}
		for _, p := range r.Parents {
			if parent, ok := api.Resources[p]; ok {
				r.parentResources = append(r.parentResources, parent)
				parent.Children = append(parent.Children, r)
			} else {
				return fmt.Errorf("parent resource %s not found for resource %s", p, r.Singular)
			}
		}
	}
	return nil
}
