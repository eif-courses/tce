package analyzer

import (
	"regexp"
	"strings"

	"github.com/eif-courses/tce/internal/config"
)

//
// ======= YAML-BASED RULES =======
// These rules dynamically use the loaded YAML configuration
//

// ruleForbiddenPhrases checks for globally forbidden phrases
type ruleForbiddenPhrases struct{}

func (ruleForbiddenPhrases) ID() string { return "forbidden_phrases_global" }

func (ruleForbiddenPhrases) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	forbiddenPhrases := config.Current.ContentRules.ForbiddenGlobalPhrases
	if len(forbiddenPhrases) == 0 {
		return nil
	}

	for _, p := range paragraphs {
		lowerText := strings.ToLower(p.Text)

		for _, phrase := range forbiddenPhrases {
			if strings.Contains(lowerText, strings.ToLower(phrase)) {
				sectionLabel := getSectionLabel(sections, p.SectionID)
				severity := config.Current.Severity.ForbiddenPhrase
				if severity == "" {
					severity = "critical"
				}

				comments = append(comments, Comment{
					ID:           "R-FORBIDDEN-PHRASE-" + p.ID,
					ParagraphID:  p.ID,
					Category:     "content",
					Severity:     severity,
					SectionLabel: sectionLabel,
					Title:        "Netinkamas formulavimas",
					Message:      "Tekste rasta netinkama fraze: \"" + phrase + "\". Tai priestarauja metodiniams reikalavimams.",
				})
				break // Only report once per paragraph
			}
		}
	}

	return comments
}

// ruleSectionForbiddenPhrases checks for forbidden phrases within specific sections
type ruleSectionForbiddenPhrases struct{}

func (ruleSectionForbiddenPhrases) ID() string { return "section_forbidden_phrases" }

func (ruleSectionForbiddenPhrases) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	for _, section := range config.Current.Sections {
		if len(section.ForbiddenPhrases) == 0 && len(section.ForbiddenPatterns) == 0 {
			continue
		}

		// Get paragraphs for this section
		sectionParas := filterParagraphsBySection(paragraphs, section.ID)
		if len(sectionParas) == 0 {
			continue
		}

		// Check forbidden phrases
		for _, para := range sectionParas {
			lowerText := strings.ToLower(para.Text)

			for _, phrase := range section.ForbiddenPhrases {
				if strings.Contains(lowerText, strings.ToLower(phrase)) {
					comments = append(comments, Comment{
						ID:           "R-SEC-FORBIDDEN-" + section.ID + "-" + para.ID,
						ParagraphID:  para.ID,
						Category:     "content",
						Severity:     "major",
						SectionLabel: section.Label,
						Title:        "Netinkamas turinys skyriuje",
						Message:      "Skyriuje \"" + section.Label + "\" rasta netinkama fraze: \"" + phrase + "\".",
					})
					break
				}
			}

			// Check forbidden patterns (regex)
			for _, pattern := range section.ForbiddenPatterns {
				matched, err := regexp.MatchString(pattern, para.Text)
				if err == nil && matched {
					comments = append(comments, Comment{
						ID:           "R-SEC-PATTERN-" + section.ID + "-" + para.ID,
						ParagraphID:  para.ID,
						Category:     "content",
						Severity:     "major",
						SectionLabel: section.Label,
						Title:        "Netinkamas turinys skyriuje",
						Message:      "Skyriuje \"" + section.Label + "\" rastas netinkamas sablonas pagal metodinius reikalavimus.",
					})
					break
				}
			}
		}
	}

	return comments
}

// ruleMustContainKeywords checks if required sections contain mandatory keywords
type ruleMustContainKeywords struct{}

func (ruleMustContainKeywords) ID() string { return "must_contain_keywords" }

func (ruleMustContainKeywords) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	for _, section := range config.Current.Sections {
		if len(section.MustContainKeywords) == 0 || !section.Required {
			continue
		}

		sectionParas := filterParagraphsBySection(paragraphs, section.ID)
		if len(sectionParas) == 0 {
			continue // Missing section handled by another rule
		}

		// Check if ANY of the keywords are present
		foundAny := false
		fullText := strings.ToLower(strings.Join(extractTexts(sectionParas), " "))

		for _, keyword := range section.MustContainKeywords {
			if strings.Contains(fullText, strings.ToLower(keyword)) {
				foundAny = true
				break
			}
		}

		if !foundAny {
			// Report on first paragraph of section
			firstPara := sectionParas[0]
			keywordList := strings.Join(section.MustContainKeywords, "\", \"")

			comments = append(comments, Comment{
				ID:           "R-KEYWORD-MISSING-" + section.ID,
				ParagraphID:  firstPara.ID,
				Category:     "content",
				Severity:     "major",
				SectionLabel: section.Label,
				Title:        "Truksta butiniu elementu",
				Message:      "Skyriuje \"" + section.Label + "\" nerastas nei vienas is butiniu elementu: \"" + keywordList + "\".",
			})
		}
	}

	return comments
}

// ruleRequireCitations checks if sections that require citations have them
type ruleRequireCitations struct{}

func (ruleRequireCitations) ID() string { return "require_citations" }

func (ruleRequireCitations) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	if len(config.Current.References.IntextPatterns) == 0 {
		return nil
	}

	for _, section := range config.Current.Sections {
		if !section.RequireCitations {
			continue
		}

		sectionParas := filterParagraphsBySection(paragraphs, section.ID)
		if len(sectionParas) == 0 {
			continue
		}

		// Check if ANY citation pattern is found
		foundCitation := false
		for _, para := range sectionParas {
			for _, pattern := range config.Current.References.IntextPatterns {
				matched, err := regexp.MatchString(pattern, para.Text)
				if err == nil && matched {
					foundCitation = true
					break
				}
			}
			if foundCitation {
				break
			}
		}

		if !foundCitation {
			firstPara := sectionParas[0]
			severity := config.Current.Severity.MissingCitationInTheory
			if severity == "" {
				severity = "major"
			}

			comments = append(comments, Comment{
				ID:           "R-CITATION-MISSING-" + section.ID,
				ParagraphID:  firstPara.ID,
				Category:     "content",
				Severity:     severity,
				SectionLabel: section.Label,
				Title:        "Truksta saltiniu nuorodu",
				Message:      "Skyriuje \"" + section.Label + "\" nerastos saltiniu nuorodos. Analitiniame skyriuje butina remtis moksline literatura.",
			})
		}
	}

	return comments
}

// ruleStrictSectionOrder checks section order according to required_order in YAML
type ruleStrictSectionOrder struct{}

func (ruleStrictSectionOrder) ID() string { return "strict_section_order" }

func (ruleStrictSectionOrder) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	requiredOrder := config.Current.Structure.RequiredOrder
	if len(requiredOrder) == 0 {
		return nil
	}

	// Map section IDs to their first paragraph index
	firstIndex := make(map[string]int)
	for i, p := range paragraphs {
		if _, seen := firstIndex[p.SectionID]; !seen {
			firstIndex[p.SectionID] = i
		}
	}

	// Check order
	for i := 0; i < len(requiredOrder)-1; i++ {
		currentID := requiredOrder[i]
		nextID := requiredOrder[i+1]

		posA, okA := firstIndex[currentID]
		posB, okB := firstIndex[nextID]

		if !okA || !okB {
			continue // Section missing or not found
		}

		if posA > posB {
			para := paragraphs[posB]
			sectionLabel := getSectionLabel(sections, nextID)
			severity := config.Current.Severity.WrongOrder
			if severity == "" {
				severity = "major"
			}

			comments = append(comments, Comment{
				ID:           "R-ORDER-" + currentID + "-" + nextID,
				ParagraphID:  para.ID,
				Category:     "structure",
				Severity:     severity,
				SectionLabel: sectionLabel,
				Title:        "Netinkama skyriu tvarka",
				Message:      "Skyriai isdelstyti netinkama tvarka. Pagal metodinius reikalavimus \"" + nextID + "\" turi buti po \"" + currentID + "\".",
			})
		}
	}

	return comments
}

// ruleMaxPages checks if sections exceed maximum page limits
type ruleMaxPages struct{}

func (ruleMaxPages) ID() string { return "max_pages" }

func (ruleMaxPages) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	for _, section := range config.Current.Sections {
		if section.MaxPages <= 0 {
			continue
		}

		sectionParas := filterParagraphsBySection(paragraphs, section.ID)
		if len(sectionParas) == 0 {
			continue
		}

		// Rough estimate: 3 paragraphs per page
		estimatedPages := len(sectionParas) / 3
		if estimatedPages > section.MaxPages {
			firstPara := sectionParas[0]

			comments = append(comments, Comment{
				ID:           "R-MAXPAGE-" + section.ID,
				ParagraphID:  firstPara.ID,
				Category:     "structure",
				Severity:     "minor",
				SectionLabel: section.Label,
				Title:        "Skyrius per ilgas",
				Message:      "Skyrius \"" + section.Label + "\" virsija rekomenduojama ilgi.",
			})
		}
	}

	return comments
}

// ruleCodeBlocks detects code blocks and validates their placement
type ruleCodeBlocks struct{}

func (ruleCodeBlocks) ID() string { return "code_blocks" }

func (ruleCodeBlocks) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	if !config.Current.Code.DetectBlocks || len(config.Current.Code.DetectPatterns) == 0 {
		return nil
	}

	for _, para := range paragraphs {
		// Check if paragraph contains code patterns
		codePatternCount := 0
		for _, pattern := range config.Current.Code.DetectPatterns {
			if strings.Contains(para.Text, pattern) {
				codePatternCount++
			}
		}

		// If multiple code patterns found, likely a code block
		if codePatternCount >= 3 {
			// Check if in forbidden section
			for _, forbiddenSec := range config.Current.Code.ForbiddenSections {
				if para.SectionID == forbiddenSec {
					sectionLabel := getSectionLabel(sections, para.SectionID)

					comments = append(comments, Comment{
						ID:           "R-CODE-FORBIDDEN-" + para.ID,
						ParagraphID:  para.ID,
						Category:     "content",
						Severity:     "major",
						SectionLabel: sectionLabel,
						Title:        "Kodo fragmentas netinkamoje vietoje",
						Message:      "Skyriuje \"" + sectionLabel + "\" rastas kodo fragmentas. Kodo listingai turetu buti prieduose, o ne pagrindiniame tekste.",
					})
					break
				}
			}
		}
	}

	return comments
}

// ruleMethodologyRequired checks if methodology is mentioned in analysis/implementation
type ruleMethodologyRequired struct{}

func (ruleMethodologyRequired) ID() string { return "methodology_required" }

func (ruleMethodologyRequired) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	methodologyRule := config.Current.ContentRules.MethodologyRequired
	if len(methodologyRule.Sections) == 0 || len(methodologyRule.MustContainAny) == 0 {
		return nil
	}

	for _, sectionID := range methodologyRule.Sections {
		sectionParas := filterParagraphsBySection(paragraphs, sectionID)
		if len(sectionParas) == 0 {
			continue
		}

		// Check if any methodology keyword is present
		foundMethodology := false
		fullText := strings.ToLower(strings.Join(extractTexts(sectionParas), " "))

		for _, keyword := range methodologyRule.MustContainAny {
			if strings.Contains(fullText, strings.ToLower(keyword)) {
				foundMethodology = true
				break
			}
		}

		if !foundMethodology {
			firstPara := sectionParas[0]
			sectionLabel := getSectionLabel(sections, sectionID)

			comments = append(comments, Comment{
				ID:           "R-METHODOLOGY-" + sectionID,
				ParagraphID:  firstPara.ID,
				Category:     "content",
				Severity:     "major",
				SectionLabel: sectionLabel,
				Title:        "Truksta metodologijos aprasymo",
				Message:      "Skyriuje \"" + sectionLabel + "\" nenurodyta naudota metodika ar tyrimo metodai.",
			})
		}
	}

	return comments
}

// ruleConclusionsQuality checks if conclusions are properly written
type ruleConclusionsQuality struct{}

func (ruleConclusionsQuality) ID() string { return "conclusions_quality" }

func (ruleConclusionsQuality) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	conclusionsRule := config.Current.ContentRules.ConclusionsQuality
	if conclusionsRule.SectionID == "" {
		return nil
	}

	sectionParas := filterParagraphsBySection(paragraphs, conclusionsRule.SectionID)
	if len(sectionParas) == 0 {
		return nil
	}

	fullText := strings.ToLower(strings.Join(extractTexts(sectionParas), " "))

	// Check for required keywords
	if len(conclusionsRule.MustContainAny) > 0 {
		foundRequired := false
		for _, keyword := range conclusionsRule.MustContainAny {
			if strings.Contains(fullText, strings.ToLower(keyword)) {
				foundRequired = true
				break
			}
		}

		if !foundRequired {
			firstPara := sectionParas[0]
			sectionLabel := getSectionLabel(sections, conclusionsRule.SectionID)

			comments = append(comments, Comment{
				ID:           "R-CONCLUSIONS-WEAK",
				ParagraphID:  firstPara.ID,
				Category:     "content",
				Severity:     "major",
				SectionLabel: sectionLabel,
				Title:        "Isvados neisamios",
				Message:      "Isvadose turetu buti apibendrinti rezultatai ir pateikti pasiulymai, o ne tik pakartoti uzdaviniai.",
			})
		}
	}

	// Check for forbidden keywords (e.g., just listing tasks instead of conclusions)
	if len(conclusionsRule.ForbiddenAny) > 0 {
		for _, forbidden := range conclusionsRule.ForbiddenAny {
			if strings.Contains(fullText, strings.ToLower(forbidden)) {
				firstPara := sectionParas[0]
				sectionLabel := getSectionLabel(sections, conclusionsRule.SectionID)

				comments = append(comments, Comment{
					ID:           "R-CONCLUSIONS-COPYPASTE",
					ParagraphID:  firstPara.ID,
					Category:     "content",
					Severity:     "critical",
					SectionLabel: sectionLabel,
					Title:        "Isvados netinkamos",
					Message:      "Isvadose rastas uzdaviniu sarasas vietoj isvadu. Isvados turi apibendrinti darbo rezultatus, o ne kartoti ivadineje dalies uzdaviniu.",
				})
				break
			}
		}
	}

	return comments
}

// ruleLongParagraphs checks for excessively long paragraphs
type ruleLongParagraphs struct{}

func (ruleLongParagraphs) ID() string { return "long_paragraphs" }

func (ruleLongParagraphs) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	maxLen := config.Current.Language.MaxParagraphLengthChars
	if maxLen <= 0 {
		return nil
	}

	for _, para := range paragraphs {
		if len(para.Text) > maxLen {
			sectionLabel := getSectionLabel(sections, para.SectionID)

			comments = append(comments, Comment{
				ID:           "R-LONG-PARA-" + para.ID,
				ParagraphID:  para.ID,
				Category:     "language",
				Severity:     "minor",
				SectionLabel: sectionLabel,
				Title:        "Per ilga pastraipa",
				Message:      "Pastraipa virsija rekomenduojama ilgi. Rekomenduojama suskaidyti i kelias trumpesnes pastraipas skaitomumo gerinimui.",
			})
		}
	}

	return comments
}

// Helper functions

func filterParagraphsBySection(paragraphs []Paragraph, sectionID string) []Paragraph {
	var result []Paragraph
	for _, p := range paragraphs {
		if p.SectionID == sectionID {
			result = append(result, p)
		}
	}
	return result
}

func extractTexts(paragraphs []Paragraph) []string {
	texts := make([]string, len(paragraphs))
	for i, p := range paragraphs {
		texts[i] = p.Text
	}
	return texts
}

func getSectionLabel(sections []Section, sectionID string) string {
	for _, s := range sections {
		if s.ID == sectionID {
			return s.Label
		}
	}
	return sectionID
}
