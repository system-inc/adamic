package lower

import (
	"errors"
	"strings"
	"testing"
)

// optionalValueCallers is a callback that counts its calls and prints what it's given, and an argument
// that counts itself, so a row shows both whether the callback ran and whether its arguments did.
const optionalValueCallers = "let calls = 0;\nlet evaluated = 0;\nfunction report(n: number): void {\n\tcalls += 1;\n\tconsole.log(`report ${n}`);\n}\nfunction arg(n: number): number {\n\tevaluated += 1;\n\treturn n;\n}\n"

// optionalValueFinal prints the two counts, so every row ends by showing them.
const optionalValueFinal = "console.log(`calls ${calls} evaluated ${evaluated}`);\n"

// A function value called through ?. lowers as one optional call (#036ve6w, optionalValueCall), and
// agrees with Node on the JavaScript backend and natively under ASan and UBSan. The rows:
//   - tsc's sys.ts pollWatchedFileQueue, first, since it's the 73 roots' path: an optional parameter called with three arguments;
//   - a parameter absent, whose side-effecting argument must not run, and present;
//   - a local typed `F | undefined`, void, and one returning a number read as `number | undefined`;
//   - a captured local, as emitter.ts's createPrinter reads its destructured handlers from nested functions;
//   - Lower's read-once rows, an argument that writes the binding, directly and through a closure, which still calls what was read.
//
// Mutant: optionalValueSignature answering nil for every call, and each row stops as optionalCall's
// NotYet; and Optional dropped in optionalValueCall, and the absent rows fail natively where Node
// prints the skip.
func TestOptionalValueCallsAgree(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"sys.ts pollWatchedFileQueue", "interface WatchedFile {\n\treadonly fileName: string;\n}\nfunction pollWatchedFileQueue<T extends WatchedFile>(queue: T[], pollIndex: number, callbackOnWatchFileStat?: (watchedFile: T, pollIndex: number, fileChanged: boolean) => void): number {\n\tfor (const watchedFile of queue) {\n\t\tcallbackOnWatchFileStat?.(watchedFile, pollIndex, watchedFile.fileName.length > 1);\n\t}\n\treturn queue.length;\n}\nconst files: WatchedFile[] = [{ fileName: 'a' }, { fileName: 'bc' }];\nconsole.log(String(pollWatchedFileQueue(files, 0)));\nconsole.log(String(pollWatchedFileQueue(files, 1, (file: WatchedFile, index: number, changed: boolean): void => {\n\tconsole.log(`${file.fileName} ${index} ${changed}`);\n})));\n", "2\na 1 false\nbc 1 true\n2\n"},
		{"a parameter absent and present", optionalValueCallers + "function run(callback?: (n: number) => void): void {\n\tcallback?.(arg(1));\n}\nrun();\nrun(report);\n" + optionalValueFinal, "report 1\ncalls 1 evaluated 1\n"},
		{"a local, void", optionalValueCallers + "function pick(present: boolean): ((n: number) => void) | undefined {\n\treturn present ? report : undefined;\n}\nconst absent = pick(false);\nabsent?.(arg(2));\nconst present = pick(true);\npresent?.(arg(3));\n" + optionalValueFinal, "report 3\ncalls 1 evaluated 1\n"},
		{"a local returning a number", optionalValueCallers + "function twice(n: number): number {\n\tcalls += 1;\n\treturn n * 2;\n}\nfunction pick(present: boolean): ((n: number) => number) | undefined {\n\treturn present ? twice : undefined;\n}\nconst absent = pick(false);\nconst none: number | undefined = absent?.(arg(4));\nconsole.log(String(none));\nconst present = pick(true);\nconst some: number | undefined = present?.(arg(5));\nconsole.log(String(some));\n" + optionalValueFinal, "undefined\n10\ncalls 1 evaluated 1\n"},
		{"a captured local, as emitter.ts's handlers", optionalValueCallers + "interface Handlers {\n\treadonly onBefore?: (n: number) => void;\n}\nfunction printer(handlers: Handlers): (n: number) => void {\n\tconst { onBefore } = handlers;\n\treturn (n: number): void => {\n\t\tonBefore?.(arg(n));\n\t};\n}\nprinter({})(6);\nprinter({ onBefore: report })(7);\n" + optionalValueFinal, "report 7\ncalls 1 evaluated 1\n"},
		{"an argument writes the binding", optionalValueCallers + "let cb: ((n: number) => void) | undefined = report;\ncb?.((cb = undefined, arg(8)));\ncb?.(arg(9));\n" + optionalValueFinal, "report 8\ncalls 1 evaluated 1\n"},
		{"a closure argument writes the binding", optionalValueCallers + "let cb: ((n: number) => void) | undefined = report;\nfunction clear(): number {\n\tcb = undefined;\n\treturn 10;\n}\ncb?.(clear());\ncb?.(arg(11));\n" + optionalValueFinal, "report 10\ncalls 1 evaluated 0\n"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			for _, extension := range []string{".a", ".ts"} {
				// lowersAndAgreesWithNodeAs holds Node and the JavaScript backend to each other; want, read
				// off Node by hand, pins what both print, and native is held to it.
				program := lowersAndAgreesWithNodeAs(t, probe.source, extension)
				// Agreement alone can't tell the optional call from an ordinary one on the present rows, so
				// the IR is read for it: these rows exist to reach optionalValueCall, not another road.
				if count := optionalCalls(program); count == 0 {
					t.Fatalf("%s: no optional call lowered", extension)
				}
				if code, out, errs := brandedNative(t, program); code != 0 || out != probe.want || errs != "" {
					t.Fatalf("%s native: exit %d stdout %q stderr %q; want exit 0 and %q, as Node prints", extension, code, out, errs, probe.want)
				}
			}
		})
	}
}

// What optionalValueCall doesn't take keeps optionalCall's NotYet, in its words: a callee that may hold
// null, which is held as another representation (includesNull); an element callee; and a call whose
// result the chain goes on reading.
func TestOptionalValueCallGapsStop(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, source, want string }{
		{"a callee that may hold null", "function run(callback: ((n: number) => void) | null | undefined): void {\n\tcallback?.(1);\n}\nrun(null);\n", optionalCallNotYet},
		{"an element callee", "const steps: ((n: number) => void)[] = [];\nsteps[0]?.(1);\n", optionalCallNotYet},
		{"the chain going on after the call", "interface Box {\n\treadonly x: number;\n}\nfunction run(make: (() => Box) | undefined): string {\n\treturn String(make?.().x);\n}\nconsole.log(run(undefined));\n", "an optional chain longer than one step"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSourceAs(t, probe.source, ".a")
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.want) {
				t.Fatalf("want NotYet %q, got %v", probe.want, err)
			}
		})
	}
}
