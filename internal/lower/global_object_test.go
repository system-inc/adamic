package lower

import (
	"strings"
	"testing"
)

// The global object as a value is one not-yet, whatever name reaches it (global_object.go, #00qn5az head
// 3): each road by which a program could install a host binding's name behind its folded typeof stops
// here, in this stop's words, on its own line, before the guarded read below it. Before head 3 these
// stopped on rules that meet the global object for other reasons (the record view, "reading globalThis",
// node:globals.global, a namespace member, Object.assign's shape, defineProperty's refusal), and lifting
// any of those would have reopened the road. Lifting this stop has to answer the host-binding fold.
// Object.defineProperty returns its first argument, typed as it, so a defineProperty of the global object
// stops at the call, whose own value is the global object, discarded, before its argument's visit.
// Mutant (f), internal/oracle/testdata/host_bindings/mutants.md: an initializer admitted, so a widening to
// object carries the global object past this stop, and the object-view rows stop elsewhere.
func TestGlobalObjectAsAValueIsNotYet(t *testing.T) {
	t.Parallel()
	const stop = "stage 0 can't lower the global object as a value (only a property read's or member call's receiver is admitted) yet"
	const handler = "const fire = (eventName: string): void => {\n\tconsole.log(eventName);\n};\n"
	const guarded = "export function run(name: string): void {\n\tif (typeof onProfilerEvent === \"function\") {\n\t\tonProfilerEvent(name);\n\t}\n}\n"
	const node = "import process from 'node:process';\n"
	for _, probe := range []struct{ name, imports, definition, where string }{
		{"Object.defineProperty on globalThis", "", "Object.defineProperty(globalThis, \"onProfilerEvent\", { value: fire });\n", "5:1"},
		{"Reflect.set on globalThis", "", "Reflect.set(globalThis, \"onProfilerEvent\", fire);\n", "5:13"},
		{"Object.assign on globalThis", "", "Object.assign(globalThis, { onProfilerEvent: fire });\n", "5:15"},
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
// of the object compared with a literal string, which lowering still reaches as "reading globalThis".
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
