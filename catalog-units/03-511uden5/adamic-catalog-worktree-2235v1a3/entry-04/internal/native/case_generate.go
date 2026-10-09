//go:build ignore

// case_generate writes runtime/case_tables.h: the tables toUpperCase and toLowerCase map by, from the
// Unicode Character Database of the Unicode version Node reports (process.versions.unicode).
//
// JavaScript's case mapping is Unicode's full, locale-independent mapping: UnicodeData.txt's simple
// mappings, overridden by SpecialCasing.txt's unconditional ones (ß uppercases to SS), with one
// context: Final_Sigma, for which Σ lowercases to ς. Final_Sigma needs the Cased and Case_Ignorable
// properties, which are in DerivedCoreProperties.txt, so that file is read too.
//
// The files come from ICU's release tag for the ICU inside the Node the oracle runs (Node 24.14.1
// carries ICU 78.2, which is Unicode 17.0), and each must hash to the SHA-256 pinned below, so the
// tables can't change without this file changing. unicode.org serves the same bytes; ICU's tag is
// used because it can't move.
//
//	go run case_generate.go             fetches the files
//	go run case_generate.go -data DIR   reads them from DIR instead
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

// pinned is each file's SHA-256.
var pinned = map[string]string{
	"UnicodeData.txt":           "2e1efc1dcb59c575eedf5ccae60f95229f706ee6d031835247d843c11d96470c",
	"SpecialCasing.txt":         "efc25faf19de21b92c1194c111c932e03d2a5eaf18194e33f1156e96de4c9588",
	"DerivedCoreProperties.txt": "1740a9d1b84b42c78f84b01d4248c0c91272f22e88cbb4ed21869921812455a7",
}

func main() {
	data := flag.String("data", "", "a directory holding the files, instead of fetching them")
	output := flag.String("output", filepath.Join("runtime", "case_tables.h"), "where to write the tables")
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
	if !bytes.HasPrefix(files["SpecialCasing.txt"], []byte("# SpecialCasing-"+unicodeVersion+".txt")) {
		log.Fatalf("SpecialCasing.txt is not version %s", unicodeVersion)
	}

	upper, lower := simpleMappings(files["UnicodeData.txt"])
	fullUpper, fullLower := specialCasing(files["SpecialCasing.txt"], upper, lower)
	cased := property(files["DerivedCoreProperties.txt"], "Cased")
	ignorable := property(files["DerivedCoreProperties.txt"], "Case_Ignorable")

	var out bytes.Buffer
	fmt.Fprintf(&out, "// case_tables.h: Unicode %s's case mappings, written by case_generate.go. Don't edit it; run\n", unicodeVersion)
	fmt.Fprintf(&out, "// go generate in internal/native. Only case.c includes it.\n//\n")
	names := []string{"UnicodeData.txt", "SpecialCasing.txt", "DerivedCoreProperties.txt"}
	for _, name := range names {
		fmt.Fprintf(&out, "// %s SHA-256 %s\n", name, pinned[name])
	}
	fmt.Fprintf(&out, "\n#define CASE_UNICODE_VERSION \"%s\"\n", unicodeVersion)
	writeRanges(&out, "case_upper", runs(upper))
	writeRanges(&out, "case_lower", runs(lower))
	writeFull(&out, "case_full_upper", fullUpper)
	writeFull(&out, "case_full_lower", fullLower)
	writeSpans(&out, "case_cased", cased)
	writeSpans(&out, "case_ignorable", ignorable)
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

func points(field string) []rune {
	var mapped []rune
	for _, part := range strings.Fields(field) {
		mapped = append(mapped, point(part))
	}
	return mapped
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

// simpleMappings is UnicodeData.txt's simple uppercase (field 12) and lowercase (field 13) mappings.
func simpleMappings(contents []byte) (map[rune]rune, map[rune]rune) {
	upper, lower := map[rune]rune{}, map[rune]rune{}
	for _, fields := range dataLines(contents) {
		if len(fields) != 15 {
			log.Fatalf("UnicodeData.txt: %d fields in %q", len(fields), fields)
		}
		code := point(fields[0])
		if fields[12] != "" {
			upper[code] = point(fields[12])
		}
		if fields[13] != "" {
			lower[code] = point(fields[13])
		}
	}
	return upper, lower
}

// fullMapping is one of SpecialCasing.txt's unconditional mappings to more than one code point.
type fullMapping struct {
	code   rune
	mapped []rune
}

// specialCasing is SpecialCasing.txt's unconditional full mappings. A mapping to one code point must
// agree with UnicodeData.txt's simple one, and the only conditional mapping outside a language is
// Final_Sigma's, which case.c carries itself; anything else stops the generator rather than being
// left out quietly.
func specialCasing(contents []byte, upper map[rune]rune, lower map[rune]rune) ([]fullMapping, []fullMapping) {
	var fullUpper, fullLower []fullMapping
	finalSigma := false
	for _, fields := range dataLines(contents) {
		// code; lower; title; upper; (condition_list;)
		code := point(fields[0])
		condition := ""
		if len(fields) >= 6 {
			condition = fields[4]
		}
		if condition != "" {
			if language := strings.Fields(condition)[0]; language == strings.ToLower(language) && len(language) <= 3 {
				// A language's own mapping (lt, tr, az): toUpperCase and toLowerCase are locale-independent.
				continue
			}
			if condition == "Final_Sigma" && code == 0x3a3 && len(points(fields[1])) == 1 && points(fields[1])[0] == 0x3c2 {
				finalSigma = true
				continue
			}
			log.Fatalf("SpecialCasing.txt: a condition case.c doesn't know: %q", fields)
		}
		for _, side := range []struct {
			field  string
			simple map[rune]rune
			full   *[]fullMapping
		}{{fields[3], upper, &fullUpper}, {fields[1], lower, &fullLower}} {
			mapped := points(side.field)
			if len(mapped) > 3 {
				log.Fatalf("SpecialCasing.txt: a mapping longer than case.c holds: %q", fields)
			}
			if len(mapped) > 1 {
				*side.full = append(*side.full, fullMapping{code, mapped})
				continue
			}
			simple, has := side.simple[code]
			if !has {
				simple = code
			}
			if mapped[0] != simple {
				log.Fatalf("SpecialCasing.txt: U+%04X maps to U+%04X where UnicodeData.txt says U+%04X", code, mapped[0], simple)
			}
		}
	}
	if !finalSigma {
		log.Fatal("SpecialCasing.txt: no Final_Sigma mapping for U+03A3")
	}
	for _, full := range [][]fullMapping{fullUpper, fullLower} {
		sort.Slice(full, func(i, j int) bool { return full[i].code < full[j].code })
	}
	return fullUpper, fullLower
}

// span is a closed range of code points.
type span struct{ first, last rune }

// property is the code points DerivedCoreProperties.txt gives a property, as sorted spans, adjacent
// ones merged.
func property(contents []byte, name string) []span {
	var spans []span
	for _, fields := range dataLines(contents) {
		if len(fields) < 2 || fields[1] != name {
			continue
		}
		first, last, isRange := strings.Cut(fields[0], "..")
		next := span{point(first), point(first)}
		if isRange {
			next.last = point(last)
		}
		spans = append(spans, next)
	}
	if len(spans) == 0 {
		log.Fatalf("DerivedCoreProperties.txt: no %s", name)
	}
	sort.Slice(spans, func(i, j int) bool { return spans[i].first < spans[j].first })
	merged := []span{spans[0]}
	for _, next := range spans[1:] {
		top := &merged[len(merged)-1]
		if next.first <= top.last {
			log.Fatalf("DerivedCoreProperties.txt: %s overlaps at U+%04X", name, next.first)
		}
		if next.first == top.last+1 {
			top.last = next.last
			continue
		}
		merged = append(merged, next)
	}
	return merged
}

// run is code points start, start+stride, ... (count of them), each mapping to itself plus delta. A
// run's members are consecutive in the sorted mapping, so a code point inside a run's span but off its
// stride has no mapping at all.
type run struct {
	start  rune
	count  int
	stride rune
	delta  rune
}

func runs(mapping map[rune]rune) []run {
	codes := make([]rune, 0, len(mapping))
	for code := range mapping {
		codes = append(codes, code)
	}
	sort.Slice(codes, func(i, j int) bool { return codes[i] < codes[j] })
	var all []run
	for _, code := range codes {
		delta := mapping[code] - code
		if len(all) > 0 {
			top := &all[len(all)-1]
			last := top.start + rune(top.count-1)*top.stride
			step := code - last
			if top.delta == delta && top.count < 0xffff && (top.count == 1 && (step == 1 || step == 2) || top.count > 1 && step == top.stride) {
				top.stride = step
				top.count++
				continue
			}
		}
		all = append(all, run{code, 1, 1, delta})
	}
	// Every mapping must come back out of the runs exactly.
	for _, each := range all {
		for index := 0; index < each.count; index++ {
			code := each.start + rune(index)*each.stride
			if mapping[code] != code+each.delta {
				log.Fatalf("runs lost U+%04X", code)
			}
		}
	}
	return all
}

func writeRanges(out *bytes.Buffer, name string, all []run) {
	fmt.Fprintf(out, "\n// %s: %d runs. A run is start, count, stride, delta.\nstatic const case_run %s[] = {\n", name, len(all), name)
	for _, each := range all {
		fmt.Fprintf(out, "\t{0x%04x, %d, %d, %d},\n", each.start, each.count, each.stride, each.delta)
	}
	fmt.Fprintf(out, "};\n")
}

func writeFull(out *bytes.Buffer, name string, all []fullMapping) {
	fmt.Fprintf(out, "\n// %s: %d mappings to more than one code point.\nstatic const case_full %s[] = {\n", name, len(all), name)
	for _, each := range all {
		mapped := make([]string, 3)
		for index := range mapped {
			mapped[index] = "0"
			if index < len(each.mapped) {
				mapped[index] = fmt.Sprintf("0x%04x", each.mapped[index])
			}
		}
		fmt.Fprintf(out, "\t{0x%04x, %d, {%s}},\n", each.code, len(each.mapped), strings.Join(mapped, ", "))
	}
	fmt.Fprintf(out, "};\n")
}

func writeSpans(out *bytes.Buffer, name string, all []span) {
	fmt.Fprintf(out, "\n// %s: %d spans, first and last.\nstatic const case_span %s[] = {\n", name, len(all), name)
	for _, each := range all {
		fmt.Fprintf(out, "\t{0x%04x, 0x%04x},\n", each.first, each.last)
	}
	fmt.Fprintf(out, "};\n")
}
