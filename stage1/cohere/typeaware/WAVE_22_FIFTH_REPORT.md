Built: twelve earlier rules complete and pushed; three React rules claimed, blocked before implementation.
Commits: completed ports 7f3b9262; fifth claim ba8ac629; this report and reproduction committed separately.
Commands and outputs: Go production tests PASS (0.191s); native JSX probe builds successfully, then exits 70; nproc 5.
Mutants: no fifth-batch rule mutants run because no native implementations exist; earlier twelve rules retain their recorded mutant checks.
Not covered: fifth-batch native findings, repairs, suggestions, corpora, sanitizer/released-handle checks, and native/Go timings.

## Claims and selection

After all twelve prior rules were tested and pushed, every origin head was fetched
with the explicit all-heads refspec. The inventory contains 389 origin refs,
33 distinct Markdown claim documents and 30 available rule names after exclusions.
The first three were react-hooks/set-state-in-effect,
react-hooks/set-state-in-render and react-hooks/static-components. Claim ba8ac629
was pushed before implementation. The full fetched inventory is in
[selection.json](validation-wave-22-fifth/selection.json).

All three remain reserved and pending. No additional rules were claimed.

## Observed parser blocker

The independent [probe](gaps/wave_22_fifth_jsx.a) uses a JSX shape from the
production static-components tests:

```tsx
function Example() { const Component = createComponent(); return <Component />; }
```

Build command, using the previously validated stage-zero compiler:

```sh
source /workspace/adamic-tools/env.sh
/workspace/wave-22-fourth-complete/adamic build stage1/cohere/typeaware/gaps/wave_22_fifth_jsx.a -o /workspace/wave-22-fifth-probes/committed_jsx
timeout 20s /workspace/wave-22-fifth-probes/committed_jsx
```

Build succeeds. Execution exits 70 with empty stdout and this stderr:

```text
adamic: panic: parser slice expected GreaterThanToken, got SlashToken at 76 in control.tsx
```

The saved build log, stdout, stderr and exit status are in
validation-wave-22-fifth. This is a runtime parser failure, not a successful
zero-finding comparison. The production Go static-components test supports this
input shape; a JSX parser repair would touch shared parser files outside this
unit's allowed territory.

## Native graph prerequisites

Reading the production Go implementations establishes these dependencies:

| Rule | Required analysis |
| --- | --- |
| set-state-in-effect | React HIR/SSA without manual memoization; capture mapping; setter propagation; ref-derived values; control dominators |
| set-state-in-render | React HIR/SSA; capture mapping; unconditional blocks using return-exit postdominance; useMemo analysis |
| static-components | JSX lowering into React HIR/SSA; compilation-unit classification; phi and value propagation for dynamically created components |

Searches for native HIR/SSA directories and ForFunctionWithoutManualMemoization,
ControlDominators, UnconditionalBlocks and HIR classes found no corresponding
native React pipeline in stage1/cohere on this checkout. The fetched shared
harness and lint-helper branches likewise have no corresponding native HIR paths.
Existing rule-specific flow graphs do not implement these production APIs.
The inference is that native JSX lowering and React graph construction are
prerequisites for faithful ports of these rules. A checker bridge question does
not supply that native analysis. No new bridge questions, shared registrations,
parser edits or harness edits were made.

Ahra's explicit instruction is: “If anything else blocks you, say exactly what
it is and stop, rather than editing shared files.” These prerequisites extend
beyond the .a loading/profile/suggestion harness work and the owned rule files.
Work stopped at this blocker. There are no partial native rule implementations
to mistake for completed ports.

## Independent validation and limits

```sh
source /workspace/adamic-tools/env.sh
go test ./internal/lint/rules/react -run '^Test(StaticComponents|SetStateInEffect|SetStateInRender)' -count=1 -v -timeout=15m
```

Run in cohere with stdout/stderr redirected to a log. Result: PASS,
github.com/system-inc/cohere/internal/lint/rules/react, 0.191s.
The complete output is [go-production.log](validation-wave-22-fifth/go-production.log).
This validates the production Go tests, not native agreement.

The earlier setup remains in /workspace/wave-22-setup.log: warm toolchain total
118s; Go 1.27.1, clang 20.1.8, Node 24.19.0; nproc is 5 (four-core quota).
Setup was not repeated for this continuation. Earlier native/Go timings,
mutants and sanitizer evidence remain in the four completed wave reports;
none establishes coverage of these three pending rules.

## Prerequisite recheck after renewed continuation

Fetched all origin heads again successfully. The updated main, tsgo-c-library,
lint-harness-dot-a and lint-helpers inventories still contain no matching native
React graph or pending rule paths, and their parser.ts files have zero JSX
mentions. Exact inspected ref SHAs are in prerequisite-recheck.json. This
path/text search is evidence about those refs, not a proof about every possible
implementation on every origin branch. The existing native reproduction again
exited 70 with the same GreaterThanToken/SlashToken parser panic; stdout, stderr
and exit status are preserved as recheck.*. No rules were marked complete and
no further claims were made. Ahra's instruction to stop at other blockers
without editing shared files still applies.
