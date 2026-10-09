Using compiler 7e18441df, condition-only reductions cover all 78 TypeScript ledger sites with 241 samples held to Node in both backends; the full census records zero non-boolean condition refusals.
Commits: initial coverage 7e18441dfd2527304b47d78f396700ea256e2625; area acceptance cdddfee346e70a5cba3e982f808b9ebe426a6dfb; area base 06877146f9563ecdc18350d6294466e9a2fa20ac.
Commands: focused oracle passes in 4.106s; full census exits 0 (79 files, 324 checker diagnostics); counts refresh passes in 22.361s with one new row.
Mutants: NaN-as-truthy and null-as-truthy fail stdout comparisons in both backends; removing one source site's calls fails the historical-site coverage check.
Limits: these are condition-only reductions; four unbound-method reads and one captured optional boolean remain independent upstream blockers. No arbitrary nullable union occurs among the 78.


## Observed baseline

The requested area head already contains truthiness lowering. internal/lower/refusals.go has no non-boolean condition refusal. internal/lower/control.go condition delegates to censusCondition. Unary ! and ternaries also use this conversion. internal/native/census_small.go distinguishes zero and NaN, string length, scalar optional presence, boxed union values, and reference presence. The JavaScript backend uses Boolean.

The source ledger is pinned at c6ea4131267ddd110e0b4c0df9367d38e2c86560. Upstream source for the two reductions is TypeScript v6.0.3, commit 050880ce59e30b356b686bd3144efe24f875ebc8. Only the expression and its necessary parameter context are reproduced; these reductions are not full function or full project compilations. conditions_sites.a records them. Its Node output is:

```text
file
file.ts
sensitive
sensitive
insensitive
```

No condition compiler patch was necessary. The initial two reductions were extended to all 78 original conditions, and the full compiler-directory census was rerun on reconstructed ledger inputs, as described below.

The compiler area already accepts non-boolean .a conditions. Commit cdddfee346e70a5cba3e982f808b9ebe426a6dfb, 'Lower unary findings 85 to 0 and overload signatures 191 to 0' (October 7), changed condition() from a boolean/optional-boolean check and Refused result to censusCondition(value), for both source extensions. Its parent still has the old refusal. This commit is an ancestor of origin/area/compiler. The earlier alternate-lineage commit 76286d77b98c69d66de1be9540e6f5d0d1d64daa also implemented truthiness, but is not an ancestor of this area head.

The user's clarified ruling is to preserve today's accepting .a behavior. This branch makes no .a semantic change. The remaining style question belongs to @system_adamic with these facts; no external message was sent.

## Coverage and mutant

All repository source fixtures are .a. TestTypeScriptConditionsAgreeWithNode copies each source unchanged to a temporary .ts path, then independently runs source on Node, the JavaScript backend on Node, sanitized native, release native, and LeakSanitizer. The normal fixture registry additionally checks the .a sources. There are no new checked-in .ts sources.

| Fixture suffix | Values or behavior |
| --- | --- |
| number | 0, -0, NaN, 1, Infinity |
| string | empty string, a, 0 as a string |
| object | an object containing zero |
| array | empty array and array containing zero |
| function | closure returning zero, and undefined |
| optional_reference | undefined and an object |
| optional_number | undefined, 0, -0, NaN, 1 |
| optional_boolean | undefined, false, true |
| optional_string | undefined, empty string, 0 as a string |
| null | missing and present RegExpExecArray results |
| union | boxed number, string, boolean and undefined |
| sites | the two upstream ternary reductions |
| evaluation | calls in if, !, ?:, while, for and do...while |
| ledger78 | all 78 original operands and branch forms, 241 labeled samples |

The representation fixtures exercise if, while, for, ?: and !. evaluation prints false, false, while, for, do, do, 9, each on its own line; repeated evaluation changes the call count or selected branch. The pure function parameter includes undefined because the checker rejects a mandatory function tested in if or a loop with TS2774.

TestConditionNaNMutant mutates actual lowered IR in memory, replacing each numeric toBoolean call with a nonzero comparison. The original Node execution remains independent. Both backends compile/run the mutant successfully and exit 0, but disagree with Node on stdout. The test requires precisely that behavioral failure and reports five changed conversions. No compiler source mutation remains afterward.

Arbitrary object | null | undefined and string | null | undefined parameters remain NotYet in the area's representation function. Nullable regexp match results are the existing supported null representation and are tested. Literal null/undefined conditions are rejected by the upstream checker as always falsy (TS2873), so the fixtures use optional parameters/results. Broader nullable storage is outside this completed coverage work.

## Commands and setup

Every test invocation wrote complete output to a log before that log was read. No full package or full repository gate was run.

```sh
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/conditions-setup.log 2>&1
source /workspace/adamic-tools/env.sh
go run ./cmd/adamic c /tmp/conditions-sites.ts > /tmp/conditions-sites.c 2> /tmp/conditions-sites.log
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestTypeScriptConditionsAgreeWithNode|TestConditionNaNMutant|TestNativeAgreesWithNode/internal/oracle/testdata/conditions_' -count=1 -v -timeout 10m > /tmp/conditions-final-focused.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/conditions-counts.log 2>&1
git diff --check
```

The first setup overlapped the branch checkout and failed compiling a mixed source view (typedArrayWrite, censusFieldSlotless, parameterProperty and related helpers undefined). Setup was rerun after checkout completed and passed: node ready 0.019s, Go ready 0.023s, submodules ready 0.054s, markdown dependencies ready 0.069s, clang ready 0.153s, Go build ready 34.969s, build cache warm 35.071s, done 35.097s. nproc is 5, CPU quota is 4. Go is 1.27.1, Node is 24.19.0, clang is 20.1.8. The environment file is /workspace/adamic-tools/env.sh; /opt/adamic-tools/env.sh does not exist here.

Counts refresh: ok github.com/system-inc/adamic/internal/oracle 22.110s. Its diff contains exactly the 13 new fixture rows. Focused output, including the mutant catcher, is preserved beside this report. Final focused run: ok github.com/system-inc/adamic/internal/oracle 3.189s, exit 0. The final focused run follows the merge of current origin/main efe9f4042049234e5a52639fe77b47c311fd530c.

No lower functions, IR definitions, backend files or runtime C helpers were edited. All new compiler-related edits are in internal/oracle/conditions_test.go, its 13 named testdata/conditions_*.a fixtures, and counts.md. Documentation and preserved evidence are under docs/verification/conditions/. No code was copied from cohere.


## Full 78-site follow-up

[SITES.md](SITES.md) lists every location, operand, original checker type, reduction function, sample count and exact current census finding. The original historical row list is ledger-sites.json. inventory.cjs independently finds exactly one AST condition at every recorded line and column, using the stock TypeScript 6.0.3 checker. It asserts all 78 matches and records hashes. generate-ledger.py retains each original operand expression and its IfStatement, WhileStatement or ConditionalExpression form. It substitutes concrete leaf data types for compiler-internal object/array element types and function result types, and defines small surrounding bindings/helpers. All enum masks are the actual checker-resolved upstream values.

These reductions isolate ToBoolean. In particular, optional methods are represented by callable properties in the reductions; they do not assert that the area's separate receiver-capture refusal is solved. The full upstream functions are neither compiled nor executed by the reduction oracle. Its false/true inputs prove the conversion and original operand evaluation on the represented shapes, not every behavior of each larger compiler function. The independent full-source census retains the complete project declarations and checker diagnostics and emits no backend output.

The four nullable rows are commandLineParser.ts:4176:9 and scanner.ts:968:9 (RegExpExecArray | null), utilities.ts:10159:12 and watch.ts:790:9 (RegExpMatchArray | null). Both missing and present results pass. No arbitrary T | null or T | null | undefined union is among the 78, so this follow-up does not add general nullable storage. Other mixed unions are string | Pattern | undefined, string | readonly string[] | undefined, and boolean | EmitOnly | undefined; their reductions cover the boxed string/object, string/array and boolean/number representations. Optional arrays, method-shaped callable properties, bit masks, and the original data?.diagnostics?.length and decl.importClause?.isTypeOnly expressions also pass.

The measurement input was reconstructed by archiving stage3/apply.py, source.json, api/ and adapt/ from c6ea4131 into /tmp/conditions-ledger-base, then running its final pipeline into /tmp/conditions-ledger-source. Every recorded location still matches its original expression exactly. This is that commit's final pipeline, rather than a claim that its source hashes equal the earlier before-census measurement tree. The stock inventory reports one project diagnostic; the Adamic census retains 324. Neither tool erases those diagnostics or claims the full project is checker-clean.

The current full compiler-directory census returns zero reasons ending in ' as a condition'. Across its 79 file records it has 4,317 unique Refused and 979 unique NotYet findings, deduplicated by kind, location, exact reason and text. Counts are measured on a checker-rejected program and are not a whole-program compilation result. The raw census is census.jsonl.gz; census-summary.json records totals and findings at all 78 exact positions. Five positions retain independent blockers:

| Site | Remaining blocker |
| --- | --- |
| moduleSpecifiers.ts:699:9 | Unbound getNearestAncestorDirectoryWithPackageJson method read |
| scanner.ts:469:12 | Unbound getPositionOfLineAndCharacter method read |
| tsbuildPublic.ts:206:12 | Unbound now method read |
| utilities.ts:6484:12 | Unbound useCaseSensitiveFileNames method read |
| tsbuildPublic.ts:289:22 | boolean | undefined variable captured by a function value |

The other 73 have no finding at that exact position; that does not prove their whole unit compiled, since census attempts still stop at their first lowering failure. None of these five is an arbitrary-nullable-union gap or a ToBoolean conversion gap. Their method/capture work remains with those compiler kinds; no other worker's lowering function was changed.

TestConditionLedgerNullMutant replaces exactly the commandLineParser match condition conversion with true. The missing-match sample then disagrees with Node; both mutant backends compile, run and exit 0, and only stdout catches the error. TestConditionLedgerMissingSiteMutant deletes the three calls for builder.ts:642:13 from real source, executes the mutant on Node successfully, and requires the independent historical-site/sample assertion to catch the missing output. The NaN mutant from the initial unit remains caught. TestTypeScriptConditionsAgreeWithNode verifies the historical set of 78 sites, original expression/form markers and all 241 sample labels before comparing both backends. All .a and temporary .ts runs include native release, ASan/UBSan, LeakSanitizer and source/backend Node.

```sh
source /workspace/adamic-tools/env.sh
python3 /tmp/conditions-ledger-base/stage3/apply.py /tmp/conditions-ledger-source > /tmp/conditions-apply-ledger.log 2>&1
NODE_PATH=/home/agent/.cache/adamic-stage3/api/node_modules node docs/verification/conditions/inventory.cjs /tmp/conditions-ledger-source docs/verification/conditions/ledger-sites.json /tmp/conditions-78-types.json > /tmp/conditions-inspect.log 2>&1
python3 stage3/census/latent/make_overlay.py "$PWD" /tmp/conditions-census-overlay > /tmp/conditions-census-overlay.log 2>&1
go build -buildvcs=false -overlay=/tmp/conditions-census-overlay/overlay.json -o /tmp/conditions-census ./stage3/census/latent/tool > /tmp/conditions-census-build.log 2>&1
LATENT_ASSERT_NO_OUTPUT=1 /tmp/conditions-census /tmp/conditions-ledger-source/src/compiler /tmp/conditions-census.jsonl > /tmp/conditions-census.log 2>&1
python3 docs/verification/conditions/generate-ledger.py > /tmp/conditions-generate.log 2>&1
ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestTypeScriptConditionsAgreeWithNode|TestConditionNaNMutant|TestConditionLedger|TestNativeAgreesWithNode/internal/oracle/testdata/conditions_' -count=1 -v -timeout 10m > /tmp/conditions-78-focused.log 2>&1
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -timeout 30m -args -update-counts > /tmp/conditions-78-counts.log 2>&1
```

Focused output: ok github.com/system-inc/adamic/internal/oracle 4.106s, exit 0. The initial corpus needed explicit element types for empty array arguments and a named helper rather than an optional-boolean capture introduced only by its stub. These fixture adjustments preserve the recorded condition operands. The final artifacts regenerate without a diff. No full gate or whole package run was used. The only implementation files touched remain oracle tests, fixtures and their counts; no runtime C helper was added.

Counts refresh: ok github.com/system-inc/adamic/internal/oracle 22.361s, exit 0. Only conditions_ledger78.a was added: allocations/frees 430/430, retains/releases 80/427, peak 6, regions 0. All prior rows are unchanged.
