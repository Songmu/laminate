package laminate

import (
	"fmt"

	"github.com/gobwas/glob"
)

// FindMatchingCommand finds the first command that matches the given language
func FindMatchingCommand(commands []Command, lang string) (*Command, error) {
	for _, cmd := range commands {
		matched, err := matchLanguage(cmd.Lang, lang)
		if err != nil {
			return nil, fmt.Errorf("failed to match language pattern %q: %w", cmd.Lang, err)
		}
		if matched {
			return &cmd, nil
		}
	}
	return nil, fmt.Errorf("no matching command found for language: %s", lang)
}

// matchLanguage checks if a language matches a pattern
func matchLanguage(pattern, lang string) (bool, error) {
	g, err := glob.Compile(pattern)
	if err != nil {
		return false, err
	}
	return g.Match(lang), nil
}
