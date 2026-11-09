package lower

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const statsOptionalReceiverSource = "import {statSync} from 'node:fs';\nconst stat = statSync('missing', {throwIfNoEntry: false});\nconsole.log(`${stat?.isFile()}`);\n"
const statsOptionalReceiverCatch = "unguarded nodeFSFile must refuse node:fs.StatsBase.isFile through an optional call"

func TestNodeFSStatsOptionalReceiverRequiresGuard(t *testing.T) {
	t.Parallel()
	program, err := load.Load([]string{writeAgreementSource(t, "main.ts", statsOptionalReceiverSource)})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	func() {
		checked, release := program.Checker(context.Background(), file)
		defer release()
		l := &lowering{program: program, checker: checked, result: &ir.Program{Locals: []ir.Local{{Name: "stat", Type: ir.Object, Global: true, Function: -1}}}, this: -1, functionIndex: -1}
		calls := 0
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindCallExpression {
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)
				if callee.Kind == ast.KindPropertyAccessExpression && callee.AsPropertyAccessExpression().QuestionDotToken != nil {
					calls++
					l.locals = map[*ast.Symbol]int{l.symbol(callee.AsPropertyAccessExpression().Expression): 0}
					value, handled, err := l.nodeFSFile(node)
					var refused *NotYet
					if !handled || !errors.As(err, &refused) || refused.What != "node:fs.StatsBase.isFile through an optional call" {
						t.Fatalf("%s: handled=%t value=%T err=%v", statsOptionalReceiverCatch, handled, value, err)
					}
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
		if calls != 1 {
			t.Fatalf("want one optional Stats call, got %d", calls)
		}
	}()
	if _, err := Lower(context.Background(), program); err != nil {
		t.Fatalf("guarded full Lower: %v", err)
	}
}

func TestNodeFSStatsOptionalReceiverFullLower(t *testing.T) {
	t.Parallel()
	if _, err := lowerSourceAs(t, statsOptionalReceiverSource, ".ts"); err != nil {
		t.Fatal(err)
	}
}

func statsOptionalReceiverMarkMutants(t *testing.T) (string, string) {
	t.Helper()
	mutations := []buildcache.Mutation{{File: "internal/lower/library_node_fs_file.go", Before: "(!receiverOptional || l.optionalReceiverCall != node)", After: "!receiverOptional"}}
	lower := buildcache.GoTestMutated(t, "", "stats-optional-mark-lower.test", "./internal/lower", nil, mutations)
	oracle := buildcache.GoTestMutated(t, "", "stats-optional-mark-oracle.test", "./internal/oracle", nil, mutations)
	return lower, oracle
}

func TestProduct_StatsOptionalReceiverMarkMutant(t *testing.T) {
	t.Parallel()
	statsOptionalReceiverMarkMutants(t)
}

func TestNodeFSStatsOptionalReceiverMarkMutantIsCaught(t *testing.T) {
	t.Parallel()
	lower, oracle := statsOptionalReceiverMarkMutants(t)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	stdout, stderr, err := childguard.Capture(ctx, exec.CommandContext(ctx, lower, "-test.run", "^TestNodeFSStatsOptionalReceiverRequiresGuard$", "-test.count=1", "-test.v"))
	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), statsOptionalReceiverCatch) {
		t.Fatalf("mark mutant must fail direct guard test: err=%v stdout=%s stderr=%s", err, stdout, stderr)
	}
	t.Logf("mark mutant caught: %s", stdout)
	controls := []struct{ binary, selection, pass string }{
		{lower, "^TestNodeFSStatsOptionalReceiverFullLower$", "--- PASS: TestNodeFSStatsOptionalReceiverFullLower"},
		{oracle, "^TestInputAgreesWithNode$/internal/oracle/testdata/node_fs_stats_optional_call.a$", "--- PASS: TestInputAgreesWithNode/internal/oracle/testdata/node_fs_stats_optional_call.a"},
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		controls = append(controls, struct{ binary, selection, pass string }{oracle, "^TestNodeFSDirectoryWASIAgreement$/node_fs_stats_optional_call$", "--- PASS: TestNodeFSDirectoryWASIAgreement/node_fs_stats_optional_call"})
	}
	for _, control := range controls {
		stdout, stderr, err := childguard.Capture(ctx, exec.CommandContext(ctx, control.binary, "-test.run", control.selection, "-test.count=1", "-test.v"))
		if err != nil || !strings.Contains(string(stdout), control.pass) || strings.Contains(string(stdout), "--- SKIP:") {
			t.Fatalf("mark mutant control %s: err=%v stdout=%s stderr=%s", control.selection, err, stdout, stderr)
		}
		t.Logf("mark mutant control green: %s", stdout)
	}
}
