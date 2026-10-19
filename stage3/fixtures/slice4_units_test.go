package fixtures

import "testing"

func TestFixturesGenerics(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "generics")
}

func TestFixturesIteration(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "iteration")
}

func TestFixturesIterationDispatch(t *testing.T) {
	t.Parallel()
	testFixtureDirectory(t, "iteration-dispatch")
}
