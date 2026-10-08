# rules/tailwind/collapse package claim — PARKED

Branch: lint-helpers/rules-tailwind-collapse; base origin/area/stage1-lint.
Triage d7ab0bc4 rank 14. Earlier eligible packages claimed; rank 13 generic strict-options already landed and its missing per-rule contracts belong in rule adapters per helpers/README.md.
160 retained / 243 required symbols; no complete port. Sources: retained owners in TRIAGE.md; slot05 NewTheme selected over wave08.
Forecast: zero alone; six additional with earlier complete packages, cumulative 113.
Consumers: better-tailwindcss/enforce-canonical-classes, enforce-consistent-class-order, enforce-consistent-variant-order, enforce-shorthand-classes, no-conflicting-classes, no-unknown-classes.
First prerequisite audit: exact nextBuildCount and recursive node clone/remove; capture every consuming-rule call, three runtimes, Node/native semantic mutants and one rule proof if prerequisites permit.

## Initial bigint probe (superseded by authorized number equivalent)

The post-push fetch found no competing collapse package claim. Claim commit c386192b7 predates implementation; no complete package port exists in the audited sources.

`rules/tailwind/collapse.nextBuildCount`, cohere/internal/lint/rules/tailwind/collapse/design_system.go:541-545, returns the process-wide signed machine-word counter under sync.Mutex. Its exact signed-64 support probe, retained from origin/codex/lint-helpers-from-lint-wave1-12 (not a production helper), still fails on current area compiler:

- Probe: ../tailwind/collapse/testdata/atomic-counter.a.txt
- Compiler output: ../tailwind/collapse/testdata/atomic-counter.compiler.txt
- Command: source /workspace/adamic-tools/env.sh; go run ./cmd/adamic build /tmp/tailwind-collapse-prerequisite/atomic-counter.a -o /tmp/tailwind-collapse-prerequisite/counter
- Refusal: atomic-counter.a:4:10: stage 0 can't lower a function returning bigint yet

At that initial stop, no number-counter implementation was delivered: Number loses consecutive signed-64 values above 2^53. At that initial stop, no helper was landed or certified, no mutant was claimed caught, no rule was ported and no newly unblocked rule was credited. Helpers/lint package gates had not run after that compiler stop. The earlier rules/core race was won by remote 351d296c0d8f7718c21981c155894f14511dcef4 (10:58:56Z), before local f3bad7d2 (10:59:08Z); local withdrawal is 03d96ca6a. Strict-options claim was withdrawn by published 5a9399498 because README assigns its missing contracts to per-rule adapters.

## Counter resumed; stylesheet byte-input boundary

The user authorized a number counter on 2026-10-08: real process build counts remain far below 2^53. The signed-bigint-return refusal above is retained as a known stage 0 gap, not a blocker. No bytewise integer output discrepancy was found: counter values do not enter lint output; the oracle decimal test agrees on every observed count.

Two fresh helpers now live in ../tailwind/collapse/: next_build_count.a (nextBuildCount) and builds_so_far.a (BuildsSoFar). They share counter_state.a and execute synchronously in a stage 1 process. Neither is a concurrent-worker primitive. Their tests capture actual upstream Tailwind test calls through a Go instrumentation overlay, replay those calls against the unchanged Go methods, and compare source Node, emitted JavaScript and sanitized native. Both semantic mutants compile, run and disagree on all three backends.

Stopped helper: rules/tailwind/collapse.*stylesheetCollector.loadFile, cohere/internal/lint/rules/tailwind/collapse/design_system.go:597. Go os.ReadFile preserves arbitrary bytes; stage 1's only file-input primitive, readTextFile, decodes invalid UTF-8 irreversibly. The actual Go loadFile installs a theme value retaining byte 0xff, while all stage 1 backends produce bytes 0xef 0xbf 0xbd from the same file. ASCII and Unicode controls match; this is not a compile refusal, crash or skipped parity case. The missing byte-preserving-input API probe reports TS2305: Module '"adamic"' has no exported member 'readFileBytes'; see raw-file-input.compiler.txt. No private filesystem or byte-decoder substitute is supplied.

This boundary prevents claiming complete collapse package parity and a full consuming-rule proof. No rule is ported and no newly unblocked rule is credited. The other retained package slices have not been landed. The complete-package forecast remains conditional, not achieved. See ../tailwind/collapse/REPORT.md for current gate counts and logs.

## Parked tested slice

Parked by the user: all Tailwind work waits on runtime regex compilation. The finished counter slice is local commit 8862e69e64e8beac93dc2e241d1050e191a3be6e: 293 Go calls agree across Node, emitted JavaScript and sanitized native; both mutants are caught.

Byte-preserving input need: `cohere/internal/lint/rules/tailwind/collapse/design_system.go:597` reads the stylesheet file as arbitrary raw bytes with `os.ReadFile`, then passes `string(content)` to `ParseCSS`. Adamic needs a file-input primitive that preserves those bytes (including invalid UTF-8) into the CSS parser, rather than the lossy Unicode decoding of `readTextFile`. First mismatch: `invalid80`, Go byte 128 versus UTF-8 replacement bytes 239, 191, 189.

Bigint note (informational): `testdata/atomic-counter.compiler.txt` records `atomic-counter.a:4:10: stage 0 can't lower a function returning bigint yet`. The supported number counter resolves this slice because process-local builds remain below 2^53; no counter value reaches production lint output.
