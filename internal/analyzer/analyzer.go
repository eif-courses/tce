package analyzer

import (
	"github.com/eif-courses/tce/internal/docparser"
)

func Analyze(raw *docparser.RawDoc) Result {
	// 1) Normalize raw paragraphs
	clean := docparser.NormalizeDoc(raw)

	// 2) Detect sections and assign paragraph IDs
	sections, paragraphs := docparser.ParseSections(clean)

	// 3) Run rules (structure/content/etc.)
	comments, sectionsWithCounts := runRules(sections, paragraphs)

	return Result{
		Sections:   sectionsWithCounts,
		Paragraphs: paragraphs,
		Comments:   comments,
	}
}
