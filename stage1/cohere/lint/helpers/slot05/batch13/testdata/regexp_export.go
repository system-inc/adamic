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
