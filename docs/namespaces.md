# Namespaces in Adamic 0.2

Decision for Kirk, October 6, 2026: a sound qualified-name subset exists, and
this branch implements it. Admit a single module-scope declaration, nested
namespaces, interfaces/type aliases, ordinary or generic functions, exported
constants, and initialized private `let`/`const` bindings. Namespace functions
can be called, detached and compared by identity. Names can coexist with an
interface or type alias of the same spelling. Type-only namespaces erase.

The namespace object's identity, reflection, escape and mutation are outside
this subset. Reopening, function/class/enum merging, mutable exports, ambient
namespaces, `var`, uninitialized state and executable namespace bodies are
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
