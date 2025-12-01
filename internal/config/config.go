// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// SectionRule describes one required/expected section from guidelines.
type SectionRule struct {
	ID                 string   `yaml:"id"`
	Label              string   `yaml:"label"`
	Requirement        string   `yaml:"requirement"`
	MinPages           int      `yaml:"min_pages"`
	MaxPages           int      `yaml:"max_pages"`
	Required           bool     `yaml:"required"`
	MustContainKeywords []string `yaml:"must_contain_keywords"`
	ForbiddenPhrases   []string `yaml:"forbidden_phrases"`
	ForbiddenPatterns  []string `yaml:"forbidden_patterns"`
	RequireCitations   bool     `yaml:"require_citations"`
}

// LanguageRule defines language-related checks (first-person, informal, etc.)
type LanguageRule struct {
	FirstPerson              []string `yaml:"first_person"`
	Informal                 []string `yaml:"informal"`
	MaxParagraphLengthChars  int      `yaml:"max_paragraph_length_chars"`
	AllowFirstPersonInSections []string `yaml:"allow_first_person_in_sections"`
}

// StructureRule defines the overall document structure requirements
type StructureRule struct {
	RequiredOrder []string `yaml:"required_order"`
}

// LayoutRule defines formatting requirements
type LayoutRule struct {
	Body     BodyLayout     `yaml:"body"`
	Headings HeadingsLayout `yaml:"headings"`
}

type BodyLayout struct {
	FontFamily       string  `yaml:"font_family"`
	FontSizePt       int     `yaml:"font_size_pt"`
	LineSpacing      float64 `yaml:"line_spacing"`
	Justify          bool    `yaml:"justify"`
	FirstLineIndentCm float64 `yaml:"first_line_indent_cm"`
}

type HeadingsLayout struct {
	H1 HeadingStyle `yaml:"h1"`
	H2 HeadingStyle `yaml:"h2"`
	H3 HeadingStyle `yaml:"h3"`
}

type HeadingStyle struct {
	FontSizePt int    `yaml:"font_size_pt"`
	Uppercase  bool   `yaml:"uppercase"`
	Bold       bool   `yaml:"bold"`
	Align      string `yaml:"align"`
}

// FiguresTablesRule defines requirements for images and tables
type FiguresTablesRule struct {
	RequireCaption         bool     `yaml:"require_caption"`
	FigureCaptionPattern   string   `yaml:"figure_caption_pattern"`
	TableCaptionPattern    string   `yaml:"table_caption_pattern"`
	RequireReferenceInText bool     `yaml:"require_reference_in_text"`
	ReferencePatterns      []string `yaml:"reference_patterns"`
	ForbidDuplicateData    bool     `yaml:"forbid_duplicate_data"`
}

// ReferencesRule defines citation and bibliography requirements
type ReferencesRule struct {
	MinSources              int      `yaml:"min_sources"`
	SectionID               string   `yaml:"section_id"`
	IntextPatterns          []string `yaml:"intext_patterns"`
	ForbiddenIntextPatterns []string `yaml:"forbidden_intext_patterns"`
	AllowWebSources         bool     `yaml:"allow_web_sources"`
	WebSourceLimitRatio     float64  `yaml:"web_source_limit_ratio"`
	RequireYear             bool     `yaml:"require_year"`
	RequireAuthorOrOrg      bool     `yaml:"require_author_or_org"`
}

// ContentRulesConfig defines content-specific validation rules
type ContentRulesConfig struct {
	ForbiddenGlobalPhrases []string          `yaml:"forbidden_global_phrases"`
	MethodologyRequired    MethodologyRule   `yaml:"methodology_required"`
	ConclusionsQuality     ConclusionsRule   `yaml:"conclusions_quality"`
}

type MethodologyRule struct {
	Sections       []string `yaml:"sections"`
	MustContainAny []string `yaml:"must_contain_any"`
}

type ConclusionsRule struct {
	SectionID      string   `yaml:"section_id"`
	MustContainAny []string `yaml:"must_contain_any"`
	ForbiddenAny   []string `yaml:"forbidden_any"`
}

// CodeRule defines rules for code fragments
type CodeRule struct {
	DetectBlocks         bool     `yaml:"detect_blocks"`
	MaxInlineLines       int      `yaml:"max_inline_lines"`
	DetectPatterns       []string `yaml:"detect_patterns"`
	ForbiddenSections    []string `yaml:"forbidden_sections"`
	AllowedOnlySection   string   `yaml:"allowed_only_section"`
	RequiredFontFamily   []string `yaml:"required_font_family"`
	CaptionRequired      bool     `yaml:"caption_required"`
	CaptionPattern       string   `yaml:"caption_pattern"`
}

// SeverityConfig defines severity levels for different violation types
type SeverityConfig struct {
	MissingRequiredSection   string `yaml:"missing_required_section"`
	SectionTooShort          string `yaml:"section_too_short"`
	WrongOrder               string `yaml:"wrong_order"`
	FormattingViolation      string `yaml:"formatting_violation"`
	MissingFigureCaption     string `yaml:"missing_figure_caption"`
	MissingFigureReference   string `yaml:"missing_figure_reference"`
	InformalLanguage         string `yaml:"informal_language"`
	FirstPersonUsage         string `yaml:"first_person_usage"`
	MissingCitationInTheory  string `yaml:"missing_citation_in_theory"`
	LowReferenceCount        string `yaml:"low_reference_count"`
	ForbiddenPhrase          string `yaml:"forbidden_phrase"`
}

// RuleProfile is one configuration profile for (program, lang).
type RuleProfile struct {
	Program       string             `yaml:"program"`
	Lang          string             `yaml:"lang"`
	Structure     StructureRule      `yaml:"structure"`
	Sections      []SectionRule      `yaml:"-"` // populated from Structure.Sections
	Language      LanguageRule       `yaml:"language"`
	Layout        LayoutRule         `yaml:"layout"`
	FiguresTables FiguresTablesRule  `yaml:"figures_tables"`
	References    ReferencesRule     `yaml:"references"`
	ContentRules  ContentRulesConfig `yaml:"content_rules"`
	Code          CodeRule           `yaml:"code"`
	Severity      SeverityConfig     `yaml:"severity"`
}

// Current holds the latest loaded profile (used by analyzer).
var Current RuleProfile

// LoadRulesForProfile loads YAML from config/rules_<program>_<lang>.yaml
// and stores it in Current. Returns error if file missing or invalid.
func LoadRulesForProfile(program, lang string) error {
	p := strings.ToLower(strings.TrimSpace(program))
	l := strings.ToLower(strings.TrimSpace(lang))
	if l == "" {
		l = "lt"
	}
	if p == "" {
		p = "pi"
	}

	filename := fmt.Sprintf("rules_%s_%s.yaml", p, l)
	path := filepath.Join("config", filename)

	data, err := os.ReadFile(path)
	if err != nil {
		// Do not crash app – just return error, caller may log + keep defaults
		return fmt.Errorf("read rules file %s: %w", path, err)
	}

	// Parse into a temporary map to handle nested structure
	var rawProfile map[string]interface{}
	if err := yaml.Unmarshal(data, &rawProfile); err != nil {
		return fmt.Errorf("unmarshal rules file %s: %w", path, err)
	}

	// Now unmarshal into the full profile struct
	var profile RuleProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("unmarshal rules file into profile %s: %w", path, err)
	}

	// Extract sections from structure.sections
	if structureData, ok := rawProfile["structure"].(map[string]interface{}); ok {
		if sectionsData, ok := structureData["sections"].([]interface{}); ok {
			for _, secData := range sectionsData {
				secBytes, _ := yaml.Marshal(secData)
				var sec SectionRule
				if err := yaml.Unmarshal(secBytes, &sec); err == nil {
					profile.Sections = append(profile.Sections, sec)
				}
			}
		}
	}

	// Fallback: if file does not contain program/lang, fill from params
	if profile.Program == "" {
		profile.Program = p
	}
	if profile.Lang == "" {
		profile.Lang = l
	}

	Current = profile
	return nil
}
