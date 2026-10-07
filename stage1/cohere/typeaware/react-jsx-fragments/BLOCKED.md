Claimed, not ported: react/jsx-fragments.
Claim commit a14936620 was pushed before this metadata; base remains current main f8013f0b.
Checks: Go production controls PASS 1.010s; three numeric declarations match Go under native sanitizers.
Mutants: three compiling listener-kind increments and three JSON metadata increments caught; no rule-semantic mutant exists yet.
Not covered: source analysis, full finding/fix/suggestion agreement, rule handle checks or native-versus-Go lint timing.

The shared parser used by current native runners has no JSX production and
ParseNode.kind remains a string with no numeric syntax kind. A fresh valid-JSX
probe using the already validated current-main runners gives Go exit 0 and
native panic 70 before dispatch:

    parser slice expected GreaterThanToken, got SlashToken at 94

The probe's Go runner selects the prior core rules, so it demonstrates parser
acceptance only, not these new rules' findings. The unchanged production tests
for the three new rule families separately pass. Shared JSX exists on
codex/stage1-jsx-lint, but current main and area/stage1-lint are both f8013f0b
and do not integrate it. Changing shared parser, harness or driver is outside
this unit's scope; Ahra's instruction says to name such a blocker and stop.

These three rules use syntax and checker facts, not the parked React compiler
HIR/SSA analysis. They are BLOCKED on the shared JSX and numeric node interface,
not parked under the HIR exception. The newly added listeners.a and rule.json
are metadata only. No zero-finding placeholder or guessed source adapter is
present. A rule handler consuming a numeric handed node cannot be implemented
against the present string-only ParseNode contract without inventing a shared
interface. Full source analysis and the actual speed contract remain unfinished.

Independent testdata/listener_oracle.go parses production Go rule.Listeners
literals and uses pinned Go AST kind constants. It imports no native rule or
bridge implementation. Numeric registrations are:
react/jsx-fragments 285,286,289;
react/jsx-no-constructed-context-values 286,287;
react/jsx-no-undef 286,287.
Normal and ASan/UBSan/LSan native declaration output matches production bytes
with empty stderr. Each first-kind-plus-one native mutation compiles and exits
0 with empty stderr; only production-registration bytes catch it. JSON metadata
mutations are checked separately. These are registration checks, not rule
semantic mutants or full rule sanitizer coverage. No new bridge query is needed
for metadata, so no new released-handle claim is made.

Evidence is in ../react-jsx-fragments/validation/. All test outputs went to
files. An initial registration-oracle invocation passed the repository root
instead of cohere, failed to locate Go source, and was corrected; both logs are
preserved. Setup succeeded: Go 0s, clang 1s, Node 1s, submodules 1s, cache warm
121s, total 121s; nproc 5, quota four cores, 17.6 GB. Go 1.27.1, clang 20.1.8,
Node 24.19.0, environment /workspace/adamic-tools/env.sh.

Commands:

    bash cloud/setup.sh > /workspace/wave20-validation/fifth-setup.log 2>&1
    source /workspace/adamic-tools/env.sh
    (cd cohere && go test ./internal/lint/rules/react -run '^(TestJsxFragments|TestJsxNoConstructedContextValues|TestJsxNoUndef)' -count=1 -v > /workspace/wave20-validation/fifth/go-controls.log 2>&1)
    python3 stage1/cohere/typeaware/react-jsx-fragments/listener_verify.py --repository /workspace/adamic --artifacts /workspace/wave20-validation/fifth/listener-guards --compiler /workspace/wave20-validation/continue-third/adamic --checker /workspace/wave20-validation/continue-third/checker.a > /workspace/wave20-validation/fifth/listener-guards.log 2>&1

The nine completed prior rules remain based on unchanged main and green from
LANDING_F8013F0B_REPORT.md. No implementation changed, so those full gates and
the full repository gate were not repeated for metadata. No new lint timings
are available because these new analyses cannot run on source. This worker
publishes only codex/typeaware-wave-20, never main or area/. No further rules
were claimed after this blocked batch.

## Named harness dependency incorporated

The branch now includes ab70f38d4 and current main c01907a7. Fresh native and
Go probes both accept the previously failing valid JSX, exit 0 with matching
bytes. JSX parsing is no longer the blocker here. The shared finding model
and optional handed-node visitor are present. ParseNode.kind and the registry
still use string kinds; the numeric shared interface remains pending. Source
analyses remain unfinished (HIR/SSA claims remain parked). See
../react-jsx-fragments/HARNESS_REBASE_REPORT.md for fresh gates and limits.

## Kind-name correction and current landing

The latest instruction explicitly uses registry `ast.Kind` names, not numeric
values. The rule.json and listeners.a declarations now use those names.
Independent Go production registrations, native output and sanitizer output
agree; replacing a declared kind with Identifier is caught after compilation.
The old numeric-interface blocker above is superseded. JSX parsing is available
on this branch through the named shared harness dependency.

Shared registration still needs a checker program/lease, configuration and root
manifest in RuleContext, which the shared API does not expose. No shared driver
or context file was changed. The JSX fragment and constructed-context source
analyses remain unfinished; jsx-no-undef has a tested private native analysis.
These syntax/checker rules are not parked HIR/SSA rules. No further claims were
made. The current landing report records revalidation against main b8fb957aa.

## Native source analysis now implemented

Source 7f646aa90 ports the full Go fragment predicate and both modes.
57 controls/34 findings, corpora, sanitizer, semantic mutant and released
question pass. Shared registration still lacks the checker lease/context;
see ANALYSIS_REPORT.md. The old source-not-ported statements are historical.
jsx-no-constructed-context-values remains unfinished, not parked.
