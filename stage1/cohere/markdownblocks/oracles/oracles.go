// Package oracles owns markdownblocks' compiler-free Go reference products.
// Paths resolve from stage1/cohere/markdownblocks in both test binaries.
package oracles

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

// LayoutGoProducts holds the two Go reference programs for a layout family.
type LayoutGoProducts struct{ List, Document string }

var oracleProducer bool

// EnableProductProduction enables recipes in the owner test binary's TestMain.
func EnableProductProduction() { oracleProducer = true }
func oracleProduct(t *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	t.Helper()
	inputs.Name += " oracle-owner-v1"
	return buildcache.Product(t, inputs, func(directory string) error {
		if !oracleProducer {
			return fmt.Errorf("oracle %s is missing: settle the markdownblocks/oracles owners first", inputs.Name)
		}
		return build(directory)
	})
}

// LeafCompositionGoProduct finds one leaf-composition Go reference program.
func LeafCompositionGoProduct(t *testing.T, root, name, driverName string) string {
	t.Helper()
	cohere := filepath.Join(root, "cohere")
	bridge, err := filepath.Abs("testdata/list_bridge.go")
	if err != nil {
		t.Fatal(err)
	}
	driver, err := filepath.Abs(filepath.Join("testdata", driverName))
	if err != nil {
		t.Fatal(err)
	}
	main := filepath.Join(cohere, "cmd", name, "main.go")
	inputs := buildcache.Inputs{Name: "markdown-leaf-go-" + name, Files: []string{
		"stage1/cohere/markdownblocks/testdata/" + driverName,
		"stage1/cohere/markdownblocks/testdata/list_bridge.go",
		"cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim",
		"cohere/go.mod", "cohere/go.sum", "cohere/go.work", "cohere/go.work.sum", "go.mod", "go.work",
	}, Flags: []string{"overlay=" + driverName, "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")}, Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH}}
	product := oracleProduct(t, inputs, func(out string) error {
		replacements := map[string]string{main: driver}
		if name == "adamic_markdown_lists" {
			replacements[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replacements})
		if err != nil {
			return err
		}
		overlayPath := filepath.Join(out, "overlay.json")
		if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
			return err
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		ctx, cancel := context.WithCancel(context.Background())
		t.Cleanup(cancel)
		command := toolchain.CommandContextIn(ctx, cohere, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+overlayPath, "-o", filepath.Join(out, "oracle"), main)...)
		command.WaitDelay = time.Second
		if output, err := childguard.CombinedOutput(command, childguard.Options{Unbounded: true}); err != nil {
			return fmt.Errorf("Go bridge: %v\n%s", err, output)
		}
		return nil
	})
	return filepath.Join(product, "oracle")
}

// QuoteLayoutGoProducts finds the requested quote-layout Go reference programs.
func QuoteLayoutGoProducts(t *testing.T, root, target string) LayoutGoProducts {
	t.Helper()
	var products LayoutGoProducts
	// Cache the unchanged Go overlay builds as products.
	cohere := filepath.Join(root, "cohere")
	for _, mode := range []struct{ name, driver, command string }{
		{"lists", "list_go.go", "adamic_markdown_lists"},
		{"layout", "document_go.go", "adamic_markdown_doclayout"},
	} {
		if target != "" && target != mode.name {
			continue
		}
		driver, err := filepath.Abs(filepath.Join("testdata", mode.driver))
		if err != nil {
			t.Fatal(err)
		}
		mainPath := filepath.Join(cohere, "cmd", mode.command, "main.go")
		replace := map[string]string{mainPath: driver}
		if mode.name == "lists" {
			bridge, err := filepath.Abs("testdata/list_bridge.go")
			if err != nil {
				t.Fatal(err)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			t.Fatal(err)
		}
		goInputs := buildcache.Inputs{
			Name:      "markdown-quote-layout-go-" + mode.name,
			Files:     []string{"cohere", "go.mod", "go.work", "stage1/cohere/markdownblocks/testdata/list_go.go", "stage1/cohere/markdownblocks/testdata/list_bridge.go", "stage1/cohere/markdownblocks/testdata/document_go.go"},
			Flags:     []string{"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "), "overlay=" + mode.name, "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH")},
			Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version")},
		}
		dir := oracleProduct(t, goInputs, func(dir string) error {
			overlayPath := filepath.Join(dir, mode.name+".json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				return err
			}
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(overlayPath)
			command := toolchain.CommandContextIn(t.Context(), cohere, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+overlayPath, "-o", filepath.Join(dir, mode.name), mainPath)...)
			command.WaitDelay = time.Second
			if output, err := childguard.CombinedOutput(command, childguard.Options{}); err != nil {
				return fmt.Errorf("Go %s: %w\n%s", mode.name, err, output)
			}
			return nil
		})
		binary := filepath.Join(dir, mode.name)
		if mode.name == "lists" {
			products.List = binary
		} else {
			products.Document = binary
		}
	}

	return products
}

// StructureLayoutGoProducts finds the requested structure-layout Go reference programs.
func StructureLayoutGoProducts(t *testing.T, root, target string) LayoutGoProducts {
	t.Helper()
	var products LayoutGoProducts
	cohere := filepath.Join(root, "cohere")
	for _, item := range []struct {
		command, driver string
		bridge          bool
	}{{"adamic_markdown_lists", "list_go.go", true}, {"adamic_markdown_doclayout", "document_go.go", false}} {
		if target != "" && target != item.command {
			continue
		}
		driver, err := filepath.Abs(filepath.Join("testdata", item.driver))
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(cohere, "cmd", item.command, "main.go")
		replace := map[string]string{main: driver}
		if item.bridge {
			bridge, err := filepath.Abs("testdata/list_bridge.go")
			if err != nil {
				t.Fatal(err)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			t.Fatal(err)
		}
		// Keep overlay files inside the content-addressed product. Their source
		// paths are stable, so warm setup fetches the Go oracle instead of
		// charging whichever shard first enters sync.Once for a new build.
		goFiles := []string{"cohere/internal", "cohere/go.mod", "cohere/go.sum", "cohere/go.work", "cohere/go.work.sum", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/mutation_aliasing", "cohere/static_single_assignment", "go.mod", "go.work", "stage1/cohere/markdownblocks/testdata/" + item.driver}
		if item.bridge {
			goFiles = append(goFiles, "stage1/cohere/markdownblocks/testdata/list_bridge.go")
		}
		goDir := oracleProduct(t, buildcache.Inputs{
			Name: "markdown-structure-go-" + item.command, Files: goFiles,
			Flags:     []string{"overlay=" + string(overlay), "main=" + main, "GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH, "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")},
			Toolchain: []string{buildcache.Tool("go", "version")},
		}, func(dir string) error {
			overlayPath := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				return err
			}
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(overlayPath)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			command := toolchain.CommandContextIn(ctx, cohere, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+overlayPath, "-o", filepath.Join(dir, "oracle"), main)...)
			command.WaitDelay = 5 * time.Second
			if output, err := childguard.CombinedOutput(command, childguard.Options{}); err != nil {
				return fmt.Errorf("Go bridge: %w\n%s", err, output)
			}
			return nil
		})
		binary := filepath.Join(goDir, "oracle")
		if item.bridge {
			products.List = binary
		} else {
			products.Document = binary
		}
	}

	return products
}

// TableLayoutGoProducts finds the requested table-layout Go reference programs.
func TableLayoutGoProducts(t *testing.T, root, target string) LayoutGoProducts {
	t.Helper()
	var products LayoutGoProducts
	cohere := filepath.Join(root, "cohere")
	for _, mode := range []string{"list", "document"} {
		if target != "" && target != mode {
			continue
		}
		driver := "list_go.go"
		command := "adamic_markdown_lists"
		if mode == "document" {
			driver = "document_go.go"
			command = "adamic_markdown_doclayout"
		}
		driverPath, err := filepath.Abs(filepath.Join("testdata", driver))
		if err != nil {
			t.Fatal(err)
		}
		main := filepath.Join(cohere, "cmd", command, "main.go")
		replacements := map[string]string{main: driverPath}
		if mode == "list" {
			bridge, err := filepath.Abs("testdata/list_bridge.go")
			if err != nil {
				t.Fatal(err)
			}
			replacements[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replacements})
		if err != nil {
			t.Fatal(err)
		}

		goInputs := buildcache.Inputs{Name: "markdown table Go " + mode, Files: []string{"cohere", "stage1/cohere/markdownblocks/testdata/" + driver}, Flags: []string{string(overlay), "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")}, Toolchain: []string{buildcache.Tool("go", "version")}}
		if mode == "list" {
			goInputs.Files = append(goInputs.Files, "stage1/cohere/markdownblocks/testdata/list_bridge.go")
		}
		product := oracleProduct(t, goInputs, func(product string) error {
			overlayPath := filepath.Join(product, mode+".json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				return err
			}
			defer os.Remove(overlayPath)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			build := toolchain.CommandContextIn(ctx, cohere, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+overlayPath, "-o", filepath.Join(product, "oracle"), main)...)
			build.WaitDelay = time.Second
			if output, err := childguard.CombinedOutput(build, childguard.Options{Unbounded: true}); err != nil {
				return fmt.Errorf("Go %s: %w\n%s", mode, err, output)
			}
			return nil
		})
		binary := filepath.Join(product, "oracle")
		if mode == "list" {
			products.List = binary
		} else {
			products.Document = binary
		}
	}

	return products
}

// WhitespaceLayoutGoProducts finds the requested whitespace-layout Go reference programs.
func WhitespaceLayoutGoProducts(t *testing.T, ctx context.Context, root, target string) LayoutGoProducts {
	t.Helper()
	var products LayoutGoProducts
	cohere := filepath.Join(root, "cohere")
	for _, mode := range []struct{ name, driver, command string }{
		{"lists", "list_go.go", "adamic_markdown_lists"},
		{"layout", "document_go.go", "adamic_markdown_doclayout"},
	} {
		if target != "" && target != mode.name {
			continue
		}
		driver, err := filepath.Abs(filepath.Join("testdata", mode.driver))
		if err != nil {
			t.Fatal(err)
		}
		mainPath := filepath.Join(cohere, "cmd", mode.command, "main.go")
		replace := map[string]string{mainPath: driver}
		if mode.name == "lists" {
			bridge, err := filepath.Abs("testdata/list_bridge.go")
			if err != nil {
				t.Fatal(err)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, err := json.Marshal(map[string]any{"Replace": replace})
		if err != nil {
			t.Fatal(err)
		}

		goInputs := whitespaceGoInputs(t, root)
		goInputs.Name = "markdown-whitespace-layout-go-" + mode.name
		goInputs.Flags = append(goInputs.Flags, "go build "+strings.Join(buildcache.ReproducibleGoFlags(), " "), "overlay="+mode.driver, "CGO_ENABLED="+os.Getenv("CGO_ENABLED"), "GOFLAGS="+os.Getenv("GOFLAGS"))
		product := oracleProduct(t, goInputs, func(product string) error {
			overlayPath := filepath.Join(product, mode.name+".json")
			if err := os.WriteFile(overlayPath, overlay, 0644); err != nil {
				return err
			}
			defer os.Remove(overlayPath)
			command := toolchain.CommandContextIn(ctx, cohere, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+overlayPath, "-o", filepath.Join(product, "oracle"), mainPath)...)
			command.WaitDelay = time.Second
			if output, err := childguard.CombinedOutput(command, childguard.Options{Unbounded: true}); err != nil {
				return fmt.Errorf("Go %s: %w\n%s", mode.name, err, output)
			}
			return nil
		})
		binary := filepath.Join(product, "oracle")
		if mode.name == "lists" {
			products.List = binary
		} else {
			products.Document = binary
		}
	}

	return products
}

func whitespaceGoInputs(t *testing.T, root string) buildcache.Inputs {
	return buildcache.Inputs{Files: []string{"cohere", "go.mod", "go.work", "stage1/cohere/markdownblocks/testdata/list_go.go", "stage1/cohere/markdownblocks/testdata/list_bridge.go", "stage1/cohere/markdownblocks/testdata/document_go.go"},
		Flags:     []string{root, os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK"), os.Getenv("ADAMIC_NATIVE_SPLIT"), os.Getenv("GOFLAGS"), os.Getenv("CGO_ENABLED"), "Split=true", "Jobs=2"},
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH, buildcache.Tool("clang", "--version"), buildcache.Tool("go", "version")}}
}

// TextFormatterProduct finds the Go cohere text formatter product directory.
func TextFormatterProduct(t *testing.T, root string) string {
	t.Helper()
	directory := oracleProduct(t, textToolInputs("formatter"), func(dir string) error {
		command := toolchain.CommandContextIn(t.Context(), filepath.Join(root, "cohere"), "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-o", filepath.Join(dir, "cohere"), "./command/cohere")...)
		command.WaitDelay = time.Second
		output, err := childguard.CombinedOutput(command, childguard.Options{})
		if err != nil {
			return fmt.Errorf("formatter: %w\n%s", err, output)
		}
		return nil
	})
	return directory
}

// TextGeneratorProduct finds the text-class generator product directory.
func TextGeneratorProduct(t *testing.T, root string) string {
	t.Helper()
	directory := oracleProduct(t, textToolInputs("generator"), func(dir string) error {
		command := toolchain.CommandContextIn(t.Context(), root, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-o", filepath.Join(dir, "generate"), "./stage1/cohere/markdownblocks/tools/generate_classes")...)
		command.WaitDelay = time.Second
		output, err := childguard.CombinedOutput(command, childguard.Options{})
		if err != nil {
			return fmt.Errorf("generator: %w\n%s", err, output)
		}
		return nil
	})
	return directory
}

type markdownCompilers struct{ c, cPlusPlus, report string }

func goCompilers() markdownCompilers {
	var compilers markdownCompilers
	for _, setting := range toolchain.Env() {
		switch name, value, _ := strings.Cut(setting, "="); name {
		case "CC":
			compilers.c = value
		case "CXX":
			compilers.cPlusPlus = value
		}
	}
	compilers.report = "no C compiler"
	if fields := strings.Fields(compilers.c); len(fields) > 0 {
		compilers.report = buildcache.Tool(fields[0], "--version")
	}
	return compilers
}

func textToolInputs(target string) buildcache.Inputs {
	compilers := goCompilers()
	flags := []string{"GOFLAGS=" + os.Getenv("GOFLAGS"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + os.Getenv("GOOS"), "GOARCH=" + os.Getenv("GOARCH"), "CC=" + compilers.c, "CXX=" + compilers.cPlusPlus, "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")}
	tools := []string{buildcache.Tool("go", "version"), compilers.report}
	if target == "formatter" {
		return buildcache.Inputs{Name: "markdownblocks-text-formatter", Files: []string{"cohere"}, Flags: flags, Toolchain: tools}
	}
	return buildcache.Inputs{Name: "markdownblocks-text-generator", Files: []string{"go.mod", "stage1/cohere/markdownblocks/tools/generate_classes"}, Flags: flags, Toolchain: tools}
}
