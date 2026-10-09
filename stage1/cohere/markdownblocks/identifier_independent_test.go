package markdownblocks

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"unicode"
)

func TestMdastIdentifierScalars(t *testing.T) {
	t.Parallel()
	products := identifierProductsReady(t)
	configureMarkdownMemory(t)
	markdownMemory.acquire(4)
	defer markdownMemory.release(4)
	ctx, finish := malformedEventsDeadline(t, t.Name())
	defer finish()
	goBinary, main, binary := products.goBinary, products.main, products.sanitized
	truth := malformedEventsExecute(t, ctx, nil, goBinary)
	clean(t, "actual Go NormalizeIdentifier", truth)
	answer := malformedEventsExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=0"}, binary)
	for _, side := range []run{answer, malformedEventsNode(t, ctx, main), malformedEventsNode(t, ctx, products.backend)} {
		clean(t, "identifier scalar oracle", side)
		equal(t, "identifier scalar oracle", side.stdout, truth.stdout)
	}
	if runtime.GOOS == "darwin" {
		report := malformedEventsExecute(t, ctx, nil, "leaks", "--atExit", "--", products.release)
		if report.exitCode != 0 {
			t.Fatalf("identifier leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
		}
	} else {
		report := malformedEventsExecute(t, ctx, []string{"ASAN_OPTIONS=detect_leaks=1"}, binary)
		if report.exitCode != 0 {
			t.Fatalf("identifier leaks: exit %d\n%s\n%s", report.exitCode, report.stdout, report.stderr)
		}
	}
	scratch := t.TempDir()
	if e := os.Mkdir(filepath.Join(scratch, "testdata"), 0755); e != nil {
		t.Fatal(e)
	}
	for _, file := range []string{"identifier.ts", "identifierCaseKeys.ts", "identifierCaseValues.ts", "testdata/identifier_probe.ts"} {
		data, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if file == "identifierCaseKeys.ts" {
			if !strings.Contains(string(data), "65,") {
				t.Fatal("case table mutation anchor")
			}
			data = []byte(strings.Replace(string(data), "65,", "64,", 1))
		}
		write(t, filepath.Join(scratch, file), data)
	}
	mutant := malformedEventsNode(t, ctx, filepath.Join(scratch, "testdata/identifier_probe.ts"))
	clean(t, "case table mutant", mutant)
	if bytes.Equal(mutant.stdout, truth.stdout) {
		t.Fatal("case table mutant survived")
	}
	t.Logf("all1112064 Unicode scalars match actual Go normalization, native/source/backend and sanitizer/leaks; output-only case-table mutant caught at byte%d; Go Unicode%s", firstDifference(string(mutant.stdout), string(truth.stdout)), unicode.Version)
}
