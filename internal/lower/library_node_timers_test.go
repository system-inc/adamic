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

// tsc's getNodeSystem (sys.ts:1499) shape, through required methods and through System's optional
// ones: each timer read as a value is one closure per program, whose body is the timer's direct call,
// held by a global, so a second literal, a view of its own and an identity comparison read the same
// one, and the program runs the loop for the setTimeout inside it.
func TestNodeTimerValuesAreOneClosureEach(t *testing.T) {
	t.Parallel()
	for name, methods := range map[string]string{
		"required": `setTimeout(callback: () => void, ms: number): unknown;
    clearTimeout(timeoutId: unknown): void;`,
		"optional": `setTimeout?(callback: () => void, ms: number): unknown;
    clearTimeout?(timeoutId: unknown): void;`,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, `import 'node:timers';
interface Host {
    `+methods+`
}
type SetTimeoutView = (callback: () => void, ms: number) => unknown;
const host: Host = { setTimeout, clearTimeout };
const again: Host = { setTimeout, clearTimeout };
const view: SetTimeoutView = setTimeout;
console.log(String(view === setTimeout) + ' ' + String(clearTimeout === clearTimeout) + ' ' + String([host, again].length));`)
			if err != nil {
				t.Fatal(err)
			}
			roles := map[string]int{}
			for _, function := range program.Functions {
				roles[function.Source.Role]++
			}
			if roles["host_set_timeout"] != 1 || roles["host_clear_timeout"] != 1 {
				t.Fatalf("want one closure for each timer, got %v", roles)
			}
			if operations, _ := timerOperations(program); operations["set_timeout"] != 1 || operations["clear_timeout"] != 1 {
				t.Fatalf("want each closure's body to be its timer's call, got %v", operations)
			}
			if !program.HasPromises() {
				t.Fatal("a setTimeout value must run the loop without an explicit Promise")
			}
		})
	}
}

// A timer's value is admitted only where its slot sees it exactly as its view (Adamic's ruling on
// #995spev, Nov 13), an optional method's undefined aside; any other slot, and none, is NotYet naming
// what it saw. An optional callback is refused before it, at the relation, by 0.1's rule for a
// method's parameters. Mutant: nodeTimerView admitting a third parameter, and the wider view's row
// goes red.
func TestNodeTimerValuesNeedTheirView(t *testing.T) {
	t.Parallel()
	const setView = "node:timers.setTimeout value needs exactly (callback: () => void, ms: number) => unknown, not "
	const clearView = "node:timers.clearTimeout value needs exactly (timeoutId: unknown) => void, not "
	for source, want := range map[string]string{
		// System's slot as rulings.md:314 had it before the ruling.
		"interface Host { setTimeout?<A extends unknown[]>(callback: (...args: NoInfer<A>) => void, ms: number, ...args: A): unknown; }\nconst host: Host = { setTimeout };": setView,
		"const wide: (callback: () => void, ms: number, label: string) => unknown = setTimeout;":                                                                             setView,
		"interface Host { setTimeout(callback: () => void, ms?: number): unknown; }\nconst host: Host = { setTimeout };":                                                     setView,
		"interface Host { setTimeout(callback: (...names: string[]) => void, ms: number): unknown; }\nconst host: Host = { setTimeout };":                                    setView,
		"interface Host { setTimeout(callback: () => void, ms: number): NodeJS.Timeout; }\nconst host: Host = { setTimeout };":                                               setView,
		"const read = setTimeout;":   setView + "typeof setTimeout with no slot",
		"const read = clearTimeout;": clearView + "(timeout: unknown) => void with no slot",
		"interface Host { clearTimeout(timeoutId: NodeJS.Timeout): void; }\nconst host: Host = { clearTimeout };":                                                   clearView,
		"class Holder { get _onTimeout(): number { return 1; } }\ninterface Host { clearTimeout(timeoutId: unknown): void; }\nconst host: Host = { clearTimeout };": "node:timers.clearTimeout value where the program declares an _onTimeout method or accessor, which Node reads",
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

// An optional callback in setTimeout's slot is refused by 0.1 before the value's view is judged, since
// tsc relates a method's parameters both ways.
func TestNodeTimerValueOptionalCallbackIsRefused(t *testing.T) {
	t.Parallel()
	_, err := lowerSource(t, "import 'node:timers';\ninterface Host { setTimeout(callback?: () => void, ms?: number): unknown; }\nconst host: Host = { setTimeout };")
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "a function taking () => void seen as one taking (() => void) | undefined") {
		t.Fatalf("want 0.1's refusal of the optional callback, got %v", err)
	}
}
