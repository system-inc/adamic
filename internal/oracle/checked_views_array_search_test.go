package oracle

import "testing"

func TestCheckedViewArraySearch(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout, diagnostic string }{
		{"native-array-search", "-Infinity:0:-1:true\n-20:0:-1:true\n-5:0:0:true\n-2.9:-1:2:true\n-1:-1:2:false\n0:0:0:true\n0:0:0:true\n1.9:2:0:true\n3:-1:2:true\n5:-1:2:false\n20:-1:2:false\nInfinity:-1:2:false\nNaN:0:0:true\n2:0:-1:true\n", ""},
		{"native-array-search-lazy", "0\ntrue\n", ""},
		{"native-array-search-object", "0:2:true\n-1\n", ""},
		{"native-array-search-sparse", "1:3:false:true\n", ""},
		{"native-array-search-evaluation", "1:2:9\n", ""},
		{"native-array-search-boolean", "1:true\n", ""},
		{"native-array-reference-write", "", "element write failed: <array write> expected { value: number; secret: number; }, found { value: number; }"},
		{"native-array-search-bad", "", "element read failed: items(raw).values[element] expected number, found string"},
		{"native-array-search-literal", "", "field read failed: items(raw).values[element] expected \"ok\", found string bad"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			source := onNode(t, path)
			if probe.diagnostic == "" {
				if difference := viewReadDisagreement(want, source, program); difference != "" {
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
				if difference := viewReadDisagreement(want, got, program); difference != "" {
					t.Fatalf("%s; got %#v", difference, got)
				}
			}
		})
	}
}
