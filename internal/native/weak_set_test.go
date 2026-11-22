package native_test

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// weakSetFreedKeys adds a key made in a call to a WeakSet that outlives it, ten times: each key is freed
// when its call returns, while the set still holds an entry for it, so the runtime's forget must take the
// entry out (weak.c) and nothing is left for the set to let go of at the end.
const weakSetFreedKeys = `const seen = new WeakSet<{ round: number }>();
function visit(round: number): boolean {
	const key = { round };
	seen.add(key);
	return seen.has(key);
}
let held = 0;
for (let round = 0; round < 10; round++) {
	if (visit(round)) held++;
}
console.log(String(held));
`

// strongKey is the mutant that keeps the key alive: add retains the key itself as well as its handle,
// which no release ever balances.
var strongKey = regexp.MustCompile(`adamic_weak_map_store\(([^;]*?), adamic_weak_key\(([^;]*?)\), \(adamic_value\)\{0\}\);`)

// A WeakSet keyed by objects a call makes and frees, lowered from the program, under ASan (with the
// size-class allocator on and off) and the counted build: each key's entry goes when the key is freed,
// so the counted build balances its allocations and frees and its weak entries line says none is live at
// exit and one at most at once, and ASan sees no use after free. Mutant: add retains the key strongly,
// red here as an unbalanced count, the ten keys never freed (#hcgy751).
func TestWeakSetFreedKeyLeavesNoEntry(t *testing.T) {
	t.Parallel()
	code := weakSetC(t, weakSetFreedKeys)
	for _, build := range []struct {
		name    string
		options native_runtime.Options
	}{
		{"sanitized", native_runtime.Options{Sanitize: true}},
		{"slabs", native_runtime.Options{Sanitize: true, Slabs: true}},
		{"counted", native_runtime.Options{Sanitize: true, Count: true}},
		{"plain counted", native_runtime.Options{Count: true}},
	} {
		t.Run(build.name, func(t *testing.T) {
			t.Parallel()
			result := weakSetRun(t, code, build.options)
			if result.ExitCode != 0 || string(result.Stdout) != "10\n" {
				t.Fatalf("exit %d, stdout %q, want 10\n%s", result.ExitCode, result.Stdout, result.Stderr)
			}
			if !build.options.Count {
				return
			}
			if report := leakcheck.Unbalanced(result); report != "" {
				t.Fatal(report)
			}
			if match := weakEntriesLine.FindSubmatch(result.Stderr); match == nil || string(match[1]) != "0" || string(match[2]) != "1" {
				t.Fatalf("want the weak entries line to say none live and a peak of one\n%s", result.Stderr)
			}
		})
	}
	t.Run("strong key mutant", func(t *testing.T) {
		t.Parallel()
		if len(strongKey.FindAllStringIndex(code, -1)) != 1 {
			t.Fatalf("the mutant's pattern matches %d stores, want 1", len(strongKey.FindAllStringIndex(code, -1)))
		}
		mutant := strongKey.ReplaceAllString(code, "adamic_retain($2); adamic_weak_map_store($1, adamic_weak_key($2), (adamic_value){0});")
		result := weakSetRun(t, mutant, native_runtime.Options{Count: true})
		// The keys outlive their calls, so every entry lives until the set goes at exit (a peak of ten), and
		// the keys the mutant retained are never freed: the counted build's balance is what goes red.
		report := leakcheck.Unbalanced(result)
		if report == "" {
			t.Fatalf("the strong key mutant survived: the counted build balanced\n%s", result.Stderr)
		}
		t.Logf("mutant caught: %s (%s)", report, weakEntriesLine.Find(result.Stderr))
	})
}

// weakSetC is the C a program lowers to.
func weakSetC(t *testing.T, source string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
	if err != nil {
		t.Fatal(err)
	}
	return native.C(program)
}

// weakSetRun builds code with options and runs it.
func weakSetRun(t *testing.T, code string, options native_runtime.Options) leakcheck.Run {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(code, binary, options); err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=0", "UBSAN_OPTIONS=halt_on_error=1")
	var stdout, stderr strings.Builder
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	exit := 0
	if err != nil {
		exit = -1
		if command.ProcessState != nil {
			exit = command.ProcessState.ExitCode()
		}
	}
	return leakcheck.Run{Stdout: []byte(stdout.String()), Stderr: []byte(stderr.String()), ExitCode: exit}
}
