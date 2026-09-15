package quickadd

import (
	"slices"
	"strings"
	"unicode/utf8"
)

// TagSuggestions builds full input values for textinput's native completion.
// ponytail: completion is end-of-input only; use token replacement if mid-line completion is needed.
func TagSuggestions(line string, cursor int, known Prefixes, defaultProject string, tags []string) (string, []string) {
	if cursor != utf8.RuneCountInString(line) {
		return "", nil
	}
	start := strings.LastIndexAny(line, " \t") + 1
	word := line[start:]
	if !strings.HasPrefix(word, "#") {
		return "", nil
	}
	// Replace the unfinished tag with title text so the normal grammar resolves
	// the project and detects invalid flags, even before a title has been typed.
	spec, err := Parse(line[:start]+"completion", known, defaultProject)
	if err != nil {
		return "", nil
	}
	for _, token := range spec.Tokens {
		if token.Kind == KindSeparator {
			return "", nil
		}
	}
	partial := word[1:]
	var out []string
	for _, tag := range tags {
		if tagRe.MatchString(tag) && tag != partial && strings.HasPrefix(tag, partial) && !slices.Contains(spec.Tags, tag) {
			out = append(out, line[:start]+"#"+tag)
		}
	}
	return spec.Project, out
}
