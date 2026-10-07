# Comment helper unit report

Branch: `codex/lint-helpers`, continuing the first helper unit. Claim commit `29990b47913aa77d542045194b512484af995d29` was pushed before implementation. Previous implementation/base: `5d13f5baaecaf11d4ea62de693426f69a1f41bba`. The implementation commit containing this report is reported in the final handoff. Pinned Go cohere: `715ba94f3608a6500086b1076ce5cb7e51b836db`.

## Built and measured

Five separate helper files provide `canBeginAt`, `collectListInteriors`, `sortByPosition`, `All`, and `ForFile`. A sixth source file is the immutable comment data model; the seventh is the comparison/adapter driver. The per-file cache must be shared by all rules of that immutable file. No parser, compiler, existing rule implementation, original inventory, or first-unit helper source was changed.

The five-symbol scan/cache bundle serves **24 consumers** and completes the remaining helper dependency set for **16 more rules**. In prerequisite order, the guard, interior anchors, sorting and uncached scan each unblock **0 alone**; `ForFile` completes **16 with those prerequisites**. Cumulative: **62 helper-ready candidates, 136 still helper-blocked, denominator 198**. The named list and residual dependencies are in `readiness.json` and the parent `HELPERS.md`.

These are the same conditional dependency counts as the previous unit. They assume a common correct AST/reporting adapter. No candidate is claimed to be an implemented or fully integrated rule. The next small single-function step, `imports.BindingsOf`, would add four after this bundle; the JSX and imports package bundles project eight and seven respectively. Those helpers are not implemented here.

## Commands and observations

Every shell sourced `/workspace/adamic-tools/env.sh`. Test output went to files, not pipes. The final successful logs are committed under `evidence/`.

| Command | Observed output |
|---|---|
| `git fetch origin` | exit 0 |
| `bash cloud/setup.sh` | Go ready 0s; clang ready 1s; Node ready 1s; submodules ready 1s; build cache warm 31s; done in 31s |
| `nproc` | 5; setup reports cgroup `cpu.max: 400000 100000`, 17.6 GB |
| `python3 stage1/cohere/lint/helpers/comments/testdata/capture.py` | 4,513 unique inputs from all 24 consumers; four original Go family test commands passed |
| `go test -count=1 -timeout 15m -v ./stage1/cohere/lint/helpers/comments` | PASS, 191.104s; 47 independent-parser boundaries, 4,513 consumer sources, five helper mutants, four explicit TSX refusals, and the compiling adapter-guard mutant |
| `go test -count=1 -timeout 10m -v ./stage1/cohere/lint/helpers/comments -run 'TestJsxParserGapIsExplicit|TestJsxAdapterGuardMutant'` | PASS, 35.857s after factoring the shared refusal predicate; four valid Go JSX sources refuse identically on Node and native; the guard mutant is rejected by that same predicate |
| `go test -count=1 -timeout 10m -v ./stage1/cohere/lint/helpers/comments -run TestConsumerCommentHelpers` | PASS, 52.363s after matching the original harness's TS/TSX/JS/JSX mode selection exactly; 32,770 lines agree byte-for-byte on Go, Node and sanitized native |
| `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 10m -v ./internal/oracle -run '^TestInputAgreesWithNode$'` | PASS, 1.671s; all six input fixtures; probe hits 0, misses 6 |
| `go vet ./stage1/cohere/lint/helpers/...` | exit 0, no output |
| From `cohere`: `go test -count=1 -v ./internal/lint/ecmascript/comments` | PASS, 0.016s; private real-tree guard corpus test skipped because its `/Users/kirkouimet/...` corpus is unavailable |
| Go cohere `--no-fix --no-format stage1/cohere/lint/helpers/comments/*.ts` | exit 0; 276 rules, 7 owned files checked; no findings |
| Go cohere `--no-fix --format-only stage1/cohere/lint/helpers/comments/*.ts` | exit 0; no formatting changes required |
| `git diff --check`; owned Go files formatted with `gofmt` | exit 0, no output |

The initial external-oracle filter selected no fixtures; it was replaced by the successful six-fixture command above and is not counted as evidence. An intermediate source import cycle and unsupported `startsWith` offset form were removed before native validation. Cohere's caller-mutation findings were resolved by returning owned interior anchors. The stable sort deliberately keeps Go's in-place contract and explains its two intentional caller writes with lint directives. Final lint is clean.

The consumer comparison uses Go's actual AST traversal geometry as the common adapter input: kind, start, end and child indexes, converted to UTF-16. It does not export comment answers into the TypeScript driver. The unmodified Go comment helpers separately produce the expected results. Both cached and uncached ranges, full comment text, block kind, line/byte-column metadata, source ordering, guard answers and shared cache identity are checked. The 47 boundary cases also run through independently parsed Adamic trees and produce 390 matching output lines. Native builds use ASan and UBSan; successful runs have no stderr or leak reports.

## Compiling semantic mutants

Every mutant built successfully and ran its witness. No compile failure, `-Werror` rejection, crash or sanitizer failure is counted as a caught semantic mutant.

| Source | Mutation | What caught it in the final full-suite log |
|---|---|---|
| `can_begin_at.ts` | Remove the shebang exception | output line 251: cached comment count 0 versus Go 1 |
| `collect_list_interiors.ts` | Stop offering the first opening delimiter's interior | output line 3: cached comment count 0 versus Go 1 |
| `sort_by_position.ts` | Reverse the insertion-sort comparison | output line 60: `/* f */` at byte 47 before `/* t */` at byte 6; Go requires source order |
| `all.ts` | Remove node-end anchors | output line 201: cached comment count 0 versus Go 1 |
| `for_file.ts` | Recompute on every `get` | output line 2: `cache 0 0 0` versus Go `cache 1 0 0`; the returned arrays no longer share identity |
| Adapter `main.ts` | Remove the explicit TSX refusal | the known unsupported JSX source finishes with exit 0, whereas the actual shared refusal predicate requires exit 70 and the exact NotYet marker |

The five helper mutants took 88.54s together in the full successful run. The adapter mutant took 18.00s there and was subsequently rechecked through the shared predicate in the focused test.

## Observed parser defect and limits

Before adding the adapter guard, the underlying parser returned exit 0 for `const x = <div>{/* c */}</div>;`. Its printed tree contains a `TypeAssertionExpression`, `ObjectLiteralExpression`, `LessThanToken`, and `RegularExpressionLiteral` instead of JSX. The other three JSX boundary sources refused. Those are observations, preserved in `parser-probe.log` and `wrong-tree.log`.

The inference is that this parser cannot be used as a JSX adapter yet. The owned independent-parser comparison driver now rejects unadapted TSX sources with `NotYet: stage-1 JSX parser adapter`. A validated AST geometry adapter continues to exercise all JSX consumer inputs successfully. The parser itself was not changed; its owner must still provide JSX support. In particular, `react/jsx-curly-brace-presence` remains conditional on that integration. Correct source kind, child geometry and byte-to-UTF-16 reporting conversion remain part of the common adapter contract. `Comment` ranges are UTF-8 bytes; the existing lint `Finding` constructor and string slicing use UTF-16.

No rule implementation, rule-local fix/default behavior, independent full consumer parser parity, checker bridge, `IsJsDoc`, `ContentLines`, `LeadingRunFor`, directive parser or concurrent cache-sharing contract was ported here. The cache is intended to be owned by one file worker, shared across its rules. No helper throughput or peak-memory improvement is claimed. The full repository gate was not run; validation covered the owned helper package, vet of helper packages, original Go consumers/helper tests, final owned-file cohere checks and six uncached external input fixtures. The missing private guard corpus was not replaced by a claim of real-tree coverage.

## Fixture hashes

- `testdata/consumers.json`: `18e4f1c954ff941e421a3158b494c11778ce94b1b631d444927d5462d698d7eb`
- `testdata/witnesses.json`: `4d0ccb9b32e5a5d95f172d513b64f85e11525440f92aada72741e34e3a04138f`
- `testdata/parser-gaps.json`: `1adc9938f10c7841f72eb282a34f94ca425462972859cdac72a48035be249cf7`
