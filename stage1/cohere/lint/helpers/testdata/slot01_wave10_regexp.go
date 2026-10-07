package regexp

var adamicWave10Calls []rewriteOptions

func adamicWave10WordCharacters(options rewriteOptions) string {
	adamicWave10Calls = append(adamicWave10Calls, options)
	return wordCharacters(options)
}
func AdamicWave10Atoms(bits int, mode string) [][][]any {
	options := rewriteOptions{ignoreCase: bits&1 != 0, unicode: bits&2 != 0, multiline: bits&4 != 0, dotAll: bits&8 != 0}
	call := func() []classAtom {
		if mode == "word" {
			return wordClassAtoms(options)
		}
		return nonWordClassAtoms(options)
	}
	rows := func(atoms []classAtom) [][]any {
		out := [][]any{}
		for _, a := range atoms {
			out = append(out, []any{int(a.kind), int(a.lo), int(a.hi), a.text})
		}
		return out
	}
	first := call()
	snapshots := [][][]any{rows(first)}
	first[0].lo = 99999
	snapshots = append(snapshots, rows(call()))
	return snapshots
}
func AdamicWave10Boundary(bits int, negated bool) (string, string, []int) {
	options := rewriteOptions{ignoreCase: bits&1 != 0, unicode: bits&2 != 0, multiline: bits&4 != 0, dotAll: bits&8 != 0}
	adamicWave10Calls = nil
	result := wordBoundary(negated, options)
	calls := []int{}
	for _, o := range adamicWave10Calls {
		b := 0
		if o.ignoreCase {
			b |= 1
		}
		if o.unicode {
			b |= 2
		}
		if o.multiline {
			b |= 4
		}
		if o.dotAll {
			b |= 8
		}
		calls = append(calls, b)
	}
	return result, wordCharacters(options), calls
}
