// internal/api/check.go
package api

import (
	"encoding/json"
	"net/http"

	"github.com/eif-courses/tce/internal/analyzer"
	"github.com/eif-courses/tce/internal/config"
	"github.com/eif-courses/tce/internal/docparser"
	"github.com/eif-courses/tce/internal/util"
	"go.uber.org/zap"
)

// POST /api/tce/check
func HandleCheck(w http.ResponseWriter, r *http.Request) {
	// Limit upload size (25MB)
	if err := r.ParseMultipartForm(25 << 20); err != nil {
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

	// Language (lt | en)
	lang := r.FormValue("lang")
	if lang != "en" {
		lang = "lt"
	}

	// Study program (pi | se | etc). Default to "pi"
	program := r.FormValue("program")
	if program == "" {
		program = "pi"
	}

	util.Log.Info("TCE: received file",
		zap.String("filename", header.Filename),
		zap.String("lang", lang),
		zap.String("program", program),
	)

	// Load program+language profile from config/rules_<program>_<lang>.yaml
	if err := config.LoadRulesForProfile(program, lang); err != nil {
		util.Log.Warn("Failed to load rules profile, using last/empty profile",
			zap.String("program", program),
			zap.String("lang", lang),
			zap.Error(err),
		)
	}

	// Step 1 — Parse DOCX → raw paragraphs
	raw, err := docparser.ParseDOCX(file)
	if err != nil {
		httpError(w, "failed to parse DOCX: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Step 2 — Run analyzer rules
	result := analyzer.Analyze(raw, lang)

	// Step 3 — Map to DTO and return JSON
	resp := CheckResponse{
		Sections:      toSectionDTO(result.Sections),
		Paragraphs:    toParagraphDTO(result.Paragraphs),
		ContentBlocks: toContentBlocksDTO(result.ContentBlocks),
		Comments:      toCommentDTO(result.Comments),
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

func toContentBlocksDTO(in []analyzer.ContentBlock) []ContentBlockDTO {
	out := make([]ContentBlockDTO, len(in))
	for i, block := range in {
		out[i] = toContentBlockDTO(block)
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
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(ErrorResponse{Error: msg})
}
