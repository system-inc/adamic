package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestMixedEqualityBoxedNullBoundary(t *testing.T) {
	for _, sameRepresentation := range []bool{false, true} {
		path := filepath.Join(t.TempDir(), "main.a")
		source := `interface Sized {readonly length:number;}
function compare(left:RegExpExecArray|null,right:Sized|undefined): void {
 console.log("" + (left === right) + ":" + (left !== right));
}
compare(null,undefined);
compare(/a/.exec('a'),{length:1});
`
		// Templates are the existing scalar spelling path, independently held to Node.
		source = strings.Replace(source, `console.log("" + (left === right) + ":" + (left !== right));`, "console.log(`${left === right}:${left !== right}`);", 1)
		if sameRepresentation {
			source = strings.Replace(source, "right:Sized|undefined", "right:RegExpExecArray|undefined", 1)
			source = strings.Replace(source, "compare(/a/.exec('a'),{length:1});", "compare(/a/.exec('a'),/a/.exec('a') ?? undefined);", 1)
		}
		if err := os.WriteFile(path, []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
		node := onNode(t, path)
		if node.exitCode != 0 || string(node.stdout) != "false:true\nfalse:true\n" {
			t.Fatalf("Node equality control: %#v", node)
		}
		program, err := lowered(t, path)
		if err == nil {
			native, _ := natively(t, program)
			backend := onJavaScriptBackend(t, program)
			t.Fatalf("expected explicit boxed-null equality boundary; lowering succeeded: Node %q native %q (exit %d) backend %q (exit %d)", node.stdout, native.stdout, native.exitCode, backend.stdout, backend.exitCode)
		}
		var notYet *lower.NotYet
		if !errors.As(err, &notYet) || !strings.Contains(notYet.What, "strict comparison requiring a boxed null distinct from undefined") {
			t.Fatalf("wrong null boundary: %v", err)
		}
	}
}
