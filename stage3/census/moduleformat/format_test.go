package moduleformat

import "testing"

// Not parallel: t.Setenv sets CENSUS_MODULE_FORMAT in the shared process environment.
func TestDeclaredFormats(t *testing.T) {
	for _, row := range []struct{ input, name, value string }{{"", "CommonJSBundle", "CommonJSBundle"}, {"CommonJSBundle", "CommonJSBundle", "CommonJSBundle"}, {"EcmaScript", "EcmaScript", ""}} {
		t.Setenv("CENSUS_MODULE_FORMAT", row.input)
		if Name() != row.name || Resolve("CommonJSBundle", "") != row.value {
			t.Fatal(row)
		}
	}
}
// Not parallel: t.Setenv sets CENSUS_MODULE_FORMAT in the shared process environment.
func TestUnknownFormatFails(t *testing.T) {
	t.Setenv("CENSUS_MODULE_FORMAT", "commonjs")
	defer func() {
		if recover() == nil {
			t.Fatal("unknown format accepted")
		}
	}()
	Name()
}
