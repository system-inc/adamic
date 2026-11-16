# The placeholder slot cache's comparison runs in four shards (#58r9613)

`TestPlaceholderSlotsCachedAgreeWithAWalk` ran 12.97 s on a quiet box and 61.6 s on a loaded line box, past the 60 s
unit budget, and turned chg_pfg0dc5g red, which touched only internal/worker. The comparison is now
`runPlaceholderSlotsShard` (placeholder_cache_test.go), and `TestPlaceholderSlotsCachedAgreeWithAWalk_0` to `_3`
(placeholder_slot_shards_test.go) each call it with the programs `nameShard` puts in their shard, as the view walk's
comparison was split (#sq4myx1).

## The cached answer stored empty (#6xt6p1f's comparison mutant)

In `notePlaceholderSlots` (placeholder.go), `sources: maps.Clone(l.result.PlaceholderSources)` becomes
`sources: ir.SortedMap[string, string]{}`, so every lowering that takes the cache's answer has no sources, and
`placeholderSlotCompare`'s fresh walk panics where it finds them.

| Test | Catch text |
| --- | --- |
| `TestPlaceholderSlotsCachedAgreeWithAWalk_0` to `_3`, each | placeholder slot comparison (#6xt6p1f): the cached answer for this checker and program isn't a fresh walk's; PlaceholderSources: only cached [], only fresh [...], at non_null_uninitialized_optional.ts and non_null_uninitialized_capture.ts among others |

## The shards cover every program

`placeholderSlotsShardPrograms` gives each shard its share. The mutant leaves out shard 3's last program:
`return selected` becomes `if shard == 3 { selected = selected[:len(selected)-1] }; return selected`.

| Test | Catch text |
| --- | --- |
| `TestPlaceholderSlotsShardsCoverEveryProgram` | the shards cover 4316 distinct programs, the comparison has 4317; yield_result_unwrap_refused/recursive.ts falls in 0 shards, want exactly one |

#6xt6p1f's three other mutants (the cache never consulted, the answer held by reference, the guard dropped) go red in
`TestPlaceholderSlotsAreWalkedOncePerCheckerAndProgram`, `TestPlaceholderSlotsAreClonedInAndOut` and
`TestPlaceholderSlotsUnderAMapperOrAKeptSlotLeaveTheCacheAlone`, which this change doesn't touch.

Proof job: J-5jsjbth on Golem at d5dea225 (proof commit d24d551d), exit 0 in 5m 30s: gofmt and vet clean; unmutated, the four shards
passed in about 20 to 23 s each, with 10 of 1076, 11 of 1100, 13 of 1053 and 13 of 1088 programs taking a cached answer
with placeholders, and the coverage test read "all 4317 programs fall in exactly one of 4 shard wrappers"; the empty
answer went 4 of 4 red and the coverage mutant 1 of 1.
