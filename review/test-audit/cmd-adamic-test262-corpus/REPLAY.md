Starting commit: 7b18d0576930caca4e22ce2eef92fcf563af52d0.

Each diffs/MNN.diff is a standalone production mutation without a selector. Apply one to the starting commit, source /workspace/adamic-tools/env.sh, run go vet ./cmd/adamic-test262/, then ADAMIC_TEST262_MEASURE=1 timeout 120 go test -json -count=1 -timeout 90s ./cmd/adamic-test262/ -run . > replay.log 2>&1. Restore before the next diff.

M06 and M13 are supplemental and excluded from every verdict. eligibility.json records this explicitly.

probes/PXX.switched.diff are separate empty-answer entry probes, selected with ADAMIC_MUTANT=PXX. Their rows are in probes.json. They are not production mutants and cannot establish sacred or subsumed verdicts.

report.md and rows.json are the deliverables; matrix.json contains all package failures, completion-checks.json verifies all 37 top-level tests completed every column, and raw logs preserve commands and failing output. Scripts and immutable source inputs document how evidence was generated. The mutation AST offsets in function-offsets.json are character offsets converted from Go UTF-8 byte offsets, with variadic forwarding corrected.
