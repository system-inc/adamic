package lower

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
)

func TestNullableScoutGenericReturn(t *testing.T) {
	t.Parallel()
	requireNullableScoutLowers(t, "generic-return")
}

func TestNullableScoutGenericValue(t *testing.T) {
	t.Parallel()
	requireNullableScoutLowers(t, "generic-value")
}

func TestNullableScoutBrandedString(t *testing.T) {
	t.Parallel()
	requireNullableScoutStop(t, "branded-string")
}

func requireNullableScoutStop(t *testing.T, name string) {
	t.Helper()
	path := "testdata/nullable_scout/" + name + ".a"
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	var stop *NotYet
	if !errors.As(err, &stop) {
		t.Fatalf("want current NotYet, got %v", err)
	}
	expected, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".notyet")
	if err != nil {
		t.Fatal(err)
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.ReplaceAll(stop.Error(), absolute, path)
	if got != strings.TrimSpace(string(expected)) {
		t.Fatalf("got %q, want %q", got, strings.TrimSpace(string(expected)))
	}
}

func requireNullableScoutLowers(t *testing.T, name string) {
	t.Helper()
	program, err := load.Load([]string{"testdata/nullable_scout/" + name + ".a"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Lower(context.Background(), program); err != nil {
		t.Fatal(err)
	}
}
