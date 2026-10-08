Read boolean option fields through a finite union of field names, toward roadmap step 30.
Implementation commit: 20c481367c76bf1d50257b3b6427a6fd8542bfc8; branch codex/hidden-08-never-array.
Focused Node/native/JavaScript oracles with sanitizers, unsupported-read checks, counts refresh and regional byte-mask audits pass.
Wrong-field and repeated-key mutants both fail from stdout disagreement in both backends.
Other value families, arbitrary string keys, accessors, methods and prototype-member names remain unsupported; no full gate was run.

Ownership was checked in stage3/hidden-briefs-2/TABLE.md on codex/hidden-wave-2 at e2c6e8df9064cc935dfc0057e8286d98054da338. It assigns no owner to this finite field-key read. Its stricter-compiler-options exclusion concerns checker settings; this change lowers an ordinary object read, without changing those settings. No other worker branch was merged.

The exact replay on b3b3276958aa125115bd911f8d492aaabd0768b7 reproduced utilities.ts:9370:12, NotYet, an ElementAccessExpression, exit 0. The failed return statement is [370572,370684). After lowering, the selected function has no findings; the historical selector exits 1 because the old finding is absent. The full-file measurement confirms that result on the final implementation.

The fixture keeps getStrictOptionValue’s real body. The reduced CompilerOptions interface retains every field that StrictOptionName permits. Calls cover missing, true and false options, and strict-mode fallback. A separate fixture checks receiver evaluation before key evaluation, a key that mutates the selected field, required fields and mixed required/optional fields. Generated helper parameters hold the receiver and key once, in JavaScript’s order. Exhaustive string-literal comparisons select one known data field. The read keeps the existing field representation, readiness and view checks. No backend file changed.

Shared files touched: internal/lower/object.go and internal/oracle/counts.md. New files: internal/lower/hidden_indexed_options.go, its focused negative test, internal/oracle/hidden_indexed_options_test.go and the two hidden_boundary_indexed_options .a fixtures.

Commands (stdout and stderr redirected to retained logs):

```sh
go test ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_indexed_options' -count=1 -v
go test ./internal/lower -run TestHiddenIndexedOptionsKeepsUnsupportedReads -count=1 -v
go test ./internal/oracle -run TestCountsAreRecorded -count=1 -args -update-counts
go test -overlay /workspace/hidden08-evidence/indexed-wrong-field-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_indexed_options[.]a$' -count=1 -v
go test -overlay /workspace/hidden08-evidence/indexed-repeat-key-overlay.json ./internal/oracle -run 'TestNativeAgreesWithNode/internal/oracle/testdata/hidden_boundary_indexed_options_order[.]a$' -count=1 -v
```

Normal checks exit 0. Each mutant exits 1 after successful compilation; Node and both backends otherwise exit 0 with empty stderr. The wrong-field mutant substitutes names[0] for names[i]. The repeated-key mutant evaluates the key twice and adds a second key line to both outputs. Negative checks require NotYet or Refused for arbitrary string indexing, getters, methods and prototype-member indexing.

Counts adds exactly two rows and changes no existing row: the real-body fixture records allocations/frees/retains/releases/high-water/region-bytes 18/18/0/18/2/0; the order fixture records 13/13/3/17/5/0. The final refresh passes in 43.306s. The existing toolchain setup remains in use, with /workspace/adamic-tools/env.sh and nproc 5; its timing lines are retained in ../hidden-08-never-array/REPORT.md.

Measurement uses the hidden.py calculation and independent audit.py byte-mask oracle from census pin 388096e6a83a4e9d287fb827f793c599ba1bf0ad. All 82 adapted source hashes match that pin. utilities.ts SHA-256 is ef43309e71a5bff946d868de2753b73de6b3cae07e5f215e2a2f1281ef0406ef. The complete unchanged compiler project was loaded for the checker, with LATENT_FULL=1 and LATENT_ASSERT_NO_OUTPUT=1. The scratch file loop selected utilities.ts while retaining every independent attempt in that file, including nested functions and checker skips. This is a regional measurement, not a new whole-corpus total. The before row comes from this branch’s previously completed full ledger. Pinned stock and both raw regional rows are retained compressed. Both independent byte-mask audits pass.

Assigned intersection [370572,370684): old hidden intersection 112 bytes, new hidden intersection 0 bytes, difference 112 bytes. No remaining boundary lies inside it. Additional changes elsewhere in utilities.ts are not counted as this unit’s revealed bytes.

The next boundary belongs to getNameOfScriptTarget at utilities.ts:9374:1, return-statement span [370796,370910), and reports utilities.ts:746:17, a function returning U | undefined. This is the generic-return family owned by codex/census-generic-returns and codex/scout-generics in the ownership table, so this unit stops there.
