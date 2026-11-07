package lane

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// A public-any exemption remains justified only while its pinned API sample still reads the unchecked value.
// This test runs git to read the pinned object, never Go or a compiler. Entries may be removed, never added here.
func TestJSONValues52PinnedAnySampleReadsRemain(test *testing.T) {
	test.Parallel()
	root := repositoryRoot(test)
	data, err := os.ReadFile(filepath.Join(root, "stage3/adapt/52-json-values/public-any-pinned.json"))
	if err != nil {
		test.Fatal(err)
	}
	var rows []struct {
		Path     []string `json:"path"`
		Kind     string   `json:"kind"`
		Baseline string   `json:"baseline"`
		Read     string   `json:"read"`
	}
	if err := json.Unmarshal(data, &rows); err != nil {
		test.Fatal(err)
	}
	if len(rows) > 1 {
		test.Fatal("the reviewed public-any exemption list grew")
	}
	source := os.Getenv("ADAMIC_TYPESCRIPT_SOURCE")
	if source == "" {
		test.Fatal("ADAMIC_TYPESCRIPT_SOURCE must name the pinned upstream git input")
	}
	command := toolchain.GitCommandIn(
		source,
		"show",
		typescriptPin(test, root)+":tests/cases/compiler/APISample_parseConfig.ts",
	)
	sample, err := command.Output()
	if err != nil {
		test.Fatalf("read pinned API sample: %v", err)
	}
	seen := map[string]bool{}
	for _, row := range rows {
		if len(row.Path) != 2 || row.Path[1] != "config" ||
			row.Path[0] != "parseConfigFileTextToJson" ||
			row.Kind != "PropertySignature" || row.Baseline != "APISample_parseConfig.errors.txt" ||
			row.Read != `config.config["compilerOptions"]` {
			test.Fatalf("unreviewed public-any exemption: %+v", row)
		}
		name := strings.Join(row.Path, ".")
		if seen[name] {
			test.Fatalf("duplicate exemption %s", name)
		}
		seen[name] = true
		if !jsonValues52SampleContainsOwnPin(string(sample), row.Path[0], row.Read) {
			test.Errorf(
				"%s is no longer pinned by %s: remove its exemption; missing own call ts.%s( or read %q",
				name,
				row.Baseline,
				row.Path[0],
				row.Read,
			)
		}
	}
}

func jsonValues52SampleContainsOwnPin(sample, function, read string) bool {
	return strings.Contains(sample, "ts."+function+"(") && strings.Contains(sample, read)
}

func TestJSONValues52ExemptionsRequireTheirOwnCall(test *testing.T) {
	test.Parallel()
	const read = `config.config["compilerOptions"]`
	const sample = `const config = ts.parseConfigFileTextToJson("config.json", text); config.config["compilerOptions"];`
	if !jsonValues52SampleContainsOwnPin(sample, "parseConfigFileTextToJson", read) {
		test.Fatal("the parser call and its read must pin the parser")
	}
	if jsonValues52SampleContainsOwnPin(sample, "readConfigFile", read) {
		test.Fatal("another function's call and read cannot pin readConfigFile")
	}
	if jsonValues52SampleContainsOwnPin(read, "parseConfigFileTextToJson", read) {
		test.Fatal("the read alone cannot pin a function")
	}
	if jsonValues52SampleContainsOwnPin(
		`ts.parseConfigFileTextToJson("config.json", text)`,
		"parseConfigFileTextToJson",
		read,
	) {
		test.Fatal("the call alone cannot pin the unchecked read")
	}
}
