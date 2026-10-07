# Batch 6: inventory syntax candidates 1 through 10

Branch `codex/stage1-lint-batch6`, cut from fetched `origin/main`
`5d4c8012a0877094134e6c6bac367ff68f9313e8`. Inventory pin:
`origin/codex/lint-inventory` at `73ac2eb`. Selection is the first ten rows,
in inventory order, with wave `syntax ready for AST/API adaptation` after
excluding statuses containing `ported with` or `partial port:`. The resulting
queue contains thirty rows. Candidates 11 through 30 are outside this unit.
[batch6-selection.json](batch6-selection.json) preserves the selection and
upstream test locations.

The Go answer is unmodified cohere
`715ba94f3608a6500086b1076ce5cb7e51b836db`. The compiler corpus is TypeScript
v6.0.3, `050880ce59e30b356b686bd3144efe24f875ebc8`, under `src/compiler`.
The cohere submodule remains clean. No compiler, lowerer or runtime file changed.

## What was built

Each rule has its own TypeScript file and one dispatch call in `registry.ts`.
Seven TypeScript visitors reuse the independently verified implementations from
`63782c53678711717f9fd4f12762ce3d103b9a3d`; the confusing assertion, module
variable and default-case visitors were added here. Current main's original
five visitors remain, making fifteen active rules in the default runner.

| Position | Rule                                                              | Distinct captured upstream cases |
| -------- | ----------------------------------------------------------------- | -------------------------------: |
| 1        | `@next/next/no-assign-module-variable`                            |                               13 |
| 2        | `@typescript-eslint/default-param-last`                           |                              117 |
| 3        | `@typescript-eslint/no-confusing-non-null-assertion`              |                               28 |
| 4        | `@typescript-eslint/no-duplicate-enum-values`                     |                               57 |
| 5        | `@typescript-eslint/no-dynamic-delete`                            |                               42 |
| 6        | `@typescript-eslint/no-extra-non-null-assertion`                  |                               43 |
| 7        | `@typescript-eslint/no-misused-new`                               |                               45 |
| 8        | `@typescript-eslint/no-unnecessary-parameter-property-assignment` |                               92 |
| 9        | `@typescript-eslint/prefer-as-const`                              |                               69 |
| 10       | `default-case-last`                                               |                               38 |

The adapter indexes parents before visitors run, keeps finding and edit spans
separate, and represents every suggestion's ID, description and ordered edits.
The confusing assertion rule's `in` and `instanceof` cases offer removal followed
by a two-edit wrap suggestion. Suggestions are never applied automatically.
The fixer handles additional edits, overlap and identity rejections, reparsing,
and iteration to convergence. The driver builds line and UTF-8 offset tables
once rather than rescanning the source prefix for every finding.

## Setup and machine

`bash cloud/setup.sh > /tmp/batch6-setup.log 2>&1` succeeded. Each toolchain
command sourced `/workspace/adamic-tools/env.sh`. `nproc` printed `5`.

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (22s)
setup: done in 22s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Go 1.27.1, clang 20.1.8, Node 24.19.0, Linux x86_64, AMD EPYC 9V74.
The container exposes five processors with a four-CPU cgroup quota.

## Commands and observations

All test output went directly to log files. Every command below exited zero.
The full repository `go test ./...` was not run; validation covers the touched
lint package's tests and an external oracle mutant, plus repository-wide vet.

```bash
go test -count=1 -run 'TestRulesAgree|TestBatch6Mutants' -timeout 20m -v ./stage1/cohere/lint > /tmp/batch6-rule-gate.log 2>&1
go test -count=1 -run 'TestRulesAgree|TestBatch6CountCheck|TestMutants|TestNestedConstructorGap' -timeout 20m -v ./stage1/cohere/lint > /tmp/batch6-fixture-gate.log 2>&1
go test -count=1 -run TestTheOracleCatchesOneByte -timeout 5m -v ./internal/oracle > /tmp/batch6-filtered-oracle.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/batch4-typescript ADAMIC_LINT_BENCH=1 go test -count=1 -run 'TestCompilerAndStage1Agree|TestThroughput' -timeout 20m -v ./stage1/cohere/lint > /tmp/batch6-final-corpus-bench.log 2>&1
go vet ./... > /tmp/batch6-vet.log 2>&1
gofmt -l cmd internal stage1/cohere/lint > /tmp/batch6-gofmt.log
```

Observed:

- Twelve new rule/repair mutants passed their kill checks, 157.87 seconds total.
  Each compiled and executed on Node and sanitized native before its bytes
  disagreed with the Go answer. The positive control matched 16,626 bytes.
- The final fixture gate captured 761 distinct source/rule/options/extension
  combinations: 544 for this unit and 217 for the original five rules. One
  exact fixture is a checked refusal below. Together with 34 generated runs,
  794 supported cases matched 327,071 output bytes. This includes human
  findings, ranges, all suggestion payloads, extra automatic edits, rejection
  records and entire fixed sources.
- Final compiler/stage1 parity covered 195 source files and 12,517,773 identical
  output bytes on Go, Node and sanitized native. The corpus includes the new
  rule implementations themselves. Sanitizers and leak checking reported no
  errors. The test passed in 27.76 seconds.
- The count mutant printed `43` where Go printed `33`, on both Node and native,
  while ordinary output remained identical. Its check passed in 13.90 seconds.
- All three inherited lint mutants were caught on Node and native, 36.95 seconds.
- The existing nested-constructor gap test passed. The external oracle's
  one-byte output mutant was caught, 4.83 seconds, with native and Node cache
  misses, so the observations were freshly executed.
- Repository-wide `go vet ./...` and the Go formatting check printed nothing.

The earlier source-corpus run also passed; the final run above supersedes its
byte count because unused adapter methods were removed. An initial oracle
build failed on incorrect rejection field names; those names were corrected
before the successful gates. The first complete fixture attempt exposed the
recovery limit below, which is now checked explicitly.

Cohere formatted the changed files. The final scoped `--no-fix --no-cache`
check names the seventeen lint TypeScript files, README, this report and the
selection JSON explicitly. The directory-wide format check also examined the
unchanged historical REPORT.md and reported its pre-existing formatting; it
was not rewritten. Repository-wide `--no-fix --no-cache --no-format` exited
one with 491 existing findings and 31 files needing fixes, all outside the
changed lint sources. Those unrelated sources were not edited.

## Every mutant and what caught it

All twelve rows below were killed by byte parity on **both Node and sanitized
native**, after successful compilation and execution. The two suggestion
mutants change only suggestion metadata or edits; findings and automatic fixed
text alone could not catch them.

| Mutant                                             | First distinguishing observation                  |
| -------------------------------------------------- | ------------------------------------------------- |
| Match module name `other` instead of `module`      | Missing whole-statement finding                   |
| Stop default-parameter loop at zero                | Missing defaulted parameter finding               |
| Invert the confusing trailing-bang guard           | Wrong confusing assertion finding                 |
| Anchor enum duplicates at the later initializer    | Wrong human position and finding span             |
| Accept an identifier as a static delete key        | Missing dynamic-key finding                       |
| Move redundant-bang deletion to an empty range     | Fix range `15 15` instead of `14 15`              |
| Widen construct keyword from three to four units   | Finding range `26 30` instead of `26 29`          |
| Remove the assignment suggestion's extra end unit  | Suggestion edit range `54 64` instead of `54 65`  |
| Insert `as never` instead of `as const`            | Wrong additional fix payload and rewritten source |
| Reverse default-clause position test               | Wrong clause finding                              |
| Repeat removal instead of offering the second wrap | Wrong second suggestion ID and description        |
| Replace wrap's closing `)` with `]`                | Wrong second suggestion edit payload              |

Additional mutants: the count-only `+1` mutant was caught exclusively by count
parity after ordinary parity passed; the inherited suggestion-to-fix,
empty-function-body exemption and duplicate-case membership mutants were
caught by ordinary byte parity on both backends; the external oracle's one-byte
mutant was caught by its own output comparison.

## Findings per second

Five fresh-process interleaved count rounds over 77 compiler files, all fifteen
rules, 481 findings per run. Before timing, release native, Node and Go matched
11,631,597 bytes of complete findings and repairs. Native used clang `-O2`
without sanitizers; correctness gates used ASan, UBSan and leak checking.
Startup, reads, scanning, parsing and visiting are included; formatting and
fixing are excluded in count mode. Compilation is excluded.

| Backend             | Best of five, seconds | Findings/second |
| ------------------- | --------------------: | --------------: |
| Go cohere oracle    |              0.311383 |        1,544.72 |
| Adamic native       |              1.632241 |          294.69 |
| Same source on Node |              0.919524 |          523.10 |

Load averages before: `1.04 0.86 0.46`; after: `1.26 0.92 0.49`. On this
measurement native took 5.24 times Go's time and 1.78 times Node's. This is an
observation of this fifteen-rule corpus runner, not a measurement of the full
cohere CLI or evidence of a speedup.

## Explicit limits

This upstream case is accepted by Go cohere but needs recovery not represented
by the stage1 parser:

```typescript
class Foo {
    constructor(public { a }: { a: string }) {
        this.a = a;
    }
}
```

Both Node and native exit 70 with
`parser slice expected CloseParenToken, got OpenBraceToken at 33`. The test
pins the exact source and exact refusal, separately executes Go, and reports
this case as a limit rather than counting it as parity. There is no general
parse-error skip.

The tests prove parity on the named corpora, not all possible TypeScript input.
Go input parse diagnostics remain refused. This unit implements no type
checker, binder, CLI configuration, file selection or suppression layer. The
fixer reparses with the stage1 parser; it does not port cohere's general
parse-rejection recovery. Candidates 11 through 30 and the other lint batches
were not changed. Commit identity is the published branch tip recorded by
`git log origin/main..codex/stage1-lint-batch6` and the accompanying summary.
