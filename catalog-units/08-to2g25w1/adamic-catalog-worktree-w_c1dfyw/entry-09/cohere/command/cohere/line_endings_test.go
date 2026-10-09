package main

import (
	"bytes"
	"strings"
	"sync"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// A file handed to the formatter with CRLF is counted once, however many passes it takes, and the note
// names the count and the .gitattributes line. A tree with none prints nothing.
func TestCRLFFilesAreNamedOnceWithTheFix(t *testing.T) {
	t.Parallel()
	var lineEndings crlfFiles
	transform := lineEndings.observing(func(fileName string, text string, _ *ast.SourceFile) (string, error) {
		return strings.ReplaceAll(text, "\r\n", "\n"), nil
	})

	var group sync.WaitGroup
	for _, file := range []struct{ name, text string }{
		{"a.ts", "export const a = 1;\r\n"},
		{"a.ts", "export const a = 1;\n"},
		{"b.ts", "export const b = 2;\r\nexport const c = 3;\r\n"},
		{"lf.ts", "export const lf = 4;\n"},
	} {
		group.Add(1)
		go func() {
			defer group.Done()
			if _, err := transform(file.name, file.text, nil); err != nil {
				t.Error(err)
			}
		}()
	}
	group.Wait()

	var checked bytes.Buffer
	lineEndings.report(&checked, false)
	if got := checked.String(); !strings.HasPrefix(got, "note: 2 files have CRLF line endings, and the house format is LF, so --no-fix reports them as unformatted.") ||
		!strings.Contains(got, "add `* text=auto eol=lf` to the repository's .gitattributes") {
		t.Fatalf("the note reads %q", got)
	}

	var written bytes.Buffer
	lineEndings.report(&written, true)
	if !strings.Contains(written.String(), "so --fix rewrote them to LF.") {
		t.Fatalf("a writing run's note reads %q", written.String())
	}

	var none crlfFiles
	var quiet bytes.Buffer
	none.report(&quiet, false)
	if quiet.Len() != 0 {
		t.Fatalf("a tree with no CRLF printed %q", quiet.String())
	}
}

// A CRLF file's would-change finding carries the cause and the .gitattributes line, since the note above
// prints only under --verbose; a file `--fix` would change for any other reason says nothing about line
// endings (#dr78rt8). Both through the binary, so the name the formatter saw and the name the finding
// prints are proven to be the same file.
func TestACRLFFilesWouldChangeFindingSaysWhy(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      "{ \"extends\": \"./NexusCohereSettings.json\", \"rules\": {} }\n",
		"NexusCohereSettings.json": "{ \"format\": { \"ignore\": [] } }\n",
		// Formatted but for its line endings.
		"Windows.ts": "export const windows = 1;\r\nexport const checkout = 2;\r\n",
		// Line endings already LF, and unformatted.
		"Ugly.ts": "export const ugly   =   1\n",
	})

	output, _ := runCohere(t, binary, root, "--no-fix", "--format-all", "--no-cache")
	lines := map[string]string{}
	for _, line := range strings.Split(output, "\n") {
		for _, name := range []string{"Windows.ts", "Ugly.ts"} {
			if strings.Contains(line, name+":1:1 - --fix would rewrite this file") {
				lines[name] = line
			}
		}
	}
	if !strings.Contains(lines["Windows.ts"], "It has CRLF line endings, and the house format is LF") ||
		!strings.Contains(lines["Windows.ts"], "add `* text=auto eol=lf` to the repository's .gitattributes") {
		t.Errorf("the CRLF file's finding does not say why or give the fix: %q\n%s", lines["Windows.ts"], output)
	}
	if lines["Ugly.ts"] == "" {
		t.Fatalf("the unformatted LF file has no would-change finding, so the control proves nothing:\n%s", output)
	}
	// "CRLF line endings" rather than "CRLF": the test's own temporary directory is named after it.
	if strings.Contains(lines["Ugly.ts"], "CRLF line endings") {
		t.Errorf("an LF file's finding names CRLF: %q", lines["Ugly.ts"])
	}
}
