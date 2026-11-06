package lower

import (
	"errors"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// watchTscSource is sys.ts's own shape (tsc 6.0.3, :891-897, :1262-1283 and :1772-1784): the options a
// run-time conditional of two literals, the listener a parameter and an arrow, the result seen as tsc's
// FsWatchWorkerWatcher through a union of the two workers, its on("error") and its close.
const watchTscSource = `import * as fs from 'node:fs';
type FsWatchCallback = (eventName: "rename" | "change", relativeFileName: string | undefined | null, modifiedTime?: Date) => void;
interface FileWatcher { close(): void; }
interface FsWatchWorkerWatcher extends FileWatcher { on(eventName: string, listener: () => void): void; }
const fsSupportsRecursiveFsWatch = process.platform === "win32" || process.platform === "darwin";
function fsWatchWorker(fileOrDirectory: string, recursive: boolean, callback: FsWatchCallback) {
 return fs.watch(fileOrDirectory, fsSupportsRecursiveFsWatch ? { persistent: true, recursive: !!recursive } : { persistent: true }, callback);
}
function fsWatchWorkerHandlingTimestamp(fileOrDirectory: string, recursive: boolean, callback: FsWatchCallback): FsWatchWorkerWatcher {
 return fsWatchWorker(fileOrDirectory, recursive, (eventName, relativeFileName, currentModifiedTime) => { callback(eventName, relativeFileName, currentModifiedTime); });
}
function watchPresent(path: string, timestamps: boolean, callback: FsWatchCallback): FileWatcher {
 const presentWatcher = (timestamps ? fsWatchWorkerHandlingTimestamp : fsWatchWorker)(path, false, callback);
 presentWatcher.on("error", () => { callback("rename", ""); });
 return presentWatcher;
}
watchPresent('.', false, (eventName, relativeFileName) => { console.log(eventName + (relativeFileName ?? 'none')); }).close();`

func watchOperations(program *ir.Program) map[string]int {
	operations := map[string]int{}
	visit := func(node any) bool {
		if call, ok := node.(ir.NodeFSFile); ok {
			operations[call.Operation]++
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return operations
}

func TestNodeWatchTscShape(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, watchTscSource)
	if err != nil {
		t.Fatal(err)
	}
	if !program.HasPromises() {
		t.Fatal("fs.watch must run the loop without an explicit Promise")
	}
	if operations := watchOperations(program); operations["host_watch"] != 1 {
		t.Fatal(operations)
	}
}

func TestNodeWatchDirectForms(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `import { watch } from 'node:fs';
const watcher = watch('.', {persistent: false}, (eventType, filename) => { console.log(eventType + (filename ?? 'none')); });
const same = watcher.on('error', () => { console.log('error'); });
console.log(same === watcher ? 'same' : 'other');
watcher.close();
watcher.close();`)
	if err != nil {
		t.Fatal(err)
	}
	if operations := watchOperations(program); operations["host_watch"] != 1 || operations["host_watch_on"] != 1 {
		t.Fatal(operations)
	}
}

func TestNodeWatchOtherFormsRefused(t *testing.T) {
	t.Parallel()
	prefix := `import * as fs from 'node:fs';
const flag = fs.existsSync('adamic-fswatch-flag');
function listener(_eventType: string, _filename: string | null): void {}
const watcher = fs.watch('.', {persistent: true}, listener);
`
	for _, probe := range []struct{ name, code, want string }{
		{"arguments", `fs.watch('.', listener);`, "node:fs.watch outside fs.watch(path, options, listener)"},
		{"option", `fs.watch('.', {persistent: true, encoding: 'utf8'}, listener);`, "node:fs.watch with the option encoding (only persistent and recursive are bound)"},
		{"persistent missing", `fs.watch('.', {recursive: false}, listener);`, "node:fs.watch without a literal persistent"},
		{"persistent run time", `fs.watch('.', {persistent: flag}, listener);`, "node:fs.watch with a persistent other than true or false"},
		{"arms differ", `fs.watch('.', flag ? {persistent: true} : {persistent: false}, listener);`, "node:fs.watch with options arms whose persistent differs"},
		{"rest listener", `fs.watch('.', {persistent: true}, (_eventType: string, _filename: string | null, ...rest: string[]) => { console.log(rest.length); });`, "node:fs.watch with a generic, this-typed or rest-parameter listener"},
		{"numeric slot", `fs.watch('.', {persistent: true}, (_eventType: string, _filename: string | null, _count?: number) => {});`, "node:fs.watch with a listener whose parameter _count can't hold what Node hands it"},
		{"other event", `watcher.on('close', () => {});`, `node:fs.FSWatcher.on with an event other than the literal "error"`},
		{"detached close", `const close = watcher.close; close();`, "node:fs.FSWatcher.close used other than called on its watcher"},
		{"detached on", `console.log(typeof watcher.on);`, "node:fs.FSWatcher.on used other than called on its watcher"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, prefix+probe.code)
			var notYet *NotYet
			if !errors.As(err, &notYet) || notYet.What != probe.want {
				t.Fatalf("want exact NotYet %q, got %v", probe.want, err)
			}
		})
	}
}

func TestNodeWatchUsesLibrarySymbols(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function watch(text: string): string { return text; }
interface Closer { close(): void; on(eventName: string, listener: () => void): void; }
const closer: Closer = { close: () => { console.log('closed'); }, on: (_eventName, listener) => { listener(); } };
closer.on('error', () => { console.log(watch('own')); });
closer.close();`)
	if err != nil {
		t.Fatal(err)
	}
	if operations := watchOperations(program); len(operations) != 0 {
		t.Fatal(operations)
	}
}

// inheritedOnSources are another host type that inherits on from the declaration FSWatcher does,
// called directly and seen through an on view: neither has an on its native shape holds, so both
// stay NotYet while only node:fs.FSWatcher.on is registered (Library's must on #j71fn59).
var inheritedOnSources = []struct{ name, code, want string }{
	{"direct", `import type { ChildProcess } from 'node:child_process';
function listen(child: ChildProcess): void { child.on('error', () => {}); }
console.log(typeof listen);`, "node:events.InternalEventEmitter.on"},
	{"view", `import type { ChildProcess } from 'node:child_process';
function listen(child: ChildProcess): void { const view: { on(eventName: string, listener: () => void): void } = child; view.on('error', () => {}); }
console.log(typeof listen);`, "node:events.InternalEventEmitter.on, a member no host registration names"},
}

func TestNodeWatchInheritedOnStaysNotYet(t *testing.T) {
	t.Parallel()
	for _, probe := range inheritedOnSources {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, probe.code)
			var notYet *NotYet
			if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.want) {
				t.Fatalf("want a NotYet naming %q, got %v", probe.want, err)
			}
		})
	}
}

// Library's mutant: the inherited declaration registered instead of FSWatcher's own name lets the
// second host type lower, which TestNodeWatchInheritedOnStaysNotYet refuses. Not parallel: it changes
// the registration every lowering reads, while the parallel tests wait.
func TestNodeWatchInheritedOnMutant(t *testing.T) {
	const inherited = "node:events.InternalEventEmitter.on"
	if implementedNodeMembers[inherited] {
		t.Fatal(inherited + " is registered: the pin's premise is gone")
	}
	implementedNodeMembers[inherited] = true
	defer delete(implementedNodeMembers, inherited)
	for _, probe := range inheritedOnSources {
		if _, err := lowerSource(t, probe.code); err != nil {
			t.Fatalf("%s: the mutant must let the second host type lower: %v", probe.name, err)
		}
	}
	t.Log("TestNodeWatchInheritedOnStaysNotYet goes red with the inherited declaration registered")
}

// An 'error' listener whose parameter Node's Error can't be type-checks only through on's general
// overload, whose listener takes ...args: any[] (events.d.ts:878). There the method relation refuses
// it before lowering, so the any never hands the program a value (host-any.md's Modelled entry). A
// listener the Error can be picks the typed overload instead. Mutant: the relation's parameter
// refusal skipped (invariance.go, widened's callable parameters) lets each of these lower.
func TestNodeWatchErrorListenerOnTheAnyOverloadIsRefused(t *testing.T) {
	t.Parallel()
	prefix := `import * as fs from 'node:fs';
function listener(_eventType: string, _filename: string | null): void {}
const watcher = fs.watch('.', {persistent: true}, listener);
`
	for _, probe := range []struct{ name, code, taking string }{
		{"map", `watcher.on('error', (seen: Map<string, number>) => { console.log(seen.get('x')); });`, "Map<string, number>"},
		{"class", "class Shape { area = 1; }\nwatcher.on('error', (seen: Shape) => { console.log(seen.area); });", "Shape"},
		{"primitive union", `watcher.on('error', (seen: number | boolean) => { console.log(typeof seen); });`, "number | boolean"},
		{"string union", `watcher.on('error', (seen: string | number) => { console.log(typeof seen); });`, "string | number"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			_, err := lowerSource(t, prefix+probe.code)
			var refused *Refused
			if !errors.As(err, &refused) || !strings.HasPrefix(refused.What, "a function taking "+probe.taking+" seen as one taking ") {
				t.Fatalf("want the method relation's refusal of a function taking %s, got %v", probe.taking, err)
			}
		})
	}
}
