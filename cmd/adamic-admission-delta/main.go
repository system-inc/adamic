// Command adamic-admission-delta checks changes in the compiler's admission set.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type observation struct {
	Exit   int    `json:"exit"`
	Stdout string `json:"stdout"`
	Stderr string `json:"stderr"`
	Error  string `json:"error,omitempty"`
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
	Corpus string      `json:"corpus"`
	Class  string      `json:"class"`
	Base   observation `json:"base_compile"`
	Head   observation `json:"compile"`
}
type report struct {
	Base          string  `json:"base"`
	Head          string  `json:"head"`
	GeneratorBlob string  `json:"generator_blob"`
	ManifestBlob  string  `json:"manifest_blob"`
	Admitted      int     `json:"admitted"`
	Programs      []entry `json:"programs"`
	Verdict       string  `json:"verdict"`
}

func execute(dir string, limit time.Duration, name string, args ...string) observation {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	command := exec.CommandContext(ctx, name, args...)
	command.Dir = dir
	var out, errout bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errout
	err := command.Run()
	o := observation{Stdout: out.String(), Stderr: errout.String()}
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
	asJSON := flags.Bool("json", false, "emit JSON to stdout")
	limit := flags.Duration("timeout", 10*time.Second, "per-command timeout")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *base == "" {
		return fmt.Errorf("--base is required")
	}
	if *limit <= 0 {
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
	h, err := build(result.Head, "head-build", *headBinary)
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
		start := filepath.Join(headTree, *corpusDir)
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
		for _, p := range c.Programs {
			clean := filepath.ToSlash(filepath.Clean(p.Path))
			if filepath.IsAbs(p.Path) || clean == ".." || strings.HasPrefix(clean, "../") {
				return fmt.Errorf("unsafe program path %q", p.Path)
			}
			blob, e := git(root, "rev-parse", result.Head+":"+p.Path)
			if e != nil {
				return e
			}
			if blob != p.Blob {
				return fmt.Errorf("blob mismatch for %s: manifest %s head %s", p.Path, p.Blob, blob)
			}
			a := execute(headTree, *limit, b, "c", p.Path)
			z := execute(headTree, *limit, h, "c", p.Path)
			a.Stdout = ""
			z.Stdout = ""
			record := entry{program: p, Corpus: c.Name, Class: classify(a, z), Base: a, Head: z}
			if record.Class == "newly-accepted" {
				result.Admitted++
			}
			if record.Class == "compiler-error" {
				result.Verdict = "fail"
			}
			result.Programs = append(result.Programs, record)
		}
	}
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
	if result.Verdict != "pass" {
		return fmt.Errorf("admission delta failed")
	}
	return nil
}
