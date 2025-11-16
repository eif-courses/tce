package analyzer

import (
	"strings"

	"github.com/eif-courses/tce/internal/config"
)

//
// RULE 7: Per daug pirmojo asmens ("aš", "mano", "mes", ...)
//

type ruleFirstPersonStyle struct{}

func (ruleFirstPersonStyle) ID() string { return "first_person_style" }

func (ruleFirstPersonStyle) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	words := config.Current.Language.FirstPerson
	if len(words) == 0 {
		// fallback defaults if YAML missing
		words = []string{" aš ", " mano ", " mes ", " mūsų "}
	}

	for _, p := range paragraphs {
		text := strings.ToLower(" " + p.Text + " ") // pad for word boundaries
		for _, w := range words {
			if strings.Contains(text, strings.ToLower(w)) {
				comments = append(comments, Comment{
					ID:           "R-LANG-FIRST-PERSON-" + p.ID,
					ParagraphID:  p.ID,
					Category:     "language",
					Severity:     "minor",
					SectionLabel: "Kalba / stilius",
					Title:        "Pirmojo asmens vartojimas",
					Message:      "Šioje pastraipoje vartojami pirmojo asmens įvardžiai. Rekomenduojama neutralesnė, akademinė formuluotė.",
				})
				break
			}
		}
	}

	return comments
}

//
// RULE 8: Šnekamoji / neformali leksika
//

type ruleInformalStyle struct{}

func (ruleInformalStyle) ID() string { return "informal_style" }

func (ruleInformalStyle) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var comments []Comment

	words := config.Current.Language.Informal
	if len(words) == 0 {
		words = []string{"šiaip", "fainas", "super", "tipo"}
	}

	for _, p := range paragraphs {
		text := strings.ToLower(p.Text)
		for _, w := range words {
			if strings.Contains(text, strings.ToLower(w)) {
				comments = append(comments, Comment{
					ID:           "R-LANG-INFORMAL-" + p.ID,
					ParagraphID:  p.ID,
					Category:     "language",
					Severity:     "minor",
					SectionLabel: "Kalba / stilius",
					Title:        "Per daug šnekamosios kalbos",
					Message:      "Aptikti šnekamosios kalbos arba per daug neformalūs žodžiai. Rekomenduojama formalesnė stilistika.",
				})
				break
			}
		}
	}

	return comments
}
