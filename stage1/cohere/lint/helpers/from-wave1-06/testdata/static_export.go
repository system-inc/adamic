package tailwind

import (
	"encoding/json"
	"sort"
)

var adamicWave06LastReading Reading

func AdamicWave06Aliases(reading Reading, found bool) (bool, bool) {
	if !found {
		return true, true
	}
	count := adamicWave06LastReading.Count
	reading.Count += 100
	independent := adamicWave06LastReading.Count == count
	shared := true
	if len(reading.Order) > 0 {
		previous := reading.Order[0]
		reading.Order[0] = previous + 100
		shared = adamicWave06LastReading.Order[0] == previous+100
	}
	return shared, independent
}

type AdamicWave06Definition struct {
	Name         string
	Id           int
	Order        []int
	OrderNil     bool
	Count        int
	Declarations []StaticDeclaration
}

func AdamicWave06Definitions() []byte {
	FrameworkStaticDeclarations["__wave06_empty_nil__"] = nil
	FrameworkStaticDeclarations["__wave06_empty_non_nil__"] = []StaticDeclaration{}
	FrameworkStaticDeclarations["__wave06_undefined__"] = []StaticDeclaration{{Property: "display", ValuePresent: false}}
	keys := make([]string, 0, len(FrameworkStaticDeclarations))
	for key := range FrameworkStaticDeclarations {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	definitions := make([]AdamicWave06Definition, 0, len(keys))
	for id, key := range keys {
		reading, found := FrameworkStaticReading(key)
		if !found {
			panic("missing static")
		}
		definitions = append(definitions, AdamicWave06Definition{Name: key, Id: id, Order: reading.Order, OrderNil: reading.Order == nil, Count: reading.Count, Declarations: FrameworkStaticDeclarations[key]})
	}
	data, err := json.Marshal(definitions)
	if err != nil {
		panic(err)
	}
	return data
}
func AdamicWave06StaticReading(name string) Reading {
	reading, _ := FrameworkStaticReading(name)
	return reading
}
