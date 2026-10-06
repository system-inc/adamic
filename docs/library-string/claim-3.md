# String third slice claim

Branch codex/library-string-3 starts at e911e46, the tip of codex/library-string-2. The previous complete mandated-runner table at that exact tip is below; a fresh Linux survey is running before any implementation. Runner: origin/codex/test262-ts-validity af12899, built in /workspace/adamic-validity-string-3. Test262 c8c798898646638cd0c24879f8e0374e847e7d74, TypeScript 6.0.3.

| Pass | Disagreement | Refused | Crashed | Skipped | Not TypeScript |
|---:|---:|---:|---:|---:|---:|
| 257 | 0 | 217 | 0 | 316 | 433 |

Priority: boxed String construction (42), locale methods (11), proven object conversion views (3), catchable repeat/fromCodePoint RangeErrors (3). Boxes will use an object with a private String slot, preserving object identity, length and UTF-16 indexing; reflective/exotic operations without a sound representation stay refused. Locale support needs Node-compatible ICU behavior; no ordinal or host-version approximation. Hidden conversion members require a proof of the complete view, not a guess.

Read codex/regex-protocol ffb3ae0: its intrinsic match/matchAll/replace/search/split protocols, replacement callbacks, exception constructors and d indices are dependencies. Merge that implementation after this claim; do not duplicate protocol/runtime algorithms. Extend String entry points for match, matchAll, replace, replaceAll, search and split using its RegExpCall IR. Custom hooks, exec, species and getters stay language dependencies unless already proven by the protocol implementation.

The 217 first obstructions partition into 59 library, 156 language and two checker/oracle diagnostic mismatches. A later obstruction can be exposed after its first one is removed. Below is the complete reason grouping; language reproducers go to @system_adamic. No language feature will be built.

| Count | Owner | Observed first refusal | One-line language reproducer |
|---:|---|---|---|
| 42 | String library | not yet: new String needs indexed exotic properties and String internal slots (a primitive or plain object is not a String box) |  |
| 16 | @system_adamic | not yet: a BinaryExpression with a string and a number | `const x = "x" + 1;` |
| 13 | @system_adamic | refuses != | `const x = "a" != "b";` |
| 8 | @system_adamic | not yet: a value of type any | `const x: any = 1; String(x);` |
| 7 | @system_adamic | not yet: new an Identifier | `new Array(); new Object(); new Boolean(true); new Number(1);` |
| 6 | @system_adamic | refuses a method read as a value (toString would lose its object, and this with it) | `const f = String.prototype.toString;` |
| 6 | String library | refuses inherited library member toLocaleLowerCase read as an own field |  |
| 6 | @system_adamic | refuses the comma operator | `const x = (1, 2);` |
| 5 | @system_adamic | refuses a method read as a value (trimEnd would lose its object, and this with it) | `const f = String.prototype.trimEnd;` |
| 5 | @system_adamic | refuses a method read as a value (trimStart would lose its object, and this with it) | `const f = String.prototype.trimStart;` |
| 5 | @system_adamic | refuses a method read as a value (valueOf would lose its object, and this with it) | `const f = String.prototype.valueOf;` |
| 5 | @system_adamic | refuses arguments | `function f(): void { String(arguments); }` |
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
| 3 | String library | not yet: String ToPrimitive needs a conversion member hidden by the object view |  |
| 3 | @system_adamic | refuses a method read as a value (match would lose its object, and this with it) | `const f = String.prototype.match;` |
| 3 | @system_adamic | refuses a method read as a value (split would lose its object, and this with it) | `const f = String.prototype.split;` |
| 2 | String library | not yet: a try around repeat, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |  |
| 2 | @system_adamic | not yet: hasOwnProperty on a function (its prototype property depends on whether it was declared or made as an arrow) | `String.hasOwnProperty("prototype");` |
| 2 | @system_adamic | refuses a method read as a value (trim would lose its object, and this with it) | `const f = String.prototype.trim;` |
| 1 | @system_adamic | error TS2345: Argument of type '…' is not assignable to parameter of type '…'. (tsc: TS1125,TS1532) | `"x".split(/\x/); "x".split(/\k<x>/);` |
| 1 | @system_adamic | error TS2741: Property '…' is missing in type '…' but required in type '…'. (tsc: TS2345) | `const obj = {}; Object.defineProperty(obj, 'raw', {get() { throw new Error(); }}); String.raw(obj);` |
| 1 | @system_adamic | not yet: RegExp with a nonconstant pattern | `const p = "x"; new RegExp(p);` |
| 1 | @system_adamic | not yet: String as a value outside equality or typeof (overloaded calls and static properties need their own representation) | `const f = String;` |
| 1 | @system_adamic | not yet: a function returning undefined | `const f = () => undefined;` |
| 1 | String library | not yet: a try around String.fromCodePoint, whose failure is a panic natively but a throw a catch can take on Node (docs/memory.md) |  |
| 1 | @system_adamic | not yet: a void call used as a value | `function f(): void {} String(f());` |
| 1 | String library | not yet: localeCompare (Node uses locale collation, not ordinal UTF-16 order) |  |
| 1 | @system_adamic | not yet: reading Boolean | `const f = Boolean;` |
| 1 | @system_adamic | not yet: reading x | `String(x); var x: undefined;` |
| 1 | @system_adamic | refuses Object.defineProperty | `Object.defineProperty({}, "x", {value: 1});` |
| 1 | @system_adamic | refuses a method read as a value (fromCharCode would lose its object, and this with it) | `const f = String.fromCharCode;` |
| 1 | @system_adamic | refuses a method read as a value (includes would lose its object, and this with it) | `const f = String.prototype.includes;` |
| 1 | @system_adamic | refuses a method read as a value (replace would lose its object, and this with it) | `const f = String.prototype.replace;` |
| 1 | @system_adamic | refuses inherited library member constructor read as an own field | `String.prototype.constructor;` |
| 1 | @system_adamic | refuses inherited library member slice read as an own field | `String.prototype.slice;` |
| 1 | @system_adamic | refuses isPrototypeOf | `String.prototype.isPrototypeOf({});` |
| 1 | @system_adamic | refuses var | `var x = "x";` |

Each built method family gets a mutant that completes with valid C, clean sanitizers and leaks, and differs only on the independent source-Node comparison. New Adamic fixtures are .a. Before/after tables and a five-line summary will be recorded; no PR.
