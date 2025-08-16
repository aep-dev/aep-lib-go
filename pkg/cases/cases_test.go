package cases

import "testing"

func TestPascalCaseToKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single word", "User", "user"},
		{"multiple words", "UserProfile", "user-profile"},
		{"consecutive capitals", "UtilityAPIResponse", "utility-api-response"},
		{"consecutive capitals", "APIResponse", "api-response"},
		{"mixed case", "backgroundColor", "background-color"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := PascalCaseToKebabCase(tt.input)
			if got != tt.expected {
				t.Errorf("PascalCaseToKebabCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestKebabToCamelCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single word", "user", "User"},
		{"multiple words", "user-profile", "UserProfile"},
		{"consecutive hyphens", "api--response", "ApiResponse"},
		{"starting with hyphen", "-background", "Background"},
		{"ending with hyphen", "foreground-", "Foreground"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KebabToCamelCase(tt.input)
			if got != tt.expected {
				t.Errorf("KebabToCamelCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestKebabToSnakeCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single word", "user", "user"},
		{"multiple words", "user-profile", "user_profile"},
		{"consecutive hyphens", "api--response", "api__response"},
		{"starting with hyphen", "-background", "_background"},
		{"ending with hyphen", "foreground-", "foreground_"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := KebabToSnakeCase(tt.input)
			if got != tt.expected {
				t.Errorf("KebabToSnakeCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestSnakeToKebabCase(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty string", "", ""},
		{"single word", "user", "user"},
		{"multiple words", "user_profile", "user-profile"},
		{"consecutive underscores", "api__response", "api--response"},
		{"starting with underscore", "_background", "-background"},
		{"ending with underscore", "foreground_", "foreground-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SnakeToKebabCase(tt.input)
			if got != tt.expected {
				t.Errorf("SnakeToKebabCase(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestUpperFirst(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"single character", "a", "A"},
		{"word", "hello", "Hello"},
		{"already capitalized", "World", "World"},
		{"with spaces", "hello world", "Hello world"},
		{"with numbers", "1st", "1st"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UpperFirst(tt.input)
			if got != tt.expected {
				t.Errorf("UpperFirst(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
