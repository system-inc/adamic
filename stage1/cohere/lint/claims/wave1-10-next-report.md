Built three `.a` candidate ports: physical Tailwind direction, ban-tslint-comment and no-invalid-this; pushed claim before code.
Commits: claim 82779e4b, helper merge e6b39bf2, Tailwind 2e0a0459, tslint 630ff1a9, this a416a1cd.
Checks: 657 compiler/stage1 cases matched 37,093,159 bytes; 220/221 upstream cases matched; 14 shapes, registry/helpers and two filtered compiler oracles passed.
Mutants: physical_direction_exemption_removed, tslint_directive_disabled and this_capital_constructor_disabled compiled and ran, then failed output comparison on Node, emitted JS and sanitized native.
Not covered: the remaining JSX parser case, unmodified `.a` registration and the complete uncached repository gate; the scratch compatibility requirements are reviewable in the Tailwind directory.

## Selection and ownership

Fetched all origin heads without recursive submodules and pushed the existing branch before claiming. Checked main at ef3d907ecdc4c771b016f7d9c52372def057a340 and all origin branch claim files. The original 46 helper-ready entries had only `structure/tailwind-no-physical-direction` unclaimed. REPORT.md's linked HELPERS.md now has 16 measured comment-ready entries from 6769b88e. Used those next, selecting `@typescript-eslint/ban-tslint-comment` and `@typescript-eslint/no-invalid-this`. No inventory fallback was needed. The original three blocked claims remain recorded. The selected rules were absent from main and unclaimed on origin. An older ban-tslint-comment implementation exists outside main on batch4; the continuation's eligibility criterion is main plus claims, so that branch served as reference rather than a reason to skip.

Claim 82779e4b was pushed before implementation. Updated comment helpers merged cleanly in e6b39bf2, also pushed. Each rule has its own directory and implementation commit. Authored Adamic modules are `.a`; raw upstream fixtures remain `.ts.txt`. Shared dispatch, corpus, oracle and compiler source files were not edited.

## Toolchain

`bash cloud/setup.sh` failed at cache warming: `profile_test.go:32:23: cannot range over portFiles (value of type func(t *testing.T) []string)`. Its timing lines were Go 0s, clang 0s, Node 0s, submodules 0s. Repeating setup with the scratch Go overlay succeeded: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm 23s, done 23s, 5 processors, cgroup cpu.max `400000 100000`, 17.6 GB. `nproc` printed `5`. Commands sourced `/workspace/adamic-tools/env.sh`.

## Reproduction

The owned validation tools are under `rules/structure-tailwind-no-physical-direction/`. Run from the repository root:

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/structure-tailwind-no-physical-direction/validate.py > /tmp/wave10-prepare.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^(TestWave10Shapes|TestWave10Upstream|TestWave10Corpus)$' -count=1 -v -timeout 20m > /tmp/wave10-parity.log 2>&1
go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestWave10Unterminated$' -count=1 -v > /tmp/wave10-unterminated.log 2>&1
go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestMutants$' -count=1 -v > /tmp/wave10-mutants.log 2>&1
ADAMIC_TYPESCRIPT_SOURCE=/path/to/TypeScript go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint -run '^TestWave10Throughput$' -count=1 -v -timeout 20m > /tmp/wave10-throughput.log 2>&1
go test -overlay /tmp/lint-wave1-10-next/overlay.json ./stage1/cohere/lint/registry ./stage1/cohere/lint/helpers > /tmp/wave10-packages.log 2>&1
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(case_mapping.a|library_map_set.a)$' -count=1 -v > /tmp/wave10-filtered-oracle.log 2>&1
```

TypeScript was checked out at the harness pin `050880ce59e30b356b686bd3144efe24f875ebc8` (77 src/compiler files). The stage1 corpus added 142 `.ts` and `.a` files. The source Node driver, emitted-JavaScript driver and ASan/UBSan native driver each returned zero, empty stderr and identical Go output over 657 explicitly selected rule/file cases. Compared diagnostic text, byte ranges, message IDs, fixes/suggestions, independent edit ranges and final fixed text. Corpus output: 37,093,159 identical bytes. Fourteen shape cases: 8,804 identical bytes. Original witness run: 8,980 identical bytes.

The upstream capture ran the actual Go tests and retained filenames/options. It produced 438 unique cases across all registered rules, of which these three own 221. The replay of 216 valid supported cases matched 103,198 bytes. Four unterminated comments (`/*`, `/* tslint:disable`, `/**`, `/*/`) are accepted as clean by Go's own rule tests, but the normal oracle rejects parse diagnostics and its editor rejects unparsable text. A separate comparison enabled recovered Go trees and, only when no diagnostics are reported, printed unchanged text without calling that editor: all four matched 81 bytes across the three runtimes. Thus ban-tslint-comment covers 37/37 upstream cases, no-invalid-this 131/131 and Tailwind 52/53.

The remaining Tailwind case is `export const c = <div className="flex ml-4" />;`. Go accepts its `.tsx` tree; Adamic's parser panics `parser slice expected GreaterThanToken, got Identifier at 22`. The complete replay failure is preserved. Exclusion logs list the initial five; the four unterminated cases were subsequently covered separately. The initial log label “5 rejected by Go parse-diagnostic guard” includes the JSX exclusion and is imprecise; the checked-in test now says “explicitly excluded”. No rule answer was substituted.

The no-invalid-this shape test caught an implementation error in binary operator handling. Adamic's parser stores the operator as child 1, not `node.operator`. Fixed logical and assignment handling; the returned-function/property-assignment case now matches Go. Initial repeated UTF-8 conversion made broad checks slow; obsolete runs were stopped and replaced after conversion became a linear scan of sorted comment positions. Final performance and parity runs use that implementation.

## Findings per second

Measured whole-process execution, compilation excluded, best of five interleaved rounds. Each rule ran the 77 compiler files plus one 1,000-positive fixture (78 files). All rounds returned the same 1,000 findings; compiler files alone yielded no findings for these rules. Native was optimized and unsanitized for timing; parity used sanitizers. Processes ran concurrently with validation, so these are observations on this worker, not an isolated performance claim.

| Rule | Native | Node | Go |
| --- | ---: | ---: | ---: |
| structure/tailwind-no-physical-direction | 917.19 | 1291.36 | 5059.50 |
| @typescript-eslint/ban-tslint-comment | 385.30 | 501.72 | 3791.49 |
| @typescript-eslint/no-invalid-this | 240.99 | 495.67 | 3285.15 |

## Mutants and integration limits

`TestMutants` passed all eight registered mutants in 113.985s. Each owned mutant was caught independently on source Node, emitted JavaScript and sanitized native after successful compilation/execution and empty stderr. Removing Tailwind's rtl/ltr exemption added an unwanted finding. Disabling the tslint directive regex removed a finding and deletion. Disabling the capital constructor convention added an unwanted `this` finding. See the raw log for the exact comparison differences.

Registry tests passed in 0.020s; helper tests in 88.298s. Filtered compiler oracle cases case_mapping.a and library_map_set.a passed in 0.799s with ordinary caching. The full uncached repository gate was not run: unmodified setup already fails the shared profile compile error, and `go run ./cmd/lint-registry` fails opening the new Tailwind `rule.ts` because shared registration has not learned `.a`.

The isolated compatibility patch and validate.py address `.a` discovery/imports/copying/mutants, emitted-JS comparison, profile invocation, script-kind selection, separate diagnostic/edit spans and comparison serialization. Both baselines and mutants receive the same changes. The inherited fixer uses independent edit spans only in the scratch copy. Shared integration still needs those changes and JSX parsing before these candidates meet the complete requested bar. Standard patch context lines produce whitespace warnings from `git diff --check`; rule and tool source changes have no whitespace errors. No protected compiler files were modified, no PR was opened, and no complete-green claim is made.

Raw evidence: [wave1-10-next-evidence](wave1-10-next-evidence/). Evidence includes captured owned fixtures, successful runs and the failures that motivated corrections or delimit coverage.
