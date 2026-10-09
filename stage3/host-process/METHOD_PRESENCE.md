Built ambient host presence descriptors; the exact parser probe prints true on native, JavaScript and WASI.
Branch codex/host-method-presence from 35bc60eb; required 722a1c5a merged as ebf310b8; existing WASI guards 047e8572 imported as 9987e378.
Commands: uncached host/WASI oracle PASS 4.663s; compiler method-presence WASI PASS 0.438s; focused lowering PASS 3.774s; native package PASS 290.570s; Linux counts PASS 119.424s; vet PASS.
Mutants: descriptors replaced with absent on native, JavaScript and WASI run with exit 0 and empty stderr, rejected only by Node stdout for all three fixtures.
Not covered: macOS execution, callback scheduling, arbitrary ambient bindings, a full oracle rerun; broader flow remains blocked by inherited private-handle casts and raw if(process.nextTick) is Checker TS2774.

## Contract and assumptions

The compiler's method-observation hook delegates to the Node descriptor before its ambient-method refusal. Callable property declarations also go through that hook in truthiness contexts. process.nextTick uses the existing adamic_node_next_tick_feature primitive. No function value is manufactured, detached or called, and no POSIX call was added. Escaping methods and callback scheduling retain their existing refusals.

The parser-proof 17c1c610 executable probe uses `declare const process: NodeJS.Process` with an otherwise unused node:fs import activating @types/node 25.3.3. This ambient declaration names the existing host global. That is the explicit assumption: it is not a null-initialized native user object. Both its typeof observation and nextTick read must name the host. Declaration/type provenance is checked; an unrelated ambient type or a different global name receives no constant descriptor. Existing compiler tests still reject an ambient user-defined Host with a nextTick method.

All fixtures print strings, including template interpolation of boolean observations, matching the compiler's console contract. Node values, including stdout, stderr and exit, are compared exactly. No private declarations or replacement loader were added.

| Presence observation | Coverage |
| --- | --- |
| process.nextTick | Imported process, Node global, exact ambient Process parser probe; negation, boolean coercion, if/while/do/for conditions, ternary, logical short circuit and nullish comparisons |
| process.cwd, process.memoryUsage | True descriptors on static host receivers; typeof function observations already implemented |
| performance.now, mark, measure, clearMarks, clearMeasures | True descriptors on the imported performance host; existing typeof function tests retained |
| fs.realpathSync.native | True only for the static fs namespace or named realpathSync import; arbitrary receiver expressions are not erased |
| stdout._handle.setBlocking | Existing dynamic handle implementation and typeof feature test retained; no constant descriptor, because handle presence depends on pipe/file/TTY |
| global.gc | No descriptor: this runtime has no GC, and Node presence depends on --expose-gc |
| inspector.Session, process.recordreplay | Not built; no constant descriptor for unsupported inspector/replay hosts |
| process.env flags, stdout.isTTY, terminal columns | Values depend on environment or descriptors, not constant method presence; existing implementations retained |

The inventory comes from sys.ts and stage3/census/data/node_derived_sites.json (core.ts nextTick, performanceCore.ts typeof tests, sys.ts realpath native, handle, gc, inspector and recordreplay). Presence does not make unsupported calls compile. In particular nextTick scheduling is still NotYet and WASI memoryUsage invocation is still target-refused by the existing guard work.

## Language and broader gate limits

Stock checker reproducer: `import process from "node:process"; if (process.nextTick) { console.log("present"); }` produces TS2774 because the method is always defined. No checker policy was changed. Checker-accepted boolean coercions and negations work in truthiness contexts; the compiler owns any change to bare conditions.

The broad flow run was started before the final descriptors and initially also saw transient fixture failures. Its remaining inherited failure is the compiler's refusal of this cast in node_process_blocking.a and node_process_terminal.a:

`import "node:process"; const handle = (process.stdout as { readonly _handle?: { setBlocking(blocking: boolean): void } | undefined })._handle;`

No cast workaround or flow exclusion was added. The final fixture oracle separately builds and verifies SSA for every function of each new fixture. The old broad lowering run caught the QualifiedName panic in the compiler's extracted castProof; the area's identifier guard was retained there, and the full lowering package then passed in 112.060s. The final focused lowering check also passed.

## Merge and verification

The requested remote SHA is on codex/method-presence-test, rather than codex/method-presence. The exact SHA was merged. Conflicts preserve Node declarations and user roots while adopting typed compiler paths; process exit status, regexp UTF-16 views, both input fixture sets and both stage-1 test sets were retained. Old input-oracle callers were adapted to the compiler's preparation callback API. Filesystem-writing fs-file leak observations receive fresh scratch directories.

The owner's existing 047e8572 WASI guards were imported unchanged rather than edited in parallel. All runtime translation units compiled as part of the actual WASI artifact builds. Linux counts were regenerated after the compiler merge. No whole-oracle or full-flow success is claimed. No main push, rebase or force-push occurred.

Setup: GOPROXY=https://proxy.golang.org|direct was set. WASI SDK 27 setup succeeded: Go ready 0.033s, Node ready 0.020s, submodules ready 0.050s, clang ready 0.147s, WASI ready 3.798s, build ready 29.735s, total 29.995s. nproc=5. The exact printed timing lines are retained in logs.

Final commands, all redirected to logs:

```sh
ADAMIC_ORACLE_WASI=1 ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestNodeProcessMethodPresence$|^TestWASIHostRuntimeRefusals$' -count=1 -timeout 30m -v
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASIAgreesWithNode$/internal/oracle/testdata/method_presence.a$' -count=1 -timeout 30m -v
go test ./internal/lower -run '^TestMethodPresence|^TestNodeFSFileQualifiedErrorType|^TestNodeLibrary' -count=1 -timeout 30m
go vet ./internal/load ./internal/lower ./internal/native ./internal/flow ./internal/oracle
```

Full affected package attempt and counts commands are retained in logs too. For macOS confirmation, run the first command without ADAMIC_ORACLE_WASI, unless a WASI SDK is installed; no macOS pass is claimed here.
