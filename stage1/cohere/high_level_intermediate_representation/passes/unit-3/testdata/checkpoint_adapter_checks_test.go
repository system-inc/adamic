//go:build lintoracle

package high_level_intermediate_representation

import (
	"bytes"
	"errors"
	"fmt"
	"testing"
)

func TestUnit3AdapterBoundaries(t *testing.T) {
	t.Parallel()
	for _, pass := range []string{"drop_manual_memoization", "inline_iife", "inline_iife_including_memo_callbacks", "inline_remap", "invoked_functions", "outline_functions", "dead_code_elimination", "merge_consecutive_blocks"} {
		t.Run(pass, func(t *testing.T) {
			t.Parallel()
			function := NewFunction(nil, "Widget", FunctionKindComponent)
			nested := NewFunction(nil, "", FunctionKindOther)
			calls := 0
			encode := func(*Function) ([]byte, error) { calls++; return []byte(fmt.Sprint(calls)), nil }
			checkpoint, err := unit3Observe(pass, function, nested, nil, encode)
			if err != nil {
				t.Fatal(err)
			}
			wantCalls := 2
			wantAfter := "2"
			if pass == "inline_remap" {
				wantCalls = 4
				wantAfter = "3"
			}
			if calls != wantCalls || string(checkpoint.Before) != "1" || string(checkpoint.After) != wantAfter {
				t.Fatalf("bad boundary capture: %+v, calls=%d", checkpoint, calls)
			}
			if pass == "inline_remap" && (string(checkpoint.NestedBefore) != "2" || string(checkpoint.NestedAfter) != "4") {
				t.Fatal("nested input not captured on both sides")
			}
		})
	}
}

func TestUnit3AdapterOwnsSnapshotsAndRetainsFailures(t *testing.T) {
	t.Parallel()
	function := NewFunction(nil, "Widget", FunctionKindComponent)
	buffer := []byte("before")
	calls := 0
	checkpoint, err := unit3Observe("dead_code_elimination", function, nil, nil, func(*Function) ([]byte, error) {
		calls++
		if calls == 2 {
			copy(buffer, "after!")
		}
		return buffer, nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(checkpoint.Before, []byte("before")) || !bytes.Equal(checkpoint.After, []byte("after!")) {
		t.Fatal("encoder reuse corrupted a snapshot")
	}
	failure := errors.New("missing source fact")
	calls = 0
	_, err = unit3Observe("dead_code_elimination", function, nil, nil, func(*Function) ([]byte, error) { calls++; return nil, failure })
	if !errors.Is(err, failure) || calls != 1 {
		t.Fatal("before-state error was lost or pass continued")
	}
	calls = 0
	checkpoint, err = unit3Observe("dead_code_elimination", function, nil, nil, func(*Function) ([]byte, error) {
		calls++
		if calls == 2 {
			return nil, failure
		}
		return []byte("before"), nil
	})
	if !errors.Is(err, failure) || string(checkpoint.Before) != "before" || checkpoint.Result == nil {
		t.Fatal("after-state error lost its recorded input or Go result")
	}
}

func TestUnit3AdapterCapturesActualRewrite(t *testing.T) {
	t.Parallel()
	function := NewFunction(nil, "Widget", FunctionKindComponent)
	block := function.NewBlock(BlockKindBlock)
	function.Entry = block.Id
	block.Terminal = &Unreachable{}
	value := function.NewIdentifier("", nil, 0)
	function.AddInstruction(block, &Instruction{LValue: Place{Identifier: value.Id}, Value: &Primitive{}})
	checkpoint, err := unit3Observe("dead_code_elimination", function, nil, nil, func(f *Function) ([]byte, error) {
		return []byte(fmt.Sprintf("%d/%d", len(f.Blocks[0].Instructions), len(f.Instructions))), nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if string(checkpoint.Before) != "1/1" || string(checkpoint.After) != "0/1" {
		t.Fatalf("wrong actual Go before/after: %+v", checkpoint)
	}
	result, ok := checkpoint.Result.(DeadCodeEliminationResult)
	if !ok || result.Instructions != 1 || result.Phis != 0 {
		t.Fatalf("lost Go rewrite counts: %+v", checkpoint.Result)
	}
}
