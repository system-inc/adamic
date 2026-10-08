# Placeholder lowering for step 09

Literal `undefined!` and `null!` slot initializers and resets retain their distinct observable values. Existing union storage, captured cells, definite-assignment analysis and readiness flow carry the unset state. A field's observable storage is distinguished from its readiness for a typed use: copying a placeholder field does not assert that it was written.

Presence observations and transfers to optional or nullable slots, placeholder slots and inferred save locals preserve the nullish arm. Every later use as T is proven or checked: property receiver, call target, nonoptional argument, return, ordinary store and container element. A saved local retains its snapshot; writing the source does not prove the saved value ready. A field write invalidates same-name facts for aliases before restoring only the written receiver's fact. The conservative assumption is that an unproven inferred transfer stays unset-capable and every eventual typed use is checked.

A checked use exits 70 with `placeholder '<slot>' is unset at <use> via <path>`. `--explain-checks` records both proven and checked placeholder boundaries, including present-arm and structural-field representation checks. Main's ordinary nonliteral assertions keep their source-specific checked .ts diagnostic and their .a refusal. Weak placeholder slots, accessor placeholder initializers and literal-reset expressions whose result escapes remain NotYet rather than relying on an unimplemented boundary.

The scanner fixture reduces scanner.ts:4097's written-before-use placeholder. Local, field, save/restore, copy, spread, strict equality, loose nullish equality, truthiness, typeof, coalesce, nullable transfer and JSON temporary-literal fixtures compare source Node with both backends. Effectful nullish receivers are evaluated once. Seven negatives pin typed-use diagnostics for direct reads, save-local leaks, assignment results, returns and alias resets. The nonliteral assertion witness is committed as .a and tested through a temporary .ts input to preserve the existing extension policy.

## Evidence and limits

The previously published read-first table covers 19 slots and 344,015 before-write observations in the 301-project measurement. Those observations were undefined: presence tests, transfers to presence tests, save/restore and copies. They motivated the amended rule; this implementation did not rerun native tsc over all 301 projects. The scanner reduction is an independent semantic fixture, not a claim of a complete scanner build.

Merged origin/main 031a1259b into the delivery branch. The merge retains namespace state initialization alongside placeholder readiness resets, enumeration-index admission alongside the literal-slot exception, and main's ordinary assertion policy. Counts were regenerated. The delivery branch contains no other worker's unlanded changes.

Two stage3 stage0 records move past eager literal assertions: assertions/06_memoize_clear.a and host/14_getCurrentDirectory.a now reach the existing captured callback-cycle refusal, at 5:28 and 11:28 respectively. Their sources and Node records are unchanged. No unrelated refusal was removed.

## Independently exercised mutants

| Change | Catcher |
| --- | --- |
| Drop each of seven typed-flow guards | The corresponding negative changes from exit 70 to Node's exit 0; both backends agree, with leak checks on completed mutant runs. |
| Collapse null storage into undefined | Strict null/undefined observations disagree with Node, exit 0 in both backends. |
| Treat observable placeholder storage as physically uninitialized during spread | Copy fixture stops in both backends where Node completes. |
| Erase readiness without proof; erase loop/capture/exception checks; initialize to zero; replace unset with represented zero | Pinned readiness witnesses change output or checked outcome. Numeric fallback mutants use a valid boxed representation. |
| Drop ordinary nonliteral assertion checks; weaken Weak assertion diagnostics | Existing migrated assertion and Weak witnesses disagree with their pinned outcomes. |
| Bypass save-local flow lowering | Saved undefined/null reaches a nonoptional parameter and prints `missing`/`present` instead of exit 70. |
| Evaluate nullish receiver twice | Node call count 3 becomes 5 in both backends, with otherwise matching exit 0. |
| Keep stale alias readiness | Reset through an alias reaches a typed local and prints `missing` instead of exit 70. |
| Treat a JSON placeholder field as an ordinary assertion | Serialization exits 70 instead of Node's null field and omitted undefined field. |
| Admit a Weak placeholder without checking its full declared type | The retained NotYet test instead lowers successfully. |

Compiler mutation runs use temporary Go overlays; no mutant is committed in delivery sources. Sanitizer failures are not used as proof that a mutant was caught.

## Local checks

All output is written to log files. The Node oracle builds native with ASan/UBSan, compares stdout, stderr and exit status with source Node and the JavaScript backend, then checks leaks on completed runs. Negative checked stops pin exit 70 separately from Node's permitted unset observation.

- `go test ./internal/lower -run 'Placeholder|Readiness|NonNull|Uninitialized' -count=1`
- `go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/(placeholder_nonnull|non_null_)|TestPlaceholder|Test.*Readiness|Test.*ViewFields|Test.*NonNull|TestLiteralDefaultPlaceholder' -count=1 -v`
- Each compiler-overlay mutant runs its corresponding filtered oracle or lower test and must fail with the intended outcome.
- Build cmd/adamic and run the fast gate's `Gate.aCheck` over every changed .a, including Go-package witnesses. The nonliteral witness's exact refusal header is asserted by its test; all other added .a inputs check cleanly.
- `npm ci` in stage3/api; `go test ./stage3/fixtures`; repeat uncached with only the host platform guard lifted through an uncommitted Go overlay.
- `go test ./stage1/... -run 'Gap|Gaps|Probes'` compared with main: 118 test outcomes and 35 package outcomes on each tree, zero changes.
- `go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts`, then the same check without update.
- Focused checked-sites CLI test, SSA contract test and go vet on affected packages.

Linux counts are recorded below. Counts at a deliberate panic describe the stopping point, not a completed leak check. No full gate or macOS gate was run locally. Setup reported 47.951 seconds, including a 47.615-second build; nproc reported 5. Environment: Go 1.27.1, Node 24.19.0, clang 20.

## Changed counts

Columns in each tuple are allocations, frees, retains, releases, peak live and values released in regions. Every row differing from origin/main is listed below.

| Fixture | Before | After | Reason |
| --- | --- | --- | --- |
| internal/oracle/testdata/placeholder_nonnull_scanner.a | new | 4, 4, 9, 8, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_local.a | new | 12, 12, 11, 24, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_field.a | new | 9, 9, 15, 23, 5, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_observe.a | new | 27, 27, 48, 79, 9, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_equality.a | new | 4, 4, 10, 16, 2, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_loose.a | new | 10, 10, 23, 35, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_truthiness.a | new | 1, 1, 7, 10, 1, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_typeof.a | new | 3, 3, 9, 11, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_coalesce.a | new | 8, 8, 15, 27, 4, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_optional.a | new | 6, 6, 16, 25, 3, 0 | New fixture; counted successful placeholder observations and ownership. |
| internal/oracle/testdata/placeholder_nonnull_null_copy.a | new | 26, 26, 102, 121, 7, 0 | New fixture; copy, save/restore and spread preserve nullish payloads and release their storage. |
| internal/oracle/testdata/placeholder_nonnull_null_json.a | new | 5, 5, 0, 6, 3, 0 | New fixture; exact null/undefined JSON temporary fields use existing schema arms and release their objects and output strings. |
| internal/oracle/testdata/placeholder_nonnull_before_use.a | new | 0, 0, 1, 1, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_saved_leak.a | new | 1, 0, 3, 4, 1, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_before_use.a | new | 0, 0, 2, 1, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_null_saved_leak.a | new | 1, 0, 4, 4, 1, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_assignment_result.a | new | 0, 0, 2, 3, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_return_assignment.a | new | 0, 0, 2, 3, 0, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/placeholder_nonnull_alias_reset.a | new | 2, 1, 6, 7, 2, 0 | New negative fixture; counts stop at the pinned typed-use check. |
| internal/oracle/testdata/non_null_initialized.ts | 0, 0, 0, 0, 0, 0 | 55, 55, 50, 109, 18, 0 | Completing writes now execute; scalar placeholder slots box their values and captured saves own their cells. |
| internal/oracle/testdata/non_null_uninitialized_field.ts | 1, 0, 0, 0, 1, 0 | 1, 0, 3, 2, 1, 0 | Allocate the object, transfer its nullish field and stop at the typed property use. |
| internal/oracle/testdata/non_null_uninitialized_capture.ts | 0, 0, 0, 0, 0, 0 | 2, 0, 2, 0, 2, 0 | Create the captured cell and closure, then stop when the callback uses the unset value as T. |
| internal/oracle/testdata/non_null_uninitialized_exception.ts | 0, 0, 0, 0, 0, 0 | 1, 1, 3, 2, 1, 0 | Execute catch handling and its string output before the later typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_loop.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 1, 0, 0, 0 | Carry unset into the loop exit and stop at the typed use, after its reference transfer. |
| internal/oracle/testdata/non_null_uninitialized_const.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 2, 1, 0, 0 | Save the nullish reference then stop at the property receiver. |
| internal/oracle/testdata/non_null_static_initialized.ts | 1, 0, 0, 1, 1, 0 | 9, 9, 12, 22, 6, 0 | Complete the static writes; scalar boxes and strings are released on the successful path. |
| internal/oracle/testdata/non_null_static_uninitialized.ts | 1, 0, 0, 1, 1, 0 | 1, 0, 3, 2, 1, 0 | Allocate static storage and transfer its nullish value before the property-use check. |
| internal/oracle/testdata/non_null_uninitialized_optional.ts | 1, 0, 0, 0, 1, 0 | 3, 3, 2, 6, 3, 0 | Observe unset through optional handling and complete rather than stopping at initialization. |
| internal/oracle/testdata/non_null_uninitialized_spread.ts | 1, 0, 0, 0, 1, 0 | 2, 0, 5, 5, 2, 0 | Copy the actual nullish field, allocate the copy and stop only at its later typed use. |
| internal/oracle/testdata/non_null_literal_assignment.ts | 0, 0, 0, 0, 0, 0 | 1, 1, 2, 2, 1, 0 | Execute the reset, later completing write and observation; own and release the resulting value. |
| internal/oracle/testdata/non_null_uninitialized_iteration.ts | 1, 0, 0, 1, 1, 0 | 6, 1, 10, 8, 6, 0 | Allocate captured iteration cells and closures before the callback typed-use stop. |
| internal/oracle/testdata/non_null_uninitialized_interface.ts | 1, 0, 1, 0, 1, 0 | 1, 0, 2, 1, 1, 0 | Read the nullish field through the checked structural view before stopping at use as T. |
| internal/oracle/testdata/non_null_uninitialized_map_entry.ts | 0, 0, 0, 0, 0, 0 | 0, 0, 1, 1, 0, 0 | Transfer the nullish reference before the map argument boundary stops. |
| stage3/interface-downcasts/default-staged.ts | 0, 0, 0, 0, 0, 0 | 2, 2, 6, 7, 2, 0 | Complete staged placeholder field writes and the structural-view reads. |
| stage3/interface-downcasts/default-boxed-write.ts | 0, 0, 0, 0, 0, 0 | 5, 5, 5, 12, 4, 0 | Complete scalar field writes using union boxes and release their values. |
| stage3/interface-downcasts/default-read-before-set.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 4, 2, 1, 0 | Allocate the placeholder object and check its typed structural field read. |
| stage3/interface-downcasts/readiness-identifier.ts | 0, 0, 0, 0, 0, 0 | 2, 2, 6, 8, 2, 0 | Complete reference field assignment and checked structural-view transfer. |
| stage3/interface-downcasts/readiness-identifier-uninitialized.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 4, 3, 1, 0 | Transfer the unset reference through a structural view and stop at use as T. |
| stage3/interface-downcasts/readiness-number.ts | 0, 0, 0, 0, 0, 0 | 5, 5, 5, 10, 5, 0 | Box completed numeric placeholder fields and release the boxes after structural reads. |
| stage3/interface-downcasts/readiness-number-uninitialized.ts | 0, 0, 0, 0, 0, 0 | 1, 0, 5, 3, 1, 0 | Preserve the numeric field nullish payload and stop at its typed structural read. |

40 changed rows: 19 new fixture rows and 21 existing rows. Other Linux rows are unchanged.
