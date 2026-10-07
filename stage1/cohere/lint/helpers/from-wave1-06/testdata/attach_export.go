package tailwind

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

var adamicAttachTrace []int

type adamicAttachInput struct {
	Names  []string
	Groups [][]FrameworkVariantComparisonGroup
}

func adamicAttachTheme(generation int) *Theme {
	if generation == 2 {
		return nil
	}
	theme := NewTheme()
	adamicAttachMutate(theme, generation)
	return theme
}
func adamicAttachMutate(theme *Theme, generation int) {
	small, large := "10rem", "40rem"
	if generation == 1 {
		small, large = "80rem", "20rem"
	}
	if err := theme.Add("--breakpoint-small", small, 0); err != nil {
		panic(err)
	}
	if err := theme.Add("--breakpoint-large", large, 0); err != nil {
		panic(err)
	}
}
func adamicAttachPairs(names []string) [][2]ParsedVariant {
	pairs := [][2]ParsedVariant{}
	for _, name := range names {
		pairs = append(pairs, [2]ParsedVariant{{Kind: ParsedVariantKindStatic, Root: name}, {Kind: ParsedVariantKindStatic, Root: "large"}}, [2]ParsedVariant{{Kind: ParsedVariantKindStatic, Root: "small"}, {Kind: ParsedVariantKindStatic, Root: name}})
	}
	return pairs
}
func AdamicWave06Attach(data []byte, prepare bool) []byte {
	var input adamicAttachInput
	if err := json.Unmarshal(data, &input); err != nil {
		panic(err)
	}
	pairs := adamicAttachPairs(input.Names)
	if prepare {
		type answer struct {
			Key   string
			Value int
		}
		answers := []answer{}
		for generation := 0; generation < 3; generation++ {
			theme := adamicAttachTheme(generation)
			for i, pair := range pairs {
				for _, ascending := range []bool{false, true} {
					answers = append(answers, answer{fmt.Sprintf("%d:%d:%d:%t", generation, i*2, i*2+1, ascending), compareBreakpointVariants(theme, pair[0], pair[1], ascending)})
				}
			}
		}
		result := struct {
			Groups  [][]FrameworkVariantComparisonGroup
			Pairs   []int
			Answers []answer
		}{Groups: input.Groups, Answers: answers}
		for i := range pairs {
			result.Pairs = append(result.Pairs, i)
		}
		encoded, err := json.Marshal(result)
		if err != nil {
			panic(err)
		}
		return encoded
	}
	originalGroups := FrameworkVariantComparisonGroups
	defer func() { FrameworkVariantComparisonGroups = originalGroups }()
	var output strings.Builder
	for _, groups := range input.Groups {
		for _, initialGeneration := range []int{0, 2} {
			FrameworkVariantComparisonGroups = append([]FrameworkVariantComparisonGroup{}, groups...)
			theme := adamicAttachTheme(initialGeneration)
			originalTheme := theme
			registry := NewVariantRegistry()
			adamicAttachTrace = nil
			attachFrameworkVariantComparisons(registry, theme)
			fmt.Fprintf(&output, "trace")
			for _, order := range adamicAttachTrace {
				fmt.Fprintf(&output, " %d", order)
			}
			fmt.Fprintln(&output)
			for i := range FrameworkVariantComparisonGroups {
				FrameworkVariantComparisonGroups[i].Ascending = !FrameworkVariantComparisonGroups[i].Ascending
			}
			keys := []int{}
			for order := range registry.comparisons {
				keys = append(keys, order)
			}
			sort.Ints(keys)
			for phase := 0; phase < 3; phase++ {
				if phase == 1 && originalTheme != nil {
					adamicAttachMutate(originalTheme, 1)
				}
				if phase == 2 {
					theme = adamicAttachTheme(0)
				}
				for _, order := range keys {
					callback := registry.comparisons[order]
					for i, pair := range pairs {
						fmt.Fprintf(&output, "%d %d %d %d\n", phase, order, i, callback(pair[0], pair[1]))
					}
				}
			}
		}
	}
	return []byte(output.String())
}
func AdamicWave06AttachGroups() []byte {
	groups := [][]FrameworkVariantComparisonGroup{FrameworkVariantComparisonGroups, {}, {{Order: 0, Ascending: false}, {Order: 1, Ascending: true}, {Order: 0, Ascending: true}, {Order: -1, Ascending: false}}}
	data, err := json.Marshal(groups)
	if err != nil {
		panic(err)
	}
	return data
}
