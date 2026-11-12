package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// optionalTools is an interface with optional methods of each result and a field holding a function,
// objects with and without them, receivers that may be missing, and an argument that counts itself.
const optionalTools = "interface Tool {\n\treadonly name: string;\n\tsize?(): number;\n\tlabel?(prefix: string): string;\n\tready?: (() => boolean) | undefined;\n\tlog?(line: string): void;\n}\n" +
	"const tools: Tool[] = [{ name: 'bare' }, { name: 'full', size: (): number => 2, label: (prefix: string): string => `${prefix}full`, ready: (): boolean => true, log: (line: string): void => {\n\tconsole.log(line);\n} }];\n" +
	"function pick(index: number): Tool | undefined {\n\treturn tools[index];\n}\n" +
	"function pickOrNull(index: number): Tool | null {\n\treturn tools[index] ?? null;\n}\n" +
	"let notes = 0;\nfunction note(text: string): string {\n\tnotes += 1;\n\treturn text;\n}\n"

// optionalCalls counts a program's calls through ?. of an optional method (ir.CallClosure's Optional).
func optionalCalls(program *ir.Program) int {
	count := 0
	visit := func(node any) bool {
		if call, isCall := node.(ir.CallClosure); isCall && call.Optional {
			count++
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return count
}

// An optional call of an optional method the program declares lowers (#tvq1eqm, optional_method_call.go),
// once, as an optional call: tsc's performance.ts shape, a receiver that's undefined, null or there, a
// method that's absent or present, an argument with an effect, object.method?.(...) with no ?. before
// it, a field holding a function, a void method as a statement and a class's method through the
// interface. The oracle (internal/oracle/optional_method_call_test.go) holds what they run to. Mutant:
// optionalMethodSignature made nil for every call, and each row stops as optionalCall's not yet.
func TestOptionalMethodCallsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"tsc's shape", "interface System {\n\treadonly debugMode?: boolean;\n\tcpuProfilingEnabled?(): boolean;\n}\nfunction performanceImpl(system: System | undefined): string {\n\tif (system?.cpuProfilingEnabled?.() || system?.debugMode) {\n\t\treturn 'native';\n\t}\n\treturn 'none';\n}\nconsole.log(performanceImpl(undefined));\nconsole.log(performanceImpl({ cpuProfilingEnabled: (): boolean => true }));\n"},
		{"a receiver that's undefined", optionalTools + "console.log(String(pick(5)?.size?.()));\n"},
		{"a receiver that's null", optionalTools + "console.log(pickOrNull(5)?.label?.(note('a')) ?? 'none');\n"},
		{"a method that's absent", optionalTools + "const bare: Tool = { name: 'only' };\nconsole.log(String(bare.size?.()));\n"},
		{"a method that's present", optionalTools + "console.log(String(pick(1)?.size?.()));\n"},
		{"an argument with an effect", optionalTools + "console.log(`${pick(0)?.label?.(note('skipped')) ?? 'none'} ${notes}`);\n"},
		{"no ?. before the method", optionalTools + "for (const tool of tools) {\n\tconsole.log(String(tool.size?.()));\n}\n"},
		{"a field holding a function", optionalTools + "console.log(String(pick(1)?.ready?.()));\n"},
		{"a void method as a statement", optionalTools + "pick(1)?.log?.('logged');\n"},
		{"a class's method through the interface", "interface Sized {\n\tsize?(): number;\n}\nclass Meter implements Sized {\n\tsize(): number {\n\t\treturn 3;\n\t}\n}\nconst meter: Sized = new Meter();\nconsole.log(String(meter.size?.()));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if count := optionalCalls(program); count != 1 {
				t.Errorf("%d optional calls, want the one", count)
			}
		})
	}
}

// What the optional method call doesn't lower keeps its stop: the chain going on after the call, a
// library method, a method behind a longer chain or in parentheses, and one whose field may hold null
// (optionalCall's not yet, or property's for the chain); an optional method held in an alias, which
// stays refused as a method read as a value, though its presence test lowers
// (optional_member_presence_test.go); a class whose method is a getter, which
// Node calls and then calls what it returns, refused before the optional lookup could miss it and give
// undefined (Lower's and Compiler's road on #vx8qdwg, probed on Home: it stops the plain call too); and the
// call through a checked view,
// check 49's road. On main the .ts row is refused first, as a checked view read with an unsupported
// representation conversion contract, so unemittableCheckedViewRead's optional call arm is a second
// stop behind it: dropping that arm is an equivalent mutant (Home, Nov 8), and the arm stays so a later
// contract for that read meets a not yet, never readUnionViewAt's compiler bug.
func TestOptionalMethodCallGapsStop(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, extension, source, want string }{
		{"the chain going on after the call", ".a", "interface Inner {\n\treadonly x: number;\n}\ninterface Maker {\n\tmake?(): Inner;\n}\nfunction made(maker: Maker | undefined): string {\n\treturn String(maker?.make?.().x);\n}\nconsole.log(made({ make: (): Inner => ({ x: 1 }) }));\n", "an optional chain longer than one step"},
		{"a library method", ".a", "function trimmed(text: string | undefined): string {\n\treturn text?.trim?.() ?? 'none';\n}\nconsole.log(trimmed(' a '));\n", optionalCallNotYet},
		{"a method behind a longer chain", ".a", "interface Sized {\n\tsize?(): number;\n}\ninterface Holder {\n\treadonly sized: Sized;\n}\nfunction size(holder: Holder | undefined): string {\n\treturn String(holder?.sized.size?.());\n}\nconsole.log(size({ sized: {} }));\n", optionalCallNotYet},
		{"parentheses around the callee", ".a", "interface Sized {\n\tsize?(): number;\n}\nfunction size(sized: Sized | undefined): string {\n\treturn String((sized?.size)?.());\n}\nconsole.log(size({}));\n", optionalCallNotYet},
		{"a field that may hold null", ".a", "interface Sized {\n\tsize?: (() => number) | null;\n}\nfunction size(sized: Sized): string {\n\treturn String(sized.size?.());\n}\nconsole.log(size({}));\n", optionalCallNotYet},
		{"an alias", ".a", "interface Sized {\n\tsize?(): number;\n}\nfunction size(sized: Sized): string {\n\tconst held = sized.size;\n\treturn held ? 'sized' : 'unsized';\n}\nconsole.log(size({}));\n", "a method read as a value (size would lose its object, and this with it)"},
		{"a class whose method is a getter", ".a", "interface Runner {\n\trun?(label: string): string;\n}\nclass Getter {\n\tget run(): (label: string) => string {\n\t\treturn (label: string) => label;\n\t}\n}\nfunction call(runner: Runner): string {\n\treturn runner.run?.('x') ?? 'absent';\n}\nconsole.log(call(new Getter()));\n", "a method sharing its name with a getter"},
		{"through a checked view", ".ts", "interface Probe {\n    readonly kind: 'Probe';\n    enabled?(): boolean;\n}\nfunction probe(raw: { readonly kind: string }): string {\n    const viewed = raw as Probe;\n    return String(viewed.enabled?.());\n}\nconsole.log(probe({ kind: ['Pro', 'be'].join('') }));\n", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var notYet *NotYet
			var refused *Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("got %v, want it stopped saying %q", err, probe.want)
			}
			if !strings.Contains(err.Error(), probe.want) {
				t.Errorf("got %v, want it stopped saying %q", err, probe.want)
			}
		})
	}
}
