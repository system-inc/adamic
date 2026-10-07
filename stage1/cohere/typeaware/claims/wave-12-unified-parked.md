Wave 12 unified-harness parking

All owned type-aware ports are parked for unified-harness landing. The twelve standalone C-checker suites remain independently tested; none is registered as a unified-harness port. No landing branch is created because every owned source rule requires checker support absent from RuleContext.

Checker-blocked rules:
- @typescript-eslint/no-redeclare
- nexus/correctness-no-test-on-global-regex
- nexus/correctness-no-write-only-collection
- nexus/correctness-no-process-exit-after-output
- nexus/correctness-no-uncleared-race-timeout
- nexus/correctness-require-blocking-standard-streams
- no-new-func
- no-new-native-nonconstructor
- no-new-wrappers
- prefer-regex-literals
- prefer-rest-params
- react-hooks/exhaustive-deps
- react/jsx-fragments
- react/jsx-no-constructed-context-values
- react/jsx-no-undef

Reproducer for each checker-blocked rule: stage1/cohere/typeaware/wave12_sixth/checker_gap.a imports tsgoProgram and tsgoRelease, opens a program and releases it. Source Node through oracle/node.mjs exits 70 because oracle/adamic.mjs exports no tsgoProgram; ordinary emitted JavaScript compilation exits 1 because the checker primitive is unlinked. stage1/cohere/lint/context.ts exposes no checker/program handle; lint_test.go builds native.C through native.Build without the C-checker profile. Existing complete commands and streams: wave12_fifth/landing/D3_REPORT.md and its evidence/d3/d3-sixth capability-probe outputs. Each rule is stopped by this same shared dependency, not by its message renderer.

Existing HIR-blocked rules also stay parked: react-hooks/set-state-in-effect, react-hooks/set-state-in-render and react-hooks/static-components. Reproducers and native HIR/SSA/capture/memo blockers remain in claims/wave-12.md and wave12_fifth reports. No blocked code is brought into the new no-multi-assign branch, which starts from origin/area/stage1-lint.
