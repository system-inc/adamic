// Native proof for this unit's six mutants, using the unchanged unified lint entry.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/stage1/cohere/lint/registry"
)

var slugs = []string{"typescript-no-unused-expressions", "typescript-unified-signatures", "class-methods-use-this", "no-async-promise-executor", "no-case-declarations", "no-compare-neg-zero"}
var imports = regexp.MustCompile(`(?m)(^import\s+[^;]*?\s+from\s+)(['"])([^'"]+)(['"])`)

func must(err error) {
	if err != nil {
		panic(err)
	}
}
func read(path string) []byte { b, e := os.ReadFile(path); must(e); return b }
func write(path string, b []byte) {
	must(os.MkdirAll(filepath.Dir(path), 0755))
	must(os.WriteFile(path, b, 0644))
}
func run(directory, output, name string, args ...string) []byte {
	c := exec.Command(name, args...)
	c.Dir = directory
	f, e := os.Create(output)
	must(e)
	defer f.Close()
	c.Stdout = f
	var stderr bytes.Buffer
	c.Stderr = &stderr
	must(c.Run())
	if stderr.Len() != 0 {
		panic(stderr.String())
	}
	return read(output)
}
func build(entry, archive, binary string) {
	p, e := load.Load([]string{entry})
	must(e)
	p.EnableTSGo()
	l, e := lower.Lower(context.Background(), p)
	must(e)
	if native.UsesTSGo(l) {
		c, e := native.TSGoC(l)
		must(e)
		must(native.BuildSplitTSGo(c, binary, archive, native.Options{Sanitize: true, Jobs: 1}))
	} else {
		must(native.Build(native.C(l), binary, native.Options{Sanitize: true, Jobs: 1}))
	}
}
func main() {
	if len(os.Args) != 3 {
		panic("usage: native-proof <oracle binary> <scratch directory>")
	}
	root, e := filepath.Abs("stage1/cohere/lint")
	must(e)
	scratch, e := filepath.Abs(os.Args[2])
	must(e)
	must(os.MkdirAll(scratch, 0755))
	oracle, e := filepath.Abs(os.Args[1])
	must(e)
	_, e = registry.Generate(root)
	must(e)
	archive := filepath.Join(scratch, "checker.a")
	c := exec.Command("go", "build", "-buildmode=c-archive", "-o", archive, "./bridge/tsgo/archive")
	c.Env = append(os.Environ(), "GOMAXPROCS=4", "CC=clang", "CGO_CFLAGS=-O1 -g -fsanitize=address,undefined -fno-sanitize-recover=all")
	f, e := os.Create(filepath.Join(scratch, "archive.log"))
	must(e)
	c.Stdout = f
	c.Stderr = f
	must(c.Run())
	must(f.Close())
	baseline := filepath.Join(scratch, "baseline")
	build(filepath.Join(root, "main.ts"), archive, baseline)
	for _, slug := range slugs {
		var change struct{ Name, File, From, To string }
		must(json.Unmarshal(read(filepath.Join(root, "rules", slug, "mutant.json")), &change))
		if change.File == "" {
			change.File = "rule.a"
		}
		copyRoot := filepath.Join(scratch, slug)
		target := filepath.Join("rules", slug, change.File)
		changed := 0
		must(filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			rel, e := filepath.Rel(root, path)
			if e != nil {
				return e
			}
			if d.IsDir() {
				if rel == "gaps" {
					return filepath.SkipDir
				}
				return nil
			}
			if !(strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a") || strings.HasSuffix(path, "rule.json") || strings.HasSuffix(path, "mutant.json") || strings.HasSuffix(path, "oracle.go") || strings.Contains(filepath.ToSlash(rel), "/testdata/")) {
				return nil
			}
			data := string(read(path))
			if rel == target {
				if strings.Count(data, change.From) != 1 {
					panic("mutant anchor count")
				}
				data = strings.Replace(data, change.From, change.To, 1)
				changed++
			}
			if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".a") {
				data = imports.ReplaceAllStringFunc(data, func(s string) string {
					p := imports.FindStringSubmatch(s)
					if !strings.HasPrefix(p[3], ".") {
						return s
					}
					abs := filepath.Clean(filepath.Join(filepath.Dir(path), p[3]))
					within, e := filepath.Rel(root, abs)
					must(e)
					if within != ".." && !strings.HasPrefix(within, ".."+string(filepath.Separator)) {
						return s
					}
					return p[1] + p[2] + filepath.ToSlash(abs) + p[4]
				})
			}
			write(filepath.Join(copyRoot, rel), []byte(data))
			return nil
		}))
		if changed != 1 {
			panic("mutation missing")
		}
		_, e = registry.Generate(copyRoot)
		must(e)
		witness, e := registry.Witnesses(filepath.Join(root, "rules", slug))
		must(e)
		var descriptor registry.Descriptor
		for _, d := range mustDescriptors(root) {
			if d.Slug == slug {
				descriptor = d
			}
		}
		var rows []string
		for _, path := range witness {
			fixture := filepath.Join(scratch, slug+"-"+filepath.Base(strings.TrimSuffix(path, ".txt")))
			write(fixture, read(path))
			rows = append(rows, fixture+"\t"+descriptor.Name)
		}
		manifest := filepath.Join(scratch, slug+".manifest")
		write(manifest, []byte(strings.Join(rows, "\n")+"\n"))
		want := run("", filepath.Join(scratch, slug+".go.log"), oracle, "--manifest", manifest)
		good := run("", filepath.Join(scratch, slug+".baseline.log"), baseline, "--manifest", manifest)
		if !bytes.Equal(want, good) {
			panic(slug + ": native baseline differs from Go")
		}
		binary := filepath.Join(copyRoot, "scanner")
		build(filepath.Join(copyRoot, "main.ts"), archive, binary)
		got := run("", filepath.Join(scratch, slug+".mutant.log"), binary, "--manifest", manifest)
		if bytes.Equal(want, got) {
			panic(slug + ": native mutant survived")
		}
		fmt.Printf("PASS %s: %s caught on ASan/UBSan native; baseline equals Go; both exit 0 with empty stderr\n", slug, change.Name)
	}
}
func mustDescriptors(root string) []registry.Descriptor {
	d, e := registry.Generate(root)
	must(e)
	return d
}
