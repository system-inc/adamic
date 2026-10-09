package printer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type printerPreflightCounts struct {
	accepted, refused, known, unexpected int
	details                              []string
}

func printerPreflightDisagreement(number int, shard printerShard, result run) (printerPreflightCounts, error) {
	counts := printerPreflightCounts{}
	if result.exitCode != 0 || len(result.stderr) > 0 {
		return counts, fmt.Errorf("shard-%03d: exit %d, %s", number, result.exitCode, result.stderr)
	}
	rows := strings.Split(string(result.stdout), "\n")
	if len(rows) != len(shard.cases)+1 || rows[len(rows)-1] != "" {
		return counts, fmt.Errorf("shard-%03d answer count: %d vs %d", number, len(rows)-1, len(shard.cases))
	}
	exactKnown := map[string]string{
		">":       "error\tSyntax Error: Unexpected <EOF>. (1:1)",
		"> ":      "error\tSyntax Error: Unexpected <EOF>. (1:2)",
		">\\n\\r": "error\tSyntax Error: Unexpected <EOF>. (3:1)",
		">\\r":    "error\tSyntax Error: Unexpected <EOF>. (2:1)",
	}
	for i, item := range shard.cases {
		got, want := rows[i], item.want
		if got == want {
			counts.accepted++
			continue
		}
		if strings.HasPrefix(got, "error\t") && strings.HasPrefix(want, "error\t") {
			counts.refused++
			continue
		}
		if expected, exists := exactKnown[item.input]; exists && got == "ok\t" && want == expected {
			counts.known++
			continue
		}
		counts.unexpected++
		if counts.unexpected <= 5 {
			counts.details = append(counts.details, fmt.Sprintf("%s: Prettier %q; Go %q", item.id, got, want))
		}
	}
	if counts.unexpected != 0 {
		return counts, fmt.Errorf("shard-%03d: %d unexpected differences: %s", number, counts.unexpected, strings.Join(counts.details, "; "))
	}
	if counts.known != 5 {
		return counts, fmt.Errorf("shard-%03d known difference count changed: %d, recorded 5 in GAPS.md", number, counts.known)
	}
	return counts, nil
}

func printerPreflightInputs(t *testing.T) (script, library, embedded string) {
	t.Helper()
	library = os.Getenv("ADAMIC_GRAPHQL_PRETTIER")
	if library == "" {
		t.Skip("set ADAMIC_GRAPHQL_PRETTIER to prettier@3.9.6 and graphql@17.0.2")
	}
	var err error
	script, err = filepath.Abs("testdata/prettier.mjs")
	if err != nil {
		t.Fatal(err)
	}
	embedded, err = filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	if err != nil {
		t.Fatal(err)
	}
	return
}

func printerUpstreamUnit(t *testing.T, unit int) {
	t.Helper()
	script, library, embedded := printerPreflightInputs(t)
	oracle := printerOracle(t)
	var whole []printerCase
	var shards []printerShard
	for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
		cases, want := printerCases(t, mode, oracle)
		enumeration := enumeratePrinter(t, mode, cases, want)
		whole = append(whole, enumeration...)
		shards = append(shards, printerShard{mode: mode, path: cases, cases: enumeration})
	}
	if len(shards) != testPrinterUpstreamPreflightShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterUpstreamPreflightShards)
	}
	if err := printerShardUnion(whole, shards); err != nil {
		t.Fatal(err)
	}
	t.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
	if unit < 0 {
		return
	}
	selected, err := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if err != nil {
		t.Fatal(err)
	}
	for number, shard := range shards {
		if number != unit || !selected[number] {
			continue
		}
		{
			start := time.Now()
			t.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					t.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			t.Logf("mode %s, cases 0..%d", shard.mode, len(shard.cases)-1)
			for _, side := range []struct{ name, directory, engine string }{{"npm Prettier", library, "npm"}, {"embedded fork", embedded, "embedded"}} {
				got := execute(t, nil, "node", script, side.directory, shard.path, shard.mode, side.engine)
				counts, err := printerPreflightDisagreement(number, shard, got)
				t.Logf("%s: %d texts: %d byte-identical formatted, %d shared refusals, %d known whitespace differences, %d unexpected differences", side.name, len(shard.cases), counts.accepted, counts.refused, counts.known, counts.unexpected)
				if err != nil {
					t.Errorf("%s: %v", side.name, err)
				}
			}
		}
	}
}

func TestPrinterPreflightPlantedDisagreement(t *testing.T) {
	t.Parallel()
	script, library, _ := printerPreflightInputs(t)
	oracle := printerOracle(t)
	var shards []printerShard
	var results []run
	for number, mode := range []string{"defaults", "tabs"} {
		cases, want := printerCases(t, mode, oracle)
		shard := printerShard{mode: mode, path: cases, cases: enumeratePrinter(t, mode, cases, want)}
		got := execute(t, nil, "node", script, library, shard.path, mode, "npm")
		if _, err := printerPreflightDisagreement(number, shard, got); err != nil {
			t.Fatal(err)
		}
		shards = append(shards, shard)
		results = append(results, got)
	}
	shards[0].cases[0].want += "planted disagreement"
	caught := 0
	for number, shard := range shards {
		if _, err := printerPreflightDisagreement(number, shard, results[number]); err != nil {
			caught++
			if number != 0 || !strings.Contains(err.Error(), "shard-000") || !strings.Contains(err.Error(), "defaults/case-000000") {
				t.Fatalf("wrong owner: %v", err)
			}
			t.Logf("planted disagreement caught: %v", err)
		}
	}
	if caught != 1 {
		t.Fatalf("planted disagreement caught by %d shards, want exactly one", caught)
	}
}
