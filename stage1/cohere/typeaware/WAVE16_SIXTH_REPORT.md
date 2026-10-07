# Wave 16, React HIR prerequisite blocker

The previous fifteen ports and their validation were pushed at a278e5e7.
Claim 58d204d3 was pushed before implementation. The sixth set is
react-hooks/set-state-in-effect, react-hooks/set-state-in-render and
react-hooks/static-components. These three are claimed but NOT ported.
No additional rules were claimed after discovering the prerequisite gap.

## Exact blocker

The production rules consume React high-level intermediate representation,
not only syntax and checker types. This branch has no native source-to-React-HIR
frontend, SSA construction, capture correspondence or post-dominator analysis
available to its rule modules. stage1/cohere/typeaware/flow.ts relates mutable
container types; it does not produce that representation. internal/flow is
compiler Go code, not an Adamic rule module or an exposed checker bridge question.
The isolated wave16 JSX parser is available and works. Neither JSX parsing nor
.a loading is the blocker identified here.

The concrete production dependencies are:

| Rule | Required representation and analyses |
| --- | --- |
| set-state-in-effect | ForFunctionWithoutManualMemoization; Lower and Construct; DropManualMemoization; memo-inclusive immediately-invoked callback inlining and SSA reconstruction; captured setter mapping; ref-derived value propagation; ControlDominators |
| set-state-in-render | ForFunction; Lower and Construct; nested capture translation; UnconditionalBlocks post-dominance; useMemo callback recognition within the compilation unit |
| static-components | ForFunction; Lower and Construct; ordered phi operands and reverse-postorder instructions; value-to-source-node correspondence; compilation-unit gate |

All three also use AsCompilationUnit to lower a component nested inside a plain
wrapper with the correct enclosing binding semantics. static-components
intentionally uses one forward pass, including its production back-edge
limitation. set-state-in-render checks its useMemo arm before unconditional
blocks. set-state-in-effect has ref-derived data and control exemptions. A
source-order approximation would change these decisions and fail the requested
byte agreement.

This requires a native HIR frontend and graph substrate, which is shared work
outside the permitted rule directories. Ahra's instruction says to stop on other
blockers instead of editing shared files. No shared generator, harness or
protected compiler file was edited. No bridge question that returns a Go lint
verdict was substituted for native analysis. No reporting-only shell was
registered as a working rule.

## Evidence and commands

validation-wave16-sixth/prerequisites.py records exact source hashes and HIR
calls, and searches every fetched origin head's stage1/cohere .a/.ts modules
for declarations of five concrete prerequisite entry points. Its JSON preserves
the individual heads and results. This name search is evidence of missing known
entry points, not a proof that differently named equivalents cannot exist.
The branch's rule interfaces and flow implementation were inspected separately.
Other React workers also document incomplete HIR prerequisites; their partial
validators are not a source-to-HIR frontend this unit can consume.

The retained toolchain comes from /workspace/adamic-tools/env.sh. Original setup
reported Go 0s, clang 1s, Node 1s, submodules 1s, cache 85s, done 85s. nproc is 5.
No toolchain installation was repeated.

```
python3 stage1/cohere/typeaware/validation-wave16-sixth/prerequisites.py \
 > stage1/cohere/typeaware/validation-wave16-sixth/prerequisites.json
# cwd cohere, after sourcing the retained environment:
go test ./internal/lint/rules/react \
 -run '^Test(SetStateInEffect|SetStateInRender|StaticComponents)' \
 -count=1 -v -timeout 10m \
 > /workspace/wave16-artifacts/sixth-go-prerequisites.log 2>&1
```

The pinned production Go tests PASS, package time 0.090s. The complete log is
preserved here as go-prerequisites.log. They exercise positive controls, alias
chains, nested captures, useMemo versus useCallback, ref exemptions, component
creation spans and creation-site messages. This confirms the Go reference is
runnable; it is NOT a native differential test or a runtime benchmark.

## Coverage and continuation

There is no sixth-set native implementation, source-corpus byte comparison,
rule mutant, released-handle check, sanitizer result or native-versus-Go timing.
Those checks cannot be reported as passing for an implementation that does not
exist. The prior fifteen ports retain their committed comparison, per-rule
mutant, released-handle, sanitizer and timing evidence in their earlier reports.
Their earlier Nexus dispatch integration gap remains separately documented.

The next step is to supply the native React HIR frontend/SSA and graph analyses,
then implement these three validators in their own .a modules and run the full
requested comparisons. Shared .a/profile/suggestion harness changes alone do
not supply that substrate. This unit stops with these three claims incomplete,
without claiming another set.

## Continuation dependency refresh

Fetched all origin heads again and reran prerequisites.py: 416 heads,
0 declarations of the five known native prerequisite entry points. The
audit now caches identical Git blobs; its search criteria are unchanged.
Main remains e011f8f6. The bridge advanced to eb6df00e with build-flag and
timing evidence, without a React HIR frontend. The harness is f4d98cab;
its profile and suggestion work does not implement React HIR. React workers
40c55da4 and 5de178b9 still explicitly document their missing HIR substrate.

No rule implementation or shared file changed, so no native rule test or
benchmark was rerun. The production Go reference test result above belongs
to the prior run. All three existing claims remain incomplete. No new claims
were made. Stop-on-prerequisite instructions still apply.
