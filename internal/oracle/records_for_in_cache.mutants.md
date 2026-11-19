# A record's cached for...in key list: mutants (#1hp4ww2)

Each mutant is one edit, applied alone, then reverted. Not run yet: written on a machine that doesn't build Go
or C, so "goes red" below is what the code says, not what a run printed. The runtime's key reads every runtime
source, so an edited map.c or record.c is a fresh build, not a cached result.

## M1. A new key doesn't let go of the list

File: internal/native_runtime/c/map.c, `adamic_map_set`.

Old:

```c
	adamic_map_forget_key_list(map);
	if (map->used == map->capacity) {
```

New:

```c
	if (map->used == map->capacity) {
```

Run `go test ./internal/oracle -run '^TestRecordForInCachedKeys$/^order.a$' -count=1 -v`.

Expected: red on the native leg with `stdout differs`. The second loop caches `b,a`; `r['c'] = 3` adds a key
without letting the list go, so the third line prints `added: b,a` where Node prints `added: b,a,c`. The
JavaScript leg stays green: it never caches.

## M2. A hidden key doesn't let go of the list

File: internal/native_runtime/c/record.c, `adamic_record_define_hidden`.

Old:

```c
		adamic_map_forget_key_list(map);
		adamic_map_entry *entry = entry_of(slot);
```

New:

```c
		adamic_map_entry *entry = entry_of(slot);
```

Run `go test ./internal/oracle -run '^TestRecordForInCachedKeys$/^hidden.a$' -count=1 -v`.

Expected: red on the native leg with `stdout differs`. The cached `a,b,c` survives the define, `b` is still
an own key, so the loop's hasOwn guard passes it, and the third line prints `b hidden: a,b,c` where Node
prints `b hidden: a,c`.

## M3. A shared record fills its table's list

File: internal/native_runtime/c/record.c, `adamic_record_for_in_keys`.

Old:

```c
	if (adamic_is_shared(&record->heap) || adamic_is_shared(&map->heap) || adamic_graph_is(record) || adamic_graph_is(map)) {
```

New:

```c
	if (adamic_graph_is(record) || adamic_graph_is(map)) {
```

Automated: `go test ./internal/native -run '^TestParallelMutant_for_in_keys_shared$' -count=1 -v`
(parallelCheckMutant's for_in_keys_shared row). adamic_share let the list go as it marked the table, so both
workers find key_list NULL, write it, and retain an array one of them made with plain counts: TSan reports
`WARNING: ThreadSanitizer: data race`. Where TSan isn't available the row skips; there the harness's check
after the join still aborts on `a shared record's loops share a list` (TestParallelForInKeys, any build).

## M4. A deleted key doesn't let go of the list

File: internal/native_runtime/c/map.c, `adamic_map_delete`.

Old:

```c
	adamic_map_forget_key_list(map);
	adamic_map_entry *entry = &map->entries[index];
```

New:

```c
	adamic_map_entry *entry = &map->entries[index];
```

Run `go test ./internal/oracle -run '^TestCountsAreRecorded_[0-9]+$/internal/oracle/testdata/records_for_in_cache/delete' -count=1 -v`.

Run on a sealed box (J-h0y3200): red, the row `36 | 36 | 102 | 88 | 11 | 0` measured as `33 | 33 | 98 | 93 | 11 | 0`,
three allocations fewer, with TestRecordForInCachedKeys green under the same mutant. No output can see this mutant: a list still
naming a deleted key is exactly what the loop's hasOwn guard skips, so delete.a prints Node's lines either
way (TestRecordForInCachedKeys stays green). Its counted row is the witness. delete.a deletes three keys the
record holds, each followed by a loop that builds the list again, so the mutant reuses the stale list three
times; its `delete r['absent']` lets go of nothing either way, and the loop after it reuses the list. What the
survivor would cost is retention, never a wrong answer or a use after free: the stale list retains every key
it names, so a deleted key's string stays alive until the next key added or the record's end (Memory's read
on #1hp4ww2).
