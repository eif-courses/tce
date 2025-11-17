package docparser

import (
	"encoding/json"
	"fmt"
	"strings"
)

// ParseSections converts a CleanDoc into Sections + Paragraphs.
func ParseSections(doc *CleanDoc, lang string) ([]Section, []Paragraph) {
	var (
		sections       []Section
		paragraphs     []Paragraph
		currentSection = "unknown"
		index          = 1
	)

	// Language-aware heading map
	sectionMap := make(map[string]string)

	lang = strings.ToLower(lang)
	switch lang {
	case "en":
		sectionMap["INTRODUCTION"] = "intro"
		sectionMap["TASK DESCRIPTION"] = "task"
		sectionMap["PRACTICE TASK"] = "task"
		sectionMap["FUNCTIONAL REQUIREMENTS"] = "fr"
		sectionMap["TASK ANALYSIS"] = "analysis"
		sectionMap["PROBLEM ANALYSIS"] = "analysis"
	default: // lt
		sectionMap["ĮVADAS"] = "intro"
		sectionMap["PRAKTIKOS UŽDUOTIS"] = "task"
		sectionMap["PRAKTIKOS UŽDUOTIS IR"] = "task"
		sectionMap["FUNKCINIAI REIKALAVIMAI"] = "fr"
		sectionMap["UŽDUOTIES ANALIZĖ"] = "analysis"
	}

	requirements := map[string]string{
		"intro":    "2–5 psl.",
		"task":     "3–5 psl.",
		"fr":       "≥3 psl.",
		"analysis": "≥10 psl.",
		"unknown":  "Nenurodyta",
	}

	// build sections list
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

		// heading detection
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

// ParseContentBlocks assigns section IDs to content blocks
func ParseContentBlocks(blocks []ContentBlock, lang string) []ContentBlock {
	if len(blocks) == 0 {
		return blocks
	}

	// Language-aware heading map
	sectionMap := make(map[string]string)

	lang = strings.ToLower(lang)
	switch lang {
	case "en":
		sectionMap["INTRODUCTION"] = "intro"
		sectionMap["TASK DESCRIPTION"] = "task"
		sectionMap["PRACTICE TASK"] = "task"
		sectionMap["FUNCTIONAL REQUIREMENTS"] = "fr"
		sectionMap["TASK ANALYSIS"] = "analysis"
		sectionMap["PROBLEM ANALYSIS"] = "analysis"
	default: // lt
		sectionMap["ĮVADAS"] = "intro"
		sectionMap["PRAKTIKOS UŽDUOTIS"] = "task"
		sectionMap["PRAKTIKOS UŽDUOTIS IR"] = "task"
		sectionMap["FUNKCINIAI REIKALAVIMAI"] = "fr"
		sectionMap["UŽDUOTIES ANALIZĖ"] = "analysis"
	}

	currentSection := "intro" // Default to intro

	for i := range blocks {
		// Check if this is a heading that matches a section
		if blocks[i].Type == "heading" {
			// Try to extract the heading text
			var headingText string
			if hc, ok := blocks[i].Content.(HeadingContent); ok {
				headingText = strings.ToUpper(strings.TrimSpace(hc.Text))
			} else if data, err := json.Marshal(blocks[i].Content); err == nil {
				var hc HeadingContent
				if err := json.Unmarshal(data, &hc); err == nil {
					headingText = strings.ToUpper(strings.TrimSpace(hc.Text))
				}
			}

			// Check if this heading indicates a section change
			for heading, id := range sectionMap {
				if strings.Contains(headingText, heading) {
					currentSection = id
					break
				}
			}
		}

		blocks[i].SectionID = currentSection
	}

	return blocks
}
