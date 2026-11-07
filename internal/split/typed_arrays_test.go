package split_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/split"
)

func typedArrayDecision(test *testing.T, source string) split.Function {
	test.Helper()
	path := filepath.Join(test.TempDir(), "typed.a")
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		test.Fatal(err)
	}
	decisions := split.Analyze(lowered(test, path))
	for _, decision := range decisions {
		if decision.Name == "tested" {
			return decision
		}
	}
	test.Fatal("tested function missing")
	return split.Function{}
}

func TestTypedArraySieveIsWasmEligible(test *testing.T) {
	test.Parallel()
	program := lowered(test, "../../workers/compute/handler.a")
	var reads, writes bool
	// The real sieve is the regression fixture, with a scalar crossing signature.
	for _, decision := range split.Analyze(program) {
		if decision.Name == "primes" {
			if !decision.Eligible || !decision.Entry {
				test.Fatalf("primes: %+v", decision)
			}
			function := program.Functions[decision.Index]
			// Lowering represents element access using the existing array operations.
			inspectTypedIndices(function.Body, &reads, &writes)
			if !reads || !writes {
				test.Fatalf("sieve lacks ArrayIndex/SetIndex: read=%t write=%t", reads, writes)
			}
			return
		}
	}
	test.Fatal("primes missing")
}

// inspectTypedIndices checks the sieve's structured loops for the existing index nodes.
func inspectTypedIndices(body []ir.Statement, reads, writes *bool) {
	for _, statement := range body {
		switch node := statement.(type) {
		case ir.SetIndex:
			if node.Array.Type().IsTypedArray() {
				*writes = true
			}
		case ir.If:
			if binary, ok := node.Condition.(ir.Binary); ok {
				if index, ok := binary.Left.(ir.ArrayIndex); ok && index.Array.Type().IsTypedArray() {
					*reads = true
				}
			}
			inspectTypedIndices(node.Then, reads, writes)
			inspectTypedIndices(node.Else, reads, writes)
		case ir.Loop:
			inspectTypedIndices(node.Body, reads, writes)
		case ir.Block:
			inspectTypedIndices(node.Body, reads, writes)
		}
	}
}

func TestTypedArrayCopiesDoNotAliasSource(test *testing.T) {
	test.Parallel()
	decision := typedArrayDecision(test, `export function tested(source: readonly number[]): number {
 const copy = new Uint8Array(source);
 for (let i=0; i<copy.length; i+=1) copy[i]=7;
 return copy[0] ?? 0;
}`)
	if !decision.Eligible {
		test.Fatalf("fresh scalar copy aliases source: %+v", decision)
	}
}
func TestTypedArrayParameterWritesAreImpure(test *testing.T) {
	test.Parallel()
	decision := typedArrayDecision(test, `export function tested(values: Uint8Array): number {
 for (let i=0; i<values.length; i+=1) values[i]=7;
 return values[0] ?? 0;
}`)
	if decision.Eligible || decision.Reason != "writes parameter 1" {
		test.Fatalf("parameter mutation: %+v", decision)
	}
}
func TestTypedArrayUnreviewedOperationsStayImpure(test *testing.T) {
	test.Parallel()
	for _, row := range []struct{ operation, expression string }{
		{"TypedArrayFill", "values.fill(1)"}, {"TypedArraySubarray", "values.subarray(0)"},
		{"TypedArraySort", "values.sort()"}, {"TypedArraySet", "values.set(new Uint8Array(2))"},
	} {
		test.Run(row.operation, func(test *testing.T) {
			decision := typedArrayDecision(
				test,
				`export function tested(n: number): number { const values = new Uint8Array(n); `+
					row.expression+
					`; for(let i=0;i<n;i+=1){} return n; }`,
			)
			if decision.Eligible || !strings.Contains(decision.Reason, "unknown operation "+row.operation) {
				test.Fatalf("unreviewed operation: %+v", decision)
			}
		})
	}
}
func TestTypedArrayCrossingsStayIneligible(test *testing.T) {
	test.Parallel()
	for _, row := range []struct{ source, reason string }{
		{
			`export function tested(values: Uint8Array): number { let sum=0; for(let i=0;i<values.length;i+=1) sum += values[i] ?? 0; return sum; }`,
			"signature: typed-array parameter 1 not crossable",
		},
		{
			`export function tested(n: number): Uint8Array { const values=new Uint8Array(n); for(let i=0;i<n;i+=1) values[i]=1; return values; }`,
			"signature: typed-array return not crossable",
		},
	} {
		decision := typedArrayDecision(test, row.source)
		if decision.Eligible || decision.Reason != row.reason {
			test.Fatalf("typed crossing: %+v", decision)
		}
	}
}
