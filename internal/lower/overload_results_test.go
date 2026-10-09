package lower

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func checkOverloadResultsFixtures(t *testing.T, selected string) {
	t.Helper()
	for _, name := range []string{"transform", "evaluate", "binding", "parameter", "scalar"} {
		func() {
			if name != selected {
				return
			}
			source, err := os.ReadFile("../oracle/testdata/overload_results_" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerSource(t, string(source))
			if err != nil {
				t.Fatal(err)
			}
			instances := 0
			for _, function := range program.Functions {
				if strings.Contains(function.Name, "_specialized_") {
					instances++
				}
			}
			if instances == 0 {
				t.Fatal("resolved overload did not get a specialized result entry")
			}
		}()
	}
}

func TestOverloadResultsFixturesTransform(t *testing.T) {
	t.Parallel()
	checkOverloadResultsFixtures(t, "transform")
}

func TestOverloadResultsFixturesEvaluate(t *testing.T) {
	t.Parallel()
	checkOverloadResultsFixtures(t, "evaluate")
}

func TestOverloadResultsFixturesBinding(t *testing.T) {
	t.Parallel()
	checkOverloadResultsFixtures(t, "binding")
}

func TestOverloadResultsFixturesParameter(t *testing.T) {
	t.Parallel()
	checkOverloadResultsFixtures(t, "parameter")
}

func TestOverloadResultsFixturesScalar(t *testing.T) {
	t.Parallel()
	checkOverloadResultsFixtures(t, "scalar")
}

func checkOverloadResultsLiarStops(t *testing.T, selected string) {
	t.Helper()
	for _, name := range []string{"liar", "evaluate-liar", "parameter-liar"} {
		func() {
			if name != selected {
				return
			}
			path, err := filepath.Abs("testdata/overload-results/" + name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			nodeRunner, _ := filepath.Abs("../../oracle/node.mjs")
			node, err := exec.Command("node", "--disable-warning=ExperimentalWarning", nodeRunner, path).CombinedOutput()
			if err != nil {
				t.Fatalf("Node: %v %s", err, node)
			}
			expectedNode, output, message := "called text\nnumber\n", "called text\n", "overload 1 of lie result string cannot be served by implementation result string | number"
			if name == "evaluate-liar" {
				expectedNode, output, message = "template\n42\n", "template\n", "overload 1 of evaluate result Result<string | undefined> cannot be served by implementation result Result<string | number | undefined>"
			}
			if name == "parameter-liar" {
				expectedNode, output, message = "before\nimplementation\nundefined\n", "before\n", "overload 1 of visitNode parameter 1 Node | undefined cannot be served by implementation parameter Node"
			}
			if string(node) != expectedNode {
				t.Fatalf("Node got %q want %q", node, expectedNode)
			}
			source, _ := os.ReadFile(path)
			program, err := lowerSource(t, string(source))
			if err != nil {
				t.Fatal(err)
			}
			js := filepath.Join(t.TempDir(), "liar.mjs")
			if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
				t.Fatal(err)
			}
			commands := []*exec.Cmd{exec.Command("node", "--disable-warning=ExperimentalWarning", nodeRunner, js)}
			for _, sanitize := range []bool{false, true} {
				binary := filepath.Join(t.TempDir(), "liar")
				if err := native.Build(native.C(program), binary, native.Options{Sanitize: sanitize}); err != nil {
					t.Fatal(err)
				}
				command := exec.Command(binary)
				command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
				commands = append(commands, command)
			}
			for _, command := range commands {
				var stdout, stderr bytes.Buffer
				command.Stdout, command.Stderr = &stdout, &stderr
				err := command.Run()
				stopped, ok := err.(*exec.ExitError)
				if !ok || stopped.ExitCode() != 70 || stdout.String() != output || stderr.String() != "adamic: panic: "+message+"\n" {
					t.Fatalf("%s: %v, stdout %q stderr %q", command.Path, err, stdout.String(), stderr.String())
				}
			}
		}()
	}
}

func TestOverloadResultsLiarStopsLiar(t *testing.T) {
	t.Parallel()
	checkOverloadResultsLiarStops(t, "liar")
}

func TestOverloadResultsLiarStopsEvaluateLiar(t *testing.T) {
	t.Parallel()
	checkOverloadResultsLiarStops(t, "evaluate-liar")
}

func TestOverloadResultsLiarStopsParameterLiar(t *testing.T) {
	t.Parallel()
	checkOverloadResultsLiarStops(t, "parameter-liar")
}

func checkOverloadResultsRefuses(t *testing.T, selected string) {
	t.Helper()
	for _, probe := range []struct{ name, source, diagnostic string }{
		{"optional result property", `interface R {value:string|undefined;}interface OptionalR{value?:string|undefined;}function take(input:string):R;function take(input:string):OptionalR{return {};}console.log(String(take('x').value));`, "cannot be served"},
		{"shared writable result", `interface R<T> {value:T;} const shared:R<string|number>={value:1}; function take(input:string):R<string>;function take(input:string):R<string|number>{return shared;} console.log(take('x').value);`, "cannot be served"},
		{"escaping checked overload", `function lie(value:string):string;function lie(value:string):string|number{return 1;}const alias=lie;console.log(alias('x'));`, "indirect value of an overload"},
		{"unserved generic domain", `interface N {readonly kind:string;}interface B extends N{readonly kind:'binding';}function visit<T extends N>(value:T):T;function visit(value:B):B{return value;} const n:N={kind:'other'};console.log(visit(n).kind);`, "parameter 1 cannot be served"},
		{"shared readonly result", `interface R<T> {readonly value:T;} const shared:R<string|number>={value:1}; function take(input:string):R<string>;function take(input:string):R<string|number>{return shared;} console.log(take('x').value);`, "cannot be served"},
	} {
		func() {
			if probe.name != selected {
				return
			}
			_, err := lowerSource(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.diagnostic) {
				t.Fatalf("wanted %s, got %v", probe.diagnostic, err)
			}
		}()
	}
}

func TestOverloadResultsRefusesOptionalResultProperty(t *testing.T) {
	t.Parallel()
	checkOverloadResultsRefuses(t, "optional result property")
}

func TestOverloadResultsRefusesSharedWritableResult(t *testing.T) {
	t.Parallel()
	checkOverloadResultsRefuses(t, "shared writable result")
}

func TestOverloadResultsRefusesEscapingCheckedOverload(t *testing.T) {
	t.Parallel()
	checkOverloadResultsRefuses(t, "escaping checked overload")
}

func TestOverloadResultsRefusesUnservedGenericDomain(t *testing.T) {
	t.Parallel()
	checkOverloadResultsRefuses(t, "unserved generic domain")
}

func TestOverloadResultsRefusesSharedReadonlyResult(t *testing.T) {
	t.Parallel()
	checkOverloadResultsRefuses(t, "shared readonly result")
}
