package high_level_intermediate_representation

import "testing"

// TestMemoizationLevelOfCoversEveryInstructionValue is the guard that makes the default safe.
//
// `MemoizationLevelOf` falls through to `MemoizationNever`, which is correct for upstream's
// `UnsupportedNode` and wrong for anything else. A new instruction value added to this package would
// silently take that default and be classified as never worth memoizing -- a value that escapes and
// is not held, which is exactly the bug the pass exists to prevent.
//
// This package already discovers the closed set by scanning source for the marker method, so the
// guard costs nothing beyond naming the one kind that is legitimately in the default.
func TestMemoizationLevelOfCoversEveryInstructionValue(t *testing.T) {
	t.Parallel()

	// The only value upstream classifies as Never. Anything else reaching the default is a gap.
	legitimateDefaults := map[string]string{
		"UnsupportedNode": "upstream's only Never arm: a node the compiler does not model",
	}

	declared := markerImplementers(t, "instructionValue")
	if len(declared) < 40 {
		t.Fatalf("found only %d instruction values, expected the full set; the source scan is broken",
			len(declared))
	}
	covered := typeSwitchCases(t, "memoization_inputs.go", "MemoizationLevelOf")

	uncovered := 0
	for _, name := range declared {
		if covered[name] {
			continue
		}
		if reason, known := legitimateDefaults[name]; known {
			t.Logf("%s falls through to Never on purpose: %s", name, reason)
			continue
		}
		uncovered++
		t.Errorf("MemoizationLevelOf has no case for %s, so it silently classifies as Never -- a "+
			"value that escapes and is not held", name)
	}

	// The count is asserted so that adding a kind to `legitimateDefaults` is a deliberate act. A
	// second entry there would mean someone decided another kind is never worth memoizing, and that
	// is a decision upstream did not make.
	if len(legitimateDefaults) != 1 {
		t.Errorf("%d kinds are excused from coverage; upstream excuses exactly one",
			len(legitimateDefaults))
	}
	if uncovered != 0 {
		t.Logf("%d instruction values are unclassified", uncovered)
	}
}

// TestMemoizationLevelOfMatchesUpstreamGroups pins the classification against upstream's own arms.
//
// Transcribed from `computeMemoizationInputs` rather than restated, including the two arms whose
// answer depends on an option: JSX under `memoizeJsxElements` and the thirteen-kind primitive group
// under `forceMemoizePrimitives`, both of which resolve true under the schema defaults.
//
// The primitive group is the one most likely to be got wrong, because `Conditional` is its forced
// branch and `Never` is its default one. A port reading the first level in the arm gets the right
// answer for the wrong reason and would flip the day the flag default changes.
func TestMemoizationLevelOfMatchesUpstreamGroups(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name  string
		want  MemoizationLevel
		value InstructionValue
	}{
		{name: "array literal allocates", want: MemoizationMemoized, value: &ArrayExpression{}},
		{name: "object literal allocates", want: MemoizationMemoized, value: &ObjectExpression{}},
		{name: "function expression allocates", want: MemoizationMemoized, value: &FunctionExpression{}},
		{name: "new allocates", want: MemoizationMemoized, value: &NewExpression{}},
		{name: "regexp literal allocates", want: MemoizationMemoized, value: &RegExpLiteral{}},
		{name: "property store", want: MemoizationMemoized, value: &PropertyStore{}},
		{name: "call", want: MemoizationMemoized, value: &CallExpression{}},
		{name: "method call", want: MemoizationMemoized, value: &MethodCall{}},
		{name: "tagged template", want: MemoizationMemoized, value: &TaggedTemplateExpression{}},
		{name: "jsx under memoizeJsxElements", want: MemoizationMemoized, value: &JsxExpression{}},
		{name: "jsx fragment under memoizeJsxElements", want: MemoizationMemoized, value: &JsxFragment{}},
		{name: "declare context", want: MemoizationMemoized, value: &DeclareContext{}},
		{name: "store context", want: MemoizationMemoized, value: &StoreContext{}},

		{name: "primitive under forceMemoizePrimitives", want: MemoizationConditional, value: &Primitive{}},
		{name: "binary under forceMemoizePrimitives", want: MemoizationConditional, value: &BinaryExpression{}},
		{name: "unary under forceMemoizePrimitives", want: MemoizationConditional, value: &UnaryExpression{}},
		{name: "template literal under forceMemoizePrimitives", want: MemoizationConditional, value: &TemplateLiteral{}},
		{name: "load global under forceMemoizePrimitives", want: MemoizationConditional, value: &LoadGlobal{}},
		{name: "jsx text under forceMemoizePrimitives", want: MemoizationConditional, value: &JsxText{}},

		{name: "load local propagates", want: MemoizationConditional, value: &LoadLocal{}},
		{name: "property load propagates", want: MemoizationConditional, value: &PropertyLoad{}},
		{name: "store local propagates", want: MemoizationConditional, value: &StoreLocal{}},
		{name: "destructure propagates", want: MemoizationConditional, value: &Destructure{}},
		{name: "await propagates", want: MemoizationConditional, value: &Await{}},
		{name: "type cast propagates", want: MemoizationConditional, value: &TypeCastExpression{}},

		{name: "declare local", want: MemoizationUnmemoized, value: &DeclareLocal{}},
		{name: "store global", want: MemoizationUnmemoized, value: &StoreGlobal{}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := MemoizationLevelOf(testCase.value); got != testCase.want {
				t.Errorf("got %s, want %s", got, testCase.want)
			}
		})
	}
}

// TestMemoizationLevelOfReactiveValueHandlesComposites covers the four kinds with no instruction
// value.
//
// They are terminals in the graph and values in the tree, so they have no `InstructionValue` to
// classify. Upstream gives all four the same answer in separate arms, which matches what they are:
// none produces an identity of its own.
func TestMemoizationLevelOfReactiveValueHandlesComposites(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name  string
		want  MemoizationLevel
		value ReactiveValue
	}{
		{name: "logical", want: MemoizationConditional, value: &ReactiveLogicalValue{}},
		{name: "ternary", want: MemoizationConditional, value: &ReactiveTernaryValue{}},
		{name: "sequence", want: MemoizationConditional, value: &ReactiveSequenceValue{}},
		{name: "optional", want: MemoizationConditional, value: &ReactiveOptionalValue{}},
		{
			name:  "a wrapped instruction value defers to the instruction classification",
			want:  MemoizationMemoized,
			value: &ReactiveInstructionValue{Value: &ArrayExpression{}},
		},
		{
			name:  "a wrapped nil is never",
			want:  MemoizationNever,
			value: &ReactiveInstructionValue{},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			if got := MemoizationLevelOfReactiveValue(testCase.value); got != testCase.want {
				t.Errorf("got %s, want %s", got, testCase.want)
			}
		})
	}

	if got := MemoizationLevelOfReactiveValue(nil); got != MemoizationNever {
		t.Errorf("a nil value answered %s, want never", got)
	}
}

// TestMemoizationInputsGapsAreDeclared pins the gap list so closing one is a visible event.
func TestMemoizationInputsGapsAreDeclared(t *testing.T) {
	t.Parallel()

	gaps := MemoizationInputsGaps()
	if len(gaps) != 2 {
		t.Fatalf("expected 2 declared gaps, found %d; a gap was added or closed without this test "+
			"and its consumers being updated", len(gaps))
	}
	seen := map[MemoizationInputsGap]bool{}
	for _, gap := range gaps {
		if seen[gap] {
			t.Errorf("gap %v is listed twice", gap)
		}
		seen[gap] = true
	}
	for _, required := range []MemoizationInputsGap{
		MemoizationInputsGapCallSignatures,
		MemoizationInputsGapOperandMutability,
	} {
		if !seen[required] {
			t.Errorf("gap %v is declared as a constant but not returned, so nothing asserts on it",
				required)
		}
	}
}
