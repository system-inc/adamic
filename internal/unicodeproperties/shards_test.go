package unicodeproperties

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const unicodeScanShardLines = 16

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
	for start := 0; start < len(lines); start += unicodeScanShardLines {
		end := min(start+unicodeScanShardLines, len(lines))
		shards = append(shards, unicodeScanShard{start, end, strings.Join(lines[start:end], "\n") + "\n"})
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
	t.Logf("%d shards cover all %d scans exactly once, %d code points", len(shards), len(lines), int64(len(lines))*0x110000)
}

// Plant a wrong expected equivalence class, leaving valid Node patterns and
// all width, count and scanner controls intact. Only the selected first shard
// fails its Node comparison; its neighboring shard remains green.
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
	dir := t.TempDir()
	replacement := filepath.Join(dir, "canonicalize_test.go")
	if err = os.WriteFile(replacement, []byte(changed), 0600); err != nil {
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
	command := exec.Command("go", "test", "-overlay", path, ".", "-count=1", "-parallel=1", "-run", "^TestCanonicalizeUnicodeNode$/(0000-0016|0016-0032)", "-v")
	output, err := command.CombinedOutput()
	if err == nil || bytes.Contains(output, []byte("build failed")) || !bytes.Contains(output, []byte("BAD CLASS iu 61 EXTRA 41")) || !bytes.Contains(output, []byte("--- FAIL: TestCanonicalizeUnicodeNode/0000-0016-")) || !bytes.Contains(output, []byte("--- PASS: TestCanonicalizeUnicodeNode/0016-0032-")) {
		t.Fatalf("plant did not fail only its Node shard: %v %s", err, output)
	}
	t.Logf("plant caught only in first shard; neighboring shard passed: %s", output)
}
