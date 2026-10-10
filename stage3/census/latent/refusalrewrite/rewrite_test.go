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
)

func TestCompiler41231d51Shape(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/refusals-41231d51.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	output, err := Rewrite(source)
	if err != nil {
		t.Fatal(err)
	}
	before, err := parser.ParseFile(token.NewFileSet(), "before.go", source, 0)
	if err != nil {
		t.Fatal(err)
	}
	after, err := parser.ParseFile(token.NewFileSet(), "after.go", output, 0)
	if err != nil {
		t.Fatal(err)
	}
	find := func(file *ast.File, name string) *ast.FuncDecl {
		for _, d := range file.Decls {
			if f, ok := d.(*ast.FuncDecl); ok && f.Name.Name == name {
				return f
			}
		}
		t.Fatalf("missing %s", name)
		return nil
	}
	if printed(find(before, "refuse")) != printed(find(after, "refuse")) {
		t.Fatal("production refuse changed")
	}
	latent := find(after, "latentRefuse")
	count := 0
	ast.Inspect(latent.Body, func(n ast.Node) bool {
		if _, ok := n.(*ast.DeferStmt); ok {
			count++
		}
		return true
	})
	if count != 4 {
		t.Fatalf("expected independently instrumented contracts and visit functions with scoped owners, got %d", count)
	}
}

func TestMissingFunctionMutantFailsLoudly(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/refusals-41231d51.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(string(source), "func (l *lowering) refuse(", "func (l *lowering) missingRefuse(", 1)
	output, err := Rewrite([]byte(mutant))
	if err == nil || !strings.Contains(err.Error(), "lowering.refuse: expected exactly one function, found 0") {
		t.Fatalf("missing-function mutant not caught: output=%d error=%v", len(output), err)
	}
	if output != nil {
		t.Fatal("failed rewrite returned partial source")
	}
}

func TestChangedVisitorMutantFailsLoudly(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/refusals-41231d51.go.txt")
	if err != nil {
		t.Fatal(err)
	}
	mutant := strings.Replace(string(source), "node.ForEachChild(visit)", "node.ForEachChild(other)", 1)
	_, err = Rewrite([]byte(mutant))
	if err == nil || !strings.Contains(err.Error(), "lowering.refuse: visit expected one child walk, found 0") {
		t.Fatalf("changed visitor mutant not caught: %v", err)
	}
}

func TestVisitorsCollectContinueAndSkipDiagnosedBodies(t *testing.T) {
	t.Parallel()
	probe := behaviorProbe(t)
	log, err := os.Create(filepath.Join(t.TempDir(), "test.log"))
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(probe, "-test.count=1")
	cmd.Dir = filepath.Dir(probe)
	cmd.Stdout = log
	cmd.Stderr = log
	err = cmd.Run()
	_ = log.Close()
	if err != nil {
		contents, _ := os.ReadFile(log.Name())
		t.Fatalf("compiled behavior probe: %v\n%s", err, contents)
	}
}

// behaviorSource is a refuse with every shape the latent rewrite instruments: directives, pragmas, a contract
// visitor and a syntax visitor that stop at diagnosed bodies.
const behaviorSource = `package lower
import "probe/ast"
func (l *lowering) refuse(module *ast.SourceFile) error {
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

// behaviorAST is the stub of the ast package Rewrite's output compiles against.
const behaviorAST = `package ast
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
 if err:=l.refuse(module);err==nil||err.Error()!="directive-one"{t.Fatalf("production first refusal: %v",err)}
 if err:=l.latentRefuse(module);err!=nil{t.Fatal(err)}
 want:=[]string{"directive-one","directive-two","pragma","contract:parent","syntax:parent","syntax:child","syntax:sibling"}
 if !reflect.DeepEqual(observations,want){t.Fatalf("findings=%v want=%v",observations,want)}
}
`

// behaviorProbe is the test binary of a module holding Rewrite's output for behaviorSource beside behaviorAST and
// behaviorTest: a product built ahead (go test -c with GOWORK off, which vets it as go test did), so the unit runs it
// without go. Rewrite runs in this process, so its source and this file key it.
func behaviorProbe(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:      "refusalrewrite behavior probe",
		Files:     []string{"stage3/census/latent/refusalrewrite/rewrite.go", "stage3/census/latent/refusalrewrite/rewrite_test.go"},
		Flags:     []string{"go test -c -trimpath -ldflags=-buildid= -buildvcs=false", "GOWORK=off"},
		Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("go", "env", "GOOS", "GOARCH", "CGO_ENABLED", "CC", "GOFLAGS", "GOEXPERIMENT")},
	}
	directory := buildcache.Product(t, inputs, func(directory string) error {
		output, err := Rewrite([]byte(behaviorSource))
		if err != nil {
			return err
		}
		module := filepath.Join(directory, "module")
		for name, text := range map[string]string{"go.mod": "module probe\n\ngo 1.21\n", "refusals.go": string(output), "ast/ast.go": behaviorAST, "behavior_test.go": behaviorTest} {
			path := filepath.Join(module, name)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
				return err
			}
		}
		command := exec.Command("go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-buildvcs=false", "-o", filepath.Join(directory, "probe.test"), ".")
		command.Dir = module
		command.Env = append(os.Environ(), "GOWORK=off")
		if combined, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("go test -c of the behavior probe: %v\n%s", err, combined)
		}
		return nil
	})
	return filepath.Join(directory, "probe.test")
}

func TestProduct_RefusalRewriteBehaviorProbe(t *testing.T) {
	t.Parallel()
	behaviorProbe(t)
}
