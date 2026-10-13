package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestNonNullAdamicRefusal(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct{ path, position string }{
		{"stage3/census/repro/r05/main.a", ":2:59"},
		{"internal/oracle/testdata/non_null_refused/initialized.a", ":3:25"},
	} {
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var refused *lower.Refused
			want := path + fixture.position + ": Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case"
			if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
				t.Fatalf("got %v, want %q", err, want)
			}
		})
	}
}
