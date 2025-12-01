package docparser

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/unidoc/unioffice/document"
)

// ParseDOCXWithUnioffice reads a DOCX file using unioffice library and extracts
// paragraphs with formatting information, images, tables, etc.
func ParseDOCXWithUnioffice(r io.Reader) (*RawDoc, error) {
	// 1) Save to temp file (unioffice needs a file path)
	tmp, err := os.CreateTemp("", "tce-unioffice-*.docx")
	if err != nil {
		return nil, fmt.Errorf("create temp file: %w", err)
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, r); err != nil {
		tmp.Close()
		return nil, fmt.Errorf("write temp file: %w", err)
	}
	tmp.Close()

	// 2) Open document with unioffice
	doc, err := document.Open(tmp.Name())
	if err != nil {
		return nil, fmt.Errorf("open docx with unioffice: %w", err)
	}
	defer doc.Close()

	var contentBlocks []ContentBlock
	var paragraphs []string
	blockID := 0

	// 3) Extract paragraphs
	for _, para := range doc.Paragraphs() {
		text := extractParagraphText(para)
		if strings.TrimSpace(text) == "" {
			continue
		}

		paragraphs = append(paragraphs, text)

		// Create content block with formatting info
		block := ContentBlock{
			ID:      fmt.Sprintf("block-%d", blockID),
			Type:    "paragraph",
			Content: TextContent{Text: text},
		}

		// Check if this paragraph might be a heading based on style
		if isHeading(para) {
			level := getHeadingLevel(para)
			block.Type = "heading"
			block.Content = HeadingContent{
				Text:  text,
				Level: level,
			}
		}

		contentBlocks = append(contentBlocks, block)
		blockID++
	}

	// 4) Extract tables
	for _, tbl := range doc.Tables() {
		tableBlock := extractTable(tbl, &blockID)
		if tableBlock != nil {
			contentBlocks = append(contentBlocks, *tableBlock)
		}
	}

	// 5) Extract images (from inline shapes and drawings)
	// Note: unioffice image extraction is complex, we'll keep the pandoc approach for now
	// or implement a simpler version

	return &RawDoc{
		Paragraphs:    paragraphs,
		ContentBlocks: contentBlocks,
	}, nil
}

// extractParagraphText concatenates all text runs in a paragraph
func extractParagraphText(para document.Paragraph) string {
	var parts []string
	for _, run := range para.Runs() {
		parts = append(parts, run.Text())
	}
	return strings.Join(parts, "")
}

// isHeading checks if a paragraph is styled as a heading
func isHeading(para document.Paragraph) bool {
	// Check paragraph style from properties
	props := para.Properties()
	if props.X() != nil && props.X().PStyle != nil {
		styleName := strings.ToLower(props.X().PStyle.ValAttr)
		return strings.Contains(styleName, "heading") ||
		       strings.Contains(styleName, "title") ||
		       strings.Contains(styleName, "antraštė")
	}

	return false
}

// getHeadingLevel extracts the heading level (1, 2, 3, etc.)
func getHeadingLevel(para document.Paragraph) int {
	props := para.Properties()

	// Check paragraph style
	if props.X() != nil && props.X().PStyle != nil {
		styleName := strings.ToLower(props.X().PStyle.ValAttr)

		// Try to extract level from style name (e.g., "Heading 1", "Heading1", "antraštė 1")
		for level := 1; level <= 6; level++ {
			if strings.Contains(styleName, fmt.Sprintf("heading%d", level)) ||
			   strings.Contains(styleName, fmt.Sprintf("heading %d", level)) ||
			   strings.Contains(styleName, fmt.Sprintf("antraštė%d", level)) ||
			   strings.Contains(styleName, fmt.Sprintf("antraštė %d", level)) {
				return level
			}
		}

		// Default heading without number
		if strings.Contains(styleName, "heading") || strings.Contains(styleName, "antraštė") {
			return 1
		}
	}

	return 1
}

// extractTable converts a unioffice table to ContentBlock
func extractTable(tbl document.Table, blockID *int) *ContentBlock {
	rows := tbl.Rows()
	if len(rows) == 0 {
		return nil
	}

	var headers []string
	var tableRows [][]string

	// Extract headers from first row
	firstRow := rows[0]
	for _, cell := range firstRow.Cells() {
		cellText := extractCellText(cell)
		headers = append(headers, cellText)
	}

	// Extract data rows
	for i := 1; i < len(rows); i++ {
		row := rows[i]
		var rowData []string
		for _, cell := range row.Cells() {
			cellText := extractCellText(cell)
			rowData = append(rowData, cellText)
		}
		if len(rowData) > 0 {
			tableRows = append(tableRows, rowData)
		}
	}

	block := &ContentBlock{
		ID:   fmt.Sprintf("block-%d", *blockID),
		Type: "table",
		Content: TableContent{
			Headers: headers,
			Rows:    tableRows,
		},
	}
	*blockID++
	return block
}

// extractCellText extracts text from a table cell
func extractCellText(cell document.Cell) string {
	var parts []string
	for _, para := range cell.Paragraphs() {
		text := extractParagraphText(para)
		if strings.TrimSpace(text) != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, " ")
}
