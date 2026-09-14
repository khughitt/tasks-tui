package tasksctl

import (
	"context"
	"errors"
	"slices"
	"testing"
)

func TestTagsUsesProjectAndDecodesWarnings(t *testing.T) {
	r := &argvRunner{reply: `{"tags":[{"tag":"bug","meaning":null,"count":2,"projects":{"tui":2}}],"warnings":["lookup notice"]}`}
	res, err := (&Client{R: r}).Tags(context.Background(), "tui")
	if err != nil || len(res.Tags) != 1 || res.Tags[0].Tag != "bug" || !slices.Equal(res.Warnings, []string{"lookup notice"}) {
		t.Fatalf("result=%+v err=%v", res, err)
	}
	if r.dirs[0] != "" || !slices.Equal(r.calls[0], []string{"tags", "--project", "tui"}) {
		t.Fatalf("unexpected invocation: %v in %v", r.calls, r.dirs)
	}
	for _, raw := range []string{`{"tags":null,"warnings":[]}`, `{"tags":[{}],"warnings":[]}`, `{"tags":[{"tag":null}],"warnings":[]}`, `{"tags":[null],"warnings":[]}`, `{"tags":[]}`} {
		_, err := (&Client{R: literalRunner(raw)}).Tags(context.Background(), "tui")
		var typed *Error
		if !errors.As(err, &typed) || typed.Kind != "decode" {
			t.Fatalf("%s: want decode error, got %v", raw, err)
		}
	}
}
