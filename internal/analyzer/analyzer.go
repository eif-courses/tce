package analyzer

import (
	"fmt"
	"strings"
)

// Core types used by the engine

type Section struct {
	ID          string
	Label       string
	Requirement string
	Issues      int
}

type Paragraph struct {
	ID        string
	SectionID string
	Text      string
}

type Issue struct {
	ID           string
	ParagraphID  string
	Category     string // structure, content, language, formatting
	Severity     string // major, minor
	SectionLabel string
	Title        string
	Message      string
	Suggestion   string
}

// Input to the analyzer (later: add language, study programme, etc.)
type Input struct {
	PlainText string
}

// Output from the analyzer
type Output struct {
	Sections   []Section
	Paragraphs []Paragraph
	Issues     []Issue
}

// Analyze is the main entry point.
// It now uses PlainText to build paragraphs and sections.
func Analyze(in Input) Output {
	plain := strings.TrimSpace(in.PlainText)
	if plain == "" {
		// Nothing to analyze, return empty but valid payload
		return Output{
			Sections:   []Section{},
			Paragraphs: []Paragraph{},
			Issues:     []Issue{},
		}
	}

	// 1) Split into raw paragraphs
	rawParas := splitParagraphs(plain)

	// 2) Assign to sections based on simple heading detection
	sections, paragraphs := assignSections(rawParas)

	// 3) Run rules
	ai := analysisInput{
		Sections:   sections,
		Paragraphs: paragraphs,
		PlainText:  plain,
	}
	issues := runRules(ai)

	// 4) Count issues per section label
	sectionIssueCount := map[string]int{}
	for _, iss := range issues {
		if iss.SectionLabel != "" {
			sectionIssueCount[iss.SectionLabel]++
		}
	}

	for i, s := range sections {
		if c, ok := sectionIssueCount[s.Label]; ok {
			sections[i].Issues = c
		}
	}

	return Output{
		Sections:   sections,
		Paragraphs: paragraphs,
		Issues:     issues,
	}
}

// internal struct passed to rules
type analysisInput struct {
	Sections   []Section
	Paragraphs []Paragraph
	PlainText  string
}

type ruleFunc func(analysisInput) []Issue

func runRules(in analysisInput) []Issue {
	rules := []ruleFunc{
		ruleIntroHasGoal,
		// TODO: add more like:
		// ruleIntroHasTasks,
		// ruleTerminaiExists,
		// ruleFRMinCount,
	}

	var issues []Issue
	for _, r := range rules {
		issues = append(issues, r(in)...)
	}
	return issues
}

//
// === Basic text processing ===
//

// splitParagraphs: very simple: split on double newlines
func splitParagraphs(plain string) []string {
	// normalize CRLF
	normalized := strings.ReplaceAll(plain, "\r\n", "\n")
	chunks := strings.Split(normalized, "\n\n")

	out := make([]string, 0, len(chunks))
	for _, c := range chunks {
		t := strings.TrimSpace(c)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// assignSections: naive section detection by keywords/headings.
// Later you can improve with regexes, roman numerals, etc.
func assignSections(rawParas []string) ([]Section, []Paragraph) {
	// Predefined sections you care about (order does not matter)
	sections := []Section{
		{ID: "intro", Label: "Įvadas", Requirement: "2–5 psl.", Issues: 0},
		{ID: "task", Label: "Praktikos užduotis ir sistema", Requirement: "3–5 psl.", Issues: 0},
		{ID: "fr", Label: "Funkciniai reikalavimai", Requirement: "≥3 psl.", Issues: 0},
	}

	// ID → label lookup
	sectionByID := map[string]string{
		"intro": "Įvadas",
		"task":  "Praktikos užduotis ir sistema",
		"fr":    "Funkciniai reikalavimai",
	}

	currentSectionID := "intro" // default if nothing matched yet
	var paragraphs []Paragraph

	for i, text := range rawParas {
		upper := strings.ToUpper(text)

		// Detect headings – you can tweak these conditions as needed
		switch {
		case strings.HasPrefix(upper, "ĮVADAS"):
			currentSectionID = "intro"
			continue // skip heading paragraph itself (we only use content)
		case strings.Contains(upper, "PRAKTIKOS UŽDUOTIS"):
			currentSectionID = "task"
			continue
		case strings.Contains(upper, "FUNKCINIAI REIKALAVIMAI"):
			currentSectionID = "fr"
			continue
		}

		para := Paragraph{
			ID:        fmt.Sprintf("p-%d", i),
			SectionID: currentSectionID,
			Text:      text,
		}
		paragraphs = append(paragraphs, para)
	}

	// If you want: drop sections that had no paragraphs at all
	// (or mark them as missing with Issues = -1 later)

	_ = sectionByID // currently unused, but nice if mapping needed later

	return sections, paragraphs
}

//
// === Example rule: Įvadas must contain "Darbo tikslas" phrase ===
//

func ruleIntroHasGoal(in analysisInput) []Issue {
	var introParas []Paragraph
	for _, p := range in.Paragraphs {
		if p.SectionID == "intro" {
			introParas = append(introParas, p)
		}
	}

	if len(introParas) == 0 {
		// no intro at all -> major issue
		return []Issue{
			{
				ID:           "r1-no-intro",
				ParagraphID:  "",
				Category:     "structure",
				Severity:     "major",
				SectionLabel: "Įvadas",
				Title:        "Trūksta Įvado skyriaus",
				Message:      "Dokumente nerastas Įvado skyrius. Pagal metodinius reikalavimus Įvadas turi būti pirmasis turinio skyrius.",
				Suggestion:   "Pridėkite Įvado skyrių, kuriame pateiksite problemos aktualumą, darbo tikslą ir uždavinius.",
			},
		}
	}

	// Search for phrase "darbo tikslas"
	for _, p := range introParas {
		if strings.Contains(strings.ToLower(p.Text), "darbo tikslas") {
			// OK – tikslas existe, no issue
			return nil
		}
	}

	// no "darbo tikslas" found in intro: attach issue to first intro paragraph
	firstParaID := introParas[0].ID

	return []Issue{
		{
			ID:           fmt.Sprintf("r1-missing-goal-%s", firstParaID),
			ParagraphID:  firstParaID,
			Category:     "structure",
			Severity:     "major",
			SectionLabel: "Įvadas",
			Title:        "Trūksta darbo tikslo formuluotės Įvade",
			Message:      "Įvade nerasta frazė su „Darbo tikslas – …“. Pagal metodinius reikalavimus tikslas turi būti aiškiai suformuluotas atskiroje pastraipoje.",
			Suggestion:   "Į Įvado skyrių įtraukite sakinio formuluotę „Darbo tikslas – …“ ir glaustai nurodykite, ką ketinate pasiekti atlikdami darbą.",
		},
	}
}
