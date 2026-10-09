//go:build ignore

// generate writes tables.go from the Unicode Character Database of the version
// Node reports (process.versions.unicode). Node 24 says 17.0, and the files
// for that version are published as 17.0.0. The same run writes
// canonicalize_tables.go: simple case folding for the u and v flags, and the
// legacy single-code-unit Canonicalize for the i flag without them.
//
//	go run generate.go              downloads the files into a scratch directory
//	go run generate.go -data DIR    reads them from DIR instead
//
// The scratch directory is not the repository: only the two Go files are
// written, and each input is pinned by SHA-256 so a different file cannot be
// silent.
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"go/format"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

const unicodeVersion = "17.0.0"

// pinned is each UCD file's SHA-256. UnicodeData.txt's hash is the same pin
// internal/native's case tables use, so the two generators cannot drift apart
// onto different copies of Unicode 17.
var pinned = map[string]string{
	"UnicodeData.txt":               "2e1efc1dcb59c575eedf5ccae60f95229f706ee6d031835247d843c11d96470c",
	"CaseFolding.txt":               "ff8d8fefbf123574205085d6714c36149eb946d717a0c585c27f0f4ef58c4183",
	"SpecialCasing.txt":             "efc25faf19de21b92c1194c111c932e03d2a5eaf18194e33f1156e96de4c9588",
	"PropList.txt":                  "130dcddcaadaf071008bdfce1e7743e04fdfbc910886f017d9f9ac931d8c64dd",
	"DerivedCoreProperties.txt":     "24c7fed1195c482faaefd5c1e7eb821c5ee1fb6de07ecdbaa64b56a99da22c08",
	"DerivedNormalizationProps.txt": "71fd6a206a2c0cdd41feb6b7f656aa31091db45e9cedc926985d718397f9e488",
	"Scripts.txt":                   "9f5e50d3abaee7d6ce09480f325c706f485ae3240912527e651954d2d6b035bf",
	"ScriptExtensions.txt":          "ec2107e58825a1586acee8e0911ce18260394ac8b87e535ca325f1ccbeb06bc6",
	"PropertyAliases.txt":           "4441f573caf952ffece1d7c892e7715bd7136dfc26f96eb6f268bf1e474715fb",
	"PropertyValueAliases.txt":      "64e9a5f76f7a1e8b5a47d6a1f9a26522a251208f5276bdfa1559dac7cf2e827a",
	"emoji-data.txt":                "2cb2bb9455cda83e8481541ecf5b6dfda66a3bb89efa3fa7c5297eccf607b72b",
	"emoji-sequences.txt":           "12cc8267dc33cbd11ed32bcf6fc5dc2ad9c7a77bae1bdfba2f41b1b9b3ead8dd",
	"emoji-zwj-sequences.txt":       "5b25441daed2322b068c5e70cda522946a4f0274df864445a1965a92e5fc5cad",
}

var sourceURL = map[string]string{
	"UnicodeData.txt":               "https://www.unicode.org/Public/17.0.0/ucd/UnicodeData.txt",
	"CaseFolding.txt":               "https://www.unicode.org/Public/17.0.0/ucd/CaseFolding.txt",
	"SpecialCasing.txt":             "https://www.unicode.org/Public/17.0.0/ucd/SpecialCasing.txt",
	"PropList.txt":                  "https://www.unicode.org/Public/17.0.0/ucd/PropList.txt",
	"DerivedCoreProperties.txt":     "https://www.unicode.org/Public/17.0.0/ucd/DerivedCoreProperties.txt",
	"DerivedNormalizationProps.txt": "https://www.unicode.org/Public/17.0.0/ucd/DerivedNormalizationProps.txt",
	"Scripts.txt":                   "https://www.unicode.org/Public/17.0.0/ucd/Scripts.txt",
	"ScriptExtensions.txt":          "https://www.unicode.org/Public/17.0.0/ucd/ScriptExtensions.txt",
	"PropertyAliases.txt":           "https://www.unicode.org/Public/17.0.0/ucd/PropertyAliases.txt",
	"PropertyValueAliases.txt":      "https://www.unicode.org/Public/17.0.0/ucd/PropertyValueAliases.txt",
	"emoji-data.txt":                "https://www.unicode.org/Public/17.0.0/ucd/emoji/emoji-data.txt",
	"emoji-sequences.txt":           "https://www.unicode.org/Public/17.0.0/emoji/emoji-sequences.txt",
	"emoji-zwj-sequences.txt":       "https://www.unicode.org/Public/17.0.0/emoji/emoji-zwj-sequences.txt",
}

// binaryProperty is one row of ECMA-262's table of binary Unicode properties,
// plus the alias spellings Node accepts for it. The canonical name is the
// long name. source is which file lists the code points where the property is
// true; ascii, any and assigned are defined by the spec rather than a file.
//
// WSpace is the short name of White_Space in PropertyAliases.txt, and Node
// accepts \p{WSpace}. The spec's HTML table lists White_Space and space but
// omits WSpace (tc39/ecma262#3286, still open against the text fetched with
// these tables). PropertyAliases is what that table says it spells, and
// leaving WSpace out would make \p{WSpace} a syntax error where Node runs it.
type binaryProperty struct {
	canonical string
	aliases   []string
	source    string
}

var binaryProperties = []binaryProperty{
	{"ASCII", nil, "ascii"},
	{"ASCII_Hex_Digit", []string{"AHex"}, "PropList.txt"},
	{"Alphabetic", []string{"Alpha"}, "DerivedCoreProperties.txt"},
	{"Any", nil, "any"},
	{"Assigned", nil, "assigned"},
	{"Bidi_Control", []string{"Bidi_C"}, "PropList.txt"},
	{"Bidi_Mirrored", []string{"Bidi_M"}, "UnicodeData.txt"},
	{"Case_Ignorable", []string{"CI"}, "DerivedCoreProperties.txt"},
	{"Cased", nil, "DerivedCoreProperties.txt"},
	{"Changes_When_Casefolded", []string{"CWCF"}, "DerivedCoreProperties.txt"},
	{"Changes_When_Casemapped", []string{"CWCM"}, "DerivedCoreProperties.txt"},
	{"Changes_When_Lowercased", []string{"CWL"}, "DerivedCoreProperties.txt"},
	{"Changes_When_NFKC_Casefolded", []string{"CWKCF"}, "DerivedNormalizationProps.txt"},
	{"Changes_When_Titlecased", []string{"CWT"}, "DerivedCoreProperties.txt"},
	{"Changes_When_Uppercased", []string{"CWU"}, "DerivedCoreProperties.txt"},
	{"Dash", nil, "PropList.txt"},
	{"Default_Ignorable_Code_Point", []string{"DI"}, "DerivedCoreProperties.txt"},
	{"Deprecated", []string{"Dep"}, "PropList.txt"},
	{"Diacritic", []string{"Dia"}, "PropList.txt"},
	{"Emoji", nil, "emoji-data.txt"},
	{"Emoji_Component", []string{"EComp"}, "emoji-data.txt"},
	{"Emoji_Modifier", []string{"EMod"}, "emoji-data.txt"},
	{"Emoji_Modifier_Base", []string{"EBase"}, "emoji-data.txt"},
	{"Emoji_Presentation", []string{"EPres"}, "emoji-data.txt"},
	{"Extended_Pictographic", []string{"ExtPict"}, "emoji-data.txt"},
	{"Extender", []string{"Ext"}, "PropList.txt"},
	{"Grapheme_Base", []string{"Gr_Base"}, "DerivedCoreProperties.txt"},
	{"Grapheme_Extend", []string{"Gr_Ext"}, "DerivedCoreProperties.txt"},
	{"Hex_Digit", []string{"Hex"}, "PropList.txt"},
	{"IDS_Binary_Operator", []string{"IDSB"}, "PropList.txt"},
	{"IDS_Trinary_Operator", []string{"IDST"}, "PropList.txt"},
	{"ID_Continue", []string{"IDC"}, "DerivedCoreProperties.txt"},
	{"ID_Start", []string{"IDS"}, "DerivedCoreProperties.txt"},
	{"Ideographic", []string{"Ideo"}, "PropList.txt"},
	{"Join_Control", []string{"Join_C"}, "PropList.txt"},
	{"Logical_Order_Exception", []string{"LOE"}, "PropList.txt"},
	{"Lowercase", []string{"Lower"}, "DerivedCoreProperties.txt"},
	{"Math", nil, "DerivedCoreProperties.txt"},
	{"Noncharacter_Code_Point", []string{"NChar"}, "PropList.txt"},
	{"Pattern_Syntax", []string{"Pat_Syn"}, "PropList.txt"},
	{"Pattern_White_Space", []string{"Pat_WS"}, "PropList.txt"},
	{"Quotation_Mark", []string{"QMark"}, "PropList.txt"},
	{"Radical", nil, "PropList.txt"},
	{"Regional_Indicator", []string{"RI"}, "PropList.txt"},
	{"Sentence_Terminal", []string{"STerm"}, "PropList.txt"},
	{"Soft_Dotted", []string{"SD"}, "PropList.txt"},
	{"Terminal_Punctuation", []string{"Term"}, "PropList.txt"},
	{"Unified_Ideograph", []string{"UIdeo"}, "PropList.txt"},
	{"Uppercase", []string{"Upper"}, "DerivedCoreProperties.txt"},
	{"Variation_Selector", []string{"VS"}, "PropList.txt"},
	{"White_Space", []string{"space", "WSpace"}, "PropList.txt"},
	{"XID_Continue", []string{"XIDC"}, "DerivedCoreProperties.txt"},
	{"XID_Start", []string{"XIDS"}, "DerivedCoreProperties.txt"},
}

// stringProperties are ECMA-262's table of binary Unicode properties of strings,
// in the table's order. RGI_Emoji is the union of the other six, which is how
// UTS #51 defines it and how emoji-sequences.txt's header says a regular
// expression should read it. None of them has an alias.
var stringPropertyNames = []string{
	"Basic_Emoji",
	"Emoji_Keycap_Sequence",
	"RGI_Emoji_Modifier_Sequence",
	"RGI_Emoji_Flag_Sequence",
	"RGI_Emoji_Tag_Sequence",
	"RGI_Emoji_ZWJ_Sequence",
	"RGI_Emoji",
}

// categoryGroups are the General_Category values that are unions. The lists are
// the file's own comments ("C = Cc | Cf | Cn | Co | Cs" and so on); the
// generator checks the comment still says this before trusting it.
var categoryGroups = map[string][]string{
	"C":  {"Cc", "Cf", "Cn", "Co", "Cs"},
	"L":  {"Ll", "Lm", "Lo", "Lt", "Lu"},
	"LC": {"Ll", "Lt", "Lu"},
	"M":  {"Mc", "Me", "Mn"},
	"N":  {"Nd", "Nl", "No"},
	"P":  {"Pc", "Pd", "Pe", "Pf", "Pi", "Po", "Ps"},
	"S":  {"Sc", "Sk", "Sm", "So"},
	"Z":  {"Zl", "Zp", "Zs"},
}

func main() {
	data := flag.String("data", "", "a directory holding the UCD files, instead of downloading them")
	output := flag.String("output", "tables.go", "where to write the tables")
	flag.Parse()

	directory := *data
	if directory == "" {
		scratch, err := os.MkdirTemp("", "adamic-ucd-")
		if err != nil {
			log.Fatal(err)
		}
		defer os.RemoveAll(scratch)
		directory = scratch
		log.Printf("downloading Unicode %s into %s", unicodeVersion, directory)
		for name, url := range sourceURL {
			if err := download(url, filepath.Join(directory, name)); err != nil {
				log.Fatal(err)
			}
		}
	}

	files := map[string][]byte{}
	for name, sum := range pinned {
		contents, err := os.ReadFile(filepath.Join(directory, name))
		if err != nil {
			log.Fatal(err)
		}
		hash := sha256.Sum256(contents)
		if got := hex.EncodeToString(hash[:]); got != sum {
			log.Fatalf("%s: SHA-256 %s, want %s", name, got, sum)
		}
		files[name] = contents
	}
	for _, name := range []string{
		"DerivedCoreProperties.txt", "DerivedNormalizationProps.txt", "PropList.txt",
		"PropertyAliases.txt", "PropertyValueAliases.txt", "ScriptExtensions.txt", "Scripts.txt",
	} {
		if !bytes.Contains(files[name][:80], []byte(unicodeVersion)) && !bytes.Contains(files[name][:120], []byte(unicodeVersion)) {
			// The version token is on the first line of these files.
			first, _, _ := bytes.Cut(files[name], []byte("\n"))
			if !bytes.Contains(first, []byte(unicodeVersion)) {
				log.Fatalf("%s is not version %s (%s)", name, unicodeVersion, first)
			}
		}
	}
	for _, name := range []string{"emoji-data.txt", "emoji-sequences.txt", "emoji-zwj-sequences.txt"} {
		if !bytes.Contains(files[name][:400], []byte("Version: 17.0")) {
			log.Fatalf("%s does not say Version: 17.0", name)
		}
	}

	checkPropertyAliases(files["PropertyAliases.txt"])

	categories := generalCategories(files["UnicodeData.txt"])
	values := categoryValues(files["PropertyValueAliases.txt"], categories)

	scripts := scriptsOf(files["Scripts.txt"], files["ScriptExtensions.txt"], files["PropertyValueAliases.txt"])

	binary := map[string]*bitSet{}
	for _, file := range []string{"PropList.txt", "DerivedCoreProperties.txt", "DerivedNormalizationProps.txt", "emoji-data.txt"} {
		for name, set := range binaryFile(files[file]) {
			if binary[name] != nil {
				log.Fatalf("property %s listed in two files", name)
			}
			binary[name] = set
		}
	}
	binary["Bidi_Mirrored"] = bidiMirrored(files["UnicodeData.txt"])
	binary["ASCII"] = rangeSet(0x0000, 0x007F)
	any := newBitSet()
	for i := range any.words {
		any.words[i] = ^uint64(0)
	}
	binary["Any"] = any
	assigned := categories["Cn"].clone()
	assigned.not()
	binary["Assigned"] = assigned

	sequences := emojiSequences(files["emoji-sequences.txt"], files["emoji-zwj-sequences.txt"])

	var out bytes.Buffer
	writeTables(&out, values, scripts, binary, sequences)
	formatted, err := format.Source(out.Bytes())
	if err != nil {
		log.Fatalf("tables.go does not format: %v", err)
	}
	if err := os.WriteFile(*output, formatted, 0o644); err != nil {
		log.Fatal(err)
	}
	writeCanonicalize(filepath.Join(filepath.Dir(*output), "canonicalize_tables.go"), files["CaseFolding.txt"], files["SpecialCasing.txt"], files["UnicodeData.txt"])
}

func download(url string, path string) error {
	client := &http.Client{Timeout: 120 * time.Second}
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	request.Header.Set("User-Agent", "adamic-unicodeproperties-generate")
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("%s: %s", url, response.Status)
	}
	contents, err := io.ReadAll(response.Body)
	if err != nil {
		return err
	}
	return os.WriteFile(path, contents, 0o644)
}

// bitSet holds one bit per code point from 0 to U+10FFFF. 0x110000 is
// divisible by 64, so the last word is full and not() needs no mask.
type bitSet struct {
	words [0x110000 / 64]uint64
}

func newBitSet() *bitSet { return &bitSet{} }

func (b *bitSet) set(cp uint32) {
	b.words[cp/64] |= 1 << (cp % 64)
}

func (b *bitSet) get(cp uint32) bool {
	return b.words[cp/64]&(1<<(cp%64)) != 0
}

func (b *bitSet) or(other *bitSet) {
	for i := range b.words {
		b.words[i] |= other.words[i]
	}
}

func (b *bitSet) not() {
	for i := range b.words {
		b.words[i] = ^b.words[i]
	}
}

func (b *bitSet) clone() *bitSet {
	next := *b
	return &next
}

func (b *bitSet) empty() bool {
	for _, word := range b.words {
		if word != 0 {
			return false
		}
	}
	return true
}

func (b *bitSet) ranges() [][2]uint32 {
	var out [][2]uint32
	in := false
	var start uint32
	for cp := uint32(0); ; cp++ {
		if b.get(cp) {
			if !in {
				start = cp
				in = true
			}
		} else if in {
			out = append(out, [2]uint32{start, cp - 1})
			in = false
		}
		if cp == 0x10FFFF {
			break
		}
	}
	if in {
		out = append(out, [2]uint32{start, 0x10FFFF})
	}
	return out
}

func rangeSet(lo, hi uint32) *bitSet {
	set := newBitSet()
	fill(set, lo, hi)
	return set
}

func fill(set *bitSet, lo, hi uint32) {
	for cp := lo; ; cp++ {
		set.set(cp)
		if cp == hi {
			return
		}
	}
}

func point(field string) uint32 {
	value, err := strconv.ParseUint(strings.TrimSpace(field), 16, 32)
	if err != nil || value > 0x10FFFF {
		log.Fatalf("not a code point: %q", field)
	}
	return uint32(value)
}

// span parses "0041" or "0041..005A".
func span(field string) (uint32, uint32) {
	field = strings.TrimSpace(field)
	loField, hiField, ranged := strings.Cut(field, "..")
	lo := point(loField)
	hi := lo
	if ranged {
		hi = point(hiField)
	}
	if lo > hi {
		log.Fatalf("range %04X..%04X is backwards", lo, hi)
	}
	return lo, hi
}

// dataLines is each line's fields, split at semicolons, with comments and
// blank lines dropped and every field trimmed. A field that is only a comment
// marker is dropped.
func dataLines(contents []byte) [][]string {
	var lines [][]string
	scanner := bufio.NewScanner(bytes.NewReader(contents))
	// ScriptExtensions lines are short; emoji names in a comment can be long,
	// and the default 64KiB token is enough, but a raised buffer costs nothing.
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if comment := strings.IndexByte(line, '#'); comment >= 0 {
			line = line[:comment]
		}
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, ";")
		for index := range fields {
			fields[index] = strings.TrimSpace(fields[index])
		}
		lines = append(lines, fields)
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	return lines
}

// generalCategories is each single-letter-pair category (Ll, Lu, Cn, ...) as a
// bit set. Code points UnicodeData.txt does not mention are Unassigned (Cn),
// which includes the noncharacters. A First/Last pair covers the whole span;
// the generator stops if the two ends disagree, rather than guess.
func generalCategories(contents []byte) map[string]*bitSet {
	categories := map[string]*bitSet{}
	for _, name := range []string{
		"Cc", "Cf", "Cn", "Co", "Cs", "Ll", "Lm", "Lo", "Lt", "Lu", "Mc", "Me", "Mn",
		"Nd", "Nl", "No", "Pc", "Pd", "Pe", "Pf", "Pi", "Po", "Ps", "Sc", "Sk", "Sm", "So",
		"Zl", "Zp", "Zs",
	} {
		categories[name] = newBitSet()
	}
	// Everything starts unassigned, and a line claims its code points.
	fill(categories["Cn"], 0, 0x10FFFF)

	var pending uint32
	var pendingCategory string
	var pendingMirror string
	open := false
	seen := newBitSet()
	for _, fields := range dataLines(contents) {
		if len(fields) != 15 {
			log.Fatalf("UnicodeData.txt: %d fields in %q", len(fields), fields)
		}
		cp := point(fields[0])
		category := fields[2]
		set := categories[category]
		if set == nil {
			log.Fatalf("UnicodeData.txt: unknown category %q", category)
		}
		name := fields[1]
		switch {
		case strings.HasSuffix(name, ", First>"):
			if open {
				log.Fatalf("UnicodeData.txt: %s opened a range inside one", name)
			}
			pending, pendingCategory, pendingMirror = cp, category, fields[9]
			open = true
		case strings.HasSuffix(name, ", Last>"):
			if !open || category != pendingCategory || fields[9] != pendingMirror || cp < pending {
				log.Fatalf("UnicodeData.txt: %s does not close %04X %s", name, pending, pendingCategory)
			}
			for code := pending; ; code++ {
				if seen.get(code) {
					log.Fatalf("UnicodeData.txt: U+%04X listed twice", code)
				}
				seen.set(code)
				categories["Cn"].clear(code)
				set.set(code)
				if code == cp {
					break
				}
			}
			open = false
		default:
			if open {
				log.Fatalf("UnicodeData.txt: %s inside an open range", name)
			}
			if seen.get(cp) {
				log.Fatalf("UnicodeData.txt: U+%04X listed twice", cp)
			}
			seen.set(cp)
			categories["Cn"].clear(cp)
			set.set(cp)
		}
	}
	if open {
		log.Fatal("UnicodeData.txt: a range was left open")
	}
	return categories
}

func (b *bitSet) clear(cp uint32) {
	b.words[cp/64] &^= 1 << (cp % 64)
}

// bidiMirrored is UnicodeData.txt field 9, which is Y or N. The category pass
// already rejected a range whose ends disagree about it.
func bidiMirrored(contents []byte) *bitSet {
	set := newBitSet()
	var pending uint32
	open := false
	for _, fields := range dataLines(contents) {
		cp := point(fields[0])
		yes := fields[9] == "Y"
		if fields[9] != "Y" && fields[9] != "N" {
			log.Fatalf("UnicodeData.txt: Bidi_Mirrored %q", fields[9])
		}
		name := fields[1]
		switch {
		case strings.HasSuffix(name, ", First>"):
			pending, open = cp, true
		case strings.HasSuffix(name, ", Last>"):
			if yes {
				fill(set, pending, cp)
			}
			open = false
		default:
			if yes {
				set.set(cp)
			}
		}
	}
	if open {
		log.Fatal("UnicodeData.txt: a mirrored range was left open")
	}
	return set
}

type categoryValue struct {
	short   string
	long    string
	aliases []string // short, long, then any extra, in the file's order
	set     *bitSet
}

// categoryValues is every General_Category value and alias in PropertyValueAliases.txt.
// A grouped value (L, LC, P, ...) is the union its comment names.
func categoryValues(contents []byte, categories map[string]*bitSet) []categoryValue {
	var values []categoryValue
	for _, fields := range dataLines(contents) {
		if fields[0] != "gc" {
			continue
		}
		if len(fields) < 3 {
			log.Fatalf("PropertyValueAliases.txt: gc line %q", fields)
		}
		short, long := fields[1], fields[2]
		names := append([]string{}, fields[1:]...)
		var set *bitSet
		if members, grouped := categoryGroups[short]; grouped {
			comment := categoryComment(contents, short)
			if comment != "" && comment != strings.Join(members, " | ") {
				log.Fatalf("General_Category %s is documented as %q, not %q", short, comment, strings.Join(members, " | "))
			}
			set = newBitSet()
			for _, member := range members {
				part := categories[member]
				if part == nil {
					log.Fatalf("General_Category %s unions unknown %s", short, member)
				}
				set.or(part)
			}
		} else {
			set = categories[short]
			if set == nil {
				log.Fatalf("General_Category %s is not a category UnicodeData.txt uses", short)
			}
		}
		values = append(values, categoryValue{short: short, long: long, aliases: names, set: set})
	}
	if len(values) != len(categoryGroups)+len(categories) {
		log.Fatalf("PropertyValueAliases.txt has %d gc values, want %d", len(values), len(categoryGroups)+len(categories))
	}
	return values
}

// categoryComment is the union a grouped gc line documents after '#',
// for example "Cc | Cf | Cn | Co | Cs". Lines that don't say return "".
func categoryComment(contents []byte, short string) string {
	for _, line := range strings.Split(string(contents), "\n") {
		body, comment, found := strings.Cut(line, "#")
		if !found {
			continue
		}
		fields := strings.Split(body, ";")
		if len(fields) < 2 {
			continue
		}
		if strings.TrimSpace(fields[0]) == "gc" && strings.TrimSpace(fields[1]) == short {
			return strings.TrimSpace(comment)
		}
	}
	return ""
}

type scriptValue struct {
	short   string
	long    string
	aliases []string
	script  *bitSet
	extend  *bitSet
}

// scriptsOf is Script and Script_Extensions for every script PropertyValueAliases
// lists. ScriptExtensions.txt names scripts by their short name and replaces
// the Script value for the code points it lists; every other code point's
// extension is the singleton of its Script. Code points Scripts.txt skips are
// Unknown, which is the file's @missing value.
func scriptsOf(scriptsFile, extensionsFile, aliasesFile []byte) []scriptValue {
	var values []scriptValue
	byShort := map[string]int{}
	byLong := map[string]int{}
	for _, fields := range dataLines(aliasesFile) {
		if fields[0] != "sc" {
			continue
		}
		if len(fields) < 3 {
			log.Fatalf("PropertyValueAliases.txt: sc line %q", fields)
		}
		value := scriptValue{short: fields[1], long: fields[2], aliases: append([]string{}, fields[1:]...), script: newBitSet(), extend: newBitSet()}
		if _, ok := byShort[value.short]; ok {
			log.Fatalf("duplicate script %s", value.short)
		}
		byShort[value.short] = len(values)
		byLong[value.long] = len(values)
		values = append(values, value)
	}
	unknown, ok := byLong["Unknown"]
	if !ok {
		log.Fatal("PropertyValueAliases.txt has no Unknown script")
	}

	scriptOf := make([]uint16, 0x110000)
	for i := range scriptOf {
		scriptOf[i] = uint16(unknown)
	}
	for _, fields := range dataLines(scriptsFile) {
		if len(fields) < 2 {
			log.Fatalf("Scripts.txt: %q", fields)
		}
		index, ok := byLong[fields[1]]
		if !ok {
			log.Fatalf("Scripts.txt names %q, which is not a script alias", fields[1])
		}
		lo, hi := span(fields[0])
		for cp := lo; ; cp++ {
			if scriptOf[cp] != uint16(unknown) && int(scriptOf[cp]) != index {
				log.Fatalf("Scripts.txt assigns U+%04X twice", cp)
			}
			scriptOf[cp] = uint16(index)
			values[index].script.set(cp)
			if cp == hi {
				break
			}
		}
	}
	// Anything still Unknown, including the explicit Unknown lines just applied.
	for cp := uint32(0); cp <= 0x10FFFF; cp++ {
		if int(scriptOf[cp]) == unknown {
			values[unknown].script.set(cp)
		}
	}

	extended := map[uint32][]int{}
	for _, fields := range dataLines(extensionsFile) {
		if len(fields) < 2 {
			log.Fatalf("ScriptExtensions.txt: %q", fields)
		}
		var ids []int
		for _, name := range strings.Fields(fields[1]) {
			index, ok := byShort[name]
			if !ok {
				log.Fatalf("ScriptExtensions.txt names %q, which is not a short script name", name)
			}
			ids = append(ids, index)
		}
		lo, hi := span(fields[0])
		for cp := lo; ; cp++ {
			if extended[cp] != nil {
				log.Fatalf("ScriptExtensions.txt lists U+%04X twice", cp)
			}
			extended[cp] = ids
			if cp == hi {
				break
			}
		}
	}
	for cp := uint32(0); ; cp++ {
		if ids, ok := extended[cp]; ok {
			for _, id := range ids {
				values[id].extend.set(cp)
			}
		} else {
			values[scriptOf[cp]].extend.set(cp)
		}
		if cp == 0x10FFFF {
			break
		}
	}
	return values
}

// binaryFile reads a PropList-shaped file: code point or range, semicolon,
// property name. The name is the long one. Lines for properties ECMAScript
// does not use are kept only so a property we do use can be found by name;
// the caller looks up the ones it wants.
func binaryFile(contents []byte) map[string]*bitSet {
	sets := map[string]*bitSet{}
	for _, fields := range dataLines(contents) {
		if len(fields) < 2 {
			log.Fatalf("binary property line %q", fields)
		}
		name := fields[1]
		set := sets[name]
		if set == nil {
			set = newBitSet()
			sets[name] = set
		}
		lo, hi := span(fields[0])
		fill(set, lo, hi)
	}
	return sets
}

// checkPropertyAliases makes the binary alias list equal PropertyAliases.txt
// for those properties. ASCII, Any and Assigned are not in that file; they
// exist because the spec says so. Any other disagreement stops the generator
// instead of shipping a name Node and the file would not agree on.
func checkPropertyAliases(contents []byte) {
	found := map[string]map[string]bool{}
	for _, fields := range dataLines(contents) {
		if len(fields) < 2 {
			log.Fatalf("PropertyAliases.txt: %q", fields)
		}
		long := fields[1]
		names := map[string]bool{}
		for _, field := range fields {
			names[field] = true
		}
		found[long] = names
	}
	for _, property := range binaryProperties {
		switch property.source {
		case "ascii", "any", "assigned":
			if found[property.canonical] != nil {
				log.Fatalf("%s is in PropertyAliases.txt; the spec defines it itself", property.canonical)
			}
			continue
		}
		want := map[string]bool{property.canonical: true}
		for _, alias := range property.aliases {
			want[alias] = true
		}
		got := found[property.canonical]
		if len(got) != len(want) {
			log.Fatalf("%s aliases: file %v, table %v", property.canonical, keys(got), keys(want))
		}
		for name := range want {
			if !got[name] {
				log.Fatalf("%s: %s is not in PropertyAliases.txt", property.canonical, name)
			}
		}
	}
}

func uniqueNames(names []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, name := range names {
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		out = append(out, name)
	}
	return out
}

func keys(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for name := range set {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func emojiSequences(sequencesFile, zwjFile []byte) map[string][]string {
	out := map[string][]string{}
	take := func(contents []byte) {
		for _, fields := range dataLines(contents) {
			if len(fields) < 2 {
				log.Fatalf("emoji sequence line %q", fields)
			}
			name := fields[1]
			for _, sequence := range expandSequence(fields[0]) {
				out[name] = append(out[name], sequence)
			}
		}
	}
	take(sequencesFile)
	take(zwjFile)
	for _, name := range stringPropertyNames {
		if name == "RGI_Emoji" {
			continue
		}
		if len(out[name]) == 0 {
			log.Fatalf("no sequences for %s", name)
		}
		out[name] = uniqueSorted(out[name])
	}
	var union []string
	for _, name := range stringPropertyNames {
		if name == "RGI_Emoji" {
			continue
		}
		union = append(union, out[name]...)
	}
	out["RGI_Emoji"] = uniqueSorted(union)
	return out
}

func expandSequence(field string) []string {
	if strings.Contains(field, "..") {
		if strings.Contains(field, " ") {
			log.Fatalf("emoji range with a space: %q", field)
		}
		lo, hi := span(field)
		var sequences []string
		for cp := lo; ; cp++ {
			sequences = append(sequences, string(rune(cp)))
			if cp == hi {
				break
			}
		}
		return sequences
	}
	parts := strings.Fields(field)
	runes := make([]rune, len(parts))
	for i, part := range parts {
		cp := point(part)
		if cp >= 0xD800 && cp <= 0xDFFF {
			log.Fatalf("emoji sequence has a surrogate: %q", field)
		}
		runes[i] = rune(cp)
	}
	return []string{string(runes)}
}

func uniqueSorted(sequences []string) []string {
	sort.Strings(sequences)
	out := sequences[:0]
	previous := ""
	for index, sequence := range sequences {
		if index > 0 && sequence == previous {
			continue
		}
		out = append(out, sequence)
		previous = sequence
	}
	return out
}

type emittedSet struct {
	comment string
	ranges  [][2]uint32
}

type emittedName struct {
	alias string
	entry string // Go struct literal body, without braces: name, value, set index
}

func writeTables(out *bytes.Buffer, categories []categoryValue, scripts []scriptValue, binary map[string]*bitSet, sequences map[string][]string) {
	var sets []emittedSet
	add := func(comment string, set *bitSet) int {
		if set == nil || set.empty() {
			log.Fatalf("%s is empty", comment)
		}
		sets = append(sets, emittedSet{comment: comment, ranges: set.ranges()})
		return len(sets) - 1
	}

	fmt.Fprintf(out, "// Code generated by generate.go from Unicode %s. Do not edit.\n", unicodeVersion)
	fmt.Fprintf(out, "// Run go generate in internal/unicodeproperties.\n")
	fmt.Fprintf(out, "//\n")
	names := make([]string, 0, len(pinned))
	for name := range pinned {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		fmt.Fprintf(out, "// %s SHA-256 %s\n", name, pinned[name])
	}
	fmt.Fprintf(out, "\npackage unicodeproperties\n\n")

	binaryIndex := map[string]int{}
	for _, property := range binaryProperties {
		set := binary[property.canonical]
		if set == nil || set.empty() {
			log.Fatalf("binary property %s has no code points (source %s)", property.canonical, property.source)
		}
		binaryIndex[property.canonical] = add(property.canonical, set)
	}
	categoryIndex := make([]int, len(categories))
	for i, value := range categories {
		categoryIndex[i] = add("gc="+value.long, value.set)
	}
	scriptIndex := make([]int, len(scripts))
	extendIndex := make([]int, len(scripts))
	used := make([]bool, len(scripts))
	for i, value := range scripts {
		// A script value no code point has, for Script or Script_Extensions, is
		// still listed in PropertyValueAliases.txt (Hrkt / Katakana_Or_Hiragana
		// in 17.0). Node rejects it as an invalid property rather than as an
		// empty set, so it is not a name we accept.
		if value.script.empty() && value.extend.empty() {
			fmt.Fprintf(os.Stderr, "skipping unused script %s (%s)\n", value.long, value.short)
			continue
		}
		used[i] = true
		if value.script.empty() {
			log.Fatalf("script %s has extensions but no Script value; Node's reading of that is untested", value.long)
		}
		if value.extend.empty() {
			log.Fatalf("script %s has a Script value but empty extensions", value.long)
		}
		scriptIndex[i] = add("sc="+value.long, value.script)
		extendIndex[i] = add("scx="+value.long, value.extend)
	}

	fmt.Fprintf(out, "var sets = []*Set{\n")
	totalRanges := 0
	for index, set := range sets {
		totalRanges += len(set.ranges)
		fmt.Fprintf(out, "\t{ // %d %s\n\t\tRanges: []Range{\n", index, set.comment)
		for _, r := range set.ranges {
			fmt.Fprintf(out, "\t\t\t{0x%04X, 0x%04X},\n", r[0], r[1])
		}
		fmt.Fprintf(out, "\t\t},\n\t},\n")
	}
	fmt.Fprintf(out, "}\n\n")

	fmt.Fprintf(out, "var binaryNames = map[string]entry{\n")
	for _, property := range binaryProperties {
		names := append([]string{property.canonical}, property.aliases...)
		sort.Strings(names)
		for _, alias := range names {
			fmt.Fprintf(out, "\t%q: {name: %q, value: \"True\", set: %d},\n", alias, property.canonical, binaryIndex[property.canonical])
		}
	}
	fmt.Fprintf(out, "}\n\n")

	fmt.Fprintf(out, "var generalCategoryNames = map[string]entry{\n")
	for i, value := range categories {
		for _, alias := range uniqueNames(value.aliases) {
			fmt.Fprintf(out, "\t%q: {name: \"General_Category\", value: %q, set: %d},\n", alias, value.long, categoryIndex[i])
		}
	}
	fmt.Fprintf(out, "}\n\n")

	fmt.Fprintf(out, "var scriptNames = map[string]entry{\n")
	for i, value := range scripts {
		if !used[i] {
			continue
		}
		for _, alias := range uniqueNames(value.aliases) {
			fmt.Fprintf(out, "\t%q: {name: \"Script\", value: %q, set: %d},\n", alias, value.long, scriptIndex[i])
		}
	}
	fmt.Fprintf(out, "}\n\n")

	fmt.Fprintf(out, "var scriptExtensionNames = map[string]entry{\n")
	for i, value := range scripts {
		if !used[i] {
			continue
		}
		for _, alias := range uniqueNames(value.aliases) {
			fmt.Fprintf(out, "\t%q: {name: \"Script_Extensions\", value: %q, set: %d},\n", alias, value.long, extendIndex[i])
		}
	}
	fmt.Fprintf(out, "}\n\n")

	for _, name := range stringPropertyNames {
		fmt.Fprintf(out, "var sequences_%s = []string{\n", name)
		for _, sequence := range sequences[name] {
			fmt.Fprintf(out, "\t%s,\n", strconv.Quote(sequence))
		}
		fmt.Fprintf(out, "}\n\n")
	}
	fmt.Fprintf(out, "var stringProperties = map[string]sequences{\n")
	for _, name := range stringPropertyNames {
		fmt.Fprintf(out, "\t%q: {name: %q, sequences: sequences_%s},\n", name, name, name)
	}
	fmt.Fprintf(out, "}\n")

	fmt.Fprintf(os.Stderr, "sets %d, ranges %d\n", len(sets), totalRanges)
	fmt.Fprintf(os.Stderr, "binary names %d, general category values %d, scripts %d\n", len(binaryProperties), len(categories), len(scripts))
	for _, name := range stringPropertyNames {
		fmt.Fprintf(os.Stderr, "%s sequences %d\n", name, len(sequences[name]))
	}
}

// writeCanonicalize writes canonicalize_tables.go. Unicode mode is simple case
// folding: CaseFolding.txt status C and S, which the file's own header says is
// how to fold when a string must not grow. Legacy mode is the full uppercase of
// one UTF-16 code unit (UnicodeData.txt's simple uppercase, overridden by
// SpecialCasing.txt where the uppercase is more than one code point), dropped
// when that uppercase is not a single code unit or when it would turn a
// non-ASCII unit into ASCII.
func writeCanonicalize(path string, caseFolding, specialCasing, unicodeData []byte) {
	if !bytes.HasPrefix(caseFolding, []byte("# CaseFolding-"+unicodeVersion+".txt")) {
		log.Fatalf("CaseFolding.txt is not version %s", unicodeVersion)
	}
	if !bytes.HasPrefix(specialCasing, []byte("# SpecialCasing-"+unicodeVersion+".txt")) {
		log.Fatalf("SpecialCasing.txt is not version %s", unicodeVersion)
	}

	fold := simpleCaseFold(caseFolding)
	legacy := legacyCanonical(unicodeData, specialCasing)

	type pair struct{ from, to uint32 }
	unicodePairs := make([]pair, 0, len(fold))
	for from, to := range fold {
		unicodePairs = append(unicodePairs, pair{from, to})
	}
	sort.Slice(unicodePairs, func(i, j int) bool { return unicodePairs[i].from < unicodePairs[j].from })
	for _, mapping := range unicodePairs {
		if next, mapped := fold[mapping.to]; mapped && next != mapping.to {
			log.Fatalf("simple case fold of U+%04X is U+%04X, which itself folds to U+%04X", mapping.from, mapping.to, next)
		}
	}

	type class struct {
		canon   uint32
		members []uint32
	}
	unicodeGroups := map[uint32][]uint32{}
	for _, mapping := range unicodePairs {
		unicodeGroups[mapping.to] = append(unicodeGroups[mapping.to], mapping.from)
	}
	unicodeClasses := make([]class, 0, len(unicodeGroups))
	for canon, members := range unicodeGroups {
		members = append(members, canon)
		sort.Slice(members, func(i, j int) bool { return members[i] < members[j] })
		unicodeClasses = append(unicodeClasses, class{canon, members})
	}
	sort.Slice(unicodeClasses, func(i, j int) bool { return unicodeClasses[i].canon < unicodeClasses[j].canon })

	legacyPairs := make([]pair, 0)
	legacyGroups := map[uint32][]uint32{}
	for cu := uint32(0); cu <= 0xFFFF; cu++ {
		canon := uint32(legacy[cu])
		if canon == cu {
			continue
		}
		legacyPairs = append(legacyPairs, pair{cu, canon})
		legacyGroups[canon] = append(legacyGroups[canon], cu)
	}
	keptASCII := legacyASCIIKept(unicodeData, specialCasing)
	legacyClasses := make([]class, 0, len(legacyGroups))
	for canon, members := range legacyGroups {
		members = append(members, canon)
		sort.Slice(members, func(i, j int) bool { return members[i] < members[j] })
		legacyClasses = append(legacyClasses, class{canon, members})
	}
	sort.Slice(legacyClasses, func(i, j int) bool { return legacyClasses[i].canon < legacyClasses[j].canon })

	var out bytes.Buffer
	fmt.Fprintf(&out, "// Code generated by generate.go from Unicode %s. Do not edit.\n", unicodeVersion)
	fmt.Fprintf(&out, "// Run go generate in internal/unicodeproperties.\n")
	fmt.Fprintf(&out, "//\n")
	fmt.Fprintf(&out, "// Canonicalize (ECMA-262 22.2.2.7.3). unicodeFold is simple case folding,\n")
	fmt.Fprintf(&out, "// the C and S lines of CaseFolding.txt, used when the u or v flag is set.\n")
	fmt.Fprintf(&out, "// legacyFold is the i flag without them: full uppercase of one UTF-16 code\n")
	fmt.Fprintf(&out, "// unit, kept only when that uppercase is one code unit and does not map a\n")
	fmt.Fprintf(&out, "// non-ASCII unit to ASCII. Each class lists every code point that shares a\n")
	fmt.Fprintf(&out, "// canonical value, which is what a character class under i has to close over.\n")
	fmt.Fprintf(&out, "//\n")
	for _, name := range []string{"CaseFolding.txt", "SpecialCasing.txt", "UnicodeData.txt"} {
		fmt.Fprintf(&out, "// %s SHA-256 %s\n", name, pinned[name])
	}
	fmt.Fprintf(&out, "//\n")
	if len(keptASCII) == 0 {
		fmt.Fprintf(&out, "// No code unit has an uppercase in ASCII that the non-ASCII filter kept.\n")
	} else {
		fmt.Fprintf(&out, "// Non-ASCII code units whose uppercase is ASCII, kept unchanged:")
		for _, cu := range keptASCII {
			fmt.Fprintf(&out, " U+%04X", cu)
		}
		fmt.Fprintf(&out, ".\n")
	}
	fmt.Fprintf(&out, "\npackage unicodeproperties\n\n")

	fmt.Fprintf(&out, "// unicodeFold is from, to. Only mappings that change the code point are stored.\n")
	fmt.Fprintf(&out, "var unicodeFold = [][2]uint32{\n")
	for _, mapping := range unicodePairs {
		fmt.Fprintf(&out, "\t{0x%04X, 0x%04X},\n", mapping.from, mapping.to)
	}
	fmt.Fprintf(&out, "}\n\n")

	fmt.Fprintf(&out, "// unicodeClassCanon[i] is the simple case fold shared by unicodeClass[i].\n")
	fmt.Fprintf(&out, "var unicodeClassCanon = []uint32{\n")
	for _, group := range unicodeClasses {
		fmt.Fprintf(&out, "\t0x%04X,\n", group.canon)
	}
	fmt.Fprintf(&out, "}\n\n")
	fmt.Fprintf(&out, "var unicodeClass = [][]uint32{\n")
	for _, group := range unicodeClasses {
		fmt.Fprintf(&out, "\t{")
		for index, member := range group.members {
			if index > 0 {
				fmt.Fprintf(&out, ", ")
			}
			fmt.Fprintf(&out, "0x%04X", member)
		}
		fmt.Fprintf(&out, "},\n")
	}
	fmt.Fprintf(&out, "}\n\n")

	fmt.Fprintf(&out, "// legacyFold is from, to, for one UTF-16 code unit. Identity is not stored.\n")
	fmt.Fprintf(&out, "var legacyFold = [][2]uint16{\n")
	for _, mapping := range legacyPairs {
		fmt.Fprintf(&out, "\t{0x%04X, 0x%04X},\n", mapping.from, mapping.to)
	}
	fmt.Fprintf(&out, "}\n\n")

	fmt.Fprintf(&out, "// legacyClassCanon[i] is the legacy canonical value shared by legacyClass[i].\n")
	fmt.Fprintf(&out, "var legacyClassCanon = []uint16{\n")
	for _, group := range legacyClasses {
		fmt.Fprintf(&out, "\t0x%04X,\n", group.canon)
	}
	fmt.Fprintf(&out, "}\n\n")
	fmt.Fprintf(&out, "var legacyClass = [][]uint16{\n")
	for _, group := range legacyClasses {
		fmt.Fprintf(&out, "\t{")
		for index, member := range group.members {
			if index > 0 {
				fmt.Fprintf(&out, ", ")
			}
			fmt.Fprintf(&out, "0x%04X", member)
		}
		fmt.Fprintf(&out, "},\n")
	}
	fmt.Fprintf(&out, "}\n")

	formatted, err := format.Source(out.Bytes())
	if err != nil {
		log.Fatalf("canonicalize_tables.go does not format: %v", err)
	}
	if err := os.WriteFile(path, formatted, 0o644); err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(os.Stderr, "simple case folds %d, unicode classes %d\n", len(unicodePairs), len(unicodeClasses))
	fmt.Fprintf(os.Stderr, "legacy non-identity %d, legacy classes %d, non-ASCII-to-ASCII kept %d\n", len(legacyPairs), len(legacyClasses), len(keptASCII))
}

// simpleCaseFold is CaseFolding.txt's C and S mappings, which together are the
// simple case folding. F grows the string and T is Turkic; neither is Canonicalize.
func simpleCaseFold(contents []byte) map[uint32]uint32 {
	fold := map[uint32]uint32{}
	counts := map[string]int{}
	for _, fields := range dataLines(contents) {
		if len(fields) < 3 {
			log.Fatalf("CaseFolding.txt: %q", fields)
		}
		status := fields[1]
		counts[status]++
		switch status {
		case "C", "S", "F", "T":
		default:
			log.Fatalf("CaseFolding.txt: status %q", status)
		}
		if status != "C" && status != "S" {
			continue
		}
		from := point(fields[0])
		parts := strings.Fields(fields[2])
		if len(parts) != 1 {
			log.Fatalf("CaseFolding.txt: %s mapping of U+%04X is not one code point: %q", status, from, fields[2])
		}
		to := point(parts[0])
		if _, exists := fold[from]; exists {
			log.Fatalf("CaseFolding.txt: two simple mappings for U+%04X", from)
		}
		if to != from {
			fold[from] = to
		}
	}
	if counts["C"] == 0 || counts["S"] == 0 {
		log.Fatalf("CaseFolding.txt: expected both C and S mappings, got %v", counts)
	}
	return fold
}

// legacyCanonical is Canonicalize of every UTF-16 code unit. Full uppercase
// comes from SpecialCasing.txt when that mapping has more than one code point,
// and from UnicodeData.txt's simple uppercase otherwise. A single-code-point
// unconditional line in SpecialCasing.txt has to agree with UnicodeData.txt;
// the generator stops rather than pick one.
func legacyCanonical(unicodeData, specialCasing []byte) []uint16 {
	simple := simpleUpper(unicodeData)
	full := specialUpper(specialCasing, simple)
	canon := make([]uint16, 0x10000)
	for cu := uint32(0); cu <= 0xFFFF; cu++ {
		mapped, expanded := full[cu]
		if !expanded {
			if upper, has := simple[cu]; has {
				mapped = []uint32{upper}
			} else {
				mapped = []uint32{cu}
			}
		}
		canon[cu] = uint16(keepLegacyUpper(cu, mapped))
	}
	for cu := uint32(0); cu <= 0xFFFF; cu++ {
		next := canon[cu]
		if canon[next] != next {
			log.Fatalf("legacy Canonicalize is not idempotent: U+%04X -> U+%04X -> U+%04X", cu, next, canon[next])
		}
	}
	return canon
}

func simpleUpper(contents []byte) map[uint32]uint32 {
	upper := map[uint32]uint32{}
	for _, fields := range dataLines(contents) {
		if len(fields) != 15 {
			log.Fatalf("UnicodeData.txt: %d fields in %q", len(fields), fields)
		}
		if fields[12] == "" {
			continue
		}
		name := fields[1]
		if strings.HasSuffix(name, ", First>") || strings.HasSuffix(name, ", Last>") {
			log.Fatalf("UnicodeData.txt: %s carries an uppercase mapping; a range would have to be expanded", name)
		}
		upper[point(fields[0])] = point(fields[12])
	}
	return upper
}

// specialUpper is SpecialCasing.txt's unconditional uppercase mappings of more
// than one code point. Conditional lines are the language tailoring (lt, tr,
// az) and Final_Sigma, which is a lowercase context; toUpperCase uses none of
// them. Anything else stops the generator instead of being skipped quietly.
func specialUpper(contents []byte, simple map[uint32]uint32) map[uint32][]uint32 {
	full := map[uint32][]uint32{}
	finalSigma := false
	for _, fields := range dataLines(contents) {
		if len(fields) < 4 {
			log.Fatalf("SpecialCasing.txt: %q", fields)
		}
		code := point(fields[0])
		condition := ""
		if len(fields) >= 5 {
			condition = fields[4]
		}
		if condition != "" {
			language := strings.Fields(condition)[0]
			if language == strings.ToLower(language) && len(language) <= 3 {
				continue
			}
			if condition == "Final_Sigma" && code == 0x03A3 {
				finalSigma = true
				continue
			}
			log.Fatalf("SpecialCasing.txt: a condition the legacy mapping doesn't know: %q", fields)
		}
		mapped := codePoints(fields[3])
		if len(mapped) == 0 {
			log.Fatalf("SpecialCasing.txt: empty uppercase %q", fields)
		}
		if len(mapped) == 1 {
			want := code
			if upper, has := simple[code]; has {
				want = upper
			}
			if mapped[0] != want {
				log.Fatalf("SpecialCasing.txt: U+%04X uppercases to U+%04X where UnicodeData.txt says U+%04X", code, mapped[0], want)
			}
			continue
		}
		full[code] = mapped
	}
	if !finalSigma {
		log.Fatal("SpecialCasing.txt: no Final_Sigma mapping for U+03A3")
	}
	return full
}

func codePoints(field string) []uint32 {
	parts := strings.Fields(field)
	mapped := make([]uint32, len(parts))
	for index, part := range parts {
		mapped[index] = point(part)
	}
	return mapped
}

// keepLegacyUpper applies the two filters in Canonicalize: an uppercase that is
// not exactly one UTF-16 code unit is dropped, and so is one that maps a
// non-ASCII code unit into ASCII. Otherwise the single code unit is the result.
func keepLegacyUpper(cu uint32, mapped []uint32) uint32 {
	units := 0
	for _, cp := range mapped {
		if cp > 0xFFFF {
			units += 2
		} else {
			units++
		}
	}
	if units != 1 {
		return cu
	}
	upper := mapped[0]
	if cu >= 128 && upper < 128 {
		return cu
	}
	return upper
}

// legacyASCIIKept is every non-ASCII code unit whose full uppercase is a single
// ASCII code unit. Canonicalize keeps the original. The list is sorted.
func legacyASCIIKept(unicodeData, specialCasing []byte) []uint32 {
	simple := simpleUpper(unicodeData)
	full := specialUpper(specialCasing, simple)
	var kept []uint32
	for cu := uint32(128); cu <= 0xFFFF; cu++ {
		mapped, expanded := full[cu]
		if !expanded {
			upper, has := simple[cu]
			if !has {
				continue
			}
			mapped = []uint32{upper}
		}
		if keepLegacyUpper(cu, mapped) == cu && utf16Units(mapped) == 1 && mapped[0] < 128 {
			kept = append(kept, cu)
		}
	}
	return kept
}

func utf16Units(mapped []uint32) int {
	units := 0
	for _, cp := range mapped {
		if cp > 0xFFFF {
			units += 2
		} else {
			units++
		}
	}
	return units
}
