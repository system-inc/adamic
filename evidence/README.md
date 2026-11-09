# Rule G compiler proof

Requested compiler head initially dda69fa165e2562031833030564bada78ae62b39, parent e3aa0bb7642163ebd004a71ba40a4657d27eab71. Updated by user to d643b42469bd229d461ca025652bf0dc360ee675 for steps 2–5; this adds only a test to rule_g_test.go.

Scope: steps 1–5. User cancelled admission delta (step 6) and census (7a/7b); those run on a house box. No claim is made for those gates here.

Every proof command has an output `.log` and matching numeric `.exit`. Mutant patches are retained; compiler and test source are restored after measurements. Logs 20 and 12 are at the original head; log 24 and the mutant/full-suite logs use d643b424. Log 24 temporarily adds `t.Logf("provider stops at: %v", err)` to the timer test, then restores it exactly.

Toolchain is loaded with `source /workspace/adamic-tools/env.sh`. The configured TMPDIR initially did not exist: first vet attempt (11) exited 1; directory created and vet retried (12). See 02-toolchain.log for versions.

Commands:

- 05: `git submodule update --init --recursive --depth 1`
- 10: `gofmt -l internal/lower/`
- 12: `go vet ./internal/lower/`
- 20, 24: `go test -count=1 -run 'TestASetTimeout|TestAGenericMethod|TestAGenericFunction|TestAHostMethod|TestNodeWatch' -v ./internal/lower/`
- 21: `node evidence/handler.ts`
- 31: mutant A, `go test -count=1 -run 'TestASetTimeout|TestAGenericMethodWithANarrowerConstraint' -v ./internal/lower/`
- 41: mutant B, `go test -count=1 -run 'TestAGenericFunctionWhoseSlotDoesntFixItsTypeParameterIsNotYet' -v ./internal/lower/`
- 42: `git diff --exit-code d643b42469bd229d461ca025652bf0dc360ee675 -- internal/lower/`

- 53: `ADAMIC_TYPESCRIPT_SOURCE=/workspace/typescript-source go test -count=1 -timeout 40m -skip '^TestProduct_' ./internal/lower/`

The upstream source is checked out at 050880ce59e30b356b686bd3144efe24f875ebc8 (see logs 50–52). No admission delta or standalone census was run.

See [steps-1-4.md](steps-1-4.md) for the focused results, verbatim failure output, and why neither mutant establishes sensitivity.

## Final result

Proof failed. The focused suite and full lower suite both fail in three tests (four failing extension subtests). Neither mutant introduces a new failure; mutant A leaves the narrower-constraint-unused test green. The new narrower-constraint-read test and timer test terminate inside the loader helper, before their stop logging can execute. See [step-5.md](step-5.md) for every full-suite failure verbatim.

Full suite exit 1 in 1751.026 seconds, no timeout. No compiler changes were made or retained. Steps 6–7 were skipped at the user's direction.

## Follow-up at 2594029

[Latest steps 2–4 rerun](2594029/results.md): the corrected timer pin passes and newly fails under mutant A; L2 still fails with `got <nil>` at baseline and under mutant B. See that report for full diagnostics and exact timer stops. No step 5 rerun was requested.

## L2 follow-up at ec5e9b2

[Baseline, condition-3 mutant, and stock tsc 6.0.3](ec5e9b2/results.md): baseline and mutant both fail during Load with TS2322; stock 6.0.3 also rejects the exact program with default and strict settings. Source restored exactly.

## L2 proof at e5531c8

[Baseline, mutant B, and stock tsc 6.0.3](e5531c8/results.md): baseline passes; condition-3 mutant newly fails with `got <nil>, want the unfixed type parameter named as NotYet`; stock tsc accepts under default and strict settings. Exact source restoration verified.

## L3 comparison

[Two-compiler native sanitizer and JavaScript probe](l3/results.md): e3aa0bb and e5531c8 both refuse the exact .ts and .a source with TS2532 before lowering. Source Node throws TypeError. No native or emitted JS programs could be run; no sanitizer report.

## L4 comparison

[Two-compiler native sanitizer and JS probe, source Node and stock tsc](l4/results.md): base refuses generic-method view; head refuses an unfixed type parameter falling back to a constraint holding any. Stock tsc 6.0.3 accepts; source Node throws TypeError. No emitted program or sanitizer report.

## L5 comparison

[Generic function property form, two compilers, stock tsc and Node](l5/results.md): both revisions refuse every native/JS .ts/.a compile with adamic/no-void-read. Stock tsc 6.0.3 accepts; source Node throws TypeError. Full refusals, outputs and exit codes retained.

## C1/C2 table casts

[Generic and non-generic table-to-record probes](c1-c2/results.md): both pinned compilers refuse both .ts/.a probes in native and JS compilation because fixed objects and records have different storage. Source Node throws node.bark TypeError for both; no sanitizer run or report.

## Canary build-only check

[22b8a0b and ad1ddfe census build checks](canary-build-check/results.md): fetch, vet, latent overlay generation/normalization/gofmt/build, and entry overlay generation/normalization/gofmt/build all exit 0 at both pins. No compile errors, tests or census runs.

## Guard proof at bbb63464

[Steps 1–3](bbb63464/steps-1-3.md) and [whole lower suite](bbb63464/step-4.md): formatting, vet and focused/full suites pass. Main and guarded head refuse the new cast at 13:14. Unguarded e5531c8 and the guard mutant also refuse at 13:14 with a different member-typed-any diagnostic, so the new pin stays green under mutation and does not establish guard sensitivity. Full suite exit 0 in 1619.275s; no failures. Source restored exactly.
