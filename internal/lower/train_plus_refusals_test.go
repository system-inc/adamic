package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

// The .ts copy exercises the same user program through the legacy source path.
// Both paths must stop at creation, before backend emission can expose a value.
func trainPlusRefusal(t *testing.T, name, location string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("..", "oracle", "testdata", "review", "refused", name+".a"))
	if err != nil {
		t.Fatal(err)
	}
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	legacy := filepath.Join(t.TempDir(), name+".ts")
	if err := os.WriteFile(legacy, source, 0600); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []string{path, legacy} {
		program, err := load.Load([]string{entry})
		if err != nil {
			t.Fatal(err)
		}
		result, err := Lower(context.Background(), program)
		var refused *Refused
		if result != nil || !errors.As(err, &refused) {
			t.Errorf("%s: want creation-site refusal, got program %v, error %v", entry, result != nil, err)
			continue
		}
		if refused.Where != entry+":"+location || refused.What != "view type has an unsupported member: value" {
			t.Errorf("%s: wrong creation refusal: %#v", entry, refused)
		}
		t.Logf("%s: %s: %s", filepath.Ext(entry), refused.Where, refused.What)
	}
}

func TestTrainPlusRefusalP26(t *testing.T) {
	t.Parallel()
	trainPlusRefusal(t, "fxspptb_oct9_native_p26_callable_or_undefined", "7:14")
}

func TestTrainPlusRefusalP33(t *testing.T) {
	t.Parallel()
	trainPlusRefusal(t, "fxspptb_oct9_native_p33_map_in_union", "7:16")
}

func TestTrainPlusRefusalP34(t *testing.T) {
	t.Parallel()
	trainPlusRefusal(t, "fxspptb_oct9_native_p34_typed_array_in_union", "5:16")
}

func TestTrainPlusRefusalP35(t *testing.T) {
	t.Parallel()
	trainPlusRefusal(t, "fxspptb_oct9_native_p35_class_in_union", "6:16")
}
