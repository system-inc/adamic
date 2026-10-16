package lint

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

func shippedLintInputs(t *testing.T, name string) buildcache.Inputs {
	t.Helper()
	profdata, err := native.Stage1ProfileDataTool()
	if err != nil {
		t.Fatal(err)
	}
	return buildcache.Inputs{Name: "stage1-shipped-lint-" + name, Files: append(compilerinputs.Lowering(), "stage1/cohere/lint", "stage1/typescript", "stage1/profiles/training.json"), Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool(profdata, "--version")}}
}
func shippedLintSource(t *testing.T) string {
	t.Helper()
	directory := buildcache.Product(t, shippedLintInputs(t, "source"), func(directory string) error {
		if _, err := registry.Generate("."); err != nil {
			return err
		}
		program, err := load.Load([]string{"main.ts"})
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Second)
		defer cancel()
		ir, err := lower.Lower(ctx, program)
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "main.c"), []byte(native.C(ir)), 0644)
	})
	data, err := os.ReadFile(filepath.Join(directory, "main.c"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
func shippedLintBinary(t *testing.T) string {
	t.Helper()
	directory := filepath.Join("profiles", runtime.GOOS+"-"+runtime.GOARCH)
	text, err := native.ReadStage1Profile(filepath.Join(directory, "profile.txt"))
	if os.IsNotExist(err) && os.Getenv("ADAMIC_REQUIRE_STAGE1_PROFILES") != "1" {
		t.Skip("profile is a subsequent target-specific piece")
	}
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var record native.Stage1ProfileManifest
	if err := json.Unmarshal(data, &record); err != nil {
		t.Fatal(err)
	}
	training, err := os.ReadFile("../../profiles/training.json")
	if err != nil {
		t.Fatal(err)
	}
	source := shippedLintSource(t)
	if err := native.CheckStage1ProfileManifest(source, text, fmt.Sprintf("%x", sha256.Sum256(training)), record); err != nil {
		t.Fatal(err)
	}
	inputs := shippedLintInputs(t, "binary")
	inputs.Flags = []string{source}
	product := buildcache.Product(t, inputs, func(output string) error {
		return native.Build(source, filepath.Join(output, "binary"), native.Options{Release: true, Split: true, Profile: filepath.Join(directory, "profile.txt")})
	})
	return filepath.Join(product, "binary")
}
func TestProduct_LintShippedSource(t *testing.T)  { t.Parallel(); shippedLintSource(t) }
func TestProduct_LintShippedProfile(t *testing.T) { t.Parallel(); t.Log(shippedLintBinary(t)) }
func TestLintShippedProfileAgreesWithNode(t *testing.T) {
	t.Parallel()
	binary := shippedLintBinary(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	row := compilationCase(t)
	// compilationCase is repository-relative; the driver reads it from the root.
	manifest := manifest(t, []string{filepath.Join("../../..", row)})
	got := execute(t, "", binary, "--manifest", manifest)
	want := execute(t, "", "node", "--disable-warning=ExperimentalWarning", "../../../oracle/node.mjs", "main.ts", "--manifest", manifest)
	oracle := execute(t, "", lintGoOracleProduct(t), "--manifest", manifest)
	if !bytes.Equal(got.output, oracle.output) {
		t.Fatal("shipping lint differs from Go")
	}
	if !bytes.Equal(got.output, want.output) {
		t.Fatal("shipping lint differs from Node", directory)
	}
}

func lintProfilePartBytes(part native.Stage1ProfileUnit, content []byte) error {
	if fmt.Sprintf("%x", sha256.Sum256(content)) != part.SHA256 {
		return fmt.Errorf("profile data part changed: %s", part.Name)
	}
	return nil
}
func TestLintProfilePartByteControl(t *testing.T) {
	t.Parallel()
	content := []byte("profile data")
	part := native.Stage1ProfileUnit{SHA256: fmt.Sprintf("%x", sha256.Sum256(content))}
	if err := lintProfilePartBytes(part, content); err != nil {
		t.Fatal(err)
	}
	content[0] ^= 1
	if err := lintProfilePartBytes(part, content); err == nil {
		t.Fatal("profile part byte mutant survived")
	}
}
func TestLintProfileDataParts(t *testing.T) {
	t.Parallel()
	directory := filepath.Join("profiles", runtime.GOOS+"-"+runtime.GOARCH)
	data, err := os.ReadFile(filepath.Join(directory, "parts.json"))
	if os.IsNotExist(err) {
		t.Skip("profile data pieces have not landed")
	}
	if err != nil {
		t.Fatal(err)
	}
	var parts []native.Stage1ProfileUnit
	if err := json.Unmarshal(data, &parts); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(directory, "profile-*.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != len(files) || len(parts) == 0 {
		t.Fatal("profile part inventory differs")
	}
	tool, err := native.Stage1ProfileDataTool()
	if err != nil {
		t.Fatal(err)
	}
	for i, part := range parts {
		if part.Name != fmt.Sprintf("profile-%02d.txt", i) {
			t.Fatal("invalid part inventory", part.Name)
		}
		content, err := os.ReadFile(filepath.Join(directory, part.Name))
		if err != nil {
			t.Fatal(err)
		}
		if err := lintProfilePartBytes(part, content); err != nil {
			t.Fatal(err)
		}
		command := toolchain.Command(tool, "merge", filepath.Join(directory, part.Name), "-o", filepath.Join(t.TempDir(), "profile.profdata"))
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("invalid profile records: %v %s", err, output)
		}
	}
}
