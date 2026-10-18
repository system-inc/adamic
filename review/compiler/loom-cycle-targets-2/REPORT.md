Fixed the stale ambient cwd refusal expectation and retained the closure-target cycle rule.
Source 06c8cb968751488b11531de609d5c81c814a3116; base 199174800998904d8d646dfa5892f512cd5cb20f; initial reproduction squash 90577c42d2835dbb957944da0da428a3b2e34016; delivery is one commit on this base.
Checks passed: 292 lower leaves, 25 census/cycle leaves, five affected oracle fixtures, closure targets/readers, 16 count rows and table structure, and vet.
All five original mutants failed their intended assertions; restoring type-matched captures also failed both new cwd admission tests.
Not covered: the full repository gate, whole oracle/native packages, unrelated stage1/cohere/json reds, or a renewed whole-corpus admission census; the first green push exceeded 30 minutes.

The cwd case reads stage3/namespace-init-sys/cwd.a whole and gives its text to lowerSource, which writes a temporary main.a before loading and lowering. This is the exact program, including its historical, inert a-check annotation:

```typescript
// a-check: refused 'callback', a variable a function value captures and can be reached from what it holds
import * as process from 'node:process';
// sys.ts:1502:51: the callback reaches the ambient process declaration.
function memoize(callback: () => string): () => string {
    let value: string | undefined;
    return () => { if (value === undefined) value = callback(); return value; };
}
const getCurrentDirectory = memoize(() => process.cwd());
namespace Debug { export let enabled = false; }
console.log(`${getCurrentDirectory() === process.cwd()}`);
```

Observed on plain origin/main: TestNamespaceAmbientHostInitialization passed in 1.233 seconds. Its cwd stop was main.a:4:18, Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free, with adamic/cycle-capable. The realpath witness independently stopped at main.a:3:22, stage 0 can't lower node:fs.native yet. See main-baseline-ready.log.

Observed on the requested squash, with Node declarations installed: the exact requested focused command failed in 1.321 seconds with namespaces_ambient_test.go:42, expected independent host member NotYet, got <nil>. See reproduce.log. Namespace preflight itself succeeded for both ambient host declarations.

The cause is a false ownership cycle from function-type matching. The old finder considered every closure assignable to () => string as a possible value of callback. The returned memoizer has that same signature and captures callback, so matching it to the callback invented a path back to callback's cell. ClosureTargets follows the actual producer supplied by the direct memoize call, () => process.cwd(). Lowering recognizes process.cwd as a host intrinsic, with no captured process object or local environment. The returned closure owns the callback cell and the string cache cell; the actual callback and immutable string cannot reach the returned closure or the callback cell. This is the closed-program interpretation already required by the rule. Unknown origins and transitive strong captures remain conservative.

Observed soundness evidence: source Node prints true and exits 0. The JavaScript backend, release native, and native with ASan, UBSan and LeakSanitizer all print exactly true followed by a newline, exit 0, and have empty stderr. The initial uncached oracle leaf took 35.19 seconds including runtime product preparation; its fixture-suite run took 3.60 seconds. The native count row is allocations 6, frees 6, retains 8, releases 12, peak 6, regions 0. The agreement and leak evidence support admitting this particular program; they do not claim a proof of every possible host program.

The fix keeps the production rule from the source branch. TestNamespaceAmbientHostInitialization retains executable namespace early-read rejection, ambient preflight, and the independent realpath host stop. The cwd expectation is now a separate top-level TestNamespaceAmbientCwdInitialization requiring source/JavaScript agreement. TestClosureCycleAmbientCwd registers and holds the exact existing stage3 source to Node, both backends and sanitizers; counts.md records its new row. This repair leaves the historical stage3 input untouched; its additional Go test edits are the named lower file and the cycle rule's oracle file.

Every one of internal/lower's 292 top-level tests ran separately, with GOMAXPROCS=4, -test.parallel=4, -test.timeout=60s and a 60-second subprocess limit. There are no Example or Fuzz roots in this package. All 292 returned 0, totaling 163.481 seconds of leaf wall time. The slowest was TestPredicateOverloadRuntime, 11.471 seconds. The touched/new leaves were TestNamespaceAmbientHostInitialization 0.882 seconds, TestNamespaceAmbientCwdInitialization 0.872 seconds, and TestNamespaceAmbientContextsDoNotExecute 0.272 seconds. Exact per-leaf commands, exits and seconds are in lower-results.json; individual logs and the complete inventory are beside it. No over-budget test leaf was rerun.

TestCallTargetReaders passed in 47.71 seconds, including its 46.06-second listing product miss. The census/cycle suite passed in 25.971 seconds; its slowest leaf was TestGettersCensusBinaryInline, 19.98 seconds. The four numbered affected fixture cases passed in that suite. user_iterators now has a dedicated family wrapper and was therefore run separately; its fixture leaf passed in 20.18 seconds. The fifteen inherited cycle-related count rows matched their records, the new cwd row matched after update, and TestCountsAreRecordedTable passed. Vet on internal/ir, internal/lower and internal/oracle exited 0 with no diagnostics.

The mutation evidence was rerun from the final production source and restored after each independent mutation. These are assertion failures, not compiler/build-warning kills:

| Mutant | Intended catcher | Observed failure |
|---|---|---|
| omit-boundary-inputs | TestClosureTargetsKnownCallerDoesNotHideUnknownCaller | known caller hid unbounded caller |
| unknown-is-empty | TestClosureTargetsBoundOnlyProvenValues | bounded without proof; Unknown failed to include throwing function |
| omit-captures | TestClosureCycleDirect | want cycle refusal, got nil |
| drop-return-target | TestClosureTargetsFollowParametersReturnsAndMapValues | want both returned targets |
| omit-map-captures | TestClosureCycleMap | want cycle refusal, got nil |
| type-matched-captures | TestNamespaceAmbientCwdInitialization and TestClosureCycleAmbientCwd | both fail with the old callback cycle refusal at line 4:18 |

Mutant sources are .patch files. The runners are run-mutants.py and run-admission-mutant.py. The first preparation of the additional mutant stopped before mutation because main has reaches rather than reachesFrom; the corrected runner restores the old type-matching behavior at the two actual producer-graph sites and both assertions fail. The final runner gives the child a 120-second limit inside the 180-second outer limit, preserving time for restoration.

Setup and delivery timing: nproc is 5, cpu.max is 400000 100000 (four CPUs). Initial setup printed Go 0.064s, Node 0.073s, clang 0.478s, markdown dependencies 1.169s, submodules 591.590s, shared cache 599.472s, then the 600-second limit expired during warming, exit 124. The 240-second retry with the optional shared cache disabled also expired during cold export/build work. Several externally bounded cold test builds expired before producing test output; their compilation was still visible before the limits. An earlier attempt failed to load cohere/TypeScript/tsc/go.mod while cloning was incomplete, and the first executable focused attempt stopped at missing locked @types/node 25.3.3. npm ci --prefix stage3/api installed the locked three packages in two seconds. Final setup, with ADAMIC_GOCACHE_OFF=1, completed: Go 0.029s, Node 0.034s, markdown 0.118s, submodules 0.141s, clang 0.307s, shared cache off 0.309s, go build 60.761s, cache warm 60.924s, done 60.997s. It printed source /workspace/adamic-tools/env.sh, which subsequent commands sourced. Tool versions were Go 1.27.1, Node v24.19.0 and clang 20.1.8. setup.log, setup-retry.log and setup-final.log preserve the output. Overlapping cold builds duplicated work; remaining builds and source-mutating runners were serialized. These setup/build delays and verification exceeded the requested first-push window.

The environment briefly returned a starting marker during verification. Its status tool still reported running/connected, wait returned ready, and retained lower-results.json and logs showed all leaves completed. Old tool session IDs were no longer available, so the complete per-leaf exit metadata and terminal PASS logs were inspected directly. No completed green tests were restarted because of that marker.

The first consolidated commit command stopped at the strict whitespace check because unified-diff context lines encode a space before Go indentation tabs. Those evidence bytes were preserved. Actual docs/internal changes were checked with the strict whitespace rules; review evidence was checked with only space-before-tab disabled. The repository ignores new .log files, so all retained logs were explicitly staged with git add -f. Commands, outputs and failed attempts are listed in COMMANDS.md. This report supersedes the source branch's historical evidence where the current grain and named wrappers differ. The source branch's original evidence remains in review/compiler/cycle-closure-targets. No code was copied out of cohere, no full gate or PR was run, no Loom submission was made, and delivery pushes only compiler/loom-cycle-targets-2.
