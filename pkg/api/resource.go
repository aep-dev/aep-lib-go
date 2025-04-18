package api

import (
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

type Resource struct {
	Singular string      `json:"singular"`
	Plural   string      `json:"plural"`
	Parents  []*Resource `json:"parents,omitempty"`
	// Children is populated on load
	Children      []*Resource     `json:"-"`
	PatternElems  []string        `json:"-"` // TOO(yft): support multiple patterns
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
	return strings.Join(r.PatternElems, "/")
}
