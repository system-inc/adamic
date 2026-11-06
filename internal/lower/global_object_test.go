package lower

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// The global object as a value is one not-yet, whatever name reaches it (global_object.go, #00qn5az head
// 3): each road by which a program could install a host binding's name behind its folded typeof stops
// here, in this stop's words, on its own line, before the guarded read below it. Before head 3 these
// stopped on rules that meet the global object for other reasons (the record view, "reading globalThis",
// node:globals.global, a namespace member, Object.assign's shape, defineProperty's refusal), and lifting
// any of those would have reopened the road. Lifting this stop has to answer the host-binding fold.
// Object.defineProperty returns its first argument, typed as it, and Object.assign its first argument as
// an intersection with the second, so either on the global object stops at the call, whose own value
// is (or holds) the global object, discarded, before its argument's visit.
// globalThis.valueOf() is the global object typed Object, so a read or call through the global object
// whose result typeof globalThis is assignable to stops the same way (globalObjectResult), through every
// spelling of the receiver. Mutant (h), internal/oracle/testdata/host_bindings/mutants.md: an initializer
// admitted, so a widening to object carries the global object past this stop, and the object-view rows
// stop elsewhere. Mutant (i): no result judged, and the valueOf rows stop elsewhere.
func TestGlobalObjectAsAValueIsNotYet(t *testing.T) {
	t.Parallel()
	const stop = "stage 0 can't lower the global object as a value (only a property read's or member call's receiver is admitted) yet"
	const handler = "const fire = (eventName: string): void => {\n\tconsole.log(eventName);\n};\n"
	const guarded = "export function run(name: string): void {\n\tif (typeof onProfilerEvent === \"function\") {\n\t\tonProfilerEvent(name);\n\t}\n}\n"
	const node = "import process from 'node:process';\n"
	for _, probe := range []struct{ name, imports, definition, where string }{
		{"Object.defineProperty on globalThis", "", "Object.defineProperty(globalThis, \"onProfilerEvent\", { value: fire });\n", "5:1"},
		{"Reflect.set on globalThis", "", "Reflect.set(globalThis, \"onProfilerEvent\", fire);\n", "5:13"},
		{"Object.assign on globalThis", "", "Object.assign(globalThis, { onProfilerEvent: fire });\n", "5:1"},
		{"a Record<string, unknown> view", "", "const g: Record<string, unknown> = globalThis;\ng[\"onProfilerEvent\"] = fire;\n", "5:36"},
		{"an index-signature view", "", "const g: { [key: string]: unknown } = globalThis;\ng[\"onProfilerEvent\"] = fire;\n", "5:39"},
		{"Reflect.set through an object view", "", "const g: object = globalThis;\nReflect.set(g, \"onProfilerEvent\", fire);\n", "5:19"},
		{"Object.assign through an object view", "", "const g: object = globalThis;\nObject.assign(g, { onProfilerEvent: fire });\n", "5:19"},
		{"Object.defineProperty through an object view", "", "const g: object = globalThis;\nObject.defineProperty(g, \"onProfilerEvent\", { value: fire });\n", "5:19"},
		{"Object.assign through an empty-object view", "", "const g: {} = globalThis;\nObject.assign(g, { onProfilerEvent: fire });\n", "5:15"},
		{"Reflect.set through an alias", "", "const g = globalThis;\nReflect.set(g, \"onProfilerEvent\", fire);\n", "5:11"},
		{"Object.assign through an alias", "", "const g = globalThis;\nObject.assign(g, { onProfilerEvent: fire });\n", "5:11"},
		{"Object.defineProperty through an alias", "", "const g = globalThis;\nObject.defineProperty(g, \"onProfilerEvent\", { value: fire });\n", "5:11"},
		{"Object.defineProperty through Node's global", node, "console.log(typeof process.argv);\nObject.defineProperty(global, \"onProfilerEvent\", { value: fire });\n", "7:1"},
		{"Reflect.set through globalThis.globalThis", "", "Reflect.set(globalThis.globalThis, \"onProfilerEvent\", fire);\n", "5:13"},
		{"Reflect.set through parentheses", "", "Reflect.set((globalThis), \"onProfilerEvent\", fire);\n", "5:13"},
		{"a return", "", "function host(): object {\n\treturn globalThis;\n}\nReflect.set(host(), \"onProfilerEvent\", fire);\n", "6:9"},
		{"a spread", "", "const copy = { ...globalThis };\nconsole.log(typeof copy);\n", "5:19"},
		{"an element of an array", "", "const hosts = [globalThis];\nReflect.set(hosts[0] ?? {}, \"onProfilerEvent\", fire);\n", "5:16"},
		{"a shorthand property", "", "const holder = { globalThis };\nReflect.set(holder.globalThis, \"onProfilerEvent\", fire);\n", "5:18"},
		{"a condition", "", "if (globalThis) {\n\tconsole.log('host');\n}\n", "5:5"},
		{"typeof read by something other than a literal comparison", "", "console.log(typeof globalThis);\n", "5:20"},
		{"Reflect.set on globalThis.valueOf()", "", "Reflect.set(globalThis.valueOf(), \"onProfilerEvent\", fire);\n", "5:13"},
		{"Object.assign on globalThis.valueOf()", "", "Object.assign(globalThis.valueOf(), { onProfilerEvent: fire });\n", "5:15"},
		{"globalThis.valueOf() widened to object", "", "const host: object = globalThis.valueOf();\nReflect.set(host, \"onProfilerEvent\", fire);\n", "5:22"},
		{"valueOf through globalThis.globalThis", "", "Reflect.set(globalThis.globalThis.valueOf(), \"onProfilerEvent\", fire);\n", "5:13"},
		{"valueOf through an element-access callee", "", "Reflect.set(globalThis[\"valueOf\"](), \"onProfilerEvent\", fire);\n", "5:13"},
		{"valueOf through an optional call", "", "Reflect.set(globalThis?.valueOf(), \"onProfilerEvent\", fire);\n", "5:13"},
		{"valueOf through Node's global", node, "console.log(typeof process.argv);\nReflect.set(global.valueOf(), \"onProfilerEvent\", fire);\n", "7:13"},
		{"Object.defineProperty on globalThis.valueOf()", "", "Object.defineProperty(globalThis.valueOf(), \"onProfilerEvent\", { value: fire });\n", "5:23"},
		{"valueOf of valueOf", "", "Reflect.set(globalThis.valueOf().valueOf(), \"onProfilerEvent\", fire);\n", "5:13"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, map[string]string{
				"performance.ts": probe.imports + profilerDeclaration + handler + probe.definition + guarded,
				"main.ts":        "import { run } from './performance.ts';\nrun('one');\n",
			})
			if err == nil || !strings.HasSuffix(err.Error(), "performance.ts:"+probe.where+": "+stop) {
				t.Errorf("got %v, want performance.ts:%s: %s", err, probe.where, stop)
			}
		})
	}
}

// The uses the ruling admits keep the stops or the lowering they had before head 3: a property read and a
// member call through Node's global (sys.ts's global.gc), a write through it, judged by rule 50, and typeof
// of the object, or of its valueOf(), compared with a literal string, which lowering still reaches as
// "reading globalThis". A member whose result is a string (toString) isn't a result the global object
// could be, and keeps its stop too.
func TestGlobalObjectAsAReceiverIsAdmitted(t *testing.T) {
	t.Parallel()
	t.Run("global.gc read and called", func(t *testing.T) {
		t.Parallel()
		if _, err := lowerModules(t, map[string]string{"main.ts": "import 'node:process';\nif (global.gc) {\n\tglobal.gc();\n}\nconsole.log(String(!!global.gc));\n"}); err != nil {
			t.Fatalf("want it to lower, got %v", err)
		}
	})
	for _, probe := range []struct{ name, source, want string }{
		{"a write through global, rule 50's", "import process from 'node:process';\nglobal.process = process;\n", "main.ts:2:1: Adamic 0.1 refuses a write through globalThis to the host binding process; " + hostBindingWriteFix},
		{"typeof compared with a literal", "console.log(String(typeof globalThis === 'object'));\n", "main.ts:1:27: stage 0 can't lower reading globalThis yet"},
		{"typeof globalThis.valueOf() compared with a literal", "console.log(String(typeof globalThis.valueOf() === 'object'));\n", "main.ts:1:27: stage 0 can't lower reading globalThis yet"},
		{"globalThis.toString(), a string", "console.log(globalThis.toString());\n", "main.ts:1:13: stage 0 can't lower reading globalThis yet"},
		{"another member of global", "import 'node:process';\nglobal.console.log('through global');\n", "main.ts:2:1: stage 0 can't lower node:globals.global other than as global.gc's receiver (fix: name the global itself, without global.) yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, map[string]string{"main.ts": probe.source})
			if err == nil || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want %s", err, probe.want)
			}
		})
	}
}

// loadedOrLowered loads modules from main.ts as lowerModules does, and returns the checker's error where
// the program doesn't check, else lowering's.
func loadedOrLowered(t *testing.T, modules map[string]string) error {
	t.Helper()
	directory := t.TempDir()
	for name, source := range modules {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	program, err := load.Load([]string{filepath.Join(directory, "main.ts")})
	if err != nil {
		return err
	}
	_, err = Lower(context.Background(), program)
	return err
}

// A host binding's name read through an admitted receiver of the global object never lowers to a
// native read where Node would say "function". onProfilerEvent can't be read so in a program the checker
// passes: its declare let is the module's (load.go's ModuleDetection), never a property of typeof
// globalThis, so the property read is TS7017 and an element read by a key that is or may be it TS7053.
// globalThis.process is Node's own process, which @types/node puts on typeof globalThis, not the process
// row (whose declaration is core.ts's declare const, matched by symbol); its reads stop today on rules
// unrelated to the row, a method's presence test and a namespace object as a value, pinned at their text
// so lifting either is seen here. A checker error's first line is pinned; TS7053 adds a second.
func TestGlobalObjectReceiverReadsOfAHostBinding(t *testing.T) {
	t.Parallel()
	const performance = profilerDeclaration + "export function run(name: string): void {\n\tif (typeof onProfilerEvent === \"function\") {\n\t\tonProfilerEvent(name);\n\t}\n}\n"
	for _, probe := range []struct{ name, main, want string }{
		{"globalThis.onProfilerEvent", "import { run } from './performance.ts';\nrun('one');\nconsole.log(typeof globalThis.onProfilerEvent);\n", "main.ts:3:31: error TS7017: Element implicitly has an 'any' type because type 'typeof globalThis' has no index signature."},
		{"globalThis['onProfilerEvent']", "import { run } from './performance.ts';\nrun('one');\nconsole.log(typeof globalThis['onProfilerEvent']);\n", "main.ts:3:20: error TS7053: Element implicitly has an 'any' type because expression of type '\"onProfilerEvent\"' can't be used to index type 'typeof globalThis'."},
		{"globalThis[k], k onProfilerEvent or gc", "import process from 'node:process';\nimport { run } from './performance.ts';\nrun('one');\nconst k: 'onProfilerEvent' | 'gc' = process.argv.length > 9 ? 'onProfilerEvent' : 'gc';\nconsole.log(typeof globalThis[k]);\n", "main.ts:5:20: error TS7053: Element implicitly has an 'any' type because expression of type '\"gc\" | \"onProfilerEvent\"' can't be used to index type 'typeof globalThis'."},
		{"globalThis.process.nextTick", "import process from 'node:process';\nconsole.log(typeof process.argv);\nconsole.log(typeof globalThis.process.nextTick);\n", "main.ts:3:20: stage 0 can't lower a presence test of a method whose presence the receiver's construction doesn't prove; a view or host value has to read it yet"},
		{"globalThis['process']", "import process from 'node:process';\nconsole.log(typeof process.argv);\nconsole.log(typeof globalThis['process']);\n", "main.ts:3:31: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet"},
		{"globalThis.process held", "import process from 'node:process';\nconsole.log(typeof process.argv);\nconst held = globalThis.process;\nconsole.log(typeof held);\n", "main.ts:3:14: stage 0 can't lower a namespace object used as a value; no runtime container is emitted, so identity, receiver behavior, live export aliases and staged properties are not represented; use qualified members or named module imports yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			err := loadedOrLowered(t, map[string]string{"main.ts": probe.main, "performance.ts": performance})
			if err == nil || !strings.Contains(err.Error(), probe.want) {
				t.Errorf("got %v, want %s", err, probe.want)
			}
		})
	}
}

// Top-level this is no road to the global object: every Adamic file is a module (load.go's
// ModuleDetection, even one with no import or export), so the checker types it undefined, as the second
// row's TS2322 shows, and lowering stops it as this outside a method. ThisKeyword still isn't among the
// kinds the global object's rule skips: a this parameter may be typeof globalThis, and install's is, so
// Reflect.set(this, ...) stops there as the global object as a value.
func TestGlobalObjectTopLevelThis(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, main, want string }{
		{"top-level this", "const g = this;\nconsole.log(typeof g);\n", "main.ts:1:11: stage 0 can't lower this outside a method yet"},
		{"top-level this as the global object", "const g: typeof globalThis = this;\nconsole.log(typeof g);\n", "main.ts:1:7: error TS2322: Type 'undefined' is not assignable to type 'typeof globalThis'."},
		{"a this parameter of the global object's type", "const fire = (eventName: string): void => {\n\tconsole.log(eventName);\n};\nfunction install(this: typeof globalThis): void {\n\tReflect.set(this, 'onProfilerEvent', fire);\n}\ninstall.call(globalThis);\n", "main.ts:5:14: stage 0 can't lower the global object as a value (only a property read's or member call's receiver is admitted) yet"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			err := loadedOrLowered(t, map[string]string{"main.ts": probe.main})
			if err == nil || !strings.HasSuffix(err.Error(), probe.want) {
				t.Errorf("got %v, want %s", err, probe.want)
			}
		})
	}
}
