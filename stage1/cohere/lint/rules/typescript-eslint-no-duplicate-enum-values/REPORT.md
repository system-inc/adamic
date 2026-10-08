Built: two directory-registered .a rule ports; no-confusing-non-null-assertion remains blocked on shared suggestion representation.
Commits: pre-code claim e9a72247; enum implementation ecb47818; delete implementation 1d8da3ab; evidence commit follows.
Commands and outputs: owned validation PASS 150.735s; overlay vet exit 0; filtered oracle PASS 0.275s; setup retry 15s, nproc 5.
Mutants: advancing numeric duplicate anchors and accepting unary plus both compile, exit cleanly and fail only Go comparison on all three runtimes.
Not covered: blocked rule port/mutant/throughput, the full repository gate, and production integration without the shared .a compatibility patch.

## Selection and claim

The existing branch was already fully pushed at 32fffb94. Fetched every origin head
without recursive submodule fetching, scanning 305 remote refs. Helper-ready order
comes from HELPERS.md, linked by helpers/REPORT.md. All 46 names occur in direct
origin claim records, including existing-port skips. The earlier instruction to
skip existing origin ports still applies to those skips.

The inventory calls its syntax-only ready wave `syntax ready for AST/API
adaptation`, rather than `syntax-only`. Its first eligible entries were:

1. @typescript-eslint/no-confusing-non-null-assertion
2. @typescript-eslint/no-duplicate-enum-values
3. @typescript-eslint/no-dynamic-delete

None had an implementation selection on main ef3d907e or a direct origin claim
before this worker's update. Claim e9a72247 was pushed before either implementation
was written. evidence/selection.json records the inventory pin, main pin and
per-rule claim matches. That audit was saved after our update, so it includes our
new claims. Reports and evidence logs were excluded from claim classification.

## Implementation

Each completed rule owns its descriptor, .a listener and message module, independent
upstream Go adapter, .a mutant and raw TypeScript witness. Neither descriptor has
an order field. No shared driver, registry, context, finding model, compiler or
cohere submodule source was edited.

Enum values use separate numeric and string tables scoped to each declaration.
The numeric table retains the first initializer anchor; the string table advances
to the previous initializer. Both use the parser's cooked literal value. Unary,
parenthesized, template and bigint initializers retain Go's silence.

Dynamic delete unwraps parentheses at all three Go sites and accepts strings,
numbers and unary minus over numbers. It checks the access's own question-dot
token rather than its inherited optional-chain flag. Findings anchor at the key;
there are no fixes or suggestions for either completed rule.

## Four-way comparison

Pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db supplies unmodified rule
bodies, message text, canonical formatter and converging fixer. The independent
capture overlay executes every original rule test and records complete input and
options. The same manifest runs on Go, Node source, emitted JavaScript and native
with ASan/UBSan plus leak checking. Every successful command must exit 0 with no
stderr before stdout is compared byte for byte.

| Rule | Captured source/options cases | With witness findings | Fixture bytes | Corpus bytes |
|---|---:|---:|---:|---:|
| no-duplicate-enum-values | 57 | 35 | 23,111 | 12,341,922 |
| no-dynamic-delete | 42 | 25 | 8,616 | 12,340,345 |

All four outputs match, including descriptions, ranges, repair metadata and fixed
source. Each corpus contains all 77 pinned TypeScript compiler files and all 139
stage1 .ts/.a sources, including generated registry sources. Compiler pin:
050880ce59e30b356b686bd3144efe24f875ebc8. There are no excluded cases for either new
port. Upstream capture also runs previously registered rules and explicitly skips
the previously documented module-alias JSX fixture; it is outside both new rules.

## Mutants

Both variants compile and run with exit 0 and no stderr on Node source, emitted
JavaScript and sanitized native. A compilation error or sanitizer failure would
fail the harness instead of counting as a caught mutant.

- Enum: always update the numeric anchor. On `A = 1, B = 0x1, C = 1`, the second
  finding moves from column 14 to column 21. Counts and message text remain correct;
  the independent location comparison catches the error on all three runtimes.
- Delete: accept unary plus in the unary-minus branch. `delete c[+7]` loses its
  finding; the independent Go comparison catches it on all three runtimes.

## Findings per second

Release native, Node source and Go run the same full compiler/stage1 manifest in
five interleaved rounds. Values below are the best elapsed time per runtime and
include process startup, reading and parsing. Native is unsanitized for timing;
parity and mutants used sanitized native. Natural finding counts are nonzero,
so no synthetic supplement was used.

| Rule | Natural findings / files | Native seconds / findings per second | Node seconds / findings per second | Go seconds / findings per second |
|---|---|---|---|---|
| enum duplicates | 4 / 216 | 1.308367 / 3.06 | 0.884200 / 4.52 | 0.205970 / 19.42 |
| dynamic delete | 2 / 216 | 1.232448 / 1.62 | 0.852239 / 2.35 | 0.206143 / 9.70 |

These are low-count natural-corpus measurements, not evidence that Adamic beats Go.
Full round observations are in evidence/validation.log.

## Concrete blocker for no-confusing-non-null-assertion

Observation: Finding carries one suggestion description and one replacement;
RuleContext.report cannot represent an ordered suggestion array or multiple edit
ranges. The independent shared Go serializer rejects suggestions unless there is
exactly one suggestion, exactly one edit and that edit covers the diagnostic range.

The owned TestWave14NextSuggestionBlocker builds that serializer with a temporary
selection of the real upstream rule. All three commands exit 2 with
`panic: unexpected suggestion shape`:

- `a! == b;`: removal edits only the bang, not the diagnostic's whole range.
- `a + b! == c;`: wrapping has two separate insertion edits.
- `a! in b;`: both removal and wrapping suggestions, in that order.

Separately, all original TestNoConfusingNonNullAssertion tests pass in Go (0.008s),
including exact descriptions and applied multi-edit suggestion rewrites.

Inference: full parity requires expanding the shared finding, context, serialization
and comparison contracts. CLAUDE.md says: "Never edit a dispatch, oracle, corpus or
copied-file list." The user's territory is the assigned rule directories. No
incomplete listener or placeholder descriptor is registered for this rule. It stays
claimed and blocked; no rule-specific mutant or throughput is claimed for it.

## Reproduce and gate

Source /workspace/adamic-tools/env.sh in each shell. Toolchain: Go 1.27.1,
clang 20.1.8, Node 24.19.0. nproc is 5; cgroup cpu.max is 400000 100000.

Initial `bash cloud/setup.sh` failed with exit 1 in cache warming because
profile_test.go:32 ranges over portFiles, now a function. Initial timing lines:
Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s.
The existing .a discovery incompatibility also remains on this branch.

validate.py creates a temporary Go overlay for .a discovery and harness copying,
plus the inherited profile typo, and injects the owned suite into the parent test
package. It reuses the prior module-alias suite's emitted-JS and corpus helpers.
The concrete minimal production patch is already committed at
../nexus-import-require-module-alias/integration.patch; it remains unapplied.
These are test-only overlays, not a claim that the branch's ordinary shared driver
can discover .a without integration work.

Commands, with every test's output sent directly to a log file:

```
ADAMIC_TYPESCRIPT_SOURCE=/tmp/lint-wave1-14-typescript-pinned python3 stage1/cohere/lint/rules/typescript-eslint-no-duplicate-enum-values/validate.py > /tmp/lint-wave1-14-next-run.log 2>&1
# Runs go test -overlay=<printed overlay> ./stage1/cohere/lint -run '^TestWave14Next' -count=1 -v -timeout=20m
GOFLAGS=-overlay=<printed overlay> bash cloud/setup.sh > <owned>/evidence/setup.log 2>&1
go vet -overlay=<printed overlay> ./... > <owned>/evidence/vet.log 2>&1
go test ./internal/oracle -run '^TestTheOracleCatchesOneByte$' -count=1 -timeout=10m > <owned>/evidence/oracle.log 2>&1
# From cohere/:
go test ./internal/lint/rules/typescript -run '^TestNoConfusingNonNullAssertion' -count=1 -v -timeout=10m > <owned>/evidence/confusing-upstream.log 2>&1
```

Results: parity, corpora, both mutants, both throughput checks and three blocker
probes PASS in 150.735s. Vet exits 0 with an empty log. Filtered external oracle
PASS 0.275s. Original confusing-assertion Go tests PASS 0.008s. Setup retry reports
Go/clang/Node/submodules ready 0s, cache warm 15s, done 15s on 5 processors, 17.6 GB.
The full repository gate was not run. Ordinary production .a registration and
the blocked rule's complete suggestion contract remain integration work.
