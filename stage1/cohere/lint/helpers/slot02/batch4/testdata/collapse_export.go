package tailwind

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

type AdamicSlot02Batch4Corpus struct {
	Texts          []string
	Want           string
	ParsedVariants int
}

func AdamicSlot02Batch4Observe(input []string) []byte {
	values := map[string]bool{"": true, "@": true, "@min": true, "@max": true, "min": true, "max": true, "sm": true, "lg": true, "x@y": true, " @min": true, "＠min": true, "TW": true, "tw": true, "a-z": true, "a0": true, "a_b": true, "a\x00b": true, "é😀_世界_@": true}
	for _, name := range []string{"a_b", "a\\_b", "a\\\\_b", "a\\\\\\_b", "_", "\\_", "\\", "\\\\", "__", "\\__", "_\\_", "a\\x_b", "_😀\\_世界_", "\n_\t\\_"} {
		values[name] = true
	}
	for count := 0; count <= 8; count++ {
		values[strings.Repeat("\\", count)+"_"] = true
		values[strings.Repeat("_", count)+"\\_"] = true
	}
	registry := NewVariantRegistry()
	registry.RegisterFrameworkVariants(FrameworkVariantRegistrations)
	system := &LoadedDesignSystem{variants: registry, theme: NewTheme()}
	corpus := AdamicSlot02Batch4Corpus{}
	for _, text := range input {
		values[text] = true
		_, prefix := parseThemeOptions(text)
		values[prefix] = true
		for _, field := range strings.Fields(text) {
			values[field] = true
			for _, part := range segment(field, ':') {
				if variant := ParseVariant(part, system); variant != nil {
					corpus.ParsedVariants++
					for current := variant; current != nil; current = current.Variant {
						values[current.Root] = true
					}
				}
			}
		}
	}
	for text := range values {
		corpus.Texts = append(corpus.Texts, text)
	}
	sort.Strings(corpus.Texts)
	var want strings.Builder
	observe := func(text string) {
		fmt.Fprintf(&want, "prefix:%t\nnamespace:%s\n", isValidThemePrefix(text), namespaceForVariantRoot(text))
		fmt.Fprintf(&want, "convert:%s\nkeep:%s\n", convertUnderscoresToWhitespace(text, false), convertUnderscoresToWhitespace(text, true))
	}
	for _, text := range corpus.Texts {
		observe(text)
	}
	// Every ASCII byte pair, including NUL, both scanner orders and every escape.
	for first := 0; first < 128; first++ {
		for second := 0; second < 128; second++ {
			observe(string([]byte{byte(first), byte(second)}))
		}
	}
	prefixCount, namespaceCount := 0, 0
	for code := 0; code <= 0x10ffff; code++ {
		value := string(rune(code))
		if isValidThemePrefix("a" + value + "b") {
			fmt.Fprintf(&want, "prefix-code:%d\n", code)
			prefixCount++
		}
		if namespace := namespaceForVariantRoot(value + "root"); namespace != "--breakpoint" {
			fmt.Fprintf(&want, "namespace-code:%d:%s\n", code, namespace)
			namespaceCount++
		}
	}
	fmt.Fprintf(&want, "prefix-count:%d\nnamespace-count:%d\n", prefixCount, namespaceCount)
	corpus.Want = want.String()
	data, err := json.Marshal(corpus)
	if err != nil {
		panic(err)
	}
	return data
}
