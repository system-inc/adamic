package tailwind

import "sort"

type AdamicArmCase struct {
	Op   string `json:"op"`
	Arms []bool `json:"arms"`
	Want bool   `json:"-"`
}

func AdamicArmCases() []AdamicArmCase {
	rows := []AdamicArmCase{}
	keys := []string{}
	for key := range GapRootDescriptions {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		description := GapRootDescriptions[key]
		arms := []bool{}
		for _, arm := range description.Arms {
			arms = append(arms, arm.IsColor)
		}
		rows = append(rows, AdamicArmCase{"arm", arms, gapRootAcceptsModifierOnArbitrary(description)})
	}
	// Exhaustive color masks through eight arms, independent of other description fields.
	for length := 0; length <= 8; length++ {
		for mask := 0; mask < 1<<length; mask++ {
			d := &FunctionalUtilityDescription{AcceptsModifierOnArbitrary: mask%2 == 0}
			arms := []bool{}
			for i := 0; i < length; i++ {
				v := mask&(1<<i) != 0
				arms = append(arms, v)
				d.Arms = append(d.Arms, FunctionalUtilityArm{IsColor: v})
			}
			rows = append(rows, AdamicArmCase{"arm", arms, gapRootAcceptsModifierOnArbitrary(d)})
		}
	}
	return rows
}
