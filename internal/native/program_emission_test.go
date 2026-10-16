package native_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
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
func programCompilerWitness(t *testing.T, source string, mutant func(string) string) *ir.Program {
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
	program, err := lower.LowerWithOptions(context.Background(), checked, lower.Options{ProgramRegion: true})
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
	if source == programMapperSource {
		allocation := regexp.MustCompile(`adamic_array \*\s*(\w+) = \((\w+) \? adamic_retain\((\w+)\) : adamic_array_new\([^\n]+;`)
		matches := allocation.FindAllStringSubmatchIndex(code, -1)
		if len(matches) != 1 {
			t.Fatalf("mapper allocation count %d, want 1", len(matches))
		}
		m := matches[0]
		result, input := code[m[2]:m[3]], code[m[6]:m[7]]
		assertion := fmt.Sprintf("\n if (!adamic_program_is(%s) || %s == %s) adamic_panic(\"Program member array storage was reused\", sizeof \"Program member array storage was reused\" - 1);\n", input, result, input)
		code = code[:m[1]] + assertion + code[m[1]:]
	}

	if mutant != nil {
		changed := mutant(code)
		if changed == code {
			t.Fatal("mutant did not change C")
		}
		code = changed
	}
	binary := filepath.Join(directory, "native")
	if err := native.Build(code, binary, native.Options{Sanitize: true, ProgramRegion: true}); err != nil {
		t.Fatal(err)
	}
	got, diagnostic, err = run([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if mutant != nil {
		if err == nil && bytes.Equal(got, truth) {
			if report := leakcheck.Report(t, code, binary); report == "" {
				t.Fatal("Program ownership mutant survived")
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
	if err := native.Build(code, binary, native.Options{Count: true, ProgramRegion: true}); err != nil {
		t.Fatal(err)
	}
	got, diagnostic, err = run(nil, binary)
	if err != nil || !bytes.Equal(got, truth) {
		t.Fatalf("counted: %v\n%s\n%s", err, got, diagnostic)
	}
	if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: got, Stderr: diagnostic}); report != "" {
		t.Fatal(report)
	}
	if source == programOwnershipSource {
		fields := leakcheck.CountsLine.FindStringSubmatch(string(diagnostic))
		if len(fields) != 7 || fields[6] != "2" {
			t.Fatalf("membership regions: %s, want 2", diagnostic)
		}
	}
	t.Logf("%s", diagnostic)
	return program
}

const programOwnershipSource = `interface RegionNode { id: string; next: RegionNode | undefined; payload: string; }
interface CountedLeaf { text: string; }
function run(): void {
 const root: RegionNode = { id: ` + "`root${1}`" + `, next: undefined, payload: '' };
 const child: RegionNode = { id: ` + "`child${2}`" + `, next: undefined, payload: '' };
 root.next = child; child.next = root;
 root.payload = ` + "`outside${3}`" + `;
 console.log(root.payload);
 root.payload = ` + "`replacement${6}`" + `;
 console.log(root.next.id);
 const leaf: CountedLeaf = { text: ` + "`leaf${4}`" + ` }; console.log(leaf.text);
 const scratch = new Map<string,string>(); scratch.set('result', ` + "`scratch${5}`" + `);
 console.log(String(scratch.get('result')));
} run();`

func TestProgramCompilerOwnership(t *testing.T) {
	t.Parallel()
	programCompilerWitness(t, programOwnershipSource, nil)
}
func TestProgramCompilerCollections(t *testing.T) {
	t.Parallel()
	programCompilerWitness(t, graphCollectionSource, nil)
}
func TestProgramCompilerCapturedCell(t *testing.T) {
	t.Parallel()
	programCompilerWitness(t, `function make():()=>string { let next:(()=>string)|undefined=undefined; const read=():string=>next===undefined?'no':'yes';next=read; return read; } console.log(make()());`, nil)
}
func TestProgramCompilerConstructor(t *testing.T) {
	t.Parallel()
	programCompilerWitness(t, `class Link { next:Link|undefined=undefined; constructor(readonly name:string){} } function run():void {const a=new Link('a'+1);a.next=a;console.log(a.next.name);}run();`, nil)
}
func TestProgramCompilerAcyclicTree(t *testing.T) {
	t.Parallel()
	programCompilerWitness(t, `interface Leaf {text:string;} interface Tree {left:Leaf;right:Leaf;} const leaf:Leaf={text:'leaf'+1};const tree:Tree={left:leaf,right:leaf};console.log(tree.left.text+tree.right.text);`, nil)
}

const programMapperSource = `interface Node {id:number;children:Node[];}
function mapped():string {const source:Node[]=[{id:1,children:[]},{id:2,children:[]}];const result=source.map(node=>String(node.id));return result.join(',');}console.log(mapped());`

func TestProgramCompilerMemberMapperStorage(t *testing.T) {
	t.Parallel()
	programCompilerWitness(t, programMapperSource, nil)
}
