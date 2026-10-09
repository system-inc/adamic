//go:build ignore

// normalize_generate writes runtime/normalize_tables.h: what String.prototype.normalize needs, from
// the Unicode Character Database of the Unicode version Node reports (process.versions.unicode).
//
// UAX #15 normalization takes three things from it: each code point's canonical combining class and
// decomposition mapping (UnicodeData.txt, fields 3 and 5; a mapping with a <tag> is a compatibility
// one), and which canonical pairs never recompose (Full_Composition_Exclusion, in
// DerivedNormalizationProps.txt, which already includes CompositionExclusions.txt). Hangul syllables
// decompose and compose by arithmetic, so they're in normalize.c, not here.
//
// As with case_generate.go, the files come from ICU's release tag for the ICU inside the Node the
// oracle runs, each hashed against a pin.
//
//	go run normalize_generate.go             fetches the files
//	go run normalize_generate.go -data DIR   reads them from DIR instead
package main

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const unicodeVersion = "17.0.0"

const source = "https://raw.githubusercontent.com/unicode-org/icu/release-78.2/icu4c/source/data/unidata/"

var pinned = map[string]string{
	"UnicodeData.txt":               "2e1efc1dcb59c575eedf5ccae60f95229f706ee6d031835247d843c11d96470c",
	"DerivedNormalizationProps.txt": "18f6857f9542dd942bc5b62e1547d156912e3fe9e9342d6bd2b544d5ece13c42",
}

func main() {
	data := flag.String("data", "", "a directory holding the files, instead of fetching them")
	output := flag.String("output", filepath.Join("runtime", "normalize_tables.h"), "where to write the tables")
	flag.Parse()

	files := map[string][]byte{}
	for name, sum := range pinned {
		contents, err := read(*data, name)
		if err != nil {
			log.Fatal(err)
		}
		hash := sha256.Sum256(contents)
		if got := hex.EncodeToString(hash[:]); got != sum {
			log.Fatalf("%s: SHA-256 %s, want %s", name, got, sum)
		}
		files[name] = contents
	}
	if !bytes.HasPrefix(files["DerivedNormalizationProps.txt"], []byte("# DerivedNormalizationProps-"+unicodeVersion+".txt")) {
		log.Fatalf("DerivedNormalizationProps.txt is not version %s", unicodeVersion)
	}

	classes, decompositions := unicodeData(files["UnicodeData.txt"])
	excluded := exclusions(files["DerivedNormalizationProps.txt"])

	// The pairs that compose: every canonical decomposition to exactly two code points whose result
	// isn't excluded. Singletons and decompositions starting with a non-starter are in the exclusions.
	var pairs [][3]rune
	for code, mapping := range decompositions {
		if mapping.compatibility || len(mapping.points) != 2 || excluded[code] {
			continue
		}
		if classes[mapping.points[0]] != 0 {
			log.Fatalf("U+%04X composes from a non-starter but isn't excluded", code)
		}
		pairs = append(pairs, [3]rune{mapping.points[0], mapping.points[1], code})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i][0] != pairs[j][0] {
			return pairs[i][0] < pairs[j][0]
		}
		return pairs[i][1] < pairs[j][1]
	})

	var out bytes.Buffer
	fmt.Fprintf(&out, "// normalize_tables.h: Unicode %s's normalization data, written by normalize_generate.go. Don't\n", unicodeVersion)
	fmt.Fprintf(&out, "// edit it; run go generate in internal/native. Only normalize.c includes it.\n//\n")
	for _, name := range []string{"UnicodeData.txt", "DerivedNormalizationProps.txt"} {
		fmt.Fprintf(&out, "// %s SHA-256 %s\n", name, pinned[name])
	}
	fmt.Fprintf(&out, "\n#define NORMALIZE_UNICODE_VERSION \"%s\"\n", unicodeVersion)
	writeClasses(&out, classes)
	writeDecompositions(&out, decompositions)
	fmt.Fprintf(&out, "\n// normalize_pairs: %d canonical pairs that compose, sorted, as first, second, composite.\nstatic const normalize_pair normalize_pairs[] = {\n", len(pairs))
	for _, pair := range pairs {
		fmt.Fprintf(&out, "\t{0x%04x, 0x%04x, 0x%04x},\n", pair[0], pair[1], pair[2])
	}
	fmt.Fprintf(&out, "};\n")
	if err := os.WriteFile(*output, out.Bytes(), 0o644); err != nil {
		log.Fatal(err)
	}
}

func read(directory string, name string) ([]byte, error) {
	if directory != "" {
		return os.ReadFile(filepath.Join(directory, name))
	}
	response, err := http.Get(source + name)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", name, response.Status)
	}
	return io.ReadAll(response.Body)
}

func point(field string) rune {
	value, err := strconv.ParseUint(strings.TrimSpace(field), 16, 32)
	if err != nil || value > 0x10ffff {
		log.Fatalf("not a code point: %q", field)
	}
	return rune(value)
}

// dataLines is each line's fields, split at semicolons, with comments and blank lines dropped.
func dataLines(contents []byte) [][]string {
	var lines [][]string
	scanner := bufio.NewScanner(bytes.NewReader(contents))
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
	return lines
}

// decomposition is one code point's mapping, one level deep, as UnicodeData.txt gives it.
type decomposition struct {
	compatibility bool
	points        []rune
}

// unicodeData is every code point's canonical combining class that isn't 0, and every decomposition
// mapping. A range (First and Last) has class 0 and no mapping in every version so far, and the
// generator stops if one ever doesn't, rather than leave its other members out.
func unicodeData(contents []byte) (map[rune]int, map[rune]decomposition) {
	classes, decompositions := map[rune]int{}, map[rune]decomposition{}
	for _, fields := range dataLines(contents) {
		if len(fields) != 15 {
			log.Fatalf("UnicodeData.txt: %d fields in %q", len(fields), fields)
		}
		code := point(fields[0])
		class, err := strconv.Atoi(fields[3])
		if err != nil || class < 0 || class > 254 {
			log.Fatalf("UnicodeData.txt: combining class %q", fields[3])
		}
		inRange := strings.HasSuffix(fields[1], ", First>") || strings.HasSuffix(fields[1], ", Last>")
		if inRange && (class != 0 || fields[5] != "") {
			log.Fatalf("UnicodeData.txt: a range with a combining class or a mapping: %q", fields)
		}
		if class != 0 {
			classes[code] = class
		}
		if fields[5] == "" {
			continue
		}
		words := strings.Fields(fields[5])
		mapping := decomposition{}
		if strings.HasPrefix(words[0], "<") {
			mapping.compatibility = true
			words = words[1:]
		}
		for _, word := range words {
			mapping.points = append(mapping.points, point(word))
		}
		if len(mapping.points) == 0 || len(mapping.points) > 18 {
			log.Fatalf("UnicodeData.txt: a mapping of %d code points: %q", len(mapping.points), fields)
		}
		decompositions[code] = mapping
	}
	return classes, decompositions
}

func exclusions(contents []byte) map[rune]bool {
	excluded := map[rune]bool{}
	for _, fields := range dataLines(contents) {
		if len(fields) < 2 || fields[1] != "Full_Composition_Exclusion" {
			continue
		}
		first, last, isRange := strings.Cut(fields[0], "..")
		end := point(first)
		if isRange {
			end = point(last)
		}
		for code := point(first); code <= end; code++ {
			excluded[code] = true
		}
	}
	if len(excluded) == 0 {
		log.Fatal("DerivedNormalizationProps.txt: no Full_Composition_Exclusion")
	}
	return excluded
}

// writeClasses writes the combining classes as runs of consecutive code points with one class.
func writeClasses(out *bytes.Buffer, classes map[rune]int) {
	codes := make([]rune, 0, len(classes))
	for code := range classes {
		codes = append(codes, code)
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })
	type run struct {
		first, last rune
		class       int
	}
	var runs []run
	for _, code := range codes {
		if count := len(runs); count > 0 && runs[count-1].last+1 == code && runs[count-1].class == classes[code] {
			runs[count-1].last = code
			continue
		}
		runs = append(runs, run{code, code, classes[code]})
	}
	fmt.Fprintf(out, "\n// normalize_classes: %d runs of code points with one canonical combining class other than 0, as\n// first, last, class.\nstatic const normalize_class normalize_classes[] = {\n", len(runs))
	for _, each := range runs {
		fmt.Fprintf(out, "\t{0x%04x, 0x%04x, %d},\n", each.first, each.last, each.class)
	}
	fmt.Fprintf(out, "};\n")
}

// writeDecompositions writes each mapping as a code point, whether it's a compatibility mapping, and
// where its code points start in one shared list and how many there are.
func writeDecompositions(out *bytes.Buffer, decompositions map[rune]decomposition) {
	codes := make([]rune, 0, len(decompositions))
	for code := range decompositions {
		codes = append(codes, code)
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })
	var list []rune
	fmt.Fprintf(out, "\n// normalize_mappings: %d decomposition mappings, one level deep: code point, compatibility,\n// start in normalize_mapped, count.\nstatic const normalize_mapping normalize_mappings[] = {\n", len(codes))
	for _, code := range codes {
		mapping := decompositions[code]
		fmt.Fprintf(out, "\t{0x%04x, %t, %d, %d},\n", code, mapping.compatibility, len(list), len(mapping.points))
		list = append(list, mapping.points...)
	}
	fmt.Fprintf(out, "};\n\n// normalize_mapped: the code points the mappings map to.\nstatic const uint32_t normalize_mapped[] = {\n")
	for start := 0; start < len(list); start += 12 {
		line := []string{}
		for _, code := range list[start:min(start+12, len(list))] {
			line = append(line, fmt.Sprintf("0x%04x", code))
		}
		fmt.Fprintf(out, "\t%s,\n", strings.Join(line, ", "))
	}
	fmt.Fprintf(out, "};\n")
}
