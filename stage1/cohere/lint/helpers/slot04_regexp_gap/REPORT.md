No new helper claimed or built: dynamic RegExp compilation blocks the highest remaining regexp prerequisites.
Existing 66-helper branch d4a80349 is green and based on main b6b1538b and lint area d3a37422; both bases unchanged.
Node probe exits 0, prints true/false; Adamic c exits 1 with explicit nonconstant-pattern NotYet; refusal test PASS 1.740s.
No new semantic mutant credited: this is blocker evidence, not a completed helper or oracle parity claim.
No shared files edited; full repository gate and the 17 broader comparisons not run.

# Observed blocker

Fetched all 20 origin codex/lint-helpers* branches and read every claims file; claims.json preserves them. The comments bundle is already delivered and reserved in the original HELPERS.md. Highest remaining unclaimed fan-out is four, shared by the regexp compiler/matcher helpers and CFG tryStatement. There is unclaimed work; this stop does not mean the inventory is exhausted. No reservation is taken for blocked work.

The user requires rule-option patterns to use new RegExp(pattern, 'u') and forbids hand-written matchers. Current internal/lower/regexp.go:39 rejects nonconstant patterns. dynamic_pattern.a uses that required operation in a function accepting pattern:string. It type-checks, runs on Node, and is loudly refused during Adamic lowering. Making arbitrary option patterns work requires a shared compiler/runtime capability outside this unit's territory. Per the instruction to name other blockers and stop, no compiler, harness, registry, rule or claim file is changed.

The frozen ledger lists regexp.Compile and *RegExp.Test for @next/next/no-html-link-for-pages, @typescript-eslint/no-empty-object-type, no-restricted-exports and no-restricted-imports. Each has other prerequisites too. blocked.json records those rows; no dependency removal or rule readiness is credited. No regex rewriter or matcher is substituted.

# Reproduce

Source /workspace/adamic-tools/env.sh. Commands write directly to the listed logs:

- node --experimental-strip-types --input-type=module-typescript < stage1/cohere/lint/helpers/slot04_regexp_gap/dynamic_pattern.a > node-final.log 2>&1: exit 0; true then false.
- go run ./cmd/adamic c stage1/cohere/lint/helpers/slot04_regexp_gap/dynamic_pattern.a > native-final.log 2>&1: exit 1; `stage 0 can't lower RegExp with a nonconstant pattern yet`, at line 2:21.
- go test -count=1 -v ./internal/lower -run '^TestRegExpNativeRefusals$' > refusal-test.log 2>&1: PASS 1.740s. This existing test explicitly contains a dynamic constructor and expects a loud refusal. It is not changed.

Earlier node.log used the wrong stdin mode; native-refusal.log used boolean console arguments rejected by the prelude. Those unsuccessful probe drafts are preserved but are not final evidence or mutants. The final probe converts its booleans to string outputs and uses Node's TypeScript stdin mode. native-exit.log and node-exit.log record final exit codes.

bash cloud/setup.sh succeeded: Go ready 1s, clang 1s, Node 1s, submodules 1s, cache warm 212s, total 212s. nproc 5. Setup and fetch logs are preserved. Existing implementation landing verification is ../slot04_landing_d3a37422/REPORT.md: all 24 helper packages, shared harness, vet, uncached Node oracle and native runtime tests passed. No implementation changes occurred after that landing run, so it was not repeated. Only codex/lint-helpers-04 is pushed; main and area branches remain untouched.
