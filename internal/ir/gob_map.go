package ir

import (
	"bytes"
	"encoding/gob"
	"slices"
)

// A SortedMap is a map whose gob encoding is a function of its entries. gob writes a plain map in Go's iteration order,
// which is random, so the products that hold a lowered program (estree's family checkpoints and lint's compilation,
// each a program.gob) came out as different bytes from the same program, build to build, and the key check refused
// them (landable-13b's tree, Oct 11). A SortedMap writes its entries sorted by their keys' own encodings.
//
// It is a map in every other way: indexed, ranged and assigned a map literal or make's result as before. Empty and nil
// encode alike, as a plain map's do: gob omits an empty map either way, and the decoded field is nil.
type SortedMap[K comparable, V any] map[K]V

type mapEntry[K comparable, V any] struct {
	Key   K
	Value V
}

// GobEncode writes the entries, sorted by the gob encoding of each key.
func (entries SortedMap[K, V]) GobEncode() ([]byte, error) {
	type ordered struct {
		order []byte
		entry mapEntry[K, V]
	}
	sorted := make([]ordered, 0, len(entries))
	for key, value := range entries {
		var order bytes.Buffer
		if err := gob.NewEncoder(&order).Encode(key); err != nil {
			return nil, err
		}
		sorted = append(sorted, ordered{order.Bytes(), mapEntry[K, V]{key, value}})
	}
	slices.SortFunc(sorted, func(a, b ordered) int { return bytes.Compare(a.order, b.order) })
	list := make([]mapEntry[K, V], len(sorted))
	for index, each := range sorted {
		list[index] = each.entry
	}
	var encoded bytes.Buffer
	err := gob.NewEncoder(&encoded).Encode(list)
	return encoded.Bytes(), err
}

// GobDecode reads what GobEncode wrote.
func (entries *SortedMap[K, V]) GobDecode(data []byte) error {
	var list []mapEntry[K, V]
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&list); err != nil {
		return err
	}
	*entries = make(SortedMap[K, V], len(list))
	for _, each := range list {
		(*entries)[each.Key] = each.Value
	}
	return nil
}
