package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"os"
	"strings"
	"testing"
)

func TestV4UnrelatedArrayFieldWrite(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/review/agree/fxspptb_oct9_native_p08_null_array_write.a")
	if err != nil {
		t.Fatal(err)
	}
	v4EscapeSource(t, string(source), "3\n2\ntrue\n")
	program, truth := v4IdentityProgram(t, string(source), "3\n2\ntrue\n")
	changed := 0
	for i, node := range program.Main {
		if write, ok := node.(ir.SetProperty); ok && write.ViewWriteUnrelated {
			write.ViewWriteUnrelated = false
			program.Main[i] = write
			changed++
		}
	}
	if changed != 2 {
		t.Fatalf("unrelated write proof mutant sites: %d", changed)
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || disagreement(truth, got) == "" || !strings.Contains(string(got.stderr), "<write>.value") || strings.Contains(string(got.stderr), "Sanitizer") {
			t.Fatalf("%s unrelated proof omission: %#v", backend, got)
		}
	}
}
