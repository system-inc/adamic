# Inspector doomed uses (#cn4yapb, head 2b)

The axiom is the already ruled no-inspector host, not general dead-code deletion.
`inspectorDoomedBinding` computes doom from const declarations, wrappers and
receiver/callee chains, without a cached doom field. `inspectorDoomedUse` is the
one predicate used by rule 800, the merged-field proof gate and the expression
constructor. Parameters, lets, imports, bare aliases and non-inspector requires
remain outside this rule. Direct member chains on require itself remain refused.

`inspectorDoomedValue` is the one constructor: an ordinary read of the root
binding, then the exact Panic, followed by the typed unreachable placeholder.
An ordinary read's named refusal is propagated unchanged. Its TDZ throw remains
observable before the Panic. The trap says what happened, not "compiler bug":
`<member> was reached through <binding>, whose require("<literal>") throws on every host Adamic targets`.

The merged-field API returns an error, not an expression. This head therefore
uses the after-lowering alternative: skipped AST reads have marker slots tied
to actual Panics in the final IR. A trap dominates its expression, including
callbacks in arguments that it never constructs; separately declared callbacks
are checked separately. A missing marker or missing Panic is a compiler-bug
stop. Marker maps copy with their paired speculative IR and are not carried as
facts into another lowering (statecopy's `speculativeState` row and
`TestInspectorDoomedMarkersStaySnapshotOwned`).

| Mutation | Named pin | Catch line |
| --- | --- | --- |
| Treat every require as a doomed initializer | `TestInspectorOtherRequireBindingsAreNotDoomed` | `fs require must not create a doomed binding`; `source-map-support require must not create a doomed binding` |
| Remove the const requirement | `TestInspectorLetReceiverStaysRefused` | `let receiver must not be doomed` |
| Remove the doomed-binding requirement for inspector receivers | `TestInspectorNonDoomedMergedMemberStaysNotYet` | `non-doomed Session.post must keep rule 800 NotYet` |
| Drop the ordinary read, retaining only the Panic | `TestInspectorDoomedEscapedReadRefused`, `TestInspectorDoomedHoistedReadKeepsDeadZone` | `d1.ts must stop at the ordinary read`; `hoisted use lost its ordinary dead-zone read` |
| A contract skip may return nil without checking its actual trap | `TestInspectorDoomedMemberValueHasPanic` | `lost trap must fail closed` |
| Remove the require's host_inspector_unavailable Evaluate | `TestInspectorDoomedRequiresThrowBeforeTrap` | `inspector require must throw before a doomed use` |

Each lower mutant is a private `buildcache.GoTestMutated` binary. The IsCaught
pin runs only its named control and requires exit 1 and its catch line. The
read-dropping oracle binary additionally runs d1 and the two-const hoisted
witness on release and sanitized native, JavaScript and WASI: each reaches the exact
post Panic with exit 70. The broken-axiom oracle binary proves the constructor
Panic on native, JavaScript and WASI, with empty stdout, exact stderr and exit 70.

The full tsc enableCPUProfiler text retains its source columns. Its isolated
version removes only the captured activeSession guard/assignment. Eleven meter
sites have no inspector refusal. The full source may still report the existing
1646 guard stop; that independent representation is not widened here.

The existing parameter meter (`TestInspectorSessionMeterSitesStayRefused`),
`TestMergedHostMemberUsesNodeNotYet`, `TestMergedHostProgramCallableKeepsContract`
and `TestMergedHostImplementedMemberKeepsContract` remain controls. The real
non-doomed inspector post keeps rule 800's NotYet; the program-defined merged
class and implemented Hash.update keep their exact callable-contract Refused.

The new ordinary oracle fixture has an explicitly approved code-only row under
the existing node-without-inspector ruling. Its adjacent .node-witness.json pins
source Node's ESM outcome separately: none, because require is not defined. The
native and JavaScript outcomes are both ERR_INSPECTOR_NOT_AVAILABLE. WASI uses
the same ruled native outcome. The row excuses no byte change on any side.
