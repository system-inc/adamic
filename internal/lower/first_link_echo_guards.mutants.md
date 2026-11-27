# Mutants for the first link's Echo guards (#nxb3xh1 head 4)

Each is one edit, run on a house box through Loom against the six pins in first_link_echo_guards_test.go, then put
back. Each turns its own pin red, and only it, save the cycle finder's pass-over dropped, which turns the live pin
red too: findCycles traps live at c, runs again, and meets c once more, now inside a trapped unit, so without the
pass-over the rerun's stop is a Gap.

| Mutant | File | Old text | New text | Red |
|---|---|---|---|---|
| The held stop's pass-over dropped | first_link_passes.go | `if l.firstLinkSkipsAt(firstLinkStopSite(stop)) {`<br>`continue`<br>`}` | nothing | TestFirstLinkUnlowerableInATrappedUnitIsPassedOver: inner's construct is an Echo, and the judgment is abandoned |
| Only the first construct held | first_link_passes.go | `if l.firstLinkTraps {`<br>`l.firstLinkUnlowerables = append(l.firstLinkUnlowerables, stop)` | `if l.firstLinkTraps && l.firstLinkUnlowerables == nil {`<br>`l.firstLinkUnlowerables = append(l.firstLinkUnlowerables, stop)` | TestFirstLinkUnlowerableInATrappedUnitIsPassedOver: live's construct is never trapped, its row missing |
| A body trap's stub of what it lowered inside dropped | first_link_traps.go | `l.firstLinkStubInside(from, declaration, stub)` | nothing | TestFirstLinkNestedFunctionOfABodyTrapIsStubbed: copy keeps its spread, and checkAccessorSpreads meets it, an Echo |
| The cycle finder's pass-over dropped | loose_slots.go | `if l.firstLinkSkips(l.localNodes[local]) {` | `if false && l.firstLinkSkips(l.localNodes[local]) {` | TestFirstLinkCapturedVariableOfATrappedUnitIsPassedOver: trapped's b is an Echo; TestFirstLinkLiveCapturedVariableIsStillJudged: the rerun meets live's c again, a Gap |
| The cycle finder's pass-over widened | loose_slots.go | `if l.firstLinkSkips(l.localNodes[local]) {` | `if l.firstLinkTraps && len(l.firstLinkTrapped) > 0 {` | TestFirstLinkLiveCapturedVariableIsStillJudged: live's c goes unjudged once trapped traps |
| stopWhere's helper arm dropped | stop_where.go | `site.function >= 0 && maker != nil {` | `false && maker != nil {` | TestFirstLinkRangeUnionStopInAHelperTrapsItsMaker: the helper's stop names no site, a Gap |
| firstLinkHeldStops' helper sweep dropped | first_link_passes.go | `l.firstLinkStubHelpers()`<br>`if l.unlowerable != nil {` | `if l.unlowerable != nil {` | TestFirstLinkHelperOrphanedByABodyTrapIsStubbed: width's helpers keep their code |
