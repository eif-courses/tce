package analyzer

import "github.com/eif-courses/tce/internal/docparser"

func Analyze(raw *docparser.RawDoc, lang string) Result {
	// 1) Normalize raw paragraphs
	clean := docparser.NormalizeDoc(raw)

	// 2) Detect sections and assign paragraph IDs (language-aware)
	sections, paragraphs := docparser.ParseSections(clean, lang)

	// 3) Process content blocks if available
	contentBlocks := docparser.ParseContentBlocks(raw.ContentBlocks, lang)

	// 4) Run rules (structure/content/etc.)
	comments, sectionsWithCounts := runRules(sections, paragraphs)

	// 5) Localize comments if necessary (EN vs LT)
	comments = LocalizeComments(comments, lang)

	return Result{
		Sections:      sectionsWithCounts,
		Paragraphs:    paragraphs,
		ContentBlocks: contentBlocks,
		Comments:      comments,
	}
}
