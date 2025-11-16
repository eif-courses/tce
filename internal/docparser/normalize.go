package docparser

import (
	"regexp"
	"strings"
)

var reMultiSpace = regexp.MustCompile(`\s+`)
var reDash = regexp.MustCompile(`[–—]`)

// CleanDoc is used after normalization.
type CleanDoc struct {
	Paragraphs []string
}

func NormalizeText(input string) string {
	s := strings.TrimSpace(input)

	// unify dashes
	s = reDash.ReplaceAllString(s, "-")

	// collapse spaces
	s = reMultiSpace.ReplaceAllString(s, " ")

	// remove weird soft breaks
	s = strings.ReplaceAll(s, "\u000b", " ")

	return strings.TrimSpace(s)
}

func NormalizeDoc(raw *RawDoc) *CleanDoc {
	result := make([]string, 0, len(raw.Paragraphs))

	for _, p := range raw.Paragraphs {
		clean := NormalizeText(p)
		if clean != "" {
			result = append(result, clean)
		}
	}

	return &CleanDoc{Paragraphs: result}
}
