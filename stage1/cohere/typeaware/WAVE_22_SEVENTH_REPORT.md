Built: three numeric .a React listeners, raw Unicode facts, and exact computed-key message text; custom patterns remain blocked.
Commits: claim 698985fdb; rebased port db9daca83; caches f82538d34; guards a2a3371f4; computed-trivia correction 5d19f7f84; main b8fb957aa.
Commands: final Go/native corpus and control comparisons PASS; all eighteen earlier/current rule gates re-green; bridge, Node oracle and vet PASS; native/compiler 26.728s versus Go 0.466s.
Mutants: eighteen rule mutants, three new regression/guard mutants, Unicode case, released registries, scope export, seven common bridge mutants and Node one-byte all caught.
Not covered: direct shared JSX/parser integration, whole-rule emitted-JavaScript comparison, native configured propNamePattern, arbitrary configuration combinations and the full repository gate.

The new rules are react/no-object-type-as-default-prop, react/no-unstable-nested-components
and react/sort-default-props. Their rule.json kinds are respectively [307], [307]
and [173, 227], matching Go's SourceFile, PropertyDeclaration and BinaryExpression
listeners. One driver builds a numeric node cache and hands only subscribed nodes
to listeners. Visitors neither refetch their subject from the parser nor compare
its kind as a string. The two SourceFile listeners reproduce Go's own syntax walks
and share one component detector. The detector needs no HIR, SSA or capture analysis.

The unicode-node-text question exports scalar identities and Go Unicode simple-case
mappings, never a React naming verdict. It has dedicated Go implementation/test
and an Adamic decoder, with one new facts.go registration line. All other edits
in this batch are inside its own rule directories, claim and evidence/report.
Shared registration generation, the test harness and compiler emission were not edited.

The initial inventory inspected 570 origin refs and 33 distinct Markdown claim blobs.
The three claimed rules were the first available entries of the combined ranking,
all tied at zero volume. selection.json records the exact refs and exclusions.
The claim was pushed before implementation. No more rules were claimed.

Validation uses the published native JSX parser a8a62d62ca49db7415e14c3887dd305022b17309
in an isolated source tree. No Go AST is substituted for native parsing. Current
main b8fb957aa839a9e8cb0b54279dd9864fa317bd30 still refuses the direct shared parser
build at parser.ts:34:20: escaping this before every field is set. The shared
harness/model ab70f38d4 remains on origin/lint-rules/harness, outside current main.
The seven-argument report wire remains unchanged. compiled-source-hashes.json
records the actual isolated compilation inputs.

The authoritative final run imports literal payloads from the three Go rule tests
using Go's parser/unquoter. Go's production parser accepted 272 and explicitly
rejected 251 non-source or malformed payloads. Added controls cover Unicode names,
Symbol and other forbidden defaults, computed keys, aliases, spreads, hooks, maps,
render props, comments, BOMs and Go Unicode whitespace. Default findings are 73,
68 and 40 respectively, totaling 181. ignoreCase yields 182; allowAsProps yields
152. Both options match byte for byte. Every diagnostic field is serialized,
including all fixes and suggestions; these three rules produce zero of either.
Go truth is the unmodified production registry with an independent program loader.

The repository manifest has 287 files and the frozen TypeScript compiler manifest
has 77. Cohere is 715ba94f3608a6500086b1076ce5cb7e51b836db; TypeScript is
050880ce59e30b356b686bd3144efe24f875ebc8. Both corpora produce zero findings
for this batch; positive controls and mutants establish that silence can fail.
All controls and both corpora also match under ASan/UBSan/LSan with empty native stderr.

| Input | Findings | Identical bytes | Native process | Go process |
| --- | ---: | ---: | ---: | ---: |
| controls | 181 | 73,462 | 0.357019s | 0.135401s |
| ignore-case | 182 | 73,739 | 0.374815s | 0.159056s |
| allow-as-props | 152 | 58,964 | 0.349760s | 0.180876s |
| repository | 0 | 18,485 | 4.086046s | 0.220424s |
| compiler | 0 | 5,857 | 26.728235s | 0.465877s |

These are recorded whole-process comparisons, including program loading, native
parsing, rule work, output and release. Native remains about 18.5x slower on the
repository and 57.4x slower on the compiler corpus. Earlier timing evidence is
retained; the previous separate compiler timing was 41.512s. Numeric per-file
facts now avoid irrelevant JSX-attribute and React-class ancestor walks. This is
not a speed target pass. The earlier instrumented corpus run made zero checker
queries, so the observed cost there is not explained by bridge query count.

The shared regex rows are pinned at origin/codex/lint-regex 071fb0128. The hook
pattern uses its translated JS literal without global iterator state, because
this rule calls test. The default render* expansion also uses a JS literal.
Configured globs are translated once and passed to new RegExp(source, 'u').
Nine option controls match Node and Go's unchanged production glob builder.
Native stage 0 refuses options.a:12:46 with `stage 0 can't lower RegExp with a
nonconstant pattern yet`. gaps/dynamic-pattern.a preserves that exact blocker.
There is no hand-written regex matching fallback. No native custom-pattern
configuration is claimed complete.

A final edge probe found an actual message mismatch at byte 127: Go preserves
leading comments and U+FEFF in computed key expressions. The correction uses
full native source spans and Go's Unicode White_Space rather than token text and
JavaScript trim. The original probe now matches Go in normal and sanitized native;
the computed-trivia mutant reproduces the old behavior and is caught only by bytes.

All six new native mutants compile and execute successfully with empty stderr:

| Mutant | Detector |
| --- | --- |
| default-id | Byte oracle, first difference 772; build/run 0, empty stderr |
| nested-id | Byte oracle, first difference 18,310; build/run 0, empty stderr |
| sort-id | Byte oracle, first difference 53,364; build/run 0, empty stderr |
| jsx-attribute-flag | Byte oracle, first difference 18,244; build/run 0, empty stderr |
| class-fallback-flag | Byte oracle, first difference 933; build/run 0, empty stderr |
| computed-trivia | Byte oracle, first difference 66,117; build/run 0, empty stderr |

The three rule mutants change diagnostic IDs; the extra mutants strip computed
trivia, erase the JSX-attribute flag, or incorrectly accept nodes when no React
class exists. The Unicode question mutant substitutes simple lower for simple
upper; TestUnicodeNodeText fails with `lost Unicode simple case` (test exit 1).
The new question rejects a released program with exactly panic 70. A valid x
Identifier probe with the delete-from-live-registry mutation executes with exit
0 and empty stderr, violating the required released refusal. released-checks.json
records both executions. The normal and sanitizer probes do not rely on an invalid
node range to reject the handle.

All fifteen previous implementations were re-green on the new main. Their rule
mutants are: nullish-suggestion, qualifier-fix, private-read, promise-condition,
spread-await-edit, lost-write-span, global-provenance, global-declaration-span,
timer-string, numeric-prefix, has-own-fix-span, spread-parens, fragment-id,
undef-id and adjacent-judgment. Every one ran cleanly and failed its byte oracle.
The scope-export normalization mutant also ran cleanly and failed at byte 29,787.
The legacy twelve-rule controls/mutants/repository gate passed in 845.836s. Its
compiler corpus was accidentally omitted by the initial environment variable;
an owned overlay, with no shared harness edits, explicitly rechecked all twelve
on the 77-file compiler corpus in normal and sanitized builds, PASS 529.963s.
The prior JSX three-rule gate likewise passed all options, mutants and both corpora.

The full bridge gate passed in 272.605s (checker package 3.510s): 1,600 independent
positions over checker.ts, parser.ts, types.ts and utilities.ts, 54,982 identical
bytes under sanitizers. Common mutants and detectors are input length +1/ASan,
output length +1/ASan, retained handle/stale-handle assertion, source-file rather
than node type/byte oracle (byte 6), removed link guard/refusal test, omitted C
output free/LSan, and heap allocation inside a region/LSan. The filtered external
Node/native/emitted-JavaScript oracle passed regexp.a and main's new inherited
static-field fixture in 5.443s. Its one-byte string mutant was caught as stdout
difference. Vet was clean. No full repository gate was run.

Exact final commands, each with stdout/stderr directed to its own log file:

```
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/typeaware/rules/wave-22-seventh/validate.py > /workspace/wave-22-seventh-final-validation.log 2>&1
python3 stage1/cohere/typeaware/rules/wave-22-seventh/prove_flags.py > /workspace/wave-22-seventh-flag-mutants.log 2>&1
python3 stage1/cohere/typeaware/rules/wave-22-sixth/validate.py > /workspace/wave-22-seventh-postrebase-sixth.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./stage1/cohere/typeaware -run '^TestWave22(AgreementAndMutants|NextAgreement|ThirdAgreement|FourthAgreement)$' -count=1 -timeout=30m -v > /workspace/wave-22-seventh-legacy12.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/wave-22-typescript-pinned ADAMIC_WAVE22_LANDING_ARTIFACTS=/workspace/wave-22-seventh-landing-work go test -overlay /workspace/wave-22-seventh-work/landing-overlay.json ./stage1/cohere/typeaware -run '^TestWave22LandingCompilerCorpus$' -count=1 -timeout=20m -v > /workspace/wave-22-seventh-landing-corpus.log 2>&1
ADAMIC_TSGO_CORPUS=/workspace/wave-22-typescript-pinned go test ./bridge/tsgo/... -count=1 -timeout=15m -v > /workspace/wave-22-seventh-postrebase-bridge.log 2>&1
go test -overlay /workspace/wave-22-seventh-work/unicode-mutant-overlay.json ./bridge/tsgo/checker -run '^TestUnicodeNodeText$' -count=1 -v > /workspace/wave-22-seventh-unicode-mutant.log 2>&1
go vet ./bridge/tsgo/... ./stage1/cohere/typeaware > /workspace/wave-22-seventh-postrebase-vet.log 2>&1
go test ./internal/oracle -run '^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^(regexp\.a|inherited_static_field_read\.a)$|^TestTheOracleCatchesOneByte$' -count=1 -v -timeout=10m > /workspace/wave-22-seventh-postrebase-node.log 2>&1
```

The Unicode mutant command is expected to fail; every production gate above
passes. logs/ retains the final runs and earlier attempts. summary.json identifies
the authoritative final comparison data; new-rules/ holds statuses, raw streams,
controls, regex rows, option oracle and released probes. landing-corpus/ holds the
legacy compiler comparisons. Earlier failed build/probe logs remain as observations.

Toolchain setup completed: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm and
done 187s. nproc is 5; cgroup quota is four CPUs. The branch was rebased onto main
b8fb957aa; a non-force ancestry merge preserves the previous published tip 698985fdb.
Only codex/typeaware-wave-22 is a push target. The three older React graph claims
remain parked on the previously recorded HIR/SSA/capture dependency #dnv6f2c.
