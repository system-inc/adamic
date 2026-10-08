# Language handoff for @system_adamic

Stock TypeScript 6.0.3 accepted each of the first 19 one-line reproducers under the validity runner's exact strict ES2024 options and prelude. Adamic refused each with exit 1. The last reproducer is the independently rejected sumPrecise control. Exact stdout/diagnostics are in [reproducers.json](reproducers.json).

Stored boxed objects require a native internal-slot representation that preserves object identity, structural views, and own-property descriptors. Modeling a Number object as an ordinary field-bearing shape would expose an invented property or misdispatch inherited Object methods. Dynamic ToPrimitive additionally needs receiver binding, method order, and abrupt completion. These are shared value/dispatch requirements; the library slice only consumes immediate intrinsic slots.

| Feature | One-line reproducer | Stock tsc |
|---|---|---|
| array_growth | `const values = new Array<number>(); values[0] = 1; console.log(`${values[0]}`);` | Accepts |
| stored_number | `const value = new Number(1); console.log(value.toString());` | Accepts |
| boxed_string | `console.log(`${Number(new String("10"))}`);` | Accepts |
| boxed_object | `console.log(`${Number(new Object(10))}`);` | Accepts |
| dynamic_toprimitive | `console.log(`${Number({valueOf: () => 1})}`);` | Accepts |
| prototype_chain | `console.log(`${Number.prototype.isPrototypeOf({})}`);` | Accepts |
| prototype_mutation | `Number.prototype.toString = () => "changed"; console.log(Number.prototype.toString());` | Accepts |
| prototype_reflection | `console.log(`${Object.getPrototypeOf(new Number(1))}`);` | Accepts |
| detached_method | `const value = {}; const print = value.toString; console.log(print());` | Accepts |
| for_in_intrinsic | `for (const key in Number) { console.log(key); }` | Accepts |
| var | `var value = 1; console.log(`${value}`);` | Accepts |
| loose_equality | `console.log(`${1 == 1}`);` | Accepts |
| void | `console.log(`${void 0}`);` | Accepts |
| mixed_add | `console.log("value:" + 1);` | Accepts |
| catch_toFixed | `try { console.log((1).toFixed(-1)); } catch { console.log("caught"); }` | Accepts |
| catch_toPrecision | `try { console.log((1).toPrecision(0)); } catch { console.log("caught"); }` | Accepts |
| catch_toExponential | `try { console.log((1).toExponential(101)); } catch { console.log("caught"); }` | Accepts |
| catch_toString | `try { console.log((1).toString(1)); } catch { console.log("caught"); }` | Accepts |
| random | `console.log(`${Math.random()}`);` | Accepts |
| sumPrecise | `console.log(`${Math.sumPrecise([1, 2])}`);` | TS2339 |

## Primary remaining counts

- Number: stored boxing/ToPrimitive construction 6, var 6, prototype chain 3, catchable formatter failures 5 (toString 2; the other three 1 each), intrinsic for...in 2, Object prototype reflection/mutation 2, dynamic ToPrimitive 1, mixed addition 1, loose equality 1, void 1. Total 28.
- Math: array construction/growth 29 and var 2. Total 31 language refusals. Math.random adds 1 explicit deterministic-library refusal.
- Math.sumPrecise adds 4 reported refusals despite real tsc rejecting the exact adapted sources. Adamic reports TS2550; stock tsc reports TS2339 under ES2024. The prescribed runner requires matching codes, so these stay refused. This unit does not alter the runner or count them as passes.

The 29 Math construction cases use `new Array()` and indexed growth. Their exact paths are recorded in [before-results.jsonl](before-results.jsonl). Fixing their first refusal may expose additional array, coercion, or type representation requirements.

Math.random cannot meet the independent deterministic byte-for-byte oracle contract for arbitrary printing callers. It stays refused with its existing reason. This is a retained library contract, not a missing arithmetic approximation.
