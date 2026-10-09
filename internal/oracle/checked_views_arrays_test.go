package oracle

import "testing"

func TestCheckedViewArrays(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"array-holes-view", "-1\n7\n", ""},
		{"array-lazy", "2\n", ""},
		{"array-iteration", "7\n8\n", ""},
		{"array-map", "8:9\n", ""},
		{"array-pop", "8:7\n", ""},
		{"array-bounds", "-1:7\n", ""},
		{"array-second", "", "element read failed: values[1] expected number, found string"},
		{"array-length-kind", "", "field read failed: items(raw).values is not a readonly number[]; expected readonly number[], found object"},
		{"array-missing", "", "field read failed: items(raw).values is not initialized; expected readonly number[], found missing"},
		{"array-boolean", "", "element read failed: values[0] expected number, found boolean"},
		{"array-iteration-bad", "", "element read failed: items(raw).values[element] expected number, found string"},
		{"array-map-bad", "", "element read failed: items(raw).values[element] expected number, found string"},
		{"array-alias", "ok\n", "field read failed: values[0] expected \"ok\", found string changed"},
		{"array-undefined", "", "element read failed: items(raw).values[1] expected number, found undefined"},
		{"array-string-literal", "", "field read failed: values[element] expected \"ok\", found string bad"},
		{"nullable-array-string", "", "element read failed: values[element] expected number, found string"},
		{"nullable-array-missing", "", "element read failed: values[0] expected number, found undefined"},
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
			node := onNode(t, path)
			t.Logf("Node: exit=%d stdout=%q stderr=%q", node.exitCode, node.stdout, node.stderr)
			if probe.diagnostic == "" {
				if difference := disagreement(want, node); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				want.exitCode = 70
				want.stderr = []byte("adamic: panic: " + probe.diagnostic + "\n")
			}
			actual, _ := nativelyUncached(t, program)
			for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
				expected := want
				if difference := disagreement(expected, got); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
