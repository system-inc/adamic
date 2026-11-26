# generic_spread_copy_reads.go's mutant

- **The stop answers nil** (`checkGenericSpreadCopyReads` returns nil before its walk): generic_method_call.ts lowers
  again and natively calls Scale's weigh on a copy that has none, so `TestGenericSpreadPrototypeMember`
  (internal/oracle/generic_spread_prototype_test.go) goes red on it, and on generic_method_typeof.ts and
  generic_setter.ts, each wanting the read stopped. generic_getter.ts stays stopped, by spreadCopyRead's words
  rather than these, so it goes red on the words alone. Run on Loom as its own red job.
