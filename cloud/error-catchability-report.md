Two exact integration probes were fixed before this unit; both now have permanent Node oracle fixtures.
Implementation commit: 7d4faab; the following audit commit records the final classifications and checks.
The uncached affected gate, scoped Cohere check and final count regeneration passed; all 266 prior fixture rows are exact.
Four mutants were caught: two native catch-versus-panic comparisons and two precise refusal assertions.
Not covered: native String.fromCodePoint catchability, the complete repository gate and performance benchmarks.

Branch: `codex/error-classes-counts`, continuing from `379ea1f`. No checkout from main and no pull request.

| Integration probe | Classification | Commit and evidence |
| --- | --- | --- |
| `9984394_lib_dispatch.a` | **fixed-before** | `be5a509` introduced checked toFixed using an ordinary nominal RangeError throw; the existing class-method exception propagation takes it through the interface call. On this unit's unchanged compiler, Node, sanitized/release native and the JS backend all finish with `1.50\ncaught\n`, exit 0. |
| `9984394_lib_codepoint.a` | **refused-with-reason** | `be5a509` already added `ir.StringFromCodes` to libraryFailure. Nonconstant code-point arguments are NotYet inside try. The native `from_codes.c` primitive calls `adamic_uncaught_library_error`, not the pending-exception cleanup path, so claiming it can reach catch would miscompile. The exact probe is retained as an oracle refusal fixture; a separate permanent boundary test first verifies Node prints `made 1\nfromCodePoint caught\n`, then requires the specific Lower refusal. No claim of native catchability is made. |
| `9984394_defined_in_try.a` | **fixed-before** | `be5a509` changed the existing narrowed property-read check to `errorDefined`, whose failing branch throws a nominal TypeError. On this unit's unchanged compiler, all backends finish with `caught\nafter\n`, exit 0. `narrowed.go` is unchanged in this unit, leaving the other worker's number-field invalidation extension independent. |
| Additional `9984394_lib_dispatch_codepoint.a` | **fixed-now**, by refusal | The stated libraryFailure traversal omission still existed even though toFixed's new wrapper protected the first exact probe. A CodePoints class implementing a Maker interface could hide unguarded String.fromCodePoint. libraryFailure now includes class methods in the conservative function-value target set, follows their bodies once using visited, and excludes already guarded helpers. Function order makes the first diagnostic deterministic. The additional fixture verifies Node's catch and requires the same specific Lower refusal. |

The historical commit attributions are from inspecting `git show be5a509`: its new error-library guards, its StringFromCodes refusal case and the three-line `narrowed.go` call to errorDefined. Exact baseline execution was at `379ea1f`; this unit did not rebuild every historical commit. Monotone exception discovery and the previous precision work remain intact.

Before changing the compiler, the registered exact probes produced two PASS results and one Lower refusal, recorded in `/tmp/adamic-catchability-before.log`. The code-point probe was initially registered as executable to observe its true boundary, then registered as refused. The additional interface-code-point probe holds the broader traversal correction rather than attributing that correction to the already working toFixed case.

The executable probes are standard entries in TestNativeAgreesWithNode, so source Node, generated JavaScript on Node, native with ASan/UBSan and release native must agree byte for byte, and successful native runs must leak nothing. Registration and new tests live in `internal/oracle/catchability_test.go`; `internal/oracle/oracle_test.go` is unchanged. Refused probes are also standard refusal entries under `internal/oracle/testdata/catchability-refused/`. This keeps them out of flow's executable-only root glob while retaining the explicit oracle and boundary checks. TestCodePointCatchabilityBoundary independently executes their source on Node and checks the precise refusal reason, so a generic refusal elsewhere or a fixture that did not actually exercise Node's catch cannot pass.

| Mutant | Fixture and intended check | Observed result |
| --- | --- | --- |
| Interface toFixed guard removed | `9984394_lib_dispatch.a`, TestIntegrationCatchabilityMutants | Successfully built native; Node exit 0, mutant exit 70. Catch output/exit comparison catches the raw runtime panic. |
| Narrowed TypeError becomes panic | `9984394_defined_in_try.a`, same permanent mutant test | Successfully built native; Node exit 0, mutant exit 70. The same message becomes uncatchable, so the fixture distinguishes throw from panic. |
| Code-point failure-list case omitted | `9984394_lib_codepoint.a`, source overlay through TestCodePointCatchabilityBoundary | Successfully compiled; Lower accepts instead of refusing. The “want String.fromCodePoint catchability refusal” assertion fails. |
| Interface method targets omitted | `9984394_lib_dispatch_codepoint.a`, source overlay through the same boundary test | Successfully compiled; Lower accepts the method-hidden runtime failure. The specific refusal assertion fails. |

The permanent IR mutant tests run as part of the oracle gate. Reproduce both source mutants with `python3 cloud/error-catchability-mutants.py` after sourcing the toolchain environment. Overlays never edit the working tree; build failures are not accepted as mutant kills.

Counts were regenerated. All 266 previously recorded fixtures keep all six counts exactly; only the two new executable rows are added. Refused fixtures do not have executable counts. This preserves the prior 63-row audit and every restored non-throwing program:

| New executable fixture | Allocations | Frees | Retains | Releases | Peak live | In regions |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `9984394_lib_dispatch.a` | 4 | 4 | 10 | 12 | 2 | 0 |
| `9984394_defined_in_try.a` | 2 | 2 | 7 | 9 | 1 | 0 |

Setup: `bash cloud/setup.sh > /tmp/adamic-catchability-setup.log 2>&1`, then `source /workspace/adamic-tools/env.sh`. Go 1.27.1, clang 20.1.8, Node 24.19.0; `nproc` printed `5`:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (69s)
setup: done in 69s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Test commands use `TMPDIR=/tmp/adamic-gate` with mode 1777. Test output is written directly to logs.

Cohere was built from the pinned submodule with `go build -o /tmp/adamic-catchability-cohere ./command/cohere`. As in the previous error unit, the binary was given byte-identical `.ts` mirrors, the original compiler options and the original prelude. The integration probes intentionally retain their interface names and shorthand method signatures; documented eslint-disable comments limit the two relevant style/soundness rules to those probes. Formatting was applied as a patch. Final `--no-fix` exited 0: 276 rules, four checked, 50% Adamic-ready (two of four), only four of five files. The prelude was loaded but not selected for linting. The four checked mirrors match the final fixture bytes. This is a scoped lint/format pass, not a claim that the entire repository is Cohere-ready.

The first affected gate caught only a fixture-layout mistake: flow loads root `.a` files as executable programs and cannot accept intentionally refused probes there. Moving the two refusal fixtures into their dedicated subdirectory fixes that placement; the oracle still registers both and the Node boundary test still executes both. No flow expectation, oracle comparison or sanitizer check was weakened. The full affected gate was rerun afterward.

Completed checks and commands:

| Command | Observation | Log |
| --- | --- | --- |
| `gofmt -l cmd internal` | exit 0, no output | `/tmp/adamic-catchability-gofmt-final.log` |
| `go vet ./...` | exit 0, no output | `/tmp/adamic-catchability-vet-final.log` |
| `git diff --check` | exit 0, no output | terminal |
| Focused oracle, boundary and permanent-mutant command below | exit 0; both executable comparisons, both Node/refusal boundaries and both native mutants passed | `/tmp/adamic-catchability-unit-final.log` |
| `python3 cloud/error-catchability-mutants.py` | exit 0; both final source overlays compiled and failed their intended assertions | `/tmp/adamic-catchability-source-mutants-final.log`, individual logs under `/tmp/adamic-catchability-source-mutants/` |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts` | exit 0, oracle 52.434s; final table regenerated, all 266 prior fixture rows unchanged | `/tmp/adamic-catchability-counts-final.log` |
| Scoped Cohere command below | exit 0, four sources checked; byte-for-byte mirror assertion passed | `/tmp/adamic-catchability-cohere-final.log` |

```sh
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(9984394_|catchability-refused/9984394_)|TestCodePointCatchabilityBoundary|TestIntegrationCatchabilityMutants' -count=1 -v > /tmp/adamic-catchability-unit-final.log 2>&1
/tmp/adamic-catchability-cohere --directory /tmp/adamic-catchability-cohere-fixtures --no-fix 9984394_lib_dispatch.ts 9984394_lib_codepoint.ts 9984394_defined_in_try.ts 9984394_lib_dispatch_codepoint.ts > /tmp/adamic-catchability-cohere-final.log 2>&1
```

The focused command's executable comparisons select the two working probes. Both refused sources are separately exercised by TestCodePointCatchabilityBoundary; the full oracle gate includes their standard refusal registration too.

Implementation commit: `7d4faab` (`Follow interface methods when refusing uncatchable library failures`). The following report commit is on the same branch and its SHA is available in `git log -1`. No pull request.

Final affected gate command:

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql > /tmp/adamic-catchability-gate-final.log 2>&1
```

```text
? ir: no test files
ok lower 53.488s
? javascript: no test files
ok flow 185.436s
ok fresh 81.884s
ok native 316.037s
ok oracle 289.502s
ok graphql 121.480s
exit 0
```

The complete affected run includes the existing error-class/generated-error/reporting mutants and the new permanent mutants, all ordinary executable comparisons, both registered refusals and both explicit Node boundary checks. ASan, UBSan and leak checking remain enabled. No gate failure is waived. Only this report changed after the final passing gate.

Not covered: the complete repository `go test ./...`, repository-wide Cohere lint, performance measurements and a rebuilt execution at every historical commit. The old source-overlay mutant suites were not rerun in this unit; their ordinary tests and all existing compiled oracle mutants are included in the affected gate. Native String.fromCodePoint range failures remain uncatchable; nonconstant or spread code-point arguments reached inside try are explicitly NotYet. This unit strengthens that refusal through interface methods rather than adding a new variadic/spread validation and unwinding implementation. Existing invariant panics and other documented library-size/stack limitations remain.
