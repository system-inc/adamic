// Package main builds an evidence scoreboard; it does not implement lint rules.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

type Snapshot struct {
	Repository string `json:"repository"`
	Root       string `json:"root"`
	SHA        string `json:"sha"`
}

type Manifest struct {
	Snapshots []Snapshot                 `json:"snapshots,omitempty"`
	Name      string                     `json:"name"`
	Files     []string                   `json:"files"`
	Program   string                     `json:"program,omitempty"`
	Rules     []string                   `json:"rules,omitempty"`
	Options   map[string]json.RawMessage `json:"options,omitempty"`
}
type Execution struct {
	Output string `json:"stdout"`
	Error  string `json:"error,omitempty"`
	Stderr string `json:"stderr,omitempty"`
}
type Cell struct {
	ReducedGo   *Execution `json:"reduced_go,omitempty"`
	ReducedNode *Execution `json:"reduced_node,omitempty"`
	File        string     `json:"file"`
	Rule        string     `json:"rule"`
	Family      string     `json:"family"`
	SHA256      string     `json:"sha256"`
	Status      string     `json:"status"`
	Go          Execution  `json:"go"`
	Node        Execution  `json:"node"`
	Findings    int        `json:"go_findings"`
	Input       string     `json:"input,omitempty"`
	Reduced     string     `json:"deletion_minimal_input,omitempty"`
	Reduction   string     `json:"reduction,omitempty"`
}
type Total struct {
	Checks   int `json:"checks"`
	Agree    int `json:"agree"`
	Diverge  int `json:"diverge"`
	Blocked  int `json:"blocked"`
	Findings int `json:"go_findings"`
}
type Missing struct {
	Rule     string `json:"rule"`
	Findings int    `json:"findings"`
	Files    int    `json:"files"`
	Status   string `json:"status"`
}
type Report struct {
	Corpus     string           `json:"corpus"`
	Canonical  bool             `json:"canonical_quiet_hundred"`
	Base       string           `json:"base"`
	Cohere     string           `json:"cohere"`
	Files      int              `json:"files"`
	Registered int              `json:"registered_port_rules"`
	Cells      []Cell           `json:"cells"`
	PerRule    map[string]Total `json:"per_rule"`
	PerFamily  map[string]Total `json:"per_family"`
	PerFile    map[string]Total `json:"per_file"`
	Missing    []Missing        `json:"unported_ranked"`
	Blockers   []string         `json:"blockers,omitempty"`
}

func family(name string) string {
	if i := strings.Index(name, "/"); i >= 0 {
		return name[:i]
	}
	return "core"
}
func hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func execute(ctx context.Context, dir, name string, args ...string) Execution {
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	var out, errout bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errout
	err := command.Run()
	result := Execution{Output: out.String(), Stderr: errout.String()}
	if err != nil {
		result.Error = err.Error()
	}
	return result
}
func checked(dir, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()
	r := execute(ctx, dir, name, args...)
	if r.Error != "" {
		return "", fmt.Errorf("%s: %s\n%s", name, r.Error, r.Stderr)
	}
	return r.Output, nil
}
func blocked(e Execution) bool {
	if e.Error != "" {
		return true
	}
	for _, line := range strings.Split(e.Output, "\n") {
		if strings.HasPrefix(line, "skipped ") || strings.HasPrefix(line, "refused ") {
			return true
		}
	}
	return false
}
func classify(a, b Execution) string {
	if blocked(a) || blocked(b) {
		return "blocked"
	}
	if a.Output == b.Output {
		return "agree"
	}
	return "diverge"
}
func aggregate(cells []Cell, key func(Cell) string) map[string]Total {
	result := map[string]Total{}
	for _, c := range cells {
		k := key(c)
		t := result[k]
		t.Checks++
		t.Findings += c.Findings
		switch c.Status {
		case "agree":
			t.Agree++
		case "diverge":
			t.Diverge++
		default:
			t.Blocked++
		}
		result[k] = t
	}
	return result
}
func summarize(r *Report) {
	r.PerRule = aggregate(r.Cells, func(c Cell) string { return c.Rule })
	r.PerFamily = aggregate(r.Cells, func(c Cell) string { return c.Family })
	r.PerFile = aggregate(r.Cells, func(c Cell) string { return c.File })
}
func validate(m Manifest, canonical bool, root string) error {
	if m.Name == "" || len(m.Files) == 0 {
		return fmt.Errorf("manifest needs name and files")
	}
	if canonical && len(m.Snapshots) != 100 {
		return fmt.Errorf("canonical quiet hundred requires 100 identified snapshots; got %d; use --fixtures for available subsets", len(m.Snapshots))
	}
	for _, snapshot := range m.Snapshots {
		if snapshot.Repository == "" || snapshot.Root == "" || len(snapshot.SHA) != 40 {
			return fmt.Errorf("incomplete snapshot identity: %s", snapshot.Repository)
		}
		actual, err := checked(snapshot.Root, "git", "rev-parse", "HEAD")
		if err != nil {
			return err
		}
		if strings.TrimSpace(actual) != snapshot.SHA {
			return fmt.Errorf("snapshot pin differs: %s", snapshot.Repository)
		}
	}
	seen := map[string]bool{}
	for _, p := range m.Files {
		abs := p
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(root, p)
		}
		abs = filepath.Clean(abs)
		if seen[abs] {
			return fmt.Errorf("duplicate source %s", p)
		}
		seen[abs] = true
		if strings.ContainsAny(p, "\t\r\n") {
			return fmt.Errorf("manifest delimiter in path %q", p)
		}
		if _, err := os.ReadFile(abs); err != nil {
			return err
		}
	}
	return nil
}

// minimize reaches a single-rune-deletion fixed point. It claims no global shortest program.
// The predicate must require two successful runs and preserve the divergence category.
func minimize(input string, fails func(string) bool) string {
	chars := []rune(input)
	for width := len(chars) / 2; width >= 1; width /= 2 {
		for start := 0; start+width <= len(chars); {
			next := append(append([]rune{}, chars[:start]...), chars[start+width:]...)
			if fails(string(next)) {
				chars = next
				start = 0
			} else {
				start++
			}
		}
	}
	return string(chars)
}

type Driver struct {
	Root, Scratch, Oracle, Catalog, Native string
	Descriptors                            []registry.Descriptor
	Limit                                  time.Duration
}

func (d *Driver) build() error {
	lint := filepath.Join(d.Root, "stage1/cohere/lint")
	descriptors, err := registry.Generate(lint)
	if err != nil {
		return err
	}
	d.Descriptors = descriptors
	replacements := map[string]string{}
	var files []string
	add := func(name, path string) {
		virtual := filepath.Join(d.Root, "cohere", "scoreboard_"+name+".go")
		replacements[virtual] = path
		files = append(files, virtual)
	}
	add("oracle", filepath.Join(lint, "testdata/oracle.go"))
	add("registry", filepath.Join(lint, ".generated/registry.go"))
	for _, desc := range descriptors {
		add(strings.ReplaceAll(desc.Slug, "-", "_"), filepath.Join(lint, "rules", desc.Slug, "oracle.go"))
	}
	overlay := filepath.Join(d.Scratch, "overlay.json")
	data, _ := json.Marshal(map[string]any{"Replace": replacements})
	if err = os.WriteFile(overlay, data, 0600); err != nil {
		return err
	}
	d.Oracle = filepath.Join(d.Scratch, "oracle")
	args := append([]string{"build", "-overlay=" + overlay, "-o", d.Oracle}, files...)
	if _, err = checked(filepath.Join(d.Root, "cohere"), "go", args...); err != nil {
		return err
	}
	virtual := filepath.Join(d.Root, "cohere", "scoreboard_catalog.go")
	data, _ = json.Marshal(map[string]any{"Replace": map[string]string{virtual: filepath.Join(d.Root, "stage1/cohere/scoreboard/testdata/catalog.go")}})
	if err = os.WriteFile(overlay, data, 0600); err != nil {
		return err
	}
	d.Catalog = filepath.Join(d.Scratch, "catalog")
	_, err = checked(filepath.Join(d.Root, "cohere"), "go", "build", "-overlay="+overlay, "-o", d.Catalog, virtual)
	return err
}
func (d *Driver) pair(path, rule, options, program string) (Execution, Execution) {
	row := path + "\t" + rule + "\t\t\t\t" + options + "\n"
	if program != "" {
		row = "program " + program + "\n" + row
	}
	manifest := filepath.Join(d.Scratch, "row.txt")
	if err := os.WriteFile(manifest, []byte(row), 0600); err != nil {
		return Execution{Error: err.Error()}, Execution{Error: err.Error()}
	}
	ctx, cancel := context.WithTimeout(context.Background(), d.Limit)
	a := execute(ctx, d.Root, d.Oracle, "--manifest", manifest)
	cancel()
	args := []string{"--disable-warning=ExperimentalWarning", filepath.Join(d.Root, "oracle/node.mjs"), filepath.Join(d.Root, "stage1/cohere/lint/main.ts"), "--manifest", manifest}
	if program != "" {
		if d.Native == "" {
			return a, Execution{Error: "typed Node requires --native to record checker facts"}
		}
		prefix := filepath.Join(d.Scratch, "transcript")
		ctx, cancel = context.WithTimeout(context.Background(), d.Limit)
		record := execute(ctx, d.Root, d.Native, "--manifest", manifest, "--record", prefix)
		cancel()
		if blocked(record) {
			return a, record
		}
		args = append(args, "--replay", prefix)
	}
	ctx, cancel = context.WithTimeout(context.Background(), d.Limit)
	b := execute(ctx, d.Root, "node", args...)
	cancel()
	return a, b
}
func (d *Driver) run(m Manifest, canonical, reduce bool) (Report, error) {
	if err := validate(m, canonical, d.Root); err != nil {
		return Report{}, err
	}
	r := Report{Corpus: m.Name, Canonical: canonical, Files: len(m.Files), Registered: len(d.Descriptors)}
	r.Base, _ = checked(d.Root, "git", "rev-parse", "HEAD")
	r.Cohere, _ = checked(filepath.Join(d.Root, "cohere"), "git", "rev-parse", "HEAD")
	r.Base = strings.TrimSpace(r.Base)
	r.Cohere = strings.TrimSpace(r.Cohere)
	selected := map[string]bool{}
	for _, name := range m.Rules {
		selected[name] = true
	}
	known := map[string]bool{}
	for _, desc := range d.Descriptors {
		known[desc.Name] = true
	}
	for name := range selected {
		if !known[name] {
			return r, fmt.Errorf("unknown port rule %s", name)
		}
	}
	for name := range m.Options {
		if !known[name] {
			return r, fmt.Errorf("options name unknown port rule %s; unported external census options need their own manifest", name)
		}
	}
	for _, path := range m.Files {
		abs := path
		if !filepath.IsAbs(abs) {
			abs = filepath.Join(d.Root, path)
		}
		source, err := os.ReadFile(abs)
		if err != nil {
			return r, err
		}
		if lintExtension(abs) {
			if len(selected) == 0 && len(m.Options) == 0 {
				a, b := d.pair(abs, "all", "", m.Program)
				for _, desc := range d.Descriptors {
					x, y := ruleAnswer(a, desc.Name), ruleAnswer(b, desc.Name)
					c := Cell{File: path, Rule: desc.Name, Family: desc.UpstreamPackage, SHA256: hash(source), Status: classify(x, y), Go: x, Node: y, Findings: strings.Count(x.Output, "\nrange ")}
					if c.Status != "agree" {
						c.Input = string(source)
						c.Reduction = "not reduced: full registry run"
						if reduce && c.Status == "diverge" && m.Program == "" {
							probe := filepath.Join(d.Scratch, filepath.Base(abs))
							c.Reduced = minimize(string(source), func(s string) bool {
								if os.WriteFile(probe, []byte(s), 0600) != nil {
									return false
								}
								x, y := d.pair(probe, desc.Name, "", "")
								return classify(ruleAnswer(x, desc.Name), ruleAnswer(y, desc.Name)) == "diverge"
							})
							c.Reduction = "single-rune deletion fixed point; relocated path; not globally shortest"
							if err := os.WriteFile(probe, []byte(c.Reduced), 0600); err != nil {
								return r, err
							}
							rx, ry := d.pair(probe, desc.Name, "", "")
							c.ReducedGo = &rx
							c.ReducedNode = &ry
						}
					}
					r.Cells = append(r.Cells, c)
				}
				c := Cell{File: path, Rule: "lint/all-fixes", Family: "fixes", SHA256: hash(source), Status: classify(a, b), Go: a, Node: b, Findings: strings.Count(a.Output, "\nrange ")}
				if c.Status != "agree" {
					c.Input = string(source)
				}
				r.Cells = append(r.Cells, c)
			} else {
				for _, desc := range d.Descriptors {
					if len(selected) > 0 && !selected[desc.Name] {
						continue
					}
					options := string(m.Options[desc.Name])
					a, b := d.pair(abs, desc.Name, options, m.Program)
					c := Cell{File: path, Rule: desc.Name, Family: desc.UpstreamPackage, SHA256: hash(source), Status: classify(a, b), Go: a, Node: b, Findings: strings.Count(a.Output, "\nrange ")}
					if c.Status != "agree" {
						c.Input = string(source)
						if reduce && c.Status == "diverge" && m.Program == "" {
							// Preserve script kind and basename. Path-sensitive reductions are not relocated silently.
							probe := filepath.Join(d.Scratch, filepath.Base(abs))
							if probe == abs {
								return r, fmt.Errorf("probe aliases input")
							}
							c.Reduced = minimize(string(source), func(s string) bool {
								if os.WriteFile(probe, []byte(s), 0600) != nil {
									return false
								}
								x, y := d.pair(probe, desc.Name, options, "")
								return classify(x, y) == "diverge"
							})
							c.Reduction = "single-rune deletion fixed point; relocated path; not globally shortest"
							if err := os.WriteFile(probe, []byte(c.Reduced), 0600); err != nil {
								return r, err
							}
							rx, ry := d.pair(probe, desc.Name, "", "")
							c.ReducedGo = &rx
							c.ReducedNode = &ry
						} else {
							c.Reduction = "not reduced: process/coverage failure or typed program"
						}
					}
					r.Cells = append(r.Cells, c)
				}
			}
		}
		a, b := d.format(abs)
		c := Cell{File: path, Rule: "format/" + formatFamily(abs), Family: "format", SHA256: hash(source), Go: a, Node: b, Status: classify(a, b)}
		if c.Status != "agree" {
			c.Input = string(source)
			c.Reduction = "not reduced"
			if reduce && c.Status == "diverge" {
				probe := filepath.Join(d.Scratch, filepath.Base(abs))
				c.Reduced = minimize(string(source), func(s string) bool {
					if os.WriteFile(probe, []byte(s), 0600) != nil {
						return false
					}
					if lintExtension(probe) {
						ctx, cancel := context.WithTimeout(context.Background(), d.Limit)
						valid := execute(ctx, d.Root, d.Catalog, "--valid", probe)
						cancel()
						if valid.Error != "" {
							return false
						}
					}
					x, y := d.format(probe)
					if strings.Contains(a.Output, "\ufeff") && !strings.Contains(b.Output, "\ufeff") {
						if !strings.Contains(x.Output, "\ufeff") || strings.Contains(y.Output, "\ufeff") {
							return false
						}
					}
					if strings.Contains(a.Output, "//") && !strings.Contains(b.Output, "//") {
						if !strings.Contains(x.Output, "//") || strings.Contains(y.Output, "//") {
							return false
						}
					}
					return classify(x, y) == "diverge"
				})
				c.Reduction = "single-rune deletion fixed point; relocated path; not globally shortest"
				if err := os.WriteFile(probe, []byte(c.Reduced), 0600); err != nil {
					return r, err
				}
				rx, ry := d.format(probe)
				c.ReducedGo = &rx
				c.ReducedNode = &ry
			}
		}
		r.Cells = append(r.Cells, c)
	}
	// Ranking is an independent unmodified Go registry census, not the port's rule list.
	censusManifest := filepath.Join(d.Scratch, "census.json")
	encoded, _ := json.Marshal(m)
	if err := os.WriteFile(censusManifest, encoded, 0600); err != nil {
		return r, err
	}
	output, err := checked(d.Root, d.Catalog, censusManifest, d.Root)
	if err != nil {
		r.Blockers = append(r.Blockers, err.Error())
	} else {
		var census []Missing
		if err = json.Unmarshal([]byte(output), &census); err != nil {
			return r, err
		}
		for _, item := range census {
			if !known[item.Rule] {
				r.Missing = append(r.Missing, item)
			}
		}
		sort.Slice(r.Missing, func(i, j int) bool {
			if r.Missing[i].Findings != r.Missing[j].Findings {
				return r.Missing[i].Findings > r.Missing[j].Findings
			}
			return r.Missing[i].Rule < r.Missing[j].Rule
		})
	}
	summarize(&r)
	return r, nil
}
func lintExtension(path string) bool {
	switch filepath.Ext(strings.TrimSuffix(path, ".txt")) {
	case ".ts", ".tsx", ".js", ".jsx", ".a":
		return true
	}
	return false
}
func formatFamily(path string) string {
	switch filepath.Ext(strings.TrimSuffix(path, ".txt")) {
	case ".ts", ".tsx", ".js", ".jsx", ".a":
		return "typescript"
	case ".json":
		return "json"
	case ".yaml", ".yml":
		return "yaml"
	case ".gql", ".graphql":
		return "graphql"
	case ".css", ".scss":
		return "css"
	case ".md", ".mdx":
		return "markdown"
	}
	return "unregistered"
}
func (d *Driver) format(path string) (Execution, Execution) {
	ctx, cancel := context.WithTimeout(context.Background(), d.Limit)
	a := execute(ctx, d.Root, d.Catalog, "--format", path)
	cancel()
	ctx, cancel = context.WithTimeout(context.Background(), d.Limit)
	b := execute(ctx, d.Root, "node", "--disable-warning=ExperimentalWarning", filepath.Join(d.Root, "oracle/node.mjs"), filepath.Join(d.Root, "stage1/cohere/scoreboard/format.a"), path, formatFamily(path))
	cancel()
	return a, b
}

// ruleAnswer retains each human finding and all its range/suggestion rows verbatim.
// Shared converged fixed text and overlap refusals are held by the all-fixes cell.
func ruleAnswer(e Execution, name string) Execution {
	if e.Error != "" {
		return e
	}
	lines := strings.SplitAfter(e.Output, "\n")
	var out strings.Builder
	active := false
	for i, line := range lines {
		if strings.HasPrefix(line, "skipped "+name+" ") || strings.HasPrefix(line, "refused "+name+" ") {
			out.WriteString(line)
			active = false
			continue
		}
		if i+1 < len(lines) && strings.HasPrefix(lines[i+1], "  ") {
			active = strings.HasPrefix(lines[i+1], "  "+name+"  ")
		}
		if strings.HasPrefix(line, "fixed\t") || strings.HasPrefix(line, "case ") || strings.HasPrefix(line, "skipped ") || strings.HasPrefix(line, "refused ") || strings.HasPrefix(line, "rejected ") {
			active = false
		}
		if active {
			out.WriteString(line)
		}
	}
	return Execution{Output: out.String(), Stderr: e.Stderr}
}
