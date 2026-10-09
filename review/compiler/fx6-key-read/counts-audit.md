# Counts audit

Attribution correction: `577c3271` (values) changed three existing rows; `218592bb` (entries) changed seven. This audits all seven entries rows against its parent `577c3271`. The values deltas are also recorded below. No unexplained row remains.

Columns are allocations / frees / retains / releases / peak / regions. All paths below are under `internal/oracle/testdata/` and end in `.a`.

| Row | Before | After | Cause |
| --- | --- | --- | --- |
| entries_checked_fit | 54 / 54 / 23 / 45 / 12 / 0 | 54 / 54 / 25 / 47 / 12 / 0 | `enumerate` and `words` each route one entries operand through `checked_view_members`. Its borrowed receiver is retained for the owned return and released by the enumeration caller: two pairs. Field loads remain borrowed. |
| entries_checked_misfit | 7 / 1 / 3 / 2 / 6 / 0 | 7 / 1 / 4 / 2 / 6 / 0 | The helper checks `z`, then retains the receiver for its return. Existing enumeration rejects hidden string `hidden` against number and exits 70 before caller cleanup, so the added retain has no release. |
| entries_checked_boolean | 18 / 18 / 6 / 19 / 7 / 0 | 18 / 18 / 7 / 20 / 7 / 0 | One entries helper reads `yes`, retains its borrowed receiver for return, and the caller releases that owned result. Boolean checks need no reference ownership. |
| entries_checked_literal | 6 / 0 / 5 / 1 / 6 / 0 | 6 / 0 / 6 / 1 / 6 / 0 | The helper reads the proven literal `value`, then retains its receiver for return. Existing hidden `bad` membership failure exits 70 before the enumeration caller releases it. |
| entries_checked_boxed | 40 / 40 / 14 / 38 / 11 / 0 | 40 / 40 / 16 / 40 / 11 / 0 | Entries in `show` and `enumEntries` now hold helper returns, one pair each. The declared number and open numeric enum reads add no boxes; hidden values still use existing enumeration conversion. |
| entries_record_alias | 65 / 65 / 89 / 118 / 28 / 0 | 65 / 65 / 90 / 119 / 28 / 0 | Three entries helper returns replace three pre-existing caller retain/release pairs on `view`, `wordView`, and `flagView`, net zero for receiver ownership. The dynamic record read of `wordView.label` uses `adamic_object_view`, retains its string reference while evaluating the checked read, then releases it: the single net pair. `visible` and `enabled` are scalars. |
| notyet_library_object_entries_const | 40 / 40 / 36 / 35 / 23 / 0 | 40 / 40 / 38 / 37 / 23 / 0 | Both fresh literal entries operands now pass through helpers. Each helper retains its borrowed literal for the returned reference; the caller releases that extra reference as well as its original literal. Their proven string field reads borrow slots and add no pairs. |

Values phase, `785b665b` to `577c3271`: `entries_checked_fit` 54/54/21/43/12/0 to 54/54/23/45/12/0 (values in `enumerate` and `words`); `entries_checked_boolean` 18/18/5/18/7/0 to 18/18/6/19/7/0 (values in `show`); `entries_checked_boxed` 40/40/12/36/11/0 to 40/40/14/38/11/0 (values in `show` and `enumEntries`). These are the same borrowed-receiver to owned-helper-return pairs, released after enumeration.

## Evidence and output

`counts-audit/results.json` records a fresh before/after counter replay on c95e58a4, with only entries routing temporarily removed in the before compiler. It exactly reproduces every historical count delta. Source was restored before tests. The paired `.c.txt` files show the emitted operations; they are evidence, not compilable repository sources.

Before/after stdout, exits and panic messages are unchanged for every row. The five successful rows agree with source Node in native sanitized, native release, and JavaScript backends. The two existing misfit/literal fixtures are explicitly ruled checked stops: both backends exit 70 with the established hidden-field diagnostic; source Node continues. They must not be described as ordinary Node-output equality. `TestEntriesProvenance` holds them to Node and pins that deliberate divergence.

Validation: `timeout 180 go test ./internal/oracle -run 'TestEntriesProvenance/(entries_checked_(fit|misfit|boolean|literal|boxed)|entries_record_alias)$|TestObjectEntriesConst' -count=1 -v -timeout 90s` passed in 1.317s. The constant fixture has a separate Node agreement and existing field-order mutant run in `counts-audit-const.log`. No compiler or test implementation changed; counts.md did not change.

Setup: GOPROXY=https://proxy.golang.org|direct; bounded cloud/setup.sh completed in 41.760s. Go ready 0.032s, Node 0.034s, submodules 0.095s, markdown 0.105s, clang 0.185s, build 41.313s, cache 41.578s. nproc=5; CPU quota=4. No new mutant was required for this evidence-only audit; existing provenance mutants passed.
