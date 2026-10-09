package estree

import ("fmt"; "testing")

const testRecoveryMutantsShards = 3

var recoveryMutations = []struct{ name, file, from, to string }{
	{"first-accessibility", "modifiers.ts", "return stringValue(kind.slice(0, -7).toLowerCase());", "return stringValue('public');"},
	{"empty-type-list-range", "typeLists.ts", "arena.node(result).set('params', listValue([]));", "arena.node(result).end -= 1; arena.node(result).set('params', listValue([]));"},
	{"module-await", "pipeline.ts", "&& externalModule(parser.nodes, root)", "&& false && externalModule(parser.nodes, root)"},
}

func recoveryMutantCases() []recoveryCase {
	var cases []recoveryCase
	// Interleave mutants so partition i%3 assigns one complete corpus per mutant.
	for i, source := range recoveredGrammar() {
		for _, m := range recoveryMutations {
			cases = append(cases, recoveryCase{fmt.Sprintf("%s/%03d", m.name, i), source})
		}
	}
	return cases
}

// ADAMIC_TEST_SHARD=i/n runs shards whose index modulo n is i; unset runs all.
// Each mutant retains the entire grammar corpus on both Node and sanitized native.
func TestRecoveryMutants(t *testing.T) {
	t.Parallel()
	// Compatibility enumeration only; execution lives in the top-level shards.
	checkRecoveryMutantUnion(t)
}

func TestRecoveryMutantsShardSurvivor(t *testing.T) {
	t.Parallel()
	proveRecoveryShard(t, recoveryMutantCases(), testRecoveryMutantsShards, true)
}

