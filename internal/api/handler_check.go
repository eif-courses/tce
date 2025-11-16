package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eif-courses/tce/internal/analyzer"
	"github.com/eif-courses/tce/internal/docparser"
)

// POST /api/tce/check
func HandleCheck(w http.ResponseWriter, r *http.Request) {
	// Limit upload size (25MB)
	err := r.ParseMultipartForm(25 << 20)
	if err != nil {
		httpError(w, "failed to parse multipart form: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Retrieve file
	file, header, err := r.FormFile("file")
	if err != nil {
		httpError(w, "no file uploaded", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Printf("[TCE] Received file: %s\n", header.Filename)

	// Step 1 — Parse DOCX → RawDoc
	doc, err := docparser.ParseDOCX(file)
	if err != nil {
		httpError(w, "failed to parse DOCX: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Step 2 — Run analyzer rules
	result := analyzer.Analyze(doc)

	// Step 3 — Return unified DTO JSON
	resp := CheckResponse{
		Sections:   toSectionDTO(result.Sections),
		Paragraphs: toParagraphDTO(result.Paragraphs),
		Comments:   toCommentDTO(result.Comments),
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}
func toSectionDTO(in []analyzer.Section) []SectionDTO {
	out := make([]SectionDTO, len(in))
	for i, s := range in {
		out[i] = SectionDTO{
			ID:          s.ID,
			Label:       s.Label,
			Requirement: s.Requirement,
			Issues:      s.Issues,
		}
	}
	return out
}

func toParagraphDTO(in []analyzer.Paragraph) []ParagraphDTO {
	out := make([]ParagraphDTO, len(in))
	for i, p := range in {
		out[i] = ParagraphDTO{
			ID:        p.ID,
			SectionID: p.SectionID,
			Text:      p.Text,
		}
	}
	return out
}

func toCommentDTO(in []analyzer.Comment) []CommentDTO {
	out := make([]CommentDTO, len(in))
	for i, c := range in {
		out[i] = CommentDTO{
			ID:           c.ID,
			ParagraphID:  c.ParagraphID,
			Category:     c.Category,
			Severity:     c.Severity,
			SectionLabel: c.SectionLabel,
			Title:        c.Title,
			Message:      c.Message,
		}
	}
	return out
}

// Generic error sender
func httpError(w http.ResponseWriter, msg string, code int) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
