package oracle

import (
	"path/filepath"
	"reflect"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func TestNodeFSStatsOptionalCallNullishMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_fs_stats_optional_call.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(t, how, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node failed: %+v", truth)
	}
	changed := 0
	mutateStringExpressions(reflect.ValueOf(program).Elem(), func(value ir.Expression) ir.Expression {
		effects, ok := value.(ir.Effects)
		if !ok || len(effects.Body) != 1 {
			return value
		}
		declare, ok := effects.Body[0].(ir.Declare)
		if !ok || program.Locals[declare.Local].Name != "optional_intrinsic_receiver" {
			return value
		}
		conditional, ok := effects.Result.(ir.Conditional)
		if !ok {
			return value
		}
		effects.Result = conditional.WhenTrue
		changed++
		return effects
	})
	if changed != 8 {
		t.Fatalf("want eight removed nullish guards, got %d", changed)
	}
	shared := sharedDirectory(t)
	bad, _ := inputNatively(t, how, program, shared)
	if disagreement(truth, bad) == "" {
		t.Fatal("native accepted missing-receiver mutant")
	}
	bad = inputBackend(t, how, program, shared)
	if disagreement(truth, bad) == "" {
		t.Fatal("JavaScript accepted missing-receiver mutant")
	}
	t.Log("removed nullish guards: undefined isFile row, stdout line 15, differs from Node on native and JavaScript")
}

func TestLibraryOptionalIntrinsicNullGuardMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/library_optional_intrinsic_null.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	how := inputRun{directory: filepath.Dir(path)}
	truth := onNodeWith(t, how, path)
	if truth.exitCode != 0 || len(truth.stderr) != 0 {
		t.Fatalf("Node failed: %+v", truth)
	}
	changed := 0
	mutateStringExpressions(reflect.ValueOf(program).Elem(), func(value ir.Expression) ir.Expression {
		effects, ok := value.(ir.Effects)
		if !ok || len(effects.Body) != 1 {
			return value
		}
		declare, ok := effects.Body[0].(ir.Declare)
		if !ok || program.Locals[declare.Local].Name != "optional_intrinsic_receiver" {
			return value
		}
		conditional, ok := effects.Result.(ir.Conditional)
		if !ok {
			return value
		}
		guard, ok := conditional.Condition.(ir.Unary)
		if !ok || guard.Operator != ir.Not {
			return value
		}
		nullish, ok := guard.Operand.(ir.Binary)
		if !ok || nullish.Operator != ir.Or {
			return value
		}
		undefined, ok := nullish.Left.(ir.IsUndefined)
		if !ok {
			return value
		}
		if _, ok := nullish.Right.(ir.IsNull); !ok {
			return value
		}
		guard.Operand = undefined
		conditional.Condition = guard
		effects.Result = conditional
		changed++
		return effects
	})
	if changed != 1 {
		t.Fatalf("want one undefined-only guard, got %d", changed)
	}
	bad := inputBackend(t, how, program, sharedDirectory(t))
	if disagreement(truth, bad) == "" {
		t.Fatal("JavaScript accepted undefined-only null receiver mutant")
	}
	t.Log("undefined-only guard: null pop row, stdout line 2, differs from Node on JavaScript; native already represents this null as missing")
}
