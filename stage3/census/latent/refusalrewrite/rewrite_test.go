package refusalrewrite

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

func TestCompiler41231d51Shape(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/refusals-41231d51.go.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	output, rewriteError := Rewrite(source)
	if rewriteError != nil {
		test.Fatal(rewriteError)
	}
	before, parseFileError := parser.ParseFile(token.NewFileSet(), "before.go", source, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	after, parseFileError := parser.ParseFile(token.NewFileSet(), "after.go", output, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	find := func(file *ast.File, name string) *ast.FuncDecl {
		for _, d := range file.Decls {
			if funcDecl, ok := d.(*ast.FuncDecl); ok && funcDecl.Name.Name == name {
				return funcDecl
			}
		}
		test.Fatalf("missing %s", name)
		return nil
	}
	if printed(find(before, "refuse")) != printed(find(after, "refuse")) {
		test.Fatal("production refuse changed")
	}
	latent := find(after, "latentRefuse")
	count := 0
	ast.Inspect(latent.Body, func(node ast.Node) bool {
		if _, ok := node.(*ast.DeferStmt); ok {
			count++
		}
		return true
	})
	if count != 4 {
		test.Fatalf(
			"expected independently instrumented contracts and visit functions with scoped owners, got %d",
			count,
		)
	}
}

func TestMissingFunctionMutantFailsLoudly(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/refusals-41231d51.go.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	mutant := strings.Replace(string(source), "func (l *lowering) refuse(", "func (l *lowering) missingRefuse(", 1)
	output, rewriteError := Rewrite([]byte(mutant))
	if rewriteError == nil ||
		!strings.Contains(rewriteError.Error(), "lowering.refuse: expected exactly one function, found 0") {
		test.Fatalf("missing-function mutant not caught: output=%d error=%v", len(output), rewriteError)
	}
	if output != nil {
		test.Fatal("failed rewrite returned partial source")
	}
}

func TestChangedVisitorMutantFailsLoudly(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/refusals-41231d51.go.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	mutant := strings.Replace(string(source), "node.ForEachChild(visit)", "node.ForEachChild(other)", 1)
	_, rewriteError := Rewrite([]byte(mutant))
	if rewriteError == nil ||
		!strings.Contains(rewriteError.Error(), "lowering.refuse: visit expected one child walk, found 0") {
		test.Fatalf("changed visitor mutant not caught: %v", rewriteError)
	}
}

func TestVisitorsCollectContinueAndSkipDiagnosedBodies(test *testing.T) {
	test.Parallel()
	probe := behaviorProbe(test)
	log, createError := os.Create(filepath.Join(test.TempDir(), "test.log"))
	if createError != nil {
		test.Fatal(createError)
	}
	command := exec.Command(probe, "-test.count=1")
	command.Dir = filepath.Dir(probe)
	command.Stdout = log
	command.Stderr = log
	runError := command.Run()
	_ = log.Close()
	if runError != nil {
		contents, _ := os.ReadFile(log.Name())
		test.Fatalf("compiled behavior probe: %v\n%s", runError, contents)
	}
}

// behaviorSource is a refuse with every shape the latent rewrite instruments: directives, pragmas, a contract
// visitor and a syntax visitor that stop at diagnosed bodies.
const behaviorSource = `package lower
import "probe/ast"
func (l *lowering) refuse(module *ast.SourceFile) error {
 if err := l.referenceRefusal(module); err != nil { return err }
 if err := l.parallelPreflight(module); err != nil { return err }
 if err := l.futurePreflight(module); err != nil { return err }
 if len(module.CommentDirectives)>0 {
  directive:=module.CommentDirectives[0]
  return &Refused{What:directive.Name}
 }
 for _,pragma:=range module.Pragmas {if pragma.Name!="" {return &Refused{What:pragma.Name}}}
 var contractError error
 var contracts ast.Visitor
 contracts=func(node *ast.Node) bool {
  if contractError!=nil{return true}
  if node.Kind==ast.KindCallExpression {contractError=l.predicateArguments(node)}
  if contractError==nil {node.ForEachChild(contracts)}
  return contractError!=nil
 }
 module.AsNode().ForEachChild(contracts)
 if contractError!=nil{return contractError}
 var found error
 var visit ast.Visitor
 visit=func(node *ast.Node) bool {
  if found!=nil{return true}
  if err:=l.nodeRefusal(node);err!=nil{found=err;return true}
  node.ForEachChild(visit)
  return false
 }
 module.AsNode().ForEachChild(visit)
 return found
}
`

// behaviorAst is the stub of the ast package Rewrite's output compiles against.
const behaviorAst = `package ast
const (KindSourceFile=1;KindFunctionDeclaration=2;KindCallExpression=3)
type Visitor func(*Node) bool
type Node struct {Kind int;Parent *Node;Children []*Node;Label string;Contract,Syntax,Diagnosed bool}
func(n *Node) Body()*Node{return n}
func(n *Node) ForEachChild(visit Visitor){for _,child:=range n.Children{if visit(child){return}}}
type Directive struct{Name string}
type SourceFile struct{CommentDirectives []Directive;Pragmas []Directive;Root *Node}
func(s *SourceFile) AsNode()*Node{return s.Root}
`

// behaviorTest runs the rewritten refuse and latentRefuse over a tree with a diagnosed body.
const behaviorTest = `package lower
import("fmt";"reflect";"testing";"probe/ast")
type Refused struct{What string}
func(r *Refused) Error()string{return r.What}
type program struct{}
func(*program) Where(n *ast.Node)string{return n.Label}
func(*program) LatentDiagnosticsIn(n *ast.Node)[]string{if n.Diagnosed{return []string{"bad body"}};return nil}
type lowering struct{program *program}
func(*lowering) referenceRefusal(*ast.SourceFile)error{return fmt.Errorf("reference")}
func(*lowering) parallelPreflight(*ast.SourceFile)error{return fmt.Errorf("parallel")}
func(*lowering) futurePreflight(*ast.SourceFile)error{return fmt.Errorf("future")}
func(*lowering) predicateArguments(n *ast.Node)error{if n.Diagnosed{panic("contract visitor entered diagnosed body")};if n.Contract{return fmt.Errorf("contract:%s",n.Label)};return nil}
func(*lowering) nodeRefusal(n *ast.Node)error{if n.Diagnosed{panic("syntax visitor entered diagnosed body")};if n.Syntax{return fmt.Errorf("syntax:%s",n.Label)};return nil}
var observations []string
var latentFindingOwner string
func(l *lowering) latentBodyDiagnostics(n *ast.Node)[]string{return l.program.LatentDiagnosticsIn(n)}
func latentFullEnabled()bool{return false}
func latentNestedDeclarations(body *ast.Node,visit ast.Visitor){}
func latentRecord(err error){if err!=nil{observations=append(observations,err.Error())}}
func TestBehavior(t *testing.T){
 root:=&ast.Node{Kind:ast.KindSourceFile}
 parent:=&ast.Node{Kind:ast.KindCallExpression,Parent:root,Label:"parent",Contract:true,Syntax:true}
 child:=&ast.Node{Parent:parent,Label:"child",Syntax:true};parent.Children=[]*ast.Node{child}
 sibling:=&ast.Node{Parent:root,Label:"sibling",Syntax:true}
 skipped:=&ast.Node{Kind:ast.KindFunctionDeclaration,Parent:root,Label:"skipped",Contract:true,Syntax:true,Diagnosed:true}
 root.Children=[]*ast.Node{parent,sibling,skipped}
 module:=&ast.SourceFile{Root:root,CommentDirectives:[]ast.Directive{{Name:"directive-one"},{Name:"directive-two"}},Pragmas:[]ast.Directive{{Name:"pragma"}}}
 l:=&lowering{program:&program{}}
 if err:=l.refuse(module);err==nil||err.Error()!="reference"{t.Fatalf("production first refusal: %v",err)}
 if err:=l.latentRefuse(module);err!=nil{t.Fatal(err)}
 want:=[]string{"reference","parallel","future","directive-one","directive-two","pragma","contract:parent","syntax:parent","syntax:child","syntax:sibling"}
 if !reflect.DeepEqual(observations,want){t.Fatalf("findings=%v want=%v",observations,want)}
}
`

// probeToolchain is the probe's go, asked as its build runs go: toolchain's, under toolchain's environment with GOWORK
// off, at the repository, whose go and toolchain directives the probe's module declares, so GOTOOLCHAIN selects the
// same release for the key as for the build, and a go env file (which toolchain's GOENV=off never reads) reaches
// neither.
func probeToolchain(test *testing.T) []string {
	test.Helper()
	var reports []string
	for _, arguments := range [][]string{
		{"version"},
		{"env", "GOOS", "GOARCH", "CGO_ENABLED", "CC", "GOFLAGS", "GOEXPERIMENT"},
	} {
		command := toolchain.CommandIn("../../../..", "go", arguments...)
		command.Env = append(command.Env, "GOWORK=off")
		output, outputError := command.Output()
		if outputError != nil {
			test.Fatalf("go %s: %v", strings.Join(arguments, " "), outputError)
		}
		reports = append(reports, strings.TrimSpace(string(output)))
	}
	return reports
}

// behaviorProbe is the test binary of a module holding Rewrite's output for behaviorSource beside behaviorAST and
// behaviorTest: a product built ahead (go test -c with GOWORK off, which vets it as go test did), so the unit runs it
// without go. Rewrite runs in this process, so its source and this file key it.
func behaviorProbe(test *testing.T) string {
	test.Helper()
	inputs := buildcache.Inputs{
		Name: "refusalrewrite behavior probe",
		Files: []string{
			"go.mod",
			"stage3/census/latent/refusalrewrite/rewrite.go",
			"stage3/census/latent/refusalrewrite/rewrite_test.go",
		},
		Flags:     []string{"go test -c -trimpath -ldflags=-buildid= -buildvcs=false", "GOWORK=off"},
		Toolchain: probeToolchain(test),
	}
	directory := buildcache.Product(test, inputs, func(directory string) error {
		output, rewriteError := Rewrite([]byte(behaviorSource))
		if rewriteError != nil {
			return rewriteError
		}
		// The probe declares the repository's go version, so GOTOOLCHAIN selects for its build the release the key's
		// go version, asked at the repository, names.
		repository, readError := os.ReadFile("../../../../go.mod")
		if readError != nil {
			return readError
		}
		directives := ""
		for _, line := range strings.Split(string(repository), "\n") {
			if strings.HasPrefix(line, "go ") || strings.HasPrefix(line, "toolchain ") {
				directives += strings.TrimSpace(line) + "\n"
			}
		}
		if !strings.HasPrefix(directives, "go ") {
			return fmt.Errorf("the repository's go.mod declares no go version")
		}
		module := filepath.Join(directory, "module")
		for name, text := range map[string]string{
			"go.mod":           "module probe\n\n" + directives,
			"refusals.go":      string(output),
			"ast/ast.go":       behaviorAst,
			"behavior_test.go": behaviorTest,
		} {
			path := filepath.Join(module, name)
			if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
				return createDirectoryError
			}
			if writeError := os.WriteFile(path, []byte(text), 0o644); writeError != nil {
				return writeError
			}
		}
		command := toolchain.CommandIn(
			module,
			"go",
			"test",
			"-c",
			"-trimpath",
			"-ldflags=-buildid=",
			"-buildvcs=false",
			"-o",
			filepath.Join(directory, "probe.test"),
			".",
		)
		command.Env = append(command.Env, "GOWORK=off")
		if combined, combinedOutputError := command.CombinedOutput(); combinedOutputError != nil {
			return fmt.Errorf("go test -c of the behavior probe: %v\n%s", combinedOutputError, combined)
		}
		return nil
	})
	return filepath.Join(directory, "probe.test")
}

func TestProduct_RefusalRewriteBehaviorProbe(test *testing.T) {
	test.Parallel()
	behaviorProbe(test)
}

func TestModulePreflightShapes(test *testing.T) {
	test.Parallel()
	for _, preflight := range []string{"referenceRefusal", "parallelPreflight", "unknownPreflight"} {
		input := strings.Replace(
			behaviorSource,
			"if len(module.CommentDirectives)",
			"if err := l."+preflight+"(module); err != nil { return err }; if len(module.CommentDirectives)",
			1,
		)
		output, rewriteError := Rewrite([]byte(input))
		if rewriteError != nil {
			test.Fatal(rewriteError)
		}
		clone := string(output)[strings.Index(string(output), "func (l *lowering) latentRefuse"):]
		if !strings.Contains(clone, "latentRecord(err)") {
			test.Fatal("preflight error not recorded")
		}
	}
}

func TestUnsupportedPreflightReturns(test *testing.T) {
	test.Parallel()
	for _, statement := range []string{
		`if err := l.unknownPreflight(module); err != nil { return fmt.Errorf("wrapped: %w", err) }`,
		`if module != nil { if err := l.unknownPreflight(module); err != nil { return err } }`,
		`if err = l.unknownPreflight(module); err != nil { return err }`,
		`if err := l.unknownPreflight(other); err != nil { return err }`,
	} {
		input := strings.Replace(
			behaviorSource,
			"if len(module.CommentDirectives)",
			statement+"; if len(module.CommentDirectives)",
			1,
		)
		output, rewriteError := Rewrite([]byte(input))
		if rewriteError == nil || output != nil {
			test.Fatalf("unsupported return published: %s", statement)
		}
	}
}

func TestCompilerSlice3Shape(test *testing.T) {
	test.Parallel()
	source, readError := os.ReadFile("testdata/refusals-slice3.go.txt")
	if readError != nil {
		test.Fatal(readError)
	}
	if _, rewriteError := Rewrite(source); rewriteError != nil {
		test.Fatal(rewriteError)
	}
}

// Main's own lowering.refuse rewrites and keeps every line make_overlay.py counts on afterwards, so a reshape in
// internal/lower reds here, at its author's change, not at the next meter run.
func TestRefusalRewriteReadsMainRefusals(test *testing.T) {
	test.Parallel()
	// Reads: internal/lower/refusals.go
	source, readError := os.ReadFile("../../../../internal/lower/refusals.go")
	if readError != nil {
		test.Fatal(readError)
	}
	output, rewriteError := Rewrite(source)
	if rewriteError != nil {
		test.Fatalf("main's lowering.refuse doesn't rewrite: %v", rewriteError)
	}
	for _, binding := range []string{"found", "contractError"} {
		if count := strings.Count(string(output), "\t\t\tlatentRecord("+binding+")\n"); count != 1 {
			test.Errorf("make_overlay.py's refusal visitor identity for %s: found %d, want 1", binding, count)
		}
	}
	for _, visitor := range []string{"contracts", "visit"} {
		if count := strings.Count(string(output), visitor+" = func(node *ast.Node) (latentStop bool) {"); count != 1 {
			test.Errorf("make_overlay.py's %s visitor entry: found %d, want 1", visitor, count)
		}
	}
}
