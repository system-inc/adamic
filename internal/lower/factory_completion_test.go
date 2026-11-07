package lower

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/load"
)

func TestParserFactoryCompletionBeforeReturn(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; node.text='ready'; return node;`,
		true,
		false,
	)
}

func TestParserFactoryCompletionAlias(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; const alias=node; alias.text='ready'; return node;`,
		true,
		false,
	)
}

func TestParserFactoryCompletionBothBranches(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; if(flag){node.text='yes';}else{node.text='no';} return node;`,
		true,
		false,
	)
}

func TestParserFactoryCompletionOneBranch(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; if(flag){node.text='yes';} return node;`,
		false,
		false,
	)
}

func TestParserFactoryCompletionBeforeEscape(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; expose(node); node.text='late'; return node;`,
		false,
		true,
	)
}

func TestParserFactoryCompletionWrongPayloadAssertion(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; node.text=42 as unknown as string; return node;`,
		false,
		false,
	)
}

func TestParserFactoryCompletionEarlyReturn(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; if(flag){return node;} node.text='late'; return node;`,
		false,
		false,
	)
}

func TestParserFactoryCompletionLoopMayNotRun(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; while(flag){node.text='late';} return node;`,
		false,
		true,
	)
}

func TestParserFactoryCompletionCaptured(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; const callback=()=>node; node.text='late'; return node;`,
		false,
		true,
	)
}

func TestParserFactoryCompletionShortCircuit(test *testing.T) {
	test.Parallel()
	assertParserFactoryCompletion(
		test,
		`const node = {kind:1,pos:-1,end:-1} as Target; flag && (node.text='late'); return node;`,
		false,
		true,
	)
}

func assertParserFactoryCompletion(test *testing.T, body string, proven, unknown bool) {
	test.Helper()
	path := filepath.Join(test.TempDir(), "factory.a")
	source := `interface Target {kind:number;pos:number;end:number;text:string;} function expose(value:Target):void{} function factory(flag:boolean):Target {` +
		body +
		`}`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	p, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	c, release := p.Checker(context.Background(), p.Files()[0])
	defer release()
	var factory *ast.Node
	p.Files()[0].AsNode().ForEachChild(func(n *ast.Node) bool {
		if n.Kind == ast.KindFunctionDeclaration && n.Name().Text() == "factory" {
			factory = n
		}
		return false
	})
	if factory == nil {
		test.Fatal("factory missing")
	}
	summary := AnalyzeFactory(factory, c.GetTypeAtLocation(factory.Type()), c)
	if summary.Proven("text") != proven || summary.Unknown != unknown {
		test.Fatalf("summary %+v, want text proven %v, unknown %v", summary, proven, unknown)
	}
	if !summary.Unknown && len(summary.Required) != 4 {
		test.Fatalf("required %+v", summary.Required)
	}
}

func TestParserFactoryReadBeforeCompletion(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "read.a")
	source := `interface Target {text:string;} function factory():Target {const node={} as Target; const earlier=node.text; node.text='ready'; return node;}`
	if writeError := os.WriteFile(path, []byte(source), 0644); writeError != nil {
		test.Fatal(writeError)
	}
	p, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	c, release := p.Checker(context.Background(), p.Files()[0])
	defer release()
	var f *ast.Node
	p.Files()[0].AsNode().ForEachChild(func(n *ast.Node) bool {
		if n.Kind == ast.KindFunctionDeclaration {
			f = n
		}
		return false
	})
	summary := AnalyzeFactory(f, c.GetTypeAtLocation(f.Type()), c)
	if !summary.Proven("text") || len(summary.Reads) != 1 || !summary.Reads[0].Checked {
		test.Fatalf("later completion must not erase earlier read: %+v", summary)
	}
}
