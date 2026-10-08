package oracle

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestCheckedViewObjectPrimitiveFixes(t *testing.T) {
	t.Run("graph", func(t *testing.T) {
		path, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4b/fixtures/fixes-graph.a"))
		if err != nil {
			t.Fatal(err)
		}
		want := run{stdout: []byte("42\n")}
		if diff := disagreement(want, onNode(t, path)); diff != "" {
			t.Fatal(diff)
		}
		loaded, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			t.Fatal(err)
		}
		if len(program.GraphTypes) == 0 {
			t.Fatal("witness must exercise graph adoption")
		}
		sanitized, binary := nativelyUncached(t, program)
		for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("%s; stderr %q", diff, got.stderr)
			}
		}
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	})

	t.Run("tuple-unread", func(t *testing.T) {
		path, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4b/fixtures/fixes-tuple-unread.a"))
		if err != nil {
			t.Fatal(err)
		}
		want := run{stdout: []byte("admitted\n")}
		if diff := disagreement(want, onNode(t, path)); diff != "" {
			t.Fatal(diff)
		}
		loaded, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			t.Fatal(err)
		}
		sanitized, binary := nativelyUncached(t, program)
		for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if diff := disagreement(want, got); diff != "" {
				t.Fatal(diff)
			}
		}
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	})
	t.Run("tuple", func(t *testing.T) {
		path, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4b/fixtures/fixes-tuple.a"))
		if err != nil {
			t.Fatal(err)
		}
		if diff := disagreement(run{stdout: []byte("string\n")}, onNode(t, path)); diff != "" {
			t.Fatal(diff)
		}
		loaded, err := load.Load([]string{path})
		if err != nil {
			t.Fatal(err)
		}
		program, err := lower.Lower(context.Background(), loaded)
		if err != nil {
			t.Fatal(err)
		}
		// Integration now supplies a checked positional contract for this tuple.
		// Retain both positions even when this producer selects the string arm.
		complete := false
		for _, contract := range program.ViewContracts {
			if !contract.FixedTuple || len(contract.Tuple) != 2 || contract.Unsupported != "" {
				continue
			}
			complete = true
			for _, id := range contract.Tuple {
				child := program.ViewContracts[id-1]
				if child.Kind != ir.ViewScalar || child.Of != ir.Number {
					t.Fatal("tuple position lost its number contract")
				}
			}
		}
		if !complete {
			t.Fatal("supported tuple alternative lost its complete positional contract")
		}
		actual, binary := nativelyUncached(t, program)
		for _, got := range []run{actual, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if diff := disagreement(run{stdout: []byte("string\n")}, got); diff != "" {
				t.Fatal(diff)
			}
		}
		if report := leaksUncached(t, program, binary); report != "" {
			t.Fatal(report)
		}
	})
	t.Run("long-message", func(t *testing.T) {
		path, err := filepath.Abs(checkedViewFixturePath("../../stage3/interface-downcasts/lane4b/fixtures/diagnostic-code-wrong.a"))
		if err != nil {
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
		declared := strings.Repeat("LongDeclaredType", 30)
		if count := changeObjectPrimitiveRead(program, func(read ir.Property) bool { return read.View == "member.code" }, func(read ir.Property) ir.Property { read.ViewType = declared; return read }); count != 1 {
			t.Fatalf("want one named scalar read, got %d", count)
		}
		want := run{exitCode: 70, stderr: []byte("adamic: panic: field read failed: member.code is not a " + declared + "; expected " + declared + ", found boolean\n")}
		sanitized, _ := nativelyUncached(t, program)
		for _, got := range []run{sanitized, releasedUncached(t, program), onJavaScriptBackend(t, program)} {
			if diff := disagreement(want, got); diff != "" {
				t.Fatalf("full diagnostic: %s; stderr %q", diff, got.stderr)
			}
		}
	})
}
