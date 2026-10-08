Built: native .a hook rules react-hooks/unsupported-syntax and react-hooks/use-memo; partial react/boolean-prop-naming, explicitly incomplete.
Commits: claim 1de5a384 pushed before source; implementation 34369377 and exact message interpolation 4f9f26dc; evidence accompanies this report.
Commands and outputs: differential PASS 115.163s; checker PASS 0.145s; four Node fixtures PASS 6.137s and byte mutant PASS 3.446s; full vet and final package vet clean; setup 19s, nproc 5.
Mutants: unsupported-syntax and use-memo span shifts plus disabled boolean-name acceptance caught only by Go byte comparison; both retained-registry mutants caught by released-handle panic; Node byte mutant caught.
Not covered: production JSX, configurable regex matching and complete BooleanPropNaming component/props integration, plus the full shared gate. No further rules were claimed.

The fourth batch was already pushed at ca3ed859. A fresh fetch inspected 389
origin refs, 33 unique Markdown claim blobs and stage1 sources on origin/main
and origin/codex/tsgo-c-library. The first eligible combined ranking entries
were ranks 171, 172 and 173, all zero-volume. Earlier entries were skipped as
ported or claimed. The complete audit is validation-wave-17-fifth/selection.json.
Claim 1de5a384 was pushed before implementation. This batch changes no shared
harness, registration generator, checker or compiler file and adds no bridge
question. Existing raw declaration, default-library and symbol facts suffice.

The hook rules reproduce Go's three unsupported-syntax messages and seven
use-memo messages. Their native gate follows ASCII names, JSX/hook-call evidence,
parameter shape, nested-function barriers and reachable-root positions. Memo
calls preserve named-import versus receiver asymmetry, local shadowing,
dependency-list classification and resolved outer-binding writes. The ordinary
native parser exercises 44 non-JSX controls with 24 findings and 12557 identical
bytes, under ordinary and ASan/UBSan builds.

For JSX decisions, the validation-only projected_react_tree.a reads raw AST
records generated independently by typescript-go. The records contain kinds,
text, positions, children and argument counts, never lint decisions. All gate,
resolution and lint decisions execute in native Adamic, against the same actual
program's checker facts. This is decision validation and does not claim native
JSX parsing. The 155 isolated upstream and edge controls match at 49120 bytes:
34 unsupported-syntax findings, 61 use-memo findings and four findings from the
implemented PropTypes portion of BooleanPropNaming. Every hook message has a
positive control. Fix and suggestion arrays are empty, exactly as Go reports.
Default and sanitizer streams match completely.

The BooleanPropNaming file ports PropTypes assignment, static and object forms,
quoted/computed keys, one-level isRequired handling, configurable boolean member
names, nested traversal, deduplication, exact message rendering and a type-member
walker. Its matcher is injected; the test suite supplies only Go's documented
default pattern. Custom message interpolation uses Go's exact ASCII placeholder
and whitespace grammar, preserving unknown placeholders, nested delimiters and
non-ASCII whitespace. Options match Go under sanitizers at 49249 bytes (nested),
49203 (message probe), and 49352 (both). The class is a partial implementation:
its type-member walker is not wired to a complete React component detector and
is not validated as a complete TypeScript props-analysis rule. Custom pattern
compilation and the Go config decoder contract are not implemented.

The precise native regex blocker is preserved in validation-wave-17-fifth/
regexp-gap.a and regexp-gap.log. Building the valid dynamic RegExp probe exits 1:

```
adamic: /workspace/wave-17-fifth-regexp-gap.a:3:13: stage 0 can't lower new an Identifier yet
```

Thus the existing native RegExp surface cannot supply the configured pattern
matcher. A complete port still needs a native matcher with Go regexp semantics;
a default-pattern function cannot substitute for it. This work stops with that
rule incomplete and does not claim another batch.

The native JSX boundary is separately observed, not inferred. Go reports two
findings for the raw jsx-boundary.tsx.txt archive (originally loaded through a .tsx fixture alias). The sanitizer
production native suite exits 70 with:

```
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 52 in /workspace/wave-17-fifth/jsx-boundary.tsx
```

The source and both complete outputs are preserved. The shared parser is
untouched. Its missing JSX support blocks production input coverage of the
JSX controls for both hook rules and the JSX component half of prop naming.

The frozen 77-root TypeScript compiler and 287-root repository corpora match Go
at 5318 and 18485 bytes respectively, with zero findings, in normal and sanitizer
builds using the ordinary parser. BooleanPropNaming has no options in those
corpus runs and is inert on both sides, matching Go's Run contract; the zero
corpus result does not validate the unfinished configured rule.

All three rule mutants exit zero with empty sanitizer stderr. The independent
Go comparator catches the unsupported-syntax span shift at byte 47, use-memo
span shift at byte 9075 and disabled boolean-name acceptance at byte 1187. The
last proves the implemented default-pattern PropTypes check, not the missing
configurable matcher. Existing binding and symbol-identity questions reject
released programs with panic 70. Registry-retention mutants finish cleanly and
fail the required-panic assertion. The external Node byte mutant also fails its
comparator as required.

Quiet three-round alternating whole-process medians, with identical byte hashes
on every round:

| Corpus | Native | Go | Native / Go | Native bridge queries |
| --- | ---: | ---: | ---: | ---: |
| compiler | 2.326133s | 0.394658s | 5.89 | 78665 |
| repository | 0.294623s | 0.122622s | 2.40 | 4446 |

These include program load, parsing, rule execution and serialization. Native is
slower. Individual rounds and phases are in timing/measurements.json; no speedup
is claimed.

Commands, with test output redirected to the preserved logs:

```
bash cloud/setup.sh
source /workspace/adamic-tools/env.sh
ADAMIC_WAVE17_FIFTH_ARTIFACTS=/workspace/wave-17-fifth ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-17-typescript TMPDIR=/workspace/wave-17-artifacts go test ./stage1/cohere/typeaware -run '^TestWave17FifthAgreementAndMutants$' -count=1 -v -timeout=20m
go test ./bridge/tsgo/checker -count=1 -v
go vet ./...
go vet ./stage1/cohere/typeaware
TMPDIR=/workspace/wave-17-artifacts go test ./internal/oracle -run '^TestNativeAgreesWithNode$/internal/oracle/testdata/(functions|closures|generic_functions|method_closures)[.]a$' -count=1 -v
python3 stage1/cohere/typeaware/validation-wave-17-fifth/measure.py /workspace/wave-17-fifth stage1/cohere/typeaware/validation-wave-17-fifth/timing /workspace/wave-17-typescript
```

An earlier Node filter matched only the parent and TestTheOracleCatchesOneByte;
node.log records that 3.446s run. The corrected full-name filter above runs all
four intended fixtures, as node-fixtures.log shows. Fixture construction also
needed reserved-variable and argument-count corrections in the raw AST decoder,
and module isolation: without export {}, a local hook in one control shadowed
hooks in other files. Positive message assertions rejected that false coverage
before final validation. Constructor parameter properties were replaced with
explicit field declarations for Adamic's erasableSyntaxOnly setting.

Setup reports Go, clang, Node and submodules ready in 0s each, cache warm 19s,
total 19s on five processors with a four-CPU cgroup quota. TypeScript remains
pinned at 050880ce59e30b356b686bd3144efe24f875ebc8 and cohere at
715ba94f3608a6500086b1076ce5cb7e51b836db. Root source hashes, full diagnostic
streams, raw fixtures, sanitizer outputs and timing streams are preserved.

The full shared gate was not run. Its existing unknown-question mutation test
still hardcodes the former facts.go refusal line and cannot construct that
mutant after the earlier dedicated-question fallback registration. This batch
changes neither that shared test nor the bridge. Prior reports document the
actual unknown-request refusal checks. Complete boolean rule parity and
production JSX remain unfinished, and this report does not mark them passed.
