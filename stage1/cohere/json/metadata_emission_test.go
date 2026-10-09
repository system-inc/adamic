package json

import (
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

func TestPortElementMetadataEmission(t *testing.T) {
	t.Parallel()
	// The port's largest C translation unit includes the Unicode tables. Keep
	// each allocation's certificate, but leave sanitizer-visible stores in the
	// runtime instead of repeating them throughout the table initializers.
	program := lowered(t, filepath.Join(portDirectory(t, nil), "main.ts"))
	source := native.C(program)
	stores := len(regexp.MustCompile(`->element_kind\s*=`).FindAllStringIndex(source, -1))
	calls := strings.Count(source, "adamic_array_view_storage(")
	t.Logf("element metadata: %d direct stores, %d runtime calls; C bytes %d", stores, calls, len(source))
	if stores > 0 {
		t.Fatalf("element metadata direct stores %d exceed bound 0", stores)
	}
	if calls != 2885 {
		t.Fatalf("element metadata runtime calls %d; want 2885 allocation certificates", calls)
	}
}
