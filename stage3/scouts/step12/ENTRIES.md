# Object.entries on an unproven shape

No ruling is made here. The reported scanner.ts:432:59 coordinate cannot be
located in this main's evidence. The same diagnostic is reproduced by
fixtures/entries.a:6:35. Pinned upstream has Object.entries(textToKeywordObj)
at scanner.ts:222:31; the raw gathered closure has it at 185:31. An additional
inline `as const` argument occurs upstream at 4073:44, raw closure 3519:44.
Those are different expressions and must not be conflated by a shifted line number.

The source Node witness takes a structural view of an object. The view names
only a numeric constructor property; the object also owns extra. Returning
`new Map(Object.entries(view))` as Map<string, number> looks sound to TypeScript.
Changing the hidden property's value from 1 to "lie" changes Node's observed
entry value type from number to string. Both programs reach Adamic's same
NotYet. This is an actual value mutation, not an output-file mutation.

ECMAScript's Object.entries first performs ToObject, then
EnumerableOwnProperties(obj, key+value), then CreateArrayFromList. Enumeration
uses actual own keys and enumerable descriptors, skips symbols, and calls Get
for each value. Ordinary own keys order array indexes numerically, then other
strings by creation order. Inherited keys are absent. A getter can run between
entries and change later descriptors or values. A snapshot of declared fields
cannot silently stand in for this behavior. See evidence/spec.json, fetched
October 8, and the cited algorithm URLs therein.

## Options inside Adamic's rules

1. **Prove the actual shape.** Track allocation provenance through const names,
   structural views, parameters, imports and calls. Prove every possible
   allocation has precisely known plain own data properties, including hidden
   properties; retain actual enumeration order and present/absent information.
   Prove aliases cannot add, delete, redefine or retype fields before enumeration.
   Evaluate the argument once; construct fresh pairs and a fresh outer array.
   Infer a union of all actual enumerable value types, not only visible fields.
   If the requested typed destination cannot hold that union, refuse that use.
   A local const alias alone is insufficient: it freezes a binding, not its object.
   This could admit the keyword literal through its MapLike view without trusting
   MapLike as a closed shape. Constant-folding the constructor name is a separate
   proof obligation, not permission to ignore arbitrary computed keys.
2. **A loud runtime check.** Use an actual-shape descriptor capable of inspecting
   all own string properties, their presence and data/accessor descriptors. Check
   each value that will be returned against a sound runtime representation of
   the destination type, including enum membership if that type requires it.
   Run the identical check in native and Adamic's JavaScript backend. A failing
   check panics with a location and reason; it never drops extra properties or
   coerces their values. Ordinary source Node remains the external oracle for
   admitted successful programs; checked failures require the language's explicit
   checked-semantics contract. If a descriptor, getter, proxy, prototype behavior,
   type or value cannot be represented and checked faithfully, refuse it.
   Checking once and then rereading values is unsound when getters or aliases
   can change them. Checking observed values during enumeration avoids that
   particular race but still needs faithful descriptor/Get ordering.
3. **Refuse the unproven operation.** Keep the current NotYet until a representation
   exists; make it Refused only after a language ruling forbids the shape.
   Suggest a Map or explicitly constructed typed entries as a source rewrite,
   and hold that rewrite against the scanner corpus and full upstream baseline.
   The scout makes no such upstream rewrite.

The first option needs compiler proof work. The second needs a runtime and
language contract. Neither is implemented by the comparator package. The third
is the current observed behavior, not a decision that reflection is forbidden.

## Questions for @system_adamic, undecided

- Is Object.entries allowed through an open structural view when allocation
  provenance proves every actual enumerable property and its value type?
- Should its inferred values include hidden actual fields, or should the compiler
  require an exact-shape capability before typed reflection is allowed?
- When proof fails, do we permit a checked reflection operation that panics on
  extra incompatible values, or require a source rewrite/refusal?
- Must the first implementation support accessors, symbols and proxies, or may it
  loudly refuse everything beyond proven ordinary objects with data properties?
- For enum-valued records, what counts as proof of membership at this boundary,
  and should a failed membership check panic rather than accepting a number?
- Which source snapshot contains the reported scanner.ts:432:59 stop, and which
  three unchecked assertion expressions does land-area-next assign to owners?
