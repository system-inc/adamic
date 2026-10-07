package tailwind

import (
	"fmt"
	"strings"
)

type Slot12ThemeValue struct {
	Key, Value string
	Options    int
}
type Slot12ThemeCase struct {
	Theme                    int
	Candidate                string
	Present                  bool
	Keys                     []string
	Options                  int
	Values                   []Slot12ThemeValue
	Key, Reference           string
	KeyFound, ReferenceFound bool
}

var slot12ThemeID int
var slot12ThemeTrace string
var slot12Keys []string

func Slot12Key(t *Theme, v string, p bool, keys []string) (string, bool) {
	same := len(keys) == len(slot12Keys)
	if same && len(keys) > 0 {
		same = &keys[0] == &slot12Keys[0]
	}
	slot12ThemeTrace += fmt.Sprintf("key:%d:%s:%t:%t;", slot12ThemeID, v, p, same)
	for _, k := range keys {
		slot12ThemeTrace += k + ","
	}
	return t.resolveKey(v, p, keys)
}
func Slot12Reference(t *Theme, k string) (string, bool) {
	slot12ThemeTrace += fmt.Sprintf("reference:%d:%s;", slot12ThemeID, k)
	return t.variableReference(k)
}
func Slot12Theme(texts []string) ([]Slot12ThemeCase, string) {
	cases := []Slot12ThemeCase{}
	var want strings.Builder
	values := []Slot12ThemeValue{{"--color-red", "red", 0}, {"--color-blue", "", 1}, {"--spacing-4", "1rem", 2}, {"--spacing", "8px", 1}, {"--color-é😀", "été😀", 0}}
	texts = append(texts, "red", "blue", "4", "é😀", "missing", "")
	seen := map[string]bool{}
	for _, candidate := range texts {
		if seen[candidate] {
			continue
		}
		seen[candidate] = true
		for _, present := range []bool{false, true} {
			limitOptions, limitThemes := 1, 1
			probe := NewTheme()
			for _, v := range values {
				probe.values[v.Key] = themeValue{value: v.Value, options: ThemeOptions(v.Options)}
			}
			_, matches := probe.resolveKey(candidate, true, []string{"--color", "--spacing"})
			if matches || candidate == "" || candidate == "missing" {
				limitOptions = 4
				limitThemes = 2
			}
			for options := 0; options < limitOptions; options++ {
				for id := 0; id < limitThemes; id++ {
					theme := NewTheme()
					for _, v := range values {
						theme.values[v.Key] = themeValue{value: v.Value, options: ThemeOptions(v.Options)}
					}
					if id == 1 {
						theme.Prefix = "tw"
					}
					keys := []string{"--color", "--spacing"}
					key, kf := theme.resolveKey(candidate, present, keys)
					reference, rf := theme.variableReference(key)
					slot12ThemeID = id
					slot12Keys = keys
					slot12ThemeTrace = ""
					value, found := theme.Resolve(candidate, present, keys, ThemeOptions(options))
					want.WriteString(fmt.Sprintf("theme:%s:%t:%s\n", slot12ThemeTrace, found, value))
					cases = append(cases, Slot12ThemeCase{id, candidate, present, keys, options, values, key, reference, kf, rf})
				}
			}
		}
	}
	return cases, want.String()
}
