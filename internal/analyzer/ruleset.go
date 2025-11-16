package analyzer

import (
	"strings"
)

// Rule represents one check that can add comments.
type Rule interface {
	ID() string
	Apply(sections []Section, paragraphs []Paragraph) []Comment
}

// AnalyzeWithRules applies all default rules.
// (You already call Analyze(raw *docparser.RawDoc) in analyzer.go – we’ll
// wire rules into that function.)
func runRules(sections []Section, paragraphs []Paragraph) ([]Comment, []Section) {
	sectionIndex := indexSections(sections)
	rules := []Rule{
		ruleIntroHasGoal{},
		ruleIntroHasTasks{},
	}

	var comments []Comment
	for _, rule := range rules {
		cs := rule.Apply(sections, paragraphs)
		comments = append(comments, cs...)
	}

	// Count issues per section
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

	// ensure “unknown” section exists if any paragraph has that
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

// ======= RULE 1: Įvadas must have “tikslas” (goal) =======

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
		}
	}

	for _, p := range paragraphs {
		if p.SectionID == "intro" {
			introParas = append(introParas, p)
		}
	}

	if len(introParas) == 0 {
		// No intro paragraphs at all → no comment, that’s another rule.
		return nil
	}

	for _, p := range introParas {
		if strings.Contains(strings.ToLower(p.Text), "tikslas") {
			return nil // OK, goal present
		}
	}

	// No goal found
	// Attach comment to first intro paragraph
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

// ======= RULE 2: Įvadas must describe tasks (“uždaviniai”) =======

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
			return nil // tasks present
		}
	}

	first := introParas[len(introParas)-1] // usually tasks are at the end

	return []Comment{
		{
			ID:           "R-INTRO-TASKS-1",
			ParagraphID:  first.ID,
			Category:     "structure",
			Severity:     "minor",
			SectionLabel: secLabel,
			Title:        "Uždavinių pateikimas",
			Message:      "Įvade nerastas aiškus uždavinių sąrašas. Rekomenduojama pateikti numeruotą sąrašą su pagrindiniais darbo uždaviniais.",
		},
	}
}
