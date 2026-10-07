# Combined host proof

Proof branch: `codex/host-proof-combined`; base `a9083be`. This is not a library landing. No main/area push and no force push.

Ordered merge commits: `003a295b` (047e857), `6215746d` (e32a414), `9f58d766` (35bc60e), `83c34ab9` (2275ca9), `ac24283b` (616870d), then `ffc1aac6` (40bf2c98). The proof also reconciles the enum/export dispatch, runtime typeof ABI, null sentinel and leak-check callback conflicts. It keeps Darwin and WASI guards.

8/25 originals and 11/25 after adaptation 47 agree with Node on both backends. After adding host-blockers, 17-22 were rerun on both backends as requested; the other 19 rows retain the full five-input proof observations. Adaptation 47 changes five fixtures (01-04, 13); the other 20 are identical controls, reused from the original run. The audited stage3 status.json and NOTICE were not changed. `Agrees` means execution, exact stdout/stderr/exit and fresh Node comparison, not just successful compilation.

| Fixture | Original native | Original JS | Adapted native | Adapted JS | First effective blocker | Owner |
|---|---|---|---|---|---|---|
| 01_readFile_utf8.a | Checker | Checker | NotYet | NotYet | array of never; original checked byte reads fail TS2322 | Compiler after adaptation 47 |
| 02_readFile_utf16le.a | Checker | Checker | Agrees | Agrees | None | None |
| 03_readFile_utf16be.a | Checker | Checker | Agrees | Agrees | None | None |
| 04_readFile_missing.a | Checker | Checker | Agrees | Agrees | None | None |
| 05_writeFile.a | Refused | Refused | Refused | Refused | driver catch uses an unchecked ErrnoException cast | Adaptation |
| 06_fileExists.a | NotYet | NotYet | NotYet | NotYet | string + boolean | Compiler |
| 07_directoryExists.a | NotYet | NotYet | NotYet | NotYet | string + boolean | Compiler |
| 08_getDirectories.a | NotYet | NotYet | NotYet | NotYet | combinePaths rest parameter | Compiler |
| 09_realpath.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 10_getModifiedTime.a | NotYet | NotYet | NotYet | NotYet | optional Date.getTime call | Compiler |
| 11_setModifiedTime.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 12_deleteFile.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 13_createDirectory.a | Checker | Checker | Refused | Refused | remaining driver ErrnoException cast; original unknown catch fails TS18046 | Adaptation |
| 14_getCurrentDirectory.a | Refused | Refused | Refused | Refused | memoize callback capture cycle | Compiler; runtime region prerequisite absent from fetched area tip |
| 15_getExecutingFilePath.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 16_getEnvironmentVariable.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 17_write.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 18_exit_0.a | NotYet | NotYet | NotYet | NotYet | function value with an optional parameter | Compiler, census-small-families |
| 19_exit_1.a | NotYet | NotYet | NotYet | NotYet | function value with an optional parameter | Compiler, census-small-families |
| 20_exit_2.a | NotYet | NotYet | NotYet | NotYet | function value with an optional parameter | Compiler, census-small-families |
| 21_createHash.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 22_createHash_fallback.a | NotYet | NotYet | NotYet | NotYet | ambient createSHA256Hash without body | Library, buffer worker |
| 23_newLine.a | Agrees | Agrees | Agrees | Agrees | None | None |
| 24_useCaseSensitiveFileNames.a | NotYet | NotYet | NotYet | NotYet | RegExp.replace callback | Library, regex |
| 25_readDirectory.a | Checker | Checker | Checker | Checker | unchecked indexed reads and assertion annotation, TS2345/TS2775 | Adaptation |

Fixture 13 supplementary checks: the exact unknown helper returns ENOENT from a real fs.statSync error on Node, native and JS. Changing only the remaining driver cast to errorCode(e) keeps recorded Node output but reaches NotYet reading directoryExists (object shorthand function value). This supplementary adaptation is not counted as a pass for adaptation 47.

Fixture 14 scratch: merge commit `57e4a94cc242ee2d712e5b71e2e92a587c73f20f`, combined merge tip plus refreshed area/runtime `915b9e05`, at `/workspace/fs-combined-runtime-14`. Native and JS remain Refused at 14_getCurrentDirectory.a:10:28, callback cycle. `origin/codex/graph-regions` is `198b1271`; it is not an ancestor of the refreshed runtime area tip, and area/runtime has no internal/lower/graph_flow.go. This scratch merge is excluded from the pushed combined branch.

Verified one-line compiler reproducers (both backends unless otherwise stated):

01_array_never.a:
```typescript
import { Buffer } from 'node:buffer'; const cases = [[], [0x61]]; for (const bytes of cases) { console.log(Buffer.from(bytes).toString('utf8')); }
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/01_array_never.a:1:54: stage 0 can't lower an array of never yet

06_string_boolean.a:
```typescript
console.log('file=' + true);
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/06_string_boolean.a:1:13: stage 0 can't lower a BinaryExpression with a string and a boolean yet

08_rest_parameter.a:
```typescript
function combinePaths(path: string, ...paths: (string | undefined)[]): string { return path; } console.log(combinePaths('a', 'b'));
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/08_rest_parameter.a:1:37: stage 0 can't lower a parameter that isn't a plain name yet

10_optional_call.a:
```typescript
function getTime(): Date | undefined { return new Date(0); } console.log(String(getTime()?.getTime()));
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/10_optional_call.a:1:81: stage 0 can't lower a call through ?. (an optional call) yet

13_method_shorthand.a:
```typescript
function directoryExists(path: string): boolean { return true; } const system = { directoryExists, createDirectory(path: string): void { if (!system.directoryExists(path)) console.log(path); } }; system.createDirectory('x');
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/13_method_shorthand.a:1:83: stage 0 can't lower reading directoryExists yet

14_callback_cycle.a:
```typescript
function memoize<T>(callback: () => T): () => T { let value: T; return () => { if (callback) { value = callback(); callback = undefined!; } return value; }; } const get = memoize(() => 'cwd'); console.log(get());
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/14_callback_cycle.a:1:21: Adamic 0.1 refuses 'callback', a variable a function value captures and can be reached from what it holds, so the function holds the variable and the variable holds the function: a cycle reference counting can't free; write the function as a function declaration (function callback() {}), which captures nothing, or declare the variable Weak<...> and keep the function somewhere strong (adamic/cycle-capable)

18_optional_method_parameter.a:
```typescript
const system = { exit(code?: number): void { console.log(String(code)); } }; system.exit(0);
```
adamic: /workspace/adamic/cloud/reports/host-proof-combined/probes/18_optional_method_parameter.a:1:23: stage 0 can't lower a function value with an optional parameter yet
Resolved compiler probes: standalone undefined bindings and exact-undefined conditions now compile on both backends. Fixtures 18-20 next stop at optional object-literal method parameters (line 21); 22 next stops at the ambient hash declaration (line 20).

Additional compiler gate blocker: narrowed scalar non-null assertion. Native emits double != NULL and clang rejects it; JS compiles.
```typescript
function narrowed(value: number | string | undefined): number { if (typeof value === 'number') return value!; return 1; } console.log(String(narrowed(0)));
```
adamic: native: clang failed: exit status 1
/tmp/adamic-gate/adamic-build-3014596281/main.c:19:28: error: invalid operands to binary expression ('double' and 'void *')
   19 |                 if (!(adamic_temporary_1 != NULL)) {
      |                       ~~~~~~~~~~~~~~~~~~ ^  ~~~~
1 error generated.

Validation commands and results:

- `GOPROXY="https://proxy.golang.org|direct" bash cloud/setup.sh --wasi-sdk`: exit 0; nproc 5, quota 4. Go 0.128s; Node 0.127s; markdown 0.295s; submodules 0.303s; clang 0.995s; WASI SDK 1.138s; go build 132.003s; cache warm 132.601s; total 132.842s. Earlier setup failed on disk full (28 GB Go cache); cache was cleared and final setup succeeded.
- `python3 internal/oracle/node_fs_file_host_check.py --all --compiler /tmp/fs-combined-adamic --logs /tmp/fs-combined-original-final --report internal/oracle/node_fs_file_combined_original_status.json`: exit 1 because named blockers remain; all 25 observed, eight agree on both.
- `NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node internal/oracle/node_fs_file_host_47_check.cjs /tmp/fs-adaptation47-tsc /tmp/fs-combined-adamic /tmp/fs-combined-47-final internal/oracle/node_fs_file_combined_47_status.json`: exit 1, five adapted fixtures observed, three agree on both, all five mutants caught.
- `go test ./internal/lower ./internal/ir ./internal/fresh ./internal/flow ./internal/load ./internal/native ./internal/javascript -count=1 -timeout 30m`: lower FAIL (old enum/refusal assertions, TestRegExpGroupCompoundRefusals and TestTasteRepresentationLimitsStayExplicit); flow FAIL (process_bad_code trace ends before basic block finishes); ir/fresh/load/native PASS, JS has no tests. Full go test ./... was not run.
- `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run "^(TestNodeFSFile.*|TestNodeFSDirectory.*|TestNodePath.*|TestNodeProcess.*|TestWASIInputAgreesWithNode|TestWASIHostRuntimeRefusals|TestWASIFileAgreesWithNode)$" -count=1 -timeout 20m -v`: exit 1, 52.476s. Native fs file/directory/path/process oracle tests pass except TestNodeProcessErrorNarrowingBlocker (expected old in refusal; now unbound hasOwnProperty read at node_process_errors.a:21:32). WASI runtime archive compiles. Input leg: 19 pass, two honest target skips, one failure (empty symlink target: Node ENOENT, WASI EINVAL, fs-directory owner). Fs file leg: nine pass, eight honest target skips. All nine runtime-refusal probes pass. An earlier overbroad TestWASI selection included the full language oracle and was stopped; no success is claimed for it.
- `go test ./internal/oracle -run "TestNodeFSFileCombinedNullClassifier|TestUnknown" -count=1 -timeout 15m -v`: PASS, 3.508s. Null-to-undefined native mutant runs cleanly under sanitizers/leak checks and is caught only by Node stdout. Unknown in_always_true and skip_inner_typeof mutants are also caught.
- `go test ./internal/oracle -run "^TestCountsAreRecorded$" -count=1 -timeout 30m -args -update-counts`: failed before rewriting (pre-null-reconciliation unknown fixture crashes and narrowed-scalar non_null.a clang error). Counts retains incoming rows once in registration order; full Linux regeneration is not claimed.
- Final incremental command: `python3 internal/oracle/node_fs_file_host_check.py --all --only 17 18 19 20 21 22 --compiler /tmp/fs-combined-adamic --logs /tmp/fs-combined-blockers --report internal/oracle/node_fs_file_combined_blockers_status.json`: exit 1 for the four named next gaps. 17/21 run and agree; 18-20 optional method parameter; 22 function without body.
- `go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(host_undefined_value|host_void_method).a$' -count=1 -timeout 10m -v`: PASS, 0.595s, both merged compiler fixtures agree with Node on native and JS and pass sanitizer/leak checks.
- Final WASI command: `ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^(TestWASIInputAgreesWithNode|TestWASIHostRuntimeRefusals|TestWASIFileAgreesWithNode|TestNodeFSFileCombinedNullClassifier)$' -count=1 -timeout 15m -v`: exit 1, 18.359s. The sole failure remains symlink empty-target Node ENOENT versus WASI EINVAL. Classifier mutant and nine runtime refusals pass; the same ten honest target skips remain.
- All 25 original semantic mutants from the audited check.py were run independently on Node and caught. Adaptation 47 adds five Node mutants; all caught. All 34 native fs file mutants passed their Node-only comparisons under sanitizers.

Not covered: no compiler fixes for surviving gaps, no fix for ambient createSHA256Hash (buffer worker), no extra catch/checked-read adaptations beyond the explicitly labelled supplementary driver probe, no graph-region branch merged into combined proof, no area/main landing, no macOS execution, no full language WASI gate. The only WASI behavioral mismatch observed is assigned to the fs-directory worker.

Logs are kept in logs/ beside this report, with full diagnostics and test output.
