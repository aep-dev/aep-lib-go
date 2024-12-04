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
}

type ApplyMethod struct {
}

type GetMethod struct {
}

type UpdateMethod struct {
}

type ListMethod struct {
	HasUnreachableResources bool
}

type DeleteMethod struct {
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
