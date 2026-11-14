package oracle

import "testing"

const objectAssignBytesFixture = "internal/oracle/testdata/library_object_assign_bytes.a"

func init() {
	fixtures = append(fixtures, struct {
		path            string
		lowers, checked bool
	}{objectAssignBytesFixture, true, false})
}

// String references print from their non-null pointer even before the byte-copy
// fix. The runtime-only null and sentinel cases live in internal/native.
func TestObjectAssignBytesAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, backend := range []struct {
		name  string
		index int
	}{{"native", 0}, {"javascript", 2}} {
		t.Run(backend.name, func(t *testing.T) {
			t.Parallel()
			fixtureAgreesOnBackend(t, objectAssignBytesFixture, backend.index)
		})
	}
}
