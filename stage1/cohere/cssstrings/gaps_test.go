package cssstrings

import (
	"context"
	"errors"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestMultiPushGap(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs("gaps/1_multi_push.ts")
	if err != nil {
		t.Fatal(err)
	}
	answer := onNode(t, path)
	clean(t, "Node gap", answer)
	equal(t, "Node gap", answer.stdout, []byte("ab\n"))
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "push with other than one value" {
		t.Fatalf("gap changed: %v; update GAPS.md and remove the workaround if closed", err)
	}
}
