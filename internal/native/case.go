package native

// runtime/case_tables.h is Unicode's case mappings for the version Node carries, and
// case_generate.go writes it; TestCaseTablesMatchNodesUnicode says when Node has moved on.
//go:generate go run case_generate.go
//go:generate go run normalize_generate.go
