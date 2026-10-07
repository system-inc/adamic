Restored all affected non-throwing programs to main's exact six-column counts.
The 63-row audit separates false exception edges from executed nominal error construction.
MayThrow remains monotone; caller readiness, stable/path presence and scalar bounds remove false seeds.
Independent precision assertions catch over-approximation; Node and sanitizers hold under-approximation and ownership.
The final affected gate, mutation results and coverage limits are recorded below.

This is the historical audit against main `5d4c801`, with final counts from `154e646`/`379ea1f`. The later merge of main `e011f8f`, all 29 remaining error rows, all 20 actual existing-fork merge deltas, and the reader's null fixes are audited in [error-reader-integration-report.md](error-reader-integration-report.md). Main's ownership fixes legitimately change some non-throwing counts from the old baseline; the merged branch equals current main for those programs.

Branch: `codex/error-classes-counts`, created from `origin/codex/error-classes` at `d0b644303a201ec319d85bc59a360a16daa6e929`, not from main. Main comparison: `5d4c801`.

Observed: exactly 63 existing rows differed between those commits. All 32 affected programs that execute no throw now equal main in every column, including `object_prototype.a`, which constructs an Error without throwing it. Two additional throwing programs also return exactly to main because their constant Error constructors inline. The remaining 29 rows execute explicit Error throws or generated nominal failures. No previously unchanged main row moves. The tuple order throughout is **allocations, frees, retains, releases, peak live, in regions**; this includes frees as well as every requested metric.

The lead's inference is largely confirmed, with two qualifications. Monotone MayThrow growth is required for correctness, not itself a defect: an unchecked read must not erase a previously discovered throw. Newly inserted guards were unconditional seeds even when their failure was excluded, and their ordinary function-call representation independently disabled operand borrowing. The resulting false cleanup paths blocked regions and reuse. `regions.a` lost all 56 region values; `reuse.a` allocated four extra objects; the tree allocated 14 extra objects; the regexp sweep allocated 888 extra objects. All four are restored exactly. Executed error constructors and corrected uncaught formatting/cleanup explain the legitimate residuals, rather than those false edges.

The causes below identify the actual source construct or analysis mechanism. Counts are observations from generated tables; attribution is based on inspecting the source, generated IR/C and the effect of the precision changes, not a runtime profile assigning individual retains to stack frames.

| Fixture | Main `5d4c801` | Original error branch `d0b6443` | Final | Cause of original delta and disposition |
| --- | --- | --- | --- | --- |
| `internal/oracle/testdata/class_oct6_release.a` | `372, 372, 180, 441, 26, 0` | `372, 372, 192, 453, 26, 0` | `372, 372, 180, 441, 26, 0` | Proven local optional class receivers were wrapped as throwing reads; remove those edges and their cleanup retains. |
| `internal/oracle/testdata/class_oct6_subclass_holder.a` | `76, 76, 49, 110, 12, 0` | `76, 76, 61, 122, 12, 0` | `76, 76, 49, 110, 12, 0` | Stable saved optional child reads were wrapped as throwing reads; preserve the proven narrowing. |
| `internal/oracle/testdata/library_array_with.a` | `68, 68, 57, 123, 10, 0` | `68, 68, 62, 123, 10, 0` | `68, 68, 62, 123, 10, 0` | Executed out-of-range Array.with constructs nominal RangeError instead of generic Error; its catch/identity bookkeeping remains. |
| `internal/oracle/testdata/library_array_flat_map.a` | `50, 50, 66, 115, 17, 0` | `50, 50, 68, 116, 17, 0` | `50, 50, 68, 116, 17, 0` | Executed callback throw new Error(`${value}`) uses nominal constructor message/cause ownership. |
| `internal/oracle/testdata/library_function_expressions.a` | `56, 56, 49, 96, 19, 0` | `56, 56, 51, 97, 19, 0` | `56, 56, 49, 96, 19, 0` | Executed callback throw new Error('expression throw') gained constructor handoff retains; direct nominal constant construction removes them. |
| `internal/load/testdata/0.1/compile/03_shapes.ts` | `13, 13, 17, 25, 7, 0` | `13, 13, 20, 28, 7, 0` | `13, 13, 17, 25, 7, 0` | Non-failing initialized-global reads acquired ready-check exception edges; intersect readiness at every caller. |
| `internal/load/testdata/0.1/compile/09_tree.ts` | `19, 19, 82, 123, 16, 0` | `33, 33, 227, 230, 16, 0` | `19, 19, 82, 123, 16, 0` | Proven optional tree receivers gained throwing wrappers, losing Perceus reuse and adding cleanup retains; keep stable narrowing. |
| `internal/oracle/testdata/dead_zone.a` | `0, 0, 0, 0, 0, 0` | `3, 3, 5, 5, 3, 0` | `3, 3, 5, 5, 3, 0` | Executed read of a global before initialization now constructs and cleans up a nominal ReferenceError. |
| `internal/oracle/testdata/string_index.a` | `89, 89, 29, 126, 7, 0` | `89, 89, 34, 131, 7, 0` | `89, 89, 29, 126, 7, 0` | Proven local optional code-unit reads gained throwing wrappers; retain the stable local narrowing. |
| `internal/oracle/testdata/visits.a` | `95, 95, 124, 208, 23, 0` | `95, 95, 126, 210, 23, 0` | `95, 95, 124, 208, 23, 0` | Proven optional find result gained throwing read wrappers; retain the stable local narrowing. |
| `internal/oracle/testdata/navigation.a` | `162, 162, 40, 189, 16, 0` | `162, 162, 48, 197, 16, 0` | `162, 162, 40, 189, 16, 0` | Proven optional route endpoints gained throwing read wrappers; retain the stable local narrowing. |
| `internal/oracle/testdata/number_formats.a` | `235, 235, 46, 263, 18, 0` | `235, 235, 88, 305, 18, 0` | `235, 235, 46, 263, 18, 0` | Safe digits/figures and pure format guards lost borrowing; prove closed field ranges, loop bounds and returned digits, and preserve successful-path purity. |
| `internal/oracle/testdata/precision_range.a` | `4, 3, 2, 5, 2, 0` | `7, 6, 7, 11, 4, 0` | `7, 6, 7, 11, 4, 0` | Executed toPrecision(count) with invalid precision now constructs and cleans up nominal RangeError. |
| `internal/oracle/testdata/radixes.a` | `696, 696, 97, 332, 25, 0` | `696, 696, 101, 336, 25, 0` | `696, 696, 97, 332, 25, 0` | Valid radix guards gained impure call boundaries; literal guards disappear and remaining pure guard calls preserve operand borrowing. |
| `internal/oracle/testdata/radix_range.a` | `3, 3, 0, 3, 1, 0` | `6, 6, 5, 9, 3, 0` | `6, 6, 5, 9, 3, 0` | Executed toString(radix) with invalid radix now constructs and cleans up nominal RangeError. |
| `internal/oracle/testdata/narrowed_reads.a` | `5, 5, 0, 5, 4, 0` | `8, 8, 11, 17, 4, 0` | `8, 8, 6, 12, 4, 0` | drop() invalidates chain; retain the actual TypeError edge and construction, remove only the earlier proven read guards. |
| `internal/oracle/testdata/narrowed_writes.a` | `1, 1, 2, 4, 1, 0` | `4, 4, 7, 10, 3, 0` | `4, 4, 7, 10, 3, 0` | The invalidated optional receiver write executes a TypeError after its right side; keep nominal failure/cleanup. |
| `internal/oracle/testdata/narrowed_methods.a` | `3, 3, 1, 5, 3, 0` | `6, 6, 9, 14, 3, 0` | `6, 6, 7, 12, 3, 0` | The invalidated optional method receiver executes a TypeError; remove earlier redundant guards only. |
| `internal/oracle/testdata/narrowed_fields.a` | `4, 3, 2, 5, 4, 0` | `7, 6, 10, 14, 4, 0` | `7, 6, 8, 12, 4, 0` | cut() invalidates head.next and executes a TypeError; remove only the read guard before cut(). |
| `internal/oracle/testdata/fills.a` | `24, 24, 27, 50, 11, 0` | `24, 24, 29, 52, 11, 0` | `24, 24, 27, 50, 11, 0` | Safe repeat wrappers and proven optional row reads added retains; preserve primitive purity and stable narrowing. |
| `internal/oracle/testdata/weak_parent.a` | `86, 86, 264, 313, 41, 0` | `86, 86, 278, 327, 41, 0` | `86, 86, 264, 313, 41, 0` | Proven optional weak-parent traversal acquired throwing wrappers; stable locals and same-path weak presence recover borrowing. |
| `internal/oracle/testdata/doubly_linked.a` | `90, 90, 545, 578, 49, 0` | `90, 90, 619, 652, 49, 0` | `90, 90, 545, 578, 49, 0` | Proven weak-link locals and repeat(1) acquired throwing/impure wrappers; keep stable narrowing and the safe primitive. |
| `internal/oracle/testdata/fresh_parser.a` | `294, 294, 338, 508, 34, 0` | `294, 294, 341, 511, 34, 0` | `294, 294, 338, 508, 34, 0` | Already-initialized parser globals acquired ready-check edges; intersect readiness over callers to remove cleanup retains. |
| `internal/oracle/testdata/fresh_writes.a` | `330, 330, 407, 523, 132, 0` | `330, 330, 415, 530, 132, 0` | `330, 330, 409, 524, 132, 0` | Initialized parser reads added false exception cleanup; remove it, retaining executed throw new Error(`too many: ${count}`) construction. |
| `internal/oracle/testdata/fresh_calls.a` | `147, 147, 266, 321, 39, 0` | `147, 147, 270, 325, 39, 0` | `147, 147, 266, 321, 39, 0` | Initialized parser globals and stable optional relatives acquired false exception edges; prove their readiness/narrowing. |
| `internal/oracle/testdata/weak_narrowed.a` | `26, 26, 50, 69, 11, 0` | `26, 26, 56, 75, 11, 0` | `26, 26, 50, 69, 11, 0` | Proven optional weak receivers gained TypeError wrappers; preserve stable locals and weak-target presence facts. |
| `internal/oracle/testdata/exceptions.a` | `133, 133, 171, 246, 20, 0` | `133, 133, 199, 260, 20, 0` | `133, 133, 193, 257, 20, 0` | Executed message-building Error throws gain nominal constructor ownership; constant Error throws now inline without extra handoffs. |
| `internal/oracle/testdata/exceptions_uncaught.a` | `3, 1, 3, 5, 2, 0` | `5, 5, 5, 9, 4, 0` | `5, 5, 5, 9, 4, 0` | Executed throw new Error(`${label} failed`) gains nominal ownership and uncaught formatting/complete cleanup. |
| `internal/oracle/testdata/exceptions_empty.a` | `2, 1, 6, 4, 2, 0` | `3, 2, 9, 8, 2, 0` | `3, 2, 9, 8, 2, 0` | Executed throw new Error() and rethrow now use nominal error representation and complete uncaught formatting/cleanup. |
| `internal/oracle/testdata/closures_throw.a` | `418, 418, 1571, 1806, 142, 0` | `418, 418, 1609, 1825, 142, 0` | `418, 418, 1599, 1820, 142, 0` | Executed dynamic-message Error throws through callbacks gain nominal ownership; constant constructors now inline. |
| `internal/oracle/testdata/closures_throw_uncaught.a` | `31, 29, 20, 36, 13, 0` | `33, 33, 22, 40, 13, 0` | `33, 33, 22, 40, 13, 0` | Executed callback throw new Error(`${item} is too loud`) gains nominal ownership and uncaught formatting/cleanup. |
| `internal/oracle/testdata/finally_leaves.a` | `157, 157, 61, 176, 10, 0` | `157, 157, 85, 188, 10, 0` | `157, 157, 85, 188, 10, 0` | Executed throw new Error(word(...)) on return/break/continue/finally paths gains nominal constructor ownership. |
| `internal/oracle/testdata/param_assigned_in_try.a` | `46, 46, 32, 67, 12, 0` | `46, 46, 36, 69, 12, 0` | `46, 46, 34, 68, 12, 0` | Executed throw new Error(word('too big ', value)) gains nominal ownership; constant 'empty' construction now inlines. |
| `internal/oracle/testdata/named_function_values.a` | `116, 116, 129, 224, 25, 0` | `116, 116, 131, 225, 25, 0` | `116, 116, 131, 225, 25, 0` | Executed throw new Error(`failing at ${value}`) through stored function values gains nominal ownership. |
| `internal/oracle/testdata/unions.a` | `1449, 1449, 1111, 2573, 16, 0` | `1449, 1449, 1113, 2575, 16, 0` | `1449, 1449, 1111, 2573, 16, 0` | Safe repeat/toFixed and initialized mutable global reads acquired guard-call overhead; prove constants/readiness and preserve purity. |
| `internal/oracle/testdata/undefined_elements.a` | `20, 20, 46, 43, 9, 0` | `20, 20, 49, 46, 9, 0` | `20, 20, 46, 43, 9, 0` | Safe literal repeat calls gained guard handoffs; prove repeat count/length and restore direct primitive reads. |
| `internal/oracle/testdata/string_too_long.a` | `0, 0, 0, 0, 0, 0` | `3, 3, 6, 7, 3, 0` | `3, 3, 6, 7, 3, 0` | Executed 'ab'.repeat(2 ** 28) now constructs and cleans up a nominal RangeError. |
| `internal/oracle/testdata/normalize.a` | `213, 213, 96, 274, 11, 0` | `213, 213, 103, 281, 11, 0` | `213, 213, 96, 274, 11, 0` | Valid normalization forms gained impure guard calls; prove forms and preserve successful-path primitive borrowing. |
| `internal/oracle/testdata/normalize_form.a` | `3, 2, 0, 3, 3, 0` | `6, 5, 6, 10, 4, 0` | `6, 5, 5, 9, 4, 0` | Executed normalize('nfc') constructs nominal RangeError; safe preliminary normalization no longer adds a retain. |
| `internal/oracle/testdata/long_literals.a` | `14518, 14518, 85, 14610, 8, 0` | `14518, 14518, 86, 14611, 8, 0` | `14518, 14518, 85, 14610, 8, 0` | Safe literal repeat gained an impure guard boundary; prove count/length and restore primitive borrowing. |
| `internal/oracle/testdata/writes_in_try.a` | `30, 30, 29, 41, 6, 0` | `30, 30, 31, 42, 6, 0` | `30, 30, 31, 42, 6, 0` | Executed throw new Error(['th', 'rown'].join('')) gains nominal constructor ownership. |
| `internal/oracle/testdata/class_as_interface.a` | `372, 372, 360, 525, 60, 0` | `372, 372, 366, 530, 60, 0` | `372, 372, 362, 526, 60, 0` | Interface calls execute throw new Error(built('negative', value)); remove false receiver edges and inline constant 'empty' errors. |
| `internal/oracle/testdata/reuse.a` | `76, 76, 57, 122, 16, 0` | `80, 80, 95, 156, 16, 0` | `76, 76, 57, 122, 16, 0` | Proven optional recursive receivers gained throw edges, losing Perceus reuse and adding cleanup retains; preserve stable narrowing. |
| `internal/oracle/testdata/reuse_weak_during_spread.a` | `10, 10, 8, 16, 8, 0` | `10, 10, 9, 17, 8, 0` | `10, 10, 8, 16, 8, 0` | Proven optional weak receiver acquired a throwing wrapper; preserve stable local narrowing through the spread callback. |
| `internal/oracle/testdata/regions.a` | `316, 260, 219, 409, 50, 56` | `316, 316, 607, 746, 50, 0` | `316, 260, 219, 409, 50, 56` | False optional-receiver/global-read exception edges prevented region placement and added cleanup retains; precision restores all 56 region values. |
| `internal/oracle/testdata/regions_throw.a` | `124, 95, 53, 143, 31, 29` | `124, 124, 103, 178, 31, 0` | `124, 95, 59, 146, 31, 29` | False checks blocked 29 region values; restore them while keeping executed throw new Error(`budget spent at depth ${depth}`) ownership. |
| `internal/oracle/testdata/move_throw.a` | `42, 42, 22, 50, 9, 0` | `42, 42, 32, 55, 9, 0` | `42, 42, 32, 55, 9, 0` | Executed dynamic-message Error throws on inner/cleanup paths gain nominal constructor ownership. |
| `internal/oracle/testdata/library_map_set_group_by.a` | `115, 115, 117, 201, 36, 0` | `115, 115, 119, 202, 36, 0` | `115, 115, 117, 201, 36, 0` | Executed callback throw new Error('callback failed') gained constructor handoff retains; direct nominal constant construction removes them. |
| `internal/oracle/testdata/normalize_long_marks.a` | `25, 25, 17, 26, 13, 0` | `25, 25, 19, 28, 13, 0` | `25, 25, 17, 26, 13, 0` | Valid normalization and safe repeats gained impure wrapper boundaries; prove forms and preserve primitive borrowing. |
| `internal/oracle/testdata/generic_method_return.a` | `23, 15, 28, 36, 5, 6` | `25, 19, 32, 41, 5, 6` | `25, 19, 32, 41, 5, 6` | Executed caught and uncaught throw new Error(`missing ${key}`) gains nominal ownership plus uncaught formatting/complete cleanup. |
| `internal/oracle/testdata/search_from_sweep.a` | `133, 133, 34, 155, 10, 0` | `133, 133, 35, 156, 10, 0` | `133, 133, 34, 155, 10, 0` | Bounded literal repeats gained guard-call borrowing overhead; prove their length/count and restore primitives. |
| `internal/oracle/testdata/reuse_throw.a` | `26, 26, 24, 41, 8, 0` | `26, 26, 30, 44, 8, 0` | `26, 26, 30, 44, 8, 0` | Executed throw new Error(`refused at ${next.count}`) gains nominal constructor ownership. |
| `internal/oracle/testdata/reuse_narrowed.a` | `7, 5, 4, 7, 5, 0` | `10, 8, 13, 17, 5, 0` | `10, 8, 10, 14, 5, 0` | clear() invalidates box.inner and executes a TypeError; remove only the earlier same-path proven read guards. |
| `internal/oracle/testdata/reuse_lent_global.a` | `10, 10, 7, 14, 7, 0` | `10, 10, 10, 17, 7, 0` | `10, 10, 7, 14, 7, 0` | Initialized box reads and same-path inner reads gained false edges; preserve the original nonmoving global snapshot with NoMove. |
| `internal/oracle/testdata/regexp.a` | `455, 455, 355, 408, 62, 0` | `455, 455, 416, 469, 62, 0` | `455, 455, 355, 408, 62, 0` | Proven null/undefined match and metadata reads acquired TypeError wrappers; preserve stable locals and same-path field presence. |
| `internal/oracle/testdata/sweeps/regexp_methods.a` | `448046, 448046, 131098, 340026, 68, 0` | `448934, 448934, 134252, 342430, 68, 0` | `448046, 448046, 131098, 340026, 68, 0` | Proven match/metadata wrappers lost borrowing and reuse (+888 allocations); recover stable/path presence proofs. |
| `internal/oracle/testdata/regexp_null_narrowed.a` | `4, 4, 7, 5, 3, 0` | `7, 7, 13, 12, 3, 0` | `7, 7, 13, 12, 3, 0` | A call invalidates a null-excluding regex match narrowing; keep the actually executed nominal TypeError. |
| `internal/oracle/testdata/regexp_exec.a` | `179, 179, 81, 159, 18, 0` | `179, 179, 107, 185, 18, 0` | `179, 179, 81, 159, 18, 0` | Proven local non-null exec results gained throwing wrappers; preserve stable narrowing and borrowing. |
| `internal/oracle/testdata/regexp_match.a` | `83, 83, 85, 103, 18, 0` | `83, 83, 94, 112, 18, 0` | `83, 83, 85, 103, 18, 0` | Proven non-null matches and present indices/groups gained throwing wrappers; preserve stable/path presence facts. |
| `internal/oracle/testdata/regexp_unicode.a` | `108, 108, 99, 86, 21, 0` | `108, 108, 108, 95, 21, 0` | `108, 108, 99, 86, 21, 0` | Proven non-null Unicode regex matches gained throwing wrappers; preserve stable narrowing and borrowing. |
| `internal/oracle/testdata/class_inheritance_exceptions.a` | `29, 29, 22, 42, 8, 0` | `29, 29, 30, 46, 8, 0` | `29, 29, 24, 43, 8, 0` | Executed dynamic throw new Error(this.label + '/' + this.detail) gains nominal ownership; constant base/derived errors inline. |
| `internal/oracle/testdata/object_prototype.a` | `142, 142, 317, 440, 22, 0` | `142, 142, 319, 441, 22, 0` | `142, 142, 317, 440, 22, 0` | Non-throwing new Error('message') gained constructor handoff retains; inline its nominal literal prefix to restore exact main counts. |
| `internal/oracle/testdata/walk.a` | `222, 222, 123, 294, 36, 0` | `222, 222, 137, 308, 36, 0` | `222, 222, 123, 294, 36, 0` | Valid radix and bounded format guards gained impure call boundaries; prove literal/counter arguments and preserve successful-path purity. |

Readiness is a closed-program fixed point over all call sites, with intersection at entry and initializer evaluation preceding declaration readiness. Indirect callbacks conservatively target possible closures and methods, so an early indirect call still throws ReferenceError. Local reads whose binding cannot be changed by a callee retain the checker's narrowing. Field facts require no relevant stores, or a same-path presence test surviving all intervening writes. Calls invalidate field facts and all local roots they may transitively assign; handlers, loop back edges and switch paths start conservatively. Unknown numeric stores poison field ranges; loop-counter facts require no capture, no body assignment and safe stepping bounds. Unknown arguments keep the throwing wrapper.

Generated pure primitive wrappers are marked pure only on successful return. Their result has its own count; their failure terminates expression evaluation before any borrowed operand can be used. Ordinary calls remain impure. Proving a global ready does not grant a new ownership move: `Read.NoMove` preserves the original snapshot policy, which avoids an unrelated one-retain decrease in `reuse_lent_global.a`. Constant built-in Error construction with absent cause emits the same nominal class, methods and name/message/cause prefix directly; dynamic messages, causes and user constructors retain their normal paths. Normalization expansion remains refused inside try unless its input is bounded; removing a form guard does not remove that separate failure boundary.

No fixture-name special cases, manually assigned counts or weakened sanitizer checks were used. `internal/native/emit.go`, `internal/lower/lower.go`, `internal/native/native.go` and `internal/oracle/oracle_test.go` are unchanged.

Setup was `bash cloud/setup.sh > /tmp/adamic-counts-setup.log 2>&1`, followed by `source /workspace/adamic-tools/env.sh`. It succeeded with Go 1.27.1, clang 20.1.8 and Node 24.19.0; `nproc` printed `5`. Its timing lines were:

```text
setup: go ready (0s)
setup: clang ready (/workspace/adamic-tools/llvm/bin/clang) (1s)
setup: node ready (1s)
setup: submodules ready (1s)
setup: build cache warm (95s)
setup: done in 95s on 5 processors (cgroup cpu.max: 400000 100000), 17.6 GB
```

`git fetch origin` succeeded, but this checkout's fetch refspec contained only main. An explicit fetch of `codex/error-classes:refs/remotes/origin/codex/error-classes` supplied the requested base before branch creation. Test commands source the tools environment and set `TMPDIR=/tmp/adamic-gate` (mode 1777); all test output goes directly to log files.

The new source mutants use Go overlays and permanent tests, independently of the counts table. Each must compile and fail the named assertion; no compiler diagnostic is accepted as a kill.

| New mutant | Intended check and observed failure |
| --- | --- |
| `all-calls-throw` | `lower.TestMayThrowPrecision` rejects CallMayThrow for two provably nonthrowing calls with “cannot throw”; no counts comparison runs. |
| `all-functions-throw` | The same test rejects inflated MayThrow on `read` and `forward`, independently of counts. |
| `readiness-unproved` | The same test rejects the retained false global-read exception seed. |
| `unknown-store-ignored` | `TestGeneratedGuardFactsIncludeUnknownStores` requires an unknown parameter store to keep the format function throwing. |
| `counter-body-write-ignored` | `TestGeneratedGuardCounterWritesRemainThrowing` requires the body write to invalidate a loop's digits range. |
| `presence-call-mutation-ignored` | Node's `error_checks.a` finishes and catches TypeError; mutant native/JS instead panic at the invalidated read, exit codes differ. |
| `all-calls-pure` | `lent_reads.a` builds strings at runtime and mutates their holders between read and use; ASan reports heap-use-after-free. |

`TestReadinessIncludesEarlyIndirectCalls` additionally holds the uninitialized-global edge through a closure. Existing generated-error mutants independently suppress that callee's MayThrow and closure propagation, and are held by Node comparisons. The existing `mayThrow-overwritten` source mutant also remains caught: changing `found = found || node.Checked` back to assignment produces work after a failed call, disagreeing with Node. Thus precision is checked in both directions. The first all-calls-pure mutant draft left a Go variable unused and was discarded; its corrected mutation builds and reaches ASan. The presence mutant initially matched the wrong diagnostic substring in the runner; the captured failure was already the intended Node/native exit mismatch, and the final runner checks that actual diagnostic.

All 21 existing source mutants were rerun, with these intended checks (full implementations and the original compiled-mutant descriptions remain in `cloud/error-classes-report.md`):

| Existing source mutant | Check |
| --- | --- |
| `mayThrow-overwritten` | Node generated-error output comparison. |
| `derived-error-descriptors` | Four derived-error descriptor refusal assertions. |
| `ready-read-flow-edge` | Node trace notices a call ending in the middle of its flow block. |
| `ready-write-flow-edge` | Node trace notices a write ending in the middle of its flow block. |
| `inherited-constructor-field` | Inherited constructor own-field refusal. |
| `nullable-generic-receiver` | Nullable generic receiver refusal. |
| `stack-read` | Stack read refusal. |
| `stack-write` | Stack write refusal. |
| `cause-cycle` | Cause closes an ownership cycle refusal. |
| `custom-toString` | Override ABI refusal. |
| `optional-generic-field` | Optional generic field refusal. |
| `erased-cause` | Erased cause representation refusal. |
| `nonliteral-options` | Nonliteral ErrorOptions refusal. |
| `normalization-expansion` | Unbounded normalization expansion refusal, including the direct primitive path introduced here. |
| `structural-view` | Direct and nested structural error view refusals. |
| `fake-error` | Nominal ancestry refusal. |
| `spread-error` | Error spread refusal. |
| `inherited-own-field` | Inherited method own-field refusal. |
| `mutable-unknown-alias` | Unknown representation refusal. |
| `unrecorded-initializer` | `fresh.TestEveryWriteIsRecordedAndKnown` detects the missing write site. |
| `mutable-cause` | Mutable cause ownership refusal. |

The full oracle package also reruns the 35 existing compiled mutants: 23 nominal/guard semantics, two reporting conventions and ten generated-error propagation/identity checks. Their checks are Node comparisons after successful native compilation, not counts drift:

| Compiled mutant | What caught it |
| --- | --- |
| Lose MyError's Error ancestry | Node identity/output comparison |
| Lose TypeError's Error ancestry | Node identity/output comparison |
| Give RangeError TypeError's definition identity | Negative instanceof comparison |
| Lose message in super | Node message/output comparison |
| Change default name to MyError | Node name/output comparison |
| Ignore cause | Narrowed cause/output comparison |
| Ignore empty name | Standard toString/output comparison |
| Ignore empty message | Standard toString/output comparison |
| Change separator from colon-space | toString/output comparison |
| Fold instanceof TypeError true | Negative user-error identity comparison |
| Remove toFixed guard | Catch output and exit comparison |
| Remove repeat guard | Catch output and exit comparison |
| Remove normalize-form guard | Catch output and exit comparison |
| Remove precision guard | Catch output and exit comparison |
| Remove radix guard | Catch output and exit comparison |
| Remove undefined-property read guard | Catch identity/output and exit comparison |
| Remove undefined-property write guard | Catch identity/evaluation-order/output and exit comparison |
| Remove exponential guard | Catch output and exit comparison |
| Lose cause through default constructor | Forwarded cause/output comparison |
| Throw RangeError for incompatible prototype receiver | TypeError catch/identity comparison |
| Lose undefined primitive receiver's text default | Exact TypeError message comparison |
| Lose undefined generic name's default | Generic standard-method output comparison |
| Lose undefined generic message's default | Generic standard-method output comparison |
| Change only uncaught status to 1 | Node exit 70 versus native 1, stdout/stderr unchanged |
| Skip buffered stdout flushing | Node before/finally output versus missing native output |
| Restore panic ready checks | A catch cannot observe ReferenceError; exit/output comparison |
| Give ready failures TypeError identity | ReferenceError catch and negative instanceof output |
| Ignore ready reads | Output after an invalid read |
| Ignore ready writes | Missing catch, right-side and later-write output |
| Suppress ready callee MayThrow | Work after a failed call becomes observable |
| Suppress closure throw propagation | Closure catch/output comparison |
| Remove null-property read guard | Nominal TypeError catch and exit comparison |
| Remove intrinsic Number prototype guard | RangeError catch/exit comparison |
| Remove intrinsic String prototype guard | RangeError catch/exit comparison |
| Give Array.with a generic Error | Nominal RangeError catch/identity comparison |


Implementation commit: `154e646` (`Prove generated exception guards precise and restore ownership counts`). The audit is a separate following commit on the same branch; its SHA is shown by `git log -1`. No pull request.

Final validation commands and observed outputs:

| Command | Result | Log |
| --- | --- | --- |
| `gofmt -l cmd internal` | exit 0, no output | `/tmp/adamic-counts-gofmt-final.log` |
| `go vet ./...` | exit 0, no output | `/tmp/adamic-counts-vet-final.log` |
| `git diff --check` | exit 0, no output | terminal |
| `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts` | exit 0; oracle 23.466s, final table regenerated after the gate | `/tmp/adamic-counts-regenerated-final.log` |
| `python3 cloud/error-classes-counts-mutants.py` | exit 0; all seven compiled and failed the intended assertions | `/tmp/adamic-counts-mutants-final.log`, individual logs under `/tmp/adamic-counts-mutants/` |
| `python3 cloud/error-classes-mutants.py --logs /tmp/adamic-error-source-mutants-final` | exit 0; all 21 compiled and failed the intended assertions | `/tmp/adamic-error-source-mutants-final.log`, individual logs under `/tmp/adamic-error-source-mutants-final/` |
| Final uncached affected gate, exact command below | exit 0; all affected packages and all 35 compiled mutants passed | `/tmp/adamic-counts-gate-final.log` |
| Read-only table assertions against both git baselines and final counts | PASS: exactly 63 audit rows, every six-column tuple matches, 34 restored, 29 error residuals, no other main-row changes | terminal |

```sh
ADAMIC_GATE_UNCACHED=1 go test -count=1 -timeout 30m ./internal/ir ./internal/lower ./internal/javascript ./internal/flow ./internal/fresh ./internal/native ./internal/oracle ./stage1/cohere/graphql > /tmp/adamic-counts-gate-final.log 2>&1
```

```text
? ir: no test files
ok lower 33.921s
? javascript: no test files
ok flow 241.697s
ok fresh 109.192s
ok native 369.777s
ok oracle 350.160s
ok graphql 167.506s
exit 0
```

The oracle compared source Node, generated JavaScript on Node, sanitized native and release native; ASan, UBSan and leak checking remained enabled. No failure was waived or omitted from the affected gate. The scalar guards, path facts, constant Error lowering and final ownership metadata were all present for that run. Only explanatory comments/documentation changed afterward. The final count regeneration then reproduced the audited tuples.

Not covered: the complete repository `go test ./...`, unrelated Unicode/property-package sweeps, repository-wide Cohere lint/format and performance benchmarks were not rerun. Repository-wide Go vet was run. This is a conservative precision analysis, not a proof that every supported future program has minimal counts: unknown stores, unknown callbacks, unbounded normalization expansion and unsupported failure kinds retain their guards or existing refusals. The existing error-unit coverage limits in `cloud/error-classes-report.md` remain; this work does not implement stack inspection, AggregateError, arbitrary mutable causes or general object coercion.
