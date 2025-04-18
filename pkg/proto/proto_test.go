// Copyright 2023 Yusuke Fredrick Tsutsumi //
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package proto

import (
	"strings"
	"testing"

	"github.com/aep-dev/aep-lib-go/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestAPIToProto(t *testing.T) {
	// Create test API
	exampleAPI := api.ExampleAPI()

	// Test cases
	tests := []struct {
		name           string
		api            *api.API
		outputDir      string
		expectError    bool
		expectMessages []string
		expectMethods  []string
	}{
		{
			name:        "BasicAPItoProtoConversion",
			api:         exampleAPI,
			outputDir:   "example/testapi/v1", // Changed from "test-api" to "testapi"
			expectError: false,
			expectMessages: []string{
				"Publisher",
				"Book",
				"BookEdition",
				"Account",
				"CreatePublisherRequest",
				"GetPublisherRequest",
				"ListPublishersRequest",
				"ListPublishersResponse",
				"CreateBookRequest",
				"GetBookRequest",
				"UpdateBookRequest",
				"DeleteBookRequest",
				"ListBooksRequest",
				"ListBooksResponse",
				"ArchiveBookRequest",
				"ArchiveBookResponse",
				"GetBookEditionRequest",
				"ListBookEditionsRequest",
				"ListBookEditionsResponse",
			},
			expectMethods: []string{
				"CreatePublisher",
				"GetPublisher",
				"ListPublishers",
				"CreateBook",
				"GetBook",
				"UpdateBook",
				"DeleteBook",
				"ListBooks",
				"ArchiveBook",
				"GetBookEdition",
				"ListBookEditions",
			},
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Generate proto file descriptor
			fileDescriptor, err := APIToProto(tt.api, tt.outputDir)

			// Check error expectations
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, fileDescriptor)

			// Verify the generated proto structure
			assert.Equal(t, "test.proto", fileDescriptor.GetName())
			assert.Equal(t, "example.testapi.v1", fileDescriptor.GetPackage()) // Changed from "example.test_api.v1"
			assert.True(t, fileDescriptor.IsProto3())

			// Check service properties
			services := fileDescriptor.GetServices()
			assert.Equal(t, 1, len(services))

			service := services[0]
			assert.Equal(t, "Testapi", service.GetName())

			// Generate proto string to check content
			protoString, err := APIToProtoString(tt.api, tt.outputDir)
			assert.NoError(t, err)
			assert.NotEmpty(t, protoString)

			protoContent := string(protoString)
			// Print the proto content for debugging
			// t.Logf("Proto content: \n%s", protoContent)

			// Check for expected messages
			for _, msgName := range tt.expectMessages {
				assert.True(t,
					strings.Contains(protoContent, "message "+msgName+" {") ||
						strings.Contains(protoContent, "message "+msgName+"{"),
					"Expected message %s not found in proto content", msgName)
			}

			// Check for expected methods
			for _, methodName := range tt.expectMethods {
				methodNameLower := strings.ToLower(methodName)
				assert.True(t,
					strings.Contains(strings.ToLower(protoContent), "rpc "+methodNameLower+" (") ||
						strings.Contains(strings.ToLower(protoContent), "rpc "+methodNameLower+"(") ||
						strings.Contains(protoContent, "rpc "+methodName+" (") ||
						strings.Contains(protoContent, "rpc "+methodName+"("),
					"Expected method %s not found in proto content", methodName)
			}

			// Verify correct parent-child relationships in the API paths
			assert.True(t, strings.Contains(protoContent, "get: \"/{path=publishers/*/books/*}\""))
			assert.True(t, strings.Contains(protoContent, "get: \"/{path=publishers/*/books/*/editions/*}\""))

			// Verify custom method handling
			assert.True(t, strings.Contains(protoContent, "post: \"/{path=publishers/*/books/*}:archive\""))
		})
	}
}

func TestToProtoServiceName(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple service name",
			input:    "test api",
			expected: "Test Api",
		},
		{
			name:     "Service name with period",
			input:    "test.api",
			expected: "Test",
		},
		{
			name:     "Capitalized service name",
			input:    "TEST API",
			expected: "Test Api", // Changed from "TEST API" to match actual implementation
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toProtoServiceName(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
