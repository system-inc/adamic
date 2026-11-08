package printer

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func TestPrinterConstructorGap(test *testing.T) {
	test.Parallel()
	path, _ := filepath.Abs("gaps/nestedConstructor.ts")
	result := onNode(test, path)
	if result.exitCode != 0 || string(result.stdout) != "ready\n" {
		test.Fatalf("Node: %+v", result)
	}
	program, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, lowerError := lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(lowerError, &refused) ||
		!strings.Contains(refused.What, "this escaping a constructor before every field is set") {
		test.Fatalf("gap changed; update GAPS.md and remove workaround: %v", lowerError)
	}
	test.Log(lowerError)
}

const printerWhitespaceWitnessShards = 5

// A real process emits one planted acceptance where the oracle requires refusal;
// only the owning whitespace shard may report the disagreement.
func whitespaceCaseId(index int) string {
	return fmt.Sprintf("stage1/cohere/graphql/printer/whitespace-witness/case-%06d", index)
}

func whitespaceOwner(key string) int {
	hash := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(hash[:8]) % printerWhitespaceWitnessShards)
}

func whitespaceShards(cases []printerCase) []printerShard {
	shards := make([]printerShard, printerWhitespaceWitnessShards)
	for index := range shards {
		shards[index].mode = "defaults"
	}
	for _, item := range cases {
		owner := whitespaceOwner(item.id)
		shards[owner].cases = append(shards[owner].cases, item)
	}
	return shards
}

func TestPrinterWhitespacePlantedDisagreement(test *testing.T) {
	test.Parallel()
	var whole []printerCase
	for index := range 5 {
		whole = append(whole, printerCase{id: whitespaceCaseId(index), want: "error\trefused"})
	}
	planted := whitespaceCaseId(3)
	caught := 0
	for number, shard := range whitespaceShards(whole) {
		var answer strings.Builder
		for _, item := range shard.cases {
			if item.id == planted {
				answer.WriteString("ok\t\n")
			} else {
				answer.WriteString(item.want + "\n")
			}
		}
		encoded, _ := json.Marshal(answer.String())
		result := execute(test, nil, "node", "-e", "process.stdout.write("+string(encoded)+")")
		printerShardDisagreementError := printerShardDisagreement(number, shard, result)
		if printerShardDisagreementError == nil {
			continue
		}
		if number != whitespaceOwner(planted) || !strings.Contains(printerShardDisagreementError.Error(), planted) {
			test.Fatalf("wrong owner: %v", printerShardDisagreementError)
		}
		test.Log(printerShardDisagreementError)
		caught++
	}
	if caught != 1 {
		test.Fatalf("%d shards caught planted disagreement, want 1", caught)
	}
}

func TestPrinterWhitespaceShardGrowth(test *testing.T) {
	test.Parallel()
	original := []printerCase{{id: whitespaceCaseId(0)}, {id: whitespaceCaseId(1)}}
	grown := append(append([]printerCase(nil), original...), printerCase{id: whitespaceCaseId(2)})
	before, after := whitespaceShards(original), whitespaceShards(grown)
	if len(before) != printerWhitespaceWitnessShards || len(after) != len(before) {
		test.Fatal("growth changed shard count")
	}
	for _, cases := range [][]printerCase{original, grown} {
		if printerShardUnionError := printerShardUnion(cases, whitespaceShards(cases)); printerShardUnionError != nil {
			test.Fatal(printerShardUnionError)
		}
	}
	for number, shard := range before {
		for _, item := range shard.cases {
			found := false
			for _, candidate := range after[number].cases {
				if candidate.id == item.id {
					found = true
				}
			}
			if !found {
				test.Fatalf("growth moved %s from shard-%03d", item.id, number)
			}
		}
	}
}
