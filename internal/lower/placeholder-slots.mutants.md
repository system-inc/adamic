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
| `TestPlaceholderSlotsCachedAgreeWithAWalk_0` to `_3`, each | PENDING |

## The shards cover every program

`placeholderSlotsShardPrograms` gives each shard its share. The mutant leaves out shard 3's last program:
`return selected` becomes `if shard == 3 { selected = selected[:len(selected)-1] }; return selected`.

| Test | Catch text |
| --- | --- |
| `TestPlaceholderSlotsShardsCoverEveryProgram` | PENDING |

#6xt6p1f's three other mutants (the cache never consulted, the answer held by reference, the guard dropped) go red in
`TestPlaceholderSlotsAreWalkedOncePerCheckerAndProgram`, `TestPlaceholderSlotsAreClonedInAndOut` and
`TestPlaceholderSlotsUnderAMapperOrAKeptSlotLeaveTheCacheAlone`, which this change doesn't touch.

Proof job: PENDING.
