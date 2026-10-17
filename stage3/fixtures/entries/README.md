# Object.entries by allocation provenance

These six fixtures hold the October 8, 11:01 ruling for step 12 (#xm3ywyc), compiler task #d9eemrs. A fresh literal bound to a const, with no writes through aliases, can prove every enumerable own data property and value type, including fields hidden by an open structural view. If that proof fails, ordinary data objects use checked reflection; an incompatible value stops loudly with exit 70. The first version refuses accessors and symbol keys.

[expectations.json](expectations.json) is the implementation contract: one entry per fixture with the required compile outcome, reflection mode, value type, runtime result and a single-edit mutant's contract. [status.json](status.json) separately records **actual** source Node bytes and current-main compiler diagnostics. A current refusal or NotYet observation is not the ruled implementation behavior.

## Cases

| File | Allocation and view | Required Adamic behavior | Source mutant |
| --- | --- | --- | --- |
| 01_scanner_keywords.a | Exact scanner keyword allocation and `new Map(Object.entries(textToKeywordObj))`, scanner.ts:136,222 | Proven; all 84 entries in source insertion order, no reflection check | Remove the real await keyword; golden count/order/value output changes |
| 02_hidden_same_type.a | Const `{ visible: 1, hidden: 2 }` through `interface Visible { readonly visible: number }` | Proven; include the hidden number | Hidden value becomes a string; checked exit 70 |
| 03_hidden_other_type.a | Same structural view hides a string | Checked; print only `before entries`, then panic with exit 70 | Hidden value becomes a number; proven success |
| 04_alias_add_key.a | Alias adds `late` before entries; union-valued string-keyed record permits the write | Checked success; include visible and late | Write a string to late; entries must stop with exit 70 |
| 05_getter.a | Hidden enumerable getter returns a number and increments a counter | Refused for getter/accessor | Replace getter with a data property; proven success, getter count 0 |
| 06_symbol_key.a | Own symbol key has a number value; view exposes only visible | Refused for symbol key | Replace symbol with string key; proven success includes hidden |

Only 01 quotes upstream declarations and statements. The other five are authored witnesses motivated by scanner.ts:222; their status.json `tsc` location is that motivating operation, not a claim that these synthetic allocations occur verbatim upstream. `origin` in expectations.json distinguishes them.

01 preserves the entire upstream table, including `["" + "constructor"]`, the MapLike declaration, and the exact entries statement. Its supporting SyntaxKind object uses the 84 numeric keyword constants from stock TypeScript 6.0.3 instead of bringing enum lowering into this test. KeywordSyntaxKind is their literal-value union. `scanner-source.json` contains the original AST spans, full source text, source hashes and constant values; only CRLF is normalized to LF. Node output also matches the complete `textToKeywordObj` exported by the independent stock compiler, with stdout hash `86e540ba23fd97c76c9fd3edf112f3b1ef34acd3c43c9b1c55a3820ada02e5a3`.

The narrower interface cases deliberately use a typed `[string, number][]` destination. Stock TypeScript chooses its `[string, any][]` fallback for an interface lacking an index signature and accepts this assignment. There is no explicit any, cast, fake library declaration or source-level check in these fixtures. `stock-audit.json` records that stock result type and the numeric contextual type. Adamic must prove or check the numeric values at the reflection boundary instead of trusting the permissive fallback. The alias record's `number | string` index value admits the mutant's string write, so its intended failure is at entries, not the write.

## Node truth and the intentional checked difference

All six raw source programs run on Node with empty stderr and exit 0. Node's entries includes hidden own enumerable string keys, invokes the getter once, and excludes the symbol key. In 03 it freely exposes `hidden:wrong`; that raw golden remains in status.json.

Adamic's checked contract for 03, and the incompatible mutants of 02 and 04, is different by the ruling: stdout is exactly `before entries\n`, no entries or final marker escape, stderr begins `adamic: panic: ` followed by a nonempty reason and newline, and exit is 70. No exact panic wording was ruled, so the contract does not invent one. Native and generated JavaScript must agree with each other on the full diagnostic bytes. Successful programs must agree exactly with raw Node on stdout, stderr and exit.

The `reflection` field records the required proven/checked classification for the compiler worker to assert in lowering tests. Output checks alone cannot prove that an unnecessary check was elided. The contract is not a substitute for inspecting that lowering fact.

## Actual measurements and headers

Base main: `45487a809f89885a3fc651cd590e7dabf31362dc`. Current results are **2 Refused, 4 NotYet**. 01 and 04 first refuse an index signature. 02, 03 and 05 stop because the structural view's exact shape is not proven; 06 stops at a computed field name. None produces a native executable. In particular, the getter and symbol cases have not yet reached the ruled Refused result.

Every fixture has a first-line a-check header. 01 and 04 use `// a-check: refused an index signature`, matching their real current diagnostic. The four NotYet inputs use `// a-check: checked`: the unchanged Gate.aCheck's default checked branch accepts clean and NotYet stops, not a false claim that code generation succeeded. Future ruled behavior is in expectations.json. Header reasons and status records need review when the implementation closes a gap.

The unchanged Gate.aCheck from `devtools/fast-gate` commit `29a6bce06472e19d6228b01bb84f8d01c69d6497` accepts all six paths. No full gate method ran. Six additional wrong-header mutants are caught by its unchanged predicate replaying captured real compiler diagnostics. The replay mutates only first-line comments, preserves every body byte and line offset, and claims no native check proof. `gate-source.json`, `evidence/a-check.json` and `evidence/header-mutants.json` pin this measurement.

The existing shared runner independently passes all six Node and stage0 subtests: `go test ./stage3/fixtures -run '^TestFixtures$/^entries$' -count=1 -timeout 10m -v`, exit 0, package time 27.476s. Its native/sanitizer hook is compiled by that harness but no native fixture runs because all six stop before compilation.

## Mutants and implementation acceptance

All six source mutants run successfully on Node, exit 0 with empty stderr, and change their fixture's stdout. `mutants.json` stores exact edits, outputs, current compiler outcomes and catchers. These prove the Node golden checks fail on altered real source input, not that the unimplemented runtime check works. The scanner changes an actual table member, not a diagnostic string; getter and symbol mutants turn the excluded property form into an ordinary data property.

`check.py` defaults to verifying raw Node goldens and the saved current observations, including all six source mutants. On newly compiling inputs it executes native and generated JavaScript, comparing success against raw Node and checked failure against the separate ruled contract; a missing or incorrect inserted check fails immediately. Generated JavaScript uses the existing oracle runner only for runtime import resolution, never as the source golden.

`check.py --acceptance` instead demands all baseline and mutant contracts in expectations.json. It deliberately fails current main: **0/12 match**, exit 1, with all twelve gaps listed in `evidence/acceptance-current-main.json`. NotYet or an unrelated refusal cannot satisfy an expected runtime panic. Getter and symbol acceptance also require a refusal reason naming their excluded property form. This red measurement belongs to the compiler feature's unfinished implementation, not to the completed fixture observation checks.

The shared runner's current status records are kept factual. Its normal native-to-source comparison is not an oracle for intentional checked differences; the bucket's acceptance command holds those separately. Refreshing a newly compiling checked fixture must retain the raw Node golden and pin its actual checked behavior, rather than replacing Node's observation with an invented panic.

## Commands run

All test output was redirected to log files, with intentional evidence copies under evidence/. No whole package or full confirmation gate ran.

```bash
export GOPROXY='https://proxy.golang.org|direct'
bash cloud/setup.sh > /tmp/step12-setup.log 2>&1
source /workspace/adamic-tools/env.sh
NODE_PATH=/tmp/step24-work/api/node_modules node stage3/fixtures/entries/extract-scanner.cjs /tmp/step24-work/typescript > /tmp/step12-extract.log 2>&1
NODE_PATH=/tmp/step24-work/api/node_modules node stage3/fixtures/entries/audit.cjs /tmp/step24-work/typescript stage3/fixtures/entries/stock-audit.json > /tmp/step12-audit.log 2>&1
python3 stage3/fixtures/entries/check.py --scratch /tmp/step12-work/final-portability-check > /tmp/step12-check-portability.log 2>&1
python3 stage3/fixtures/entries/check-headers.py /tmp/step12-work/gate.py --scratch /tmp/step12-work/headers > /tmp/step12-headers.log 2>&1
go test ./stage3/fixtures -run '^TestFixtures$/^entries$' -count=1 -timeout 10m -v > /tmp/step12-fixtures-test.log 2>&1
python3 stage3/fixtures/entries/check.py --acceptance --scratch /tmp/step12-work/acceptance > /tmp/step12-acceptance.log 2>&1
# Acceptance above intentionally exits 1 on this base: 12 unresolved implementation contracts.
```

Each source observation uses `node --disable-warning=ExperimentalWarning oracle/node.mjs <file>` and each build uses `go run ./cmd/adamic build <file> -o <scratch-binary>`. Initial status recording used `check.py --record`; subsequent verification preserves the goldens. The stock audit uses TypeScript 6.0.3 with strict, exact optional properties and unchecked-index protection, ES2024 library plus a minimal console declaration. All six programs have zero stock diagnostics. Virtual .ts filenames exist only in the compiler host; no authored .ts file is written.

For reproduction, install typescript@6.0.3 in a scratch dependency directory and point NODE_PATH there. Use a pristine v6.0.3 checkout at `050880ce59e30b356b686bd3144efe24f875ebc8` for source audits. Extract the pinned gate with `git show 29a6bce06472e19d6228b01bb84f8d01c69d6497:cloud/fast-gate/run.py` into scratch for the header check. Use new scratch paths; mutant diagnostic paths are normalized to `<scratch>/` for portability. counts.md is refreshed locally; these bucket inputs are not registered in the shared internal/oracle table and cannot supply native heap-event counts.

Setup succeeded: Go ready 0.087s, Node 0.101s, markdown 0.352s, submodules 0.386s, clang 0.913s, go build 101.494s, deferred test binaries 101.957s, warm cache 101.963s, done 102.070s. `nproc=5`, cgroup CPU quota four (`400000 100000`); Go 1.27.1, Node 24.19.0, clang 20.1.8. See evidence/setup.txt.

## Limits

No allocation-provenance analysis, runtime reflection, native layout or compiler code was changed. Runtime exit-70 checks and native/JavaScript behavior cannot yet be executed on this base. No sanitizer or native leak result is claimed. The cases do not cover inherited/non-enumerable properties, proxies, class instances, prototype mutations, alias writes after entries, deletion/reinsertion, or all possible scalar types. No other bucket, shared fixture runner, shared NOTICE or oracle registry was edited.
