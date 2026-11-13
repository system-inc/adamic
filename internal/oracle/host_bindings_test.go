package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Every host-binding row's columns are held to Node's real typeof, whether or not the row lowers today
// (#00qn5az, Adamic's ruling; internal/load/host_bindings.json). Each row has a program at
// testdata/host_bindings/typeof/<name>/<file>, its declaration as the row names it, in a file named as
// the row's declaration is, and console.log(typeof <name>); a name with a row in two files, as process
// has in core.ts and tracing.ts, has a program for each. Node, running it as oracle/node.mjs runs every
// program, an ES module, must print the row's recorded answer; a column saying Absent where Node prints
// anything else fails, and so does a column saying Provided where Node prints "undefined", unless the
// row names a ruled divergence (ruledBackendDivergences). Every row names the Node it was measured on, and
// that has to be the Node the test runs, cloud/setup.sh's pin, or the rows are held to a Node they don't
// name (Workers' should on #93w9rtz, Adamic's yes). A recorded CommonJS answer is held the same
// way, to Node running the probe as a CommonJS script (onNodeCommonJS), the mode a bundle build answers
// for (#93w9rtz). A row whose columns are cleared lowers, and each backend must print Node's answer; a
// Pending row stops as not yet, naming itself, and never prints. Mutants (d) and (p),
// testdata/host_bindings/mutants.md: process folded Absent goes red here, and so does a CommonJS answer
// flipped.
func TestHostBindingsHeldToNode(t *testing.T) {
	t.Parallel()
	rows, err := load.HostBindings()
	if err != nil {
		t.Fatal(err)
	}
	directory := "internal/oracle/testdata/host_bindings/typeof"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	var programs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			programs = append(programs, entry.Name())
			continue
		}
		files, err := os.ReadDir(filepath.Join(repository, directory, entry.Name()))
		if err != nil {
			t.Fatal(err)
		}
		for _, file := range files {
			programs = append(programs, entry.Name()+"/"+file.Name())
		}
	}
	var declared []string
	for _, row := range rows {
		declared = append(declared, row.Name+"/"+row.Declaration.File)
	}
	version := onNodeCommonJS(t, "process.version")
	if version.exitCode != 0 {
		t.Fatalf("Node's version: exit %d, stderr %q", version.exitCode, version.stderr)
	}
	for _, row := range rows {
		if string(version.stdout) != row.Node.Version+"\n" {
			t.Errorf("%s in %s: the row names Node %s, and this test runs %s", row.Name, row.Declaration.File, row.Node.Version, strings.TrimSpace(string(version.stdout)))
		}
	}
	slices.Sort(programs)
	slices.Sort(declared)
	if !slices.Equal(programs, declared) {
		t.Fatalf("%s holds programs %q, want one for each row, %q", directory, programs, declared)
	}
	for _, row := range rows {
		t.Run(row.Name+"/"+row.Declaration.File, func(t *testing.T) {
			t.Parallel()
			relative := directory + "/" + row.Name + "/" + row.Declaration.File
			path, err := filepath.Abs(filepath.Join(repository, relative))
			if err != nil {
				t.Fatal(err)
			}
			node := onNode(t, path)
			if node.exitCode != 0 || string(node.stdout) != row.Node.TypeOf+"\n" {
				t.Fatalf("Node: exit %d, stdout %q, stderr %q, want the row's typeof %q", node.exitCode, node.stdout, node.stderr, row.Node.TypeOf)
			}
			if row.Node.CommonJS != "" {
				script := onNodeCommonJS(t, "typeof "+row.Name)
				if script.exitCode != 0 || string(script.stdout) != row.Node.CommonJS+"\n" {
					t.Fatalf("Node, as a CommonJS script: exit %d, stdout %q, stderr %q, want the row's CommonJS typeof %q", script.exitCode, script.stdout, script.stderr, row.Node.CommonJS)
				}
			}
			divergence := false
			for _, ruled := range ruledBackendDivergences {
				divergence = divergence || row.Divergence != "" && ruled.name == row.Divergence
			}
			if row.Divergence != "" && !divergence {
				t.Fatalf("the row names the divergence %q, which no ruled backend divergence is", row.Divergence)
			}
			cleared := true
			for _, target := range row.Targets() {
				cleared = cleared && target.Column.Answer != load.HostBindingPending
				if divergence {
					continue
				}
				if target.Column.Answer == load.HostBindingAbsent && row.Node.TypeOf != "undefined" {
					t.Errorf("the %s column says Absent where Node's typeof is %q", target.Target, row.Node.TypeOf)
				}
				if target.Column.Answer == load.HostBindingProvided && row.Node.TypeOf == "undefined" {
					t.Errorf("the %s column says Provided where Node's typeof is \"undefined\"", target.Target)
				}
			}
			_, err = lowered(t, path)
			if !cleared {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "stage 0 can't lower the host binding "+row.Name+", ") {
					t.Errorf("got %v, want a not-yet naming the host binding %s", err, row.Name)
				}
				return
			}
			if err != nil {
				t.Fatalf("a row cleared on every target stops: %v", err)
			}
			generatorAgrees(t, relative)
		})
	}
}

// onNodeCommonJS is Node running console.log(<expression>) as a CommonJS script, node probe.cjs, the mode
// tsc's bundle _tsc.js runs in. A row's probe is its declaration and its typeof; the declaration is a
// type alone, which stripping a .ts file's types erases, so the script is the probe's JavaScript.
func onNodeCommonJS(t *testing.T, expression string) run {
	t.Helper()
	path := filepath.Join(t.TempDir(), "probe.cjs")
	if err := os.WriteFile(path, []byte("console.log("+expression+");\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return execute(t, "node", path)
}

// onNodeCommonJS runs a CommonJS script, not a module: the module wrapper's names are defined there, and
// none is in the ES module the oracle runs (#93w9rtz). Without this, a helper that ran an ES module would
// hold every row's CommonJS answer to its ES one and pass.
func TestNodeCommonJSProbeIsAScript(t *testing.T) {
	t.Parallel()
	script := onNodeCommonJS(t, "typeof require, typeof module, typeof exports, typeof __filename")
	if script.exitCode != 0 || string(script.stdout) != "function object object string\n" {
		t.Fatalf("Node, as a CommonJS script: exit %d, stdout %q, stderr %q, want the module wrapper's names", script.exitCode, script.stdout, script.stderr)
	}
}

// An Absent host binding's guarded reads agree with Node natively, sanitized and on the JavaScript
// backend, and leak nothing: tsc's own if (performance.ts:80), the right of && as a statement and as a
// boolean, the true arm of ?: holding a number and a string, a parameter that shadows the declared
// name, judged as the parameter, whose call is made, and core.ts's require in isNodeLikeSystem's test,
// under a guard and in typeof (node_like), which Node in its module mode answers false, "no loader" and
// "undefined". Mutants (a), (c) and (k) are in testdata/host_bindings/mutants.md; (c) goes red here, on
// shadow, as well as in lower's pin, and (k) on node_like.
func TestHostBindingsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/host_bindings"
	programs := 0
	for _, name := range []string{"if", "and", "conditional", "shadow", "performance_provided", "process_provided", "process_browser_boolean", "process_browser_string", "process_browser_number", "node_like"} {
		programs++
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			if strings.HasPrefix(name, "process_browser_") {
				hostBindingTypedProcessAgrees(t, strings.TrimPrefix(name, "process_browser_"))
				return
			}
			generatorAgrees(t, directory+"/"+name+"/main.ts")
		})
	}
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	directories := 0
	for _, entry := range entries {
		if entry.IsDir() && entry.Name() != "typeof" && entry.Name() != "global_object" {
			directories++
		}
	}
	if directories != programs {
		t.Fatalf("%d programs in %s, want %d: a program added or removed needs this list changed with it", directories, directory, programs)
	}
}

// Language's p11 (#00qn5az): Reflect.set(globalThis, "onProfilerEvent", f) installs the global on Node,
// so typeof turns "function" there and the guarded call prints, where native, having folded typeof to
// "undefined", would print nothing: the divergence the host-binding fold has to rule out. Adamic stops
// it as the global object as a value (lower/global_object.go), and Node is held to printing, so the pin
// shows what the stop prevents.
func TestHostBindingGlobalObjectRoadIsNotYet(t *testing.T) {
	t.Parallel()
	relative := "internal/oracle/testdata/host_bindings/global_object/reflect_set/main.ts"
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	if node := onNode(t, path); node.exitCode != 0 || string(node.stdout) != "set mark\n" {
		t.Fatalf("Node: exit %d, stdout %q, stderr %q, want Reflect.set to install onProfilerEvent and the guarded call to print", node.exitCode, node.stdout, node.stderr)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), "main.ts:6:13: stage 0 can't lower the global object as a value (only a property read's or member call's receiver is admitted) yet") {
		t.Errorf("got %v, want the global object as a value to stop at main.ts:6:13", err)
	}
}

// generatorAgrees holds native, sanitized and JavaScript; this pin executes the same process row
// fixture as a WASI command, with no permitted refusal (#zqz01vh, Adamic and Lower, Nov 8 2026).
func TestHostBindingProcessWASIAgreesWithNode(t *testing.T) {
	t.Parallel()
	if os.Getenv("ADAMIC_ORACLE_WASI") != "1" {
		t.Skip("set ADAMIC_ORACLE_WASI=1")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/host_bindings/process_provided/main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	truth, got := onNode(t, path), onWASI(t, native.C(program))
	if difference := disagreement(truth, got); difference != "" {
		t.Fatalf("%s: Node %+v WASI %+v", difference, truth, got)
	}
}
