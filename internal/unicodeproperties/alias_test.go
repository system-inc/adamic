package unicodeproperties

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"testing"

	cohere "github.com/system-inc/cohere/unicodeproperties"
)

// specBinary is Table 66 of ECMA-262 (binary Unicode properties) plus WSpace.
// WSpace is the short name in PropertyAliases.txt. The spec's rendered table
// omits it (tc39/ecma262#3286) and Node 24 accepts it, so the lookup does too.
// The canonical name is the map value. This list is the check that an alias
// cannot quietly disappear from cohere's tables, which Lookup reads.
var specBinary = map[string]string{
	"ASCII":                        "ASCII",
	"AHex":                         "ASCII_Hex_Digit",
	"ASCII_Hex_Digit":              "ASCII_Hex_Digit",
	"Alpha":                        "Alphabetic",
	"Alphabetic":                   "Alphabetic",
	"Any":                          "Any",
	"Assigned":                     "Assigned",
	"Bidi_C":                       "Bidi_Control",
	"Bidi_Control":                 "Bidi_Control",
	"Bidi_M":                       "Bidi_Mirrored",
	"Bidi_Mirrored":                "Bidi_Mirrored",
	"CI":                           "Case_Ignorable",
	"Case_Ignorable":               "Case_Ignorable",
	"Cased":                        "Cased",
	"CWCF":                         "Changes_When_Casefolded",
	"Changes_When_Casefolded":      "Changes_When_Casefolded",
	"CWCM":                         "Changes_When_Casemapped",
	"Changes_When_Casemapped":      "Changes_When_Casemapped",
	"CWL":                          "Changes_When_Lowercased",
	"Changes_When_Lowercased":      "Changes_When_Lowercased",
	"CWKCF":                        "Changes_When_NFKC_Casefolded",
	"Changes_When_NFKC_Casefolded": "Changes_When_NFKC_Casefolded",
	"CWT":                          "Changes_When_Titlecased",
	"Changes_When_Titlecased":      "Changes_When_Titlecased",
	"CWU":                          "Changes_When_Uppercased",
	"Changes_When_Uppercased":      "Changes_When_Uppercased",
	"Dash":                         "Dash",
	"DI":                           "Default_Ignorable_Code_Point",
	"Default_Ignorable_Code_Point": "Default_Ignorable_Code_Point",
	"Dep":                          "Deprecated",
	"Deprecated":                   "Deprecated",
	"Dia":                          "Diacritic",
	"Diacritic":                    "Diacritic",
	"Emoji":                        "Emoji",
	"EComp":                        "Emoji_Component",
	"Emoji_Component":              "Emoji_Component",
	"EMod":                         "Emoji_Modifier",
	"Emoji_Modifier":               "Emoji_Modifier",
	"EBase":                        "Emoji_Modifier_Base",
	"Emoji_Modifier_Base":          "Emoji_Modifier_Base",
	"EPres":                        "Emoji_Presentation",
	"Emoji_Presentation":           "Emoji_Presentation",
	"ExtPict":                      "Extended_Pictographic",
	"Extended_Pictographic":        "Extended_Pictographic",
	"Ext":                          "Extender",
	"Extender":                     "Extender",
	"Gr_Base":                      "Grapheme_Base",
	"Grapheme_Base":                "Grapheme_Base",
	"Gr_Ext":                       "Grapheme_Extend",
	"Grapheme_Extend":              "Grapheme_Extend",
	"Hex":                          "Hex_Digit",
	"Hex_Digit":                    "Hex_Digit",
	"IDSB":                         "IDS_Binary_Operator",
	"IDS_Binary_Operator":          "IDS_Binary_Operator",
	"IDST":                         "IDS_Trinary_Operator",
	"IDS_Trinary_Operator":         "IDS_Trinary_Operator",
	"IDC":                          "ID_Continue",
	"ID_Continue":                  "ID_Continue",
	"IDS":                          "ID_Start",
	"ID_Start":                     "ID_Start",
	"Ideo":                         "Ideographic",
	"Ideographic":                  "Ideographic",
	"Join_C":                       "Join_Control",
	"Join_Control":                 "Join_Control",
	"LOE":                          "Logical_Order_Exception",
	"Logical_Order_Exception":      "Logical_Order_Exception",
	"Lower":                        "Lowercase",
	"Lowercase":                    "Lowercase",
	"Math":                         "Math",
	"NChar":                        "Noncharacter_Code_Point",
	"Noncharacter_Code_Point":      "Noncharacter_Code_Point",
	"Pat_Syn":                      "Pattern_Syntax",
	"Pattern_Syntax":               "Pattern_Syntax",
	"Pat_WS":                       "Pattern_White_Space",
	"Pattern_White_Space":          "Pattern_White_Space",
	"QMark":                        "Quotation_Mark",
	"Quotation_Mark":               "Quotation_Mark",
	"Radical":                      "Radical",
	"RI":                           "Regional_Indicator",
	"Regional_Indicator":           "Regional_Indicator",
	"STerm":                        "Sentence_Terminal",
	"Sentence_Terminal":            "Sentence_Terminal",
	"SD":                           "Soft_Dotted",
	"Soft_Dotted":                  "Soft_Dotted",
	"Term":                         "Terminal_Punctuation",
	"Terminal_Punctuation":         "Terminal_Punctuation",
	"UIdeo":                        "Unified_Ideograph",
	"Unified_Ideograph":            "Unified_Ideograph",
	"Upper":                        "Uppercase",
	"Uppercase":                    "Uppercase",
	"VS":                           "Variation_Selector",
	"Variation_Selector":           "Variation_Selector",
	"space":                        "White_Space",
	"WSpace":                       "White_Space",
	"White_Space":                  "White_Space",
	"XIDC":                         "XID_Continue",
	"XID_Continue":                 "XID_Continue",
	"XIDS":                         "XID_Start",
	"XID_Start":                    "XID_Start",
}

// specGeneralCategory is every General_Category alias in PropertyValueAliases.txt,
// including the grouped values (L, LC, P, ...) and the extra spellings cntrl,
// digit, punct and Combining_Mark. The value is the long name.
var specGeneralCategory = map[string]string{
	"C": "Other", "Other": "Other",
	"Cc": "Control", "Control": "Control", "cntrl": "Control",
	"Cf": "Format", "Format": "Format",
	"Cn": "Unassigned", "Unassigned": "Unassigned",
	"Co": "Private_Use", "Private_Use": "Private_Use",
	"Cs": "Surrogate", "Surrogate": "Surrogate",
	"L": "Letter", "Letter": "Letter",
	"LC": "Cased_Letter", "Cased_Letter": "Cased_Letter",
	"Ll": "Lowercase_Letter", "Lowercase_Letter": "Lowercase_Letter",
	"Lm": "Modifier_Letter", "Modifier_Letter": "Modifier_Letter",
	"Lo": "Other_Letter", "Other_Letter": "Other_Letter",
	"Lt": "Titlecase_Letter", "Titlecase_Letter": "Titlecase_Letter",
	"Lu": "Uppercase_Letter", "Uppercase_Letter": "Uppercase_Letter",
	"M": "Mark", "Mark": "Mark", "Combining_Mark": "Mark",
	"Mc": "Spacing_Mark", "Spacing_Mark": "Spacing_Mark",
	"Me": "Enclosing_Mark", "Enclosing_Mark": "Enclosing_Mark",
	"Mn": "Nonspacing_Mark", "Nonspacing_Mark": "Nonspacing_Mark",
	"N": "Number", "Number": "Number",
	"Nd": "Decimal_Number", "Decimal_Number": "Decimal_Number", "digit": "Decimal_Number",
	"Nl": "Letter_Number", "Letter_Number": "Letter_Number",
	"No": "Other_Number", "Other_Number": "Other_Number",
	"P": "Punctuation", "Punctuation": "Punctuation", "punct": "Punctuation",
	"Pc": "Connector_Punctuation", "Connector_Punctuation": "Connector_Punctuation",
	"Pd": "Dash_Punctuation", "Dash_Punctuation": "Dash_Punctuation",
	"Pe": "Close_Punctuation", "Close_Punctuation": "Close_Punctuation",
	"Pf": "Final_Punctuation", "Final_Punctuation": "Final_Punctuation",
	"Pi": "Initial_Punctuation", "Initial_Punctuation": "Initial_Punctuation",
	"Po": "Other_Punctuation", "Other_Punctuation": "Other_Punctuation",
	"Ps": "Open_Punctuation", "Open_Punctuation": "Open_Punctuation",
	"S": "Symbol", "Symbol": "Symbol",
	"Sc": "Currency_Symbol", "Currency_Symbol": "Currency_Symbol",
	"Sk": "Modifier_Symbol", "Modifier_Symbol": "Modifier_Symbol",
	"Sm": "Math_Symbol", "Math_Symbol": "Math_Symbol",
	"So": "Other_Symbol", "Other_Symbol": "Other_Symbol",
	"Z": "Separator", "Separator": "Separator",
	"Zl": "Line_Separator", "Line_Separator": "Line_Separator",
	"Zp": "Paragraph_Separator", "Paragraph_Separator": "Paragraph_Separator",
	"Zs": "Space_Separator", "Space_Separator": "Space_Separator",
}

// familyNames is every spelling cohere lists for a property of family:
// "Binary", "General_Category", "Script" or "Script_Extensions".
func familyNames(family string) []string {
	var names []string
	for _, name := range cohere.Names() {
		if property, _ := cohere.Lookup(name); property.Family == family {
			names = append(names, name)
		}
	}
	return names
}

// An alias is its canonical name's property: the same Set, and no other
// canonical name's.
func TestBinaryAliases(t *testing.T) {
	t.Parallel()
	binary := familyNames("Binary")
	if len(binary) != len(specBinary) {
		t.Fatalf("%d binary names, spec lists %d", len(binary), len(specBinary))
	}
	canonicalSets := map[*Set]string{}
	for alias, canonical := range specBinary {
		property, ok := Lookup(alias, false)
		named, _ := Lookup(canonical, false)
		if !ok || property.Kind != KindCodePoints || property.Set == nil || property.Set != named.Set {
			t.Errorf("\\p{%s}: ok %v, and not the set of %s", alias, ok, canonical)
		}
		if other, taken := canonicalSets[named.Set]; taken && other != canonical {
			t.Errorf("%s and %s are one set", canonical, other)
		}
		canonicalSets[named.Set] = canonical
		if _, ok := Lookup(alias+"=Yes", false); ok {
			t.Errorf("\\p{%s=Yes} is accepted; Node rejects a value on a binary property", alias)
		}
	}
	for _, alias := range binary {
		if _, ok := specBinary[alias]; !ok {
			t.Errorf("table has %s, which the spec list does not", alias)
		}
	}
}

func TestGeneralCategoryAliases(t *testing.T) {
	t.Parallel()
	// Each value is spelled alone, after gc= and after General_Category=.
	categories := familyNames("General_Category")
	if len(categories) != 3*len(specGeneralCategory) {
		t.Fatalf("%d general category names, spec lists %d values, each three ways", len(categories), len(specGeneralCategory))
	}
	for _, name := range categories {
		value := strings.TrimPrefix(strings.TrimPrefix(name, "General_Category="), "gc=")
		if _, ok := specGeneralCategory[value]; !ok {
			t.Errorf("table has %s, which the spec list does not", name)
		}
	}
	canonicalSets := map[*Set]string{}
	for alias, canonical := range specGeneralCategory {
		named, _ := Lookup("General_Category="+canonical, false)
		for _, expression := range []string{alias, "gc=" + alias, "General_Category=" + alias} {
			property, ok := Lookup(expression, false)
			if !ok || property.Kind != KindCodePoints || property.Set == nil || property.Set != named.Set {
				t.Errorf("\\p{%s}: ok %v, and not the set of %s", expression, ok, canonical)
			}
		}
		if other, taken := canonicalSets[named.Set]; taken && other != canonical {
			t.Errorf("%s and %s are one set", canonical, other)
		}
		canonicalSets[named.Set] = canonical
	}
}

// Script aliases are every short name, long name and extra alias in
// PropertyValueAliases.txt except Hrkt (Katakana_Or_Hiragana), which no code
// point has and which Node rejects. Scripts whose short and long names are the
// same word (Ahom, Cham, and six others) are counted once. The digest is of
// the sorted names, so a missing alias changes it. 344 = 175 scripts, with
// eight of those using one spelling for both names, plus Qaac and Qaai.
func TestScriptAliasSet(t *testing.T) {
	t.Parallel()
	const wantCount = 344
	const wantDigest = "717bd8ab5e94ba064f3187a545e08be10ba1f5bc5da3ba93d419f74ef13126de"
	// Each value is spelled after sc= and after Script=, sorted, so the sc= ones
	// are the values in sorted order. Script_Extensions is the same, after scx=
	// and Script_Extensions=.
	var names []string
	for _, name := range familyNames("Script") {
		if value, short := strings.CutPrefix(name, "sc="); short {
			names = append(names, value)
		}
	}
	sum := sha256.Sum256([]byte(strings.Join(names, "\n")))
	got := hex.EncodeToString(sum[:])
	if len(names) != wantCount || got != wantDigest {
		t.Fatalf("script aliases: %d names, sha256 %s", len(names), got)
	}
	if scripts, extensions := len(familyNames("Script")), len(familyNames("Script_Extensions")); scripts != 2*wantCount || extensions != scripts {
		t.Fatalf("script spellings %d, script extensions spellings %d, want %d each", scripts, extensions, 2*wantCount)
	}
	for _, required := range []string{"Latin", "Latn", "Unknown", "Zzzz", "Inherited", "Zinh", "Qaai", "Common", "Zyyy", "Coptic", "Copt", "Qaac", "Adlam", "Adlm"} {
		script, ok := Lookup("sc="+required, false)
		extensions, extOK := Lookup("scx="+required, false)
		if !ok || !extOK || script.Set == extensions.Set {
			t.Errorf("script %s: sc %v scx %v same set %v", required, ok, extOK, ok && extOK && script.Set == extensions.Set)
		}
		if _, ok := Lookup("Script="+required, false); !ok {
			t.Errorf("Script=%s rejected", required)
		}
		if _, ok := Lookup("Script_Extensions="+required, false); !ok {
			t.Errorf("Script_Extensions=%s rejected", required)
		}
	}
	// A script name alone is not a property. Latin is not a General_Category value.
	if _, ok := Lookup("Latin", false); ok {
		t.Error(`\p{Latin} accepted; a script needs Script= or sc=`)
	}
	if _, ok := Lookup("sc=Hrkt", false); ok {
		t.Error(`\p{sc=Hrkt} accepted; Node rejects Katakana_Or_Hiragana`)
	}
	if _, ok := Lookup("sc=Katakana_Or_Hiragana", false); ok {
		t.Error(`\p{sc=Katakana_Or_Hiragana} accepted`)
	}
}

func TestRejectedNames(t *testing.T) {
	t.Parallel()
	rejected := []string{
		"", "=", "ascii", "ASCII=Yes", "ASCII=Y", "ASCII=No", "White_Space=Yes",
		"Hyphen", "Block=Basic_Latin", "InBasic_Latin", "Age=1.1", "gc=ll",
		"gc=LL", "Cntrl", "Script_Extensions", "General_Category", "gc", "sc",
		"sc=xpeo", "sc=Old_Persian ", "Wspace", "wspace", "Emoji_Presentation=Yes",
		"RGI_Emoji", "Basic_Emoji", "RGI_Emoji=Yes",
	}
	for _, expression := range rejected {
		if _, ok := Lookup(expression, false); ok {
			t.Errorf("accepted %q without the v flag", expression)
		}
	}
	if _, ok := Lookup("RGI_Emoji", true); !ok {
		t.Error(`\p{RGI_Emoji} rejected with the v flag`)
	}
	if _, ok := Lookup("Basic_Emoji=Yes", true); ok {
		t.Error(`\p{Basic_Emoji=Yes} accepted`)
	}
	if _, ok := Lookup("Basic_Emoji", false); ok {
		t.Error(`\p{Basic_Emoji} accepted without the v flag`)
	}
	// The same string property is still rejected as a value form with the v flag,
	// and a code point property is not affected by the flag.
	if property, ok := Lookup("ASCII", true); !ok || property.Kind != KindCodePoints {
		t.Error(`\p{ASCII} with the v flag is not the binary property`)
	}
}

// stringPropertyCensus is the sequence count and the sha256 of the sequences
// joined by newlines. A generator that drops one sequence, or a table someone
// edits by hand, disagrees with this. The Node test then checks each sequence
// that remains actually matches.
var stringPropertyCensus = map[string]struct {
	count  int
	digest string
}{
	"Basic_Emoji":                 {1400, "5a66f7a20529ae7e948f2af51a291961e0f5f44a92f12dd6a4138d2f00027131"},
	"Emoji_Keycap_Sequence":       {12, "defc0fce01656a8b0b733cb57a58a9954345b6987f12dfad45cf9d42f3ebdfe3"},
	"RGI_Emoji_Modifier_Sequence": {665, "c5301d56b0753510b94b7bbfe2f6542fb3ffc45036d49a8b983c57dad31ac4f2"},
	"RGI_Emoji_Flag_Sequence":     {259, "11ef68a2075eb039f4996360cc7d8a2a37424fce1aa74322183c7e4f0ac60b07"},
	"RGI_Emoji_Tag_Sequence":      {3, "65b22ba0aaa68493d7d80906607a51cbc18d224451a86abd1a6c8aeba90c1373"},
	"RGI_Emoji_ZWJ_Sequence":      {1614, "934fc66f03ad8f55f9430aba8e40df016cc3d286546e9218e8ccd83e30a3720f"},
	"RGI_Emoji":                   {3953, "e59a1ad6cad887d2c1ac15cb04859ec50aad8b0c1ef53d9714a586bd5c2ff400"},
}

func TestStringPropertyCensus(t *testing.T) {
	t.Parallel()
	if len(stringProperties) != len(stringPropertyCensus) {
		t.Fatalf("%d string properties, want %d", len(stringProperties), len(stringPropertyCensus))
	}
	for name, want := range stringPropertyCensus {
		sequences, ok := stringProperties[name]
		if !ok {
			t.Errorf("missing %s", name)
			continue
		}
		sum := sha256.Sum256([]byte(strings.Join(sequences, "\n")))
		got := hex.EncodeToString(sum[:])
		if len(sequences) != want.count || got != want.digest {
			t.Errorf("%s: %d sequences, sha256 %s", name, len(sequences), got)
		}
		resolved, ok := Lookup(name, true)
		if !ok || resolved.Kind != KindStrings || len(resolved.Sequences) != len(sequences) {
			t.Errorf("Lookup %s: ok %v kind %v len %d", name, ok, resolved.Kind, len(resolved.Sequences))
		}
		if len(sequences) > 0 && !resolved.ContainsSequence(sequences[0]) {
			t.Errorf("%s does not contain its first sequence", name)
		}
		if resolved.ContainsSequence(sequences[0] + "\u0000") {
			t.Errorf("%s contains a sequence with a NUL added", name)
		}
	}
}

// Every set Lookup answers keeps a Set's shape: sorted, and no two ranges
// overlapping or adjacent, which cohere promises only in part (sorted and
// disjoint) and does by merging.
func TestSetBoundaries(t *testing.T) {
	t.Parallel()
	checked := map[*Set]bool{}
	for _, name := range cohere.Names() {
		property, _ := Lookup(name, false)
		set := property.Set
		if checked[set] {
			continue
		}
		checked[set] = true
		previous := int64(-2)
		for _, r := range set.Ranges {
			if int64(r.Start) <= previous || r.Start > r.End || r.End > 0x10FFFF {
				t.Fatalf("%s range %#v after end %d", name, r, previous)
			}
			if !set.Contains(rune(r.Start)) || !set.Contains(rune(r.End)) {
				t.Fatalf("%s misses an endpoint of %#v", name, r)
			}
			if r.Start > 0 && set.Contains(rune(r.Start-1)) {
				t.Fatalf("%s contains U+%04X, just before %#v", name, r.Start-1, r)
			}
			if r.End < 0x10FFFF && set.Contains(rune(r.End+1)) {
				t.Fatalf("%s contains U+%04X, just after %#v", name, r.End+1, r)
			}
			previous = int64(r.End)
		}
		complement := set.Complement()
		if set.Contains(0) == complement.Contains(0) || set.Contains(0x10FFFF) == complement.Contains(0x10FFFF) {
			t.Fatalf("%s complement agrees at an end", name)
		}
		back := complement.Complement()
		if len(back.Ranges) != len(set.Ranges) {
			t.Fatalf("%s complement twice has %d ranges, want %d", name, len(back.Ranges), len(set.Ranges))
		}
		for i := range set.Ranges {
			if back.Ranges[i] != set.Ranges[i] {
				t.Fatalf("%s complement twice differs at %d", name, i)
			}
		}
	}
	if len(checked) != 441 {
		t.Fatalf("%d sets, want 441: 53 binary, 38 General_Category values, 175 scripts and their extensions", len(checked))
	}
}

func TestKnownMembership(t *testing.T) {
	t.Parallel()
	must := func(expression string, codePoint rune, want bool) {
		t.Helper()
		property, ok := Lookup(expression, false)
		if !ok {
			t.Errorf("Lookup %s failed", expression)
			return
		}
		if property.Contains(codePoint) != want {
			t.Errorf("\\p{%s} U+%04X = %v, want %v", expression, codePoint, property.Contains(codePoint), want)
		}
	}
	must("L", 'A', true)
	must("Lu", 'A', true)
	must("Ll", 'A', false)
	must("gc=Ll", 'a', true)
	must("ASCII", 'A', true)
	must("ASCII", 0x80, false)
	must("Any", 0, true)
	must("Any", 0x10FFFF, true)
	must("Any", 0xD800, true)
	must("Assigned", 0xD800, true)
	must("Assigned", 0xFFFF, false)
	must("gc=Cn", 0xFFFF, true)
	must("NChar", 0xFFFF, true)
	must("Cs", 0xD800, true)
	must("Cs", 'A', false)
	must("sc=Latin", 'A', true)
	must("sc=Latin", 0x00B7, false)
	must("sc=Common", 0x00B7, true)
	must("scx=Common", 0x00B7, false)
	must("scx=Latin", 0x00B7, true)
	must("sc=Inherited", 0x0300, true)
	must("scx=Inherited", 0x0300, false)
	must("scx=Qaai", 0x0300, false)
	must("scx=Latn", 0x0300, true)
	must("Script=Unknown", 0xD800, true)
	must("Emoji", '#', true)
	must("White_Space", ' ', true)
	must("space", '\t', true)
	must("WSpace", '\n', true)
	must("Math", '+', true)

	watch, ok := Lookup("Basic_Emoji", true)
	if !ok || !watch.ContainsSequence("\u231A") || watch.ContainsSequence("A") {
		t.Errorf("Basic_Emoji watch: ok %v", ok)
	}
	keycap, ok := Lookup("Emoji_Keycap_Sequence", true)
	if !ok || !keycap.ContainsSequence("#\uFE0F\u20E3") || keycap.ContainsSequence("#\u20E3") {
		t.Errorf("keycap: ok %v sequences %d", ok, len(keycap.Sequences))
	}
}
