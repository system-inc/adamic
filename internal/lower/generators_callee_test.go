package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// A generator called other than by its name, through a namespace import's member or a conditional
// callee ((0, gen) is a TypeScript error), routes by its origin like a call by name, so a declared
// SetIterator never sends it down the collection iterator's path, and stops with its own reason:
// its callee could run code first (Compiler's seam read of S1c on #qk8rztp). Read by testing, the
// road was already closed before this: each probe stopped not yet at the call guard or as a value.
// These pins hold it closed by origin. Mutant: generatorCallee reads only a bare name, and each probe
// stops for another reason, so the test goes red.
func TestGeneratorCalledOtherThanByItsNameIsNotYet(t *testing.T) {
	t.Parallel()
	library := "export function* letters(word: string): SetIterator<string> { for (let index = 0; index < word.length; index++) { yield word.charAt(index); } }\n"
	for _, probe := range []struct{ name, main string }{
		{"namespace for-of", "import * as library from './library';\nfor (const letter of library.letters('ab')) console.log(letter);\n"},
		{"namespace spread", "import * as library from './library';\nconsole.log([...library.letters('ab')].join(','));\n"},
		{"namespace Array.from", "import * as library from './library';\nconsole.log(Array.from(library.letters('ab')).join(','));\n"},
		{"conditional callee for-of", "import { letters } from './library';\nconst pick: boolean = 'ab'.length > 1;\nfor (const letter of (pick ? letters : letters)('ab')) console.log(letter);\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			for name, source := range map[string]string{"library.ts": library, "main.ts": probe.main} {
				if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
					t.Fatal(err)
				}
			}
			program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			_, err = Lower(context.Background(), program)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, "called other than by its name") {
				t.Fatalf("got %v, want the generator's called-other-than-by-its-name stop", err)
			}
		})
	}
}
