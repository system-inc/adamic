package scanall

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/toolchain"
)

// The unoverlaid lower package supplies the complete transitive Go source closure,
// including the originals that generation copies. Generation is recipe work, never
// lookup work: no Go build or go run is needed to fetch a settled product.
func scanAllProduct(t *testing.T, instrumented bool) string {
	t.Helper()
	inputs, err := buildcache.GoInputs("scan-all-probe", "./internal/lower", buildcache.ReproducibleGoFlags(), nil)
	if err != nil {
		t.Fatal(err)
	}
	inputs.Files = append(inputs.Files,
		"stage3/census/scan-all/products_test.go",
		"stage3/census/scan-all/make_overlay.py",
		"stage3/census/scan-all/main.go.txt",
		"stage3/census/scan-all/scout.go.txt",
		"stage3/census/scan-all/production.go.txt",
		"stage3/census/latent/refusalrewrite")
	inputs.Name = "scan-all-production"
	if instrumented {
		inputs.Name = "scan-all-instrumented"
	}
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("python3", "--version"))
	directory := buildcache.Product(t, inputs, func(directory string) error {
		root, err := filepath.Abs("../../..")
		if err != nil {
			return err
		}
		scratch, err := os.MkdirTemp("", "scan-all-build-")
		if err != nil {
			return err
		}
		defer os.RemoveAll(scratch)
		run := func(name string, arguments ...string) error {
			command := toolchain.CommandIn(root, name, arguments...)
			output, err := command.CombinedOutput()
			if err != nil {
				return fmt.Errorf("%s %v: %w\n%s", name, arguments, err, output)
			}
			return nil
		}
		overlay := filepath.Join(scratch, "overlay.json")
		main := "stage3/census/scan-all/production/main.go"
		if instrumented {
			if err := run("python3", "stage3/census/scan-all/make_overlay.py", scratch); err != nil {
				return err
			}
			main = "stage3/census/scan-all/tool/main.go"
		} else {
			content, err := json.Marshal(map[string]any{"Replace": map[string]string{
				filepath.Join(root, main): filepath.Join(root, "stage3/census/scan-all/production.go.txt"),
			}})
			if err != nil {
				return err
			}
			if err := os.WriteFile(overlay, content, 0600); err != nil {
				return err
			}
		}
		arguments := append([]string{"build"}, buildcache.ReproducibleGoFlags()...)
		arguments = append(arguments, "-overlay", overlay, "-o", filepath.Join(directory, "probe"), "./"+main)
		return run("go", arguments...)
	})
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "probe"})
	t.Logf("%s probe: %s", inputs.Name, filepath.Join(directory, "probe"))
	return filepath.Join(directory, "probe")
}

func TestProduct_ScanAllInstrumentedProbe(t *testing.T) { t.Parallel(); scanAllProduct(t, true) }
func TestProduct_ScanAllProductionProbe(t *testing.T)   { t.Parallel(); scanAllProduct(t, false) }
