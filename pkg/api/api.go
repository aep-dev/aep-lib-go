package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/cases"
	"github.com/aep-dev/aep-lib-go/pkg/constants"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

type API struct {
	ServerURL string `json:"server_url"`
	Name      string
	Contact   *Contact
	Schemas   map[string]*openapi.Schema
	// A list of the resources that are exposed by the API.
	//
	// The key "operation" carries a special meaning, and must
	// map to an aep.dev/151 Operation resource.
	Resources map[string]*Resource
}

type Contact struct {
	Name  string
	Email string
	URL   string
}

func LoadAPIFromJson(data []byte) (*API, error) {
	api := &API{}
	err := json.Unmarshal(data, api)
	if err != nil {
		return nil, fmt.Errorf("error unmarshalling API: %v", err)
	}
	err = addImplicitFields(api)
	if err != nil {
		return nil, fmt.Errorf("error adding defaults to API: %v", err)
	}
	return api, nil
}

func GetAPI(api *openapi.OpenAPI, serverURL, pathPrefix string) (*API, error) {
	if api.OASVersion() == "" {
		return nil, fmt.Errorf("unable to detect OAS version. Please add a openapi field or a swagger field")
	}
	slog.Debug("parsing openapi", "pathPrefix", pathPrefix)
	resourceBySingular := make(map[string]*Resource)
	customMethodsByPattern := make(map[string][]*CustomMethod)
	// we try to parse the paths to find possible resources, since
	// they may not always be annotated as such.
	for path, pathItem := range api.Paths {
		path = path[len(pathPrefix):]
		slog.Debug("path", "path", path)
		var r Resource
		var sRef *openapi.Schema
		p := getPatternInfo(path)
		var lroDetails *openapi.XAEPLongRunningOperation
		if p == nil { // not a resource pattern
			slog.Debug("path is not a resource", "path", path)
			continue
		}
		slog.Debug("parsing path for resource", "path", path)
		if p.CustomMethodName != "" && p.IsResourcePattern {
			// strip the leading slash and the custom method suffix
			pattern := strings.Split(path, ":")[0][1:]
			if _, ok := customMethodsByPattern[pattern]; !ok {
				customMethodsByPattern[pattern] = []*CustomMethod{}
			}
			if pathItem.Post != nil {
				if resp, ok := pathItem.Post.Responses["200"]; ok {
					lroDetails = pathItem.Post.XAEPLongRunningOperation
					schema := api.GetSchemaFromResponse(resp)
					responseSchema := &openapi.Schema{}
					if lroDetails != nil {
						schema = lroDetails.Response.Schema
					}
					if schema != nil {
						var err error
						responseSchema, err = api.DereferenceSchema(*schema)
						if err != nil {
							return nil, fmt.Errorf("error dereferencing schema %v: %v", schema, err)
						}
					}
					if pathItem.Post.RequestBody == nil {
						return nil, fmt.Errorf("custom method %q has a POST response, but no request body", p.CustomMethodName)
					}
					schema = api.GetSchemaFromRequestBody(*pathItem.Post.RequestBody)
					requestSchema, err := api.DereferenceSchema(*schema)
					if err != nil {
						return nil, fmt.Errorf("error dereferencing schema %q: %v", schema.Ref, err)
					}
					customMethodsByPattern[pattern] = append(customMethodsByPattern[pattern], &CustomMethod{
						Name:          p.CustomMethodName,
						Method:        "POST",
						Request:       requestSchema,
						Response:      responseSchema,
						IsLongRunning: lroDetails != nil,
					})
				}
			}
			if pathItem.Get != nil {
				if resp, ok := pathItem.Get.Responses["200"]; ok {
					lroDetails = pathItem.Post.XAEPLongRunningOperation
					schema := api.GetSchemaFromResponse(resp)
					responseSchema := &openapi.Schema{}
					if lroDetails != nil {
						schema = lroDetails.Response.Schema
					}
					if schema != nil {
						var err error
						responseSchema, err = api.DereferenceSchema(*schema)
						if err != nil {
							return nil, fmt.Errorf("error dereferencing schema %v: %v", schema.Ref, err)
						}
					}
					customMethodsByPattern[pattern] = append(r.CustomMethods, &CustomMethod{
						Name:          p.CustomMethodName,
						Method:        "GET",
						Response:      responseSchema,
						IsLongRunning: lroDetails != nil,
					})
				}
			}
		} else if p.IsResourcePattern {
			// treat it like a collection pattern (update, delete, get)
			if pathItem.Delete != nil {
				lroDetails = pathItem.Delete.XAEPLongRunningOperation
				r.Methods.Delete = &DeleteMethod{
					IsLongRunning: lroDetails != nil,
				}
			}
			if pathItem.Get != nil {
				if resp, ok := pathItem.Get.Responses["200"]; ok {
					sRef = api.GetSchemaFromResponse(resp)
					r.Methods.Get = &GetMethod{}
				}
			}
			if pathItem.Patch != nil {
				lroDetails = pathItem.Patch.XAEPLongRunningOperation
				if resp, ok := pathItem.Patch.Responses["200"]; ok {
					sRef = api.GetSchemaFromResponse(resp)
					r.Methods.Update = &UpdateMethod{
						IsLongRunning: lroDetails != nil,
					}
				}
			}
		} else {
			// create method
			if pathItem.Post != nil {
				// check if there is a query parameter "id"
				lroDetails = pathItem.Post.XAEPLongRunningOperation
				if resp, ok := pathItem.Post.Responses["200"]; ok {
					sRef = api.GetSchemaFromResponse(resp)
					supportsUserSettableCreate := false
					for _, param := range pathItem.Post.Parameters {
						if param.Name == "id" {
							supportsUserSettableCreate = true
							break
						}
					}
					r.Methods.Create = &CreateMethod{
						SupportsUserSettableCreate: supportsUserSettableCreate,
						IsLongRunning:              lroDetails != nil,
					}
				}
			}
			// list method
			if pathItem.Get != nil {
				if resp, ok := pathItem.Get.Responses["200"]; ok {
					respSchema := api.GetSchemaFromResponse(resp)
					if respSchema == nil {
						slog.Warn(fmt.Sprintf("resource %q has a LIST method with a response schema, but the response schema is nil.", path))
					} else {
						resolvedSchema, err := api.DereferenceSchema(*respSchema)
						if err != nil {
							return nil, fmt.Errorf("error dereferencing schema %q: %v", respSchema.Ref, err)
						}
						found := false
						for _, property := range resolvedSchema.Properties {
							if property.Type == "array" {
								sRef = property.Items
								r.Methods.List = &ListMethod{}
								found = true
								break
							}
						}
						if found {
							for _, param := range pathItem.Get.Parameters {
								if param.Name == constants.FIELD_SKIP_NAME {
									r.Methods.List.SupportsSkip = true
								}
								if param.Name == constants.FIELD_UNREACHABLE_NAME {
									r.Methods.List.HasUnreachableResources = true
								}
								if param.Name == constants.FIELD_FILTER_NAME {
									r.Methods.List.SupportsFilter = true
								}
							}
						} else {
							slog.Warn(fmt.Sprintf("resource %q has a LIST method with a response schema, but the items field is not present or is not an array.", path))
						}
					}
				}
			}
		}
		if lroDetails != nil {
			sRef = lroDetails.Response.Schema
		}
		if sRef != nil {
			// s should always be a reference to a schema in the components section.
			parts := strings.Split(sRef.Ref, "/")
			key := parts[len(parts)-1]
			dereferencedSchema, err := api.DereferenceSchema(*sRef)
			if err != nil {
				return nil, fmt.Errorf("error dereferencing schema %q: %v", sRef.Ref, err)
			}
			singular := cases.PascalCaseToKebabCase(key)
			pattern := strings.Split(path, "/")[1:]
			if !p.IsResourcePattern {
				// deduplicate the singular, if applicable
				finalSingular := singular
				parent := ""
				if len(pattern) >= 3 {
					parent = pattern[len(pattern)-3]
					parent = parent[0 : len(parent)-1] // strip curly surrounding
					if strings.HasPrefix(singular, parent) {
						finalSingular = strings.TrimPrefix(singular, parent+"-")
					}
				}
				pattern = append(pattern, fmt.Sprintf("{%s}", finalSingular))
			}
			r2, err := getOrPopulateResource(singular, pattern, dereferencedSchema, resourceBySingular, api)
			if err != nil {
				return nil, fmt.Errorf("error populating resource %q: %v", r.Singular, err)
			}
			foldResourceMethods(&r, r2)
		}
	}
	// the custom methods are trickier - because they may not respond with the schema of the resource
	// (which would allow us to map the resource via looking at it's reference), we instead will have to
	// map it by the pattern.
	// we also have to do this by longest pattern match - this helps account for situations where
	// the custom method doesn't match the resource pattern exactly with things like deduping.
	for pattern, customMethods := range customMethodsByPattern {
		found := false
		for _, r := range resourceBySingular {
			if r.GetPattern() == pattern {
				r.CustomMethods = customMethods
				found = true
				break
			}
		}
		if !found {
			slog.Debug(fmt.Sprintf("custom methods with pattern %q have no resource associated with it", pattern))
		}
	}
	if serverURL == "" {
		for _, s := range api.Servers {
			serverURL = s.URL + pathPrefix
		}
	}

	if serverURL == "" {
		return nil, fmt.Errorf("no server URL found in openapi, and none was provided")
	}

	// any schemas that are not a resource are added to the API's schemas
	schemas := make(map[string]*openapi.Schema)
	for k, v := range api.Components.Schemas {
		if _, ok := resourceBySingular[k]; ok {
			continue
		}
		schemas[k] = &v
	}

	return &API{
		ServerURL: serverURL,
		Name:      api.Info.Title,
		Contact:   getContact(api.Info.Contact),
		Resources: resourceBySingular,
		Schemas:   schemas,
	}, nil
}

func (s *API) GetResource(resource string) (*Resource, error) {
	r, ok := (*s).Resources[resource]
	if !ok {
		return nil, fmt.Errorf("Resource %q not found", resource)
	}
	return r, nil
}

type PatternInfo struct {
	// if true, the pattern represents an individual resource,
	// otherwise it represents a path to a collection of resources
	IsResourcePattern bool
	CustomMethodName  string
}

// getPatternInfo returns true if the path is an alternating pairing of collection and id,
// and returns the collection names if so.
func getPatternInfo(path string) *PatternInfo {
	customMethodName := ""
	if strings.Contains(path, ":") {
		parts := strings.Split(path, ":")
		path = parts[0]
		customMethodName = parts[1]
	}
	// we ignore the first segment, which is empty.
	pattern := strings.Split(path, "/")[1:]
	for i, segment := range pattern {
		// check if segment is wrapped in curly brackets
		wrapped := strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
		wantWrapped := i%2 == 1
		if wrapped != wantWrapped {
			return nil
		}
	}
	return &PatternInfo{
		IsResourcePattern: len(pattern)%2 == 0,
		CustomMethodName:  customMethodName,
	}
}

// getOrPopulateResource populates the resource via a variety of means:
// - if the resource already exists in the map, it returns it
// - if the schema has the x-aep-resource annotation, it parses the resource
// - otherwise, it attempts to infer the resource from the schema and name.
func getOrPopulateResource(singular string, pattern []string, s *openapi.Schema, resourceBySingular map[string]*Resource, api *openapi.OpenAPI) (*Resource, error) {
	if r, ok := resourceBySingular[singular]; ok {
		return r, nil
	}
	var r *Resource
	// use the X-AEP-Resource annotation to populate the resource,
	// if it exists.
	if s.XAEPResource != nil {
		parents := []*Resource{}
		for _, parentSingular := range s.XAEPResource.Parents {
			parentSchema, ok := api.Components.Schemas[parentSingular]
			if !ok {
				return nil, fmt.Errorf("resource %q parent %q not found", singular, parentSingular)
			}
			parentResource, err := getOrPopulateResource(parentSingular, []string{}, &parentSchema, resourceBySingular, api)
			if err != nil {
				return nil, fmt.Errorf("error parsing resource %q parent %q: %v", singular, parentSingular, err)
			}
			parents = append(parents, parentResource)
			parentResource.Children = append(parentResource.Children, r)
		}
		r = &Resource{
			Singular:     s.XAEPResource.Singular,
			Plural:       s.XAEPResource.Plural,
			Parents:      parents,
			Children:     []*Resource{},
			PatternElems: strings.Split(strings.TrimPrefix(s.XAEPResource.Patterns[0], "/"), "/"),
			Schema:       s,
		}
	} else {
		// best effort otherwise
		r = &Resource{
			Schema:       s,
			PatternElems: pattern,
			Singular:     singular,
			Parents:      []*Resource{},
			Children:     []*Resource{},
			Plural:       plural(singular),
		}
	}
	resourceBySingular[singular] = r
	return r, nil
}

func foldResourceMethods(from, into *Resource) {
	if from.Methods.Get != nil {
		into.Methods.Get = from.Methods.Get
	}
	if from.Methods.List != nil {
		into.Methods.List = from.Methods.List
	}
	if from.Methods.Create != nil {
		into.Methods.Create = from.Methods.Create
	}
	if from.Methods.Update != nil {
		into.Methods.Update = from.Methods.Update
	}
	if from.Methods.Delete != nil {
		into.Methods.Delete = from.Methods.Delete
	}
}

func getContact(contact openapi.Contact) *Contact {
	if contact.Name != "" || contact.Email != "" || contact.URL != "" {
		return &Contact{
			Name:  contact.Name,
			Email: contact.Email,
			URL:   contact.URL,
		}
	}
	return nil
}

// plural returns the plural form of a singular resource name
// This is a simple implementation that just adds 's' to the end
// of the singular form, which works for most cases.
func plural(singular string) string {
	return singular + "s"
}

// addImplicitFields adds implicit fields to the API object,
// such as the "path" variable in the resource.
func addImplicitFields(api *API) error {
	// add the path variable to the resource
	for _, r := range api.Resources {
		if r.Schema.Properties != nil {
			r.Schema.Properties[constants.FIELD_PATH_NAME] = openapi.Schema{
				Type:            "string",
				Description:     "The server-assigned path of the resource, which is unique within the service.",
				XAEPFieldNumber: constants.FIELD_PATH_NUMBER,
			}
		}
	}
	return nil
}
