package cases

import (
	"strings"
)

func PascalCaseToKebabCase(s string) string {
	// a uppercase char is a sign of a deiimiter
	// except for acronyms. if it's an acronym, then you delimit
	// on the previous character.
	delimiterIndices := []int{}
	previousIsUpper := false
	isAcronym := false
	for i, r := range s {
		if 'A' <= r && r <= 'Z' {
			if previousIsUpper && !isAcronym {
				isAcronym = true
				// assuming multiple uppers in sequence is an uppercase character.
				// so there should be a delimiter there.
				delimiterIndices = append(delimiterIndices, i-1)
			}
			previousIsUpper = true
		} else {
			if previousIsUpper {
				delimiterIndices = append(delimiterIndices, i-1)
			}
			isAcronym = false
			previousIsUpper = false
		}
	}
	parts := []string{}
	prevDelimIndex := 0
	for _, d := range delimiterIndices {
		if d != prevDelimIndex {
			parts = append(parts, s[prevDelimIndex:d])
			prevDelimIndex = d
		}
	}
	parts = append(parts, s[prevDelimIndex:])
	return strings.ToLower(strings.Join(parts, "-"))
}

func KebabToCamelCase(s string) string {
	parts := strings.Split(s, "-")
	for i := range parts {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(string(parts[i][0])) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

func KebabToSnakeCase(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}

func UpperFirst(s string) string {
	return strings.ToUpper(s[:1]) + s[1:]
}
