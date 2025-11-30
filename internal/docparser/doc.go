package docparser

// Section represents a logical document section.
type Section struct {
	ID          string
	Label       string
	Requirement string
	Issues      int
}

// ContentBlock represents different types of content in a document
type ContentBlock struct {
	ID        string      `json:"id"`
	SectionID string      `json:"sectionId"`
	Type      string      `json:"type"` // "paragraph", "image", "table", "heading", "list"
	Content   interface{} `json:"content"`
}

// TextContent for regular paragraphs
type TextContent struct {
	Text string `json:"text"`
}

// ImageContent for embedded images
type ImageContent struct {
	Alt      string `json:"alt"`
	Base64   string `json:"base64,omitempty"` // Base64 encoded image data
	Caption  string `json:"caption,omitempty"`
	Width    int    `json:"width,omitempty"`
	Height   int    `json:"height,omitempty"`
}

// TableContent for tables
type TableContent struct {
	Headers []string   `json:"headers"`
	Rows    [][]string `json:"rows"`
	Caption string     `json:"caption,omitempty"`
}

// HeadingContent for headings
type HeadingContent struct {
	Text  string `json:"text"`
	Level int    `json:"level"` // 1-6
}

// ListContent for bulleted or numbered lists
type ListContent struct {
	Items   []string `json:"items"`
	Ordered bool     `json:"ordered"`
}

// Paragraph is a paragraph linked to a section (legacy, kept for compatibility).
type Paragraph struct {
	ID        string
	SectionID string
	Text      string
}
