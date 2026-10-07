# Historical harness parking, superseded by the regex contract blocker

Current status: NOT landing-ready under the new regex rule. The matcher fallback was removed; native/emitted lowering of runtime RegExp and Go/JS option-dialect parity block default-case and no-fallthrough. The historical green runs below apply to 551c529c8, before this migration. See REGEX_BLOCKER.md. No new helper is claimed until these blockers are resolved.

This branch was rebased onto origin/main c01907a7036a22c2ea7ee686ed5fe4c6cd4bbc06. Only owned rule directories and rule claims differ from main; the registration, helper and inventory foundations were removed from its history during the rebase. Integration owns shared registration and the harness.

The landing blocker is missing shared directory registration and the unified .a/emitted-JavaScript/finding harness on main. The twelve ports still need that integration to connect their numeric listener declarations to the handed-node driver and the shared Diagnostic model. Existing arrow-body-style/no-extra-bind comparisons use diagnostics in the legacy driver; their repair metadata and fixed bytes have a separate 184-case three-runtime comparison. JSX self-closing behavior uses 79 actual Go AST snapshots because the independent parser cannot parse JSX. General runtime regular expressions explicitly refuse unsupported forms. These bounded coverage limits remain in the individual rule reports.

parking_validate.py archives the previously pushed harness at 641b887ad10c109ba8a30f54966b8c6a87d7c579 into /tmp/wave10-park-validation, overlays the current owned rules, and uses compiler/internal code from the current main-based helper worktree. Shared repository files are not modified. It refreshes Go fixture answers, compares Go with source Node, emitted JavaScript and sanitized native, runs compiling output-only mutants, checks numeric listeners, and vets the touched package. Set ADAMIC_PARK_COMPILER_ROOT to another current-main checkout when reproducing. The frozen compatibility harness is evidence for parking, not a claim that the absent main harness passes.

Command (test output goes to files):

```sh
source /workspace/adamic-tools/env.sh
python3 stage1/cohere/lint/rules/default-case/parking_validate.py > /tmp/wave10-parking-validation.log 2>&1
```

Raw logs and exact results are recorded in evidence/parking-*.log.txt after the complete run. No shared finding.ts, context.ts, main.ts, registry generator, oracle or lint_test comparison is changed on this branch. Once the unified harness SHA is announced, rebase and re-green this parked branch before taking further work.

Observed results: original Go upstream capture/replay PASS 203.716s (696 ordinary owned cases and 79 separate actual Go JSX AST snapshots). The broad first run retained a FAIL after 859.029s: repair regeneration accidentally included the three documented parser exclusions. Preparation now filters exactly those same exclusions. The corrected 184-case repair comparison and rebased witnesses PASS 82.035s, with 49,634 repair bytes per runtime. After that correction, all four corpus batches PASS 407.133s: 250 frozen-source files / 750 combinations per batch, three runtimes, 152,205,712 identical output bytes total per runtime. The two legacy multi-edit rules compare corpus diagnostics; their separate upstream comparison checks all disjoint edits. The four corpus byte totals are 38,200,720; 38,030,499; 37,990,121; 37,984,372. Numeric declarations match pinned Go (322 bytes), and all twelve compiling ID-changing mutants fail only output comparison on every runtime. Focused vet exits zero. Setup ready steps each 0s, cache warm and total 124s; nproc 5.

All twelve compiling rule mutants were caught by external Go output comparison on source Node, emitted JavaScript and sanitized native:

- unsafe_component_detection_disabled
- selfclosing_fix_suppressed
- sort_vars_comparison_disabled
- physical_direction_exemption_removed
- tslint_directive_disabled
- this_capital_constructor_disabled
- arrow_findings_suppressed
- max_lines_findings_suppressed
- bind_argument_count_disabled
- default_case_missing_disabled
- redundant_label_reporting_disabled
- fallthrough_findings_disabled

The unsupported_pattern_accepted guard mutant is independently caught by the explicit-refusal comparison on all three runtimes. Raw first failures are retained; they are not counted as passing gates. The main rebase from f8013f0b to c01907a7 changes Stage 3 sources and internal/oracle/stage3_hook_test.go, not the runtime compiler or frozen lint dependencies used here. The final witnesses, repair, corpus and listener checks ran after that rebase.

Coverage limit: this parking corpus consists of TypeScript src/compiler and the frozen historical Stage 1 foundation plus current owned rule sources. Newly landed non-lint Stage 1 sources are not exhaustively replayed. The full repository gate and the unavailable unmodified main lint harness are not claimed green. Individual reports retain the measured native/Node/Go findings rates; no new speed measurement is inferred from the parking run.
