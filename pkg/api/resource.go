package api

import (
	"strings"

	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

type Resource struct {
	Singular      string
	Plural        string
	Parents       []*Resource
	Children      []*Resource
	PatternElems  []string // TOO(yft): support multiple patterns
	Schema        *openapi.Schema
	GetMethod     *GetMethod
	ListMethod    *ListMethod
	ApplyMethod   *ApplyMethod
	CreateMethod  *CreateMethod
	UpdateMethod  *UpdateMethod
	DeleteMethod  *DeleteMethod
	CustomMethods []*CustomMethod
}

type CreateMethod struct {
	SupportsUserSettableCreate bool
	IsLongRunning              bool
}

type ApplyMethod struct {
	IsLongRunning bool
}

type GetMethod struct {
}

type UpdateMethod struct {
	IsLongRunning bool
}

type ListMethod struct {
	HasUnreachableResources bool
	SupportsFilter          bool
	SupportsSkip            bool
}

type DeleteMethod struct {
	IsLongRunning bool
}

type CustomMethod struct {
	Name          string
	Method        string
	Request       *openapi.Schema
	Response      *openapi.Schema
	IsLongRunning bool
}

func (r *Resource) GetPattern() string {
	return strings.Join(r.PatternElems, "/")
}
