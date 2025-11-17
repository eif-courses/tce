package api

import "github.com/eif-courses/tce/internal/docparser"

// === Incoming DTO ===

type CheckResponse struct {
	Sections      []SectionDTO      `json:"sections"`
	Paragraphs    []ParagraphDTO    `json:"paragraphs"`
	ContentBlocks []ContentBlockDTO `json:"contentBlocks"` // New rich content
	Comments      []CommentDTO      `json:"comments"`
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

type ContentBlockDTO struct {
	ID        string      `json:"id"`
	SectionID string      `json:"sectionId"`
	Type      string      `json:"type"` // "paragraph", "image", "table", "heading", "list"
	Content   interface{} `json:"content"`
}

type TextContentDTO struct {
	Text string `json:"text"`
}

type ImageContentDTO struct {
	Alt     string `json:"alt"`
	Base64  string `json:"base64,omitempty"`
	Caption string `json:"caption,omitempty"`
	Width   int    `json:"width,omitempty"`
	Height  int    `json:"height,omitempty"`
}

type TableContentDTO struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
	Caption string     `json:"caption,omitempty"`
}

type HeadingContentDTO struct {
	Text  string `json:"text"`
	Level int    `json:"level"`
}

type ListContentDTO struct {
	Items   []string `json:"items"`
	Ordered bool     `json:"ordered"`
}

type CommentDTO struct {
	ID           string `json:"id"`
	ParagraphID  string `json:"paragraphId"` // Can also reference ContentBlock ID
	Category     string `json:"category"`
	Severity     string `json:"severity"`
	SectionLabel string `json:"sectionLabel"`
	Title        string `json:"title"`
	Message      string `json:"message"`
}

// Helper to convert docparser ContentBlock to DTO
func toContentBlockDTO(block docparser.ContentBlock) ContentBlockDTO {
	return ContentBlockDTO{
		ID:        block.ID,
		SectionID: block.SectionID,
		Type:      block.Type,
		Content:   block.Content,
	}
}
