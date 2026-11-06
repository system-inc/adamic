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
)

// Every host-binding row's columns are held to Node's real typeof, whether or not the row lowers today
// (#00qn5az, Adamic's ruling; internal/load/host_bindings.json). Each row has a program in
// testdata/host_bindings/typeof/<name>, its declaration as the row names it, in a file of the row's name,
// and console.log(typeof <name>). Node, running it as oracle/node.mjs runs every program, an ES module,
// must print the row's recorded answer; a column saying Absent where Node prints anything else fails,
// and so does a column saying Provided where Node prints "undefined", unless the row names a ruled
// divergence (ruledBackendDivergences). A row every column of which is Absent lowers, and each backend
// must print Node's answer; any other row stops as not yet, naming itself, and never prints. Mutant (d),
// testdata/host_bindings/mutants.md: process folded Absent goes red here.
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
		programs = append(programs, entry.Name())
	}
	var names []string
	for _, row := range rows {
		names = append(names, row.Name)
	}
	slices.Sort(programs)
	slices.Sort(names)
	if !slices.Equal(programs, names) {
		t.Fatalf("%s holds programs for %q, want one for each row, %q", directory, programs, names)
	}
	for _, row := range rows {
		t.Run(row.Name, func(t *testing.T) {
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
			divergence := false
			for _, ruled := range ruledBackendDivergences {
				divergence = divergence || row.Divergence != "" && ruled.name == row.Divergence
			}
			if row.Divergence != "" && !divergence {
				t.Fatalf("the row names the divergence %q, which no ruled backend divergence is", row.Divergence)
			}
			absent := true
			for _, target := range row.Targets() {
				absent = absent && target.Column.Answer == load.HostBindingAbsent
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
			if !absent {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "stage 0 can't lower the host binding "+row.Name+", ") {
					t.Errorf("got %v, want a not-yet naming the host binding %s", err, row.Name)
				}
				return
			}
			if err != nil {
				t.Fatalf("a row Absent on every target stops: %v", err)
			}
			generatorAgrees(t, relative)
		})
	}
}

// An Absent host binding's guarded reads agree with Node natively, sanitized and on the JavaScript
// backend, and leak nothing: tsc's own if (performance.ts:80), the right of && as a statement and as a
// boolean, the true arm of ?: holding a number and a string, and a parameter that shadows the declared
// name, judged as the parameter, whose call is made. Mutants (a) and (c) are in
// testdata/host_bindings/mutants.md; (c) goes red here, on shadow, as well as in lower's pin.
func TestHostBindingsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/host_bindings"
	programs := 0
	for _, name := range []string{"if", "and", "conditional", "shadow"} {
		programs++
		t.Run(name, func(t *testing.T) {
			t.Parallel()
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
