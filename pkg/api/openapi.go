package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/constants"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

func (api *API) ConvertToOpenAPIBytes() ([]byte, error) {
	openAPI, err := ConvertToOpenAPI(api)
	if err != nil {
		return nil, err
	}

	jsonData, err := json.MarshalIndent(openAPI, "", "  ")
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func ConvertToOpenAPI(api *API) (*openapi.OpenAPI, error) {
	paths := map[string]*openapi.PathItem{}
	components := openapi.Components{
		Schemas: map[string]openapi.Schema{},
	}
	for _, r := range api.Resources {
		d := r.Schema
		// if it is a resource, add paths
		collection, parentPWPS := generateParentPatternsWithParams(r)
		// add an empty PathWithParam, if there are no parents.
		// This will add paths for the simple resource case.
		if len(*parentPWPS) == 0 {
			*parentPWPS = append(*parentPWPS, PathWithParams{
				Pattern: "", Params: []openapi.Parameter{},
			})
		}
		patterns := []string{}
		schemaRef := fmt.Sprintf("#/components/schemas/%v", r.Singular)
		singular := r.Singular
		// declare some commonly used objects, to be used later.
		bodyParam := openapi.RequestBody{
			Required: true,
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Ref: schemaRef,
					},
				},
			},
		}
		idParam := openapi.Parameter{
			In:       "path",
			Name:     singular,
			Required: true,
			Schema: &openapi.Schema{
				Type: "string",
			},
		}
		resourceResponse := openapi.Response{
			Description: "Successful response",
			Content: map[string]openapi.MediaType{
				"application/json": {
					Schema: &openapi.Schema{
						Ref: schemaRef,
					},
				},
			},
		}
		for _, pwp := range *parentPWPS {
			resourcePath := fmt.Sprintf("%s%s/{%s}", pwp.Pattern, collection, singular)
			patterns = append(patterns, resourcePath[1:])
			if r.ListMethod != nil {
				listPath := fmt.Sprintf("%s%s", pwp.Pattern, collection)
				responseProperties := map[string]openapi.Schema{
					"results": {
						Type: "array",
						Items: &openapi.Schema{
							Ref: schemaRef,
						},
					},
				}
				if r.ListMethod.HasUnreachableResources {
					responseProperties[constants.FIELD_UNREACHABLE_NAME] = openapi.Schema{
						Type: "array",
						Items: &openapi.Schema{
							Type: "string",
						},
					}
				}
				addMethodToPath(paths, listPath, "get", openapi.Operation{
					OperationID: r.Singular + ".list",
					Description: fmt.Sprintf("List method for %s", r.Singular),
					Parameters: append(pwp.Params,
						openapi.Parameter{
							In:       "query",
							Name:     constants.FIELD_MAX_PAGE_SIZE_NAME,
							Required: false,
							Schema: &openapi.Schema{
								Type: "integer",
							},
						},
						openapi.Parameter{
							In:       "query",
							Name:     constants.FIELD_PAGE_TOKEN_NAME,
							Required: false,
							Schema: &openapi.Schema{
								Type: "string",
							},
						},
					),
					Responses: map[string]openapi.Response{
						"200": {
							Description: "Successful response",
							Content: map[string]openapi.MediaType{
								"application/json": {
									Schema: &openapi.Schema{
										Type:       "object",
										Properties: responseProperties,
									},
								},
							},
						},
					},
				})
			}
			if r.CreateMethod != nil {
				createPath := fmt.Sprintf("%s%s", pwp.Pattern, collection)
				params := pwp.Params
				if !r.CreateMethod.SupportsUserSettableCreate {
					params = append(params, openapi.Parameter{
						In:       "query",
						Name:     "id",
						Required: false,
						Type:     "string",
					})
				}
				addMethodToPath(paths, createPath, "post", openapi.Operation{
					OperationID: r.Singular + ".create",
					Description: fmt.Sprintf("Create method for %s", r.Singular),
					Parameters:  params,
					RequestBody: &bodyParam,
					Responses: map[string]openapi.Response{
						"200": resourceResponse,
					},
				})
			}
			if r.GetMethod != nil {
				addMethodToPath(paths, resourcePath, "get", openapi.Operation{
					OperationID: r.Singular + ".get",
					Description: fmt.Sprintf("Get method for %s", r.Singular),
					Parameters:  append(pwp.Params, idParam),
					Responses: map[string]openapi.Response{
						"200": resourceResponse,
					},
				})
			}
			if r.UpdateMethod != nil {
				addMethodToPath(paths, resourcePath, "patch", openapi.Operation{
					OperationID: r.Singular + ".update",
					Description: fmt.Sprintf("Update method for %s", r.Singular),
					Parameters:  append(pwp.Params, idParam),
					RequestBody: &bodyParam,
					Responses: map[string]openapi.Response{
						"200": resourceResponse,
					},
				})
			}
			if r.DeleteMethod != nil {
				params := append(pwp.Params, idParam)
				if len(r.Children) > 0 {
					params = append(params, openapi.Parameter{
						In:       "query",
						Name:     constants.FIELD_FORCE_NAME,
						Required: false,
						Schema: &openapi.Schema{
							Type: "boolean",
						},
					})
				}
				addMethodToPath(paths, resourcePath, "delete", openapi.Operation{
					OperationID: r.Singular + ".delete",
					Description: fmt.Sprintf("Delete method for %s", r.Singular),
					Parameters:  params,
					Responses: map[string]openapi.Response{
						"204": {
							Description: "Successful response",
							Content: map[string]openapi.MediaType{
								"application/json": {
									Schema: &openapi.Schema{},
								},
							},
						},
					},
				})
			}
			if r.ApplyMethod != nil {
				addMethodToPath(paths, resourcePath, "put", openapi.Operation{
					OperationID: r.Singular + ".apply",
					Description: fmt.Sprintf("Apply method for %s", r.Singular),
					Parameters:  append(pwp.Params, idParam),
					RequestBody: &bodyParam,
					Responses: map[string]openapi.Response{
						"200": resourceResponse,
					},
				})
			}
			for _, custom := range r.CustomMethods {
				methodType := "get"
				if custom.Method == "POST" {
					methodType = "post"
				}
				cmPath := fmt.Sprintf("%s:%s", resourcePath, custom.Name)
				methodInfo := openapi.Operation{
					OperationID: r.Singular + ":" + custom.Name,
					Description: fmt.Sprintf("Custom method %s for %s", custom.Name, r.Singular),
					Parameters:  append(pwp.Params, idParam),
					Responses: map[string]openapi.Response{
						"200": resourceResponse,
					},
				}
				if custom.Method == "POST" {
					methodInfo.RequestBody = &openapi.RequestBody{
						Required: true,
						Content: map[string]openapi.MediaType{
							"application/json": {
								Schema: &openapi.Schema{},
							},
						},
					}
				}
				addMethodToPath(paths, cmPath, methodType, methodInfo)
			}
		}
		parents := []string{}
		for _, p := range r.Parents {
			parents = append(parents, p.Singular)
		}
		d.XAEPResource = &openapi.XAEPResource{
			Singular: r.Singular,
			Plural:   r.Plural,
			Patterns: patterns,
			Parents:  parents,
		}
		components.Schemas[r.Singular] = *d
	}
	for k, v := range api.Schemas {
		components.Schemas[k] = *v
	}
	openAPI := &openapi.OpenAPI{
		OpenAPI: "3.1.0",
		Servers: []openapi.Server{
			{URL: api.ServerURL},
		},
		Info: openapi.Info{
			Title:       api.Name,
			Version:     "version not set",
			Description: "An API for " + api.Name,
		},
		Paths:      paths,
		Components: components,
	}
	return openAPI, nil
}

// PathWithParams passes an http path
// with the OpenAPI parameters it contains.
// helpful to bundle them both when iterating.
type PathWithParams struct {
	Pattern string
	Params  []openapi.Parameter
}

// generate the x-aep-patterns for the parent resources, along with the patterns
// they need. Return a tuple of the collection name for the resource, and the
// patterns.
//
// This is helpful when you're constructing methods on resources with a parent.
//
// There are two algorithms that are used:
//
// 1. if PatternElems are present, then those will be used. This helps
// handle the situation where the resource structs were retrieved from a parsed
// OpenAPI definition, where the plural of the parents aren't necessarily clear,
// or the pattern element naming may not completely match the resource names.
//
// 2. Otherwise, we'll use the parent resources, and generate the collection
// names. This works for the case where the resource hierarchy is generated from
// scratch. This Algorithm will result in the fully AEP-compliant collection
// names.
func generateParentPatternsWithParams(r *Resource) (string, *[]PathWithParams) {
	// case 1: pattern elems are present, so we use them.
	// TODO(yft): support multiple patterns
	if len(r.PatternElems) > 0 {
		collection := fmt.Sprintf("/%s", r.PatternElems[len(r.PatternElems)-2])
		params := []openapi.Parameter{}
		for i := 0; i < len(r.PatternElems)-2; i += 2 {
			pElem := r.PatternElems[i+1]
			params = append(params, openapi.Parameter{
				In:       "path",
				Name:     pElem[1 : len(pElem)-1],
				Required: true,
				Schema: &openapi.Schema{
					Type: "string",
				},
			})
		}
		pattern := strings.Join(r.PatternElems[0:len(r.PatternElems)-2], "/")
		if pattern != "" {
			pattern = fmt.Sprintf("/%s", pattern)
		}
		return collection, &[]PathWithParams{
			{Pattern: pattern, Params: params},
		}
	}
	// case 2: no pattern elems, so we need to generate the collection names
	collection := fmt.Sprintf("/%s", CollectionName(r))
	pwps := []PathWithParams{}
	for _, parent := range r.Parents {
		singular := parent.Singular
		basePattern := fmt.Sprintf("/%s/{%s}", CollectionName(parent), singular)
		baseParam := openapi.Parameter{
			In:       "path",
			Name:     singular,
			Required: true,
			Schema: &openapi.Schema{
				Type: "string",
			},
		}
		if len(parent.Parents) == 0 {
			pwps = append(pwps, PathWithParams{
				Pattern: basePattern,
				Params:  []openapi.Parameter{baseParam},
			})
		} else {
			_, parentPWPS := generateParentPatternsWithParams(parent)
			for _, parentPWP := range *parentPWPS {
				params := append(parentPWP.Params, baseParam)
				pattern := fmt.Sprintf("%s%s", parentPWP.Pattern, basePattern)
				pwps = append(pwps, PathWithParams{Pattern: pattern, Params: params})
			}
		}
	}
	return collection, &pwps
}

func addMethodToPath(paths map[string]*openapi.PathItem, path, method string, methodInfo openapi.Operation) {
	methods, ok := paths[path]
	if !ok {
		methods = &openapi.PathItem{}
		paths[path] = methods
	}
	switch method {
	case "get":
		methods.Get = &methodInfo
	case "post":
		methods.Post = &methodInfo
	case "patch":
		methods.Patch = &methodInfo
	case "put":
		methods.Put = &methodInfo
	case "delete":
		methods.Delete = &methodInfo
	}
}
