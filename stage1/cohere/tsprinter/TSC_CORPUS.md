# Deliberate tsc-driver corpus

Branch: `stage1-format/tsprinter-tsc-corpus`. Base: `bfa0bfec9ab6f49c81e22984d3fca4ee21a472e8`.
The integration exclusion was absent at branch creation. No `internal/` implementation changed.

## Every original disagreement

[TestTSCCorpusAgreement](tsc_corpus_test.go) runs the existing Go selectors against only the
300 tracked files under `stage3/drivers/tsc/corpus`, then compares every selected fragment on
Node, sanitized native and the JavaScript backend. It reports all differences, rather than stopping
at the first. The baseline Node audit found **seven differences: two NotYet answers, five differing
texts, zero errors**. [results/tsc-corpus-before.json](results/tsc-corpus-before.json) records each
source, file and byte-exact port/Go answer.

Paths below are relative to `stage3/drivers/tsc/corpus`; offsets identify the original AST visits.

| Cause | File and offsets | Original port answer | Go answer | Repair |
|---|---|---|---|---|
| Yield literal lookahead | `080_builtinIterator/builtinIterator.ts:422` | NotYet `expression-file` | Formats yield statement | Same-line identifier/keyword/literal lookahead in shared parser |
| Standalone leading block | `146_noImplicitAnyIndexing/noImplicitAnyIndexing.ts:519,577` | Parenthesized object receiver | Block followed by array statement | Share complete-program composition for a leading brace |
| Bodyless declaration | `065_callOverloads2/callOverloads2.ts:211` | NotYet `FunctionDeclaration` | Formats declaration | Validate all signature children, print terminating semicolon |
| Ungrouped statement ternary | `114_truthinessPromiseCoercion/truthinessPromiseCoercion.ts:314,350,388` | Unconditionally broken layout | Fits on one line | Group conditionals with no parent |

After repair, **1,729 expression fragments and 874 statement fragments** match Go on all three
port builds. Generated boundary cases also cover literal forms, ordinary calls to yield, newline
boundaries, declaration parameters and nested conditional layouts. The shared parser regression
[yield_test.go](../../typescript/parser/yield_test.go) holds 17 AST cases to Go; the existing
compiler-expression and generated-expression parser tests also pass.

These are selected supported fragments, not 300 formatted complete files. Both Go selectors refuse
`239_expressionWithJSDocTypeArguments/expressionWithJSDocTypeArguments.ts` with `Type expected.`.
The statement selector accepts zero complete files from this root. Existing unsupported types,
comments, declarations and control-flow families remain explicit in [GAPS.md](GAPS.md).
The focused coverage reports are [results/tsc-expressions-coverage.json](results/tsc-expressions-coverage.json)
and [results/tsc-statements-coverage.json](results/tsc-statements-coverage.json). Their `cases`
field includes the selector's generated cases; the focused test filters to originating-file labels.

## Go and external Prettier outcomes

The port's byte comparison to Go has no exceptions. The broader gate exposed external outcomes
that differ from Go. Both npm Prettier and cohere's embedded Prettier are version **3.9.6**.
An exhaustive comparison of **169,887 expression cases and 57,799 statement cases** found:

| Pattern | Expression cases | Statement cases | Files | Smallest input for routing |
|---|---:|---:|---|---|
| Await operator versus identifier call | 2 | 1 | `114_truthinessPromiseCoercion/truthinessPromiseCoercion.ts`, `229_awaitCallExpressionInSyncFunction/awaitCallExpressionInSyncFunction.ts` | `await(x)` |
| Invalid update target accepted by Go, refused by both Prettiers | 13 | 13 | `170_decrementAndIncrementOperators/decrementAndIncrementOperators.ts` | `1++` |

Observation: Go prints the standalone await text as a call; both Prettiers print an await operator.
Inference: extracting an expression from its enclosing function loses context, and the standalone
parsers choose differently. The synchronous-function fixture also deliberately calls an identifier
named await. No port change makes these cases disagree with Go.

Observation: both Prettiers throw an invalid-left-hand-side parser error for the update targets;
Go formats them. These inputs are negative TypeScript tests. The port preserves Go's formatting
rather than rejecting a case Go accepts. [results/tsc-smallest-upstream.json](results/tsc-smallest-upstream.json) verifies the smallest inputs
against Go and both libraries. These inputs and the exact cases are ready for routing upstream;
this unit did not contact a maintainer.

[testdata/tsc-upstream-differences.json](testdata/tsc-upstream-differences.json) pins every one of
these **29 family-labelled outcomes**: source, originating file/offset, Go bytes, both external byte
strings or both exact parser errors. External scripts emit their actual text or error. The Go test
compares that output to the separate pin; it never substitutes Go text. Every unpinned outcome must
match Go, and every pinned case must still be present. `TestTSCCorpusUpstreamDifferences` independently
runs all 29 outcomes against both libraries.

The leading-brace example is another extraction boundary, not a new external disagreement:
smallest input `{}[0]` is a block followed by an array statement in a standalone program. Go's
interpretation is preserved even though the original AST visit was an object receiver.

## Named tracked roots

[corpus_test.go](corpus_test.go) uses `git ls-files`, not an on-disk recursive walk. Both agreement
tests log each named root and fail when it is missing or has no tracked `.ts` files. Untracked files
and tracked files outside these roots are ignored. Individual missing tracked files are still errors
when the Go selector reads them.

| Checkout / root | Tracked `.ts` files |
|---|---:|
| Adamic / `bench` | 6 |
| Adamic / `bridge` | 1 |
| Adamic / `cmd` | 5 |
| Adamic / `internal` | 18 |
| Adamic / `stage1` | 476 |
| Adamic / `stage3/drivers/tsc/corpus` | 300 |
| TypeScript / `src/compiler` | 77 |
| Total | **883** |

TypeScript is pinned to `050880ce59e30b356b686bd3144efe24f875ebc8` through
`ADAMIC_TYPESCRIPT_SOURCE`; npm Prettier is provided by `ADAMIC_TS_PRETTIER`.
Both were provisioned on this machine, so no input skip is needed.

## Scratch mutants

[testdata/run_tsc_mutants.py](testdata/run_tsc_mutants.py) copies the port and parser into a temporary
directory and overlays only the test's entry path. It never mutates committed source. Each undone
repair fails the focused agreement test on **Node, native and backend**, naming the original file:

| Mutant | Named failing file under the tsc root |
|---|---|
| `yield-literal` restores old identifier-only lookahead | `080_builtinIterator/builtinIterator.ts` |
| `leading-block` disables complete-program dispatch | `146_noImplicitAnyIndexing/noImplicitAnyIndexing.ts` |
| `bodyless-declaration` restores body-assuming validation | `065_callOverloads2/callOverloads2.ts` |
| `statement-conditional` removes root conditional grouping | `114_truthinessPromiseCoercion/truthinessPromiseCoercion.ts` |

All four mutated programs finish normally; the byte assertions catch them. Additional scratch
controls remove missing/empty-root checks and admit an untracked file; `TestTrackedCorpusRoots`
catches both. `--upstream-only` changes a text pin and an error pin; the actual external-output
comparisons fail naming `114_truthinessPromiseCoercion` and `170_decrementAndIncrementOperators`.
This proves that the external report is checked evidence, not an unchecked allow-list.

## Validation

Setup: `GOPROXY='https://proxy.golang.org|direct' bash cloud/setup.sh`, then source
`/workspace/adamic-tools/env.sh`. Timings: Go 0.024s; Node 0.024s; markdown 0.070s; submodules 0.078s;
clang 0.196s; build 42.233s; cache 42.336s; complete 42.364s. `nproc`: **5**, CPU quota: **4**.

Commands and retained log paths:

- Focused audit: `go test -v -count=1 -run '^TestTSCCorpusAgreement$' ./stage1/cohere/tsprinter`;
  `/tmp/tsprinter-tsc-before.log` records all seven initial failures;
  `/tmp/tsprinter-tsc-focus.log` records the repaired pass, 63.962s.
- Parser regressions: `go test -v -count=1 -run 'TestGeneratedExpressionsAgree|TestCompilerExpressionsAgree|TestYieldLookaheadAgrees' ./stage1/typescript/parser`;
  `/tmp/tsprinter-tsc-parser.log`: PASS, 41.915s, 1,676 generated cases, 77 compiler files and 17 lookahead cases.
- Scratch fixes: `python3 stage1/cohere/tsprinter/testdata/run_tsc_mutants.py`;
  `/tmp/tsprinter-tsc-mutants.log` and `/tmp/tsprinter-tsc-root-mutants.log`.
- External controls: same runner with `--upstream-only`;
  `/tmp/tsprinter-tsc-upstream-mutants.log`; all individual expected-failure logs under
  `/tmp/tsprinter-tsc-mutants/`.
- `go vet ./stage1/cohere/tsprinter ./stage1/typescript/parser`;
  `/tmp/tsprinter-tsc-vet-final.log`: exit 0; Go files are gofmt-clean.
- Whole package: `go test -json -count=1 -parallel=4 -timeout 30m ./stage1/cohere/tsprinter` with both pins;
  `/tmp/tsprinter-tsc-package-complete.jsonl`, machine observations in
  `/tmp/tsprinter-tsc-package-complete-machine.log`.

Whole package: **45 passed, 0 failed, 0 skipped** including subtests; **12 top-level tests passed**.
Skipped test names: none. Go package time: **883.834s**; command wall time: **885.906s**.
`nproc`: **5**, quota: **4 CPUs**. Load averages before: **1.38 / 1.60 / 1.68**; after:
**3.01 / 3.42 / 2.57**. All **29 existing mutants** were caught in this run, in addition to
the four new fix mutants and four additional scratch controls recorded above.
[results/tsc-validation.json](results/tsc-validation.json) retains the census and machine observations.

The first broader run passed the port-to-Go comparisons but failed the two external-library checks
before the upstream audit/pins were added. `/tmp/tsprinter-tsc-package.jsonl` preserves that result.
A subsequent run exposed a label-assumption bug in the new external harness when an older
upstream test supplied unlabeled cases. The lookup now accepts those cases without applying any
tsc outcome pin; `/tmp/tsprinter-tsc-upstream-compatibility.log` records both report tests passing.
That package run was stopped after the failure, before restarting the corrected whole package.
No compiler/runtime language gap blocked these four repairs. Unsupported printer families and the
one Go full-file parse refusal remain outside this increment's accepted-fragment claim.


## Main merge at 679af4df

Merge commit `de3fbba16c7bf8d474cf43191eaf71eb765fa8ff` has parents `eb3750d1` and main `679af4df`.
This is a merge, not a rebase. The two conflicts in `expressions_test.go` and
`statements_test.go` retain `printerCorpusFiles`. Integration's `cfa787fc2` exclusion
helper, its excluded-file census and both repository-wide walks are removed.
Its missing/empty-directory guarantee remains through `trackedRootFiles`: every named
root fails by name when missing or without tracked `.ts` files. The tsc corpus remains
an intentional 300-file root. The per-root table above is unchanged: **883 files**.
The printer and shared parser code are unchanged from `eb3750d1`.

`git merge-tree --write-tree HEAD 42caf641` exited 0 with no conflicts. The generated
merge tree `093a554e87a83f9ec4d1c0598054d87e4a10c2eb` retains the yield helper and its call.
This checks Git merge compatibility; the recovery branch is not part of this merge commit.

Both inputs were set: `ADAMIC_TYPESCRIPT_SOURCE` at `050880ce` and
`ADAMIC_TS_PRETTIER` at Prettier 3.9.6. Reruns:

- `go test -json -count=1 -parallel=4 -timeout 30m ./stage1/cohere/tsprinter`:
  **45 pass, 0 fail, 0 skip**, including subtests; all 12 top-level tests and 29 existing
  mutants passed. Package time **918.209s**, command wall time
  **920.283s**. Log: `/tmp/tsprinter-tsc-merge-package.jsonl`.
- `go test -json -count=1 -timeout 30m -run '^(TestYieldLookaheadAgrees|TestCompilerExpressionsAgree|TestGeneratedExpressionsAgree)$' ./stage1/typescript/parser`:
  **3 pass, 0 fail, 0 skip**, **40.621s**. This holds 17 yield
  cases, 77 compiler files and 1,676 generated expressions. Log:
  `/tmp/tsprinter-tsc-merge-parser.jsonl`.
- `go vet ./stage1/cohere/tsprinter ./stage1/typescript/parser`: exit 0;
  `/tmp/tsprinter-tsc-merge-vet.log`. gofmt output is empty in
  `/tmp/tsprinter-tsc-merge-gofmt.log`.

No skipped test names. `nproc`: 5; CPU quota: 4. Load before:
3.67 / 2.45 / 2.25; after: 2.75 / 3.38 / 2.75. Setup succeeded: Go 0.027s,
Node 0.029s, submodules 0.095s, markdown 0.099s, clang 0.215s, build 35.235s,
cache 35.341s, done 35.368s; `/tmp/tsprinter-tsc-refresh-setup.log`.
[results/tsc-main-merge-validation.json](results/tsc-main-merge-validation.json)
retains the counts, root census and merge evidence.
