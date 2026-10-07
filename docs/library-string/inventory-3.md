# String third slice remaining inventory

Fresh mandated-runner results at implementation f5f66da: 187 refused rows. These are first obstructions; removing one can expose another. Language features below are handed to @system_adamic and were not implemented.

| Group | Refused rows |
|---|---:|
| language | 168 |
| String library | 16 |
| diagnostic mismatch | 2 |
| stock TypeScript rejection | 1 |

| Count | Owner | Observed refusal | One-line reproducer or limit |
|---:|---|---|---|
| 26 | language | not yet: a BinaryExpression with a string and a number | `const x = "x" + 1;` |
| 13 | language | refuses != | `const x = "a" != "b";` |
| 8 | language | not yet: a value of type any | `const x: any = 1; String(x);` |
| 7 | language | not yet: new an Identifier | `new Array(); new Object(); new Boolean(true); new Number(1);` |
| 6 | String library | refuses inherited library member toLocaleLowerCase read as an own field | See the explicit library limit below. |
| 6 | language | refuses the comma operator | `const x = (1, 2);` |
| 5 | language | refuses a method read as a value (toString would lose its object, and this with it) | `const f = String.prototype.toString;` |
| 5 | language | refuses a method read as a value (trimEnd would lose its object, and this with it) | `const f = String.prototype.trimEnd;` |
| 5 | language | refuses a method read as a value (trimStart would lose its object, and this with it) | `const f = String.prototype.trimStart;` |
| 5 | language | refuses arguments | `function f(): void { String(arguments); }` |
| 4 | String library | not yet: String ToPrimitive needs a conversion member hidden by the object view | See the explicit library limit below. |
| 4 | language | refuses a method read as a value (charAt would lose its object, and this with it) | `const f = String.prototype.charAt;` |
| 4 | language | refuses a method read as a value (charCodeAt would lose its object, and this with it) | `const f = String.prototype.charCodeAt;` |
| 4 | language | refuses a method read as a value (concat would lose its object, and this with it) | `const f = String.prototype.concat;` |
| 4 | language | refuses a method read as a value (indexOf would lose its object, and this with it) | `const f = String.prototype.indexOf;` |
| 4 | language | refuses a method read as a value (lastIndexOf would lose its object, and this with it) | `const f = String.prototype.lastIndexOf;` |
| 4 | language | refuses a method read as a value (localeCompare would lose its object, and this with it) | `const f = String.prototype.localeCompare;` |
| 4 | language | refuses a method read as a value (search would lose its object, and this with it) | `const f = String.prototype.search;` |
| 4 | language | refuses a method read as a value (slice would lose its object, and this with it) | `const f = String.prototype.slice;` |
| 4 | language | refuses a method read as a value (substring would lose its object, and this with it) | `const f = String.prototype.substring;` |
| 4 | language | refuses a method read as a value (toLocaleLowerCase would lose its object, and this with it) | `const f = String.prototype.toLocaleLowerCase;` |
| 4 | language | refuses a method read as a value (toLocaleUpperCase would lose its object, and this with it) | `const f = String.prototype.toLocaleUpperCase;` |
| 4 | language | refuses a method read as a value (toLowerCase would lose its object, and this with it) | `const f = String.prototype.toLowerCase;` |
| 4 | language | refuses a method read as a value (toUpperCase would lose its object, and this with it) | `const f = String.prototype.toUpperCase;` |
| 4 | language | refuses a method read as a value (valueOf would lose its object, and this with it) | `const f = String.prototype.valueOf;` |
| 4 | String library | refuses inherited library member toLocaleUpperCase read as an own field | See the explicit library limit below. |
| 4 | language | refuses the void operator | `String(void 0);` |
| 3 | language | refuses a method read as a value (match would lose its object, and this with it) | `const f = String.prototype.match;` |
| 3 | language | refuses a method read as a value (split would lose its object, and this with it) | `const f = String.prototype.split;` |
| 2 | language | not yet: a BinaryExpression with a string and a value | `const box = new String('x'); const x = 'x' + box;` |
| 2 | language | not yet: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow) | `String.hasOwnProperty("prototype");` |
| 2 | language | not yet: overwriting a String box prototype member | `const box = new String('x'); box.valueOf = () => 'y';` |
| 2 | language | not yet: reading x | `String(x); var x: undefined;` |
| 2 | language | refuses a method read as a value (trim would lose its object, and this with it) | `const f = String.prototype.trim;` |
| 1 | diagnostic mismatch | error TS2345: Argument of type '…' is not assignable to parameter of type '…'. (tsc: TS1125,TS1532) | `"x".split(/\x/); "x".split(/\k<x>/);` |
| 1 | diagnostic mismatch | error TS2741: Property '…' is missing in type '…' but required in type '…'. (tsc: TS2345) | `const obj = {}; Object.defineProperty(obj, 'raw', {get() { throw new Error(); }}); String.raw(obj);` |
| 1 | stock TypeScript rejection | not yet: RegExp with a nonconstant pattern: the native runtime has no ECMAScript pattern compiler | `function f(p: string): RegExp { return new RegExp(p); }` |
| 1 | language | not yet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | `const f = String;` |
| 1 | language | not yet: a BinaryExpression with a string and a boolean | `const x = 'x' + true;` |
| 1 | String library | not yet: a String box viewed through erased, mutable or prototype fields | `const box = new String('x'); const view: {readonly length: number} = box;` |
| 1 | language | not yet: a function returning undefined | `const f = () => undefined;` |
| 1 | language | not yet: a void call used as a value | `function f(): void {} String(f());` |
| 1 | String library | not yet: localeCompare (Node uses locale collation, not ordinal UTF-16 order) | See the explicit library limit below. |
| 1 | language | not yet: reading Boolean | `const f = Boolean;` |
| 1 | language | refuses Object.defineProperty | `Object.defineProperty({}, "x", {value: 1});` |
| 1 | language | refuses a method read as a value (fromCharCode would lose its object, and this with it) | `const f = String.fromCharCode;` |
| 1 | language | refuses a method read as a value (includes would lose its object, and this with it) | `const f = String.prototype.includes;` |
| 1 | language | refuses a method read as a value (replace would lose its object, and this with it) | `const f = String.prototype.replace;` |
| 1 | language | refuses inherited library member constructor read as an own field | `String.prototype.constructor;` |
| 1 | language | refuses isPrototypeOf | `String.prototype.isPrototypeOf({});` |
| 1 | language | refuses var | `var x = "x";` |

Remaining String library limits, largest first:

- Locale casing/collation: 11 rows. Exact Node ICU behavior is absent; do not approximate with ordinary casing or ordinal comparison.
- Hidden object conversion members: 4 rows. The view cannot establish an own callable conversion member or the default prototype. A complete stable view proof is required.
- Erased box views: 1 row. The admitted private-slot representation cannot expose arbitrary exotic/prototype fields through a wider structural view. Explicit guards prevent that representation escaping through nested containers. Prototype mutation itself is a language dependency (2 rows), listed above.

The dynamic-pattern row is replace/S15.5.4.11_A1_T17.js. Stock tsc rejects the adapted source with TS2769 using both baseline and current preludes. The imported checker declarations let Adamic reach lowering, so the runner stops at its nonconstant-pattern refusal without consulting stock tsc. It moves from not-typescript to refused but is not work owed. The two diagnostic mismatches are also stock TypeScript rejections with different diagnostic codes. All three are excluded from the feature work list; the measured table is left unchanged. Dynamic patterns in stock-valid programs would need the RegExp owner's ECMAScript pattern compiler and are not duplicated here.

All 31 newly passing paths were independently rerun against the frozen baseline and each was refused there. Stock TypeScript 6.0.3 also accepts every exact adapted new-pass source (new-pass-stock-3.jsonl). Every baseline pass is retained. New passes agree with Node; counts, sanitizers and leak checks pass. The 31-path baseline report is new-pass-baseline-3.json.
