Rebased all nine completed wave 20 ports and twelve listener declarations onto current origin/main f8013f0b.
Validated source SHA: 9c1867f971acdd7c4d253a0de7990178470fe8fb; evidence commit follows it on the same branch.
Commands: all three full rule gates PASS; checker PASS 0.663s; Node oracle PASS 27.524s; vet empty; listeners and manifests PASS.
Mutants: nine rule, six raw-fact, five registry, twelve compiling declaration and twelve JSON metadata mutants caught.
Not covered: three React source analyses, shared numeric node handoff, new Diagnostic integration, full repository gate, own-rule emitted JavaScript comparison.

The all-heads fetch found main advanced from e8ba3d5d to
f8013f0baac41ddc340d76f83bddde38536a8f07. Explicit user landing instructions
authorize rebasing this own branch, overriding CLAUDE's ordinary no-rewrite
rule. All 17 commits replayed without conflicts. No source implementation
changed beyond the rebase. A final remote check still found the same main.
Only codex/typeaware-wave-20 is published, with an exact lease against its
previous 7cd43425c378388e3db3613343ff9425b2dd7541 tip. Never main or area/.

The complete byte-comparison gates cover first-batch 48 control findings,
Nexus 121 projects and 160 findings, core 682 projects and 502 findings with
365 suggestions. Messages, spans, fixes, suggestions and edit bytes agree
with unchanged production Go in ordinary and ASan/UBSan/LSan runs. Each gate
also matches the frozen repository 287 roots and compiler 77 roots:
18,485 and 5,241 bytes, zero findings, with positive control suites above.
The known invalid upstream TSX exclusion is unchanged and remains documented
in REPORT.md. First gate completed in 246.076s while gates ran concurrently.

All nine semantic mutants compile, exit 0, emit empty stderr and differ only
under the independent finding comparison, including the regex suggestion-span
mutation. Six raw checker-fact mutations are caught. Released checker queries
panic 70; five retained-registry mutations exit normally and fail that expected
panic. Numeric declarations match 136 production-registration bytes under
normal and sanitizer execution; twelve compiling kind mutations are caught.
Twelve JSON first-kind increments are rejected independently; these metadata
mutants do not establish correctness of the three absent React analyses.

Commands, with source /workspace/adamic-tools/env.sh first:

```sh
ADAMIC_WAVE20_ARTIFACTS=/workspace/wave20-validation/continue-first \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave20-typescript \
go test ./stage1/cohere/typeaware -run '^TestWave20AgreementAndMutants$' -count=1 -v -timeout 30m > /workspace/wave20-validation/continue-first.log 2>&1
python3 stage1/cohere/typeaware/no-process-exit-after-output/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/continue-second --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/next/restored > /workspace/wave20-validation/continue-second.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/continue-third --compiler /workspace/wave20-typescript --cases /workspace/wave20-validation/third/valid-cases > /workspace/wave20-validation/continue-third.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/continue-listeners --compiler /workspace/wave20-validation/continue-third/adamic --checker /workspace/wave20-validation/continue-third/checker.a > /workspace/wave20-validation/continue-listeners.log 2>&1
python3 stage1/cohere/typeaware/prefer-promise-reject-errors/manifest_verify.py --oracle /workspace/wave20-validation/continue-listeners/oracle --artifacts /workspace/wave20-validation/continue-manifests > /workspace/wave20-validation/continue-manifests.log 2>&1
go test ./bridge/tsgo/checker -count=1 > /workspace/wave20-validation/continue-checker-correct.log 2>&1
go test ./internal/oracle -run 'TestTheOracleCatchesOneByte|TestNativeAgreesWithNode/internal/oracle/testdata/(strings|closures|generic_functions)[.]a$' -count=1 -v -timeout 5m > /workspace/wave20-validation/continue-node.log 2>&1
go vet ./... > /workspace/wave20-validation/continue-vet.log 2>&1
```

An initial checker command incorrectly named nonexistent ./internal/checker
and failed before running tests; corrected ./bridge/tsgo/checker passed.
Both outputs are preserved. Toolchain reused from prior documented setup:
Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc rechecked 5, quota 4 cores.
No new setup or full repository gate was run. Logs and command outputs are
archived under validation/landing-f8013f0b, excluding executables.

Observed whole-process seconds (native / Go): first repository .662655 /
.355306, compiler 5.499247 / 1.156424; Nexus repository .468726 / .267182,
compiler 2.223828 / .515211; core repository .365569 / .216161, compiler
2.171961 / .468023. Gates overlapped, so these observations are not isolated
performance benchmarks. Native remains slower; no dispatch speedup claimed.

Main still has string-only ParseNode.kind and no integrated JSX parser or
native React source-to-HIR/SSA/capture pipeline. Wave 29 now additionally has
tested supplied-graph post-dominance/control kernels, but its CONTROL_REPORT
explicitly leaves native source-to-SSA and full React comparisons unfinished.
Duplicating its partial kernels would not close those shared gaps. The three
React claims overlap waves 06 and 29 and remain unfinished; no replacement
claims were taken. Shared parser/driver/harness modifications remain forbidden.
No batch 8 Diagnostic landing SHA was named, and none is inferred from branch
existence. Full per-node dispatch conversion remains blocked on that interface.
