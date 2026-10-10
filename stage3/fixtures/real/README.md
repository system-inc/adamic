Built: ten real tsc source extracts, each with a deterministic source-Node oracle and a-check header.
Base: origin/main 89ac4a8c1de0b02d95965be72f7f8bf1c92433d2; branch codex/stage3-real-fixtures.
Checks: ten fixtures, twenty Node/full-diagnostic comparisons; three Refused and seven NotYet.
Mutants: one stdout byte, one diagnostic byte, and a contradictory a-check header, each caught independently.
Limits: historical census counts, reduced dependencies, no compiler changes or native execution of these negative fixtures.

# Real compiler fixtures

The ranking comes from main's [latent report](../../census/latent/REPORT.md),
`main-unmerged`, measured at `b8fb957aa839a9e8cb0b54279dd9864fa317bd30`.
That report combines NotYet and Refused and explicitly measures a
checker-rejected program. Its counts below are historical priorities, not
new current-main census totals. Continue down that ranking to obtain ten
surviving families after checking landed representations and earlier refusals.
The current fixture compiler is the unedited topic base `89ac4a8c`.

| Historical count | Reason | Function and source |
| ---: | --- | --- |
| 588 | a type predicate | [isIdentifier](type-predicate/README.md), factory/nodeTests.ts:318 |
| 484 | the non-null assertion ! | [first](non-null-assertion/README.md), core.ts:1098 |
| 191 | a function without a body | [getDirectoryPath](function-without-body/README.md), path.ts:310 |
| 85 | a PrefixUnaryExpression on a value | [isPlainJsFile](prefix-unary-value/README.md), utilities.ts:997 |
| 80 | structural method call in a program with statics | [readFile callback](structural-method-statics/README.md), commandLineParser.ts:2239 |
| 66 | a function returning T \| undefined | [firstOrUndefined](generic-optional-return/README.md), core.ts:1083 |
| 56 | a value as a condition | [numberOfDirectorySeparators](value-condition/README.md), utilities.ts:10157 |
| 32 | a BinaryExpression with a value and a value | [getWatchOptionsNameMap](binary-value-value/README.md), commandLineParser.ts:2317 |
| 30 | a value of type any | [compareDataObjects](any-value/README.md), utilities.ts:8147 |
| 25 | a field of type boolean \| undefined | [getStartsOnNewLine](optional-boolean-field/README.md), factory/emitNode.ts:166 |

Every program has a source note, a first-line a-check header, its own
`expected.stdout`, and a [shared status entry](status.json) with the complete
current diagnostic, including line and column. Nine current reason strings
match their historical census strings exactly. Predicate proof has landed
partially: the open Node-to-Identifier contract still refuses, now with the
precise reason `a type predicate whose return is not proven (return expression
is not a trusted check on node)`. Both spellings are recorded; the old blanket
predicate diagnostic is not claimed to remain available.

# Skipped ranked rows

| Historical count | Row | Evidence and scope |
| ---: | --- | --- |
| 484 | reading SyntaxKind | Enum member reads are landed; enum control compiles. |
| 162 | enum | Numeric enums are landed; flags control compiles. |
| 154 | an EnumDeclaration | Same landed enum declaration representation. |
| 77 | an ExportDeclaration | Named exports are landed; export control compiles. |
| 60 | a NonNullExpression | Earlier refusal catches this syntax as the non-null assertion !; covered once by first. This is diagnostic precedence, not a claim that ! compiles. |
| 50 | a function returning T | Generic scalar return specialization is landed; the real identity function compiles for number and string. |
| 37 | reading CharacterCodes | Same landed enum representation; CharacterCodes member control compiles. |
| 30 | a value of type T | Ordinary generic parameter specialization is landed; the real equateValues function compiles with numeric parameters. The latent meter attempts unspecialized functions independently. This does not claim that every higher-order generic signature works. |

[landed-controls.json](logs/landed-controls.json) contains all six control
programs and compiler exits; enum controls cover both declaration and member
rows. The generic optional return remains separately reproducible, so it was
retained. Counts for unproven original predicates cannot be inferred from their
old blanket count. No full corpus recount was performed.

# Provenance and trimming

TypeScript 6.0.3 is pinned at
`050880ce59e30b356b686bd3144efe24f875ebc8`. A fresh
`bash stage3/apply.sh /tmp/real-fixtures-main-source` applied this main's
registered adaptations before extraction. Source lines refer to that
`src/compiler` tree. This distinction matters for first: strict indexed-access
adaptation inserts its non-null assertion, while upstream spells array[0].
[The fixture NOTICE](../NOTICE) applies to these excerpts.

The selected function bodies are retained. The directory fixture keeps its
string overload and drops its branded Path overload and long documentation.
The host fixture retains the original callback expression and adds the
contextual parameter type explicitly. Supporting declarations and helpers
are intentionally small; each fixture note identifies them and the inputs
whose behavior they supply. This suite does not claim to run complete tsc,
its host, path library, open node hierarchy, or configuration parser.

# Reproduction and evidence

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/real-fixtures-setup.log 2>&1
source /workspace/adamic-tools/env.sh
python3 stage3/fixtures/real/check.py /tmp/real-fixtures-verification > /tmp/real-fixtures-verification.log 2>&1
```

The script uses the unchanged shared fixture harness:
`go test ./stage3/fixtures -run '^TestFixtures$/^real$' -count=1 -timeout 10m -v`.
The final unmutated run exits 0, with all ten Node and all ten full-diagnostic
checks passing. Source Node runs through `oracle/node.mjs`; its output is
compared byte for byte with the recorded stdout, stderr and exit. Checker
errors are not accepted in place of the intended lowering diagnostic.

The scratch output mutant changes exactly one byte of the first fixture's
recorded stdout, `true:false\n` to `Xrue:false\n`. Node comparison fails and its
full diagnostic check passes. The scratch diagnostic mutant changes one byte
of `predicate` in the recorded diagnostic. Diagnostic comparison fails and
Node comparison passes. Each mutant invocation exits 1; the script exits 0
only after confirming both intended catches and independent passing checks.
No working-tree source or golden is modified by these mutants.

The unchanged `Gate.aCheck` from devtools/fast-gate
`bebd5966c6747a9ae107c7e23f56f3a90532bde7` also accepts all ten headers. Its
build step used the already built main compiler for this focused check.
Refused headers name their actual rule; `// a-check: checked` is the gate's
ordinary expectation for NotYet fixtures. Replacing the any fixture's checked
header with `refused nonexistent-mutant` makes that method fail. This header
check does not replace the stricter full-diagnostic fixture check.

[logs](logs/) retain the complete passing fixture log, both byte-mutant logs,
header observations and mutant, landed controls, setup output, and command
results. The first fixture run found absolute paths in the new status entries;
those paths were normalized before the passing run. Nothing in an existing
fixture bucket or the compiler was edited. `internal/oracle/counts.md` has no
rows for these stage3 negative fixtures: none reaches native allocation
counting, and the counted oracle fixture set was not changed.

Setup reported Node 0.026s, Go 0.029s, markdown dependencies ready 0.078s,
submodules 0.078s, clang 0.206s, Go build 35.631s, test binaries deferred
35.754s, cache warm 35.755s, done 35.786s. `nproc=5`, cgroup CPU quota four;
Go 1.27.1, Node 24.19.0, clang 20.1.8. No whole-package confirmation or full
repository gate was run. The tests here are the requested fixture measurement.
