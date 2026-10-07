package tailwind

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"sync"
)

func AdamicObserveVariants(names, kinds []string, root string) {
	registry := &VariantRegistry{registrations: map[string]VariantRegistration{}}
	for i, name := range names {
		registry.registrations[name] = VariantRegistration{Name: name, Kind: ParsedVariantKind(kinds[i])}
	}
	system := &LoadedDesignSystem{variants: registry}
	fmt.Printf("has %t\nkind %s\n", system.HasVariant(root), system.VariantKind(root))
	for _, kind := range []string{"", "arbitrary", "static", "functional", "compound", "foreign"} {
		fmt.Printf("compound %s %t\n", kind, system.VariantCompoundsWith(root, ParsedVariant{Kind: ParsedVariantKind(kind), Root: string([]byte{255})}))
	}
	child := ParsedVariant{Kind: ParsedVariantKindCompound}
	delete(registry.registrations, root)
	fmt.Printf("deleted %t|%s|%t\n", system.HasVariant(root), system.VariantKind(root), system.VariantCompoundsWith(root, child))
	registry.registrations[root] = VariantRegistration{}
	fmt.Printf("empty %t|%s|%t\n", system.HasVariant(root), system.VariantKind(root), system.VariantCompoundsWith(root, ParsedVariant{Kind: ParsedVariantKindArbitrary}))
	registry.registrations[root] = VariantRegistration{Kind: ParsedVariantKindCompound}
	fmt.Printf("added %t|%s|%t\n", system.HasVariant(root), system.VariantKind(root), system.VariantCompoundsWith(root, ParsedVariant{Kind: ParsedVariantKindArbitrary}))
	registry.registrations[root] = VariantRegistration{Kind: ParsedVariantKindFunctional}
	fmt.Printf("changed %t\n", system.VariantCompoundsWith(root, child))
}
func adamicInts(s string) []int {
	out := []int{}
	for _, b := range []byte(s) {
		out = append(out, int(b))
	}
	return out
}

var adamicLock sync.Mutex

func AdamicRecordVariants(helper string, registry *VariantRegistry, root string) {
	path := os.Getenv("ADAMIC_SLOT04_VARIANTS")
	if path == "" {
		return
	}
	adamicLock.Lock()
	defer adamicLock.Unlock()
	names := []string{}
	for k := range registry.registrations {
		names = append(names, k)
	}
	sort.Strings(names)
	type Entry struct {
		Key  []int
		Kind string
	}
	entries := []Entry{}
	for _, name := range names {
		entries = append(entries, Entry{adamicInts(name), string(registry.registrations[name].Kind)})
	}
	row := struct {
		Name    string
		Helper  string
		Root    []int
		Entries []Entry
	}{"control:call", helper, adamicInts(root), entries}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	if err = json.NewEncoder(f).Encode(row); err != nil {
		panic(err)
	}
}
