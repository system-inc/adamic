# The view site index asks an access again when its name gains a key (#q1hyhzc)

`viewWalkIndexed` (view_sites.go) skips an access not yet found checked while its name's count of CheckedFields
writes is the one it was last judged at. The mutant drops the count from that test, so only a name never judged is
asked: `asked = !judged || seen != written[name]` becomes `asked = !judged; _ = seen`.

| Test | Catch text |
| --- | --- |
| `TestViewSiteAskedAgainWhenItsNameGainsAKey` | view walk comparison (#q1hyhzc): the indexed walk of the view at main.ts:10:18 isn't the whole walk's; CheckedFields: only whole [3823:b], only indexed [] |
| `TestViewWalkIndexedAgreesWithWhole` | the same comparison at casts.ts:10:18, and at internal/oracle/testdata/cast_enum_const.ts:11:53 |

Both run with `viewWalkCompare` on, which runs `viewWalkWhole` beside every indexed walk from one state and panics
where the two differ. Proof jobs: J-hftswrk at a8d9fa0d, J-rf2pdd5 at e08e51d5 and J-w36rk32 at ca3f457d, each 2 of 2
red. The earlier `asked = !judged` form left `seen` unused, so it never compiled. The script now fails a mutant arm
that doesn't build.

`TestViewReceiverCallersAreClassified` holds `judgeViewAccess` as the one caller of `checkedViewReceiverField` for
both walks, and `TestCheckedFieldsAreWrittenThroughMarkCheckedField` holds every CheckedFields write before the last
view to `markCheckedField`, whose count the skip reads.
