Fifteen earlier rules remain implemented and pushed; the three sixth-batch React rules are blocked and unported.
Claim 8100c5a4 was pushed before this dependency audit; this report is committed separately.
Origin audit: 389 refs, 383 trees, 33 distinct claim documents; unchanged Go React tests PASS 0.081s.
No sixth-batch native mutants, released-handle checks, sanitizer runs or byte comparisons were performed.
Missing coverage: all three native ports and timings, pending native HIR lowering/SSA/control-flow support.

## Reservation

The previous fifteen production-default native ports and their evidence were
already pushed through 5e14644c. The fetch/selection audit selected:

1. react-hooks/set-state-in-effect
2. react-hooks/set-state-in-render
3. react-hooks/static-components

The descending combined-volume ranking and lexical ties exclude the 26 base
ports and every rule named in claims on every fetched origin branch, including
skipped or released reservations. No main/bridge native implementation or prior
claim was found for these three. The claim was pushed in 8100c5a4 before any
implementation. sixth-selection.json and its reconstruction script retain the
origin snapshot and all claim names. No further rules have been reserved.

## Exact dependency gap

All three unchanged Go rules use
cohere/internal/lint/ecmascript/high_level_intermediate_representation.
The 148 native stage-1 .a/.ts sources contain no definitions or imports of the
required HIR APIs; sixth-dependency-scan.json records the search and its needles.
That observation establishes the missing current substrate, not an impossibility
of implementing it. No native rule implementation has been written for this batch.
The Go source files were read in full before this report.

| Rule | Required prerequisite | Why an AST-only approximation would differ |
| --- | --- | --- |
| set-state-in-effect | Lower/Construct, capture translation, ref-value taint, ControlDominators, manual memo erasure and callback inlining | Ref-controlled setters are exempt; outer helper calls and memoized callbacks require value/capture propagation; reports use the first setter and its reference span. |
| set-state-in-render | Lower/Construct, captured setter propagation, UnconditionalBlocks/post-dominators | Early returns and loop exits change whether a call runs unconditionally; nested setter wrappers report at the outer call; useMemo has a separate diagnostic path. |
| static-components | Lower/Construct, SSA phis, reverse-postorder taint, compilation-unit gates | Branch merges carry creation taint; loop back edges deliberately are not revisited; tag and creation spans depend on value provenance. |

The production entry points are cache.go:ForFunction and
ForFunctionWithoutManualMemoization. They call Lower and Construct; the erased
variant also calls DropManualMemoization and
InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks. The effect
and render validators then use ControlDominators and UnconditionalBlocks from
postdominator.go. The bridge currently exposes checker and syntax facts, not this
native graph. This is separate from the shared .a loading, profile compilation
and suggestion serialization work named by Ahra.

A raw checker question does not replace a native control-flow/SSA implementation.
Returning Go's derived graph or rule decisions through a new bridge question would
move this analysis to Go and would not establish the requested native port.
A prerequisite is a native HIR implementation with matching value identities,
closure captures, block order, phi order, source provenance and memo transforms,
followed by these validators and independent complete-byte comparisons.

Ahra's instruction was: "If anything else blocks you, say exactly what it is and
stop, rather than editing shared files." This batch stops at that dependency gap.
No shared generator, shared harness, compiler implementation or Go production
rule was edited. No placeholder reports success or silently returns zero findings.
The three claims remain marked blocked, not complete.

## Oracle check and limits

The unchanged pinned Go production rules and their positive/negative fixtures
were checked with output redirected to /workspace/wave-11-logs/sixth-go-react.log:

```
source /workspace/adamic-tools/env.sh
go test -v -count=1 -timeout 10m ./internal/lint/rules/react -run '^(TestSetStateInEffect|TestSetStateInRender|TestStaticComponents)'
# From cohere: PASS, package 0.081s.
```

checks.json records the exact command, test names and counts; go-react.log.gz
retains full output. This verifies oracle availability only. It is not native
agreement. No sixth-batch native implementation exists to compare on controls,
the frozen repository corpus or TypeScript's compiler corpus. No sixth-batch
mutants, released-handle checks, sanitizers, applied-fix checks or native-versus-Go
performance measurements were run. Earlier batches' passing evidence and timings
remain in their own reports and are not credited to these three rules.

Toolchain setup is reused from the earlier work: ready 0s, cache warm 86s,
total 86s; nproc 5. No PR was opened.
