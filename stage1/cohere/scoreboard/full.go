package main

import (
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Receipt struct {
	Index       int         `json:"index"`
	Path        string      `json:"path"`
	SHA         string      `json:"source_sha256"`
	Bytes       int         `json:"bytes"`
	Lint        string      `json:"lint_states,omitempty"`
	Format      string      `json:"format_state"`
	GoLint      *WorkAnswer `json:"go_lint,omitempty"`
	NodeLint    *WorkAnswer `json:"node_lint,omitempty"`
	GoFormat    WorkAnswer  `json:"go_format"`
	NodeFormat  WorkAnswer  `json:"node_format"`
	Census      []Missing   `json:"census,omitempty"`
	Divergences []string    `json:"divergences,omitempty"`
}
type RuleTotal struct {
	OracleFiles int `json:"oracle_finding_files"`
	Total
	NodeFindings   int            `json:"node_findings"`
	Causes         map[string]int `json:"blocked_causes,omitempty"`
	OracleStatuses map[string]int `json:"oracle_census_statuses,omitempty"`
}
type Sweep struct {
	Complete     bool                 `json:"complete"`
	Dependencies DependencyInputs     `json:"dependency_inputs"`
	Corpus       string               `json:"corpus"`
	Files        int                  `json:"files"`
	LintFiles    int                  `json:"lint_files"`
	Bytes        int64                `json:"bytes"`
	Base         string               `json:"base"`
	Cohere       string               `json:"cohere"`
	PortRules    []string             `json:"port_rule_order"`
	PerRule      map[string]RuleTotal `json:"per_rule"`
	PerFamily    map[string]RuleTotal `json:"per_family"`
	Blocked      map[string]int       `json:"blocked_cells_by_cause"`
	Missing      []Missing            `json:"unported_ranked"`
	Total        Total                `json:"total"`
}

func stateCode(a, b WorkAnswer, format bool) (string, string) {
	if a.Error != "" || b.Error != "" || a.Exit != 0 || b.Exit != 0 {
		cause := "execution fault"

		if strings.Contains(a.Error, "invalid corpus") || strings.Contains(b.Stack, "typescript/parser") || strings.Contains(b.Stack, "typescript/scanner") || b.Phase == "parser" {
			cause = "parser"
		}
		return "blocked", cause
	}
	if !format {
		if hasLinePrefix(a.Output, "skipped ") || hasLinePrefix(b.Output, "skipped ") {
			return "blocked", "checker fact"
		}
	}
	if strings.HasPrefix(a.Output, "error\t") || strings.HasPrefix(b.Output, "error\t") || strings.HasPrefix(b.Output, "refused ") {
		return "blocked", "formatter refusal"
	}
	if a.Output == b.Output {
		return "agree", ""
	}
	return "diverge", ""
}
func stateLetter(status, cause string) string {
	switch status {
	case "agree":
		return "A"
	case "diverge":
		return "D"
	}
	switch cause {
	case "checker fact":
		return "C"
	case "parser":
		return "P"
	case "formatter refusal":
		return "F"
	case "unported rule":
		return "U"
	}
	return "E"
}
func bump(t RuleTotal, status, cause string, goN, nodeN int) RuleTotal {
	t.Checks++
	t.Findings += goN
	t.NodeFindings += nodeN
	switch status {
	case "agree":
		t.Agree++
	case "diverge":
		t.Diverge++
	default:
		t.Blocked++
		if t.Causes == nil {
			t.Causes = map[string]int{}
		}
		t.Causes[cause]++
	}
	return t
}
func readManifest(path string) (m Manifest, err error) {
	f, e := os.Open(path)
	if e != nil {
		return m, e
	}
	defer f.Close()
	var r io.Reader = f
	if strings.HasSuffix(path, ".gz") {
		g, e := gzip.NewReader(f)
		if e != nil {
			return m, e
		}
		defer g.Close()
		r = g
	}
	err = json.NewDecoder(r).Decode(&m)
	return
}
func inputBytes(m Manifest, path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil {
		return nil, err
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return os.ReadFile(path)
	}
	for _, snap := range m.Snapshots {
		if strings.HasPrefix(path, snap.Root+"/") {
			rel, err := filepath.Rel(snap.Root, path)
			if err != nil {
				return nil, err
			}
			text, err := checked(snap.Root, "git", "show", snap.SHA+":"+filepath.ToSlash(rel))
			return []byte(text), err
		}
	}
	return nil, fmt.Errorf("symlink not in an identified snapshot: %s", path)
}
func ruleFamily(name string) string {
	if strings.HasPrefix(name, "@next/next/") {
		return "next"
	}
	if strings.HasPrefix(name, "@typescript-eslint/") {
		return "typescript"
	}
	return family(name)
}
func sweep(root, manifest, out string, workers, maxFiles int, limit time.Duration) error {
	if workers < 1 || workers > 16 {
		return fmt.Errorf("workers must be 1 through 16")
	}
	m, err := readManifest(manifest)
	if err != nil {
		return err
	}
	if err = m.Dependencies.validate(); err != nil {
		return err
	}
	if m.Program != "" || len(m.Options) > 0 || len(m.Rules) > 0 {
		return fmt.Errorf("full-tree worker currently requires default syntax-only options; checker recordings and per-rule option transport are separate inputs")
	}
	if err = os.MkdirAll(out, 0755); err != nil {
		return err
	}
	scratch := filepath.Join(out, ".work")
	if err = os.MkdirAll(scratch, 0755); err != nil {
		return err
	}
	d := Driver{Root: root, Scratch: scratch, Limit: limit}
	if err = d.build(); err != nil {
		return err
	}
	goBinary, lintModule, formatModule, err := d.buildWorkers()
	if err != nil {
		return err
	}
	// Validate identified pins without following tracked symlinks as source files.
	seen := map[string]bool{}
	for _, path := range m.Files {
		if seen[path] {
			return fmt.Errorf("duplicate source %s", path)
		}
		seen[path] = true
		if strings.ContainsAny(path, "\t\r\n") {
			return fmt.Errorf("delimited source path %q", path)
		}
	}
	for _, snap := range m.Snapshots {
		got, err := checked(snap.Root, "git", "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		if strings.TrimSpace(got) != snap.SHA {
			return fmt.Errorf("pin differs: %s", snap.Repository)
		}
	}
	base, _ := checked(root, "git", "rev-parse", "HEAD")
	cohere, _ := checked(filepath.Join(root, "cohere"), "git", "rev-parse", "HEAD")
	summary := Sweep{Dependencies: m.Dependencies, Corpus: m.Name, Base: strings.TrimSpace(base), Cohere: strings.TrimSpace(cohere), PerRule: map[string]RuleTotal{}, PerFamily: map[string]RuleTotal{}, Blocked: map[string]int{}}
	known := map[string]bool{}
	for _, desc := range d.Descriptors {
		summary.PortRules = append(summary.PortRules, desc.Name)
		known[desc.Name] = true
	}
	n := len(m.Files)
	if maxFiles > 0 && maxFiles < n {
		n = maxFiles
	}
	var next, done atomic.Int64
	var lock sync.Mutex
	var firstError error
	record := func(name, fam, status, cause string, goN, nodeN int) {
		summary.PerRule[name] = bump(summary.PerRule[name], status, cause, goN, nodeN)
		summary.PerFamily[fam] = bump(summary.PerFamily[fam], status, cause, goN, nodeN)
		if status == "blocked" {
			summary.Blocked[cause]++
		}
	}
	account := func(r Receipt) {
		lock.Lock()
		defer lock.Unlock()
		if r.GoLint != nil {
			a, b := *r.GoLint, *r.NodeLint
			for _, desc := range d.Descriptors {
				x, y := a, b
				x.Output = ruleAnswer(a.execution(), desc.Name).Output
				y.Output = ruleAnswer(b.execution(), desc.Name).Output
				status, cause := stateCode(x, y, false)
				record(desc.Name, desc.UpstreamPackage, status, cause, strings.Count(x.Output, "\nrange "), strings.Count(y.Output, "\nrange "))
			}
			status, cause := stateCode(a, b, false)
			record("lint/all-fixes", "fixes", status, cause, 0, 0)
			status, cause = stateCode(syntaxFixes(a), syntaxFixes(b), false)
			record("lint/syntax-fixes", "fixes", status, cause, 0, 0)
			summary.LintFiles++
			for _, row := range r.Census {
				if known[row.Rule] {
					continue
				}
				record(row.Rule, ruleFamily(row.Rule), "blocked", "unported rule", row.Findings, 0)
				t := summary.PerRule[row.Rule]
				if t.OracleStatuses == nil {
					t.OracleStatuses = map[string]int{}
				}
				t.OracleStatuses[row.Status]++
				t.OracleFiles += row.Files
				summary.PerRule[row.Rule] = t
			}
		}
		status, cause := stateCode(r.GoFormat, r.NodeFormat, true)
		record("format/"+formatFamily(r.Path), "format", status, cause, 0, 0)
		summary.Files++
		summary.Bytes += int64(r.Bytes)
	}
	completed := map[int]bool{}
	oldChunks, _ := filepath.Glob(filepath.Join(out, "receipts-*.jsonl.gz"))
	for _, chunkPath := range oldChunks {
		f, e := os.Open(chunkPath)
		if e != nil {
			return e
		}
		g, e := gzip.NewReader(f)
		if e != nil {
			f.Close()
			return e
		}
		data, e := io.ReadAll(g)
		g.Close()
		f.Close()
		if e != nil {
			if e = os.Rename(chunkPath, chunkPath+".partial"); e != nil {
				return e
			}
			continue
		}
		decoder := json.NewDecoder(strings.NewReader(string(data)))
		for {
			var r Receipt
			e = decoder.Decode(&r)
			if e == io.EOF {
				break
			}
			if e != nil {
				return e
			}
			if r.Index < 0 || r.Index >= n || m.Files[r.Index] != r.Path || completed[r.Index] {
				return fmt.Errorf("invalid or duplicate saved receipt %s index %d", chunkPath, r.Index)
			}
			source, e := inputBytes(m, r.Path)
			if e != nil {
				return e
			}
			if hash(source) != r.SHA {
				return fmt.Errorf("saved source changed: %s", r.Path)
			}
			completed[r.Index] = true
			account(r)
			done.Add(1)
		}
	}
	saveSummary := func(complete bool) error {
		lock.Lock()
		defer lock.Unlock()
		summary.Complete = complete
		summary.Total = Total{}
		summary.Missing = nil
		for name, t := range summary.PerRule {
			summary.Total.Checks += t.Checks
			summary.Total.Agree += t.Agree
			summary.Total.Diverge += t.Diverge
			summary.Total.Blocked += t.Blocked
			summary.Total.Findings += t.Findings
			if !known[name] && !strings.HasPrefix(name, "format/") && !strings.HasPrefix(name, "lint/") {
				status := "complete"
				for s := range t.OracleStatuses {
					if s != "complete" {
						status = "incomplete"
					}
				}
				summary.Missing = append(summary.Missing, Missing{Rule: name, Findings: t.Findings, Files: t.OracleFiles, Status: status})
			}
		}
		sort.Slice(summary.Missing, func(i, j int) bool {
			if summary.Missing[i].Findings != summary.Missing[j].Findings {
				return summary.Missing[i].Findings > summary.Missing[j].Findings
			}
			return summary.Missing[i].Rule < summary.Missing[j].Rule
		})
		data, e := json.MarshalIndent(summary, "", "  ")
		if e != nil {
			return e
		}
		return os.WriteFile(filepath.Join(out, "summary.json"), append(data, '\n'), 0644)
	}
	var group sync.WaitGroup
	for workerIndex := 0; workerIndex < workers; workerIndex++ {
		group.Add(1)
		go func(id int) {
			defer group.Done()
			gw := Worker{args: []string{goBinary}, dir: root, Limit: limit}
			nw := Worker{args: []string{"node", "--max-old-space-size=512", "--disable-warning=ExperimentalWarning", filepath.Join(root, "oracle/node.mjs"), filepath.Join(root, "stage1/cohere/scoreboard/node-worker.mjs"), lintModule, formatModule}, dir: root, Limit: limit}
			defer gw.close()
			defer nw.close()
			var file *os.File
			var gz *gzip.Writer
			var encoder *json.Encoder
			chunk := len(oldChunks) + 1
			count := 0
			closeChunk := func() {
				if gz != nil {
					if e := gz.Close(); e != nil {
						lock.Lock()
						firstError = e
						lock.Unlock()
					}
					file.Close()
					gz = nil
				}
			}
			defer closeChunk()
			for {
				index := int(next.Add(1) - 1)
				if index >= n {
					return
				}
				if completed[index] {
					continue
				}
				path := m.Files[index]
				source, e := inputBytes(m, path)
				if e != nil {
					lock.Lock()
					firstError = e
					lock.Unlock()
					return
				}
				repoRoot := root
				for _, snap := range m.Snapshots {
					if strings.HasPrefix(path, snap.Root+"/") {
						repoRoot = snap.Root
						break
					}
				}
				q := WorkRequest{Path: path, Root: repoRoot, Rule: "all", Source: source}
				r := Receipt{Index: index, Path: path, SHA: hash(source), Bytes: len(source)}
				if lintExtension(path) {
					q.Op = "lint"
					a := gw.call(q)
					b := nw.call(q)
					r.GoLint = &a
					r.NodeLint = &b
					for _, desc := range d.Descriptors {
						x := a
						x.Output = ruleAnswer(a.execution(), desc.Name).Output
						y := b
						y.Output = ruleAnswer(b.execution(), desc.Name).Output
						status, cause := stateCode(x, y, false)

						r.Lint += stateLetter(status, cause)
						if status == "diverge" {
							r.Divergences = append(r.Divergences, desc.Name)
						}

					}
					status, _ := stateCode(a, b, false)
					if status == "diverge" {
						r.Divergences = append(r.Divergences, "lint/all-fixes")
					}
					q.Op = "census"
					c := gw.call(q)
					if c.Error != "" {
						lock.Lock()
						firstError = fmt.Errorf("census %s: %s", path, c.Error)
						lock.Unlock()
						return
					}
					r.Census = c.Rows

				}
				q.Op = "format"
				q.Family = formatFamily(path)
				r.GoFormat = gw.call(q)
				r.NodeFormat = nw.call(q)
				status, cause := stateCode(r.GoFormat, r.NodeFormat, true)
				r.Format = stateLetter(status, cause)
				if status == "diverge" {
					r.Divergences = append(r.Divergences, "format/"+q.Family)
				}
				if gz == nil {
					p := filepath.Join(out, fmt.Sprintf("receipts-%02d-%04d.jsonl.gz", id, chunk))
					file, e = os.Create(p)
					if e != nil {
						lock.Lock()
						firstError = e
						lock.Unlock()
						return
					}
					gz = gzip.NewWriter(file)
					encoder = json.NewEncoder(gz)
					chunk++
					count = 0
				}
				if e = encoder.Encode(r); e != nil {
					lock.Lock()
					firstError = e
					lock.Unlock()
					return
				}
				account(r)
				count++
				if count == 1024 {
					closeChunk()
				}
				done.Add(1)
			}
		}(workerIndex)
	}
	finished := make(chan struct{})
	go func() { group.Wait(); close(finished) }()
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(os.Stderr, "scored %d/%d files\n", done.Load(), n)
			if err = saveSummary(false); err != nil {
				return err
			}
		case <-finished:
			if firstError != nil {
				return firstError
			}
			if err = saveSummary(n == len(m.Files) && int(done.Load()) == n); err != nil {
				return err
			}
			return nil
		}
	}
}

func hasLinePrefix(text, prefix string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, prefix) {
			return true
		}
	}
	return false
}
