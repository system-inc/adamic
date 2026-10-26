package lower

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
)

// A refusal becoming graph ownership must agree with source Node and finish
// with no heap values left. Empty output is valid for ownership-only programs.
func graphAccepted(t *testing.T, source string) *ir.Program {
	t.Helper()
	path := writeAgreementSource(t, "main.a", source)
	truth := runAgreementNode(t, path)
	if truth.code != 0 {
		t.Fatalf("graph acceptance source must finish: %d %s", truth.code, truth.stderr)
	}
	program := lowerAgreementEntry(t, path)
	if len(program.GraphTypes) == 0 {
		t.Fatal("cycle was accepted without graph ownership")
	}
	js := filepath.Join(t.TempDir(), "main.mjs")
	if err := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	compareAgreement(t, runAgreementNode(t, js), truth)
	code := native.C(program)
	binary := filepath.Join(t.TempDir(), "native")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var out, diagnostic bytes.Buffer
	command.Stdout = &out
	command.Stderr = &diagnostic
	if err := command.Run(); err != nil || !bytes.Equal(out.Bytes(), truth.stdout) {
		t.Fatalf("graph native disagrees: %v\n%s\n%s, Node %s", err, diagnostic.Bytes(), out.Bytes(), truth.stdout)
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
	return program
}

func TestGraphClosedFrameInputStaysCounted(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function parser(scanner: { scan: () => number }): () => number {
 let token = 0;
 function read(): number { return token; }
 function next(): number { return token = scanner.scan(); }
 next(); return read;
 } console.log(String(parser({scan: () => 1})()));`)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) != 0 {
		t.Fatal("closed frame input seeded graph ownership")
	}
}

func TestGraphFlowReturnAndConditional(t *testing.T) {
	t.Parallel()
	program := graphAccepted(t, `interface Link {next: Link | undefined}
 function empty(){return {next:undefined as Link|undefined};}
 function choose(flag:boolean):Link {return flag?empty():empty();}
 function run(){const a=choose(true);const b=choose(false);a.next=b;b.next=a;console.log(String(a.next===b));}
 run();`)
	literals := 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if value, ok := node.(ir.ObjectLiteral); ok {
				literals++
				if len(value.GraphTypes) == 0 || !program.GraphTypes[value.GraphTypes[len(value.GraphTypes)-1]] {
					t.Fatal("return/conditional allocation stayed counted")
				}
			}
			return true
		})
	}
	if literals == 0 {
		t.Fatal("missing allocation")
	}
}

const graphReviewCycle = `interface Link {next:Link|undefined} const link:Link={next:undefined};link.next=link;`

func TestGraphViewAdoptionIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, interfaceCastSource+graphReviewCycle+`console.log(identifier({kind:'identifier',pos:0,name:'ok'} as Identifier).name);`)
	if err == nil || !strings.Contains(err.Error(), "checked-view adapter") {
		t.Fatalf("view graph adoption: %v", err)
	}
}
func TestGraphDynamicShapeAdoptionIsRefused(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../oracle/testdata/entries_record_alias.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(data)+graphReviewCycle)
	if err == nil || !strings.Contains(err.Error(), "dynamic-shape") {
		t.Fatalf("dynamic graph adoption: %v", err)
	}
}
func TestGraphRecordCycleStaysRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, `interface Ring {self:Ring|undefined;[key:string]:Ring|undefined} const ring:Ring={self:undefined};ring.self=ring;console.log(Object.keys(ring).join(','));`)
	if err == nil {
		t.Fatal("record cycle accepted")
	}
	t.Log(err)
}
func TestGraphCountTypesSurviveClassification(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function make():()=>boolean {let saved:(()=>boolean)|undefined;function f():boolean{return arguments.length===0 && saved===f;}saved=f;return f;}console.log(String(make()()));`)
	if err != nil {
		t.Fatal(err)
	}
	for i, function := range program.Functions {
		if function.GraphClosure && function.ReadsArguments {
			if !program.PackedCountNeeded(i) {
				t.Fatal("graph closure lost argument-count ABI")
			}
			if len(program.FunctionTypeTargets) == 0 {
				t.Fatal("count target proof did not run after graph selection")
			}
			return
		}
	}
	t.Fatal("missing graph argument-count reader")
}

// Main keeps derived computed members outside lowering until their implementation gap is resolved.
func graphComputedIteratorGap(t *testing.T, source string) {
	t.Helper()
	_, err := lowerSource(t, source)
	var gap *NotYet
	if !errors.As(err, &gap) || !strings.Contains(gap.What, "a computed member name in a derived class") {
		t.Fatalf("want main's derived computed-member gap, got %v", err)
	}
}

func TestGraphComputedIteratorInheritance(t *testing.T) {
	t.Parallel()
	graphComputedIteratorGap(t, graphReviewCycle+`class Base {[Symbol.iterator](){return {next(){return {value:1,done:false};}};}} class Child extends Base {[Symbol.iterator](){return {next(){return {value:2,done:false};}};}} for(const n of new Child()){console.log(String(n));break;}console.log(String(link.next===link));`)
}
func TestGraphComputedIteratorIntroduced(t *testing.T) {
	t.Parallel()
	graphComputedIteratorGap(t, graphReviewCycle+`class Base {} class Child extends Base {[Symbol.iterator](){return {next(){return {value:2,done:false};}};}} for(const n of new Child()){console.log(String(n));break;}console.log(String(link.next===link));`)
}

func TestGraphCheckerTypeArguments(t *testing.T) {
	t.Parallel()
	graphAccepted(t, `// a-check: refused adamic/cycle-capable
// Cycle projection of tsc types.ts, cohere/TypeScript d92d9bfee114c80be2c375d72edae966176e3a4f.
// Node:942-955; Declaration:1756-1760; NodeArray:1589-1592; Symbol:6037-6054.
// Node.parent includes undefined for an honest unbound root. NodeArray omits
// metadata fields this base cannot construct; readonly elements keep the cycle.
interface ReadonlyTextRange {
    readonly pos: number;
    readonly end: number;
}
interface Node extends ReadonlyTextRange {
    readonly parent: Node | undefined;
}
interface Declaration extends Node {
    symbol: Symbol;
}
interface Symbol {
    declarations?: Declaration[];
}
type NodeArray<T extends Node> = readonly T[];

class BaseNode implements Declaration {
    readonly pos: number;
    readonly end: number;
    parent: Node | undefined;
    symbol: Symbol;
    constructor(parent: Node | undefined, symbol: Symbol, pos: number) {
        this.pos = pos;
        this.end = pos + 1;
        this.parent = parent;
        this.symbol = symbol;
    }
}
function createBaseNode(parent: Node | undefined, symbol: Symbol, pos: number): BaseNode {
    return new BaseNode(parent, symbol, pos);
}
function createNodeArray<T extends Node>(elements: readonly T[]): NodeArray<T> {
    return elements;
}
function run(): void {
    const symbol: Symbol = {declarations: []};
    const root = createBaseNode(undefined, symbol, 0);
    const first = createBaseNode(root, symbol, 1);
    const second = createBaseNode(root, symbol, 2);
    const declarations: Declaration[] = [root, first, second];
    symbol.declarations = declarations;
    const nodes = createNodeArray<Node>([root, first, second]);
    const node = nodes[1];
    const declaration = symbol.declarations[1];
    if (node !== undefined && declaration !== undefined) {
        console.log(String(nodes.length)+"|"+String(node.parent?.pos)+"|"+String(declaration.symbol.declarations?.length));
        console.log(String(node === first));
        console.log(String(declaration.symbol === symbol));
        console.log(String(first.parent === root));
    }
}
run();
`)
}

func TestGraphFlowVirtualTargets(t *testing.T) {
	t.Parallel()
	program := graphAccepted(t, `interface Link { readonly label: string; next: Link | undefined; }
class Maker {
 make() { return { label: 'base' + String(0), next: undefined as Link | undefined }; }
}
class Override extends Maker {
 override make() { return { label: 'override' + String(1), next: undefined as Link | undefined }; }
}
function cycle(maker: Maker): Link {
 const view: Link = maker.make();
 view.next = view;
 return view;
}
function run(): void {
 const maker: Maker = new Override();
 const escaped = cycle(maker);
 console.log(escaped.label + ' ' + (escaped.next?.label ?? '-'));
}
run();
`)
	sites := 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			if literal, ok := node.(ir.ObjectLiteral); ok && len(literal.Fields) > 0 {
				for _, field := range literal.Fields {
					if field.Name == "label" {
						sites++
						if len(literal.GraphTypes) == 0 || !program.GraphTypes[literal.GraphTypes[len(literal.GraphTypes)-1]] {
							t.Fatal("virtual return allocation has no graph flow identity")
						}
					}
				}
			}
			return true
		})
	}
	if sites != 2 {
		t.Fatalf("wanted base and override allocation, got %d", sites)
	}
}

func TestGraphLegacyViewAdoptionIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, graphReviewCycle+`
 type First={readonly kind:'a';readonly name:string};
 type Second={readonly kind:'b';readonly name:string};
 type Third={readonly kind:'c';readonly name:string};
 function narrow(value:First|Second|Third):First|Second{return value as First|Second;}
 console.log(narrow({kind:'a',name:'kept'}).name);`)
	if err == nil || !strings.Contains(err.Error(), "checked-view adapter") {
		t.Fatalf("legacy view graph adoption: %v", err)
	}
}

func TestGraphFlowStoredAllocation(t *testing.T) {
	t.Parallel()
	program := graphAccepted(t, `interface Link {label:string;next:Link|undefined}
 const root:Link={label:'root',next:undefined};
 root.next={label:'inner',next:root};console.log(root.next.label);`)
	stored := false
	walk(program.Main, func(node any) bool {
		if write, ok := node.(ir.SetProperty); ok && write.Name == "next" {
			if literal, ok := write.Value.(ir.ObjectLiteral); ok {
				stored = true
				if len(literal.GraphTypes) == 0 || !program.GraphTypes[literal.GraphTypes[len(literal.GraphTypes)-1]] {
					t.Fatal("slot initializer has no graph flow identity")
				}
			}
		}
		return true
	})
	if !stored {
		t.Fatal("missing inline slot initializer")
	}
}
