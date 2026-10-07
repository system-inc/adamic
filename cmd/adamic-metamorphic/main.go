// adamic-metamorphic rewrites the oracle's fixtures into programs that mean the same thing by another
// road (internal/metamorphic), keeps each variant only when Node prints exactly what it printed for
// the original, and then holds a compiler to Node on the variants the way the oracle holds it on the
// fixtures: native sanitized and release, the JavaScript backend, and the leak check. Any
// disagreement, sanitizer report or crash is a finding; a variant the checker refuses or stage 0
// can't lower is discarded and counted, never a finding.
//
//	adamic-metamorphic generate -transforms wrap -fixtures 20 -import-free -o variants
//	adamic-metamorphic generate -transforms wrap,alias,method,identity,temporaries,combined -o variants
//	adamic-metamorphic run -variants variants -o results.jsonl
//	adamic-metamorphic run -root ../other -variants variants -o other.jsonl
//	adamic-metamorphic run -variants variants -transforms original,wrap -fixtures fills,sorts -o some.jsonl
//
// generate writes each variant under <out>/<transform>/<fixture>/, every original under
// <out>/original/<fixture>/, and what came of each try to <out>/manifest.jsonl. run builds the
// root's adamic command (with -trimpath), runs every kept variant and every original through it,
// writes a line per program to -o and prints a summary.
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/system-inc/adamic/internal/fuzz"
	"github.com/system-inc/adamic/internal/metamorphic"
)

// Entry is one try at a fixture: a transform's variant (or the original), and what came of it.
type Entry struct {
	Fixture   string `json:"fixture"`
	Transform string `json:"transform"`
	// Status is kept, skipped (nothing to rewrite, or every site refused), or node-differs.
	Status   string   `json:"status"`
	Reason   string   `json:"reason,omitempty"`
	File     string   `json:"file,omitempty"`
	Sites    int      `json:"sites,omitempty"`
	Refused  int      `json:"refused,omitempty"`
	Refusals []string `json:"refusals,omitempty"`
	Seconds  float64  `json:"seconds"`
}

// Outcome is one program's result on a checkout.
type Outcome struct {
	Fixture   string       `json:"fixture"`
	Transform string       `json:"transform"`
	File      string       `json:"file"`
	Verdict   fuzz.Verdict `json:"verdict"`
	Key       string       `json:"key,omitempty"`
	Line      string       `json:"line,omitempty"`
	Detail    string       `json:"detail,omitempty"`
	Seconds   float64      `json:"seconds"`
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "generate":
		os.Exit(generate(os.Args[2:]))
	case "run":
		os.Exit(run(os.Args[2:]))
	}
	usage()
	os.Exit(2)
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: adamic-metamorphic generate [-root .] [-transforms wrap,...,combined] [-fixtures N] [-import-free] -o <dir>")
	fmt.Fprintln(os.Stderr, "       adamic-metamorphic run [-root .] -variants <dir> -o <results.jsonl>")
}

func generate(arguments []string) int {
	flags := flag.NewFlagSet("generate", flag.ExitOnError)
	root := flags.String("root", ".", "the checkout whose fixtures, checker and oracle/node.mjs are used")
	written := flags.String("transforms", "wrap", "the transforms, each its own variant, comma separated; combined is all of them composed")
	count := flags.Int("fixtures", 0, "how many fixtures, chosen by a fixed stride over the sorted list (0: all)")
	importFree := flags.Bool("import-free", false, "only fixtures that import nothing")
	output := flags.String("o", "", "where the variants and manifest.jsonl go")
	parallel := flags.Int("parallel", 6, "how many fixtures at once")
	_ = flags.Parse(arguments)
	if *output == "" {
		usage()
		return 2
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	var plans [][]metamorphic.Transform
	var names []string
	for name := range strings.SplitSeq(*written, ",") {
		if name == "combined" {
			plans, names = append(plans, metamorphic.Combined), append(names, name)
			continue
		}
		transform, err := metamorphic.ParseTransform(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		plans, names = append(plans, []metamorphic.Transform{transform}), append(names, name)
	}
	fixtures, err := metamorphic.Fixtures(absoluteRoot)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if *importFree {
		var free []string
		for _, fixture := range fixtures {
			source, err := os.ReadFile(filepath.Join(absoluteRoot, fixture))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
			if imports, _ := metamorphic.Imports(fixture, string(source)); !imports {
				free = append(free, fixture)
			}
		}
		fixtures = free
	}
	fixtures = metamorphic.Stride(fixtures, *count)
	checkout := metamorphic.Checkout{Root: absoluteRoot}

	// Each fixture's original, with its Node run, which every variant has to match.
	originals := make(map[string]fuzz.Run)
	var lock sync.Mutex
	var entries []Entry
	each(len(fixtures), *parallel, func(index int) {
		fixture := fixtures[index]
		started := time.Now()
		file, err := place(absoluteRoot, fixture, *output, "original", nil)
		entry := Entry{Fixture: fixture, Transform: "original", Status: "kept", File: file}
		if err != nil {
			entry.Status, entry.Reason = "skipped", err.Error()
		} else {
			observed := checkout.Node(filepath.Join(*output, file), filepath.Dir(filepath.Join(*output, file)))
			lock.Lock()
			originals[fixture] = observed
			lock.Unlock()
		}
		entry.Seconds = time.Since(started).Seconds()
		lock.Lock()
		entries = append(entries, entry)
		lock.Unlock()
	})
	type job struct {
		fixture string
		plan    int
	}
	var jobs []job
	for _, fixture := range fixtures {
		for plan := range plans {
			jobs = append(jobs, job{fixture, plan})
		}
	}
	each(len(jobs), *parallel, func(index int) {
		fixture, plan := jobs[index].fixture, jobs[index].plan
		started := time.Now()
		entry := Entry{Fixture: fixture, Transform: names[plan]}
		defer func() {
			entry.Seconds = time.Since(started).Seconds()
			lock.Lock()
			entries = append(entries, entry)
			lock.Unlock()
			fmt.Fprintf(os.Stderr, "%s %s: %s %s\n", entry.Transform, entry.Fixture, entry.Status, entry.Reason)
		}()
		path := filepath.Join(absoluteRoot, fixture)
		source, err := os.ReadFile(path)
		if err != nil {
			entry.Status, entry.Reason = "skipped", err.Error()
			return
		}
		variant := metamorphic.Apply(path, string(source), plans[plan])
		entry.Sites, entry.Refused, entry.Refusals = variant.Sites, variant.Refused, variant.Refusals
		if variant.Skipped != "" {
			entry.Status, entry.Reason = "skipped", variant.Skipped
			return
		}
		file, err := place(absoluteRoot, fixture, *output, names[plan], []byte(variant.Source))
		if err != nil {
			entry.Status, entry.Reason = "skipped", err.Error()
			return
		}
		entry.File = file
		lock.Lock()
		original, found := originals[fixture]
		lock.Unlock()
		if !found {
			entry.Status, entry.Reason = "skipped", "the original didn't run"
			return
		}
		observed := checkout.Node(filepath.Join(*output, file), filepath.Dir(filepath.Join(*output, file)))
		if !metamorphic.Same(original, observed) {
			entry.Status = "node-differs"
			entry.Reason = nodeDifference(original, observed)
			return
		}
		entry.Status = "kept"
	})
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Transform != entries[j].Transform {
			return entries[i].Transform < entries[j].Transform
		}
		return entries[i].Fixture < entries[j].Fixture
	})
	if err := writeLines(filepath.Join(*output, "manifest.jsonl"), entries); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	summarizeManifest(entries)
	return 0
}

// place writes a fixture (or its variant, when source isn't nil) under out/<kind>/<fixture's slug>/,
// with the files beside it when it imports one, and returns its path relative to out.
func place(root string, fixture string, out string, kind string, source []byte) (string, error) {
	original := filepath.Join(root, fixture)
	text, err := os.ReadFile(original)
	if err != nil {
		return "", err
	}
	slug := strings.ReplaceAll(strings.TrimSuffix(fixture, filepath.Ext(fixture)), "/", "__")
	directory := filepath.Join(out, kind, slug)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return "", err
	}
	if _, relative := metamorphic.Imports(fixture, string(text)); relative {
		siblings, err := os.ReadDir(filepath.Dir(original))
		if err != nil {
			return "", err
		}
		for _, sibling := range siblings {
			if sibling.IsDir() || sibling.Name() == filepath.Base(fixture) {
				continue
			}
			content, err := os.ReadFile(filepath.Join(filepath.Dir(original), sibling.Name()))
			if err != nil {
				return "", err
			}
			if err := os.WriteFile(filepath.Join(directory, sibling.Name()), content, 0o644); err != nil {
				return "", err
			}
		}
	}
	if source == nil {
		source = text
	}
	if err := os.WriteFile(filepath.Join(directory, filepath.Base(fixture)), source, 0o644); err != nil {
		return "", err
	}
	return filepath.Join(kind, slug, filepath.Base(fixture)), nil
}

func nodeDifference(original fuzz.Run, variant fuzz.Run) string {
	switch {
	case variant.TimedOut:
		return "the variant never finished on Node"
	case original.ExitCode != variant.ExitCode:
		return fmt.Sprintf("exit %d, the variant's %d: %s", original.ExitCode, variant.ExitCode, firstLine(string(variant.Stderr)))
	case string(original.Stdout) != string(variant.Stdout):
		return "stdout differs"
	}
	return "stderr differs: " + firstLine(string(variant.Stderr))
}

func run(arguments []string) int {
	flags := flag.NewFlagSet("run", flag.ExitOnError)
	root := flags.String("root", ".", "the checkout whose compiler is under test")
	variants := flags.String("variants", "", "what generate wrote")
	only := flags.String("transforms", "", "only these transforms' variants, comma separated, original among them (default: all)")
	fixtures := flags.String("fixtures", "", "only fixtures whose path contains one of these, comma separated (default: all)")
	output := flags.String("o", "", "where a line per program goes")
	work := flags.String("work", filepath.Join(os.TempDir(), "adamic-metamorphic"), "where programs are built and run")
	parallel := flags.Int("parallel", 6, "how many programs at once")
	_ = flags.Parse(arguments)
	if *variants == "" || *output == "" {
		usage()
		return 2
	}
	absoluteRoot, err := filepath.Abs(*root)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if err := os.MkdirAll(*work, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	adamic := filepath.Join(*work, "adamic")
	build := exec.Command("go", "build", "-trimpath", "-o", adamic, "./cmd/adamic")
	build.Dir = absoluteRoot
	if combined, err := build.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "building adamic in %s: %v\n%s", absoluteRoot, err, combined)
		return 1
	}
	checkout := metamorphic.Checkout{Root: absoluteRoot, Adamic: adamic}
	entries, err := readManifest(filepath.Join(*variants, "manifest.jsonl"))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	wanted := map[string]bool{}
	for name := range strings.SplitSeq(*only, ",") {
		if name != "" {
			wanted[name] = true
		}
	}
	var named []string
	for name := range strings.SplitSeq(*fixtures, ",") {
		if name != "" {
			named = append(named, name)
		}
	}
	var programs []Entry
	for _, entry := range entries {
		if entry.Status != "kept" || (len(wanted) > 0 && !wanted[entry.Transform]) {
			continue
		}
		matched := len(named) == 0
		for _, name := range named {
			matched = matched || strings.Contains(entry.Fixture, name)
		}
		if matched {
			programs = append(programs, entry)
		}
	}
	absoluteVariants, err := filepath.Abs(*variants)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	outcomes := make([]Outcome, len(programs))
	each(len(programs), *parallel, func(index int) {
		entry := programs[index]
		started := time.Now()
		directory, err := os.MkdirTemp(*work, "program-")
		outcome := Outcome{Fixture: entry.Fixture, Transform: entry.Transform, File: entry.File}
		if err != nil {
			outcome.Verdict, outcome.Key = fuzz.Finding, "harness: "+err.Error()
		} else {
			result := checkout.Run(filepath.Join(absoluteVariants, entry.File), directory)
			outcome.Verdict, outcome.Key, outcome.Line, outcome.Detail = result.Verdict, result.Key, result.Line, result.Detail
			// The builds are large and every one is kept in the variants; only the directory goes.
			_ = os.RemoveAll(directory)
		}
		outcome.Seconds = time.Since(started).Seconds()
		outcomes[index] = outcome
		if outcome.Verdict == fuzz.Crash || outcome.Verdict == fuzz.Finding {
			fmt.Fprintf(os.Stderr, "%s %s: %s %s\n", outcome.Transform, outcome.Fixture, outcome.Verdict, outcome.Key)
		}
	})
	if err := writeLines(*output, outcomes); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	summarizeOutcomes(outcomes)
	return 0
}

// each runs work for every index, at most parallel at once.
func each(count int, parallel int, work func(int)) {
	slots := make(chan struct{}, max(parallel, 1))
	var group sync.WaitGroup
	for index := range count {
		slots <- struct{}{}
		group.Go(func() {
			defer func() { <-slots }()
			work(index)
		})
	}
	group.Wait()
}

func writeLines[T any](path string, values []T) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	encoder := json.NewEncoder(writer)
	for _, value := range values {
		if err := encoder.Encode(value); err != nil {
			file.Close()
			return err
		}
	}
	if err := writer.Flush(); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}

func readManifest(path string) ([]Entry, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for line := range strings.SplitSeq(strings.TrimSpace(string(content)), "\n") {
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func summarizeManifest(entries []Entry) {
	type tally struct {
		kept, skipped, differs, sites, refused int
		seconds                                float64
		count                                  int
	}
	byTransform := map[string]*tally{}
	var order []string
	for _, entry := range entries {
		counts := byTransform[entry.Transform]
		if counts == nil {
			counts = &tally{}
			byTransform[entry.Transform] = counts
			order = append(order, entry.Transform)
		}
		counts.count++
		counts.seconds += entry.Seconds
		counts.sites += entry.Sites
		counts.refused += entry.Refused
		switch entry.Status {
		case "kept":
			counts.kept++
		case "skipped":
			counts.skipped++
		case "node-differs":
			counts.differs++
		}
	}
	sort.Strings(order)
	fmt.Println("transform\tfixtures\tkept\tskipped\tnode-differs\tsites kept\tsites refused\tseconds per fixture")
	for _, name := range order {
		counts := byTransform[name]
		fmt.Printf("%s\t%d\t%d\t%d\t%d\t%d\t%d\t%.2f\n", name, counts.count, counts.kept, counts.skipped, counts.differs, counts.sites, counts.refused, counts.seconds/float64(counts.count))
	}
}

func summarizeOutcomes(outcomes []Outcome) {
	byTransform := map[string]map[fuzz.Verdict]int{}
	seconds := map[string]float64{}
	totals := map[string]int{}
	var order []string
	for _, outcome := range outcomes {
		if byTransform[outcome.Transform] == nil {
			byTransform[outcome.Transform] = map[fuzz.Verdict]int{}
			order = append(order, outcome.Transform)
		}
		byTransform[outcome.Transform][outcome.Verdict]++
		seconds[outcome.Transform] += outcome.Seconds
		totals[outcome.Transform]++
	}
	sort.Strings(order)
	for _, name := range order {
		var parts []string
		for _, verdict := range fuzz.Verdicts {
			if count := byTransform[name][verdict]; count > 0 {
				parts = append(parts, fmt.Sprintf("%s %d", verdict, count))
			}
		}
		fmt.Printf("%s: %d programs, %s, %.2f seconds each\n", name, totals[name], strings.Join(parts, ", "), seconds[name]/float64(totals[name]))
	}
	for _, outcome := range outcomes {
		if outcome.Verdict == fuzz.Crash || outcome.Verdict == fuzz.Finding {
			fmt.Printf("%s %s %s: %s\n", outcome.Verdict, outcome.Transform, outcome.Fixture, outcome.Key)
		}
	}
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}
