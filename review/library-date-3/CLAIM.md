# Date 3 claim

Branch codex/library-date-3 starts at codex/library-date-refusals 4ec07a3.
Claim Date-specific fixtures and lowering/runtime hooks, runner integration and Date
adaptation, plus this review directory. No language extension and no new use of
`dateNullableString`. New Adamic programs use .a.

## Measured scope

Runner origin/codex/test262-ts-validity af12899, compiler 4ec07a3, --adapt,
test262 3fd3eab12309bd7732f4b5ddeaae19c5d95ad9dd, stock tsc 6.0.3, TZ=UTC:
111 pass, 0 disagreements, 90 refused, 123 not-typescript, 0 crashes, 270 skipped,
594 total. This differs from the supplied 201-refusal inventory. The previous
Date metadata implementation already covers the 42 prototype own-property tests
and UTC name/length observations. This runner lacks the Date adaptations already
present on our compiler branch.

## Largest first

| Measured group | Count | Classification and action |
|---|---:|---|
| Untyped Date result locals represented as any | 43 | Harness/source adaptation: preserve the existing proven Date-local annotation adaptation while integrating the TS-validity runner. Do not implement runtime any. |
| Detached toString | 13 | Language: detached inherited method values; list repro, do not build. |
| Other detached getter/ISO methods | 19 | Same language feature; list repro, do not build. |
| Date.now and no-argument construction | 5 | Intentional nondeterminism refusal; leave with reason. |
| Date.prototype read by a call | 3 | Observable prototype objects plus catchable TypeError; language, do not materialize an invalid Date instance. |
| Catch around toISOString validation | 3 | Catchable native validation failures; language, do not build. |
| Borrowed ISO method on a non-Date | 2 | Catchable native TypeError and absent internal slots; language, do not build. |
| new String | 1 | Boxed primitive constructor; language, do not build. |
| Function.prototype.isPrototypeOf(Date) | 1 | Observable intrinsic prototype chain and constructor values; language, do not build. |

No unimplemented library arithmetic, parser or Date method family is exposed by
these 90 first refusals. Integrate only the runner commits, retaining our existing
Date adaptations, then audit every resulting refusal for later diagnostics.
The 123 matching stock-tsc rejections are not compiler work; adapted harness
restrictions can move out of this outcome when existing safe adaptations apply.

## One-line language reproducers for @system_adamic

Each line is stored verbatim in gaps/*.a; compile with TZ=UTC. For stock tsc,
copy the .a source to a scratch .ts path and use the runner's strict options and
internal/load/prelude.d.ts. These are language boundaries, not implementation targets.

| Feature | Reproducer |
|---|---|
| General evolving any local | `let result; result = new Date(0).getTime(); console.log(String(result));` |
| Detached inherited method | `const f = new Date(0).toString; console.log(f.call(new Date(0)));` |
| Observable Date.prototype | `console.log(typeof Date.prototype);` |
| Native library validation caught by user code | `try { new Date(NaN).toISOString(); } catch (e) { console.log(String(e instanceof Error)); }` |
| Missing Date slot throws catchable TypeError | `try { Date.prototype.toISOString.call({}); } catch (e) { console.log(String(e instanceof Error)); }` |
| Boxed String constructor | `const s = new String("");` |
| Observable intrinsic prototype ancestry | `console.log(String(Function.prototype.isPrototypeOf(Date)));` |

The any language repro is not an excuse to add dynamic values: its test262
instances already admit an erasable, checked annotation in the existing adapter.
Clock repros are `Date.now()` and `new Date()`; neither is a language gap.
