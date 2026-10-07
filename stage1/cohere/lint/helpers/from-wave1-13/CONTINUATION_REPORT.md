Built: ParseAtRule, parseModifier and NodeStylesheetResolver, one .a file each, three isolated owned comparison packages.
Commits: rule branch remains pushed through 89d25f6f; helper ports e0b8c98c, 1b1382bd and c8dfc12e; all pushed on codex/lint-helpers-from-lint-wave1-13.
Commands: owned tests PASS 5.797s, 13.573s and 5.615s; 820,676 Go bytes over 4,470 helper cases agree with source Node, emitted JavaScript and sanitized native; owned vet PASS; setup 77s, nproc 5.
Mutants: three at-rule, six modifier and three resolver mutations compile, exit cleanly and are caught by every comparison. NewTheme's two earlier mutants are withdrawn evidence and excluded.
Not covered: whole-rule findings, complete integration gate, missing live design-system fixtures; next loading-helper selection is blocked, no new claim or implementation is published.

## Delivered scope

All three helpers serve the same six rules:

- better-tailwindcss/enforce-canonical-classes
- better-tailwindcss/enforce-consistent-class-order
- better-tailwindcss/enforce-consistent-variant-order
- better-tailwindcss/enforce-shorthand-classes
- better-tailwindcss/no-conflicting-classes
- better-tailwindcss/no-unknown-classes

Each removes one frozen ledger occurrence per consumer, eighteen occurrences total. Zero rules lose their final blocker. Direct helper behavior is compared against unmodified Go on every consumer's test-file literals and controls. No whole-rule findings or fixes parity is inferred from these helper tests. Modifier decoding/validation and at-rule allocation remain explicit separately owned dependencies. See at-rule/REPORT.md, modifier/REPORT.md and resolver/REPORT.md for the exact observation contracts and limits.

## Ownership

Every reservation was pushed before code. A later remote fetch exposed earlier NewTheme reservations by slots 08 and 06. Slot 08 e069e11a at 02:30:35 UTC precedes our f0b136e5 at 02:32:43 UTC. We withdrew in 2ef2c711, removed the active duplicate source and tests, and retained inert snapshots and logs only. NewTheme is not a delivered helper or a dependency removal. Latest origin claim scan shows the three retained symbols in our claim file only.

The final refresh read 427 origin refs and 17 distinct helper claim blobs. ParseValue had already been reserved when it was considered as the next target; no duplicate claim or code was written. selection-audit.json records the remaining named candidates and provenance. Unclaimed helpers still exist. loadDesignSystemForProgram, loadDesignSystemThrough and LoadDesignSystem remain six-consumer candidates, but their required live comparison inputs are missing here.

## Concrete blocker and stop

Ran, without an upstream overlay:

```
cd cohere
go test ./internal/lint/rules/tailwind -run '^TestClassOrderFixturesActuallyRan$' -count=1 -v -timeout=5m > /tmp/lint-helper-wave13-live-fixture-gate.log 2>&1
```

Observed FAIL, exit 1, package duration 0.009s. TestClassOrderFixturesActuallyRan reports no installed tailwindcss reachable from /Users/kirkouimet/Projects/ahra/app/_theme/styles. It explicitly fails to prevent the other fixtures' skips from reading as a pass. The pinned theme_fixtures.json also names these absent repository inputs:

- /Users/kirkouimet/Projects/ahra/app/_theme/styles/theme.css
- /Users/kirkouimet/Projects/connected/www-connected-app/app/_theme/styles/theme.css

The committed corpus holds resolved observations, not those original stylesheet/import graphs. A guessed replacement would not be the original live input. Under the instruction to stop and identify blockers instead of editing shared files, continuation stops before another loading-helper claim. No fake package, rewritten fixture path or shared harness change is used to hide this gap. The three standalone helpers' successful direct comparisons are unaffected; live whole-rule and next loading-helper comparisons remain unmeasured until the actual fixtures are supplied or published by their owner.

No full repository gate was run. Test output is retained in owned evidence directories. All new active Adamic files use .a. No PR was opened.
