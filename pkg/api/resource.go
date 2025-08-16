package api

import (
	"fmt"
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/cases"
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

type Resource struct {
	Singular        string      `json:"singular"`
	Plural          string      `json:"plural"`
	Parents         []string    `json:"parents,omitempty"`
	parentResources []*Resource `json:"-"`
	// Children is populated on load
	Children     []*Resource `json:"-"`
	patternElems []string    `json:"-"` // TOO(yft): support multiple patterns
	// the API reference is used to retrieve things like the parent resources.
	API           *API            `json:"-"`
	Schema        *openapi.Schema `json:"schema,omitempty"`
	Methods       Methods         `json:"methods,omitempty"`
	CustomMethods []*CustomMethod `json:"custom_methods,omitempty"`
}

type Methods struct {
	Get    *GetMethod    `json:"get,omitempty"`
	List   *ListMethod   `json:"list,omitempty"`
	Apply  *ApplyMethod  `json:"apply,omitempty"`
	Create *CreateMethod `json:"create,omitempty"`
	Update *UpdateMethod `json:"update,omitempty"`
	Delete *DeleteMethod `json:"delete,omitempty"`
}

type CreateMethod struct {
	SupportsUserSettableCreate bool `json:"supports_user_settable_create"`
	IsLongRunning              bool `json:"is_long_running"`
}

type ApplyMethod struct {
	IsLongRunning bool `json:"is_long_running"`
}

type GetMethod struct {
}

type UpdateMethod struct {
	IsLongRunning bool `json:"is_long_running"`
}

type ListMethod struct {
	HasUnreachableResources bool `json:"has_unreachable_resources"`
	SupportsFilter          bool `json:"supports_filter"`
	SupportsSkip            bool `json:"supports_skip"`
}

type DeleteMethod struct {
	IsLongRunning bool `json:"is_long_running"`
}

type CustomMethod struct {
	Name          string
	Method        string
	Request       *openapi.Schema
	Response      *openapi.Schema
	IsLongRunning bool `json:"is_long_running"`
}

func (r *Resource) GetPattern() string {
	return strings.Join(r.PatternElems(), "/")
}

// return the parent resources of the resource.
//
// This function should only be called until after
// the API has been validated.
func (r *Resource) ParentResources() []*Resource {
	if r.parentResources == nil {
		r.parentResources = []*Resource{}
		for _, parent := range r.Parents {
			parentResource, ok := r.API.Resources[parent]
			if !ok {
				panic(fmt.Sprintf("parent resource %s not found", parent))
			}
			r.parentResources = append(r.parentResources, parentResource)
		}
	}
	return r.parentResources
}

// return the collection name of the resource, but deduplicate
// the name of the previous parent
// e.g:
// - book-editions becomes editions under the parent resource book.
func CollectionName(r *Resource) string {
	collectionName := r.Plural
	if len(r.Parents) > 0 {
		parent := r.ParentResources()[0].Singular
		// if collectionName has a prefix of parent, remove it
		if strings.HasPrefix(collectionName, parent) {
			collectionName = strings.TrimPrefix(collectionName, parent+"-")
		}
	}
	// Convert to kebab-case for path elements
	return cases.SnakeToKebabCase(collectionName)
}

// GeneratePatternStrings generates the pattern strings for a resource
// TODO(yft): support multiple parents
func (r *Resource) PatternElems() []string {
	if len(r.patternElems) == 0 {
		// Convert kebab-case singular to snake_case for path variables
		singularSnake := cases.KebabToSnakeCase(r.Singular)
		// Base pattern without params
		patternElems := []string{CollectionName(r), fmt.Sprintf("{%s_id}", singularSnake)}
		if len(r.Parents) > 0 {
			patternElems = append(
				r.ParentResources()[0].PatternElems(),
				patternElems...,
			)
		}
		r.patternElems = patternElems
	}
	return r.patternElems
}
