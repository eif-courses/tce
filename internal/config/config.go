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
	ID          string `yaml:"id"`
	Label       string `yaml:"label"`
	Requirement string `yaml:"requirement"`
	MinPages    int    `yaml:"min_pages"`
	Required    bool   `yaml:"required"`
}

// LanguageRule defines language-related checks (first-person, informal, etc.)
type LanguageRule struct {
	FirstPerson []string `yaml:"first_person"`
	Informal    []string `yaml:"informal"`
}

// RuleProfile is one configuration profile for (program, lang).
type RuleProfile struct {
	Program  string        `yaml:"program"`
	Lang     string        `yaml:"lang"`
	Sections []SectionRule `yaml:"sections"`
	Language LanguageRule  `yaml:"language"`
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

	var profile RuleProfile
	if err := yaml.Unmarshal(data, &profile); err != nil {
		return fmt.Errorf("unmarshal rules file %s: %w", path, err)
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
