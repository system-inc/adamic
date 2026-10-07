package tailwind

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

var adamicRepoSorted map[string]Reading

func AdamicWave06RepositoryCases(data []byte) []byte {
	var input struct {
		Definitions []AdamicWave06Definition
		Cases       []struct {
			Roots []string
			Ids   []int
		}
	}
	if err := json.Unmarshal(data, &input); err != nil {
		panic(err)
	}
	var out strings.Builder
	for _, c := range input.Cases {
		system := &LoadedDesignSystem{staticUtilityNodes: map[string][]*Node{}}
		initial := Reading{Order: []int{-1}, Count: -99}
		table := &Table{Statics: map[string]Reading{"untouched": initial}}
		for i, root := range c.Roots {
			definition := input.Definitions[c.Ids[i]]
			system.staticUtilityNodes[root] = nodesFromStaticDeclarations(definition.Declarations)
			table.Statics[root] = initial
		}
		before := table.Statics
		adamicRepoSorted = map[string]Reading{}
		table.addRepositoryStatics(system)
		fmt.Fprintf(&out, "map %t %d\n", len(before) == len(table.Statics), len(table.Statics))
		keys := make([]string, 0, len(table.Statics))
		for key := range table.Statics {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			reading := table.Statics[key]
			order := []string{}
			for _, v := range reading.Order {
				order = append(order, strconv.Itoa(v))
			}
			shared, independent := true, true
			if sorted, found := adamicRepoSorted[key]; found {
				count := sorted.Count
				reading.Count += 100
				independent = adamicRepoSorted[key].Count == count
				reading.Count -= 100
				if len(reading.Order) > 0 {
					previous := reading.Order[0]
					reading.Order[0] = previous + 100
					shared = adamicRepoSorted[key].Order[0] == previous+100
					reading.Order[0] = previous
				}
			}
			encoded, _ := json.Marshal(key)
			fmt.Fprintf(&out, "%s %t %d %s %t %t\n", encoded, reading.Order == nil, reading.Count, strings.Join(order, ","), shared, independent)
		}
	}
	return []byte(out.String())
}
