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
		{"array-second", "", "cast failed: element read failed: values[1] expected number, found string"},
		{"array-length-kind", "", "cast failed: field read failed: items(raw).values is not a readonly number[]; expected readonly number[], found object"},
		{"callable-missing", "", "cast failed: field read failed: runner(raw).run is not initialized; expected (value: number) => number, found missing"},
		{"callable-uninitialized", "", "cast failed: field read failed: runner(raw).run is not initialized; expected (value: number) => number, found uninitialized"},
		{"callable-kind", "", "cast failed: field read failed: runner(raw).run is not a (value: number) => number; expected (value: number) => number, found number"},
		{"array-missing", "", "cast failed: field read failed: items(raw).values is not initialized; expected readonly number[], found missing"},
		{"array-uninitialized", "", "cast failed: field read failed: items(raw).values is not initialized; expected readonly number[], found uninitialized"},
		{"array-boolean", "", "cast failed: element read failed: values[0] expected number, found boolean"},
		{"array-iteration-bad", "", "cast failed: element read failed: items(raw).values[element] expected number, found string"},
		{"array-map-bad", "", "cast failed: element read failed: items(raw).values[element] expected number, found string"},
		{"array-alias", "ok\n", "cast failed: field read failed: values[0] expected \"ok\", found string changed"},
		{"array-undefined", "", "cast failed: element read failed: items(raw).values[1] expected number, found undefined"},
		{"array-string-literal", "", "cast failed: field read failed: values[element] expected \"ok\", found string bad"},
		{"nullable-array-string", "", "cast failed: element read failed: values[element] expected number, found string"},
		{"nullable-array-element", "", "cast failed: element read failed: values[0] expected number, found string"},
		{"generic-array-element", "", "cast failed: element read failed: values[0] expected string, found number"},
		{"array-field", "2\n", ""},
		{"array-element", "7\n", ""},
		{"object-element", "ok\n", ""},
		{"non-array", "", "cast failed: field read failed: items(raw).values is not a readonly number[]; expected readonly number[], found number"},
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
			for index, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				expected := want
				if probe.name == "callable-kind" && index == 2 {
					expected.stderr = []byte("adamic: panic: cast failed: field read failed: runner(raw).run expected (value: number) => number, found number\n")
				}
				if difference := disagreement(expected, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
