Built checked JavaScript witnesses, exact named host/runtime assertions and opt-in V8 coverage of all input fixtures.  
Engine implementation: `7d3f5ed459e114a82b2df5937a38f3a82cf69a9a`; runtime fix separately pushed as `90d3a822b537fcca370ab8c468447c1887baa7c4` on codex/wasi-empty-path.  
Green: Wasmtime 329 fixtures, 324 equivalent + 5 asserted limitations; V8 329 fixtures, 328 equivalent + 1 asserted runtime bug; zero failures/skips.  
Mutants: compiled stdout/stderr/exit changes caught by both runners; every named exception rejects independently changed output or status.  
Not covered: full repository gate, ARM, reactor, deployment or integration of the separate runtime fix.

The base of `codex/wasm-engines` remains unchanged from the previous pushed `5c040b0`. No runtime hooks or empty-path.a fixture are included here. The historical report probe was removed because the input fixture now belongs to the runtime branch. `oracle/wasi.mjs`, `internal/oracle/wasi_test.go`, directory.c and the engine pin/setup block were unchanged in this follow-up.

Checked fixtures use `onJavaScriptBackend`, matching wasi_test.go. The `writes_past_end.a` control observed raw source exit 0 and JavaScript backend exit 70; Wasmtime agrees byte for byte with the backend, including stderr. All ten checked fixtures now pass with this witness.

| Engine and scope | Fixtures | Equivalent | Asserted limitations | Fail | Skip |
|---|---:|---:|---:|---:|---:|
| Wasmtime, ordinary + input | 329 | 324 | 5 | 0 | 0 |
| V8, ordinary | 323 | 323 | 0 | 0 | 0 |
| V8, input | 6 | 5 | 1 | 0 | 0 |
| V8, total | 329 | 328 | 1 | 0 | 0 |

Limitations are reported as `LIMIT`, never equivalent passes or skips. For each named observation, the test requires exact stdout, stderr and exit status; an unknown difference fails, and a changed known result fails.

| Named fixture | Exact Wasmtime observation | Classification |
|---|---|---|
| write_stdout_order.a | Exit 0; stdout `first\nthird, after cannot write /dev/stdout: permission denied\n`; empty stderr | Host special-file access |
| write_stderr_order.a | Exit 0; stdout `first\nthird\n`; empty stderr, including absent source `second\n` | Host special-file access |
| prompt_then_read.a | Exit 0; stdout `ready\ncannot read /dev/stdin: permission denied\n`; empty stderr | Host special-file access |
| arguments.a | Exit 1; empty stdout; stderr `Error: failed to convert "a\xFFb" to utf-8\n` | Host CLI rejects invalid UTF-8 argv before execution |
| walk.a | The exact eight-line scratch listing block becomes `cannot read directory <dir>: failed\n`; all other output is compared | Host rejects non-UTF-8 filenames |

Walk also has a distinct **known Adamic runtime bug** on this branch, under both engines: its empty listing error becomes `listed <comma-joined cwd names>`, and its empty status error becomes `directory false`. The expected directory names are independently read and sorted from the fixture cwd. Only those two exact lines are transformed; the rest of the source witness is compared unchanged. The controls require the original Node errors exactly once. This is explicitly named as pending `codex/wasi-empty-path`, whose hooks fix the bug. Integrating that fix must change these assertions; the assertions deliberately fail when either old bug disappears.

No new engine semantic difference was found. Existing minimal C host witnesses are [stdio-path.c](stdio-path.c) and [directory-bytes.c](directory-bytes.c); their original Node/Wasmtime logs and the pinned 38.0.3 release hashes remain in this report directory. Node is still the semantic witness, except checked fixtures use the independent JavaScript backend's inserted checks.

The new `internal/oracle/wasi_input_test.go` preserves wasi.mjs's `/` preopen. It adds libc's `.` alias for the host cwd and explicit absolute scratch and `/dev` aliases, identically for V8 and Wasmtime. Without the dot alias, relative file fixtures see `/`; without the explicit absolute aliases, the dot mapping shadows scratch and device access. Each writing fixture uses the same printed argument path, reset between source and module runs; filesystem bytes and permissions are compared too. Permission probes run unprivileged, dropping to uid/gid 65534 when launched as root. Every module is built once per engine test with `adamic build --target wasm32-wasi`.

Commands and observed outputs, all logged and retained compressed:

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_ORACLE_WASI=1 ADAMIC_ORACLE_WASMTIME=1 go test ./internal/oracle -run 'Test(WASIInput|WasmtimeAgrees)' -count=1 -timeout 30m -v
# PASS, 193.915 s; complete Wasmtime and V8 input counts above.
ADAMIC_ORACLE_WASI=1 go test ./internal/oracle -run '^TestWASI(AgreesWithNode|OracleCatchesMutants|RunnerCatchesMutants)$' -count=1 -timeout 30m -v
# PASS, 150.452 s; all 323 ordinary V8 fixtures.
ADAMIC_ORACLE_WASI=1 ADAMIC_ORACLE_WASMTIME=1 go test ./internal/oracle -run '^Test(WasmtimeCheckedWitness|WasmtimeOracleCatchesMutants|WASINamedBehaviorCatchesMutants|WASIInputRunnerCatchesMutants|WASIWalkBehaviorCatchesMutants)$' -count=1 -v
# PASS, 1.243 s.
go test ./internal/oracle -run '^TestInputAgreesWithNode$' -count=1 -timeout 15m
# PASS, 3.114 s; all six existing native/JavaScript input fixtures.
go vet ./internal/oracle
# PASS, empty output. gofmt and git diff --check also clean.
```

The compiled-module mutants changed dedication's stdout bytes, main return to 23, and stderr independently. Both Wasmtime and the new V8 input runner caught each by its corresponding comparison. Four named host observations rejected stdout, stderr and exit mutations independently (12 checks). Walk's two runtime observations rejected changes under each engine, and its invalid-filename host observation rejected a changed result (5 checks). These latter probes mutate the captured observation, while the runner probes mutate compiled C; the report distinguishes them.

Tools were reused from W5: Node 24.19.0, Wasmtime 38.0.3, WASI SDK 27, Go 1.27.1, native clang 20.1.8, nproc 5. Original setup completed in 200 s, with the SDK retry recorded at 79 s. Every input run and Wasmtime source run is uncached. Complete names and classifications are in [follow-up-fixtures.md](follow-up-fixtures.md).
