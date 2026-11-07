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

	"github.com/system-inc/adamic/internal/childguard"
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

func consumerRows(test *testing.T) []consumerRow {
	test.Helper()
	data, err := os.ReadFile("testdata/consumers.json")
	if err != nil {
		test.Fatal(err)
	}
	var raws []json.RawMessage
	if err := json.Unmarshal(data, &raws); err != nil {
		test.Fatal(err)
	}
	rows := make([]consumerRow, len(raws))
	for i, raw := range raws {
		rows[i].raw = raw
		if err := json.Unmarshal(raw, &rows[i]); err != nil {
			test.Fatal(err)
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
func consumerShard(test *testing.T, shard int) []consumerRow {
	test.Helper()
	var owned []consumerRow
	for _, row := range consumerRows(test) {
		if consumerOwner(row) == shard {
			owned = append(owned, row)
		}
	}
	return owned
}

func TestConsumerCommentHelpers_000(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 0)
}
func TestConsumerCommentHelpers_001(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 1)
}
func TestConsumerCommentHelpers_002(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 2)
}
func TestConsumerCommentHelpers_003(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 3)
}
func TestConsumerCommentHelpers_004(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 4)
}
func TestConsumerCommentHelpers_005(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 5)
}
func TestConsumerCommentHelpers_006(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 6)
}
func TestConsumerCommentHelpers_007(test *testing.T) {
	test.Parallel()
	runConsumerShard(test, 7)
}

// The inventory assumes a common AST adapter. This comparison isolates that
// contract from the separate stage-1 parser, using Go's actual traversal spans.
// Each shard holds its rows on Go, Node and sanitized native, as the whole corpus was.
func runConsumerShard(test *testing.T, shard int) {
	test.Helper()
	rows := consumerShard(test, shard)
	if len(rows) == 0 {
		test.Fatalf("consumer shard %03d owns no rows", shard)
	}
	// The planted-failure subprocess names one row by its corpus index: its owner's comparison sees a wrong output,
	// every other shard passes without running the corpus.
	if probe := os.Getenv("ADAMIC_COMMENT_CONSUMER_PROBE"); probe != "" {
		index, err := strconv.Atoi(probe)
		if err != nil {
			test.Fatal(err)
		}
		planted := consumerRows(test)[index]
		for _, row := range rows {
			if row.Name == planted.Name && row.Source == planted.Source {
				compare(test, []byte("planted\n"), []byte("case "+row.Name+"\n"))
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
		test.Fatal(err)
	}
	original := filepath.Join(test.TempDir(), "consumers.json")
	if err := os.WriteFile(original, corpus, 0644); err != nil {
		test.Fatal(err)
	}
	goOracle := oracle(test)
	adapted := run(test, "", goOracle, "--ast", original)
	path := filepath.Join(test.TempDir(), "adapted.json")
	if err := os.WriteFile(path, adapted, 0644); err != nil {
		test.Fatal(err)
	}
	want := run(test, "", goOracle, original)
	runner, _ := filepath.Abs("../../../../../oracle/node.mjs")
	entry, _ := filepath.Abs("main.ts")
	compare(test, run(test, "", "node", "--disable-warning=ExperimentalWarning", runner, entry, path), want)
	compare(test, run(test, "", helpersNative(test, nativeMutation{}), path), want)
	test.Logf(
		"shard %03d: Go, Node and sanitized native agree on %d consumer rows, %d output lines with the stated AST adapter",
		shard,
		len(rows),
		bytes.Count(want, []byte("\n")),
	)
}

func TestConsumerCommentHelpersUnion(test *testing.T) {
	test.Parallel()
	rows := consumerRows(test)
	expected := map[string]bool{}
	for _, row := range rows {
		key := row.Name + "\x00" + row.Source
		if row.Name == "" || expected[key] {
			test.Fatalf("empty or repeated consumer row %q", row.Name)
		}
		expected[key] = true
	}
	actual := map[string]bool{}
	for shard := 0; shard < consumerShards; shard++ {
		owned := consumerShard(test, shard)
		if len(owned) == 0 {
			test.Fatalf("consumer shard %03d owns no rows", shard)
		}
		for _, row := range owned {
			key := row.Name + "\x00" + row.Source
			if actual[key] || !expected[key] {
				test.Fatalf("repeated or unexpected shard row %q", row.Name)
			}
			actual[key] = true
		}
	}
	if len(actual) != len(expected) {
		test.Fatalf("union %d rows, unsplit %d", len(actual), len(expected))
	}
	test.Logf("union: %d consumer rows over %d shards, each exactly once", len(actual), consumerShards)
}

func TestConsumerCommentHelpersPlantedFailure(test *testing.T) {
	test.Parallel()
	binary, err := os.Executable()
	if err != nil {
		test.Fatal(err)
	}
	planted := consumerRows(test)[0]
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
		for _, value := range os.Environ() {
			if !strings.HasPrefix(value, "ADAMIC_COMMENT_CONSUMER_PROBE=") {
				command.Env = append(command.Env, value)
			}
		}
		command.Env = append(command.Env, "ADAMIC_COMMENT_CONSUMER_PROBE=0")
		// Read through childguard.Capture, to its end with no timer deciding when: exec's own copying under a
		// one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after the
		// shard exited, so a shard that had finished could fail this check under load (#n3gk22h's shape). ctx's 90 s
		// deadline ends a hung shard. Stdout comes before stderr; the checks below look for their lines anywhere in
		// either.
		stdout, stderr, err := childguard.Capture(ctx, command)
		output := append(stdout, stderr...)
		cancel()
		if shard == owner {
			if err == nil ||
				!bytes.Contains(output, []byte(`output line 1: got "planted"`)) ||
				!bytes.Contains(output, []byte("--- FAIL: "+name)) {
				test.Fatalf("%s did not catch the planted row: %v\n%s", name, err, output)
			}
			test.Logf("planted row %q caught only by %s", planted.Name, name)
		} else if err != nil {
			test.Fatalf("planted row leaked into %s: %v\n%s", name, err, output)
		}
	}
}
