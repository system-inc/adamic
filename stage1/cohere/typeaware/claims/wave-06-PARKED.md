# Wave 06 parked outside unified landing

All owned rule implementations remain on codex/typeaware-wave-06 as private-driver work, outside the unified lint/rules registry. There is no green unified landing subset; step 1 is skipped.

Checker-context/link/oracle blockers apply to: adamic/nominal-class; @typescript-eslint/no-duplicate-type-constituents; no-useless-return; nexus/correctness-require-child-process-error-listener; nexus/correctness-require-response-status-check; nexus/performance-no-independent-await-in-loop; no-new-func; no-new-native-nonconstructor; no-new-wrappers; no-throw-literal; no-useless-backreference; prefer-arrow-callback; react/jsx-fragments; react/jsx-no-undef; react/no-adjacent-inline-elements.

Reproducer: build the owned actual-source checker driver without --tsgo, as recorded with exact command and output in wave_06_jsx/harness_landing_evidence and HARNESS_LANDING_REPORT.md. Compilation exits 1: `Adamic 0.1 refuses an unlinked typescript-go library call; build with --tsgo <checker archive>`. The shared RuleContext has no checker program/filename; native build has no checker archive; Go oracle has no TypeChecker. These adapters are not registered production rules, and no private-driver parity is claimed as unified certification.

react-hooks/set-state-in-effect; react-hooks/set-state-in-render; react-hooks/static-components remain parked for source-to-HIR, SSA, capture and compilation-unit analysis. Reproducer: their source entry points in wave_06_react_state refuse with panic 70, exercised by validate_partial.py and its archived refusal logs. Prepared-HIR comparisons do not certify source analysis.

No blocked rule is carried into the separate codex/lint-port-no-iterator landing branch, which starts directly from origin/area/stage1-lint.
