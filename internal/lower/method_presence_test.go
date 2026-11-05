package lower

import (
	"errors"
	"strings"
	"testing"
)

// presenceCounter is a class with a method, for the presence probes below; its read is on line 9.
const presenceCounter = "class Counter {\n\tcount = 0;\n\tstep(): number {\n\t\tthis.count += 1;\n\t\treturn this.count;\n\t}\n}\nconst counter = new Counter();\n"

// A method read only to test for it lowers at each consumer that only tests it (#4q1b2er,
// method_presence.go): typeof, !, a condition and the left of &&, through parentheses and through ?..
// The oracle holds what these run to (testdata/method_presence in internal/oracle); this holds that
// they lower at all.
func TestMethodPresenceTestsLower(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source string }{
		{"typeof a library global's method", "console.log(String(typeof Object.setPrototypeOf === 'function'));\n"},
		{"an array's method as a condition", "function fill<T>(array: T[], value: T): T[] {\n\tif (array.fill) {\n\t\tarray.fill(value);\n\t}\n\treturn array;\n}\nconsole.log(fill([1, 2], 3).join());\n"},
		{"a string's method as a condition", "function twice(text: string): string {\n\tif (text.repeat) {\n\t\treturn text.repeat(2);\n\t}\n\treturn text + text;\n}\nconsole.log(twice('ab'));\n"},
		{"typeof", presenceCounter + "console.log(typeof counter.step);\n"},
		{"!", presenceCounter + "console.log(String(!counter.step));\n"},
		{"the left of &&", presenceCounter + "console.log(String(counter.step && counter.step()));\n"},
		{"the test of ?:", presenceCounter + "console.log(String(counter.step ? counter.step() : 0));\n"},
		{"the test of while", presenceCounter + "while (counter.step) {\n\tcounter.step();\n\tbreak;\n}\n"},
		{"the test of for", presenceCounter + "for (; counter.step; ) {\n\tcounter.step();\n\tbreak;\n}\n"},
		{"the test of do...while", presenceCounter + "do {\n\tif (counter.step() > 2) {\n\t\tbreak;\n\t}\n} while (counter.step);\n"},
		{"in parentheses", presenceCounter + "console.log(typeof ((counter.step)));\n"},
		{"through ?.", presenceCounter + "function pick(): Counter | undefined {\n\treturn counter;\n}\nconsole.log(typeof pick()?.step);\nconsole.log(String(!pick()?.step));\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			if _, err := lowerSource(t, probe.source); err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
		})
	}
}

// Every other consumer keeps the method value, and is refused with the same words a method read as a
// value always is, at the read: an alias (refused at the read, not where the alias is tested), ||, ??,
// an arm of ?:, an assignment, an argument, a return and a comma operand. So is an optional method's
// read, which nothing proves is there. Mutant (testdata/method_presence/mutants.md in internal/oracle):
// presenceConsumer's check dropped admits every consumer, and these rows stop as not yet instead.
func TestMethodReadsThatKeepTheValueAreRefused(t *testing.T) {
	t.Parallel()
	const refusal = "Adamic 0.1 refuses a method read as a value (step would lose its object, and this with it)"
	for _, probe := range []struct{ name, source, want string }{
		{"an alias", presenceCounter + "const step = counter.step;\nconsole.log(typeof step);\n", "main.a:9:14: " + refusal},
		{"||", presenceCounter + "const either = counter.step || ((): number => 0);\nconsole.log(String(either()));\n", "main.a:9:16: " + refusal},
		{"??", presenceCounter + "const either = counter.step ?? ((): number => 0);\nconsole.log(String(either()));\n", "main.a:9:16: " + refusal},
		{"an arm of ?:", presenceCounter + "const pick = counter.count > 0 ? counter.step : (): number => 0;\nconsole.log(String(pick()));\n", "main.a:9:34: " + refusal},
		{"an assignment", presenceCounter + "let held = (): number => 0;\nheld = counter.step;\nconsole.log(String(held()));\n", "main.a:10:8: " + refusal},
		{"an argument", presenceCounter + "function call(f: () => number): number {\n\treturn f();\n}\nconsole.log(String(call(counter.step)));\n", "main.a:12:25: " + refusal},
		{"a return", presenceCounter + "function detach(): () => number {\n\treturn counter.step;\n}\n", "main.a:10:9: " + refusal},
		{"a comma operand", presenceCounter + "console.log(typeof (counter.step(), counter.step));\n", "main.a:9:37: " + refusal},
		{"an optional method", "interface Maybe {\n\tm?(): number;\n}\nconst maybe: Maybe = { m() { return 1; } };\nconsole.log(typeof maybe.m);\n", "main.a:5:20: Adamic 0.1 refuses a method read as a value (m would lose its object, and this with it)"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.Contains(refused.Error(), probe.want) {
				t.Errorf("got %v, want a refusal containing %q", err, probe.want)
			}
		})
	}
}

// A presence test is a constant only where the receiver's construction proves the method there. A
// method signature on a program's interface reaches values Adamic didn't build as that type: an object
// literal handed to a parameter, a value let in by a checked cast whose real object may lack the
// method (Node says "undefined"), a host value; and a library method off anything but Adamic's own
// array or string, a Map's. Each is not yet, never folded to "function". Mutant
// (testdata/method_presence/mutants.md in internal/oracle): methodPresence's guard dropped folds these.
func TestMethodPresenceTheReceiverDoesntProveIsNotYet(t *testing.T) {
	t.Parallel()
	const unproven = "stage 0 can't lower " + methodPresenceUnproven + " yet"
	for _, probe := range []struct{ name, extension, source, want string }{
		{"an object literal handed to an interface parameter", ".a", "interface Greeter {\n\tgreet(): string;\n}\nfunction probe(greeter: Greeter): string {\n\treturn typeof greeter.greet;\n}\nconsole.log(probe({ greet: () => 'hi' }));\n", "main.a:5:16: " + unproven},
		{"an interface condition", ".a", "interface Greeter {\n\tgreet(): string;\n}\nfunction probe(greeter: Greeter): string {\n\tif (greeter.greet) {\n\t\treturn greeter.greet();\n\t}\n\treturn 'none';\n}\nconsole.log(probe({ greet: () => 'hi' }));\n", "main.a:5:6: " + unproven},
		// .a refuses this downcast outright; .ts keeps it a checked cast, whose view is the receiver.
		{"a cast receiver", ".ts", "interface Circle {\n\treadonly kind: 'circle';\n\tarea(): number;\n}\nfunction probe(raw: { readonly kind: string }): string {\n\treturn typeof (raw as Circle).area;\n}\nconsole.log(probe({ kind: 'circle' }));\n", "main.ts:6:16: " + unproven},
		{"a host value's method", ".a", "import * as fs from 'node:fs';\nconst stats = fs.statSync('.');\nconsole.log(typeof stats.isFile);\n", "main.a:3:20: " + unproven},
		{"a library method on a Map", ".a", "const counts = new Map<string, number>();\ncounts.set('a', 1);\nconsole.log(typeof counts.get);\n", "main.a:3:20: " + unproven},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, probe.extension)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want a not-yet ending %q", err, probe.want)
			}
		})
	}
}

// What the IR can't yet say through ?. is not yet, in its own words: the left of && whose missing
// receiver would be the whole value, and a receiver held as a string, whose truthiness isn't its
// presence ("" has repeat).
func TestMethodPresenceThroughOptionalChainGaps(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"the left of &&", presenceCounter + "function pick(): Counter | undefined {\n\treturn counter;\n}\nconsole.log(String(pick()?.step && 1));\n", "main.a:12:20: stage 0 can't lower the left of && reading a method through ?., whose missing receiver would be the value yet"},
		{"a string receiver", "function pick(): string | undefined {\n\treturn '';\n}\nconsole.log(typeof pick()?.repeat);\n", "main.a:4:20: stage 0 can't lower a presence test of a method read through ?. on a string yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want a not-yet ending %q", err, probe.want)
			}
		})
	}
}
