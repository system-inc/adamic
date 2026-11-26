# Mutants for a judging pass's stop where (#nxb3xh1 head 2)

Each is one edit, run on a house box through Loom, then put back.

| Mutant | File | Old text | New text | Red |
|---|---|---|---|---|
| stopWhere names the file alone | stop_where.go | `func (l *lowering) stopWhere(site stopSite) string {` | `func (l *lowering) stopWhere(site stopSite) string {`<br>`return l.result.Source` | TestFirstLinkAccessorSpreadStopTrapsItsUnit: the spread's stop is a Gap again, checkAccessorSpreads' judgment abandoned, and without the option the stop names the file alone; TestFirstLinkRangeUnionStopTrapsItsUnit: the narrow in width is a Gap, checkRangeUnionReads' judgment abandoned |
| checkRangeUnionReads names the file alone | range_unions.go | `refused = &NotYet{Where: l.stopWhere(site), What: road` | `refused = &NotYet{Where: l.result.Source, What: road` | TestJudgingPassStopsNameTheirWhereThroughStopWhere: range_unions.go checkRangeUnionReads has a stop whose Where isn't stopWhere's, and no allowance |
