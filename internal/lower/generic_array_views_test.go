package lower

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestGenericNullableArrayViewRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/generic_nullable_array_view_refused.a")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSource(t, string(source))
	var refused *Refused
	want := "type argument T = string | undefined breaks invariance: a value of type NonNullable<T>[] seen as T[], which can write string | undefined where string is read"
	fix := "use a type argument that includes neither null nor undefined, or keep the array view readonly (adamic/invariant-mutable)"
	if !errors.As(err, &refused) || refused.What != want || refused.Fix != fix || !strings.HasSuffix(refused.Where, ":6:13") {
		t.Fatalf("want pinned call-site nullable type argument refusal, got %v", err)
	}
	t.Log(err)
}

func TestGenericNullableArrayViewNullRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/generic_nullable_array_view_refused.a")
	if err != nil {
		t.Fatal(err)
	}
	probe := strings.ReplaceAll(string(source), "string | undefined", "string | null")
	probe = strings.ReplaceAll(probe, "], undefined)", "], null)")
	_, err = lowerSource(t, probe)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "type argument T = string | null breaks invariance") {
		t.Fatalf("want null-admitting type argument refusal, got %v", err)
	}
}

func TestGenericNullableArrayViewInClosureRefused(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("testdata/generic_nullable_array_view_refused.a")
	if err != nil {
		t.Fatal(err)
	}
	probe := strings.Replace(string(source), "write<T>(result, next);", "const append = () => write<T>(result, next); append();", 1)
	_, err = lowerSource(t, probe)
	var refused *Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "type argument T = string | undefined breaks invariance") {
		t.Fatalf("want closure's nullable array view refused at instantiation, got %v", err)
	}
}
