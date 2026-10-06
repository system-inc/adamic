# Newly exposed language blockers

The initial claim provides the complete original language handoff. These additional
families appeared after the Array library blocker was removed. Counts below are
first blockers only; complete sources and exact compiler messages are in
second-pass-final-refusals.json. They remain refused rather than approximated.

| Count | Feature | One-line reproducer for @system_adamic |
|---:|---|---|
| 12 | Sparse array presence | `const a = new Array<number>(3);` |
| 12 | Property descriptors | `Object.defineProperty([1], "0", {value: 2});` |
| 10 | Indexed extension | `const a = new Array(1, 2); a[2] = 3;` |
| 10 | Empty/heterogeneous slots | `const a = new Array();` |
| 7 | Generic array element slots | `function f<T>(a: T[]): T[] { return a.toSorted(); }` |
| 2 | Constructor aliases in instanceof | `const C = Array; const a = new Array(1, 2); const b = a instanceof C;` |
| 2 | In operator | `const a = new Array(1, 2); const b = "0" in a;` |
| 2 | Var declarations | `var a = new Array(1, 2);` |
| 1 | Object coercion in default sort | `[ {toString: (): string => "a"} ].sort();` |

Two newly exposed Object.freeze blockers are other-library work:
`const a = new Array(1, 2); Object.freeze(a);`.

The original 437-input ownership ledger is preserved. Of the final 386 first
blockers, 341 are language, 29 other-library and 16 Array-library. Array-library
blockers are 14 first-class intrinsic reads/aliases, one generic map receiver,
and one function receiver to shift. No handoff message was sent to the maintainer.
