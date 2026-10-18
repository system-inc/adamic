package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func v4SurfaceUnsupported(t *testing.T, invocation, reason, fix string) {
	t.Helper()
	source := v4EscapeDeclarations + `const raw={kind:'Receiver' as const,run:(value:'ok'):string=>{console.log('producer');return value;}};
function probe(base:Base):void {const view=base as Target;const escaped=view.run;const argumentsList:[string]=['bad'];console.log(` + invocation + `);}
probe(raw);`
	path := filepath.Join(t.TempDir(), "unsupported-surface.ts")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "producer\nbad\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node truth: %#v", truth)
	}
	_, err := lowered(t, path)
	var refusal *lower.Refused
	if !errors.As(err, &refusal) || !strings.Contains(refusal.Where, "unsupported-surface.ts:4:") || refusal.What != reason || refusal.Fix != fix {
		t.Fatalf("compile-time refusal with path and fix: %v", err)
	}
	t.Logf("Node succeeds; compiler refuses before either backend: %v", err)
}

func TestV4EscapeAdapterDynamicApplyRefusal(t *testing.T) {
	t.Parallel()
	v4SurfaceUnsupported(t, "escaped.apply(undefined,argumentsList)", "apply with a dynamic argument list", "spell the arguments in a dense literal or use call until dynamic apply argument lists are checked")
}
func TestV4EscapeAdapterPrimitiveCallRefusal(t *testing.T) {
	t.Parallel()
	v4SurfaceUnsupported(t, "escaped.call('receiver','bad')", "call or apply with a primitive or unrepresented thisArg", "use an object or undefined receiver until checked primitive receivers are supported")
}
func TestV4EscapeAdapterPrimitiveApplyRefusal(t *testing.T) {
	t.Parallel()
	v4SurfaceUnsupported(t, "escaped.apply(7,['bad'])", "call or apply with a primitive or unrepresented thisArg", "use an object or undefined receiver until checked primitive receivers are supported")
}
func TestV4EscapeAdapterNullablePrimitiveCallRefusal(t *testing.T) {
	t.Parallel()
	v4SurfaceUnsupported(t, "escaped.call('receiver' as string|null,'bad')", "call or apply with a primitive or unrepresented thisArg", "use an object or undefined receiver until checked primitive receivers are supported")
}
