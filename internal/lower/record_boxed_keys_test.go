package lower

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

func TestRecordBoxedKeyEvaluatesOnce(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/record_boxed_keys/p3-once.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSourceAs(t, string(source), ".ts")
	if err != nil {
		t.Fatal(err)
	}
	main := fmt.Sprintf("%#v", program.Main)
	if calls := strings.Count(main, "ir.Call{"); calls != 1 {
		t.Fatalf("main should call the key producer once, has %d calls: %s", calls, main)
	}
	if conversions := strings.Count(main, "ir.NumberToString{Value:ir.Narrow{"); conversions != 1 {
		t.Fatalf("main should have one numeric key conversion, has %d: %s", conversions, main)
	}
}

func TestRecordBoxedKeyKeepsNonReadStops(t *testing.T) {
	t.Parallel()
	const recordStop = "a record key other than string or number"
	for _, probe := range []struct{ name, reason string }{
		{"p5-map", "a Map whose keys aren't strings, numbers, booleans, objects, arrays, maps or functions"},
		{"p6-write", recordStop},
		{"p7-in", recordStop},
		{"p7-has-own", recordStop},
		{"w01-nullish-assign", recordStop},
		{"w02-delete", recordStop},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			source, err := os.ReadFile(filepath.Join("../oracle/testdata/record_boxed_keys", probe.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowerSourceAs(t, string(source), ".ts")
			var stop *NotYet
			if !errors.As(err, &stop) || stop.What != probe.reason {
				t.Fatalf("want the original stop %q, got %v", probe.reason, err)
			}
		})
	}
}

func TestRecordBoxedKeyChecksEveryArm(t *testing.T) {
	t.Parallel()
	// TypeScript rejects these arms at element access before lowering, so read
	// the live parameter types from a valid program and exercise the key gate.
	path := writeAgreementSource(t, "main.ts", `
function keys(
    mixed: string | number, literals: "a" | 1, absent: string | number | undefined,
    flag: string | number | boolean, nilKey: string | number | null,
    objectKey: string | number | object, symbolKey: string | number | symbol
): void {}
`)
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	file := program.Files()[0]
	checked, release := program.Checker(context.Background(), file)
	defer release()
	for _, param := range file.Statements.Nodes[0].AsFunctionDeclaration().Parameters.Nodes {
		name := param.Name()
		want := name.Text() == "mixed" || name.Text() == "literals"
		if got := recordBoxedKeyType(checked.GetTypeAtLocation(name)); got != want {
			t.Errorf("%s: admitted %t, want %t", name.Text(), got, want)
		}
	}
}

func TestRecordBoxedKeyChecksStringNarrowing(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/record_boxed_keys/s01-narrowed-local.ts")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSourceAs(t, string(source), ".ts")
	if err != nil {
		t.Fatal(err)
	}
	const message = "union member where the checker narrowed it away: a call since the narrowing put it back"
	want := fmt.Sprintf("To:%d, Checked:true, Message:%q", ir.String, message)
	if !strings.Contains(fmt.Sprintf("%#v", program.Main), want) {
		t.Fatalf("expected checked string narrowing in the record key: %#v", program.Main)
	}
}

func TestRecordBoxedKeyRejectsIntersectionArm(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/record_boxed_keys/w03-intersection-arm.ts")
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowerSourceAs(t, string(source), ".ts")
	var stop *NotYet
	if !errors.As(err, &stop) || stop.What != "a record key other than string or number" || !strings.HasSuffix(stop.Where, ":2:78") {
		t.Fatalf("want the original record key stop at the index expression, got %v", err)
	}
}
