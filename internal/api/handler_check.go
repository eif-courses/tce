package api

import (
	"encoding/json"
	"net/http"

	"github.com/eif-courses/tce/internal/analyzer"
	"github.com/eif-courses/tce/internal/docx"
)

// DTOs returned to frontend (match your Nuxt interfaces)

type sectionDTO struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Requirement string `json:"requirement"`
	Issues      int    `json:"issues"`
}

type paragraphDTO struct {
	ID        string `json:"id"`
	SectionID string `json:"sectionId"`
	Text      string `json:"text"`
}

type commentDTO struct {
	ID           string `json:"id"`
	ParagraphID  string `json:"paragraphId"`
	Category     string `json:"category"`
	Severity     string `json:"severity"`
	SectionLabel string `json:"sectionLabel"`
	Title        string `json:"title"`
	Message      string `json:"message"`
	Suggestion   string `json:"suggestion,omitempty"`
}

type checkResponse struct {
	Sections   []sectionDTO   `json:"sections"`
	Paragraphs []paragraphDTO `json:"paragraphs"`
	Comments   []commentDTO   `json:"comments"`
}

// CheckHandler handles POST /api/tce/check
func CheckHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		http.Error(w, "cannot parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file field 'file' is required", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// ✨ NEW: extract plain text from DOCX using pandoc
	plain, err := docx.ExtractTextFromDocx(file)
	if err != nil {
		http.Error(w, "failed to extract text: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Feed extracted text into analyzer
	out := analyzer.Analyze(analyzer.Input{
		PlainText: plain,
	})

	// map to DTOs for Nuxt
	resp := checkResponse{
		Sections:   make([]sectionDTO, 0, len(out.Sections)),
		Paragraphs: make([]paragraphDTO, 0, len(out.Paragraphs)),
		Comments:   make([]commentDTO, 0, len(out.Issues)),
	}

	for _, s := range out.Sections {
		resp.Sections = append(resp.Sections, sectionDTO{
			ID:          s.ID,
			Label:       s.Label,
			Requirement: s.Requirement,
			Issues:      s.Issues,
		})
	}

	for _, p := range out.Paragraphs {
		resp.Paragraphs = append(resp.Paragraphs, paragraphDTO{
			ID:        p.ID,
			SectionID: p.SectionID,
			Text:      p.Text,
		})
	}

	for _, issue := range out.Issues {
		resp.Comments = append(resp.Comments, commentDTO{
			ID:           issue.ID,
			ParagraphID:  issue.ParagraphID,
			Category:     issue.Category,
			Severity:     issue.Severity,
			SectionLabel: issue.SectionLabel,
			Title:        issue.Title,
			Message:      issue.Message,
			Suggestion:   issue.Suggestion,
		})
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
		return
	}
}
