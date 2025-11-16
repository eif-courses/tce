package docparser

// Section represents a logical document section.
type Section struct {
	ID          string
	Label       string
	Requirement string
	Issues      int
}

// Paragraph is a paragraph linked to a section.
type Paragraph struct {
	ID        string
	SectionID string
	Text      string
}
