Built: rebased wave-09 onto lint area b84a9d931, including current main c7991b900, and re-greened its owned oracles.
Commits: old remote a109073bf; rebased implementation 85cd25b73; evidence commit follows on wave-09 only.
Commands/output: original rules with both frozen corpora, nineteen owned verifiers, bridge/checker, registry, filtered Node, required parser checks and vet PASS.
Mutants: all existing semantic, handed-node, handle and sanitizer checks reran and detected their deliberate changes.
Not covered: full regex rule parity, shared checker installation, unrelated required checks or the full repository gate; no new claims.

The area had advanced to b84a9d9314b65d3d0261ee017e233287b4f071da,
merging current main c7991b900362796aefd111474e65eb5398e91953.
All 34 owned commits rebased without conflicts. Shared changes were retained;
no shared compiler, parser, registry, harness or generator files were edited.
The native compiler was rebuilt before verification. Publication is solely
to codex/typeaware-wave-09 with an exact lease on old remote
a109073bf1148a85b6d577476825f4033e643609. No main or area push.

After sourcing /workspace/adamic-tools/env.sh, commands were:

```
go build -o /workspace/wave-09-core/adamic ./cmd/adamic
ADAMIC_TYPESCRIPT_SOURCE=/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8 ADAMIC_WAVE09_REPOSITORY_MANIFEST=/workspace/wave-09-validation/repository.manifest ADAMIC_WAVE09_COMPILER_MANIFEST=/workspace/wave-09-validation/compiler.manifest ADAMIC_WAVE09_ARTIFACTS=/workspace/wave-09-fresh-original-artifacts go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v
go test ./bridge/tsgo/checker ./bridge/tsgo -count=1 -timeout=15m -v
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-09-required-typescript go test ./stage1/typescript/parser -run '^TestCompilerExpressionsAgree$|^TestWholeCompilerAgrees$' -count=1 -timeout=30m -v
go test ./stage1/cohere/lint/registry -count=1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|inherited_static_field_read|runtime_last_index_of)\.a$' -count=1 -timeout=10m -v
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware
```

Every command wrote a separate log, never a pipe. All nineteen Python
verifiers enumerated in validation-landing-c799/checks.json ran sequentially
on the rebuilt compiler, including both handed-node contract checks. The
first original-suite invocation omitted its two manifest variables; its log
explicitly reported unconfigured corpora. That invocation is retained but
does not establish corpus coverage. The corrected invocation above ran both
frozen manifests and passed in 102.390 seconds. It matches 44 control, 14
compiler and 4 repository findings, fixes and suggestions in normal and
sanitized native executions.

Original semantic mutants compile/run cleanly and differ from Go at byte
63 (prefer-const), 10007 (radix), 12900 (type-parameter flag) and 12301
(non-nullable assertion fix). Released-registry retention removes required
panic 70 and is caught. Bridge ownership/memory/length mutants are caught
by its sanitizer/refusal checks; bridge PASS 165.972s, checker PASS 0.376s.
Filtered Node PASS 18.341s, registry PASS 0.038s; vet output is empty.

Label controls retain 18 findings/7700 bytes; compiler 77 sources and
repository 287 sources retain zero findings/7859 and 18485 bytes. The one
parser-refused label fixture remains explicitly recorded. Scope membership,
value-mask, handle retention and handed-node refetch checks pass. The flags
frontend retains 40 controls, 35 findings and 6020 bytes, normal/sanitized;
handed report-refetch differs at byte 58. Named descriptor mutations differ
at byte 276. Literal-slice findings/suggestions retain all four execution
modes and corpora. All existing helper mutations reran: flag precedence,
sequence bounds, surrogate folding, malformed-pattern guard, rune quoting,
error trimming, suggestion range, cooked/raw offsets, constructor span,
constant and alias eligibility, Unicode folding/equivalence/class escapes.
Detailed differences are in the nineteen fresh verifier logs. Historical
custom regex helpers remain isolated evidence, not full rule completion
under the no-hand-rolled-matcher requirement. No new matcher was added.

Required parser checks use the real Git checkout at
050880ce59e30b356b686bd3144efe24f875ebc8, not the tar extraction.
They pass on 77 files with 28836875 expression-tree bytes and 44766682
whole-AST bytes identical across Go, source Node and sanitized native.
Package PASS 40.635s. No inputs or correctness checks were weakened or
skipped. Other required stage1 packages and the full gate were not run.

Original whole-process native/Go timings: compiler 2.383843/0.605300s;
repository 0.366465/0.196368s. Concurrent validation is not isolated
throughput evidence. Toolchain setup is reused from its successful 88-second
run; nproc 5. Larger helper compilation took 90 seconds but still passed.

Dynamic new RegExp(pattern, 'u') remains explicitly refused on this base.
Source Node passes; the static-literal mutant compiles and runs under
sanitizers, proving the required-refusal check can fail. The dependency
origin/codex/regex-runtime-compiler remains at a69dcd648. Both regex claims
remain incomplete and no React parking exception applies. No new claims.
Existing shared checker installation and remaining visitor migration are
unfinished integration work. Fresh logs and the complete check list are
retained in validation-landing-c799.
