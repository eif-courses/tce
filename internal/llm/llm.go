// internal/llm/llm.go
package llm

import (
	"context"
	"fmt"
	"os"
	"sync"

	openai "github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
)

var (
	client     openai.Client
	clientOnce sync.Once
)

// lazy-init global client (thread-safe)
func getClient() *openai.Client {
	clientOnce.Do(func() {
		client = openai.NewClient(
			option.WithAPIKey(os.Getenv("OPENAI_API_KEY")), // will also read env automatically
		)
	})
	return &client
}

// Parameters for the rewrite suggestion – matches what front-end sends.
type SuggestParams struct {
	Lang           string
	Program        string
	ParagraphText  string
	CommentTitle   string
	CommentMessage string
}

// SuggestRewrite calls GPT-4o-mini and returns an improved paragraph.
func SuggestRewrite(ctx context.Context, p SuggestParams) (string, error) {
	if p.Lang == "" {
		p.Lang = "lt"
	}

	systemMsg := "You are an academic writing assistant. Rewrite the text in a formal, clear academic style. Return ONLY the improved paragraph."
	if p.Lang == "lt" {
		systemMsg = "Tu esi akademinio rašymo asistentas. Pataisyk tekstą į taisyklingą, formalų, akademinį stilių. Grąžink TIK pataisytą pastraipą lietuvių kalba."
	}

	userMsg := fmt.Sprintf(
		"Study program: %s\nLanguage: %s\nIssue: %s – %s\nOriginal paragraph:\n%s",
		p.Program, p.Lang, p.CommentTitle, p.CommentMessage, p.ParagraphText,
	)

	resp, err := getClient().Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Model: openai.ChatModelGPT4oMini,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(systemMsg),
			openai.UserMessage(userMsg),
		},
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices from OpenAI")
	}

	// In v3, Content is a plain string (see README example)
	return resp.Choices[0].Message.Content, nil
}
