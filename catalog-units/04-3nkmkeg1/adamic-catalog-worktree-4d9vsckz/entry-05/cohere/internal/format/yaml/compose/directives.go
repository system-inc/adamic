package compose

// Ported from eemeli/yaml 2.9.0, dist/doc/directives.js: the parts parsing reaches (the constructor,
// atDocument, add, tagName). clone, tagString and toString serve stringify and the compat schema option,
// which yaml-unist-parser does not set.

import (
	"maps"
	"regexp"
	"unicode/utf8"
)

// DirectivesYAML is the %YAML directive's state.
type DirectivesYAML struct {
	Explicit bool
	Version  string
}

// Directives is a document's (or the stream's) directives.
type Directives struct {
	// DocStart is the directives-end/doc-start marker ---: upstream's null when false, never false.
	DocStart bool
	// DocEnd is the doc-end marker ...
	DocEnd bool
	YAML   DirectivesYAML
	// Tags maps each %TAG handle to its prefix.
	Tags map[string]string

	atNextDocument bool
}

// defaultTags is Directives.defaultTags.
func defaultTags() map[string]string {
	return map[string]string{"!!": "tag:yaml.org,2002:"}
}

// newDirectives is `new Directives(yaml, tags)`: Object.assign over the defaults, so both are copies.
func newDirectives(yaml DirectivesYAML, tags map[string]string) *Directives {
	merged := defaultTags()
	maps.Copy(merged, tags)
	return &Directives{YAML: yaml, Tags: merged}
}

// atDocument is upstream's: during parsing, get a Directives instance for the current document and
// update the stream state according to the current version's spec.
func (directives *Directives) atDocument() *Directives {
	result := newDirectives(directives.YAML, directives.Tags)
	switch directives.YAML.Version {
	case "1.1":
		directives.atNextDocument = true
	case "1.2":
		directives.atNextDocument = false
		directives.YAML = DirectivesYAML{Explicit: false, Version: "1.2"}
		directives.Tags = defaultTags()
	}
	return result
}

var versionPattern = regexp.MustCompile(`^\d+\.\d+$`)

// add is upstream's. onError may be called even if the action was successful. Returns true on
// success.
func (directives *Directives) add(line []uint16, onError func(offset int, message string, warning bool)) bool {
	if directives.atNextDocument {
		directives.YAML = DirectivesYAML{Explicit: false, Version: "1.1"}
		directives.Tags = defaultTags()
		directives.atNextDocument = false
	}
	parts := splitSpaceOrTab(trim(line))
	name := parts[0]
	parts = parts[1:]
	switch {
	case equalsASCII(name, "%TAG"):
		if len(parts) != 2 {
			onError(0, "%TAG directive should contain exactly two parts", false)
			if len(parts) < 2 {
				return false
			}
		}
		handle, prefix := parts[0], parts[1]
		directives.Tags[unitsToString(handle)] = unitsToString(prefix)
		return true
	case equalsASCII(name, "%YAML"):
		directives.YAML.Explicit = true
		if len(parts) != 1 {
			onError(0, "%YAML directive should contain exactly one part", false)
			return false
		}
		version := parts[0]
		if equalsASCII(version, "1.1") || equalsASCII(version, "1.2") {
			directives.YAML.Version = unitsToString(version)
			return true
		}
		isValid := versionPattern.MatchString(asciiView(version))
		onError(6, "Unsupported YAML version "+unitsToString(version), isValid)
		return false
	default:
		onError(0, "Unknown directive "+unitsToString(name), true)
		return false
	}
}

// splitSpaceOrTab is `text.split(/[ \t]+/)`.
func splitSpaceOrTab(text []uint16) [][]uint16 {
	parts := [][]uint16{}
	start := 0
	for i := 0; i < len(text); {
		if text[i] == ' ' || text[i] == '\t' {
			parts = append(parts, text[start:i:i])
			for i < len(text) && (text[i] == ' ' || text[i] == '\t') {
				i++
			}
			start = i
			continue
		}
		i++
	}
	return append(parts, text[start:len(text):len(text)])
}

// tagName resolves a tag, matching handles to those defined in %TAG directives. It returns the resolved
// tag, which may also be the non-specific tag '!' or a '!local' tag, or "" where upstream returns null
// (unresolvable). Upstream can also return the empty string, for !<>; every caller treats the two alike.
func (directives *Directives) tagName(source []uint16, onError func(message string)) string {
	if equalsASCII(source, "!") {
		return "!" // non-specific tag
	}
	if characterAt(source, 0) != '!' {
		onError("Not a valid tag: " + unitsToString(source))
		return ""
	}
	if characterAt(source, 1) == '<' {
		verbatim := slice(source, 2, -1)
		if equalsASCII(verbatim, "!") || equalsASCII(verbatim, "!!") {
			onError("Verbatim tags aren't resolved, so " + unitsToString(source) + " is invalid.")
			return ""
		}
		if characterAt(source, len(source)-1) != '>' {
			onError("Verbatim tags must end with a >")
		}
		return unitsToString(verbatim)
	}
	// `source.match(/^(.*!)([^!]*)$/s)`: the handle runs to the last !, which source[0] guarantees.
	last := 0
	for index, unit := range source {
		if unit == '!' {
			last = index
		}
	}
	handle, suffix := source[:last+1], source[last+1:]
	if len(suffix) == 0 {
		onError("The " + unitsToString(source) + " tag has no suffix")
	}
	prefix := directives.Tags[unitsToString(handle)]
	if prefix != "" {
		decoded, ok := decodeURIComponent(suffix)
		if !ok {
			onError("URIError: URI malformed")
			return ""
		}
		return prefix + unitsToString(decoded)
	}
	if equalsASCII(handle, "!") {
		return unitsToString(source) // local tag
	}
	onError("Could not resolve tag: " + unitsToString(source))
	return ""
}

// decodeURIComponent is upstream's global, ECMAScript's Decode with an empty reserved set: each %XX
// escape is a UTF-8 byte, and a malformed escape or byte sequence is a URIError (ok false).
func decodeURIComponent(text []uint16) (decoded []uint16, ok bool) {
	decoded = make([]uint16, 0, len(text))
	hexByte := func(index int) (byte, bool) {
		if index+2 >= len(text) || text[index] != '%' {
			return 0, false
		}
		high, highOK := hexDigit(text[index+1])
		low, lowOK := hexDigit(text[index+2])
		if !highOK || !lowOK {
			return 0, false
		}
		return high<<4 | low, true
	}
	for k := 0; k < len(text); k++ {
		if text[k] != '%' {
			decoded = append(decoded, text[k])
			continue
		}
		first, isByte := hexByte(k)
		if !isByte {
			return nil, false
		}
		k += 2
		if first < 0x80 {
			decoded = append(decoded, uint16(first))
			continue
		}
		// The number of leading 1 bits is the sequence's length; 1 or more than 4 is malformed.
		n := 0
		for first<<n&0x80 != 0 {
			n++
		}
		if n == 1 || n > 4 {
			return nil, false
		}
		octets := []byte{first}
		if k+3*(n-1) >= len(text) {
			return nil, false
		}
		for j := 1; j < n; j++ {
			k++
			octet, isOctet := hexByte(k)
			if !isOctet || octet&0xC0 != 0x80 {
				return nil, false
			}
			k += 2
			octets = append(octets, octet)
		}
		character, size := utf8.DecodeRune(octets)
		if character == utf8.RuneError && size <= 1 || size != len(octets) {
			return nil, false
		}
		decoded = append(decoded, stringToUnits(string(character))...)
	}
	return decoded, true
}

func hexDigit(unit uint16) (byte, bool) {
	switch {
	case unit >= '0' && unit <= '9':
		return byte(unit - '0'), true
	case unit >= 'a' && unit <= 'f':
		return byte(unit - 'a' + 10), true
	case unit >= 'A' && unit <= 'F':
		return byte(unit - 'A' + 10), true
	}
	return 0, false
}
