package oracle

import (
	"testing"
)

func fx7Depth(test *testing.T, name string) {
	test.Helper()
	fx7DepthBackend(test, name, "")
}
func fx7DepthBackend(test *testing.T, name, backend string) {
	test.Helper()
	program, path := interfaceFixture(test, "fx7/"+name)
	want := onNode(test, path)
	if want.exitCode != 0 {
		test.Fatalf("Node: %#v", want)
	}
	for _, candidate := range []string{"native", "sanitized", "javascript"} {
		if backend != "" && backend != candidate {
			continue
		}
		var got run
		switch candidate {
		case "native":
			got = releasedUncached(test, program)
		case "sanitized":
			got, _ = nativelyUncached(test, program)
		case "javascript":
			got = onJavaScriptBackend(test, program)
		}
		if diff := disagreement(want, got); diff != "" {
			test.Errorf("%s: %s; stderr %q", candidate, diff, got.stderr)
		}
	}
}
func TestFX7Depth65(test *testing.T) { test.Parallel(); fx7DepthBackend(test, "depth-65", "native") }
func TestFX7Depth65Sanitized(test *testing.T) {
	test.Parallel()
	fx7DepthBackend(test, "depth-65", "sanitized")
}
func TestFX7Depth65JavaScript(test *testing.T) {
	test.Parallel()
	fx7DepthBackend(test, "depth-65", "javascript")
}
func TestFX7Depth200(test *testing.T)  { test.Parallel(); fx7Depth(test, "depth-200") }
func TestFX7Depth2000(test *testing.T) { test.Parallel(); fx7Depth(test, "depth-2000") }

func init() {
	for _, name := range []string{
		"depth-65",
		"depth-200",
		"depth-2000",
		"p09",
		"p32",
		"p33",
		"p67",
		"p72",
		"scalar-tuple",
		"scalar-static",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{
			"stage3/interface-downcasts/fx7/" + name + ".ts",
			true,
			name != "depth-65" && name != "depth-200" && name != "depth-2000",
		})
	}
}

func fx7Text(test *testing.T, name string) {
	test.Helper()
	fx7TextBackend(test, name, "")
}
func fx7TextBackend(test *testing.T, name, backend string) {
	test.Helper()
	program, _ := interfaceFixture(test, "fx7/"+name)
	js := onJavaScriptBackend(test, program)
	test.Logf("JavaScript: exit %d stdout %q stderr %q", js.exitCode, js.stdout, js.stderr)
	want := run{exitCode: 70}
	switch name {
	case "p32":
		want.stdout = []byte("3\n")
		want.stderr = []byte(
			"adamic: panic: field read failed: viewed.name is not a string; expected string, found number\n",
		)
	case "p33":
		want.stdout = []byte("3\n")
		want.stderr = []byte(
			"adamic: panic: field read failed: value.name is not a string; expected string, found number\n",
		)
	case "p67":
		want.stderr = []byte(
			"adamic: panic: field read failed: viewed.value matches no member of number | Inner; expected number | Inner, found function\n",
		)
	case "scalar-static":
		want.stderr = []byte(
			"adamic: panic: field read failed: viewed.value is not a boolean; expected boolean, found function\n",
		)
	case "scalar-tuple":
		want.stderr = []byte(
			"adamic: panic: field read failed: viewed.value is not a boolean; expected boolean, found array\n",
		)
	case "p09", "p72":
		want.stderr = []byte(
			"adamic: panic: field read failed: viewed.value matches no member of readonly [string, number]; expected readonly [string, number], found array\n",
		)
	default:
		test.Fatal("missing diagnostic pin")
	}
	if diff := disagreement(want, js); diff != "" {
		test.Fatalf("javascript: %s", diff)
	}
	for _, candidate := range []string{"native", "sanitized"} {
		if backend != "" && backend != candidate {
			continue
		}
		var got run
		if candidate == "native" {
			got = releasedUncached(test, program)
		} else {
			got, _ = nativelyUncached(test, program)
		}
		if diff := disagreement(want, got); diff != "" {
			test.Errorf("%s: %s; stderr %q", candidate, diff, got.stderr)
		}
	}
}
func TestFX7TextP09(test *testing.T) { test.Parallel(); fx7Text(test, "p09") }
func TestFX7TextP32(test *testing.T) { test.Parallel(); fx7Text(test, "p32") }
func TestFX7TextP33(test *testing.T) { test.Parallel(); fx7Text(test, "p33") }
func TestFX7TextP67(test *testing.T) { test.Parallel(); fx7Text(test, "p67") }
func TestFX7TextP72(test *testing.T) { test.Parallel(); fx7Text(test, "p72") }

func TestFX7TextScalarTuple(test *testing.T) { test.Parallel(); fx7Text(test, "scalar-tuple") }

func TestFX7TextScalarStatic(test *testing.T) {
	test.Parallel()
	fx7TextBackend(test, "scalar-static", "native")
}

func TestFX7TextScalarStaticSanitized(test *testing.T) {
	test.Parallel()
	fx7TextBackend(test, "scalar-static", "sanitized")
}
func TestFX7TextScalarStaticJavaScript(test *testing.T) {
	test.Parallel()
	fx7TextBackend(test, "scalar-static", "javascript")
}
