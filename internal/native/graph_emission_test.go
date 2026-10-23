package native_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// These are compiler ownership witnesses, not the later graph corpus. Each
// control runs source Node first, both backends, sanitizers and the leak check.
func graphCompilerWitness(t *testing.T, source string, mutant func(string) string) *ir.Program {
	t.Helper()
	directory := t.TempDir()
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	run := func(env []string, name string, args ...string) ([]byte, []byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, name, args...)
		command.Env = append(os.Environ(), env...)
		var out, diagnostic bytes.Buffer
		command.Stdout = &out
		command.Stderr = &diagnostic
		err := command.Run()
		return out.Bytes(), diagnostic.Bytes(), err
	}
	truth, diagnostic, err := run(nil, "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path)
	if err != nil {
		t.Fatalf("source Node: %v\n%s", err, diagnostic)
	}
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), checked)
	if err != nil {
		t.Fatal(err)
	}
	js := filepath.Join(directory, "main.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	got, diagnostic, err := run(nil, "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js)
	if err != nil || !bytes.Equal(got, truth) {
		t.Fatalf("JavaScript: %v\n%s\n%s, Node %s", err, diagnostic, got, truth)
	}
	code := native.C(program)
	if mutant != nil {
		changed := mutant(code)
		if changed == code {
			t.Fatal("mutant did not change C")
		}
		code = changed
	}
	binary := filepath.Join(directory, "native")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got, diagnostic, err = run([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if mutant != nil {
		if err == nil && bytes.Equal(got, truth) {
			if report := leakcheck.Report(t, code, binary); report == "" {
				t.Fatal("graph ownership mutant survived")
			}
		} else if err != nil && !strings.Contains(string(diagnostic), "AddressSanitizer") && !strings.Contains(string(diagnostic), "runtime error:") && !strings.Contains(string(diagnostic), "LeakSanitizer") && !strings.Contains(string(diagnostic), "compiler bug: graph edge crosses regions") && !strings.Contains(string(diagnostic), "compiler bug: graph adoption truncates object storage") {
			t.Fatalf("mutant failed outside sanitizer: %v\n%s", err, diagnostic)
		}
		return program
	}
	if err != nil || !bytes.Equal(got, truth) {
		t.Fatalf("native: %v\n%s\n%s, Node %s", err, diagnostic, got, truth)
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
	if err := native.Build(code, binary, native.Options{Count: true}); err != nil {
		t.Fatal(err)
	}
	got, diagnostic, err = run(nil, binary)
	if err != nil || !bytes.Equal(got, truth) {
		t.Fatalf("counted: %v\n%s\n%s", err, got, diagnostic)
	}
	if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: got, Stderr: diagnostic}); report != "" {
		t.Fatal(report)
	}
	t.Logf("%s", diagnostic)
	return program
}

const graphLinkSource = `interface Link { name: string; next: Link | undefined }
function make(name: string): Link { return {name, next: undefined}; }
function run(): void { const a=make('a'+1); const b=make('b'+2); a.next=b; b.next=a; console.log(a.next.name+b.next.name); }
run();`

func TestGraphCompilerLinks(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphLinkSource, nil)
}
func TestGraphCompilerAdoptionMutant(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphLinkSource, func(code string) string {
		return regexp.MustCompile(`adamic_graph_adopt_owned\(([^,]+), [^;]+\)`).ReplaceAllString(code, `$1`)
	})
}
func TestGraphCompilerSlotHoldMutant(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphLinkSource, func(code string) string {
		return regexp.MustCompile(`adamic_graph_hold\([^,]+, ([^)]+)\)`).ReplaceAllString(code, `adamic_retain($1)`)
	})
}

const graphCanonicalSource = `function make(): () => boolean {
 let saved: (() => boolean) | undefined = undefined;
 function f(): boolean { return saved === f; }
 saved = f;
 return f;
}
console.log(String(make()()));`

func TestGraphCompilerCanonical(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphCanonicalSource, nil)
}
func TestGraphCompilerCanonicalCacheMutant(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphCanonicalSource, func(code string) string {
		return strings.ReplaceAll(code, "adamic_closure_canonical_graph(", "adamic_closure_canonical(")
	})
}
func TestGraphCompilerCanonicalCounted(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, strings.ReplaceAll(graphCanonicalSource, "return saved === f;", "return arguments.length === 0 && saved === f;"), nil)
}

const graphCollectionSource = `interface Link { children: Link[]; map: Map<string,Link>; set: Set<Link> }
function run(): void {
 const children: Link[]=[]; const map=new Map<string,Link>(); const set=new Set<Link>();
 const link: Link={children,map,set}; children.push(link); map.set('a',link); set.add(link);
 const copied=children.map(x=>x); const filtered=copied.filter(x=>true); const sliced=filtered.slice();
 const from=Array.from({length:2},()=>link); const filled=new Array<Link>(2).fill(link);
 console.log(String(sliced[0]===link)+String(map.get('a')===link)+String(set.has(link))+String(from[0]===filled[1]));
} run();`

func TestGraphCompilerCollections(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphCollectionSource, nil)
}
func TestGraphCompilerArrayTakeMutant(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphCollectionSource, func(code string) string {
		return regexp.MustCompile(`adamic_graph_take\([^,]+, ([^)]+)\)`).ReplaceAllString(code, `$1`)
	})
}

func TestGraphCompilerObjectBytesMutant(t *testing.T) {
	t.Parallel()
	graphCompilerWitness(t, graphLinkSource, func(code string) string {
		return regexp.MustCompile(`adamic_object_size\([^)]*\)`).ReplaceAllString(code, `sizeof(adamic_object)`)
	})
}

// Break the links before scope exit so both the counted control and the graph
// program finish balanced. Their cost difference measures interior edge counts.
func TestGraphCompilerLinkCountsDecrease(t *testing.T) {
	t.Parallel()
	program := graphCompilerWitness(t, strings.Replace(graphLinkSource, "console.log(a.next.name+b.next.name);", "console.log(a.next.name+b.next.name); a.next=undefined; b.next=undefined;", 1), nil)
	counts := func(code string) []string {
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(code, binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		output, err := exec.CommandContext(ctx, binary).CombinedOutput()
		if err != nil {
			t.Fatalf("counted: %v\n%s", err, output)
		}
		match := leakcheck.CountsLine.FindStringSubmatch(string(output))
		if match == nil {
			t.Fatalf("missing counts: %s", output)
		}
		return match[1:]
	}
	after := counts(native.C(program))
	program.GraphTypes = nil
	for i := range program.Functions {
		program.Functions[i].GraphClosure = false
	}
	for i := range program.Locals {
		program.Locals[i].GraphCell = false
	}
	for i := range program.Classes {
		program.Classes[i].Graph = false
	}
	before := counts(native.C(program))
	t.Logf("Link witness A/F/R/L/P/G: counted %s; graph %s", strings.Join(before, "/"), strings.Join(after, "/"))
	for _, field := range []int{2, 3} {
		old, err := strconv.Atoi(before[field])
		if err != nil {
			t.Fatal(err)
		}
		now, err := strconv.Atoi(after[field])
		if err != nil {
			t.Fatal(err)
		}
		if now >= old {
			t.Fatalf("graph ownership did not reduce count %d: %d -> %d", field, old, now)
		}
	}
}
