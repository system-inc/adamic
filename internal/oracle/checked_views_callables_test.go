package oracle

import "testing"

func TestCheckedViewCallables(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, stdout string }{{"function-field", "8\n"}, {"class-method", "10\n"}} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal("Node: " + difference)
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

func TestCheckedViewOpaqueSignature(t *testing.T) {
	t.Parallel()
	program, path := interfaceFixture(t, "lane2/opaque-signature")
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "wrong\n" {
		t.Fatalf("Node %#v", truth)
	}
	sanitized, _ := nativelyUncached(t, program)
	expected := "adamic: panic: field read failed: (node as Runner).run expected (value: number) => number, found function with incompatible result representation\n"
	for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if got.exitCode != 70 || len(got.stdout) != 0 || string(got.stderr) != expected {
			t.Fatalf("got %#v want %q", got, expected)
		}
	}
}

func TestCheckedViewGenericArrayCast(t *testing.T) {
	t.Parallel()
	program, path := interfaceFixture(t, "lane2/native-generic-array-cast")
	want := run{stdout: []byte("1\n")}
	if difference := disagreement(want, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s; got %#v", difference, got)
		}
	}
}

func TestCheckedViewSameMap(t *testing.T) {
	program, path := interfaceFixture(t, "lane2/same-map")
	want := run{stdout: []byte("true:1:20:3:true\n")}
	if difference := disagreement(want, onNode(t, path)); difference != "" {
		t.Fatal(difference)
	}
	actual, _ := nativelyUncached(t, program)
	for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Fatalf("%s; got %#v", difference, got)
		}
	}
}

func TestCheckedViewParserReturns(t *testing.T) {
	for _, probe := range []struct{ name, stdout string }{{"native-same-map-return-cast", "2\nundefined\n"}, {"native-sorted-empty-conditional", "0\n1\n"}, {"optional-sort", "1,10,9\n1,9,10\nfalse,true,true\na,𝄞,\uE000\n"}} {
		t.Run(probe.name, func(t *testing.T) {
			program, path := interfaceFixture(t, "lane2/"+probe.name)
			want := run{stdout: []byte(probe.stdout)}
			if difference := disagreement(want, onNode(t, path)); difference != "" {
				t.Fatal(difference)
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
