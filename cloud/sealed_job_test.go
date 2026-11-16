package cloud

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// sealed_products gives its go test no clock but the job's own ceiling: go's default 10 minutes killed internal/oracle's
// products before they settled on Lunatic's 8 cores (Library's J-1xphvcc, routed by Runtime on #jobs), so a slow box
// read as the head's red. The test sources the script in a temporary tree beside a stand-in go that writes down the
// arguments it was given, and holds that the products' go test asked `-timeout 0`.
//
// Mutant: the `-timeout 0` taken out of sealed_products' go test (the stand-in's arguments name no timeout, red here).
func TestSealedProductsGivesGoNoClockButTheJobs(t *testing.T) {
	t.Parallel()
	script, err := filepath.Abs("sealed-job.sh")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	tree, bin := filepath.Join(directory, "tree"), filepath.Join(directory, "bin")
	for _, made := range []string{filepath.Join(tree, "cloud", "fast-gate"), filepath.Join(tree, "internal", "oracle"), bin} {
		if err := os.MkdirAll(made, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(tree, "cloud", "fast-gate", "compiler-dependencies.json"), []byte(`{"packages":{}}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	said := filepath.Join(directory, "go-arguments")
	stand := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$GO_ARGUMENTS\"\n"
	if err := os.WriteFile(filepath.Join(bin, "go"), []byte(stand), 0o755); err != nil {
		t.Fatal(err)
	}
	// The TypeScript copy is the image's road, not this test's: it stands in as done.
	sourced := `. "$SEALED_JOB_SCRIPT" || exit 1
sealed_typescript_source() { :; }
sealed_products ./internal/oracle
`
	command := toolchain.Command("bash", "-c", sourced)
	if command.Err != nil {
		t.Fatal(command.Err)
	}
	command.Dir = tree
	environment := command.Env
	if environment == nil {
		environment = os.Environ()
	}
	path := os.Getenv("PATH")
	kept := make([]string, 0, len(environment)+4)
	for _, variable := range environment {
		name, value, _ := strings.Cut(variable, "=")
		switch name {
		case "PATH":
			path = value
		case "TMPDIR", "ADAMIC_BUILD_CACHE_DIR", "LOOM_JOB_SCRATCH":
		default:
			kept = append(kept, variable)
		}
	}
	command.Env = append(kept, "PATH="+bin+":"+path, "TMPDIR="+filepath.Join(directory, "tmp"), "GO_ARGUMENTS="+said,
		"SEALED_JOB_SCRIPT="+script)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("sealed_products: %v\n%s", err, output)
	}
	data, err := os.ReadFile(said)
	if err != nil {
		t.Fatalf("the stand-in go was never run: %v", err)
	}
	arguments := strings.Split(strings.TrimSpace(string(data)), "\n")
	for index, argument := range arguments {
		if argument == "-timeout" && index+1 < len(arguments) && arguments[index+1] == "0" {
			return
		}
		if argument == "-timeout=0" {
			return
		}
	}
	t.Fatalf("sealed_products' go test named no -timeout 0, so go's own 10 minutes would end a slow settle: %q", arguments)
}
