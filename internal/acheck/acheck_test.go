package acheck

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"go/build/constraint"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// repository is the module's root, where the gate ran adamic c from, each program named relative to it.
const repository = "../.."

// knownFile lists the programs outside a package that broke their first line's promise when this test came (#7xsbxbk,
// 53 on Oct 18), one a line as "<path>: <how it broke>", so a listed program that starts breaking another way is named
// rather than sheltered (Node's must on 23a97e701c). It only shrinks (Node's ruling): a listed program that keeps its
// promise again is named and must come off, and none is added. Each is its owner's to settle: its header updated, its directory exempted
// as evidence, or the compiler's regression fixed.
const knownFile = "known.txt"

// evidenceDirectories hold dated evidence of what the compiler did on a given day, not live fixtures, so their programs
// are exempt by directory (@system_adamic's ruling on #7xsbxbk); a live fixture of the same behavior belongs in
// internal/oracle/testdata, which stays checked. TestTheEvidenceExemptionIsThoseThreeDirectories keeps it from spreading.
var evidenceDirectories = []string{"docs/verification", "notes", "review"}

// TestProduct_Adamic declares the stage 0 compiler this package's test runs, so its unit is handed it; internal/native's
// TestProduct_WASICompiler builds it (buildcache.AdamicOwner), and this one finds it.
func TestProduct_Adamic(t *testing.T) {
	t.Parallel()
	buildcache.AdamicBuiltElsewhere(t)
}

// Every .a program outside a Go package keeps the promise its first line makes, through stage 0's front end (adamic c:
// the checker's proven types, then the refusal pass, stopping before C): a clean result or a "can't lower ... yet" stop
// passes, since nothing was refused, and a program whose first line is "// a-check: refused <rule>" or "// a-check: type
// error <code>" must fail exactly that way instead. A program inside a package is that package's to test. This was the
// gate's a-check phase, which checked only the programs a change touched, so a compiler change that broke one it didn't
// touch was never seen; here every program is checked whenever one of them or the compiler moves. The programs that
// broke their promise when it came are known.txt's, which only shrinks; the evidence directories' are exempt.
func TestEveryProgramOutsideAPackageKeepsItsPromise(t *testing.T) {
	t.Parallel()
	adamic := buildcache.Adamic(t)
	programs, err := programsOutsidePackages(repository)
	if err != nil {
		t.Fatal(err)
	}
	programs = slices.DeleteFunc(programs, evidence)
	if len(programs) == 0 {
		t.Fatal("no .a program outside a package was found, so nothing was checked")
	}
	broken := map[string]string{}
	var mutex sync.Mutex
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	work := make(chan string)
	var workers sync.WaitGroup
	for range 8 {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for program := range work {
				if why := stable(checked(t, adamic, program), root); why != "" {
					mutex.Lock()
					broken[program] = why
					mutex.Unlock()
				}
			}
		}()
	}
	for _, program := range programs {
		work <- program
	}
	close(work)
	workers.Wait()
	known, err := readKnown(knownFile)
	if err != nil {
		t.Fatal(err)
	}
	for _, problem := range compare(programs, broken, known) {
		t.Error(problem)
	}
}

// The evidence exemption is exactly docs/verification, notes and review, and a program is exempt only under one of
// them, never a directory whose name merely starts the same: what @system_adamic ruled, so it can't spread. Mutants that
// each fail it: a fourth directory exempted; a prefix taken as a directory.
func TestTheEvidenceExemptionIsThoseThreeDirectories(t *testing.T) {
	t.Parallel()
	if want := []string{"docs/verification", "notes", "review"}; !slices.Equal(evidenceDirectories, want) {
		t.Fatalf("exempt %q, want %q", evidenceDirectories, want)
	}
	for program, exempt := range map[string]bool{"notes/devirtualize/constructor_refused.a": true, "review/oct6/a.a": true,
		"docs/verification/enums-open/probe.a": true, "notes2/a.a": false, "docs/other/a.a": false, "reviewer/a.a": false,
		"stage3/drivers/notes/a.a": false} {
		if evidence(program) != exempt {
			t.Errorf("%s exempt %v, want %v", program, evidence(program), exempt)
		}
	}
}

// evidence says whether a program sits under one of evidenceDirectories.
func evidence(program string) bool {
	return slices.ContainsFunc(evidenceDirectories, func(directory string) bool { return strings.HasPrefix(program, directory+"/") })
}

// The test finds the programs it must and judges them as the gate did: a planted repository's programs outside a
// package are found, a Go file under testdata making no package, and those in a package, its testdata, a submodule or
// node_modules are not; each outcome is held to its first line; and known.txt only shrinks, naming a broken program it
// doesn't list, a listed one that breaks another way than listed, a listed one that keeps its promise again, and a
// listed one that's gone; and how a program broke is the same on every box, its absolute path taken off. Mutants that each fail it: a
// Go file under testdata taken as a package; a program in a package's subdirectory taken as outside it; a submodule or
// node_modules walked; a "can't lower" stop judged a failure; a refusal or a type error judged by its kind alone,
// without its rule or code; a listed program compared by its path alone; a listed program that keeps its promise
// passing; a gone listed path passing; the absolute path kept in how a program broke.
func TestTheCheckFindsAndJudgesAsTheGateDid(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	plant := func(path, content string) {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	plant("go.mod", "module example.com/a\n")
	plant("stage3/drivers/a.a", "let a = 1;\n")
	plant("notes/deep/b.a", "// a-check: refused capture\n")
	plant("pkg/p.go", "package p\n")
	plant("pkg/testdata/owned.a", "owned\n")
	plant("pkg/sub/owned.a", "owned\n")
	plant("tests/only/only_test.go", "package only\n")
	plant("tests/only/owned.a", "owned\n")
	plant("cohere/.git", "gitdir: elsewhere\n")
	plant("cohere/x.a", "a submodule's\n")
	plant("stage3/node_modules/y.a", "an install's\n")
	plant("notes/testdata/fixture.go", "package fixture\n")
	plant("notes/testdata/c.a", "no package holds it\n")
	programs, err := programsOutsidePackages(root)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"notes/deep/b.a", "notes/testdata/c.a", "stage3/drivers/a.a"}; !slices.Equal(programs, want) {
		t.Fatalf("found %q, want %q", programs, want)
	}
	for _, judged := range []struct {
		name, expect, stderr string
		code                 int
		broken               bool
	}{
		{"a clean check", "", "", 0, false},
		{"a stop before C", "", "adamic: x.a:1:1: stage 0 can't lower node:process.Process.nextTick yet", 1, false},
		{"a refusal unpromised", "", "adamic: x.a:1:1: Adamic 0.1 refuses 'callback', a variable", 1, true},
		{"the refusal promised", "refused 'callback'", "adamic: x.a:1:1: Adamic 0.1 refuses 'callback', a variable", 1, false},
		{"another refusal than promised", "refused 'other'", "adamic: x.a:1:1: Adamic 0.1 refuses 'callback', a variable", 1, true},
		{"the type error promised", "type error TS2591", "x.a(1,1): error TS2591: Cannot find name", 1, false},
		{"a type error promised and a clean check", "type error TS2591", "", 0, true},
		{"another type error than promised", "type error TS2591", "x.a(1,1): error TS2304: Cannot find name", 1, true},
		{"a failure", "", "adamic: panic", 2, true},
	} {
		if why := judge(judged.expect, judged.code, judged.stderr); (why != "") != judged.broken {
			t.Errorf("%s: %q, want broken %v", judged.name, why, judged.broken)
		}
	}
	if why := stable("refused (adamic: "+root+"/stage3/a.a:1:28: Adamic 0.1 refuses 'callback')", root); why != "refused (adamic: stage3/a.a:1:28: Adamic 0.1 refuses 'callback')" {
		t.Errorf("stable: %q, want the repository's path off it", why)
	}
	programs = []string{"a.a", "b.a", "c.a", "d.a"}
	problems := compare(programs, map[string]string{"a.a": "refused", "b.a": "refused", "d.a": "type error"},
		map[string]string{"b.a": "refused", "c.a": "refused", "d.a": "refused", "gone.a": "refused"})
	want := []string{
		"a.a breaks its first line's promise: refused",
		"c.a keeps its promise now: take it off internal/acheck/known.txt",
		"d.a breaks differently now: refused -> type error",
		"gone.a is on internal/acheck/known.txt and isn't a program outside a package: take it off",
	}
	if !slices.Equal(problems, want) {
		t.Fatalf("named:\n%s\nwant:\n%s", strings.Join(problems, "\n"), strings.Join(want, "\n"))
	}
}

// A directory is a package only by a Go file some box builds, as go list had it: one whose only Go is //go:build ignore
// holds none of its programs (stage3/drivers/scanner's, which build-metrics.go left unchecked), while a linux-only file
// makes a package on every box, darwin too, as do !linux, linux && !darwin and ignore || linux, each true under some
// tags, while linux && !linux, true under none, makes none; and a plain package holds its programs and its
// subdirectories'. Mutants that fail it: the //go:build line not read, every Go file a package; the line evaluated with
// every tag but ignore set (Node's reading), which drops the !linux file's package.
func TestAnIgnoredGoFileMakesNoPackage(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	plant := func(path, content string) {
		full := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	plant("go.mod", "module example.com/a\n")
	plant("ignored/metrics.go", "//go:build ignore\n\n// Run this with go run.\npackage main\n")
	plant("ignored/outside.a", "let a = 1;\n")
	plant("linux/only.go", "//go:build linux\n\npackage linux\n")
	plant("linux/owned.a", "owned\n")
	plant("notlinux/only.go", "//go:build !linux\n\npackage notlinux\n")
	plant("notlinux/owned.a", "owned\n")
	plant("linuxnotdarwin/only.go", "//go:build linux && !darwin\n\npackage linuxnotdarwin\n")
	plant("linuxnotdarwin/owned.a", "owned\n")
	plant("ignoreorlinux/only.go", "//go:build ignore || linux\n\npackage ignoreorlinux\n")
	plant("ignoreorlinux/owned.a", "owned\n")
	plant("never/only.go", "//go:build linux && !linux\n\npackage never\n")
	plant("never/outside.a", "let a = 1;\n")
	plant("plain/p.go", "// Package p is plain.\npackage p\n")
	plant("plain/owned.a", "owned\n")
	plant("plain/sub/owned.a", "owned\n")
	programs, err := programsOutsidePackages(root)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"ignored/outside.a", "never/outside.a"}; !slices.Equal(programs, want) {
		t.Fatalf("found %q, want %q", programs, want)
	}
}

// programsOutsidePackages is every .a program under root that no Go package holds, sorted, relative to root: none in a
// package's directory or under it (its testdata, its subdirectories), as the gate's ownership has it, nor in a
// submodule (a directory holding its own .git) or node_modules, which aren't the module's. A directory is a package only
// by a .go file some box builds (buildable), as go list had it, so one whose only Go is //go:build ignore holds nothing.
func programsOutsidePackages(root string) ([]string, error) {
	packages, programs := map[string]bool{}, []string{}
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch {
			case entry.Name() == ".git" || entry.Name() == "node_modules":
				return filepath.SkipDir
			case relative != ".":
				if _, err := os.Lstat(filepath.Join(path, ".git")); err == nil {
					return filepath.SkipDir
				}
			}
			return nil
		}
		switch filepath.Ext(entry.Name()) {
		case ".go":
			if !slices.Contains(strings.Split(filepath.ToSlash(relative), "/"), "testdata") && buildable(path) {
				packages[filepath.ToSlash(filepath.Dir(relative))] = true
			}
		case ".a":
			programs = append(programs, filepath.ToSlash(relative))
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	outside := []string{}
	for _, program := range programs {
		owned := false
		for directory := filepath.ToSlash(filepath.Dir(program)); ; directory = filepath.ToSlash(filepath.Dir(directory)) {
			if packages[directory] {
				owned = true
				break
			}
			if directory == "." {
				break
			}
		}
		if !owned {
			outside = append(outside, program)
		}
	}
	slices.Sort(outside)
	return outside, nil
}

// buildable says whether the Go file at path builds on some box: its //go:build line, read from its header before
// package as go/build reads it, holds under some assignment of its tags with ignore false (satisfiable, Node's ruling),
// so a darwin-, linux- or !linux-only file counts on every box alike and only one no box builds, ignore's, doesn't. A
// file with no
// //go:build line counts, a // +build line alone too: that's the syntax before Go 1.17, which gofmt rewrites into a
// //go:build line, so none is left to read here. A file whose header or //go:build line doesn't parse counts, as go
// list still lists its package, with the error.
func buildable(path string) bool {
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.PackageClauseOnly|parser.ParseComments)
	if err != nil {
		return true
	}
	for _, group := range file.Comments {
		if group.Pos() >= file.Package {
			break
		}
		for _, comment := range group.List {
			if !constraint.IsGoBuild(comment.Text) {
				continue
			}
			expression, err := constraint.Parse(comment.Text)
			if err != nil {
				return true
			}
			return satisfiable(expression)
		}
	}
	return true
}

// satisfiableTags is the most distinct tags satisfiable tries every assignment of; a constraint naming more counts as
// buildable, since a file read as a package only leaves its programs to its package, never checks one it shouldn't.
const satisfiableTags = 16

// satisfiable says whether expression holds under some assignment of the tags it names, ignore held false: a file only
// ignore gates, or one false under every assignment, is one no box builds.
func satisfiable(expression constraint.Expr) bool {
	tags := []string{}
	var collect func(constraint.Expr)
	collect = func(expression constraint.Expr) {
		switch node := expression.(type) {
		case *constraint.TagExpr:
			if node.Tag != "ignore" && !slices.Contains(tags, node.Tag) {
				tags = append(tags, node.Tag)
			}
		case *constraint.NotExpr:
			collect(node.X)
		case *constraint.AndExpr:
			collect(node.X)
			collect(node.Y)
		case *constraint.OrExpr:
			collect(node.X)
			collect(node.Y)
		}
	}
	collect(expression)
	if len(tags) > satisfiableTags {
		return true
	}
	for assignment := range 1 << len(tags) {
		if expression.Eval(func(tag string) bool {
			index := slices.Index(tags, tag)
			return index >= 0 && assignment&(1<<index) != 0
		}) {
			return true
		}
	}
	return false
}

// checked runs adamic c on program from the repository's root, as the gate did, and says how it broke its first
// line's promise, "" when it kept it.
func checked(t *testing.T, adamic, program string) string {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(repository, program))
	if err != nil {
		return err.Error()
	}
	header, _, _ := strings.Cut(string(content), "\n")
	expect := ""
	if rest, promised := strings.CutPrefix(strings.TrimSpace(header), "// a-check:"); promised {
		expect = strings.TrimSpace(rest)
	}
	command := exec.Command(adamic, "c", program)
	command.Dir = repository
	var stderr bytes.Buffer
	command.Stderr = &stderr
	code := 0
	if err := command.Run(); err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			return "running adamic c: " + err.Error()
		}
		code = exit.ExitCode()
	}
	return judge(expect, code, stderr.String())
}

// judge is the gate's rule for one program's run, from its promise (what its first line names after "// a-check:"),
// its exit code and its standard error, stderr: why it broke the promise, "" when it kept it.
func judge(expect string, code int, stderr string) string {
	outcome := "failed"
	switch {
	case code == 0 || strings.Contains(stderr, "can't lower") && strings.Contains(stderr, " yet"):
		outcome = "checked"
	case strings.Contains(stderr, "Adamic 0.1 refuses"):
		outcome = "refused"
	case strings.Contains(stderr, " error TS"):
		outcome = "type error"
	}
	kept := false
	switch {
	case strings.HasPrefix(expect, "refused"):
		kept = outcome == "refused" && strings.Contains(stderr, strings.TrimSpace(strings.TrimPrefix(expect, "refused")))
	case strings.HasPrefix(expect, "type error"):
		code := strings.TrimSpace(strings.TrimPrefix(expect, "type error"))
		kept = outcome == "type error" && (code == "" || strings.Contains(stderr, "error "+code))
	default:
		kept = outcome == "checked"
	}
	if kept {
		return ""
	}
	first, _, _ := strings.Cut(strings.TrimSpace(stderr), "\n")
	why := outcome
	if expect != "" {
		why += ", and its first line promises " + expect
	}
	if first != "" {
		why += " (" + first + ")"
	}
	return why
}

// stable is how a program broke as known.txt holds it: the same on every box, the repository's absolute path, which
// adamic names in its first line, taken off.
func stable(why, root string) string {
	for _, prefix := range []string{root + string(filepath.Separator), resolved(root) + string(filepath.Separator)} {
		why = strings.ReplaceAll(why, prefix, "")
	}
	return why
}

// resolved is a path with every link on it followed, or the path itself when it can't be.
func resolved(path string) string {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		return real
	}
	return path
}

// readKnown reads known.txt, "<path>: <how it broke>" a line; blank lines and # comments are skipped.
func readKnown(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	known := map[string]string{}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 64*1024), 1<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		program, why, found := strings.Cut(line, ": ")
		if !found || why == "" {
			return nil, fmt.Errorf("%s: %q isn't \"<path>: <how it broke>\"", path, line)
		}
		known[program] = why
	}
	return known, scanner.Err()
}

// compare names, sorted, each program that broke its promise and isn't on known.txt, each listed program that broke
// another way than listed, each listed program that kept it, and each listed path that isn't a program outside a
// package any more.
func compare(programs []string, broken, known map[string]string) []string {
	problems := []string{}
	for _, program := range programs {
		why, breaks := broken[program]
		listed, isListed := known[program]
		switch {
		case breaks && !isListed:
			problems = append(problems, program+" breaks its first line's promise: "+why)
		case breaks && why != listed:
			problems = append(problems, program+" breaks differently now: "+listed+" -> "+why)
		case !breaks && isListed:
			problems = append(problems, program+" keeps its promise now: take it off internal/acheck/"+knownFile)
		}
	}
	for path := range known {
		if !slices.Contains(programs, path) {
			problems = append(problems, path+" is on internal/acheck/"+knownFile+" and isn't a program outside a package: take it off")
		}
	}
	slices.Sort(problems)
	return problems
}
