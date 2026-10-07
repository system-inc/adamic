// adamic-stage1-progress inventories recorded stage 1 coverage without checking out branches.
package stage1progress

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

type coverage struct {
	File         string
	Declarations []string
}
type slice struct {
	Directory    string
	Required     []string
	Coverage     []coverage
	Packages     []string
	Partial      []string
	NativePhrase string
	Note         string
}
type packageRow struct {
	Package     string   `json:"package"`
	GoLines     int      `json:"non_test_go_lines"`
	PortedLines int      `json:"credited_ported_lines"`
	Status      string   `json:"status"`
	Evidence    []string `json:"evidence,omitempty"`
}
type sliceRow struct {
	Directory  string `json:"directory"`
	Credited   bool   `json:"credited"`
	Note       string `json:"note"`
	GapSummary string `json:"gap_summary"`
}
type report struct {
	Ref        string       `json:"ref"`
	Commit     string       `json:"commit"`
	Cohere     string       `json:"cohere_commit"`
	Packages   []packageRow `json:"packages"`
	Slices     []sliceRow   `json:"slices"`
	External   []string     `json:"external_dependencies,omitempty"`
	Total      int          `json:"non_test_go_lines"`
	Ported     int          `json:"credited_ported_lines"`
	Percent    float64      `json:"ported_percent_lower_bound"`
	Complete   int          `json:"complete_packages"`
	Partial    int          `json:"partial_packages"`
	NotStarted int          `json:"not_started_packages"`
	marks      map[string]map[int]bool
	partial    map[string]bool
}
type inventory struct {
	sources  map[string][]byte
	lines    map[string]int
	packages map[string]int
}

func git(ctx context.Context, root string, args ...string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "git", append([]string{"-C", root}, args...)...)
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return output, nil
}
func physicalLines(source []byte) int {
	n := bytes.Count(source, []byte{'\n'})
	if len(source) > 0 && source[len(source)-1] != '\n' {
		n++
	}
	return n
}
func production(name string) bool {
	if !strings.HasPrefix(name, "internal/") || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
		return false
	}
	for _, part := range strings.Split(name, "/") {
		if part == "testdata" || strings.HasPrefix(part, ".") {
			return false
		}
	}
	return true
}
func loadInventory(ctx context.Context, root, pin string) (*inventory, error) {
	tree, err := git(ctx, root, "ls-tree", "-r", pin, "--", "internal")
	if err != nil {
		return nil, err
	}
	var names, hashes []string
	for _, line := range strings.Split(strings.TrimSpace(string(tree)), "\n") {
		fields := strings.SplitN(line, "\t", 2)
		if len(fields) != 2 || !production(fields[1]) {
			continue
		}
		metadata := strings.Fields(fields[0])
		if len(metadata) != 3 || metadata[1] != "blob" {
			continue
		}
		names = append(names, fields[1])
		hashes = append(hashes, metadata[2])
	}
	if len(names) == 0 {
		return nil, fmt.Errorf("no cohere internal Go sources at %s", pin)
	}
	batchCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	command := exec.CommandContext(batchCtx, "git", "-C", root, "cat-file", "--batch")
	command.Stdin = strings.NewReader(strings.Join(hashes, "\n") + "\n")
	output, err := command.Output()
	if err != nil {
		return nil, err
	}
	reader := bufio.NewReader(bytes.NewReader(output))
	result := &inventory{sources: map[string][]byte{}, lines: map[string]int{}, packages: map[string]int{}}
	for _, name := range names {
		header, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		fields := strings.Fields(header)
		if len(fields) != 3 || fields[1] != "blob" {
			return nil, fmt.Errorf("invalid blob header %q", header)
		}
		size, err := strconv.Atoi(fields[2])
		if err != nil {
			return nil, err
		}
		data := make([]byte, size+1)
		if _, err = io.ReadFull(reader, data); err != nil {
			return nil, err
		}
		source := data[:size]
		result.sources[name] = source
		result.lines[name] = physicalLines(source)
		result.packages[path.Dir(name)] += result.lines[name]
	}
	return result, nil
}
func credit(source []byte, declarations []string) (map[int]bool, error) {
	marks := map[int]bool{}
	if len(declarations) == 0 {
		for n := 1; n <= physicalLines(source); n++ {
			marks[n] = true
		}
		return marks, nil
	}
	files := token.NewFileSet()
	file, err := parser.ParseFile(files, "coverage.go", source, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	wanted := map[string]bool{}
	for _, name := range declarations {
		wanted[name] = false
	}
	for _, declaration := range file.Decls {
		function, ok := declaration.(*ast.FuncDecl)
		if !ok {
			continue
		}
		name := function.Name.Name
		if _, ok := wanted[name]; !ok {
			continue
		}
		if wanted[name] {
			return nil, fmt.Errorf("ambiguous declaration %s", name)
		}
		wanted[name] = true
		start := function.Pos()
		if function.Doc != nil {
			start = function.Doc.Pos()
		}
		for n := files.Position(start).Line; n <= files.Position(function.End()).Line; n++ {
			marks[n] = true
		}
	}
	for name, found := range wanted {
		if !found {
			return nil, fmt.Errorf("mapped declaration %s missing", name)
		}
	}
	return marks, nil
}
func blob(ctx context.Context, root, ref, name string) ([]byte, error) {
	return git(ctx, root, "show", ref+":"+name)
}
func measureContext(ctx context.Context, root, ref string, cache map[string]*inventory) (*report, error) {
	commit, err := git(ctx, root, "rev-parse", ref+"^{commit}")
	if err != nil {
		return nil, err
	}
	link, err := git(ctx, root, "ls-tree", ref, "--", "cohere")
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(link))
	if len(fields) != 4 || fields[1] != "commit" {
		return nil, fmt.Errorf("%s has no cohere gitlink", ref)
	}
	pin := fields[2]
	inv := cache[pin]
	if inv == nil {
		inv, err = loadInventory(ctx, path.Join(root, "cohere"), pin)
		if err != nil {
			return nil, err
		}
		cache[pin] = inv
	}
	tree, err := git(ctx, root, "ls-tree", "-r", "--name-only", ref, "--", "stage1")
	if err != nil {
		return nil, err
	}
	present := map[string]bool{}
	directories := map[string]bool{}
	for _, name := range strings.Split(strings.TrimSpace(string(tree)), "\n") {
		present[name] = true
		if strings.HasPrefix(name, "stage1/cohere/") {
			parts := strings.Split(name, "/")
			if len(parts) >= 4 {
				directories[parts[2]] = true
			}
		}
	}
	result := &report{Ref: ref, Commit: strings.TrimSpace(string(commit)), Cohere: pin, marks: map[string]map[int]bool{}, partial: map[string]bool{}}
	evidence := map[string][]string{}
	known := map[string]bool{}
	for _, spec := range slices {
		known[spec.Directory] = true
		if !directories[spec.Directory] {
			continue
		}
		directory := "stage1/cohere/" + spec.Directory
		gaps, err := blob(ctx, root, ref, directory+"/GAPS.md")
		if err != nil {
			return nil, fmt.Errorf("slice %s needs GAPS.md: %w", directory, err)
		}
		active := true
		for _, required := range spec.Required {
			if !present[directory+"/"+required] {
				active = false
			}
		}
		if spec.NativePhrase != "" && !strings.Contains(string(gaps), spec.NativePhrase) {
			active = false
		}
		result.Slices = append(result.Slices, sliceRow{directory, active, spec.Note, firstParagraph(string(gaps))})
		for _, pkg := range spec.Packages {
			if _, ok := inv.packages[pkg]; !ok {
				return nil, fmt.Errorf("mapped package %s absent", pkg)
			}
			evidence[pkg] = appendUnique(evidence[pkg], directory+"/GAPS.md")
			if !active {
				result.partial[pkg] = true
			}
		}
		for _, item := range spec.Coverage {
			pkg := path.Dir(item.File)
			if _, ok := inv.packages[pkg]; !ok {
				return nil, fmt.Errorf("mapped package %s absent", pkg)
			}
			evidence[pkg] = appendUnique(evidence[pkg], directory+"/GAPS.md")
			if !active {
				result.partial[pkg] = true
			}
		}
		if !active {
			continue
		}
		for _, pkg := range spec.Partial {
			result.partial[pkg] = true
		}
		for _, item := range spec.Coverage {
			source, ok := inv.sources[item.File]
			if !ok {
				return nil, fmt.Errorf("mapped source %s absent", item.File)
			}
			marks, err := credit(source, item.Declarations)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", item.File, err)
			}
			if result.marks[item.File] == nil {
				result.marks[item.File] = map[int]bool{}
			}
			for line := range marks {
				result.marks[item.File][line] = true
			}
			pkg := path.Dir(item.File)
			evidence[pkg] = appendUnique(evidence[pkg], directory+"/GAPS.md")
		}
	}
	for directory := range directories {
		if !known[directory] {
			return nil, fmt.Errorf("unmapped stage1/cohere/%s: add reviewed source coverage before reporting", directory)
		}
	}
	for _, name := range []string{"scanner", "parser"} {
		directory := "stage1/typescript/" + name
		if present[directory+"/GAPS.md"] {
			gaps, err := blob(ctx, root, ref, directory+"/GAPS.md")
			if err != nil {
				return nil, err
			}
			result.External = append(result.External, directory+": "+firstParagraph(string(gaps))+" (outside cohere/internal denominator)")
		}
	}
	for pkg, total := range inv.packages {
		row := packageRow{Package: pkg, GoLines: total, Evidence: evidence[pkg], Status: "not started"}
		for file, marks := range result.marks {
			if path.Dir(file) == pkg {
				row.PortedLines += len(marks)
			}
		}
		if len(row.Evidence) > 0 {
			row.Status = "partly ported"
		}
		if row.PortedLines == total && !result.partial[pkg] {
			row.Status = "ported and held byte-identical"
		}
		result.Packages = append(result.Packages, row)
	}
	finish(result)
	return result, nil
}
func finish(result *report) {
	sort.Slice(result.Packages, func(i, j int) bool { return result.Packages[i].Package < result.Packages[j].Package })
	for _, row := range result.Packages {
		result.Total += row.GoLines
		result.Ported += row.PortedLines
		switch row.Status {
		case "not started":
			result.NotStarted++
		case "partly ported":
			result.Partial++
		default:
			result.Complete++
		}
	}
	if result.Total > 0 {
		result.Percent = 100 * float64(result.Ported) / float64(result.Total)
	}
}

// This is an evidence union, not a merge or a new executable composition.
func union(reports []*report) (*report, error) {
	first := reports[0]
	result := &report{Ref: "main + pending evidence union (not a merge)", Commit: first.Commit, Cohere: first.Cohere, marks: map[string]map[int]bool{}, partial: map[string]bool{}}
	rows := map[string]*packageRow{}
	for _, report := range reports {
		if report.Cohere != first.Cohere {
			return nil, fmt.Errorf("cannot union different cohere pins")
		}
		for file, marks := range report.marks {
			if result.marks[file] == nil {
				result.marks[file] = map[int]bool{}
			}
			for line := range marks {
				result.marks[file][line] = true
			}
		}
		for pkg, partial := range report.partial {
			if partial {
				result.partial[pkg] = true
			}
		}
		for _, row := range report.Packages {
			saved := rows[row.Package]
			if saved == nil {
				saved = &packageRow{Package: row.Package, GoLines: row.GoLines}
				rows[row.Package] = saved
			}
			for _, evidence := range row.Evidence {
				saved.Evidence = appendUnique(saved.Evidence, report.Ref+":"+evidence)
			}
		}
	}
	for pkg, row := range rows {
		for file, marks := range result.marks {
			if path.Dir(file) == pkg {
				row.PortedLines += len(marks)
			}
		}
		row.Status = "not started"
		if len(row.Evidence) > 0 {
			row.Status = "partly ported"
		}
		if row.PortedLines == row.GoLines && !result.partial[pkg] {
			row.Status = "ported and held byte-identical"
		}
		result.Packages = append(result.Packages, *row)
	}
	finish(result)
	return result, nil
}

func firstParagraph(text string) string {
	lines := strings.Split(text, "\n")
	var paragraph []string
	started := false
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !started && (line == "" || strings.HasPrefix(line, "#")) {
			continue
		}
		if line == "" {
			break
		}
		started = true
		paragraph = append(paragraph, line)
	}
	return strings.Join(paragraph, " ")
}
func appendUnique(items []string, item string) []string {
	for _, existing := range items {
		if existing == item {
			return items
		}
	}
	return append(items, item)
}
func pending(root string) ([]string, []string, error) {
	ctx := context.Background()
	output, err := git(ctx, root, "for-each-ref", "--format=%(refname:short)", "refs/remotes/origin")
	if err != nil {
		return nil, nil, err
	}
	var refs, integrated []string
	for _, ref := range strings.Fields(string(output)) {
		if !strings.HasPrefix(ref, "origin/codex/stage1-") && ref != "origin/codex/typescript-scanner" {
			continue
		}
		command := exec.CommandContext(ctx, "git", "-C", root, "merge-base", "--is-ancestor", ref, "origin/main")
		err := command.Run()
		if err == nil {
			integrated = append(integrated, ref)
		} else if exit, ok := err.(*exec.ExitError); ok && exit.ExitCode() == 1 {
			refs = append(refs, ref)
		} else {
			return nil, nil, err
		}
	}
	return refs, integrated, nil
}

// Report is the existing conservative stage 1 source inventory.
type Report = report

// Measure inventories one Git snapshot without executing a gate.
func Measure(root, ref string) (*Report, error) { return measure(root, ref, map[string]*inventory{}) }

func measure(root, ref string, cache map[string]*inventory) (*report, error) {
	return measureContext(context.Background(), root, ref, cache)
}

// MeasureContext bounds all Git reads by the caller's deadline.
func MeasureContext(ctx context.Context, root, ref string) (*Report, error) {
	return measureContext(ctx, root, ref, map[string]*inventory{})
}

// NewMeasurer shares pinned inventories and identical trees across sequential snapshots.
func NewMeasurer(ctx context.Context) func(string, string) (*Report, error) {
	inventories := map[string]map[string]*inventory{}
	snapshots := map[string]*Report{}
	failures := map[string]error{}
	return func(root, ref string) (*Report, error) {
		tree, err := git(ctx, root, "ls-tree", ref, "--", "stage1", "cohere")
		if err != nil {
			return nil, err
		}
		key := root + "\x00" + string(tree)
		if err := failures[key]; err != nil {
			return nil, err
		}
		if saved := snapshots[key]; saved != nil {
			commit, err := git(ctx, root, "rev-parse", ref)
			if err != nil {
				return nil, err
			}
			copy := *saved
			copy.Ref = ref
			copy.Commit = strings.TrimSpace(string(commit))
			return &copy, nil
		}
		if inventories[root] == nil {
			inventories[root] = map[string]*inventory{}
		}
		result, err := measureContext(ctx, root, ref, inventories[root])
		if err == nil {
			snapshots[key] = result
		} else {
			failures[key] = err
		}
		return result, err
	}
}
