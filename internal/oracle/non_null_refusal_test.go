package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

func TestNonNullAdamicRefusal(test *testing.T) {
	test.Parallel()
	for _, fixture := range []struct{ path, position string }{
		{"stage3/census/repro/r05/main.a", ":2:59"},
		{"internal/oracle/testdata/non_null_refused/initialized.a", ":47:16"},
	} {
		test.Run(fixture.path, func(test *testing.T) {
			test.Parallel()
			path, absolutePathError := filepath.Abs(filepath.Join(repository, fixture.path))
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			_, loweredError := lowered(test, path)
			var refused *lower.Refused
			want := path +
				fixture.position +
				": Adamic 0.1 refuses the non-null assertion !; write ?? panic('why it can't be missing'), or narrow and handle the missing case"
			if !errors.As(loweredError, &refused) || !strings.HasSuffix(loweredError.Error(), want) {
				test.Fatalf("got %v, want %q", loweredError, want)
			}
		})
	}
}
