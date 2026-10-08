//go:build lintoracle

package high_level_intermediate_representation

import (
	"bytes"
	"fmt"
	"sort"
)

// unit3Encode is supplied by the replay owner. It must encode all graph tables,
// orphan instructions, nesting identities and AST/memo facts, not a construction
// dump alone. The adapter does not choose a pipeline or synthesize missing input.
type unit3Encode func(*Function) ([]byte, error)

type unit3Checkpoint struct {
	Pass                      string
	Before, After             []byte
	NestedBefore, NestedAfter []byte
	Result                    any
}

// unit3Observe captures exactly one Go entry point. The caller supplies the real
// prepass state, including Go's reactive facts at preservation's drop boundary.
// Encoder errors propagate; they must remain outcomes in the census manifest.
func unit3Observe(pass string, function, nested *Function, captures []Place, encode unit3Encode) (unit3Checkpoint, error) {
	checkpoint := unit3Checkpoint{Pass: pass}
	switch pass {
	case "drop_manual_memoization", "inline_iife", "inline_iife_including_memo_callbacks", "inline_remap", "invoked_functions", "outline_functions", "dead_code_elimination", "merge_consecutive_blocks":
	default:
		return checkpoint, fmt.Errorf("unknown unit-3 pass %q", pass)
	}
	if function == nil || encode == nil {
		return checkpoint, fmt.Errorf("%s: missing graph or checkpoint encoder", pass)
	}
	snapshot := func(value *Function) ([]byte, error) {
		encoded, err := encode(value)
		if err != nil {
			return nil, err
		}
		return bytes.Clone(encoded), nil
	}
	var err error
	checkpoint.Before, err = snapshot(function)
	if err != nil {
		return checkpoint, fmt.Errorf("%s before: %w", pass, err)
	}
	if pass == "inline_remap" {
		if nested == nil {
			return checkpoint, fmt.Errorf("inline_remap: missing nested input")
		}
		checkpoint.NestedBefore, err = snapshot(nested)
		if err != nil {
			return checkpoint, fmt.Errorf("%s nested before: %w", pass, err)
		}
	}
	switch pass {
	case "drop_manual_memoization":
		checkpoint.Result = DropManualMemoization(function)
	case "inline_iife":
		checkpoint.Result = InlineImmediatelyInvokedFunctionExpressions(function)
	case "inline_iife_including_memo_callbacks":
		checkpoint.Result = InlineImmediatelyInvokedFunctionExpressionsIncludingMemoCallbacks(function)
	case "inline_remap":
		remap, copied := CopyNestedBodyInto(function, nested, captures)
		checkpoint.Result = struct {
			Remap  *InlineRemap
			Copied bool
		}{remap, copied}
	case "invoked_functions":
		// Go's map is a set. Sort numeric function identities before framing it.
		invoked := CollectAssumedInvokedFunctions(function)
		ids := make([]FunctionId, 0, len(invoked))
		for id, yes := range invoked {
			if yes {
				ids = append(ids, id)
			}
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		checkpoint.Result = ids
	case "outline_functions":
		checkpoint.Result = OutlineFunctions(function)
	case "dead_code_elimination":
		checkpoint.Result = EliminateDeadCode(function)
	case "merge_consecutive_blocks":
		checkpoint.Result = MergeConsecutiveBlocks(function)
	}
	checkpoint.After, err = snapshot(function)
	if err != nil {
		return checkpoint, fmt.Errorf("%s after: %w", pass, err)
	}
	if pass == "inline_remap" {
		checkpoint.NestedAfter, err = snapshot(nested)
		if err != nil {
			return checkpoint, fmt.Errorf("%s nested after: %w", pass, err)
		}
	}
	return checkpoint, nil
}
