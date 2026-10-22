package cssstrings

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
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
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(t.TempDir(), "gap")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	actual := stringsExecute(t, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	clean(t, "native ASan/UBSan/LSan", actual)
	equal(t, "native", actual.stdout, answer.stdout)
	emitted := filepath.Join(t.TempDir(), "gap.mjs")
	write(t, emitted, []byte(javascript.JavaScript(lowered)))
	backend := onNode(t, emitted)
	clean(t, "JavaScript backend", backend)
	equal(t, "JavaScript backend", backend.stdout, answer.stdout)
}
