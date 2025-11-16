package api

// === Incoming DTO ===

type CheckResponse struct {
	Sections   []SectionDTO   `json:"sections"`
	Paragraphs []ParagraphDTO `json:"paragraphs"`
	Comments   []CommentDTO   `json:"comments"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

// === Outgoing DTOs ===

type SectionDTO struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Requirement string `json:"requirement"`
	Issues      int    `json:"issues"`
}

type ParagraphDTO struct {
	ID        string `json:"id"`
	SectionID string `json:"sectionId"`
	Text      string `json:"text"`
}

type CommentDTO struct {
	ID           string `json:"id"`
	ParagraphID  string `json:"paragraphId"`
	Category     string `json:"category"`
	Severity     string `json:"severity"`
	SectionLabel string `json:"sectionLabel"`
	Title        string `json:"title"`
	Message      string `json:"message"`
}
