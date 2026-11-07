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

func shapeFixtureRows(test *testing.T) []struct{ ID string } {
	test.Helper()
	data, readError := os.ReadFile("table.json")
	if readError != nil {
		test.Fatal(readError)
	}
	var rows []struct{ ID string }
	if decodeError := json.Unmarshal(data, &rows); decodeError != nil {
		test.Fatal(decodeError)
	}
	return rows
}

// invokeShapesGate runs the gate binary from the repository, its output and error together.
func invokeShapesGate(test *testing.T, binary string, arguments ...string) ([]byte, error) {
	test.Helper()
	repository, absolutePathError := filepath.Abs("../../../..")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	command := exec.Command(binary, arguments...)
	command.Dir = repository
	var output bytes.Buffer
	command.Stdout = &output
	command.Stderr = &output
	runError := command.Run()
	return output.Bytes(), runError
}

// shapeFixtureOwned is the rows shard index owns, as the gate's dry run names them after their Go side passes.
func shapeFixtureOwned(test *testing.T, binary string, rows, index int) []string {
	test.Helper()
	shard := fmt.Sprintf("%d/%d", index, shapeFixtureShards)
	output, invokeShapesGateError := invokeShapesGate(test, binary, "-dry", "-shard", shard)
	if invokeShapesGateError != nil {
		test.Fatalf("shard %s dry run: %v\n%s", shard, invokeShapesGateError, output)
	}
	var owned []string
	for _, line := range strings.Split(string(output), "\n") {
		if id, ok := strings.CutPrefix(line, "owns "); ok {
			owned = append(owned, id)
		}
	}
	if !strings.Contains(string(output), fmt.Sprintf("DRY fixtures=%d shard=%s ran=%d ", rows, shard, len(owned))) {
		test.Fatalf("shard %s dry coverage count missing: %s", shard, output)
	}
	return owned
}

func shapeFixtureShard(test *testing.T, index int) {
	test.Helper()
	rows := shapeFixtureRows(test)
	binary := shapesGate(test)
	shard := fmt.Sprintf("%d/%d", index, shapeFixtureShards)
	owned := shapeFixtureOwned(test, binary, len(rows), index)
	arguments := []string{"-shard", shard}
	if os.Getenv("ADAMIC_REGEX_TRANSLATION_MUTANT") == "1" {
		arguments = append(arguments, "-mutant")
	}
	output, invokeShapesGateError := invokeShapesGate(test, binary, arguments...)
	if invokeShapesGateError != nil {
		test.Fatalf("per-shape fixture comparison failed: %v\n%s", invokeShapesGateError, output)
	}
	if !strings.Contains(
		string(output),
		fmt.Sprintf("PASS fixtures=%d shard=%s ran=%d ", len(rows), shard, len(owned)),
	) {
		test.Fatalf("fixture coverage count missing: %s", output)
	}
	test.Log(strings.TrimSpace(string(output)))
}

// TestShapeFixturesUnion holds the shards to the table: the declared shard tests are exactly shapeFixtureShards, every
// row is owned by exactly one shard, and a failure planted in one row is caught by its owner alone.
func TestShapeFixturesUnion(test *testing.T) {
	test.Parallel()
	tree, parseFileError := parser.ParseFile(token.NewFileSet(), "shapes_test.go", nil, 0)
	if parseFileError != nil {
		test.Fatal(parseFileError)
	}
	declared := 0
	for _, declaration := range tree.Decls {
		if function, ok := declaration.(*ast.FuncDecl); ok &&
			strings.HasPrefix(function.Name.Name, "TestShapeFixtures_") {
			if function.Name.Name != fmt.Sprintf("TestShapeFixtures_%03d", declared) {
				test.Fatalf("noncontiguous shard declaration %s", function.Name.Name)
			}
			declared++
		}
	}
	if declared != shapeFixtureShards {
		test.Fatalf("shard declarations: got %d want %d", declared, shapeFixtureShards)
	}
	rows := shapeFixtureRows(test)
	binary := shapesGate(test)
	owners := map[string][]int{}
	ran := 0
	for index := 0; index < shapeFixtureShards; index++ {
		for _, id := range shapeFixtureOwned(test, binary, len(rows), index) {
			owners[id] = append(owners[id], index)
			ran++
		}
	}
	for _, row := range rows {
		if len(owners[row.ID]) != 1 {
			test.Fatalf("%s owned by shards %v", row.ID, owners[row.ID])
		}
	}
	if ran != len(rows) || len(owners) != len(rows) {
		test.Fatalf("shards ran %d rows, %d distinct, of %d", ran, len(owners), len(rows))
	}
	test.Logf("union: %d rows, each owned by exactly one of %d shards", len(rows), shapeFixtureShards)
	planted := rows[len(rows)-1].ID
	caught := 0
	for index := 0; index < shapeFixtureShards; index++ {
		output, invokeShapesGateError := invokeShapesGate(
			test,
			binary,
			"-dry",
			"-plant",
			planted,
			"-shard",
			fmt.Sprintf("%d/%d", index, shapeFixtureShards),
		)
		if invokeShapesGateError == nil {
			continue
		}
		if index != owners[planted][0] || !strings.Contains(string(output), planted+": Go oracle drift") {
			test.Fatalf(
				"planted failure in %s failed shard %d outside its comparison: %v\n%s",
				planted,
				index,
				invokeShapesGateError,
				output,
			)
		}
		caught++
	}
	if caught != 1 {
		test.Fatalf("planted failure in %s caught by %d shards", planted, caught)
	}
	test.Logf("planted failure in %s caught by its owner, shard %d, alone", planted, owners[planted][0])
}

// TestShapeFixtureControls proves the comparison can fail. Both controls fire at the table's first row, so they run
// against the shard that owns it.
func TestShapeFixtureControls(test *testing.T) {
	test.Parallel()
	rows := shapeFixtureRows(test)
	binary := shapesGate(test)
	first := -1
	for index := 0; index < shapeFixtureShards && first < 0; index++ {
		if slices.Contains(shapeFixtureOwned(test, binary, len(rows), index), rows[0].ID) {
			first = index
		}
	}
	if first < 0 {
		test.Fatalf("no shard owns %s", rows[0].ID)
	}
	shard := fmt.Sprintf("%d/%d", first, shapeFixtureShards)
	test.Run("one-character-translation-mutant", func(test *testing.T) {
		output, invokeShapesGateError := invokeShapesGate(test, binary, "-mutant", "-shard", shard)
		if invokeShapesGateError == nil ||
			!strings.Contains(string(output), rows[0].ID+": source Node fixture mismatch") {
			test.Fatalf("mutant escaped or failed outside comparison: %v\n%s", invokeShapesGateError, output)
		}
		test.Logf("one-character mutant caught by fixture comparison: %s", rows[0].ID)
	})
	test.Run("forced-native-transition", func(test *testing.T) {
		// Whether every row awaits the runtime compiler is a whole-table fact, so it is counted over the whole table.
		census, invokeShapesGateError := invokeShapesGate(test, binary, "-pending")
		if invokeShapesGateError != nil || !strings.Contains(
			string(census),
			fmt.Sprintf("PENDING fixtures=%d shard=0/1 ran=%d ", len(rows), len(rows)),
		) {
			test.Fatalf("runtime lowering census failed: %v\n%s", invokeShapesGateError, census)
		}
		forced, invokeShapesGateError := invokeShapesGate(test, binary, "-force-native", "-shard", shard)
		if strings.Contains(string(census), fmt.Sprintf("awaits codex/regex-runtime-compiler=%d\n", len(rows))) {
			if invokeShapesGateError == nil || !strings.Contains(
				string(forced),
				"awaits codex/regex-runtime-compiler: forced native requirement caught refusal",
			) {
				test.Fatalf("native refusal control failed: %v\n%s", invokeShapesGateError, forced)
			}
			test.Log("awaits codex/regex-runtime-compiler: forced native requirement caught named refusal")
		} else if invokeShapesGateError != nil {
			test.Fatalf("native is now required: %v\n%s", invokeShapesGateError, forced)
		}
	})
}

func TestShapeFixtures_000(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 0)
}

func TestShapeFixtures_001(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 1)
}

func TestShapeFixtures_002(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 2)
}

func TestShapeFixtures_003(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 3)
}

func TestShapeFixtures_004(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 4)
}

func TestShapeFixtures_005(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 5)
}

func TestShapeFixtures_006(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 6)
}

func TestShapeFixtures_007(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 7)
}

func TestShapeFixtures_008(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 8)
}

func TestShapeFixtures_009(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 9)
}

func TestShapeFixtures_010(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 10)
}

func TestShapeFixtures_011(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 11)
}

func TestShapeFixtures_012(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 12)
}

func TestShapeFixtures_013(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 13)
}

func TestShapeFixtures_014(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 14)
}

func TestShapeFixtures_015(test *testing.T) {
	test.Parallel()
	shapeFixtureShard(test, 15)
}

var shapesGateOnce sync.Once
var shapesGateBinary string

// shapesGate is testdata/shapes/gate.go, a product built ahead (#5qykzj5), looked up once per process, since a lookup
// values the product's whole read set and every shard asks for it.
func shapesGate(test *testing.T) string {
	test.Helper()
	shapesGateOnce.Do(func() {
		shapesGateBinary = buildcache.GoBuild(
			test,
			"shapes-gate",
			"./stage1/cohere/lint/regex/testdata/shapes/gate.go",
			nil,
		)
	})
	if shapesGateBinary == "" {
		test.Fatal("shapes gate product lookup did not complete")
	}
	return shapesGateBinary
}
