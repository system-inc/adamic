package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func init() {
	for _, path := range []string{"library_object_numeric_keys.a", "library_object_own_keys_receivers.a", "library_object_collection_integrity.a", "library_object_descriptor_errors.a", "library_object_descriptors.a", "library_object_descriptor_order.a", "library_object_descriptor_same_value.a", "library_object_define_properties.a"} {
		fixtures = append(fixtures, struct {
			path    string
			lowers  bool
			checked bool
		}{"internal/oracle/testdata/" + path, true, false})
	}
}

func TestObjectDescriptorConsumerStaysRefused(t *testing.T) {
	assertObjectDescriptorRefusal(t, "library_object_descriptor_shape_consumer.a", "static complete shape")
}

func TestObjectDescriptorMapStaysRefused(t *testing.T) {
	assertObjectDescriptorRefusal(t, "library_object_descriptor_map_mutated.a", "descriptor maps")
}

func assertObjectDescriptorRefusal(t *testing.T, fixture, reason string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_descriptor_probes", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		var refused *lower.Refused
		var notYet *lower.NotYet
		if (!errors.As(err, &refused) && !errors.As(err, &notYet)) || !strings.Contains(err.Error(), reason) {
			t.Fatal(err)
		}
		return
	}
	node := onNode(t, path)
	native, _ := natively(t, program)
	if difference := disagreement(node, native); difference != "" {
		t.Fatalf("%s: node stdout %q native stdout %q", difference, node.stdout, native.stdout)
	}
	t.Fatal("probe now agrees with Node; implement and register the newly supported behavior")
}

func TestObjectNumericTypedSlotStaysRefused(t *testing.T) {
	assertObjectDescriptorRefusal(t, "library_object_numeric_typed_slot.a", "keys must be constant strings")
}
