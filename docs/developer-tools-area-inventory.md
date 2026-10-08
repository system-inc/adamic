# Developer-tools area inventory and fix-forward plan

Roadmap step 02, #2s8kq6y. Inventory only, no code landing. Snapshot fetched on 2026-10-08.

| Reference | Full SHA |
| --- | --- |
| origin/main | `45487a809f89885a3fc651cd590e7dabf31362dc` |
| origin/area/developer-tools | `5fb868e3f08583b9a0d5cd3fc0fcac474eb12f68` |
| Merge base | `855d114e9b37776ec3739f25d63dbf4da968d02e` |

The area has 203 commits absent from main; current main has 319 commits absent from the area, rather than the reported 233. `git diff origin/main...origin/area/developer-tools` has 837 paths, 190,227 added and 3,661 removed text lines. These are delta lines, not file sizes, and binary content is excluded from line totals. A three-dot diff measures what the area added since the merge base; it does not describe all changes needed to replace current main with the area.

Do not land the area merge. There are useful independent checks, but there are stale global inventories, shared-file conflicts, and independently owned extractions. Keep current main as the source of truth and port selected hunks. In particular, main's CPU-time fuzzer limits must survive.

## Classification and counting

A means the whole carried delta has a demonstrated replacement on main. B means live checking code, test inputs or support worth extracting, subject to the listed conflicts and validation. C means historical evidence or inert documentation, including the scripts kept with a particular measurement; these scripts can be executed manually but do not install a gate. D means demonstrated dead experiments. U means classification is unresolved or mixed and explains why. A file is counted once in the ledger below; shared-file hunks are discussed separately. B is a recommendation to port a check, not a claim that the area blob can safely replace main's blob.

| Class | Files | Added lines | Removed lines | Binary files |
| --- | ---: | ---: | ---: | ---: |
| A | 0 | 0 | 0 | 0 |
| B | 214 | 28,672 | 3,606 | 0 |
| C | 619 | 161,286 | 7 | 23 |
| D | 0 | 0 | 0 | 0 |
| U | 4 | 269 | 48 | 0 |

A and D are zero at whole-file granularity. No carried file equals its current-main counterpart (`git diff --quiet main area -- path` for every ledger path), and this audit did not establish a complete alternative implementation superseding a whole carried delta. Absence of a static reader alone does not prove a manually invoked tool dead. Already-landed foundations are outside this three-dot inventory: main has the original sharding/resume code (`b790232b`, `64c91449`), generator judging/reduction (`50e99c57`, `c81ece50`), shared runtime/split compilation, and later runtime work through `8677d41b` and `4e0bfda5`. Do not re-land those foundations merely because the area has more files in the same package.

The four unresolved paths are not quietly counted as dead: `CLAUDE.md` combines protected-owner documentation, pinning instructions and policy; `CohereSettings.json` broadens ignores and changes gate coverage, requiring the cohere extraction owner's decision; `docs/lint-registration.md` mixes an operational command contract with a historical report; `internal/fuzz/run.go` mixes useful API/judging/instrumentation changes with replacement of main's CPU deadlines. Their exact counts are included in U. Extract specific hunks with their owners rather than landing those blobs.

## Live families and main dependencies

The path ledger is the exact set of surviving paths. For a family below, its B rows are the recommended candidate paths. Scratch attempts used raw family sets defined before disputed run.go was moved to U; that raw overlay is explicitly discussed. Protected families were not independently tested. This is a dependency inventory, not a proposal to land each family as one large commit.

| Family | Package or script and purpose | Main dependency or conflict | Smallest independent extraction |
| --- | --- | --- | --- |
| 01-boundedrun | `internal/boundedrun`: deadline-bound Go, Python and shell children, killing their process groups | New leaf package; consumers must preserve main's CPU-time semantics where those already exist | Go: `identity.go`, `run.go`, `run_test.go`, `testfixture/hang.go`; Python/shell adapters and their tests may follow separately |
| 02-nodepin | `internal/nodepin`, `cloud/setup-node.py`: require v24.19.0 and validate downloaded platform artifacts | Needs boundedrun; enabling pin in callers must come after installer support | `internal/nodepin/{nodepin.go,nodepin_test.go}`; installer separately: `cloud/{node-pin.json,setup-node.py,test_node_setup.py,node-setup-mutants.py}` |
| 03-refusalprobe | `internal/refusalprobe`, `cmd/adamic-refusals`: generate forbidden constructs and accepted neighbors and audit catalog completeness | Uses current load/lower/native APIs; enum and optional-widening helper owners are stale | `internal/refusalprobe/{catalog.go,probe.go,probe_test.go}`, `cmd/adamic-refusals/main.go`; historical lies/results are unnecessary |
| 04-metamorphic | `internal/metamorphic`, `cmd/adamic-metamorphic`: compare semantics-preserving transforms across source Node, JS, sanitized and release native | Depends on area-only exported fuzz verdict/execution/comparison APIs; parser shim adjustment `c5e950e6` already adapts this piece to main's shim | Ten package Go files and CLI, plus narrowly ported fuzz API prerequisite; no REPORT |
| 05-skipcensus | `internal/skipcensus`: compare AST skip declarations and raw JSON test events, refusing unknown/required-input skips | Static `testdata/skips.json` describes area, not main; must scan and classify main afresh. Parser benchmark guard can land alone | Seven Go/data files plus `testdata/proofs.py`, with regenerated `skips.json`; separate three-line change in `stage1/typescript/parser/parser_test.go` |
| 06-fuzz | `internal/fuzz`, fuzz/reduce CLIs: new liveness/field/operator scenes, stronger severity-preserving reduction and execution APIs | Shared generate/reduce/run files; main now has CPU-time fixes `bc0100a1`, `e266de6b`, `dc30d0fe`. Raw run.go reverses them. Coverage calls need native instrumentation | Split scenes, judge/reducer and exported API hunks into distinct commits as listed in the plan |
| 07-test262 | `cmd/adamic-test262`: bounded compilation/execution and stronger run/cache deadline checks | boundedrun and nodepin; reconcile cache identity and preserve main changes | Deadline/runtime adapter files and bounded call-site/test hunks first; pinning hunks only after installer |
| 08-gate | `cmd/adamic-gate`: selectable units, complementary coverage, frozen provenance, required inputs, affinity, archive policy, WASI and timing calibration | Starts from main's existing shard/resume harness; new main tests invalidate old timing and discovery assumptions. boundedrun and skipcensus prerequisites; live timings.json is scheduling input | Extract integrated call-site hunks with each helper/test pair, not the entire 72-path package; regenerate timing data last |
| 09-cohere-owned | `cloud/cohere_gate.py`, baseline and Go test: repository cohere baseline gate | Already being carried on `devtools/cohere-gate-main`; broad ignores remain an owner decision | Reserved three paths; no validation or landing by this unit |
| 10-setup | cloud setup/input scripts and testdata: checksum/pin/content validation, module warming, cross-platform setup and optional corpus/archive preparation | boundedrun; setup.sh is shared with current main's bootstrap and markdown work. Node installer is family 02 | Split module, stage3 API, corpus/npm and archive-mode helpers with only their setup.sh wiring; retain main bootstrap |
| 11-lint | selected-rule CLI, lint cache/registry and tests: certify one rule against all runtimes, prove cache invalidation and mutant rejection | Main's lint registry/rule corpus are authoritative; shared lint_test.go/registry.go need hunk port | CLI + selected-rule helper/tests + optional-selection registry hunks; cache helpers/tests separately with their call sites |
| 12-leaks-owned | shared leakcheck and stage1 port tests: enforce native leak checks on Linux/macOS | internal/leakcheck and oracle leak tests reserved to `devtools/mac-base-green`. Stage1 callers should wait for its API | Reserved package/oracle extraction owned elsewhere; then one port's import/call-site change per commit |
| 13-native-oracle | native flags and oracle opt-in lanes: GCC, release flag, slab/malloc and JSON helper checks | Multiple features share native.go/library.go/cache_test.go; oracle files overlap leak owner's work. Node pin + boundedrun required by copied cache harness. Main compiler/helper and runtime behavior are authoritative | GCC selection+tests; release_flags+tests; Slabs option+tests; coverage flags; JSON helper preparation/cache hunks, each separately |
| 14-catalog | `verify/catalog/check.sh` and refreshed narrowed-field mutant: bound catalog runner and preserve a reproducible regression mutation | boundedrun Python/shell; patch targets old lower/object.go context and must be checked/refreshed on current main | check.sh's wrapper separately; `08-narrowed-number-field-39638d9e.patch` only after `git apply --check` and current mutant proof |
| 15-coverage | `verify/coverage/{measure.sh,analyze.py,runtimelibrary/main.go}`: measure compiler Go coverage and native C coverage under generators/oracle | Needs native coverage option/env hook and fuzz per-run profile naming; LLVM tools and runtime library build | Those three live scripts plus native/native.go coverage and fuzz/run.go profile hunks; recorded REPORTs unnecessary |

## Evidence readers

C is inert with respect to ordinary compiler/gate execution. No current-main gate reader of these added records was found by source-reference inspection. The main `stage3/census/latent/*` REPORT readers refer to their own reports, not the inventory's added coverage reports.

* `cloud/reports/` holds 446 historical records/drivers. `cloud/run-css-input-proof.py` writes `cloud/reports/css-gate-inputs`; colocated measurement/proof scripts manually consume their own records. These are investigation tools, not inputs to an ordinary gate verdict.
* `cmd/adamic-gate/evidence/` archives, `timings.json.audit.json`, and child-deadline audit records are stored proofs. The gate **does** read `cmd/adamic-gate/timings.json`, which is B. Its timing command produces an audit sidecar; the committed sidecar is not authority for a pass. Tests synthesize reference/evidence fixtures in temporary directories rather than loading the historical archives.
* `internal/oracle/json_types_results/`, `internal/refusalprobe/results/`, slab evidence and reports are historical outputs. `internal/refusalprobe/lies/` has no live Go/script reader found; retain it as historical witness evidence, not required probe package input.
* `verify/catalog/detection.json` and DETECTION.md describe old runs; current check.py does not read them. The `.patch` mutant is live B input, not a report.
* `verify/coverage/REPORT.{json,md}` are measurement outputs. B analyze.py writes them (and the document points humans to JSON); it reads fresh profiles from the measurement directory, not these committed reports.
* docs gate instructions and reports are inert as executable inputs. fleet.sh refers operators to `docs/gate-shards.md`; archive.go names it in a diagnostic. The operational part of lint-registration.md is U because it belongs with a live command contract, not solely a recorded run.
* In contrast, npm/bootstrap/lock manifests, boundedrun's hang fixture, and skipcensus `testdata/{skips.json,plain-skips.jsonl}` are read by B code/tests. They remain B even though they are data.

Static references cannot establish that no external automation or person consumes a report. This inventory makes no such claim.

## Scratch checks against main

Main's `go build -o <scratch>/adamic ./cmd/adamic` and `go build ./...` passed. Dependencies were initialized at main's cohere pin `7945d102a6c18dd36adf9114a758ce646e8b2359`, TypeScript `d92d9bfee114c80be2c375d72edae966176e3a4f`. Toolchain: Go 1.27.1, Node v24.19.0; GOMAXPROCS=4, TMPDIR=/tmp/adamic-gate, uncached tests.

Scratch branch `devtools/inventory-scratch` starts at the recorded main, overlays the family's candidate paths (including the disputed fuzz/run.go in the raw fuzz attempt) and the explicitly listed prerequisites, builds the touched packages and runs their tests. Each overlay is restored before the next. Tests have Go timeout 90s and an external 110s process-group ceiling; build ceiling 120s. These bounded attempts establish failures or passes, not full integration readiness. Larger suites need their normal 30m/90m ceilings before landing. Tests in a package with no production Go files are built by `go test`; `go build` on such a package reports “no non-test Go files” and is not a product compile failure. Python/shell tools require their own syntax/unit checks below.

The scratch worktree shares initialized cohere through a symlink. Initial attempts hit Git's refusal to stamp a symlinked submodule and cold-build limits; reruns use GOFLAGS=-buildvcs=false. This is a scratch accommodation, not a proposed product setting. Prior attempt failures are retained below. Prerequisite overlays are disclosed rather than labeled independent success. Protected workers' branches and main were not changed; the cohere-gate family and stage1 leak ports were not overlaid/tested here. The combined native/oracle attempt used the area's oracle leak wiring plus leakcheck only inside scratch; it timed out and is not an independent check or proposed landing of the protected extraction.

### Current outcome by candidate

| Candidate | Builds against main | Touched-package/check outcome | Independent green? |
| --- | --- | --- | --- |
| Boundedrun | Yes | Go tests pass; six Python tests pass | Yes for tested Go/Python package; shell adapter needs its own full proof |
| Node pin helper/installer | Yes with boundedrun | Go tests and installer Python tests pass | Yes with disclosed boundedrun prerequisite |
| Refusalprobe | Yes | Raw catalog fails; two-helper-map repair passes existing package tests | No, semantic enum catalog repair still needed |
| Metamorphic | Raw fails missing fuzz API; builds with fuzz/native/pin/bounded prerequisites | Package tests pass with prerequisites | No on main alone; prerequisite full fuzz suite is red |
| Skip census and parser guard | Census builds; parser is test-only | Census stale/missing rows; full parser test attempt times out | No; regenerate census; guard-specific proof still needed |
| Fuzz/reduce | Yes with instrumentation/pin/bounded prerequisites | Full fuzz suite fails main CPU-limit tests | No; port preserving main CPU limits |
| Test262 | Production CLI builds | Test compilation fails: main tests still call runProgram | No; reconcile API with current main |
| Gate | Builds with census+bounded prerequisites | Current WASI/cohere source audits fail on stale area census/absent owner extraction | No; adapt contracts/table to each actual main landing |
| Cloud setup | Script syntax and existing cloud Go tests pass | With Node prerequisites, 49 Python tests pass, five skip | Unit checks green; real cold/warm/platform/optional input integration not established |
| Selected lint/cache | CLI/packages build | Full touched lint/... tests exceed bounded attempt | Unknown, requires normal 90m suite |
| Native/oracle lanes | Native builds; oracle is test-only | Combined package test attempt exceeds external ceiling before a complete result | Unknown; each new lane needs supplied opt-in inputs and normal suite |
| Catalog | Shell syntax passes | Refreshed field patch applies to recorded main (`git apply --check` exit 0) | Runner/mutant verdict not established by syntax/applicability |
| Coverage | Runtime-library helper builds | Shell/Python syntax pass, Go helper has no tests | Coverage measurement/cache separation not established |
| Reserved cohere/leak extractions | Not independently checked | Owned by other workers | Defer to owners |

No fresh full repository test gate, external-input integration, macOS run, selected-rule parity/mutant campaign, complete coverage measurement or fleet execution is claimed. The following detailed attempts supply actual commands and exits; a timeout means unverified, not dead.

### 01-boundedrun attempt

Exact scratch overlay paths: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`. Additional scratch prerequisites: none.

* build: `go build ./internal/boundedrun`, exit **0**, 0.17s.

* test: `go test -count=1 -timeout=90s ./internal/boundedrun`, exit **0**, 5.04s.

```text
ok  	github.com/system-inc/adamic/internal/boundedrun	4.555s
```

* Additional: `python3 -m unittest discover -s internal/boundedrun -p python_test.py`, exit **0**, 3.99s.

```text
......
----------------------------------------------------------------------
Ran 6 tests in 3.618s

OK

```

### 02-nodepin attempt

Exact scratch overlay paths: `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`. Additional scratch prerequisites: none.

* build: `go build ./internal/nodepin`, exit **1**, 0.17s.

```text
internal/nodepin/nodepin.go:9:2: no required module provides package github.com/system-inc/adamic/internal/boundedrun; to add it:
	go get github.com/system-inc/adamic/internal/boundedrun
```

* test: `go test -count=1 -timeout=90s ./internal/nodepin`, exit **1**, 0.15s.

```text
# github.com/system-inc/adamic/internal/nodepin
internal/nodepin/nodepin.go:9:2: no required module provides package github.com/system-inc/adamic/internal/boundedrun; to add it:
	go get github.com/system-inc/adamic/internal/boundedrun
FAIL	github.com/system-inc/adamic/internal/nodepin [setup failed]
FAIL
```

### 03-refusalprobe attempt

Exact scratch overlay paths: `cmd/adamic-refusals/main.go`, `internal/refusalprobe/README.md`, `internal/refusalprobe/catalog.go`, `internal/refusalprobe/lies/function_type_length.a`, `internal/refusalprobe/lies/function_type_length.expected`, `internal/refusalprobe/lies/widening_boolean_as_number.a`, `internal/refusalprobe/lies/widening_boolean_as_number.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_array.a`, `internal/refusalprobe/lies/widening_boolean_as_number_array.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_ops.a`, `internal/refusalprobe/lies/widening_boolean_as_number_ops.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_parameter.a`, `internal/refusalprobe/lies/widening_boolean_as_number_parameter.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_return.a`, `internal/refusalprobe/lies/widening_boolean_as_number_return.expected`, `internal/refusalprobe/lies/widening_false_as_number.a`, `internal/refusalprobe/lies/widening_false_as_number.expected`, `internal/refusalprobe/probe.go`, `internal/refusalprobe/probe_test.go`. Additional scratch prerequisites: none.

* build: `go build ./internal/refusalprobe ./cmd/adamic-refusals`, exit **1**, 2.99s.

```text
error obtaining VCS status: exit status 128
	Use -buildvcs=false to disable VCS stamping.
```

* test: `go test -count=1 -timeout=90s ./internal/refusalprobe ./cmd/adamic-refusals`, exit **TIMEOUT**, 110.01s.

### 04-metamorphic attempt

Exact scratch overlay paths: `cmd/adamic-metamorphic/main.go`, `internal/metamorphic/alias.go`, `internal/metamorphic/fixtures.go`, `internal/metamorphic/identity.go`, `internal/metamorphic/metamorphic_test.go`, `internal/metamorphic/method.go`, `internal/metamorphic/run.go`, `internal/metamorphic/source.go`, `internal/metamorphic/symbols.go`, `internal/metamorphic/temporaries.go`, `internal/metamorphic/wrap.go`. Additional scratch prerequisites: none.

* build: `go build ./internal/metamorphic ./cmd/adamic-metamorphic`, exit **1**, 1.4s.

```text
error obtaining VCS status: exit status 128
	Use -buildvcs=false to disable VCS stamping.
```

* test: `go test -count=1 -timeout=90s ./internal/metamorphic ./cmd/adamic-metamorphic`, exit **1**, 81.38s.

```text
# github.com/system-inc/adamic/internal/metamorphic
internal/metamorphic/run.go:49:14: undefined: fuzz.Execute
internal/metamorphic/run.go:58:18: undefined: fuzz.Execute
internal/metamorphic/run.go:60:19: undefined: fuzz.Refusal
internal/metamorphic/run.go:76:17: undefined: fuzz.Execute
internal/metamorphic/run.go:91:17: undefined: fuzz.Execute
internal/metamorphic/run.go:92:17: undefined: fuzz.Execute
internal/metamorphic/run.go:102:18: undefined: fuzz.Compare
internal/metamorphic/run.go:115:18: undefined: fuzz.Execute
internal/metamorphic/run.go:142:18: undefined: fuzz.Execute
internal/metamorphic/run.go:148:14: undefined: fuzz.Execute
internal/metamorphic/run.go:148:14: too many errors
FAIL	github.com/system-inc/adamic/internal/metamorphic [build failed]
FAIL	github.com/system-inc/adamic/cmd/adamic-metamorphic [build failed]
FAIL
```

### 02-nodepin attempt

Exact scratch overlay paths: `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`.

* build: `go build ./internal/nodepin`, exit **0**, 0.16s.

* test: `go test -count=1 -timeout=90s ./internal/nodepin`, exit **0**, 0.22s.

```text
ok  	github.com/system-inc/adamic/internal/nodepin	0.004s
```

### 03-refusalprobe attempt

Exact scratch overlay paths: `cmd/adamic-refusals/main.go`, `internal/refusalprobe/README.md`, `internal/refusalprobe/catalog.go`, `internal/refusalprobe/lies/function_type_length.a`, `internal/refusalprobe/lies/function_type_length.expected`, `internal/refusalprobe/lies/widening_boolean_as_number.a`, `internal/refusalprobe/lies/widening_boolean_as_number.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_array.a`, `internal/refusalprobe/lies/widening_boolean_as_number_array.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_ops.a`, `internal/refusalprobe/lies/widening_boolean_as_number_ops.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_parameter.a`, `internal/refusalprobe/lies/widening_boolean_as_number_parameter.expected`, `internal/refusalprobe/lies/widening_boolean_as_number_return.a`, `internal/refusalprobe/lies/widening_boolean_as_number_return.expected`, `internal/refusalprobe/lies/widening_false_as_number.a`, `internal/refusalprobe/lies/widening_false_as_number.expected`, `internal/refusalprobe/probe.go`, `internal/refusalprobe/probe_test.go`. Additional scratch prerequisites: none.

* build: `go build ./internal/refusalprobe ./cmd/adamic-refusals`, exit **0**, 5.49s.

* test: `go test -count=1 -timeout=90s ./internal/refusalprobe ./cmd/adamic-refusals`, exit **1**, 11.48s.

```text
--- FAIL: TestCatalogCoverage (0.00s)
    probe_test.go:15: refusal catalog is incomplete: helper enumRefusal, helper refuseOptionalWidening
FAIL
FAIL	github.com/system-inc/adamic/internal/refusalprobe	6.229s
?   	github.com/system-inc/adamic/cmd/adamic-refusals	[no test files]
FAIL
```

### 04-metamorphic attempt

Exact scratch overlay paths: `cmd/adamic-metamorphic/main.go`, `internal/metamorphic/alias.go`, `internal/metamorphic/fixtures.go`, `internal/metamorphic/identity.go`, `internal/metamorphic/metamorphic_test.go`, `internal/metamorphic/method.go`, `internal/metamorphic/run.go`, `internal/metamorphic/source.go`, `internal/metamorphic/symbols.go`, `internal/metamorphic/temporaries.go`, `internal/metamorphic/wrap.go`. Additional scratch prerequisites: none.

* build: `go build ./internal/metamorphic ./cmd/adamic-metamorphic`, exit **1**, 0.11s.

```text
# github.com/system-inc/adamic/internal/metamorphic
internal/metamorphic/run.go:49:14: undefined: fuzz.Execute
internal/metamorphic/run.go:58:18: undefined: fuzz.Execute
internal/metamorphic/run.go:60:19: undefined: fuzz.Refusal
internal/metamorphic/run.go:76:17: undefined: fuzz.Execute
internal/metamorphic/run.go:91:17: undefined: fuzz.Execute
internal/metamorphic/run.go:92:17: undefined: fuzz.Execute
internal/metamorphic/run.go:102:18: undefined: fuzz.Compare
internal/metamorphic/run.go:115:18: undefined: fuzz.Execute
internal/metamorphic/run.go:142:18: undefined: fuzz.Execute
internal/metamorphic/run.go:148:14: undefined: fuzz.Execute
internal/metamorphic/run.go:148:14: too many errors
```

* test: `go test -count=1 -timeout=90s ./internal/metamorphic ./cmd/adamic-metamorphic`, exit **1**, 0.17s.

```text
# github.com/system-inc/adamic/internal/metamorphic
internal/metamorphic/run.go:49:14: undefined: fuzz.Execute
internal/metamorphic/run.go:58:18: undefined: fuzz.Execute
internal/metamorphic/run.go:60:19: undefined: fuzz.Refusal
internal/metamorphic/run.go:76:17: undefined: fuzz.Execute
internal/metamorphic/run.go:91:17: undefined: fuzz.Execute
internal/metamorphic/run.go:92:17: undefined: fuzz.Execute
internal/metamorphic/run.go:102:18: undefined: fuzz.Compare
internal/metamorphic/run.go:115:18: undefined: fuzz.Execute
internal/metamorphic/run.go:142:18: undefined: fuzz.Execute
internal/metamorphic/run.go:148:14: undefined: fuzz.Execute
internal/metamorphic/run.go:148:14: too many errors
FAIL	github.com/system-inc/adamic/internal/metamorphic [build failed]
FAIL	github.com/system-inc/adamic/cmd/adamic-metamorphic [build failed]
FAIL
```

### 05-skipcensus attempt

Exact scratch overlay paths: `internal/skipcensus/census.go`, `internal/skipcensus/census_test.go`, `internal/skipcensus/cmd/main.go`, `internal/skipcensus/log.go`, `internal/skipcensus/optin.go`, `internal/skipcensus/testdata/plain-skips.jsonl`, `internal/skipcensus/testdata/proofs.py`, `internal/skipcensus/testdata/skips.json`, `stage1/typescript/parser/parser_test.go`. Additional scratch prerequisites: none.

* build: `go build ./internal/skipcensus/... ./stage1/typescript/parser`, exit **1**, 0.22s.

```text
github.com/system-inc/adamic/stage1/typescript/parser: no non-test Go files in /tmp/devtools-inventory-scratch/stage1/typescript/parser
```

* test: `go test -count=1 -timeout=90s ./internal/skipcensus/... ./stage1/typescript/parser`, exit **1**, 95.64s.

```text
, 0x35}, 0x16e4054d1728, {0x1c62060, 0x19, 0x19}, {0xc2aa101788a85e50, 0x14f4f6d93e, ...})
	/workspace/adamic-tools/go/src/testing/testing.go:2740 +0x510
testing.(*M).Run(0x16e4058cefa0)
	/workspace/adamic-tools/go/src/testing/testing.go:2600 +0x6af
main.main()
	_testmain.go:94 +0x9b

goroutine 178 [syscall]:
syscall.Syscall6(0xf7, 0x3, 0xa, 0x16e405a13618, 0x4, 0x16e4056fe090, 0x0)
	/workspace/adamic-tools/go/src/syscall/syscall_linux.go:96 +0x39
internal/syscall/unix.Waitid(0x16e405a13646?, 0x16e405a13770?, 0x53614b?, 0x18?, 0x1b6e990?)
	/workspace/adamic-tools/go/src/internal/syscall/unix/waitid_linux.go:18 +0x39
os.(*Process).pidfdWait.func1(...)
	/workspace/adamic-tools/go/src/os/pidfd_linux.go:108
os.ignoringEINTR(...)
	/workspace/adamic-tools/go/src/os/file_posix.go:263
os.(*Process).pidfdWait(0x16e405a7ed00)
	/workspace/adamic-tools/go/src/os/pidfd_linux.go:107 +0x1cf
os.(*Process).wait(0x1b?)
	/workspace/adamic-tools/go/src/os/exec_unix.go:25 +0x1a
os.(*Process).Wait(...)
	/workspace/adamic-tools/go/src/os/exec.go:347
os/exec.(*Cmd).Wait(0x16e4058ec4e0)
	/workspace/adamic-to
...
 {0x1bc5e00, 0x16e406e9e058})
	/workspace/adamic-tools/go/src/bytes/buffer.go:229 +0x98
io.copyBuffer({0x1bc5f20, 0x16e4056ec420}, {0x1bc5e00, 0x16e406e9e058}, {0x0, 0x0, 0x0})
	/workspace/adamic-tools/go/src/io/io.go:415 +0x151
io.Copy(...)
	/workspace/adamic-tools/go/src/io/io.go:388
os.genericWriteTo(0x16e406e9e058?, {0x1bc5f20?, 0x16e4056ec420?})
	/workspace/adamic-tools/go/src/os/file.go:295 +0x35
os.(*File).WriteTo(0x16e406e9e058, {0x1bc5f20, 0x16e4056ec420})
	/workspace/adamic-tools/go/src/os/file.go:273 +0x9c
io.copyBuffer({0x1bc5f20, 0x16e4056ec420}, {0x1bc5e80, 0x16e406e9e058}, {0x0, 0x0, 0x0})
	/workspace/adamic-tools/go/src/io/io.go:411 +0x9d
io.Copy(...)
	/workspace/adamic-tools/go/src/io/io.go:388
os/exec.(*Cmd).writerDescriptor.func1()
	/workspace/adamic-tools/go/src/os/exec/exec.go:613 +0x34
os/exec.(*Cmd).Start.func2(0x16e406cae0e0?)
	/workspace/adamic-tools/go/src/os/exec/exec.go:768 +0x2c
created by os/exec.(*Cmd).Start in goroutine 178
	/workspace/adamic-tools/go/src/os/exec/exec.go:767 +0x92b
FAIL	github.com/system-inc/adamic/stage1/typescript/parser	90.059s
FAIL
```

### 06-fuzz attempt

Exact scratch overlay paths: `cmd/adamic-fuzz/main.go`, `cmd/adamic-fuzz/main_test.go`, `cmd/adamic-reduce/main.go`, `internal/fuzz/deadline_test.go`, `internal/fuzz/field_representation.go`, `internal/fuzz/fuzz_test.go`, `internal/fuzz/generate.go`, `internal/fuzz/judge_test.go`, `internal/fuzz/liveness.go`, `internal/fuzz/liveness_fields_test.go`, `internal/fuzz/operators.go`, `internal/fuzz/operators_test.go`, `internal/fuzz/reduce.go`, `internal/fuzz/reduce_test.go`, `internal/fuzz/run.go`, `internal/fuzz/runtime_deadline.go`, `internal/fuzz/runtime_test.go`, `internal/fuzz/shrink.go`, `internal/fuzz/undefined_numbers.go`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`, `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`, `internal/native/compiler_test.go`, `internal/native/heap_test.go`, `internal/native/library.go`, `internal/native/library_test.go`, `internal/native/map_hash_test.go`, `internal/native/native.go`, `internal/native/release_flags.go`, `internal/native/wasm_test.go`.

* build: `go build ./internal/fuzz ./cmd/adamic-fuzz ./cmd/adamic-reduce`, exit **0**, 20.01s.

* test: `go test -count=1 -timeout=90s ./internal/fuzz ./cmd/adamic-fuzz ./cmd/adamic-reduce`, exit **1**, 63.42s.

```text
child /tmp/adamic-gate/TestExecuteDeadlineKillsGrandchild2087506511/001/fake-compiler []: deadline/cancellation expired; killing process group; descendants=[]
child go [build -o /tmp/adamic-gate/TestPrepareBuildDeadline4035590220/002/adamic ./cmd/adamic]: deadline/cancellation expired; killing process group; descendants=[]
child /tmp/adamic-gate/go-build2309592905/b241/fuzz.test [-test.run=^TestExecuteCPUHelper$]: deadline/cancellation expired; killing process group; descendants=[]
child /tmp/adamic-gate/go-build2309592905/b241/fuzz.test [-test.run=^TestExecuteCPUHelper$]: deadline/cancellation expired; killing process group; descendants=[]
--- FAIL: TestExecuteCPULimit (3.12s)
    --- FAIL: TestExecuteCPULimit/SIGXCPU (0.01s)
        run_test.go:53: SIGXCPU death was not timed out: {Stdout:[] Stderr:[] ExitCode:-1 Signal:CPU time limit exceeded TimedOut:false}
    --- FAIL: TestExecuteCPULimit/saturated (2.10s)
        run_test.go:74: 1 s CPU workload under saturation: TimedOut=true ExitCode=-1 elapsed=2.036128033s stderr=child /tmp/adamic-gate/go-build2309592905/b241/fuzz.test: deadline exceeded; process group killed
            child /tmp/adamic-gate/go-build2309592905/b241/fuzz.test: context deadline exceeded; process group killed
FAIL
FAIL	github.com/system-inc/adamic/internal/fuzz	44.735s
ok  	github.com/system-inc/adamic/cmd/adamic-fuzz	0.003s
?   	github.com/system-inc/adamic/cmd/adamic-reduce	[no test files]
FAIL
```

### 07-test262 attempt

Exact scratch overlay paths: `cmd/adamic-test262/cache.go`, `cmd/adamic-test262/cache_test.go`, `cmd/adamic-test262/compiler.go`, `cmd/adamic-test262/compiler_test.go`, `cmd/adamic-test262/deadline_test.go`, `cmd/adamic-test262/edit_test.go`, `cmd/adamic-test262/main.go`, `cmd/adamic-test262/measure.py`, `cmd/adamic-test262/measure_edits.py`, `cmd/adamic-test262/native_deadline_test.go`, `cmd/adamic-test262/performance_test.go`, `cmd/adamic-test262/run.go`, `cmd/adamic-test262/runtime_deadline.go`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`, `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`.

* build: `go build ./cmd/adamic-test262`, exit **0**, 10.54s.

* test: `go test -count=1 -timeout=90s ./cmd/adamic-test262`, exit **1**, 0.54s.

```text
# github.com/system-inc/adamic/cmd/adamic-test262 [github.com/system-inc/adamic/cmd/adamic-test262.test]
cmd/adamic-test262/run_test.go:41:13: undefined: runProgram
cmd/adamic-test262/run_test.go:66:13: undefined: runProgram
FAIL	github.com/system-inc/adamic/cmd/adamic-test262 [build failed]
FAIL
```

### 08-gate attempt

Exact scratch overlay paths: `cloud/gate/fleet.sh`, `cmd/adamic-gate/affinity.go`, `cmd/adamic-gate/affinity_test.go`, `cmd/adamic-gate/archive.go`, `cmd/adamic-gate/archive_test.go`, `cmd/adamic-gate/complement.go`, `cmd/adamic-gate/complement_test.go`, `cmd/adamic-gate/concurrency.go`, `cmd/adamic-gate/concurrency_test.go`, `cmd/adamic-gate/deadline_test.go`, `cmd/adamic-gate/discovery.go`, `cmd/adamic-gate/frozen.go`, `cmd/adamic-gate/frozen_test.go`, `cmd/adamic-gate/historical_test.go`, `cmd/adamic-gate/host_darwin.go`, `cmd/adamic-gate/host_linux.go`, `cmd/adamic-gate/host_test.go`, `cmd/adamic-gate/json_command.go`, `cmd/adamic-gate/literal_children.go`, `cmd/adamic-gate/literal_children_test.go`, `cmd/adamic-gate/main.go`, `cmd/adamic-gate/main_test.go`, `cmd/adamic-gate/package_order.go`, `cmd/adamic-gate/package_order_test.go`, `cmd/adamic-gate/provenance.go`, `cmd/adamic-gate/provenance_test.go`, `cmd/adamic-gate/reference.go`, `cmd/adamic-gate/reference_test.go`, `cmd/adamic-gate/required_environment.go`, `cmd/adamic-gate/required_environment_test.go`, `cmd/adamic-gate/resume.go`, `cmd/adamic-gate/skip_policy.go`, `cmd/adamic-gate/skip_policy_test.go`, `cmd/adamic-gate/submodules.go`, `cmd/adamic-gate/submodules_test.go`, `cmd/adamic-gate/timing.go`, `cmd/adamic-gate/timing_test.go`, `cmd/adamic-gate/timings.json`, `cmd/adamic-gate/wasi.go`, `cmd/adamic-gate/wasi_test.go`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`.

* build: `go build ./cmd/adamic-gate`, exit **1**, 0.04s.

```text
cmd/adamic-gate/skip_policy.go:12:2: no required module provides package github.com/system-inc/adamic/internal/skipcensus; to add it:
	go get github.com/system-inc/adamic/internal/skipcensus
```

* test: `go test -count=1 -timeout=90s ./cmd/adamic-gate`, exit **1**, 0.03s.

```text
# github.com/system-inc/adamic/cmd/adamic-gate
cmd/adamic-gate/skip_policy.go:12:2: no required module provides package github.com/system-inc/adamic/internal/skipcensus; to add it:
	go get github.com/system-inc/adamic/internal/skipcensus
FAIL	github.com/system-inc/adamic/cmd/adamic-gate [setup failed]
FAIL
```

### 10-setup attempt

Exact scratch overlay paths: `cloud/css-input-mutants.py`, `cloud/cycle-ledger-mutants.py`, `cloud/gate-input-mutants.py`, `cloud/gate-inputs/README.md`, `cloud/gate-inputs/css-printer/npm-bootstrap.json`, `cloud/gate-inputs/css-printer/package-lock.json`, `cloud/gate-inputs/css-printer/package.json`, `cloud/gate-inputs/css/npm-bootstrap.json`, `cloud/gate-inputs/css/package-lock.json`, `cloud/gate-inputs/css/package.json`, `cloud/gate-inputs/graphql/npm-bootstrap.json`, `cloud/gate-inputs/graphql/package-lock.json`, `cloud/gate-inputs/graphql/package.json`, `cloud/gate-inputs/json-prettier/npm-bootstrap.json`, `cloud/gate-inputs/json-prettier/package-lock.json`, `cloud/gate-inputs/json-prettier/package.json`, `cloud/gate-inputs/media-query/npm-bootstrap.json`, `cloud/gate-inputs/media-query/package-lock.json`, `cloud/gate-inputs/media-query/package.json`, `cloud/gate-inputs/selector/npm-bootstrap.json`, `cloud/gate-inputs/selector/package-lock.json`, `cloud/gate-inputs/selector/package.json`, `cloud/gate-inputs/values/npm-bootstrap.json`, `cloud/gate-inputs/values/package-lock.json`, `cloud/gate-inputs/values/package.json`, `cloud/measure-css-inputs.py`, `cloud/measure-gate-inputs.py`, `cloud/measure-stage3-setup.py`, `cloud/mutate_setup_modules.py`, `cloud/run-css-input-proof.py`, `cloud/run-gate-input-tests.py`, `cloud/setup-darwin.py`, `cloud/setup-gate-inputs.py`, `cloud/setup-gate-npm.py`, `cloud/setup-key.py`, `cloud/setup-modules.py`, `cloud/setup-stage3-api.py`, `cloud/setup.sh`, `cloud/stage3-setup-mutants.py`, `cloud/test_archive_modes.py`, `cloud/test_css_gate_inputs.py`, `cloud/test_cycle_ledger_setup.py`, `cloud/test_gate_inputs.py`, `cloud/test_markdown_setup.py`, `cloud/test_setup.py`, `cloud/test_setup_modules.py`, `cloud/test_stage3_setup.py`, `cloud/testdata/stage3-api/package-lock.json`, `cloud/testdata/stage3-api/package.json`, `cloud/verify-width-preflight.py`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`.

* build: `go build ./cloud`, exit **1**, 0.02s.

```text
github.com/system-inc/adamic/cloud: no non-test Go files in /tmp/devtools-inventory-scratch/cloud
```

* test: `go test -count=1 -timeout=90s ./cloud`, exit **0**, 10.95s.

```text
ok  	github.com/system-inc/adamic/cloud	10.685s
```

* Additional: `python3 -m unittest discover -s cloud -p test_*.py`, exit **1**, 3.58s.

```text
ne_reinstall
    upstream, pin, node, inputs = self.fixture(Path(temporary))
                                  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/tmp/devtools-inventory-scratch/cloud/test_cycle_ledger_setup.py", line 39, in fixture
    version = json.loads((SOURCE / 'node-pin.json').read_text())['version']
                         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/pathlib.py", line 1027, in read_text
    with self.open(mode='r', encoding=encoding, errors=errors) as f:
         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/pathlib.py", line 1013, in open
    return io.open(self, mode, buffering, encoding, errors, newline)
           ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
FileNotFoundError: [Errno 2] No such file or directory: '/tmp/devtools-inventory-scratch/cloud/node-pin.json'

======================================================================
ERROR: test_install_skip_repair_output_and_uncached_bytes (test_cycle_ledger_setup.Ledger.test_install_skip_repair_output_and_uncached_bytes)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/tmp/devtools-inventory-scratch/cloud/test_cycle_ledger_setup.py", line 70, in test_install_skip_repair_output_and_uncached_bytes
    upstream, pin, node, inputs = self.fixture(Path(temporary))
                                  ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/tmp/devtools-inventory-scratch/cloud/test_cycle_ledger_setup.py", line 39, in fixture
    version = json.loads((SOURCE / 'node-pin.json').read_text())['version']
                         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/pathlib.py", line 1027, in read_text
    with self.open(mode='r', encoding=encoding, errors=errors) as f:
         ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
  File "/opt/codex/runtimes/codex-primary-runtime/dependencies/python/lib/python3.12/pathlib.py", line 1013, in open
    return io
```

### 11-lint attempt

Exact scratch overlay paths: `cmd/adamic-lint-check/main.go`, `stage1/cohere/lint/cache_mutants_test.go`, `stage1/cohere/lint/cache_test.go`, `stage1/cohere/lint/lint_test.go`, `stage1/cohere/lint/registry/registry.go`, `stage1/cohere/lint/registry/registry_test.go`, `stage1/cohere/lint/rule_check_test.go`, `stage1/cohere/lint/selection_test.go`, `stage1/cohere/lint/semantic_test.go`, `stage1/cohere/lint/split_test.go`, `stage1/cohere/lint/timing_test.go`. Additional scratch prerequisites: none.

* build: `go build ./stage1/cohere/lint/... ./cmd/adamic-lint-check`, exit **0**, 1.08s.

* test: `go test -count=1 -timeout=90s ./stage1/cohere/lint/... ./cmd/adamic-lint-check`, exit **TIMEOUT**, 110.01s.

```text
{0x1c3ecb0, 0x5, 0x5}, {0xc2aa104eb4d77c7c, 0x14f57632e9, ...})
	/workspace/adamic-tools/go/src/testing/testing.go:2740 +0x510
testing.(*M).Run(0x35c83f0f8b40)
	/workspace/adamic-tools/go/src/testing/testing.go:2600 +0x6af
main.main()
	_testmain.go:54 +0x9b

goroutine 19 [syscall]:
syscall.Syscall6(0xf7, 0x3, 0xa, 0x35c83ee51868, 0x4, 0x35c83edf6090, 0x0)
	/workspace/adamic-tools/go/src/syscall/syscall_linux.go:96 +0x39
internal/syscall/unix.Waitid(0x35c83ee51896?, 0x35c83ee519c0?, 0x53294b?, 0x18?, 0x1b517b0?)
	/workspace/adamic-tools/go/src/internal/syscall/unix/waitid_linux.go:18 +0x39
os.(*Process).pidfdWait.func1(...)
	/workspace/adamic-tools/go/src/os/pidfd_linux.go:108
os.ignoringEINTR(...)
	/workspace/adamic-tools/go/src/os/file_posix.go:263
os.(*Process).pidfdWait(0x35c83ece8180)
	/workspace/adamic-tools/go/src/os/pidfd_linux.go:107 +0x1cf
os.(*Process).wait(0x1b?)
	/workspace/adamic-tools/go/src/os/exec_unix.go:25 +0x1a
os.(*Process).Wait(...)
	/workspace/adamic-tools/go/src/os/exec.go:347
os/exec.(*Cmd).Wait(0x35c83f0fe340)
	/workspace/adamic-tools/go/src/os/exec/exec.go:9
...
1ba8950, 0x35c83ecc6190})
	/workspace/adamic-tools/go/src/bytes/buffer.go:229 +0x98
io.copyBuffer({0x1ba8a70, 0x35c8402e6150}, {0x1ba8950, 0x35c83ecc6190}, {0x0, 0x0, 0x0})
	/workspace/adamic-tools/go/src/io/io.go:415 +0x151
io.Copy(...)
	/workspace/adamic-tools/go/src/io/io.go:388
os.genericWriteTo(0x35c83ecc6190?, {0x1ba8a70?, 0x35c8402e6150?})
	/workspace/adamic-tools/go/src/os/file.go:295 +0x35
os.(*File).WriteTo(0x35c83ecc6190, {0x1ba8a70, 0x35c8402e6150})
	/workspace/adamic-tools/go/src/os/file.go:273 +0x9c
io.copyBuffer({0x1ba8a70, 0x35c8402e6150}, {0x1ba89d0, 0x35c83ecc6190}, {0x0, 0x0, 0x0})
	/workspace/adamic-tools/go/src/io/io.go:411 +0x9d
io.Copy(...)
	/workspace/adamic-tools/go/src/io/io.go:388
os/exec.(*Cmd).writerDescriptor.func1()
	/workspace/adamic-tools/go/src/os/exec/exec.go:613 +0x34
os/exec.(*Cmd).Start.func2(0x5ffc0b80626?)
	/workspace/adamic-tools/go/src/os/exec/exec.go:768 +0x2c
created by os/exec.(*Cmd).Start in goroutine 19
	/workspace/adamic-tools/go/src/os/exec/exec.go:767 +0x92b
FAIL	github.com/system-inc/adamic/stage1/cohere/lint/helpers/comments	90.062s
```

### 13-native-oracle attempt

Exact scratch overlay paths: `internal/native/compiler_test.go`, `internal/native/heap_test.go`, `internal/native/library.go`, `internal/native/library_test.go`, `internal/native/map_hash_test.go`, `internal/native/native.go`, `internal/native/release_flags.go`, `internal/native/wasm_test.go`, `internal/oracle/cache_test.go`, `internal/oracle/counts_test.go`, `internal/oracle/gcc_lane_test.go`, `internal/oracle/input_test.go`, `internal/oracle/json_types_test.go`, `internal/oracle/oct6_mutant_test.go`, `internal/oracle/oracle_test.go`, `internal/oracle/release_flags_test.go`, `internal/oracle/release_guard_test.go`, `internal/oracle/slabs_test.go`, `internal/oracle/wasi_test.go`, `oracle/node.mjs`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`, `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`, `internal/leakcheck/leakcheck.go`.

* build: `go build ./internal/native ./internal/oracle`, exit **1**, 0.36s.

```text
github.com/system-inc/adamic/internal/oracle: no non-test Go files in /tmp/devtools-inventory-scratch/internal/oracle
```

* test: `go test -count=1 -timeout=90s ./internal/native ./internal/oracle`, exit **TIMEOUT**, 110.0s.

### 14-catalog attempt

Exact scratch overlay paths: `verify/catalog/08-narrowed-number-field-39638d9e.patch`, `verify/catalog/DETECTION.md`, `verify/catalog/check.sh`, `verify/catalog/detection.json`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`.

* build: `go build ./verify/catalog`, exit **1**, 0.03s.

```text
no Go files in /tmp/devtools-inventory-scratch/verify/catalog
```

* test: `go test -count=1 -timeout=90s ./verify/catalog`, exit **1**, 0.04s.

```text
# ./verify/catalog
no Go files in /tmp/devtools-inventory-scratch/verify/catalog
FAIL	./verify/catalog [setup failed]
FAIL
```

### 15-coverage attempt

Exact scratch overlay paths: `verify/coverage/analyze.py`, `verify/coverage/measure.sh`, `verify/coverage/runtimelibrary/main.go`. Additional scratch prerequisites: `internal/native/compiler_test.go`, `internal/native/heap_test.go`, `internal/native/library.go`, `internal/native/library_test.go`, `internal/native/map_hash_test.go`, `internal/native/native.go`, `internal/native/release_flags.go`, `internal/native/wasm_test.go`.

* build: `go build ./verify/coverage/runtimelibrary`, exit **0**, 1.18s.

* test: `go test -count=1 -timeout=90s ./verify/coverage/runtimelibrary`, exit **0**, 0.3s.

```text
?   	github.com/system-inc/adamic/verify/coverage/runtimelibrary	[no test files]
```

### 02-nodepin-complete attempt

Exact scratch overlay paths: `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`.

* build: `go build ./internal/nodepin`, exit **0**, 0.07s.

* test: `go test -count=1 ./internal/nodepin`, exit **0**, 0.22s.

```text
ok  	github.com/system-inc/adamic/internal/nodepin	0.003s
```

* Additional: `python3 -m unittest discover -s cloud -p test_node_setup.py`, exit **1**, 0.16s.

```text
..E....
======================================================================
ERROR: test_portable_darwin_flow (test_node_setup.NodeSetup.test_portable_darwin_flow)
----------------------------------------------------------------------
Traceback (most recent call last):
  File "/tmp/devtools-inventory-scratch/cloud/test_node_setup.py", line 111, in test_portable_darwin_flow
    spec.loader.exec_module(darwin)
  File "<frozen importlib._bootstrap_external>", line 995, in exec_module
  File "<frozen importlib._bootstrap_external>", line 1132, in get_code
  File "<frozen importlib._bootstrap_external>", line 1190, in get_data
FileNotFoundError: [Errno 2] No such file or directory: '/tmp/devtools-inventory-scratch/cloud/setup-darwin.py'

----------------------------------------------------------------------
Ran 7 tests in 0.013s

FAILED (errors=1)

```

### 10-setup-with-pin attempt

Exact scratch overlay paths: `cloud/css-input-mutants.py`, `cloud/cycle-ledger-mutants.py`, `cloud/gate-input-mutants.py`, `cloud/gate-inputs/README.md`, `cloud/gate-inputs/css-printer/npm-bootstrap.json`, `cloud/gate-inputs/css-printer/package-lock.json`, `cloud/gate-inputs/css-printer/package.json`, `cloud/gate-inputs/css/npm-bootstrap.json`, `cloud/gate-inputs/css/package-lock.json`, `cloud/gate-inputs/css/package.json`, `cloud/gate-inputs/graphql/npm-bootstrap.json`, `cloud/gate-inputs/graphql/package-lock.json`, `cloud/gate-inputs/graphql/package.json`, `cloud/gate-inputs/json-prettier/npm-bootstrap.json`, `cloud/gate-inputs/json-prettier/package-lock.json`, `cloud/gate-inputs/json-prettier/package.json`, `cloud/gate-inputs/media-query/npm-bootstrap.json`, `cloud/gate-inputs/media-query/package-lock.json`, `cloud/gate-inputs/media-query/package.json`, `cloud/gate-inputs/selector/npm-bootstrap.json`, `cloud/gate-inputs/selector/package-lock.json`, `cloud/gate-inputs/selector/package.json`, `cloud/gate-inputs/values/npm-bootstrap.json`, `cloud/gate-inputs/values/package-lock.json`, `cloud/gate-inputs/values/package.json`, `cloud/measure-css-inputs.py`, `cloud/measure-gate-inputs.py`, `cloud/measure-stage3-setup.py`, `cloud/mutate_setup_modules.py`, `cloud/run-css-input-proof.py`, `cloud/run-gate-input-tests.py`, `cloud/setup-darwin.py`, `cloud/setup-gate-inputs.py`, `cloud/setup-gate-npm.py`, `cloud/setup-key.py`, `cloud/setup-modules.py`, `cloud/setup-stage3-api.py`, `cloud/setup.sh`, `cloud/stage3-setup-mutants.py`, `cloud/test_archive_modes.py`, `cloud/test_css_gate_inputs.py`, `cloud/test_cycle_ledger_setup.py`, `cloud/test_gate_inputs.py`, `cloud/test_markdown_setup.py`, `cloud/test_setup.py`, `cloud/test_setup_modules.py`, `cloud/test_stage3_setup.py`, `cloud/testdata/stage3-api/package-lock.json`, `cloud/testdata/stage3-api/package.json`, `cloud/verify-width-preflight.py`. Additional scratch prerequisites: `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`.

* build: `python3 -c import pathlib; [compile(p.read_text(),str(p),"exec") for p in pathlib.Path("cloud").glob("*.py")]`, exit **0**, 0.06s.

* test: `python3 -m unittest discover -s cloud -p test_*.py`, exit **0**, 2.72s.

```text
hint:
hint: Disable this message with "git config set advice.defaultBranchName false"
Initialized empty Git repository in /tmp/adamic-gate/tmpncv7pj2d/inputs/typescript-myvkianw/checkout/.git/
From /tmp/adamic-gate/tmpncv7pj2d/source
 * branch            74335dbb1cb7a9ddb8054ef96053fd6a2668fa83 -> FETCH_HEAD
HEAD is now at 74335db Proof
setup: TypeScript source checkout dirty (tracked, untracked or ignored files); reinstalling
hint: Using 'master' as the name for the initial branch. This default branch name
hint: will change to "main" in Git 3.0. To configure the initial branch name
hint: to use in all of your new repositories, which will suppress this warning,
hint: call:
hint:
hint: 	git config --global init.defaultBranch <name>
hint:
hint: Names commonly chosen instead of 'master' are 'main', 'trunk' and
hint: 'development'. The just-created branch can be renamed via this command:
hint:
hint: 	git branch -m <name>
hint:
hint: Disable this message with "git config set advice.defaultBranchName false"
Initialized empty Git repository in /tmp/adamic-gate/tmpncv7pj2d/inputs/typescript-
...
8fa83 -> FETCH_HEAD
HEAD is now at 74335db Proof
hint: Using 'master' as the name for the initial branch. This default branch name
hint: will change to "main" in Git 3.0. To configure the initial branch name
hint: to use in all of your new repositories, which will suppress this warning,
hint: call:
hint:
hint: 	git config --global init.defaultBranch <name>
hint:
hint: Names commonly chosen instead of 'master' are 'main', 'trunk' and
hint: 'development'. The just-created branch can be renamed via this command:
hint:
hint: 	git branch -m <name>
hint:
hint: Disable this message with "git config set advice.defaultBranchName false"
Initialized empty Git repository in /tmp/adamic-gate/tmpncv7pj2d/inputs/typescript-2hr268si/checkout/.git/
From /tmp/adamic-gate/tmpncv7pj2d/source
 * branch            74335dbb1cb7a9ddb8054ef96053fd6a2668fa83 -> FETCH_HEAD
HEAD is now at 74335db Proof
....ss...........s.....setup: downloading module dependencies in /tmp/adamic-gate/tmpmo4hm63n
....s...
----------------------------------------------------------------------
Ran 49 tests in 2.515s

OK (skipped=5)
```

* Additional: `bash -n cloud/setup.sh`, exit **0**, 0.0s.

```text

```

### 08-gate-with-census attempt

Exact scratch overlay paths: `cloud/gate/fleet.sh`, `cmd/adamic-gate/affinity.go`, `cmd/adamic-gate/affinity_test.go`, `cmd/adamic-gate/archive.go`, `cmd/adamic-gate/archive_test.go`, `cmd/adamic-gate/complement.go`, `cmd/adamic-gate/complement_test.go`, `cmd/adamic-gate/concurrency.go`, `cmd/adamic-gate/concurrency_test.go`, `cmd/adamic-gate/deadline_test.go`, `cmd/adamic-gate/discovery.go`, `cmd/adamic-gate/frozen.go`, `cmd/adamic-gate/frozen_test.go`, `cmd/adamic-gate/historical_test.go`, `cmd/adamic-gate/host_darwin.go`, `cmd/adamic-gate/host_linux.go`, `cmd/adamic-gate/host_test.go`, `cmd/adamic-gate/json_command.go`, `cmd/adamic-gate/literal_children.go`, `cmd/adamic-gate/literal_children_test.go`, `cmd/adamic-gate/main.go`, `cmd/adamic-gate/main_test.go`, `cmd/adamic-gate/package_order.go`, `cmd/adamic-gate/package_order_test.go`, `cmd/adamic-gate/provenance.go`, `cmd/adamic-gate/provenance_test.go`, `cmd/adamic-gate/reference.go`, `cmd/adamic-gate/reference_test.go`, `cmd/adamic-gate/required_environment.go`, `cmd/adamic-gate/required_environment_test.go`, `cmd/adamic-gate/resume.go`, `cmd/adamic-gate/skip_policy.go`, `cmd/adamic-gate/skip_policy_test.go`, `cmd/adamic-gate/submodules.go`, `cmd/adamic-gate/submodules_test.go`, `cmd/adamic-gate/timing.go`, `cmd/adamic-gate/timing_test.go`, `cmd/adamic-gate/timings.json`, `cmd/adamic-gate/wasi.go`, `cmd/adamic-gate/wasi_test.go`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`, `internal/skipcensus/census.go`, `internal/skipcensus/census_test.go`, `internal/skipcensus/cmd/main.go`, `internal/skipcensus/log.go`, `internal/skipcensus/optin.go`, `internal/skipcensus/testdata/plain-skips.jsonl`, `internal/skipcensus/testdata/proofs.py`, `internal/skipcensus/testdata/skips.json`, `stage1/typescript/parser/parser_test.go`.

* build: `go build ./cmd/adamic-gate`, exit **0**, 1.08s.

* test: `go test -count=1 -timeout=90s ./cmd/adamic-gate`, exit **1**, 22.59s.

```text
ql/printer/testdata/cohere_side_test.go:TestAdamicPrinter:071bee04ebc1306585f9ad6a0a4aa5333046920fcb7bc52d74b659e6bbc9199d (TestAdamicPrinter:19)
        undeclared skip stage1/cohere/markdownblocks/width_test.go:TestMarkdownUnicodeWidths:18f38ccf9cb523f1097ccdd6e3e20139492062d476c53e6c8b8a182a10c6f8f0 (TestMarkdownUnicodeWidths:20)
        undeclared skip stage1/cohere/selector/selector_test.go:TestCorpusKeepsEveryParseableFile:d477d9e8698360d4235fd6018e2b01d8580d25ea79afd8417ad44c9dfcdd86fc (TestCorpusKeepsEveryParseableFile:478)
        undeclared skip stage1/cohere/static_single_assignment/testdata/cohere_side_test.go:TestAdamicPortCases:4fd418791213190c8c02147c8709ae05919a7aa60a46f77c98fc6a9fcb512052 (TestAdamicPortCases:43)
        undeclared skip stage1/cohere/tsprinter/corpus_test.go:TestExpressionsAgainstGoAndPrettier,TestMutants,TestStatementsAgainstGoAndPrettier:1dbf140729f17ef866b29b574189ddf438e2a4dd6762d9bdcd8e5dae68bf6b7e (printerCorpusFiles:47)
        undeclared skip stage1/cohere/tsprinter/doc_test.go:TestDocumentsAgainstGoAndPrettier:d477d9e8698360d4235fd6018e2b01d
...
1)
        undeclared skip stage1/cohere/yaml/lexer_test.go:TestLexerMatchesGo:d477d9e8698360d4235fd6018e2b01d8580d25ea79afd8417ad44c9dfcdd86fc (TestLexerMatchesGo:214)
        undeclared skip stage1/cohere/yaml/props_test.go:TestPropsMatchGo:d477d9e8698360d4235fd6018e2b01d8580d25ea79afd8417ad44c9dfcdd86fc (TestPropsMatchGo:109)
        undeclared skip stage1/cohere/yaml/scalar_test.go:TestScalarsMatchGo:d477d9e8698360d4235fd6018e2b01d8580d25ea79afd8417ad44c9dfcdd86fc (TestScalarsMatchGo:136)
        undeclared skip stage1/cohere/yaml/schema_test.go:TestSchemaMatchesGo:d477d9e8698360d4235fd6018e2b01d8580d25ea79afd8417ad44c9dfcdd86fc (TestSchemaMatchesGo:106)
        undeclared skip stage1/cohere/yaml/unist_test.go:TestUnistMatchesGo:d477d9e8698360d4235fd6018e2b01d8580d25ea79afd8417ad44c9dfcdd86fc (TestUnistMatchesGo:84)
        undeclared skip stage3/fixtures/runner_guard_test.go:TestTransformedNodeRunnerGuardHook:365ddb0b6660a8b806fdffbcf304774accc4879f412d115bfd7d912a4f669ab2 (TestTransformedNodeRunnerGuardHook:13)
FAIL
FAIL	github.com/system-inc/adamic/cmd/adamic-gate	21.758s
FAIL
```

### 04-metamorphic-with-fuzz attempt

Exact scratch overlay paths: `cmd/adamic-metamorphic/main.go`, `internal/metamorphic/alias.go`, `internal/metamorphic/fixtures.go`, `internal/metamorphic/identity.go`, `internal/metamorphic/metamorphic_test.go`, `internal/metamorphic/method.go`, `internal/metamorphic/run.go`, `internal/metamorphic/source.go`, `internal/metamorphic/symbols.go`, `internal/metamorphic/temporaries.go`, `internal/metamorphic/wrap.go`. Additional scratch prerequisites: `cloud/node-pin.json`, `cloud/node-setup-mutants.py`, `cloud/setup-node.py`, `cloud/test_node_setup.py`, `cmd/adamic-fuzz/main.go`, `cmd/adamic-fuzz/main_test.go`, `cmd/adamic-reduce/main.go`, `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`, `internal/fuzz/deadline_test.go`, `internal/fuzz/field_representation.go`, `internal/fuzz/fuzz_test.go`, `internal/fuzz/generate.go`, `internal/fuzz/judge_test.go`, `internal/fuzz/liveness.go`, `internal/fuzz/liveness_fields_test.go`, `internal/fuzz/operators.go`, `internal/fuzz/operators_test.go`, `internal/fuzz/reduce.go`, `internal/fuzz/reduce_test.go`, `internal/fuzz/run.go`, `internal/fuzz/runtime_deadline.go`, `internal/fuzz/runtime_test.go`, `internal/fuzz/shrink.go`, `internal/fuzz/undefined_numbers.go`, `internal/nodepin/nodepin.go`, `internal/nodepin/nodepin_test.go`, `internal/native/compiler_test.go`, `internal/native/heap_test.go`, `internal/native/library.go`, `internal/native/library_test.go`, `internal/native/map_hash_test.go`, `internal/native/native.go`, `internal/native/release_flags.go`, `internal/native/wasm_test.go`.

* build: `go build ./cmd/adamic-metamorphic`, exit **0**, 5.65s.

* test: `go test -count=1 -timeout=90s ./internal/metamorphic ./cmd/adamic-metamorphic`, exit **0**, 7.71s.

```text
ok  	github.com/system-inc/adamic/internal/metamorphic	1.893s
?   	github.com/system-inc/adamic/cmd/adamic-metamorphic	[no test files]
```

### 14-catalog-script attempt

Exact scratch overlay paths: `verify/catalog/08-narrowed-number-field-39638d9e.patch`, `verify/catalog/DETECTION.md`, `verify/catalog/check.sh`, `verify/catalog/detection.json`. Additional scratch prerequisites: `internal/boundedrun/.gitignore`, `internal/boundedrun/identity.go`, `internal/boundedrun/prove.py`, `internal/boundedrun/python.py`, `internal/boundedrun/python_test.py`, `internal/boundedrun/run.go`, `internal/boundedrun/run_test.go`, `internal/boundedrun/shell.sh`, `internal/boundedrun/testfixture/hang.go`.

* build: `bash -n verify/catalog/check.sh`, exit **0**, 0.0s.

* test: `git apply --check verify/catalog/08-narrowed-number-field-39638d9e.patch`, exit **0**, 0.0s.

### 15-coverage-script attempt

Exact scratch overlay paths: `verify/coverage/analyze.py`, `verify/coverage/measure.sh`, `verify/coverage/runtimelibrary/main.go`. Additional scratch prerequisites: `internal/native/compiler_test.go`, `internal/native/heap_test.go`, `internal/native/library.go`, `internal/native/library_test.go`, `internal/native/map_hash_test.go`, `internal/native/native.go`, `internal/native/release_flags.go`, `internal/native/wasm_test.go`.

* build: `go build ./verify/coverage/runtimelibrary`, exit **0**, 0.06s.

* test: `bash -n verify/coverage/measure.sh`, exit **0**, 0.0s.

* Additional: `python3 -c import pathlib; compile(pathlib.Path("verify/coverage/analyze.py").read_text(),"analyze.py","exec")`, exit **0**, 0.06s.

```text

```

## What refusalprobe is and how to repair it

This is a developer test-program generator, not compiler functionality. It creates a rejected construct inside varied surroundings and an accepted neighbor; it checks the specific refusal diagnostic, treats broken neighbors as invalid probes, emits findings, and AST-audits refusal coverage. `cmd/adamic-refusals` exposes it. Main does not need it to compile or to enforce enum/optional-widening rules: main already has `internal/lower/enums.go`, `internal/lower/optional_widening.go` and their tests. Main can benefit from the independent audit/generator after its promises are brought up to date.

The reported red is reproducible on main plus the three package files. `ValidateCatalog` parses refusals.go, notices `if err := l.<helper>(node)` calls, and requires every helper to have a fixed owner in its helpers map. Add `enumRefusal` and `refuseOptionalWidening` owners in probe.go. The optional-widening catalog entry already has a bad structural-view program, a repaired neighbor and `adamic/no-optional-widening` identity.

Two added map keys are enough to remove the **audit** failure, but are not a sufficient semantic repair. The enum entry currently says the loader rejects enum syntax, whereas main supports enums. Give enumRefusal explicit entries for the actual enum-tag/enum-object refusals and supported neighbors (plus distinguish NotYet boundaries such as nested/merged/ambient enum declarations). Generate and lower the bad and good programs against main. Retain a mutant that removes each mapping/entry and proves the audit can fail, and a refusal/neighbor mutant that proves the generator can catch an incorrect acceptance or wrong diagnostic. Do not weaken the audit or classify accepted enums as forbidden. Small repair scope is catalog.go, probe.go and probe_test.go; CLI needs no behavioral change. A separate main-based scratch test below checks the two-map-key mechanical repair, without claiming its stale enum boundary is correct.


refusal-map-build:

```text

```

refusal-map-test:

```text
ok  	github.com/system-inc/adamic/internal/refusalprobe	10.184s
?   	github.com/system-inc/adamic/cmd/adamic-refusals	[no test files]

```
## Ordered fix-forward plan

Every row is a new short branch from freshly fetched main, one commit-sized behavior change. Preserve main's newer code and fix forward on that branch; do not cherry-pick the area merge or copy shared files wholesale. Expand the braces below into exact paths; the ledger lists all carried source paths and their contributing merges. A successful family check does not certify smaller subsets that have not been separately exercised.

| Order | Small landing | Smallest paths / hunks | Required check before landing |
| ---: | --- | --- | --- |
| 1 | Bounded Go children | `internal/boundedrun/{identity.go,run.go,run_test.go,testfixture/hang.go}` | go build ./internal/boundedrun; go test -count=1 ./internal/boundedrun; prove hanging process group dies |
| 2 | Python/shell bounded adapters | `internal/boundedrun/{python.py,python_test.py,shell.sh,prove.py,.gitignore}` | Python unit suite and prove.py/shell hanging-child proof; bash -n |
| 3 | Node pin primitive, then installer | First `internal/nodepin/{nodepin.go,nodepin_test.go}`; separate `cloud/{node-pin.json,setup-node.py,test_node_setup.py,node-setup-mutants.py}` and setup.sh Node-only wiring | Package build/tests; test_node_setup.py and checksum/wrong-version mutant; run setup and inspect final PATH/version |
| 4 | Parser opt-in must fail when input absent | Three added lines in `stage1/typescript/parser/parser_test.go` | ADAMIC_PARSER_BENCH=1 without source must fail; ordinary unset input remains declared skip; supplied pinned source passes |
| 5 | Skip census policy | internal/skipcensus source/tests/cmd and live testdata, regenerated from current main after above guards | go run ./internal/skipcensus/cmd -scan; manually classify declarations; all tests; unknown/new/missing/required-input skip mutants fail |
| 6 | Repair refusal catalog, then add generator to main | `internal/refusalprobe/{catalog.go,probe.go,probe_test.go}`, `cmd/adamic-refusals/main.go` | Build and all package tests; real enum/optional bad+neighbor probes; audit and wrong-diagnostic mutants |
| 7 | Preserve fuzz severity during reduction | `internal/fuzz/{reduce.go,reduce_test.go,shrink.go,judge_test.go,fuzz_test.go}`, corresponding `cmd/adamic-{fuzz,reduce}/main.go` call/signature hunks and fuzz CLI tests; only required judge/API hunks in run.go | Build three touched packages; all tests; crash must remain crash after shrinking; preserve current CPU-limit tests |
| 8 | Export fuzz execution/judging API | Only export/new-verdict hunks in internal/fuzz/run.go plus exact consumer prerequisites; boundedrun/nodepin call sites separately | Full internal/fuzz and CLI tests; main's TestExecuteCPULimit remains green under saturation |
| 9 | Fuzz liveness and field representation | `internal/fuzz/{liveness.go,field_representation.go,liveness_fields_test.go}`, generate.go feature/call hunks | Build/test internal/fuzz; seeded new-scene run on main; mutant demonstrating catch/finally or field presence disagreement |
| 10 | Operator generator scenes | `internal/fuzz/{operators.go,operators_test.go}`, generate.go and undefined_numbers.go feature hunks; required reducer filtering hunks | Full fuzz tests, every enabled feature compiled/run against main; keep unsupported features opt-in until proven |
| 11 | Metamorphic checker | Ten internal/metamorphic Go files + cmd/adamic-metamorphic/main.go, after API prerequisite | Package/CLI build and tests; run representative identity/alias/wrap/method/temporary transformations on current main, release+sanitized+leak paths |
| 12 | Test262 bounded children | `cmd/adamic-test262/{runtime_deadline.go,deadline_test.go,native_deadline_test.go}` and compiler/cache/run/edit/performance-test deadline hunks | CLI build and full package tests; hanging compiler/native child proof; cache invalidation after timeout; pin wiring separately |
| 13 | Modules and stage3 setup | Separate `cloud/{setup-modules.py,test_setup_modules.py,mutate_setup_modules.py}` plus module wiring; then `{setup-stage3-api.py,test_stage3_setup.py,stage3-setup-mutants.py,testdata/stage3-api/*}` plus setup-key.py and stage3-only setup.sh hunks | Each Python suite and mutants, cold/warm main setup; current workspace/source sums restored on exit |
| 14 | Optional corpus/npm setup | `cloud/{setup-gate-inputs.py,setup-gate-npm.py,test_gate_inputs.py,gate-input-mutants.py,gate-inputs/*/{package.json,package-lock.json,npm-bootstrap.json}}` and corpus-only setup.sh hunks; CSS and cycle-ledger tests/hunks as separate follow-ups | Corresponding Python tests/mutants; unset/ordinary setup clears optional inputs; opt-in runs real main correctness tests, not recorded reports |
| 15 | Archive-once setup | test_archive_modes.py plus archive-mode setup-gate-inputs.py/setup.sh hunks, after corpus inputs and Node installer | Unit conflict tests; ordinary/no-archive does not build archive; requested archive verifies provenance; real cold/warm execution |
| 16 | Darwin and preflight setup | setup-darwin.py, verify-width-preflight.py, test_markdown_setup.py and only relevant setup.sh/test_setup.py hunks | Platform unit tests; real macOS setup before claiming macOS green; existing markdown bootstrap tests |
| 17 | GCC differential lane | Native compiler-selection hunks in native.go/library.go, compiler_test.go, library_test.go compiler hunks, oracle/gcc_lane_test.go | Native build/tests; ADAMIC_ORACLE_GCC lane against Node, sanitizer and leak failure mutant; wait for protected leak API |
| 18 | Release-flags lane | native/release_flags.go, oracle/{release_flags_test.go,release_guard_test.go} and required native helper hunks | Native package tests and full opt-in release lane; faulty release-only compiler/runtime mutant |
| 19 | Slab/malloc coverage and WASI site checks | Slabs export in native.go and native/{heap_test.go,map_hash_test.go}; oracle/slabs_test.go separately; wasm_test.go and oracle/wasi_test.go guard hunks separately | Corresponding opt-in slab/malloc/WASI tests with tools present, plus known allocator mutant; no silent skips |
| 20 | JSON oracle helper/cache | oracle/node.mjs, internal/oracle/json_types_test.go and only JSON-helper/cache context/TestMain hunks in cache_test.go | After pinning and leak owner, full oracle suite; current oracle/json_types.go helper when present; corrupt helper/source cache mutant; source Node still truth |
| 21 | Leak callers beyond protected extraction | Each stage1/cohere port_test/support_test/etc. import and leakChecked call-site from family 12; markdown width preflight separately | Wait for mac-base-green owner; each port's uncached full test suite and leak mutant; one port per commit |
| 22 | Selected lint rule authoring | cmd/adamic-lint-check/main.go, registry.go/registry_test.go selection hunks; rule_check_test.go, selection_test.go, semantic_test.go, split_test.go, timing_test.go, required lint_test.go hunks | Build CLI; selected rule plus mutant; selected/full parity; all lint/... tests with normal 90m ceiling |
| 23 | Lint cache invalidation | lint/{cache_test.go,cache_mutants_test.go} plus precise lint_test.go call-site hunks | All lint tests uncached and cached; each key-removal mutant fails; edited rule invalidates semantic artifact |
| 24 | Gate child deadlines and package order | Gate main/resume call-site hunks with deadline_test.go; package_order.go/package_order_test.go separately | Build/test cmd/adamic-gate; hung children killed; main discovery complete; ordering preserves verdict |
| 25 | Gate independent parent selection | discovery.go, literal_children.go/literal_children_test.go, complement.go/complement_test.go and only integrated main/resume selectors | Full gate package tests; raw go test JSON proves every discovered unit exactly once; omitted/duplicate selector mutant |
| 26 | Gate required-input and WASI contract | required_environment.go/tests, skip_policy.go/tests, wasi.go/tests, archive.go/tests and integrated plan/shard hunks; separate commits per contract | Package tests; missing input must fail before work; real fully supplied shard; absence is never claimed coverage |
| 27 | Frozen/reference provenance | frozen.go/tests, reference.go/tests, provenance.go/tests, historical_test.go, submodules.go/tests, host_*.go/host_test.go, required main integration hunks | Full package tests; wrong source/submodule/env/raw-evidence mutations fail; current-main raw reference generated afresh |
| 28 | Gate timing, affinity and concurrency | timing.go/tests, affinity.go/tests, concurrency.go/tests, json_command.go and main plan/timing hooks; regenerate timings.json only after code | Full tests; full current-main partition coverage versus plain reference; measured per-shard memory/time. Archived timings/audit are not proof |
| 29 | Regression catalog bounded runner / refreshed mutant | verify/catalog/check.sh wrapper first; field patch separately after adaptation | bash -n; bounded helper tests; git apply --check on main; uncached control+mutant, not old detection.json |
| 30 | Coverage instrumentation and tools | native/native.go coverage+library_test.go coverage hunks; fuzz/run.go profile hunks; verify/coverage/{measure.sh,analyze.py,runtimelibrary/main.go} | Native/fuzz build/tests, ordinary cache separated from coverage cache, LLVM profiles generated; fresh representative coverage report |
| 31 | Fleet execution and ancillary proof/measurement scripts | cloud/gate/fleet.sh after gate contracts; remaining cloud measure/run/mutant scripts individually with needed setup helpers | Shell/Python syntax; fixture/mock command tests; explicit supplied environment; run each proof against fresh main. Publishing a fleet is outside this inventory |

The cohere-gate extraction is already owned on `devtools/cohere-gate-main`, and leakcheck plus oracle leak tests on `devtools/mac-base-green`. They are prerequisites/coordination points, not extra landings for this unit. Resolve U policy/docs with those owners. Each later landing should rerun the relevant checks on its actual newest main, refresh generated inventories, and record its full base/result SHA. No row above is authorized to skip a failing check just because another row needs it.

After the worthwhile independent checks land, there is no demonstrated reason to keep a developer-tools integration area. What remains is historical evidence, measurement inputs that should be regenerated, or unresolved/superseded hunks to archive after review. Preserve the branch as a historical reference if desired; retire its integration role. This inventory does not delete any branch or evidence.

## By directory

Directories are the first two components (root files stay separate). These rows partition the 837-path delta and count each path once.

| Directory | Class | Files | Added | Removed | Binary |
| --- | --- | ---: | ---: | ---: | ---: |
| `CLAUDE.md` | U | 1 | 12 | 1 | 0 |
| `CohereSettings.json` | U | 1 | 1 | 1 | 0 |
| `cloud/cohere-baseline.json` | B | 1 | 985 | 0 | 0 |
| `cloud/cohere_gate.py` | B | 1 | 216 | 0 | 0 |
| `cloud/cohere_gate_test.go` | B | 1 | 30 | 0 | 0 |
| `cloud/css-input-mutants.py` | B | 1 | 43 | 0 | 0 |
| `cloud/cycle-ledger-mutants.py` | B | 1 | 39 | 0 | 0 |
| `cloud/gate` | B | 1 | 231 | 0 | 0 |
| `cloud/gate-input-mutants.py` | B | 1 | 67 | 0 | 0 |
| `cloud/gate-inputs` | B | 21 | 749 | 0 | 0 |
| `cloud/gate-inputs` | C | 1 | 52 | 0 | 0 |
| `cloud/measure-css-inputs.py` | B | 1 | 55 | 0 | 0 |
| `cloud/measure-gate-inputs.py` | B | 1 | 36 | 0 | 0 |
| `cloud/measure-stage3-setup.py` | B | 1 | 50 | 0 | 0 |
| `cloud/mutate_setup_modules.py` | B | 1 | 33 | 0 | 0 |
| `cloud/node-pin.json` | B | 1 | 9 | 0 | 0 |
| `cloud/node-setup-mutants.py` | B | 1 | 31 | 0 | 0 |
| `cloud/reports` | C | 446 | 17,211 | 0 | 0 |
| `cloud/run-css-input-proof.py` | B | 1 | 58 | 0 | 0 |
| `cloud/run-gate-input-tests.py` | B | 1 | 85 | 0 | 0 |
| `cloud/setup-darwin.py` | B | 1 | 134 | 0 | 0 |
| `cloud/setup-gate-inputs.py` | B | 1 | 416 | 0 | 0 |
| `cloud/setup-gate-npm.py` | B | 1 | 105 | 0 | 0 |
| `cloud/setup-key.py` | B | 1 | 2 | 1 | 0 |
| `cloud/setup-modules.py` | B | 1 | 110 | 0 | 0 |
| `cloud/setup-node.py` | B | 1 | 106 | 0 | 0 |
| `cloud/setup-stage3-api.py` | B | 1 | 90 | 0 | 0 |
| `cloud/setup.sh` | B | 1 | 159 | 45 | 0 |
| `cloud/stage3-setup-mutants.py` | B | 1 | 61 | 0 | 0 |
| `cloud/test_archive_modes.py` | B | 1 | 69 | 0 | 0 |
| `cloud/test_css_gate_inputs.py` | B | 1 | 115 | 0 | 0 |
| `cloud/test_cycle_ledger_setup.py` | B | 1 | 145 | 0 | 0 |
| `cloud/test_gate_inputs.py` | B | 1 | 141 | 0 | 0 |
| `cloud/test_markdown_setup.py` | B | 1 | 4 | 1 | 0 |
| `cloud/test_node_setup.py` | B | 1 | 136 | 0 | 0 |
| `cloud/test_setup.py` | B | 1 | 7 | 2 | 0 |
| `cloud/test_setup_modules.py` | B | 1 | 79 | 0 | 0 |
| `cloud/test_stage3_setup.py` | B | 1 | 115 | 0 | 0 |
| `cloud/testdata` | B | 2 | 53 | 0 | 0 |
| `cloud/verify-width-preflight.py` | B | 1 | 47 | 0 | 0 |
| `cmd/adamic-fuzz` | B | 2 | 90 | 18 | 0 |
| `cmd/adamic-gate` | B | 39 | 11,127 | 2,993 | 0 |
| `cmd/adamic-gate` | C | 33 | 96,434 | 0 | 21 |
| `cmd/adamic-lint-check` | B | 1 | 24 | 0 | 0 |
| `cmd/adamic-metamorphic` | B | 1 | 493 | 0 | 0 |
| `cmd/adamic-metamorphic` | C | 1 | 557 | 0 | 0 |
| `cmd/adamic-reduce` | B | 1 | 8 | 2 | 0 |
| `cmd/adamic-refusals` | B | 1 | 102 | 0 | 0 |
| `cmd/adamic-test262` | B | 13 | 217 | 39 | 0 |
| `docs/gate-inputs.md` | C | 1 | 91 | 0 | 0 |
| `docs/gate-shards.md` | C | 1 | 1,533 | 7 | 0 |
| `docs/gcc-lane-completion-evidence.json.gz` | C | 1 | 0 | 0 | 1 |
| `docs/gcc-lane-completion-timings.json` | C | 1 | 242 | 0 | 0 |
| `docs/gcc-lane-completion-warnings.tsv` | C | 1 | 28,905 | 0 | 0 |
| `docs/gcc-lane-completion.md` | C | 1 | 162 | 0 | 0 |
| `docs/gcc-lane-diagnostics.tsv` | C | 1 | 28 | 0 | 0 |
| `docs/gcc-lane-measure.py` | C | 1 | 52 | 0 | 0 |
| `docs/gcc-lane-timings.json` | C | 1 | 122 | 0 | 0 |
| `docs/gcc-lane.md` | C | 1 | 173 | 0 | 0 |
| `docs/lint-registration.md` | U | 1 | 81 | 0 | 0 |
| `docs/skip-census` | C | 26 | 822 | 0 | 0 |
| `docs/skip-census.md` | C | 1 | 66 | 0 | 0 |
| `internal/boundedrun` | B | 9 | 885 | 0 | 0 |
| `internal/fuzz` | B | 15 | 1,875 | 51 | 0 |
| `internal/fuzz` | U | 1 | 175 | 46 | 0 |
| `internal/leakcheck` | B | 1 | 168 | 0 | 0 |
| `internal/metamorphic` | B | 10 | 1,867 | 0 | 0 |
| `internal/native` | B | 8 | 192 | 17 | 0 |
| `internal/nodepin` | B | 2 | 80 | 0 | 0 |
| `internal/oracle` | B | 11 | 1,182 | 119 | 0 |
| `internal/oracle` | C | 39 | 3,700 | 0 | 1 |
| `internal/refusalprobe` | B | 4 | 598 | 0 | 0 |
| `internal/refusalprobe` | C | 58 | 2,550 | 0 | 0 |
| `internal/skipcensus` | B | 8 | 2,726 | 0 | 0 |
| `oracle/node.mjs` | B | 1 | 17 | 2 | 0 |
| `stage1/cohere` | B | 24 | 1,411 | 314 | 0 |
| `stage1/typescript` | B | 1 | 3 | 0 | 0 |
| `verify/catalog` | B | 2 | 30 | 2 | 0 |
| `verify/catalog` | C | 2 | 1,177 | 0 | 0 |
| `verify/coverage` | B | 3 | 766 | 0 | 0 |
| `verify/coverage` | C | 2 | 7,409 | 0 | 0 |

## By area merge

All 47 `Merge devtools/<x> at <sha> into area/developer-tools` commits on the area's exclusive first-parent history are listed. “Surviving paths” intersects that merge's first-parent path changes with the final three-dot ledger. The class counts below count those paths' **final** delta lines, not the historical merge's transient additions. A shared path can occur in multiple merge rows; do not sum these rows. The full ledger's provenance columns make that overlap explicit. Direct commits and main merges may subsequently change a path; these provenance labels identify contributing area merges, not sole authorship of every final line.

| Merge (full SHA) | Subject | Surviving paths | B paths / added / removed | C paths / added / removed | U paths / added / removed |
| --- | --- | ---: | --- | --- | --- |
| `88a6cf07cb8e484a32584ef089db19ddb6286c61` | Merge devtools/gate-shards at 72836ef1 into area/developer-tools | 6 | 4 / 1,100 / 161 | 2 / 1,533 / 7 | 0 / 0 / 0 |
| `b3dd43e5e004f1ae7441b3f8141027addbefb92b` | Merge devtools/gate-fleet at 8226f4f3 into area/developer-tools | 1 | 1 / 231 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `3f904d5364d39c9a1d03845fe2981537d5227561` | Merge devtools/detection-table at 77fdd85b into area/developer-tools | 2 | 0 / 0 / 0 | 2 / 1,177 / 0 | 0 / 0 / 0 |
| `bf35c1cbf0cfba856abb081d094da7a1946c8e4d` | Merge devtools/child-deadlines at 97ac3d6e into area/developer-tools | 32 | 29 / 1,814 / 247 | 2 / 1,562 / 0 | 1 / 175 / 46 |
| `c1ef38bb2058d40604ccc574ed93fea796bfa96c` | Merge devtools/refusal-probes at 5e3c7b73 into area/developer-tools | 50 | 5 / 700 / 0 | 45 / 2,626 / 0 | 0 / 0 / 0 |
| `e765be2885bdec79a3960bc0d3abed1bb045910c` | Merge devtools/generator-liveness-fields-merge at aec16918 into area/developer-tools | 7 | 5 / 835 / 31 | 2 / 1,177 / 0 | 0 / 0 / 0 |
| `8117291c2fa94f443ac791e7594ff68f8a9578d1` | Merge devtools/generator-operators at b058a914 into area/developer-tools | 8 | 7 / 873 / 45 | 0 / 0 / 0 | 1 / 175 / 46 |
| `c2c76c4b233ae08c5b50feef9b4d5ef792791736` | Merge devtools/refusal-lies at 4fb0b5ed into area/developer-tools | 15 | 0 / 0 / 0 | 15 / 184 / 0 | 0 / 0 / 0 |
| `3335b274d12fb0aa39f1d8ce207949b3f2f8c0cc` | Merge devtools/gate-inputs at 7598edc9 into area/developer-tools | 209 | 30 / 1,769 / 48 | 179 / 6,497 / 0 | 0 / 0 / 0 |
| `7863216ed419151286c34df3896496dd15f4ec11` | Merge devtools/gate-fleet at f3f03825 into area/developer-tools | 1 | 1 / 231 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `15ab80659cf70a3cd0ddd292b7c4bdb6084f2d33` | Merge devtools/judge-severity at 58894bd0 into area/developer-tools | 9 | 8 / 437 / 37 | 0 / 0 / 0 | 1 / 175 / 46 |
| `82fb7b469b2f2f4aa1f25e47cb0f2aab667ac68a` | Merge devtools/release-lane at 5cb0e9b7 into area/developer-tools | 3 | 3 / 565 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `664a608d38aa62f09ffb88f76bc98c381f20ad3b` | Merge devtools/gate-shards at 9a79ab0d into area/developer-tools | 31 | 22 / 2,617 / 189 | 8 / 1,657 / 7 | 1 / 175 / 46 |
| `1f9eabad5627a579978741556b16fed9245236ef` | Merge devtools/gcc-lane at 774d0cac into area/developer-tools | 12 | 4 / 399 / 8 | 8 / 29,684 / 0 | 0 / 0 / 0 |
| `fa93eed4afca4d7762f8779cdb9e9e674b6c6786` | Merge devtools/generator-coverage at b2bf06e2 into area/developer-tools | 8 | 5 / 858 / 7 | 2 / 7,409 / 0 | 1 / 175 / 46 |
| `0de7704410b74a1b4cb09e6399dc8aaf72267508` | Merge devtools/refusal-catalog-predicate at b646946d into area/developer-tools | 2 | 2 / 421 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `3848a8401ba2d99804884e71d097e39d33d65559` | Merge devtools/setup-modules at a8909116 into area/developer-tools | 24 | 5 / 388 / 47 | 19 / 1,089 / 0 | 0 / 0 / 0 |
| `2adf65c2514eefbbb073bfdab922d00a4e80a3d3` | Merge devtools/gate-fleet at 96644baa into area/developer-tools | 1 | 1 / 231 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `52cfc645ba85a3561ad1b64498a99205b484b0b4` | Merge devtools/gate-fleet at a24c0874 into area/developer-tools | 1 | 1 / 231 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `5794c87661b030d24cee457bbec4900f260e025d` | Merge devtools/slab-lane at a811b867 into area/developer-tools | 6 | 4 / 276 / 8 | 2 / 56 / 0 | 0 / 0 / 0 |
| `ce477b0460ca2ebb3a1428c20130a5c8d8194e0c` | Merge devtools/cohere-gate at 7d6d85e7 into area/developer-tools | 5 | 3 / 1,231 / 0 | 0 / 0 / 0 | 2 / 13 / 2 |
| `478f4c6fe7f7f3edb50956c3fcc3fd1d0bffab06` | Merge devtools/merge-main-4e0bfda at 04f0746c into area/developer-tools | 14 | 12 / 353 / 173 | 0 / 0 / 0 | 2 / 93 / 1 |
| `6bfb3279eeb17e59c3bd79b723473316e7b402c9` | Merge devtools/metamorphic at 0caa80b1 into area/developer-tools | 13 | 11 / 2,360 / 0 | 1 / 557 / 0 | 1 / 175 / 46 |
| `a80542b00faefabcb60ba58f248392a8fa7398ad` | Merge devtools/node-pin at d01540b8 into area/developer-tools | 193 | 24 / 2,006 / 52 | 167 / 2,707 / 0 | 2 / 187 / 47 |
| `30ba5a31fd47a4a65282c2d6c255347ca389d71a` | Merge devtools/goproxy-fallback at 15004457 into area/developer-tools | 1 | 1 / 159 / 45 | 0 / 0 / 0 | 0 / 0 / 0 |
| `f13e632eebcf9233047f6401db097e6331526cc4` | Merge devtools/gate-shards at 2a017877 into area/developer-tools | 16 | 10 / 7,343 / 2,980 | 6 / 2,746 / 7 | 0 / 0 / 0 |
| `cbd532a7fe8a09855708c27854f60612c3bbf61a` | Merge devtools/gate-fleet at 7bfab950 into area/developer-tools | 1 | 1 / 231 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `8df9eca528a4ba9169e4f0f76c26b45e71b9e352` | Merge devtools/gate-shards at 23c2cca2 into area/developer-tools | 7 | 5 / 1,503 / 145 | 2 / 1,533 / 7 | 0 / 0 / 0 |
| `6edcd86de24b1b5cfca609e96171880e66b5c8ed` | Merge devtools/oracle-json-types at c9ed0639 into area/developer-tools | 40 | 3 / 151 / 5 | 37 / 3,644 / 0 | 0 / 0 / 0 |
| `837ae7e79832713cf1f9b1ad1aefd2b6a1acbba2` | Merge devtools/ignore-lies at 1f3f99f3 into area/developer-tools | 2 | 0 / 0 / 0 | 0 / 0 / 0 | 2 / 13 / 2 |
| `e2a8b7a1e4b18e1f150877a10bd0d02a8132369c` | Merge devtools/cohere-gate-fixtures at 077e8df4 into area/developer-tools | 4 | 2 / 1,201 / 0 | 0 / 0 / 0 | 2 / 13 / 2 |
| `e754e6a9cc1e30bc7767148b38b8532fbd862632` | Merge devtools/cohere-gate-fixtures at b39468a3 into area/developer-tools | 2 | 1 / 216 / 0 | 0 / 0 / 0 | 1 / 12 / 1 |
| `47fbaf174d168f14cd50dedcaabfaa09c51cdefd` | Merge devtools/gate-shards at b2001113 into area/developer-tools | 5 | 3 / 534 / 145 | 2 / 1,533 / 7 | 0 / 0 / 0 |
| `540fa7f0ad2c770a16040d671e7c2d6c6c148aea` | Merge devtools/skip-census at 10e709ce into area/developer-tools | 24 | 7 / 2,498 / 0 | 17 / 674 / 0 | 0 / 0 / 0 |
| `1b40bbeea623a2784e9287c91fc95c7a9f0cc4c9` | Merge devtools/stage1-leaks at f6eef5df into area/developer-tools | 19 | 18 / 260 / 415 | 0 / 0 / 0 | 1 / 12 / 1 |
| `d2d1a8208a8aebd3489bca42309f36ba62afcff3` | Merge devtools/work-sums-area at 42484be5 into area/developer-tools | 1 | 1 / 159 / 45 | 0 / 0 / 0 | 0 / 0 / 0 |
| `976ce7fa3f4e157353947fcdd60d24ae60ae4f9a` | Merge devtools/gate-shards at 0081e98d into area/developer-tools | 8 | 6 / 1,567 / 145 | 2 / 1,533 / 7 | 0 / 0 / 0 |
| `6b33709083b75368ee35e18340a70836abc72b68` | Merge devtools/gcc-lane-leakcheck at ae3075b7 into area/developer-tools | 1 | 1 / 263 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |
| `08e0db2036acd710df716f4f6856191da3adb0be` | Merge devtools/skip-census at e165f424 into area/developer-tools | 21 | 8 / 2,537 / 9 | 13 / 462 / 0 | 0 / 0 / 0 |
| `5586b7739e99af44965e45aa48723ee8779eec21` | Merge devtools/lint-rule-check-2 at ec04fd64 into area/developer-tools | 12 | 11 / 1,375 / 11 | 0 / 0 / 0 | 1 / 81 / 0 |
| `ac3d541c88fd30f13aa354eb80ec1e3615044983` | Merge devtools/gate-inputs-css at bd575e8b into area/developer-tools | 58 | 7 / 987 / 45 | 51 / 5,288 / 0 | 0 / 0 / 0 |
| `18a920c8c832205f2d31f5cfc3c7475955f0a596` | Merge devtools/gate-archive-once at 77fc4632 into area/developer-tools | 38 | 4 / 778 / 45 | 34 / 1,838 / 0 | 0 / 0 / 0 |
| `cbe51cb4d32822793d5d02ce4f7713d0fa4c1529` | Merge devtools/gate-inputs-ledger at 4d12276a into area/developer-tools | 6 | 5 / 900 / 45 | 1 / 52 / 0 | 0 / 0 / 0 |
| `28746d7dbb966af5e0c6082eb9beacffffaeeacd` | Merge devtools/slab-sites at af9a505e into area/developer-tools | 3 | 3 / 1,755 / 3 | 0 / 0 / 0 | 0 / 0 / 0 |
| `68ab1a29829ee3cd05b5001e8a872e819f6aeb64` | Merge devtools/wasi-path at 206da79a into area/developer-tools | 2 | 2 / 1,577 / 6 | 0 / 0 / 0 | 0 / 0 / 0 |
| `2d89ae23f15659bc16786d51aed4095a370b5493` | Merge devtools/gate-affinity at ab1c3d04 into area/developer-tools | 37 | 21 / 8,996 / 2,980 | 16 / 95,068 / 7 | 0 / 0 / 0 |
| `358dbccd95a909a849d1d9db370ae7122a87fc5b` | Merge devtools/census-refresh at d1ce3568 into area/developer-tools | 1 | 1 / 1,568 / 0 | 0 / 0 / 0 | 0 / 0 / 0 |

## Complete path ledger

Every path in the requested diff appears exactly once. Family numbers refer to the live-family table, even for that family's historical records. Provenance is the ordered list of contributing area merge IDs from the merge table. `-` line counts mean binary. C inherits the reader assessment above; B inherits the family purpose/conflicts/check result above. U's specific ambiguity is explained above. The ledger is reproducible with `git diff --numstat 855d114e9b37776ec3739f25d63dbf4da968d02e 5fb868e3f08583b9a0d5cd3fc0fcac474eb12f68`.

| Path | Class | Family | Added | Removed | Area merge provenance |
| --- | --- | --- | ---: | ---: | --- |
| `CLAUDE.md` | U | unresolved | 12 | 1 | `ce477b04`, `478f4c6f`, `a80542b0`, `837ae7e7`, `e2a8b7a1`, `e754e6a9`, `1b40bbee` |
| `CohereSettings.json` | U | 09-cohere-owned | 1 | 1 | `ce477b04`, `837ae7e7`, `e2a8b7a1` |
| `cloud/cohere-baseline.json` | B | 09-cohere-owned | 985 | 0 | `ce477b04`, `e2a8b7a1` |
| `cloud/cohere_gate.py` | B | 09-cohere-owned | 216 | 0 | `ce477b04`, `e2a8b7a1`, `e754e6a9` |
| `cloud/cohere_gate_test.go` | B | 09-cohere-owned | 30 | 0 | `ce477b04` |
| `cloud/css-input-mutants.py` | B | 10-setup | 43 | 0 | `ac3d541c` |
| `cloud/cycle-ledger-mutants.py` | B | 10-setup | 39 | 0 | `cbe51cb4` |
| `cloud/gate-input-mutants.py` | B | 10-setup | 67 | 0 | `3335b274` |
| `cloud/gate-inputs/README.md` | C | 10-setup | 52 | 0 | `3335b274`, `a80542b0`, `ac3d541c`, `18a920c8`, `cbe51cb4` |
| `cloud/gate-inputs/css-printer/npm-bootstrap.json` | B | 10-setup | 5 | 0 | `3335b274` |
| `cloud/gate-inputs/css-printer/package-lock.json` | B | 10-setup | 318 | 0 | `3335b274`, `a80542b0` |
| `cloud/gate-inputs/css-printer/package.json` | B | 10-setup | 13 | 0 | `3335b274`, `a80542b0` |
| `cloud/gate-inputs/css/npm-bootstrap.json` | B | 10-setup | 5 | 0 | `3335b274` |
| `cloud/gate-inputs/css/package-lock.json` | B | 10-setup | 103 | 0 | `3335b274` |
| `cloud/gate-inputs/css/package.json` | B | 10-setup | 9 | 0 | `3335b274` |
| `cloud/gate-inputs/graphql/npm-bootstrap.json` | B | 10-setup | 5 | 0 | `3335b274` |
| `cloud/gate-inputs/graphql/package-lock.json` | B | 10-setup | 24 | 0 | `3335b274` |
| `cloud/gate-inputs/graphql/package.json` | B | 10-setup | 8 | 0 | `3335b274` |
| `cloud/gate-inputs/json-prettier/npm-bootstrap.json` | B | 10-setup | 5 | 0 | `3335b274` |
| `cloud/gate-inputs/json-prettier/package-lock.json` | B | 10-setup | 30 | 0 | `3335b274` |
| `cloud/gate-inputs/json-prettier/package.json` | B | 10-setup | 8 | 0 | `3335b274` |
| `cloud/gate-inputs/media-query/npm-bootstrap.json` | B | 10-setup | 5 | 0 | `3335b274` |
| `cloud/gate-inputs/media-query/package-lock.json` | B | 10-setup | 21 | 0 | `3335b274` |
| `cloud/gate-inputs/media-query/package.json` | B | 10-setup | 8 | 0 | `3335b274` |
| `cloud/gate-inputs/selector/npm-bootstrap.json` | B | 10-setup | 5 | 0 | `3335b274` |
| `cloud/gate-inputs/selector/package-lock.json` | B | 10-setup | 107 | 0 | `3335b274` |
| `cloud/gate-inputs/selector/package.json` | B | 10-setup | 9 | 0 | `3335b274` |
| `cloud/gate-inputs/values/npm-bootstrap.json` | B | 10-setup | 5 | 0 | `3335b274` |
| `cloud/gate-inputs/values/package-lock.json` | B | 10-setup | 48 | 0 | `3335b274` |
| `cloud/gate-inputs/values/package.json` | B | 10-setup | 8 | 0 | `3335b274` |
| `cloud/gate/fleet.sh` | B | 08-gate | 231 | 0 | `b3dd43e5`, `7863216e`, `2adf65c2`, `52cfc645`, `cbd532a7`, `2d89ae23` |
| `cloud/measure-css-inputs.py` | B | 10-setup | 55 | 0 | `ac3d541c` |
| `cloud/measure-gate-inputs.py` | B | 10-setup | 36 | 0 | `3335b274` |
| `cloud/measure-stage3-setup.py` | B | 10-setup | 50 | 0 | `a80542b0` |
| `cloud/mutate_setup_modules.py` | B | 10-setup | 33 | 0 | `3848a840` |
| `cloud/node-pin.json` | B | 02-nodepin | 9 | 0 | `a80542b0` |
| `cloud/node-setup-mutants.py` | B | 02-nodepin | 31 | 0 | `a80542b0` |
| `cloud/reports/css-gate-inputs/REPORT.txt` | C | 10-setup | 61 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestCSSNumbers.log` | C | 10-setup | 24 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestCSSParserOptimizedMatchesNode.log` | C | 10-setup | 12 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestCSSPrinterAgreesWithGo.log` | C | 10-setup | 59 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestCSSPrinterOptimizedMatchesGo.log` | C | 10-setup | 23 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestCSSStrings.log` | C | 10-setup | 24 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestCompositionMatchesGo.log` | C | 10-setup | 26 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestMarkdownInline.log` | C | 10-setup | 25 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/TestThePortParsesAsGoCohereDoes.log` | C | 10-setup | 29 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/all-setup-tests.log` | C | 10-setup | 211 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/cost-cold-1.log` | C | 10-setup | 18 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/cost-cold-2.log` | C | 10-setup | 18 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/cost-cold-3.log` | C | 10-setup | 18 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/cost-warm-1.log` | C | 10-setup | 1 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/cost-warm-2.log` | C | 10-setup | 1 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/cost-warm-3.log` | C | 10-setup | 1 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/costs.json` | C | 10-setup | 240 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/costs.log` | C | 10-setup | 6 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/existing-unit.log` | C | 10-setup | 90 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/exports.log` | C | 10-setup | 18 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-degraded-counts.log` | C | 10-setup | 14 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-drop-commit.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-drop-counts.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-drop-helper.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-drop-sparse.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-drop-url.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-missing-ADAMIC_CSSNUMBERS_LIBRARY.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-missing-ADAMIC_CSSSTRINGS_LIBRARY.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-missing-ADAMIC_CSS_FIXTURES.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-missing-ADAMIC_MARKDOWNINLINE_LIBRARY.log` | C | 10-setup | 13 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-skip-checkout-bytes.log` | C | 10-setup | 32 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-skip-package-prettier-version.log` | C | 10-setup | 29 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-skip-prefix-prettier-version.log` | C | 10-setup | 14 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-skip-prettier-version.log` | C | 10-setup | 14 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-wrong-head.log` | C | 10-setup | 18 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutant-wrong-sparse.log` | C | 10-setup | 18 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutants.json` | C | 10-setup | 77 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/mutants.log` | C | 10-setup | 15 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/ordinary-unset.log` | C | 10-setup | 4 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/parity-run.log` | C | 10-setup | 8 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/recorded-postcss-discrepancies.json` | C | 10-setup | 26 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/recorded-printer-discrepancies.json` | C | 10-setup | 3146 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/setup-first.log` | C | 10-setup | 131 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/setup-ordinary.log` | C | 10-setup | 63 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/setup-tip.log` | C | 10-setup | 111 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/setup-warm.log` | C | 10-setup | 77 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/uncached-fixtures.log` | C | 10-setup | 18 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/unit.log` | C | 10-setup | 110 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/verdicts.json` | C | 10-setup | 289 | 0 | `ac3d541c` |
| `cloud/reports/css-gate-inputs/vet.log` | C | 10-setup | 0 | 0 | `ac3d541c` |
| `cloud/reports/estree-gate-input/README.md` | C | 10-setup | 24 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/cache-proof.log` | C | 10-setup | 16 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/cache-proof.py` | C | 10-setup | 15 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/estree-test.log` | C | 10-setup | 9 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/export-mutant.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/exports.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/install.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/npm-tests.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/pin-mutant.log` | C | 10-setup | 19 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/prettier-resolution.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/raw-resolution.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/setup-tests.log` | C | 10-setup | 91 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/shared-resolution.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/shared-resolution.mjs` | C | 10-setup | 9 | 0 | `a80542b0` |
| `cloud/reports/estree-gate-input/without-input.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/gate-archive-once/1-with-cold.log` | C | 10-setup | 77 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/1-with-warm.log` | C | 10-setup | 77 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/1-without-cold.log` | C | 10-setup | 36 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/1-without-warm.log` | C | 10-setup | 76 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/2-with-cold.log` | C | 10-setup | 77 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/2-with-warm.log` | C | 10-setup | 77 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/2-without-cold.log` | C | 10-setup | 76 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/2-without-warm.log` | C | 10-setup | 76 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/3-with-cold.log` | C | 10-setup | 77 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/3-with-warm.log` | C | 10-setup | 77 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/3-without-cold.log` | C | 10-setup | 76 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/3-without-warm.log` | C | 10-setup | 76 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/REPORT.txt` | C | 10-setup | 101 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/consumer-with.log` | C | 10-setup | 8 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/consumer-without.log` | C | 10-setup | 5 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/empty-go-cache-attempt-metadata.txt` | C | 10-setup | 1 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/empty-go-cache-attempt.json` | C | 10-setup | 13 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/empty-go-cache-attempt.log` | C | 10-setup | 67 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/measure.log` | C | 10-setup | 12 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/measure.py` | C | 10-setup | 27 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutant-archive-only-ignored.log` | C | 10-setup | 20 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutant-conflict-diagnostic-lost.log` | C | 10-setup | 13 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutant-conflict-exits-success.log` | C | 10-setup | 14 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutant-default-archive-dropped.log` | C | 10-setup | 22 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutant-leak-archive.log` | C | 10-setup | 14 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutant-leave-stale-variable.log` | C | 10-setup | 14 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutant-no-archive-builds-archive.log` | C | 10-setup | 22 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutants.log` | C | 10-setup | 7 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/mutants.py` | C | 10-setup | 29 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/setup-only.log` | C | 10-setup | 64 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/setup-tests.log` | C | 10-setup | 211 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/setup-without.log` | C | 10-setup | 110 | 0 | `18a920c8` |
| `cloud/reports/gate-archive-once/timings.json` | C | 10-setup | 134 | 0 | `18a920c8` |
| `cloud/reports/gate-inputs/README.md` | C | 10-setup | 84 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-baseline-setup.log` | C | 10-setup | 16 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-branch-archive.log` | C | 10-setup | 1 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-cache-tests-final.log` | C | 10-setup | 90 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-cache-tests.log` | C | 10-setup | 85 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-gitignore-main-clean-repeat.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-main-cold-retry.log` | C | 10-setup | 45 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-main-cold.log` | C | 10-setup | 76 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-main-final-ordinary.log` | C | 10-setup | 16 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-main-final-warm-repeat.log` | C | 10-setup | 28 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-main-final-warm.log` | C | 10-setup | 28 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-main-ordinary-setup.log` | C | 10-setup | 25 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-mutants-final.log` | C | 10-setup | 31 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-npm-tests.log` | C | 10-setup | 7 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-real-uncached.json` | C | 10-setup | 27 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-real-uncached.log` | C | 10-setup | 71 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-setup-tests-final.log` | C | 10-setup | 6 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-setup-tests.log` | C | 10-setup | 14 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-split-branch-with.log` | C | 10-setup | 8 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-split-branch-without.log` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-timing-driver-final.log` | C | 10-setup | 19 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-vet-final.log` | C | 10-setup | 0 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-with-driver.log` | C | 10-setup | 17 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/gate-inputs-without-driver.log` | C | 10-setup | 17 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-artifact-bytes.log` | C | 10-setup | 14 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-artifact-entries.log` | C | 10-setup | 14 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-artifact-modes.log` | C | 10-setup | 14 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-artifact-names.log` | C | 10-setup | 14 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-artifact-root-mode.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-cc_version.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-commit.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-content.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-environment.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-flags.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-head.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-helper.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-kind.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-packages.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-size.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-url.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-validation_flags.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/drop-version.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/list-emits-header.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-bootstrap.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-bytes.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-helper.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-integrity.log` | C | 10-setup | 14 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-lock.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-manifest.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-modes.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-names.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-node.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/npm-drop-root-mode.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/mutants/real-dirty-source.log` | C | 10-setup | 14 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/1-cold-archive.log` | C | 10-setup | 2 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/1-cold-corpora.log` | C | 10-setup | 20 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/1-cold-npm.log` | C | 10-setup | 24 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/1-warm-archive.log` | C | 10-setup | 2 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/1-warm-corpora.log` | C | 10-setup | 3 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/1-warm-npm.log` | C | 10-setup | 8 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/2-cold-archive.log` | C | 10-setup | 2 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/2-cold-corpora.log` | C | 10-setup | 20 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/2-cold-npm.log` | C | 10-setup | 24 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/2-warm-archive.log` | C | 10-setup | 2 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/2-warm-corpora.log` | C | 10-setup | 3 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/2-warm-npm.log` | C | 10-setup | 8 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/3-cold-archive.log` | C | 10-setup | 2 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/3-cold-corpora.log` | C | 10-setup | 20 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/3-cold-npm.log` | C | 10-setup | 24 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/3-warm-archive.log` | C | 10-setup | 2 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/3-warm-corpora.log` | C | 10-setup | 3 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/3-warm-npm.log` | C | 10-setup | 8 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/timings/timings.json` | C | 10-setup | 932 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/checker-split.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/checker-split.jsonl` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/checker-split.log` | C | 10-setup | 3 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-boundaries.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-boundaries.jsonl` | C | 10-setup | 23 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-boundaries.log` | C | 10-setup | 19 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-default.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-default.jsonl` | C | 10-setup | 47 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-default.log` | C | 10-setup | 35 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-narrow.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-narrow.jsonl` | C | 10-setup | 47 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/css-printer-narrow.log` | C | 10-setup | 35 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/gitignore-100MiB.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/gitignore-100MiB.jsonl` | C | 10-setup | 23 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/gitignore-100MiB.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/graphql.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/graphql.jsonl` | C | 10-setup | 21 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/graphql.log` | C | 10-setup | 11 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-corpus-upstream-differences.txt` | C | 10-setup | 27 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-corpus.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-corpus.jsonl` | C | 10-setup | 29 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-corpus.log` | C | 10-setup | 23 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-mutants.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-mutants.jsonl` | C | 10-setup | 66 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-mutants.log` | C | 10-setup | 48 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-numeric.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-numeric.jsonl` | C | 10-setup | 24 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/json-numeric.log` | C | 10-setup | 18 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/lint.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/lint.jsonl` | C | 10-setup | 10 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/lint.log` | C | 10-setup | 6 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/media-query.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/media-query.jsonl` | C | 10-setup | 22 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/media-query.log` | C | 10-setup | 12 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/parser-expressions.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/parser-expressions.jsonl` | C | 10-setup | 9 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/parser-expressions.log` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/parser-whole.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/parser-whole.jsonl` | C | 10-setup | 9 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/parser-whole.log` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/postcss.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/postcss.jsonl` | C | 10-setup | 21 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/postcss.log` | C | 10-setup | 15 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/selector-nontermination.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/selector-nontermination.jsonl` | C | 10-setup | 48 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/selector-nontermination.log` | C | 10-setup | 26 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/selector.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/selector.jsonl` | C | 10-setup | 38 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/selector.log` | C | 10-setup | 28 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/values.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/values.jsonl` | C | 10-setup | 21 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/values.log` | C | 10-setup | 11 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/with/verdicts.json` | C | 10-setup | 716 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/checker-split.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/checker-split.jsonl` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/checker-split.log` | C | 10-setup | 3 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-boundaries.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-boundaries.jsonl` | C | 10-setup | 19 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-boundaries.log` | C | 10-setup | 15 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-default.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-default.jsonl` | C | 10-setup | 46 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-default.log` | C | 10-setup | 34 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-narrow.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-narrow.jsonl` | C | 10-setup | 46 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/css-printer-narrow.log` | C | 10-setup | 34 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/gitignore-100MiB.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/gitignore-100MiB.jsonl` | C | 10-setup | 22 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/gitignore-100MiB.log` | C | 10-setup | 12 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/graphql.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/graphql.jsonl` | C | 10-setup | 22 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/graphql.log` | C | 10-setup | 12 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-corpus.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-corpus.jsonl` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-corpus.log` | C | 10-setup | 7 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-mutants.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-mutants.jsonl` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-mutants.log` | C | 10-setup | 7 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-numeric.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-numeric.jsonl` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/json-numeric.log` | C | 10-setup | 7 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/lint.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/lint.jsonl` | C | 10-setup | 9 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/lint.log` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/media-query.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/media-query.jsonl` | C | 10-setup | 23 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/media-query.log` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/parser-expressions.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/parser-expressions.jsonl` | C | 10-setup | 9 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/parser-expressions.log` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/parser-whole.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/parser-whole.jsonl` | C | 10-setup | 9 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/parser-whole.log` | C | 10-setup | 5 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/postcss.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/postcss.jsonl` | C | 10-setup | 21 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/postcss.log` | C | 10-setup | 15 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/selector-nontermination.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/selector-nontermination.jsonl` | C | 10-setup | 13 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/selector-nontermination.log` | C | 10-setup | 7 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/selector.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/selector.jsonl` | C | 10-setup | 30 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/selector.log` | C | 10-setup | 20 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/values.json` | C | 10-setup | 42 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/values.jsonl` | C | 10-setup | 22 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/values.log` | C | 10-setup | 12 | 0 | `3335b274` |
| `cloud/reports/gate-inputs/without/verdicts.json` | C | 10-setup | 716 | 0 | `3335b274` |
| `cloud/reports/graphql-printer-input/README.md` | C | 10-setup | 28 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/cache-proof.log` | C | 10-setup | 11 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/cache-proof.py` | C | 10-setup | 9 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/exports.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/install.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/missing-export.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/mutants.log` | C | 10-setup | 7 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/mutants.py` | C | 10-setup | 19 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/preflight-disk-failure.log` | C | 10-setup | 53 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/preflight.log` | C | 10-setup | 80 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/resolution.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/setup-tests.log` | C | 10-setup | 91 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/whitespace.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/width-confirmation.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/without-input.log` | C | 10-setup | 8 | 0 | `a80542b0` |
| `cloud/reports/graphql-printer-input/wrong-pin.log` | C | 10-setup | 18 | 0 | `a80542b0` |
| `cloud/reports/node-pin/README.md` | C | 10-setup | 20 | 0 | `a80542b0` |
| `cloud/reports/node-pin/before-node.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/node-pin/consumers.log` | C | 10-setup | 2 | 0 | `a80542b0` |
| `cloud/reports/node-pin/drift-install.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/node-pin/final-check-proof.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/node-pin/final-check-proof.py` | C | 10-setup | 21 | 0 | `a80542b0` |
| `cloud/reports/node-pin/final-drift-refused.log` | C | 10-setup | 22 | 0 | `a80542b0` |
| `cloud/reports/node-pin/fixture-add.log` | C | 10-setup | 0 | 0 | `a80542b0` |
| `cloud/reports/node-pin/fixture-commit.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/node-pin/fixture-git.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/node-pin/full-setup-0.log` | C | 10-setup | 27 | 0 | `a80542b0` |
| `cloud/reports/node-pin/full-setup-1.log` | C | 10-setup | 22 | 0 | `a80542b0` |
| `cloud/reports/node-pin/go-mutants.log` | C | 10-setup | 2 | 0 | `a80542b0` |
| `cloud/reports/node-pin/go-mutants.py` | C | 10-setup | 9 | 0 | `a80542b0` |
| `cloud/reports/node-pin/missing-final-check.log` | C | 10-setup | 22 | 0 | `a80542b0` |
| `cloud/reports/node-pin/missing-oracle-guard.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants.log` | C | 10-setup | 10 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/corrupt-archive.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/drop-architecture.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/drop-checksum.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/drop-helper.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/drop-pin.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/drop-system.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/drop-version.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/installed-bytes.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/installed-mode.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/mutants/published-checksum.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/node-pin/nodepin-test.log` | C | 10-setup | 8 | 0 | `a80542b0` |
| `cloud/reports/node-pin/oracle-after-setup.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/node-pin/oracle-drift-refused.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/node-pin/oracle.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/node-pin/path-proof.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/node-pin/path-proof.py` | C | 10-setup | 25 | 0 | `a80542b0` |
| `cloud/reports/node-pin/pin-drift.log` | C | 10-setup | 8 | 0 | `a80542b0` |
| `cloud/reports/node-pin/real-cache-proof.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/node-pin/real-install.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/node-pin/setup-tests.log` | C | 10-setup | 91 | 0 | `a80542b0` |
| `cloud/reports/node-pin/vet.log` | C | 10-setup | 0 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/README.md` | C | 10-setup | 26 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/alias-mutants.log` | C | 10-setup | 2 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/gate-npm-integration.log` | C | 10-setup | 7 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/markdown-integration.log` | C | 10-setup | 7 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/setup-integration-retry.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/shared-prettier.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/stage3-integration.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/test_gate_inputs.log` | C | 10-setup | 90 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/test_markdown_setup.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/test_setup.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/test_setup_modules.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/setup-fast-merge/test_stage3_setup.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/setup-modules/1-cold.log` | C | 10-setup | 41 | 0 | `3848a840` |
| `cloud/reports/setup-modules/1-warm.log` | C | 10-setup | 1 | 0 | `3848a840` |
| `cloud/reports/setup-modules/2-cold.log` | C | 10-setup | 41 | 0 | `3848a840` |
| `cloud/reports/setup-modules/2-warm.log` | C | 10-setup | 1 | 0 | `3848a840` |
| `cloud/reports/setup-modules/3-cold.log` | C | 10-setup | 41 | 0 | `3848a840` |
| `cloud/reports/setup-modules/3-warm.log` | C | 10-setup | 1 | 0 | `3848a840` |
| `cloud/reports/setup-modules/README.md` | C | 10-setup | 65 | 0 | `3848a840` |
| `cloud/reports/setup-modules/baseline-integration.log` | C | 10-setup | 14 | 0 | `3848a840` |
| `cloud/reports/setup-modules/common-unit.log` | C | 10-setup | 5 | 0 | `3848a840` |
| `cloud/reports/setup-modules/empty-cache-oracle.stderr` | C | 10-setup | 0 | 0 | `3848a840` |
| `cloud/reports/setup-modules/empty-cache-setup-complete.log` | C | 10-setup | 30 | 0 | `3848a840` |
| `cloud/reports/setup-modules/empty-cache-setup-first.log` | C | 10-setup | 85 | 0 | `3848a840` |
| `cloud/reports/setup-modules/integration.log` | C | 10-setup | 14 | 0 | `3848a840` |
| `cloud/reports/setup-modules/invalidation.log` | C | 10-setup | 10 | 0 | `3848a840` |
| `cloud/reports/setup-modules/lint.log` | C | 10-setup | 506 | 0 | `3848a840` |
| `cloud/reports/setup-modules/mutants.log` | C | 10-setup | 16 | 0 | `3848a840` |
| `cloud/reports/setup-modules/timings.json` | C | 10-setup | 206 | 0 | `3848a840` |
| `cloud/reports/setup-modules/unit.log` | C | 10-setup | 6 | 0 | `3848a840` |
| `cloud/reports/setup-modules/without-step.stderr` | C | 10-setup | 6 | 0 | `3848a840` |
| `cloud/reports/stage3-and-width/README.md` | C | 10-setup | 74 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-integration/01.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-integration/02.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-integration/03.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-integration/04.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-integration/05.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-integration/06.log` | C | 10-setup | 3 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-integration/07.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-timings/cold-1.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-timings/cold-2.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-timings/cold-3.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-timings/timings.json` | C | 10-setup | 194 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-timings/warm-1.log` | C | 10-setup | 2 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-timings/warm-2.log` | C | 10-setup | 2 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/api-timings/warm-3.log` | C | 10-setup | 2 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-bootstrap.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-helper.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-link-target.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-lock.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-manifest.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-node.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-tree-bytes.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/drop-tree-modes.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/omit-go-package-lock.json.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/omit-go-package.json.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/omit-go-setup-stage3-api.py.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/cache-mutants/real-drop-lock.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-empty-after-1.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-empty-after-2.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-empty-after-3.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-empty-before-1.log` | C | 10-setup | 23 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-empty-before-2.log` | C | 10-setup | 23 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-empty-before-3.log` | C | 10-setup | 23 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-empty-timings.json` | C | 10-setup | 128 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-node-open-stdin.json` | C | 10-setup | 11 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-node-open-stdin.log` | C | 10-setup | 20 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-phase-before.log` | C | 10-setup | 31 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/markdown-present-after.log` | C | 10-setup | 16 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/setup-two-baseline.log` | C | 10-setup | 16 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/setup-two-integration-final.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/setup-two-integration.log` | C | 10-setup | 14 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/setup-two-vet.log` | C | 10-setup | 0 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-load-before.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-load-final.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-lock-changed.log` | C | 10-setup | 20 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-lock-proof.json` | C | 10-setup | 26 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-lock-restored.log` | C | 10-setup | 20 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-lock-warm.log` | C | 10-setup | 18 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-lower-before.log` | C | 10-setup | 38 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-lower-final.log` | C | 10-setup | 26 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-setup-final.log` | C | 10-setup | 20 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-setup-integration-final.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/stage3-setup-mutants.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/emoji-regex-build.log` | C | 10-setup | 0 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/emoji-regex-mutant.log` | C | 10-setup | 23 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/emoji-regex-present-check.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/emoji-regex.json` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/get-east-asian-width-build.log` | C | 10-setup | 0 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/get-east-asian-width-mutant.log` | C | 10-setup | 23 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/get-east-asian-width-present-check.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/get-east-asian-width.json` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/narrow-emojis-build.log` | C | 10-setup | 0 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/narrow-emojis-mutant.log` | C | 10-setup | 23 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/narrow-emojis-present-check.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-mutants/narrow-emojis.json` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/stage3-and-width/width-preflight-mutants.log` | C | 10-setup | 4 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/README.md` | C | 10-setup | 32 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/cache-proof.log` | C | 10-setup | 15 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/modules-unit.log` | C | 10-setup | 6 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutant-assertions/drop-bootstrap.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutant-assertions/drop-helper.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutant-assertions/drop-lock.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutant-assertions/drop-manifest.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutant-assertions/drop-node.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutant-assertions/missing-export.log` | C | 10-setup | 13 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutant-assertions/wrong-yaml-pin.log` | C | 10-setup | 18 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/mutants.log` | C | 10-setup | 8 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/npm-unit.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/resolution.log` | C | 10-setup | 12 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/setup-unit.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/setup.log` | C | 10-setup | 32 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/stage3-unit.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/unist-resolution.log` | C | 10-setup | 1 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/unit.log` | C | 10-setup | 90 | 0 | `a80542b0` |
| `cloud/reports/yaml-gate-input/yaml-test.log` | C | 10-setup | 5 | 0 | `a80542b0` |
| `cloud/run-css-input-proof.py` | B | 10-setup | 58 | 0 | `ac3d541c` |
| `cloud/run-gate-input-tests.py` | B | 10-setup | 85 | 0 | `3335b274` |
| `cloud/setup-darwin.py` | B | 10-setup | 134 | 0 | `a80542b0`, `18a920c8` |
| `cloud/setup-gate-inputs.py` | B | 10-setup | 416 | 0 | `3335b274`, `a80542b0`, `ac3d541c`, `18a920c8`, `cbe51cb4` |
| `cloud/setup-gate-npm.py` | B | 10-setup | 105 | 0 | `3335b274` |
| `cloud/setup-key.py` | B | 10-setup | 2 | 1 | `a80542b0` |
| `cloud/setup-modules.py` | B | 10-setup | 110 | 0 | `3848a840` |
| `cloud/setup-node.py` | B | 02-nodepin | 106 | 0 | `a80542b0` |
| `cloud/setup-stage3-api.py` | B | 10-setup | 90 | 0 | `a80542b0` |
| `cloud/setup.sh` | B | 10-setup | 159 | 45 | `bf35c1cb`, `3335b274`, `3848a840`, `478f4c6f`, `a80542b0`, `30ba5a31`, `d2d1a820`, `ac3d541c`, `18a920c8`, `cbe51cb4` |
| `cloud/stage3-setup-mutants.py` | B | 10-setup | 61 | 0 | `a80542b0` |
| `cloud/test_archive_modes.py` | B | 10-setup | 69 | 0 | `18a920c8` |
| `cloud/test_css_gate_inputs.py` | B | 10-setup | 115 | 0 | `ac3d541c` |
| `cloud/test_cycle_ledger_setup.py` | B | 10-setup | 145 | 0 | `cbe51cb4` |
| `cloud/test_gate_inputs.py` | B | 10-setup | 141 | 0 | `3335b274`, `a80542b0`, `ac3d541c`, `cbe51cb4` |
| `cloud/test_markdown_setup.py` | B | 10-setup | 4 | 1 | `3335b274` |
| `cloud/test_node_setup.py` | B | 02-nodepin | 136 | 0 | `a80542b0` |
| `cloud/test_setup.py` | B | 10-setup | 7 | 2 | `3335b274`, `3848a840`, `a80542b0` |
| `cloud/test_setup_modules.py` | B | 10-setup | 79 | 0 | `3848a840` |
| `cloud/test_stage3_setup.py` | B | 10-setup | 115 | 0 | `a80542b0` |
| `cloud/testdata/stage3-api/package-lock.json` | B | 10-setup | 44 | 0 | `a80542b0` |
| `cloud/testdata/stage3-api/package.json` | B | 10-setup | 9 | 0 | `a80542b0` |
| `cloud/verify-width-preflight.py` | B | 10-setup | 47 | 0 | `a80542b0` |
| `cmd/adamic-fuzz/main.go` | B | 06-fuzz | 60 | 18 | `15ab8065` |
| `cmd/adamic-fuzz/main_test.go` | B | 06-fuzz | 30 | 0 | `15ab8065` |
| `cmd/adamic-gate/affinity.go` | B | 08-gate | 186 | 0 | `2d89ae23` |
| `cmd/adamic-gate/affinity_test.go` | B | 08-gate | 90 | 0 | `2d89ae23` |
| `cmd/adamic-gate/archive.go` | B | 08-gate | 67 | 0 | `2d89ae23` |
| `cmd/adamic-gate/archive_test.go` | B | 08-gate | 93 | 0 | `2d89ae23` |
| `cmd/adamic-gate/child-deadlines.json` | C | 08-gate | 1428 | 0 | `bf35c1cb` |
| `cmd/adamic-gate/child-deadlines.md` | C | 08-gate | 134 | 0 | `bf35c1cb` |
| `cmd/adamic-gate/complement.go` | B | 08-gate | 148 | 0 | `664a608d`, `f13e632e`, `8df9eca5`, `2d89ae23` |
| `cmd/adamic-gate/complement_test.go` | B | 08-gate | 152 | 0 | `664a608d` |
| `cmd/adamic-gate/concurrency.go` | B | 08-gate | 90 | 0 | `f13e632e` |
| `cmd/adamic-gate/concurrency_test.go` | B | 08-gate | 75 | 0 | `f13e632e` |
| `cmd/adamic-gate/deadline_test.go` | B | 08-gate | 90 | 0 | `bf35c1cb`, `664a608d` |
| `cmd/adamic-gate/discovery.go` | B | 08-gate | 81 | 0 | `664a608d` |
| `cmd/adamic-gate/evidence/affinity-47fb-measurement.json` | C | 08-gate | 230 | 0 | `2d89ae23` |
| `cmd/adamic-gate/evidence/affinity-47fb-plan.json` | C | 08-gate | 18612 | 0 | `2d89ae23` |
| `cmd/adamic-gate/evidence/affinity-47fb-shard-0.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/archive-shard.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/concurrency-trials.tar.gz` | C | 08-gate | - | - | `f13e632e` |
| `cmd/adamic-gate/evidence/coverage-evidence.tar.gz` | C | 08-gate | - | - | `664a608d` |
| `cmd/adamic-gate/evidence/coverage-followup.tar.gz` | C | 08-gate | - | - | `664a608d` |
| `cmd/adamic-gate/evidence/deadline-merge.tar.gz` | C | 08-gate | - | - | `664a608d` |
| `cmd/adamic-gate/evidence/layout-memory-guard.json` | C | 08-gate | 20 | 0 | `2d89ae23` |
| `cmd/adamic-gate/evidence/layout-split-deadline.json` | C | 08-gate | 22 | 0 | `2d89ae23` |
| `cmd/adamic-gate/evidence/layout-split-estimated-plan.json` | C | 08-gate | 18628 | 0 | `2d89ae23` |
| `cmd/adamic-gate/evidence/layout-split-half-a.json` | C | 08-gate | 64 | 0 | `2d89ae23` |
| `cmd/adamic-gate/evidence/layout-split-proof.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/lint-planner.tar.gz` | C | 08-gate | - | - | `47fbaf17` |
| `cmd/adamic-gate/evidence/missing-47.json` | C | 08-gate | 53 | 0 | `664a608d` |
| `cmd/adamic-gate/evidence/oracle-once-33935158-blocked.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/oracle-once-47b8b8d5.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/oracle-partitions.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/plain-47-check.json` | C | 08-gate | 71 | 0 | `664a608d` |
| `cmd/adamic-gate/evidence/plan-once.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/provenance-checks.tgz` | C | 08-gate | - | - | `2d89ae23` |
| `cmd/adamic-gate/evidence/readiness-proofs.tar.gz` | C | 08-gate | - | - | `664a608d` |
| `cmd/adamic-gate/evidence/replan-proofs.tar.gz` | C | 08-gate | - | - | `f13e632e` |
| `cmd/adamic-gate/evidence/replan-provisioned.json` | C | 08-gate | 1213 | 0 | `f13e632e` |
| `cmd/adamic-gate/evidence/required-environment.tar.gz` | C | 08-gate | - | - | `8df9eca5` |
| `cmd/adamic-gate/evidence/skip-census-773d5fc.tgz` | C | 08-gate | - | - | `976ce7fa` |
| `cmd/adamic-gate/evidence/stderr-portability.tar.gz` | C | 08-gate | - | - | `664a608d` |
| `cmd/adamic-gate/evidence/submodule-portability.tar.gz` | C | 08-gate | - | - | `f13e632e` |
| `cmd/adamic-gate/evidence/submodule-provenance.tar.gz` | C | 08-gate | - | - | `f13e632e` |
| `cmd/adamic-gate/evidence/wasi-evidence.tar.gz` | C | 08-gate | - | - | `88a6cf07` |
| `cmd/adamic-gate/frozen.go` | B | 08-gate | 489 | 0 | `2d89ae23` |
| `cmd/adamic-gate/frozen_test.go` | B | 08-gate | 199 | 0 | `2d89ae23` |
| `cmd/adamic-gate/historical_test.go` | B | 08-gate | 30 | 0 | `2d89ae23` |
| `cmd/adamic-gate/host_darwin.go` | B | 08-gate | 25 | 0 | `664a608d` |
| `cmd/adamic-gate/host_linux.go` | B | 08-gate | 32 | 0 | `664a608d` |
| `cmd/adamic-gate/host_test.go` | B | 08-gate | 81 | 0 | `664a608d` |
| `cmd/adamic-gate/json_command.go` | B | 08-gate | 51 | 0 | `664a608d` |
| `cmd/adamic-gate/literal_children.go` | B | 08-gate | 57 | 0 | `47fbaf17` |
| `cmd/adamic-gate/literal_children_test.go` | B | 08-gate | 103 | 0 | `47fbaf17` |
| `cmd/adamic-gate/main.go` | B | 08-gate | 374 | 145 | `88a6cf07`, `bf35c1cb`, `664a608d`, `f13e632e`, `8df9eca5`, `47fbaf17`, `976ce7fa`, `2d89ae23` |
| `cmd/adamic-gate/main_test.go` | B | 08-gate | 14 | 13 | `bf35c1cb`, `664a608d` |
| `cmd/adamic-gate/package_order.go` | B | 08-gate | 54 | 0 | `2d89ae23` |
| `cmd/adamic-gate/package_order_test.go` | B | 08-gate | 19 | 0 | `2d89ae23` |
| `cmd/adamic-gate/provenance.go` | B | 08-gate | 115 | 0 | `2d89ae23` |
| `cmd/adamic-gate/provenance_test.go` | B | 08-gate | 78 | 0 | `2d89ae23` |
| `cmd/adamic-gate/reference.go` | B | 08-gate | 158 | 0 | `f13e632e`, `2d89ae23` |
| `cmd/adamic-gate/reference_test.go` | B | 08-gate | 139 | 0 | `f13e632e`, `2d89ae23` |
| `cmd/adamic-gate/required_environment.go` | B | 08-gate | 266 | 0 | `8df9eca5`, `976ce7fa`, `2d89ae23` |
| `cmd/adamic-gate/required_environment_test.go` | B | 08-gate | 195 | 0 | `8df9eca5`, `976ce7fa` |
| `cmd/adamic-gate/resume.go` | B | 08-gate | 60 | 16 | `88a6cf07`, `664a608d`, `f13e632e`, `2d89ae23` |
| `cmd/adamic-gate/skip_policy.go` | B | 08-gate | 113 | 0 | `664a608d`, `976ce7fa` |
| `cmd/adamic-gate/skip_policy_test.go` | B | 08-gate | 99 | 0 | `664a608d`, `976ce7fa` |
| `cmd/adamic-gate/submodules.go` | B | 08-gate | 131 | 0 | `f13e632e` |
| `cmd/adamic-gate/submodules_test.go` | B | 08-gate | 307 | 0 | `f13e632e` |
| `cmd/adamic-gate/timing.go` | B | 08-gate | 224 | 0 | `2d89ae23` |
| `cmd/adamic-gate/timing_test.go` | B | 08-gate | 115 | 0 | `2d89ae23` |
| `cmd/adamic-gate/timings.json` | B | 08-gate | 5861 | 2819 | `f13e632e`, `2d89ae23` |
| `cmd/adamic-gate/timings.json.audit.json` | C | 08-gate | 55959 | 0 | `2d89ae23` |
| `cmd/adamic-gate/wasi.go` | B | 08-gate | 520 | 0 | `88a6cf07`, `664a608d`, `8df9eca5`, `976ce7fa` |
| `cmd/adamic-gate/wasi_test.go` | B | 08-gate | 146 | 0 | `88a6cf07` |
| `cmd/adamic-lint-check/main.go` | B | 11-lint | 24 | 0 | `5586b773` |
| `cmd/adamic-metamorphic/REPORT.md` | C | 04-metamorphic | 557 | 0 | `6bfb3279` |
| `cmd/adamic-metamorphic/main.go` | B | 04-metamorphic | 493 | 0 | `6bfb3279` |
| `cmd/adamic-reduce/main.go` | B | 06-fuzz | 8 | 2 | `8117291c`, `15ab8065` |
| `cmd/adamic-refusals/main.go` | B | 03-refusalprobe | 102 | 0 | `c1ef38bb` |
| `cmd/adamic-test262/cache.go` | B | 07-test262 | 12 | 2 | `bf35c1cb` |
| `cmd/adamic-test262/cache_test.go` | B | 07-test262 | 2 | 4 | `bf35c1cb` |
| `cmd/adamic-test262/compiler.go` | B | 07-test262 | 32 | 8 | `bf35c1cb` |
| `cmd/adamic-test262/compiler_test.go` | B | 07-test262 | 10 | 8 | `bf35c1cb`, `664a608d` |
| `cmd/adamic-test262/deadline_test.go` | B | 07-test262 | 53 | 0 | `bf35c1cb`, `664a608d` |
| `cmd/adamic-test262/edit_test.go` | B | 07-test262 | 19 | 2 | `bf35c1cb` |
| `cmd/adamic-test262/main.go` | B | 07-test262 | 6 | 0 | `a80542b0` |
| `cmd/adamic-test262/measure.py` | B | 07-test262 | 5 | 2 | `bf35c1cb` |
| `cmd/adamic-test262/measure_edits.py` | B | 07-test262 | 6 | 3 | `bf35c1cb` |
| `cmd/adamic-test262/native_deadline_test.go` | B | 07-test262 | 30 | 0 | `bf35c1cb` |
| `cmd/adamic-test262/performance_test.go` | B | 07-test262 | 7 | 3 | `bf35c1cb` |
| `cmd/adamic-test262/run.go` | B | 07-test262 | 17 | 7 | `bf35c1cb`, `664a608d` |
| `cmd/adamic-test262/runtime_deadline.go` | B | 07-test262 | 18 | 0 | `bf35c1cb` |
| `docs/gate-inputs.md` | C | unresolved | 91 | 0 | `540fa7f0`, `08e0db20` |
| `docs/gate-shards.md` | C | unresolved | 1533 | 7 | `88a6cf07`, `664a608d`, `f13e632e`, `8df9eca5`, `47fbaf17`, `976ce7fa`, `2d89ae23` |
| `docs/gcc-lane-completion-evidence.json.gz` | C | unresolved | - | - | `1f9eabad` |
| `docs/gcc-lane-completion-timings.json` | C | unresolved | 242 | 0 | `1f9eabad` |
| `docs/gcc-lane-completion-warnings.tsv` | C | unresolved | 28905 | 0 | `1f9eabad` |
| `docs/gcc-lane-completion.md` | C | unresolved | 162 | 0 | `1f9eabad` |
| `docs/gcc-lane-diagnostics.tsv` | C | unresolved | 28 | 0 | `1f9eabad` |
| `docs/gcc-lane-measure.py` | C | unresolved | 52 | 0 | `1f9eabad` |
| `docs/gcc-lane-timings.json` | C | unresolved | 122 | 0 | `1f9eabad` |
| `docs/gcc-lane.md` | C | unresolved | 173 | 0 | `1f9eabad` |
| `docs/lint-registration.md` | U | unresolved | 81 | 0 | `478f4c6f`, `5586b773` |
| `docs/skip-census.md` | C | unresolved | 66 | 0 | `540fa7f0`, `08e0db20` |
| `docs/skip-census/area-merge.md` | C | unresolved | 43 | 0 | `540fa7f0` |
| `docs/skip-census/opt-in-proof.md` | C | unresolved | 71 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/flow-mutant.log` | C | unresolved | 16 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/native-clang.log` | C | unresolved | 8 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/oracle-off.log` | C | unresolved | 12 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/parser-missing.log` | C | unresolved | 6 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/restored-skip.log` | C | unresolved | 7 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/rule-mutant.log` | C | unresolved | 16 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/setup.log` | C | unresolved | 29 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/tests.log` | C | unresolved | 60 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/vet.log` | C | unresolved | 0 | 0 | `08e0db20` |
| `docs/skip-census/opt-in-proof/wasi-clang.log` | C | unresolved | 80 | 0 | `08e0db20` |
| `docs/skip-census/report.md` | C | unresolved | 93 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-allow-added-mutant.log` | C | unresolved | 8 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-allow-removed-mutant.log` | C | unresolved | 9 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-allow-required-mutant.log` | C | unresolved | 9 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-area-lane-class-mutant.log` | C | unresolved | 9 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-area-merge-tests.log` | C | unresolved | 50 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-area-release-reason-mutant.log` | C | unresolved | 10 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-latest-area-tests.log` | C | unresolved | 50 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-new-skip-mutant.log` | C | unresolved | 6 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-oracle.log` | C | unresolved | 13 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-paired.json` | C | unresolved | 131 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-plain-check.log` | C | unresolved | 35 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-proofs.log` | C | unresolved | 5 | 0 | `540fa7f0` |
| `docs/skip-census/skip-census-tests.log` | C | unresolved | 46 | 0 | `540fa7f0` |
| `internal/boundedrun/.gitignore` | B | 01-boundedrun | 1 | 0 | `bf35c1cb` |
| `internal/boundedrun/identity.go` | B | 01-boundedrun | 35 | 0 | `bf35c1cb` |
| `internal/boundedrun/prove.py` | B | 01-boundedrun | 101 | 0 | `bf35c1cb` |
| `internal/boundedrun/python.py` | B | 01-boundedrun | 70 | 0 | `bf35c1cb` |
| `internal/boundedrun/python_test.py` | B | 01-boundedrun | 156 | 0 | `bf35c1cb`, `664a608d` |
| `internal/boundedrun/run.go` | B | 01-boundedrun | 203 | 0 | `bf35c1cb`, `664a608d` |
| `internal/boundedrun/run_test.go` | B | 01-boundedrun | 144 | 0 | `bf35c1cb`, `664a608d` |
| `internal/boundedrun/shell.sh` | B | 01-boundedrun | 36 | 0 | `bf35c1cb` |
| `internal/boundedrun/testfixture/hang.go` | B | 01-boundedrun | 139 | 0 | `bf35c1cb`, `664a608d` |
| `internal/fuzz/deadline_test.go` | B | 06-fuzz | 55 | 0 | `bf35c1cb`, `664a608d` |
| `internal/fuzz/field_representation.go` | B | 06-fuzz | 297 | 0 | `e765be28` |
| `internal/fuzz/fuzz_test.go` | B | 06-fuzz | 25 | 1 | `15ab8065` |
| `internal/fuzz/generate.go` | B | 06-fuzz | 55 | 31 | `e765be28`, `8117291c` |
| `internal/fuzz/judge_test.go` | B | 06-fuzz | 150 | 1 | `15ab8065` |
| `internal/fuzz/liveness.go` | B | 06-fuzz | 362 | 0 | `e765be28` |
| `internal/fuzz/liveness_fields_test.go` | B | 06-fuzz | 98 | 0 | `e765be28` |
| `internal/fuzz/operators.go` | B | 06-fuzz | 495 | 0 | `8117291c` |
| `internal/fuzz/operators_test.go` | B | 06-fuzz | 149 | 0 | `8117291c` |
| `internal/fuzz/reduce.go` | B | 06-fuzz | 54 | 12 | `8117291c`, `15ab8065` |
| `internal/fuzz/reduce_test.go` | B | 06-fuzz | 106 | 0 | `8117291c`, `15ab8065` |
| `internal/fuzz/run.go` | U | 06-fuzz | 175 | 46 | `bf35c1cb`, `8117291c`, `15ab8065`, `664a608d`, `fa93eed4`, `6bfb3279`, `a80542b0` |
| `internal/fuzz/runtime_deadline.go` | B | 06-fuzz | 18 | 0 | `bf35c1cb` |
| `internal/fuzz/runtime_test.go` | B | 06-fuzz | 1 | 3 | `bf35c1cb` |
| `internal/fuzz/shrink.go` | B | 06-fuzz | 4 | 3 | `15ab8065` |
| `internal/fuzz/undefined_numbers.go` | B | 06-fuzz | 6 | 0 | `8117291c` |
| `internal/leakcheck/leakcheck.go` | B | 12-leaks-owned | 168 | 0 | `1b40bbee` |
| `internal/metamorphic/alias.go` | B | 04-metamorphic | 62 | 0 | `6bfb3279` |
| `internal/metamorphic/fixtures.go` | B | 04-metamorphic | 71 | 0 | `6bfb3279` |
| `internal/metamorphic/identity.go` | B | 04-metamorphic | 79 | 0 | `6bfb3279` |
| `internal/metamorphic/metamorphic_test.go` | B | 04-metamorphic | 159 | 0 | `6bfb3279` |
| `internal/metamorphic/method.go` | B | 04-metamorphic | 308 | 0 | `6bfb3279` |
| `internal/metamorphic/run.go` | B | 04-metamorphic | 239 | 0 | `6bfb3279` |
| `internal/metamorphic/source.go` | B | 04-metamorphic | 460 | 0 | `6bfb3279` |
| `internal/metamorphic/symbols.go` | B | 04-metamorphic | 126 | 0 | `6bfb3279` |
| `internal/metamorphic/temporaries.go` | B | 04-metamorphic | 215 | 0 | `6bfb3279` |
| `internal/metamorphic/wrap.go` | B | 04-metamorphic | 148 | 0 | `6bfb3279` |
| `internal/native/compiler_test.go` | B | 13-native-oracle | 76 | 0 | `1f9eabad` |
| `internal/native/heap_test.go` | B | 13-native-oracle | 1 | 1 | `5794c876` |
| `internal/native/library.go` | B | 13-native-oracle | 6 | 2 | `1f9eabad`, `478f4c6f` |
| `internal/native/library_test.go` | B | 13-native-oracle | 38 | 1 | `fa93eed4`, `5794c876` |
| `internal/native/map_hash_test.go` | B | 13-native-oracle | 1 | 1 | `478f4c6f` |
| `internal/native/native.go` | B | 13-native-oracle | 54 | 6 | `1f9eabad`, `fa93eed4`, `5794c876`, `478f4c6f` |
| `internal/native/release_flags.go` | B | 13-native-oracle | 7 | 0 | `82fb7b46` |
| `internal/native/wasm_test.go` | B | 13-native-oracle | 9 | 6 | `478f4c6f`, `08e0db20`, `68ab1a29` |
| `internal/nodepin/nodepin.go` | B | 02-nodepin | 31 | 0 | `a80542b0` |
| `internal/nodepin/nodepin_test.go` | B | 02-nodepin | 49 | 0 | `a80542b0` |
| `internal/oracle/cache_test.go` | B | 13-native-oracle | 24 | 3 | `a80542b0`, `6edcd86d` |
| `internal/oracle/counts_test.go` | B | 12-leaks-owned | 2 | 5 | `1b40bbee` |
| `internal/oracle/gcc_lane_test.go` | B | 13-native-oracle | 263 | 0 | `1f9eabad`, `6b337090` |
| `internal/oracle/input_test.go` | B | 12-leaks-owned | 18 | 22 | `478f4c6f`, `1b40bbee` |
| `internal/oracle/json_types_results/base-focused.log` | C | 13-native-oracle | 1 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/base-oracle.log` | C | 13-native-oracle | 1 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/base-vet.log` | C | 13-native-oracle | 0 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/count-builds.py` | C | 13-native-oracle | 13 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/evidence.json` | C | 13-native-oracle | 39 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/full-build-counts.json` | C | 13-native-oracle | 433 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/go-wrapper.py` | C | 13-native-oracle | 4 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/initial-trimpath-timings.json` | C | 13-native-oracle | 122 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-1-after-go.jsonl` | C | 13-native-oracle | 1 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-1-after.log` | C | 13-native-oracle | 104 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-1-before-go.jsonl` | C | 13-native-oracle | 24 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-1-before.log` | C | 13-native-oracle | 104 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-2-after-go.jsonl` | C | 13-native-oracle | 1 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-2-after.log` | C | 13-native-oracle | 104 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-2-before-go.jsonl` | C | 13-native-oracle | 24 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-2-before.log` | C | 13-native-oracle | 104 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-3-after-go.jsonl` | C | 13-native-oracle | 1 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-3-after.log` | C | 13-native-oracle | 104 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-3-before-go.jsonl` | C | 13-native-oracle | 24 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/json-types-final-loaded-3-before.log` | C | 13-native-oracle | 104 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/key-mutant.log` | C | 13-native-oracle | 10 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/loaded-timings.json` | C | 13-native-oracle | 122 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/loader-mutant.log` | C | 13-native-oracle | 19 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/measure.py` | C | 13-native-oracle | 21 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/native-flags.json` | C | 13-native-oracle | 1 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/node-before-after.json` | C | 13-native-oracle | 246 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/node-proof.py` | C | 13-native-oracle | 22 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/scratch-go.jsonl` | C | 13-native-oracle | 1 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/scratch-oracle.log` | C | 13-native-oracle | 4 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/scratch-vet.log` | C | 13-native-oracle | 0 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/source-before.log` | C | 13-native-oracle | 17 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/source-mutant.json` | C | 13-native-oracle | 66 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/source-mutant.log` | C | 13-native-oracle | 23 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/source-mutant.py` | C | 13-native-oracle | 24 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/source-repeated.log` | C | 13-native-oracle | 17 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/source-restored.log` | C | 13-native-oracle | 17 | 0 | `6edcd86d` |
| `internal/oracle/json_types_results/wrong-cwd.log` | C | 13-native-oracle | 1722 | 0 | `6edcd86d` |
| `internal/oracle/json_types_test.go` | B | 13-native-oracle | 110 | 0 | `6edcd86d` |
| `internal/oracle/oct6_mutant_test.go` | B | 12-leaks-owned | 1 | 9 | `1b40bbee` |
| `internal/oracle/oracle_test.go` | B | 12-leaks-owned | 19 | 77 | `478f4c6f`, `1b40bbee` |
| `internal/oracle/release_flags_test.go` | B | 13-native-oracle | 82 | 0 | `82fb7b46`, `08e0db20` |
| `internal/oracle/release_guard_test.go` | B | 13-native-oracle | 476 | 0 | `82fb7b46` |
| `internal/oracle/slabs_evidence.jsonl.gz` | C | 13-native-oracle | - | - | `5794c876` |
| `internal/oracle/slabs_report.md` | C | 13-native-oracle | 56 | 0 | `5794c876` |
| `internal/oracle/slabs_test.go` | B | 13-native-oracle | 183 | 0 | `5794c876`, `28746d7d` |
| `internal/oracle/wasi_test.go` | B | 13-native-oracle | 4 | 3 | `478f4c6f`, `08e0db20`, `28746d7d` |
| `internal/refusalprobe/README.md` | B | 03-refusalprobe | 47 | 0 | `c1ef38bb` |
| `internal/refusalprobe/REPORT.md` | C | 03-refusalprobe | 104 | 0 | `c1ef38bb`, `c2c76c4b` |
| `internal/refusalprobe/catalog.go` | B | 03-refusalprobe | 78 | 0 | `c1ef38bb`, `0de77044` |
| `internal/refusalprobe/lies/function_type_length.a` | C | 03-refusalprobe | 12 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/function_type_length.expected` | C | 03-refusalprobe | 4 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number.a` | C | 03-refusalprobe | 6 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number.expected` | C | 03-refusalprobe | 3 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_array.a` | C | 03-refusalprobe | 8 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_array.expected` | C | 03-refusalprobe | 3 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_ops.a` | C | 03-refusalprobe | 8 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_ops.expected` | C | 03-refusalprobe | 5 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_parameter.a` | C | 03-refusalprobe | 8 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_parameter.expected` | C | 03-refusalprobe | 3 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_return.a` | C | 03-refusalprobe | 8 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_boolean_as_number_return.expected` | C | 03-refusalprobe | 3 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_false_as_number.a` | C | 03-refusalprobe | 6 | 0 | `c2c76c4b` |
| `internal/refusalprobe/lies/widening_false_as_number.expected` | C | 03-refusalprobe | 3 | 0 | `c2c76c4b` |
| `internal/refusalprobe/probe.go` | B | 03-refusalprobe | 343 | 0 | `c1ef38bb`, `0de77044` |
| `internal/refusalprobe/probe_test.go` | B | 03-refusalprobe | 130 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-after-1.jsonl` | C | 03-refusalprobe | 2 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-after-1.log` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-after-2.jsonl` | C | 03-refusalprobe | 2 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-after-2.log` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-after-3.jsonl` | C | 03-refusalprobe | 2 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-after-3.log` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-before-1.jsonl` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-before-1.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-before-2.jsonl` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-before-2.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-before-3.jsonl` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/11-before-3.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-after-1.jsonl` | C | 03-refusalprobe | 2 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-after-1.log` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-after-2.jsonl` | C | 03-refusalprobe | 2 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-after-2.log` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-after-3.jsonl` | C | 03-refusalprobe | 2 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-after-3.log` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-before-1.jsonl` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-before-1.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-before-2.jsonl` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-before-2.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-before-3.jsonl` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/12-before-3.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/catalog-mutant.log` | C | 03-refusalprobe | 7 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/coverage-mutant.log` | C | 03-refusalprobe | 5 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/format.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/judge-mutant.log` | C | 03-refusalprobe | 8 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/lower.log` | C | 03-refusalprobe | 1 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/main-2000.jsonl` | C | 03-refusalprobe | 323 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/main-2000.log` | C | 03-refusalprobe | 322 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/measurements.json` | C | 03-refusalprobe | 1449 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/mode-default.jsonl` | C | 03-refusalprobe | 34 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/mode-default.log` | C | 03-refusalprobe | 33 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/mode-uncached.jsonl` | C | 03-refusalprobe | 34 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/mode-uncached.log` | C | 03-refusalprobe | 33 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/neighbor-mutant.log` | C | 03-refusalprobe | 14 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/oracle.log` | C | 03-refusalprobe | 13 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/patch-11.diff` | C | 03-refusalprobe | 24 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/patch-12.diff` | C | 03-refusalprobe | 22 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/setup.log` | C | 03-refusalprobe | 18 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/tests.log` | C | 03-refusalprobe | 2 | 0 | `c1ef38bb` |
| `internal/refusalprobe/results/vet.log` | C | 03-refusalprobe | 0 | 0 | `c1ef38bb` |
| `internal/skipcensus/census.go` | B | 05-skipcensus | 330 | 0 | `540fa7f0`, `08e0db20` |
| `internal/skipcensus/census_test.go` | B | 05-skipcensus | 313 | 0 | `540fa7f0`, `08e0db20` |
| `internal/skipcensus/cmd/main.go` | B | 05-skipcensus | 58 | 0 | `540fa7f0` |
| `internal/skipcensus/log.go` | B | 05-skipcensus | 100 | 0 | `540fa7f0` |
| `internal/skipcensus/optin.go` | B | 05-skipcensus | 228 | 0 | `08e0db20` |
| `internal/skipcensus/testdata/plain-skips.jsonl` | B | 05-skipcensus | 99 | 0 | `540fa7f0` |
| `internal/skipcensus/testdata/proofs.py` | B | 05-skipcensus | 30 | 0 | `540fa7f0` |
| `internal/skipcensus/testdata/skips.json` | B | 05-skipcensus | 1568 | 0 | `540fa7f0`, `08e0db20`, `28746d7d`, `68ab1a29`, `358dbccd` |
| `oracle/node.mjs` | B | 13-native-oracle | 17 | 2 | `6edcd86d` |
| `stage1/cohere/css/css_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/cssnumbers/port_test.go` | B | 12-leaks-owned | 4 | 24 | `1b40bbee` |
| `stage1/cohere/cssstrings/port_test.go` | B | 12-leaks-owned | 4 | 24 | `1b40bbee` |
| `stage1/cohere/formatfiles/formatfiles_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/gitignore/gitignore_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/graphql/graphql_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/json/port_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/lint/cache_mutants_test.go` | B | 11-lint | 161 | 0 | `5586b773` |
| `stage1/cohere/lint/cache_test.go` | B | 11-lint | 566 | 0 | `5586b773` |
| `stage1/cohere/lint/lint_test.go` | B | 11-lint | 23 | 9 | `478f4c6f`, `5586b773` |
| `stage1/cohere/lint/registry/registry.go` | B | 11-lint | 19 | 2 | `478f4c6f`, `5586b773` |
| `stage1/cohere/lint/registry/registry_test.go` | B | 11-lint | 38 | 0 | `478f4c6f`, `5586b773` |
| `stage1/cohere/lint/rule_check_test.go` | B | 11-lint | 283 | 0 | `5586b773` |
| `stage1/cohere/lint/selection_test.go` | B | 11-lint | 107 | 0 | `5586b773` |
| `stage1/cohere/lint/semantic_test.go` | B | 11-lint | 44 | 0 | `5586b773` |
| `stage1/cohere/lint/split_test.go` | B | 11-lint | 56 | 0 | `5586b773` |
| `stage1/cohere/lint/timing_test.go` | B | 11-lint | 54 | 0 | `5586b773` |
| `stage1/cohere/markdownblocks/support_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/markdownblocks/width_test.go` | B | 12-leaks-owned | 8 | 1 | `a80542b0` |
| `stage1/cohere/markdowninline/port_test.go` | B | 12-leaks-owned | 4 | 24 | `1b40bbee` |
| `stage1/cohere/mediaquery/mediaquery_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/selector/selector_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/suppression/suppression_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/cohere/values/values_test.go` | B | 12-leaks-owned | 4 | 23 | `1b40bbee` |
| `stage1/typescript/parser/parser_test.go` | B | 05-skipcensus | 3 | 0 | `478f4c6f`, `08e0db20` |
| `verify/catalog/08-narrowed-number-field-39638d9e.patch` | B | 14-catalog | 23 | 0 | `e765be28` |
| `verify/catalog/DETECTION.md` | C | 14-catalog | 156 | 0 | `3f904d53`, `c1ef38bb`, `e765be28` |
| `verify/catalog/check.sh` | B | 14-catalog | 7 | 2 | `bf35c1cb` |
| `verify/catalog/detection.json` | C | 14-catalog | 1021 | 0 | `3f904d53`, `e765be28` |
| `verify/coverage/REPORT.json` | C | 15-coverage | 7099 | 0 | `fa93eed4` |
| `verify/coverage/REPORT.md` | C | 15-coverage | 310 | 0 | `fa93eed4` |
| `verify/coverage/analyze.py` | B | 15-coverage | 552 | 0 | `fa93eed4` |
| `verify/coverage/measure.sh` | B | 15-coverage | 187 | 0 | `fa93eed4` |
| `verify/coverage/runtimelibrary/main.go` | B | 15-coverage | 27 | 0 | `fa93eed4` |
