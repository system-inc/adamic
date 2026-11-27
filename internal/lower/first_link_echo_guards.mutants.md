# Mutants for the first link's Echo guards (#nxb3xh1 heads 4 and 5)

Each is one edit, run on a house box through Loom against the nine pins in first_link_echo_guards_test.go, then put
back. Each turns its own pin red, and only it, save each of the cycle finder's pass-overs dropped, which turns its
live pin red too: findCycles traps the live unit, runs again, and meets the same stop once more, now inside a trapped
unit, so without the pass-over the rerun's stop is a Gap. The captured variable's old text runs on to the line after it,
so it matches once.

| Mutant | File | Old text | New text | Red |
|---|---|---|---|---|
| The held stop's pass-over dropped | first_link_passes.go | `if l.firstLinkSkipsAt(firstLinkStopSite(stop)) {`<br>`continue`<br>`}` | nothing | TestFirstLinkUnlowerableInATrappedUnitIsPassedOver: inner's construct is an Echo, and the judgment is abandoned |
| Only the first construct held | first_link_passes.go | `if l.firstLinkTraps {`<br>`l.firstLinkUnlowerables = append(l.firstLinkUnlowerables, stop)` | `if l.firstLinkTraps && l.firstLinkUnlowerables == nil {`<br>`l.firstLinkUnlowerables = append(l.firstLinkUnlowerables, stop)` | TestFirstLinkUnlowerableInATrappedUnitIsPassedOver: live's construct is never trapped, its row missing |
| A body trap's stub of what it lowered inside dropped | first_link_traps.go | `l.firstLinkStubInside(from, declaration, stub)` | nothing | TestFirstLinkNestedFunctionOfABodyTrapIsStubbed: copy keeps its spread, and checkAccessorSpreads meets it, an Echo |
| The cycle finder's pass-over dropped | loose_slots.go | `if l.firstLinkSkips(l.localNodes[local]) {`<br>`continue`<br>`}`<br>`what := "'" + declared.Name + "', a variable a function value captures` | `if false && l.firstLinkSkips(l.localNodes[local]) {`<br>`continue`<br>`}`<br>`what := "'" + declared.Name + "', a variable a function value captures` | TestFirstLinkCapturedVariableOfATrappedUnitIsPassedOver: trapped's b is an Echo; TestFirstLinkLiveCapturedVariableIsStillJudged: the rerun meets live's c again, a Gap |
| The cycle finder's pass-over widened | loose_slots.go | `if l.firstLinkSkips(l.localNodes[local]) {`<br>`continue`<br>`}`<br>`what := "'" + declared.Name + "', a variable a function value captures` | `if l.firstLinkTraps && len(l.firstLinkTrapped) > 0 {`<br>`continue`<br>`}`<br>`what := "'" + declared.Name + "', a variable a function value captures` | TestFirstLinkLiveCapturedVariableIsStillJudged: live's c goes unjudged once trapped traps |
| stopWhere's helper arm dropped | stop_where.go | `site.function >= 0 && maker != nil {` | `false && maker != nil {` | TestFirstLinkRangeUnionStopInAHelperTrapsItsMaker: the helper's stop names no site, a Gap |
| firstLinkHeldStops' helper sweep dropped | first_link_passes.go | `l.firstLinkStubHelpers()`<br>`if l.unlowerable != nil {` | `if l.unlowerable != nil {` | TestFirstLinkHelperOrphanedByABodyTrapIsStubbed: width's helpers keep their code |
| The initialization's pass-over dropped | loose_slots.go | `if l.firstLinkSkips(made.node) {` | `if false && l.firstLinkSkips(made.node) {` | TestFirstLinkLooseInitOfATrappedUnitIsPassedOver: copy's spread is an Echo; TestFirstLinkLiveLooseInitIsStillJudged: the rerun meets copy's spread again, a Gap |
| The initialization's pass-over widened | loose_slots.go | `if l.firstLinkSkips(made.node) {` | `if l.firstLinkTraps && len(l.firstLinkTrapped) > 0 {` | TestFirstLinkLiveLooseInitIsStillJudged: copy's spread goes unjudged once trapped traps |
