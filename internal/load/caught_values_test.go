package load

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestAdamicCaughtMessageRequiresProof(t *testing.T) {
	t.Parallel()
	paths := writeProgram(t, [2]string{"main.a", "function accepts(message: string): void { console.log(message); }\ntry { throw 'x'; } catch (e) { accepts(e.message); }\n"})
	diagnostics := checkErrors(t, paths)
	if len(diagnostics) != 1 {
		t.Fatalf("diagnostics: %q", diagnostics)
	}
	got := strings.TrimPrefix(diagnostics[0], filepath.Dir(paths[0])+"/")
	want := "main.a:2:40: error TS18046: 'e' is of type 'unknown'."
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}
