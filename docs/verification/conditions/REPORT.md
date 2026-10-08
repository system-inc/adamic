At e43f8c16204c13a1b48adc415df534b8a217ae45, reductions of src/compiler/path.ts:385:12 and src/compiler/core.ts:2080:12 compile successfully. The former tests an optional string in a ternary; the latter tests boolean | undefined. Both now agree with Node in both backends.
Built: 13 condition fixtures, explicit TypeScript differential runs, and a NaN truthiness mutant.
Commits: area base 06877146f9563ecdc18350d6294466e9a2fa20ac; dependency merge e43f8c16204c13a1b48adc415df534b8a217ae45; current-main merge c2003e0cd2480de2376358140045ca29ef01d165.
Commands: reductions compile with exit 0; focused oracle passes; counts refresh passes, adding 13 rows with no existing changes.
Mutant: replacing five numeric ToBoolean conversions with value !== 0 produces stdout mismatches in both backends, with exit 0 and successful native compilation.
Uncovered: full 78-site census replay, arbitrary nullable unions, and restoring a historical boolean-only .a style rule.

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

No condition compiler patch was necessary for these observations. They do not establish that every one of the old 78 findings has disappeared: TypeScript must rerun its census on this branch.

The resolved area also already allows non-boolean .a conditions, as do its existing census_unary_truthiness.a and other fixtures. Conservative assumption: preserve the resolved base's behavior, rather than reinstate a historical refusal and break existing oracle fixtures. This differs from the prompt's description of today's .a refusal; it does not add any .a language permission.

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
