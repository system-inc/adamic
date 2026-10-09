package oracle

import (
	"testing"
)

func fx7Depth(t *testing.T, name string) {
	t.Helper()
	program, path := interfaceFixture(t, "fx7/"+name)
	want := onNode(t, path)
	if want.exitCode != 0 {
		t.Fatalf("Node: %#v", want)
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized, "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s; stderr %q", backend, diff, got.stderr)
		}
	}
}
func TestFX7Depth65(t *testing.T)   { t.Parallel(); fx7Depth(t, "depth-65") }
func TestFX7Depth200(t *testing.T)  { t.Parallel(); fx7Depth(t, "depth-200") }
func TestFX7Depth2000(t *testing.T) { t.Parallel(); fx7Depth(t, "depth-2000") }

func init() { additionalFixtureCounts = append(additionalFixtureCounts, fx7Counts) }
func fx7Counts(t *testing.T) []string {
	t.Helper()
	rows := []string{}
	for _, name := range []string{"depth-65", "depth-200", "depth-2000", "p09", "p32", "p33", "p67", "p72", "scalar-tuple"} {
		rows = append(rows, counted(t, "stage3/interface-downcasts/fx7/"+name+".a", false, nil, false, false))
	}
	return rows
}

func fx7Text(t *testing.T, name string) {
	t.Helper()
	program, _ := interfaceFixture(t, "fx7/"+name)
	js := onJavaScriptBackend(t, program)
	t.Logf("JavaScript: exit %d stdout %q stderr %q", js.exitCode, js.stdout, js.stderr)
	want := run{exitCode: 70}
	switch name {
	case "p32", "p33":
		want.stdout = []byte("3\n")
		want.stderr = []byte("adamic: panic: field read failed: <write>.name is not a string | number; expected string | number, found string\n")
	case "p67":
		want.stderr = []byte("adamic: panic: field read failed: viewed.value matches no member of number | Inner; expected number | Inner, found function\n")
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
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native": releasedUncached(t, program), "sanitized": sanitized} {
		if diff := disagreement(want, got); diff != "" {
			t.Errorf("%s: %s; stderr %q", backend, diff, got.stderr)
		}
	}
}
func TestFX7TextP09(t *testing.T) { t.Parallel(); fx7Text(t, "p09") }
func TestFX7TextP32(t *testing.T) { t.Parallel(); fx7Text(t, "p32") }
func TestFX7TextP33(t *testing.T) { t.Parallel(); fx7Text(t, "p33") }
func TestFX7TextP67(t *testing.T) { t.Parallel(); fx7Text(t, "p67") }
func TestFX7TextP72(t *testing.T) { t.Parallel(); fx7Text(t, "p72") }

func TestFX7TextScalarTuple(t *testing.T) { t.Parallel(); fx7Text(t, "scalar-tuple") }
