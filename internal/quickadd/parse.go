// Package quickadd parses the one-line grammar of spec §6.1 into a tasks add argv.
// A word that starts with a marker and does not match its rule is an error, never
// title text; nothing is defaulted.
package quickadd

import (
	"regexp"
	"slices"
	"strconv"
	"strings"
)

type Kind int

const (
	KindTag Kind = iota + 1
	KindPriority
	KindSize
	KindComplexity
	KindEvery
	KindProject
	KindIdea
	KindSeparator
	KindError
)

// Token is a byte span of the input line and what it parsed as, for styling.
type Token struct {
	Start, End int
	Kind       Kind
}

type Prefixes []string

func (p Prefixes) Has(s string) bool {
	return slices.Contains(p, s)
}

type Spec struct {
	Title      string
	Project    string
	Body       string
	Idea       bool
	Priority   *int
	Size       string
	Complexity string
	Every      string
	Tags       []string
	Tokens     []Token
}

// ParseError names the offending word; Msg is shown in the preview line.
type ParseError struct {
	Word string
	Msg  string
}

func (e *ParseError) Error() string {
	if e.Word == "" {
		return e.Msg
	}
	return e.Word + ": " + e.Msg
}

var (
	tagRe   = regexp.MustCompile(`^[A-Za-z0-9_:./-]+$`)
	everyRe = regexp.MustCompile(`^([1-9][0-9]*)([dw])$`)
	sizes   = []string{"xs", "s", "m", "l", "xl"}
	levels  = []string{"low", "mid", "high"}
)

type word struct {
	text       string
	start, end int
}

func words(line string) []word {
	var out []word
	i := 0
	for i < len(line) {
		for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
			i++
		}
		j := i
		for j < len(line) && line[j] != ' ' && line[j] != '\t' {
			j++
		}
		if j > i {
			out = append(out, word{line[i:j], i, j})
		}
		i = j
	}
	return out
}

// Parse reads line. known is the registry's prefixes; defaultProject is the view's
// project ("" when the view has none). Tokens are returned even on error, with the
// failing word as KindError, so the input can stay styled while the error shows.
func Parse(line string, known Prefixes, defaultProject string) (Spec, error) {
	spec := Spec{Project: defaultProject}
	var title []string
	explicitProject := false
	fail := func(w word, msg string) (Spec, error) {
		spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindError})
		return spec, &ParseError{Word: w.text, Msg: msg}
	}
	for i, w := range words(line) {
		if w.text == "--" {
			spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindSeparator})
			spec.Body = strings.TrimPrefix(line[w.end:], " ")
			break
		}
		if i == 0 && strings.HasPrefix(w.text, "?") {
			spec.Idea = true
			spec.Tokens = append(spec.Tokens, Token{w.start, w.start + 1, KindIdea})
			if rest := w.text[1:]; rest != "" {
				title = append(title, rest)
			}
			continue
		}
		switch w.text[0] {
		case '#':
			tag := w.text[1:]
			if !tagRe.MatchString(tag) {
				return fail(w, "a tag is letters, digits, and _ : . / -")
			}
			spec.Tags = append(spec.Tags, tag)
			spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindTag})
		case '!':
			n, err := strconv.Atoi(w.text[1:])
			if err != nil || n < 0 || n > 4 {
				return fail(w, "priority must be 0–4")
			}
			if spec.Priority != nil {
				return fail(w, "priority given twice")
			}
			spec.Priority = &n
			spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindPriority})
		case '~':
			if !slices.Contains(sizes, w.text[1:]) {
				return fail(w, "size must be xs, s, m, l, or xl")
			}
			if spec.Size != "" {
				return fail(w, "size given twice")
			}
			spec.Size = w.text[1:]
			spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindSize})
		case '^':
			if !slices.Contains(levels, w.text[1:]) {
				return fail(w, "complexity must be low, mid, or high")
			}
			if spec.Complexity != "" {
				return fail(w, "complexity given twice")
			}
			spec.Complexity = w.text[1:]
			spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindComplexity})
		case '@':
			if !everyRe.MatchString(w.text[1:]) {
				return fail(w, "recurrence is <n>d or <n>w, n ≥ 1")
			}
			if spec.Every != "" {
				return fail(w, "recurrence given twice")
			}
			spec.Every = w.text[1:]
			spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindEvery})
		case '>':
			project := w.text[1:]
			if !known.Has(project) {
				return fail(w, "not a registered project")
			}
			if explicitProject {
				return fail(w, "project given twice")
			}
			explicitProject = true
			spec.Project = project
			spec.Tokens = append(spec.Tokens, Token{w.start, w.end, KindProject})
		case '?':
			return fail(w, "? marks an idea only as the first word")
		default:
			title = append(title, w.text)
		}
	}
	spec.Title = strings.Join(title, " ")
	if spec.Title == "" {
		return spec, &ParseError{Msg: "title is empty"}
	}
	if spec.Project == "" {
		return spec, &ParseError{Msg: "no project: add >prefix"}
	}
	return spec, nil
}

// Args is the tasks argv, project always explicit, nothing defaulted, no --agent (§6.1).
func (s Spec) Args() []string {
	args := []string{"add", s.Title, "--project", s.Project}
	if s.Idea {
		args = append(args, "--status", "idea")
	}
	if s.Priority != nil {
		args = append(args, "-p", strconv.Itoa(*s.Priority))
	}
	if s.Size != "" {
		args = append(args, "--size", s.Size)
	}
	if s.Complexity != "" {
		args = append(args, "--complexity", s.Complexity)
	}
	if s.Every != "" {
		args = append(args, "--every", s.Every)
	}
	for _, tag := range s.Tags {
		args = append(args, "--tag", tag)
	}
	if s.Body != "" {
		args = append(args, "-b", s.Body)
	}
	return args
}
