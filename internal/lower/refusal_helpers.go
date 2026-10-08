package lower

// A refusal helper is an error-returning method that the up-front refusal pass in refusals.go
// calls. Each one registers itself beside its definition, naming the internal/refusalprobe
// catalog entry that owns its refusals, so an area that adds a helper finds a missing
// registration in this package's own tests rather than in the catalog's, after a merge.
var refusalHelpers = map[string]string{}

// registerRefusalHelper is called from a package-level declaration next to the helper:
//
//	var _ = registerRefusalHelper("enumRefusal", "enum-object-write")
func registerRefusalHelper(helper, entry string) struct{} {
	if _, duplicate := refusalHelpers[helper]; duplicate {
		panic("refusal helper " + helper + " is registered twice")
	}
	refusalHelpers[helper] = entry
	return struct{}{}
}

// RefusalHelpers returns each registered refusal helper's catalog entry, keyed by method name.
func RefusalHelpers() map[string]string {
	helpers := make(map[string]string, len(refusalHelpers))
	for helper, entry := range refusalHelpers {
		helpers[helper] = entry
	}
	return helpers
}
