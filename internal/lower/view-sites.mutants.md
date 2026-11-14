# The view site index asks an access again when its name gains a key (#q1hyhzc)

`viewWalkIndexed` (view_sites.go) skips an access not yet found checked while its name's count of CheckedFields
writes is the one it was last judged at. The mutant drops the count from that test, so only a name never judged is
asked: `asked = !judged || seen != written[name]` becomes `asked = !judged; _ = seen`.

| Test | Catch text |
| --- | --- |
| `TestViewSiteAskedAgainWhenItsNameGainsAKey` | view walk comparison (#q1hyhzc): the indexed walk of the view at main.ts:10:18 isn't the whole walk's; CheckedFields: only whole [3823:b], only indexed [] |
| `TestViewWalkIndexedAgreesWithWhole_1`, the shard holding both since #sq4myx1 | the same comparison at casts.ts:10:18, and at internal/oracle/testdata/cast_enum_const.ts:11:53 |

Both run with `viewWalkCompare` on, which runs `viewWalkWhole` beside every indexed walk from one state and panics
where the two differ. Proof jobs: J-hftswrk at a8d9fa0d, J-rf2pdd5 at e08e51d5 and J-w36rk32 at ca3f457d, each 2 of 2
red. The earlier `asked = !judged` form left `seen` unused, so it never compiled. The script now fails a mutant arm
that doesn't build.

`TestViewReceiverCallersAreClassified` holds `judgeViewAccess` as the one caller of `checkedViewReceiverField` for
both walks, and `TestCheckedFieldsAreWrittenThroughMarkCheckedField` holds every CheckedFields write before the last
view to `markCheckedField`, whose count the skip reads.

# The view walk's shards cover every program (#sq4myx1)

`viewWalkShardPrograms` (view_sites_test.go) gives each of `TestViewWalkIndexedAgreesWithWhole_0` to `_3` its share of
the comparison's programs. The mutant leaves out shard 3's last program: `return selected` becomes
`if shard == 3 { selected = selected[:len(selected)-1] }; return selected`.

| Test | Catch text |
| --- | --- |
| `TestViewWalkShardsCoverEveryProgram` | the shards cover 4213 distinct programs, the comparison has 4214 |

Proof job: J-gvskk0c at 665a7a8b, 1 of 1 red, with the coverage test green unmutated ("all 4214 programs fall in exactly
one of 4 shard wrappers") and the four shards at 5.90 s, 5.47 s, 8.19 s and 8.70 s, 211 of 4214 programs lowered with
views between them.
