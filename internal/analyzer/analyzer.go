package analyzer

import "github.com/eif-courses/tce/internal/docparser"

func Analyze(raw *docparser.RawDoc, lang string) Result {
	// 1) Normalize raw paragraphs
	clean := docparser.NormalizeDoc(raw)

	// 2) Detect sections and assign paragraph IDs (language-aware)
	sections, paragraphs := docparser.ParseSections(clean, lang)

	// 3) Run rules (structure/content/etc.)
	comments, sectionsWithCounts := runRules(sections, paragraphs)

	// 4) Localize comments if necessary (EN vs LT)
	comments = LocalizeComments(comments, lang)

	return Result{
		Sections:   sectionsWithCounts,
		Paragraphs: paragraphs,
		Comments:   comments,
	}
}
