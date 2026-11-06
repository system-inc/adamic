package oracle

import "testing"

// Every step result a collection iterator gives is built in one place, map_set.c's
// adamic_iterator_result (#ak92fcp). testdata/iterator_result_builder/steps.a steps a Map's values(),
// a Set's values() and a string-keyed Map's keys() by hand, one for each kind of value slot, and
// agrees with Node on each result's keys and value at a step and at done, natively, released and
// sanitized, and on the JavaScript backend, and leaks nothing. The boolean Set's value is read by name
// and by a run-time key, which reads the result's representation byte (union.c's dynamic_slot). Its
// mutants, each with the test it turns red, are in testdata/iterator_result_builder/mutants.md.
func TestIteratorResultBuilderAgreesWithNode(t *testing.T) {
	t.Parallel()
	generatorAgrees(t, "internal/oracle/testdata/iterator_result_builder/steps.a")
}
