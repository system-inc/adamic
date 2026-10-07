# Array holes claims

Base c762b555e7c0b39b105e2d9208732a5a311b2e6f; oracle Node 24.19.0.
The inventory was pushed before implementation in 1b86df6e. The scanner probe
from 8d1b1141289fcd7340231b6950eacf6aa7f3a91f prints 4 on both backends.

## Representation and admission

Length-form arrays store present numeric properties in the existing ordered
map. Missing entries are holes; a present entry containing undefined is still
present. Construction takes constant space even at length 4294967295. Reference
slots retain their usual ownership. For sparse arrays the otherwise unused
capacity field counts holes; an indexed write decrements it only for a new own
slot. Dense arrays retain their storage and dense-only programs retain their
original generated indexed access and assignment paths. In a compilation that
contains a length constructor or length write, accessors branch between dense
and sparse storage. This conservative policy does not prove that dense arrays
in such a mixed compilation incur no overhead.

Every array in such a compilation is considered potentially holey, including
aliases, arguments, fields, returns and callback inputs. The lowerer checks both
source operations and generated IR. Unsupported consumers cause a named NotYet
with an operation reason; there is no local-variable-only hole test. This can
refuse an unrelated dense operation in the same program. Untyped any[] arrays
may expose length but cannot escape into a differently represented array.

## Existing operation contracts

"Refused" means refused whenever the compilation may contain holes under the
rule above, unless an independent earlier type/lowering restriction applies.
The fixtures execute the rejected sources on Node as well as testing refusals.

| Existing operation | Hole contract |
| --- | --- |
| new Array(n), Array(n), new Array<T>(n) | Built for one number and a supported slot representation; integral 0..4294967295 creates holes; invalid lengths throw catchable nominal RangeError, message Invalid array length |
| Other constructor forms, Array.of | Named NotYet; no element-list constructor was added |
| length read | Built, counts absent slots |
| length = number | Built, growth adds holes; shrink deletes own indexed properties and releases references; non-index properties survive; invalid length throws without changing the array |
| indexed read, optional indexed read | Built, absence yields undefined |
| indexed write and compound update | Built through hole-aware slot lookup/store; fills an own slot; valid index can grow length; non-index numeric properties do not grow length |
| at | Refused |
| push, pop | Refused |
| fill, new Array(n).fill(v) | Refused in hole-containing compilations; the existing immediate complete-fill dense specialization remains available in dense-only programs under its existing restrictions |
| slice, concat, splice, reverse | Refused |
| sort, toSorted | Refused |
| map | Built, visits only existing indices and preserves absent output slots; snapshots initial length and rechecks presence after callback mutations |
| forEach, filter, some, every | Built, skip absent indices; filter produces dense output; callbacks observe original indices |
| reduce with initial value | Built, skips absent indices |
| reduce without initial value | Existing independent NotYet; no initializer-free reduction added |
| find, findIndex, findLast, findLastIndex | Refused |
| indexOf, lastIndexOf | Built, skip holes; explicit undefined remains searchable; existing fromIndex coercion retained |
| includes | Built, treats absent indices as undefined; present slots use existing SameValueZero comparison |
| join, toString, String(array) | Built for existing scalar join representations, holes and undefined produce empty text |
| nested join | Refused |
| flat, flatMap, copyWithin | Refused |
| with, toReversed, toSpliced | Refused |
| for-of, values(), entries() | Refused |
| keys() in for-of | Built, visits every index; does not read elements |
| array spread and call spread | Refused; includes Math and String spread consumers |
| destructuring fixed elements | Existing indexed-read lowering uses hole-aware reads; defaults retain existing restrictions |
| rest bindings | Refused by the generated slice/spread consumer checks |
| Array.from({length}, mapper), Array.from(iterable) | Refused in hole-containing compilations, including the existing dense mapper form |
| Array.isArray | Existing type-based observation remains valid; untyped array escape restrictions still apply |
| JSON.stringify | Refused |
| console.log(array), console.error(array) | Existing string-only console type check refuses array arguments; Node empty-item inspection was not implemented |
| in | Existing array operand refusal; no property-presence implementation added |
| hasOwnProperty | Refused by source receiver check, including methods that the old lowerer could fold |
| Object operations on arrays | Refused by existing operand restrictions and the conservative object-consumer check |
| constructor / Array.prototype method name, length, typeof | Existing observations remain independent of element presence |
| Map/Set construction, collection consumption of arrays | Refused |
| Buffer/Node host array consumers | Refused |

Already independently refused: array literal elisions, delete, shift, unshift,
reduceRight, toLocaleString, unsupported element types, and compound length
writes. Prototype mutation and inherited indexed properties remain outside the
existing admitted language. The implementation supports own holes within that
language; it does not claim arbitrary JavaScript Array object semantics.

## Evidence

Accepted fixtures cover length, boundary lengths, indexed writes, references,
explicit undefined, callbacks, searches, joins, toString, keys, resizing and
callback mutation. Operation fixtures exercise all-holes, a slot at the start,
middle or end, and a partly filled array. Accepted sources compare sanitized
native, release native and generated JavaScript with Node; native leak checks
use detect_leaks=1 only on Linux. Thirty independently executed Node sources
verify named compile/type refusals. Alias tests cover locals, fields, returns,
parameters, operations hidden beneath length reads, and untyped escapes.

The requested forEach, map, join, RangeError-bound and hole-count mutants each
produce a Node-output disagreement. Disabling the conservative admission guard
makes the alias refusal test fail. Timing and test262 observations are recorded
in internal/native/performance/array-holes/REPORT.md.
