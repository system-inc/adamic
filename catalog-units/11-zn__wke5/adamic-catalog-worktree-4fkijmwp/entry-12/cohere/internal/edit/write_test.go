package edit

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// Atomicity is the one guard whose absence is invisible in a single-threaded test: writing in place
// produces the same final bytes as temp-plus-rename, so every assertion about the result passes
// either way. The difference only appears to a concurrent reader, which is exactly the situation
// this repo is in — several agents editing one tree, where a half-written file failing to build and
// then healing two minutes later was observed rather than imagined.
//
// So the test reads the file continuously while it is rewritten, and asserts that every read saw a
// complete version. An in-place write fails this: os.WriteFile truncates and then fills, so a
// reader landing in that window sees a prefix, or nothing.
//
// Confirmed to fail: replacing WriteAtomically's body with os.WriteFile makes this test report
// torn reads. That mutation passes every other test in this package, which is why this one exists.
func TestConcurrentReadersNeverSeeAPartialFile(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	fileName := filepath.Join(directory, "contended.ts")

	// Large enough that a non-atomic write cannot complete between two reads. A small file is
	// written in one syscall on most systems and would hide the defect.
	oldBody := strings.Repeat("const alpha = 'old';\n", 4000)
	newBody := strings.Repeat("const alpha = 'new';\n", 4000)

	if err := os.WriteFile(fileName, []byte(oldBody), 0o644); err != nil {
		t.Fatal(err)
	}

	stop := make(chan struct{})
	var waitGroup sync.WaitGroup
	var mutex sync.Mutex
	torn := []string{}

	// Several readers, because one reader polling is unlikely to land inside a narrow window.
	for reader := range 4 {
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()
			for {
				select {
				case <-stop:
					return
				default:
				}
				contents, err := os.ReadFile(fileName)
				if err != nil {
					mutex.Lock()
					torn = append(torn, "reader "+string(rune('a'+reader))+": "+err.Error())
					mutex.Unlock()
					continue
				}
				text := string(contents)
				// A complete file is exactly one of the two versions. Anything else — a prefix, a
				// mixture, an empty file — is a partial write reaching a reader.
				if text != oldBody && text != newBody {
					mutex.Lock()
					torn = append(torn, describeTear(text, len(oldBody)))
					mutex.Unlock()
				}
			}
		}()
	}

	// Rewrite repeatedly while the readers run, alternating so both directions are exercised.
	for round := range 40 {
		body := newBody
		if round%2 == 1 {
			body = oldBody
		}
		if err := WriteAtomically(fileName, body); err != nil {
			close(stop)
			waitGroup.Wait()
			t.Fatalf("write failed on round %d: %v", round, err)
		}
	}

	close(stop)
	waitGroup.Wait()

	if len(torn) > 0 {
		shown := torn
		if len(shown) > 5 {
			shown = shown[:5]
		}
		t.Fatalf("%d readers saw a partial file (writes are not atomic):\n  %s",
			len(torn), strings.Join(shown, "\n  "))
	}
}

// describeTear summarizes a partial read without dumping eighty kilobytes into the failure output.
func describeTear(text string, expectedLength int) string {
	switch {
	case len(text) == 0:
		return "read an empty file"
	case len(text) < expectedLength:
		return "read a truncated file: " + itoa(len(text)) + " of " + itoa(expectedLength) + " bytes"
	default:
		return "read a file of unexpected length: " + itoa(len(text))
	}
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	digits := []byte{}
	for value > 0 {
		digits = append([]byte{byte('0' + value%10)}, digits...)
		value /= 10
	}
	return string(digits)
}

// The temp file must be created beside the target rather than in the system temp directory. A
// rename across filesystems is not atomic — it degrades to copy-and-delete — so a temp file on
// another device silently removes the guarantee this function exists to provide.
func TestTemporaryFileIsCreatedBesideTheTarget(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	nested := filepath.Join(directory, "nested")
	if err := os.Mkdir(nested, 0o755); err != nil {
		t.Fatal(err)
	}
	fileName := filepath.Join(nested, "target.ts")

	if err := os.WriteFile(fileName, []byte("const a = 1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Watch the directory during the write by racing a reader that records what it sees.
	seen := map[string]bool{}
	stop := make(chan struct{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			select {
			case <-stop:
				return
			default:
			}
			entries, err := os.ReadDir(nested)
			if err != nil {
				continue
			}
			for _, entry := range entries {
				seen[entry.Name()] = true
			}
		}
	}()

	for range 200 {
		if err := WriteAtomically(fileName, "const a = 2;\n"); err != nil {
			close(stop)
			<-done
			t.Fatal(err)
		}
		if err := WriteAtomically(fileName, "const a = 1;\n"); err != nil {
			close(stop)
			<-done
			t.Fatal(err)
		}
	}
	close(stop)
	<-done

	sawTemporary := false
	for name := range seen {
		if strings.HasPrefix(name, ".target.ts.cohere-") {
			sawTemporary = true
		}
	}
	if !sawTemporary {
		t.Fatalf("no temp file was ever observed beside the target; writes may not be going through temp+rename (saw: %v)", keysOf(seen))
	}

	// And none may survive.
	entries, err := os.ReadDir(nested)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() != "target.ts" {
			t.Fatalf("a temp file survived the write: %s", entry.Name())
		}
	}
}

func keysOf(set map[string]bool) []string {
	names := make([]string, 0, len(set))
	for name := range set {
		names = append(names, name)
	}
	return names
}

// A write into a directory that does not exist must fail loudly rather than silently doing nothing.
func TestWriteIntoAMissingDirectoryFails(t *testing.T) {
	t.Parallel()
	missing := filepath.Join(t.TempDir(), "absent", "target.ts")
	if err := WriteAtomically(missing, "const a = 1;\n"); err == nil {
		t.Fatalf("expected an error writing into a directory that does not exist")
	}
}

// The parse guard must only judge files the TypeScript parser owns.
//
// The guard is a TypeScript parser, so it answers a question about TypeScript. Pointed at css, json
// or markdown it reports TS1128 and TS1434 and refuses the file, which is a correct answer to a
// question nobody asked: those files are not malformed, they are not TypeScript.
//
// This was live rather than theoretical. When the format phase's universe became the working tree
// rather than the type graph, five files in a six-file fixture came back "could not be processed"
// with TypeScript syntax errors against css, json and markdown. Before that split every file
// reaching this package was TypeScript by construction, so the guard's scope was adequate by
// accident, in exactly the way the type graph was an adequate format universe by accident.
func TestTheParseGuardOnlyJudgesTypeScript(t *testing.T) {
	t.Parallel()
	// Real content in each language, none of which is valid TypeScript.
	for _, testCase := range []struct{ fileName, text string }{
		{"styles.css", ".foo{color:red;background:blue}\n"},
		{"data.json", "{\"b\":2, \"a\":1}\n"},
		{"notes.md", "# heading\n\nsome text\n"},
		{"config.yaml", "key: value\nlist:\n  - one\n"},
	} {
		if parses, reason := Parses(testCase.fileName, testCase.text); !parses {
			t.Fatalf("%s was refused by a TypeScript parser: %s", testCase.fileName, reason)
		}
	}
}

// And the guard must still fire on the languages it does own, or scoping it would have disabled it.
//
// This is the half that makes the change safe rather than merely permissive: a broken .ts is still
// refused, so the discipline that lets autofix write to source at all is intact.
func TestTheParseGuardStillFiresOnTypeScript(t *testing.T) {
	t.Parallel()
	for _, fileName := range []string{"a.ts", "b.tsx", "c.js", "d.jsx", "e.mjs", "f.cjs", "g.mts", "h.cts"} {
		if parses, _ := Parses(fileName, "export function alpha( {\n"); parses {
			t.Fatalf("%s: a broken file was accepted, so the guard is off for a language it owns", fileName)
		}
	}
}

// TypeScriptParsable must discriminate, not merely answer. A predicate that said yes to everything
// would restore the bug; one that said no to everything would disable the guard entirely.
func TestTypeScriptParsableDiscriminates(t *testing.T) {
	t.Parallel()
	for _, fileName := range []string{"a.ts", "b.tsx", "c.js", "d.mjs"} {
		if !TypeScriptParsable(fileName) {
			t.Fatalf("%s is TypeScript-family and was not recognized", fileName)
		}
	}
	for _, fileName := range []string{"a.css", "b.json", "c.md", "d.yaml", "Makefile", "e.rb"} {
		if TypeScriptParsable(fileName) {
			t.Fatalf("%s is not TypeScript-family and was claimed", fileName)
		}
	}
}

// An Adamic `.a` file is TypeScript to the guard (#6mhafvb): it is parsable, valid TypeScript in it parses, and
// a broken one is refused with a reason, which is the refusal that lets autofix write a `.a` at all. Dropping
// sourcename.TreatedAs from TypeScriptParsable passes the broken text through unguarded, and this fails.
func TestTheParseGuardJudgesAnAdamicFileAsTypeScript(t *testing.T) {
	t.Parallel()
	if !TypeScriptParsable("/repo/geometry.a") {
		t.Fatal("an Adamic .a file is not TypeScript to the guard")
	}
	if parses, reason := Parses("/repo/geometry.a", "const widened = <number>value;\nexport const area = (width: number): number => width;\n"); !parses {
		t.Fatalf("valid TypeScript in a .a file was refused: %s", reason)
	}
	parses, reason := Parses("/repo/geometry.a", "export function alpha( {\n")
	if parses || reason == "" {
		t.Fatalf("a broken .a file was accepted (reason %q), so the guard is off for it", reason)
	}
}
