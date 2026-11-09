package lower

import (
	"context"
	"errors"
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

func statsOptionalReceiverMarkMutant(t *testing.T) string {
	t.Helper()
	mutations := []buildcache.Mutation{{File: "internal/lower/library_node_fs_file.go", Before: "(!receiverOptional || l.optionalReceiverCall != node)", After: "!receiverOptional"}}
	return buildcache.GoTestMutated(t, "", "stats-optional-mark-lower.test", "./internal/lower", nil, mutations)
}

func TestProduct_StatsOptionalReceiverMarkMutant(t *testing.T) {
	t.Parallel()
	statsOptionalReceiverMarkMutant(t)
}

func TestNodeFSStatsOptionalReceiverMarkMutantIsCaught(t *testing.T) {
	t.Parallel()
	lower := statsOptionalReceiverMarkMutant(t)
	checkEntryReceiverMarkMutant(t, lower, "TestNodeFSStatsOptionalReceiverRequiresGuard", statsOptionalReceiverCatch)
}

func checkEntryReceiverMarkMutant(t *testing.T, lower, direct, catch string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Minute)
	defer cancel()
	stdout, stderr, err := childguard.Capture(ctx, exec.CommandContext(ctx, lower, "-test.run", "^"+direct+"$", "-test.count=1", "-test.v"))
	var exited *exec.ExitError
	if !errors.As(err, &exited) || exited.ExitCode() != 1 || !strings.Contains(string(stdout), catch) {
		t.Fatalf("mark mutant must fail direct guard test: err=%v stdout=%s stderr=%s", err, stdout, stderr)
	}
	t.Logf("mark mutant caught: %s", stdout)
}
