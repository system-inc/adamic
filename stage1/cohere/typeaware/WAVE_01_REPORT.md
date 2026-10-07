Built: native default-option ports of no-implicit-return, no-deprecated and no-else-return, plus three raw checker questions.
Commits: base 0d540f413625f016f20fea39761c7b184f335de6; claim 9887237b19faafdcdff5fdd5c92d0da3bc556adb; implementation follows on codex/typeaware-wave-01.
Commands and outputs: final agreement PASS 88.175s, checker PASS 0.152s, bridge PASS 66.942s, filtered Node oracle PASS 12.806s, vet clean, source lint zero findings.
Mutants: all three rule mutations exit normally and fail byte comparison; released-registry, ABI, ownership, linkage, source-lint and Node byte controls are caught as detailed below.
Not covered: nondefault rule options, full upstream rule suites, full repository Go gate, full CLI/fix engine integration, or a performance improvement.

## Selection and scope

Fetched all origin branches before claiming. The top three unported checker-dependent rules in the combined by-volume ranking were `nexus/correctness-no-implicit-return` (251), `@typescript-eslint/no-deprecated` (238), and `no-else-return` (202). None was already ported or claimed on the fetched branches, so no candidate was skipped. The claim was pushed before implementation. The unit's explicit tsgo-c-library base takes precedence over the generic main-base instruction.

Each rule lives in its own `.a` file. `wave_01_suite.a` loads a checker program once, runs just these three rules, emits the existing canonical finding protocol and releases the program. Existing 26-rule runners and protected compiler files were not edited. The independent Go oracle loads a separate program and invokes the pinned production rules unchanged, with default options. It imports no bridge code.

The three question files on both sides are `function_return_flow`, `symbol_documentation`, and `scope_symbol_declarations`. They expose compiler return-flow/type facts, JSDoc/signature/alias metadata, and symbols with declaration locations. Adamic makes the lint decisions. Documentation transport preserves raw comments; Adamic trims them as the Go rule does. The only pre-existing shared file edit is one line in `bridge/tsgo/checker/facts.go`, delegating unknown questions to the isolated extension registry. Each Go question registers in its own one-line `init`; Adamic requests questions from its own file through the existing generic `Rules.ask` API. No new ABI operation or shared Adamic registration is needed. Two additional `.a` files keep documentation value classes separate to satisfy the source lint gate.

## Toolchain

`bash cloud/setup.sh > /tmp/wave-01-setup.log 2>&1` passed; subsequent shells sourced `/workspace/adamic-tools/env.sh`.

```text
Go 1.27.1: ready 0s
clang 20.1.8: ready 0s
Node v24.19.0: ready 0s
submodules: ready 0s
build cache warm: 82s
setup done: 82s
nproc: 5; cgroup cpu.max: 400000 100000; memory: 17.6 GB
```

## Byte agreement and sanitizers

TypeScript is v6.0.3, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. Both corpora use exactly the manifests already recorded in validation-coverage: all 77 compiler roots and the frozen 287 repository roots. The repository manifest is not enlarged with this unit's new files. The generated controls separately exercise the new implementation.

| Corpus/config | Findings | Fixes | Suggestions | Canonical bytes |
| --- | ---: | ---: | ---: | ---: |
| Compiler | 691 | 186 | 0 | 276439 |
| Frozen repository | 0 | 0 | 0 | 18485 |
| 28 generated controls, strict | 49 | 8 | 0 | 13825 |
| Controls, strict false | 51 | compared | compared | 14707 |
| Controls, noImplicitReturns true | 34 | compared | compared | 7217 |

Compiler counts are exactly 251 implicit-return, 238 deprecated, and 202 else-return findings. Strict controls have 15, 21, and 13 findings respectively. Byte comparison includes rule/message IDs, descriptions, UTF-8 ranges, duplicate findings, all replacement text and ranges, suggestion messages and every suggestion fix. Sorting uses the complete canonical diagnostic. Repository bytes include all file headers; zero findings is not treated as positive rule evidence.

Normal and ASan/UBSan/LSan binaries agree on strict controls and both corpora. All three new checker questions reject a released handle with exit 70 and exactly `adamic: panic: invalid or released checker handle`. The registry-retention mutant succeeds with exit zero instead, proving the stale-handle checks discriminate. Tests save stdout/stderr independently and never pipe test output.

Controls cover opposing reachability, exhaustive switches, never-returning calls, annotations including any/void/undefined/never, async/generators, bare and nested returns, methods/getters, deprecated overloads and constructors, namespace aliases, properties/literal and computed element access, binding and shorthand uses, private identifiers, Unicode/CRLF and NEL/BOM deprecation reasons, else-if chains, fix scope collisions, nested declarations, and ASI-sensitive edit declines. Positive deprecated controls explicitly require Alias, private, method and property findings. External `.a` module resolution and JSX deprecation are not established by these controls.

## Mutants that were run

| Mutation | Required discriminator and observed result |
| --- | --- |
| Implicit-return: invert `facts.implicit` | Exit 0, empty stderr; canonical comparison differs at byte 1685 |
| Deprecated: change reason-bearing message ID to `deprecated` | Exit 0, empty stderr; comparison differs at byte 98 |
| Else-return: extend repair end by one byte | Exit 0, empty stderr; comparison differs at byte 6743 |
| Retain released registry entry | Each new question exits 0; fails the required panic-70 check |
| C input length off by one | ASan heap-buffer-overflow |
| C output string length off by one | ASan heap-buffer-overflow |
| Retain released handle in existing ABI tests | Stale-handle assertion fails |
| Query source-file type instead of position type | Independent Go answer comparison differs at byte 6 |
| Remove link opt-in guard | Unlinked-call refusal test fails |
| Remove C output free | LSan detects leaked buffers |
| Allocate region result on heap | LSan detects unowned result |
| Plant explicit `any` in source-lint control | Configured no-explicit-any reports one finding |
| Existing Node oracle one-byte mutation | TestTheOracleCatchesOneByte passes by detecting it |

The three rule mutants preserve successful execution; only the independent diagnostic comparison catches their changed behavior. Shared-file mutations are Go overlays in scratch storage, not repository edits. Bridge tests also prove outputs survive program release, zero/stale handles are rejected, subsequent handles differ, and 1,600 queries over four compiler files match under sanitizers.

## Commands and logs

```sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE01_ARTIFACTS=/workspace/wave-01-complete \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-01-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave01AgreementAndMutants$' \
  -count=1 -timeout 30m -v > /tmp/wave-01-complete.log 2>&1

go test ./bridge/tsgo/checker -count=1 -timeout 30m -v \
  > /tmp/wave-01-checker-final.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-01-typescript \
go test ./bridge/tsgo -count=1 -timeout 30m -v \
  > /tmp/wave-01-bridge-final.log 2>&1
go vet ./bridge/tsgo/checker ./bridge/tsgo/archive ./stage1/cohere/typeaware \
  > /tmp/wave-01-vet.log 2>&1

go test ./internal/oracle \
  -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures)\.a$' \
  -count=1 -timeout 30m -v > /tmp/wave-01-node-oracle.log 2>&1
```

The filtered Node oracle actually selected eight fixtures, including method_closures and generic_functions, plus the one-byte control. The full `go test ./...` and every older typeaware comparison suite were not run; the targeted new suite, all checker tests, all bridge tests and the filtered external oracle were run.

The pinned cohere CLI's `--no-cache --no-fix --lint` rejects physical `.a` paths as non-TypeScript inputs. This refusal is saved and is not counted as a passing CLI gate. `testdata/source_gate_wave_01.go`, built as an overlay inside cohere, instead applies the unchanged configured production rules using virtual `.a.ts` names and virtual import remapping. It writes no `.ts` files. The diagnostic oracle and native bridge do not use that remapping. All nine new `.a` sources produce zero lint findings; a planted `any` produces one. `testdata/format_wave_01.go` uses the pinned native formatter with virtual parser names and writes only physical `.a` files; a second formatting pass changes nothing.

To reproduce those auxiliary gates, make an overlay mapping the virtual `cohere/adamic_wave01_source_gate.go` to `testdata/source_gate_wave_01.go`, build it from the cohere module, and pass the repository tsconfig, `validation-wave-01/source.manifest`, and CohereSettings.json. The formatter uses the same overlay method with format_wave_01.go and accepts source paths. These are explicit workarounds for the pinned CLI limitation, not claims that its original discovery path passed.

Committed evidence is in [validation-wave-01](validation-wave-01): full gate summaries, setup, CLI refusal, source lint/positive outputs, corpus manifests, gzip finding streams, SHA-256 hashes and all twelve timing records. The complete build/run artifacts remain at `/workspace/wave-01-complete`; raw timing outputs at `/workspace/wave-01-timings`.

## Native time against Go

Three interleaved normal-binary rounds were run without competing builds/tests, with full output comparison in every round. Timer definitions are the existing bridge/oracle definitions: load initializes the checker; run covers native parse/lint/render or Go walk/lint/render; native query time includes checker work and adapter overhead. These phase timers are not isolated C-crossing costs. Whole-process wall time is measured externally.

| Corpus | Go median run | Native median run | Go median wall | Native median wall |
| --- | ---: | ---: | ---: | ---: |
| Compiler | 2.539s | 6.950s | 2.813s | 7.201s |
| Repository | 0.133s | 0.542s | 0.200s | 0.609s |

Native run is 2.74 times Go on the compiler and 4.08 times Go on the repository. Whole-process ratios are 2.56 and 3.05. Native issues 367,142 compiler queries and 39,200 repository queries. This is an agreement milestone, with an observed performance regression relative to Go, not a speed win. No profiling experiment establishes a specific cause.

The measurement procedure is saved in `testdata/bench_wave_01.py`:

```sh
python3 stage1/cohere/typeaware/testdata/bench_wave_01.py \
  --artifacts /workspace/wave-01-complete --typescript /workspace/wave-01-typescript \
  --repository /workspace/adamic --output /workspace/wave-01-timings \
  > /tmp/wave-01-bench.log 2>&1
```

Recorded compiler stream SHA-256 is `442208214231cec9f38b07369740fdf84493e73516918787a70d31fd2a5788a3`; repository is `20dc789d505315aefcd366d819ba20b780f64bcdce0e5fbf2f26dc54ac120a2f`. Absolute file headers make hashes dependent on the recorded root paths.

## Remaining limits

This follows the existing standalone stage-1 diagnostic harness pattern. It does not add these rules to the 26-rule runner or a complete CLI; it does not apply edits or implement suppression/config routing. no-deprecated's nonempty allow list and no-else-return's nondefault allowElseIf option are not implemented. Default options are held to the unchanged Go oracle on the named corpora and controls. Zero suggestions on these workloads does not prove every possible suggestion branch. The frozen repository corpus omits new wave sources; they have separate source-lint and native-build checks. No general equivalence proof outside the measured corpora is claimed.
