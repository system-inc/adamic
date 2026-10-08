# Step 29 library startup and output binding

Branch `codex/host29-startup-output`, based on library `71f91687` with
host surface `91215715` merged. The supplied refs have no `AGENTS.md`;
repository guidance was read from `CLAUDE.md`, `README.md`, `docs/0.1.md`
and `docs/memory.md`. This unit changes library lowering and its tests only.
Runtime `8a0d1803` is merged at `f2173aef` into checkpoint `5cb3db47`.
Runtime's C primitive files are untouched after that merge.

The existing library already lowers `process.cwd`, `argv`, environment reads,
stdout writes, numeric exit and exitCode, performance and memoryUsage to native
calls. This checkpoint makes environment assignments and deletion explicitly
Refused under the step's live-read-only ruling. The Node declarations accept
string exit codes, but this base's native ABI represents numeric codes only;
those overloads now produce NotYet before emission instead of an invalid fit.

## Unchanged scout sources

The tests use the source fixtures retained by runtime scout `2cdc201b` without
rewriting their sys.ts bodies. Node is 24.19.0, on Linux.

| Fixture | Native and JavaScript | WASI | First compiler prerequisite |
| --- | --- | --- | --- |
| 14 getCurrentDirectory | Refused | No emission | `callback = undefined!` is a refused non-null assertion at 16:24 |
| 15 getExecutingFilePath | Agrees with Node | TargetRefused | Its scratch setup calls `fs.mkdtempSync`, unavailable on WASI |
| 16 getEnvironmentVariable | NotYet | No emission | String `process.env[name] \|\| ""` at 11:12; fixture environment writes would also be Refused |
| 17 write | NotYet | No emission | Object literal void method result at 17:5 is unrepresented |
| 18, 19, 20 exit | NotYet | No emission | Declared `activeSession: undefined` at 13:5 is unrepresented |
| 23 newLine | Agrees with Node | TargetRefused | Its scratch setup calls `fs.mkdtempSync`, unavailable on WASI |

Fixture 15 shadows `__filename` and `__dirname` with local strings. Its success
holds the sys.ts branch expression; the separate identity probe holds the
actual executable globals.
No source-shape approximation is advertised as closing a compiler prerequisite.
WASI refusal is recorded separately so it cannot hide native/JavaScript results.

The binding-only `node_startup_environment.a` probe uses a nullish fallback to
exercise the live getter and distinguish missing from empty. It is explicitly
not fixture 16's unsupported truthy fallback. Externally supplied empty, zero
and Unicode values, and an absent key, agree on native, JavaScript and WASI.
These environment-dependent runs disable the oracle cache: its source Node
cache does not distinguish the externally supplied values.

## Mutants

Every scout source mutant runs cleanly on Node and disagrees only with the
original Node result: retain the memoize callback (14), test `sys.cjs` instead
of `sys.js` (15), return empty environment text (16), append `?` to writes (17),
change exit 0 to 1, 1 to 2, and 2 to 0 (18,19,20), and use CRLF (23).
The 15 and 23 mutants also compile and run without sanitizer errors on native
and JavaScript, and fail only the Node stdout comparison. Compiler-blocked
fixtures have no backend mutant success claim. The separate getter's empty-text
mutant compiles and runs cleanly on native, JavaScript and WASI and fails only
the Node stdout comparison.

## Runtime seam bound

The binding delegates to the C ABI documented in the runtime contract section
of [scout-29-host.md](scout-29-host.md). Environment reads call the live getter;
cwd preserves Node's getter cache until a successful chdir. Generated adapters
convert cwd failures to catchable Errors with code, errno, syscall and message.
argv and execArgv each retain one stable array.

process.execPath and CommonJS __filename/__dirname use OS executable identity.
Checked constant builtin requires with matching `typeof import(...)` annotations
lower through the same native namespace bindings as imports. Dynamic requests,
untyped any, observing module objects and source-map hooks remain NotYet. This
does not claim tsc's unannotated any requires or dynamic module loading.
Native identity builds ship every real bundled lib.*.d.ts beside the executable;
different existing files are refused. A completeness test checks every file.

stdout is one stable owned receiver; its private handle probe is undefined.
Unsupported reflection on the opaque receiver is refused. Raw writes delegate
to runtime's blocking writer without adding a newline. Numeric validation
precedes exitCode writes; getters and natural main status use runtime's status
primitives. Explicit exit calls runtime's terminating primitive, without
unwinding finally. JavaScript calls Node's exit directly. String exit overloads,
write callbacks and observed backpressure remain NotYet.

performance.now/timeOrigin and Date.now use runtime's new clocks, including the
stable performance receiver's clock method. Existing mark/measure behavior is
retained. memoryUsage exposes all five native fields. These are native metrics,
not V8 heap totals: comparisons cover field names and finite/nonnegative bounds,
not exact totals or V8 relationships. Runtime tracks numeric typed-array backing;
additional foreign/Buffer backing accounting is not claimed here.

The new CommonJS identity probe passes on native and JavaScript, including
stable argv and default-library lookup, under a raw Node CommonJS control.
Memory passes on those backends. Clocks/stdout/cwd, environment and direct exit
0/1/2 with finally bodies pass on native, JavaScript and WASI. Identity/argv and
memory are explicitly TargetRefused on WASI under runtime's contract.
These probes do not approximate or close the blocked scout source shapes.

Binding source mutants reverse clock ordering, negate the RSS bound, break
dirname identity and change exit 0/1/2 to 1/2/0. Each compiles and runs cleanly,
and fails only Node stdout/status comparison on every supported backend.
Related process mutants now target generated adapters when the old implementation
was replaced. Runtime's primitive contract and mutant tests also pass.

## Validation

Run output stays outside the checkout. The finalized targeted command is:

```text
source /workspace/adamic-tools/env.sh
ADAMIC_GATE_UNCACHED=1 ADAMIC_ORACLE_WASI=1 go test ./internal/lower ./internal/oracle -run '^TestNodeStartup' -count=1 -v -timeout 10m
```

Scoped lower/native/oracle comparisons, default-library completeness, field
layout proofs and vet pass, with the explicit compiler/target skips above.
Logs: `/tmp/host29-final-verification.log`,
`/tmp/host29-exit-and-counts-final.log`, `/tmp/host29-binding-first.log` and
`/tmp/host29-vet-final.log`. Full repository tests are left to the fast gate.

Linux startup/process count rows are refreshed with scoped TestCountsAreRecorded,
which preserves unmeasured rows; full runs retain the canonical table check.
The getter row is 2/2/5/8/2/0, clocks 2/2/3/5/2/0, memory 1/1/0/2/1/0 and
identity 16/16/5/19/4/0. The directory-system row changes to
3876/3876/7131/5813/2430/0 because it enumerates the expanded fixture directory.
The full-table refresh failed on eleven graph-region fixtures with invalid
frees, including graph_regions_coverage_readonly.a. That integration failure
is recorded in `/tmp/host29-counts-current.log`; no graph primitive is changed
by this binding unit. The readonly graph failure also reproduces at merge
checkpoint f2173aef before binding edits, in `/tmp/host29-baseline-graph.log`.
macOS and Windows were not run.
