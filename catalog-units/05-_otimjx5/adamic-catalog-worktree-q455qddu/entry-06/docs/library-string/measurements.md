# String test262 measurements

Pinned test262 commit: `7ab7fafa0003f73fc85c1b95d88094d33f7eb8bd`. The runner compares every passing case with Node. There are 45 newly passing tests, no lost passes, zero disagreements, and zero crashes. Refusal counts can rise when an earlier refusal is removed and a later one becomes visible.

| Run | Pass | Disagreement | Refused | Crash | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| Before | 85 | 0 | 822 | 0 | 316 | 1223 |
| After | 130 | 0 | 777 | 0 | 316 | 1223 |

## Before directory table

| Directory | Pass | Disagreement | Refused | Crash | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| built-ins/String | 0 | 0 | 80 | 0 | 12 | 92 |
| built-ins/String/fromCharCode | 0 | 0 | 14 | 0 | 3 | 17 |
| built-ins/String/fromCodePoint | 2 | 0 | 4 | 0 | 5 | 11 |
| built-ins/String/prototype | 0 | 0 | 5 | 0 | 2 | 7 |
| built-ins/String/prototype/Symbol.iterator | 0 | 0 | 0 | 0 | 6 | 6 |
| built-ins/String/prototype/at | 0 | 0 | 7 | 0 | 4 | 11 |
| built-ins/String/prototype/charAt | 0 | 0 | 26 | 0 | 4 | 30 |
| built-ins/String/prototype/charCodeAt | 0 | 0 | 21 | 0 | 4 | 25 |
| built-ins/String/prototype/codePointAt | 5 | 0 | 5 | 0 | 6 | 16 |
| built-ins/String/prototype/concat | 0 | 0 | 19 | 0 | 3 | 22 |
| built-ins/String/prototype/constructor | 0 | 0 | 2 | 0 | 0 | 2 |
| built-ins/String/prototype/endsWith | 2 | 0 | 17 | 0 | 8 | 27 |
| built-ins/String/prototype/includes | 5 | 0 | 14 | 0 | 8 | 27 |
| built-ins/String/prototype/indexOf | 0 | 0 | 35 | 0 | 12 | 47 |
| built-ins/String/prototype/isWellFormed | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/String/prototype/lastIndexOf | 1 | 0 | 21 | 0 | 3 | 25 |
| built-ins/String/prototype/localeCompare | 0 | 0 | 10 | 0 | 3 | 13 |
| built-ins/String/prototype/match | 0 | 0 | 36 | 0 | 15 | 51 |
| built-ins/String/prototype/matchAll | 0 | 0 | 1 | 0 | 24 | 25 |
| built-ins/String/prototype/normalize | 0 | 0 | 8 | 0 | 6 | 14 |
| built-ins/String/prototype/padEnd | 3 | 0 | 4 | 0 | 6 | 13 |
| built-ins/String/prototype/padStart | 3 | 0 | 4 | 0 | 6 | 13 |
| built-ins/String/prototype/repeat | 1 | 0 | 9 | 0 | 6 | 16 |
| built-ins/String/prototype/replace | 0 | 0 | 41 | 0 | 14 | 55 |
| built-ins/String/prototype/replaceAll | 0 | 0 | 15 | 0 | 30 | 45 |
| built-ins/String/prototype/search | 0 | 0 | 29 | 0 | 14 | 43 |
| built-ins/String/prototype/slice | 0 | 0 | 34 | 0 | 4 | 38 |
| built-ins/String/prototype/split | 0 | 0 | 106 | 0 | 14 | 120 |
| built-ins/String/prototype/startsWith | 0 | 0 | 13 | 0 | 8 | 21 |
| built-ins/String/prototype/substring | 0 | 0 | 42 | 0 | 4 | 46 |
| built-ins/String/prototype/toLocaleLowerCase | 0 | 0 | 24 | 0 | 4 | 28 |
| built-ins/String/prototype/toLocaleUpperCase | 0 | 0 | 22 | 0 | 4 | 26 |
| built-ins/String/prototype/toLowerCase | 4 | 0 | 21 | 0 | 5 | 30 |
| built-ins/String/prototype/toString | 0 | 0 | 2 | 0 | 5 | 7 |
| built-ins/String/prototype/toUpperCase | 2 | 0 | 20 | 0 | 4 | 26 |
| built-ins/String/prototype/toWellFormed | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/String/prototype/trim | 57 | 0 | 70 | 0 | 2 | 129 |
| built-ins/String/prototype/trimEnd | 0 | 0 | 5 | 0 | 18 | 23 |
| built-ins/String/prototype/trimStart | 0 | 0 | 5 | 0 | 18 | 23 |
| built-ins/String/prototype/valueOf | 0 | 0 | 2 | 0 | 5 | 7 |
| built-ins/String/raw | 0 | 0 | 23 | 0 | 7 | 30 |

## Before refusal table

| Reason | Count |
| --- | ---: |
| error TS2345: Argument of type '…' is not assignable to parameter of type '…'. | 135 |
| refuses var | 132 |
| error TS2339: Property '…' does not exist on type '…'. | 109 |
| error TS2769: No overload matches this call. | 55 |
| refuses a method read as a value (trim would lose its object, and this with it) | 54 |
| error TS2554: Expected 1-2 arguments, but got 0. | 21 |
| error TS2531: Object is possibly '…'. | 17 |
| error TS2704: The operand of a '…' operator cannot be a read-only property. | 15 |
| not yet: reading String | 15 |
| error TS7009: '…' expression, whose target lacks a construct signature, implicitly has an '…' type. | 14 |
| refuses != | 13 |
| refuses throwing a Test262Error | 13 |
| error TS2554: Expected 1 arguments, but got 0. | 11 |
| not yet: a function expression (an arrow function captures this as written) | 9 |
| error TS2551: Property '…' does not exist on type '…'. Did you mean '…'? | 8 |
| error TS2322: Type '…' is not assignable to type '…'. | 6 |
| refuses a method read as a value (toString would lose its object, and this with it) | 6 |
| refuses the comma operator | 6 |
| error TS2554: Expected 2 arguments, but got 1. | 5 |
| error TS7015: Element implicitly has an '…' type because index expression is not of type '…'. | 5 |
| not yet: .toLocaleLowerCase on a string | 5 |
| not yet: endsWith with these arguments | 5 |
| refuses a method read as a value (slice would lose its object, and this with it) | 5 |
| refuses a method read as a value (toLowerCase would lose its object, and this with it) | 5 |
| refuses a method read as a value (trimEnd would lose its object, and this with it) | 5 |
| refuses a method read as a value (trimStart would lose its object, and this with it) | 5 |
| refuses a method read as a value (valueOf would lose its object, and this with it) | 5 |
| refuses arguments | 5 |
| error TS2790: The operand of a '…' operator must be optional. | 4 |
| error TS7034: Variable '…' implicitly has type '…' in some locations where its type cannot be determined. | 4 |
| refuses a method read as a value (at would lose its object, and this with it) | 4 |
| refuses a method read as a value (charAt would lose its object, and this with it) | 4 |
| refuses a method read as a value (charCodeAt would lose its object, and this with it) | 4 |
| refuses a method read as a value (concat would lose its object, and this with it) | 4 |
| refuses a method read as a value (includes would lose its object, and this with it) | 4 |
| refuses a method read as a value (indexOf would lose its object, and this with it) | 4 |
| refuses a method read as a value (lastIndexOf would lose its object, and this with it) | 4 |
| refuses a method read as a value (localeCompare would lose its object, and this with it) | 4 |
| refuses a method read as a value (raw would lose its object, and this with it) | 4 |
| refuses a method read as a value (search would lose its object, and this with it) | 4 |
| refuses a method read as a value (substring would lose its object, and this with it) | 4 |
| refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) | 4 |
| refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) | 4 |
| refuses a method read as a value (toUpperCase would lose its object, and this with it) | 4 |
| refuses the void operator | 4 |
| error TS2351: This expression is not constructable. | 3 |
| error TS2554: Expected 2-3 arguments, but got 1. | 3 |
| error TS2554: Expected 3 arguments, but got 1. | 3 |
| not yet: .toLocaleUpperCase on a string | 3 |
| refuses a method read as a value (codePointAt would lose its object, and this with it) | 3 |
| refuses a method read as a value (endsWith would lose its object, and this with it) | 3 |
| refuses a method read as a value (isWellFormed would lose its object, and this with it) | 3 |
| refuses a method read as a value (match would lose its object, and this with it) | 3 |
| refuses a method read as a value (normalize would lose its object, and this with it) | 3 |
| refuses a method read as a value (split would lose its object, and this with it) | 3 |
| refuses a method read as a value (startsWith would lose its object, and this with it) | 3 |
| refuses a method read as a value (toWellFormed would lose its object, and this with it) | 3 |
| refuses a value of type { raw: string[] | 3 |
| error TS2349: This expression is not callable. | 2 |
| error TS2554: Expected 1-3 arguments, but got 0. | 2 |
| error TS7006: Parameter '…' implicitly has an '…' type. | 2 |
| not yet: .charAt on a string | 2 |
| refuses for...in | 2 |
| error TS1532: There is no capturing group named '…' in this regular expression. | 1 |
| error TS18048: '…' is possibly '…'. | 1 |
| error TS2345: Argument of type '…'0'…' is not assignable to parameter of type '…'. | 1 |
| error TS2362: The left-hand side of an arithmetic operation must be of type '…', '…', '…' or an enum type. | 1 |
| error TS2363: The right-hand side of an arithmetic operation must be of type '…', '…', '…' or an enum type. | 1 |
| error TS2403: Subsequent variable declarations must have the same type.  Variable '…' must be of type '…', but here has type '…'. | 1 |
| error TS2554: Expected 0 arguments, but got 3. | 1 |
| error TS2554: Expected 2 arguments, but got 0. | 1 |
| error TS2683: '…' implicitly has type '…' because it does not have a type annotation. | 1 |
| error TS2741: Property '…' is missing in type '…' but required in type '…'. | 1 |
| error TS7053: Element implicitly has an '…' type because expression of type '…' can'…'{ length: number; }'. | 1 |
| not yet: .concat on a string | 1 |
| not yet: new an Identifier | 1 |
| not yet: reading Function | 1 |
| refuses a method read as a value (fromCharCode would lose its object, and this with it) | 1 |
| refuses a method read as a value (replace would lose its object, and this with it) | 1 |

## Before skip table

| Reason | Count |
| --- | ---: |
| harness include propertyHelper.js | 99 |
| harness include isConstructor.js | 38 |
| feature Symbol | 32 |
| feature Symbol.toPrimitive | 32 |
| feature Symbol.replace | 20 |
| feature Symbol.matchAll | 16 |
| feature Symbol.match | 14 |
| eval | 12 |
| Symbol | 9 |
| feature Symbol.search | 9 |
| feature Symbol.split | 7 |
| features Symbol, Symbol.toPrimitive | 3 |
| noStrict (sloppy mode only; Adamic modules are always strict) | 3 |
| feature Symbol.iterator | 2 |
| feature cross-realm | 2 |
| features BigInt, Symbol.toPrimitive | 2 |
| features Symbol, Symbol.match | 2 |
| features Symbol.search, regexp-v-flag | 2 |
| harness includes compareIterator.js, regExpUtils.js | 2 |
| feature BigInt | 1 |
| feature regexp-duplicate-named-groups | 1 |
| features Reflect, cross-realm | 1 |
| features Symbol, Symbol.match, Symbol.replace | 1 |
| features Symbol, Symbol.split, Symbol.toPrimitive | 1 |
| features Symbol.match, Symbol.replace | 1 |
| features Symbol.match, regexp-v-flag | 1 |
| features Symbol.matchAll, regexp-unicode-property-escapes, regexp-v-flag | 1 |
| features Symbol.replace, regexp-v-flag | 1 |
| features regexp-duplicate-named-groups, regexp-match-indices | 1 |

## After directory table

| Directory | Pass | Disagreement | Refused | Crash | Skipped | Total |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| built-ins/String | 0 | 0 | 80 | 0 | 12 | 92 |
| built-ins/String/fromCharCode | 0 | 0 | 14 | 0 | 3 | 17 |
| built-ins/String/fromCodePoint | 2 | 0 | 4 | 0 | 5 | 11 |
| built-ins/String/prototype | 0 | 0 | 5 | 0 | 2 | 7 |
| built-ins/String/prototype/Symbol.iterator | 0 | 0 | 0 | 0 | 6 | 6 |
| built-ins/String/prototype/at | 4 | 0 | 3 | 0 | 4 | 11 |
| built-ins/String/prototype/charAt | 1 | 0 | 25 | 0 | 4 | 30 |
| built-ins/String/prototype/charCodeAt | 0 | 0 | 21 | 0 | 4 | 25 |
| built-ins/String/prototype/codePointAt | 5 | 0 | 5 | 0 | 6 | 16 |
| built-ins/String/prototype/concat | 0 | 0 | 19 | 0 | 3 | 22 |
| built-ins/String/prototype/constructor | 0 | 0 | 2 | 0 | 0 | 2 |
| built-ins/String/prototype/endsWith | 2 | 0 | 17 | 0 | 8 | 27 |
| built-ins/String/prototype/includes | 5 | 0 | 14 | 0 | 8 | 27 |
| built-ins/String/prototype/indexOf | 0 | 0 | 35 | 0 | 12 | 47 |
| built-ins/String/prototype/isWellFormed | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/String/prototype/lastIndexOf | 1 | 0 | 21 | 0 | 3 | 25 |
| built-ins/String/prototype/localeCompare | 0 | 0 | 10 | 0 | 3 | 13 |
| built-ins/String/prototype/match | 0 | 0 | 36 | 0 | 15 | 51 |
| built-ins/String/prototype/matchAll | 0 | 0 | 1 | 0 | 24 | 25 |
| built-ins/String/prototype/normalize | 0 | 0 | 8 | 0 | 6 | 14 |
| built-ins/String/prototype/padEnd | 3 | 0 | 4 | 0 | 6 | 13 |
| built-ins/String/prototype/padStart | 3 | 0 | 4 | 0 | 6 | 13 |
| built-ins/String/prototype/repeat | 1 | 0 | 9 | 0 | 6 | 16 |
| built-ins/String/prototype/replace | 0 | 0 | 41 | 0 | 14 | 55 |
| built-ins/String/prototype/replaceAll | 0 | 0 | 15 | 0 | 30 | 45 |
| built-ins/String/prototype/search | 0 | 0 | 29 | 0 | 14 | 43 |
| built-ins/String/prototype/slice | 0 | 0 | 34 | 0 | 4 | 38 |
| built-ins/String/prototype/split | 0 | 0 | 106 | 0 | 14 | 120 |
| built-ins/String/prototype/startsWith | 0 | 0 | 13 | 0 | 8 | 21 |
| built-ins/String/prototype/substring | 0 | 0 | 42 | 0 | 4 | 46 |
| built-ins/String/prototype/toLocaleLowerCase | 0 | 0 | 24 | 0 | 4 | 28 |
| built-ins/String/prototype/toLocaleUpperCase | 0 | 0 | 22 | 0 | 4 | 26 |
| built-ins/String/prototype/toLowerCase | 4 | 0 | 21 | 0 | 5 | 30 |
| built-ins/String/prototype/toString | 0 | 0 | 2 | 0 | 5 | 7 |
| built-ins/String/prototype/toUpperCase | 2 | 0 | 20 | 0 | 4 | 26 |
| built-ins/String/prototype/toWellFormed | 0 | 0 | 3 | 0 | 5 | 8 |
| built-ins/String/prototype/trim | 93 | 0 | 34 | 0 | 2 | 129 |
| built-ins/String/prototype/trimEnd | 0 | 0 | 5 | 0 | 18 | 23 |
| built-ins/String/prototype/trimStart | 0 | 0 | 5 | 0 | 18 | 23 |
| built-ins/String/prototype/valueOf | 0 | 0 | 2 | 0 | 5 | 7 |
| built-ins/String/raw | 4 | 0 | 19 | 0 | 7 | 30 |

## After refusal table

| Reason | Count |
| --- | ---: |
| refuses var | 151 |
| error TS2345: Argument of type '…' is not assignable to parameter of type '…'. | 135 |
| error TS2339: Property '…' does not exist on type '…'. | 109 |
| error TS2769: No overload matches this call. | 55 |
| refuses throwing a Test262Error | 27 |
| not yet: a function expression (an arrow function captures this as written) | 25 |
| error TS2554: Expected 1-2 arguments, but got 0. | 21 |
| error TS2531: Object is possibly '…'. | 17 |
| error TS2704: The operand of a '…' operator cannot be a read-only property. | 15 |
| error TS7009: '…' expression, whose target lacks a construct signature, implicitly has an '…' type. | 14 |
| refuses != | 13 |
| error TS2554: Expected 1 arguments, but got 0. | 11 |
| error TS2551: Property '…' does not exist on type '…'. Did you mean '…'? | 8 |
| error TS2322: Type '…' is not assignable to type '…'. | 6 |
| refuses a method read as a value (toString would lose its object, and this with it) | 6 |
| refuses the comma operator | 6 |
| error TS2554: Expected 2 arguments, but got 1. | 5 |
| error TS7015: Element implicitly has an '…' type because index expression is not of type '…'. | 5 |
| not yet: .toLocaleLowerCase on a string | 5 |
| not yet: endsWith with these arguments | 5 |
| not yet: new an Identifier | 5 |
| refuses a method read as a value (trimEnd would lose its object, and this with it) | 5 |
| refuses a method read as a value (trimStart would lose its object, and this with it) | 5 |
| refuses a method read as a value (valueOf would lose its object, and this with it) | 5 |
| refuses arguments | 5 |
| error TS2790: The operand of a '…' operator must be optional. | 4 |
| error TS7034: Variable '…' implicitly has type '…' in some locations where its type cannot be determined. | 4 |
| refuses a method read as a value (charAt would lose its object, and this with it) | 4 |
| refuses a method read as a value (charCodeAt would lose its object, and this with it) | 4 |
| refuses a method read as a value (concat would lose its object, and this with it) | 4 |
| refuses a method read as a value (indexOf would lose its object, and this with it) | 4 |
| refuses a method read as a value (lastIndexOf would lose its object, and this with it) | 4 |
| refuses a method read as a value (localeCompare would lose its object, and this with it) | 4 |
| refuses a method read as a value (search would lose its object, and this with it) | 4 |
| refuses a method read as a value (slice would lose its object, and this with it) | 4 |
| refuses a method read as a value (substring would lose its object, and this with it) | 4 |
| refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) | 4 |
| refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) | 4 |
| refuses a method read as a value (toLowerCase would lose its object, and this with it) | 4 |
| refuses a method read as a value (toUpperCase would lose its object, and this with it) | 4 |
| refuses the void operator | 4 |
| error TS2351: This expression is not constructable. | 3 |
| error TS2554: Expected 2-3 arguments, but got 1. | 3 |
| error TS2554: Expected 3 arguments, but got 1. | 3 |
| not yet: .toLocaleUpperCase on a string | 3 |
| not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) | 3 |
| not yet: reading String as a first-class constructor (its any-typed call signature, construction and static members need an intrinsic value representation) | 3 |
| refuses a method read as a value (match would lose its object, and this with it) | 3 |
| refuses a method read as a value (split would lose its object, and this with it) | 3 |
| error TS2349: This expression is not callable. | 2 |
| error TS2554: Expected 1-3 arguments, but got 0. | 2 |
| error TS7006: Parameter '…' implicitly has an '…' type. | 2 |
| refuses a method read as a value (trim would lose its object, and this with it) | 2 |
| refuses for...in | 2 |
| error TS1532: There is no capturing group named '…' in this regular expression. | 1 |
| error TS18048: '…' is possibly '…'. | 1 |
| error TS2345: Argument of type '…'0'…' is not assignable to parameter of type '…'. | 1 |
| error TS2362: The left-hand side of an arithmetic operation must be of type '…', '…', '…' or an enum type. | 1 |
| error TS2363: The right-hand side of an arithmetic operation must be of type '…', '…', '…' or an enum type. | 1 |
| error TS2403: Subsequent variable declarations must have the same type.  Variable '…' must be of type '…', but here has type '…'. | 1 |
| error TS2554: Expected 0 arguments, but got 3. | 1 |
| error TS2554: Expected 2 arguments, but got 0. | 1 |
| error TS2683: '…' implicitly has type '…' because it does not have a type annotation. | 1 |
| error TS2741: Property '…' is missing in type '…' but required in type '…'. | 1 |
| error TS7053: Element implicitly has an '…' type because expression of type '…' can'…'{ length: number; }'. | 1 |
| not yet: reading Function | 1 |
| refuses a method read as a value (fromCharCode would lose its object, and this with it) | 1 |
| refuses a method read as a value (includes would lose its object, and this with it) | 1 |
| refuses a method read as a value (replace would lose its object, and this with it) | 1 |

## After skip table

| Reason | Count |
| --- | ---: |
| harness include propertyHelper.js | 99 |
| harness include isConstructor.js | 38 |
| feature Symbol | 32 |
| feature Symbol.toPrimitive | 32 |
| feature Symbol.replace | 20 |
| feature Symbol.matchAll | 16 |
| feature Symbol.match | 14 |
| eval | 12 |
| Symbol | 9 |
| feature Symbol.search | 9 |
| feature Symbol.split | 7 |
| features Symbol, Symbol.toPrimitive | 3 |
| noStrict (sloppy mode only; Adamic modules are always strict) | 3 |
| feature Symbol.iterator | 2 |
| feature cross-realm | 2 |
| features BigInt, Symbol.toPrimitive | 2 |
| features Symbol, Symbol.match | 2 |
| features Symbol.search, regexp-v-flag | 2 |
| harness includes compareIterator.js, regExpUtils.js | 2 |
| feature BigInt | 1 |
| feature regexp-duplicate-named-groups | 1 |
| features Reflect, cross-realm | 1 |
| features Symbol, Symbol.match, Symbol.replace | 1 |
| features Symbol, Symbol.split, Symbol.toPrimitive | 1 |
| features Symbol.match, Symbol.replace | 1 |
| features Symbol.match, regexp-v-flag | 1 |
| features Symbol.matchAll, regexp-unicode-property-escapes, regexp-v-flag | 1 |
| features Symbol.replace, regexp-v-flag | 1 |
| features regexp-duplicate-named-groups, regexp-match-indices | 1 |

## Newly passing tests

All of the following passed the comparison with Node:

- `built-ins/String/prototype/at/returns-code-unit.js`
- `built-ins/String/prototype/at/returns-item-relative-index.js`
- `built-ins/String/prototype/at/returns-item.js`
- `built-ins/String/prototype/at/returns-undefined-for-out-of-range-index.js`
- `built-ins/String/prototype/charAt/pos-rounding.js`
- `built-ins/String/prototype/trim/15.5.4.20-1-3.js`
- `built-ins/String/prototype/trim/15.5.4.20-1-4.js`
- `built-ins/String/prototype/trim/15.5.4.20-1-7.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-1.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-10.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-11.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-12.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-13.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-14.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-15.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-16.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-17.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-18.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-19.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-2.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-20.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-21.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-22.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-23.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-24.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-25.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-26.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-27.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-28.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-29.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-3.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-30.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-31.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-32.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-33.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-4.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-5.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-6.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-7.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-8.js`
- `built-ins/String/prototype/trim/15.5.4.20-2-9.js`
- `built-ins/String/raw/return-the-string-value-from-template.js`
- `built-ins/String/raw/special-characters.js`
- `built-ins/String/raw/template-substitutions-are-appended-on-same-index.js`
- `built-ins/String/raw/zero-literal-segments.js`
