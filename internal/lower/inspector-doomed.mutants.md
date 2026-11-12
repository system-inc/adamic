# Inspector doomed uses (#cn4yapb, head 2b)

The axiom is the already ruled no-inspector host, not general dead-code deletion.
`inspectorDoomedBinding` computes doom from const declarations, wrappers,
receiver/callee chains and proven function-local lets, without a cached doom field. `inspectorDoomedUse` is the
one predicate used by rule 800, the merged-field proof gate and the expression
constructor. Parameters, module-level or exported lets, imports and non-inspector requires
remain outside this rule. Const aliases inherit their initializer's proof. Direct member chains on require itself remain refused.

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
| Treat a module let as a const | `TestInspectorLetReceiverStaysRefused` | `let receiver must not be doomed` |
| Remove the doomed-binding requirement for inspector receivers | `TestInspectorNonDoomedMergedMemberStaysNotYet` | `non-doomed Session.post must keep rule 800 NotYet` |
| Drop the ordinary read, retaining only the Panic | `TestInspectorDoomedEscapedReadRefused`, `TestInspectorDoomedHoistedReadKeepsDeadZone` | `d1.ts must stop at the ordinary read`; `hoisted use lost its ordinary dead-zone read` |
| A contract skip may return nil without checking its actual trap | `TestInspectorDoomedMemberValueHasPanic` | `lost trap must fail closed` |
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


## Function-local lets (#tk5ggga)

One binding proof returns the throwing require and whether the provenance passes
through a let. The let must be function-local and unexported. Every assignment
target identifier with the same symbol is visited in the declaring function's
body, including nested functions and generators. A shorthand assignment name is
resolved to its value symbol after the ordinary l.symbol lookup, since the
checker otherwise names the property rather than the written binding. An inner
`=` that is itself an assignment target is a destructuring default, not a plain
write, and is refused. Only a plain `=` assignment to
that identifier is classified. Compound, logical, increment, decrement,
destructuring and iteration targets fail closed. Declaration lists that are
assignment targets also fail closed, since iteration writes a value rather than
performing a bare undefined initialization.

The initializer is a write. A bare declaration is a write of undefined only if
the declared checker type contains undefined. Every explicit right side must
root at a recursively proven doomed const or let, or have checker type undefined
or a string literal type. A union of those outside-arm types also counts. Null,
void, never, broad string, booleans, numbers, any, unknown and unproven objects do
not count. At least one write must provide the throwing require provenance.

The shared use predicate requires a let-derived receiver's flow type to be its
object arm alone, and excludes optional chains. A const alias of a doomed let
inherits this gate. The ordinary receiver read retains outer expressions so a
non-null assertion or cast gets its own check first. The diagnostic names only
the identifier at that receiver's root. Existing 2b const messages are unchanged.
No lowering state field was added.

| Pin | Requirement |
| --- | --- |
| `TestInspectorDoomedLetTSCShape` | The committed fixture has exactly four writes, including the bare declaration, and both guarded post and const-alias disconnect are doomed; the program lowers with final contract checks |
| `TestInspectorLetParenthesizedCalleeKeepsReceiverFlow` | Parenthesized callees cannot substitute the method function type for the underlying union receiver type |
| `TestInspectorLetWrappedOptionalCalleeStaysRefused` | A non-null callee cannot admit its inner optional member |
| `TestInspectorLetNarrowedUse` | A truthiness guard admits the inspector-only receiver |
| `TestInspectorLetOptionalStaysRefused` | An optional member remains NotYet |
| `TestInspectorLetParameterStaysRefused` | An assignment from a Session parameter refuses doom |
| `TestInspectorLetControlsStayRefused` | Imported Session, module/exported lets, for-of/in targets, inferred for-of declarations, logical assignments, destructuring, bare reachable unproven read, and closure/generator parameter writes remain NotYet |
| `TestInspectorBareLetDeclaredTypeRequired` | A bare Session lacking declared undefined is unproven |
| `TestInspectorLetLanguageClosure` | Language's pre-assignment closure read stays NotYet, including when the only explicit write roots at the throwing require |
| `TestInspectorLetHatchesCheckBeforeTrap` | Non-null retains its check; the cast stops with `adamic/no-unchecked-cast` |
| `TestInspectorBareLetInitializationAgrees` | The exact bare union declaration reads undefined after declaration; its pre-declaration control throws ReferenceError; native, sanitized, counted and JavaScript agree with Node |
| `TestInspectorDoomedLetAgrees` | With profiling unrequested, the four-write fixture prints disabled and agrees with Node on native, sanitized, counted, JavaScript and WASI |
| `TestInspectorLetNonNullCheckPrecedesTrap` | Native, sanitized and JavaScript stop at the non-null assertion's exact diagnostic and exit 70, before the doomed-use trap |
| `TestInspectorLetCountsRecords` | Both ordinary oracle fixtures have independently measured committed runtime count records |

| Mutation | Named pin | Red line |
| --- | --- | --- |
| Insert a fifth non-doomed Session write into the fixture text through its test reader | `TestInspectorDoomedLetTSCShape` | `tsc shape post must be doomed` |
| Treat any let as a const and ignore its later writes | `TestInspectorLetParameterStaysRefused` | `unclassified let must stay refused` |
| Skip the symbol-reference write scan | `TestInspectorLetParameterStaysRefused` | `unclassified let must stay refused` |
| Drop the let receiver flow/optional gate | `TestInspectorLetOptionalStaysRefused` | `unclassified let must stay refused` |
| Drop the declared-type test for a bare declaration | `TestInspectorLetLanguageClosure` | `Language closure must stay refused` |

Each is run in a private Go overlay binary, never by editing the compiler under
another test. The fifth-write overlay plants `function fifth(other: Session):
void { activeSession = other; }` into the exact fixture text before its final
`return disable()`. The lower pin must go red when that extra write removes its
doomed answer. This is a source mutant, not an admission of the fifth write.
The other four IsCaught tests require exit 1 and the listed diagnostic line.
The fixtures are ordinary top-level oracle programs because their no-profiler
outcomes agree with Node; neither needs a backend-divergence exception.


## Paired sys.ts measurement

Baseline: `2f32716054f245622c2ac3c8f692d00fa51e66e9` (`origin/main` when
fetched). The final change was restored cleanly onto main tip
`718697482c85ac53c1cde878b9d86084f655e15b`; its parent census also reports
62 / 44 and 140 / 322 on the same frozen input. Both sides use TypeScript `050880ce59e30b356b686bd3144efe24f875ebc8`,
the same completed stage3 adapted tree, and the stock latent census overlay.
The bare declaration remains bare; no tsc adaptation was added. The toolchain
is Go 1.27.1, Node 24.19.0 and clang 20.1.8 with WASI SDK 27.

| Mode | Baseline units / findings | Tip units / findings | Attempted / deferred, both sides |
| --- | --- | --- | --- |
| First-error | 62 / 44 | 62 / 39 | 60 / 2 |
| Full | 140 / 322 | 140 / 305 | 137 / 3 |

Both modes have five inspector findings at baseline and zero at tip. Both loader
headers report zero checker diagnostics. The meter's fixed measurement label
says "measured on a checker-rejected program" even when `checker_rejected` is
false; the diagnostic arrays and that boolean, not the label, establish this.

The module refusal scan's baseline stop is upstream 1696, Session.post (adapted
1692:17). That stop is gone at tip. Distinguish it from actual lowering-phase
findings in the isolated getNodeSystem unit: its first lowering-phase stop moves
from upstream 1699, Date.toISOString in the post callback (adapted 1695:70), to
upstream 1728, __filename in isFileSystemCaseSensitive (adapted 1724:41). The
post's read-then-trap dominates the callback, so its Date work is not built. The
latter CommonJS path is the next surviving lowering stop after 1696. The earlier
refusal-scan findings for __filename/__dirname at upstream 1498 remain in the
isolated getNodeSystem unit on both sides; this change does not implement them.

Reproduce after setting `ADAMIC_TYPESCRIPT_SOURCE` to the pinned checkout and
building the complete adapted tree with `stage3/apply.sh`. Build the latent
binary with `stage3/census/latent/make_overlay.py` and its generated overlay,
then run it on adapted `src/compiler/sys.ts` with `LATENT_ASSERT_NO_OUTPUT=1`.
For full mode also set `LATENT_FULL=1` and `LATENT_ONLY_FILE` to that exact file.
The measurement returns no backend program and emits no native or JavaScript.


The paired full-mode guard probe retains each successful statement unit's IR in
its disposable Main and applies checkInspectorDoomedContracts after successful
statement or function construction. It still emits no program. The guarded
baseline reports 140 units / 322 findings; guarded tip reports 140 / 305. Neither
adds a missing-Panic or skipped-inspector-contract finding.

The bare-let fixture's committed count row is `5 | 5 | 8 | 11 | 3 | 0`;
the no-profiler tsc-shape row is `2 | 2 | 4 | 6 | 2 | 0`, in the runtime record's
allocation, free, retain, release, peak and region columns. Both counted runs
agree with Node and balance their allocations and frees. The required five
IsCaught tests all passed with the red lines in the mutation table.
