package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProgramDiscoveryAllocationIdentities(test *testing.T) {
	test.Parallel()
	loaded, loadError := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_extended.a")})
	if loadError != nil {
		test.Fatal(loadError)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	defer release()
	first, lowerCheckedError := lowerChecked(loaded, check, entry, nil, true, false)
	if lowerCheckedError != nil {
		test.Fatal(lowerCheckedError)
	}
	second, lowerCheckedError := lowerChecked(loaded, check, entry, nil, true, false)
	if lowerCheckedError != nil {
		test.Fatal(lowerCheckedError)
	}
	a, b := first.programPlan, second.programPlan
	if !reflect.DeepEqual(a.types, b.types) ||
		!reflect.DeepEqual(a.cells, b.cells) ||
		!reflect.DeepEqual(a.functions, b.functions) ||
		!reflect.DeepEqual(a.classes, b.classes) {
		test.Fatal("repeated discovery changed membership")
	}
	test.Logf(
		"same-input discovery membership stable: types=%d locals=%d functions=%d classes=%d",
		len(a.types),
		len(a.cells),
		len(a.functions),
		len(a.classes),
	)
	if !reflect.DeepEqual(a.identities, b.identities) {
		test.Fatal("repeated discovery changed allocation identities")
	}
	built, lowerCheckedError := lowerChecked(loaded, check, entry, a, false, false)
	if lowerCheckedError != nil {
		test.Fatal(lowerCheckedError)
	}
	if !reflect.DeepEqual(programAllocationIdentities(built.result), a.identities) {
		test.Fatal("final lowering changed allocation identities")
	}
}
