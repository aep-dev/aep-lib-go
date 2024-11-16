package api

import (
	"fmt"
	"strings"
)

// return the collection name of the resource, but deduplicate
// the name of the previous parent
// e.g:
// - book-editions becomes editions under the parent resource book.
func CollectionName(r *Resource) string {
	collectionName := r.Plural
	if len(r.Parents) > 0 {
		parent := r.Parents[0].Singular
		// if collectionName has a prefix of parent, remove it
		if strings.HasPrefix(collectionName, parent) {
			collectionName = strings.TrimPrefix(collectionName, parent+"-")
		}
	}
	return collectionName
}

// GeneratePatternStrings generates the pattern strings for a resource
// TODO(yft): support multiple parents
func GeneratePatternStrings(r *Resource) []string {

	// Base pattern without params
	pattern := fmt.Sprintf("%v/{%v}", CollectionName(r), r.Singular)
	if len(r.Parents) > 0 {
		parentParts := GeneratePatternStrings(r.Parents[0])
		pattern = fmt.Sprintf("%v/%v", parentParts[0], pattern)
	}
	return []string{pattern}
}
