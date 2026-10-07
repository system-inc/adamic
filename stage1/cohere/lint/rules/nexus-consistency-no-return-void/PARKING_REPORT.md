Built: 17 owned rule ports rebased onto origin/main b8fb957a; no shared harness edits or new claims.
Commits: this current-main evidence commit follows the rebased rule history; prior pushed rule head b6ae2ce0f and helper head 94c8e5995.
Commands: six fresh witness/mutant suites and overlay vet PASS on b8fb957a; the inherited-static-field Node oracle PASS uncached; earlier broader corpus evidence remains historical.
Mutants: all 17 rule and 12 helper semantic mutants compile, exit cleanly and are caught on source Node, emitted JavaScript and sanitized native again on b8fb957a.
Not covered: full Go parity on eleven known parser/adapter exclusions, normal shared harness integration, direct node handoff and full repository gate; these gaps extend beyond the parking exception's only-harness condition.

## Rebase and ownership

The previous full-history rebase conflicted in four inherited shared files. Under the new parking instruction, replayed only owned commits with `git rebase --onto origin/main 33b9c9e53`. This excludes the inherited registration/helper foundation history and retains current main's shared implementations. The resulting diff touches owned rule directories and stage1/cohere/lint/claims/wave1-13.md only. No shared dispatcher, finding model, context, registry generator, Go oracle or comparison file was edited on this branch.

During verification main advanced from f8013f0b to c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Both owned branches rebased cleanly again. Compiler sources, stage1 sources, cohere pin, module manifests and Node loader are unchanged between these revisions; internal/oracle/stage3_hook_test.go is the only changed file under the tested compiler/oracle paths. Fresh witness and semantic-mutant suites passed after that rebase. No main or area/ reference is pushed. Publication uses exact old-head leases to protect concurrent changes.

## Shared harness blocker and parking limit

Blocker: #zmh9v36, the shared lint harness/registration and full finding model are not on this main. Directly running the retained owned corpus overlay on the rebased rule checkout fails because the registry directory is absent and ownedWitnesses is undefined. The runtime RuleContext and generated dispatcher contracts also belong to that absent registration foundation. direct-harness-failure.log retains the actual failure. Normal integration is not claimed green.

The scratch compatibility harness comes from previously pushed 89d25f6f. Current-main compiler packages are used unchanged; non-lint stage1 sources were copied from current main. All 88 owned code, descriptor and Go adapter files were checked byte-identical between that scratch harness and the rebased branch. Scratch edits and Go overlays never alter shared tracked files. The 356-file compiler/stage1 corpus includes the pinned TypeScript compiler, current non-lint stage1 sources, owned rules and the retained lint foundation; it does not assert an exact replay of current main's replaced monolithic shared lint entry points. Earlier incomplete scratch parser-directory links failed before comparison and were corrected; those failures are not mutant catches.

Beyond registration, the following Go-valid or Go-recovered input cases remain outside parity: one top-level-await delay case; one JSX return-void callback omitted by the capture adapter; one Tailwind JSX attribute case; five multiline-string JSX ancestry cases; three legacy-octal recovery cases (01.5, 0777.5 and 0755n). Existing owned reports retain the inputs. All executed refusal controls agree on the Adamic modes, but a refusal differs from Go and is not a passing finding comparison. Therefore this report does not claim the branch has only a harness obstacle or qualifies for unrestricted landing under the parking exception. No new helper is claimed.

The corrected listener contract uses typescript-go ast.Kind names in rule.json kinds, not numeric kinds. All existing owned descriptors already declare kind names. Existing visitors retain their index-based compatibility API; new rules must take the handed node and leave relevance dispatch to the shared driver. No new rule is added. The named shared harness is ab70f38d4 on origin/lint-rules/harness; it includes JSX parser work but is not an ancestor of current main c01907a7. Its effects on the eleven exclusions have not been measured. No shared file is edited to invent an interface. Future regex ports must use the shared translated JS RegExp row or a JS RegExp literal, and option patterns use `new RegExp(pattern, 'u')`; this documentation change adds no matcher.

## Commands and evidence

Toolchain reused from the logged setup: 128s total, nproc 5; source /workspace/adamic-tools/env.sh before commands. The six retained reproduce.py snapshots supply scratch overlays. Corpus commands used `go test -overlay=<group overlay> ./stage1/cohere/lint -run '^TestWave13<group>' -count=1 -v -timeout=20m`, with ADAMIC_TYPESCRIPT_SOURCE=/workspace/scratch/typescript-6.0.3. The TypeScript pin is 050880ce59e30b356b686bd3144efe24f875ebc8. Source Node, emitted JavaScript and native with ASan/UBSan are compared against pinned Go cohere 715ba94f3608a6500086b1076ce5cb7e51b836db. The initial and strings corpus commands were rerun after correcting source-walk directory coverage, with exact test selectors and their final logs retained.

Fresh c01907a7 commands used `go test -overlay=<current-main group overlay> ./stage1/cohere/lint -run '^(TestParkingWitnesses|TestWave13.*Mutants)$' -count=1 -v -timeout=20m`. Their complete oracle-only Go test snapshots are retained as *-current-main.go.txt. TestParkingWitnesses selects the group's owned raw witnesses and checks actual Go output against all three unmutated Adamic modes. The existing semantic-mutant tests then require clean compilation and execution before checking a Go-output mismatch. `go vet -overlay=<strings overlay> ./...` passed with empty output. Full gate and the new Stage 3 oracle hook were not run.

## Fresh rule mutants

Every variant below was caught on all three Adamic modes by the independent comparison:

- TestWave13Mutants/delay-resolve-name-ignored
- TestWave13Mutants/return-void-braces-omitted
- TestWave13StringsMutants/no-multi-str-semantic-mutant
- TestWave13StringsMutants/no-nonoctal-decimal-escape-semantic-mutant
- TestWave13StringsMutants/no-octal-semantic-mutant
- TestWave13NextMutants/next-no-assign-module-variable-wrong-answer
- TestWave13NextMutants/typescript-default-param-last-wrong-answer
- TestWave13NextMutants/tailwind-direction-aware-exemption-removed
- TestWave13RepairsMutants/typescript-no-unnecessary-type-constraint-repair-mutant
- TestWave13RepairsMutants/typescript-prefer-as-const-repair-mutant
- TestWave13RepairsMutants/typescript-prefer-enum-initializers-repair-mutant
- TestWave13LoopsMutants/default-case-last-semantic-mutant
- TestWave13LoopsMutants/for-direction-semantic-mutant
- TestWave13LoopsMutants/guard-for-in-semantic-mutant
- TestWave13ReturnsMutants/no-constructor-return-semantic-mutant
- TestWave13ReturnsMutants/no-delete-var-semantic-mutant
- TestWave13ReturnsMutants/no-eq-null-semantic-mutant

## Findings per second

Best of three mixed-corpus release/process rounds; builds excluded, sanitizers used in correctness runs. Measurements were made concurrently with other owned oracle batches and should not be treated as isolated visitor performance. Inputs and compiler are unchanged on c01907a7.

| Rule | Native | Node | Go |
| --- | ---: | ---: | ---: |
| nexus/consistency-no-hand-rolled-delay | 13.73 | 19.79 | 70.51 |
| nexus/consistency-no-return-void | 12.57 | 16.11 | 57.68 |
| no-multi-str | 6.48 | 9.35 | 29.34 |
| no-nonoctal-decimal-escape | 11.05 | 15.64 | 40.19 |
| no-octal | 9.26 | 10.23 | 30.23 |
| structure/tailwind-no-physical-direction | 27.02 | 35.59 | 110.89 |
| @next/next/no-assign-module-variable | 5.18 | 7.79 | 24.11 |
| @typescript-eslint/default-param-last | 69.55 | 99.17 | 422.10 |
| @typescript-eslint/no-unnecessary-type-constraint | 20.18 | 21.87 | 67.92 |
| @typescript-eslint/prefer-as-const | 18.10 | 21.88 | 83.95 |
| @typescript-eslint/prefer-enum-initializers | 524.24 | 650.17 | 2183.96 |
| default-case-last | 9.57 | 12.19 | 49.92 |
| for-direction | 15.28 | 21.86 | 79.63 |
| guard-for-in | 11.98 | 19.30 | 68.73 |
| no-constructor-return | 16.42 | 24.10 | 78.13 |
| no-delete-var | 2.30 | 3.73 | 11.03 |
| no-eq-null | 5.23 | 7.87 | 29.83 |

## Helper inventory scan

Fetched all 532 origin refs and read all 17 distinct recursive helper claim blobs plus the complete readiness ledger and helper README. Unclaimed helpers remain. No reservation is added while the rule branch still has the parser/adapter limitations above. The retained three helper ports remove eighteen dependency occurrences across six Tailwind rules, zero final blockers; their own reports define their narrower direct-helper contracts.

## Revalidation on b8fb957a

Both owned branches rebased cleanly onto b8fb957aa839a9e8cb0b54279dd9864fa317bd30. This takes the main native inherited-static-field-read fix without modifying or reverting shared files. Main changed internal/native/emit_objects.go and its oracle fixture/count records. Stage1 inputs are unchanged. The six retained scratch overlays reran `go test -overlay=<overlay> ./stage1/cohere/lint -run "^(TestParkingWitnesses|TestWave13.*Mutants)$" -count=1 -v -timeout=20m`; all passed. These are fresh raw-witness baseline comparisons against Go in all three modes, plus all seventeen compiling semantic mutants. Broad corpus and throughput results above were measured on the preceding backend and were not rerun or remeasured on this backend. The full repository gate was not run.

`go vet -overlay=<strings overlay> ./...` passed with empty output. `ADAMIC_GATE_UNCACHED=1 go test ./internal/oracle -run "^TestNativeAgreesWithNode$/^internal$/^oracle$/^testdata$/^inherited_static_field_read.a$" -count=1 -v -timeout=10m` passed and executed the named fixture in 3.39s; cache native hits=0 misses=3 and node hits=0 misses=2. This upstream check is retained evidence, not a newly claimed mutant. The twelve helper mutants and their 4,470 baseline cases were also rerun on this main and passed. The harness ab70f38d4 still is not on main, and the eleven frontend/capture exclusions remain beyond the only-harness parking condition on this tested configuration. No new helper claim follows.

## Dedup scope supersedes the historical counts above

The 41eb6eab2 dedup ledger removes nine losing copies from this unit. Current scope is eight retained rules and two Nexus frontend/capture exclusions. DEDUP_REPORT.md records the winner mapping and fresh scoped oracle checks. The earlier seventeen-rule counts, eleven exclusions and rate table are historical, and the other workers' winning copies are not tested or altered here. The named harness is now 41eb6eab2, still absent from current main b8fb957a. No new claim follows the remaining landing cap.
