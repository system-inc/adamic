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
	force := flag.Bool("force-native", false, "run the first fixture through the required runtime native path")
	mutant := flag.Bool("mutant", false, "change one translation character in the first fixture and require comparison failure")
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
	if len(fixtures) == 0 {
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
			fail("%s: runtime RegExp required after area/library merge; pattern=%q flags=%q: %v", f.ID, f.Pattern, f.Flags, e)
		}
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
		if *force && i == 0 {
			fmt.Printf("PASS forced native fixture=%s native=1, identical matches and UTF-16 spans\n", f.ID)
			return
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
