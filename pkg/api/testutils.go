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
				SupportsUserSettableCreate: false,
			},
		},
	}

	// Create book resource
	book := &Resource{
		Singular:        "book",
		Plural:          "books",
		Parents:         []string{"publisher"},
		parentResources: []*Resource{publisher},
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

	// Resource to test operation logic
	tome := &Resource{
		Singular:        "tome",
		Plural:          "tomes",
		Parents:         []string{"publisher"},
		parentResources: []*Resource{publisher},
		Schema: &openapi.Schema{
			Type: "object",
			Properties: map[string]openapi.Schema{
				"name": {Type: "string", XAEPFieldNumber: 1},
				"id":   {Type: "string", XAEPFieldNumber: 2},
			},
		},
		Methods: Methods{
			List: &ListMethod{},
			Get:  &GetMethod{},
			Apply: &ApplyMethod{
				IsLongRunning: true,
			},
			Create: &CreateMethod{
				IsLongRunning: true,
			},
			Update: &UpdateMethod{
				IsLongRunning: true,
			},
			Delete: &DeleteMethod{
				IsLongRunning: true,
			},
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
				IsLongRunning: true,
			},
		},
	}
	publisher.Children = append(publisher.Children, tome)

	// Create book_edition resource
	bookEdition := &Resource{
		Singular:        "book_edition",
		Plural:          "book_editions",
		Parents:         []string{"book"},
		parentResources: []*Resource{book},
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
	api := &API{
		Name:      "TestAPI",
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
			"book_edition": bookEdition,
			"publisher":    publisher,
			"operation":    OperationResourceWithDefaults(),
			"tome":         tome,
		},
	}
	if err := AddImplicitFieldsAndValidate(api); err != nil {
		panic(err)
	}
	return api
}
