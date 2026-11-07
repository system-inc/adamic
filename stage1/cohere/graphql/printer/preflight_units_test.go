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
	for index, item := range shard.cases {
		got, want := rows[index], item.want
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
		return counts, fmt.Errorf(
			"shard-%03d: %d unexpected differences: %s",
			number,
			counts.unexpected,
			strings.Join(counts.details, "; "),
		)
	}
	if counts.known != 5 {
		return counts, fmt.Errorf(
			"shard-%03d known difference count changed: %d, recorded 5 in GAPS.md",
			number,
			counts.known,
		)
	}
	return counts, nil
}

func printerPreflightInputs(test *testing.T) (script, library, embedded string) {
	test.Helper()
	library = os.Getenv("ADAMIC_GRAPHQL_PRETTIER")
	if library == "" {
		test.Skip("set ADAMIC_GRAPHQL_PRETTIER to prettier@3.9.6 and graphql@17.0.2")
	}
	var absolutePathError error
	script, absolutePathError = filepath.Abs("testdata/prettier.mjs")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	embedded, absolutePathError = filepath.Abs(filepath.Join(repository, "cohere/internal/format/prettier/bundles"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return
}

func printerUpstreamUnit(test *testing.T, unit int) {
	setupStart := time.Now()
	test.Helper()
	script, library, embedded := printerPreflightInputs(test)
	var whole []printerCase
	var shards []printerShard
	for _, mode := range []string{"defaults", "narrow", "tight", "tabs"} {
		cases, want := printerCases(test, mode)
		enumeration := enumeratePrinter(test, mode, cases, want)
		whole = append(whole, enumeration...)
		shards = append(shards, printerShard{mode: mode, path: cases, cases: enumeration})
	}
	if len(shards) != testPrinterUpstreamPreflightShards {
		test.Fatalf("enumerated %d shards, declared %d", len(shards), testPrinterUpstreamPreflightShards)
	}
	if printerShardUnionError := printerShardUnion(whole, shards); printerShardUnionError != nil {
		test.Fatal(printerShardUnionError)
	}
	test.Logf("union: %d unique mode/case ids across %d shards", len(whole), len(shards))
	if unit < 0 {
		test.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
		return
	}
	selected, printerShardSelectionError := printerShardSelection(os.Getenv("ADAMIC_TEST_SHARD"), len(shards))
	if printerShardSelectionError != nil {
		test.Fatal(printerShardSelectionError)
	}
	for number, shard := range shards {
		if number != unit || !selected[number] {
			continue
		}
		{
			test.Logf("setup wall %.3fs", time.Since(setupStart).Seconds())
			start := time.Now()
			defer func() { test.Logf("own work wall %.3fs", time.Since(start).Seconds()) }()
			test.Cleanup(func() {
				if elapsed := time.Since(start); elapsed > 30*time.Second {
					test.Errorf("invalid test unit: %.3fs exceeds 30s", elapsed.Seconds())
				}
			})
			test.Logf("mode %s, cases 0..%d", shard.mode, len(shard.cases)-1)
			for _, side := range []struct{ name, directory, engine string }{
				{"npm Prettier", library, "npm"},
				{"embedded fork", embedded, "embedded"},
			} {
				got := execute(test, nil, "node", script, side.directory, shard.path, shard.mode, side.engine)
				counts, printerPreflightDisagreementError := printerPreflightDisagreement(number, shard, got)
				test.Logf(
					"%s: %d texts: %d byte-identical formatted, %d shared refusals, %d known whitespace differences, %d unexpected differences",
					side.name,
					len(shard.cases),
					counts.accepted,
					counts.refused,
					counts.known,
					counts.unexpected,
				)
				if printerPreflightDisagreementError != nil {
					test.Errorf("%s: %v", side.name, printerPreflightDisagreementError)
				}
			}
		}
	}
}

func TestPrinterPreflightPlantedDisagreement(test *testing.T) {
	test.Parallel()
	script, library, _ := printerPreflightInputs(test)
	var shards []printerShard
	var results []run
	for number, mode := range []string{"defaults", "tabs"} {
		cases, want := printerCases(test, mode)
		shard := printerShard{mode: mode, path: cases, cases: enumeratePrinter(test, mode, cases, want)}
		got := execute(test, nil, "node", script, library, shard.path, mode, "npm")
		if _, printerPreflightDisagreementError := printerPreflightDisagreement(
			number,
			shard,
			got,
		); printerPreflightDisagreementError != nil {
			test.Fatal(printerPreflightDisagreementError)
		}
		shards = append(shards, shard)
		results = append(results, got)
	}
	shards[0].cases[0].want += "planted disagreement"
	caught := 0
	for number, shard := range shards {
		if _, printerPreflightDisagreementError := printerPreflightDisagreement(
			number,
			shard,
			results[number],
		); printerPreflightDisagreementError != nil {
			caught++
			if number != 0 ||
				!strings.Contains(printerPreflightDisagreementError.Error(), "shard-000") ||
				!strings.Contains(printerPreflightDisagreementError.Error(), "defaults/case-000000") {
				test.Fatalf("wrong owner: %v", printerPreflightDisagreementError)
			}
			test.Logf("planted disagreement caught: %v", printerPreflightDisagreementError)
		}
	}
	if caught != 1 {
		test.Fatalf("planted disagreement caught by %d shards, want exactly one", caught)
	}
}
