Built: rebased wave-09 onto lint area d3a37422c, including main b6b1538b0 and its typeof/null fixes, then re-greened all owned checks.
Commits: old remote 88212dee8; rebased implementation d1226b259; evidence commit follows on wave-09 only.
Commands/output: original rules with both corpora, twenty owned verifiers, bridge/checker, pinned parser comparisons, registry, filtered Node and vet PASS.
Mutants: every existing semantic, options, handed-node, handle and sanitizer check reran and caught its mutation.
Not covered: full regex rule parity, shared checker/JSON installation, unrelated required checks or the full gate; no new claims.

Area d3a37422c6c2c3dd4a90b8721a2067a4ba0d8898 merges main
b6b1538b0cebc4ba6741ac34f1aedb60293c1d06. All 37 owned commits
rebased without conflicts and the compiler was rebuilt. Shared typeof/null
fixes were retained unchanged. No shared source, registry or harness edits.
Push uses an exact lease on old remote
88212dee80f9c8bf0eed3435603916908eee7705 and exclusively targets
codex/typeaware-wave-09; no main or area push.

After sourcing /workspace/adamic-tools/env.sh, test output went to separate
files under /workspace/wave-09-typeof-*.log, never a pipe. Commands:

```
go build -o /workspace/wave-09-core/adamic ./cmd/adamic
ADAMIC_TYPESCRIPT_SOURCE=/workspace/TypeScript-050880ce59e30b356b686bd3144efe24f875ebc8 ADAMIC_WAVE09_REPOSITORY_MANIFEST=/workspace/wave-09-validation/repository.manifest ADAMIC_WAVE09_COMPILER_MANIFEST=/workspace/wave-09-validation/compiler.manifest ADAMIC_WAVE09_ARTIFACTS=/workspace/wave-09-typeof-original-artifacts go test ./stage1/cohere/typeaware -run '^TestWave09' -count=1 -timeout=30m -v
go test ./bridge/tsgo/checker ./bridge/tsgo -count=1 -timeout=15m -v
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-09-required-typescript go test ./stage1/typescript/parser -run '^TestCompilerExpressionsAgree$|^TestWholeCompilerAgrees$' -count=1 -timeout=30m -v
go test ./stage1/cohere/lint/registry -count=1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$|^TestTypeof|^TestNativeAgreesWithNode$/internal/oracle/testdata/(maps_and_text|sorting|string_index|lone_surrogates|functions|closures|inherited_static_field_read|runtime_last_index_of|typeof_.*)\.a$' -count=1 -timeout=15m -v
go vet ./bridge/tsgo/checker ./stage1/cohere/typeaware ./stage1/cohere/lint/registry
```

The twenty owned Python verifiers are enumerated with their log paths in
validation-landing-typeof/checks.json. All ran sequentially on the rebuilt
compiler; constructor flag options are newly included alongside the preceding
nineteen. Original suite PASS 136.747s, checker 0.156s, bridge 98.953s,
required parser 43.180s, filtered Node 6.797s, registry 0.084s; vet empty.
The filtered Node run covers typeof fixtures; the separate TestTypeOf* mutant
tests are not claimed as run by the differently capitalized filter above.

Original findings/fixes/suggestions match production Go: 44 controls,
14 compiler and 4 repository findings, normal and sanitized. Frozen corpus
manifests remain 77 compiler/287 repository sources. The original clean
semantic mutants are caught at byte 64 (prefer-const), 10017 (radix),
12914 (type-parameter flag) and 12314 (non-nullable fix). Required panic 70
catches released-registry retention. Bridge ownership/length/memory mutations
are caught by its sanitizer/refusal checks. Label controls retain 18/7700
bytes, both corpora zero/7859 and 18485 bytes; its one parser-refused fixture
remains explicit. Scope membership/value-mask, handle retention and handed
node-refetch checks pass. Flag controls retain 35/6020 bytes, option profiles
13/2216, 11/1998 and 13/2218 bytes. Ignored options differ at byte 49 and
report-refetch at byte 58. Duplicate options fail for the named uniqueness
reason in Go, normal and sanitized native.

Literal findings/suggestions and both corpora retain Go/native/sanitized/
source-Node/emitted-JS agreement. Every existing helper mutation reran:
flag precedence, sequence bounds, surrogate folding, malformed-pattern guard,
rune quoting, error trimming, suggestion range, cooked/raw offsets, constructor
span, constants, scoped writes/aliases, named listener kinds and Unicode
folding/equivalence/class escapes. Complete fresh logs retain every catcher.
Historical custom regex helpers are isolated evidence, not full completion
under the no-hand-rolled-matcher rule. No new matcher or parser was added.

The required parser checks use a real Git checkout pinned to
050880ce59e30b356b686bd3144efe24f875ebc8. They pass on all 77 files,
28836875 expression-tree and 44766682 whole-AST bytes identical across
Go/source Node/sanitized native. No input pin, check or options guard was
weakened or skipped. Other required packages and the full gate were not run.

Original native/Go whole-process timing: compiler 3.048214/0.754975s;
repository 0.349310/0.185527s. Joined flag-options profile 0.064263/0.064079s.
These concurrent single runs are not isolated throughput evidence. Successful
88-second setup was reused; nproc 5.

Dynamic new RegExp(pattern, 'u') remains explicitly refused on the rebuilt
compiler; Node passes and the static-pattern mutant builds/runs sanitized,
proving the refusal assertion can fail. Runtime dependency remains ca71f1deb.
Both regex claims remain incomplete, with shared checker/JSON installation
and remaining visitor migration also unfinished. No React parking or new
claims. Fresh logs and measurements are retained in validation-landing-typeof.
