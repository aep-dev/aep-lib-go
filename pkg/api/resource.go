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
	Parameters                 []openapi.Parameter
}

type ApplyMethod struct {
	Parameters []openapi.Parameter
}

type GetMethod struct {
	Parameters []openapi.Parameter
}

type UpdateMethod struct {
	Parameters []openapi.Parameter
}

type ListMethod struct {
	HasUnreachableResources bool
	SupportsFilter          bool
	SupportsSkip            bool
	Parameters              []openapi.Parameter
}

type DeleteMethod struct {
	Parameters []openapi.Parameter
}

type CustomMethod struct {
	Name     string
	Method   string
	Request  *openapi.Schema
	Response *openapi.Schema
}

func (r *Resource) GetPattern() string {
	return strings.Join(r.PatternElems, "/")
}
