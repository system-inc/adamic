package comments

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The 4,513 consumer sources split into buckets by a hash of each row's name and source, so a row's owner is fixed by
// its content and growing the corpus never moves another row. Seven max-lines sources of about 55 KB each cost the
// sanitized native about 2.7 s apiece on one core; whole, with its own native build, the corpus was 91 s alone on one
// core and 263 s on a loaded box (Workshop, Oct 11). Eight shards are each at most about 7 s alone.
const consumerShards = 8

type consumerRow struct {
	raw          json.RawMessage
	Name, Source string
}

func consumerRows(t *testing.T) []consumerRow {
	t.Helper()
	data, err := os.ReadFile("testdata/consumers.json")
	if err != nil {
		t.Fatal(err)
	}
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		t.Fatal(err)
	}
	rows := make([]consumerRow, len(raws))
	for i, raw := range raws {
		rows[i].raw = raw
		if err := json.Unmarshal(raw, &rows[i]); err != nil {
			t.Fatal(err)
		}
	}
	return rows
}

func consumerOwner(row consumerRow) int {
	hash := fnv.New64a()
	hash.Write([]byte(row.Name))
	hash.Write([]byte{0})
	hash.Write([]byte(row.Source))
	return int(hash.Sum64() % consumerShards)
}

// consumerShard is the rows shard owns, in corpus order.
func consumerShard(t *testing.T, shard int) []consumerRow {
	t.Helper()
	var owned []consumerRow
	for _, row := range consumerRows(t) {
		if consumerOwner(row) == shard {
			owned = append(owned, row)
		}
	}
	return owned
}

func TestConsumerCommentHelpers_000(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 0)
}
func TestConsumerCommentHelpers_001(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 1)
}
func TestConsumerCommentHelpers_002(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 2)
}
func TestConsumerCommentHelpers_003(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 3)
}
func TestConsumerCommentHelpers_004(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 4)
}
func TestConsumerCommentHelpers_005(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 5)
}
func TestConsumerCommentHelpers_006(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 6)
}
func TestConsumerCommentHelpers_007(t *testing.T) {
	t.Parallel()
	runConsumerShard(t, 7)
}

// The inventory assumes a common AST adapter. This comparison isolates that
// contract from the separate stage-1 parser, using Go's actual traversal spans.
// Each shard holds its rows on Go, Node and sanitized native, as the whole corpus was.
func runConsumerShard(t *testing.T, shard int) {
	t.Helper()
	rows := consumerShard(t, shard)
	if len(rows) == 0 {
		t.Fatalf("consumer shard %03d owns no rows", shard)
	}
	// The planted-failure subprocess names one row by its corpus index: its owner's comparison sees a wrong output,
	// every other shard passes without running the corpus.
	if probe := os.Getenv("ADAMIC_COMMENT_CONSUMER_PROBE"); probe != "" {
		index, err := strconv.Atoi(probe)
		if err != nil {
			t.Fatal(err)
		}
		planted := consumerRows(t)[index]
		for _, row := range rows {
			if row.Name == planted.Name && row.Source == planted.Source {
				compare(t, []byte("planted\n"), []byte("case "+row.Name+"\n"))
			}
		}
		return
	}
	raws := make([]json.RawMessage, len(rows))
	for i, row := range rows {
		raws[i] = row.raw
	}
	corpus, err := json.Marshal(raws)
	if err != nil {
		t.Fatal(err)
	}
	original := filepath.Join(t.TempDir(), "consumers.json")
	if err := os.WriteFile(original, corpus, 0644); err != nil {
		t.Fatal(err)
	}
	goOracle := oracle(t)
	adapted := run(t, "", goOracle, "--ast", original)
	path := filepath.Join(t.TempDir(), "adapted.json")
	if err := os.WriteFile(path, adapted, 0644); err != nil {
		t.Fatal(err)
	}
	want := run(t, "", goOracle, original)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	compare(t, run(t, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(t, run(t, "", helpersNative(t, nativeMutation{}), path), want)
	t.Logf("shard %03d: Go, Node and sanitized native agree on %d consumer rows, %d output lines with the stated AST adapter", shard, len(rows), bytes.Count(want, []byte("\n")))
}

func TestConsumerCommentHelpersUnion(t *testing.T) {
	t.Parallel()
	rows := consumerRows(t)
	expected := map[string]bool{}
	for _, row := range rows {
		key := row.Name + "\x00" + row.Source
		if row.Name == "" || expected[key] {
			t.Fatalf("empty or repeated consumer row %q", row.Name)
		}
		expected[key] = true
	}
	actual := map[string]bool{}
	for shard := 0; shard < consumerShards; shard++ {
		owned := consumerShard(t, shard)
		if len(owned) == 0 {
			t.Fatalf("consumer shard %03d owns no rows", shard)
		}
		for _, row := range owned {
			key := row.Name + "\x00" + row.Source
			if actual[key] || !expected[key] {
				t.Fatalf("repeated or unexpected shard row %q", row.Name)
			}
			actual[key] = true
		}
	}
	if len(actual) != len(expected) {
		t.Fatalf("union %d rows, unsplit %d", len(actual), len(expected))
	}
	t.Logf("union: %d consumer rows over %d shards, each exactly once", len(actual), consumerShards)
}

func TestConsumerCommentHelpersPlantedFailure(t *testing.T) {
	t.Parallel()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	planted := consumerRows(t)[0]
	owner := consumerOwner(planted)
	for shard := 0; shard < consumerShards; shard++ {
		name := fmt.Sprintf("TestConsumerCommentHelpers_%03d", shard)
		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		command := exec.CommandContext(ctx, binary, "-test.run=^"+name+"$", "-test.timeout=75s", "-test.v")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		command.Cancel = func() error {
			err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
			if err == syscall.ESRCH {
				return os.ErrProcessDone
			}
			return err
		}
		command.WaitDelay = time.Second
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "ADAMIC_COMMENT_CONSUMER_PROBE=") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, "ADAMIC_COMMENT_CONSUMER_PROBE=0")
		output, err := command.CombinedOutput()
		cancel()
		if shard == owner {
			if err == nil || !bytes.Contains(output, []byte(`output line 1: got "planted"`)) || !bytes.Contains(output, []byte("--- FAIL: "+name)) {
				t.Fatalf("%s did not catch the planted row: %v\n%s", name, err, output)
			}
			t.Logf("planted row %q caught only by %s", planted.Name, name)
		} else if err != nil {
			t.Fatalf("planted row leaked into %s: %v\n%s", name, err, output)
		}
	}
}
