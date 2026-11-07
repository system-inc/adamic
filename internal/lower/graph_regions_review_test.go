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
func graphAccepted(test *testing.T, source string) *ir.Program {
	test.Helper()
	path := writeAgreementSource(test, "main.a", source)
	truth := runAgreementNode(test, path)
	if truth.code != 0 {
		test.Fatalf("graph acceptance source must finish: %d %s", truth.code, truth.stderr)
	}
	program := lowerAgreementEntry(test, path)
	if len(program.GraphTypes) == 0 {
		test.Fatal("cycle was accepted without graph ownership")
	}
	js := filepath.Join(test.TempDir(), "main.mjs")
	if writeError := os.WriteFile(js, []byte(javascript.JavaScript(program)), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	compareAgreement(test, runAgreementNode(test, js), truth)
	code := native.C(program)
	binary := filepath.Join(test.TempDir(), "native")
	if buildError := native.Build(code, binary, native.Options{Sanitize: true}); buildError != nil {
		test.Fatal(buildError)
	}
	timeoutContext, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(timeoutContext, binary)
	command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
	var out, diagnostic bytes.Buffer
	command.Stdout = &out
	command.Stderr = &diagnostic
	if runError := command.Run(); runError != nil || !bytes.Equal(out.Bytes(), truth.stdout) {
		test.Fatalf(
			"graph native disagrees: %v\n%s\n%s, Node %s",
			runError,
			diagnostic.Bytes(),
			out.Bytes(),
			truth.stdout,
		)
	}
	if report := leakcheck.Report(test, code, binary); report != "" {
		test.Fatal(report)
	}
	return program
}

func TestGraphClosedFrameInputStaysCounted(test *testing.T) {
	test.Parallel()
	program, lowerSourceError := lowerSource(test, `function parser(scanner: { scan: () => number }): () => number {
 let token = 0;
 function read(): number { return token; }
 function next(): number { return token = scanner.scan(); }
 next(); return read;
 } console.log(String(parser({scan: () => 1})()));`)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	if len(program.GraphTypes) != 0 {
		test.Fatal("closed frame input seeded graph ownership")
	}
}

func TestGraphFlowReturnAndConditional(test *testing.T) {
	test.Parallel()
	program := graphAccepted(test, `interface Link {next: Link | undefined}
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
					test.Fatal("return/conditional allocation stayed counted")
				}
			}
			return true
		})
	}
	if literals == 0 {
		test.Fatal("missing allocation")
	}
}

const graphReviewCycle = `interface Link {next:Link|undefined} const link:Link={next:undefined};link.next=link;`

// In .a the view's cast is a downcast whose checks no proof erases, refused (#m8dvvmf) before graph
// adoption is reached; the .ts twin keeps the checked view and pins the adoption refusal.
func TestGraphViewAdoptionIsRefused(test *testing.T) {
	test.Parallel()
	source := interfaceCastSource +
		graphReviewCycle +
		`console.log(identifier({kind:'identifier',pos:0,name:'ok'} as Identifier).name);`
	_, lowerSourceError := lowerSource(test, source)
	if lowerSourceError == nil ||
		!strings.Contains(
			lowerSourceError.Error(),
			"a downcast from Node to Identifier in .a whose checks no proof erases",
		) {
		test.Fatalf("view graph adoption in .a: %v", lowerSourceError)
	}
	_, lowerSourceAsError := lowerSourceAs(test, source, ".ts")
	if lowerSourceAsError == nil || !strings.Contains(lowerSourceAsError.Error(), "checked-view adapter") {
		test.Fatalf("view graph adoption: %v", lowerSourceAsError)
	}
}
func TestGraphDynamicShapeAdoptionIsRefused(test *testing.T) {
	test.Parallel()
	data, readError := os.ReadFile("../oracle/testdata/entries_record_alias.a")
	if readError != nil {
		test.Fatal(readError)
	}
	_, lowerSourceError := lowerSource(test, string(data)+graphReviewCycle)
	if lowerSourceError == nil || !strings.Contains(lowerSourceError.Error(), "dynamic-shape") {
		test.Fatalf("dynamic graph adoption: %v", lowerSourceError)
	}
}
func TestGraphRecordCycleStaysRefused(test *testing.T) {
	test.Parallel()
	_, lowerSourceError := lowerSource(
		test,
		`interface Ring {self:Ring|undefined;[key:string]:Ring|undefined} const ring:Ring={self:undefined};ring.self=ring;console.log(Object.keys(ring).join(','));`,
	)
	if lowerSourceError == nil {
		test.Fatal("record cycle accepted")
	}
	test.Log(lowerSourceError)
}
func TestGraphCountTypesSurviveClassification(test *testing.T) {
	test.Parallel()
	program, lowerSourceError := lowerSource(
		test,
		`function make():()=>boolean {let saved:(()=>boolean)|undefined;function f():boolean{return arguments.length===0 && saved===f;}saved=f;return f;}console.log(String(make()()));`,
	)
	if lowerSourceError != nil {
		test.Fatal(lowerSourceError)
	}
	for index, function := range program.Functions {
		if function.GraphClosure && function.ReadsArguments {
			if !program.PackedCountNeeded(index) {
				test.Fatal("graph closure lost argument-count ABI")
			}
			if len(program.FunctionTypeTargets) == 0 {
				test.Fatal("count target proof did not run after graph selection")
			}
			return
		}
	}
	test.Fatal("missing graph argument-count reader")
}

// Main keeps derived computed members outside lowering until their implementation gap is resolved.
func graphComputedIteratorGap(test *testing.T, source string) {
	test.Helper()
	_, lowerSourceError := lowerSource(test, source)
	var gap *NotYet
	if !errors.As(lowerSourceError, &gap) || !strings.Contains(gap.What, "a computed member name in a derived class") {
		test.Fatalf("want main's derived computed-member gap, got %v", lowerSourceError)
	}
}

func TestGraphComputedIteratorInheritance(test *testing.T) {
	test.Parallel()
	graphComputedIteratorGap(test, graphReviewCycle+
		`class Base {[Symbol.iterator](){return {next(){return {value:1,done:false};}};}} class Child extends Base {[Symbol.iterator](){return {next(){return {value:2,done:false};}};}} for(const n of new Child()){console.log(String(n));break;}console.log(String(link.next===link));`)
}
func TestGraphComputedIteratorIntroduced(test *testing.T) {
	test.Parallel()
	graphComputedIteratorGap(test, graphReviewCycle+
		`class Base {} class Child extends Base {[Symbol.iterator](){return {next(){return {value:2,done:false};}};}} for(const n of new Child()){console.log(String(n));break;}console.log(String(link.next===link));`)
}

func TestGraphCheckerTypeArguments(test *testing.T) {
	test.Parallel()
	graphAccepted(test, `// a-check: refused adamic/cycle-capable
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

func TestGraphFlowVirtualTargets(test *testing.T) {
	test.Parallel()
	program := graphAccepted(test, `interface Link { readonly label: string; next: Link | undefined; }
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
						if len(literal.GraphTypes) == 0 ||
							!program.GraphTypes[literal.GraphTypes[len(literal.GraphTypes)-1]] {
							test.Fatal("virtual return allocation has no graph flow identity")
						}
					}
				}
			}
			return true
		})
	}
	if sites != 2 {
		test.Fatalf("wanted base and override allocation, got %d", sites)
	}
}

// In .a the cast to a sub-union is a downcast whose checks no proof erases, refused (#m8dvvmf) before
// graph adoption is reached; the .ts twin keeps the checked view and pins the adoption refusal.
func TestGraphLegacyViewAdoptionIsRefused(test *testing.T) {
	test.Parallel()
	source := graphReviewCycle + `
 type First={readonly kind:'a';readonly name:string};
 type Second={readonly kind:'b';readonly name:string};
 type Third={readonly kind:'c';readonly name:string};
 function narrow(value:First|Second|Third):First|Second{return value as First|Second;}
 console.log(narrow({kind:'a',name:'kept'}).name);`
	_, lowerSourceError := lowerSource(test, source)
	if lowerSourceError == nil || !strings.Contains(
		lowerSourceError.Error(),
		"a downcast from First | Second | Third to First | Second in .a whose checks no proof erases",
	) {
		test.Fatalf("legacy view graph adoption in .a: %v", lowerSourceError)
	}
	_, lowerSourceAsError := lowerSourceAs(test, source, ".ts")
	if lowerSourceAsError == nil || !strings.Contains(lowerSourceAsError.Error(), "checked-view adapter") {
		test.Fatalf("legacy view graph adoption: %v", lowerSourceAsError)
	}
}

func TestGraphFlowStoredAllocation(test *testing.T) {
	test.Parallel()
	program := graphAccepted(test, `interface Link {label:string;next:Link|undefined}
 const root:Link={label:'root',next:undefined};
 root.next={label:'inner',next:root};console.log(root.next.label);`)
	stored := false
	walk(program.Main, func(node any) bool {
		if write, ok := node.(ir.SetProperty); ok && write.Name == "next" {
			if literal, ok := write.Value.(ir.ObjectLiteral); ok {
				stored = true
				if len(literal.GraphTypes) == 0 || !program.GraphTypes[literal.GraphTypes[len(literal.GraphTypes)-1]] {
					test.Fatal("slot initializer has no graph flow identity")
				}
			}
		}
		return true
	})
	if !stored {
		test.Fatal("missing inline slot initializer")
	}
}
