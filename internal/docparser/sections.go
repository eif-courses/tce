package docparser

import (
	"fmt"
	"strings"
)

// ParseSections converts a CleanDoc into Sections + Paragraphs.
func ParseSections(doc *CleanDoc) ([]Section, []Paragraph) {
	var (
		sections       []Section
		paragraphs     []Paragraph
		currentSection = "unknown"
		index          = 1
	)

	sectionMap := map[string]string{
		"ĮVADAS":                  "intro",
		"PRAKTIKOS UŽDUOTIS":      "task",
		"PRAKTIKOS UŽDUOTIS IR":   "task",
		"FUNKCINIAI REIKALAVIMAI": "fr",
		"UŽDUOTIES ANALIZĖ":       "analysis",
	}

	requirements := map[string]string{
		"intro":    "2–5 psl.",
		"task":     "3–5 psl.",
		"fr":       "≥3 psl.",
		"analysis": "≥10 psl.",
		"unknown":  "Nenurodyta",
	}

	// build section list
	for label, id := range sectionMap {
		sections = append(sections, Section{
			ID:          id,
			Label:       label,
			Requirement: requirements[id],
			Issues:      0,
		})
	}
	sections = append(sections, Section{
		ID:          "unknown",
		Label:       "Nepriskirta",
		Requirement: requirements["unknown"],
		Issues:      0,
	})

	for _, text := range doc.Paragraphs {
		upper := strings.ToUpper(strings.TrimSpace(text))

		// check if this is a heading
		for heading, id := range sectionMap {
			if strings.HasPrefix(upper, heading) {
				currentSection = id
				goto nextParagraph
			}
		}

	nextParagraph:
		paragraphs = append(paragraphs, Paragraph{
			ID:        fmt.Sprintf("p-%d", index),
			SectionID: currentSection,
			Text:      text,
		})
		index++
	}

	return sections, paragraphs
}
