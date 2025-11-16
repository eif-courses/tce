package analyzer

import "github.com/eif-courses/tce/internal/docparser"

type Section = docparser.Section
type Paragraph = docparser.Paragraph

type Comment struct {
	ID           string
	ParagraphID  string
	Category     string // structure/content/language/formatting
	Severity     string // major/minor
	SectionLabel string
	Title        string
	Message      string
}

type Result struct {
	Sections   []Section
	Paragraphs []Paragraph
	Comments   []Comment
}
