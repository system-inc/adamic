package oracle

// A union of records with one slot representation is the record pointer, read as a record, held to Node (#6c4swvt G2b).
// Mutants: recordLiteralElement's union slot dropped, and the bare literal is built as a fixed object; recordReadTarget
// falling back to recordTarget, and the fixture is refused at the read. Its three pins stop as their headers say.
func init() {
	fixtures = append(fixtures, struct {
		path    string
		lowers  bool
		checked bool
	}{"internal/oracle/testdata/records_union_of_records.a", true, false})
	recordViewPins = append(recordViewPins, "records_union_of_records_storage", "records_union_of_records_representation", "records_union_of_records_write")
}
