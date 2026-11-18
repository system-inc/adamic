package generate

import (
	"strings"
)

// Deprecation: Apple's "don't use". A declaration deprecated on the platform is left out of the
// bindings, with its reason in the module's comments, as an unavailable one is (docs/apple.md). The
// JSON syntax tree carries an availability attribute but not its arguments, so they're read from
// the attribute's source text, the macro as the header writes it.

// platformNamed reports whether an availability clause's platform word is the one generated.
func (g *generator) platformNamed(word string) bool {
	platform := g.configuration.Platform
	word = strings.ToLower(strings.TrimSpace(word))
	return word == platform || platform == "macos" && word == "macosx" || platform == "visionos" && word == "xros"
}

// deprecatedIn reads one availability attribute's text: API_DEPRECATED("...", macos(10.0, 10.4)),
// NS_DEPRECATED_MAC(10_0, 10_4), availability(macos, deprecated=10.4), or a framework's own macro
// over one of those, expanded from its #define. API_TO_BE_DEPRECATED is a version not reached yet,
// not a deprecation.
func (g *generator) deprecatedIn(text, file string) bool {
	if strings.HasPrefix(text, "availability(") || strings.HasPrefix(text, "__attribute__") {
		return g.availabilityDeprecated(text)
	}
	upper := strings.ToUpper(text)
	if !strings.Contains(upper, "DEPRECATED") {
		return false
	}
	name, arguments, _ := strings.Cut(text, "(")
	arguments = strings.TrimSuffix(strings.TrimSpace(arguments), ")")
	name = strings.ToUpper(strings.TrimSpace(name))
	versions := func(clause string) []string {
		parts := []string{}
		for _, part := range strings.Split(clause, ",") {
			parts = append(parts, strings.TrimSpace(part))
		}
		return parts
	}
	switch {
	case strings.HasPrefix(name, "API_DEPRECATED"):
		// The platforms it names, each with its introduced and deprecated versions.
		for rest := arguments; ; {
			open := strings.IndexByte(rest, '(')
			if open < 0 {
				return false
			}
			close := strings.IndexByte(rest[open:], ')')
			if close < 0 {
				return false
			}
			word := rest[:open]
			if at := strings.LastIndexAny(word, ", \t"); at >= 0 {
				word = word[at+1:]
			}
			clause := versions(rest[open+1 : open+close])
			if g.platformNamed(word) {
				return len(clause) >= 2 && clause[1] != "API_TO_BE_DEPRECATED"
			}
			rest = rest[open+close+1:]
		}
	case strings.Contains(name, "BUT_DEPRECATED"):
		// AVAILABLE_MAC_OS_X_VERSION_10_0_AND_LATER_BUT_DEPRECATED_IN_MAC_OS_X_VERSION_10_4,
		// __OSX_AVAILABLE_BUT_DEPRECATED: the Mac's.
		return g.platformNamed("macos")
	case strings.HasSuffix(name, "_MAC"):
		return g.platformNamed("macos")
	case strings.HasSuffix(name, "_IOS"):
		return g.platformNamed("ios")
	case name == "NS_DEPRECATED" || name == "NS_CLASS_DEPRECATED" || name == "NS_ENUM_DEPRECATED" || name == "NS_DEPRECATED_WITH_REPLACEMENT":
		// The Mac's introduced and deprecated versions, then iOS's; NA where it has none.
		clause := versions(arguments)
		if name == "NS_DEPRECATED_WITH_REPLACEMENT" && len(clause) > 0 {
			clause = clause[1:]
		}
		at := 1
		if g.platformNamed("ios") {
			at = 3
		}
		return len(clause) > at && clause[at] != "NA"
	case name == "DEPRECATED_ATTRIBUTE" || name == "DEPRECATED_MSG_ATTRIBUTE" || name == "__DEPRECATED":
		return true
	case strings.Contains(name, "TO_BE_DEPRECATED"):
		return false
	}
	// A framework's own macro (CG_AVAILABLE_BUT_DEPRECATED): what it expands to. One that can't be
	// read is taken as deprecated: left out, which is the cautious reading of "don't use".
	if expanded, found := g.expandMacro(text, file); found && expanded != text {
		return g.deprecatedIn(strings.TrimSpace(expanded), file)
	}
	return true
}

// availabilityDeprecated reads availability(platform, ..., deprecated=version) clauses. A version of
// 100000 is API_TO_BE_DEPRECATED expanded.
func (g *generator) availabilityDeprecated(text string) bool {
	for rest := text; ; {
		at := strings.Index(rest, "availability(")
		if at < 0 {
			return strings.Contains(text, "deprecated") && !strings.Contains(text, "availability(")
		}
		rest = rest[at+len("availability("):]
		end := strings.IndexByte(rest, ')')
		if end < 0 {
			return false
		}
		clauses := strings.Split(rest[:end], ",")
		if g.platformNamed(clauses[0]) {
			for _, clause := range clauses[1:] {
				clause = strings.Join(strings.Fields(clause), "")
				if version, found := strings.CutPrefix(clause, "deprecated="); found && version != "100000" {
					return true
				}
			}
		}
		rest = rest[end:]
	}
}

// deprecatedByName reports whether a region's macro, which names its platforms instead of taking
// them, deprecates its declarations here.
func (g *generator) deprecatedByName(macro string) bool {
	at := strings.Index(macro, "DEPRECATED")
	if at < 0 || strings.Contains(macro, "TO_BE_DEPRECATED") {
		return false
	}
	words := strings.Split(strings.ToLower(macro[at+len("DEPRECATED"):]), "_")
	named := false
	for _, word := range words {
		switch word {
		case "", "begin", "end", "with", "replacement":
			continue
		}
		named = true
		if g.platformNamed(word) {
			return true
		}
	}
	return !named
}
