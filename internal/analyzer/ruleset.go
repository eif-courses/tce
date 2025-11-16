package analyzer

import (
	"strings"

	"github.com/eif-courses/tce/internal/config"
)

// Rule represents a single analysis/check that can produce comments.
type Rule interface {
	ID() string
	Apply(sections []Section, paragraphs []Paragraph) []Comment
}

// runRules executes all registered rules and also updates section issue counts.
func runRules(sections []Section, paragraphs []Paragraph) ([]Comment, []Section) {
	sectionIndex := indexSections(sections)

	rules := []Rule{
		ruleIntroHasGoal{},
		ruleIntroHasTasks{},
		ruleRequiredSections{},
		ruleSectionOrder{},
		ruleSectionMinLength{},
		ruleFunctionalRequirements{},

		// language / style
		ruleFirstPersonStyle{},
		ruleInformalStyle{},
	}

	var comments []Comment
	for _, rule := range rules {
		cs := rule.Apply(sections, paragraphs)
		comments = append(comments, cs...)
	}

	// Count issues per section based on comments
	issuesBySection := map[string]int{}
	paraByID := map[string]Paragraph{}
	for _, p := range paragraphs {
		paraByID[p.ID] = p
	}
	for _, c := range comments {
		if p, ok := paraByID[c.ParagraphID]; ok {
			issuesBySection[p.SectionID]++
		}
	}

	for i := range sections {
		if count, ok := issuesBySection[sections[i].ID]; ok {
			sections[i].Issues = count
		}
	}

	// Ensure "unknown" section exists if there are paragraphs with that ID
	if _, ok := sectionIndex["unknown"]; !ok {
		for _, p := range paragraphs {
			if p.SectionID == "unknown" {
				sections = append(sections, Section{
					ID:          "unknown",
					Label:       "Nepriskirta",
					Requirement: "Nenurodyta",
					Issues:      issuesBySection["unknown"],
				})
				break
			}
		}
	}

	return comments, sections
}

func indexSections(sections []Section) map[string]Section {
	m := make(map[string]Section, len(sections))
	for _, s := range sections {
		m[s.ID] = s
	}
	return m
}

//
// ======= RULE 1: Įvadas must contain a clear goal (“tikslas”) =======
//

type ruleIntroHasGoal struct{}

func (ruleIntroHasGoal) ID() string { return "intro_has_goal" }

func (ruleIntroHasGoal) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var (
		introParas []Paragraph
		secLabel   = "Įvadas"
	)

	for _, s := range sections {
		if s.ID == "intro" && s.Label != "" {
			secLabel = s.Label
			break
		}
	}

	for _, p := range paragraphs {
		if p.SectionID == "intro" {
			introParas = append(introParas, p)
		}
	}

	if len(introParas) == 0 {
		// No intro at all – missing section handled by another rule.
		return nil
	}

	for _, p := range introParas {
		if strings.Contains(strings.ToLower(p.Text), "tikslas") {
			return nil // goal found
		}
	}

	first := introParas[0]

	return []Comment{
		{
			ID:           "R-INTRO-GOAL-1",
			ParagraphID:  first.ID,
			Category:     "structure",
			Severity:     "major",
			SectionLabel: secLabel,
			Title:        "Darbo tikslo trūkumas",
			Message:      "Įvade nerastas aiškus darbo tikslas. Rekomenduojama pridėti sakinį su formuluote „Darbo tikslas – …“.",
		},
	}
}

//
// ======= RULE 2: Įvadas must contain tasks (“uždaviniai”) =======
//

type ruleIntroHasTasks struct{}

func (ruleIntroHasTasks) ID() string { return "intro_has_tasks" }

func (ruleIntroHasTasks) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var (
		introParas []Paragraph
		secLabel   = "Įvadas"
	)

	for _, s := range sections {
		if s.ID == "intro" && s.Label != "" {
			secLabel = s.Label
			break
		}
	}

	for _, p := range paragraphs {
		if p.SectionID == "intro" {
			introParas = append(introParas, p)
		}
	}

	if len(introParas) == 0 {
		return nil
	}

	for _, p := range introParas {
		l := strings.ToLower(p.Text)
		if strings.Contains(l, "uždaviniai") || strings.Contains(l, "uždavinys") {
			return nil // tasks found
		}
	}

	// Usually tasks are near the end of intro
	last := introParas[len(introParas)-1]

	return []Comment{
		{
			ID:           "R-INTRO-TASKS-1",
			ParagraphID:  last.ID,
			Category:     "structure",
			Severity:     "minor",
			SectionLabel: secLabel,
			Title:        "Uždavinių pateikimas",
			Message:      "Įvade nerastas aiškus uždavinių sąrašas. Rekomenduojama pateikti numeruotą sąrašą su pagrindiniais darbo uždaviniais.",
		},
	}
}

//
// ======= RULE 3: Required sections must exist =======
//

type ruleRequiredSections struct{}

func (ruleRequiredSections) ID() string { return "required_sections" }

func (ruleRequiredSections) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	sectionMap := indexSections(sections)
	var comments []Comment

	// Use rules from config.Current.Sections which came from YAML
	for _, sr := range config.Current.Sections {
		if !sr.Required {
			continue
		}

		if _, ok := sectionMap[sr.ID]; !ok {
			comments = append(comments, Comment{
				ID:           "R-SEC-MISSING-" + sr.ID,
				ParagraphID:  "",
				Category:     "structure",
				Severity:     "major",
				SectionLabel: sr.Label,
				Title:        "Trūksta skyriaus: " + sr.Label,
				Message:      "Pagal metodinius reikalavimus darbe turi būti skyrius „" + sr.Label + "“. Šio skyriaus nerasta.",
			})
		}
	}

	return comments
}

//
// ======= RULE 4: Section order (Įvadas -> Užduotis -> FR -> Analizė) =======
//

type ruleSectionOrder struct{}

func (ruleSectionOrder) ID() string { return "section_order" }

func (ruleSectionOrder) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	order := []string{"intro", "task", "fr", "analysis"}

	// first paragraph index for each section
	firstIndex := map[string]int{}
	for i, p := range paragraphs {
		if _, seen := firstIndex[p.SectionID]; !seen {
			firstIndex[p.SectionID] = i
		}
	}

	var comments []Comment

	for i := 0; i < len(order)-1; i++ {
		a := order[i]
		b := order[i+1]

		posA, okA := firstIndex[a]
		posB, okB := firstIndex[b]

		if !okA || !okB {
			continue
		}

		if posA > posB {
			para := paragraphs[posB]
			comments = append(comments, Comment{
				ID:           "R-SEC-ORDER-" + a + "-" + b,
				ParagraphID:  para.ID,
				Category:     "structure",
				Severity:     "minor",
				SectionLabel: "Skyriai",
				Title:        "Netinkama skyrių seka",
				Message:      "Skyrius, susijęs su „" + b + "“, atsiranda darbe anksčiau nei „" + a + "“. Rekomenduojama laikytis metodinėje medžiagoje nurodytos sekos.",
			})
		}
	}

	return comments
}

//
// ======= RULE 5: Section minimum length (approximate by paragraph count) =======
//

type ruleSectionMinLength struct{}

func (ruleSectionMinLength) ID() string { return "section_min_length" }

func (ruleSectionMinLength) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	// --- Build minParas from config.Current.Sections ---
	// Simple heuristic: 1 page ≈ 3 paragraphs.
	minParas := map[string]int{}
	for _, sr := range config.Current.Sections {
		if sr.MinPages <= 0 {
			continue
		}
		minParas[sr.ID] = sr.MinPages * 3
	}

	// Count paragraphs per section in this document
	counts := map[string]int{}
	for _, p := range paragraphs {
		counts[p.SectionID]++
	}

	sectionMap := indexSections(sections)
	var comments []Comment

	for id, min := range minParas {
		actual := counts[id]
		if actual == 0 {
			// missing section: handled by ruleRequiredSections
			continue
		}
		if actual < min {
			s := sectionMap[id]

			// attach to first paragraph of that section
			var firstParaID string
			for _, p := range paragraphs {
				if p.SectionID == id {
					firstParaID = p.ID
					break
				}
			}

			comments = append(comments, Comment{
				ID:           "R-SEC-LEN-" + id,
				ParagraphID:  firstParaID,
				Category:     "structure",
				Severity:     "minor",
				SectionLabel: s.Label,
				Title:        "Per trumpas skyrius",
				Message:      "Skyriuje „" + s.Label + "“ yra mažai pastraipų. Pagal metodinius reikalavimus rekomenduojama išsamiau išdėstyti turinį.",
			})
		}
	}

	return comments
}

//
// ======= RULE 6: Funkciniai reikalavimai turi turėti F1, F2, ... =======
//

type ruleFunctionalRequirements struct{}

func (ruleFunctionalRequirements) ID() string { return "fr_must_have_f_items" }

func (ruleFunctionalRequirements) Apply(sections []Section, paragraphs []Paragraph) []Comment {
	var frParas []Paragraph
	for _, p := range paragraphs {
		if p.SectionID == "fr" {
			frParas = append(frParas, p)
		}
	}

	if len(frParas) == 0 {
		return nil // missing FR handled elsewhere
	}

	hasF := false
	for _, p := range frParas {
		txt := strings.ToUpper(p.Text)
		if strings.Contains(txt, "F1") || strings.Contains(txt, "F2") || strings.Contains(txt, "F3") {
			hasF = true
			break
		}
	}

	if hasF {
		return nil
	}

	first := frParas[0]

	return []Comment{
		{
			ID:           "R-FR-NO-F-ITEMS",
			ParagraphID:  first.ID,
			Category:     "content",
			Severity:     "major",
			SectionLabel: "Funkciniai reikalavimai",
			Title:        "Funkciniai reikalavimai neišskirti",
			Message:      "„Funkciniai reikalavimai“ skyriuje nerasta aiškių reikalavimų formuluočių (pvz., F1, F2, ...). Rekomenduojama pateikti struktūruotą reikalavimų sąrašą.",
		},
	}
}
