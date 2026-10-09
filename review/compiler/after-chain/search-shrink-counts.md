Audited: all four changed existing rows come from search callback-admission routing in merge 5aa9df2e.
Commits: compare pre-merge 7a151f7f with merged 5aa9df2e; the audit commit records the four measured rows.
Commands and outputs: all four fixtures agree with source Node in release native, sanitized native and JavaScript; focused run PASS 0.701s.
Mutant: TestSearchShrinkMutantSkipCheck catches missing admission checks in both backends by required exit 70 versus observed exit 0, 0.53s.
Not covered: whole packages, the full gate, WASI, or other array-method policy changes.

This audits task #b75jjs3's search-shrink port on compiler/after-chain. Counts below are allocations/frees/retains/releases/peak live/in regions. Every path is under internal/oracle/testdata/callback_widening/; each fixture contains exactly one search call, on line 1.

| Fixture | Exact search call and visited values | Before 7a151f7f | After 5aa9df2e | Node and both backend stdout |
|---|---|---|---|---|
| number-find.a | `[1,2,3].find((x: number \| string): boolean => (typeof x === 'number' ? x : 0) === 2)`; visits 1, then 2 | 8/8/8/14/5/0 | 6/6/2/8/3/0 | `2\n` |
| number-findIndex.a | `[1,2,3].findIndex((x:number\|string):boolean=>x===2)`; visits 1, then 2 | 10/10/8/16/6/0 | 8/8/2/10/4/0 | `1\n` |
| number-findLast.a | `[1,2,3].findLast((x:number\|string):boolean=>x===2)`; visits 3, then 2 | 10/10/8/16/6/0 | 8/8/2/10/4/0 | `2\n` |
| object-find.a | `[{id:1},{id:2},{id:3}].find((x: Item \| number): boolean => (typeof x === 'number' ? x : x.id) === 2)`; visits objects with id 1, then 2 | 9/9/17/22/7/0 | 7/7/9/14/5/0 | `2\n` |

All outputs have exit 0 and empty stderr. The tests also run LeakSanitizer. The arrays never shrink in these four fixtures, so every new callback-admission check sees a present element. Neither union admits undefined; SearchUndefined is false. The new guard remains necessary on removed indices, as the skip-check mutant demonstrates.

## The common cause

The lowering chain already captured the initial array length once in both versions. That length capture does not explain these count reductions.

The relevant search-shrink change is in internal/lower/object.go's arrayVisit: it now calls arraySearchContract before adaptArrayCallback. The original checker signature gives SearchFirst=ir.Union for all four callbacks. Lowering passes this declared search argument representation to adaptArrayCallback instead of the array's raw element representation (ir.Number for the numeric fixtures, ir.Object for object-find).

In internal/lower/callback_widening.go, `changed = changed || given != of` now stays false for these signatures. `if !changed { return callback, kind, nil }` bypasses the old `intrinsic_callback` capture cell and `array_callback_adapter` closure. The callback's unused index and array slots do not require another adapter.

internal/native/emit_arrays.go's arrayVisit converts the present element to SearchFirst immediately before the original callback call. Numeric values still allocate one adamic_box_number per visited value and release that box after the call. Object values use a borrowed adamic_heap pointer cast, which allocates nothing. This conversion belongs to the new admission path: present values receive the declared callback representation, while an admitted missing index would supply the representation's undefined value. It is not a change to native lifetime policy.

## Exact accounting for each row

Generated C was obtained independently at both commits using `go run ./cmd/adamic c <fixture>`, with each command limited by timeout 120. For each of the four before/after pairs, the old C contains an array_callback_adapter function and adamic_cell_new; the new C contains neither, and calls the original closure directly from the search loop.

Removing that cell and adapter closure removes exactly two allocations and two frees per fixture. The two removed live owners account for peak live falling by two. Array, original callback, element values and result lifetimes remain present in both versions; numeric boxing is still once per visit, not removed.

The old wrapper adds two setup retain/release pairs: retaining the captured cell into the adapter and retaining the adapter as the search callback. Its cleanup releases the temporary cell owner and the two adapter owners; the new path instead releases the original callback owner, giving the same net reduction of two release calls.

For number-find, number-findIndex and number-findLast, the removed adapter additionally retains/releases its array parameter and its captured callback once per invocation. Each listed call invokes the callback exactly twice, so it loses `2 + 2*2 = 6` retains and six releases. The numeric element box and the original callback's union-parameter ownership remain in the generated code on both sides. Together with the two removed allocations/frees and peak owners, this accounts for every changed column in each of the three numeric rows.

For object-find, the removed adapter also retains/releases its object parameter, in addition to its array parameter and captured callback. It is invoked exactly twice, so it loses `2 + 3*2 = 8` retains and eight releases. The search loop still retains the object across its callback, transfers the successful held object to the found result, and releases the unsuccessful element. The original callback still owns its union parameter. Together with the two removed allocations/frees and peak owners, this accounts for every changed column in object-find.

Regions stay zero for all four. No unaccounted count change remains.

Representative old numeric adapter body:

```c
adamic_retain(array_argument);
adamic_closure *callback = adamic_retain(self->cells[0]->value.reference);
adamic_heap *boxed = adamic_box_number(number_argument);
callback->code(callback, (adamic_value[]){{.reference = boxed}, /* index, array */});
adamic_release(boxed);
adamic_release(callback);
adamic_release(array_argument);
```

Representative new numeric loop body:

```c
bool present = index < array->length;
/* If not present and the declared callback disallows undefined, panic. */
adamic_value argument = {.reference = present ? adamic_box_number(element.number) : NULL};
callback->code(callback, (adamic_value[]){argument, /* index, array */});
adamic_release(argument.reference);
```

For object-find the old adapter additionally retains/releases object_argument; the new argument is `present ? (adamic_heap *)element.reference : NULL`, with no fresh box. These excerpts normalize generated temporary names; exact generated files and unified diffs are retained locally under review/compiler/after-chain/search-counts-before/, search-counts-after/ and search-counts-<fixture>.patch.

## Validation

```sh
source /workspace/adamic-tools/env.sh
timeout 120 go test ./internal/oracle -run '^TestCallbackWidening(NumberFind|NumberFindIndex|NumberFindLast|ObjectFind)$|^TestSearchShrinkMutantSkipCheck$' -count=1 -v -timeout 90s > review/compiler/after-chain/search-counts-output-check.log 2>&1
```

PASS 0.701s. Leaf seconds: number-find 0.64, number-findIndex 0.63, number-findLast 0.62, object-find 0.67, skip-check mutant 0.53. These existing tests compare source Node with release native, ASan/UBSan native and backend JavaScript byte for byte, then check leaks. The skip-check mutant compiles and runs without stderr, returning 0 in both backends instead of the required terminal 70; the intended exit-code assertion catches both.

The prior merged search controls, fourteen fixtures and old-stop mutant also passed (11.392s); prior backend-stop and self-comparison regression checks passed (6.540s). Those observations are retained in search-merged.log and member-final.log. No compiler source was edited during this audit.

Canonical regeneration (`timeout 150 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s -args -update-counts`) passed in 65.252s. A row-key comparison against 5aa9df2e confirms that numeric changes are exactly the four rows audited above: no row was added or removed, and every other value is unchanged. The generator also moves the three unchanged self-comparison rows after the fourteen search-shrink rows, matching fixture registration order. This is a positional repair to the search port's hand-merged table. Exact diff: search-counts-final.diff.

The full recorded-counts check passed in 28.065s: `timeout 120 go test ./internal/oracle -run '^TestCountsAreRecorded$' -count=1 -timeout 90s`, output in search-counts-check.log. The old recorded values had failed the same check on the merged compiler, reporting exactly these four numeric changes; updating the justified rows and canonical order makes it pass.
