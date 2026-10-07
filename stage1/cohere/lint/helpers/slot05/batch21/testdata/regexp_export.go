package regexp

import (
	"fmt"
	"slices"
	"sort"
	"strings"
	"unicode"
)

type AdamicAtom struct {
	Kind int    `json:"kind"`
	Lo   int    `json:"lo"`
	Hi   int    `json:"hi"`
	Text string `json:"text"`
}
type AdamicSample struct {
	Atoms   []AdamicAtom `json:"atoms"`
	Unicode bool         `json:"unicode"`
}
type AdamicDeps struct {
	Ranges      [][2]int    `json:"ranges"`
	Canonical   [][3]int    `json:"canonical"`
	Fold        [][2]int    `json:"fold"`
	UpperGroups [][]rune    `json:"upperGroups"`
	FoldGroups  [][]rune    `json:"foldGroups"`
	Escapes     [][2]string `json:"escapes"`
}

var adamicTrace strings.Builder
var adamicCan = map[[2]int]int{}
var adamicFolds = map[int]int{}
var adamicGroupSnapshot [][]rune

func adamicCanonical(r rune, u bool) rune {
	v := Canonicalize(r, u)
	flag := 0
	if u {
		flag = 1
	}
	adamicCan[[2]int{int(r), flag}] = int(v)
	fmt.Fprintf(&adamicTrace, "c:%d:%t;", r, u)
	return v
}
func adamicFold(r rune) rune {
	v := unicode.SimpleFold(r)
	adamicFolds[int(r)] = int(v)
	fmt.Fprintf(&adamicTrace, "f:%d;", r)
	return v
}
func adamicGroups(u bool) [][]rune {
	fmt.Fprintf(&adamicTrace, "groups:%t;", u)
	return adamicGroupSnapshot
}
func adamicLiteral(r rune) string { fmt.Fprintf(&adamicTrace, "literal:%d;", r); return literalRune(r) }
func adamicAtoms(in []AdamicAtom) []classAtom {
	out := []classAtom{}
	for _, a := range in {
		out = append(out, classAtom{kind: classAtomKind(a.Kind), lo: rune(a.Lo), hi: rune(a.Hi), text: a.Text})
	}
	return out
}
func adamicPublic(in []classAtom) []AdamicAtom {
	out := []AdamicAtom{}
	for _, a := range in {
		out = append(out, AdamicAtom{int(a.kind), int(a.lo), int(a.hi), a.text})
	}
	return out
}
func AdamicParse(body string, u bool) ([]AdamicAtom, bool) {
	atoms, _, e := classAtoms(body, rewriteOptions{unicode: u, ignoreCase: true}, escapeContext{unicode: u})
	return adamicPublic(atoms), e == nil
}
func AdamicJoin(s AdamicSample) string {
	out, e := joinRanges(adamicAtoms(s.Atoms), rewriteOptions{unicode: s.Unicode})
	errText := ""
	if e != nil {
		errText = e.Error()
	}
	var result strings.Builder
	fmt.Fprintf(&result, "%s|", errText)
	for _, a := range out {
		fmt.Fprintf(&result, "%d,%d,%d,%s;", a.kind, a.lo, a.hi, a.text)
	}
	return result.String()
}
func AdamicExtras(s AdamicSample) string {
	_, adamicGroupSnapshot = caseTables(s.Unicode)
	adamicTrace.Reset()
	result := caseExtras(adamicAtoms(s.Atoms), s.Unicode)
	return result + "|" + adamicTrace.String()
}
func AdamicBuild(u bool) string {
	adamicTrace.Reset()
	by, groups := buildCaseTables(u)
	sort.Slice(groups, func(i, j int) bool { return groups[i][0] < groups[j][0] })
	var result strings.Builder
	for _, g := range groups {
		for _, r := range g {
			fmt.Fprintf(&result, "%d,", r)
		}
		result.WriteByte(';')
	}
	result.WriteByte('|')
	keys := []int{}
	for r := range by {
		keys = append(keys, int(r))
	}
	sort.Ints(keys)
	for _, r := range keys {
		g := by[rune(r)]
		index := slices.IndexFunc(groups, func(group []rune) bool { return len(group) > 0 && len(g) > 0 && &group[0] == &g[0] })
		fmt.Fprintf(&result, "%d:%d;", r, index)
	}
	return result.String() + "|" + adamicTrace.String()
}
func AdamicDependencies() AdamicDeps {
	d := AdamicDeps{Ranges: [][2]int{}, Canonical: [][3]int{}, Fold: [][2]int{}, Escapes: [][2]string{}}
	for _, r := range unicode.CaseRanges {
		d.Ranges = append(d.Ranges, [2]int{int(r.Lo), int(r.Hi)})
	}
	_, d.UpperGroups = caseTables(false)
	_, d.FoldGroups = caseTables(true)
	seen := map[rune]bool{}
	for _, groups := range [][][]rune{d.UpperGroups, d.FoldGroups} {
		for _, g := range groups {
			for _, r := range g {
				if !seen[r] {
					seen[r] = true
					d.Escapes = append(d.Escapes, [2]string{fmt.Sprint(r), literalRune(r)})
				}
			}
		}
	}
	for key, v := range adamicCan {
		d.Canonical = append(d.Canonical, [3]int{key[0], key[1], v})
	}
	sort.Slice(d.Canonical, func(i, j int) bool {
		a, b := d.Canonical[i], d.Canonical[j]
		return a[0] < b[0] || a[0] == b[0] && a[1] < b[1]
	})
	for r, v := range adamicFolds {
		d.Fold = append(d.Fold, [2]int{r, v})
	}
	sort.Slice(d.Fold, func(i, j int) bool { return d.Fold[i][0] < d.Fold[j][0] })
	return d
}
