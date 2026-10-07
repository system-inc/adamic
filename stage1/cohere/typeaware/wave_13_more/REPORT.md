Built: native .a ports of no-obj-calls, no-object-constructor and no-promise-executor-return, plus a dedicated raw read-symbol question; the object rule has three parser-blocked controls.
Commits: prior six ports pushed through 60e1618c; this claim a215986d was pushed before implementation cc7b89d0.
Commands and outputs: production upstream tests pass; normal and sanitized comparisons match 430/433 controls, 289 findings, 171 suggestions and 162621 bytes; both frozen corpora agree.
Mutants: namespace alias, literal parentheses, executor span and read-shorthand compile, exit 0 and are caught only by byte comparison; suffix and wrong-kind mutants fail the dedicated checker test; sanitized released handle refuses access.
Not covered: three shared-parser refusals in no-object-constructor, full repository Go gate, emitted-JavaScript comparison, configured-globals surfaces and whole-source cohere CLI lint of .a.

## Selection, scope and status

All previous six claims were tested and pushed before these were selected.
Fetched all origin heads, scanned 347 refs, and ranked the 197 checker rules by
combined compiler/repository count with lexical ties. Excluded 25 checker ports
on origin/main and origin/codex/tsgo-c-library and 120 rules named in origin claim
Markdown. These were the first three of 52 remaining, each with combined volume
zero. selection.json records every excluded source/claim and remaining rule;
scan.py reproduces the scan. Claim a215986d was pushed before code.

NoObjCalls and NoPromiseExecutorReturn agree on all exported controls belonging
to them. NoObjectConstructor implements findings and suggestions but cannot run
on three inputs refused by the shared parser. This claim is therefore partially
blocked, not fully complete. No further claims were taken. BLOCKED.md records the
exact input shapes, parser positions and refusal messages.

The only shared edit in this batch is one dispatcher case, two Go-formatted
lines in bridge/tsgo/checker/facts.go. The new read_symbol.go and read_symbol_test.go
are dedicated files. Existing harnesses, registration generators, protected
compiler files, parser files and submodule pins are unchanged. All new executable
Adamic sources are .a. TypeScript fixture programs are exported test data.

## Native decisions and raw checker facts

NoObjCalls natively follows the production reference tracker: unmodified globals,
namespace objects, aliases through assignments/defaults/destructuring, both
conditional branches, logical/comma expressions, type-only wrappers, constant
computed keys and cycle guards. It retains duplicates when one call reaches two
globals and names the callee exactly as Go does. A new read-symbol question
returns the raw symbol identity, flags and declaration ancestry. It uses the
compiler's shorthand-value and local-export-target accessors where those sites
read a different symbol than GetSymbolAtLocation. Go returns no tracked paths,
lint verdicts or findings. Its native decoder lives in read_symbol.a.

NoObjectConstructor uses the bare callee, actual argument count (excluding type
arguments and optional punctuation), first declaration-file predicate, native
parenthesis decision and native preceding-character/tree classification. It
emits the exact suggestion id, explanation, replacement and byte span, including
the semicolon repair. NoPromiseExecutorReturn uses the actual first argument,
global Promise declaration, nearest function, unwrapped report span and original
body repair spans. Default allowVoid is false; --allow-void runs the production
true option on both sides. Anonymous function/class brace suggestions, loose
precedence, return adjacency and all insertion ordering are retained.

This batch reuses the prior raw platform-symbol and source-context questions.
It does not use the Go raw syntax graph builder needed by the previous output
rules. All rule decisions and reference traversal in this batch run native.

## Exact agreement and honest refusals

Upstream tests were exported through overlays of only the four rule-owned test
files. Shared test helpers were called unchanged, so all upstream assertions,
including applied suggestion outputs, passed. The export retains all 433 programs,
their exact fixture transformation and their decoded options. The independent
Go oracle loads the files and invokes the unchanged three production core rules;
it imports no bridge helpers. These upstream controls include deliberately
invalid TypeScript shapes which the production test harness still passes to
rules. The oracle follows that harness behavior rather than refusing its parse
diagnostics. The native parser explicitly refuses the three shapes below.

| Population | Roots/programs | Findings | Bytes | Normal / sanitized |
| --- | ---: | ---: | ---: | --- |
| Supported upstream controls | 430 | 289 | 162621 | equal / equal |
| Blocked upstream controls | 3 | Go reports 3 | 2009 Go bytes | native exit 70 / exit 70 |
| TypeScript compiler | 77 | 0 | 4933 | equal / equal |
| Frozen repository | 287 | 0 | 18485 | equal / equal |

The supported controls contain 120 namespace findings, 53 Object findings and
116 executor findings. They carry 171 suggestions and no automatic fixes. Full
finding/fix/suggestion serialization, including ordering, spans and every edit,
is compared. Controls with allowVoid true and false are both retained.

The blocked cases are 53d1c0a9ffc2fefa (`<foo />` in a .ts file), aff6ced2ca61fe89
(`<foo></foo>` in a .ts file) and defcb4c4ce6a2921 (a yield label and break yield).
The native shared parser reports respectively expected GreaterThanToken/got
SlashToken at 5, expected GreaterThanToken/got Identifier at 7, and expected
semicolon at 37. The control command exits 1 and prints `cases 433 failures 3`;
this is retained for both normal and sanitized runs, not turned into a green
whole-control result. There are no additional failures or sanitizer findings.
Ahra's stop instruction applies to these parser gaps; no shared parser edits or
source-rewriting workaround were made. The remaining rules were finished.

Both frozen manifests are the original populations: TypeScript v6.0.3 at
050880ce59e30b356b686bd3144efe24f875ebc8 and the pre-port repository's 287 roots.
The compiler canonical hash is
`e940d25193f84b9d9a2631c1c83d0a52a5f5b90a7246811d12445a1b08ffebcf`,
and repository hash is
`20dc789d505315aefcd366d819ba20b780f64bcdce0e5fbf2f26dc54ac120a2f`.
Zero corpus findings alone would let an empty port pass; positive controls and
semantic mutants prevent that. Complete streams, normal/sanitized per-case
hashes, fixture snapshots, options and refusal stderr are archived in validation/.

## Every mutant and lifetime check

All four semantic mutants compile, exit 0 and have empty native stderr. Only
full output comparison detects them:

| Mutant | Change | Catching case | First different byte |
| --- | --- | --- | ---: |
| namespace-alias | read conditional child 3 instead of false branch 4 | 15ed2ae55b6ded46 | 72 |
| literal-parentheses | invert replacement-parentheses predicate | 0741b4d4821fe4db | 587 |
| executor-span | report parenthesized body instead of unwrapped expression | 01a8b5f32591f310 | 101 |
| read-shorthand | use property symbol instead of shorthand read binding | 83efcbf16f6baaec | 72 |

Mutants.py and question_mutants.py reproduce them. The three known parser-refused
cases are explicitly skipped by the mutant search; every accepted catcher itself
must exit normally with empty stderr, so a parser refusal cannot kill a mutant.
The new read-symbol suffix guard mutant and wrong-kind guard mutant separately
compile and cause TestReadSymbolQuestion to fail with accepted suffix / accepted
wrong kind (exit 1). Their logs are retained. The complete touched checker
package test run passes in 0.205s; Go vet, gofmt and git diff --check pass.

The native read-symbol released-handle probe, linked to the sanitized archive,
exits 70 with `invalid or released checker handle` and no sanitizer report.
ASan/UBSan/LSan runs cover both full corpora and all supported controls, and even
the three parser refusals have no sanitizer report. The prior full bridge gate
and seven foundation mutants remain recorded in wave_13_next/validation:
149.709s, 162 positions, 3261 identical bytes, off-by-one input/output caught by
ASan, retained handle caught by stale assertion, wrong position caught at byte 6,
missing opt-in caught by refusal, missing C free and heap-for-region caught by
LSan. That unchanged infrastructure gate was not redundantly repeated here.

## Native time against production Go

After builds and validation ended, three alternating full-output runs per corpus
retained identical hashes. The medians include checker loading:

| Population | Native | Go | Native / Go | Queries |
| --- | ---: | ---: | ---: | ---: |
| Compiler | 3.235730s | 0.459277s | 7.05x | 189 |
| Repository | 0.506252s | 0.131094s | 3.86x | 401 |

Native is slower on both measured populations. Phase timings and all rounds are
in validation/measurements.json; earlier concurrent timings are not the primary
comparison. The toolchain was the existing setup: Go 0s, clang 1s, Node 1s,
submodule 1s, compiler cache 148s, done 148s; nproc 5 (four-core cgroup). Original
setup logs are linked from WAVE_13_REPORT.md. Setup was not run again.

## Reproduction

Source /workspace/adamic-tools/env.sh. Build the normal C archive with
`go build -buildmode=c-archive -o /workspace/wave13-more-checker.a ./bridge/tsgo/archive`
and native with
`/workspace/wave-13-adamic build stage1/cohere/typeaware/wave_13_more/suite.a -o /workspace/wave13-more-native --tsgo /workspace/wave13-more-checker.a`.
Build the oracle inside cohere with a Go overlay mapping cohere/wave13_more_oracle.go
to this directory's testdata/oracle.go:
`go build -overlay <overlay.json> -o /workspace/wave13-more-oracle ./wave13_more_oracle.go`.

Run this directory's scripts with the following absolute arguments, directing
each command's stdout/stderr to a log file:

```
python3 extract_controls.py /workspace/adamic /workspace/wave13-more-upstream
python3 validate.py /workspace/wave13-more-native /workspace/wave13-more-oracle /workspace/wave13-more-upstream/controls /workspace/wave13-more-validation
python3 corpora.py /workspace/wave13-more-native /workspace/wave13-more-oracle /workspace/wave13-more-corpora
python3 mutants.py /workspace/adamic /workspace/wave13-more-mutants /workspace/wave13-more-upstream/controls /workspace/wave13-more-validation /workspace/wave-13-adamic /workspace/wave13-more-checker.a
python3 question_mutants.py /workspace/adamic /workspace/wave13-more-question-mutants /workspace/wave13-more-upstream/controls /workspace/wave13-more-validation /workspace/wave-13-adamic /workspace/wave13-more-checker.a
python3 benchmark.py /workspace/wave13-more-native /workspace/wave13-more-oracle /workspace/wave13-more-bench
```

Sanitized archive uses `go build -asan -buildmode=c-archive`; native build adds
`--sanitize`. Repeat validate.py/corpora.py with the -asan executable. Known control
exit 1 is documented above. Portable manifests substitute TYPESCRIPT and ADAMIC
with checkout paths; corpora.py/benchmark.py use the original /workspace locations.
The dedicated Go command is `go test ./bridge/tsgo/checker -count=1 -v`.
No test output was piped. The full repository Go test gate and emitted-JavaScript
comparison were not run. Pinned cohere CLI does not discover .a sources, so no
whole-source cohere lint pass is claimed. No PR was opened.
