package lint

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

type mutantRuleOwner struct{ name, slug string }

// Read the actual gate-visible functions and their helper arguments. A second
// maintained list could agree with itself while a real wrapper dropped a rule.
func mutantRuleDeclarations(t *testing.T, prefix, helper string) []mutantRuleOwner {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), "mutant_rule_shards_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	var owners []mutantRuleOwner
	for _, decl := range file.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || !strings.HasPrefix(fn.Name.Name, prefix) {
			continue
		}
		if fn.Body == nil || len(fn.Body.List) != 2 {
			t.Fatalf("%s must parallelize then call its shared helper", fn.Name.Name)
		}
		first, ok := fn.Body.List[0].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("%s does not parallelize first", fn.Name.Name)
		}
		parallel, ok := first.X.(*ast.CallExpr)
		if !ok {
			t.Fatalf("%s does not parallelize first", fn.Name.Name)
		}
		selector, ok := parallel.Fun.(*ast.SelectorExpr)
		if !ok || selector.Sel.Name != "Parallel" || len(parallel.Args) != 0 {
			t.Fatalf("%s does not parallelize first", fn.Name.Name)
		}
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Name != "t" {
			t.Fatalf("%s parallelizes the wrong receiver", fn.Name.Name)
		}
		second, ok := fn.Body.List[1].(*ast.ExprStmt)
		if !ok {
			t.Fatalf("%s does not call its shared helper", fn.Name.Name)
		}
		call, ok := second.X.(*ast.CallExpr)
		if !ok || len(call.Args) != 2 {
			t.Fatalf("%s does not call its shared helper", fn.Name.Name)
		}
		callee, ok := call.Fun.(*ast.Ident)
		if !ok || callee.Name != helper {
			t.Fatalf("%s calls the wrong helper", fn.Name.Name)
		}
		argument, ok := call.Args[0].(*ast.Ident)
		if !ok || argument.Name != "t" {
			t.Fatalf("%s passes the wrong test", fn.Name.Name)
		}
		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			t.Fatalf("%s needs a literal rule slug", fn.Name.Name)
		}
		slug, err := strconv.Unquote(literal.Value)
		if err != nil {
			t.Fatal(err)
		}
		owners = append(owners, mutantRuleOwner{fn.Name.Name, slug})
	}
	return owners
}

func checkMutantRuleUnion(descriptors []registry.Descriptor, owners []mutantRuleOwner, prefix string) error {
	want := make(map[string]bool, len(descriptors))
	for _, d := range descriptors {
		if want[d.Slug] {
			return fmt.Errorf("duplicate registry mutant %s", d.Slug)
		}
		want[d.Slug] = true
	}
	seen := make(map[string]bool, len(owners))
	for _, owner := range owners {
		if !want[owner.slug] {
			return fmt.Errorf("unexpected mutant shard %s", owner.slug)
		}
		if seen[owner.slug] {
			return fmt.Errorf("duplicate mutant shard %s", owner.slug)
		}
		if owner.name != prefix+strings.ReplaceAll(owner.slug, "-", "_") {
			return fmt.Errorf("%s calls the wrong rule %s", owner.name, owner.slug)
		}
		seen[owner.slug] = true
	}
	for _, d := range descriptors {
		if !seen[d.Slug] {
			return fmt.Errorf("missing mutant shard %s", d.Slug)
		}
	}
	if len(owners) != len(descriptors) {
		return fmt.Errorf("mutant union has %d owners, want %d", len(owners), len(descriptors))
	}
	return nil
}

func TestMutantsUnion(t *testing.T) {
	t.Parallel()
	// Discover validates the owned mutant.json for every registry rule, including
	// the JSX canary previously skipped by the parent TestMutants.
	descriptors := prepareRegistry(t, packageDirectory)
	shards := mutantRuleDeclarations(t, "TestMutants_", "mutantRuleShard")
	products := mutantRuleDeclarations(t, "TestProduct_LintMutant_", "mutantRuleProduct")
	for _, group := range []struct {
		owners []mutantRuleOwner
		prefix string
	}{{shards, "TestMutants_"}, {products, "TestProduct_LintMutant_"}} {
		if err := checkMutantRuleUnion(descriptors, group.owners, group.prefix); err != nil {
			t.Fatal(err)
		}
	}
	if checkMutantRuleUnion(descriptors, shards[1:], "TestMutants_") == nil {
		t.Fatal("missing rule survived the union check")
	}
	repeated := append(append([]mutantRuleOwner(nil), shards...), shards[0])
	if checkMutantRuleUnion(descriptors, repeated, "TestMutants_") == nil {
		t.Fatal("duplicate rule survived the union check")
	}
	wrong := append([]mutantRuleOwner(nil), shards...)
	wrong[0].slug = "unregistered-planted-rule"
	if checkMutantRuleUnion(descriptors, wrong, "TestMutants_") == nil {
		t.Fatal("unexpected rule survived the union check")
	}
	wrong = append([]mutantRuleOwner(nil), shards...)
	wrong[0].name = "TestMutants_planted_wrong_wrapper"
	if checkMutantRuleUnion(descriptors, wrong, "TestMutants_") == nil {
		t.Fatal("wrong wrapper survived the union check")
	}
	t.Logf("mutant union: %d registry mutants exactly once in %d rule shards; %d matching product declarations", len(descriptors), len(shards), len(products))
}

func TestMutantsPlantedFailure(t *testing.T) {
	t.Parallel()
	const slug = "no-underscore-dangle"
	const shard = "TestMutants_no_underscore_dangle"
	descriptor := mutantRuleDescriptor(t, slug)
	change := mutantRuleMutation(t, descriptor)
	// Prepare the same declared products independently before the probe's work.
	lintGoOracleProduct(t)
	mutantRuleProduct(t, slug)
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, side := range []string{"Node", "emitted JavaScript"} {
		// No timeout context surrounds a setup child. Its test timeout and Loom's
		// whole-unit ceiling bound it; its own work deadline starts after fetching.
		command := exec.CommandContext(t.Context(), binary, "-test.run=^"+shard+"$", "-test.skip=^TestProduct_", "-test.timeout=90s", "-test.v")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if err == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return err
		}
		command.WaitDelay = time.Second
		command.Env = append(os.Environ(), "ADAMIC_MUTANT_SHARD_PROBE="+slug+":"+side)
		output, err := command.CombinedOutput()
		if err == nil {
			t.Fatalf("planted %s mutant survivor passed", side)
		}
		message := change.Name + " mutant survived on " + side
		if !bytes.Contains(output, []byte(message)) || bytes.Count(output, []byte("--- FAIL: "+shard+" ")) != 1 || bytes.Count(output, []byte("--- FAIL:")) != 1 {
			t.Fatalf("wrong failure for planted %s survivor: %v\n%s", side, err, output)
		}
		t.Logf("planted %s survivor caught by %s", side, shard)
	}
}
