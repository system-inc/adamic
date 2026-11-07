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
func programCompilerWitness(test *testing.T, source string, mutant func(string) string) *ir.Program {
	test.Helper()
	directory := test.TempDir()
	path := filepath.Join(directory, "main.a")
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
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
	program, lowerWithOptionsError := lower.LowerWithOptions(
		context.Background(),
		checked,
		lower.Options{ProgramRegion: true},
	)
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
	if source == programMapperSource {
		allocation := regexp.MustCompile(
			`adamic_array \*\s*(\w+) = \((\w+) \? adamic_retain\((\w+)\) : adamic_array_make\([^\n]+;`,
		)
		matches := allocation.FindAllStringSubmatchIndex(code, -1)
		if len(matches) != 1 {
			test.Fatalf("mapper allocation count %d, want 1", len(matches))
		}
		m := matches[0]
		result, input := code[m[2]:m[3]], code[m[6]:m[7]]
		assertion := fmt.Sprintf(
			"\n if (!adamic_program_is(%s) || %s == %s) adamic_panic(\"Program member array storage was reused\", sizeof \"Program member array storage was reused\" - 1);\n",
			input,
			result,
			input,
		)
		code = code[:m[1]] + assertion + code[m[1]:]
	}

	if mutant != nil {
		changed := mutant(code)
		if changed == code {
			test.Fatal("mutant did not change C")
		}
		code = changed
	}
	binary := filepath.Join(directory, "native")
	if buildError := native.Build(
		code,
		binary,
		native.Options{Sanitize: true, ProgramRegion: true},
	); buildError != nil {
		test.Fatal(buildError)
	}
	got, diagnostic, runError = run([]string{"ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if mutant != nil {
		if runError == nil && bytes.Equal(got, truth) {
			if report := leakcheck.Report(test, code, binary); report == "" {
				test.Fatal("Program ownership mutant survived")
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
	if buildError := native.Build(code, binary, native.Options{Count: true, ProgramRegion: true}); buildError != nil {
		test.Fatal(buildError)
	}
	got, diagnostic, runError = run(nil, binary)
	if runError != nil || !bytes.Equal(got, truth) {
		test.Fatalf("counted: %v\n%s\n%s", runError, got, diagnostic)
	}
	if report := leakcheck.Unbalanced(leakcheck.Run{Stdout: got, Stderr: diagnostic}); report != "" {
		test.Fatal(report)
	}
	if source == programOwnershipSource {
		fields := leakcheck.CountsLine.FindStringSubmatch(string(diagnostic))
		if len(fields) != 7 || fields[6] != "2" {
			test.Fatalf("membership regions: %s, want 2", diagnostic)
		}
	}
	if source == programContainerSource {
		fields := leakcheck.CountsLine.FindStringSubmatch(string(diagnostic))
		if len(fields) != 7 || fields[6] != "6" {
			test.Fatalf("container membership regions: %s, want 6", diagnostic)
		}
	}
	test.Logf("%s", diagnostic)
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

func TestProgramCompilerOwnership(test *testing.T) {
	test.Parallel()
	programCompilerWitness(test, programOwnershipSource, nil)
}
func TestProgramCompilerCollections(test *testing.T) {
	test.Parallel()
	programCompilerWitness(test, graphCollectionSource, nil)
}
func TestProgramCompilerCapturedCell(test *testing.T) {
	test.Parallel()
	programCompilerWitness(
		test,
		`function make():()=>string { let next:(()=>string)|undefined=undefined; const read=():string=>next===undefined?'no':'yes';next=read; return read; } console.log(make()());`,
		nil,
	)
}
func TestProgramCompilerConstructor(test *testing.T) {
	test.Parallel()
	programCompilerWitness(
		test,
		`class Link { next:Link|undefined=undefined; constructor(readonly name:string){} } function run():void {const a=new Link('a'+1);a.next=a;console.log(a.next.name);}run();`,
		nil,
	)
}
func TestProgramCompilerAcyclicTree(test *testing.T) {
	test.Parallel()
	programCompilerWitness(
		test,
		`interface Leaf {text:string;} interface Tree {left:Leaf;right:Leaf;} const leaf:Leaf={text:'leaf'+1};const tree:Tree={left:leaf,right:leaf};console.log(tree.left.text+tree.right.text);`,
		nil,
	)
}

const programMapperSource = `interface Node {id:number;children:Node[];}
function mapped():string {const source:Node[]=[{id:1,children:[]},{id:2,children:[]}];const result=source.map(node=>String(node.id));return result.join(',');}console.log(mapped());`

func TestProgramCompilerMemberMapperStorage(test *testing.T) {
	test.Parallel()
	programCompilerWitness(test, programMapperSource, nil)
}

// None of these containers has a backlink from its element. Their concrete
// element, key or value type selects them after the owning SCC is established.
const programContainerSource = `interface Link {next:Link|undefined;}
function run():void {
 const root:Link={next:undefined};root.next=root;
 const keys=new Map<Link,string>();keys.set(root,'root'+1);
 const values=new Map<string,Link>();values.set('root',root);
 const nodes=new Set<Link>();nodes.add(root);
 const tuple:[Link,number]=[root,1];const list:Link[]=[root];
 console.log(String(keys.get(root))+String(values.get('root')===root)+String(nodes.has(root))+String(tuple[0]===root)+String(list[0]===root));
}run();`

func TestProgramCompilerContainerPropagation(test *testing.T) {
	test.Parallel()
	programCompilerWitness(test, programContainerSource, nil)
}
