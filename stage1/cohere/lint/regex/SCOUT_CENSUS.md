# Scout census

Cohere f5d1934a has 89 standard regexp calls: 88 MustCompile (82 fixed plus six constructed) and one constructed Compile. All 89 are checked against the pinned Go AST. The stale 107-row translation baseline was regenerated, while retaining all 34 JavaScript-semantics esregexp sites in option-sites.json and combining both engines in all-patterns.json.

The option engine migration is cohere c2e39b75 (#7mztrdd), explicitly named by the fleet; option source strings now remain JavaScript, with their upstream flags. scout-census-review.json lists every removed/added Go row by full source expression. The live esregexp AST census is independently checked, so option families are retained rather than dropped when they leave RE2. Earlier 715ba94f..7945d102 provenance remains in CENSUS.md and census-review.json.

Combined table: 123 rows, plain 80, dynamic 39, (?i) 2, (?s) 1, (?m) 1. The 89 Go shapes all pass their concrete string fixtures on source Node, emitted JavaScript and runtime sanitized native; no selected row is outside the library's shapes or one of its five V8-divergence refusals. Infinite future option strings are not claimed to have a finite fixture census.
