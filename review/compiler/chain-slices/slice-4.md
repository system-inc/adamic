# Slice 4

Built on origin/main 5e33a17b186a8a2218d27b69b21e2de5acc5b750 alone. No stack envelope or other delivery branch was merged. All five candidates are kept; no required outside-slice dependency was found. Task closure here means the slice implementation and local proof are complete; main landing remains integration's action. The whole lowering chain and #wj4pmt1 remain larger than this unit.

| Member | Source SHA | Status | Dependency evidence | Tasks closed by this slice proof |
| --- | --- | --- | --- | --- |
| generic-body-relations | 2cc109e5 | kept | Checker shim operations plus main's instantiate.go:20, class.go:783 and invariance.go:309; no outside-chain symbol. | #js89dcw |
| generics-scout-main | 48391dcf | kept | Fixture-only compiler coverage. step16BodyRefusals at internal/oracle/step16_generics_test.go:33 depends on generic-body-relations within this slice. | Step 16; no additional named task |
| iteration-main | 518eca83 | kept | IteratorMethod/IteratorField belong to this member at internal/ir/iteration.go:5 and :15; main supplies object/call/exception machinery. Generator outcome registration uses generators-main within this slice. | #p6afa8g, #mtm9tke |
| project-references-main | e2aa3750 | kept | Uses main's loader/checker and its own project helpers. Probe imports only internal/load; declared in cloud/fast-gate/compiler-dependencies.json:22. | #acxgkff, step 32 |
| generators-main | 6f0ebd15 | kept | Own GeneratorTypes/GeneratorFrame at internal/ir/generator.go:5 and :35; iterationLocal/memberFunction existed on main. Combined iterator dispatch uses iteration-main within this slice. | #qk8rztp, step 20 |

Own ranges are 7a10c877..2cc109e5, 7a10c877..48391dcf, e511df28..518eca83, 0a3bea5c..e2aa3750 and b5245943..6f0ebd15. Main remains unchanged at the original base after the final fetch.

Repairs: aea87960's indexed-witness oracle hunk; 91b0148f's ordinary-file selection, test and mixed-file oracle hunks; 2ca18b1b's project probe declaration only; 71972e18's generator outcome registration; the composed scout's exact generic-body refusal expectations described by 9e1e73e0. The 2b2b1095 maybeBind isolation assumed checked-any's explicit-any refusal and was removed after its main-alone control disproved that assumption. No checked-any code was imported. Counts are regenerated locally rather than copied from the chain.

The five named member commits are 113d76a5, a018fc10, 64b77e64, 19ca1506 and 704edd9c. fd6f0ef6 completes the project's fixture staging; 162d3d21 reconciles owning repairs, guards and counts; 4f30d11d fixes three isolated test parallel annotations. History was preserved.

Proof: 292 internal/lower top-level tests in 12 shards, maximum command duration 19.256 seconds; each member's source Node comparisons and both backends with native ASan/UBSan and leaks; 46 caught mutants; 1,814 admission witnesses against main with 55 attributed deltas and no tracked-main accepted witness losing admission; TestCallTargetReaders including oracle tests; all 1,066 regenerated counts rows attributed, with 55 additions and four existing changes. Counts wrote once successfully in 50.921 seconds after one timed-out attempt wrote nothing.

Lane output: lane checks 4.3 s: gofmt and tools on 71 Go files, t.Parallel on 6 test packages; a-check 13 .a files; vet 6 packages.

[Full report](../chain-slice-4/REPORT.md), [mutants](../chain-slice-4/mutants.md), [admission delta](../chain-slice-4/admission-delta.json), [counts attribution](../chain-slice-4/counts-attribution.md). Two exploratory review inputs panic identically on main and the slice; the opt-in full compiler census is not claimed run. No complete package gate, full gate or full native tsc execution was run.
