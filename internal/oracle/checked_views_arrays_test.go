package oracle

import "testing"

type checkedViewArrayProbe struct{ name, stdout, diagnostic string }

func checkedViewArrayProbeCheck(t *testing.T, probe checkedViewArrayProbe) {
	t.Helper()
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
}

func TestCheckedViewArraysArrayHolesView(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-holes-view", "-1\n7\n", ""})
}

func TestCheckedViewArraysArrayLazy(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-lazy", "2\n", ""})
}

func TestCheckedViewArraysArrayIteration(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-iteration", "7\n8\n", ""})
}

func TestCheckedViewArraysArrayMap(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-map", "8:9\n", ""})
}

func TestCheckedViewArraysArrayPop(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-pop", "8:7\n", ""})
}

func TestCheckedViewArraysArrayBounds(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-bounds", "-1:7\n", ""})
}

func TestCheckedViewArraysArraySecond(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-second", "", "element read failed: values[1] expected number, found string"})
}

func TestCheckedViewArraysArrayLengthKind(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-length-kind", "", "field read failed: items(raw).values is not a readonly number[]; expected readonly number[], found object"})
}

func TestCheckedViewArraysArrayMissing(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-missing", "", "field read failed: items(raw).values is not initialized; expected readonly number[], found missing"})
}

func TestCheckedViewArraysArrayBoolean(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-boolean", "", "element read failed: values[0] expected number, found boolean"})
}

func TestCheckedViewArraysArrayIterationBad(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-iteration-bad", "", "element read failed: items(raw).values[element] expected number, found string"})
}

func TestCheckedViewArraysArrayMapBad(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-map-bad", "", "element read failed: items(raw).values[element] expected number, found string"})
}

func TestCheckedViewArraysArrayAlias(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-alias", "ok\n", "field read failed: values[0] expected \"ok\", found string changed"})
}

func TestCheckedViewArraysArrayUndefined(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-undefined", "", "element read failed: items(raw).values[1] expected number, found undefined"})
}

func TestCheckedViewArraysArrayStringLiteral(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-string-literal", "", "field read failed: values[element] expected \"ok\", found string bad"})
}

func TestCheckedViewArraysNullableArrayString(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"nullable-array-string", "", "element read failed: values[element] expected number, found string"})
}

func TestCheckedViewArraysNullableArrayMissing(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"nullable-array-missing", "", "element read failed: values[0] expected number, found undefined"})
}

func TestCheckedViewArraysNullableArrayElement(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"nullable-array-element", "", "element read failed: values[0] expected number, found string"})
}

func TestCheckedViewArraysGenericArrayElement(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"generic-array-element", "", "element read failed: values[0] expected string, found number"})
}

func TestCheckedViewArraysArrayField(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-field", "2\n", ""})
}

func TestCheckedViewArraysArrayElement(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"array-element", "7\n", ""})
}

func TestCheckedViewArraysObjectElement(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"object-element", "ok\n", ""})
}

func TestCheckedViewArraysNonArray(t *testing.T) {
	t.Parallel()
	checkedViewArrayProbeCheck(t, checkedViewArrayProbe{"non-array", "", "field read failed: items(raw).values is not a readonly number[]; expected readonly number[], found number"})
}
