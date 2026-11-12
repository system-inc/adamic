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

The refusal scans only admit syntax: both rule 800 and merged-field rulings see
uncalled generic bodies too. Neither scan records a pending obligation. The
shared `inspectorRecordBuiltUse` helper records only at lowering entries:
expression/value/statement, direct property/element/new, and call/builtin roads.
It retains an existing covered marker rather than resetting it. A member-name
identifier canonicalizes to its parent PropertyAccessExpression, while a call
keeps its own node. The trap's cover walk visits that call and its callee property
and canonicalizes names the same way. `TestInspectorDoomedBuiltNodesMatchCover`
pins the exact pointer identities; `TestInspectorDoomedDormantGenericsHaveNoObligation`
pins both uninstantiated generic cases. `TestInspectorDoomedDirectMemberRequiresPanic`
bypasses value and holds the ordinary property road to the final guard.

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
| Gate always answers false | `TestInspectorDoomedMemberValueHasPanic`, `TestInspectorDoomedEnableMeterSites` | `doomed member value lost its read-then-Panic`; `all 11 isolated sites must lower` |
| Remove the value() trap hook | `TestInspectorDoomedValueHookRequired` | `skipped inspector contract without its doomed-use Panic` |
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

## Platforms' dormant-body and sys.ts probes (-4)

The -3 scan marker is observable even for unbuilt bodies: an uncalled generic
containing `session.post` fails with the missing-Panic compiler-bug guard. An
uncalled generic taking `inspector.Session` as a value lowers on -3. The initial
-4 prototype recording both refusal skips makes both cases fail with that guard.
The final -4 lowering-time recorder makes both lower successfully, pinned by
`TestInspectorDoomedDormantGenericsHaveNoObligation`.

The no-output full-mode sys.ts unit probe used the pinned 6.0.3 adapted source.
Both frozen -3 and final -4 report 140 units: 137 attempted and three deferred
for generic instantiation. Neither reports `skipped inspector contract`.
The stock unit meter does not run production Lower's final judges, so the scratch
probe applies this final guard after successful unit construction, retaining the
statement IR in the disposable unit and checking actual function IR. It never
returns or emits a program. Both runs' loader headers report zero checker diagnostics.

Inspector reasons are unchanged on both runs (adapted coordinates):

| Attempt owner | Sites | Reason |
| --- | --- | --- |
| sys.ts:1:1 (module refusal scan) | sys.ts:1692:17 | `node:inspector.Session.post` |
| sys.ts:1689:9 (disableCPUProfiler unit) | sys.ts:1692:17, sys.ts:1692:31 | `node:inspector.Session.post` |
| sys.ts:1689:9 (disableCPUProfiler unit) | sys.ts:1706:21, sys.ts:1706:23 | `node:inspector.Session.disconnect` |

The value-hook mutant removes only the `inspectorDoomedValue` hook in `value`.
The entry recorder and ordinary property road still record the built access.
`const ctor = inspector.Session` then fails at the final guard, with the exact
catch line in the mutation table, rather than silently lowering the member.

## Program gate (#pncjepa)

`inspectorProgramHasRequire` memoizes a graph-wide fact on the lowering. It scans
all non-declaration source files from the loaded compiler program, filtering the
inspector literal before confirming the existing `nodeCommonJSRequireRoot` form.
The fact only skips recorder/trap work; the admission predicate is unchanged.
`TestInspectorDoomedAbsentGateSkipsMemberLookups` clears the program and checker
after caching false, proving both entries return before a lookup or another scan.
`TestInspectorDoomedGateSeesImportedModule` keeps the only require in an imported
second module, outside the named root list, and pins its exact trap text.
Both scalar fields are audited carried program facts; the latent disposable
program probe initializes them before capture for its fresh unit lowerings.
