package oracle

import (
	"testing"
)

func fx7Depth(t *testing.T, name string) {
	t.Helper()
	fx7DepthBackend(t, name, "")
}
func fx7DepthBackend(t *testing.T, name, backend string) {
	t.Helper()
	program, path := interfaceFixture(t, "fx7/"+name)
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node: %#v", want)
	}
	for _, candidate := range []string{"native", "sanitized", "javascript"} {
		if backend != "" && backend != candidate {
			continue
		}
		var got run
		switch candidate {
		case "native":
			got = releasedUncached(t, program)
		case "sanitized":
			got, _ = nativelyUncached(t, program)
		case "javascript":
			got = onJavaScriptBackend(t, program)
		}
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s; stderr %q", candidate, diff, got.stderr)
		}
	}
}
func TestFX7Depth65(t *testing.T)          { t.Parallel(); fx7DepthBackend(t, "depth-65", "native") }
func TestFX7Depth65Sanitized(t *testing.T) { t.Parallel(); fx7DepthBackend(t, "depth-65", "sanitized") }
func TestFX7Depth65JavaScript(t *testing.T) {
	t.Parallel()
	fx7DepthBackend(t, "depth-65", "javascript")
}
func TestFX7Depth200(t *testing.T)  { t.Parallel(); fx7Depth(t, "depth-200") }
func TestFX7Depth2000(t *testing.T) { t.Parallel(); fx7Depth(t, "depth-2000") }

func init() {
	for _, name := range []string{"depth-65", "depth-200", "depth-2000", "p09", "p32", "p33", "p67", "p72", "scalar-tuple", "scalar-static"} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"stage3/interface-downcasts/fx7/" + name + ".ts", true, name != "depth-65" && name != "depth-200" && name != "depth-2000"})
	}
}

func fx7Text(t *testing.T, name string) {
	t.Helper()
	fx7TextBackend(t, name, "")
}
func fx7TextBackend(t *testing.T, name, backend string) {
	t.Helper()
	program, _ := interfaceFixture(t, "fx7/"+name)
	js := onJavaScriptBackend(t, program)
	t.Logf("JavaScript: exit %d stdout %q stderr %q", js.exitCode, js.stdout, js.stderr)
	want := run{exitCode: 70}
	switch name {
	case "p32":
		want.stdout = []byte("3\n")
		want.stderr = []byte("adamic: panic: field read failed: viewed.name is not a string; expected string, found number\n")
	case "p33":
		want.stdout = []byte("3\n")
		want.stderr = []byte("adamic: panic: field read failed: value.name is not a string; expected string, found number\n")
	case "p67":
		want.stderr = []byte("adamic: panic: field read failed: viewed.value matches no member of number | Inner; expected number | Inner, found function\n")
	case "scalar-static":
		want.stderr = []byte("adamic: panic: field read failed: viewed.value is not a boolean; expected boolean, found function\n")
	case "scalar-tuple":
		want.stderr = []byte("adamic: panic: field read failed: viewed.value is not a boolean; expected boolean, found array\n")
	case "p09", "p72":
		want.stderr = []byte("adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found array\n")
	default:
		t.Fatal("missing diagnostic pin")
	}
	if diff := disagreement(want, js); diff != "" {
		t.Fatalf("javascript: %s", diff)
	}
	for _, candidate := range []string{"native", "sanitized"} {
		if backend != "" && backend != candidate {
			continue
		}
		var got run
		if candidate == "native" {
			got = releasedUncached(t, program)
		} else {
			got, _ = nativelyUncached(t, program)
		}
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s; stderr %q", candidate, diff, got.stderr)
		}
	}
}
func TestFX7TextP09(t *testing.T) { t.Parallel(); fx7Text(t, "p09") }
func TestFX7TextP32(t *testing.T) { t.Parallel(); fx7Text(t, "p32") }
func TestFX7TextP33(t *testing.T) { t.Parallel(); fx7Text(t, "p33") }
func TestFX7TextP67(t *testing.T) { t.Parallel(); fx7Text(t, "p67") }
func TestFX7TextP72(t *testing.T) { t.Parallel(); fx7Text(t, "p72") }

func TestFX7TextScalarTuple(t *testing.T) { t.Parallel(); fx7Text(t, "scalar-tuple") }

func TestFX7TextScalarStatic(t *testing.T) {
	t.Parallel()
	fx7TextBackend(t, "scalar-static", "native")
}

func TestFX7TextScalarStaticSanitized(t *testing.T) {
	t.Parallel()
	fx7TextBackend(t, "scalar-static", "sanitized")
}
func TestFX7TextScalarStaticJavaScript(t *testing.T) {
	t.Parallel()
	fx7TextBackend(t, "scalar-static", "javascript")
}
