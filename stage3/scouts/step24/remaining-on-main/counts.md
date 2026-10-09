# Local counts

Three .a fixtures: optional-field-presence, nullable-result-relation, nullable-argument.
Three Node goldens pass; three source mutants caught by differing stdout.
Three in-place main checker rejections verified: TS2375, TS2322, TS2345.
Four evidence mutants caught: fabricated clearance, count, first stop, combined pass.
Each fixture unit has a 30-second subprocess deadline and total-duration assertion.
No native fixture execution, combined-compiler clearance, or runtime guard proof.
