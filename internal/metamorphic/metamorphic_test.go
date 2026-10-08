package metamorphic

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// sample has something for every transform to find: a class (which stays at the top level), a
// function the module's statements call, a constant read twice, a run of statements in a function
// that reads only constants, values handed to calls, and a temporary read once in the next statement.
const sample = `class Counter {
	count = 0;

	add(amount: number): number {
		this.count += amount;
		return this.count;
	}
}

function describe(label: string, value: number): string {
	const shown = label.toUpperCase();
	console.log(String(shown.length + value));
	return shown + '=' + String(value * 2);
}

const counter = new Counter();
const name = 'total';
counter.add(name.length);
counter.add(name.length + 1);
const once = counter.add(3) * 2;
console.log(describe(name, once));
`

func samplePath(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "program.a")
	if err := os.WriteFile(path, []byte(sample), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// Every transform finds a site in the sample, and what it makes still checks and lowers.
func TestEveryTransformRewritesTheSample(t *testing.T) {
	t.Parallel()
	path := samplePath(t)
	for _, transform := range Transforms {
		variant := Apply(path, sample, []Transform{transform})
		if variant.Skipped != "" || variant.Sites == 0 {
			t.Errorf("%s: no variant: %s (refusals %v)", transform, variant.Skipped, variant.Refusals)
			continue
		}
		if variant.Source == sample {
			t.Errorf("%s: the variant is the original", transform)
		}
		if refusal := accepts(path, variant.Source); refusal != "" {
			t.Errorf("%s: the variant is refused: %s\n%s", transform, refusal, variant.Source)
		}
	}
	combined := Apply(path, sample, Combined)
	if combined.Skipped != "" || combined.Sites < len(Combined) {
		t.Errorf("combined: %d sites, %s", combined.Sites, combined.Skipped)
	}
}

// Wrap keeps the class at the top level, moves the function in as a closure, and calls the wrapper
// once.
func TestWrapKeepsClassesAtTheTopLevel(t *testing.T) {
	t.Parallel()
	path := samplePath(t)
	variant := Apply(path, sample, []Transform{Wrap})
	wrapper := strings.Index(variant.Source, "function metamorphicMain(): void {")
	class := strings.Index(variant.Source, "class Counter")
	closure := strings.Index(variant.Source, "const describe = function describe(")
	if wrapper < 0 || class < 0 || class > wrapper || closure < wrapper || !strings.HasSuffix(strings.TrimSpace(variant.Source), "metamorphicMain();") {
		t.Errorf("wrap made:\n%s", variant.Source)
	}
}

// A module that exports is left alone by Wrap.
func TestWrapLeavesAnExportingModuleAlone(t *testing.T) {
	t.Parallel()
	source := "export const value = 1;\nconsole.log(`${value}`);\n"
	path := filepath.Join(t.TempDir(), "module.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	if variant := Apply(path, source, []Transform{Wrap}); !strings.Contains(variant.Skipped, "it exports") {
		t.Errorf("want it skipped for its export, got %q:\n%s", variant.Skipped, variant.Source)
	}
}

// A site the checker refuses is counted and dropped, never kept: an alias of a value the program
// narrows loses the narrowing.
func TestARefusedSiteIsCountedAndDropped(t *testing.T) {
	t.Parallel()
	source := "function pick(flag: boolean): string | undefined {\n\treturn flag ? 'a' : undefined;\n}\nconst found = pick(true);\nif (found !== undefined) {\n\tconsole.log(found);\n\tconsole.log(found);\n}\n"
	path := filepath.Join(t.TempDir(), "narrowed.a")
	if err := os.WriteFile(path, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	variant := Apply(path, source, []Transform{Alias})
	if variant.Refused == 0 || variant.Sites != 0 || !strings.Contains(variant.Skipped, "every site refused") {
		t.Errorf("want the alias refused and dropped, got %d sites, %d refused, %q: %q", variant.Sites, variant.Refused, variant.Skipped, variant.Source)
	}
}

// The fixtures are the oracle's list, sorted, and a stride takes the first and spreads the rest.
func TestFixturesAreTheOraclesList(t *testing.T) {
	t.Parallel()
	fixtures, err := Fixtures(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	if len(fixtures) < 200 || fixtures[0] > fixtures[len(fixtures)-1] {
		t.Fatalf("%d fixtures, %v", len(fixtures), fixtures[:min(3, len(fixtures))])
	}
	chosen := Stride(fixtures, 20)
	if len(chosen) != 20 || chosen[0] != fixtures[0] || chosen[1] != fixtures[len(fixtures)/20] {
		t.Errorf("stride chose %v", chosen)
	}
}

// The leak check reads a counted build's counts: a program whose allocations aren't its frees and
// its values in regions leaked, and one that ends without its counts line didn't finish.
func TestTheLeakCheckReadsTheCounts(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "darwin" {
		// census: not-applicable macOS-only counted-build leak witness; Linux uses LeakSanitizer on the sanitized binary.
		t.Skip("Not on macOS: Linux's leak check is LeakSanitizer's, run on the sanitized binary")
	}
	directory := t.TempDir()
	for _, test := range []struct {
		name, script, want string
	}{
		{"unbalanced", "echo 'adamic: counts: allocations 5 frees 3 retains 0 releases 0 peak 2 regions 1' >&2\n", "heap values leaked: 1 (allocations 5, frees 3, in regions 1)"},
		{"no counts", "echo 'out'\n", "the counted build didn't finish with its counts"},
	} {
		counted := filepath.Join(directory, test.name)
		if err := os.WriteFile(counted, []byte("#!/bin/sh\n"+test.script), 0o755); err != nil {
			t.Fatal(err)
		}
		if report := (Checkout{}).leaks(directory, "", counted); !strings.HasPrefix(report, test.want) {
			t.Errorf("%s: got %q, want %q", test.name, report, test.want)
		}
	}
}

// Two outputs are judged by their digests only when they're over a megabyte, and the digests are
// equal exactly when the bytes are.
func TestDigestsTellOutputsApart(t *testing.T) {
	t.Parallel()
	large := strings.Repeat("x", 1<<20+1)
	if string(digest([]byte(large))) != string(digest([]byte(large))) || string(digest([]byte(large))) == string(digest([]byte(large[1:]+"y"))) {
		t.Error("digests don't tell outputs apart")
	}
}
