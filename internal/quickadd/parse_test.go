package quickadd

import (
	"errors"
	"slices"
	"testing"
)

var known = Prefixes{"tui", "tasks", "ops"}

func TestSpecExampleFromSpec(t *testing.T) {
	line := "?tint the projects strip with each accent #ui #identity !3 ~s"
	spec, err := Parse(line, known, "tui")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"add", "tint the projects strip with each accent", "--project", "tui", "--status", "idea", "-p", "3", "--size", "s", "--tag", "ui", "--tag", "identity"}
	if got := spec.Args(); !slices.Equal(got, want) {
		t.Fatalf("args\n got %q\nwant %q", got, want)
	}
	kinds := []Kind{}
	for _, tok := range spec.Tokens {
		kinds = append(kinds, tok.Kind)
	}
	if !slices.Equal(kinds, []Kind{KindIdea, KindTag, KindTag, KindPriority, KindSize}) {
		t.Fatalf("token kinds %v", kinds)
	}
	if spec.Tokens[0] != (Token{0, 1, KindIdea}) || spec.Tokens[1].Start != 42 {
		t.Fatalf("token spans %+v", spec.Tokens)
	}
}

func TestEveryComplexityProjectAndBody(t *testing.T) {
	spec, err := Parse("rotate the log @30d ^mid >ops -- body text  with   spaces  ", known, "tui")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"add", "rotate the log", "--project", "ops", "--complexity", "mid", "--every", "30d", "-b", "body text  with   spaces  "}
	if got := spec.Args(); !slices.Equal(got, want) {
		t.Fatalf("args %q", got)
	}
	if spec.Every != "30d" || spec.Project != "ops" || spec.Body != "body text  with   spaces  " {
		t.Fatalf("spec %+v (body is verbatim, trailing spaces kept)", spec)
	}
	if spec, _ := Parse("x --", known, "tui"); spec.Body != "" {
		t.Fatalf("a bare separator has an empty body, got %q", spec.Body)
	}
}

func TestTitleTextKeepsMidWordMarkers(t *testing.T) {
	spec, err := Parse("C# wow! and a/b", known, "tui")
	if err != nil {
		t.Fatal(err)
	}
	if spec.Title != "C# wow! and a/b" || len(spec.Tokens) != 0 {
		t.Fatalf("spec %+v", spec)
	}
}

func TestErrors(t *testing.T) {
	cases := map[string]string{
		"fix !9":        "!9",
		"fix ~huge":     "~huge",
		"fix ^very":     "^very",
		"fix @1h":       "@1h",
		"fix @0d":       "@0d",
		"fix >nope":     ">nope",
		"fix #":         "#",
		"fix #ba!d":     "#ba!d",
		"!1 fix !2":     "!2",
		"fix >tui >ops": ">ops",
		"#a -- body":    "",
		"   ":           "",
		"fix ? later":   "?",
	}
	for line, word := range cases {
		_, err := Parse(line, known, "tui")
		var pe *ParseError
		if !errors.As(err, &pe) {
			t.Errorf("%q: want ParseError, got %v", line, err)
			continue
		}
		if pe.Word != word {
			t.Errorf("%q: error on %q, want %q (%s)", line, pe.Word, word, pe.Msg)
		}
	}
	if _, err := Parse("fix it", known, ""); err == nil {
		t.Fatal("no project anywhere must error")
	}
}

func TestErrorStillReportsTokensSoFar(t *testing.T) {
	spec, err := Parse("fix #x !9", known, "tui")
	if err == nil || len(spec.Tokens) != 2 || spec.Tokens[1].Kind != KindError {
		t.Fatalf("spec=%+v err=%v", spec, err)
	}
}
