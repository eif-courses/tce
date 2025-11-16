package analyzer

// LocalizeComments converts Lithuanian comment texts to English when lang == "en".
// If lang is anything else, comments are returned unchanged.
func LocalizeComments(in []Comment, lang string) []Comment {
	if lang != "en" {
		return in
	}

	out := make([]Comment, len(in))
	copy(out, in)

	for i, c := range out {
		switch c.ID {
		case "R-INTRO-GOAL-1":
			out[i].Title = "Missing thesis goal"
			out[i].Message = "No clear thesis goal was found in the introduction. It is recommended to add a sentence like \"The aim of this thesis is ...\"."

		case "R-INTRO-TASKS-1":
			out[i].Title = "Missing list of tasks"
			out[i].Message = "The introduction does not contain a clear list of tasks. It is recommended to present a numbered list with the main tasks of the thesis."

		case "R-SEC-MISSING-intro":
			out[i].Title = "Missing section: Introduction"
			out[i].Message = "According to the methodological guidelines, the thesis must contain an \"Introduction\" section, but it was not found."

		case "R-SEC-MISSING-task":
			out[i].Title = "Missing section: Practice task"
			out[i].Message = "According to the methodological guidelines, the thesis must contain a section describing the practice task and system, but it was not found."

		case "R-SEC-MISSING-fr":
			out[i].Title = "Missing section: Functional requirements"
			out[i].Message = "According to the methodological guidelines, the thesis must contain a \"Functional requirements\" section, but it was not found."

		case "R-SEC-MISSING-analysis":
			out[i].Title = "Missing section: Task analysis"
			out[i].Message = "According to the methodological guidelines, the thesis must contain a \"Task analysis\" section, but it was not found."

		case "R-SEC-LEN-intro":
			out[i].Title = "Introduction is too short"
			out[i].Message = "The \"Introduction\" section seems too short. According to the guidelines, it should provide more detailed context and explanation."

		case "R-SEC-LEN-task":
			out[i].Title = "Practice task section is too short"
			out[i].Message = "The section describing the practice task and system seems too short. It should describe the context, system, and your role more clearly."

		case "R-FR-NO-F-ITEMS":
			out[i].Title = "Functional requirements not clearly listed"
			out[i].Message = "In the \"Functional requirements\" section no clear requirement list (e.g., F1, F2, ...) was detected. It is recommended to list the requirements in a structured way."

		// language rules you will add later:
		case "R-LANG-FIRST-PERSON":
			out[i].Title = "Too much first-person usage"
			out[i].Message = "This paragraph uses first-person pronouns (\"I\", \"we\"). For academic writing a more neutral style is recommended."

		case "R-LANG-INFORMAL":
			out[i].Title = "Informal wording"
			out[i].Message = "Informal or colloquial terms were detected. A more formal wording is recommended for the thesis."

		default:
			// fallback: translate category labels a bit
			switch c.Category {
			case "structure":
				out[i].SectionLabel = "Structure"
			case "content":
				out[i].SectionLabel = "Content / methodology"
			case "language":
				out[i].SectionLabel = "Language / style"
			case "formatting":
				out[i].SectionLabel = "Formatting"
			}
		}
	}

	return out
}
