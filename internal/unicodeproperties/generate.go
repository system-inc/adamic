//go:build ignore

// generate writes tables.go, the v flag's properties of strings, from the
// Unicode Character Database of the version cohere's unicodeproperties is
// built from, which is the version Node reports (process.versions.unicode).
// Node 24 says 17.0, and the files for that version are published as 17.0.0.
// The code-point properties are cohere's alone: this package reads them
// through cohere's Lookup and generates no copy. The same run writes
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

	cohere "github.com/system-inc/cohere/unicodeproperties"
)

// unicodeVersion is cohere's, so these tables and the code-point ones Lookup
// reads from cohere are one Unicode. A cohere on a new version makes this run
// fetch that version's files, which the pins below then refuse until they are
// updated.
const unicodeVersion = cohere.Version

// pinned is each UCD file's SHA-256. UnicodeData.txt's hash is the same pin
// internal/native's case tables use, so the two generators cannot drift apart
// onto different copies of Unicode 17.
var pinned = map[string]string{
	"UnicodeData.txt":         "2e1efc1dcb59c575eedf5ccae60f95229f706ee6d031835247d843c11d96470c",
	"CaseFolding.txt":         "ff8d8fefbf123574205085d6714c36149eb946d717a0c585c27f0f4ef58c4183",
	"SpecialCasing.txt":       "efc25faf19de21b92c1194c111c932e03d2a5eaf18194e33f1156e96de4c9588",
	"emoji-sequences.txt":     "12cc8267dc33cbd11ed32bcf6fc5dc2ad9c7a77bae1bdfba2f41b1b9b3ead8dd",
	"emoji-zwj-sequences.txt": "5b25441daed2322b068c5e70cda522946a4f0274df864445a1965a92e5fc5cad",
}

var sourceURL = map[string]string{
	"UnicodeData.txt":         "https://www.unicode.org/Public/" + unicodeVersion + "/ucd/UnicodeData.txt",
	"CaseFolding.txt":         "https://www.unicode.org/Public/" + unicodeVersion + "/ucd/CaseFolding.txt",
	"SpecialCasing.txt":       "https://www.unicode.org/Public/" + unicodeVersion + "/ucd/SpecialCasing.txt",
	"emoji-sequences.txt":     "https://www.unicode.org/Public/" + unicodeVersion + "/emoji/emoji-sequences.txt",
	"emoji-zwj-sequences.txt": "https://www.unicode.org/Public/" + unicodeVersion + "/emoji/emoji-zwj-sequences.txt",
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
	for _, name := range []string{"emoji-sequences.txt", "emoji-zwj-sequences.txt"} {
		if !bytes.Contains(files[name][:400], []byte("Version: 17.0")) {
			log.Fatalf("%s does not say Version: 17.0", name)
		}
	}

	sequences := emojiSequences(files["emoji-sequences.txt"], files["emoji-zwj-sequences.txt"])

	var out bytes.Buffer
	writeTables(&out, sequences)
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
	// Emoji names in a comment can be long, and the default 64KiB token is
	// enough, but a raised buffer costs nothing.
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

// writeTables writes tables.go: each property of strings' sequences, sorted,
// and the map Lookup reads them from.
func writeTables(out *bytes.Buffer, sequences map[string][]string) {
	fmt.Fprintf(out, "// Code generated by generate.go from Unicode %s. Do not edit.\n", unicodeVersion)
	fmt.Fprintf(out, "// Run go generate in internal/unicodeproperties.\n")
	fmt.Fprintf(out, "//\n")
	fmt.Fprintf(out, "// The v flag's properties of strings. The code-point properties are cohere's\n")
	fmt.Fprintf(out, "// unicodeproperties, which Lookup reads.\n")
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

	for _, name := range stringPropertyNames {
		fmt.Fprintf(out, "var sequences_%s = []string{\n", name)
		for _, sequence := range sequences[name] {
			fmt.Fprintf(out, "\t%s,\n", strconv.Quote(sequence))
		}
		fmt.Fprintf(out, "}\n\n")
	}
	fmt.Fprintf(out, "var stringProperties = map[string][]string{\n")
	for _, name := range stringPropertyNames {
		fmt.Fprintf(out, "\t%q: sequences_%s,\n", name, name)
	}
	fmt.Fprintf(out, "}\n")

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
