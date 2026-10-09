package fresh

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// These reads really add outside. Removing just that addition leaves their placeholder, which
// still defers the write locally. At a concrete caller, a leak or an earlier escape makes the
// placeholder's translation include outside again. This is different from ignoring clobbers in
// made's judgment: that clause must count objects exposed by the replay itself.
func TestClobberedPlaceholderReadsCoveredAtCaller(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"child", anyField} {
		name := "load"
		if field == anyField {
			name = "everything"
		}
		for _, access := range []string{"argument", "weak argument", "descendant back link", "escaped before"} {
			t.Run(name+"/"+access, func(t *testing.T) {
				callee, parameter, holder := clobberedReadAnalysis()
				operands := []ir.Expression{}
				if access == "argument" {
					operands = append(operands, ir.Read{Local: 0, Of: ir.Object})
				} else if access == "weak argument" {
					operands = append(operands, ir.WeakOf{Value: ir.Read{Local: 0, Of: ir.Object}})
				} else if access == "descendant back link" {
					operands = append(operands, ir.Property{Object: ir.Read{Local: 0, Of: ir.Object}, Name: "child", Of: ir.Object})
				}
				callee.value(ir.CallClosure{Closure: ir.Read{Local: 2, Of: ir.Closure}, Arguments: operands})
				var loaded value
				if field == anyField {
					loaded = callee.everything(parameter)
				} else {
					loaded = callee.load(parameter, field)
				}
				if !loaded.strong[outside] || !loaded.weak[outside] {
					t.Fatal("clobbered read did not add outside")
				}
				without := loaded.copy()
				delete(without.strong, outside)
				delete(without.weak, outside)
				if without.empty() {
					t.Fatal("removal lost the field placeholder too")
				}
				for _, withMerge := range []bool{true, false} {
					stored := without
					if withMerge {
						stored = loaded
					}
					callee.deferred = map[writeKey]*deferral{}
					callee.judge(writeKey{function: 0}, Write{Site: 1, Kind: WriteField, Function: 0}, holder, stored, "", callee.state.exposed)
					if len(callee.deferred) != 1 {
						t.Fatalf("merge=%v: local write was not deferred", withMerge)
					}
					callee.state.locals[callee.returnedLocal] = stored
					transfer := callee.summarize(callee.state)
					leakedParameter := false
					for _, leaked := range transfer.leaks {
						want := term{param: 0}
						if access == "descendant back link" {
							want = term{param: 0, field: "child", level: 1}
						}
						leakedParameter = leakedParameter || leaked == want
					}
					if leakedParameter != (access != "escaped before") {
						t.Fatal("opaque call did not record the parameter leak")
					}
					caller, arguments := clobberedReadCaller(callee.proof.program)
					if access == "descendant back link" {
						child := value{strong: objects{newest(1): true}}
						caller.state.store(newest(0), "child", child)
						caller.state.store(newest(1), "parent", value{weak: arguments[0].strong})
					}
					if access == "escaped before" {
						caller.state.escape(arguments[0])
					}
					result := caller.made(0, transfer, arguments)
					if !result.strong[outside] || !result.weak[outside] {
						t.Fatalf("merge=%v: translated read lost outside", withMerge)
					}
					write := caller.proof.writes[writeKey{function: 0}]
					if write == nil || write.Proven {
						t.Fatalf("merge=%v: caller proved the deferred write", withMerge)
					}
					t.Logf("merge=%v: returned read includes outside; deferred write refused", withMerge)
				}
			})
		}
	}
}

// Clobbered is not an access set. An unrelated callback can make these merges refuse a safe
// write too: its parameter's field is empty at the caller, and the callback cannot reach it.
// Thus equal judgments in the tests above are not a claim that all judgments are unchanged.
func TestClobberedPlaceholderReadCanLosePrecision(t *testing.T) {
	t.Parallel()
	for _, field := range []string{"child", anyField} {
		callee, parameter, holder := clobberedReadAnalysis()
		callee.value(ir.CallClosure{Closure: ir.Read{Local: 2, Of: ir.Closure}})
		loaded := callee.load(parameter, field)
		if field == anyField {
			loaded = callee.everything(parameter)
		}
		if !loaded.strong[outside] {
			t.Fatal("unrelated callback did not clobber the read")
		}
		for _, withMerge := range []bool{true, false} {
			stored := loaded.copy()
			if !withMerge {
				delete(stored.strong, outside)
				delete(stored.weak, outside)
			}
			callee.deferred = map[writeKey]*deferral{}
			callee.judge(writeKey{function: 0}, Write{Site: 1, Kind: WriteField, Function: 0}, holder, stored, "", callee.state.exposed)
			transfer := callee.summarize(callee.state)
			caller, arguments := clobberedReadCaller(callee.proof.program)
			caller.made(0, transfer, arguments)
			write := caller.proof.writes[writeKey{function: 0}]
			if write.Proven == withMerge {
				t.Fatalf("field=%q merge=%v: proven=%v", field, withMerge, write.Proven)
			}
			t.Logf("field=%q merge=%v: safe write proven=%v", field, withMerge, write.Proven)
		}
	}
}

// Reading before the opaque call leaves exactly the placeholder that removing the merge would
// leave after it: call changes escapes and clobbered, but not the symbolic heap. Exercise the
// whole graph and summary fixed point as well as the controlled replay states above.
func TestClobberedReadProgramKeepsRefusal(t *testing.T) {
	t.Parallel()
	for _, spread := range []bool{false, true} {
		for _, readAfter := range []bool{false, true} {
			parameter := ir.Read{Local: 0, Of: ir.Object}
			var read ir.Expression = ir.Property{Object: parameter, Name: "child", Of: ir.Object}
			if spread {
				read = ir.ObjectLiteral{Spread: parameter}
			}
			load := ir.Declare{Local: 3, Value: read}
			call := ir.Evaluate{Value: ir.CallClosure{
				Closure: ir.Read{Local: 2, Of: ir.Closure}, Arguments: []ir.Expression{parameter},
			}}
			body := []ir.Statement{load, call}
			if readAfter {
				body = []ir.Statement{call, load}
			}
			body = append(body, ir.SetProperty{
				Object: ir.Read{Local: 1, Of: ir.Object}, Name: "link",
				Value: ir.Read{Local: 3, Of: ir.Object}, Site: 1,
			})
			program := &ir.Program{
				Locals: []ir.Local{
					{Type: ir.Object}, {Type: ir.Object}, {Type: ir.Closure}, {Type: ir.Object},
					{Type: ir.Object}, {Type: ir.Object, Global: true}, {Type: ir.Object},
				},
				Functions: []ir.Function{
					{Name: "readAfterCallback", Parameters: []int{0, 1, 2}, Body: body},
					{Name: "callback", Parameters: []int{6}, Closure: true},
				},
				Main: []ir.Statement{
					ir.Declare{Local: 4, Value: ir.ObjectLiteral{}},
					ir.Declare{Local: 5, Value: ir.ObjectLiteral{}},
					ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{
						ir.Read{Local: 4, Of: ir.Object}, ir.Read{Local: 5, Of: ir.Object},
						ir.MakeClosure{Function: 1},
					}}},
				},
			}
			writes := ProveWrites(program)
			if len(writes) != 1 || writes[0].Site != 1 || writes[0].Proven {
				t.Fatalf("spread=%v readAfter=%v: want the link write refused, got %v", spread, readAfter, writes)
			}
			t.Logf("spread=%v readAfter=%v: link write refused", spread, readAfter)
		}
	}
}

func clobberedReadAnalysis() (*analysis, value, value) {
	program := &ir.Program{
		Functions: []ir.Function{{Name: "readAfterCallback", Parameters: []int{0, 1, 2}}},
		Locals:    []ir.Local{{Type: ir.Object}, {Type: ir.Object}, {Type: ir.Closure}},
	}
	a := &analysis{
		proof:    &freshness{program: program, writes: map[writeKey]*Write{}},
		function: 0, state: newState(), placeholders: true,
		termObjects: map[term]object{}, deferred: map[writeKey]*deferral{}, returnedLocal: 3,
	}
	parameter := value{strong: objects{a.placeholder(term{param: 0}): true}}
	holder := value{strong: objects{a.placeholder(term{param: 1}): true}}
	callback := value{strong: objects{a.placeholder(term{param: 2}): true}}
	a.state.locals[0], a.state.locals[1], a.state.locals[2] = parameter, holder, callback
	return a, parameter, holder
}

func clobberedReadCaller(program *ir.Program) (*analysis, []value) {
	a := &analysis{
		proof:    &freshness{program: program, writes: map[writeKey]*Write{}},
		function: -1, state: newState(), judging: true,
	}
	// A confined parameter object with no reference fields, and a global destination. A real
	// callback handed the parameter may fill its field with an object reaching that destination.
	parameter := value{strong: objects{newest(0): true}}
	// The callee records its write before callers judge it.
	a.proof.record(writeKey{function: 0}, Write{Site: 1, Kind: WriteField, Function: 0})
	return a, []value{parameter, outsideValue(), outsideValue()}
}
