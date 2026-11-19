package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A value built as a Node, whose kind is the open SyntaxKind, holding kind ElementAccessExpression and no flags, cast
// to a union of kinded interfaces and written (#b5w3ycg head 2): Adamic's amendment of Nov 19 judges the write by
// what reaches it, and the Node built without flags holds no slot for it, so stage 0 refuses the program at the
// write, at compile time, never a stop at run time. Node runs on.
func TestKindUnionWriteLieIsRefusedAtTheWrite(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/cast_kind_union_write_lie.ts"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) {
		t.Fatalf("want NotYet at the write, got %v", err)
	}
	want := "a write of flags through a kind-checked cast's union, where a value reaching it is built as Node without flags"
	if notYet.What != want || !strings.HasSuffix(notYet.Where, "cast_kind_union_write_lie.ts:9:2") {
		t.Fatalf("want NotYet at cast_kind_union_write_lie.ts:9:2, %q, got %v", want, err)
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
	}{"internal/oracle/testdata/cast_kind_union_write_lie.ts", false, false})
}
