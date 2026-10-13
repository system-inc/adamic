package unicodeproperties

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/testguard"
)

const (
	// Eighty lines keep one Node process at runUnicodeBatch's four-minute deadline. A dense shard is almost all
	// Node's start, so fewer, larger shards do the same scans in a fifth of the processes.
	unicodeScanShardLines = 80
	// A stride line is about half a second of Node's scan on one core, so two-line shards keep a stride range, and
	// the planted failure that runs one range and its neighbor, well under the twenty-second budget.
	unicodeStrideScanShardLines = 2
	// The ordered census has 4050 class and dense-group lines, then 1024 strides.
	unicodeStrideScanStart = 4050
)

// The first four blocks contain the cheap class and dense singleton scans,
// thirteen eighty-line shards each and twelve in the last. Each remaining block
// owns four two-line singleton-stride shards, one wave on four CPUs. The last
// dense shard stops at the first stride line.
// These boundaries partition all 563 shards without changing the scan inputs.
type unicodeNodeTopRange struct {
	name       string
	start, end int
}

var unicodeNodeTopRanges = []unicodeNodeTopRange{
	{"TestCanonicalizeUnicodeNodeRange0", 0, 13},
	{"TestCanonicalizeUnicodeNodeRange1", 13, 26},
	{"TestCanonicalizeUnicodeNodeRange2", 26, 39},
	{"TestCanonicalizeUnicodeNodeRange3", 39, 51},
	{"TestCanonicalizeUnicodeNodeRange4", 51, 55},
	{"TestCanonicalizeUnicodeNodeRange5", 55, 59},
	{"TestCanonicalizeUnicodeNodeRange6", 59, 63},
	{"TestCanonicalizeUnicodeNodeRange7", 63, 67},
	{"TestCanonicalizeUnicodeNodeRange8", 67, 71},
	{"TestCanonicalizeUnicodeNodeRange9", 71, 75},
	{"TestCanonicalizeUnicodeNodeRange10", 75, 79},
	{"TestCanonicalizeUnicodeNodeRange11", 79, 83},
	{"TestCanonicalizeUnicodeNodeRange12", 83, 87},
	{"TestCanonicalizeUnicodeNodeRange13", 87, 91},
	{"TestCanonicalizeUnicodeNodeRange14", 91, 95},
	{"TestCanonicalizeUnicodeNodeRange15", 95, 99},
	{"TestCanonicalizeUnicodeNodeRange16", 99, 103},
	{"TestCanonicalizeUnicodeNodeRange17", 103, 107},
	{"TestCanonicalizeUnicodeNodeRange18", 107, 111},
	{"TestCanonicalizeUnicodeNodeRange19", 111, 115},
	{"TestCanonicalizeUnicodeNodeRange20", 115, 119},
	{"TestCanonicalizeUnicodeNodeRange21", 119, 123},
	{"TestCanonicalizeUnicodeNodeRange22", 123, 127},
	{"TestCanonicalizeUnicodeNodeRange23", 127, 131},
	{"TestCanonicalizeUnicodeNodeRange24", 131, 135},
	{"TestCanonicalizeUnicodeNodeRange25", 135, 139},
	{"TestCanonicalizeUnicodeNodeRange26", 139, 143},
	{"TestCanonicalizeUnicodeNodeRange27", 143, 147},
	{"TestCanonicalizeUnicodeNodeRange28", 147, 151},
	{"TestCanonicalizeUnicodeNodeRange29", 151, 155},
	{"TestCanonicalizeUnicodeNodeRange30", 155, 159},
	{"TestCanonicalizeUnicodeNodeRange31", 159, 163},
	{"TestCanonicalizeUnicodeNodeRange32", 163, 167},
	{"TestCanonicalizeUnicodeNodeRange33", 167, 171},
	{"TestCanonicalizeUnicodeNodeRange34", 171, 175},
	{"TestCanonicalizeUnicodeNodeRange35", 175, 179},
	{"TestCanonicalizeUnicodeNodeRange36", 179, 183},
	{"TestCanonicalizeUnicodeNodeRange37", 183, 187},
	{"TestCanonicalizeUnicodeNodeRange38", 187, 191},
	{"TestCanonicalizeUnicodeNodeRange39", 191, 195},
	{"TestCanonicalizeUnicodeNodeRange40", 195, 199},
	{"TestCanonicalizeUnicodeNodeRange41", 199, 203},
	{"TestCanonicalizeUnicodeNodeRange42", 203, 207},
	{"TestCanonicalizeUnicodeNodeRange43", 207, 211},
	{"TestCanonicalizeUnicodeNodeRange44", 211, 215},
	{"TestCanonicalizeUnicodeNodeRange45", 215, 219},
	{"TestCanonicalizeUnicodeNodeRange46", 219, 223},
	{"TestCanonicalizeUnicodeNodeRange47", 223, 227},
	{"TestCanonicalizeUnicodeNodeRange48", 227, 231},
	{"TestCanonicalizeUnicodeNodeRange49", 231, 235},
	{"TestCanonicalizeUnicodeNodeRange50", 235, 239},
	{"TestCanonicalizeUnicodeNodeRange51", 239, 243},
	{"TestCanonicalizeUnicodeNodeRange52", 243, 247},
	{"TestCanonicalizeUnicodeNodeRange53", 247, 251},
	{"TestCanonicalizeUnicodeNodeRange54", 251, 255},
	{"TestCanonicalizeUnicodeNodeRange55", 255, 259},
	{"TestCanonicalizeUnicodeNodeRange56", 259, 263},
	{"TestCanonicalizeUnicodeNodeRange57", 263, 267},
	{"TestCanonicalizeUnicodeNodeRange58", 267, 271},
	{"TestCanonicalizeUnicodeNodeRange59", 271, 275},
	{"TestCanonicalizeUnicodeNodeRange60", 275, 279},
	{"TestCanonicalizeUnicodeNodeRange61", 279, 283},
	{"TestCanonicalizeUnicodeNodeRange62", 283, 287},
	{"TestCanonicalizeUnicodeNodeRange63", 287, 291},
	{"TestCanonicalizeUnicodeNodeRange64", 291, 295},
	{"TestCanonicalizeUnicodeNodeRange65", 295, 299},
	{"TestCanonicalizeUnicodeNodeRange66", 299, 303},
	{"TestCanonicalizeUnicodeNodeRange67", 303, 307},
	{"TestCanonicalizeUnicodeNodeRange68", 307, 311},
	{"TestCanonicalizeUnicodeNodeRange69", 311, 315},
	{"TestCanonicalizeUnicodeNodeRange70", 315, 319},
	{"TestCanonicalizeUnicodeNodeRange71", 319, 323},
	{"TestCanonicalizeUnicodeNodeRange72", 323, 327},
	{"TestCanonicalizeUnicodeNodeRange73", 327, 331},
	{"TestCanonicalizeUnicodeNodeRange74", 331, 335},
	{"TestCanonicalizeUnicodeNodeRange75", 335, 339},
	{"TestCanonicalizeUnicodeNodeRange76", 339, 343},
	{"TestCanonicalizeUnicodeNodeRange77", 343, 347},
	{"TestCanonicalizeUnicodeNodeRange78", 347, 351},
	{"TestCanonicalizeUnicodeNodeRange79", 351, 355},
	{"TestCanonicalizeUnicodeNodeRange80", 355, 359},
	{"TestCanonicalizeUnicodeNodeRange81", 359, 363},
	{"TestCanonicalizeUnicodeNodeRange82", 363, 367},
	{"TestCanonicalizeUnicodeNodeRange83", 367, 371},
	{"TestCanonicalizeUnicodeNodeRange84", 371, 375},
	{"TestCanonicalizeUnicodeNodeRange85", 375, 379},
	{"TestCanonicalizeUnicodeNodeRange86", 379, 383},
	{"TestCanonicalizeUnicodeNodeRange87", 383, 387},
	{"TestCanonicalizeUnicodeNodeRange88", 387, 391},
	{"TestCanonicalizeUnicodeNodeRange89", 391, 395},
	{"TestCanonicalizeUnicodeNodeRange90", 395, 399},
	{"TestCanonicalizeUnicodeNodeRange91", 399, 403},
	{"TestCanonicalizeUnicodeNodeRange92", 403, 407},
	{"TestCanonicalizeUnicodeNodeRange93", 407, 411},
	{"TestCanonicalizeUnicodeNodeRange94", 411, 415},
	{"TestCanonicalizeUnicodeNodeRange95", 415, 419},
	{"TestCanonicalizeUnicodeNodeRange96", 419, 423},
	{"TestCanonicalizeUnicodeNodeRange97", 423, 427},
	{"TestCanonicalizeUnicodeNodeRange98", 427, 431},
	{"TestCanonicalizeUnicodeNodeRange99", 431, 435},
	{"TestCanonicalizeUnicodeNodeRange100", 435, 439},
	{"TestCanonicalizeUnicodeNodeRange101", 439, 443},
	{"TestCanonicalizeUnicodeNodeRange102", 443, 447},
	{"TestCanonicalizeUnicodeNodeRange103", 447, 451},
	{"TestCanonicalizeUnicodeNodeRange104", 451, 455},
	{"TestCanonicalizeUnicodeNodeRange105", 455, 459},
	{"TestCanonicalizeUnicodeNodeRange106", 459, 463},
	{"TestCanonicalizeUnicodeNodeRange107", 463, 467},
	{"TestCanonicalizeUnicodeNodeRange108", 467, 471},
	{"TestCanonicalizeUnicodeNodeRange109", 471, 475},
	{"TestCanonicalizeUnicodeNodeRange110", 475, 479},
	{"TestCanonicalizeUnicodeNodeRange111", 479, 483},
	{"TestCanonicalizeUnicodeNodeRange112", 483, 487},
	{"TestCanonicalizeUnicodeNodeRange113", 487, 491},
	{"TestCanonicalizeUnicodeNodeRange114", 491, 495},
	{"TestCanonicalizeUnicodeNodeRange115", 495, 499},
	{"TestCanonicalizeUnicodeNodeRange116", 499, 503},
	{"TestCanonicalizeUnicodeNodeRange117", 503, 507},
	{"TestCanonicalizeUnicodeNodeRange118", 507, 511},
	{"TestCanonicalizeUnicodeNodeRange119", 511, 515},
	{"TestCanonicalizeUnicodeNodeRange120", 515, 519},
	{"TestCanonicalizeUnicodeNodeRange121", 519, 523},
	{"TestCanonicalizeUnicodeNodeRange122", 523, 527},
	{"TestCanonicalizeUnicodeNodeRange123", 527, 531},
	{"TestCanonicalizeUnicodeNodeRange124", 531, 535},
	{"TestCanonicalizeUnicodeNodeRange125", 535, 539},
	{"TestCanonicalizeUnicodeNodeRange126", 539, 543},
	{"TestCanonicalizeUnicodeNodeRange127", 543, 547},
	{"TestCanonicalizeUnicodeNodeRange128", 547, 551},
	{"TestCanonicalizeUnicodeNodeRange129", 551, 555},
	{"TestCanonicalizeUnicodeNodeRange130", 555, 559},
	{"TestCanonicalizeUnicodeNodeRange131", 559, 563},
}

func checkUnicodeTopRangeCoverage(lines []string, shards []unicodeScanShard, groups []unicodeNodeTopRange) error {
	if err := checkUnicodeShardCoverage(lines, shards); err != nil {
		return err
	}
	names := map[string]bool{}
	covered := make([]int, len(shards))
	for _, group := range groups {
		if names[group.name] {
			return fmt.Errorf("duplicate top-level range %s", group.name)
		}
		names[group.name] = true
		if group.start < 0 || group.end <= group.start || group.end > len(shards) {
			return fmt.Errorf("invalid top-level range %s", group.name)
		}
		for i := group.start; i < group.end; i++ {
			covered[i]++
		}
	}
	for i, count := range covered {
		if count != 1 {
			return fmt.Errorf("shard %d owned %d times, want 1", i, count)
		}
	}
	return nil
}

type unicodeScanShard struct {
	start, end int
	input      string
}

func (s unicodeScanShard) name() string {
	hash := sha256.Sum256([]byte(s.input))
	return fmt.Sprintf("%04d-%04d-%x", s.start, s.end, hash[:6])
}

func unicodeCanonicalizeShards(lines []string) []unicodeScanShard {
	var shards []unicodeScanShard
	for start := 0; start < len(lines); {
		end := min(start+unicodeScanShardLines, unicodeStrideScanStart, len(lines))
		if start >= unicodeStrideScanStart {
			end = min(start+unicodeStrideScanShardLines, len(lines))
		}
		shards = append(shards, unicodeScanShard{start, end, strings.Join(lines[start:end], "\n") + "\n"})
		start = end
	}
	return shards
}

func checkUnicodeShardCoverage(lines []string, shards []unicodeScanShard) error {
	covered := make([]int, len(lines))
	names := map[string]bool{}
	for _, shard := range shards {
		if shard.start < 0 || shard.end <= shard.start || shard.end > len(lines) || shard.end-shard.start > unicodeScanShardLines {
			return fmt.Errorf("invalid shard range %d:%d", shard.start, shard.end)
		}
		if shard.start < unicodeStrideScanStart && shard.end > unicodeStrideScanStart || shard.start >= unicodeStrideScanStart && shard.end-shard.start > unicodeStrideScanShardLines {
			return fmt.Errorf("invalid stride shard range %d:%d", shard.start, shard.end)
		}
		if shard.input != strings.Join(lines[shard.start:shard.end], "\n")+"\n" {
			return fmt.Errorf("shard %s changed its input", shard.name())
		}
		if names[shard.name()] {
			return fmt.Errorf("duplicate shard %s", shard.name())
		}
		names[shard.name()] = true
		for i := shard.start; i < shard.end; i++ {
			covered[i]++
		}
	}
	for i, count := range covered {
		if count != 1 {
			return fmt.Errorf("scan %d covered %d times, want 1", i, count)
		}
	}
	return nil
}

func TestUnicodeNodeShardCoverage(t *testing.T) {
	t.Parallel()
	lines := unicodeCanonicalizeLines()
	if len(lines) != 5074 {
		t.Fatalf("scan census changed: %d", len(lines))
	}
	// This digest pins the ordered pre-split inputs, including both class flags
	// and every block and stride singleton group. It is independent of batching.
	hash := sha256.Sum256([]byte(strings.Join(lines, "\n") + "\n"))
	const originalInputs = "d808be2501a06c80d27f1962fd47218bf070a51a2cfb904beac5692c36b956ac"
	if fmt.Sprintf("%x", hash) != originalInputs {
		t.Fatalf("pre-split scan inputs changed: %x", hash)
	}
	t.Logf("original inputs SHA256 %x", hash)
	shards := unicodeCanonicalizeShards(lines)
	if err := checkUnicodeShardCoverage(lines, shards); err != nil {
		t.Fatal(err)
	}
	if err := checkUnicodeShardCoverage(lines, shards[1:]); err == nil || !strings.Contains(err.Error(), "covered 0 times") {
		t.Fatalf("dropped shard survived: %v", err)
	}
	duplicated := append(append([]unicodeScanShard{}, shards...), shards[0])
	if err := checkUnicodeShardCoverage(lines, duplicated); err == nil || !strings.Contains(err.Error(), "duplicate shard") {
		t.Fatalf("duplicate shard survived: %v", err)
	}
	overlap := append([]unicodeScanShard{}, shards...)
	overlap[1] = unicodeScanShard{79, 159, strings.Join(lines[79:159], "\n") + "\n"}
	if err := checkUnicodeShardCoverage(lines, overlap); err == nil || !strings.Contains(err.Error(), "covered 2 times") {
		t.Fatalf("overlapping shard survived: %v", err)
	}
	crossingStride := append([]unicodeScanShard{}, shards...)
	crossingStride[50] = unicodeScanShard{4048, 4051, strings.Join(lines[4048:4051], "\n") + "\n"}
	if err := checkUnicodeShardCoverage(lines, crossingStride); err == nil || !strings.Contains(err.Error(), "invalid stride shard range") {
		t.Fatalf("shard crossing into strides survived: %v", err)
	}
	wideStride := append([]unicodeScanShard{}, shards...)
	wideStride[51] = unicodeScanShard{4050, 4053, strings.Join(lines[4050:4053], "\n") + "\n"}
	if err := checkUnicodeShardCoverage(lines, wideStride); err == nil || !strings.Contains(err.Error(), "invalid stride shard range") {
		t.Fatalf("three-line stride shard survived: %v", err)
	}
	t.Log("crossing stride boundary and oversized stride shard mutants caught")
	t.Logf("%d shards cover all %d scans exactly once, %d code points", len(shards), len(lines), int64(len(lines))*0x110000)
	if err := checkUnicodeTopRangeCoverage(lines, shards, unicodeNodeTopRanges); err != nil {
		t.Fatal(err)
	}
	if len(shards) != 563 || shards[51].start != unicodeStrideScanStart {
		t.Fatal("dense and stride shard census changed")
	}
	if len(unicodeNodeTopRanges) != 132 {
		t.Fatal("top-level range census changed")
	}
	if err := checkUnicodeTopRangeCoverage(lines, shards, unicodeNodeTopRanges[1:]); err == nil || !strings.Contains(err.Error(), "owned 0 times") {
		t.Fatalf("dropped top-level range survived: %v", err)
	}
	droppedStride := append([]unicodeNodeTopRange{}, unicodeNodeTopRanges...)
	droppedStride[4].start++
	if err := checkUnicodeTopRangeCoverage(lines, shards, droppedStride); err == nil || !strings.Contains(err.Error(), "shard 51 owned 0 times") {
		t.Fatalf("dropped stride shard survived: %v", err)
	}
	duplicateGroups := append(append([]unicodeNodeTopRange{}, unicodeNodeTopRanges...), unicodeNodeTopRanges[0])
	if err := checkUnicodeTopRangeCoverage(lines, shards, duplicateGroups); err == nil {
		t.Fatal("duplicate top-level range survived")
	}
	overlappingGroups := append([]unicodeNodeTopRange{}, unicodeNodeTopRanges...)
	overlappingGroups[1].start--
	if err := checkUnicodeTopRangeCoverage(lines, shards, overlappingGroups); err == nil || !strings.Contains(err.Error(), "owned 2 times") {
		t.Fatalf("overlapping top-level range survived: %v", err)
	}
	t.Log("dropped stride shard, dropped range, duplicate range and overlapping range mutants caught")
	// The actual Test declarations must equal the dispatch inventory. The runner
	// also verifies t.Name(), so a wrapper cannot dispatch another range silently.
	file, err := parser.ParseFile(token.NewFileSet(), "canonicalize_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	declarations := map[string]int{}
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok && strings.HasPrefix(fn.Name.Name, "TestCanonicalizeUnicodeNode") {
			declarations[fn.Name.Name]++
		}
	}
	if len(declarations) != len(unicodeNodeTopRanges) {
		t.Fatal("top-level Test declarations differ from range inventory")
	}
	for _, group := range unicodeNodeTopRanges {
		if declarations[group.name] != 1 {
			t.Fatalf("%s declared %d times", group.name, declarations[group.name])
		}
	}
	t.Logf("all %d top-level functions partition all %d shards exactly once", len(unicodeNodeTopRanges), len(shards))

}

// unicodeShardPlants are the two planted failures a shard must catch alone, each a change to canonicalize_test.go
// that leaves valid Node patterns and every width, count and scanner control intact.
var unicodeShardPlants = map[string]buildcache.Mutation{
	// A wrong expected equivalence class: only the selected first shard fails its Node comparison.
	"class": {File: "internal/unicodeproperties/canonicalize_test.go", Before: "body := text.String()", After: "body := text.String()" + "\n if index == 0 { body = fmt.Sprintf(\"%X %X\", canon, canon) }"},
	// A non-singleton appended to the first stride: a valid GROUP pattern whose Node fold adds U+0061.
	"stride": {File: "internal/unicodeproperties/canonicalize_test.go", Before: "for _, group := range strides {", After: `for index, group := range strides {
		if index == 0 { group = append(append([]uint32{}, group...), 0x41) }`},
}

// unicodeShardPlantBinary is this package's test binary with a plant from unicodeShardPlants, a product built ahead
// (buildcache.GoTestMutated): the mutated source lives in the product, and the unit only runs it (#5qykzj5).
func unicodeShardPlantBinary(t testing.TB, plant string) string {
	t.Helper()
	return buildcache.GoTestMutated(t, "", "unicodeproperties.test", "./internal/unicodeproperties", nil, []buildcache.Mutation{unicodeShardPlants[plant]})
}

func TestProduct_UnicodePropertiesClassPlant(t *testing.T) {
	t.Parallel()
	unicodeShardPlantBinary(t, "class")
}

func TestProduct_UnicodePropertiesStridePlant(t *testing.T) {
	t.Parallel()
	unicodeShardPlantBinary(t, "stride")
}

// Plant a wrong expected equivalence class, leaving valid Node patterns and
// all width, count and scanner controls intact. Only the selected first shard
// fails its Node comparison; a neighboring top-level range remains green.
func TestUnicodeNodeShardPlantedFailure(t *testing.T) {
	t.Parallel()
	runUnicodeShardPlant(t, "class", "^TestCanonicalizeUnicodeNodeRange[01]$", "BAD CLASS iu 61 EXTRA 41", "TestCanonicalizeUnicodeNodeRange0/0000-0080-", "TestCanonicalizeUnicodeNodeRange1")
}

// A non-singleton appended to the first stride is a valid GROUP pattern whose
// Node fold adds U+0061. Only that stride's shard may report BAD; all other
// shards in its range and the neighboring stride range must pass.
func TestUnicodeNodeStridePlantedFailure(t *testing.T) {
	t.Parallel()
	runUnicodeShardPlant(t, "stride", "^TestCanonicalizeUnicodeNodeRange[45]$", "BAD GROUP iu n", "TestCanonicalizeUnicodeNodeRange4/4050-4052-", "TestCanonicalizeUnicodeNodeRange5")
}

// runUnicodeShardPlant runs the planted test binary in this package's directory, as go test ran it. Keying refuses a
// plant whose Before isn't in canonicalize_test.go exactly once, so a moved plant site fails here. The child runs as
// many Node shards at once as this test's share of the unit's CPUs (testguard.Share).
func runUnicodeShardPlant(t *testing.T, plant, run, bad, failedShard, passedRange string) {
	t.Helper()
	command := exec.Command(unicodeShardPlantBinary(t, plant), "-test.count=1", "-test.parallel="+strconv.Itoa(testguard.Share()), "-test.timeout=75s", "-test.run", run, "-test.v")
	output, err := command.CombinedOutput()
	if err == nil || bytes.Contains(output, []byte("build failed")) || !bytes.Contains(output, []byte(bad)) || !bytes.Contains(output, []byte("--- FAIL: "+failedShard)) || !bytes.Contains(output, []byte("--- PASS: "+passedRange+" (")) {
		t.Fatalf("plant did not fail only its Node shard: %v %s", err, output)
	}
	if strings.HasPrefix(bad, "BAD GROUP") && (!bytes.Contains(output, []byte("EXTRA 61 MISSING")) || bytes.Count(output, []byte("BAD GROUP")) != 1) {
		t.Fatalf("stride plant did not report only Node's extra folded member: %s", output)
	}
	if bytes.Count(output, []byte("--- FAIL:")) != 2 {
		t.Fatalf("plant failed more than its own shard and owning top-level function: %s", output)
	}
	t.Logf("plant caught only in %s; every other shard and %s passed: %s", failedShard, passedRange, output)
}
