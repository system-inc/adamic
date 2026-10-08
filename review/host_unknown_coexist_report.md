Built: merged codex/unknown-narrowing and made Array.isArray an additive observation of tagged unknown values.
Commit: this merge commit on codex/host-blockers; parents a85a9cb1 and 616870dd. No main merge.
Checks: Node agrees with native sanitized/release and JavaScript for errorCode, existing host fixtures and array predicates; library's exact host 05/13 also pass on its combined base.
Mutants: seven local source mutants caught and restored; the imported unknown presence/type guard mutants also pass their detection tests.
Limits: no full repository gate; erased unknown array elements and nullable reference erasure remain named refusals.

The prior opaque array-only representation shared null and undefined and prohibited all other unknown observations. The merged unknown unit supplies tagged values, a distinct null sentinel, property presence and dynamic field reads. Array.isArray now uses that representation and the checker-proven declared array member; the unknown user predicate signature remains `value: unknown`, returning `value is readonly unknown[]`. General `typeof`, non-null guards and `'code' in` coexist with it. `any` remains refused.

The new `internal/oracle/testdata/host_unknown_error_code.a` exercises the exact guarded errorCode shape, missing/numeric/undefined code, unknown returns and calls with effects, null/undefined, array wrapper and direct brand tests, null truthiness and nullish fallback. The distinct sentinel must be recognized by typeof, both native truthiness paths and coalescing. Reading erased unknown array elements is still refused because their raw slot representation is not recoverable; nullable reference-to-unknown erasure is refused rather than conflating null with undefined.

Merge compatibility: qualified NodeJS assertions now reach the existing optional-field proof instead of panicking. Two older incoming positive tests are changed to expect the established optional-widening refusal for Error-to-ErrnoException. Four incoming fs fixtures use guarded unknown errorCode instead of that unproven cast, and bind plain options objects where inline literal splitting would refuse evaluated fields. Their filesystem behavior remains held to Node. The fs leak harness uses the current callback API. Historical stage3 NOTICE/status snapshots retain this branch's version, without claiming new snapshot results.

Commands and observations (all output redirected to logs):

- `export GOPROXY='https://proxy.golang.org|direct'; bash cloud/setup.sh > /tmp/host-unknown-coexist-setup.log 2>&1`; source `/workspace/adamic-tools/env.sh`. Timing lines: Node 0.023s, Go 0.024s, submodules 0.066s, markdown dependencies 0.078s, clang 0.179s, build 26.766s, deferred test binaries 26.920s, cache warm 26.922s, done 26.952s. `nproc=5`, cpu.max `400000 100000`.
- `go test ./internal/lower -count=1`: PASS, 17.180s, `/tmp/host-unknown-coexist-lower-final.log`.
- `go test ./internal/load ./internal/fresh ./internal/ir ./internal/javascript ./internal/flow -count=1`: load PASS 9.646s, fresh PASS 130.342s, IR PASS 59.038s, JavaScript has no package tests. The initial flow run caught the four old fs fixtures described above; final `go test ./internal/flow -count=1` PASS 77.567s, `/tmp/host-unknown-coexist-flow-final.log`.
- `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestUnknownNarrowingMutants|TestNativeAgreesWithNode/internal/oracle/testdata/(host_.*|unknown_narrowing.*)|TestNodeFSFileAgreesWithNode' -count=1`: PASS 7.318s, `/tmp/host-unknown-coexist-final-oracle.log`. Native sanitized and release plus JavaScript all held to Node; imported presence and inner-typeof mutants detected.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`: PASS 21.775s, `/tmp/host-unknown-coexist-counts-final.log`. Initial run caught the incompatible fs fixtures; final measured table regenerated after adapting them.
- `go test ./internal/native -run 'TestRuntime|TestCEndsInNewline|TestMaybeNumbers|TestNodeBufferRuntimeWithoutDeclarations' -count=1`: PASS 21.785s, `/tmp/host-unknown-coexist-native.log`.
- `go vet ./...`: PASS, `/tmp/host-unknown-coexist-vet.log`.
- Isolated detached worktree `/workspace/host-unknown-combined-review` at library combined proof `57fd72ae`, with the additive array/sentinel fixes applied and three review fixtures registered: `ADAMIC_GATE_UNCACHED=1 go test -v -buildvcs=false ./internal/oracle -run 'TestNativeAgreesWithNode/(stage3|internal)/(fixtures|oracle)/(host|testdata)/(05_writeFile|13_createDirectory|host_unknown_error_code|host_array.*)' -count=1`: PASS 21.907s, all five tests explicitly ran. `/tmp/host-unknown-combined-oracle-final.log`. This worktree uses its own pinned cohere revision; nothing from it is merged into host-blockers.

Source mutants, each actually run, failed as intended, and restored (`/tmp/host-unknown-coexist-mutants.log`, per-mutant logs with the same prefix):

| Mutant | Check that caught it |
|---|---|
| Restore a85's unknown observation restriction to Array.isArray only | New errorCode oracle refuses at 3:16 with the reported regression diagnostic |
| Remove null sentinel handling from native typeof | New oracle trips ASan global-buffer-overflow |
| Remove null sentinel handling from census truthiness | Node comparison catches `!!null` becoming true |
| Remove null sentinel handling from branch truthiness | Node comparison catches `if (null)` taking the true branch |
| Remove null sentinel handling from union coalescing | Node comparison catches null failing to take its fallback |
| Admit nullable reference erasure into unknown | `TestUnknownReflectionRefusals/nullable_reference` catches absent named refusal |
| Admit erased unknown array element reads | `TestUnknownArrayPredicateRefusesUnrepresentedObservations` catches absent named refusal |

No full gate or all 25 host fixtures claimed. Actual typed arrays are outside this branch's constructor support; the previous separate typed-array review remains the evidence for their false brand test. This task rechecks this branch's array/scalar/object/null/undefined predicate fixtures and the two reported host regressions.

The new errorCode count row is 33 allocations, 33 frees, 40 retains, 70 releases, peak 4, no region frees. Formatting is clean. The staged merge whitespace check reports only two pre-existing incoming lines: a trailing space in node_fs_file_buffer.a and intentional whitespace in the base64 input data; neither is introduced by this fix.
