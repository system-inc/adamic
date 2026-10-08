package naming

import "strings"

// The prefix is a framework namespace, removed only before an uppercase word.
// NS: Foundation and AppKit legacy NeXT namespace; UI: UIKit; CG: CoreGraphics;
// CF: CoreFoundation; CA: QuartzCore Core Animation; AV: AVFoundation.
// No other prefixes are inferred: adding one needs a documented namespace.
var prefixes = []string{"NS", "UI", "CG", "CF", "CA", "AV"}

func dropPrefix(s string) string {
	for _, p := range prefixes {
		if strings.HasPrefix(s, p) && len(s) > len(p) && upper(s[len(p)]) {
			return s[len(p):]
		}
	}
	return s
}

// Full-word expansions intentionally do not contain max, min, or index.
var expansions = map[string]string{
	"rect": "Rectangle",  // Apple's geometric abbreviation means rectangle.
	"id":   "Identifier", // Identifier suffixes name identity, not a new acronym.
}

func upper(c byte) bool { return c >= 'A' && c <= 'Z' }
func lower(c byte) bool { return c >= 'a' && c <= 'z' }
func digit(c byte) bool { return c >= '0' && c <= '9' }

// words ports camel_case::WordIterator::computeNextPosition from
// swift/lib/Basic/StringExtras.cpp at f79ab15f11ffea6cad859da22b17da6cd9252167
// (swift-6.0-RELEASE). Changed to return Go slices instead of LLVM iterators.
// Copyright (c) 2014 - 2017 Apple Inc. and the Swift project authors.
// Apache-2.0 WITH Swift-exception; see THIRD_PARTY_NOTICES.md.
func words(s string) []string {
	var result []string
	for start := 0; start < len(s); {
		i := start
		if s[i] == '_' {
			result = append(result, "_")
			start++
			continue
		}
		for i < len(s) && upper(s[i]) {
			i++
		}
		if i-start > 1 {
			end := i
			for end < len(s) && lower(s[end]) {
				end++
			}
			suffix := s[i:end]
			if i == len(s) || ((suffix == "s" || suffix == "es" || suffix == "ies") && s[i-1:end] != "Is") {
				i = end
			} else if lower(s[i]) {
				i--
			}
		} else {
			for i < len(s) && !upper(s[i]) && s[i] != '_' {
				i++
			}
		}
		result = append(result, s[start:i])
		start = i
	}
	return result
}

// Adamic recognizes adjacent initialisms within an uppercase run. Swift's word
// iterator deliberately treats HTTPURL as one word; splitting it is rule 3.
var acronyms = []string{"HTTPS", "HTTP", "JSON", "UTF", "URL", "XML", "UUID", "ASCII", "PDF", "PNG", "JPEG", "TIFF", "HTML", "CPU", "GPU", "TCP", "IP", "DNS", "ID", "RGBA", "RGB", "API", "OS"}

func adamicWords(s string) []string {
	var out []string
	for _, w := range words(s) {
		if w == "_" {
			continue
		}
		for len(w) > 0 {
			found := false
			for _, a := range acronyms {
				if strings.HasPrefix(w, a) {
					out = append(out, a)
					w = w[len(a):]
					found = true
					break
				}
			}
			if !found {
				out = append(out, w)
				break
			}
		}
	}
	return out
}
func normalize(s string, pascal bool) string {
	s = dropPrefix(s)
	ws := adamicWords(s)
	for i, w := range ws {
		w = strings.ToLower(w)
		if full, ok := expansions[w]; ok && (w != "id" || i == len(ws)-1) {
			w = strings.ToLower(full)
		}
		if i > 0 || pascal {
			w = strings.ToUpper(w[:1]) + w[1:]
		}
		ws[i] = w
	}
	return strings.Join(ws, "")
}
func kebab(s string) string {
	ws := adamicWords(s)
	for i, w := range ws {
		ws[i] = strings.ToLower(w)
	}
	return strings.Join(ws, "-")
}

// Swift lowercases only initial acronym words, preserving later URL spelling.
func swiftLower(s string) string {
	ws := words(s)
	if len(ws) == 0 {
		return s
	}
	return strings.ToLower(ws[0]) + strings.Join(ws[1:], "")
}
