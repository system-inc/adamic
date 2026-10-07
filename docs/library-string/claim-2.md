# String second slice claim

Branch: `codex/library-string-2`, main `50045bd`. Linux, test262 `c8c798898646638cd0c24879f8e0374e847e7d74`, validity runner `af12899`, stock tsc 6.0.3, adaptation on.

| Pass | Disagreement | Refused | Not TypeScript | Crashed | Skipped |
|---:|---:|---:|---:|---:|---:|
| 205 | 1 | 267 | 433 | 1 | 316 |

First: reproduce and refuse catchable fromCodePoint RangeError; reproduce and fix S9.8_A4_T1 clang failure.

Library work is ordered by current first refusal count. Boxed String construction is first (42 of the 49 `new` refusals; the other seven are Array/Object/Boolean/Number), but indexed exotic objects, inherited methods, reflection and internal-slot identity require language representation support. Keep these refused rather than represent a box as a primitive. Then ToPrimitive (22), null receiver exceptions (14), positioned affixes (12), locale methods (10), raw array-like conversion (4), well-formed UTF-16 (2). Implement sound subsets and document dependencies; Node decides every new pass.

## Complete first-refusal classification

The partition is 109 library, 156 language, and two checker/oracle diagnostic mismatches. These counts partition the 267 current refusals, not every obstruction inside each test. A later refusal can become visible after a library feature lands. Detached functions, dynamic constructors, any, loose operators, reflection and library intrinsic values are language dependencies for @system_adamic.

| Count | Owner | First refusal | One-line reproducer for language owner |
|---:|---|---|---|
| 49 | String library (42) / @system_adamic (7) | not yet: new an Identifier | `new Array(); new Object(); new Boolean(true); new Number(1);` |
| 22 | String library | not yet: String conversion of an object, array, map or function (ToPrimitive is not lowered) |  |
| 16 | @system_adamic | not yet: a BinaryExpression with a string and a number | `const x = "x" + 1;` |
| 14 | String library | not yet: String prototype call on null or undefined (its TypeError is not catchable natively yet) |  |
| 13 | @system_adamic | refuses != | `const x = "a" != "b";` |
| 8 | @system_adamic | not yet: a value of type any | `const x: any = 1; String(x);` |
| 8 | String library | not yet: endsWith with these arguments |  |
| 6 | @system_adamic | refuses a method read as a value (toString would lose its object, and this with it) | `const f = String.prototype.toString;` |
| 6 | String library | refuses inherited library member toLocaleLowerCase read as an own field |  |
| 6 | @system_adamic | refuses the comma operator | `const x = (1, 2);` |
| 5 | @system_adamic | refuses a method read as a value (trimEnd would lose its object, and this with it) | `const f = String.prototype.trimEnd;` |
| 5 | @system_adamic | refuses a method read as a value (trimStart would lose its object, and this with it) | `const f = String.prototype.trimStart;` |
| 5 | @system_adamic | refuses a method read as a value (valueOf would lose its object, and this with it) | `const f = String.prototype.valueOf;` |
| 5 | @system_adamic | refuses arguments | `function f(): void { String(arguments); }` |
| 4 | String library | not yet: startsWith with these arguments |  |
| 4 | @system_adamic | refuses a method read as a value (charAt would lose its object, and this with it) | `const f = String.prototype.charAt;` |
| 4 | @system_adamic | refuses a method read as a value (charCodeAt would lose its object, and this with it) | `const f = String.prototype.charCodeAt;` |
| 4 | @system_adamic | refuses a method read as a value (concat would lose its object, and this with it) | `const f = String.prototype.concat;` |
| 4 | @system_adamic | refuses a method read as a value (indexOf would lose its object, and this with it) | `const f = String.prototype.indexOf;` |
| 4 | @system_adamic | refuses a method read as a value (lastIndexOf would lose its object, and this with it) | `const f = String.prototype.lastIndexOf;` |
| 4 | @system_adamic | refuses a method read as a value (localeCompare would lose its object, and this with it) | `const f = String.prototype.localeCompare;` |
| 4 | @system_adamic | refuses a method read as a value (search would lose its object, and this with it) | `const f = String.prototype.search;` |
| 4 | @system_adamic | refuses a method read as a value (slice would lose its object, and this with it) | `const f = String.prototype.slice;` |
| 4 | @system_adamic | refuses a method read as a value (substring would lose its object, and this with it) | `const f = String.prototype.substring;` |
| 4 | @system_adamic | refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) | `const f = String.prototype.toLocaleLowerCase;` |
| 4 | @system_adamic | refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) | `const f = String.prototype.toLocaleUpperCase;` |
| 4 | @system_adamic | refuses a method read as a value (toLowerCase would lose its object, and this with it) | `const f = String.prototype.toLowerCase;` |
| 4 | @system_adamic | refuses a method read as a value (toUpperCase would lose its object, and this with it) | `const f = String.prototype.toUpperCase;` |
| 4 | String library | refuses inherited library member toLocaleUpperCase read as an own field |  |
| 4 | @system_adamic | refuses the void operator | `String(void 0);` |
| 3 | String library | not yet: String.raw without a present array of strings in raw |  |
| 3 | @system_adamic | refuses a method read as a value (match would lose its object, and this with it) | `const f = String.prototype.match;` |
| 3 | @system_adamic | refuses a method read as a value (split would lose its object, and this with it) | `const f = String.prototype.split;` |
| 2 | String library | not yet: a try around repeat, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |  |
| 2 | @system_adamic | not yet: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow) | `String.hasOwnProperty("prototype");` |
| 2 | @system_adamic | refuses a method read as a value (trim would lose its object, and this with it) | `const f = String.prototype.trim;` |
| 1 | @system_adamic | error TS2345: Argument of type '…' is not assignable to parameter of type '…'. (tsc: TS1125,TS1532) | `"x".split(/\x/); "x".split(/\k<x>/);` |
| 1 | @system_adamic | error TS2741: Property '…' is missing in type '…' but required in type '…'. (tsc: TS2345) | `const obj = {}; Object.defineProperty(obj, 'raw', {get() { throw new Error(); }}); String.raw(obj);` |
| 1 | @system_adamic | not yet: RegExp with a nonconstant pattern | `const p = "x"; new RegExp(p);` |
| 1 | @system_adamic | not yet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | `const f = String;` |
| 1 | String library | not yet: String.raw with raw elements that are not strings |  |
| 1 | @system_adamic | not yet: a function returning undefined | `const f = () => undefined;` |
| 1 | @system_adamic | not yet: a void call used as a value | `function f(): void {} String(f());` |
| 1 | String library | not yet: localeCompare (Node uses locale collation, not ordinal UTF-16 order) |  |
| 1 | @system_adamic | not yet: reading Boolean | `const f = Boolean;` |
| 1 | @system_adamic | not yet: reading x | `String(x); var x: undefined;` |
| 1 | @system_adamic | refuses Object.defineProperty | `Object.defineProperty({}, "x", {value: 1});` |
| 1 | @system_adamic | refuses a method read as a value (fromCharCode would lose its object, and this with it) | `const f = String.fromCharCode;` |
| 1 | @system_adamic | refuses a method read as a value (includes would lose its object, and this with it) | `const f = String.prototype.includes;` |
| 1 | @system_adamic | refuses a method read as a value (replace would lose its object, and this with it) | `const f = String.prototype.replace;` |
| 1 | @system_adamic | refuses inherited library member constructor read as an own field | `String.prototype.constructor;` |
| 1 | String library | refuses inherited library member isWellFormed read as an own field |  |
| 1 | @system_adamic | refuses inherited library member slice read as an own field | `String.prototype.slice;` |
| 1 | String library | refuses inherited library member toWellFormed read as an own field |  |
| 1 | @system_adamic | refuses isPrototypeOf | `String.prototype.isPrototypeOf({});` |
| 1 | @system_adamic | refuses var | `var x = "x";` |

No language features will be built in this unit. RangeError-producing runtime calls remain refused around a try; the existing cleanup-aware object throw can support statically known nullish receiver TypeErrors. Locale-sensitive operations remain refused until Node-compatible ICU is available.

Mutants will change real behavior at a boundary and must complete with valid C and no sanitizer failure, with only the comparison against Node rejecting them.
