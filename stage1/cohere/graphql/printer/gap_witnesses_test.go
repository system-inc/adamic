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

func TestPrinterConstructorGap(t *testing.T) {
	t.Parallel()
	path, _ := filepath.Abs("gaps/nestedConstructor.ts")
	result := onNode(t, path)
	if result.exitCode != 0 || string(result.stdout) != "ready\n" {
		t.Fatalf("Node: %+v", result)
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = lower.Lower(context.Background(), program)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "this escaping a constructor before every field is set") {
		t.Fatalf("gap changed; update GAPS.md and remove workaround: %v", err)
	}
	t.Log(err)
}

const testPrinterWhitespaceGapShards = 5

// A real process emits one planted acceptance where the oracle requires refusal;
// only the owning whitespace shard may report the disagreement.
func whitespaceCaseID(index int) string {
	return fmt.Sprintf("stage1/cohere/graphql/printer/gaps/whitespace-cases.json/case-%06d", index)
}

func whitespaceOwner(key string) int {
	hash := sha256.Sum256([]byte(key))
	return int(binary.BigEndian.Uint64(hash[:8]) % testPrinterWhitespaceGapShards)
}

func whitespaceShards(cases []printerCase) []printerShard {
	shards := make([]printerShard, testPrinterWhitespaceGapShards)
	for i := range shards {
		shards[i].mode = "defaults"
	}
	for _, item := range cases {
		owner := whitespaceOwner(item.id)
		shards[owner].cases = append(shards[owner].cases, item)
	}
	return shards
}

func TestPrinterWhitespacePlantedDisagreement(t *testing.T) {
	t.Parallel()
	var whole []printerCase
	for i := range 5 {
		whole = append(whole, printerCase{id: whitespaceCaseID(i), want: "error\trefused"})
	}
	planted := whitespaceCaseID(3)
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
		result := execute(t, nil, "node", "-e", "process.stdout.write("+string(encoded)+")")
		err := printerShardDisagreement(number, shard, result)
		if err == nil {
			continue
		}
		if number != whitespaceOwner(planted) || !strings.Contains(err.Error(), planted) {
			t.Fatalf("wrong owner: %v", err)
		}
		t.Log(err)
		caught++
	}
	if caught != 1 {
		t.Fatalf("%d shards caught planted disagreement, want 1", caught)
	}
}

func TestPrinterWhitespaceShardGrowth(t *testing.T) {
	t.Parallel()
	original := []printerCase{{id: whitespaceCaseID(0)}, {id: whitespaceCaseID(1)}}
	grown := append(append([]printerCase(nil), original...), printerCase{id: whitespaceCaseID(2)})
	before, after := whitespaceShards(original), whitespaceShards(grown)
	if len(before) != testPrinterWhitespaceGapShards || len(after) != len(before) {
		t.Fatal("growth changed shard count")
	}
	for _, cases := range [][]printerCase{original, grown} {
		if err := printerShardUnion(cases, whitespaceShards(cases)); err != nil {
			t.Fatal(err)
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
				t.Fatalf("growth moved %s from shard-%03d", item.id, number)
			}
		}
	}
}
