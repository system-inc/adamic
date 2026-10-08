# Detached own-property helper

The library's `Object.prototype.hasOwnProperty` can be read into a const or
used directly when it is called only through `.call(target, key)`. Recognition
uses checker symbols, not the spelling of a user function. The call lowers to
Object.hasOwn semantics: a missing `toString` is false, an own `constructor`
is true, and a receiver field named `hasOwnProperty` does not override it.

Records use `RecordCall.hasOwn`. Plain objects and supported class instances
use `ObjectCall.hasOwn`, including dynamic string or numeric keys. Target and
key evaluate once in source order. The const is a nonescaping intrinsic: its
runtime local holds a boolean initialization marker. Initialization and capture
checks still run before the target and key; cycle analysis sees the marker.

The parser probe from `codex/stage3-parser-proof` at `2179dd8` is included
unchanged as `detached_own_parser.a`. Its `object` parameter has an opaque object
pointer representation only when every value use is the target of this idiom.
A caller cannot pass an array, dictionary or callable through that parameter.
An explicitly typed record parameter uses the records representation instead.
General representation erasure for TypeScript's `object` keyword remains NotYet.

Other uses of the detached method remain refused: mutable bindings, aliasing,
exports, storage in fields or shorthand properties, bare calls, apply, bind,
observations, saved call methods, optional calls and spread arguments. Unsupported
target representations remain NotYet. Existing descriptor and storage checks
remain in force. Own-key guards in the supported record-read shapes also recognize
this idiom, so discarded inherited snapshots continue using has_own and get_own.

The Node oracle covers the exact parser probe, present and absent keys, inherited
names, own prototype-name entries, the direct form, a shadowing receiver property,
side-effect order, local and captured bindings, guarded reads and early global
initialization. Mutants report inherited names as own for records and objects;
separate checks remove opaque-storage, shorthand-escape and readiness checks.
Whole-parser compilation and general detached functions are outside this unit.
