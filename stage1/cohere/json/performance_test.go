package json

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The profiled -O2 -g binaries must answer the full corpus, not just the profile sample.
const testProfileSnapshotsAgreeShards = 518

// Profile binaries are provided inputs. ADAMIC_TEST_SHARD=i/n selects ordinal modulo n locally.
func TestProfileSnapshotsAgree(t *testing.T) {
	t.Parallel()
	snapshots := os.Getenv("ADAMIC_JSON_PROFILE_BINARIES")
	if snapshots == "" {
		t.Skip("set ADAMIC_JSON_PROFILE_BINARIES to profile snapshot binaries")
	}
	cases := corpusCases(t)
	shards := jsonPortShards(cases)
	if len(shards) != testProfileSnapshotsAgreeShards {
		t.Fatalf("enumerated %d shards, declared %d", len(shards), testProfileSnapshotsAgreeShards)
	}
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	index, count, err := jsonPortSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	tools, err := jsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(jsonBuildUnit(t, "build-go-oracle", jsonGoOracleInputs(tools), buildJSONGoOracle), "go-cohere")
	for ordinal, shard := range shards {
		if ordinal%count != index {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", ordinal), func(t *testing.T) {
			t.Parallel()
			items := cases[shard.start:shard.end]
			answers := jsonOracleAnswers(t, oracle, items)
			input, expected := protocol(items, answers)
			path := filepath.Join(t.TempDir(), "cases.txt")
			if err := os.WriteFile(path, []byte(input), 0644); err != nil {
				t.Fatal(err)
			}
			for _, binary := range filepath.SplitList(snapshots) {
				compare(t, t.Name()+" "+binary, execute(t, nil, binary, "--cases", path), expected, items)
			}
			t.Logf("case range [%d:%d]; %d snapshots; %d cases", shard.start, shard.end, len(filepath.SplitList(snapshots)), len(items))
		})
	}
	t.Logf("exact union: %d cases", len(cases))
}

const testCachedWidthMutantIsCaughtShards = 1

// ADAMIC_TEST_SHARD=i/n selects ordinal modulo n; unset runs every shard.
func TestCachedWidthMutantIsCaught(t *testing.T) {
	t.Parallel()
	cases := []textCase{{"probe.json", `["` + strings.Repeat("wide", 25) + `",1]`}}
	shards := []nativeChunk{{start: 0, end: len(cases)}}
	if len(shards) != testCachedWidthMutantIsCaughtShards {
		t.Fatal("cached-width shard count changed")
	}
	if err := jsonPortUnion(cases, shards); err != nil {
		t.Fatal(err)
	}
	index, count, err := jsonPortSelection(os.Getenv("ADAMIC_TEST_SHARD"))
	if err != nil {
		t.Fatal(err)
	}
	tools, err := jsonGoToolchain()
	if err != nil {
		t.Fatal(err)
	}
	oracle := filepath.Join(jsonBuildUnit(t, "build-go-oracle", jsonGoOracleInputs(tools), buildJSONGoOracle), "go-cohere")
	mutation := printerMutation{"zero cached width", "doc.ts", "width: stringWidth(text)", "width: 0"}
	inputs := jsonLoweredPortInputs(tools)
	inputs.Name = "json zero-cached-width lowered mutant"
	inputs.Files = append(inputs.Files, "stage1/cohere/json/performance_test.go")
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutation=%+v", mutation))
	build := func(dir string) error { return buildJSONLoweredPortMutation(dir, &mutation) }
	lowered := jsonBuildUnit(t, "build-lowered-mutant", inputs, build)
	source, err := os.ReadFile(filepath.Join(lowered, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	clang, err := jsonClangToolchain()
	if err != nil {
		t.Fatal(err)
	}
	nativeInputs := jsonNativePortInputs(string(source), false, clang)
	nativeInputs.Name = "json zero-cached-width native mutant"
	binary := filepath.Join(jsonBuildUnit(t, "build-native-mutant", nativeInputs, buildJSONReleasePort(string(source))), "port")
	for ordinal, shard := range shards {
		if ordinal%count != index {
			continue
		}
		t.Run(fmt.Sprintf("shard-%03d", ordinal), func(t *testing.T) {
			t.Parallel()
			items := cases[shard.start:shard.end]
			answers := jsonOracleAnswers(t, oracle, items)
			input, expected := protocol(items, answers)
			path := filepath.Join(t.TempDir(), "cases.txt")
			if err := os.WriteFile(path, []byte(input), 0644); err != nil {
				t.Fatal(err)
			}
			for _, side := range []struct {
				name   string
				result run
			}{{"native release", execute(t, nil, binary, "--cases", path)}, {"Node", onNode(t, filepath.Join(lowered, "main.ts"), "--cases", path)}} {
				if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
					t.Fatalf("%s %s mutant must execute successfully: %+v", t.Name(), side.name, side.result)
				}
				if string(side.result.stdout) == expected {
					t.Fatalf("%s %s failed to catch zero cached width", t.Name(), side.name)
				}
				t.Logf("%s caught zero cached width in case range [%d:%d]", side.name, shard.start, shard.end)
			}
		})
	}
}
