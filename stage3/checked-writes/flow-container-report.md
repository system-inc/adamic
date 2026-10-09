Built FlowNode member and direct array/Map slot checks toward step 10, #63x2441.
Base commit: b94f41ee; delivery branch: codex/checked-wider-writes.
Node oracles, local a-check and recorded counts pass; exact commands are in flow-container-commands.txt.
Nineteen mutants are caught, including thirteen new allocation-contract mutants.
Not covered: 78 relation refusals, 11 unresolved source matches, and native execution of the complete stock tsc.

The actual allocation declares the slot contract. A FlowNode view uses the member's
actual field declaration. Pure undefined contracts reject present references.
Arrays keep an element contract, and Maps keep a value contract. Writes validate
before changing ownership or storage; splice checks every inserted item and fill
checks only when its range contains a slot. Checks are listed by --explain-checks.
Map keys stay invariant. Tuple and class element views and representation changes
keep their refusals. Checked programs use fresh arrays when constructing a new
allocation declaration. Receivers stay alive while checked-write operands run.

Fifteen new fitting/misfitting pairs cover FlowNode structural unions, undefined
object and array fields, numeric/string/boolean arrays, array index/push/splice/fill,
object elements, nested arrays, required nested fields, numeric and object Map
values, array slice and Map value spread. All fitting cases match Node in JavaScript,
sanitized native and release native; successful native runs pass leak checks.
Misfits exit 70 with the pinned path, declared type and incoming value or kind.
Node stores those misfits. The unit now has 73 runtime TypeScript witnesses,
six refused Adamic controls and two proven Adamic controls. Adamic refusal sidecars
name the source path and readonly or copy fix. Proven controls insert no write check.

The same corrected census matcher is used for each group. It visits the enclosing
assertion when the pinned receiver is an assertion. Missing source previews are
matched only when the pinned source type identifies one expression at that location.
A row without a unique match remains unresolved. It verifies all 603 source hashes
and keeps the latent loader's no-output guards. These are relation decisions on
checker-rejected adapted sources, not executable lowering of the whole compiler.

| Relation measurement | Checked | Proven | Refused | Unresolved | Total |
| --- | ---: | ---: | ---: | ---: | ---: |
| b94f41ee with corrected matcher | 325 | 0 | 179 | 11 | 515 |
| FlowNode contracts | 398 | 0 | 106 | 11 | 515 |
| Direct-container contracts | 426 | 0 | 78 | 11 | 515 |

The FlowNode group adds 73 admissions, including all 19 previously largest
FlowNode-to-FlowNode-or-undefined sites. Supported container contracts add 28.
The 89 remaining rows consist of 78 measured refusals and 11 unresolved matches.
The largest remaining contract family is 38 sites whose EmitNode.autoGenerate
promise becomes wider. Its exact first reasons, all other refusals and all unresolved
locations are recorded in flow-container-results.json. No unresolved row is counted
as proven. The previously reported 25 rows without a current refusal were re-audited,
not promoted to proofs.

The original stock tsc small programs from 3255eb1e run against stock TypeScript
6.0.3 with a pre-store check inserted at each original assignment. The flags
program stops with 'write failed: (node as Mutable<T>).flags expects
NodeFlags.Synthesized, got 0'. The parent program stops with 'write failed:
(visited as Mutable<T>).parent expects Node, got undefined'. Both exit 70.
All 301 CLI inputs match their existing golden output and exit status; 61 flags
writes pass and the parent write executes zero times. The stock broad flags domain
is a bit mask; the erased narrow witness registers [16], as the original evidence
probe does. These are Node controls. The reduced flags and parent programs are
also held to Node in Adamic's two backends. No native stock tsc build or upstream
compiler/conformance harness is claimed here.

The two original check/literal-set mutants, three diagnostic reference mutants and
never-element mutant still fail their pinned exit-70 oracle. The thirteen new
mutants remove the union, undefined-only, numeric/string/boolean array, Map value,
object element, nested-array, splice, fill or required-field allocation contract.
Each mutant executes successfully and matches Node in both backends. The
undefined-to-array mutant removes the complete reference contract, including its
kind; the nested required-number mutant removes its presence promise. Native
mutants pass sanitizers and leak checks. The exact catcher names and outputs are
recorded in evidence/flow-container-final.log and flow-container-slot-final.log.
Dropping the original check is caught by flags-misfit.ts; dropping its literal set
is caught by the same NodeFlags.Synthesized witness.

Setup completed with GOPROXY=https://proxy.golang.org|direct and the printed
/workspace/adamic-tools/env.sh sourced. Timing lines: Node 0.045s, Go 0.046s,
submodules 0.130s, markdown 0.137s, clang 0.252s, build 67.618s, total 71.986s.
nproc is 5; the cgroup quota is four CPUs. Final uncached targeted verification
passed: oracle 97.056s, driver 4.758s, native 1.589s, lower 0.944s. The final
container/FlowNode verification after retaining tuple/class refusals passed in
32.831s. The count refresh passed in 111.693s. Local a-check exited 0.

No full package test, whole gate, Windows run, native stock tsc execution, Map key
contract, Set contract, tuple/class allocation identity or unrestricted callable
contract was run or added. No cohere source was copied. All comparisons use
scratch overlays and preserve the no-backend-output guard. No evidence branch is
merged. The only push is to the worker's own delivery branch.
