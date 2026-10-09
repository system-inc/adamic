package ir

import (
	"crypto/sha256"
	"fmt"
	"sort"
)

// TypeTags returns the representation ABI. Keep every Type constant here.
func TypeTags() map[string]Type {
	return map[string]Type{
		"Number":       Number,
		"Boolean":      Boolean,
		"String":       String,
		"Object":       Object,
		"Array":        Array,
		"Map":          Map,
		"MaybeNumber":  MaybeNumber,
		"Closure":      Closure,
		"MaybeBoolean": MaybeBoolean,
		"Union":        Union,
		"Weak":         Weak,
		"Uint8Array":   Uint8Array,
		"Int32Array":   Int32Array,
		"Float64Array": Float64Array,
		"Uint16Array":  Uint16Array,
		"Promise":      Promise,
	}
}

// TypeTagFingerprint changes whenever a representation name or number changes.
func TypeTagFingerprint() string {
	return typeTagFingerprint(TypeTags())
}

func typeTagFingerprint(tags map[string]Type) string {
	names := make([]string, 0, len(tags))
	for name := range tags {
		names = append(names, name)
	}
	sort.Strings(names)
	hash := sha256.New()
	for _, name := range names {
		fmt.Fprintf(hash, "%d:%s=%d;", len(name), name, tags[name])
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}
