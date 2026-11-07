// Package oracles owns markdownblocks' compiler-free Go reference products.
// Paths resolve from stage1/cohere/markdownblocks in both test binaries.
package oracles

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
	"unicode/utf16"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/corpusfiles"
	"github.com/system-inc/adamic/internal/toolchain"
)

// LayoutGoProducts holds the two Go reference programs for a layout family.
type LayoutGoProducts struct{ List, Document string }

var oracleProducer bool

// EnableProductProduction enables recipes in the owner test binary's TestMain.
func EnableProductProduction() { oracleProducer = true }

// oracleProduct builds an oracle in the owner test binary and finds it everywhere else, naming the owner package as
// buildcache.BuiltElsewhere does in every other oracles package, so ahra loom check's products step, which settles the
// owner packages a "<directory> owners" literal names, settles these too. With the gate inside the recipe instead, the
// check settled no markdownblocks/oracles owner and every check reaching markdownblocks went red (Operations, Oct 28).
func oracleProduct(test *testing.T, inputs buildcache.Inputs, build func(string) error) string {
	test.Helper()
	inputs.Name += " oracle-owner-v1"
	if oracleProducer {
		return buildcache.Product(test, inputs, build)
	}
	return buildcache.BuiltElsewhere(test, inputs.Name, inputs, "stage1/cohere/markdownblocks/oracles owners", func(
		string,
	) error {
		return fmt.Errorf(
			"oracle %s is missing: settle the stage1/cohere/markdownblocks/oracles owners first",
			inputs.Name,
		)
	})
}

// LeafCompositionGoProduct finds one leaf-composition Go reference program.
func LeafCompositionGoProduct(test *testing.T, root, name, driverName string) string {
	test.Helper()
	cohere := filepath.Join(root, "cohere")
	bridge, absolutePathError := filepath.Abs("testdata/list_bridge.go")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	driver, absolutePathError := filepath.Abs(filepath.Join("testdata", driverName))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	main := filepath.Join(cohere, "cmd", name, "main.go")
	inputs := buildcache.Inputs{Name: "markdown-leaf-go-" + name, Files: []string{
		"stage1/cohere/markdownblocks/testdata/" + driverName,
		"stage1/cohere/markdownblocks/testdata/list_bridge.go",
		"cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim",
		"cohere/go.mod", "cohere/go.sum", "cohere/go.work", "cohere/go.work.sum", "go.mod", "go.work",
	}, Flags: []string{
		"overlay=" + driverName,
		"GOFLAGS=" + os.Getenv("GOFLAGS"),
		"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
		"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
	}, Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH}}
	product := oracleProduct(test, inputs, func(out string) error {
		replacements := map[string]string{main: driver}
		if name == "adamic_markdown_lists" {
			replacements[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, encodeError := json.Marshal(map[string]any{"Replace": replacements})
		if encodeError != nil {
			return encodeError
		}
		overlayPath := filepath.Join(out, "overlay.json")
		if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
			return writeError
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		cancelContext, cancel := context.WithCancel(context.Background())
		test.Cleanup(cancel)
		if output, toolchainGoError := toolchainGo(cancelContext, cohere, append(
			append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
			"-overlay="+overlayPath,
			"-o",
			filepath.Join(out, "oracle"),
			main,
		)...); toolchainGoError != nil {
			return fmt.Errorf("Go bridge: %v\n%s", toolchainGoError, output)
		}
		return nil
	})
	return filepath.Join(product, "oracle")
}

// QuoteLayoutGoProducts finds the requested quote-layout Go reference programs.
func QuoteLayoutGoProducts(test *testing.T, root, target string) LayoutGoProducts {
	test.Helper()
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
		driver, absolutePathError := filepath.Abs(filepath.Join("testdata", mode.driver))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		mainPath := filepath.Join(cohere, "cmd", mode.command, "main.go")
		replace := map[string]string{mainPath: driver}
		if mode.name == "lists" {
			bridge, absolutePathError := filepath.Abs("testdata/list_bridge.go")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		goInputs := buildcache.Inputs{
			Name: "markdown-quote-layout-go-" + mode.name,
			Files: []string{
				"cohere",
				"go.mod",
				"go.work",
				"stage1/cohere/markdownblocks/testdata/list_go.go",
				"stage1/cohere/markdownblocks/testdata/list_bridge.go",
				"stage1/cohere/markdownblocks/testdata/document_go.go",
			},
			Flags: []string{
				"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
				"overlay=" + mode.name,
				"GOFLAGS=" + os.Getenv("GOFLAGS"),
				"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
				"GOOS=" + os.Getenv("GOOS"),
				"GOARCH=" + os.Getenv("GOARCH"),
			},
			Toolchain: []string{runtime.Version(), buildcache.Tool("go", "version")},
		}
		dir := oracleProduct(test, goInputs, func(dir string) error {
			overlayPath := filepath.Join(dir, mode.name+".json")
			if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
				return writeError
			}
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(overlayPath)
			if output, toolchainGoError := toolchainGo(test.Context(), cohere, append(
				append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
				"-overlay="+overlayPath,
				"-o",
				filepath.Join(dir, mode.name),
				mainPath,
			)...); toolchainGoError != nil {
				return fmt.Errorf("Go %s: %w\n%s", mode.name, toolchainGoError, output)
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
func StructureLayoutGoProducts(test *testing.T, root, target string) LayoutGoProducts {
	test.Helper()
	var products LayoutGoProducts
	cohere := filepath.Join(root, "cohere")
	for _, item := range []struct {
		command, driver string
		bridge          bool
	}{{"adamic_markdown_lists", "list_go.go", true}, {"adamic_markdown_doclayout", "document_go.go", false}} {
		if target != "" && target != item.command {
			continue
		}
		driver, absolutePathError := filepath.Abs(filepath.Join("testdata", item.driver))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		main := filepath.Join(cohere, "cmd", item.command, "main.go")
		replace := map[string]string{main: driver}
		if item.bridge {
			bridge, absolutePathError := filepath.Abs("testdata/list_bridge.go")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
		if encodeError != nil {
			test.Fatal(encodeError)
		}
		// Keep overlay files inside the content-addressed product. Their source
		// paths are stable, so warm setup fetches the Go oracle instead of
		// charging whichever shard first enters sync.Once for a new build.
		goFiles := []string{
			"cohere/internal",
			"cohere/go.mod",
			"cohere/go.sum",
			"cohere/go.work",
			"cohere/go.work.sum",
			"cohere/TypeScript/tsc",
			"cohere/TypeScript-shim",
			"cohere/mutation_aliasing",
			"cohere/static_single_assignment",
			"go.mod",
			"go.work",
			"stage1/cohere/markdownblocks/testdata/" + item.driver,
		}
		if item.bridge {
			goFiles = append(goFiles, "stage1/cohere/markdownblocks/testdata/list_bridge.go")
		}
		goDir := oracleProduct(test, buildcache.Inputs{
			Name: "markdown-structure-go-" + item.command, Files: goFiles,
			Flags: []string{
				"overlay=" + string(overlay),
				"main=" + main,
				"GOFLAGS=" + os.Getenv("GOFLAGS"),
				"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
				"GOOS=" + runtime.GOOS,
				"GOARCH=" + runtime.GOARCH,
				"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
			},
			Toolchain: []string{buildcache.Tool("go", "version")},
		}, func(dir string) error {
			overlayPath := filepath.Join(dir, "overlay.json")
			if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
				return writeError
			}
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(overlayPath)
			cancelContext, cancel := context.WithCancel(context.Background())
			defer cancel()
			if output, toolchainGoError := toolchainGo(cancelContext, cohere, append(
				append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
				"-overlay="+overlayPath,
				"-o",
				filepath.Join(dir, "oracle"),
				main,
			)...); toolchainGoError != nil {
				return fmt.Errorf("Go bridge: %w\n%s", toolchainGoError, output)
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
func TableLayoutGoProducts(test *testing.T, root, target string) LayoutGoProducts {
	test.Helper()
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
		driverPath, absolutePathError := filepath.Abs(filepath.Join("testdata", driver))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		main := filepath.Join(cohere, "cmd", command, "main.go")
		replacements := map[string]string{main: driverPath}
		if mode == "list" {
			bridge, absolutePathError := filepath.Abs("testdata/list_bridge.go")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			replacements[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, encodeError := json.Marshal(map[string]any{"Replace": replacements})
		if encodeError != nil {
			test.Fatal(encodeError)
		}

		goInputs := buildcache.Inputs{
			Name:      "markdown table Go " + mode,
			Files:     []string{"cohere", "stage1/cohere/markdownblocks/testdata/" + driver},
			Flags:     []string{string(overlay), "go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")},
			Toolchain: []string{buildcache.Tool("go", "version")},
		}
		if mode == "list" {
			goInputs.Files = append(goInputs.Files, "stage1/cohere/markdownblocks/testdata/list_bridge.go")
		}
		product := oracleProduct(test, goInputs, func(product string) error {
			overlayPath := filepath.Join(product, mode+".json")
			if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
				return writeError
			}
			defer os.Remove(overlayPath)
			cancelContext, cancel := context.WithCancel(test.Context())
			defer cancel()
			if output, toolchainGoError := toolchainGo(cancelContext, cohere, append(
				append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
				"-overlay="+overlayPath,
				"-o",
				filepath.Join(product, "oracle"),
				main,
			)...); toolchainGoError != nil {
				return fmt.Errorf("Go %s: %w\n%s", mode, toolchainGoError, output)
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
func WhitespaceLayoutGoProducts(test *testing.T, callerContext context.Context, root, target string) LayoutGoProducts {
	test.Helper()
	var products LayoutGoProducts
	cohere := filepath.Join(root, "cohere")
	for _, mode := range []struct{ name, driver, command string }{
		{"lists", "list_go.go", "adamic_markdown_lists"},
		{"layout", "document_go.go", "adamic_markdown_doclayout"},
	} {
		if target != "" && target != mode.name {
			continue
		}
		driver, absolutePathError := filepath.Abs(filepath.Join("testdata", mode.driver))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		mainPath := filepath.Join(cohere, "cmd", mode.command, "main.go")
		replace := map[string]string{mainPath: driver}
		if mode.name == "lists" {
			bridge, absolutePathError := filepath.Abs("testdata/list_bridge.go")
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			replace[filepath.Join(cohere, "internal/format/markdown/adamic_lists.go")] = bridge
		}
		overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
		if encodeError != nil {
			test.Fatal(encodeError)
		}

		goInputs := whitespaceGoInputs(test, root)
		goInputs.Name = "markdown-whitespace-layout-go-" + mode.name
		goInputs.Flags = append(
			goInputs.Flags,
			"go build "+strings.Join(buildcache.ReproducibleGoFlags(), " "),
			"overlay="+mode.driver,
			"CGO_ENABLED="+os.Getenv("CGO_ENABLED"),
			"GOFLAGS="+os.Getenv("GOFLAGS"),
		)
		product := oracleProduct(test, goInputs, func(product string) error {
			overlayPath := filepath.Join(product, mode.name+".json")
			if writeError := os.WriteFile(overlayPath, overlay, 0644); writeError != nil {
				return writeError
			}
			defer os.Remove(overlayPath)
			if output, toolchainGoError := toolchainGo(callerContext, cohere, append(
				append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
				"-overlay="+overlayPath,
				"-o",
				filepath.Join(product, "oracle"),
				mainPath,
			)...); toolchainGoError != nil {
				return fmt.Errorf("Go %s: %w\n%s", mode.name, toolchainGoError, output)
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

func whitespaceGoInputs(test *testing.T, root string) buildcache.Inputs {
	return buildcache.Inputs{Files: []string{
		"cohere",
		"go.mod",
		"go.work",
		"stage1/cohere/markdownblocks/testdata/list_go.go",
		"stage1/cohere/markdownblocks/testdata/list_bridge.go",
		"stage1/cohere/markdownblocks/testdata/document_go.go",
	},
		Flags: []string{
			root,
			os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK"),
			os.Getenv("ADAMIC_NATIVE_SPLIT"),
			os.Getenv("GOFLAGS"),
			os.Getenv("CGO_ENABLED"),
			"Split=true",
			"Jobs=2",
		},
		Toolchain: []string{
			runtime.Version(),
			runtime.GOOS,
			runtime.GOARCH,
			buildcache.Tool("clang", "--version"),
			buildcache.Tool("go", "version"),
		}}
}

// TextFormatterProduct finds the Go cohere text formatter product directory.
func TextFormatterProduct(test *testing.T, root string) string {
	test.Helper()
	directory := oracleProduct(test, textToolInputs("formatter"), func(dir string) error {
		output, toolchainGoError := toolchainGo(test.Context(), filepath.Join(root, "cohere"), append(
			append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
			"-o",
			filepath.Join(dir, "cohere"),
			"./command/cohere",
		)...)
		if toolchainGoError != nil {
			return fmt.Errorf("formatter: %w\n%s", toolchainGoError, output)
		}
		return nil
	})
	return directory
}

// TextGeneratorProduct finds the text-class generator product directory.
func TextGeneratorProduct(test *testing.T, root string) string {
	test.Helper()
	directory := oracleProduct(test, textToolInputs("generator"), func(dir string) error {
		output, toolchainGoError := toolchainGo(test.Context(), root, append(
			append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
			"-o",
			filepath.Join(dir, "generate"),
			"./stage1/cohere/markdownblocks/tools/generate_classes",
		)...)
		if toolchainGoError != nil {
			return fmt.Errorf("generator: %w\n%s", toolchainGoError, output)
		}
		return nil
	})
	return directory
}

func textToolInputs(target string) buildcache.Inputs {
	return textGoInputs(target, goCompilerIdentity())
}

const repository = "../../.."

type AuditInput struct {
	Corpus bool   `json:"-"`
	Name   string `json:"name"`
	Text   string `json:"text"`
}

func Census() bool {
	return os.Getenv("ADAMIC_MARKDOWNBLOCKS_CENSUS") != ""
}

func AuditCorpus(test *testing.T, root string) ([]AuditInput, int) {
	test.Helper()
	var inputs []AuditInput
	paths := auditCorpusFiles(test, root)
	for _, path := range paths {
		content, readError := os.ReadFile(path)
		if readError != nil {
			test.Fatal(readError)
		}
		name, relativePathError := filepath.Rel(root, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		inputs = append(inputs, AuditInput{Name: filepath.ToSlash(name), Text: string(content)})
	}
	files := len(inputs)
	for _, marker := range []string{"*", "-", "+", "1.", "1)", "10.", "999999999."} {
		for depth := 1; depth <= 5; depth++ {
			var lines []string
			for level := 0; level < depth; level++ {
				lines = append(lines, strings.Repeat("  ", level)+marker+" item *text* `x`")
			}
			inputs = append(inputs, AuditInput{
				Name: "generated/list/" + marker + "/" + strings.Repeat("n", depth),
				Text: strings.Join(lines, "\n") + "\n",
			})
		}
	}
	for _, text := range []string{"# x #", "## x", "x\n=", "x\n---", "a\nb\n==="} {
		inputs = append(inputs, AuditInput{Name: "generated/heading/" + text, Text: text + "\n"})
	}
	for _, prefix := range []string{"", "> ", "- ", "  "} {
		for _, block := range []string{
			"```\nx\n```",
			"~~~\nx\n~~~",
			"    x\n    y",
			"<div>\nx\n</div>",
			"<!-- x -->",
			"***",
			"---\na: b\n---",
			"|a|b|\n|-|-|\n|x|y|",
		} {
			inputs = append(inputs, AuditInput{
				Name: "generated/block/" + prefix + block,
				Text: prefix + strings.ReplaceAll(block, "\n", "\n"+prefix) + "\n",
			})
		}
	}
	for _, code := range []string{
		`if (true) { console.log("x"); }`,
		`for (let i=0;i<2;i++) console.log(i);`,
		`const x = <div className="x">x</div>;`,
	} {
		inputs = append(inputs, AuditInput{Name: "generated/embedding/" + code, Text: "```tsx\n" + code + "\n```\n"})
	}
	for _, name := range []string{"embedded_jsonc.md", "embedded_flow.md"} {
		content, readError := os.ReadFile(filepath.Join("gaps", name))
		if readError != nil {
			test.Fatal(readError)
		}
		inputs = append(inputs, AuditInput{Name: "generated/" + name, Text: string(content)})
	}
	return inputs, files
}

func BlockCorpus(test *testing.T, root, slice string) ([]AuditInput, int) {
	test.Helper()
	inputs, files := AuditCorpus(test, root)
	for _, marker := range []string{"-", "*", "+", "0.", "1.", "1)", "9.", "10)", "99.", "999999999."} {
		for _, spacing := range []string{" ", "  ", "   ", "    ", "     ", "\t"} {
			for _, task := range []string{"", "[ ] ", "[x] ", "[X] "} {
				for _, body := range []string{
					"a *b _c_* `x`",
					"a\n\n    b\n\n    - c",
					"a\n\n    > q\n    > r",
					"a\n<div>\nx\n</div>",
					"\n\n\n    code",
				} {
					text := marker + spacing + task + body + "\n"
					inputs = append(inputs, AuditInput{Corpus: true, Name: "generated/list-layout/" + text, Text: text})
				}
			}
		}
	}
	for _, text := range []string{
		"0. a\n1. b\n", "0. a\n1. b\n1. c\n", "2. a\n1. b\n7. c\n", "999999999. a\n1. b\n",
		"- a\n* b\n+ c\n- d\n", "1. a\n2) b\n3. c\n", "-\n\n\n    x\n", "1.\n\n\n        x\n",
		"- [x] a\n\n    - [ ] b\n      c\n", "- a\n\n  ```js\n  let x = 1;\n  ```\n",
		"<!-- prettier-ignore -->\n*  a\n+   b\n", "> <!-- prettier-ignore -->\n> *  a\n>\n",
		"- a\n\n  |a|b|\n  |-|-|\n  |x|y|\n", "- a\n\n  ### h\n\n  ---\n",
		"- <div>\n  x\n  </div>\n", "- a\n\n  [x]: https://example.test\n  [y]: /y\n",
		"- a\n\n  [^x]: footnote\n\n- [^x]\n", "1.    中😀\n1.    _a*b*_\n",
	} {
		inputs = append(inputs, AuditInput{Name: "generated/list-layout/edge/" + text, Text: text})
	}
	if slice != "lists" {
		for _, prefix := range []string{">", "> ", " > ", "  >  ", "   >\t", "> > ", ">> ", "> > > ", ">\t> "} {
			for _, body := range []string{
				"",
				"text",
				"a\nb",
				"a\n\nb",
				"# h\n\na",
				"a\n---",
				"- a\n- b",
				"1. a\n\n   - b",
				"- [x] a",
				"```js\nlet x=1;\n```",
				"    x\n    y",
				"<div>\nx\n</div>",
				"a\n<div>\nx",
				"[a]: /x\n[b]: /y",
				"| a | b |\n| - | - |\n| x | y |",
				"<!-- prettier-ignore -->\n+    a",
				"中😀 _a*b*_",
			} {
				text := prefix + strings.ReplaceAll(body, "\n", "\n"+prefix) + "\n"
				inputs = append(inputs, AuditInput{Corpus: true, Name: "generated/quote-layout/" + text, Text: text})
			}
		}
	}
	if slice != "lists" && slice != "quotes" {
		for _, align := range []string{"---", ":--", "--:", ":-:"} {
			for _, cell := range []string{
				"",
				"a",
				"abcde",
				"abcdef",
				"中😀",
				"*a _b_*",
				"`a|b`",
				"a\\|b",
				"[a](/b)",
				"<em>a</em>",
			} {
				for _, prefix := range []string{"", "> ", "- ", "> - "} {
					text := "| a | b |\n| " + align + " | " + align + " |\n| " + cell + " | x |\n| long text | y |\n"
					lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
					text = prefix + lines[0] + "\n"
					continuation := prefix
					if strings.HasSuffix(prefix, "- ") {
						continuation = strings.TrimSuffix(prefix, "- ") + "  "
					}
					for _, line := range lines[1:] {
						text += continuation + line + "\n"
					}
					inputs = append(
						inputs,
						AuditInput{Corpus: true, Name: "generated/table-layout/" + text, Text: text},
					)
				}
			}
		}
		for _, text := range []string{
			"a | b\n- | -\nc\nd | e | f\n",
			"a | b\n:- | -:\n | \n",
			"| a |\n| - |\n| x | y |\n",
			"| 😀 | 中 |\n| :-: | -: |\n| é | 👨‍👩‍👧‍👦 |\n",
		} {
			inputs = append(inputs, AuditInput{Name: "generated/table-layout/edge/" + text, Text: text})
		}
	}
	if slice == "code" ||
		(slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace") {
		for _, marker := range []string{"```", "````", "~~~~", "```````"} {
			for _, info := range []string{
				"",
				"js",
				"json",
				"yaml",
				"toml",
				"css",
				"html",
				"text title=foo",
				"x {#id .class}",
			} {
				for _, body := range []string{"", "a", "a\nb", "\n\nx\n", "a  \nb\t", "`a`", "```", "~~~~", "中😀"} {
					for _, prefix := range []string{"", "> ", "- ", "> - "} {
						text := marker + info + "\n" + body + "\n" + marker + "\n"
						lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
						text = prefix + lines[0] + "\n"
						continuation := prefix
						if strings.HasSuffix(prefix, "- ") {
							continuation = strings.TrimSuffix(prefix, "- ") + "  "
						}
						for _, line := range lines[1:] {
							text += continuation + line + "\n"
						}
						inputs = append(
							inputs,
							AuditInput{Corpus: true, Name: "generated/code-layout/" + text, Text: text},
						)
					}
				}
			}
		}
		for _, indent := range []string{"    ", "\t", "     ", "  \t", "\t "} {
			for _, body := range []string{"a", "a  \nb\t", "\n\nx\n", "中😀"} {
				text := indent + strings.ReplaceAll(body, "\n", "\n"+indent) + "\n"
				inputs = append(
					inputs,
					AuditInput{Corpus: true, Name: "generated/code-layout/indent/" + text, Text: text},
				)
			}
		}
	}
	if slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, body := range []string{
			"<div>\nx  \n</div>",
			"<script>\nx  \n</script>",
			"<style>\nx\t\n</style>",
			"<pre>\nx  \n</pre>",
			"<!-- a  \nb\t -->",
			"<!-->",
			"<!--->",
			"<!--a-->",
			"<?xml\nx  \n?>",
			"<!DOCTYPE html>",
			"<![CDATA[\nx  \n]]>",
			"<table>\n<tr>\nx\n</tr>\n</table>",
			"<x-a a='b'>\nx\n</x-a>",
			"a <em>\nx  \n</em> b",
			"<div>\n\n# h\n\n</div>",
			"<!-- prettier-ignore -->\n<div>  \nx\n</div>",
			"<div>中😀</div>",
		} {
			for _, prefix := range []string{"", "> ", "> > ", "- ", "> - "} {
				for _, ending := range []string{"", "  ", "\t", "\u00a0", "\u2000", "\ufeff"} {
					text := body + ending + "\n"
					lines := strings.Split(strings.TrimSuffix(text, "\n"), "\n")
					text = prefix + lines[0] + "\n"
					continuation := prefix
					if strings.HasSuffix(prefix, "- ") {
						continuation = strings.TrimSuffix(prefix, "- ") + "  "
					}
					for _, line := range lines[1:] {
						text += continuation + line + "\n"
					}
					inputs = append(inputs, AuditInput{Corpus: true, Name: "generated/html-layout/" + text, Text: text})
				}
			}
		}
	}
	if slice == "html" || slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, cell := range []string{
			"中",
			"Ａ",
			"α",
			"é",
			"©",
			"©️",
			"👩🏽‍⚕️",
			"👨🏻‍❤️‍💋‍👨🏿",
			"🧑🏿‍🦽‍➡️",
			"🏳️‍🌈",
			"🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f",
		} {
			for _, align := range []string{"---", ":--", "--:", ":-:"} {
				text := "| " + cell + " | a |\n| " + align + " | --- |\n| a | " + cell + " |\n"
				inputs = append(
					inputs,
					AuditInput{Corpus: true, Name: "generated/native-width/table/" + text, Text: text},
				)
			}
		}
	}
	if slice == "structure" || slice == "root" || slice == "leaves" || slice == "whitespace" {
		for depth := 1; depth <= 6; depth++ {
			for _, body := range []string{"", "a", "*a _b_*", "中 👩🏽‍⚕️", "[a](/b)", "a #", "`a  b`"} {
				for _, prefix := range []string{"", "> ", "- ", "> - "} {
					text := prefix + strings.Repeat("#", depth) + " " + body + "\n"
					inputs = append(
						inputs,
						AuditInput{Corpus: true, Name: "generated/structure/atx/" + text, Text: text},
					)
				}
			}
		}
		for _, marker := range []string{"=", "===", "---", "-----"} {
			for _, body := range []string{"a", "a\nb", "*a _b_*", "中\n文", "a\n\ntext", "[x](/y)"} {
				for _, spacing := range []string{"", "  ", "\t"} {
					text := body + "\n" + marker + spacing + "\n"
					inputs = append(
						inputs,
						AuditInput{Corpus: true, Name: "generated/structure/setext/" + text, Text: text},
					)
				}
			}
		}
	}
	if slice == "root" || slice == "leaves" || slice == "whitespace" {
		for _, space := range []string{"", " ", "\t", "\u00a0", "\u2028", "\ufeff"} {
			for _, body := range []string{
				"+    a\n*   b",
				"a\n====",
				"|a|b|\n|-|-|\n|中|😀|",
				"```js\na  b\n```",
				"<div>  \nx\n</div>",
			} {
				for _, marker := range []string{"next", "range", "nested", "unmatched"} {
					comment := "<!--" + space + "prettier-ignore" + space + "-->"
					text := comment + "\n" + body + "\n\n# after\n"
					if marker != "next" {
						text = "before\n\n<!--" + space + "prettier-ignore-start" + space + "-->\n" + body + "\n"
						if marker == "nested" {
							text += "<!-- prettier-ignore-start -->\n+   extra\n"
						}
						if marker != "unmatched" {
							text += "<!--" + space + "prettier-ignore-end" + space + "-->\n\n# after\n"
						}
					}
					inputs = append(
						inputs,
						AuditInput{Corpus: true, Name: "generated/root/ignore/" + marker + "/" + text, Text: text},
					)
				}
			}
		}
		for _, text := range []string{
			"",
			" \n",
			"a\n<div>\nx\n</div>\n",
			"[a]: /x\n[b]: /y\n",
			"<!-- prettier-ignore-start -->\n+    a\n<!-- prettier-ignore-end -->\n\n<!-- prettier-ignore-start -->\n*    b\n<!-- prettier-ignore-end -->\n",
			"<!-- prettier-ignore-end -->\n+    a\n",
		} {
			inputs = append(inputs, AuditInput{Name: "generated/root/edge/" + text, Text: text})
		}
	}
	if slice == "leaves" || slice == "whitespace" {
		for _, label := range []string{"a", "中", "a b", "x-y"} {
			for _, url := range []string{"/x", "", "<x>", "/a%20b", "https://example.test"} {
				for _, title := range []string{"", " \"title\"", " 'a \"b\"'", " (a'b)"} {
					text := "[" +
						label +
						"](" +
						url +
						title +
						") ![" +
						label +
						"](" +
						url +
						title +
						")\n\n[" +
						label +
						"]: " +
						url +
						title +
						"\n"
					inputs = append(
						inputs,
						AuditInput{Corpus: true, Name: "generated/leaves/links/" + text, Text: text},
					)
				}
			}
		}
		for _, text := range []string{
			"*a* **b** ~~c~~ _d_ 1*2*3 1***2***3\n",
			"_<https://example.test>_\n",
			"a  \nb\nc\\\nd\n",
			"[^x]\n\n[^x]: first\n    second\n\n    - a\n    - b\n",
			"- a\n\n  ***\n\n- b\n\n  ---\n",
			"$$ title\na+b\n$$\n\n$x + y$\n",
			"{{ a }} {% b %}\n",
			"[[wiki link]]\n",
			"![a][b] [a][b] ![b][] [b][] ![b] [b]\n\n[b]: /x 'title'\n",
		} {
			inputs = append(inputs, AuditInput{Name: "generated/leaves/edge/" + text, Text: text})
		}
	}
	if slice == "whitespace" {
		for _, previous := range []string{"a", "中", "한", "。", "α", "…"} {
			for _, next := range []string{"a", "中", "한", "1.", "123)", "。", "-"} {
				for _, space := range []string{" ", "\n", "\t"} {
					text := previous + space + next + "\n"
					inputs = append(inputs, AuditInput{Corpus: true, Name: "generated/whitespace/" + text, Text: text})
				}
			}
		}
		for _, text := range []string{
			"[a  中][b]\n\n[b]: /x\n",
			"![a  中][b] [a\n中][b]\n\n[b]: /x\n",
			"|a b|中 c|\n|-|-|\n|x y|한 z|\n",
			"[^x]\n\n[^x]: a\n\n    {{ b }}\n",
		} {
			inputs = append(inputs, AuditInput{Name: "generated/whitespace/edge/" + text, Text: text})
		}
	}
	return inputs, files
}

func NumericBatch(inputs [][]uint16) []byte {
	var batch strings.Builder
	for _, units := range inputs {
		if len(units) == 0 {
			batch.WriteString("-")
		} else {
			for index, unit := range units {
				if index > 0 {
					batch.WriteByte(',')
				}
				fmt.Fprint(&batch, unit)
			}
		}
		batch.WriteByte('\n')
	}
	return []byte(batch.String())
}

func TokenizerEventLiveEnumeration(test *testing.T) (string, [][]uint16, []string, int) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	corpus, files := BlockCorpus(test, root, "whitespace")
	// Upstream selection verifies the two fixed checkout pins against Git.
	// The live repository portion must independently remain non-empty.
	repositoryFiles, cohereFiles, checkerFiles := 0, 0, 0
	for _, item := range corpus {
		switch {
		case strings.HasPrefix(item.Name, "cohere/TypeScript/"):
			checkerFiles++
		case strings.HasPrefix(item.Name, "cohere/"):
			cohereFiles++
		case !strings.HasPrefix(item.Name, "generated/"):
			repositoryFiles++
		}
	}
	if repositoryFiles == 0 {
		test.Fatal("repository Markdown corpus is empty")
	}
	// Only pinned upstream counts are fixed: cohere 7945d102 and
	// TypeScript d92d9bfe, using AuditCorpus's explicitly selected roots.
	if Census() && (cohereFiles != 786 || checkerFiles != 67) {
		test.Fatalf("pinned upstream file counts: cohere %d/786, TypeScript %d/67", cohereFiles, checkerFiles)
	}
	inputs := [][]uint16{{}}
	names := []string{"empty"}
	for _, item := range corpus {
		inputs = append(inputs, utf16.Encode([]rune(item.Text)))
		names = append(names, item.Name)
	}
	for unit := 0; unit <= 65535; unit++ {
		inputs = append(inputs, []uint16{uint16(unit)})
		names = append(names, fmt.Sprintf("unit/%d", unit))
	}
	alphabet := []uint16{0, 9, 10, 13, 65279, 65}
	for length := 1; length <= 5; length++ {
		count := 1
		for index := 0; index < length; index++ {
			count *= len(alphabet)
		}
		for n := 0; n < count; n++ {
			units := make([]uint16, length)
			number := n
			for unitIndex := range units {
				units[unitIndex] = alphabet[number%len(alphabet)]
				number /= len(alphabet)
			}
			inputs = append(inputs, units)
			names = append(names, fmt.Sprintf("sequence/%d/%d", length, n))
		}
	}
	for _, prefix := range []string{"", "a", "ab", "abc", "abcd", "abcde", "中", "😀", "a😀"} {
		for _, tail := range []string{"\t", "\t\t", "\r\n", "\rX\n", "\rX", "\x00\t", "\ufeff\t"} {
			inputs = append(inputs, utf16.Encode([]rune(prefix+tail)))
			names = append(names, fmt.Sprintf("columns/%q/%q", prefix, tail))
		}
	}
	return root, inputs, names, files
}

func GrainWidthInputs(test *testing.T) (string, []AuditInput, int) {
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	inputs, files := AuditCorpus(test, root)
	original := append([]AuditInput(nil), inputs...)
	for _, input := range original {
		for index, line := range strings.Split(input.Text, "\n") {
			inputs = append(inputs, AuditInput{Name: fmt.Sprintf("%s:%d", input.Name, index+1), Text: line})
		}
	}
	for code := rune(0); code <= 0x10ffff; code++ {
		if code >= 0xd800 && code <= 0xdfff {
			continue
		}
		inputs = append(inputs, AuditInput{Name: fmt.Sprintf("U+%06X", code), Text: string(code)})
	}
	atoms := []string{
		"a",
		"\x7f",
		"中",
		"α",
		"\u0300",
		"\ufe0f",
		"\ufe0e",
		"\u200d",
		"©",
		"❤",
		"☀",
		"😀",
		"👩",
		"👨",
		"🧑",
		"🤝",
		"🏳",
		"🏴",
		"🇦",
		"🇿",
		"🏽",
		"🏻",
		"⚕",
		"♂",
		"♀",
		"1",
		"#",
		"\u20e3",
		"🦯",
		"➡",
	}
	for _, a := range atoms {
		for _, b := range atoms {
			for _, c := range atoms {
				inputs = append(inputs, AuditInput{Name: "sequence/" + a + b + c, Text: a + b + c})
			}
		}
	}
	for _, text := range []string{
		"👩🏽‍⚕️",
		"👨🏻‍❤️‍💋‍👨🏿",
		"🏴\U000e0067\U000e0062\U000e0065\U000e006e\U000e0067\U000e007f",
		"👩‍👩‍👧‍👦",
		"🏳️‍🌈",
		"🧑🏿‍🦽‍➡️",
		"a\x7f",
		"中\x7f",
	} {
		inputs = append(inputs, AuditInput{Name: "long/" + text, Text: text})
	}
	return root, inputs, files
}

func GrainWidthBatch(inputs []AuditInput) []byte {
	escape := strings.NewReplacer(`\`, `\\`, "\n", `\n`, "\r", `\r`, "\t", `\t`)
	var batch strings.Builder
	for _, input := range inputs {
		batch.WriteString("x" + escape.Replace(input.Text) + "\n")
	}
	return []byte(batch.String())
}

func GrainWidthFork(root string) string {
	if fork := os.Getenv("ADAMIC_MARKDOWNBLOCKS_FORK"); fork != "" {
		return fork
	}
	return filepath.Join(root, "cohere/internal/format/prettier/bundles")
}

func root(test testing.TB) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return root
}

const ownerDirectory = "stage1/cohere/markdownblocks/oracles"

var errorOwnedHere = errors.New("built by " + ownerDirectory + "'s product tests")

func goOracle(test *testing.T, owner, module, output, packagePath string, arguments []string) string {
	test.Helper()
	output += "-oracle-owner-v1"
	if oracleProducer {
		if module == "" {
			return buildcache.GoBuild(test, output, packagePath, arguments)
		}
		return buildcache.GoBuildIn(test, module, output, packagePath, arguments)
	}
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs(module, output, packagePath, arguments, nil)
	if goBuildInInputsError != nil {
		test.Fatalf("go build %s: %v", packagePath, goBuildInInputsError)
	}
	directory := buildcache.BuiltElsewhere(
		test,
		output,
		inputs,
		ownerDirectory+" "+owner,
		func(string) error { return errorOwnedHere },
	)
	return filepath.Join(directory, output)
}

func writeOverlay(directory, name string, replace map[string]string) (string, error) {
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		return "", encodeError
	}
	overlayPath := filepath.Join(directory, name)
	return overlayPath, os.WriteFile(overlayPath, overlay, 0644)
}

// toolchainGo runs the toolchain's go in directory and returns its stdout, then its stderr, for an error message.
// Output is read through childguard.Capture, to its end with no timer deciding when: CombinedOutput's copying under a
// one-second WaitDelay reported exec.ErrWaitDelay whenever the reads hadn't reached EOF a second after go exited, so a
// build that had finished could fail its product under load (#n3gk22h's shape). Capture has no stall guard, and the
// builds that had one lose nothing: a go build writes nothing until it fails, so only childguard's first-output window
// (30 minutes) could fire, and ctx and the unit's own ceiling end it sooner (Loom kills a test unit at 90 s and a
// product unit at 600 s).
func toolchainGo(callerContext context.Context, directory string, arguments ...string) ([]byte, error) {
	stdout, stderr, captureError := childguard.Capture(
		callerContext,
		toolchain.CommandContextIn(callerContext, directory, "go", arguments...),
	)
	return append(stdout, stderr...), captureError
}

// callerGo is toolchainGo with the caller's environment.
func callerGo(callerContext context.Context, directory string, arguments ...string) ([]byte, error) {
	command := toolchain.CommandContextIn(callerContext, directory, "go", arguments...)
	// The caller's environment, with PWD set to the directory, as a command with no environment of its own gets.
	command.Env = nil
	stdout, stderr, captureError := childguard.Capture(callerContext, command)
	return append(stdout, stderr...), captureError
}

type run struct {
	stdout, stderr []byte
	exitCode       int
}

// execute reads a run's output through childguard.Capture, as toolchainGo does and for its reason. Here the old
// WaitDelay was worse than red: an exec.ErrWaitDelay came with a ProcessState, so it passed as exit 0 with the output
// cut where the reads were closed, and an answers product could store that.
func execute(callerContext context.Context, environment []string, name string, arguments ...string) (run, error) {
	command := toolchain.CommandContextIn(callerContext, "", name, arguments...)
	command.Env = nil
	if environment != nil {
		command.Env = append(os.Environ(), environment...)
	}
	stdout, stderr, captureError := childguard.Capture(callerContext, command)
	if callerContext.Err() != nil {
		return run{}, callerContext.Err()
	}
	if captureError != nil && command.ProcessState == nil {
		return run{}, fmt.Errorf("running %s: %w", name, captureError)
	}
	return run{stdout, stderr, command.ProcessState.ExitCode()}, nil
}

type compilerIdentity struct{ C, CPlusPlus, Report string }

func goCompilerIdentity() compilerIdentity {
	var compilers compilerIdentity
	for _, setting := range toolchain.Env() {
		switch name, value, _ := strings.Cut(setting, "="); name {
		case "CC":
			compilers.C = value
		case "CXX":
			compilers.CPlusPlus = value
		}
	}
	compilers.Report = "no C compiler"
	if fields := strings.Fields(compilers.C); len(fields) > 0 {
		compilers.Report = buildcache.Tool(fields[0], "--version")
	}
	return compilers
}

func overlayGoOracle(test *testing.T, owner, command, output string, replacements map[string]string) string {
	test.Helper()
	cohere := filepath.Join(root(test), "cohere")
	mainPath := filepath.Join(cohere, "cmd", command, "main.go")
	replace := map[string]string{}
	for target, source := range replacements {
		absolute, absolutePathError := filepath.Abs(filepath.Join("testdata", source))
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		if target == "" {
			target = mainPath
		} else {
			target = filepath.Join(cohere, target)
		}
		replace[target] = absolute
	}
	overlayPath, writeOverlayError := writeOverlay(test.TempDir(), "overlay.json", replace)
	if writeOverlayError != nil {
		test.Fatal(writeOverlayError)
	}
	return goOracle(test, owner, "cohere", output, mainPath, []string{"-overlay=" + overlayPath})
}

func ChunksGoOracle(test *testing.T) string {
	test.Helper()
	return overlayGoOracle(test, "TestProduct_MarkdownChunksGo", "adamic_chunks", "go-chunks", map[string]string{
		"": "chunks_go.go",
		"internal/format/markdown/micromark/adamic_chunks.go": "chunks_bridge.go",
	})
}

func ASTGoOracle(test *testing.T) string {
	test.Helper()
	return overlayGoOracle(test, "TestProduct_MarkdownASTGo", "adamic_ast", "go-ast", map[string]string{
		"":                                       "ast_go.go",
		"internal/format/markdown/adamic_ast.go": "ast_bridge.go",
		"internal/format/markdown/adamic_ast_facts.go": "ast_facts.go",
	})
}

func FrontMatterGoOracle(test *testing.T) string {
	test.Helper()
	return overlayGoOracle(
		test,
		"TestProduct_MarkdownFrontMatterGo",
		"adamic_frontmatter",
		"go-parser",
		map[string]string{"": "frontmatter_go.go"},
	)
}

func PrefixGoOracle(test *testing.T) string {
	test.Helper()
	return overlayGoOracle(test, "TestProduct_MarkdownPrefixGo", "adamic_prefix", "go-prefix", map[string]string{
		"": "prefix_go.go",
		"internal/format/markdown/micromark/adamic_prefix.go": "prefix_bridge.go",
	})
}

func PilotGoProduct(test *testing.T, family string) string {
	test.Helper()
	root := root(test)
	cohere := filepath.Join(root, "cohere")
	command, output := "adamic_identifier", "go-identifier"
	targets := map[string]string{"": "identifier_go.go"}
	if family == "events" {
		command, output = "adamic_chunks", "go-chunks"
		targets = map[string]string{
			"": "events_go.go",
			"internal/format/markdown/micromark/adamic_chunks.go":          "events_bridge.go",
			"internal/format/markdown/micromark/adamic_event_transport.go": "events_transport.go",
		}
	}
	files := []string{"cohere"}
	replace := make(map[string]string)
	for target, source := range targets {
		files = append(files, "stage1/cohere/markdownblocks/testdata/"+source)
		if target == "" {
			target = filepath.Join(cohere, "cmd", command, "main.go")
		} else {
			target = filepath.Join(cohere, target)
		}
		replace[target] = filepath.Join(root, "stage1/cohere/markdownblocks/testdata", source)
	}
	// Sort map-derived inputs to keep the recipe key deterministic.
	sort.Strings(files)
	flags := []string{"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "), "overlay:" + command}
	for _, name := range []string{"GOFLAGS", "GOOS", "GOARCH", "CGO_ENABLED", "GOAMD64", "CC", "CXX"} {
		flags = append(flags, name+"="+os.Getenv(name))
	}
	dir := oracleProduct(test, buildcache.Inputs{
		Name:      "markdownblocks-pilot-" + family + "-go",
		Files:     files,
		Flags:     flags,
		Toolchain: []string{buildcache.Tool("go", "version")},
	}, func(dir string) error {
		overlayPath, writeOverlayError := writeOverlay(dir, "overlay.json", replace)
		if writeOverlayError != nil {
			return writeOverlayError
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		if result, callerGoError := callerGo(test.Context(), cohere, append(
			append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
			"-overlay="+overlayPath,
			"-o",
			filepath.Join(dir, output),
			filepath.Join(cohere, "cmd", command, "main.go"),
		)...); callerGoError != nil {
			return fmt.Errorf("%s Go product: %w\n%s", family, callerGoError, result)
		}
		return nil
	})
	return filepath.Join(dir, output)
}

func TokenizerEventAnswers(test *testing.T, goBinary string, batch []byte) []byte {
	test.Helper()
	binary, readError := os.ReadFile(goBinary)
	if readError != nil {
		test.Fatal(readError)
	}
	inputs := buildcache.Inputs{
		Name:  "markdownblocks-tokenizer-events-answers",
		Files: []string{"stage1/cohere/markdownblocks/oracles/oracles.go"},
		Flags: []string{
			fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary)),
			fmt.Sprintf("cases-sha256=%x", sha256.Sum256(batch)),
		},
		Toolchain: []string{runtime.GOOS, runtime.GOARCH},
	}
	// A key names its inputs relative to the repository root, while auditCorpusFiles returns the absolute paths
	// AuditCorpus reads, so convert here (#y6ff7ea).
	repositoryRoot := root(test)
	for _, path := range auditCorpusFiles(test, repositoryRoot) {
		relative, relativePathError := filepath.Rel(repositoryRoot, path)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		if relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			test.Fatalf("markdown corpus file %s is outside the repository", path)
		}
		inputs.Files = append(inputs.Files, filepath.ToSlash(relative))
	}
	directory := oracleProduct(test, inputs, func(directory string) error {
		scratch, temporaryDirectoryError := os.MkdirTemp("", "markdown-events-cases-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		defer os.RemoveAll(scratch)
		cases := filepath.Join(scratch, "cases.txt")
		if writeError := os.WriteFile(cases, batch, 0644); writeError != nil {
			return writeError
		}
		want, executeError := execute(test.Context(), nil, goBinary, cases)
		if executeError != nil {
			return executeError
		}
		if want.exitCode != 0 || len(want.stderr) > 0 {
			return fmt.Errorf("full Go mutant oracle exit %d stderr %s", want.exitCode, want.stderr)
		}
		return os.WriteFile(filepath.Join(directory, "answers.txt"), want.stdout, 0644)
	})
	answers, readError := os.ReadFile(filepath.Join(directory, "answers.txt"))
	if readError != nil {
		test.Fatal(readError)
	}
	return answers
}

func TokenizerEventsAnswers(test *testing.T) []byte {
	test.Helper()
	_, inputs, _, _ := TokenizerEventLiveEnumeration(test)
	return TokenizerEventAnswers(test, PilotGoProduct(test, "events"), NumericBatch(inputs))
}

func textGoInputs(target string, compilers compilerIdentity) buildcache.Inputs {
	flags := []string{
		"GOFLAGS=" + os.Getenv("GOFLAGS"),
		"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
		"GOOS=" + os.Getenv("GOOS"),
		"GOARCH=" + os.Getenv("GOARCH"),
		"CC=" + compilers.C,
		"CXX=" + compilers.CPlusPlus,
		"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
	}
	tools := []string{buildcache.Tool("go", "version"), compilers.Report}
	switch target {
	case "go":
		return buildcache.Inputs{Name: "markdownblocks-text-go", Files: []string{
			"cohere",
			"stage1/cohere/markdownblocks/testdata/text_go.go",
			"stage1/cohere/markdownblocks/testdata/text_bridge.go",
		}, Flags: append(flags, "overlay text_go.go,text_bridge.go"), Toolchain: tools}
	case "formatter":
		return buildcache.Inputs{
			Name:      "markdownblocks-text-formatter",
			Files:     []string{"cohere"},
			Flags:     flags,
			Toolchain: tools,
		}
	case "generator":
		return buildcache.Inputs{
			Name:      "markdownblocks-text-generator",
			Files:     []string{"go.mod", "stage1/cohere/markdownblocks/tools/generate_classes"},
			Flags:     flags,
			Toolchain: tools,
		}
	}
	panic("unknown text Go product " + target)
}

func TextGoBuild(test *testing.T, target string) string {
	test.Helper()
	if target == "formatter" {
		return filepath.Join(TextFormatterProduct(test, root(test)), "cohere")
	}
	if target == "generator" {
		return filepath.Join(TextGeneratorProduct(test, root(test)), "generate")
	}
	root := root(test)
	cohere := filepath.Join(root, "cohere")
	binary, directory, packagePath := "go-text", cohere, filepath.Join(cohere, "cmd/adamic_text/main.go")
	switch target {
	case "formatter":
		binary, packagePath = "cohere", "./command/cohere"
	case "generator":
		binary, directory, packagePath = "generate", root, "./stage1/cohere/markdownblocks/tools/generate_classes"
	}
	dir := oracleProduct(test, textGoInputs(target, goCompilerIdentity()), func(dir string) error {
		arguments := append([]string{"build"}, buildcache.ReproducibleGoFlags()...)
		if target == "go" {
			overlayPath, writeOverlayError := writeOverlay(dir, "overlay.json", map[string]string{
				packagePath: filepath.Join(root, "stage1/cohere/markdownblocks/testdata/text_go.go"),
				filepath.Join(
					cohere,
					"internal/format/markdown/adamic_text.go",
				): filepath.Join(root, "stage1/cohere/markdownblocks/testdata/text_bridge.go"),
			})
			if writeOverlayError != nil {
				return writeOverlayError
			}
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(overlayPath)
			arguments = append(arguments, "-overlay="+overlayPath)
		}
		if output, toolchainGoError := toolchainGo(
			test.Context(),
			directory,
			append(arguments, "-o", filepath.Join(dir, binary), packagePath)...,
		); toolchainGoError != nil {
			return fmt.Errorf("text %s: %w\n%s", target, toolchainGoError, output)
		}
		return nil
	})
	return filepath.Join(dir, binary)
}

func MalformedEventsGoOracle(test *testing.T) string {
	test.Helper()
	cohere := filepath.Join(root(test), "cohere")
	dir := oracleProduct(test, buildcache.Inputs{
		Name: "markdownblocks-malformed-events-go-errors",
		Files: []string{
			"cohere",
			"stage1/cohere/markdownblocks/testdata/mdast_go.go",
			"stage1/cohere/markdownblocks/testdata/mdast_bridge.go",
			"stage1/cohere/markdownblocks/testdata/events_transport.go",
		},
		Flags: []string{
			"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
			"overlay: mdast_go.go, mdast_bridge.go, events_transport.go",
			"GOFLAGS=" + os.Getenv("GOFLAGS"),
			"GOOS=" + os.Getenv("GOOS"),
			"GOARCH=" + os.Getenv("GOARCH"),
			"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
			"GOAMD64=" + os.Getenv("GOAMD64"),
		},
		Toolchain: []string{buildcache.Tool("go", "version")},
	}, func(dir string) error {
		mainPath := filepath.Join(cohere, "cmd/adamic_mdast_errors/main.go")
		replace := map[string]string{}
		for _, p := range []struct{ target, source string }{
			{mainPath, "testdata/mdast_go.go"},
			{filepath.Join(cohere, "internal/format/markdown/mdast/adamic_mdast.go"), "testdata/mdast_bridge.go"},
			{
				filepath.Join(cohere, "internal/format/markdown/micromark/adamic_events.go"),
				"testdata/events_transport.go",
			},
		} {
			source, absolutePathError := filepath.Abs(p.source)
			if absolutePathError != nil {
				return absolutePathError
			}
			replace[p.target] = source
		}
		overlayPath, writeOverlayError := writeOverlay(dir, "overlay.json", replace)
		if writeOverlayError != nil {
			return writeOverlayError
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		if output, callerGoError := callerGo(test.Context(), cohere, append(
			append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
			"-overlay="+overlayPath,
			"-o",
			filepath.Join(dir, "go-errors"),
			mainPath,
		)...); callerGoError != nil {
			return fmt.Errorf("Go errors: %w\n%s", callerGoError, output)
		}
		return nil
	})
	return filepath.Join(dir, "go-errors")
}

type grainRecipe struct {
	command string
	overlay map[string]string
}

var grainRecipes = map[string]grainRecipe{
	"path": {"adamic_path", map[string]string{
		"cmd/adamic_path/main.go":                       "path_go.go",
		"internal/format/markdown/adamic_path_facts.go": "path_facts.go",
		"internal/format/markdown/adamic_path.go":       "path_bridge.go",
		"internal/format/markdown/adamic_ast.go":        "ast_bridge.go",
	}},
	"mdast": {"adamic_mdast", map[string]string{
		"cmd/adamic_mdast/main.go":                            "mdast_go.go",
		"internal/format/markdown/mdast/adamic_mdast.go":      "mdast_bridge.go",
		"internal/format/markdown/micromark/adamic_events.go": "events_transport.go",
	}},
	"decode": {"adamic_decode", map[string]string{"cmd/adamic_decode/main.go": "decode_go.go"}},
	"width":  {"adamic_width", map[string]string{"cmd/adamic_width/main.go": "width_go.go"}},
}

func grainGoInputs(name string, compilers compilerIdentity) buildcache.Inputs {
	inputs := buildcache.Inputs{
		Name: "markdown-grain-" + name + "-go-v2",
		Files: []string{
			"go.mod",
			"go.work",
			"cohere",
			"stage1/cohere/markdownblocks/testdata",
			"stage1/cohere/markdownblocks/oracles/oracles.go",
		},
		Toolchain: []string{buildcache.Tool("go", "version"), compilers.Report},
		Flags:     []string{"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " ")},
	}
	for _, variable := range []string{
		"GOFLAGS",
		"CGO_ENABLED",
		"GOOS",
		"GOARCH",
		"GOAMD64",
		"GOARM64",
		"GOEXPERIMENT",
		"CGO_CFLAGS",
		"CGO_CPPFLAGS",
		"CGO_CXXFLAGS",
		"CGO_LDFLAGS",
	} {
		inputs.Flags = append(inputs.Flags, variable+"="+os.Getenv(variable))
	}
	inputs.Flags = append(
		inputs.Flags,
		"CC="+compilers.C,
		"CXX="+compilers.CPlusPlus,
		buildcache.Tool("go", "env", "GOWORK"),
	)
	return inputs
}

func GrainGoOracle(test *testing.T, name string) string {
	test.Helper()
	recipe := grainRecipes[name]
	root := root(test)
	cohere := filepath.Join(root, "cohere")
	dir := oracleProduct(test, grainGoInputs(name, goCompilerIdentity()), func(dir string) error {
		replace := map[string]string{}
		for to, from := range recipe.overlay {
			target := filepath.Join(cohere, "cmd", recipe.command, "main.go")
			if to != "" {
				target = filepath.Join(cohere, to)
			}
			replace[target] = filepath.Join(root, "stage1/cohere/markdownblocks/testdata", from)
		}
		overlayPath, writeOverlayError := writeOverlay(dir, "overlay.json", replace)
		if writeOverlayError != nil {
			return writeOverlayError
		}
		// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
		defer os.Remove(overlayPath)
		if output, toolchainGoError := toolchainGo(test.Context(), cohere, append(
			append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
			"-overlay="+overlayPath,
			"-o",
			filepath.Join(dir, "oracle"),
			filepath.Join(cohere, "cmd", recipe.command, "main.go"),
		)...); toolchainGoError != nil {
			return fmt.Errorf("Go oracle: %w\n%s", toolchainGoError, output)
		}
		return nil
	})
	return filepath.Join(dir, "oracle")
}

type grainToolPreparation struct {
	once sync.Once
	path string
}

var grainToolPreparations sync.Map

func GrainTool(test *testing.T, name string) string {
	test.Helper()
	state, _ := grainToolPreparations.LoadOrStore(name, &grainToolPreparation{})
	prepared := state.(*grainToolPreparation)
	prepared.once.Do(func() {
		packagePath := "./stage1/cohere/markdownblocks/tools/generate_" + name
		owner := "TestProduct_GrainMarkdownEntitiesGenerator"
		switch name {
		case "formatter":
			packagePath, owner = "github.com/system-inc/cohere/command/cohere", "TestProduct_GrainMarkdownFormatter"
		case "width":
			owner = "TestProduct_GrainMarkdownWidthGenerator"
		}
		prepared.path = goOracle(test, owner, "", "grain-"+name, packagePath, nil)
	})
	return prepared.path
}

func GrainWidthFullAnswers(test *testing.T, goBinary, fork string, batch []byte) ([]byte, []byte) {
	test.Helper()
	binary, readError := os.ReadFile(goBinary)
	if readError != nil {
		test.Fatal(readError)
	}
	standalone, readError := os.ReadFile(filepath.Join(fork, "standalone.js"))
	if readError != nil {
		test.Fatal(readError)
	}
	inputs := buildcache.Inputs{
		Name: "markdown-grain-width-full-answers",
		Files: []string{
			"stage1/cohere/markdownblocks/testdata/width_fork.mjs",
			"stage1/cohere/markdownblocks/oracles/oracles.go",
		},
		Flags: []string{
			fmt.Sprintf("oracle-sha256=%x", sha256.Sum256(binary)),
			fmt.Sprintf("fork-sha256=%x", sha256.Sum256(standalone)),
			fmt.Sprintf("cases-sha256=%x", sha256.Sum256(batch)),
		},
		Toolchain: []string{runtime.GOOS, runtime.GOARCH, buildcache.Tool("node", "--version")},
	}
	directory := oracleProduct(test, inputs, func(directory string) error {
		scratch, temporaryDirectoryError := os.MkdirTemp("", "markdown-width-cases-")
		if temporaryDirectoryError != nil {
			return temporaryDirectoryError
		}
		defer os.RemoveAll(scratch)
		cases := filepath.Join(scratch, "cases.txt")
		if writeError := os.WriteFile(cases, batch, 0644); writeError != nil {
			return writeError
		}
		for _, side := range []struct {
			name, file, command string
			args                []string
		}{
			{"Go StringWidth", "go.txt", goBinary, []string{cases}},
			{"actual fork group-fit widths", "fork.txt", "node", []string{"testdata/width_fork.mjs", fork, cases}},
		} {
			result, executeError := execute(test.Context(), nil, side.command, side.args...)
			if executeError != nil {
				return executeError
			}
			if result.exitCode != 0 || len(result.stderr) > 0 {
				return fmt.Errorf("%s exit %d stderr %s", side.name, result.exitCode, result.stderr)
			}
			if writeError := os.WriteFile(filepath.Join(directory, side.file), result.stdout, 0644); writeError != nil {
				return writeError
			}
		}
		return nil
	})
	read := func(file string) []byte {
		data, readError := os.ReadFile(filepath.Join(directory, file))
		if readError != nil {
			test.Fatal(readError)
		}
		return data
	}
	return read("go.txt"), read("fork.txt")
}

func GrainMarkdownWidthFullAnswers(test *testing.T) ([]byte, []byte) {
	test.Helper()
	root, inputs, _ := GrainWidthInputs(test)
	return GrainWidthFullAnswers(test, GrainGoOracle(test, "width"), GrainWidthFork(root), GrainWidthBatch(inputs))
}

var preflightMutations = []struct{ name, path, from, to string }{
	{
		"heading prefix",
		"heading.go",
		`strings.Repeat("#", currentNode(path).Depth)`,
		`strings.Repeat("#", currentNode(path).Depth+1)`,
	},
	{"list marker", "list.go", `rawPrefix = "- "`, `rawPrefix = "+ "`},
	{
		"fence length",
		"print.go",
		`max(3, getMaxContinuousCount(value, "` + "`" + `")+1)`,
		`max(4, getMaxContinuousCount(value, "` + "`" + `")+1)`,
	},
}

func PreflightGoProducts(test *testing.T) (string, []string) {
	test.Helper()
	budget := newPreflightProductBudget(test)
	defer budget.close()
	root := root(test)
	driver, absolutePathError := filepath.Abs("testdata/go_driver.go")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_markdown_blocks/main.go")
	alteredSources := make([][]byte, len(preflightMutations))
	for index, mutation := range preflightMutations {
		content, readError := os.ReadFile(filepath.Join(cohere, "internal/format/markdown", mutation.path))
		if readError != nil {
			test.Fatal(readError)
		}
		if strings.Count(string(content), mutation.from) != 1 {
			test.Fatal("mutant must alter exactly one site")
		}
		alteredSources[index] = []byte(strings.Replace(string(content), mutation.from, mutation.to, 1))
	}
	buildFiles := []string{
		"go.mod",
		"go.work",
		"cohere",
		"stage1/cohere/markdownblocks/testdata/go_driver.go",
		"stage1/cohere/markdownblocks/oracles/oracles.go",
	}
	for _, name := range []string{"go.sum", "go.work.sum"} {
		if _, statError := os.Stat(filepath.Join(root, name)); statError == nil {
			buildFiles = append(buildFiles, name)
		} else if !os.IsNotExist(statError) {
			test.Fatal(statError)
		}
	}
	product := oracleProduct(test, buildcache.Inputs{
		Name:  "markdownblocks-whole-document-preflight-go-products",
		Files: buildFiles,
		Flags: []string{
			"go build -overlay -o cmd/adamic_markdown_blocks/main.go",
			"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
		},
		Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool(
			"go",
			"env",
			"GOOS",
			"GOARCH",
			"CGO_ENABLED",
			"GOEXPERIMENT",
			"GOFLAGS",
			"GOTOOLCHAIN",
			"GOWORK",
		)},
	}, func(dir string) error {

		build := func(replacements map[string]string, binary string) error {
			replacements[mainPath] = driver
			overlayPath, writeOverlayError := writeOverlay(
				filepath.Dir(binary),
				filepath.Base(binary)+".overlay.json",
				replacements,
			)
			if writeOverlayError != nil {
				return writeOverlayError
			}
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(overlayPath)
			if output, preflightGoError := preflightGo(budget, cohere, append(
				append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
				"-overlay="+overlayPath,
				"-o",
				binary,
				mainPath,
			)...); preflightGoError != nil {
				return fmt.Errorf("build: %w\n%s", preflightGoError, output)
			}
			return nil
		}
		if buildError := build(map[string]string{}, filepath.Join(dir, "go-printer")); buildError != nil {
			return buildError
		}
		for index, mutation := range preflightMutations {
			altered := filepath.Join(dir, mutation.path)
			if writeError := os.WriteFile(altered, alteredSources[index], 0644); writeError != nil {
				return writeError
			}
			source := filepath.Join(cohere, "internal/format/markdown", mutation.path)
			if buildError := build(
				map[string]string{source: altered},
				filepath.Join(dir, fmt.Sprintf("mutant-%d", index)),
			); buildError != nil {
				return buildError
			}
		}
		return nil
	})
	mutants := make([]string, len(preflightMutations))
	for index := range preflightMutations {
		mutants[index] = filepath.Join(product, fmt.Sprintf("mutant-%d", index))
	}
	return filepath.Join(product, "go-printer"), mutants
}

type preflightProductBudget struct {
	label  string
	limit  time.Duration
	ctx    context.Context
	cancel context.CancelFunc
	timer  *time.Timer
	groups sync.Map
}

func newPreflightProductBudget(test *testing.T) *preflightProductBudget {
	test.Helper()
	label, limit := test.Name(), 590*time.Second
	timeoutContext, cancel := context.WithTimeout(context.Background(), limit)
	budget := &preflightProductBudget{ctx: timeoutContext, cancel: cancel, label: label, limit: limit}
	budget.timer = time.AfterFunc(limit, budget.cooked)
	return budget
}
func (budget *preflightProductBudget) cooked() {
	budget.cancel()
	budget.groups.Range(func(pid, _ any) bool { _ = syscall.Kill(-pid.(int), syscall.SIGKILL); return true })
	fmt.Fprintf(os.Stderr, "cooked: %s exceeded %.0fs hard deadline\n", budget.label, budget.limit.Seconds())
	os.Exit(124)
}
func (budget *preflightProductBudget) close() { budget.timer.Stop(); budget.cancel() }
func preflightGo(budget *preflightProductBudget, directory string, arguments ...string) ([]byte, error) {
	defer func() {
		if errors.Is(budget.ctx.Err(), context.DeadlineExceeded) {
			budget.cooked()
		}
	}()
	command := toolchain.CommandContextIn(budget.ctx, directory, "go", arguments...)
	command.Env = nil
	// Still CombinedOutput, not toolchainGo's Capture: cooked kills the groups Started registers before it exits the
	// process, and Capture has no Started to register this one by, so its group would outlive the exit. No WaitDelay,
	// so exec reads the output to its end and a build slow to close its pipes under load isn't red by chance
	// (#n3gk22h's shape): go waits for the compilers it starts before it exits, so nothing holds a pipe after it, and
	// the budget's hard deadline ends the wait if something did.
	var processID int
	defer func() { budget.groups.Delete(processID) }()
	return childguard.CombinedOutput(
		command,
		childguard.Options{Unbounded: true, Started: func(pid int) { processID = pid; budget.groups.Store(pid, true) }},
	)
}

// TextGoProductInputs describes the actual text recipe for compiler-key checks.
func TextGoProductInputs(target, compiler, compilerPlusPlus, report string) buildcache.Inputs {
	return textGoInputs(target, compilerIdentity{compiler, compilerPlusPlus, report})
}

// GrainGoProductInputs describes the actual grain recipe for compiler-key checks.
func GrainGoProductInputs(name, compiler, compilerPlusPlus, report string) buildcache.Inputs {
	return grainGoInputs(name, compilerIdentity{compiler, compilerPlusPlus, report})
}

func auditCorpusFiles(test *testing.T, root string) []string {
	test.Helper()
	patterns := []string{"*.md", "*.markdown", "*.mdown", "*.mkd"}
	roots := []string{"stage1/cohere/markdownblocks"}
	if Census() {
		roots = []string{"."}
	}
	paths := corpusfiles.Repository(test, root, roots, patterns)
	if Census() {
		paths = append(
			paths,
			corpusfiles.Upstream(test, filepath.Join(root, "cohere"), corpusfiles.CohereCommit, []string{
				"CHANGELOG.md",
				"CONTRIBUTING.md",
				"README.md",
				"THIRD_PARTY_NOTICES.md",
				"TypeScript-shim",
				"editors",
				"internal",
				"schema",
				"swift",
			}, patterns)...,
		)
		paths = append(paths, corpusfiles.Upstream(
			test,
			filepath.Join(root, "cohere/TypeScript"),
			corpusfiles.TypeScriptGoCommit,
			[]string{
				".github",
				"CODE_OF_CONDUCT.md",
				"CONTRIBUTING.md",
				"README.md",
				"SECURITY.md",
				"SUPPORT.md",
				"packages",
				"tsc",
			},
			patterns,
		)...)
	}
	sort.Strings(paths)

	return paths
}
