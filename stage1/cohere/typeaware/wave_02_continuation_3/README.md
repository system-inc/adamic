Built: native prefer-regex-literals and prefer-rest-params; supplemental React exhaustive-deps draft with explicit JSX refusal.
Commits: claim 882381a7, corrected claim 7e909f47, implementation ff7f6c9e; all pushed on codex/typeaware-wave-02.
Checks: 481 supported controls, 433 findings, 77 compiler roots and 287 repository roots match Go bytes; sanitizers and released handles pass.
Mutants: all three rule mutants caught by byte comparison; JSX refusal and released registry mutants caught; existing bridge mutants pass.
Limits: core prefer-promise-reject-errors skipped because claimed elsewhere; 14 JSX controls blocked; defaults only; emitted-JavaScript comparison not run.

The initial selection used substring matching and incorrectly counted the namespaced @typescript-eslint/prefer-promise-reject-errors claim as the core rule. Exact-token reconstruction of the original 355 origin references identifies the correct first three as core prefer-promise-reject-errors, prefer-regex-literals and prefer-rest-params. The correction was pushed before any core promise implementation. A refreshed fetch found exact core reservations on waves 07, 18, 20, 24 and 26. That rule was skipped without implementation. The React reservation was released; its implementation remains supplemental work. selection-audit.json records the original tips, exact-name analysis and correction. The branch's previous nine claimed ports were already tested and pushed at 0a3a176e before this continuation.

All implementation files are .a. No shared harness, registration generator or protected compiler file was edited. These rules use existing checker questions; no new bridge registration was needed. verify.py is a standalone verifier inside this rule directory, and source_gate.py checks a temporary source snapshot because the pinned cohere CLI does not enumerate .a files.

Validation observations

The independent Go executable calls production cohere rules and serializes complete canonical findings, fixes and suggestions. Native and sanitized native output matched it byte for byte:

| Population | Roots | Identical output bytes | Findings |
| --- | ---: | ---: | ---: |
| Supported controls | 481 of 495 | 221982 | 433 |
| TypeScript v6.0.3 src/compiler | 77 | 5318 | 0 |
| Frozen repository manifest | 287 | 13319 | 0 |

All 495 controls parse in Go. Native refuses 14 because the shared parser lacks JSX support: some fail parsing directly, others are incorrectly represented as TypeAssertionExpression. The owned suite explicitly refuses the latter instead of silently comparing an incorrect AST. validation/blocked.json lists the inputs and exact errors. Control 467 constructs a JSX value used as a hook dependency; Go reports a construction finding, while disabling the refusal yields a successful native run with no finding. The independent comparison catches that silent AST error. Completing React parity requires shared parser JSX support; this draft does not claim parity on those 14 controls.

| Rule | Supported-control findings | Direct fixes | Suggestions | Suggestion edits |
| --- | ---: | ---: | ---: | ---: |
| prefer-rest-params | 7 | 0 | 0 | 0 |
| prefer-regex-literals | 196 | 0 | 166 | 166 |
| react-hooks/exhaustive-deps, supplemental | 230 | 0 | 138 | 0 |

The React suggestion summaries intentionally contain no edits, matching production Go. Regex validation and spelling are implemented natively, rather than requesting Go lint verdicts.

Mutant evidence

| Mutant | Result and independent check |
| --- | --- |
| Rest: invert symbol declaration-count predicate | Compiles, exits 0; Go byte comparison fails at byte 62 |
| Regex: invert syntax validity | Compiles, exits 0; Go byte comparison fails at byte 15082 |
| React: invert missing-dependency effect predicate | Compiles, exits 0; Go byte comparison fails at byte 155714 |
| JSX refusal: disable TSX guard | Compiles, exits 0; Go construction finding differs on control 467 |
| Released registry: omit live-handle deletion in a build overlay | Compiles, exits 0; required released-handle panic 70 is missing |

The unmutated released-handle probe exits 70 with `adamic: panic: invalid or released checker handle`. ASAN, UBSAN and LSAN runs match all supported outputs with empty stderr. Existing bridge tests also exercise input/output length corruption, stale registry, wrong position, missing linkage, omitted C free and region heap ownership mutants. Only build overlays and scratch copies are mutated.

Commands and outcomes

Run from the repository root after sourcing /workspace/adamic-tools/env.sh; redirect test output to files:

```sh
python3 stage1/cohere/typeaware/wave_02_continuation_3/verify.py --scratch /workspace/wave-02/continuation-3/final4 --stage0 /workspace/wave-02/continuation-2/run/adamic --archive /workspace/wave-02/continuation-2/run/checker.a --sanitized-archive /workspace/wave-02/continuation-2/run/checker-asan.a --typescript-source /workspace/wave-02/typescript > /workspace/wave-02/continuation-3/verification-hardened.log 2>&1
python3 stage1/cohere/typeaware/wave_02_continuation_3/source_gate.py --scratch /workspace/wave-02/continuation-3/source-final --cohere /workspace/wave-02/cohere > /workspace/wave-02/continuation-3/source-final.log 2>&1
go test ./bridge/tsgo/... > /workspace/wave-02/continuation-3/bridge.log 2>&1
go test -v ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(strings.a|maps_and_text.a|reuse_arrays.a|closures.a)$' > /workspace/wave-02/continuation-3/node.log 2>&1
go vet ./... > /workspace/wave-02/continuation-3/vet-all.log 2>&1
```

Verification passes. Source formatting and all 276 lint rules pass on nine owned modules. Bridge tests pass (bridge 80.328s; checker 0.624s). The filtered Node oracle passes in 23.069s, including its one-byte mutant. go vet ./... and git diff --check pass. This is scoped testing, not the full repository test gate. The original setup was reused: Go 0s, clang 0s, Node 0s, submodules 0s, build/cache 76s, total 76s. nproc is 5; cgroup CPU allocation is 4. Toolchain: Go 1.27.1, clang 20.1.8, Node 24.19.0.

Performance observations

Three alternating warmed runs include process startup, checker creation, traversal and output; medians are wall-clock seconds for the complete three-rule suite:

| Population | Native | Go | Native / Go |
| --- | ---: | ---: | ---: |
| Compiler | 2.198067849 | 0.362358381 | 6.07 |
| Repository | 0.249143279 | 0.130059104 | 1.92 |

These observations show native is slower on these populations; no speedup is inferred. Raw timing rounds are in validation/timings.json.

Evidence and remaining limits

validation/streams.json records original paths, sizes and SHA-256 hashes for 99 archived streams. Large streams use deterministic gzip; hashes refer to decompressed bytes. validation/source-hashes.json pins 377 implementation and corpus sources. commands.json records 40 aggregate commands, exits and timings; individual parse-probe errors are preserved in blocked.json. Only the final successful final4 run is archived. Earlier failed/stale runs are not presented as passing evidence.

Nondefault configuration is not implemented (including redundant regex wrapping and additional React hooks or dangerous autofix). JSX-dependent React behavior remains blocked. Emitted-JavaScript comparison is not run; the shared harness worker owns that integration. No further rules were claimed after discovering the reservation conflict.
