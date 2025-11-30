package analyzer

import "github.com/eif-courses/tce/internal/docparser"

type Section = docparser.Section
type Paragraph = docparser.Paragraph
type ContentBlock = docparser.ContentBlock

type Comment struct {
	ID           string
	ParagraphID  string // Can also reference ContentBlock ID
	Category     string // structure/content/language/formatting
	Severity     string // major/minor
	SectionLabel string
	Title        string
	Message      string
}

type Result struct {
	Sections      []Section
	Paragraphs    []Paragraph
	ContentBlocks []ContentBlock // Rich content blocks
	Comments      []Comment
}
