// Command gate holds testdata string patterns to Go, source Node and both compiler paths.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
	"unicode/utf16"
)

type Match struct {
	Start, End int
	Text       string
}
type Sample struct {
	Input string
	Go    []Match `json:"go_matches"`
	Node  []Match `json:"node_matches"`
}
type Fixture struct {
	ID, Shape, Pattern, Flags, Outside, File string
	Go                                       string `json:"go_pattern"`
	Runtime                                  string `json:"runtime_file"`
	Samples                                  []Sample
}

func fail(format string, args ...any) { fmt.Fprintf(os.Stderr, format+"\n", args...); os.Exit(1) }
func run(command string, args ...string) []byte {
	c := exec.Command(command, args...)
	var out, errors bytes.Buffer
	c.Stdout = &out
	c.Stderr = &errors
	if err := c.Run(); err != nil || errors.Len() != 0 {
		fail("%s: %v: %s", command, err, errors.String())
	}
	return out.Bytes()
}
func expected(f Fixture) []byte {
	var b bytes.Buffer
	for input, s := range f.Samples {
		pattern := regexp.MustCompile(f.Go)
		matches := []Match{}
		for _, span := range pattern.FindAllStringIndex(s.Input, -1) {
			matches = append(matches, Match{len(utf16.Encode([]rune(s.Input[:span[0]]))), len(utf16.Encode([]rune(s.Input[:span[1]]))), s.Input[span[0]:span[1]]})
		}
		fresh, _ := json.Marshal(matches)
		recorded, _ := json.Marshal(s.Go)
		if !bytes.Equal(fresh, recorded) {
			fail("%s: Go oracle drift", f.ID)
		}

		a, _ := json.Marshal(s.Go)
		n, _ := json.Marshal(s.Node)
		if !bytes.Equal(a, n) {
			fail("%s: stored Go/Node spans differ", f.ID)
		}
		for _, m := range s.Go {
			fmt.Fprintf(&b, "%d:%d:%d:", input, m.Start, m.End)
			for _, u := range utf16.Encode([]rune(m.Text)) {
				fmt.Fprintf(&b, "%d,", u)
			}
			fmt.Fprintln(&b)
		}
	}
	return b.Bytes()
}
func equal(f Fixture, backend string, want, actual []byte) {
	if !bytes.Equal(want, actual) {
		fail("%s: %s fixture mismatch: expected %q got %q", f.ID, backend, want, actual)
	}
}
func main() {
	external := flag.String("runtime-js-compiler", "", "optional approved library compiler binary for runtime-string JS")
	force := flag.Bool("force-native", false, "require the first runtime fixture to pass native even before compiler landing")
	mutant := flag.Bool("mutant", false, "change one translation character in the first fixture and require comparison failure")
	prepare := flag.String("prepare", "", "build every fixture product into this directory without running comparisons")
	products := flag.String("products", "", "use previously prepared fixture products")
	selected := flag.Int("case", -1, "fixture index; -1 runs the entire inventory")
	flag.Parse()
	root, e := filepath.Abs("stage1/cohere/lint/regex/testdata/shapes")
	if e != nil {
		fail("%v", e)
	}
	repo, e := filepath.Abs(".")
	if e != nil {
		fail("%v", e)
	}
	data, e := os.ReadFile(filepath.Join(root, "fixtures.json"))
	if e != nil {
		fail("%v", e)
	}
	var fixtures []Fixture
	if e = json.Unmarshal(data, &fixtures); e != nil {
		fail("%v", e)
	}
	if len(fixtures) != 107 {
		fail("lost table rows: %d", len(fixtures))
	}
	table, e := os.ReadFile(filepath.Join(root, "../../table.json"))
	if e != nil {
		fail("%v", e)
	}
	var ids []struct{ ID string }
	if e = json.Unmarshal(table, &ids); e != nil {
		fail("%v", e)
	}
	if len(ids) != len(fixtures) {
		fail("table coverage drift")
	}

	for i, f := range fixtures {
		if f.ID != ids[i].ID {
			fail("table row %d coverage drift", i)
		}
	}
	if *prepare != "" {
		// Inputs: fixtures.json, table.json, every fixture and transitive import;
		// compiler Go sources/toolchain; native runtime and sanitizer flags; optional external compiler.
		// Product writes all lowered JS, refusal inputs and accepted native binaries into dir.
		product := func(dir string) error {
			for i, f := range fixtures {
				start := time.Now()
				if err := prepareShape(f, root, filepath.Join(dir, fmt.Sprintf("%03d", i)), *external); err != nil {
					return fmt.Errorf("%s: %w", f.ID, err)
				}
				fmt.Printf("BUILD fixture=%03d id=%s wall=%.6fs\n", i, f.ID, time.Since(start).Seconds())
			}
			return nil
		}
		if e := product(*prepare); e != nil {
			fail("prepare: %v", e)
		}
		pending := 0
		for i := range fixtures {
			data, err := os.ReadFile(filepath.Join(*prepare, fmt.Sprintf("%03d", i), "product.json"))
			if err != nil {
				fail("%v", err)
			}
			var product shapeProduct
			if err := json.Unmarshal(data, &product); err != nil {
				fail("%v", err)
			}
			if product.Refusal != "" {
				pending++
			}
		}
		fmt.Printf("PREPARED fixtures=%d pending=%d\n", len(fixtures), pending)
		return
	}
	if *products != "" {
		if *selected < -1 || *selected >= len(fixtures) {
			fail("invalid fixture index %d", *selected)
		}
		counts := map[string]int{}
		pending, nativePass, externalPass, total := 0, 0, 0, 0
		for i, f := range fixtures {
			if *selected >= 0 && i != *selected {
				continue
			}
			total++
			counts[f.Shape]++
			if f.Outside != "" {
				fmt.Printf("outside the port's shapes: %s: %s\n", f.ID, f.Outside)
				continue
			}
			p, n, x := checkPreparedShape(f, i, root, repo, filepath.Join(*products, fmt.Sprintf("%03d", i)), *mutant, *force)
			pending += p
			nativePass += n
			externalPass += x
		}
		fmt.Printf("PASS fixtures=%d shapes=%v source Node and area constant-string JS; runtime Node=%d library runtime JS=%d native=%d awaits codex/regex-runtime-compiler=%d\n", total, counts, total, externalPass, nativePass, pending)
		return
	}
	temp, e := os.MkdirTemp("", "regex-shapes-gate-")
	if e != nil {
		fail("%v", e)
	}
	defer os.RemoveAll(temp)
	nodeRunner := filepath.Join(repo, "oracle/node.mjs")
	counts := map[string]int{}
	pending, nativePass, externalPass := 0, 0, 0
	for i, f := range fixtures {
		if f.ID != ids[i].ID {
			fail("table row %d coverage drift", i)
		}
		counts[f.Shape]++
		if f.Outside != "" {
			fmt.Printf("outside the port's shapes: %s: %s\n", f.ID, f.Outside)
			continue
		}
		want := expected(f)
		entry := filepath.Join(root, f.File)
		if *mutant && i == 0 {
			content, e := os.ReadFile(entry)
			if e != nil {
				fail("%v", e)
			}
			changed := bytes.Replace(content, []byte(`"PaginationInput(`), []byte(`"PaginationInpuX(`), 1)
			if bytes.Equal(changed, content) {
				fail("mutant anchor drift")
			}
			changed = bytes.ReplaceAll(changed, []byte("'../observe.a'"), []byte("'"+filepath.Join(root, "observe.a")+"'"))
			entry = filepath.Join(temp, "mutant.a")
			if e = os.WriteFile(entry, changed, 0644); e != nil {
				fail("%v", e)
			}
		}
		actual := run("node", "--disable-warning=ExperimentalWarning", nodeRunner, entry)
		equal(f, "source Node", want, actual)
		p, e := load.Load([]string{entry})
		if e != nil {
			fail("%s: %v", f.ID, e)
		}
		ir, e := lower.Lower(context.Background(), p)
		if e != nil {
			fail("%s: constant-string emitted JS: %v", f.ID, e)
		}
		js := filepath.Join(temp, "static.js")
		if e = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); e != nil {
			fail("%v", e)
		}
		equal(f, "area emitted JavaScript", want, run("node", "--disable-warning=ExperimentalWarning", nodeRunner, js))
		runtimeEntry := filepath.Join(root, f.Runtime)
		equal(f, "runtime source Node", want, run("node", "--disable-warning=ExperimentalWarning", nodeRunner, runtimeEntry, f.Pattern, f.Flags))
		p, e = load.Load([]string{runtimeEntry})
		if e != nil {
			fail("%s: %v", f.ID, e)
		}
		ir, e = lower.Lower(context.Background(), p)
		if e != nil {
			if !strings.Contains(e.Error(), "RegExp with a nonconstant pattern") {
				fail("%s: unexpected native refusal: %v", f.ID, e)
			}
			pending++
			if *force && i == 0 {
				fail("%s: awaits codex/regex-runtime-compiler: forced native requirement caught refusal: %v", f.ID, e)
			}
		} else { // No skip or grandfathering: native is required immediately after lowering accepts strings.
			binary := filepath.Join(temp, "native")
			if e = native.Build(native.C(ir), binary, native.Options{Sanitize: true}); e != nil {
				fail("%s: %v", f.ID, e)
			}
			equal(f, "sanitized runtime native", want, run(binary, f.Pattern, f.Flags))
			if e = os.WriteFile(js, []byte(javascript.JavaScript(ir)), 0644); e != nil {
				fail("%v", e)
			}
			equal(f, "area runtime emitted JS", want, run("node", "--disable-warning=ExperimentalWarning", nodeRunner, js, f.Pattern, f.Flags))
			nativePass++
		}
		if *external != "" {
			output := run(*external, "js", runtimeEntry)
			if e = os.WriteFile(js, output, 0644); e != nil {
				fail("%v", e)
			}
			equal(f, "library runtime emitted JS", want, run("node", "--disable-warning=ExperimentalWarning", nodeRunner, js, f.Pattern, f.Flags))
			externalPass++
		}
	}
	fmt.Printf("PASS fixtures=%d shapes=%v source Node and area constant-string JS; runtime Node=%d library runtime JS=%d native=%d awaits codex/regex-runtime-compiler=%d\n", len(fixtures), counts, len(fixtures), externalPass, nativePass, pending)
}

// shapeProduct records build inputs, never comparison observations.
type shapeProduct struct {
	ID       string
	Refusal  string
	Native   bool
	External bool
}

func prepareShape(f Fixture, root, dir, external string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	product := shapeProduct{ID: f.ID}
	if f.Outside == "" {
		// Constant program: lowered JavaScript is a build product.
		program, err := load.Load([]string{filepath.Join(root, f.File)})
		if err != nil {
			return err
		}
		ir, err := lower.Lower(context.Background(), program)
		if err != nil {
			return fmt.Errorf("constant-string emitted JS: %w", err)
		}
		if err := os.WriteFile(filepath.Join(dir, "static.js"), []byte(javascript.JavaScript(ir)), 0644); err != nil {
			return err
		}
		runtimeEntry := filepath.Join(root, f.Runtime)
		program, err = load.Load([]string{runtimeEntry})
		if err != nil {
			return err
		}
		ir, err = lower.Lower(context.Background(), program)
		if err != nil {
			product.Refusal = err.Error()
		} else {
			// Accepted runtime strings immediately require both sanitized native and JS.
			source := native.C(ir)
			if err := os.WriteFile(filepath.Join(dir, "native.c"), []byte(source), 0644); err != nil {
				return err
			}
			if err := native.Build(source, filepath.Join(dir, "native"), native.Options{Sanitize: true}); err != nil {
				return err
			}
			if err := os.WriteFile(filepath.Join(dir, "runtime.js"), []byte(javascript.JavaScript(ir)), 0644); err != nil {
				return err
			}
			product.Native = true
		}
		if external != "" {
			command := exec.Command(external, "js", runtimeEntry)
			var output, errors bytes.Buffer
			command.Stdout, command.Stderr = &output, &errors
			if err := command.Run(); err != nil || errors.Len() != 0 {
				return fmt.Errorf("external compiler: %v: %s", err, &errors)
			}
			if err := os.WriteFile(filepath.Join(dir, "external.js"), output.Bytes(), 0644); err != nil {
				return err
			}
			product.External = true
		}
	}
	data, err := json.Marshal(product)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "product.json"), data, 0644)
}

func checkPreparedShape(f Fixture, index int, root, repo, dir string, mutant, force bool) (pending, nativePass, externalPass int) {
	data, err := os.ReadFile(filepath.Join(dir, "product.json"))
	if err != nil {
		fail("%s: product: %v", f.ID, err)
	}
	var product shapeProduct
	if err := json.Unmarshal(data, &product); err != nil || product.ID != f.ID {
		fail("%s: product identity drift: %v", f.ID, err)
	}
	want := expected(f)
	entry := filepath.Join(root, f.File)
	if mutant && index == 0 {
		content, err := os.ReadFile(entry)
		if err != nil {
			fail("%v", err)
		}
		changed := bytes.Replace(content, []byte(`"PaginationInput(`), []byte(`"PaginationInpuX(`), 1)
		if bytes.Equal(changed, content) {
			fail("mutant anchor drift")
		}
		changed = bytes.ReplaceAll(changed, []byte("'../observe.a'"), []byte("'"+filepath.Join(root, "observe.a")+"'"))
		temp, err := os.MkdirTemp("", "regex-shape-mutant-")
		if err != nil {
			fail("%v", err)
		}
		defer os.RemoveAll(temp)
		entry = filepath.Join(temp, "mutant.a")
		if err := os.WriteFile(entry, changed, 0644); err != nil {
			fail("%v", err)
		}
	}
	runner := filepath.Join(repo, "oracle/node.mjs")
	equal(f, "source Node", want, run("node", "--disable-warning=ExperimentalWarning", runner, entry))
	equal(f, "area emitted JavaScript", want, run("node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "static.js")))
	equal(f, "runtime source Node", want, run("node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(root, f.Runtime), f.Pattern, f.Flags))
	if product.Refusal != "" {
		if !strings.Contains(product.Refusal, "RegExp with a nonconstant pattern") {
			fail("%s: unexpected native refusal: %s", f.ID, product.Refusal)
		}
		pending++
		if force && index == 0 {
			fail("%s: awaits codex/regex-runtime-compiler: forced native requirement caught refusal: %s", f.ID, product.Refusal)
		}
	} else {
		if !product.Native {
			fail("%s: missing accepted native product", f.ID)
		}
		equal(f, "sanitized runtime native", want, run(filepath.Join(dir, "native"), f.Pattern, f.Flags))
		equal(f, "area runtime emitted JS", want, run("node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "runtime.js"), f.Pattern, f.Flags))
		nativePass++
	}
	if product.External {
		equal(f, "library runtime emitted JS", want, run("node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(dir, "external.js"), f.Pattern, f.Flags))
		externalPass++
	}
	return
}
