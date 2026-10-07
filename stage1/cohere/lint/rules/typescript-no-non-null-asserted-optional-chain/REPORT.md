Built: three registered TypeScript syntax-rule implementations with complete owned diagnostic validation; shared suggestion integration is blocked.
Commits: claim 717e5418; optional-chain implementation/evidence 201bde4f; non-null assertion 83d16b57; this-alias bc6e6be6; final report commit follows.
Commands: 100 Go configurations plus five owned witnesses and 224 whole sources match Go/Node/emitted JS/sanitized native; three mutants and throughput pass.
Mutants: wrong optional-chain deletion range, missing dot replacement, and RHS parenthesis unwrapping caught only by Go comparison on all three execution paths; integration-guard mutant caught by explicit refusal check.
Not covered: shared-driver suggestion integration and full repository gate; no additional rules claimed after Ahra's correction.

## Rules and source format

This unit implements:

- @typescript-eslint/no-non-null-asserted-optional-chain
- @typescript-eslint/no-non-null-assertion
- @typescript-eslint/no-this-alias

Each rule owns its descriptor, listener, exact messages, independent Go adapter,
raw witnesses and compiling mutant. The claim was pushed as 717e5418 before code.
Selection used origin/main ef3d907, 320 origin refs and 39 unique Markdown claim
blobs; the 46-rule helper list was exhausted. No further claim was made.

Ahra's later correction explicitly permits .ts while .a harness support is being
added. The final new source modules are .ts. The initial .a implementations were
validated with a scratch-only Go registry overlay; that patch was removed after
switching to the authorized fallback. Normal `go run ./cmd/lint-registry` now
succeeds and discovers all eight descriptors, including these three. Shared
registration, traversal, finding, harness and compiler files are unchanged.
The codex/lint-harness-dot-a ref was not yet available when fetched; its exact
remote-ref failure is preserved in evidence/dot-a-fetch.log.

## Behavior

The optional-chain rule distinguishes a final assertion from one reached through
by a continuing chain, unwraps parentheses only at the specified boundary, and
reports an assertion before a new optional-chain root. Its suggestion removes
only the final exclamation token, preserving its exact range and message.

The general non-null rule reports every assertion and reproduces Go's conditional
suggestions. Object/callee positions, already-optional links, destructuring and
assignment/update/delete targets, wrappers and comments are held separately.
Dot access keeps the two distinct edits returned by Go, including its actual
first-dot search through the intervening source. None of the suggestions is an
automatic fix. All three rules' fixed source is unchanged.

The this-alias rule preserves case-sensitive TypeScript filename gating, default
allowance of destructuring, allowed-name exemptions, assignment operators, narrow
unwrapped RHS matching, and recursive LHS unwrapping. A type-only `: this` is not
an initializer. The oracle adapter consumes the real captured decoded Go options;
the listener also understands the wire spellings and the destructuring inversion.
Arbitrary malformed configuration acceptance/diagnostic parity is not claimed.

## Validation boundary and exact blocker

The shared Finding model cannot represent complete suggestion edit lists. The
unmodified shared Go oracle independently reproduces the blocker on `foo!.bar`:
it panics with `unexpected suggestion shape`, because it requires one edit whose
range equals the diagnostic range. Real Go returns two edits, and even the
optional-chain rule's single token deletion has a different range from its
whole-expression diagnostic.

The owned diagnostic.ts retains every suggestion ID, description, edit range and
replacement. driver.ts uses the normal generated RuleSet/Linter traversal and
explicitly enables its complete-record adapter, then serializes those records.
The owned independent Go oracle executes unmodified Go rules and emits the same
protocol: rule, ID, message, UTF-8 diagnostic range, fix count, suggestion count,
suggestion IDs/descriptions, every edit in order, and unchanged fixed source.
No Go AST projection is used: Go and Adamic parse every source independently.

An ordinary driver that does not enable the complete-record adapter explicitly
refuses a suggestion with exit 70 and:

```
NotYet: shared Finding model cannot serialize full suggestion edits
```

It therefore cannot silently drop suggestions. No-suggestion findings, including
all this-alias findings, use the normal context. The refusal is tested on source
Node, emitted JavaScript and ASan/UBSan native. A compiling guard mutant replaces
its condition with false, finishes normally, and is caught by the exact refusal
check. Shared integration of the two suggestion-bearing rules is blocked until
the owner supplies a complete suggestion model and reporting protocol. This unit
stops there as Ahra instructed; it does not repair shared files or claim the
ordinary shared harness passes.

## Corpus and results

Original Go tests are executed through an overlay of their capture hook, including
cases that assert fields directly instead of calling ExpectFindings. It changes
no Go rule. Deduplication preserves filename and decoded options. The original
suites supply 24 optional-chain, 36 non-null, and 40 this-alias configurations.
`evidence/upstream-cases.json` preserves the exact 100 inputs and options. Five
owned witnesses add comment preservation, a new optional root and a parenthesized
RHS control. Go, Node source, emitted JavaScript and sanitized native agree on
68,287 canonical bytes over those 105 configurations.

The compiler is TypeScript v6.0.3, pinned at
050880ce59e30b356b686bd3144efe24f875ebc8. Every .ts/.a source under stage1 is also
included, without excluding gaps or new rule modules. There are 77 compiler and
147 stage1 files: 224 files, 672 rule/file pairs. The final .ts run compares
38,679,837 identical canonical bytes on all four sides. It does not compare the
human CLI renderer's formatting; it compares every diagnostic and repair field
listed above. This is rule validation through the owned complete-record adapter,
not certification of the blocked shared renderer.

Final bounded gate:

```
source /workspace/adamic-tools/env.sh
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-04-typescript ADAMIC_LINT_BENCH=1 \
ADAMIC_WAVE104_EVIDENCE=/workspace/adamic/stage1/cohere/lint/rules/typescript-no-non-null-asserted-optional-chain/evidence \
python3 stage1/cohere/lint/rules/typescript-no-non-null-asserted-optional-chain/validate.py > /tmp/lint-wave1-04-next-final.log 2>&1
```

PASS, 141.777s: original configurations and witnesses, all three mutants, complete
compiler/stage1 source parity and throughput. The refusal test was added while
that compiled test binary was already running, so it was run separately:
`validate.py -run '^TestSharedSuggestionRefusal$'`, PASS 24.647s. The independent
unmodified shared-oracle blocker probe also passes separately in 5.493s.
Re-running validate.py now includes both probes. All test output goes to files,
never pipes. Successful mutant runs must exit 0 with empty stderr; compiler,
clang, sanitizer or runtime failures receive no semantic-mutant credit.

The first validator compilation failed on a Go slice literal. The first semantic
comparison found a suggestion incorrectly retained on an object assignment
because the stage1 parser stores a binary operator as child 1, not node.operator.
A subsequent correction accidentally changed unary-operator lookup too and was
caught as a runtime panic. Both sites now use their actual parser representation.
Every superseded failure is preserved in evidence; the final gate passes.

## Throughput

Best of three complete-corpus count runs, including process startup, source reads,
independent parsing and rule execution. Same 224 files and per-rule finding counts
on all three sides. Native timing is unsanitized; correctness above is sanitized.

| Rule | Findings | Native findings/s | Node findings/s | Go findings/s |
|---|---:|---:|---:|---:|
| no-non-null-asserted-optional-chain | 7 | 5.64 | 8.10 | 34.21 |
| no-non-null-assertion | 1,123 | 898.08 | 1,348.45 | 5,321.02 |
| no-this-alias | 0 | 0.00 | 0.00 | 0.00 |

Zero aliases is an observed corpus count, not an unmeasured benchmark. Its native,
Node and Go elapsed times are 1.225242s, 0.936337s and 0.211346s; its positive
upstream/witness cases prove it fires. Native is slower than both Go and Node on
these measurements. No speedup is claimed.

## Mutants

- Optional-chain suggestion: move the deletion back one character. Findings still
  match; only the independent comparison of edit ranges catches the mutant.
- Non-null assertion: leave the dot as `.` rather than replacing it with `?.`.
  Findings still match; only suggestion replacement comparison catches it.
- This alias: unwrap a parenthesized initializer. The source stays valid and the
  mutant finishes, but it adds a finding Go deliberately omits.
- Integration guard: disable the explicit adapter requirement. This is caught by
  the refusal check, not credited as a fourth rule-semantic mutant.

All three rule mutants compile and finish normally on Node source, emitted JS
and ASan/UBSan native before their wrong output is compared to Go.

## Setup and remaining checks

Setup still exits 1 warming the shared lint package because profile_test.go:32
ranges over portFiles instead of calling portFiles(t). Its exact error is in
setup.log. All setup timing phases report 0s: Go, clang, Node and submodules.
There is no cache-ready/done timing. Go 1.27.1, clang 20.1.8 and Node 24.19.0 are
available; the setup sanitizer probe succeeds. nproc is 5 and cpu.max is
400000 100000. The owned validator works around the shared package compilation
failure without changing it.

`go vet ./stage1/cohere/lint/rules/...` passes with empty output. An earlier vet
attempt explicitly named the two adapter-only Go directories and failed because
their ordinary builds exclude the lintoracle-tagged files; that attempt is also
preserved. `git diff --check` passes. The full repository gate is not run and the
known profile compilation failure remains. No protected compiler file, shared
registry/harness file or cohere submodule source was edited. No new rules were
claimed after Ahra's correction, and no pull request is opened.
