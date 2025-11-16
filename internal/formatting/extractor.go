package formatting

// ParagraphFormat describes formatting-related info for a paragraph.
// Later you can fill this from DOCX (alignment, spacing, font size, etc.).
type ParagraphFormat struct {
	Align       string  // "left", "right", "center", "justify", ""
	LineSpacing float64 // e.g. 1.0, 1.5, 2.0
	IsTitle     bool
}

// For now we just return an empty format map.
func ExtractFormats(paragraphIDs []string) map[string]ParagraphFormat {
	m := make(map[string]ParagraphFormat, len(paragraphIDs))
	for _, id := range paragraphIDs {
		m[id] = ParagraphFormat{}
	}
	return m
}
