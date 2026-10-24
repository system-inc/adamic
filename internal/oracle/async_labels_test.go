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
func TestAsyncLabelsAsyncFinallyCompletion(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_finally_completion", "")
}
func TestAsyncLabelsAsyncLabels(t *testing.T) { t.Parallel(); asyncLoopWitness(t, "async_labels", "") }
func TestAsyncLabelsAsyncLabelsFinally(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_labels_finally", "")
}
func TestAsyncLabelsAsyncLabelsBlockFinally(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_labels_block_finally", "")
}
func TestAsyncLabelsLabelBlockEdges(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_edges", "")
}
func TestAsyncLabelsAsyncIterablesCellsLabels(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_iterables_cells_labels", "")
}
func TestAsyncLabelsLabelBlockBreakSync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_break_sync", "")
}
func TestAsyncLabelsLabelBlockContinueSync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_continue_sync", "")
}
func TestAsyncLabelsLabelBlockSwitchBreakSync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_switch_break_sync", "")
}
func TestAsyncLabelsLabelBlockOuterContinueSync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_outer_continue_sync", "")
}
func TestAsyncLabelsLabelBlockBreakAsync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_break_async", "")
}
func TestAsyncLabelsLabelBlockContinueAsync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_continue_async", "")
}
func TestAsyncLabelsLabelBlockSwitchBreakAsync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_switch_break_async", "")
}
func TestAsyncLabelsLabelBlockOuterContinueAsync(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "label_block_outer_continue_async", "")
}
func TestAsyncLabelsAsyncFinallyReturn(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_finally_return", "")
}
func TestAsyncLabelsAsyncFinallyThrow(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_finally_throw", "")
}
func TestAsyncLabelsAsyncFinallyBreak(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_finally_break", "")
}
func TestAsyncLabelsAsyncFinallyContinue(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_finally_continue", "")
}
func TestAsyncLabelsAsyncFinallyVoidReturn(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_finally_void_return", "")
}
func TestAsyncLabelsAsyncFinallySnapshot(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_finally_snapshot", "")
}
func TestAsyncLabelsAsyncGapLabelAdmission(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_gap_label_admission", "")
}
func TestAsyncLabelsAsyncGapCatchAdmission(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_gap_catch_admission", "")
}
func TestAsyncLabelsAsyncGapFinallyAdmission(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_gap_finally_admission", "")
}
func TestAsyncLabelsAsyncGapFinallyReturnAdmission(t *testing.T) {
	t.Parallel()
	asyncLoopWitness(t, "async_gap_finally_return_admission", "")
}

// These mutants preserve valid IR, clean C builds and balanced ownership. Only
// the source oracle may catch the lost completion, snapshot or destination.
func asyncLabelsMutation(t *testing.T, name, kind string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/"+name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	p, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
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
			t.Fatal("snapshot source not found")
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
		t.Fatal("mutant changed no completion")
	}
	code := native.C(p)
	binary := filepath.Join(t.TempDir(), "wrong-completion")
	if err := native.Build(code, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	if report := leakcheck.Report(t, code, binary); report != "" {
		t.Fatal(report)
	}
	for backend, result := range map[string]run{"native": execute(t, binary), "JavaScript": onJavaScriptBackend(t, p)} {
		if disagreement(truth, result) != "stdout differs" || result.exitCode != 0 || len(result.stderr) != 0 {
			t.Fatalf("%s: only Node stdout must catch %s: %+v", backend, kind, result)
		}
	}
	t.Logf("Node stdout catches %s in native and JavaScript", kind)
}
func TestAsyncLabelsFinallyReturnMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "async_finally_return", "finally")
}
func TestAsyncLabelsFinallyThrowMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "async_finally_throw", "finally")
}
func TestAsyncLabelsFinallyBreakMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "async_finally_break", "finally")
}
func TestAsyncLabelsFinallyContinueMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "async_finally_continue", "finally")
}
func TestAsyncLabelsReturnSnapshotMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "async_finally_snapshot", "snapshot")
}
func TestAsyncLabelsNameResolutionMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "async_labels", "name")
}
func TestAsyncLabelsBlockWrapperBreakSyncMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "label_block_break_sync", "block")
}
func TestAsyncLabelsBlockWrapperBreakAsyncMutant(t *testing.T) {
	t.Parallel()
	asyncLabelsMutation(t, "label_block_break_async", "block")
}
