package docparser

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// RawDoc is lowest-level text extracted from DOCX.
type RawDoc struct {
	Paragraphs []string
}

// ParseDOCX reads a DOCX file from r and extracts paragraphs using pandoc.
func ParseDOCX(r io.Reader) (*RawDoc, error) {
	// 1) Save to temp file
	tmp, err := os.CreateTemp("", "tce-*.docx")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()

	if _, err := io.Copy(tmp, r); err != nil {
		return nil, fmt.Errorf("write temp file: %w", err)
	}

	// 2) Run pandoc: docx -> plain text
	var out, stderr bytes.Buffer
	cmd := exec.Command("pandoc", "-f", "docx", "-t", "plain", tmp.Name())
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pandoc error: %v (%s)", err, stderr.String())
	}

	text := out.String()

	// 3) Split into paragraphs on empty lines
	blocks := strings.Split(text, "\n\n")
	paras := make([]string, 0, len(blocks))

	for _, b := range blocks {
		t := strings.TrimSpace(b)
		if t != "" {
			paras = append(paras, t)
		}
	}

	return &RawDoc{Paragraphs: paras}, nil
}
