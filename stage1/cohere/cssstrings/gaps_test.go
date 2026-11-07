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

func TestMultiPushGap(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs("gaps/1_multi_push.ts")
	if err != nil {
		test.Fatal(err)
	}
	answer := onNode(test, path)
	clean(test, "Node gap", answer)
	equal(test, "Node gap", answer.stdout, []byte("ab\n"))
	program, err := load.Load([]string{path})
	if err != nil {
		test.Fatal(err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		test.Fatal(err)
	}
	binary := filepath.Join(test.TempDir(), "gap")
	if err := native.Build(native.C(lowered), binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	actual := stringsExecute(test, []string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	clean(test, "native ASan/UBSan/LSan", actual)
	equal(test, "native", actual.stdout, answer.stdout)
	emitted := filepath.Join(test.TempDir(), "gap.mjs")
	write(test, emitted, []byte(javascript.JavaScript(lowered)))
	backend := onNode(test, emitted)
	clean(test, "JavaScript backend", backend)
	equal(test, "JavaScript backend", backend.stdout, answer.stdout)
}
