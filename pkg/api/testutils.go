package api

import (
	"github.com/aep-dev/aep-lib-go/pkg/openapi"
)

func ExampleAPI() *API {
	// Create publisher resource
	publisher := &Resource{
		Singular: "publisher",
		Plural:   "publishers",
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"title": {Type: "string", XAEPFieldNumber: 1},
				"id":    {Type: "string", XAEPFieldNumber: 2},
			},
		},
		Methods: Methods{
			List: &ListMethod{},
			Get:  &GetMethod{},
			Create: &CreateMethod{
				SupportsUserSettableCreate: true,
			},
		},
	}

	// Create book resource
	book := &Resource{
		Singular: "book",
		Plural:   "books",
		Parents:  []*Resource{publisher},
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"name": {Type: "string", XAEPFieldNumber: 1},
				"id":   {Type: "string", XAEPFieldNumber: 2},
			},
		},
		Methods: Methods{
			List: &ListMethod{
				HasUnreachableResources: true,
				SupportsFilter:          true,
				SupportsSkip:            true,
			},
			Get: &GetMethod{},
			Create: &CreateMethod{
				SupportsUserSettableCreate: true,
			},
			Update: &UpdateMethod{},
			Delete: &DeleteMethod{},
		},
		CustomMethods: []*CustomMethod{
			{
				Name:   "archive",
				Method: "POST",
				Request: &openapi.Schema{
					Type:       "object",
					Properties: map[string]openapi.Schema{},
				},
				Response: &openapi.Schema{
					Type: "object",
					Properties: map[string]openapi.Schema{
						"archived": {Type: "boolean", XAEPFieldNumber: 1},
					},
				},
			},
		},
	}
	publisher.Children = append(publisher.Children, book)

	// Create book-edition resource
	bookEdition := &Resource{
		Singular: "book-edition",
		Plural:   "book-editions",
		Parents:  []*Resource{book},
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"date": {Type: "string", XAEPFieldNumber: 1},
			},
		},
		Methods: Methods{
			List: &ListMethod{},
			Get:  &GetMethod{},
		},
	}
	book.Children = append(book.Children, bookEdition)

	// Return the complete example API
	return &API{
		Name:      "TestAPI", // Changed from "Test API" to "TestAPI" (removed space)
		ServerURL: "https://api.example.com",
		Contact: &Contact{
			Name:  "John Doe",
			Email: "john.doe@example.com",
			URL:   "https://example.com",
		},
		Schemas: map[string]*openapi.Schema{
			"account": {
				Type: "object",
				Properties: map[string]openapi.Schema{
					"name": {Type: "string", XAEPFieldNumber: 1},
				},
			},
		},
		Resources: map[string]*Resource{
			"book":         book,
			"book-edition": bookEdition,
			"publisher":    publisher,
		},
	}
}
