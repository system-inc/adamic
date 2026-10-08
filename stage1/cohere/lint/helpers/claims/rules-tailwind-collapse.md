# rules/tailwind/collapse package claim

Branch: lint-helpers/rules-tailwind-collapse; base origin/area/stage1-lint.
Triage d7ab0bc4 rank 14. Earlier eligible packages claimed; rank 13 generic strict-options already landed and its missing per-rule contracts belong in rule adapters per helpers/README.md.
160 retained / 243 required symbols; no complete port. Sources: retained owners in TRIAGE.md; slot05 NewTheme selected over wave08.
Forecast: zero alone; six additional with earlier complete packages, cumulative 113.
Consumers: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes, no-unknown-classes.
First prerequisite audit: exact nextBuildCount and recursive node clone/remove; capture every consuming-rule call, three runtimes, Node/native semantic mutants and one rule proof if prerequisites permit.

## Stopped on compiler prerequisite

The post-push fetch found no competing collapse package claim. Claim commit c386192b7 predates implementation; no complete package port exists in the audited sources.

`rules/tailwind/collapse.nextBuildCount`, cohere/internal/lint/rules/tailwind/collapse/design_system.go:541-545, returns the process-wide signed machine-word counter under sync.Mutex. Its exact signed-64 support probe, retained from origin/codex/lint-helpers-from-lint-wave1-12 (not a production helper), still fails on current area compiler:

- Probe: ../tailwind/collapse/testdata/atomic-counter.a.txt
- Compiler output: ../tailwind/collapse/testdata/atomic-counter.compiler.txt
- Command: source /workspace/adamic-tools/env.sh; go run ./cmd/adamic build /tmp/tailwind-collapse-prerequisite/atomic-counter.a -o /tmp/tailwind-collapse-prerequisite/counter
- Refusal: atomic-counter.a:4:10: stage 0 can't lower a function returning bigint yet

No number-counter approximation is delivered: Number loses consecutive signed-64 values above 2^53. No helper is landed or certified, no mutant is claimed caught, no rule is ported and no newly unblocked rule is credited. Helpers/lint package gates were not run after this explicit compiler stop. The earlier rules/core race was won by remote 351d296c0d8f7718c21981c155894f14511dcef4 (10:58:56Z), before local f3bad7d2 (10:59:08Z); local withdrawal is 03d96ca6a. Strict-options claim was withdrawn by published 5a9399498 because README assigns its missing contracts to per-rule adapters.
