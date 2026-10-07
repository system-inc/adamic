package oracle

import "testing"

func TestCheckedViewArrays(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"array-lazy", "2\n", ""},
		{"array-iteration", "7\n8\n", ""},
		{"array-map", "8:9\n", ""},
		{"array-pop", "8:7\n", ""},
		{"array-bounds", "-1:7\n", ""},
		{"array-second", "", "element read failed: values[1] expected number, found string"},
		{"array-length-kind", "", "field read failed: items(raw).values is not a readonly number[]; expected readonly number[], found object"},
		{"callable-missing", "", "field read failed: runner(raw).run is not initialized; expected function, found missing"},
		{"callable-uninitialized", "", "field read failed: runner(raw).run is not initialized; expected function, found uninitialized"},
		{"callable-kind", "", "field read failed: runner(raw).run is not a function; expected function, found number"},
		{"array-missing", "", "field read failed: items(raw).values is not initialized; expected readonly number[], found missing"},
		{"array-uninitialized", "", "field read failed: items(raw).values is not initialized; expected readonly number[], found uninitialized"},
		{"array-boolean", "", "element read failed: values[0] expected number, found boolean"},
		{"array-iteration-bad", "", "element read failed: items(raw).values[element] expected number, found string"},
		{"array-map-bad", "", "element read failed: items(raw).values[element] expected number, found string"},
		{"array-alias", "ok\n", "field read failed: values[0] expected \"ok\", found string changed"},
		{"array-undefined", "", "element read failed: items(raw).values[1] expected number, found undefined"},
		{"array-string-literal", "", "field read failed: values[element] expected \"ok\", found string bad"},
		{"nullable-array-string", "", "element read failed: values[element] expected number, found string"},
		{"nullable-array-element", "", "element read failed: values[0] expected number, found string"},
		{"generic-array-element", "", "element read failed: values[0] expected string, found number"},
		{"array-field", "2\n", ""},
		{"array-element", "7\n", ""},
		{"object-element", "ok\n", ""},
		{"non-array", "", "field read failed: items(raw).values is not a readonly number[]; expected readonly number[], found number"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if probe.diagnostic == "" {
				if difference := disagreement(want, onNode(t, path)); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
