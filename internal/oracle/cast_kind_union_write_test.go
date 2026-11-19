package oracle

import (
	"strings"
	"testing"
)

// A value built as a Node, whose kind is the open SyntaxKind, holding kind ElementAccessExpression and no flags, cast
// to a union of kinded interfaces and written (#b5w3ycg head 2, Adamic's ruling of Nov 17): the kind test passes, so
// main's checked write is what stops it, at the write, with the write's own message, on both backends alike, never
// at the cast and never storing. Node runs on.
func TestKindUnionWriteLieStopsAtTheWrite(t *testing.T) {
	t.Parallel()
	program, path := castFixture(t, "cast_kind_union_write_lie.ts")
	native, _ := natively(t, program)
	javaScript := onJavaScriptBackend(t, program)
	want := "adamic: panic: field read failed: <write>.flags is not initialized; expected number, found missing\n"
	for backend, result := range map[string]run{"native": native, "JavaScript": javaScript} {
		if result.exitCode != 70 || string(result.stdout) != "writing\n" || string(result.stderr) != want || strings.Contains(string(result.stderr), "cast failed") {
			t.Errorf("%s: exit %d, stdout %q, stderr %q; want 70 at the write and %q", backend, result.exitCode, result.stdout, result.stderr, want)
		}
	}
	if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "writing\nafter\n" {
		t.Fatalf("source on Node must run on: exit %d stdout %q", node.exitCode, node.stdout)
	}
}

func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/cast_kind_union_write.ts", true, false})
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/cast_kind_union_write_lie.ts", true, true})
}
