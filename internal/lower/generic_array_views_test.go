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
