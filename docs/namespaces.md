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
| Debug | debug.ts:113 | 75 functions, mutable exported logging/debug state, private caches, a class, and nested log. Needs observable live storage and runtime merge support. |
| Debug.log | debug.ts:137 | Four functions merged into the callable `log`; needs a callable object whose attached properties keep identity and ownership. |
| BinaryExpressionState | factory/utilities.ts:1273 | Nine functions, seven exported, coexisting with a generic callable type alias. Its namespace structure fits the subset. |
| Parser | parser.ts:1437 | 437 functions, 30 variable statements, two enums and a nested namespace. The singleton deliberately uses var and uninitialized mutable parser state. Needs explicit initialization/TDZ policy and those enum scopes. |
| Parser.JSDocParser | parser.ts:8790 | Six functions and two enums; nested access and ownership of surrounding parser state. |
| IncrementalParser | parser.ts:9946 | 13 functions, an interface and an enum; needs namespace-scoped enums as well as existing body features. |
| tracingEnabled | tracing.ts:37 | 12 functions, nine variable statements and an enum; private mutable tracing state and namespace-object escape through `tracing = tracingEnabled`. Needs a real runtime object, not just qualification. |

Inference: the subset removes namespace syntax itself for four of the ten
runtime wrappers. It does not prove their function bodies compile, their casts
are sound, or the TypeScript compiler runs. The parser, debugging and tracing
cases require a larger design. Rewriting them into modules may work after an
explicit porting review; it is not an unconditional behavior-preserving rewrite.

## Why qualification is sound

Flatten declarations while retaining checker symbols, never textual names.
Thus `First.read` and `Second.read` remain distinct, private state stays attached
to its original scope, and imported aliases resolve to the same binding.
Function values use the existing canonical forwarder, preserving `===`.
A structural object typed `typeof N` uses ordinary property calls; its members
must never be redirected to N's functions merely because their types match.

The subset forbids observing the namespace container or replacing its exports.
Consequently its object can be omitted, while its owned values and private
state live in ordinary module storage. Existing flow, borrow, reuse, region,
exception and cycle analyses see ordinary declarations and calls; none is
disabled. Mutable contents of exported objects still obey ordinary slot rules.

JavaScript hoists namespace variables and assigns exports in stages. Until
partial namespace objects and early reads are represented exactly, Adamic
conservatively refuses calls or construction before all runtime namespaces in
the dependency closure are initialized, as well as namespace reads before their
own initialization. Put namespaces before executable module code. Initializer
calls and executable namespace bodies remain NotYet; function bodies are deferred.
This preserves accepted initialization order without substituting a native TDZ
error for JavaScript's undefined property behavior.

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
undefined. Function-local var, destructuring namespace declarations, mutable
exports and calls during namespace initialization remain outside this step.

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
merged and prematurely observed enums remain NotYet. The existing closed-domain
checks also guard namespace enum values through writable views and containers.

The parser-enum fixture covers two parser enum scopes, a nested JSDoc parser,
const member inlining, numeric aliases/reverse mapping, and IncrementalParser.
A mutant replacing SourceElements with 99 must finish cleanly and disagree with
independent Node stdout.

## Stage 3: ordered nested bodies

Namespace statements now run in source order at module evaluation, with nested
bodies evaluated at their declaration. Console calls, local assignments and
ordinary control flow are admitted. Arbitrary calls and construction while any
runtime namespace is pending remain NotYet: they could see partially assigned
exports through deferred code. Initializer calls have the same restriction.
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
