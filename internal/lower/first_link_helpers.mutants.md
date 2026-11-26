# Mutants for a first-link helper's stub (#nxb3xh1 head 3)

Each is one edit to first_link_helpers.go, run on a house box through Loom, then put back.

| Mutant | Old text | New text | Red |
|---|---|---|---|
| The helper stubbing loop dropped | `if trappedWhole(index) {` | `if false && trappedWhole(index) {` | TestFirstLinkHelperIsStubbedWithItsUnit: width's helpers keep their reads, the rerun meets the site again, a Gap, and checkHelperReads' judgment is abandoned |
| The check of every owner dropped | `for _, use := range uses[index] {` | `for _, use := range uses[index][:0] {` | TestFirstLinkHelperWithALiveOwnerIsAGap: the helper other uses too is stubbed with width, so the pass judges on with no Gap |
