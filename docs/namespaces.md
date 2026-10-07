# Namespaces in Adamic 0.2

Decision for Kirk, October 6, 2026: a sound qualified-name subset exists, and
this branch implements it. Admit a single module-scope declaration, nested
namespaces, interfaces/type aliases, ordinary or generic functions, exported
constants, classes, and initialized private `let`/`const` bindings. Namespace functions
can be called, detached and compared by identity. Names can coexist with an
interface or type alias of the same spelling. Type-only namespaces erase.

The namespace object's identity, reflection, escape and mutation are outside
this subset. A function merge admits the restricted qualified uses below.
Reopening, class/enum merging, mutable exports, ambient
namespaces, executable namespace bodies are
`NotYet`, with an alternative using a file module or private state behind fixed
exported functions. A namespace function using `this` is refused: qualified and
detached calls have different receivers; pass state explicitly.

This supersedes the blanket namespace refusal in [0.1](0.1.md), alongside the
[parameter-property decision](parameter-properties.md). The loader and tsconfig
agree on `erasableSyntaxOnly: false`; explicit refusals still guard other syntax.

## What TypeScript's ten runtime declarations need

Observed from an independent stock-TypeScript AST parse of the pinned 6.0.3
source, commit `050880ce59e30b356b686bd3144efe24f875ebc8`. These are the ten
runtime declarations in the survey, including nested ones. An eleventh,
`Status` at `tsbuild.ts:58`, contains fourteen interfaces and is type-only.

| Declaration | Location | Namespace requirements |
| --- | --- | --- |
| BuilderState | builderState.ts:100 | 25 functions, three interfaces; merges with an interface of the same name. Its namespace structure fits the subset. |
| JsxNames | checker.ts:54223 | Ten exported string constants. The wrapper fits; their branded casts are an independent obligation. |
| ReactNames | checker.ts:54236 | One exported string constant, likewise branded. |
| Debug | debug.ts:113 | 75 functions, mutable exported logging/debug state, private caches, a class, and nested log. Live state is covered; its class, overloads and runtime merging remain; calls use reachability checks. |
| Debug.log | debug.ts:137 | Four functions merged into the callable `log`; needs a callable object whose attached properties keep identity and ownership. |
| BinaryExpressionState | factory/utilities.ts:1273 | Nine functions, seven exported, coexisting with a generic callable type alias. Its namespace structure fits the subset. |
| Parser | parser.ts:1437 | 437 functions, 30 variable statements, two enums and a nested namespace. The singleton deliberately uses var and uninitialized mutable parser state. Direct singleton storage, enum scopes and flat factory var binding structure are covered; binding value types and overloads remain; calls use reachability checks. |
| Parser.JSDocParser | parser.ts:8790 | Six functions and two enums; its isolated namespace declaration structure now lowers. Surrounding parser state and function bodies remain independent obligations. |
| IncrementalParser | parser.ts:9946 | 13 functions, an interface and an enum; its enum scope is covered; two overload signatures and function bodies remain. |
| tracingEnabled | tracing.ts:37 | 12 functions, nine variable statements and an enum; private mutable tracing state and namespace-object escape through `tracing = tracingEnabled`. Needs a real runtime object, not just qualification. |

Observed: five of the ten declaration-shape probes lower, with external types
and function bodies normalized as described below. This does not prove their
original function bodies compile, their casts are sound, or TypeScript runs.
Parser, IncrementalParser, Debug and tracing retain explicit blockers. Rewriting them into modules may work after an
explicit porting review; it is not an unconditional behavior-preserving rewrite.

## Why qualification is sound

Flatten declarations while retaining checker symbols, never textual names.
Thus `First.read` and `Second.read` remain distinct, private state stays attached
to its original scope, and imported aliases resolve to the same binding.
Function values use the existing canonical forwarder, preserving `===`.
A structural object typed `typeof N` uses ordinary property calls; its members
must never be redirected to N's functions merely because their types match.

The subset forbids observing the namespace container or replacing exported
functions. Mutable exported variables use the same live singleton storage.
Consequently its object can be omitted, while its owned values and private
state live in ordinary module storage. Existing flow, borrow, reuse, region,
exception and cycle analyses see ordinary declarations and calls; none is
disabled. Mutable contents of exported objects still obey ordinary slot rules.

JavaScript hoists namespace variables and assigns exports in stages. Before
module evaluation, a private generated boolean is false for each runtime
namespace; completing its body sets that boolean to true. The static walk follows
known function declarations, immutable aliases, callbacks, default arguments and
recursive helpers by checker symbol. It refuses reads of pending namespaces that
those calls can reach. An unrelated Map construction or pure helper call may run
before a namespace, including inside a namespace initializer.

Parameters, mutable function values, virtual methods and constructors can leave
an unresolved edge. Their eventual qualified namespace reads check the generated
boolean using ordinary IR shared by both backends, including reads of pure
exported functions. No edge is assumed harmless merely because its target is
unknown. Calls check the callee before argument effects; simple writes evaluate
the right side before their final write check. Direct known premature reads stay
NotYet. Partial namespace containers still have no representation: an unresolved
read while the body is incomplete stops loudly, even if some exports have been
assigned. Neither container escape nor callable/class merging is admitted.

Node fixtures exercise private mutable state, nested generic functions,
same-spelled constants/functions, type/value name coexistence, detached function
identity, structural copies and module import/evaluation order.
[Validation](parameter-properties-namespaces-validation.md) records the gate,
counts, mutants and limits.

## ESM namespace qualifications (October 8)

The `performance` namespace in TypeScript's generated barrels is an ESM namespace
import, rather than a `namespace` declaration. Qualified members now resolve to
the checker's actual export bindings through named and star re-exports. A
structural object with the same type keeps ordinary object semantics. Function
declarations retain ESM hoisting and canonical detached identity; mutable module
exports are live reads of their original storage.

The existing import-cycle initialization proof applies to these reads. A provider
that has completed before the consumer runs needs no readiness check. Deferred
reads and reads whose ordering is undecided retain the runtime readiness check
at the original binding, including reads reached through a helper. ESM namespace
object escape, reflection, computed access and functions observing the namespace
receiver remain named NotYet. A narrowed boxed union member also remains NotYet
on this base; reading it into an ordinary local uses existing checked narrowing.

Three sampled `reading Debug` census signatures did not reproduce with complete
project bindings. Two imported `reading performance` signatures did reproduce
and disappear with this qualification fix. This is a five-site replay result,
not a claim that all 557 Debug or 51 performance roots lower. Exact sites, other
blockers, Node fixtures and real readiness mutants are recorded in
[the namespace-reads report](../stage3/notyet-namespace-reads/REPORT.md).

## Nested receiver scopes in parser Debug (October 8)

The parser proof's debug.ts:429 and 526 receiver probe contains an object method
inside a namespace function. That method owns its receiver; it does not use the
namespace receiver. The namespace scope check now follows lexical arrows but
stops at nested ordinary functions, object methods and classes. Their existing
lowering paths check their receivers. ESM-qualified functions use the same scope
boundary. Actual namespace-owned `this` retains its named receiver refusal.

The unchanged parser probe lowers in both backends. An executed returned method
reads and then mutates its object's flags, separately from Debug's flags. A mutant
substituting Debug's flags for the method receiver is caught by Node stdout in
both backends. [Parser-stop evidence](../stage3/namespace-parser-stops/REPORT.md)
records the exact programs, baseline refusals, validation and remaining scope.

## Namespace class declarations (October 8)

Classes inside a namespace use the existing class registration, instance layouts,
methods and constructor lowering. Their constructor binding becomes ready in
declaration order, including classes with no static members. Construction reads
that binding before its arguments run. Scope remains the checker's symbol scope.
Existing class limitations and the conservative early namespace initialization
refusals remain in force; class/namespace merging remains NotYet.

The parser's DebugTypeMapper declaration probe and an executed two-instance
fixture match Node in both backends. Removing the actual constructor registration
must trigger the existing readiness check. See [the parser-stop report](../stage3/namespace-parser-stops/REPORT.md).

## Qualified callable namespaces (October 8)

A function and one namespace declaration in the same scope may merge. Direct
calls, fixed qualified properties, typeof and strict identity comparisons with
other fixed namespace functions lower. The callable value uses the existing
canonical closure identity; its attached exports retain their checker bindings.
Functions whose receivers are unused can be called qualified or detached with
the same behavior. Namespace-owned this and explicit receiver parameters retain
their named refusals.

This subset emits no callable property container. Passing, returning or storing
the merged callable, reflection, computed properties, export replacement and
function-intrinsic collisions remain NotYet. In particular name, length,
prototype, caller, arguments, call, apply, bind and toString are not represented
as attached exports. Reopened namespaces and class merges still remain NotYet.
Initialization preflight is unchanged on this base.

The unchanged Debug.log parser probe and executed callable/attached function
fixture match Node in native and JavaScript. Redirecting attached calls to the
root function is caught by Node stdout. [The report](../stage3/namespace-parser-stops/REPORT.md)
records all three units and their retained limits; the whole Debug body is not
claimed to lower.

## Stage 3: parser singleton storage

Direct private namespace `var` bindings are hoisted to namespace entry. Their
initializers run in source order; an uninitialized declaration does not overwrite
an earlier assignment. Uninitialized private `let` is also admitted. Storage
whose declared type includes undefined begins with undefined. Storage whose type
excludes undefined stays unready until assigned; reads use the existing ready
check in both backends, rather than treating native zero bits as a typed value.
This check intentionally stops a checker-accepted type lie where Node reads
undefined. Flat object namespace var bindings hoist their identifier leaves and assign them
in source order. Arrays, nested patterns, defaults and rest remain NotYet.
Function-local var remains unsupported. Initializer calls use the namespace
reachability and readiness rule above.
Mutable exports are admitted by the later live-storage step.

The parser-state fixture observes a var before its declaration, initializes a
built string through an exported function, and resets optional token state.
The unready fixture proves a typed number read before assignment stops with the
same ready check in native and generated JavaScript. This is singleton storage
support, not proof that parser.ts or all its function bodies compile.

## Stage 3: namespace enums

The constant-valued enum implementation from `codex/enums` (`7127080`) is
integrated through the split lowering files. Ordinary and const enums may be
declared directly in a namespace, including nested JSDocParser. Symbols retain
the enum's declaration identity. Ordinary objects use the existing forward and
reverse maps and ownership rules; const members inline. Computed, ambient,
merged and prematurely observed enums remain NotYet. Namespace enums now follow current main's numeric-enum policy: a whole numeric
enum is open to numbers, while string enum identity, member-specific literal
promises and writable enum-object views retain their checks.

The parser-enum fixture covers two parser enum scopes, a nested JSDoc parser,
const member inlining, numeric aliases/reverse mapping, and IncrementalParser.
A mutant replacing SourceElements with 99 must finish cleanly and disagree with
independent Node stdout.

## Stage 3: ordered nested bodies

Namespace statements now run in source order at module evaluation, with nested
bodies evaluated at their declaration. Console calls, local assignments and
ordinary control flow are admitted. Calls and construction are admitted when they do not reach a pending namespace.
Known premature namespace reads stay NotYet; unresolved targets carry runtime
readiness checks at their eventual namespace accesses.
Direct singleton var is hoisted; var hidden inside namespace control flow remains
NotYet until its block and function scope are modeled. Early qualified namespace
reads remain NotYet, including reads before a nested namespace's body finishes.

The ordered-body fixture prints before the module namespace, at Parser entry,
inside JSDocParser, after the nested body, and after Parser completes. Reordering
two output statements is caught by Node stdout comparison.

## Stage 3: live mutable exports

Exported let and var bindings now use the same singleton storage for private
identifier access, qualified access, and imports. Plain, compound and increment
writes update that storage; qualified reads preserve checker narrowing and its
runtime checks. Exported function identities still cannot be replaced. Namespace
objects cannot be aliased to write through a structural view.

The Debug-state fixture checks inside/outside numeric writes, compound updates,
increment, optional built-string replacement and narrowing, and boolean state.
Changing an outside write from 3 to 30 is caught by Node stdout. This admits the
state portion of Debug; its class and callable log merge remain separate limits.

## Stage 3: escaped tracing object remains NotYet

`tracing = tracingEnabled` requires a real namespace container. Flattening its
qualified names does not implement container identity, live exported storage
through aliases, callable receivers or JavaScript's staged export properties.
A plain object of current member values would snapshot mutable exports and could
change receiver behavior. That is not a small sound extension of this lowering,
so escape, reflection and observation keep an explicit NotYet explaining the
missing runtime container. `tracing_escape.a` preserves the compiler's alias
shape, with singleton state and an enum, and checks that exact reason.

The escape-guard mutant removes only the container observation guard. The
regression must lose the explicit reason and fail before C emission; the later
unbound-name NotYet is not mistaken for runtime object support.

## Stage 3: callable and constructor merges remain NotYet

Debug.log's function/namespace merge requires one callable object whose attached
functions preserve identity and receivers. Class merging similarly needs one
constructor with staged static properties. The flattened singleton bindings do
not prove those behaviors, so each form has its own explicit NotYet reason.
A class declaration inside a namespace also remains NotYet until constructor
registration and namespace initialization are proven. `debug_log.a`,
`class_merge.a` and `debug_class.a` retain those independent shapes.

Removing the function-merge or class-merge branch is caught by the regression
requiring its specific reason. These are diagnostic guard mutants, not evidence
that native callable namespace objects work.

## Ten declaration shapes, checked again on October 7

`stage3/namespaces/shape-census.cjs` uses stock npm TypeScript 6.0.3 on the pinned
source. Its committed census records source hashes, locations, member-kind
counts, generic function counts and overload counts. The generated .a inputs
retain names, exports, var/let/const, missing initializers, eager initializer
calls, binding patterns, generic arity, bodyless signatures and nesting. They
replace external types, function bodies and enum expressions, and preserve
ordinary string/boolean literals. BuilderState's interface merge,
BinaryExpressionState's callable type alias, Debug.log's function merge and
tracingEnabled's alias are retained. These are declaration-shape experiments,
not adapted original compiler implementations.

| Shape | Result | Remaining first blocker |
| --- | --- | --- |
| BuilderState | Lowers | Original types and bodies untested |
| JsxNames | Lowers | Original branded casts untested |
| ReactNames | Lowers | Original branded casts untested |
| BinaryExpressionState | Lowers | Original callable types and bodies untested |
| Parser.JSDocParser | Lowers | Original bodies and surrounding parser state untested |
| Parser | NotYet | Bodyless overload signature after safe initializer calls |
| IncrementalParser | NotYet | Two overload signatures without bodies |
| Debug | NotYet | Namespace class; overloads and log merge also remain |
| Debug.log | NotYet | Callable namespace object |
| tracingEnabled | NotYet | Escaped runtime container |

Parser also has a flat destructured var factory binding, now admitted, and 18 bodyless overload
signatures. The isolation deliberately retains them; normalizing those away
would overstate progress. IncrementalParser's enum now lowers in the cut-down
runtime fixture, but that is weaker than its complete declaration shape.
The day 3 proof of unchanged parser.ts running natively is **not achieved**.

Regressions live in `TestTscNamespaceDeclarationShapes`. All five accepted
shapes are also oracle fixtures. The [validation report](../stage3/namespaces/REPORT.md) records commands,
outputs, commits, mutants and the setup retry.

## Original-source fixtures after the fixture branch merge

The fixture branch a75ba00 is merged with a merge commit. Its twelve unchanged
slices cover all eleven census declarations, including the type-only Status
namespace. All twelve were built at the base, each of the six namespace steps,
and the merge. The accepted count moves 1 -> 2 -> 3 -> 3 -> 4 -> 4 -> 4;
the fixture branch merge keeps it at 4. Singleton state adds Parser countNode,
enums add Status, and mutable exports add Debug shouldLog. BuilderState
releaseCache already compiled on the base. Every successful native execution
matches independent Node; the merged successes also pass sanitizers and leak
checks. These are selected real executable bodies, stronger than normalized
shapes but weaker than complete namespace or parser.ts support. See the
[per-step matrix and all twelve outcomes](../stage3/namespaces/REAL_FIXTURES.md).

## Original Parser.JSDocParser slice

The original-source fixture 08 now runs natively. A direct returned assignment
of a number or boolean namespace singleton evaluates its right side once,
stores it, then returns the stored scalar through ordinary statement IR.
Reference/optional storage and assignment expressions in other positions stay
NotYet. The unchanged nextTokenJSDoc and parseOptionalJsdoc bodies print
`true 1`, `false 1`, `3 3`, matching independent Node. Dropping the returned
store is caught by Node stdout with clean native execution and sanitizers.
Removing the scalar-only boundary is caught by its NotYet regression.

## Parser factory bindings and located missing symbols

The cut at parser.ts:1472 binds factory methods through a flat object var pattern.
The checker has symbols for its identifier leaves, not the pattern itself.
Namespace hoisting now registers those leaves, and destructuring assigns the
hoisted singleton storage. Unsupported nested/default/rest bindings retain
NotYet. A remaining missing-symbol declaration error is a structured NotYet
with its source location, so census consumers can locate it.
The factory oracle compares both callable bindings with Node; a same-signature
wrong-method mutant is caught by stdout. Replacing the structured diagnostic
with an ordinary error is caught by its location regression.

## Original bodies at f893faf2

At f893faf2 the twelve-slice matrix was five Compiles, five NotYet and two Refused.
BuilderState canReuseOldState and Debug.log also need the boolean-only operator
decision; JsxNames and ReactNames need a primitive-brand construction decision.
Tracing still requires a canonical runtime container with alias-write semantics;
a snapshot plus flattened calls can disagree with Node. The two original
Refused fixtures were left unchanged. BinaryExpressionState appears overly
strict for alpha-equivalent generic binders; IncrementalParser is required by
the current exhaustive-switch policy, though its post-switch return is sound.
The [exact programs, diagnostics, Node outputs and narrow acceptance judgments](../stage3/namespaces/DECISIONS.md)
are recorded for integration review. No language decision was made here.

## Parser discovery Debug probe after batch 3

The exact debug.ts:26 singleton probe from parser discovery row 11 now lowers
and prints `false`, matching Node in native code and checked JavaScript. The
current-main merge preserves namespace ready checks together with main's
undefined-observation and union-narrowing checks. Both wrong initialization and
restoring the old observation guard are caught by Node regressions. The twelve
original namespace slices still have five Compiles, five NotYet and two Refused;
this minimal Debug probe does not establish full Debug or parser execution.
[Integration repair, exact source, mutants and fresh counts](../stage3/namespaces/DEBUG_PROBE.md)
record the evidence.

## Reachable initialization calls from core.ts

The exact `new Map<never, never>()` before Debug from parser probe 77aaea49 now
prints `0:false`, matching Node. Empty never-key/value maps use storage slots
without inventing inhabitants; writable widening remains refused. Direct and
helper calls reaching pending Debug stay located NotYet. Indirect state reads,
pure function calls, void calls, const-enum reads and writes stop with the same error and effects
as Node in the new fixtures. Three compiler mutants prove the static refusal pin, runtime check and
const-enum hook independently.

The refreshed original-source matrix is six Compiles, three NotYet and three
Refused. Current main's open numeric-enum policy admits fixture 09; its source
was not edited. Main's checked-cast policy changes 02 and 03 from NotYet to
Refused. Fixture 06 and its refusal are unchanged. The initialization unit
clears the separate core Map probe; it does not change those three policy rows.
[Reachability report, programs, diagnostics, counts and validation](../stage3/namespaces/REACHABILITY.md)
record the observations and remaining limits.

## Memoized initialization reachability

Namespace initialization discovers each statically known function body once.
A Tarjan active stack finds strongly connected call components. Each component
receives the union of its own namespace and regular namespace-enum reads and
its outgoing components' reach sets; no partial cycle result is cached.
Module evaluation checks these cached sets against the declarations initialized
at that call site. Unknown targets retain the existing runtime readiness checks.
Graph discovery is linear in function bodies and call edges; materializing
transitive sets additionally costs their union operations.

The exact 24-level, two-call dead-branch witness from parser proof 33110b86
emits C in milliseconds. A deterministic expansion counter pins 13 bodies at
depth 12 and 25 at depth 24, including repeat queries; dropping completed-body
memoization produces 8,191 walks at depth 12 and fails the pin. A three-function
cycle must give every member both reachable namespaces. Native cycle execution
matches Node after initialization; the same cycle before initialization remains
located NotYet. The [component report](../stage3/namespaces/CALL_GRAPH.md) records
exact timings, sources, diagnostics, mutants, commands and the unchanged matrix.

Parser and IncrementalParser declaration-shape tests now require the same
bodyless-overload outcome as the module-level implementation capability. This
branch still returns the existing bodyless-function NotYet; integrations that
admit checked overloads must also lower these namespace shapes. Namespace regular
enum initialization has its own reachability pin, independent of the module
analysis, so the early-enum refusal survives integration of module ready checks.


The October 7 paired meter's 55 `reading Debug` NotYet sites measure as zero on
this branch across the same 79-source-file adapted input. This is a latent
measurement on a checker-rejected program. Thirteen other NotYet sites remain
inside debug.ts; the production whole-file build still fails its dependency
checker errors. The component report includes the exact remaining-site ledger,
raw findings and failed production build. Existing post-initialization and
unresolved-before-initialization Debug reads remain held to Node.
