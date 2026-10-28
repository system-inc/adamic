# The entries record memo: its mutants (#pv935yw)

entriesRecordLiteral keeps its answer per literal (`lowering.entriesRecords`), and entriesProgramHasRecords keeps the
program's (`entriesProgramRecords`, once `entriesProgramRecordsKnown`), both on the lowering. Three mutants in
entries_records.go, each run with
`go test ./internal/lower -run '^TestEntriesRecordMemoAnswersAsAFreshLowering$' -count=1 -v` at b7b5ac9e with the
change, on macOS. Restored, it passes.

1. The program's answer kept across lowerings (a package-level pair of bools in place of the lowering's). The program
   without records reads the first program's true:
   ```
   entries_provenance_test.go:127: no records: the program alone has records=true, want false
   ```
2. A kept program answer read as false (`return l.entriesProgramRecords && false`):
   ```
   entries_provenance_test.go:127: records: the program alone has records=false, want true
   ```
3. One answer kept for every literal (the memo keyed by `nil` in place of the literal). The second literal reads the
   first's:
   ```
   entries_provenance_test.go:136: records: literal 1 from the memo is a record=true, want false
   ```
