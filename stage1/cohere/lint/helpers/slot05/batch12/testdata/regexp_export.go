package regexp

import (
	"fmt"
	"strings"
)

var adamicTrace string

func adamicFoldTables() (map[rune][]rune, [][]rune) { adamicTrace += "fold;"; return foldTables() }
func adamicUppercaseTables() (map[rune][]rune, [][]rune) {
	adamicTrace += "upper;"
	return uppercaseTables()
}
func adamicSameGroup(a, b []rune) bool {
	return len(a) > 0 && len(b) > 0 && &a[0] == &b[0] && len(a) == len(b)
}
func adamicSameGroups(a, b [][]rune) bool {
	return len(a) > 0 && len(b) > 0 && &a[0] == &b[0] && len(a) == len(b)
}
func adamicSameMap(a, b map[rune][]rune) bool {
	const marker rune = -2147483648
	old, present := a[marker]
	a[marker] = []rune{123456789}
	same := len(b[marker]) == 1 && b[marker][0] == 123456789
	if present {
		a[marker] = old
	} else {
		delete(a, marker)
	}
	return same
}

type AdamicTable struct {
	Groups  [][]rune `json:"groups"`
	Members [][2]int `json:"members"`
}

func AdamicTables(u bool) AdamicTable {
	var by map[rune][]rune
	var groups [][]rune
	if u {
		by, groups = foldTables()
	} else {
		by, groups = uppercaseTables()
	}
	members := [][2]int{}
	for index, group := range groups {
		for _, r := range group {
			if !adamicSameGroup(by[r], group) {
				panic("Go alias drift")
			}
			members = append(members, [2]int{int(r), index})
		}
	}
	return AdamicTable{groups, members}
}
func AdamicObserve(mode string, r rune, u bool) string {
	adamicTrace = ""
	if mode == "tables" {
		by, groups := caseTables(u)
		fold, fg := foldTables()
		upper, ug := uppercaseTables()
		label := "neither"
		if adamicSameMap(by, fold) {
			label = "fold"
		} else if adamicSameMap(by, upper) {
			label = "upper"
		}
		shared := adamicSameGroups(groups, fg) || adamicSameGroups(groups, ug)
		return fmt.Sprintf("%s:%t:%s", label, shared, adamicTrace)
	}
	if mode == "groups" {
		groups := CaseEquivalenceGroups(u)
		trace := adamicTrace
		var base [][]rune
		if u {
			_, base = foldTables()
		} else {
			_, base = uppercaseTables()
		}
		var result strings.Builder
		for _, g := range groups {
			for i, r := range g {
				if i > 0 {
					result.WriteByte(',')
				}
				fmt.Fprint(&result, r)
			}
			result.WriteByte(';')
		}
		return fmt.Sprintf("%s:%t:%s", result.String(), adamicSameGroups(groups, base), trace)
	}
	result := CaseEquivalents(r, u)
	trace := adamicTrace
	var by map[rune][]rune
	var groups [][]rune
	if u {
		by, groups = foldTables()
	} else {
		by, groups = uppercaseTables()
	}
	lookup := result == nil && by[r] == nil || adamicSameGroup(result, by[r])
	shared := false
	if result != nil {
		for _, g := range groups {
			if adamicSameGroup(result, g) {
				shared = true
				break
			}
		}
	}
	value := "nil"
	if result != nil {
		var text strings.Builder
		for i, v := range result {
			if i > 0 {
				text.WriteByte(',')
			}
			fmt.Fprint(&text, v)
		}
		value = text.String()
	}
	return fmt.Sprintf("%s:%t:%t:%s", value, lookup, shared, trace)
}
