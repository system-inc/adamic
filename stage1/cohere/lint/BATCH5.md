# Syntax-only lint batch 5

Branch `codex/stage1-lint-batch5`, cut from `origin/main` at
`5d4c801`. The final ten names are listed below. The first claim commit
`406910f` reserved nine of them plus `default-case-last`; claim update `04877da`
replaced that duplicate before the replacement implementation:

1. `@typescript-eslint/no-dynamic-delete`
2. `@typescript-eslint/no-import-type-side-effects`
3. `@typescript-eslint/no-misused-new`
4. `@typescript-eslint/no-this-alias`
5. `@typescript-eslint/prefer-as-const`
6. `@typescript-eslint/no-confusing-non-null-assertion`
7. `@typescript-eslint/no-extra-non-null-assertion`
8. `@typescript-eslint/no-duplicate-enum-values`
9. `@typescript-eslint/no-explicit-any`
10. `@typescript-eslint/no-useless-empty-export`

Each rule lives in its own file with a one-line registration. The oracle
is unmodified Go Cohere at the submodule pin, compared byte for byte with the
same Adamic source running on Node and sanitized native, including complete
finding and repair payloads. A successfully executing mutant per family must
be caught on both backends. Parser or compiler gaps will have proving programs
and will be excluded explicitly from agreeing-byte totals.

Original claim check: `origin/codex/stage1-lint-batch2` at `c4373c0` and batch3 at
`fa9781c` contain none of these ten names in implementations or batch claims.
At the original claim check, the remote had no `codex/stage1-lint-batch4` ref; that
absence is observed, not evidence about unpublished work. The branch tips were
refreshed before each rule. The later published overlap was replaced as recorded
below. Frequency-ranking logs listing unimplemented rules
are not treated as claims.

The first two commits contain reservations only, with no implementation or
parity claim. The completed implementation and evidence follow below.

## Updated claim before replacement implementation

Batch 4 appeared during final refresh at `d486b03` and holds
`default-case-last`. That original tenth reservation is cancelled; no copy of
its implementation will ship in batch 5. The replacement tenth name is
`@typescript-eslint/no-useless-empty-export`, checked against published batch 2,
batch 3 and batch 4 before this claim update. None holds it. A considered
alternative, `no-unnecessary-type-constraint`, was rejected because batch 2
already holds it. The nine TypeScript rules originally reserved remain unchanged.
Claim update `04877da` was pushed before writing the replacement implementation.

## Completed implementation

The claim-only first commit was `406910f`, pushed before source implementation.
This branch is deliberately main's five-rule baseline plus the ten claims,
not a merge of scanner/batch 2/batch 3. Each new rule has its own root file;
`batch5_registry.ts` supplies one call per rule in Go listener order. All
changes are inside `stage1/cohere/lint/`.

`Batch5Context` borrows syntax indexes, source and scanner, retaining ancestry as
numbers. Findings now distinguish finding ranges from edit ranges and retain all
automatic edits and suggestion IDs, descriptions and edit lists. The repair
loop uses the earlier batch 3 machinery for deterministic ordering, reparsing,
and bounded iteration. Settings accept the actual decoded Go options for
`no-this-alias` and `no-explicit-any`, not ESLint option syntax.

Pinned oracles and corpus:

- Cohere `715ba94f3608a6500086b1076ce5cb7e51b836db`; unchanged rule implementations,
  `report.Write` human diagnostics and `edit.FixText` final source.
- TypeScript v6.0.3 corpus `050880ce59e30b356b686bd3144efe24f875ebc8` in scratch.
- The same TS port on Node 24.19.0 and Adamic native built with clang 20.1.8,
  ASan/UBSan, `ASAN_OPTIONS=detect_leaks=1`, `UBSAN_OPTIONS=halt_on_error=1`.
- Go 1.27.1. No compiler, runtime, scanner or parser code was changed.

The upstream capture overlay records every harness `Run`, preserving the
original assertions. Uniqueness includes source, rule, decoded options and file
name, retaining TypeScript and JavaScript extension behavior.

| New rule | Captured cases | Oracle-held details |
| --- | ---: | --- |
| no-dynamic-delete | 42 | parentheses, signed/static keys, optional access, key range |
| no-import-type-side-effects | 26 | top-level phase, alias source name, all removals and insertion |
| no-misused-new | 45 | interface construct keyword, constructor member, class signature return name |
| no-this-alias | 40 | assignment wrappers, destructuring, allowed names, TS-only selection |
| prefer-as-const | 69 | cooked values, annotation removal, initializer suffix, no-fix patterns |
| no-confusing-non-null-assertion | 28 | operator messages, suggestion order, both wrap edits |
| no-extra-non-null-assertion | 19 | parenthesis ancestry, optional receiver, bang removal |
| no-duplicate-enum-values | 57 | separate value tables, first numeric versus previous string range |
| no-explicit-any | 207 | unbounded rest ancestry, opt-in unknown fix, file extensions |
| no-useless-empty-export | 64 | top-level module markers, export-equals distinction, declaration-file exemption |

This is 598 new cases plus 217 baseline cases. Two illegal method-body inputs
from `no-explicit-any` recover only in Go and are explicitly proved as parser
limits, with exit 70 on both port backends. The 813 supported upstream cases,
24 baseline generated runs and 131 batch 5 generated runs total **968 agreeing
cases**. Generated cases cover Unicode byte columns, CRLF, comments, cooked
escapes and numeric spellings, assignment wrappers/operators, triple duplicates,
rest-type nesting and composed repairs. A separate full corpus compares **198
files**, including every compiler `.ts` and stage1 `.ts`; precisely the two new
invalid-source proving programs are tested separately, not counted as agreement.

Final byte totals (paths are test-temporary, so totals can differ across runs):

```
TestBatch5Controls:            75,283 bytes identical
TestRulesAgree:               414,115 bytes identical
TestCompilerAndStage1Agree:   12,640,839 bytes identical
```

## Mutants and checks

`TestBatch5Mutants` replaces the selected rule constant with an omitted rule in
that rule's own scratch file, once for each of the ten families. Each mutant
must compile and execute successfully on Node and sanitized native, then differ
from the matching Go control. All ten were caught on both backends: the complete
finding disappears, and fix-producing families also change the final source.
This checks missing family behavior; it is not a claim that every internal
branch was independently mutated.

Three additional complete-payload mutants were caught on both backends:

1. Skip the first additional automatic edit. The import control loses the second
   specifier's removal; `extra-fix 22 27` is missing, and fixed source differs.
2. Drop the second suggestion. The `in` control loses `wrapUpLeft` and its two
   insertion edits while retaining the first removal suggestion.
3. Promote suggestions to automatic fixes. The confusing assertion control
   changes repair category and removes assertions without consent.

The original three mutants were also rerun: apply eqeqeq's suggestion as a fix,
report empty function bodies, and invert duplicate-case membership. These add
three successful-execution catches on both backends, for **16 lint mutants**.
The compiler oracle's one-byte guard passed separately.

Final commands, run from the repo after sourcing `/workspace/adamic-tools/env.sh`;
all test output goes to regular log files. Final checks use the three runs below
rather than another complete repository gate:

```
ASAN_OPTIONS=detect_leaks=1 UBSAN_OPTIONS=halt_on_error=1 \
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3 \
go test ./stage1/cohere/lint \
  -run 'Test(Batch5Controls|RulesAgree|CompilerAndStage1Agree|Batch5RecoveryGaps|NestedConstructorGap)$' \
  -count=1 -timeout=15m -v > /tmp/lint-batch5-parity-final.log 2>&1

ASAN_OPTIONS=detect_leaks=1 UBSAN_OPTIONS=halt_on_error=1 \
go test ./stage1/cohere/lint -run '^TestBatch5Mutants$' \
  -count=1 -timeout=10m -v > /tmp/lint-batch5-family-mutants-final.log 2>&1

ASAN_OPTIONS=detect_leaks=1 UBSAN_OPTIONS=halt_on_error=1 \
go test ./stage1/cohere/lint -run '^Test(Batch5RepairPayloadMutants|Mutants)$' \
  -count=1 -timeout=10m -v > /tmp/lint-batch5-payload-mutants-final.log 2>&1

go test ./internal/oracle \
  -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(classes|collections|maps_and_text|optional|regexp|sort)[.]a$' \
  -count=1 -timeout=10m -v > /tmp/lint-batch5-oracle-filter.log 2>&1

go test ./internal/oracle \
  -run 'Test(TheOracleCatchesOneByte|NativeAgreesWithNode/internal/oracle/testdata/(classes|collections|maps_and_text|optional|regexp|sort)[.]a)' \
  -count=1 -timeout=10m -v > /tmp/lint-batch5-oracle.log 2>&1

go vet ./... > /tmp/lint-batch5-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint > /tmp/lint-batch5-gofmt.log
/workspace/scratch/cohere --no-fix --no-cache stage1/cohere/lint/*.ts \
  > /tmp/lint-batch5-source-gate.log 2>&1
git diff --check > /tmp/lint-batch5-diff-check.log
```

The filtered compiler run passed six fixtures on source Node, native with
sanitizers, and the JavaScript backend: classes, collections, maps_and_text,
maybe_collections, regexp and timsort. The second filter ran the one-byte guard.
Worker oracle caching follows CLAUDE.md; integration's uncached full gate was
not run. Vet, gofmt and final diff checks have empty output. `.gitattributes` preserves
raw evidence whitespace rather than changing oracle bytes. Cohere reports 20 checked
files, 100% Adamic-ready, with no findings.

Final package results were PASS: parity and gap checks 107.571s, ten family
mutants 159.289s, payload and baseline mutants 95.446s. The final logs are copied
under `evidence/batch5/`. The initial broader owned
run launched before the two invalid proving programs were separated from the
ordinary corpus; that corpus check failed its Go parse guard. The final parity
run above checks their explicit refusal separately and passes. No parser result
was silently accepted or included in the agreeing-byte total.

## Setup, claim discipline and limits

`bash cloud/setup.sh` succeeded. Its exact timing lines were:

```
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (103s)
setup: done in 103s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
nproc: 5
```

`testdata/check_batch5_claims.py` refreshed the published batch 2, batch 3 and
batch 4 tips immediately before each rule's first source write. The original ten pre-write
checks, replacement checks and final all-ten refresh are in `batch5_tip_checks.json`.
The original batch 4 ref was absent; the final refresh found it at `d486b03`.
Its `default-case-last` was dropped and its replacement claimed and pushed in
`04877da` before implementation. The nine other claims remained unheld by any
published batch. A rejected `no-unnecessary-type-constraint` candidate already
belonged to batch 2 and was never claimed or implemented. Final all-ten checks
used batch 2 `c4373c0`, batch 3 `fa9781c`, batch 4 `d486b03` and found no overlap.
Ranking logs listing unimplemented rule names are not classified as claims.

The two recovered method-body gaps and inherited constructor-proof gap have
proving programs; details are in [GAPS.md](GAPS.md). No binder/type-checker rules,
CLI config/suppression integration, general JSX or malformed-source recovery,
non-UTF-8 input, or arbitrary invalid third-party fixes are claimed. The copied
edit engine's competing and oscillating fix machinery has no new control in
these ten rules. No throughput measurement was requested or taken for batch 5;
the historical five-rule rates in REPORT.md are not rates for this batch.
