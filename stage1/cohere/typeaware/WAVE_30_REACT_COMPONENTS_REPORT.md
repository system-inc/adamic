Built: .a ref-lattice, purity-property and manual-memo-scope components for the three retained React claims; full rules remain incomplete.
Commits: retained claim d8478862; components 1f2e7bbc; component evidence committed separately and pushed on codex/typeaware-wave-30.
Commands and outputs: component gate PASS 18.002s; JSX prerequisite reproduction PASS 8.547s; vet PASS; setup 21s, nproc 5.
Mutants: ref identity comparison caught at byte 86865; reversed purity-kind guard at byte 1368; memo-pruning kind removal at byte 435; all exited 0 with empty stderr.
Not covered: complete React verdicts, findings/fixes/suggestions parity, corpus gates, full-rule mutants, released-handle checks or whole-rule native/Go times.

The components make progress without moving the decisions into Go or editing
shared files. They are not registered complete rules and do not make an empty
corpus result look like a port.

react_refs_lattice.a ports the six-element ref domain, convergence equality,
function equality, join, ref-carrying join, join-many and structure destructuring.
The 40 finite input values cover absent values, every domain kind, ref identity
presence, access spans, differing provenance, nested structures, function effects
and nullable function returns. Every ordered pair exercises equality, join,
join-many and destructuring: 1600 rows, 279296 exact bytes.

react_purity_properties.a ports the property propagation used by purityLoadProperty:
transparent global containers resolve Math/Date/performance; only Math.random,
Date.now and performance.now resolve to impure signatures. Its 245 combinations
cover absent values, all four purity kinds, seven object names, seven property
names and retaining an existing output value when a property is unknown:
6523 exact bytes. It does not decide render context, true-global identity,
callback invocation or propagation through the missing HIR.

react_manual_memo_scope.a ports manualMemoValidator.check's scope verdict,
including live scopes, dependency-only pruning, walked versus absent scopes,
pruned-by-chain scopes and the immutable-parameter exemption. Every combination
of six input-state flags is tested for both finding kinds: 128 rows, 900 exact
bytes. It assumes those HIR phase inputs are available; it does not construct
reactive scopes, walk the reactive tree, pair memo markers, infer dependencies,
or perform the complete memoization analysis.

Independent Go overlay tests call the private production helpers in the pinned
cohere packages directly. They do not duplicate the helper algorithms. The
private Adamic test compares full serialized component outputs with those
helpers, sanitized native output and emitted JavaScript on Node. The existing
oracle/adamic.mjs runtime is exposed in the scratch Node package; no replacement
runtime was authored. The first emitted-JavaScript attempt had a missing runtime
package; the final run installs that existing runtime and passes. Both logs are
preserved.

Each component has a decision mutant operating on real input states: ref equality
compares the identity that Go intentionally ignores for convergence; the purity
container-kind check is reversed; and pruning is applied to the finish/value
condition as well as the start/dependency condition. Mutant native executables
compile and exit successfully with empty stderr. Only the independent Go byte
comparison detects each change. ASan, UBSan and LeakSanitizer runs have empty
stderr and identical results. These are component mutants, not complete-rule
finding/fix/suggestion mutants.

The three JSX positive controls are rechecked. Go accepts their .tsx aliases and
reports one finding per rule, with zero fixes and suggestions. Native parsing
rejects the purity and memo examples with panic 70. The refs example exits 0 but
is parsed as TypeAssertionExpression, with no JSX node. A successful parse alone
therefore does not satisfy these rules' source requirements. The prerequisite
test records those gaps; it is not a successful complete-rule comparison.

A broader remote scan found JSX support on codex/stage1-jsx-lint, tip
a8a62d62ca49db7415e14c3887dd305022b17309. That branch changes the shared parser,
lookahead and scanner and adds a JSX module; it is absent from this branch and
requires shared integration. Ahra's instruction to keep changes inside our own
rule files prevents applying those shared edits here. Native React HIR/SSA,
ForFunction/AsCompilationUnit, memo erasure, reactive-scope construction and
pruning remain absent. Even with JSX integrated, those native analysis inputs
and complete transfers remain necessary for faithful Go parity.

The refreshed origin also contains an overlapping wave-04 React claim. Our claim
d8478862 has commit time 2026-10-07T01:58:25Z; wave-04's f3b2e6f9 has commit time
2026-10-07T01:59:57Z. Our saved selection snapshot did not contain that later
claim. Wave-04 has partial reporting components and documents the same missing
HIR. This is recorded as an overlap, not treated as proof of integration or a
completed native rule. Our claims remain retained and partial; no new claims
were taken while these remain unfinished.

Commands source /workspace/adamic-tools/env.sh first and redirect output to logs:

- bash cloud/setup.sh: Go 0s, clang 0s, Node 0s, submodules 0s, cache warm 21s;
  total 21s, nproc 5, cgroup quota 4 cores.
- ADAMIC_WAVE30_COMPONENT_ORACLE=/workspace/wave-30-react-components-final
  go test ./stage1/cohere/typeaware -run '^TestWave30ReactComponents$' -count=1 -v.
- ADAMIC_WAVE_30_REACT_ARTIFACTS=/workspace/wave-30-resume-prerequisites
  go test ./stage1/cohere/typeaware -run '^TestWave30ReactPrerequisites$' -count=1 -v.
- go vet ./stage1/cohere/typeaware; git diff --check.

The gate is limited to the changed package's private component and prerequisite
tests. No compiler implementation or bridge question was changed, so no new
released-handle check is applicable to these pure components. Full-rule native
versus Go times cannot be measured faithfully until the rules execute. The prior
completed process rules' whole-corpus native/Go medians remain documented in
WAVE_30_PROCESS_REPORT.md. Exact streams, hashes, source controls, logs and pins
are in validation-wave-30-react-components.
