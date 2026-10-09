package unicodeproperties

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	unicodeScanShardLines       = 16
	unicodeStrideScanShardLines = 8
	// The ordered census has 4050 class and dense-group lines, then 1024 strides.
	unicodeStrideScanStart = 4050
)

// The first two blocks contain the cheap class and dense singleton scans.
// Each remaining block owns four eight-line singleton-stride shards, one wave
// on four CPUs. The last dense shard stops at the first stride line.
// These boundaries partition all 382 shards without changing the scan inputs.
type unicodeNodeTopRange struct {
	name       string
	start, end int
}

var unicodeNodeTopRanges = []unicodeNodeTopRange{
	{"TestCanonicalizeUnicodeNodeRange0", 0, 186},
	{"TestCanonicalizeUnicodeNodeRange1", 186, 254},
	{"TestCanonicalizeUnicodeNodeRange2", 254, 258},
	{"TestCanonicalizeUnicodeNodeRange3", 258, 262},
	{"TestCanonicalizeUnicodeNodeRange4", 262, 266},
	{"TestCanonicalizeUnicodeNodeRange5", 266, 270},
	{"TestCanonicalizeUnicodeNodeRange6", 270, 274},
	{"TestCanonicalizeUnicodeNodeRange7", 274, 278},
	{"TestCanonicalizeUnicodeNodeRange8", 278, 282},
	{"TestCanonicalizeUnicodeNodeRange9", 282, 286},
	{"TestCanonicalizeUnicodeNodeRange10", 286, 290},
	{"TestCanonicalizeUnicodeNodeRange11", 290, 294},
	{"TestCanonicalizeUnicodeNodeRange12", 294, 298},
	{"TestCanonicalizeUnicodeNodeRange13", 298, 302},
	{"TestCanonicalizeUnicodeNodeRange14", 302, 306},
	{"TestCanonicalizeUnicodeNodeRange15", 306, 310},
	{"TestCanonicalizeUnicodeNodeRange16", 310, 314},
	{"TestCanonicalizeUnicodeNodeRange17", 314, 318},
	{"TestCanonicalizeUnicodeNodeRange18", 318, 322},
	{"TestCanonicalizeUnicodeNodeRange19", 322, 326},
	{"TestCanonicalizeUnicodeNodeRange20", 326, 330},
	{"TestCanonicalizeUnicodeNodeRange21", 330, 334},
	{"TestCanonicalizeUnicodeNodeRange22", 334, 338},
	{"TestCanonicalizeUnicodeNodeRange23", 338, 342},
	{"TestCanonicalizeUnicodeNodeRange24", 342, 346},
	{"TestCanonicalizeUnicodeNodeRange25", 346, 350},
	{"TestCanonicalizeUnicodeNodeRange26", 350, 354},
	{"TestCanonicalizeUnicodeNodeRange27", 354, 358},
	{"TestCanonicalizeUnicodeNodeRange28", 358, 362},
	{"TestCanonicalizeUnicodeNodeRange29", 362, 366},
	{"TestCanonicalizeUnicodeNodeRange30", 366, 370},
	{"TestCanonicalizeUnicodeNodeRange31", 370, 374},
	{"TestCanonicalizeUnicodeNodeRange32", 374, 378},
	{"TestCanonicalizeUnicodeNodeRange33", 378, 382},
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
	overlap[1] = unicodeScanShard{15, 31, strings.Join(lines[15:31], "\n") + "\n"}
	if err := checkUnicodeShardCoverage(lines, overlap); err == nil || !strings.Contains(err.Error(), "covered 2 times") {
		t.Fatalf("overlapping shard survived: %v", err)
	}
	crossingStride := append([]unicodeScanShard{}, shards...)
	crossingStride[253] = unicodeScanShard{4048, 4051, strings.Join(lines[4048:4051], "\n") + "\n"}
	if err := checkUnicodeShardCoverage(lines, crossingStride); err == nil || !strings.Contains(err.Error(), "invalid stride shard range") {
		t.Fatalf("shard crossing into strides survived: %v", err)
	}
	wideStride := append([]unicodeScanShard{}, shards...)
	wideStride[254] = unicodeScanShard{4050, 4059, strings.Join(lines[4050:4059], "\n") + "\n"}
	if err := checkUnicodeShardCoverage(lines, wideStride); err == nil || !strings.Contains(err.Error(), "invalid stride shard range") {
		t.Fatalf("nine-line stride shard survived: %v", err)
	}
	t.Log("crossing stride boundary and oversized stride shard mutants caught")
	t.Logf("%d shards cover all %d scans exactly once, %d code points", len(shards), len(lines), int64(len(lines))*0x110000)
	if err := checkUnicodeTopRangeCoverage(lines, shards, unicodeNodeTopRanges); err != nil {
		t.Fatal(err)
	}
	if len(shards) != 382 || shards[254].start != unicodeStrideScanStart {
		t.Fatal("dense and stride shard census changed")
	}
	if len(unicodeNodeTopRanges) != 34 {
		t.Fatal("top-level range census changed")
	}
	if err := checkUnicodeTopRangeCoverage(lines, shards, unicodeNodeTopRanges[1:]); err == nil || !strings.Contains(err.Error(), "owned 0 times") {
		t.Fatalf("dropped top-level range survived: %v", err)
	}
	droppedStride := append([]unicodeNodeTopRange{}, unicodeNodeTopRanges...)
	droppedStride[2].start++
	if err := checkUnicodeTopRangeCoverage(lines, shards, droppedStride); err == nil || !strings.Contains(err.Error(), "shard 254 owned 0 times") {
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

// Plant a wrong expected equivalence class, leaving valid Node patterns and
// all width, count and scanner controls intact. Only the selected first shard
// fails its Node comparison; a neighboring top-level range remains green.
func TestUnicodeNodeShardPlantedFailure(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("canonicalize_test.go")
	if err != nil {
		t.Fatal(err)
	}
	old := "body := text.String()"
	if strings.Count(string(source), old) != 1 {
		t.Fatal("plant site changed")
	}
	changed := strings.Replace(string(source), old, old+"\n if index == 0 { body = fmt.Sprintf(\"%X %X\", canon, canon) }", 1)
	runUnicodeShardPlant(t, changed, "^TestCanonicalizeUnicodeNodeRange[01]$", "BAD CLASS iu 61 EXTRA 41", "TestCanonicalizeUnicodeNodeRange0/0000-0016-", "TestCanonicalizeUnicodeNodeRange1")
}

// A non-singleton appended to the first stride is a valid GROUP pattern whose
// Node fold adds U+0061. Only that stride's shard may report BAD; all other
// shards in its range and the neighboring stride range must pass.
func TestUnicodeNodeStridePlantedFailure(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("canonicalize_test.go")
	if err != nil {
		t.Fatal(err)
	}
	old := "for _, group := range strides {"
	if strings.Count(string(source), old) != 1 {
		t.Fatal("stride plant site changed")
	}
	changed := strings.Replace(string(source), old, `for index, group := range strides {
		if index == 0 { group = append(append([]uint32{}, group...), 0x41) }`, 1)
	runUnicodeShardPlant(t, changed, "^TestCanonicalizeUnicodeNodeRange[23]$", "BAD GROUP iu n", "TestCanonicalizeUnicodeNodeRange2/4050-4058-", "TestCanonicalizeUnicodeNodeRange3")
}

func runUnicodeShardPlant(t *testing.T, changed, run, bad, failedShard, passedRange string) {
	t.Helper()
	dir := t.TempDir()
	replacement := filepath.Join(dir, "canonicalize_test.go")
	if err := os.WriteFile(replacement, []byte(changed), 0600); err != nil {
		t.Fatal(err)
	}
	original, err := filepath.Abs("canonicalize_test.go")
	if err != nil {
		t.Fatal(err)
	}
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{original: replacement}})
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "overlay.json")
	if err = os.WriteFile(path, overlay, 0600); err != nil {
		t.Fatal(err)
	}
	command := exec.Command("go", "test", "-overlay", path, ".", "-count=1", "-parallel=4", "-timeout=75s", "-run", run, "-v")
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
