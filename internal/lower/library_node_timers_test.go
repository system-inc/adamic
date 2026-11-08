package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// timerOperations counts the timer operations a program lowered to, and each set_timeout's extra
// arguments by their representation.
func timerOperations(program *ir.Program) (map[string]int, [][]ir.Type) {
	operations, extras := map[string]int{}, [][]ir.Type{}
	visit := func(node any) bool {
		if call, ok := node.(ir.NodeFSFile); ok && (call.Operation == "set_timeout" || call.Operation == "clear_timeout") {
			operations[call.Operation]++
			if call.Operation == "set_timeout" {
				types := []ir.Type{}
				for _, argument := range call.Arguments[2:] {
					types = append(types, argument.Type())
				}
				extras = append(extras, types)
			}
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return operations, extras
}

// tsc's three shapes (sys.ts:377, :490, :767 and :770) lower, run the loop without an explicit
// Promise, and hand each extra argument over as its callback's parameter takes it: a number where a
// number | undefined goes is packed as one.
func TestNodeTimersTscShapes(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `import 'node:timers';
interface Queue { names: string[] }
function poll(label: string, queue: Queue): void { console.log(label + queue.names.join()); }
function pollQueue(_label: string): void { console.log('poll'); }
function counted(count?: number): void { console.log(String(count)); }
let timer: NodeJS.Timeout | undefined;
setTimeout(poll, 250, 'pollPollingIntervalQueue', { names: ['a'] });
const scheduled = setTimeout(pollQueue, 2000, 'pollQueue');
if (scheduled) { console.log('held'); }
if (timer) { clearTimeout(timer); }
timer = setTimeout(pollQueue, 1000, 'timerToUpdateChildWatches');
setTimeout(counted, 1, 4);
setTimeout(() => { console.log('none'); });`)
	if err != nil {
		t.Fatal(err)
	}
	if !program.HasPromises() {
		t.Fatal("setTimeout must run the loop without an explicit Promise")
	}
	operations, extras := timerOperations(program)
	if operations["set_timeout"] != 5 || operations["clear_timeout"] != 1 {
		t.Fatal(operations)
	}
	want := [][]ir.Type{{ir.String, ir.Object}, {ir.String}, {ir.String}, {ir.MaybeNumber}, {}}
	for index, types := range want {
		if len(extras[index]) != len(types) {
			t.Fatalf("set_timeout %d: extra arguments %v, want %v", index, extras[index], types)
		}
		for position, of := range types {
			if extras[index][position] != of {
				t.Fatalf("set_timeout %d: extra arguments %v, want %v", index, extras[index], types)
			}
		}
	}
}

// What this head doesn't hold stays NotYet, in words that say what.
func TestNodeTimersStayLoud(t *testing.T) {
	t.Parallel()
	for source, want := range map[string]string{
		"setTimeout((...names: string[]) => { console.log(names.join()); }, 1, 'a');":                                     "node:timers.setTimeout with a generic, this-typed or rest-parameter callback",
		"setTimeout(async () => { console.log('a'); }, 1);":                                                               "node:timers.setTimeout with a callback returning a promise",
		"setTimeout((flag: boolean): Promise<number> | number => flag ? Promise.resolve(1) : 1, 1, true);":                "node:timers.setTimeout with a callback returning a promise",
		"const handle = setTimeout(() => {}, 1); handle.unref();":                                                         "node:timers.Timeout.unref",
		"const handle = setTimeout(() => {}, 1); handle.refresh();":                                                       "node:timers.Timeout.refresh",
		"const handle = setTimeout(() => {}, 1); const cleared = clearTimeout(handle);":                                   "node:timers.clearTimeout used as a value; call it as a statement",
		"clearTimeout({ _onTimeout: 1 });":                                                                                "node:timers.clearTimeout of an object with an _onTimeout property, which Node unenrolls",
		"class Holder { get _onTimeout(): number { return 1; } } const held: unknown = new Holder(); clearTimeout(held);": "node:timers.clearTimeout of a value that may be an object with an _onTimeout method or accessor, which Node reads",
		"const read = setTimeout;":                                                                                        "observing a callable namespace object",
	} {
		t.Run(source, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, "import 'node:timers';\n"+source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, want) {
				t.Fatalf("want NotYet %q, got %v", want, err)
			}
		})
	}
}
