# String second slice report

## Built

Statically proven ToPrimitive for object conversion members, scalar/nested arrays and Map/Set; evaluation order and cleanup-aware TypeErrors; positioned startsWith/endsWith; isWellFormed/toWellFormed over canonical WTF-8; String.raw substitution conversion only when used, captured raw length, and exact empty literals. String conversion literals pass through a typed helper so typeof does not trigger clang's address comparison warning.

Shared changes are dispatch hooks: StringCall result/emission, runtime declarations, library failure reachability, and a precise boxed String refusal. Protected emit.go, lower.go, native.go and oracle_test.go are unchanged. The original concat mutant follows the new helper receiver. Counts are regenerated, including extra retain/release pairs introduced by typed helpers in existing fixtures.

## Measurement

Linux, main 50045bd, validity runner af12899 built from the separate /workspace/adamic-ts-validity checkout, test262 c8c798898646638cd0c24879f8e0374e847e7d74, stock TypeScript 6.0.3, Node v24.19.0, clang 20.1.8. Before and final tables use the unmodified runner. The observation-only audit build adds one JSON log line per result; its baseline exactly matches the unmodified baseline.

```sh
source /workspace/adamic-tools/env.sh
export PATH=/workspace/string-tsc/node_modules/.bin:$PATH
/tmp/library-string-runner -adapt -json -root /workspace/adamic -test262 /workspace/test262 -work /tmp/library-string-final-work built-ins/String > /tmp/library-string-final.json 2> /tmp/library-string-final.log
```

| Survey | Pass | Disagreement | Refused | Crashed | Skipped | Not TypeScript |
|---|---:|---:|---:|---:|---:|---:|
| Before | 205 | 1 | 267 | 1 | 316 | 433 |
| After | 257 | 0 | 217 | 0 | 316 | 433 |

All 205 baseline passes are retained. The 52 new passes comprise 51 previously refused tests and the original crash. The original disagreement moves into refused, so the refused column falls by 50 overall. Full runner artifacts are before-2.json and after-2.json; 1,223 tests in each.

The isolated fromCodePoint test originally had Node exit 0 and native exit 70. It is now refused: a try may catch the RangeError on Node, while that runtime operation still panics natively. The proof follows named calls, closures and spreads; valid literal integer code points remain accepted. The isolated S9.8_A4_T1 test originally crashed compiling C at `&adamic_string_11 == NULL` under -Werror; it now passes. Both have oracle reproducers; the intentional refusal lives in the String refusal subdirectory because flow globs top-level runnable fixtures.

## Refusal ownership and limits

The baseline partitions into 109 library first obstructions, 156 language first obstructions, and two checker/oracle diagnostic mismatches. The original coarse `new` reason contains 42 boxed String tests and seven Array/Object/Boolean/Number tests. claim-2.md supplies one-line language reproducers, and inventory-2.md records all 267 observed baseline paths and reasons. The mandated runner retains two checker refusals where stock tsc also rejects but emits different diagnostic codes (split/separator-regexp.js and raw/template-raw-throws.js). These are measurement-classifier dependencies for @system_adamic, not library features. The reported table is unchanged. No language feature is implemented here.

Boxed String construction remains refused: JavaScript String boxes have indexed exotic own properties, internal slots, prototype identity and distinct primitive/object observations. A plain object or primitive representation would silently miscompile; the required representation reaches beyond this library slice. Remaining object conversions with hidden, optional or parameterized members are refused rather than infer erased behavior. Locale operations need Node-compatible ICU and locale selection. Existing catchable repeat RangeErrors and the newly guarded fromCodePoint RangeError still need cleanup-aware runtime error propagation. Nonempty general raw array-like objects, getters, function conversion and mixed primitive/object return views remain unsupported even where a preceding refusal hides them.

## Mutants

Every automated mutant below generates valid C, exits zero and has clean ASan/UBSan and leak output. Only independently executed source on Node rejects stdout. TestLibraryStringSecondMutants runs these ten:

| Family | Mutation |
|---|---|
| String literal conversion | Identity helper slices off the first unit |
| Object ToPrimitive | Invoke valueOf instead of toString |
| Array conversion | Join with semicolon instead of comma |
| Map conversion | Return the Set tag |
| startsWith | Increase clamped start by one unit |
| endsWith | Increase clamped end by one unit |
| isWellFormed | Invert the Boolean result |
| toWellFormed | Return the original unpaired-surrogate string |
| Nullish receiver error | Name it RangeError instead of TypeError |
| raw substitutions | Use <= at the final-substitution boundary |

The twelve existing TestLibraryStringMutants also pass, with the same Node-only rejection:

| Family | Mutation |
|---|---|
| conversion | Negate the number before converting |
| prototype | Replace trim with trimStart |
| charAt | Increase the unit index by one |
| at | Increase the relative index by one |
| codePointAt | Increase the unit index by one |
| substring | Replace the start with the end |
| concat | Drop the final part |
| raw | Use <= at the substitution limit |
| rawTemplate | Cook a raw backslash-n into a newline |
| fromCharCode | Use code-point interpretation instead of code-unit interpretation |
| normalize | Use NFD instead of the requested form |
| replace | Replace every occurrence instead of the first |

A runtime boundary mutant changes the unpaired-surrogate byte threshold from >= 0xa0 to > 0xa0. The wellformed fixture rejects it only on stdout; Node and native both exit 0 with empty stderr. ASan/UBSan and leak comparison do not reject it. This specifically exercises the U+D800 boundary and paired characters assembled at runtime.

Deleting libraryStringCodePointFailure's guard makes TestLibraryStringRangeProof fail with a nil error on `try { String.fromCodePoint(-1); } catch {}`. The isolated mandated runner also recovers the original disagreement: Node 0, native 70. This is a refusal-proof control, separate from the 23 Node-only behavior mutants. Both temporary source mutations were restored before the final checks.

## Validation

- `gofmt -l cmd internal`: empty output; `git diff --check`: clean.
- `go vet ./...`: exit 0, empty output.
- `go test ./internal/lower -run TestLibraryString -count=1`: exit 0, `ok` (0.379s).
- `ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./...`: cancelled with exit 143 after roughly ten minutes in unrelated Unicode and stage-1 suites. Its completed lower (28.157s), native (271.200s) and oracle (212.582s) packages passed. Flow initially failed because the intentional refusal fixture was in the runnable glob; after moving it, `go test ./internal/flow -count=1 -timeout 30m` passed (185.560s). This is a focused Linux gate, not a claimed full green gate.
- After restoring both source mutants, `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run 'TestNativeAgreesWithNode/.*library_string_|TestLibraryString(Second)?Mutants|TestCountsAreRecorded' -count=1 -timeout 30m -v`: exit 0, `ok` (12.388s), no observation-cache hits. Covers all String fixtures, 22 automated mutants and all recorded ownership counts.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`: exit 0, generated counts; final recorded-count check passes.

Logs live under /tmp/library-string-*.log. before/final JSON and the full baseline inventory are committed with this report. First claim commit: eecec7f. Implementation commit: 87ba3f3. No PR is opened.

Setup timing: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 73s, total 73s. `nproc` is 5; cgroup cpu.max is 400000 100000. Tool environment is /workspace/adamic-tools/env.sh. Initial setup succeeded; /opt/adamic-tools/env.sh was not the printed path.
