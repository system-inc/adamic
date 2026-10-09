package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/load"
	"path/filepath"
	"reflect"
	"testing"
)

func TestProgramDiscoveryAllocationIdentities(t *testing.T) {
	loaded, err := load.Load([]string{filepath.Join("..", "oracle", "testdata", "arguments_length_extended.a")})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	check, release := loaded.Checker(context.Background(), entry)
	defer release()
	first, err := lowerChecked(loaded, check, entry, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	second, err := lowerChecked(loaded, check, entry, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	a, b := first.programPlan, second.programPlan
	if !reflect.DeepEqual(a.types, b.types) || !reflect.DeepEqual(a.cells, b.cells) || !reflect.DeepEqual(a.functions, b.functions) || !reflect.DeepEqual(a.classes, b.classes) {
		t.Fatal("repeated discovery changed membership")
	}
	t.Logf("same-input discovery membership stable: types=%d locals=%d functions=%d classes=%d", len(a.types), len(a.cells), len(a.functions), len(a.classes))
	if !reflect.DeepEqual(a.identities, b.identities) {
		t.Fatal("repeated discovery changed allocation identities")
	}
	built, err := lowerChecked(loaded, check, entry, a, false)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(programAllocationIdentities(built.result), a.identities) {
		t.Fatal("final lowering changed allocation identities")
	}
}
