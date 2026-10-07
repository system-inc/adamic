package regexp

import "fmt"

var adamicTrace string
var adamicTraceRunes bool
var adamicAtoms []classAtom

func adamicTraceRune(r rune) string {
	if adamicTraceRunes {
		adamicTrace += fmt.Sprintf("%d:", r)
	}
	return literalRune(r)
}
func AdamicClassAtomWrite(kind int, lo, hi rune, text string) (string, string) {
	adamicTraceRunes = true
	adamicTrace = ""
	result := (classAtom{kind: classAtomKind(kind), lo: lo, hi: hi, text: text}).write()
	return result, adamicTrace
}
func AdamicLiteralRune(r rune) string { return literalRune(r) }
func adamicTraceWordAtoms(options rewriteOptions) []classAtom {
	adamicTrace += fmt.Sprintf("atoms:%t:%t:%t:%t;", options.unicode, options.ignoreCase, options.multiline, options.dotAll)
	adamicAtoms = wordClassAtoms(options)
	return adamicAtoms
}
func adamicTraceWriteClass(atoms []classAtom, negated bool, options rewriteOptions) string {
	same := len(atoms) > 0 && len(adamicAtoms) > 0 && &atoms[0] == &adamicAtoms[0]
	adamicTrace += fmt.Sprintf("write:%t:%t:%t:%t:%t:%t;", options.unicode, options.ignoreCase, options.multiline, options.dotAll, negated, same)
	return writeClass(atoms, negated, options)
}

type AdamicAtom struct {
	Kind int    `json:"kind"`
	Lo   int    `json:"lo"`
	Hi   int    `json:"hi"`
	Text string `json:"text"`
}

func AdamicWordCharacters(u, i, m, d bool) (string, string, []AdamicAtom) {
	adamicTraceRunes = false
	adamicTrace = ""
	result := wordCharacters(rewriteOptions{unicode: u, ignoreCase: i, multiline: m, dotAll: d})
	atoms := []AdamicAtom{}
	for _, a := range adamicAtoms {
		atoms = append(atoms, AdamicAtom{int(a.kind), int(a.lo), int(a.hi), a.text})
	}
	return result, adamicTrace, atoms
}
func AdamicExpandsOnUppercase(r rune) bool { return expandsOnUppercase(r) }
