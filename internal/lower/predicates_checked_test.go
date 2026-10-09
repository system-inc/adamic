package lower

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// The .a witnesses are also source Node inputs. A temporary .ts copy explicitly
// opts into checked admission; the identical .a lying bodies remain refused.
func TestCheckedPredicateBranches(t *testing.T) {
	for _, probe := range []struct{ name, nodeOut, checkedOut, branch string }{
		{"true", "claimed\n", "", "true"},
		{"false", "continued\n", "", "false"},
		{"asserts", "called\nasserted\n", "called\n", "asserts"},
		{"valid", "argument\ncalled\nnumber\nargument\ncalled\nstring\n", "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			path, err := filepath.Abs("testdata/predicates_checked/" + probe.name + ".a")
			if err != nil {
				t.Fatal(err)
			}
			runner, err := filepath.Abs("../../oracle/node.mjs")
			if err != nil {
				t.Fatal(err)
			}
			run := func(command *exec.Cmd, stdout, stderr string, code int) {
				t.Helper()
				var out, errors bytes.Buffer
				command.Stdout = &out
				command.Stderr = &errors
				err := command.Run()
				got := 0
				if err != nil {
					if exit, ok := err.(*exec.ExitError); ok {
						got = exit.ExitCode()
					} else {
						t.Fatal(err)
					}
				}
				if out.String() != stdout || errors.String() != stderr || got != code {
					t.Fatalf("%s: exit %d stdout %q stderr %q; want %d %q %q", command.Path, got, out.String(), errors.String(), code, stdout, stderr)
				}
			}
			run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, path), probe.nodeOut, "", 0)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowerTypeScriptAssertionSource(t, string(source))
			if err != nil {
				t.Fatal(err)
			}
			if program.PredicateChecks.Checked == 0 || len(program.PredicateChecks.Sites) == 0 {
				t.Fatalf("missing checked site: %+v", program.PredicateChecks)
			}
			stdout, stderr, code := probe.nodeOut, "", 0
			if probe.branch != "" {
				stdout = probe.checkedOut
				code = 70
				site := program.PredicateChecks.Sites[0]
				message := fmt.Sprintf("predicate %s at %s %s branch: source string | number, target number", site.Function, site.Where, probe.branch)
				stderr = "adamic: panic: " + message + "\n"
				_, err = lowerSource(t, string(source))
				if err == nil || !strings.Contains(err.Error(), "type predicate") {
					t.Fatalf("lying .a body admitted: %v", err)
				}
			}
			binary := filepath.Join(t.TempDir(), "native")
			if err = native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			command := exec.Command(binary)
			command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
			run(command, stdout, stderr, code)
			generated := filepath.Join(t.TempDir(), "generated.mjs")
			if err = os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0644); err != nil {
				t.Fatal(err)
			}
			run(exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated), stdout, stderr, code)
		})
	}
}

func TestCheckedPredicatePendingContracts(t *testing.T) {
	for _, probe := range []struct{ name, source, diagnostic string }{
		{"untagged", `interface Narrow { value: number } function claim(value: {value:number|string}): value is Narrow {return true;} claim({value:"text"});`, "checked predicate overload target Narrow"},
		{"open tag", `interface Numeric {kind:1;value:number} interface Textual {kind:1;value:string} function claim(value:Numeric|Textual):value is Numeric{return false;} function use(value:Numeric|Textual):void{if(claim(value)){}else{}} use({kind:1,value:"text"});`, "checked predicate membership over an open tag domain"},
		{"indirect", `function claim(value:number|string):value is number{return true;} const alias=claim; alias("text");`, "an indirect call of a checked predicate"},
		{"overload in .a", `function claim(value:number|string):value is number;function claim(value:number|string):boolean{return true;} claim("text");`, "the overload implementation does not prove both directions"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			lowerInput := lowerTypeScriptAssertionSource
			if probe.name == "overload in .a" {
				lowerInput = lowerSource
			}
			_, err := lowerInput(t, probe.source)
			if err == nil || !strings.Contains(err.Error(), probe.diagnostic) {
				t.Fatalf("want pending/refused %q, got %v", probe.diagnostic, err)
			}
		})
	}
}
