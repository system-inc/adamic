package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

type Reduction struct {
	Index     int        `json:"index"`
	File      string     `json:"file"`
	Rule      string     `json:"rule"`
	Signature string     `json:"signature"`
	SourceSHA string     `json:"source_sha256"`
	Input     string     `json:"deletion_minimal_input"`
	Go        WorkAnswer `json:"go"`
	Node      WorkAnswer `json:"node"`
	PortSites []string   `json:"port_sites"`
	GoSites   []string   `json:"go_sites"`
	Verified  bool       `json:"verified_single_rune_deletions"`
	Calls     int        `json:"predicate_calls"`
	Cache     bool       `json:"reused_revalidated_witness"`
}

func mismatchSignature(name string, a, b WorkAnswer) string {
	if name == "lint/syntax-fixes" {
		a, b = syntaxFixes(a), syntaxFixes(b)
		if strings.Contains(a.Output, "rejected ") || strings.Contains(b.Output, "rejected ") {
			return "fix rejection"
		}
		if strings.Contains(a.Output, "unconverged\t") || strings.Contains(b.Output, "unconverged\t") {
			return "fix convergence"
		}
		return "fixed source"
	}
	if strings.HasPrefix(name, "format/") {
		switch {
		case strings.Contains(a.Output, "\ufeff") && !strings.Contains(b.Output, "\ufeff"):
			return "BOM preservation"
		case strings.Contains(a.Output, "//") && !strings.Contains(b.Output, "//"):
			return "comment-loss line"
		case strings.Contains(a.Output, "/*") && !strings.Contains(b.Output, "/*"):
			return "comment-loss block"
		case a.Output == "ok\t\n":
			return "Go empty"
		case b.Output == "ok\t\n":
			return "Node empty"
		}
		return "formatted text"
	}
	x, y := ruleAnswer(a.execution(), name).Output, ruleAnswer(b.execution(), name).Output
	if x == "" {
		return "Node-only finding"
	}
	if y == "" {
		return "Go-only finding"
	}
	if strings.Count(x, "\nrange ") != strings.Count(y, "\nrange ") {
		return "finding count"
	}
	xx, yy := strings.Split(x, "\n"), strings.Split(y, "\n")
	for i := 0; i < len(xx) && i < len(yy); i++ {
		if xx[i] == yy[i] {
			continue
		}
		if strings.HasPrefix(xx[i], "range ") && strings.HasPrefix(yy[i], "range ") {
			u, v := strings.Split(xx[i], "\t"), strings.Split(yy[i], "\t")
			for j := 0; j < len(u) && j < len(v); j++ {
				if u[j] != v[j] {
					switch j {
					case 0:
						return "range/id/repair"
					case 1:
						return "replacement"
					case 2:
						return "suggestion description"
					default:
						return "edit range"
					}
				}
			}
		}
		if strings.HasPrefix(xx[i], "  ") {
			return "human message"
		}
		if strings.HasPrefix(xx[i], "fix-edit") {
			return "extra fix"
		}
		if strings.HasPrefix(xx[i], "suggestion") {
			return "suggestion"
		}
		return "human location/order"
	}
	return "protocol length"
}
func subsequence(short, long string) bool {
	small := []rune(short)
	i := 0
	for _, r := range long {
		if i < len(small) && r == small[i] {
			i++
		}
	}
	return i == len(small)
}
func site(root, path, anchor string) string {
	data, err := os.ReadFile(filepath.Join(root, path))
	if err != nil {
		return path + ":unresolved"
	}
	at := strings.Index(string(data), anchor)
	if at < 0 {
		return path + ":unresolved"
	}
	return fmt.Sprintf("%s:%d", path, strings.Count(string(data[:at]), "\n")+1)
}
func provenance(d *Driver, name, signature string) (port, goSites []string) {
	if name == "lint/syntax-fixes" {
		return []string{site(d.Root, "stage1/cohere/lint/lint.ts", "fixed():")}, []string{site(d.Root, "cohere/internal/edit/engine.go", "func FixText(")}
	}
	if strings.HasPrefix(name, "format/") {
		switch name {
		case "format/typescript":
			port = []string{site(d.Root, "stage1/cohere/tsprinter/expressions.ts", "export function formatProgram")}
			goSites = []string{site(d.Root, "cohere/internal/format/native/native.go", "hasByteOrderMark :=")}
		case "format/json":
			port = []string{site(d.Root, "stage1/cohere/json/formatter.ts", "export function format(")}
			goSites = []string{site(d.Root, "cohere/internal/format/native/json.go", "Register(\".json\"")}
		case "format/yaml":
			port = []string{site(d.Root, "stage1/cohere/yaml/format.ts", "export function format(")}
			goSites = []string{site(d.Root, "cohere/internal/format/native/yaml.go", "Register(")}
		case "format/css":
			port = []string{site(d.Root, "stage1/cohere/css/print.ts", "export function format(")}
			goSites = []string{site(d.Root, "cohere/internal/format/native/css.go", "Register(")}
		}
		return
	}
	for _, desc := range d.Descriptors {
		if desc.Name != name {
			continue
		}
		path := "stage1/cohere/lint/rules/" + desc.Slug + "/" + desc.Module
		port = append(port, site(d.Root, path, "export class "+desc.Class))
		adapter, _ := os.ReadFile(filepath.Join(d.Root, "stage1/cohere/lint/rules", desc.Slug, "oracle.go"))
		symbolRE := regexp.MustCompile(`return [A-Za-z_]+\.([A-Za-z_]+)`)
		match := symbolRE.FindStringSubmatch(string(adapter))
		if len(match) == 2 {
			files, _ := filepath.Glob(filepath.Join(d.Root, "cohere/internal/lint/rules", desc.UpstreamPackage, "*.go"))
			for _, p := range files {
				if strings.HasSuffix(p, "_test.go") {
					continue
				}
				text, _ := os.ReadFile(p)
				anchor := "var " + match[1] + " ="
				if strings.Contains(string(text), anchor) {
					rel, _ := filepath.Rel(d.Root, p)
					goSites = append(goSites, site(d.Root, rel, anchor))
					for line, s := range strings.Split(string(text), "\n") {
						if strings.Contains(s, "ctx.Report(") || strings.Contains(s, "ctx.ReportRange(") || strings.Contains(s, "ctx.ReportNode(") {
							goSites = append(goSites, fmt.Sprintf("%s:%d", rel, line+1))
						}
					}
				}
			}
		}
		if signature == "human location/order" {
			port = append(port, site(d.Root, "stage1/cohere/lint/main.ts", "const offsets:"))
			goSites = append(goSites, site(d.Root, "cohere/internal/lint/rule/rule.go", "func (diagnostic Diagnostic) Location"))
		}
		return
	}
	return
}
func readReceipts(path string) ([]Receipt, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	g, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	defer g.Close()
	decoder := json.NewDecoder(g)
	var rows []Receipt
	for {
		var lean struct {
			Index       int         `json:"index"`
			Path        string      `json:"path"`
			SHA         string      `json:"source_sha256"`
			Bytes       int         `json:"bytes"`
			GoLint      *WorkAnswer `json:"go_lint"`
			NodeLint    *WorkAnswer `json:"node_lint"`
			GoFormat    WorkAnswer  `json:"go_format"`
			NodeFormat  WorkAnswer  `json:"node_format"`
			Divergences []string    `json:"divergences"`
		}
		err = decoder.Decode(&lean)
		r := Receipt{Index: lean.Index, Path: lean.Path, SHA: lean.SHA, Bytes: lean.Bytes, GoLint: lean.GoLint, NodeLint: lean.NodeLint, GoFormat: lean.GoFormat, NodeFormat: lean.NodeFormat, Divergences: lean.Divergences}
		if err == io.EOF {
			return rows, nil
		}
		if err != nil {
			return nil, err
		}
		r.Divergences = divergences(r)
		if len(r.Divergences) > 0 {
			rows = append(rows, r)
		}
	}
}

func reduceSweep(root, manifest, in, out string, limit time.Duration, watch bool) error {
	m, err := readManifest(manifest)
	if err != nil {
		return err
	}
	if err = m.Dependencies.validate(); err != nil {
		return err
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		return err
	}
	scratch := filepath.Join(out, ".work")
	os.MkdirAll(scratch, 0755)
	d := Driver{Root: root, Scratch: scratch, Limit: limit}
	if err = d.build(); err != nil {
		return err
	}
	goBinary, lintModule, formatModule, err := d.buildWorkers()
	if err != nil {
		return err
	}
	gw := Worker{args: []string{goBinary}, dir: root, Limit: limit}
	nw := Worker{args: []string{"node", "--max-old-space-size=512", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(root, "stage1/cohere/scoreboard/node-worker.mjs"), lintModule, formatModule}, dir: root, Limit: limit}
	defer gw.close()
	defer nw.close()
	cache := map[string][]Reduction{}
	reduced := 0
	for {
		files, _ := filepath.Glob(filepath.Join(in, "receipts-*.jsonl.gz"))
		for _, file := range files {
			destination := filepath.Join(out, "reductions-"+filepath.Base(file))
			var priorRows []Reduction
			present := map[string]bool{}
			if _, err = os.Stat(destination); err == nil {
				old, e := os.Open(destination)
				if e != nil {
					return e
				}
				z, e := gzip.NewReader(old)
				if e != nil {
					return e
				}
				decoder := json.NewDecoder(z)
				for {
					var r Reduction
					e = decoder.Decode(&r)
					if e == io.EOF {
						break
					}
					if e != nil {
						return e
					}
					priorRows = append(priorRows, r)
					present[fmt.Sprintf("%d|%s", r.Index, r.Rule)] = true
					key := r.Rule + "|" + r.Signature + "|" + filepath.Ext(r.File)
					cache[key] = append(cache[key], r)
				}
				z.Close()
				old.Close()
			}
			rows, err := readReceipts(file)
			if err != nil {
				continue
			}
			missing := false
			for _, row := range rows {
				for _, name := range row.Divergences {
					if !present[fmt.Sprintf("%d|%s", row.Index, name)] {
						missing = true
					}
				}
			}
			if !missing {
				continue
			}
			f, err := os.Create(destination + ".tmp")
			if err != nil {
				return err
			}
			g := gzip.NewWriter(f)
			encoder := json.NewEncoder(g)
			for _, r := range priorRows {
				if err = encoder.Encode(r); err != nil {
					return err
				}
			}
			for _, receipt := range rows {
				for _, name := range receipt.Divergences {
					if present[fmt.Sprintf("%d|%s", receipt.Index, name)] {
						continue
					}
					fmt.Fprintf(os.Stderr, "reducing index %d %s %s\n", receipt.Index, name, receipt.Path)
					source, err := inputBytes(m, receipt.Path)
					if err != nil {
						return err
					}
					if hash(source) != receipt.SHA {
						return fmt.Errorf("divergent source changed: %s", receipt.Path)
					}
					a, b := receipt.GoFormat, receipt.NodeFormat
					if !strings.HasPrefix(name, "format/") {
						a, b = *receipt.GoLint, *receipt.NodeLint
					}
					signature := mismatchSignature(name, a, b)
					q := WorkRequest{Path: receipt.Path, Rule: "all", Family: formatFamily(receipt.Path)}
					if strings.HasPrefix(name, "format/") {
						q.Op = "format"
					} else {
						q.Op = "lint"
					}
					key := name + "|" + signature + "|" + filepath.Ext(receipt.Path)
					calls := 0
					var lastA, lastB WorkAnswer
					predicate := func(text string) bool {
						calls++
						q.Source = []byte(text)
						if lintExtension(q.Path) {
							valid := q
							valid.Op = "valid"
							if v := gw.call(valid); v.Error != "" || !v.Valid {
								return false
							}
						}
						x, y := gw.call(q), nw.call(q)
						u, v := comparison(x, name), comparison(y, name)
						status, _ := stateCode(u, v, q.Op == "format")
						if status != "diverge" || mismatchSignature(name, x, y) != signature {
							return false
						}
						lastA, lastB = x, y
						return true
					}
					if !predicate(string(source)) {
						return fmt.Errorf("original divergence does not replay: %s %s", receipt.Path, name)
					}
					input := ""
					reused := false
					for _, prior := range cache[key] {
						if subsequence(prior.Input, string(source)) && predicate(prior.Input) {
							input = prior.Input
							reused = true
							break
						}
					}
					if !reused && name == "lint/syntax-fixes" {
						// Any previously proved witness may seed a syntax-fix
						// mismatch, including one just reduced for another rule.
						// Same-path replay, signature and source subsequence are
						// still required; no witness substitutes for this check.
						var candidates []string
						unique := map[string]bool{}
						for _, rows := range cache {
							for _, prior := range rows {
								if !unique[prior.Input] {
									candidates = append(candidates, prior.Input)
									unique[prior.Input] = true
								}
							}
						}
						sort.Slice(candidates, func(i, j int) bool {
							if len(candidates[i]) != len(candidates[j]) {
								return len(candidates[i]) < len(candidates[j])
							}
							return candidates[i] < candidates[j]
						})
						for _, candidate := range candidates {
							if subsequence(candidate, string(source)) && predicate(candidate) {
								input = candidate
								reused = true
								break
							}
						}
					}
					if !reused {
						input = string(source)
						for _, seed := range []string{"\ufeff", "//", "/* */", "// \n0", "0// ", "\"\"// ", "{}// ", "{// \n}", "// \n\"\"", "#!\n//\n//\n//\n//\n//"} {
							if subsequence(seed, input) && predicate(seed) {
								input = seed
								break
							}
						}
						if lintExtension(q.Path) {
							for {
								spansQ := q
								spansQ.Op = "spans"
								spansQ.Source = []byte(input)
								answer := gw.call(spansQ)
								if answer.Error != "" || !answer.Valid {
									break
								}
								sort.Slice(answer.Spans, func(i, j int) bool {
									return answer.Spans[i][1]-answer.Spans[i][0] > answer.Spans[j][1]-answer.Spans[j][0]
								})
								changed := false
								for _, span := range answer.Spans {
									candidate := input[:span[0]] + input[span[1]:]
									if predicate(candidate) {
										input = candidate
										changed = true
										break
									}
								}
								if !changed {
									break
								}
							}
						}
						input = minimize(input, predicate)
					}
					if !predicate(input) {
						return fmt.Errorf("reduction no longer diverges: %s %s", receipt.Path, name)
					}
					// Context can change a cached witness: recheck every one-rune deletion here.
					chars := []rune(input)
					verified := true
					for i := range chars {
						candidate := string(append(append([]rune{}, chars[:i]...), chars[i+1:]...))
						if predicate(candidate) {
							verified = false
							break
						}
					}
					if !verified {
						input = minimize(input, predicate)
						if !predicate(input) {
							return fmt.Errorf("failed context-specific re-reduction")
						}
						verified = true
					}
					if !predicate(input) {
						return fmt.Errorf("final reduction lost divergence")
					}
					port, goSites := provenance(&d, name, signature)
					r := Reduction{Index: receipt.Index, File: receipt.Path, Rule: name, Signature: signature, SourceSHA: receipt.SHA, Input: input, Go: lastA, Node: lastB, PortSites: port, GoSites: goSites, Verified: verified, Calls: calls, Cache: reused}
					if len(port) == 0 || len(goSites) == 0 {
						return fmt.Errorf("missing responsible implementation sites: %s", name)
					}
					if err = encoder.Encode(r); err != nil {
						return err
					}
					if !reused {
						cache[key] = append(cache[key], r)
					}
					reduced++
					if reduced%20 == 0 {
						fmt.Fprintf(os.Stderr, "reduced %d divergences; latest %s (%d runes)\n", reduced, name, len([]rune(input)))
					}
				}
			}
			if err = g.Close(); err != nil {
				return err
			}
			if err = f.Close(); err != nil {
				return err
			}
			if err = os.Rename(destination+".tmp", destination); err != nil {
				return err
			}
		}
		data, _ := os.ReadFile(filepath.Join(in, "summary.json"))
		var status Sweep
		json.Unmarshal(data, &status)
		if !watch || status.Complete {
			break
		}
		time.Sleep(20 * time.Second)
	}
	fmt.Fprintf(os.Stderr, "reduced %d divergences from available completed receipt chunks\n", reduced)
	return nil
}
