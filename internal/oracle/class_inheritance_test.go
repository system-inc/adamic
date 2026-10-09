package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

// The shared oracle runs this fixture against source Node, generated JavaScript, native release,
// ASan/UBSan and the leak check. Register separately to avoid changing oracle_test.go's fixture list.
// Accessor spread remains refused when its getter can throw; library concat is now catchable.
func init() {
	for _, path := range []string{
		"internal/oracle/testdata/class_features_static.a",
		"internal/oracle/testdata/class_features_static_private.a",
		"internal/oracle/testdata/class_features_private.a",
		"internal/oracle/testdata/class_features_refused/class_features_accessors.a",
		"internal/oracle/testdata/class_features_twice.a",
		"internal/oracle/testdata/class_features_retained.a",
		"internal/oracle/testdata/class_features_distinct.a",
		"internal/oracle/testdata/class_inheritance.a",
		"internal/oracle/testdata/class_inheritance_exceptions.a",
		"internal/oracle/testdata/class_inheritance_order.a",
		"internal/oracle/testdata/class_identity.a",
		"internal/oracle/testdata/class_inheritance_memory.a",
		"internal/oracle/testdata/class_inheritance_generic.a",
		"internal/oracle/testdata/class_inheritance_interface.a",
		"internal/oracle/testdata/class_private_generic.a",
		"internal/oracle/testdata/class_private_members.a",
		"internal/oracle/testdata/class_inheritance_conditional.a",
		"internal/oracle/testdata/class_super_closure.a",
	} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{path, path != "internal/oracle/testdata/class_features_refused/class_features_accessors.a", false})
	}
}

func TestScout22AccessorSpreadRefusal(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/class_features_refused/class_features_accessors.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refusal *lower.NotYet
	if !errors.As(err, &refusal) || refusal.What != "spreading an accessor literal whose getter may throw" {
		t.Fatalf("want throwing getter spread refusal, got %v", err)
	}
}
