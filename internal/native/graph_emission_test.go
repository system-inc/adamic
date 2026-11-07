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
func graphCompilerWitness(test *testing.T, source string, mutant func(string) string) *ir.Program {
	test.Helper()
	return graphCompilerModulesWitness(test, source, nil, mutant)
}

// graphCompilerModulesWitness is graphCompilerWitness of a main.a that imports modules, each written
// beside it under its name, so './name' reaches it: load follows the imports, and the program lowers
// as one.
func graphCompilerModulesWitness(
	test *testing.T,
	source string,
	modules map[string]string,
	mutant func(string) string,
) *ir.Program {
	test.Helper()
	directory := test.TempDir()
	path := filepath.Join(directory, "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	for name, module := range modules {
		if writeError := os.WriteFile(filepath.Join(directory, name), []byte(module), 0600); writeError != nil {
			test.Fatal(writeError)
		}
	}
	run := func(env []string, name string, args ...string) ([]byte, []byte, error) {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, name, args...)
		command.Env = append(os.Environ(), env...)
		var out, diagnostic bytes.Buffer
		command.Stdout = &out
		command.Stderr = &diagnostic
		runError := command.Run()
		return out.Bytes(), diagnostic.Bytes(), runError
	}
	truth, diagnostic, runError := run(
		nil,
		"node",
		"--disable-warning=ExperimentalWarning",
		"../../oracle/node.mjs",
		path,
	)
	if runError != nil {
		test.Fatalf("source Node: %v\n%s", runError, diagnostic)
	}
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	// This witness targets batch 8 ownership even when the process opts into Program mode.
	program, lowerWithOptionsError := lower.LowerWithOptions(context.Background(), checked, lower.Options{})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	js := filepath.Join(directory, "main.mjs")
	if writeError := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	got, diagnostic, runError := run(nil, "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", js)
	if runError != nil || !bytes.Equal(got, truth) {
		test.Fatalf("JavaScript: %v\n%s\n%s, Node %s", runError, diagnostic, got, truth)
	}
	code := native.C(program)
	if mutant != nil {
		changed := mutant(code)
		if changed == code {
			test.Fatal("mutant did not change C")
		}
		code = changed
	}
	binary := filepath.Join(directory, "native")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got, diagnostic, runError = run([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if mutant != nil {
		if runError == nil && bytes.Equal(got, truth) {
			if report := leakcheck.Report(test, code, binary); report == "" {
				// A sanitizer may report leaks with exit status zero, or have leak
				// detection disabled. Check ownership counts independently.
				if buildError := native.Build(code, binary, native.Options{Count: true}); buildError != nil {
					test.Fatal(buildError)
				}
				got, diagnostic, runError = run(nil, binary)
				counts := leakcheck.CountsLine.FindSubmatch(diagnostic)
				if runError != nil || !bytes.Equal(got, truth) || counts == nil {
					test.Fatalf("counted mutant: %v\n%s\n%s, Node %s", runError, diagnostic, got, truth)
				}
				// Counts are emitted as canonical decimal integers. Unlike the
				// general leak check, this witness permits no surviving region.
				if bytes.Equal(counts[1], counts[2]) && string(counts[6]) == "0" {
					test.Fatal("graph ownership mutant survived")
				}
				test.Logf("graph mutant ownership count check caught mutant: %s", counts[0])
			} else {
				test.Logf("leak check caught mutant: %s", report)
			}
		} else if runError != nil &&
			!strings.Contains(string(diagnostic), "AddressSanitizer") &&
			!strings.Contains(string(diagnostic), "runtime error:") &&
			!strings.Contains(string(diagnostic), "LeakSanitizer") &&
			!strings.Contains(string(diagnostic), "compiler bug: graph edge crosses regions") &&
			!strings.Contains(string(diagnostic), "compiler bug: graph adoption truncates object storage") {
			test.Fatalf("mutant failed outside sanitizer: %v\n%s", runError, diagnostic)
		}
		return program
	}
	if runError != nil || !bytes.Equal(got, truth) {
		test.Fatalf("native: %v\n%s\n%s, Node %s", runError, diagnostic, got, truth)
	}
	if report := leakcheck.Report(test, code, binary); report != "" {
		test.Fatal(report)
	}
	if buildError := native.Build(code, binary, native.Options{Count: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got, diagnostic, runError = run(nil, binary)
	if runError != nil || !bytes.Equal(got, truth) {
		test.Fatalf("counted: %v\n%s\n%s", runError, got, diagnostic)
	}
	if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: got, Stderr: diagnostic}); report != "" {
		test.Fatal(report)
	}
	test.Logf("%s", diagnostic)
	return program
}

const graphLinkSource = `interface Link { name: string; next: Link | undefined }
function make(name: string): Link { return {name, next: undefined}; }
function run(): void { const a=make('a'+1); const b=make('b'+2); a.next=b; b.next=a; console.log(a.next.name+b.next.name); }
run();`

func TestGraphCompilerLinks(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphLinkSource, nil)
}
func TestGraphCompilerAdoptionMutant(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphLinkSource, func(code string) string {
		return regexp.MustCompile(`adamic_graph_adopt_owned\(([^,]+), [^;]+\)`).ReplaceAllString(code, `$1`)
	})
}
func TestGraphCompilerSlotHoldMutant(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphLinkSource, func(code string) string {
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

func TestGraphCompilerCanonical(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphCanonicalSource, nil)
}
func TestGraphCompilerCanonicalCacheMutant(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphCanonicalSource, func(code string) string {
		return strings.ReplaceAll(code, "adamic_closure_canonical_graph(", "adamic_closure_canonical(")
	})
}
func TestGraphCompilerCanonicalCounted(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, strings.ReplaceAll(
		graphCanonicalSource,
		"return saved === f;",
		"return arguments.length === 0 && saved === f;",
	), nil)
}

const graphCollectionSource = `interface Link { children: Link[]; map: Map<string,Link>; set: Set<Link> }
function run(): void {
 const children: Link[]=[]; const map=new Map<string,Link>(); const set=new Set<Link>();
 const link: Link={children,map,set}; children.push(link); map.set('a',link); set.add(link);
 const copied=children.map(x=>x); const filtered=copied.filter(x=>true); const sliced=filtered.slice();
 const from=Array.from({length:2},()=>link); const filled=new Array<Link>(2).fill(link);
 console.log(String(sliced[0]===link)+String(map.get('a')===link)+String(set.has(link))+String(from[0]===filled[1]));
} run();`

func TestGraphCompilerCollections(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphCollectionSource, nil)
}
func TestGraphCompilerArrayTakeMutant(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphCollectionSource, func(code string) string {
		return regexp.MustCompile(`adamic_graph_take\([^,]+, ([^)]+)\)`).ReplaceAllString(code, `$1`)
	})
}

func TestGraphCompilerObjectBytesMutant(test *testing.T) {
	test.Parallel()
	graphCompilerWitness(test, graphLinkSource, func(code string) string {
		return regexp.MustCompile(`adamic_object_allocation_size\([^)]*\)`).ReplaceAllString(
			code,
			`sizeof(adamic_object)`,
		)
	})
}

// Break the links before scope exit so both the counted control and the graph
// program finish balanced. Their cost difference measures interior edge counts.
func TestGraphCompilerLinkCountsDecrease(test *testing.T) {
	test.Parallel()
	program := graphCompilerWitness(test, strings.Replace(
		graphLinkSource,
		"console.log(a.next.name+b.next.name);",
		"console.log(a.next.name+b.next.name); a.next=undefined; b.next=undefined;",
		1,
	), nil)
	counts := func(code string) []string {
		binary := filepath.Join(test.TempDir(), "counted")
		if buildError := native.Build(code, binary, native.Options{Count: true}); buildError != nil {
			test.Fatal(buildError)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		output, combinedOutputError := exec.CommandContext(ctx, binary).CombinedOutput()
		if combinedOutputError != nil {
			test.Fatalf("counted: %v\n%s", combinedOutputError, output)
		}
		match := leakcheck.CountsLine.FindStringSubmatch(string(output))
		if match == nil {
			test.Fatalf("missing counts: %s", output)
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
	test.Logf("Link witness A/F/R/L/P/G: counted %s; graph %s", strings.Join(before, "/"), strings.Join(after, "/"))
	for _, field := range []int{2, 3} {
		old, parseError := strconv.Atoi(before[field])
		if parseError != nil {
			test.Fatal(parseError)
		}
		now, parseError := strconv.Atoi(after[field])
		if parseError != nil {
			test.Fatal(parseError)
		}
		if now >= old {
			test.Fatalf("graph ownership did not reduce count %d: %d -> %d", field, old, now)
		}
	}
}
