# Wave 21 parked work for unified landing

No wave-21 implementation is connected to the unified registry. All remain on
codex/typeaware-wave-21; none is proposed in a landing branch.

The unified RuleContext has no project/checker handle or binding/type queries.
This blocks no-mixed-enums, nexus/correctness-no-collection-misuse,
nexus/correctness-no-discarded-outcome, nexus/correctness-no-discarded-pure-result,
nexus/correctness-no-process-exit-after-output,
nexus/correctness-no-uncleared-race-timeout,
nexus/correctness-require-blocking-standard-streams, no-obj-calls,
no-object-constructor and no-promise-executor-return. Their private-driver
witnesses and complete oracle comparisons are retained in validation-wave-21*,
and their implementations are not registered in stage1/cohere/lint/rules.
A direct reproducer is no-obj-calls on `Object();`: distinguishing a local
Object binding from the global requires a checker identity unavailable on
RuleContext. The same missing program lifecycle prevents their existing raw
bridge queries from running through the unified context.

no-object-constructor additionally hits this native parser reproducer:

```typescript
var yield = 5;
yield: while (foo) {
    if (bar) break yield;
    new Object();
}
```

The parser treats the label as a yield expression and refuses; Go reports
useLiteral. The original exact whitespace witness and byte-37 refusal are in
wave21_jsx/validation/landing-area/REPORT.md.

react/jsx-fragments, react/jsx-no-undef,
react/jsx-no-constructed-context-values and react/style-prop-object retain
complete prepared cores but need native source/checker adaptation. Reproducers
are their own testdata controls; minimal binding/type controls are `<Missing />`,
`<Context.Provider value={{}} />` and `<div style={binding} />`. The current
context cannot supply the ordered declarations, symbols or resolved types.

react-hooks/set-state-in-render, react-hooks/set-state-in-effect and
react-hooks/static-components are parked on native HIR, SSA and captures.
Reproducers: a component calling its state setter unconditionally; an effect
calling its captured setter; and a component returning a nested component.
Raw HIR witnesses and mutants are retained in wave21_react. No source parity
or unified registration is claimed for these prepared cores.

The 17th required correctness check TestSplitTSGoAgrees is absent from this
area base: `go test ./internal/native -run '^TestSplitTSGoAgrees$' -count=1 -v`
selects no test. Existing evidence is 16 PASS / 1 ABSENT, not a green gate.

Step 1 is skipped because every owned rule has a unified source/checker or
analysis blocker. No descriptor was removed from the integration base.
