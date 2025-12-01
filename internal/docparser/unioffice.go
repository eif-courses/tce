package docparser

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// RawDoc is lowest-level text extracted from DOCX.
type RawDoc struct {
	Paragraphs    []string
	ContentBlocks []ContentBlock // Rich content blocks (new)
}

// ParseDOCX reads a DOCX file from r and extracts paragraphs.
// It tries unioffice first (native Go), then falls back to pandoc if needed.
func ParseDOCX(r io.Reader) (*RawDoc, error) {
	// Save input to a buffer so we can try multiple parsers
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, r); err != nil {
		return nil, fmt.Errorf("buffer docx: %w", err)
	}

	// Try unioffice parser first (native Go, no external dependencies)
	doc, err := ParseDOCXWithUnioffice(bytes.NewReader(buf.Bytes()))
	if err == nil && doc != nil {
		// Return unioffice result even if empty - it successfully parsed the document
		return doc, nil
	}

	// Only fallback to pandoc if unioffice had an error
	// Check if pandoc is available first
	if _, err := exec.LookPath("pandoc"); err != nil {
		// Pandoc not available, return unioffice result or error
		if doc != nil {
			return doc, nil
		}
		return nil, fmt.Errorf("failed to parse DOCX: unioffice error (%v) and pandoc not available", err)
	}

	// Fallback to pandoc if available
	return parseDOCXWithPandoc(bytes.NewReader(buf.Bytes()))
}

// parseDOCXWithPandoc uses external pandoc command (legacy method)
func parseDOCXWithPandoc(r io.Reader) (*RawDoc, error) {
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

	// 2) Extract rich content using markdown format
	contentBlocks, err := extractRichContent(tmp.Name())
	if err != nil {
		// Fallback to plain text if rich extraction fails
		contentBlocks = []ContentBlock{}
	}

	// 3) Also extract plain text for backwards compatibility
	var out, stderr bytes.Buffer
	cmd := exec.Command("pandoc", "-f", "docx", "-t", "plain", tmp.Name())
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pandoc error: %v (%s)", err, stderr.String())
	}

	text := out.String()

	// 4) Split into paragraphs on empty lines
	blocks := strings.Split(text, "\n\n")
	paras := make([]string, 0, len(blocks))

	for _, b := range blocks {
		t := strings.TrimSpace(b)
		if t != "" {
			paras = append(paras, t)
		}
	}

	return &RawDoc{
		Paragraphs:    paras,
		ContentBlocks: contentBlocks,
	}, nil
}

// extractRichContent uses pandoc to extract structured content including images, tables, etc.
func extractRichContent(docxPath string) ([]ContentBlock, error) {
	// Create temp directory for media files
	mediaDir, err := os.MkdirTemp("", "tce-media-*")
	if err != nil {
		return nil, fmt.Errorf("create media dir: %w", err)
	}
	defer os.RemoveAll(mediaDir)

	// Run pandoc: docx -> markdown with image extraction
	var out, stderr bytes.Buffer
	cmd := exec.Command("pandoc",
		"-f", "docx",
		"-t", "markdown",
		"--extract-media", mediaDir,
		docxPath,
	)
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("pandoc markdown error: %v (%s)", err, stderr.String())
	}

	markdown := out.String()
	return parseMarkdownToBlocks(markdown, mediaDir)
}

// parseMarkdownToBlocks converts markdown to ContentBlocks
func parseMarkdownToBlocks(markdown string, mediaDir string) ([]ContentBlock, error) {
	var blocks []ContentBlock
	blockID := 0

	lines := strings.Split(markdown, "\n")
	var currentParagraph []string
	var inTable bool
	var tableLines []string

	for i := 0; i < len(lines); i++ {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Check for headings (# ## ### etc.)
		if strings.HasPrefix(trimmed, "#") {
			// Save any pending paragraph
			if len(currentParagraph) > 0 {
				blocks = append(blocks, createTextBlock(&blockID, strings.Join(currentParagraph, " ")))
				currentParagraph = nil
			}

			level := 0
			for _, ch := range trimmed {
				if ch == '#' {
					level++
				} else {
					break
				}
			}
			headingText := strings.TrimSpace(trimmed[level:])
			blocks = append(blocks, ContentBlock{
				ID:      fmt.Sprintf("block-%d", blockID),
				Type:    "heading",
				Content: HeadingContent{Text: headingText, Level: level},
			})
			blockID++
			continue
		}

		// Check for images ![alt](path)
		imgRegex := regexp.MustCompile(`!\[(.*?)\]\((.*?)\)`)
		if imgRegex.MatchString(trimmed) {
			// Save any pending paragraph
			if len(currentParagraph) > 0 {
				blocks = append(blocks, createTextBlock(&blockID, strings.Join(currentParagraph, " ")))
				currentParagraph = nil
			}

			matches := imgRegex.FindStringSubmatch(trimmed)
			if len(matches) >= 3 {
				alt := matches[1]
				imgPath := matches[2]

				// Try to read and encode image
				var base64Data string
				fullPath := filepath.Join(mediaDir, imgPath)
				if data, err := os.ReadFile(fullPath); err == nil {
					base64Data = base64.StdEncoding.EncodeToString(data)
				}

				blocks = append(blocks, ContentBlock{
					ID:   fmt.Sprintf("block-%d", blockID),
					Type: "image",
					Content: ImageContent{
						Alt:    alt,
						Base64: base64Data,
					},
				})
				blockID++
			}
			continue
		}

		// Check for table (markdown tables have |)
		if strings.Contains(trimmed, "|") && trimmed != "" {
			// Save any pending paragraph
			if len(currentParagraph) > 0 {
				blocks = append(blocks, createTextBlock(&blockID, strings.Join(currentParagraph, " ")))
				currentParagraph = nil
			}

			if !inTable {
				inTable = true
				tableLines = []string{line}
			} else {
				tableLines = append(tableLines, line)
			}
			continue
		} else if inTable {
			// End of table
			inTable = false
			if len(tableLines) > 0 {
				if tableBlock := parseMarkdownTable(tableLines, &blockID); tableBlock != nil {
					blocks = append(blocks, *tableBlock)
				}
				tableLines = nil
			}
		}

		// Check for lists (- or * for unordered, 1. for ordered)
		if regexp.MustCompile(`^[\-\*]\s+`).MatchString(trimmed) || regexp.MustCompile(`^\d+\.\s+`).MatchString(trimmed) {
			// Save any pending paragraph
			if len(currentParagraph) > 0 {
				blocks = append(blocks, createTextBlock(&blockID, strings.Join(currentParagraph, " ")))
				currentParagraph = nil
			}

			// Collect list items
			var items []string
			ordered := regexp.MustCompile(`^\d+\.\s+`).MatchString(trimmed)

			for i < len(lines) {
				listLine := strings.TrimSpace(lines[i])
				if regexp.MustCompile(`^[\-\*]\s+`).MatchString(listLine) {
					item := regexp.MustCompile(`^[\-\*]\s+`).ReplaceAllString(listLine, "")
					items = append(items, item)
					i++
				} else if regexp.MustCompile(`^\d+\.\s+`).MatchString(listLine) {
					item := regexp.MustCompile(`^\d+\.\s+`).ReplaceAllString(listLine, "")
					items = append(items, item)
					i++
				} else {
					break
				}
			}
			i-- // Back up one since we'll increment in the loop

			if len(items) > 0 {
				blocks = append(blocks, ContentBlock{
					ID:   fmt.Sprintf("block-%d", blockID),
					Type: "list",
					Content: ListContent{
						Items:   items,
						Ordered: ordered,
					},
				})
				blockID++
			}
			continue
		}

		// Regular paragraph
		if trimmed != "" {
			currentParagraph = append(currentParagraph, trimmed)
		} else if len(currentParagraph) > 0 {
			// Empty line - end current paragraph
			blocks = append(blocks, createTextBlock(&blockID, strings.Join(currentParagraph, " ")))
			currentParagraph = nil
		}
	}

	// Save any remaining paragraph
	if len(currentParagraph) > 0 {
		blocks = append(blocks, createTextBlock(&blockID, strings.Join(currentParagraph, " ")))
	}

	// Handle remaining table
	if inTable && len(tableLines) > 0 {
		if tableBlock := parseMarkdownTable(tableLines, &blockID); tableBlock != nil {
			blocks = append(blocks, *tableBlock)
		}
	}

	return blocks, nil
}

func createTextBlock(blockID *int, text string) ContentBlock {
	block := ContentBlock{
		ID:      fmt.Sprintf("block-%d", *blockID),
		Type:    "paragraph",
		Content: TextContent{Text: text},
	}
	*blockID++
	return block
}

func parseMarkdownTable(lines []string, blockID *int) *ContentBlock {
	if len(lines) < 2 {
		return nil
	}

	// Parse headers (first line)
	headerLine := strings.Trim(lines[0], " ")
	headers := strings.Split(headerLine, "|")
	for i := range headers {
		headers[i] = strings.TrimSpace(headers[i])
	}
	// Remove empty first/last elements if present
	if len(headers) > 0 && headers[0] == "" {
		headers = headers[1:]
	}
	if len(headers) > 0 && headers[len(headers)-1] == "" {
		headers = headers[:len(headers)-1]
	}

	// Skip separator line (second line with dashes)
	var rows [][]string
	for i := 2; i < len(lines); i++ {
		rowLine := strings.Trim(lines[i], " ")
		if rowLine == "" {
			continue
		}
		cells := strings.Split(rowLine, "|")
		for j := range cells {
			cells[j] = strings.TrimSpace(cells[j])
		}
		// Remove empty first/last elements
		if len(cells) > 0 && cells[0] == "" {
			cells = cells[1:]
		}
		if len(cells) > 0 && cells[len(cells)-1] == "" {
			cells = cells[:len(cells)-1]
		}
		if len(cells) > 0 {
			rows = append(rows, cells)
		}
	}

	block := &ContentBlock{
		ID:   fmt.Sprintf("block-%d", *blockID),
		Type: "table",
		Content: TableContent{
			Headers: headers,
			Rows:    rows,
		},
	}
	*blockID++
	return block
}
