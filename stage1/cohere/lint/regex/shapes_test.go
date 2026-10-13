package regex

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"
)

// shapeFixtureShards is how many tests share the per-shape fixtures. The gate compares a row in the shard its ID
// hashes to (gate.go's owner), so every shard still checks the whole table's coverage and compares only its own rows.
const shapeFixtureShards = 16

func shapeFixtureRows(t *testing.T) []struct{ ID string } {
	t.Helper()
	data, err := os.ReadFile("table.json")
	if err != nil {
		t.Fatal(err)
	}
	var rows []struct{ ID string }
	if err := json.Unmarshal(data, &rows); err != nil {
		t.Fatal(err)
	}
	return rows
}

// invokeShapesGate runs the gate binary from the repository, its output and error together.
func invokeShapesGate(t *testing.T, binary string, arguments ...string) ([]byte, error) {
	t.Helper()
	repository, err := filepath.Abs("../../../..")
	if err != nil {
		t.Fatal(err)
	}
	command := exec.Command(binary, arguments...)
	command.Dir = repository
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	err = command.Run()
	return output.Bytes(), err
}

// shapeFixtureOwned is the rows shard index owns, as the gate's dry run names them after their Go side passes.
func shapeFixtureOwned(t *testing.T, binary string, rows, index int) []string {
	t.Helper()
	shard := fmt.Sprintf("%d/%d", index, shapeFixtureShards)
	output, err := invokeShapesGate(t, binary, "-dry", "-shard", shard)
	if err != nil {
		t.Fatalf("shard %s dry run: %v\n%s", shard, err, output)
	}
	var owned []string
	for _, line := range strings.Split(string(output), "\n") {
		if id, ok := strings.CutPrefix(line, "owns "); ok {
			owned = append(owned, id)
		}
	}
	if !strings.Contains(string(output), fmt.Sprintf("DRY fixtures=%d shard=%s ran=%d ", rows, shard, len(owned))) {
		t.Fatalf("shard %s dry coverage count missing: %s", shard, output)
	}
	return owned
}

func shapeFixtureShard(t *testing.T, index int) {
	t.Helper()
	rows := shapeFixtureRows(t)
	binary := shapesGate(t)
	shard := fmt.Sprintf("%d/%d", index, shapeFixtureShards)
	owned := shapeFixtureOwned(t, binary, len(rows), index)
	arguments := []string{"-shard", shard}
	if os.Getenv("ADAMIC_REGEX_TRANSLATION_MUTANT") == "1" {
		arguments = append(arguments, "-mutant")
	}
	output, err := invokeShapesGate(t, binary, arguments...)
	if err != nil {
		t.Fatalf("per-shape fixture comparison failed: %v\n%s", err, output)
	}
	if !strings.Contains(string(output), fmt.Sprintf("PASS fixtures=%d shard=%s ran=%d ", len(rows), shard, len(owned))) {
		t.Fatalf("fixture coverage count missing: %s", output)
	}
	t.Log(strings.TrimSpace(string(output)))
}

// TestShapeFixturesUnion holds the shards to the table: the declared shard tests are exactly shapeFixtureShards, every
// row is owned by exactly one shard, and a failure planted in one row is caught by its owner alone.
func TestShapeFixturesUnion(t *testing.T) {
	t.Parallel()
	tree, err := parser.ParseFile(token.NewFileSet(), "shapes_test.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	declared := 0
	for _, declaration := range tree.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok && strings.HasPrefix(function.Name.Name, "TestShapeFixtures_") {
			if function.Name.Name != fmt.Sprintf("TestShapeFixtures_%03d", declared) {
				t.Fatalf("noncontiguous shard declaration %s", function.Name.Name)
			}
			declared++
		}
	}
	if declared != shapeFixtureShards {
		t.Fatalf("shard declarations: got %d want %d", declared, shapeFixtureShards)
	}
	rows := shapeFixtureRows(t)
	binary := shapesGate(t)
	owners := map[string][]int{}
	ran := 0
	for index := 0; index < shapeFixtureShards; index++ {
		for _, id := range shapeFixtureOwned(t, binary, len(rows), index) {
			owners[id] = append(owners[id], index)
			ran++
		}
	}
	for _, row := range rows {
		if len(owners[row.ID]) != 1 {
			t.Fatalf("%s owned by shards %v", row.ID, owners[row.ID])
		}
	}
	if ran != len(rows) || len(owners) != len(rows) {
		t.Fatalf("shards ran %d rows, %d distinct, of %d", ran, len(owners), len(rows))
	}
	t.Logf("union: %d rows, each owned by exactly one of %d shards", len(rows), shapeFixtureShards)
	planted := rows[len(rows)-1].ID
	caught := 0
	for index := 0; index < shapeFixtureShards; index++ {
		output, err := invokeShapesGate(t, binary, "-dry", "-plant", planted, "-shard", fmt.Sprintf("%d/%d", index, shapeFixtureShards))
		if err == nil {
			continue
		}
		if index != owners[planted][0] || !strings.Contains(string(output), planted+": Go oracle drift") {
			t.Fatalf("planted failure in %s failed shard %d outside its comparison: %v\n%s", planted, index, err, output)
		}
		caught++
	}
	if caught != 1 {
		t.Fatalf("planted failure in %s caught by %d shards", planted, caught)
	}
	t.Logf("planted failure in %s caught by its owner, shard %d, alone", planted, owners[planted][0])
}

// TestShapeFixtureControls proves the comparison can fail. Both controls fire at the table's first row, so they run
// against the shard that owns it.
func TestShapeFixtureControls(t *testing.T) {
	t.Parallel()
	rows := shapeFixtureRows(t)
	binary := shapesGate(t)
	first := -1
	for index := 0; index < shapeFixtureShards && first < 0; index++ {
		if slices.Contains(shapeFixtureOwned(t, binary, len(rows), index), rows[0].ID) {
			first = index
		}
	}
	if first < 0 {
		t.Fatalf("no shard owns %s", rows[0].ID)
	}
	shard := fmt.Sprintf("%d/%d", first, shapeFixtureShards)
	t.Run("one-character-translation-mutant", func(t *testing.T) {
		output, err := invokeShapesGate(t, binary, "-mutant", "-shard", shard)
		if err == nil || !strings.Contains(string(output), rows[0].ID+": source Node fixture mismatch") {
			t.Fatalf("mutant escaped or failed outside comparison: %v\n%s", err, output)
		}
		t.Logf("one-character mutant caught by fixture comparison: %s", rows[0].ID)
	})
	t.Run("forced-native-transition", func(t *testing.T) {
		// Whether every row awaits the runtime compiler is a whole-table fact, so it is counted over the whole table.
		census, err := invokeShapesGate(t, binary, "-pending")
		if err != nil || !strings.Contains(string(census), fmt.Sprintf("PENDING fixtures=%d shard=0/1 ran=%d ", len(rows), len(rows))) {
			t.Fatalf("runtime lowering census failed: %v\n%s", err, census)
		}
		forced, err := invokeShapesGate(t, binary, "-force-native", "-shard", shard)
		if strings.Contains(string(census), fmt.Sprintf("awaits codex/regex-runtime-compiler=%d\n", len(rows))) {
			if err == nil || !strings.Contains(string(forced), "awaits codex/regex-runtime-compiler: forced native requirement caught refusal") {
				t.Fatalf("native refusal control failed: %v\n%s", err, forced)
			}
			t.Log("awaits codex/regex-runtime-compiler: forced native requirement caught named refusal")
		} else if err != nil {
			t.Fatalf("native is now required: %v\n%s", err, forced)
		}
	})
}

func TestShapeFixtures_000(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 0)
}

func TestShapeFixtures_001(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 1)
}

func TestShapeFixtures_002(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 2)
}

func TestShapeFixtures_003(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 3)
}

func TestShapeFixtures_004(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 4)
}

func TestShapeFixtures_005(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 5)
}

func TestShapeFixtures_006(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 6)
}

func TestShapeFixtures_007(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 7)
}

func TestShapeFixtures_008(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 8)
}

func TestShapeFixtures_009(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 9)
}

func TestShapeFixtures_010(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 10)
}

func TestShapeFixtures_011(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 11)
}

func TestShapeFixtures_012(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 12)
}

func TestShapeFixtures_013(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 13)
}

func TestShapeFixtures_014(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 14)
}

func TestShapeFixtures_015(t *testing.T) {
	t.Parallel()
	shapeFixtureShard(t, 15)
}

var shapesGateOnce sync.Once
var shapesGateBinary string

// shapesGate is testdata/shapes/gate.go, a product built ahead (#5qykzj5), looked up once per process, since a lookup
// values the product's whole read set and every shard asks for it.
func shapesGate(t *testing.T) string {
	t.Helper()
	shapesGateOnce.Do(func() {
		shapesGateBinary = buildcache.GoBuild(t, "shapes-gate", "./stage1/cohere/lint/regex/testdata/shapes/gate.go", nil)
	})
	if shapesGateBinary == "" {
		t.Fatal("shapes gate product lookup did not complete")
	}
	return shapesGateBinary
}
