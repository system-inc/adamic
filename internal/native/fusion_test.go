package native

import (
	"bufio"
	"encoding/hex"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	goruntime "runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

// Node's arm64 release is built by GCC 12, which contracts a * b + c into one fused multiply-add by
// default, so its last bits differ from V8's own build, from x86-64 Node and from Adamic, which never
// fuses (Flags' -ffp-contract=off). Compiler ruled on #4z1nx1s, and Adamic confirmed: Adamic stays
// unfused on every target, and the oracle's truth for arithmetic is V8's unfused result. So a sweep
// holds a case where arm64 Node differs to its unfused value only when the case is on a fusion list:
// closed, audited, and only for an operation ECMAScript lets an engine approximate (Math's
// transcendentals, and parseInt in a radix other than 10 or a power of two). Every other difference is
// red, on every architecture.
//
// A list lives in testdata/fusion/<name>.txt, one case a line, tab-separated: the sweep's question as
// its harness reads it, the unfused value native must answer on every architecture, and arm64 Node's
// fused answer. A parseInt case is audited by V8's own loop, emulated both ways (v8ParseInt), on every
// architecture; an IEEE 754 case by the fused twin (fusedTwin), on arm64. A difference no audit
// reproduces goes in <name>-pending.txt, which allows nothing: it's red on arm64 until an auditor
// explains it (#t3rw158 owes the GCC one).

// updateFusion rewrites the fusion lists from an arm64 run's differences, each sorted by its audit.
var updateFusion = flag.Bool("update-fusion", false, "rewrite testdata/fusion from this arm64 run's differences from Node")

// fusionCase is one case on a fusion list: the unfused value native answers, and arm64 Node's fused one.
type fusionCase struct {
	pinned, fused string
}

// fusionList is one sweep's list, and the questions the sweep has asked of it so far.
type fusionList struct {
	name    string
	cases   map[string]fusionCase
	pending map[string]fusionCase
	asked   map[string]bool
}

func readFusion(t *testing.T, name string) *fusionList {
	t.Helper()
	list := &fusionList{name: name, cases: readFusionCases(t, name+".txt"), pending: readFusionCases(t, name+"-pending.txt"), asked: map[string]bool{}}
	for question := range list.pending {
		if _, listed := list.cases[question]; listed {
			t.Errorf("fusion list %s: %q is both audited and pending", name, question)
		}
	}
	return list
}

func readFusionCases(t *testing.T, file string) map[string]fusionCase {
	t.Helper()
	cases := map[string]fusionCase{}
	contents, err := os.ReadFile(filepath.Join("testdata", "fusion", file))
	if os.IsNotExist(err) {
		return cases
	}
	if err != nil {
		t.Fatal(err)
	}
	scanner := bufio.NewScanner(strings.NewReader(string(contents)))
	scanner.Buffer(nil, 1<<20)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 || fields[1] == fields[2] {
			t.Fatalf("fusion list %s: %q isn't a question, its pinned value and a different fused one", file, line)
		}
		if _, repeated := cases[fields[0]]; repeated {
			t.Fatalf("fusion list %s: %q is listed twice", file, fields[0])
		}
		cases[fields[0]] = fusionCase{fields[1], fields[2]}
	}
	return cases
}

// handled holds one of the sweep's answers to the list. A listed question must answer its pinned value
// natively, everywhere, and Node either that (x86-64) or the fused answer the list records (arm64).
// accounted says the answer was dealt with here, explained or reported, so the sweep doesn't count it
// again, and explained that Node differed and an audited case says why. A pending question is
// reported where Node differs, with what it waits on.
func (list *fusionList) handled(t *testing.T, question, native, oracle string) (accounted, explained bool) {
	t.Helper()
	listed, audited := list.cases[question]
	if !audited {
		var pending bool
		if listed, pending = list.pending[question]; !pending {
			return false, false
		}
	}
	list.asked[question] = true
	if native != listed.pinned {
		t.Errorf("%s: native %s, where the %s fusion list pins %s", question, native, list.name, listed.pinned)
		return true, false
	}
	switch {
	case oracle == listed.pinned:
		return true, false
	case oracle == listed.fused && audited:
		return true, true
	case oracle == listed.fused:
		t.Errorf("%s: native %s, arm64 Node %s; pending on the %s fusion list, which no auditor has explained yet (#t3rw158)", question, native, oracle, list.name)
	default:
		t.Errorf("%s: Node %s is neither the pinned %s nor the fused %s the %s fusion list records", question, oracle, listed.pinned, listed.fused, list.name)
	}
	return true, false
}

// complete says every listed question was asked: a list over cases the sweep no longer asks is stale.
func (list *fusionList) complete(t *testing.T) {
	t.Helper()
	for _, cases := range []map[string]fusionCase{list.cases, list.pending} {
		for question := range cases {
			if !list.asked[question] {
				t.Errorf("fusion list %s names %q, which the sweep no longer asks", list.name, question)
			}
		}
	}
}

// writeFusion writes a list and its pending list, each in the sweep's order, under -update-fusion.
func writeFusion(t *testing.T, name, header string, audited, pending []string) {
	t.Helper()
	for _, file := range []struct {
		name, why string
		lines     []string
	}{
		{name + ".txt", header, audited},
		{name + "-pending.txt", "Differences from arm64 Node no auditor reproduces yet. These allow nothing: each is red on arm64 until it's audited onto " + name + ".txt or fixed (#t3rw158).", pending},
	} {
		contents := "# " + file.why + "\n# Generated by go test ./internal/native -args -update-fusion on arm64 (fusion_test.go).\n" + strings.Join(file.lines, "")
		if err := os.WriteFile(filepath.Join("testdata", "fusion", file.name), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("fusion list %s rewritten: %d audited, %d pending", name, len(audited), len(pending))
}

// fusedTwin is, on arm64, what harness answers to input when built as Node's arm64 release is: with
// contraction on (Options.contract). It's nil anywhere else, where Node can't have fused.
func fusedTwin(t *testing.T, harness, input string) []string {
	t.Helper()
	if goruntime.GOARCH != "arm64" {
		return nil
	}
	binary := filepath.Join(t.TempDir(), "fused")
	if err := Build(harness, binary, Options{Sanitize: true, contract: true}); err != nil {
		t.Fatal(err)
	}
	return strings.Split(strings.TrimSuffix(runWithInput(t, input, binary), "\n"), "\n")
}

// approximatedParseInt says a parse harness question is parseInt in a radix ECMAScript lets an engine
// approximate past 2^53: any but 10 and the powers of two, which V8 computes exactly. It returns the
// text and the radix.
func approximatedParseInt(question string) (string, int, bool) {
	if len(question) < 17 || question[0] != 'i' {
		return "", 0, false
	}
	bits, err := strconv.ParseUint(question[1:17], 16, 64)
	if err != nil {
		return "", 0, false
	}
	text, err := hex.DecodeString(question[17:])
	if err != nil {
		return "", 0, false
	}
	radix := int(toInt32(math.Float64frombits(bits)))
	if radix < 2 || radix > 36 || radix == 10 || radix&(radix-1) == 0 {
		return "", 0, false
	}
	return string(text), radix, true
}

// toInt32 is ECMAScript's ToInt32.
func toInt32(value float64) int32 {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return 0
	}
	return int32(uint32(int64(math.Mod(math.Trunc(value), 4294967296))))
}

// v8ParseInt is V8's parseInt in a radix that isn't 10 or a power of two
// (NumberParseIntHelper::HandleGenericCase): runs of digits whose multiplier stays within 32 bits, each
// folded in as result * multiplier + part, rounded twice where fused is false, as V8's own build and
// x86-64 Node do, and once where it's true, as GCC's arm64 build of Node does. ok is false for a text
// with no digit to parse.
func v8ParseInt(text string, radix int, fused bool) (float64, bool) {
	runes := []rune(text)
	at := 0
	for at < len(runes) && javaScriptSpace(runes[at]) {
		at++
	}
	negative := false
	if at < len(runes) && (runes[at] == '+' || runes[at] == '-') {
		negative = runes[at] == '-'
		at++
	}
	digit := func(at int) int {
		if at >= len(runes) {
			return -1
		}
		value := -1
		switch character := runes[at]; {
		case character >= '0' && character <= '9':
			value = int(character - '0')
		case character >= 'a' && character <= 'z':
			value = int(character-'a') + 10
		case character >= 'A' && character <= 'Z':
			value = int(character-'A') + 10
		}
		if value >= radix {
			return -1
		}
		return value
	}
	if digit(at) < 0 {
		return 0, false
	}
	// StringToIntHelper skips leading zeros before the loop starts, which moves where each part's
	// digits begin, and so how the sum rounds.
	for at < len(runes) && runes[at] == '0' {
		at++
	}
	if digit(at) < 0 {
		if negative {
			return math.Copysign(0, -1), true
		}
		return 0, true
	}
	// V8's kMaximumMultiplier: the largest a part's multiplier may grow to before one more digit of
	// radix 36 could overflow 32 bits.
	const maximumMultiplier = 0xFFFFFFFF / 36
	result := 0.0
	for done := false; !done; {
		part, multiplier := uint32(0), uint32(1)
		for {
			value := digit(at)
			if value < 0 {
				done = true
				break
			}
			next := multiplier * uint32(radix)
			if next > maximumMultiplier {
				break
			}
			part = part*uint32(radix) + uint32(value)
			multiplier = next
			at++
			if at == len(runes) {
				done = true
				break
			}
		}
		if fused {
			result = math.FMA(result, float64(multiplier), float64(part))
		} else {
			// The conversion rounds the product on its own, which Go's spec says no fusion may skip.
			result = float64(result*float64(multiplier)) + float64(part)
		}
	}
	if negative {
		return -result, true
	}
	return result, true
}

// javaScriptSpace is ECMAScript's StrWhiteSpaceChar: white space and line terminators.
func javaScriptSpace(character rune) bool {
	switch character {
	case '\t', '\n', '\v', '\f', '\r', ' ', 0xa0, 0x2028, 0x2029, 0xfeff:
		return true
	}
	return unicode.Is(unicode.Zs, character)
}

// approximatedIEEE754 says an IEEE 754 sweep question is a function ECMAScript lets an engine
// approximate: every one the sweep asks, Math's transcendentals and hypot.
func approximatedIEEE754(question string) bool {
	operation, _, _ := strings.Cut(question, " ")
	return slices.Contains(ieee754Unary, operation) || operation == "atan2" || operation == "hypot"
}

// Every parseInt case on the list is what V8's loop gives: pinned unfused, as V8's own build and x86-64
// Node round, and arm64 Node's answer fused, as GCC's build does. It reads the list alone, so it holds
// on every architecture, Node or no Node. A mutant that flips a pinned value's last bit is red here and
// in the sweep, on both.
func TestParseIntFusionListIsV8sLoop(t *testing.T) {
	t.Parallel()
	list := readFusion(t, "parseInt")
	if len(list.cases) == 0 {
		t.Fatal("the parseInt fusion list is empty")
	}
	bits := func(value float64) string { return fmt.Sprintf("%016x", math.Float64bits(value)) }
	for question, listed := range list.cases {
		text, radix, approximated := approximatedParseInt(question)
		if !approximated {
			t.Errorf("%q isn't parseInt in a radix ECMAScript lets an engine approximate", question)
			continue
		}
		unfused, parsed := v8ParseInt(text, radix, false)
		fused, _ := v8ParseInt(text, radix, true)
		if !parsed || bits(unfused) != listed.pinned || bits(fused) != listed.fused {
			t.Errorf("parseInt(%q, %d): V8's loop gives %s unfused and %s fused, and the list records %s and %s", text, radix, bits(unfused), bits(fused), listed.pinned, listed.fused)
		}
	}
	for question := range list.pending {
		t.Errorf("parseInt has a pending case, %q, which its audit would explain or refuse", question)
	}
}
