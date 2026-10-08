# Caught values

Ruling by @system_adamic, October 7, 21:41:

1. A caught value keeps what was thrown. `instanceof Error` answers as Node does.
2. A caught member read tells the true type. In the message cases below, `e.message` is `string | undefined`: the actual field when present, otherwise `undefined`. No check is inserted at the read. Values in arbitrary object fields retain their actual tags too.
3. Reading a member of thrown `undefined` or `null` raises a catchable Adamic TypeError, as Node does, and unwinds through cleanup.
4. A definite string parameter, field, or return checks the narrowing from `string | undefined`. A missing message stops with exit 70 and `adamic/catch-type: caught value does not match its typed use`. The check also rejects a field holding a different primitive type.
5. Tests compare both backends byte for byte with Node for thrown `'x'`, `{}`, `{ message: 'm' }`, `new Error('e')`, and `undefined`. They additionally cover null, numbers, scalar fields and returned records. Treating a read as a definite string must be caught by these tests. The project `.ts` parameter probe stops loudly; the corresponding `.a` program is refused.

The `.a` program:

```ts
function accepts(message: string): void { console.log(message); }
try { throw 'x'; } catch (e) { accepts(e.message); }
```

Its pinned diagnostic is `main.a:2:40: error TS18046: 'e' is of type 'unknown'.` Adamic `.a` requires a proof before reading an unknown catch binding. In a `.ts` project declaring `useUnknownInCatchVariables: false`, TypeScript retains its declared catch type for checking; Adamic stores it as an unknown tagged value and inserts the definite-use check. The same source on Node prints `undefined` and continues.

`internal/lower/catch_values.go:caughtMemberRead` is the member-read policy boundary. Catch-derived record fields carry the generic value; their declared Error-like view cannot turn a thrown primitive into an Error. Error's runtime object fields remain `name` and `message`; the generic thrown carrier and the immortal null tag do not change that layout.

Dynamic properties that could invoke a source accessor are currently refused with a named reason until their getter targets can be represented soundly. Optional access on an unclassified caught value is likewise refused; explicitly test for undefined first.

Unclassified dynamic reads are currently admitted for `message` and `code`. Other property names require narrowing to a declared receiver type and otherwise are NotYet: prototype methods and primitive metadata cannot be replaced by an own-field lookup. This includes a string's `length`, Error's `stack`, and an unclassified function's `name`.

A potential source method named `message` or `code` also needs a narrowed receiver: a native class method lives in its dispatch table, not an ordinary field slot. The dynamic read refuses that case instead of silently returning undefined.

Potential static class fields of those names likewise require the declared constructor receiver, because their constructor layout is separate from instance own fields.
