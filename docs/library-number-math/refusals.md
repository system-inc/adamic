# Refusal census before implementation

Each count is the first refusal in one adapted test. A first refusal can mask further language requirements. Counts are observations from the independent runner, not estimates. Number totals 70; Math totals 38.

## built-ins/Number

| Count | Exact normalized reason | Disposition |
|---:|---|---|
| 37 | not yet: toString through an object view (a value with a different native representation may be hidden by the view) | Library: immediate Number internal slot, built |
| 6 | not yet: new an Identifier | Language dependency: stored boxed objects and dynamic ToPrimitive |
| 6 | refuses var | Language boundary: var |
| 3 | refuses isPrototypeOf | Language dependency: observable prototype chains |
| 2 | not yet: a try around toString, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | Language dependency: catchable library RangeError and cleanup paths |
| 2 | not yet: for...in without a proven fixed plain-object origin (arrays, prototypes and absent synthetic fields cannot be enumerated soundly) | Language dependency: enumerating intrinsic/prototype objects |
| 2 | not yet: reading isFinite | Library: global finite predicate, built |
| 2 | refuses a method read as a value (toString would lose its object, and this with it) | Language dependency: prototype reflection/call or prototype mutation |
| 1 | not yet: Number conversion of an object | Language dependency: dynamic ToPrimitive, boxed allocation, and abrupt completion |
| 1 | not yet: a BinaryExpression with a string and a number | Language dependency: implicit mixed-type addition |
| 1 | not yet: a try around toExponential, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | Language dependency: catchable library RangeError and cleanup paths |
| 1 | not yet: a try around toFixed, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | Language dependency: catchable library RangeError and cleanup paths |
| 1 | not yet: a try around toPrecision, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) | Language dependency: catchable library RangeError and cleanup paths |
| 1 | not yet: a value argument to toExponential | Library: explicit undefined formatting argument, built |
| 1 | not yet: valueOf through an object view (a value with a different native representation may be hidden by the view) | Library: immediate Number internal slot, built |
| 1 | refuses == | Language boundary: == |
| 1 | refuses a method read as a value (toFixed would lose its object, and this with it) | Library: intrinsic function own-property metadata, built |
| 1 | refuses the void operator | Language boundary: the void operator |

## built-ins/Math

| Count | Exact normalized reason | Disposition |
|---:|---|---|
| 29 | not yet: new an Identifier | Language dependency: growing arrays |
| 4 | error TS2550: Property '…' does not exist on type '…'. Do you need to change your target library? Try changing the '…' compiler option to '…' or later. (tsc: TS2339) | Checker classification mismatch: stock tsc rejects TS2339, not TypeScript-valid work |
| 2 | refuses var | Language boundary: var |
| 1 | refuses Math.random | Library contract: nondeterminism remains refused |
| 1 | refuses a method read as a value (max would lose its object, and this with it) | Library: intrinsic function type and length metadata, built |
| 1 | refuses a method read as a value (min would lose its object, and this with it) | Library: intrinsic function type and length metadata, built |

## Library work, largest first

| Tests recovered | Feature |
|---:|---|
| 38 Number | Immediate boxed Number formatting and valueOf, including explicit prototype .call receivers |
| 2 Number | Global isFinite dispatch, using the existing exact Number conversion |
| 2 Math | Intrinsic function typeof and length metadata; name metadata covered in fixtures |
| 1 Number | Explicit undefined in numeric formatting |
| 1 Number | Numeric intrinsic functions own name and length properties |

The Number slot lowering and four edge fixtures were brought forward from `735a702`. The rest is new on current main. Existing conversion, radix, decimal formatting, and Math algorithms are called unchanged. No new V8 port was required, so THIRD_PARTY_NOTICES.md did not change.
