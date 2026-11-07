package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"reflect"
	"testing"
)

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_finally_completion.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_labels.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_labels_finally.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_labels_block_finally.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_edges.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_iterables_cells_labels.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_break_sync.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_continue_sync.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_switch_break_sync.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_outer_continue_sync.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_break_async.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_continue_async.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_switch_break_async.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/label_block_outer_continue_async.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_finally_return.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_finally_throw.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_finally_break.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_finally_continue.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_finally_void_return.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_finally_snapshot.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_gap_label_admission.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_gap_catch_admission.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_gap_finally_admission.a", true, false})
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{"internal/oracle/testdata/async_gap_finally_return_admission.a", true, false})
}
func TestAsyncLabelsAsyncFinallyCompletion(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_finally_completion", "")
}
func TestAsyncLabelsAsyncLabels(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_labels", "")
}
func TestAsyncLabelsAsyncLabelsFinally(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_labels_finally", "")
}
func TestAsyncLabelsAsyncLabelsBlockFinally(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_labels_block_finally", "")
}
func TestAsyncLabelsLabelBlockEdges(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_edges", "")
}
func TestAsyncLabelsAsyncIterablesCellsLabels(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_iterables_cells_labels", "")
}
func TestAsyncLabelsLabelBlockBreakSync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_break_sync", "")
}
func TestAsyncLabelsLabelBlockContinueSync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_continue_sync", "")
}
func TestAsyncLabelsLabelBlockSwitchBreakSync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_switch_break_sync", "")
}
func TestAsyncLabelsLabelBlockOuterContinueSync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_outer_continue_sync", "")
}
func TestAsyncLabelsLabelBlockBreakAsync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_break_async", "")
}
func TestAsyncLabelsLabelBlockContinueAsync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_continue_async", "")
}
func TestAsyncLabelsLabelBlockSwitchBreakAsync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_switch_break_async", "")
}
func TestAsyncLabelsLabelBlockOuterContinueAsync(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "label_block_outer_continue_async", "")
}
func TestAsyncLabelsAsyncFinallyReturn(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_finally_return", "")
}
func TestAsyncLabelsAsyncFinallyThrow(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_finally_throw", "")
}
func TestAsyncLabelsAsyncFinallyBreak(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_finally_break", "")
}
func TestAsyncLabelsAsyncFinallyContinue(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_finally_continue", "")
}
func TestAsyncLabelsAsyncFinallyVoidReturn(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_finally_void_return", "")
}
func TestAsyncLabelsAsyncFinallySnapshot(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_finally_snapshot", "")
}
func TestAsyncLabelsAsyncGapLabelAdmission(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_gap_label_admission", "")
}
func TestAsyncLabelsAsyncGapCatchAdmission(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_gap_catch_admission", "")
}
func TestAsyncLabelsAsyncGapFinallyAdmission(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_gap_finally_admission", "")
}
func TestAsyncLabelsAsyncGapFinallyReturnAdmission(test *testing.T) {
	test.Parallel()
	asyncLoopWitness(test, "async_gap_finally_return_admission", "")
}

// These mutants preserve valid IR, clean C builds and balanced ownership. Only
// the source oracle may catch the lost completion, snapshot or destination.
func asyncLabelsMutation(test *testing.T, name, kind string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	p, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	changed := false
	label := ""
	original := -1
	if kind == "snapshot" {
		for i, l := range p.Locals {
			if l.Name == "value" {
				original = i
			}
		}
		if original < 0 {
			test.Fatal("snapshot source not found")
		}
	}
	var rewrite func(reflect.Value) reflect.Value
	rewrite = func(v reflect.Value) reflect.Value {
		if v.Kind() == reflect.Interface {
			out := reflect.New(v.Type()).Elem()
			if !v.IsNil() {
				out.Set(rewrite(v.Elem()))
			}
			return out
		}
		if !changed {
			switch node := v.Interface().(type) {
			case ir.WriteLine:
				if kind == "finally" {
					changed = true
					return reflect.ValueOf(ir.Block{})
				}
			case ir.Read:
				if kind == "snapshot" && p.Locals[node.Local].Name == "pending return" {
					changed = true
					return reflect.ValueOf(ir.Read{Local: original, Of: node.Of})
				}
			case ir.Break:
				if kind == "block" && label != "" && node.Label == "" && node.Depth == 0 {
					changed = true
					node.Label = label
					return reflect.ValueOf(node)
				}
			case ir.Continue:
				if kind == "name" && node.Label != "" {
					changed = true
					node.Label = ""
					return reflect.ValueOf(node)
				}
			}
		}
		previous := label
		if node, ok := v.Interface().(ir.Labeled); ok {
			label = node.Name
		}
		defer func() { label = previous }()
		switch v.Kind() {
		case reflect.Struct:
			out := reflect.New(v.Type()).Elem()
			for i := 0; i < v.NumField(); i++ {
				out.Field(i).Set(rewrite(v.Field(i)))
			}
			return out
		case reflect.Slice:
			if v.IsNil() {
				return v
			}
			out := reflect.MakeSlice(v.Type(), v.Len(), v.Len())
			for i := 0; i < v.Len(); i++ {
				out.Index(i).Set(rewrite(v.Index(i)))
			}
			return out
		}
		return v
	}
	for i, f := range p.Functions {
		if kind == "finally" && f.Name != "work" {
			continue
		}
		p.Functions[i].Body = rewrite(reflect.ValueOf(f.Body)).Interface().([]ir.Statement)
	}
	if !changed {
		test.Fatal("mutant changed no completion")
	}
	code := native.C(p)
	binary := filepath.Join(test.TempDir(), "wrong-completion")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatal(err)
	}
	if report := leakcheck.Report(test, code, binary); report != "" {
		test.Fatal(report)
	}
	for backend, result := range map[string]run{
		"native":     execute(test, binary),
		"JavaScript": onJavaScriptBackend(test, p),
	} {
		if disagreement(truth, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			test.Fatalf("%s: only Node stdout must catch %s: %+v", backend, kind, result)
		}
	}
	test.Logf("Node stdout catches %s in native and JavaScript", kind)
}
func TestAsyncLabelsFinallyReturnMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "async_finally_return", "finally")
}
func TestAsyncLabelsFinallyThrowMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "async_finally_throw", "finally")
}
func TestAsyncLabelsFinallyBreakMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "async_finally_break", "finally")
}
func TestAsyncLabelsFinallyContinueMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "async_finally_continue", "finally")
}
func TestAsyncLabelsReturnSnapshotMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "async_finally_snapshot", "snapshot")
}
func TestAsyncLabelsNameResolutionMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "async_labels", "name")
}
func TestAsyncLabelsBlockWrapperBreakSyncMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "label_block_break_sync", "block")
}
func TestAsyncLabelsBlockWrapperBreakAsyncMutant(test *testing.T) {
	test.Parallel()
	asyncLabelsMutation(test, "label_block_break_async", "block")
}
