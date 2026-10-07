package tailwind

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

func AdamicPrefixKey(prefix, key string) string { return (&Theme{Prefix: prefix}).PrefixKey(key) }
func AdamicPrefixSequence(before, after string) (string, string) {
	theme := &Theme{Prefix: before}
	system := &LoadedDesignSystem{theme: theme}
	first := system.Prefix()
	theme.Prefix = after
	return first, system.Prefix()
}
func AdamicHasSequence(keys []string, root string) (bool, bool, bool) {
	registry := &VariantRegistry{registrations: map[string]VariantRegistration{}}
	for _, key := range keys {
		registry.registrations[key] = VariantRegistration{}
	}
	first := registry.Has(root)
	delete(registry.registrations, root)
	deleted := registry.Has(root)
	registry.registrations[root] = VariantRegistration{}
	return first, deleted, registry.Has(root)
}
func adamicInts(s string) []int {
	out := []int{}
	for _, b := range []byte(s) {
		out = append(out, int(b))
	}
	return out
}

var adamicLock sync.Mutex

func AdamicRecordState(prefix, key string, registry *VariantRegistry, root string) {
	path := os.Getenv("ADAMIC_SLOT04_STATES")
	if path == "" {
		return
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	names := []string{}
	if registry != nil {
		for k := range registry.registrations {
			names = append(names, k)
		}
	}
	sort.Strings(names)
	keys := [][]int{}
	for _, name := range names {
		keys = append(keys, adamicInts(name))
	}
	row := struct {
		Name              string
		Prefix, Key, Root []int
		Keys              [][]int
	}{"control:call", adamicInts(prefix), adamicInts(key), adamicInts(root), keys}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(row); err != nil {
		panic(err)
	}
}
