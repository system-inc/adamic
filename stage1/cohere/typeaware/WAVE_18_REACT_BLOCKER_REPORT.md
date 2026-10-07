# Wave 18 fifth batch: React HIR dependency blocker

Status: claimed, unported. Do not count these three rules as native coverage.
Previous preference ports and evidence are pushed through eb1251ab. The new
claim was pushed as ee56c863 before implementation. No further rules were claimed.

The claim audit fetched every origin head: 389 refs, 142 claimed names and 25
ranked baseline ports. Baselines were main e011f8f6 and tsgo-c-library 5afbdb83.
The first three remaining names were react-hooks/set-state-in-effect,
react-hooks/set-state-in-render and react-hooks/static-components.

## Observed dependencies

All three Go validators consume CoHere's React high-level intermediate
representation, not just TypeScript checker types. Its cache.go ForFunction
calls Lower and Construct (SSA); AsCompilationUnit re-lowers nested compilation
units to distinguish external bindings from captures.

* set-state-in-effect: ForFunctionWithoutManualMemoization additionally erases
  manual memoization, inlines immediately invoked callbacks and reconstructs SSA.
  findSetStateCall uses ControlDominators and phi/capture propagation for the
  ref-derived value and branch-control exemptions.
* set-state-in-render: reportSetStateInRender uses UnconditionalBlocks, SSA
  aliases, nested-function captures and a separate useMemo reporting path.
* static-components: reportDynamicComponents propagates creation identities
  through SSA phi operands, loads and stores to JSX tag Places. Correct finding
  ordering and creation-site text depend on the lowered graph.

Searching native stage1/cohere .a and .ts sources found no implementations of
HIR, ControlDominators, AsCompilationUnit or UnconditionalBlocks. The checker
bridge facts.go dispatch exposes AST/symbol/type facts and rejects unsupported
questions; it has no React HIR graph query. CoHere's HIR is a separate internal
Go compiler-analysis package, not a typescript-go checker operation.
The previously documented native JSX parser integration gap also remains on
this branch; see WAVE_18_TITLE_REPORT.md.

This is not the .a loader, shared profile compilation or suggestion serialization
work on codex/lint-harness-dot-a. Those changes alone cannot supply these graphs.
A faithful continuation needs a native React HIR lowering/SSA implementation
with source identity and capture semantics, its two memoization views, and the
control/post-dominance analyses. No shared parser, harness, registration generator
or protected compiler files were edited. No approximate AST validator or empty
native rule was installed. There is no completed independent validator to publish
without this input representation. Work stops at the dependency boundary under
Ahra's instruction to report other blockers and stop.

## Validation and limits

From cohere/, with /workspace/adamic-tools/env.sh sourced:

    go test ./internal/lint/rules/react -run 'Test(SetStateInEffect|SetStateInRender|StaticComponents)' -count=1

Output: ok github.com/system-inc/cohere/internal/lint/rules/react 0.089s.
The log is validation-wave-18-react-blocker/go-oracle-tests.log. This validates
existing Go behavior only, not an Adamic port. Setup completed in 25 seconds
(build cache warm 25s, other stages 0s); nproc is 5, CPU quota is 4 processors.

No native findings/fixes/suggestions comparison, mutant, sanitizer, released-handle
check or Go/native timing was run for these unported rules. Prior preference
coverage is unchanged and recorded in WAVE_18_PREFERENCE_REPORT.md. No full gate
was run for this documentation-only blocker report.

## Reporting-only continuation

Commit 57614fd8 adds native diagnostic rendering with explicit source-analysis
refusal in owned directories. See wave_18_react_partial/REPORT.md for byte
comparisons, sanitizers and reporting/refusal mutants. The earlier absence of
native validation described above applies to the initial blocker-only commit;
source analysis and the full corpus/timing requirements remain unfulfilled.
