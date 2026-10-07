Built: three .a candidates for no-unnecessary-type-constraint, prefer-as-const and prefer-enum-initializers, with every fix and suggestion represented.
Commits: claim 9c4b547c pushed before code; as-const af336eb6; enums 5f780bf7; constraints and comparison transport 67737231.
Commands and outputs: 133 upstream cases PASS, 48,589 identical bytes; 222 compiler/stage1 files PASS, 13,038,944 identical bytes; setup 29s, nproc 5.
Mutants: constraint_suggestion_wrong, const_append_wrong and enum_second_suggestion_wrong compiled and ran normally, then failed solely by Go byte comparison on source Node, emitted JavaScript and sanitized native.
Not covered: default integration remains blocked by .a registration, profile compilation and the old edit representation; the full registered upstream gate fails on the prior Tailwind JSX case; full repository tests were not run.

The branch was pushed first, then every origin head was fetched without recursing
into submodules. Selection inspected 320 origin refs and 16 distinct actual claim
documents, excluding report/evidence files from claim assertions. The original
46-rule helper-ready handoff linked by helpers/REPORT.md was fully claimed.
Inventory syntax-ready positions 1-8 were already claimed; positions 9-11 were
neither present on main ef3d907ecdc4c771b016f7d9c52372def057a340 nor claimed.
The exact ref snapshots, ordered candidate checks and claim blob provenance are
in ../rules/typescript-no-unnecessary-type-constraint/selection.json. The claim
was committed and pushed as 9c4b547c before the new directories were written.

Each rule owns its descriptor without an order field, .a implementation and
messages, upstream Go adapter, raw witness and mutant. No shared source was
edited, including dispatch, corpus, registry, oracle, parser, native or lowering
files. No authored Adamic .ts file was added. The constraints directory also owns
a complete repair transport, comparison entry point, external Go comparison
artifact, scratch-only compatibility patch, virtual Go tests and reproduction
runner. Its sibling rules import its repair types and transport by named imports.

The constraint rule reports only bare any/unknown constraints. It preserves the
name's token range, the distinct constraint-removal suggestion range, and the
filename-dependent trailing comma for a single generic arrow parameter in .tsx,
.mts and .cts, accounting for defaults, real commas and comments. The as-const
rule matches cooked string/canonical number values only, declines parentheses,
booleans, bigint and negative-number shapes, preserves destructuring's lack of
repair, and carries both independent edits for an annotation. The enum rule
preserves three ordered suggestions, source spelling including quoted names,
and member positions rather than computed enum values. None has options.

The inherited finding shape can carry only one repair and one suggestion over
the diagnostic range. These rules need different ranges, two fixes or three
suggestions. Silently dropping extra edits would violate the requested bar.
repairs.a retains complete edit lists in a file-scoped comparison transport.
main.a is an owned entry point that prints every fix and every suggestion's ID,
message, edit ranges, replacement bytes and applied text. Fixes are applied and
reparsed; suggestions remain unapplied to the reported fixed file. The existing
Go rules and Go converging fixer provide the independent expected findings and
fixed source. comparison_oracle.go.txt extends only the comparison serialization;
it does not change cohere rule bodies. UTF-16 Adamic indexes are converted to Go's
UTF-8 byte offsets at the output boundary.

Shared compatibility is proposed in compatibility.patch and applied only to
scratch copies with Go overlays. It retains the earlier .a discovery/import
rewriting and profile-function fix, preserves captured filename extensions and
uses the owned rich entry points for source Node, emitted JavaScript and native.
Default repository commands still fail: registry generation tries to open
next-no-assign-module-variable/rule.ts, and lint tests fail to compile because
profile_test.go ranges over the portFiles function. These observations are in
evidence/default-registry.txt and default-test.txt. Applying the shared changes
would cross CLAUDE.md's explicit directory ownership boundary. This unit does not
claim default production integration or a green full repository gate.

After sourcing /workspace/adamic-tools/env.sh, reproduce against TypeScript
v6.0.3 commit 050880ce59e30b356b686bd3144efe24f875ebc8:

```
python3 stage1/cohere/lint/rules/typescript-no-unnecessary-type-constraint/validate.py \
  --scratch /tmp/lint-wave1-08-third-reproduce \
  --typescript /tmp/lint-wave1-08-typescript
```

The runner writes output directly to separate logs, runs the independent gates
even if the full inherited upstream gate fails, and returns nonzero for that
failure. Its --prepare-only mode was run successfully, and all three reproduced
scratch shared files matched the tested scratch copies byte for byte. The
observed commands below used -overlay=/tmp/lint-wave1-08-third/overlay.json.

| Command | Observation |
| --- | --- |
| GOFLAGS=-overlay=/tmp/lint-wave1-08-next/overlay.json bash cloud/setup.sh | PASS; Go/clang/Node/submodules ready 0s each; build cache warm 29s; total 29s; nproc 5 |
| go run -overlay=... ./cmd/lint-registry | PASS; all eleven descriptors, including these three |
| go test -overlay=... ./stage1/cohere/lint -run '^TestOwnedWitnesses$' -count=1 -v -timeout=10m | PASS 29.128s; 13,350 identical bytes |
| go test -overlay=... ./stage1/cohere/lint -run '^(TestThirdUpstream|TestThirdShapes|TestCompilerAndStage1Agree)$' -count=1 -v -timeout=20m | PASS 65.923s for upstream and edge cases; compiler test initially SKIP because its env variable was absent, then run explicitly below |
| ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-08-typescript ADAMIC_GATE_UNCACHED=1 go test -overlay=... ./stage1/cohere/lint -run '^(TestCompilerAndStage1Agree|TestThirdThroughput)$' -count=1 -v -timeout=20m | PASS 103.432s; corpus 49.03s, throughput 54.39s |
| ADAMIC_GATE_UNCACHED=1 go test -overlay=... ./stage1/cohere/lint -run '^TestMutants$/(constraint_suggestion_wrong|const_append_wrong|enum_second_suggestion_wrong)$' -count=1 -v -timeout=10m | PASS 78.306s; three semantic mutants caught on all three Adamic executions |
| ADAMIC_GATE_UNCACHED=1 go test -overlay=... ./stage1/cohere/lint -run '^TestRulesAgree$' -count=1 -v -timeout=10m | FAIL 37.670s on inherited Tailwind JSX source; 535 captured cases total |
| go test -overlay=... ./stage1/cohere/lint/registry -count=1 -v | PASS 0.051s; deterministic regeneration and eleven descriptor rejections |
| go vet -overlay=... ./... | PASS, exit 0 with empty output |
| ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m | PASS 0.621s; native and Node cache hits 0, misses 1 |

All 43 constraint cases, all 69 as-const cases and all 21 enum cases captured
from the actual Go tests passed without exclusions. This includes the upstream
fix-vector and suggestion assertions, whose Run calls are captured even when the
assertion is not a standard harness finding check. The 133 cases produce 48,589
identical bytes on Go, source Node, emitted JavaScript and ASan/UBSan native.
Thirteen additional cases produce 13,282 identical bytes, covering default versus
constraint syntax, infer parameters, modifier tokens, file extensions, commas in
comments, numeric spellings, Unicode/astral strings and byte ranges, computed
properties and enum keys, optional/definite properties and initialized members.

The compiler/stage1 run covers 77 pinned compiler files and 145 stage1 .ts/.a
files, including gaps and the new source modules. All eleven rules run. It
compares every finding, every fix, all suggestion IDs/messages/ranges/replacement
bytes, and fixed source, producing 13,038,944 identical bytes. The corpus uses a
compact manifest marker to omit the redundant entire-file rendering after each
suggestion; the edits themselves are never omitted. Complete applied suggestion
text remains checked on all 133 upstream cases and thirteen edge cases. The
initial expanded corpus reference and source Node each rendered 1,412,802,629
bytes; native was still processing that redundant output. That run was stopped
and is recorded as a terminated failure, never as a pass. The compact run
replaced it. Native sanitizer executions must exit zero with no stderr, including
Linux's default leak check.

The full eleven-rule upstream gate still cannot complete: the prior Tailwind
case `export const c = <div className="flex ml-4" />;` causes source Node to
exit 70 with `parser slice expected GreaterThanToken, got Identifier at 22`.
This failure occurs before the four-backend comparison completes. The selected
new-rule corpus has no such exclusion and passes in full. No claim is made that
the parser handles arbitrary JSX or recovers all invalid syntax.

The three mutants compiled and ran with exit zero and empty stderr; no compiler
error, parser failure, runtime panic or sanitizer error killed them. Changing
removeTheConstraint to removeOtherConstraint changes only the suggestion ID, so
a findings-only comparison misses it. Appending ` as unknown` instead of
` as const` changes the second fix and fixed source while preserving the finding.
Changing the enum's second suggestion from position+1 to position+2 changes a
later suggestion that a first-suggestion-only comparison misses. Go comparison
caught each on source Node, emitted JavaScript and sanitized native. The extra
core oracle test proves the one-byte external comparison mutant still fails.

Throughput below is findings per second over 77 compiler files plus one file
with 1,000 positive declarations for the selected rule, 78 files total. Counts
match Go in all five rotating rounds. Native timing uses the release build;
sanitized native is used for correctness. Measurements include process startup,
parsing and count output, use the best of five durations, and do not render fixes
or applied suggestions. They are synthetic-positive workload measurements rather
than throughput for the unmodified compiler corpus alone.

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
| --- | ---: | ---: | ---: | ---: |
| @typescript-eslint/no-unnecessary-type-constraint | 1,000 | 700.47 | 1,030.35 | 4,061.34 |
| @typescript-eslint/prefer-as-const | 1,000 | 629.10 | 1,002.91 | 4,030.34 |
| @typescript-eslint/prefer-enum-initializers | 2,680 | 1,731.85 | 2,274.08 | 10,619.83 |

Raw logs are in ../rules/typescript-no-unnecessary-type-constraint/evidence/.
Patch context and diagnostic log whitespace is retained byte for byte; those
artifacts can trigger git diff --check. Authored .a sources, descriptors and Go
adapters pass the whitespace check. Earlier reservations and their reported
integration blockers remain unchanged. No PR was opened.
