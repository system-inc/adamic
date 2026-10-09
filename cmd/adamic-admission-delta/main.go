// Command adamic-admission-delta checks changes in the compiler's admission set.
package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"
)

type observation struct {
	WallSeconds float64 `json:"wall_seconds"`
	Exit        int     `json:"exit"`
	Stdout      string  `json:"stdout"`
	Stderr      string  `json:"stderr"`
	Error       string  `json:"error,omitempty"`
}
type program struct {
	Path string `json:"path"`
	Blob string `json:"blob"`
}
type corpus struct {
	Name     string    `json:"name"`
	Programs []program `json:"programs"`
}
type manifest struct {
	SHA       string   `json:"sha"`
	Generator string   `json:"generator"`
	Corpora   []corpus `json:"corpora"`
}
type entry struct {
	program
	Corpus            string       `json:"corpus"`
	Class             string       `json:"class"`
	Base              observation  `json:"base_compile"`
	Head              observation  `json:"compile"`
	Node              *observation `json:"node"`
	JavaScript        *observation `json:"javascript"`
	Native            *observation `json:"native"`
	JavaScriptCompile *observation `json:"javascript_compile,omitempty"`
	NativeCompile     *observation `json:"native_compile,omitempty"`
	Agree             *bool        `json:"agree"`
	Sampled           bool         `json:"sampled"`
}
type report struct {
	Diff                  []program `json:"diff"`
	DiffCount             int       `json:"diff_count"`
	CompileTimeoutSeconds float64   `json:"compile_timeout_seconds"`
	RuntimeTimeoutSeconds float64   `json:"runtime_timeout_seconds"`
	Base                  string    `json:"base"`
	Head                  string    `json:"head"`
	GeneratorBlob         string    `json:"generator_blob"`
	ManifestBlob          string    `json:"manifest_blob"`
	Admitted              int       `json:"admitted"`
	Programs              []entry   `json:"programs"`
	Verdict               string    `json:"verdict"`
	Corpora               []corpus  `json:"corpora"`
	SamplingSeed          string    `json:"sampling_seed"`
	SamplingSize          int       `json:"sampling_size"`
	Omitted               int       `json:"omitted"`
	BudgetSeconds         float64   `json:"budget_seconds"`
	BudgetUsedSeconds     float64   `json:"budget_used_seconds"`
}

func execute(dir string, limit time.Duration, name string, args ...string) observation {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = time.Second
	var out, errout bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errout
	started := time.Now()
	err := command.Run()
	o := observation{Stdout: out.String(), Stderr: errout.String(), WallSeconds: time.Since(started).Seconds()}
	if err != nil {
		o.Exit = -1
		if e, ok := err.(*exec.ExitError); ok {
			o.Exit = e.ExitCode()
		} else {
			o.Error = err.Error()
		}
	}
	if ctx.Err() != nil {
		o.Error = "timeout"
	} else if o.Exit < 0 && o.Error == "" {
		o.Error = "crash"
	}
	return o
}
func git(dir string, args ...string) (string, error) {
	o := execute(dir, 60*time.Second, "git", args...)
	if o.Exit != 0 || o.Error != "" {
		return "", fmt.Errorf("git %v: %s %s", args, o.Error, o.Stderr)
	}
	return strings.TrimSpace(o.Stdout), nil
}
func compileClass(o observation) string {
	if o.Error != "" || o.Exit < 0 || strings.Contains(o.Stderr, "panic:") || strings.Contains(o.Stderr, "fatal error:") {
		return "error"
	}
	if o.Exit == 0 {
		return "accepted"
	}
	if o.Exit == 1 && (strings.Contains(o.Stderr, "stage 0 can't lower") || strings.Contains(o.Stderr, "Adamic 0.1 refuses") || strings.Contains(o.Stderr, "error TS")) {
		return "refused"
	}
	return "error"
}
func classify(a, b observation) string {
	if a.Error == "timeout" || b.Error == "timeout" {
		return "compiler-timeout"
	}
	for _, o := range []observation{a, b} {
		if o.Error == "crash" || strings.Contains(o.Stderr, "panic:") || strings.Contains(o.Stderr, "fatal error:") {
			return "compiler-crash"
		}
	}
	x, y := compileClass(a), compileClass(b)
	if x == "error" || y == "error" {
		return "compiler-error"
	}
	if x == y {
		return x + "-by-both"
	}
	if y == "accepted" {
		return "newly-accepted"
	}
	return "newly-refused"
}
func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
func run(args []string) error {
	flags := flag.NewFlagSet("adamic-admission-delta", flag.ContinueOnError)
	base := flags.String("base", "", "base revision")
	head := flags.String("head", "HEAD", "head revision")
	baseBinary := flags.String("base-binary", "", "built base compiler")
	headBinary := flags.String("head-binary", "", "built head compiler")
	manifestPath := flags.String("manifest", "", "pinned JSON corpus manifest")
	generator := flags.String("manifest-generator", "", "generator path at head")
	corpusDir := flags.String("corpus", "", "ad hoc corpus directory relative to head")
	budget := flags.Float64("budget", 0, "runtime budget in seconds; 0 means all, witnesses always run")
	asJSON := flags.Bool("json", false, "emit JSON to stdout")
	limit := flags.Duration("timeout", 10*time.Second, "per-runtime command timeout")
	compileLimit := flags.Duration("compile-timeout", 45*time.Second, "per-program compiler timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *base == "" {
		return fmt.Errorf("--base is required")
	}
	if *budget < 0 {
		return fmt.Errorf("budget must be nonnegative")
	}
	if *limit <= 0 || *compileLimit <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	root, err := git(".", "rev-parse", "--show-toplevel")
	if err != nil {
		return err
	}
	result := report{Programs: []entry{}, Verdict: "pass"}
	result.Base, err = git(root, "rev-parse", *base+"^{commit}")
	if err != nil {
		return err
	}
	result.Head, err = git(root, "rev-parse", *head+"^{commit}")
	if err != nil {
		return err
	}
	scratch, err := os.MkdirTemp("", "admission-delta-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(scratch)
	checkout := func(sha, name string) (string, error) {
		path := filepath.Join(scratch, name)
		_, e := git(root, "worktree", "add", "--detach", path, sha)
		return path, e
	}
	headTree, err := checkout(result.Head, "head")
	if err != nil {
		return err
	}
	defer git(root, "worktree", "remove", "--force", headTree)
	if err = prepareCheckout(headTree); err != nil {
		return err
	}
	build := func(sha, name, supplied string) (string, error) {
		if supplied != "" {
			return filepath.Abs(supplied)
		}
		tree, e := checkout(sha, name)
		if e != nil {
			return "", e
		}
		defer git(root, "worktree", "remove", "--force", tree)
		if _, e = git(tree, "submodule", "update", "--init", "--recursive"); e != nil {
			return "", e
		}
		binary := filepath.Join(scratch, name+"-adamic")
		o := execute(tree, 120*time.Second, "go", "build", "-o", binary, "./cmd/adamic")
		if o.Exit != 0 || o.Error != "" {
			return "", fmt.Errorf("build %s: %s %s", name, o.Error, o.Stderr)
		}
		return binary, nil
	}
	b, err := build(result.Base, "base-build", *baseBinary)
	if err != nil {
		return err
	}
	h := b
	if result.Base != result.Head || *baseBinary != "" || *headBinary != "" {
		h, err = build(result.Head, "head-build", *headBinary)
	}
	if err != nil {
		return err
	}
	var m manifest
	if *manifestPath != "" {
		data, e := os.ReadFile(*manifestPath)
		if e != nil {
			return e
		}
		if e = json.Unmarshal(data, &m); e != nil {
			return e
		}
		if m.SHA != result.Head {
			return fmt.Errorf("manifest sha %s differs from head %s", m.SHA, result.Head)
		}
		result.ManifestBlob, err = git(root, "hash-object", *manifestPath)
		if err != nil {
			return err
		}
	} else if *corpusDir != "" {
		c := corpus{Name: "adhoc", Programs: []program{}}
		cleanCorpus := filepath.Clean(*corpusDir)
		if filepath.IsAbs(cleanCorpus) || cleanCorpus == ".." || strings.HasPrefix(cleanCorpus, "../") {
			return fmt.Errorf("corpus must be inside head checkout")
		}
		start := filepath.Join(headTree, cleanCorpus)
		err = filepath.WalkDir(start, func(path string, d os.DirEntry, e error) error {
			if e != nil {
				return e
			}
			if !d.IsDir() && (strings.HasSuffix(path, ".a") || strings.HasSuffix(path, ".ts")) {
				rel, e := filepath.Rel(headTree, path)
				if e != nil {
					return e
				}
				blob, e := git(root, "rev-parse", result.Head+":"+filepath.ToSlash(rel))
				if e != nil {
					return e
				}
				c.Programs = append(c.Programs, program{filepath.ToSlash(rel), blob})
			}
			return nil
		})
		if err != nil {
			return err
		}
		m.Corpora = []corpus{c}
	} else {
		return fmt.Errorf("--manifest or --corpus is required")
	}
	if *generator == "" {
		*generator = m.Generator
	}
	if *generator != "" {
		result.GeneratorBlob, err = git(root, "rev-parse", result.Head+":"+*generator)
		if err != nil {
			return err
		}
	}
	for _, c := range m.Corpora {
		if c.Name == "diff" {
			return fmt.Errorf("manifest corpus name diff is reserved for computed revision coverage")
		}
	}
	result.Diff, err = changedPrograms(root, result.Base, result.Head)
	if err != nil {
		return err
	}
	result.DiffCount = len(result.Diff)
	m.Corpora = append([]corpus{{Name: "diff", Programs: result.Diff}}, m.Corpora...)
	result.Corpora = m.Corpora
	result.SamplingSeed = result.Head
	result.BudgetSeconds = *budget
	result.CompileTimeoutSeconds = compileLimit.Seconds()
	result.RuntimeTimeoutSeconds = limit.Seconds()
	seen := map[string]string{}
	for _, c := range m.Corpora {
		for _, p := range c.Programs {
			clean := filepath.ToSlash(filepath.Clean(p.Path))
			if filepath.IsAbs(p.Path) || clean == ".." || strings.HasPrefix(clean, "../") {
				return fmt.Errorf("unsafe program path %q", p.Path)
			}
			info, e := os.Lstat(filepath.Join(headTree, p.Path))
			if e != nil {
				return e
			}
			if !info.Mode().IsRegular() {
				return fmt.Errorf("program is not a regular file: %s", p.Path)
			}
			blob, e := git(root, "rev-parse", result.Head+":"+p.Path)
			if e != nil {
				return e
			}
			if blob != p.Blob {
				return fmt.Errorf("blob mismatch for %s: manifest %s head %s", p.Path, p.Blob, blob)
			}
			if previous, exists := seen[p.Path]; exists {
				if previous != p.Blob {
					return fmt.Errorf("conflicting blob for %s", p.Path)
				}
				continue
			}
			seen[p.Path] = p.Blob
			a := execute(headTree, *compileLimit, b, "c", p.Path)
			z := execute(headTree, *compileLimit, h, "c", p.Path)
			a.Stdout = ""
			z.Stdout = ""
			record := entry{program: p, Corpus: c.Name, Class: classify(a, z), Base: a, Head: z}
			if record.Class == "newly-accepted" {
				result.Admitted++
			}
			if strings.HasPrefix(record.Class, "compiler-") {
				result.Verdict = "fail"
			}
			result.Programs = append(result.Programs, record)
		}
	}

	selected, omitted := sample(result.Programs, result.Head, *budget, *limit)
	result.Omitted = omitted
	if omitted > 0 && result.Verdict == "pass" {
		result.Verdict = "sampled"
	}
	started := time.Now()
	for _, index := range selected {
		record := &result.Programs[index]
		record.Sampled = true
		source := filepath.Join(headTree, record.Path)
		node := execute(headTree, *limit, "node", "--disable-warning=ExperimentalWarning", filepath.Join(headTree, "oracle/node.mjs"), source)
		record.Node = &node
		jsCompile := execute(headTree, *limit, h, "js", record.Path)
		jsPath := filepath.Join(scratch, fmt.Sprintf("program-%d.mjs", index))
		javascript := observation{Exit: -1, Error: "JavaScript compilation failed"}
		if jsCompile.Exit == 0 && jsCompile.Error == "" {
			if e := os.WriteFile(jsPath, []byte(jsCompile.Stdout), 0600); e != nil {
				return e
			}
			javascript = execute(headTree, *limit, "node", "--disable-warning=ExperimentalWarning", filepath.Join(headTree, "oracle/node.mjs"), jsPath)
		}
		jsCompile.Stdout = ""
		record.JavaScriptCompile = &jsCompile
		record.JavaScript = &javascript
		nativePath := filepath.Join(scratch, fmt.Sprintf("program-%d", index))
		nativeCompile := execute(headTree, *limit, h, "build", record.Path, "-o", nativePath)
		native := observation{Exit: -1, Error: "native compilation failed"}
		if nativeCompile.Exit == 0 && nativeCompile.Error == "" {
			native = execute(headTree, *limit, nativePath)
		}
		record.NativeCompile = &nativeCompile
		record.Native = &native
		agree := outputsAgree(node, javascript, native)
		record.Agree = &agree
		if !agree {
			result.Verdict = "fail"
		}
	}
	result.SamplingSize = len(selected)
	result.BudgetUsedSeconds = time.Since(started).Seconds()
	if *asJSON {
		if err = json.NewEncoder(os.Stdout).Encode(result); err != nil {
			return err
		}
	} else {
		fmt.Printf("admitted: %d\nverdict: %s\n", result.Admitted, result.Verdict)
		for _, p := range result.Programs {
			fmt.Printf("%s %s\n", p.Class, p.Path)
		}
	}
	if result.Verdict == "fail" {
		return fmt.Errorf("admission delta failed")
	}
	return nil
}

// Each selected program reserves five command timeouts. Diff and witnesses override the budget.
func sample(programs []entry, seed string, seconds float64, limit time.Duration) ([]int, int) {
	diff := []int{}
	witnesses := []int{}
	others := []int{}
	for i, p := range programs {
		if p.Class != "newly-accepted" {
			continue
		}
		if p.Corpus == "diff" {
			diff = append(diff, i)
		} else if p.Corpus == "witnesses" {
			witnesses = append(witnesses, i)
		} else {
			others = append(others, i)
		}
	}
	sort.Slice(others, func(i, j int) bool { return programs[others[i]].Path < programs[others[j]].Path })
	digest := sha256.Sum256([]byte(seed))
	random := rand.New(rand.NewSource(int64(binary.LittleEndian.Uint64(digest[:8]))))
	random.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
	take := len(others)
	if seconds > 0 {
		capacity := int(seconds/(5*limit.Seconds())) - len(diff) - len(witnesses)
		if capacity < 0 {
			capacity = 0
		}
		if take > capacity {
			take = capacity
		}
	}
	return append(append(diff, witnesses...), others[:take]...), len(others) - take
}
func outputsAgree(a, b, c observation) bool {
	return a.Error == "" && b.Error == "" && c.Error == "" && a.Exit >= 0 && a.Exit == b.Exit && a.Exit == c.Exit && a.Stdout == b.Stdout && a.Stdout == c.Stdout
}
