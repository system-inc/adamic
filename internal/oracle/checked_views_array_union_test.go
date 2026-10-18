package oracle

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Source Node controls retain the cast's erased behavior. The backends additionally
// reject arrays whose represented elements cannot inhabit any declared array arm.
type checkedViewArrayUnionProbe struct {
	name, cells, target, length string
	rejects                     bool
}

func checkedViewArrayUnionProbeCheck(t *testing.T, probe checkedViewArrayUnionProbe) {
	t.Helper()
	source := "interface Element{readonly label:string;} type Values=readonly " + probe.target + "[]|readonly string[]; interface Base{readonly kind:'Holder'} interface View extends Base{readonly values:Values;} " + probe.cells + " const raw={kind:'Holder' as const,values:cells};const base:Base=raw;const view=base as View;console.log(`${view.values.length}`);"
	path := filepath.Join(t.TempDir(), "array-union.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if diff := disagreement(run{stdout: []byte(probe.length + "\n")}, node); diff != "" {
		t.Fatal("Node: " + diff)
	}
	want := node
	if probe.rejects {
		want = run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: view.values matches no member of Values; expected Values, found array\n")}
	}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native-sanitized": sanitized, "native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s: %s; stdout %q stderr %q exit %d", backend, diff, got.stdout, got.stderr, got.exitCode)
		}
	}
}

func TestCheckedViewArrayUnionMembershipNumber(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"number", "const cells=[7];", "number", "1", false})
}

func TestCheckedViewArrayUnionMembershipString(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"string", "const cells=['ok'];", "number", "1", false})
}

func TestCheckedViewArrayUnionMembershipBoolean(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"boolean", "const cells=[true];", "number", "1", true})
}

func TestCheckedViewArrayUnionMembershipTupleObject(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"tuple-object", "const pair:readonly [number,number]=[1,2]; const cells=[pair];", "Element", "1", true})
}

func TestCheckedViewArrayUnionMembershipBoxedMixed(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"boxed-mixed", "const cells=[7,'ok'];", "number", "2", true})
}

func TestCheckedViewArrayUnionMembershipMaybeNumber(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"maybe-number", "const cells=[7,undefined];", "number", "2", true})
}

func TestCheckedViewArrayUnionMembershipBooleanGood(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"boolean-good", "const cells=[false,true];", "boolean", "2", false})
}

func TestCheckedViewArrayUnionMembershipSparseMax(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"sparse-max", "const cells=new Array<number>(4294967295); cells[4294967294]=7;", "number", "4294967295", false})
}

func TestCheckedViewArrayUnionMembershipSparseWrong(t *testing.T) {
	t.Parallel()
	checkedViewArrayUnionProbeCheck(t, checkedViewArrayUnionProbe{"sparse-wrong", "const cells=new Array<number>(4294967295); cells[4294967294]=7;", "boolean", "4294967295", true})
}

// The checker keeps narrowing across this call, but the held union has changed.
// Array.isArray must recheck it before an array representation is read.
func TestCheckedViewArrayUnionNarrowing(t *testing.T) {
	t.Parallel()
	source := "let member:string|readonly number[]|undefined=[7];\nfunction change():void {member='wrong';}\nif(typeof member==='object'){change();console.log(`${member[0] ?? -1}`);}"
	path := filepath.Join(t.TempDir(), "narrowing.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
	}
	node := onNode(t, path)
	if diff := disagreement(run{stdout: []byte("w\n")}, node); diff != "" {
		t.Fatal("Node: " + diff)
	}
	want := run{exitCode: 70, stderr: []byte("adamic: panic: union member where the checker narrowed it away: a call since the narrowing put it back\n")}
	sanitized, _ := nativelyUncached(t, program)
	for backend, got := range map[string]run{"native-sanitized": sanitized, "native": releasedUncached(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if diff := disagreement(want, got); diff != "" {
			t.Fatalf("%s: %s; stdout %q stderr %q exit %d", backend, diff, got.stdout, got.stderr, got.exitCode)
		}
	}
}
