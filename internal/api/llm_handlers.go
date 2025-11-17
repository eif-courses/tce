package api

import (
	"encoding/json"
	"net/http"

	"github.com/eif-courses/tce/internal/llm"
	"github.com/eif-courses/tce/internal/util"
	"go.uber.org/zap"
)

// ==== DTOs for LLM rewrite ====

type LlmSuggestRewriteRequest struct {
	ParagraphID    string `json:"paragraphId"`
	ParagraphText  string `json:"paragraphText"`
	Lang           string `json:"lang"`           // "lt" | "en"
	Program        string `json:"program"`        // "pi" | "se" (or others later)
	CommentTitle   string `json:"commentTitle"`   // rule title shown to student
	CommentMessage string `json:"commentMessage"` // rule explanation shown to student
}

type LlmSuggestRewriteResponse struct {
	ParagraphID string `json:"paragraphId"`
	Suggestion  string `json:"suggestion"`
}

// POST /api/tce/suggest-rewrite
func HandleLLMSuggestRewrite(w http.ResponseWriter, r *http.Request) {
	var req LlmSuggestRewriteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpError(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.ParagraphText == "" {
		httpError(w, "paragraphText is required", http.StatusBadRequest)
		return
	}
	if req.Lang != "en" {
		req.Lang = "lt"
	}

	ctx := r.Context()
	suggestion, err := llm.SuggestRewrite(ctx, llm.SuggestParams{
		Lang:           req.Lang,
		Program:        req.Program,
		ParagraphText:  req.ParagraphText,
		CommentTitle:   req.CommentTitle,
		CommentMessage: req.CommentMessage,
	})
	if err != nil {
		util.Log.Error("LLM suggest rewrite failed",
			zap.Error(err),
			zap.String("lang", req.Lang),
			zap.String("program", req.Program),
		)
		httpError(w, "LLM suggestion failed", http.StatusInternalServerError)
		return
	}

	resp := LlmSuggestRewriteResponse{
		ParagraphID: req.ParagraphID,
		Suggestion:  suggestion,
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	_ = json.NewEncoder(w).Encode(resp)
}
