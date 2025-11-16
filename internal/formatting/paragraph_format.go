package formatting

func IsJustified(f ParagraphFormat) bool {
	return f.Align == "justify"
}

func HasAtLeastLineSpacing(f ParagraphFormat, min float64) bool {
	if f.LineSpacing == 0 {
		return false
	}
	return f.LineSpacing >= min
}
