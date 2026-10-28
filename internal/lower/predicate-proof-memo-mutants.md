# The predicate proof memo: its mutants (#pv935yw)

provePredicate reads one fixed point per verifier mode from `lowering.predicateProofs`, run to its end the first time
a mode is asked. Three mutants in predicates_proof.go's `predicateFixedPoint`, each run with
`go test ./internal/lower -run '^(TestPredicateProofMemoAnswersAsAFreshLowering|TestPredicateBodyProof)$' -count=1 -v`
at b7b5ac9e with the change, on macOS. Restored, both pass.

1. The fixed point stopped after its first pass (`if !changed || true { break }`). A predicate proved only through a
   helper declared after it never converges:
   ```
   predicates_proof_test.go:77: want proof tagged=true, got {TaggedView:false}, ...: unconverged or opaque helper on this return path ...
   --- FAIL: TestPredicateBodyProof/kind_helpers_fixed_point
   --- FAIL: TestPredicateProofMemoAnswersAsAFreshLowering (isModuleName alone: refused ... want proved, tagged=true)
   ```
2. One fixed point for both modes (`field = ""` before the lookup), so a tag predicate reads the typeof mode's:
   ```
   --- FAIL: TestPredicateBodyProof/kind (want proof tagged=true ... the target has an unsupported runtime contract)
   --- FAIL: TestPredicateProofMemoAnswersAsAFreshLowering (isModuleName alone: refused ... want proved, tagged=true)
   ```
3. A cached mode answered empty (`return &predicateFixedPoint{}` on a hit), so a second ask in one lowering refuses:
   ```
   predicates_proof_test.go:340: isIdentifier from the memo: refused: ... unconverged helper proof ..., want what a fresh lowering gives: proved, tagged=true
   --- FAIL: TestPredicateProofMemoAnswersAsAFreshLowering
   --- FAIL: TestPredicateBodyProof/kind_helpers_fixed_point (want checked-view admission ...)
   ```
