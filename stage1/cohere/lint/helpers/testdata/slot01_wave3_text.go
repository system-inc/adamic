package text

import "sort"

func AdamicWave3Decode(item string) (string, bool) { return decodeEntity(item) }
func AdamicWave3Entities() [][]string {
	keys := []string{}
	for key := range xhtmlEntities {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	rows := [][]string{}
	for _, key := range keys {
		rows = append(rows, []string{key, xhtmlEntities[key]})
	}
	return rows
}
func AdamicWave3EmptyPanics() (panics bool) {
	defer func() { panics = recover() != nil }()
	decodeEntity("")
	return
}
