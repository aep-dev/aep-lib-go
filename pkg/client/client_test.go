package client

import (
	"context"
	"net/http"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/jarcoal/httpmock"
)

func TestCreate(t *testing.T) {
	// Create a test server
	httpmock.Activate()
	httpmock.RegisterResponder("POST", "http://localhost:8081/publishers/my-pub/books",
		httpmock.NewStringResponder(200, "{\"path\":\"/publishers/my-pub/books/1\"}"))

	// Create a test resource
	r := &api.Resource{
		CreateMethod: &api.CreateMethod{
			SupportsUserSettableCreate: false,
		},
		PatternElems: []string{"publishers", "{publisher}", "books", "{book}"},
	}

	// Create a test context
	ctx := context.Background()

	// Create a test body
	body := map[string]interface{}{
		"price": "1",
	}

	parameters := map[string]string{
		"publisher": "my-pub",
	}

	// Call the Create method
	data, err := Create(ctx, r, http.DefaultClient, "http://localhost:8081/", body, parameters)
	if err != nil {
		t.Fatal(err)
	}

	// Check the response
	if data["path"] != "/publishers/my-pub/books/1" {
		t.Errorf("expected id to be '/publishers/my-pub/books/1', got '%v'", data["path"])
	}
}

func TestCreateWithUserSpecifiedId(t *testing.T) {
	// Create a test server
	httpmock.Activate()
	httpmock.RegisterResponder("POST", "http://localhost:8081/publishers/my-pub/books?id=my-book",
		httpmock.NewStringResponder(200, "{\"path\":\"/publishers/my-pub/books/my-book\"}"))

	// Create a test resource
	r := &api.Resource{
		CreateMethod: &api.CreateMethod{
			SupportsUserSettableCreate: true,
		},
		PatternElems: []string{"publishers", "{publisher}", "books", "{book}"},
	}

	// Create a test context
	ctx := context.Background()

	// Create a test body
	body := map[string]interface{}{
		"price": "1",
		"id":    "my-book",
	}

	parameters := map[string]string{
		"publisher": "my-pub",
	}

	// Call the Create method
	data, err := Create(ctx, r, http.DefaultClient, "http://localhost:8081/", body, parameters)
	if err != nil {
		t.Fatal(err)
	}

	// Check the response
	if data["path"] != "/publishers/my-pub/books/my-book" {
		t.Errorf("expected id to be '/publishers/my-pub/books/my-book', got '%v'", data["path"])
	}
}

func TestGet(t *testing.T) {
	// Create a test server
	httpmock.Activate()
	httpmock.RegisterResponder("GET", "http://localhost:8081/publishers/my-pub/books/1",
		httpmock.NewStringResponder(200, "{\"path\":\"/publishers/my-pub/books/1\"}"))

	// Create a test context
	ctx := context.Background()

	// Call the Read method
	data, err := Get(ctx, http.DefaultClient, "http://localhost:8081", "/publishers/my-pub/books/1")
	if err != nil {
		t.Fatal(err)
	}

	// Check the response
	if data["path"] != "/publishers/my-pub/books/1" {
		t.Errorf("expected id to be '/publishers/my-pub/books/1', got '%v'", data["path"])
	}
}

func TestDelete(t *testing.T) {
	// Create a test server
	httpmock.Activate()
	httpmock.RegisterResponder("DELETE", "http://localhost:8081/publishers/my-pub/books/1",
		httpmock.NewStringResponder(200, ""))

	// Create a test context
	ctx := context.Background()

	// Call the Delete method
	err := Delete(ctx, http.DefaultClient, "http://localhost:8081", "/publishers/my-pub/books/1")
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdate(t *testing.T) {
	// Create a test server
	httpmock.Activate()
	httpmock.RegisterResponder("PATCH", "http://localhost:8081/publishers/my-pub/books/1",
		httpmock.NewStringResponder(200, "{\"path\":\"/publishers/my-pub/books/1\", \"price\":\"2\"}"))

	// Create a test context
	ctx := context.Background()

	// Create a test body
	body := map[string]interface{}{
		"path":  "/publishers/my-pub/books/1",
		"price": "2",
	}

	// Call the Update method
	err := Update(ctx, http.DefaultClient, "http://localhost:8081", "/publishers/my-pub/books/1", body)
	if err != nil {
		t.Fatal(err)
	}
}
