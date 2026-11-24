# spread_copy_reads.go's mutant

- **The read check answers nil** (`spreadCopyRead` returns nil before its walk): c13b lowers again and loads a slot
  its copy doesn't have, so `TestSpreadPrototypeField` (internal/oracle/spread_prototype_test.go) goes red on
  c13b.ts, c13c.ts and c13b_a.a, each wanting the read stopped. Run on Loom as its own red job.
