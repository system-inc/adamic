package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// recordSlotOptions is tsc's CompilerOptions in small: typed members beside an index signature whose
// values are a union, so every member is stored boxed in one Union slot, and string[], (string | number)[]
// and MapLike<string[]> share it with the scalars (#y852c3c).
const recordSlotOptions = `interface MapLike<T> { [index: string]: T; }
interface Options {
    target?: number;
    strict?: boolean;
    lib: string[];
    rootDirs?: string[];
    types?: string[];
    suffixes?: (string | number)[];
    paths?: MapLike<string[]>;
    [option: string]: string | number | boolean | string[] | (string | number)[] | MapLike<string[]> | undefined;
}
const opts: Options = { target: 9, lib: ["es2020", "dom"], rootDirs: ["src", "test"], suffixes: [".ios", 2] };
`

// recordSlotView is whether a program carries the slot read's view: an origin, and V3's array contract
// for it, readonly whatever the member's declared type (recordReadSlotView's condition, Runtime's 1).
func recordSlotView(program *ir.Program, viewed string) bool {
	if len(program.ViewOrigins) == 0 {
		return false
	}
	for _, contract := range program.ViewContracts {
		if contract.Kind == ir.ViewArray && contract.Name == viewed && contract.ArrayReadonly && contract.Element != 0 {
			return true
		}
	}
	return false
}

// checkedHandouts is the ViewType of every for-of handout a lowered program keeps an element check on.
func checkedHandouts(program *ir.Program) []string {
	var found []string
	inspect := func(node any) bool {
		if loop, ok := node.(ir.ForOf); ok && loop.ElementCheck != nil && loop.ElementCheck.Contract != 0 && !loop.ElementCheck.Unrelated {
			found = append(found, loop.ElementCheck.ViewType)
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return found
}

// checkedElementRead is whether a lowered program keeps view's element read checked: a ViewElement with
// a contract, never put back as a raw ArrayIndex.
func checkedElementRead(program *ir.Program, view string) bool {
	found := false
	inspect := func(node any) bool {
		if read, ok := node.(ir.ViewElement); ok && read.View == view && read.Contract != 0 && !read.Unrelated {
			found = true
		}
		return !found
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return found
}

// tsc's shape (moduleNameResolver.ts reads o.rootDirs and o.typeRoots so): the member read out of the Union
// slot lowers, no longer NotYet, as a view origin whose elements are each checked as the loop hands them
// out. A (string | number)[] member is read the same way, its element held to string | number.
func TestRecordSlotArrayReadIsACheckedView(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, member, element, viewed string }{
		{"string[]", "rootDirs", "string", "readonly string[]"},
		{"(string | number)[]", "suffixes", "string | number", "readonly (string | number)[]"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, recordSlotOptions+`function each(options: Options): void {
    const dirs = options.`+row.member+`;
    if (dirs) {
        for (const d of dirs) {
            console.log(String(d));
        }
    }
}
each(opts);
`, ".ts")
			if err != nil {
				t.Fatalf("the slot read stopped: %v", err)
			}
			if !recordSlotView(program, row.viewed) {
				t.Fatalf("no view origin with a readonly %s contract: %d origins, contracts %#v", row.viewed, len(program.ViewOrigins), program.ViewContracts)
			}
			if handouts := checkedHandouts(program); len(handouts) != 1 || handouts[0] != row.element {
				t.Fatalf("the loop's handout isn't checked as %s: %v", row.element, handouts)
			}
		})
	}
	// A number[] beside no other array member was stopped all the same, the TestNamedRecordRefusals row it
	// was, since the representation alone sent every array to the stop. It's a view origin now, in .a and .ts.
	for _, extension := range []string{".a", ".ts"} {
		program, err := lowerSourceAs(t, `interface R {named?:number[];[key:string]:number[]|string|undefined} function read(r:R):number|undefined {return r.named?.length;}`, extension)
		if err != nil {
			t.Fatalf("%s: a number[] member stopped: %v", extension, err)
		}
		if !recordSlotView(program, "readonly number[]") {
			t.Fatalf("%s: the number[] read made no view origin", extension)
		}
	}
}

// Writes through the reach stop at compile time, never at run time (Runtime's condition 3, Language's 1):
// push and an index set are roads checkViewArrays stops on a demanded reach, so a later widening of the
// readonly treatment reds here first. Object.assign, one of writingLibraryMembers, never lowers with an
// array as its target, slot reach or not, so V3 needs no stop of its own for it. The alias row is the
// contract's alias program: a (string | number)[] stored under lib through the index signature can't push
// its 1 either, since every array in a program with a slot origin is demanded, so the lie can only be
// there from the start (internal/oracle's views_v3/record-slot-alias-lie.ts).
func TestRecordSlotArrayWritesStop(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ name, source, want string }{
		{"push", `opts.lib.push("x");`, "a checked view's array read through push (not checked yet)"},
		{"index write", `opts.lib[0] = "x";`, "a checked view's array read through SetIndex (not checked yet)"},
		{"Object.assign", `Object.assign(opts.lib, ["y"]);`, "Object.assign on a shape not proven by a plain literal or its const binding"},
		{"alias push", `const keys: string[] = ["lib"];
const key = keys[0] ?? "lib";
const alias: (string | number)[] = ["a"];
const shared: Options = { lib: [], [key]: alias };
alias.push(1);
console.log(String(shared.lib[1]));`, "a checked view's array read through push (not checked yet)"},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, recordSlotOptions+row.source+"\nconsole.log('ran');\n", ".ts")
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != row.want {
				t.Fatalf("got %v, want NotYet %q", err, row.want)
			}
		})
	}
}

// Language's A: a road that reads elements inside the runtime hands none out to check, so on a slot reach
// each stops, one per group: reading (join), copying (slice, a spread) and callbacks (sort).
func TestRecordSlotArrayNonHandoutReadsStop(t *testing.T) {
	t.Parallel()
	for _, row := range []struct{ road, source string }{
		{"join", `console.log(opts.lib.join(","));`},
		{"slice", "const copy = opts.lib.slice();\nconsole.log(String(copy.length));"},
		{"sort", "opts.lib.sort((a, b) => (a < b ? -1 : a > b ? 1 : 0));"},
		{"spread", "const copy = [...opts.lib];\nconsole.log(String(copy.length));"},
	} {
		t.Run(row.road, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, recordSlotOptions+row.source+"\nconsole.log('ran');\n", ".ts")
			want := "a checked view's array read through " + row.road + " (not checked yet)"
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != want {
				t.Fatalf("got %v, want NotYet %q", err, want)
			}
		})
	}
}

// Language's B: a reach that comes back out of a Map didn't come from the slot as lowering sees it, and
// its element read stays checked all the same, never raw: the read is a ViewElement with a contract in the
// lowered program (checkViewArrays puts back as a raw ArrayIndex every read it doesn't demand). It's
// demanded because a Map's value reaches allocations the flow graph can't name, and the slot origin turns
// demand on for the program.
func TestRecordSlotArrayMapRoadStaysChecked(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, recordSlotOptions+`const m = new Map<string, string[]>();
m.set("k", opts.lib);
console.log(m.get("k")![0] ?? "none");
`, ".ts")
	if err != nil {
		t.Fatalf("the Map road stopped: %v", err)
	}
	if !checkedElementRead(program, `m.get("k")![0]`) {
		t.Fatal("the element read back out of the Map isn't checked")
	}
}

// paths isn't an array: MapLike<string[]> is a dictionary, the slot's one, read by recordReadHeapMember's
// checked heap narrow as before, and no view origin. Beside a second dictionary it keeps today's stop and
// words. A member whose element V3 can't contract, a function here, keeps the array stop's words (Runtime's
// condition 2).
func TestRecordSlotViewLeavesTheRest(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, recordSlotOptions+`const paths = opts.paths;
console.log(paths ? Object.keys(paths).join(",") : "none");
`, ".ts")
	if err != nil {
		t.Fatalf("paths stopped: %v", err)
	}
	if len(program.ViewOrigins) != 0 {
		t.Fatalf("paths' read made %d view origins", len(program.ViewOrigins))
	}
	beside := strings.Replace(recordSlotOptions, "| MapLike<string[]> | undefined;", "| MapLike<string[]> | MapLike<number> | undefined;", 1)
	_, err = lowerSourceAs(t, beside+"const paths = opts.paths;\nconsole.log(paths ? Object.keys(paths).join(\",\") : \"none\");\n", ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "a dictionary member narrowed to an object kind without a checked runtime tag" {
		t.Fatalf("paths beside a second dictionary: got %v, want today's object-kind NotYet", err)
	}
	_, err = lowerSourceAs(t, `interface Hooks { handlers?: (() => void)[]; [name: string]: string | string[] | (() => void)[] | undefined; }
function count(hooks: Hooks): number { const handlers = hooks.handlers; return handlers ? handlers.length : 0; }
`, ".ts")
	if !errors.As(err, &notYet) || notYet.What != recordSlotArrayStop {
		t.Fatalf("a function element: got %v, want NotYet %q", err, recordSlotArrayStop)
	}
}

// recordSlotEach is tsc's slot read in a function, the loop handing out each element of rootDirs.
const recordSlotEach = `function each(options: Options): void {
    const dirs = options.rootDirs;
    if (dirs) {
        for (const d of dirs) {
            console.log(d);
        }
    }
}
each(opts);
`

// recordValuesOf is the lowered program's record node (recordValuesNode): whether it can hold a value the
// flow graph can't place, and why.
func recordValuesOf(program *ir.Program) (bool, string) {
	assignAllocationSites(program)
	return newAllocationFlowGraph(program).recordValuesUnknown()
}

// Every record's values are one allocation-flow node (Runtime's ruling on #y852c3c), so the slot origin
// reaches the arrays records hold, here the literal's, and not the whole program. The unrelated number[]
// is the regression pin for the widening 2c8671ab made: there the RecordCall was untracked, the origin's
// reach Unknown, and xs.push and xs.join stopped as reads through a checked view.
func TestRecordSlotValuesAreOneFlowNode(t *testing.T) {
	t.Parallel()
	program, err := lowerSourceAs(t, recordSlotOptions+recordSlotEach+`const xs = [1, 2];
xs.push(3);
console.log(xs.join(","));
`, ".ts")
	if err != nil {
		t.Fatalf("an unrelated array stopped beside the slot read: %v", err)
	}
	if unknown, why := recordValuesOf(program); unknown {
		t.Fatalf("a record built from a literal of known arrays has Unknown values: %s", why)
	}
	if !recordSlotView(program, "readonly string[]") {
		t.Fatal("the slot read made no view origin")
	}
	if handouts := checkedHandouts(program); len(handouts) != 1 || handouts[0] != "string" {
		t.Fatalf("the slot reach's handout isn't checked: %v", handouts)
	}
}

// A record written a value the graph can't place, a closure's parameter, which a call the graph doesn't
// see may pass anything, makes the record node Unknown, named by that write. Every slot reach is then
// Unknown too, so its writes still stop: fail closed, as before the node.
func TestRecordSlotValuesUnplacedFailClosed(t *testing.T) {
	t.Parallel()
	unplaced := recordSlotOptions + `const put = (dirs: string[]): void => {
    opts.rootDirs = dirs;
};
put(["x"]);
` + recordSlotEach
	program, err := lowerSourceAs(t, unplaced, ".ts")
	if err != nil {
		t.Fatalf("the unplaced write's program stopped: %v", err)
	}
	unknown, why := recordValuesOf(program)
	if !unknown || !strings.Contains(why, "a record set in function ") || !strings.Contains(why, "parameter may receive untracked or omitted arguments") {
		t.Fatalf("got unknown %t, %q; want Unknown, named by the record set of the closure's parameter", unknown, why)
	}
	_, err = lowerSourceAs(t, unplaced+"opts.lib.push(\"x\");\n", ".ts")
	var notYet *NotYet
	if !errors.As(err, &notYet) || notYet.What != "a checked view's array read through push (not checked yet)" {
		t.Fatalf("got %v, want the push through the slot reach to stop", err)
	}
}
