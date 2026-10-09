package load

import "testing"

func TestDateJSONRequiresNullHandling(t *testing.T) {
	t.Parallel()
	diagnostics := checkErrors(t, writeProgram(t, [2]string{"main.a", "const value: string = new Date(0).toJSON(); console.log(value);"}))
	if len(diagnostics) == 0 {
		t.Fatal("nullable Date.toJSON result accepted as an always-present string")
	}
}
