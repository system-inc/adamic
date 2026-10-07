Rebased the owned wave-29 history onto area/stage1-lint and re-green all available profiles and kernels.
Source commit 7f09aab3e8541fe4fa252686a0e6bb62d679a345; area d65a8f931; current main 39638d9e2 is an ancestor.
Setup 103s, nproc 5; default oracle PASS 93.864s, next seven profiles and two corpora PASS, focused bridge/runtime gates PASS.
Rule, provenance, released-handle, listener, JSX, analysis-kernel, resolution and regex-contract mutants caught as described in preserved logs.
Full React source analysis, checker-backed source adapters and configured id-match remain blocked; no new claim or full repository gate.

This direct rebase contains the requested 50a5f105 integration and harness 41eb6eab2,
plus newer inherited runtime profiling and string fixes. It had no conflicts.
No shared compiler, harness or registry implementation was edited manually.

Validation commands, all with output redirected to the compressed logs beside this report:

- bash cloud/setup.sh; source /workspace/adamic-tools/env.sh; nproc; go run ./cmd/lint-registry; go build -o /workspace/wave29-area-adamic ./cmd/adamic
- ADAMIC_WAVE29_ARTIFACTS=/workspace/wave29-area-default ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave29-typescript go test ./stage1/cohere/typeaware -run '^TestWave29AgreementAndMutants$' -count=1 -timeout=30m -v
- ADAMIC_COMPILER=/workspace/wave29-area-adamic python3 -u stage1/cohere/typeaware/wave-29-next/check.py /workspace/wave29-area-next
- With that same compiler: wave-29-fourth/check.py, wave-29-third/check_static_core.py, check_control_core.py, check_listeners.py, wave-29-fourth/prove_gaps.py, wave-29-third/prove_gaps.py, wave-29-configured/check_regex_contract.py and wave-29-fourth/rules/react-jsx-no-undef/testdata/check_resolution.py, each with its corresponding /workspace/wave29-area-* artifact directory.
- go test ./bridge/tsgo/checker -count=1 -timeout=30m -v: PASS 0.349s
- go test ./internal/native -run '^(TestRuntimeReleasePaths|TestRuntimeStringEquality)$' -count=1 -timeout=30m -v: PASS 22.356s
- go test ./internal/oracle -run '^TestRuntimeLastIndexOfMatchesNode$' -count=1 -timeout=30m -v: PASS 0.449s

Original profiles: 9840 control bytes/15 findings; compiler 77 files/5241 bytes and
repository 287 files/18485 bytes, zero findings in either corpus. Sanitized streams
match. Mutants denylist, match and concurrency plus provenance compile, exit zero,
and fail only byte comparison; released-handle mutant is caught by required panic.
Original native/Go timing: compiler 2.108307973s/0.334159075s (6.31x), repository
0.306956230s/0.145938141s (2.10x).

Next profiles: 400 cases/252 findings/seven options groups. All three compiling
mutants (globals, setter, shadow) caught; corpus and sanitizer streams match.
Native/Go compiler 2.276489924s/0.370732292s (6.14x); repository
0.425217063s/0.207029410s (2.05x). These concurrent process timings are observations,
not isolated benchmarks, and still use the legacy scan driver.

Partial kernels: JSX 79 rows/7359 bytes; static 34 supplied graphs/27 findings;
control 17796 graphs/28 sources; nine named listener declarations/562 bytes.
Their existing compiling mutants all caught. Resolution 148 source/option/checker
controls/28 findings/18408 bytes agrees with Go, native, source Node, emitted JS,
ASan/UBSan/leaks; all seven mutants (globals, commonjs, declaration, checker,
resolved, message, positions) caught. Raw facts remain a test adapter, not a
production checker-backed source registration. The regex constant-pattern mutant
is caught by Node bytes; nonconstant native RegExp refusal and five Go/JS dialect
witnesses remain explicit blockers. React source probes now compile JSX but the
source-to-SSA/analysis integration remains parked.

An initial default run failed with ENOSPC, not an oracle mismatch. Automatic review
rejected a broad size-based cleanup; a narrower cleanup verified ELF/archive
signatures in three superseded owned build directories, preserving sources/logs.
The clean rerun above passed. No new rules are claimed while remaining existing
checker/source and regex gaps prevent complete ports.
