package main

// This scout command models ownership in Go. It does not add a production facts API.
import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	bridge "github.com/system-inc/adamic/bridge/tsgo/checker"
)

type censusPin struct{ Repo, SHA, Root string }
type censusRow struct {
	Repo, SHA, Path, Digest                     string
	Bytes, ParseDiagnostics, Options, TypeShape int
	Declaration                                 bool
}

func continuation(args []string) bool {
	if len(args) == 0 {
		return false
	}
	if args[0] == "census" {
		if len(args) != 3 {
			panic("census pins.json output.jsonl")
		}
		data, err := os.ReadFile(args[1])
		must(err)
		var pins []censusPin
		must(json.Unmarshal(data, &pins))
		out, err := os.Create(args[2])
		must(err)
		defer out.Close()
		enc := json.NewEncoder(out)
		for _, pin := range pins {
			count := 0
			must(filepath.WalkDir(pin.Root, func(path string, d os.DirEntry, e error) error {
				if e != nil {
					return e
				}
				if d.IsDir() {
					if d.Name() == ".git" {
						return filepath.SkipDir
					}
					return nil
				}
				if !strings.HasSuffix(path, ".ts") && !strings.HasSuffix(path, ".tsx") {
					return nil
				}
				text, e := os.ReadFile(path)
				if e != nil {
					return e
				}
				name := tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path))
				kind := core.ScriptKindTS
				if strings.HasSuffix(path, ".tsx") {
					kind = core.ScriptKindTSX
				}
				f := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: name, PathKey: tspath.CaseSensitive.PathKey(tspath.RootedPathFromAbsolute(filepath.ToSlash(path)))}, string(text), kind)
				asks := planSource(f, path)
				rel, e := filepath.Rel(pin.Root, path)
				if e != nil {
					return e
				}
				row := censusRow{Repo: pin.Repo, SHA: pin.SHA, Path: filepath.ToSlash(rel), Digest: fmt.Sprintf("%x", sha256.Sum256(text)), Bytes: len(text), ParseDiagnostics: len(f.Diagnostics()), Declaration: strings.HasSuffix(path, ".d.ts")}
				// A syntax-error row remains visible but is not represented as an executed rule.
				if row.ParseDiagnostics == 0 {
					row.Options = 1
					row.TypeShape = len(asks) - 1
				}
				if e = enc.Encode(row); e != nil {
					return e
				}
				count++
				return nil
			}))
			if count == 0 {
				panic("empty corpus: " + pin.Repo)
			}
			fmt.Fprintf(os.Stderr, "%s %s files=%d\n", pin.Repo, pin.SHA, count)
		}
		return true
	}
	if args[0] == "experiment" {
		if len(args) != 4 {
			panic("experiment config manifest output.json")
		}
		experiment(args[1], args[2], args[3])
		return true
	}
	return false
}
func must(err error) {
	if err != nil {
		panic(err)
	}
}

type ownerHandle struct {
	Owner *checker.Checker
	ID    uint64
}
type ownerTable struct {
	Types        map[*checker.Type]uint64
	Values       []*checker.Type
	Symbols      map[*ast.Symbol]uint64
	SymbolValues []*ast.Symbol
}

func (t *ownerTable) intern(c *checker.Checker, v *checker.Type) ownerHandle {
	id, ok := t.Types[v]
	if !ok {
		id = uint64(len(t.Values))
		t.Types[v] = id
		t.Values = append(t.Values, v)
	}
	return ownerHandle{c, id}
}
func (t *ownerTable) read(c *checker.Checker, h ownerHandle) (*checker.Type, error) {
	if h.Owner != c || h.ID >= uint64(len(t.Values)) {
		return nil, fmt.Errorf("foreign or stale handle")
	}
	return t.Values[h.ID], nil
}

type symbolHandle struct {
	Owner *checker.Checker
	ID    uint64
}

func (t *ownerTable) internSymbol(c *checker.Checker, value *ast.Symbol) symbolHandle {
	id, ok := t.Symbols[value]
	if !ok {
		id = uint64(len(t.SymbolValues))
		t.Symbols[value] = id
		t.SymbolValues = append(t.SymbolValues, value)
	}
	return symbolHandle{c, id}
}
func (t *ownerTable) readSymbol(c *checker.Checker, h symbolHandle) (*ast.Symbol, error) {
	if h.Owner != c || h.ID >= uint64(len(t.SymbolValues)) {
		return nil, fmt.Errorf("foreign or stale symbol")
	}
	return t.SymbolValues[h.ID], nil
}
func probeNode(f *ast.SourceFile) *ast.Node {
	var found *ast.Node
	var visit func(*ast.Node)
	visit = func(node *ast.Node) {
		if found != nil {
			return
		}
		if node.Kind == ast.KindIdentifier {
			found = node
			return
		}
		node.ForEachChild(func(child *ast.Node) bool { visit(child); return found != nil })
	}
	visit(f.AsNode())
	if found == nil {
		return ast.GetNodeAtPosition(f, len(f.Text())/2, false)
	}
	return found
}

type experimentResult struct {
	GoVersion                                  string
	Cores                                      int
	Repetitions, Questions, AnswerBytes, Files int
	RecomputeNS, CacheNS                       int64
	Pools                                      []poolResult
}
type poolResult struct {
	Requested, Observed, TypeIDs, SymbolIDs, ForeignMutants int
	BuildNS, PerQuestionNS, PerFileNS                       int64
}

func experiment(config, manifest, out string) {
	data, err := os.ReadFile(manifest)
	must(err)
	paths := strings.Split(strings.TrimSpace(string(data)), "\n")
	p, err := bridge.Open(config, nil)
	must(err)
	asks := []request{}
	for _, path := range paths {
		q, e := plan(p, path)
		must(e)
		asks = append(asks, q...)
	}
	result := experimentResult{GoVersion: runtime.Version(), Cores: runtime.GOMAXPROCS(0), Repetitions: 1000, Questions: len(asks), Files: len(paths)}
	cache := map[string]string{}
	for _, q := range asks {
		v, e := p.Inspect(q.File, q.Start, q.End, q.Kind, q.Question)
		must(e)
		cache[key(q)] = v
		result.AnswerBytes += len(v)
	}
	start := time.Now()
	for r := 0; r < result.Repetitions; r++ {
		for _, q := range asks {
			v, e := p.Inspect(q.File, q.Start, q.End, q.Kind, q.Question)
			must(e)
			if v != cache[key(q)] {
				panic("recomputed answer drift")
			}
		}
	}
	result.RecomputeNS = time.Since(start).Nanoseconds()
	start = time.Now()
	for r := 0; r < result.Repetitions; r++ {
		for _, q := range asks {
			if cache[key(q)] == "" {
				panic("cache miss")
			}
		}
	}
	result.CacheNS = time.Since(start).Nanoseconds()
	baseline := map[string]string{}
	for _, n := range []int{1, 4, 12} {
		// NewProgram shares the parsed configuration/host, but owns a fresh checker pool.
		cfg := p.Compiler.CommandLine()
		cfg.CompilerOptions().Checkers = &n
		start = time.Now()
		program := compiler.NewProgram(compiler.ProgramOptions{Config: cfg, Host: p.Compiler.Host(), SingleThreaded: core.TSFalse})
		tables := map[*checker.Checker]*ownerTable{}
		files := []*ast.SourceFile{}
		want := map[*ast.SourceFile]string{}
		for _, path := range paths {
			f := program.GetSourceFile(tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
			if f == nil {
				panic("missing file")
			}
			files = append(files, f)
			c, release := program.GetTypeCheckerForFileExclusive(context.Background(), f)
			table := tables[c]
			if table == nil {
				table = &ownerTable{Types: map[*checker.Type]uint64{}, Symbols: map[*ast.Symbol]uint64{}}
				tables[c] = table
			}
			node := probeNode(f)
			h := table.intern(c, c.GetTypeAtLocation(node))
			v, e := table.read(c, h)
			must(e)
			want[f] = c.TypeToString(v)
			if n == 1 {
				baseline[path] = want[f]
			} else if baseline[path] != want[f] {
				panic("cross-pool answer drift")
			}
			sym := c.GetSymbolAtLocation(node)
			if sym != nil {
				h := table.internSymbol(c, sym)
				value, e := table.readSymbol(c, h)
				must(e)
				if value != sym {
					panic("symbol round trip")
				}
			}

			release()
		}
		row := poolResult{Requested: n, Observed: len(tables), BuildNS: time.Since(start).Nanoseconds()}
		for c, t := range tables {
			row.TypeIDs += len(t.Values)
			row.SymbolIDs += len(t.SymbolValues)
			for other := range tables {
				if other != c {
					_, e := t.read(c, ownerHandle{other, 0})
					if e == nil {
						panic("owner mutant survived")
					}
					row.ForeignMutants++
					if len(t.SymbolValues) > 0 {
						_, e := t.readSymbol(c, symbolHandle{other, 0})
						if e == nil {
							panic("symbol owner mutant survived")
						}
						row.ForeignMutants++
					}
				}
			}
		}
		// Multiple goroutines exercise the compiler's actual exclusive leases. Each
		// question resolves the same node; batching holds one lease for all 1000.
		measure := func(batch bool) int64 {
			start := time.Now()
			var wg sync.WaitGroup
			for _, file := range files {
				wg.Add(1)
				go func(f *ast.SourceFile) {
					defer wg.Done()
					node := probeNode(f)
					var c *checker.Checker
					var release func()
					if batch {
						c, release = program.GetTypeCheckerForFileExclusive(context.Background(), f)
						defer release()
					}
					for r := 0; r < result.Repetitions; r++ {
						if !batch {
							c, release = program.GetTypeCheckerForFileExclusive(context.Background(), f)
						}
						h := tables[c].intern(c, c.GetTypeAtLocation(node))
						typed, e := tables[c].read(c, h)
						must(e)
						v := c.TypeToString(typed)
						if v != want[f] {
							panic("checker answer drift")
						}
						if !batch {
							release()
						}
					}
				}(file)
			}
			wg.Wait()
			return time.Since(start).Nanoseconds()
		}
		row.PerQuestionNS = measure(false)
		row.PerFileNS = measure(true)
		result.Pools = append(result.Pools, row)
	}
	for i := range asks {
		asks[i].Value = cache[key(asks[i])]
	}
	must(os.WriteFile(out+".requests", []byte(encodeRequests(asks)), 0644))
	encoded, err := json.MarshalIndent(result, "", "  ")
	must(err)
	must(os.WriteFile(out, append(encoded, '\n'), 0644))
}
