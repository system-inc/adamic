package compilerinputs

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// CheckKey passes a key that names Lowering and the product's own files, and fails the two ways a key goes wrong:
// declaring tsc whole (a baseline edit moves it) and declaring no compiler (a linked package's edit keeps it).
func TestCheckKeyTellsTheCompilerFromTscWhole(t *testing.T) {
	t.Parallel()
	own := []string{"stage1/stand-in", "stage1/stand-in/port.ts"}
	keyed := func(files []string) func(string) (string, error) {
		return func(root string) (string, error) {
			return buildcache.Key(root, buildcache.Inputs{Name: "stand-in", Files: files})
		}
	}
	if err := CheckKey(own, keyed(append(Lowering(), own...))); err != nil {
		t.Fatalf("a key naming Lowering: %v", err)
	}
	for _, wrong := range []struct {
		name  string
		files []string
		want  string
	}{
		{"tsc whole", append(append(Lowering(), own...), "cohere/TypeScript/tsc"), "names tsc whole"},
		{"no compiler", own, "doesn't name the compiler"},
	} {
		err := CheckKey(own, keyed(wrong.files))
		if err == nil || !strings.Contains(err.Error(), wrong.want) {
			t.Errorf("a key declaring %s: got %v, want an error saying it %s", wrong.name, err, wrong.want)
		}
	}
}
