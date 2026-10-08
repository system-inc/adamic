package oracle

import (
	"errors"
	"github.com/system-inc/adamic/internal/lower"
	"path/filepath"
	"testing"
)

// This reduction is registered separately because its expected outcome is a
// design Refused, rather than a stage-zero NotYet or an executable fixture.
var objectSmallArgsRefused = []string{"internal/oracle/testdata/object_small_args_refused.a"}

func TestObjectSmallArgsRefusal(t *testing.T) {
	for _, fixture := range objectSmallArgsRefused {
		path, err := filepath.Abs(filepath.Join(repository, fixture))
		if err != nil {
			t.Fatal(err)
		}
		observed := onNode(t, path)
		if observed.exitCode != 0 || string(observed.stdout) != "event\ntrue\nfalse\ntrue\n" || len(observed.stderr) != 0 {
			t.Fatalf("Node: %+v", observed)
		}
		_, err = lowered(t, path)
		var refused *lower.Refused
		if !errors.As(err, &refused) || refused.What != "an index signature" {
			t.Fatalf("want the index-signature design refusal, got %v", err)
		}
	}
}

func TestObjectSmallOptionalWriteRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_small_index_optional_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "false\ntrue\n12\n" || len(observed.stderr) != 0 {
		t.Fatalf("Node: %+v", observed)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || refused.What != "adding an own field through a computed optional-field write" {
		t.Fatalf("want the fixed-shape design refusal, got %v", err)
	}
}

func TestObjectSmallSpreadNULRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_small_spread_nul_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "[\"id\",\"a\\u0000x\",\"a\"]\n" || len(observed.stderr) != 0 {
		t.Fatalf("Node: %+v", observed)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "an extending object spread in a program with NUL-bearing property names" {
		t.Fatalf("want the explicit name-storage stop, got %v", err)
	}
}

func TestObjectSmallIndexNULRefusal(t *testing.T) {
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/object_small_index_nul_refused.a"))
	if err != nil {
		t.Fatal(err)
	}
	observed := onNode(t, path)
	if observed.exitCode != 0 || string(observed.stdout) != "1,5\n" || len(observed.stderr) != 0 {
		t.Fatalf("Node: %+v", observed)
	}
	_, err = lowered(t, path)
	var notYet *lower.NotYet
	if !errors.As(err, &notYet) || notYet.What != "assigning a computed object field whose name contains NUL" {
		t.Fatalf("want the explicit name-storage stop, got %v", err)
	}
}
