Built: a pushed reservation and independent prerequisite evidence for set-state-in-effect, set-state-in-render and static-components; no native implementations of these three are claimed.
Commits: previous work pushed through 6af00212; claim 3e7267e2 pushed before investigation code; prerequisite/report commit recorded in git history.
Commands/output: TestWave21ReactPrerequisites PASS 47.819s; six Go-positive controls, compiler77 and repository287 Go baselines; parser checks normal and ASan/UBSan/LSan; vet/gofmt clean; nproc 5.
Mutants: three Go control-selection mutants remove required messages; parser-bypass mutant removes the required JSX refusal. These are prerequisite guard mutants, not qualifying native rule mutants.
Not covered: all three native rule implementations, native/Go agreement or speed comparison, native rule mutants, new-question released-handle checks, shared harness integration and the full gate.

Selection

Fetched every origin head, including branches excluded by origin's default fetch
refspec. The audit covers 389 refs and 33 Markdown claim documents, with 26
base ports; main is e011f8f6 and the bridge is 5afbdb83. The selected rules have
full-ranking positions 168, 169 and 170 and zero combined volume. Their names
occur in no audited claim and no native rule source on either base. The audit
is preserved in validation-wave-21-react/selection.json.xz.

Explicit reservation releases were checked. The old wave-02 exhaustive-deps
reservation is released, but waves 12 and 27 actively reserve the same rule.
Core prefer-promise-reject-errors has active continuation claims. Promises,
spread and lost-update remain actively claimed by wave 22. Those entries are
therefore unavailable. No further batch is reserved.

Observed controls and limits

The independent oracle imports cohere's unchanged production registry, builds
its own program and serializes complete findings, fixes and suggestions. It
imports no bridge code. Each of six controls produces one finding:

- effect-direct: a synchronous setter in an effect.
- effect-captures: two closures between the effect and the setter.
- render-direct: a setter called unconditionally during render.
- render-captures: the unconditional setter reached through closures.
- static-direct: a component returned by a call and used as a JSX tag.
- static-phi: a component that is dynamic on only one incoming branch.

The real native parser accepts all four non-JSX controls, normally and under
ASan, UBSan and LeakSanitizer, with empty stderr. Both JSX controls exit 70:

- static-direct: parser slice expected GreaterThanToken, got SlashToken at 76.
- static-phi: parser slice expected GreaterThanToken, got SlashToken at 133.

The parser-only probe is explicitly labelled as a prerequisite probe. It runs
no lint rule. Its successful parses and sanitizer results are not native rule
agreement or a sanitizer gate for the requested ports. The two rejected JSX
inputs remain valid positive inputs to production Go.

The frozen compiler77 and repository287 manifests produce zero Go findings,
5087 and 18485 canonical output bytes respectively. Go reports compiler load
234778163 ns and run 3737238224 ns (including rule time 3701155953 ns), and
repository load 55091049 ns and run 55490413 ns (rule time 10455259 ns).
No native rule timing exists, so no native-versus-Go speed claim is made.

Exact prerequisite and inference

All three production rules consume
cohere/internal/lint/ecmascript/high_level_intermediate_representation.
The ordinary entry point is ForFunction: Lower followed by Construct, with
separate identifier namespaces for each function, capture/context edges,
reverse-postorder blocks, instructions and phis. AsCompilationUnit re-lowers
nested compilation units. The effect rule uses a separate transformed entry
point: ForFunctionWithoutManualMemoization, which erases useMemo/useCallback,
inlines callbacks and reconstructs SSA when necessary.

The validators require different properties of this graph:

- set-state-in-render uses UnconditionalBlocks (post-dominators), translates
  captured setter identifiers and distinguishes useMemo callbacks.
- set-state-in-effect tracks ref-derived instruction operands and patterns,
  uses ControlDominators and follows setter functions across capture edges.
- static-components propagates creator identities through instructions and
  phis, then reports the JSX tag using the creator's original source range.

The inspected native process_graph.a represents syntax events, reads/writes,
successors and liveness. It has no React HIR value/instruction tables, SSA phis,
nested-function arenas or capture/context correspondence. No native equivalent
of the above HIR producer is present on this branch. This is a missing native
prerequisite, not an .a-loader or suggestion-serialization limitation.

Inference: using the available syntax graph as a substitute would not preserve
those production algorithms. The closure and branch controls show why a name
heuristic cannot be reported as a completed byte-for-byte port. Producing React
HIR is separate prerequisite implementation work; it has not been implemented
here. No Go lint verdict or Go React lowering was added to the checker bridge
as a substitute for native rule execution.

Under Ahra's instruction to stop on other blockers, this batch stops before
native rule implementation. The reservations are retained, not released. No
shared registration generator, shared test harness, native parser, protected
compiler file or bridge registration was changed. The only new Adamic source
is the .a parser probe; JSX fixture text is preserved as JSON evidence.

Verification and guard mutants

The private prerequisite test uses existing harness helpers without editing them.
Its three control-selection mutants change the Go oracle's selected-rule map
entry to false, one per rule. Every mutant builds and exits 0, and the required
Go message disappears, demonstrating that the positive-control guard can fail.
The parser-bypass mutant builds and exits 0 on the unsupported JSX control,
with empty stderr; the required parser-refusal guard catches the lost panic.
These are not native rule-judgment mutants and do not satisfy the requested
per-rule mutation bar. Released-handle checks for new rule questions were not
run: no rule or checker question was implemented.

Command, redirected directly to /workspace/wave21-react-final.log:

    source /workspace/adamic-tools/env.sh
    ADAMIC_WAVE21_REACT_ARTIFACTS=/workspace/wave21-react-final ADAMIC_WAVE21_COMPILER_CONFIG=/workspace/wave21-compiler/src/compiler/tsconfig.json ADAMIC_WAVE21_COMPILER_MANIFEST=/workspace/wave21-compiler.manifest ADAMIC_WAVE21_REPOSITORY_MANIFEST=/workspace/wave21-repository.manifest go test ./stage1/cohere/typeaware -run '^TestWave21ReactPrerequisites$' -count=1 -timeout=10m -v

PASS 47.819s. Touched-package go vet and gofmt output are empty. The initial
run passed its controls but failed at a mistakenly supplied compiler config
path; that log is retained, and the final run uses the correct frozen config.

The original setup timing remains tools 0s, submodules 0s, cache 81s, total 81s;
no fresh setup was needed for this continuation. nproc reports 5. Complete
outputs, fixture text, manifests, guard-mutant evidence and hashes are in
validation-wave-21-react/. The full repository gate was not run.
