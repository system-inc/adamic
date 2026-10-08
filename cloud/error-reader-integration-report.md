Fixed null Error causes and catchable null prototype receivers on the feature branch; no main or area branch was pushed or merged into.
Implementation commits: `2a8a59b`; merge `ca8c83c` brings in requested main `e011f8f`; reader import/precision fix `e21c5ae`; ownership test/mutant integration `2d2a598`; final report tip is recorded below.
Validation: all eight required affected packages pass; all 22 probes, counts, vet and scoped Cohere pass. The broader repository attempt has additional failures listed below.
Mutants: null-tag loss and folded null comparisons fail Node stdout on both backends; four additional compiled mutants and nine precision/receiver overlays are detailed below.
Coverage limits: full repository green is not claimed: CSS reaches the documented code-point refusal, and a workspace restart interrupted the remaining broad tests/Unicode retry; external Prettier audits and the reader's 28 unsupported probes are not covered.

Branch `codex/error-classes-counts`, starting at `4627ced`. Reader source is `coverage/error-classes` at `849de8e`. Only this feature branch may be pushed. Requested main `e011f8f` was merged into it, never the other way around. No PR was opened.

Both bugs were reproduced first with source Node, sanitized native, release native and generated JavaScript. `/tmp/adamic-reader-before.log` records exit 0 with differing stdout. `cause_null.a` is fixed-now in `2a8a59b`: boxed unknown null has its own immortal heap tag rather than the undefined pointer, nullable reference causes preserve that tag, and an unknown comparison is not folded from a non-nullable object representation. Both backends print `object true false`. `prototype_null.a` is fixed-now in the same commit: an incompatible null receiver throws a nominal, catchable TypeError through cleanup. This uses the existing ordinary IR throw machinery, so no refusal was needed. Both print `TypeError: Method Error.prototype.toString called on incompatible receiver null`.

All 19 `coverage_error_*.a` programs added by the reader are registered in the ordinary Node oracle. The two notes probes and one nullable-reference cause supplement add three more. Reader expressions are preserved; formatting and locally explained Cohere exemptions retain intentional temporal-dead-zone, mutation, generic, finally-override and heap-built-string tests. The reader's unsupported notes remain separate.

Merge conflicts and compatibility

The four code conflicts were resolved as follows; counts were a fifth textual conflict and were regenerated rather than resolved by choosing numbers.

| File | Resolution |
| --- | --- |
| `internal/ir/ir.go` | Keep main's accessor/setter call metadata and HasAccessor operation together with the branch's virtual targets, Pure call flag and nominal error data. |
| `internal/lower/class.go` | Keep main's static/instance method separation, accessor registration and super-accessor writes together with built-in error instances, Error override refusals and cause/stack restrictions. |
| `internal/lower/class_inheritance.go` | Keep both main's static constructor ancestry and nominal Error ancestry/structural-view refusals. |
| `internal/lower/object.go` | Keep both main's object-key support and the branch's Error prototype and instance library-call hooks. |

Automatic merge compatibility also required allowing a recognized built-in Error base through static class discovery, and replacing main's legacy private-brand TypeError and super-call ReferenceError objects with nominal error values. Their existing Node fixtures reproduced missing catch output before that repair. Generated errors use constant prefix literals, avoiding another checked call while constructing a failure. The merge commit is `ca8c83c`.

Main's new method-origin refusal rejected three reader interface calls (`interface_throw`, `padding_dispatch`, `tdz_interface`) whose signatures are callback-typed properties. Immediate property calls already use `Property.Method`, which binds class prototype methods and preserves own arrow closures. The import change allows those calls while refusing detached reads and destructuring. An existing negative test included an immediate call, so that entry became a positive assertion of receiver binding; actual detached-method refusal probes remain. The reader's class and subclass dispatch probes hold the behavior against Node.

Main also adds a JSON-slice gap test expecting a try around dynamic repeat to be refused. That gap closed in the earlier error-runtime unit `eb9fcd2`. The import commit `e21c5ae` keeps the original source as `TestRepeatInTryGapClosed`, comparing source Node, generated JavaScript and native with ASan/UBSan and leak checks for zero and three arguments. GAPS.md now describes the closed gap; no formatter program changed.

Main's `TestPassThroughsAreNotConsumers` exhaustively names every pure operation but did not include the branch's conditionally pure Call. Its failure was observed in the broad affected gate. Commit `2d2a598` now explicitly classifies `Call{Pure:true}` as a consumer with its own counted result, and asserts that ordinary calls and pure calls with impure arguments cannot lend operands. No production ownership rule or count was changed. The existing all-calls-pure source mutant still fails under ASan on allocated strings; the final matrix test passes.

Counts and attribution

Tuples are allocations, frees, retains, releases, peak live, in regions. Counts are measured, not edited by hand. There are two different comparisons: the earlier audit has 29 legitimate error-class residuals; merging main changes 20 existing fork rows, not 29. Both comparisons are fully listed below. The regenerated table has 342 numeric fixture rows, including all 19 imports. Relative to the requested current main, 40 common rows differ: those 29 earlier error rows plus 11 newly added throwing programs. No common non-throwing program differs from current main.

The original 29 rows are shown with both main baselines and the integration reader's fork. `be5a509` introduced nominal error construction and generated library/narrowing errors; `8465ed2` first recorded these changed counts, added nominal readiness failures and corrected uncaught output/cleanup. `154e646` removed false guard edges and inlined safe constant Error construction while keeping actual failures. Where main subsequently changed ownership, `fe62d30` prevents a narrowed read from being lent beyond the call that can free it. Attribution identifies source constructs and inspected IR/C mechanisms, not a profile assigning each retain to an instruction.

| Existing error row | Main `5d4c801` | Main `e011f8f` | Fork `4627ced` | Final | Commit and reason |
| --- | --- | --- | --- | --- | --- |
| `internal/oracle/testdata/library_array_with.a` | `68, 68, 57, 123, 10, 0` | `68, 68, 57, 123, 10, 0` | `68, 68, 62, 123, 10, 0` | `68, 68, 62, 123, 10, 0` | `be5a509`, recorded `8465ed2`: Executed out-of-range Array.with constructs nominal RangeError instead of generic Error; its catch/identity bookkeeping remains. |
| `internal/oracle/testdata/library_array_flat_map.a` | `50, 50, 66, 115, 17, 0` | `50, 50, 66, 115, 17, 0` | `50, 50, 68, 116, 17, 0` | `50, 50, 68, 116, 17, 0` | `be5a509`, recorded `8465ed2`: Executed callback throw new Error(`${value}`) uses nominal constructor message/cause ownership. |
| `internal/oracle/testdata/dead_zone.a` | `0, 0, 0, 0, 0, 0` | `0, 0, 0, 0, 0, 0` | `3, 3, 5, 5, 3, 0` | `3, 3, 5, 5, 3, 0` | `8465ed2`: Executed read of a global before initialization now constructs and cleans up a nominal ReferenceError. |
| `internal/oracle/testdata/precision_range.a` | `4, 3, 2, 5, 2, 0` | `4, 3, 2, 5, 2, 0` | `7, 6, 7, 11, 4, 0` | `7, 6, 7, 11, 4, 0` | `be5a509`, recorded `8465ed2`: Executed toPrecision(count) with invalid precision now constructs and cleans up nominal RangeError. |
| `internal/oracle/testdata/radix_range.a` | `3, 3, 0, 3, 1, 0` | `3, 3, 0, 3, 1, 0` | `6, 6, 5, 9, 3, 0` | `6, 6, 5, 9, 3, 0` | `be5a509`, recorded `8465ed2`: Executed toString(radix) with invalid radix now constructs and cleans up nominal RangeError. |
| `internal/oracle/testdata/narrowed_reads.a` | `5, 5, 0, 5, 4, 0` | `5, 5, 3, 7, 4, 0` | `8, 8, 6, 12, 4, 0` | `8, 8, 8, 14, 4, 0` | `be5a509`, recorded `8465ed2`; main ownership `fe62d30`: drop() invalidates chain; retain the actual TypeError edge and construction, remove only the earlier proven read guards. |
| `internal/oracle/testdata/narrowed_writes.a` | `1, 1, 2, 4, 1, 0` | `1, 1, 2, 4, 1, 0` | `4, 4, 7, 10, 3, 0` | `4, 4, 7, 10, 3, 0` | `be5a509`, recorded `8465ed2`: The invalidated optional receiver write executes a TypeError after its right side; keep nominal failure/cleanup. |
| `internal/oracle/testdata/narrowed_methods.a` | `3, 3, 1, 5, 3, 0` | `3, 3, 3, 6, 3, 0` | `6, 6, 7, 12, 3, 0` | `6, 6, 8, 13, 3, 0` | `be5a509`, recorded `8465ed2`; main ownership `fe62d30`: The invalidated optional method receiver executes a TypeError; remove earlier redundant guards only. |
| `internal/oracle/testdata/narrowed_fields.a` | `4, 3, 2, 5, 4, 0` | `4, 3, 4, 6, 4, 0` | `7, 6, 8, 12, 4, 0` | `7, 6, 9, 13, 4, 0` | `be5a509`, recorded `8465ed2`; main ownership `fe62d30`: cut() invalidates head.next and executes a TypeError; remove only the read guard before cut(). |
| `internal/oracle/testdata/fresh_writes.a` | `330, 330, 407, 523, 132, 0` | `330, 330, 409, 525, 132, 0` | `330, 330, 409, 524, 132, 0` | `330, 330, 411, 526, 132, 0` | `be5a509`, recorded `8465ed2`; main ownership `fe62d30`: Initialized parser reads added false exception cleanup; remove it, retaining executed throw new Error(`too many: ${count}`) construction. |
| `internal/oracle/testdata/exceptions.a` | `133, 133, 171, 246, 20, 0` | `133, 133, 171, 246, 20, 0` | `133, 133, 193, 257, 20, 0` | `133, 133, 193, 257, 20, 0` | `be5a509`, recorded `8465ed2`: Executed message-building Error throws gain nominal constructor ownership; constant Error throws now inline without extra handoffs. |
| `internal/oracle/testdata/exceptions_uncaught.a` | `3, 1, 3, 5, 2, 0` | `3, 1, 3, 5, 2, 0` | `5, 5, 5, 9, 4, 0` | `5, 5, 5, 9, 4, 0` | `be5a509`, recorded `8465ed2`; uncaught cleanup `8465ed2`: Executed throw new Error(`${label} failed`) gains nominal ownership and uncaught formatting/complete cleanup. |
| `internal/oracle/testdata/exceptions_empty.a` | `2, 1, 6, 4, 2, 0` | `2, 1, 6, 4, 2, 0` | `3, 2, 9, 8, 2, 0` | `3, 2, 9, 8, 2, 0` | `be5a509`, recorded `8465ed2`; uncaught cleanup `8465ed2`: Executed throw new Error() and rethrow now use nominal error representation and complete uncaught formatting/cleanup. |
| `internal/oracle/testdata/closures_throw.a` | `418, 418, 1571, 1806, 142, 0` | `418, 418, 1571, 1806, 142, 0` | `418, 418, 1599, 1820, 142, 0` | `418, 418, 1599, 1820, 142, 0` | `be5a509`, recorded `8465ed2`: Executed dynamic-message Error throws through callbacks gain nominal ownership; constant constructors now inline. |
| `internal/oracle/testdata/closures_throw_uncaught.a` | `31, 29, 20, 36, 13, 0` | `31, 29, 20, 36, 13, 0` | `33, 33, 22, 40, 13, 0` | `33, 33, 22, 40, 13, 0` | `be5a509`, recorded `8465ed2`; uncaught cleanup `8465ed2`: Executed callback throw new Error(`${item} is too loud`) gains nominal ownership and uncaught formatting/cleanup. |
| `internal/oracle/testdata/finally_leaves.a` | `157, 157, 61, 176, 10, 0` | `157, 157, 61, 176, 10, 0` | `157, 157, 85, 188, 10, 0` | `157, 157, 85, 188, 10, 0` | `be5a509`, recorded `8465ed2`: Executed throw new Error(word(...)) on return/break/continue/finally paths gains nominal constructor ownership. |
| `internal/oracle/testdata/param_assigned_in_try.a` | `46, 46, 32, 67, 12, 0` | `46, 46, 32, 67, 12, 0` | `46, 46, 34, 68, 12, 0` | `46, 46, 34, 68, 12, 0` | `be5a509`, recorded `8465ed2`: Executed throw new Error(word('too big ', value)) gains nominal ownership; constant 'empty' construction now inlines. |
| `internal/oracle/testdata/named_function_values.a` | `116, 116, 129, 224, 25, 0` | `116, 116, 129, 224, 25, 0` | `116, 116, 131, 225, 25, 0` | `116, 116, 131, 225, 25, 0` | `be5a509`, recorded `8465ed2`: Executed throw new Error(`failing at ${value}`) through stored function values gains nominal ownership. |
| `internal/oracle/testdata/string_too_long.a` | `0, 0, 0, 0, 0, 0` | `0, 0, 0, 0, 0, 0` | `3, 3, 6, 7, 3, 0` | `3, 3, 6, 7, 3, 0` | `be5a509`, recorded `8465ed2`: Executed 'ab'.repeat(2 ** 28) now constructs and cleans up a nominal RangeError. |
| `internal/oracle/testdata/normalize_form.a` | `3, 2, 0, 3, 3, 0` | `3, 2, 0, 3, 3, 0` | `6, 5, 5, 9, 4, 0` | `6, 5, 5, 9, 4, 0` | `be5a509`, recorded `8465ed2`: Executed normalize('nfc') constructs nominal RangeError; safe preliminary normalization no longer adds a retain. |
| `internal/oracle/testdata/writes_in_try.a` | `30, 30, 29, 41, 6, 0` | `30, 30, 29, 41, 6, 0` | `30, 30, 31, 42, 6, 0` | `30, 30, 31, 42, 6, 0` | `be5a509`, recorded `8465ed2`: Executed throw new Error(['th', 'rown'].join('')) gains nominal constructor ownership. |
| `internal/oracle/testdata/class_as_interface.a` | `372, 372, 360, 525, 60, 0` | `372, 372, 360, 525, 60, 0` | `372, 372, 362, 526, 60, 0` | `372, 372, 362, 526, 60, 0` | `be5a509`, recorded `8465ed2`: Interface calls execute throw new Error(built('negative', value)); remove false receiver edges and inline constant 'empty' errors. |
| `internal/oracle/testdata/regions_throw.a` | `124, 95, 53, 143, 31, 29` | `124, 95, 53, 143, 31, 29` | `124, 95, 59, 146, 31, 29` | `124, 95, 59, 146, 31, 29` | `be5a509`, recorded `8465ed2`: False checks blocked 29 region values; restore them while keeping executed throw new Error(`budget spent at depth ${depth}`) ownership. |
| `internal/oracle/testdata/move_throw.a` | `42, 42, 22, 50, 9, 0` | `42, 42, 22, 50, 9, 0` | `42, 42, 32, 55, 9, 0` | `42, 42, 32, 55, 9, 0` | `be5a509`, recorded `8465ed2`: Executed dynamic-message Error throws on inner/cleanup paths gain nominal constructor ownership. |
| `internal/oracle/testdata/generic_method_return.a` | `23, 15, 28, 36, 5, 6` | `23, 15, 28, 36, 5, 6` | `25, 19, 32, 41, 5, 6` | `25, 19, 32, 41, 5, 6` | `be5a509`, recorded `8465ed2`; uncaught cleanup `8465ed2`: Executed caught and uncaught throw new Error(`missing ${key}`) gains nominal ownership plus uncaught formatting/complete cleanup. |
| `internal/oracle/testdata/reuse_throw.a` | `26, 26, 24, 41, 8, 0` | `26, 26, 24, 41, 8, 0` | `26, 26, 30, 44, 8, 0` | `26, 26, 30, 44, 8, 0` | `be5a509`, recorded `8465ed2`: Executed throw new Error(`refused at ${next.count}`) gains nominal constructor ownership. |
| `internal/oracle/testdata/reuse_narrowed.a` | `7, 5, 4, 7, 5, 0` | `7, 5, 6, 8, 5, 0` | `10, 8, 10, 14, 5, 0` | `10, 8, 11, 15, 5, 0` | `be5a509`, recorded `8465ed2`; main ownership `fe62d30`: clear() invalidates box.inner and executes a TypeError; remove only the earlier same-path proven read guards. |
| `internal/oracle/testdata/regexp_null_narrowed.a` | `4, 4, 7, 5, 3, 0` | `4, 4, 8, 5, 3, 0` | `7, 7, 13, 12, 3, 0` | `7, 7, 13, 12, 3, 0` | `be5a509` and null guard repair `8465ed2`: A call invalidates a null-excluding regex match narrowing; keep the actually executed nominal TypeError. |
| `internal/oracle/testdata/class_inheritance_exceptions.a` | `29, 29, 22, 42, 8, 0` | `29, 29, 22, 42, 8, 0` | `29, 29, 24, 43, 8, 0` | `29, 29, 24, 43, 8, 0` | `be5a509`, recorded `8465ed2`: Executed dynamic throw new Error(this.label + '/' + this.detail) gains nominal ownership; constant base/derived errors inline. |

Every existing fork row changed by the merge:

| Fixture | Fork `4627ced` | Final | Commit and why correct |
| --- | --- | --- | --- |
| `internal/oracle/testdata/class_oct6_subclass_holder.a` | `76, 76, 49, 110, 12, 0` | `76, 76, 55, 116, 12, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/load/testdata/0.1/compile/09_tree.ts` | `19, 19, 82, 123, 16, 0` | `19, 19, 85, 126, 16, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/casts.a` | `9, 9, 16, 24, 6, 0` | `9, 9, 17, 25, 6, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/visits.a` | `95, 95, 124, 208, 23, 0` | `95, 95, 125, 209, 23, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/narrowed_reads.a` | `8, 8, 6, 12, 4, 0` | `8, 8, 8, 14, 4, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/narrowed_methods.a` | `6, 6, 7, 12, 3, 0` | `6, 6, 8, 13, 3, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/narrowed_fields.a` | `7, 6, 8, 12, 4, 0` | `7, 6, 9, 13, 4, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/fills.a` | `24, 24, 27, 50, 11, 0` | `24, 24, 28, 51, 11, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/fresh_writes.a` | `330, 330, 409, 524, 132, 0` | `330, 330, 411, 526, 132, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/fresh_calls.a` | `147, 147, 266, 321, 39, 0` | `147, 147, 268, 323, 39, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/weak_narrowed.a` | `26, 26, 50, 69, 11, 0` | `26, 26, 53, 72, 11, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/undefined_keys.a` | `122, 122, 171, 236, 24, 0` | `122, 122, 173, 238, 24, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/undefined_strings.a` | `15, 15, 20, 42, 6, 0` | `15, 15, 21, 43, 6, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/undefined_references.a` | `17, 17, 36, 51, 11, 0` | `17, 17, 37, 52, 11, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/reuse_narrowed.a` | `10, 8, 10, 14, 5, 0` | `10, 8, 11, 15, 5, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/reuse_lent_global.a` | `10, 10, 7, 14, 7, 0` | `10, 10, 8, 15, 7, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/regexp.a` | `455, 455, 355, 408, 62, 0` | `455, 455, 369, 422, 62, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/regexp_exec.a` | `179, 179, 81, 159, 18, 0` | `179, 179, 82, 160, 18, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/regexp_unicode.a` | `108, 108, 99, 86, 21, 0` | `108, 108, 101, 88, 21, 0` | `fe62d30`: Defined is a pass-through, not an ownership consumer; keep a count across later arguments/calls that can free the narrowed value. These ownership safeguards are independent of MayThrow and exactly match current main for non-throwing programs. |
| `internal/oracle/testdata/class_inheritance_generic.a` | `64, 64, 74, 121, 24, 0` | `89, 89, 104, 167, 39, 0` | `7b30af6`: fixture adds makePair/forwardPair/genericRead and Projected factory/constructor exercises; the program performs additional work. |

The eleven additional error rows introduced on main after the original audit:

| Fixture | Main `e011f8f` | Final | Commit and source failure |
| --- | --- | --- | --- |
| `internal/oracle/testdata/borrow_element_throw.a` | `38, 38, 34, 54, 8, 0` | `38, 38, 44, 59, 8, 0` | `be5a509`: five executed dynamic-message Error throws, including removeForMessage(items), require nominal message ownership. |
| `internal/oracle/testdata/throw_keeps_old_value.a` | `18, 18, 9, 21, 6, 0` | `18, 18, 13, 23, 6, 0` | `be5a509`: fail throws a dynamic-message Error on both spread and consumed-call paths. |
| `internal/oracle/testdata/throw_keeps_old_value_variants.a` | `26, 26, 11, 32, 6, 0` | `26, 26, 17, 35, 6, 0` | `be5a509`: three dynamic-message Error throws through swallow, finally and array-spread paths. |
| `internal/oracle/testdata/throw_in_writes.a` | `22, 22, 8, 22, 8, 0` | `22, 22, 12, 24, 8, 0` | `be5a509`: failNumber throws a dynamic-message Error in both the field-write and element-write case. |
| `internal/oracle/testdata/throw_global_move.a` | `18, 18, 12, 26, 6, 0` | `18, 18, 18, 29, 6, 0` | `be5a509`: three executed Error throws in fail and insertOrThrow use dynamic messages. |
| `internal/oracle/testdata/class_features_static.a` | `110, 110, 84, 191, 28, 0` | `110, 110, 90, 194, 28, 0` | `be5a509`: StaticFailure method, getter and setter execute three dynamic-message Error throws. |
| `internal/oracle/testdata/class_features_static_private.a` | `42, 42, 71, 112, 12, 0` | `42, 42, 66, 107, 12, 0` | `ca8c83c`: five actual private-brand failures now construct nominal TypeErrors; direct constant prefixes omit legacy name-store ownership, reducing retains/releases by five. |
| `internal/oracle/testdata/class_features_accessors.a` | `67, 67, 53, 110, 21, 0` | `67, 67, 57, 112, 21, 0` | `be5a509`: Fallible getter and setter each execute a dynamic-message Error throw. |
| `internal/oracle/testdata/class_inheritance_conditional.a` | `164, 164, 148, 251, 35, 0` | `164, 164, 144, 247, 35, 0` | `ca8c83c`: missing or repeated super executes four nominal ReferenceErrors; direct constant prefixes omit legacy name-store ownership, reducing retains/releases by four. Constant explicit Error throws inline. |
| `internal/oracle/testdata/user_iterators.a` | `669, 669, 466, 915, 67, 0` | `669, 669, 480, 922, 67, 0` | `be5a509`: seven executed dynamic-message Error throws in loop bodies, mapper, destructuring expression and overridden iterator; constant Error throws inline. |
| `internal/oracle/testdata/user_iterators_rest_tdz.a` | `5, 0, 5, 3, 5, 0` | `8, 3, 10, 9, 8, 0` | `8465ed2`: iterator.next reads rest before destructuring initializes it and constructs a nominal ReferenceError; uncaught formatting and cleanup now execute. |

The new main non-throwing `regexp_cycle_fields.a` initially rose from `61, 61, 134, 132, 44, 0` to `61, 61, 158, 156, 44, 0` on the merge. The imported analysis previously dropped every loop-condition presence fact. The final precision fix derives presence at each successful condition, starts back edges without earlier-iteration facts, and preserves only binding facts across ArrayPush/RegExpCall runtime operations that cannot reassign bindings. Field facts still disappear, calls and writes retain their invalidation behavior, and the final six counts exactly equal main. `TestLoopPresenceSurvivesRuntimeMutation` tests this precision independently of counts; a source overlay removing the loop proof fails it.

`regions.a`, `reuse.a`, `09_tree.ts`, the regexp sweep and every other common non-throwing program equal current main. The tree now has 85 retains rather than old main's 82 because of `fe62d30`, not a reintroduced exception edge. The original 63-row report remains historical and explicitly names its old main comparison.

Mutation evidence

| Mutant | What catches it |
| --- | --- |
| Drop boxed null tag | `TestReaderNullMutants/drop_boxed_null_tag`: both backends still compile and exit 0 but print `undefined false true`; Node prints `object true false`. |
| Fold unknown null comparison again | Same test's fold case: both backends exit 0 but print `object false false`; Node prints `object true false`. |
| Drop nullable-reference null tag | Nullable RegExp.exec cause probe catches native `undefined false true` at exit 0. JavaScript already retains null by identity, so this native-only representation mutant does not alter it. |
| Null prototype receiver returns Error | Both backends exit 0 printing Error instead of Node's caught TypeError line. |
| Remove TypeError nominal ancestry | `TestMergedGeneratedErrorMutants/TypeError`: existing private static fixture's catch output differs from Node on both backends at exit 0. |
| Remove ReferenceError nominal ancestry | Same test's ReferenceError case: conditional-super fixture loses Node's catch output on both backends at exit 0. |
| Disable callback-typed method receiver binding | `TestCallbackTypedClassMethodCallBindsReceiver` fails its receiver assertion independently of Node and counts. |
| Remove loop-condition presence proof | `TestLoopPresenceSurvivesRuntimeMutation` fails without any count-table comparison. |
| Mark every call may-throw | `TestMayThrowPrecision` rejects false call edges independently of counts. |
| Mark every function may-throw | The same precision assertions reject false function edges. |
| Leave readiness unproved | The same precision assertions reject a false global-read exception edge. |
| Ignore unknown numeric field stores | `TestGeneratedGuardFactsIncludeUnknownStores` requires the format guard to remain throwing. |
| Ignore counter body writes | `TestGeneratedGuardCounterWritesRemainThrowing` requires the loop format guard to remain throwing. |
| Ignore call mutation in presence facts | Node's `error_checks.a` catches the TypeError; mutant exits differently. |
| Mark every call pure | ASan catches use-after-free in heap-built `lent_reads.a`. |

The six compiled mutants run as ordinary oracle tests. The nine source overlays run with `python3 cloud/error-classes-counts-mutants.py`; only the intended failing assertion counts, never a compiler diagnostic. Existing exception under-approximation mutants remain in the full oracle gate.

Setup and commands

`bash cloud/setup.sh > /tmp/adamic-reader-setup.log 2>&1` succeeded. Then every Go command sources `/workspace/adamic-tools/env.sh` and uses `TMPDIR=/tmp/adamic-gate`, mode 1777. `nproc` is 5. Go 1.27.1, clang 20.1.8 and Node 24.19.0 were used.

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (0s)
setup: node ready (0s)
setup: submodules ready (0s)
setup: build cache warm (111s)
setup: done in 111s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

Final validation output and implementation SHAs follow below. All test output went directly to logs; intermediate failures were corrected and are not waived.

Already completed final checks:

| Command | Output and log |
| --- | --- |
| `go test -count=1 -v ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/coverage_error'` | exit 0; all 22 probes pass, 21.465s; `/tmp/adamic-reader-fixtures-final.log` |
| `go test -count=1 -v ./internal/oracle -run 'TestReaderNullMutants|TestMergedGeneratedErrorMutants|TestFixturesAgreeWithNode/coverage_error'` | six compiled mutants pass by detecting Node stdout differences; `/tmp/adamic-reader-final-probes.log` (also contains an unused fixture-test alternative; actual fixture coverage is the preceding row) |
| `python3 cloud/error-classes-counts-mutants.py` | exit 0; all nine intended assertions caught, with no build failures; `/tmp/adamic-reader-precision-mutants-final.log` |
| `go test -count=1 ./internal/oracle -run TestCountsAreRecorded -args -update-counts` | exit 0, 77.675s, 342 numeric rows; `/tmp/adamic-reader-final-counts.log` |
| `go test -count=1 ./stage1/cohere/json -run 'TestDocumentedStageZeroGaps|TestRepeatInTryGapClosed'` | exit 0, 4.560s; `/tmp/adamic-reader-json-gap.log` |
| `go test -count=1 ./internal/native -run TestPassThroughsAreNotConsumers` | exit 0, 0.050s; `/tmp/adamic-reader-consumers.log` |
| `gofmt -l cmd internal stage1/cohere/json` and `go vet ./...` | exit 0, no output; `/tmp/adamic-reader-gofmt-final.log`, `/tmp/adamic-reader-vet-final.log` |
| Cohere `--no-fix --no-cache` on byte-identical `.ts` mirrors of the 22 oracle fixtures | exit 0, 276 rules, 22 checked, 100% Adamic-ready; `/tmp/adamic-reader-cohere-final.log`. Prelude is loaded for types but not one of the 22 scoped paths. |

Intermediate gate failures were the outdated immediate-call refusal expectation and the missing pure Call entry in main's ownership matrix. Both tests were integrated, not suppressed. The affected uncached run at `/tmp/adamic-reader-gate.log` otherwise passed flow, fresh, oracle and GraphQL, with all native runtime tests completing and only the matrix assertion failing. The subsequent repository runs and final gate outcome are recorded below.

Final affected-package results

The final source state is `2d2a5983f11c57e7f11b65af570910443d22c6e9`. The command was:

```sh
source /workspace/adamic-tools/env.sh
go test -count=1 -timeout 30m ./... > /tmp/adamic-reader-final-gate.log 2>&1
```

Every required affected package completed successfully in that run:

```text
internal/ir: no test files
internal/lower: ok 93.322s
internal/javascript: no test files
internal/flow: ok 419.700s
internal/fresh: ok 163.173s
internal/native: ok 802.656s
internal/oracle: ok 656.366s
stage1/cohere/graphql: ok 160.149s
```

This is not a claim that the full repository invocation was green. It additionally observed:

- `cmd/adamic-test262/TestLargeCompilerOutputIsComplete` failed with an empty `clang:` reason while broad runs overlapped. Its clang phase has a two-minute budget. The unchanged test passed on sequential retry in 50.573s, exit 0: `/tmp/adamic-reader-isolated-large-output.log`.
- `internal/unicodeproperties/TestCanonicalizeUnicodeNode` reported `node: signal: killed`. Its Node batches have four-minute budgets. Contention is an inference from the overlapping runs and these limits, not a proved cause. Its sequential retry was running when the workspace restarted; no passing result is claimed for that retry.
- `stage1/cohere/css` refuses `compose.ts:290:5`, a try reaching `String.fromCodePoint`. This is the deliberate, documented boundary introduced in the earlier catchability unit `7d4faab`: an unproven code-point failure still panics natively and cannot unwind yet. The new main CSS slice reaches that boundary. The refusal was retained; neither it nor the CSS expectations were weakened to produce a green broad gate. A separate code-point guard/range-proof extension is required to make that slice compile soundly.

The superseded repository run was stopped after its old ownership-matrix failure to release its remaining subprocesses. The final run was interrupted by the environment restart after recording all eight required passing package results above, the two subprocess failures and the CSS refusal. Later stage-1 package results and a complete repository exit status are not claimed. This distinction is also in the five-line summary.

The 19 imported reader fixtures were compared with `849de8e`: their tokens agree after ignoring comments, whitespace and trailing commas. No expression was simplified to an immortal literal to hide an ownership failure. The two null probes, the nullable supplement and all reader imports have ordinary oracle flags, with no sanitizer exemptions.

Implementation commits, in order:

- `2a8a59b`: preserve null causes and throw for the incompatible null receiver, before merging main.
- `ca8c83c`: merge requested main `e011f8f` and repair nominal generated class failures.
- `e21c5ae27cc1106fcdbb2761628a00290f00ff45`: import the reader, restore loop precision and receiver dispatch, close the stale JSON gap expectation, regenerate counts.
- `2d2a5983f11c57e7f11b65af570910443d22c6e9`: integrate main's ownership classification and prove receiver binding with a source mutant.

The following report commit changes only this report and the historical audit link. Only `codex/error-classes-counts` is pushed. No main or area branch was pushed or merged into; no PR was opened.
