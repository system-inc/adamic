# Stream R, round eight (second half): Perceus meets exceptions, on cloud/integrate-7-flow (3ab8f94)

I ran everything on Linux x86-64 with Go 1.27.0, clang 18.1.3 and Node v24.21.0. `review/fxspptb/probe.sh` runs
each probe four ways:

1. Node on the source.
2. Native under ASan, UBSan and LeakSanitizer.
3. Native as `adamic build` makes it.
4. The JavaScript backend.

## Findings, worst first

1. **A2's claim that the `ir.Defined` check reuse looks through can never fire is false for a global.**
   `defined_global.a`: a global narrowed by `if (tree !== undefined)`, put back to undefined by a call
   (`drop()`), then `tree = insert(tree, 1)`, which moves it into a consumed parameter.

   The check fires before the move (the generated C tests `adamic_global_0_tree == NULL`, then moves). So it
   isn't a memory bug: no NULL is moved, and native agrees with the JavaScript backend. It's the only thing
   standing between that move and a NULL dereference, so it must stay.

   For an owned, uncaptured local or a parameter, the claim does hold by construction: no call can write those.
   `defined.a` narrows each and spreads it after a call, and agrees on all four runs.

   The messages differ from Node's:

   | Run | stderr |
   |---|---|
   | Node | `adamic: panic: TypeError: Cannot read properties of undefined (reading 'value')`, thrown inside `insert` |
   | Native | `adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back` |

   Both exit 70.
2. **A narrowing check inside a `try` panics natively, where Node catches the TypeError.** `defined_try.a` is
   the same program inside `try`/`catch`:

   | Run | Output | Exit |
   |---|---|---|
   | Node | `caught Cannot read properties of undefined (reading 'value')`, `after` | 0 |
   | Native, sanitized and -O2 | panic | 70 |
   | JavaScript backend | panic | 70 |

   The exceptions design lists the panics that Node could catch inside a `try` (running out of stack, the
   string limit, the temporal dead zone). Narrowing checks aren't among them, and nothing refuses a `try` that
   can reach one. Two choices:
   - make the check a catchable TypeError through the pending-exception word, as the design plans for library
     failures;
   - or say NotYet for a `try` that can reach one.
3. **A spread of a value that may be undefined dereferences NULL, older than this branch.**
   `spread_undefined.a`: `const spread = { ...maybe }` where `maybe` is `Tree | undefined` and undefined.

   | Run | Output |
   |---|---|
   | Node | `0 none` (JavaScript's `{...undefined}` is `{}`) |
   | My branch's base | UBSan in `adamic_object_copy` (object.c:15) |
   | integrate-7-flow | UBSan earlier, at reuse's `maybe->heap.references`, since the spread is now a reuse source |
   | -O2 | segfault |

   Lowering accepts the spread of a `T | undefined` without NotYet or a NULL path.

## What held

`family.a` agrees on all four runs, leak-clean:

- A throw out of a callee that reused its parameter in place: the caller's catch reads its variable unchanged.
- A finally reading a variable that a throwing statement would have overwritten.
- A spread interrupted mid-field by a throw, the source read in the catch.
- A global moved in a statement that can throw, read in the catch.
- A loop whose throwing rounds keep the old value and whose good rounds reuse.
- A rethrow from a catch that reads the moved value.

The rules A2 added hold in each: no overwrite in a statement that can throw, no global moved there, temporaries
kept until every argument is evaluated, and exception edges in liveness.

## Not covered

- Regions meeting exceptions: they're still on different branches.
- macOS and arm64.
