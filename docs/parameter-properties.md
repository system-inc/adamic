# Parameter properties in Adamic 0.2

Decision for Kirk, October 6, 2026: admit constructor parameter properties,
including `public`, `private`, `protected` and `readonly`, defaults and optional
parameters. They are ordinary instance slots, with the same nominal identity,
variance, ownership and visibility rules as written fields. Generic classes
monomorphize them with the rest of their fields.

This supersedes the parameter-property entry in [0.1](0.1.md). Current main still
sets `erasableSyntaxOnly`; this branch disables it in the loader and tsconfig.
Enums remain explicitly refused on this branch until the independent enum unit
is integrated. The source oracle uses Node's transform mode for this syntax.

The pinned TypeScript 6.0.3 compiler has six parameter properties, all readonly
callbacks in `BinaryExpressionStateMachine` at `factory/utilities.ts:1415..1420`.
This supports that declaration shape. It does not establish that the whole
compiler or every callback return representation now lowers.

## Initialization and soundness

Node's transform declares these slots before the written class fields. Base
field initializers run first, then parameter defaults, then parameter-property
stores, then the constructor body. For a derived class: its defaults and code
before `super` run first; base construction completes; derived parameter-property declarations reset their
slots, including inherited names; written derived field initializers run; parameter properties are copied from their current lexical parameters;
the rest of the derived body runs. A parameter reassigned before `super` is
copied with its new value. A readonly property does not make that lexical
parameter readonly.

The checker supplies distinct parameter and property symbols. Lowering binds
the lexical symbol to the incoming/defaulted local and uses the property symbol
for the instance slot. This keeps `x` and `this.x` distinct after construction.
Each stored reference is retained by the object and released by the existing
class destructor chain. Weak slots keep weak handles, not strong owners.

The checker enforces accessibility and direct readonly writes. Adamic also
refuses readonly-to-writable views and narrowed mutable overrides, including
parameter properties. The ordinary cycle proof sees these slots and their
implicit writes; no memory analysis is disabled.

One checker-admitted hole is a default such as `y: number = this.x` when `x` is
a parameter property. Its implicit store has not run, so JavaScript reads
undefined. Adamic refuses that default with the initialization explanation.
Use `y: number = x`, or assign it in the constructor. Defaults and field
initializers may read only fields already initialized; this proof is conservative
about methods and deferred callbacks.

Named callbacks with an explicit dynamic `this` parameter remain NotYet, as
function expressions with such a receiver already were. Use an arrow or pass
the receiver as an ordinary parameter.

Constructors returning a replacement object, rest/destructured parameters,
unsupported slot representations and conditional/repeated `super` remain the
existing `NotYet` cases on this main-based branch. Use written fields and a
factory when these are needed. The supported constructors must complete through
the existing class initialization path.

Node fixtures cover defaults and side effects, base/derived order, mutable and
readonly slots, private/protected reads in their allowed scopes, generic
hierarchies, dispatch, optional numbers, callbacks and a Weak child-parent link.
Validation run evidence stays outside the checkout.
