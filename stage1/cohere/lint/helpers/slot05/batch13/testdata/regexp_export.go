package regexp

import (
	"fmt"
	"unicode"
)

var adamicTrace string

func adamicCanonicalFold(r rune) rune {
	adamicTrace += fmt.Sprintf("fold:%d;", r)
	return simpleFold(r)
}
func adamicCanonicalExpand(r rune) bool {
	adamicTrace += fmt.Sprintf("expand:%d;", r)
	return expandsOnUppercase(r)
}
func adamicCanonicalUpper(r rune) rune {
	adamicTrace += fmt.Sprintf("upper:%d;", r)
	return unicode.ToUpper(r)
}
func adamicClassEquivalents(r rune, u bool) []rune {
	adamicTrace += fmt.Sprintf("equivalents:%d:%t;", r, u)
	return CaseEquivalents(r, u)
}
func adamicClassEscape(r rune) string {
	adamicTrace += fmt.Sprintf("escape:%d;", r)
	return EscapeClassRune(r)
}

type AdamicTable struct {
	Groups  [][]rune `json:"groups"`
	Members [][2]int `json:"members"`
}

func AdamicTables(u bool) AdamicTable {
	_, groups := caseTables(u)
	members := [][2]int{}
	for index, g := range groups {
		for _, r := range g {
			members = append(members, [2]int{int(r), index})
		}
	}
	return AdamicTable{groups, members}
}

type AdamicEscape struct {
	Rune int    `json:"rune"`
	Text string `json:"text"`
}
type AdamicData struct {
	Upper    [][2]int       `json:"upper"`
	Fold     [][2]int       `json:"fold"`
	Expanded []int          `json:"expanded"`
	Escapes  []AdamicEscape `json:"escapes"`
}

func AdamicDependencies() AdamicData {
	d := AdamicData{Upper: [][2]int{}, Fold: [][2]int{}, Expanded: []int{}, Escapes: []AdamicEscape{}}
	for r := rune(0); r <= 0x10ffff; r++ {
		if upper := unicode.ToUpper(r); upper != r {
			d.Upper = append(d.Upper, [2]int{int(r), int(upper)})
		}
		if folded := simpleFold(r); folded != r {
			d.Fold = append(d.Fold, [2]int{int(r), int(folded)})
		}
		if expandsOnUppercase(r) {
			d.Expanded = append(d.Expanded, int(r))
		}
	}
	seen := map[rune]bool{}
	for _, u := range []bool{false, true} {
		_, groups := caseTables(u)
		for _, g := range groups {
			for _, r := range g {
				if !seen[r] {
					seen[r] = true
					d.Escapes = append(d.Escapes, AdamicEscape{int(r), EscapeClassRune(r)})
				}
			}
		}
	}
	return d
}
func AdamicObserve(mode string, r rune, u bool) string {
	adamicTrace = ""
	if mode == "canonical" {
		value := Canonicalize(r, u)
		return fmt.Sprintf("%d:%s", value, adamicTrace)
	}
	text, widened := CaseClass(r, u)
	return fmt.Sprintf("%t:%s:%s", widened, text, adamicTrace)
}

var adamicObservedAtoms []classAtom

func adamicAtomWrite(a classAtom) string {
	adamicTrace += fmt.Sprintf("atom:%d:%d:%d:%s;", a.kind, a.lo, a.hi, a.text)
	return a.write()
}
func adamicWriteExtras(atoms []classAtom, u bool) string {
	same := len(atoms) > 0 && len(adamicObservedAtoms) > 0 && len(atoms) == len(adamicObservedAtoms) && &atoms[0] == &adamicObservedAtoms[0]
	adamicTrace += fmt.Sprintf("extras:%t:%t;", u, same)
	return caseExtras(atoms, u)
}

type AdamicAtom struct {
	Kind int    `json:"kind"`
	Lo   int    `json:"lo"`
	Hi   int    `json:"hi"`
	Text string `json:"text"`
}
type AdamicWriteSample struct {
	Atoms      []AdamicAtom `json:"atoms"`
	Texts      []string     `json:"texts"`
	Extras     string       `json:"extras"`
	Negated    bool         `json:"negated"`
	Unicode    bool         `json:"unicode"`
	IgnoreCase bool         `json:"ignoreCase"`
	Multiline  bool         `json:"multiline"`
	DotAll     bool         `json:"dotAll"`
}

func AdamicWrite(atoms []AdamicAtom, neg, u, ignore, m, d bool) (AdamicWriteSample, string) {
	originals := []classAtom{}
	texts := []string{}
	for _, a := range atoms {
		atom := classAtom{kind: classAtomKind(a.Kind), lo: rune(a.Lo), hi: rune(a.Hi), text: a.Text}
		originals = append(originals, atom)
		texts = append(texts, atom.write())
	}
	extras := caseExtras(originals, u)
	adamicTrace = ""
	adamicObservedAtoms = originals
	result := writeClass(originals, neg, rewriteOptions{unicode: u, ignoreCase: ignore, multiline: m, dotAll: d})
	return AdamicWriteSample{atoms, texts, extras, neg, u, ignore, m, d}, result + "\n" + adamicTrace + "\n"
}
