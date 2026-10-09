package flow

import "testing"

func TestGeneratorBasic(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/basic.a")
}

func TestGeneratorCancellation(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/cancellation.a")
}

func TestGeneratorThrow(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/throw.a")
}

func TestGeneratorDefaults(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/defaults.a")
}

func TestGeneratorClose(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/close.a")
}

func TestGeneratorDelegate(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/delegate.a")
}

func TestGeneratorDelegateThrow(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/delegate_throw.a")
}

func TestGeneratorReentrant(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/reentrant.a")
}

func TestGeneratorOwned(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/owned.a")
}

func TestGeneratorCaptures(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/captures.a")
}

func TestGeneratorDefaultError(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/default_error.a")
}

func TestGeneratorMapIterator(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/map_iterator.a")
}

func TestGeneratorChecker(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/checker.a")
}

func TestGeneratorFirstReference(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/first_reference.a")
}

func TestGeneratorCompletions(t *testing.T) {
	t.Parallel()
	checkAllFlowProgram(t, "../oracle/testdata/generators/completions.a")
}
