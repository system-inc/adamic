package oracle

import "testing"

func TestCheckedViewNativeArrays(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"native-array-join", "7:8\n7,8\n8\n", ""},
		{"native-array-join-evaluation", "7:8:9\n2\n", ""},
		{"native-array-large-field", "1000000000\n", ""},
		{"native-array-join-boolean", "true:false\n", ""},
		{"native-array-mutation", "3\n10\n9:8\n", ""},
		{"native-array-string-mutation", "3\nd\nc:b\n", ""},
		{"native-array-sparse", ":7::9\n:8::10\n16\n9\n3\n7::10\n", ""},
		{"native-array-sparse-pop-hole", "-1\n2\n-1\n7\n", ""},
		{"native-array-join-bad", "2\n", "cast failed: element read failed: values[element] expected number, found string"},
		{"native-array-join-literal", "", "cast failed: field read failed: items(raw).values[element] expected \"ok\", found string bad"},
		{"native-array-write-bad", "", "cast failed: element read failed: <array write> expected number, found boolean"},
		{"native-array-push-bad", "", "cast failed: element read failed: <array write> expected number, found boolean"},
		{"native-array-copy-bad", "", "cast failed: element read failed: values[0] expected number, found boolean"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			// Unsafe source casts execute on Node too; their checked failure is the
			// language's additional contract, pinned below for both backends.
			source := onNode(t, path)
			if probe.diagnostic == "" {
				if difference := disagreement(want, source); difference != "" {
					t.Fatal("Node: " + difference)
				}
			} else {
				if source.exitCode != 0 {
					t.Fatalf("Node control: %#v", source)
				}
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
