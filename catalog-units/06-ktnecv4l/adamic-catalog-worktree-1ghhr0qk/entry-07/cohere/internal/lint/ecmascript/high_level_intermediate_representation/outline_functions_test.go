package high_level_intermediate_representation

import "testing"

// The name an outlined function takes must resolve back to that function.
//
// This is the invariant the `Outlined` map cannot hold on its own. That map is keyed by the
// identifier the `LoadGlobal` was written into, and a second `Construct` renumbers every identifier
// in the function -- which inlining an immediately invoked function expression forces. The key then
// names a value that no longer exists while the map still reports a hit for the old one, so the
// lookup fails silently and a callback resolver finds nothing.
//
// Measured on `error.validate-object-values-mutation`: with the map alone and the memo callback
// inlined, `values.map(cb)` stopped resolving `cb`, so it emitted no mutation for the element it is
// called over, `object`'s mutable range collapsed from `{3,20}` to `{3,4}`, and the scope no longer
// spanned the mutation the golden is about. Two goldens turned silent for that reason alone.
//
// The name is the durable half: it is synthesized from the function id, nothing rewrites it, and it
// survives renumbering. Pinning the round trip here keeps the two halves from drifting apart.
func TestOutlinedFunctionNameRoundTrip(t *testing.T) {
	t.Parallel()

	function := &Function{Functions: make([]*Function, 8)}
	for id := FunctionId(0); int(id) < len(function.Functions); id++ {
		name := outlinedFunctionName(id)
		resolved, found := outlinedFunctionByName(function, name)
		if !found {
			t.Fatalf("outlinedFunctionName(%d) = %q did not resolve", id, name)
		}
		if resolved != id {
			t.Fatalf("outlinedFunctionName(%d) = %q resolved to %d", id, name, resolved)
		}
	}
}

// A name that is not an outlined function's must not resolve to one.
//
// The resolver runs over every `LoadGlobal` in the function, and most of them name real globals:
// `Object`, `Stringify`, `useMemo`. Answering for one of those would attribute a callback's effects
// to whatever function happened to sit at that index.
func TestOutlinedFunctionByNameDeclinesOrdinaryGlobals(t *testing.T) {
	t.Parallel()

	function := &Function{Functions: make([]*Function, 4)}
	for _, name := range []string{"Object", "Stringify", "useMemo", "", "_tempest", "_temp0",
		"_temp-1", "_temp99"} {
		if _, found := outlinedFunctionByName(function, name); found {
			t.Errorf("outlinedFunctionByName(%q) resolved, want declined", name)
		}
	}
}
