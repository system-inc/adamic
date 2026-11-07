package compilerinputs

import (
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// CheckKey passes a key that names Lowering and the product's own files, and fails the two ways a key goes wrong:
// declaring tsc whole (a baseline edit moves it) and declaring no compiler (a linked package's edit keeps it).
func TestCheckKeyTellsTheCompilerFromTscWhole(test *testing.T) {
	test.Parallel()
	own := []string{"stage1/stand-in", "stage1/stand-in/port.ts"}
	keyed := func(files []string) func(string) (string, error) {
		return func(root string) (string, error) {
			return buildcache.Key(root, buildcache.Inputs{Name: "stand-in", Files: files})
		}
	}
	if checkKeyError := CheckKey(own, keyed(append(Lowering(), own...))); checkKeyError != nil {
		test.Fatalf("a key naming Lowering: %v", checkKeyError)
	}
	for _, wrong := range []struct {
		name  string
		files []string
		want  string
	}{
		{"tsc whole", append(append(Lowering(), own...), "cohere/TypeScript/tsc"), "names tsc whole"},
		{"no compiler", own, "doesn't name the compiler"},
	} {
		checkKeyError := CheckKey(own, keyed(wrong.files))
		if checkKeyError == nil || !strings.Contains(checkKeyError.Error(), wrong.want) {
			test.Errorf("a key declaring %s: got %v, want an error saying it %s", wrong.name, checkKeyError, wrong.want)
		}
	}
}
