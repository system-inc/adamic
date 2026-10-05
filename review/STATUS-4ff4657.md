# Stream R: what is still open on main 4ff4657

Each earlier probe was re-run on the combined integration 7 and 8, four ways (`review/fxspptb/probe.sh`), after
merging main into claude/task-k9saws. That branch's whole gate is green there: gofmt and vet are clean, and
every package passes, the oracle and stage 1's gitignore included.

## Silent: wrong output, exit 0, no sanitizer report

| Probe | Node | Native | From |
|---|---|---|---|
| `round9/c3.a`: a tuple used where an array goes, `[3, 4] as [number, number]` seen as `readonly number[]` | length `2` | a pointer as its length, `94115887495456` | round 9; the collections branch lowers any tuple-typed literal to an object |
| `round9/fs_flush.a`, with stdout redirected to a file | `file 26 true` | `file 0 true`: fileStatus doesn't flush stdout first | round 9; pipes hide it |

## Loud: crash, compiler panic, or a wrong exit

| Probe | What happens | From |
|---|---|---|
| `round10/try_assign.a` | the compiler panics, "a store into the borrowed parameter box": `findAssigned` doesn't look inside `try` | round 10 |
| `round7/early_this.a`, `round7/half_built.a` | `this` read before a field is set: UBSan sanitized, exit 139 at -O2; Node prints undefined | round 7; the refusal is on P's branch, not yet on main |
| `round9/a5.a` | a Set or Map keyed by `string \| undefined` holding undefined: UBSan at map.c, exit 139 | round 9 (older) |
| `round8b/defined_try.a` | a narrowing check inside a `try`: Node catches the TypeError (exit 0); native and the JavaScript backend panic (exit 70) | round 8b |
| `round8b/defined_global.a` | Node's TypeError, against Adamic's own words, in the panic message; both exit 70 | round 8b |
| `integrate1/method_value.a` | a method read as a value panics "compiler bug" natively; Node prints `t1a1` | round 2 (older) |

## Closed on main (checked here)

- `round8b/spread_undefined.a`: a spread of undefined gives `{}`.
- `integrate5/weak/a_filter2.a`: a filtered Weak array.
- `integrate5/weak/a_objview3.a`: now NotYet.
- `integrate5/weak/b_genericclass2.a`: now refused by the cycle finder.
- `integrate6/stack_small.a`: panics under `ulimit -s 1024`, as Node does.
- Every round 2 to 6 finding re-checked in earlier rounds.
