package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type wasiRequirement struct {
	Shard   int
	Name    string
	Command string
	Gates   []string
}

func wasiVariableLiteral(n ast.Node) bool {
	literal, ok := n.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(literal.Value)
	if err != nil || !strings.Contains(value, "WASI") {
		return false
	}
	for _, ch := range value {
		if !(ch >= 'A' && ch <= 'Z' || ch >= '0' && ch <= '9' || ch == '_') {
			return false
		}
	}
	return true
}

func hasSkip(n ast.Node) bool {
	found := false
	ast.Inspect(n, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok && (s.Sel.Name == "Skip" || s.Sel.Name == "Skipf" || s.Sel.Name == "SkipNow") {
				found = true
			}
		}
		return true
	})
	return found
}

// Recognize direct os.Getenv comparisons and local aliases, not arbitrary predicates.
// Unknown predicates fail closed: their dependence on the environment needs an audit.
func wasiGates(root string) (map[string]bool, error) {
	gates := map[string]bool{}
	moduleBytes, err := os.ReadFile(filepath.Join(root, "go.mod"))
	if err != nil {
		return nil, err
	}
	fields := strings.Fields(string(moduleBytes))
	if len(fields) < 2 || fields[0] != "module" {
		return nil, fmt.Errorf("cannot derive module")
	}
	module := fields[1]
	err = filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && (entry.Name() == ".git" || entry.Name() == "cohere" || entry.Name() == "vendor" || entry.Name() == "node_modules" || entry.Name() == "testdata") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			return err
		}
		osAlias := ""
		for _, imp := range f.Imports {
			if imp.Path.Value == `"os"` {
				osAlias = "os"
				if imp.Name != nil {
					osAlias = imp.Name.Name
				}
				if osAlias == "." {
					return fmt.Errorf("unsupported dot os import: %s", path)
				}
			}
		}
		getter := func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok {
				return false
			}
			sel, ok := c.Fun.(*ast.SelectorExpr)
			if !ok {
				return false
			}
			id, ok := sel.X.(*ast.Ident)
			return ok && id.Name == osAlias && (sel.Sel.Name == "Getenv" || sel.Sel.Name == "LookupEnv")
		}
		// Gates routed through global variables or helpers are intentionally unsupported.
		// Reject them rather than claiming a complete platform audit.
		for _, declaration := range f.Decls {
			fn, isFunction := declaration.(*ast.FuncDecl)
			if isFunction && strings.HasPrefix(fn.Name.Name, "Test") {
				continue
			}
			hidden := false
			ast.Inspect(declaration, func(n ast.Node) bool {
				if wasiVariableLiteral(n) {
					hidden = true
				}
				if getter(n) {
					c := n.(*ast.CallExpr)
					for _, arg := range c.Args {
						if lit, ok := arg.(*ast.BasicLit); ok {
							value, _ := strconv.Unquote(lit.Value)
							if strings.Contains(value, "WASI") {
								hidden = true
							}
						}
					}
				}
				return true
			})
			if hidden {
				return fmt.Errorf("WASI environment dependency outside a test in %s requires audit", path)
			}
		}
		for _, declaration := range f.Decls {
			fn, ok := declaration.(*ast.FuncDecl)
			if !ok || fn.Body == nil || !hasSkip(fn.Body) {
				continue
			}
			aliases := map[*ast.Object]bool{}
			badAliases := map[*ast.Object]bool{}
			calls := map[ast.Node]bool{}
			var bad error
			variableMention := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if wasiVariableLiteral(n) {
					variableMention = true
				}
				if call, ok := n.(*ast.CallExpr); ok && !getter(n) {
					for _, argument := range call.Args {
						if wasiVariableLiteral(argument) {
							bad = fmt.Errorf("unrecognized WASI variable access in %s::%s", path, fn.Name.Name)
						}
					}
					if selector, ok := call.Fun.(*ast.SelectorExpr); ok && (selector.Sel.Name == "Getenv" || selector.Sel.Name == "LookupEnv") {
						bad = fmt.Errorf("unrecognized environment getter in %s::%s", path, fn.Name.Name)
					}
				}

				if getter(n) {
					c := n.(*ast.CallExpr)
					if len(c.Args) != 1 {
						bad = fmt.Errorf("unknown environment gate in %s::%s", path, fn.Name.Name)
						return true
					}
					lit, ok := c.Args[0].(*ast.BasicLit)
					if !ok {
						bad = fmt.Errorf("dynamic environment skip gate in %s::%s", path, fn.Name.Name)
						return true
					}
					value, _ := strconv.Unquote(lit.Value)
					if strings.Contains(value, "WASI") {
						calls[n] = true
						if c.Fun.(*ast.SelectorExpr).Sel.Name != "Getenv" {
							bad = fmt.Errorf("unrecognized WASI LookupEnv skip gate in %s::%s", path, fn.Name.Name)
						}
					}
				}
				return true
			})
			if bad != nil {
				return bad
			}
			mentions := func(n ast.Node) bool {
				yes := false
				ast.Inspect(n, func(n ast.Node) bool {
					if calls[n] {
						yes = true
					}
					if id, ok := n.(*ast.Ident); ok && id.Obj != nil && aliases[id.Obj] {
						yes = true
					}
					return true
				})
				return yes
			}
			// Propagate aliases before examining predicates; a fixed point handles indirect aliases.
			for changed := true; changed; {
				changed = false
				ast.Inspect(fn.Body, func(n ast.Node) bool {
					if a, ok := n.(*ast.AssignStmt); ok {
						for i, rhs := range a.Rhs {
							// Only aliases of environment predicates are supported, not command outputs.
							aliasExpression := true
							switch rhs.(type) {
							case *ast.CompositeLit, *ast.FuncLit:
								aliasExpression = false
							}
							if aliasExpression && mentions(rhs) && i < len(a.Lhs) {
								invalid := false
								ast.Inspect(rhs, func(n ast.Node) bool {
									if call, ok := n.(*ast.CallExpr); ok && !calls[call] {
										invalid = true
									}
									if id, ok := n.(*ast.Ident); ok && id.Obj != nil && badAliases[id.Obj] {
										invalid = true
									}
									return true
								})
								if id, ok := a.Lhs[i].(*ast.Ident); ok && id.Obj != nil && invalid && !badAliases[id.Obj] {
									badAliases[id.Obj] = true
									changed = true
								}
								if id, ok := a.Lhs[i].(*ast.Ident); ok && id.Obj != nil && !aliases[id.Obj] {
									aliases[id.Obj] = true
									changed = true
								}
							}
						}
					}
					return true
				})
			}
			if bad != nil {
				return bad
			}
			gated := false
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				condition, ok := n.(*ast.IfStmt)
				if !ok || !hasSkip(condition.Body) || !mentions(condition.Cond) {
					return true
				}
				valid := true
				ast.Inspect(condition.Cond, func(n ast.Node) bool {
					if n == nil {
						return true
					}
					switch x := n.(type) {
					case *ast.Ident:
						if x.Obj != nil && badAliases[x.Obj] {
							valid = false
						}
					case *ast.BinaryExpr:
						if x.Op != token.EQL && x.Op != token.NEQ && x.Op != token.LAND && x.Op != token.LOR {
							valid = false
						}
					case *ast.CallExpr:
						if !calls[x] {
							valid = false
						}
					case *ast.UnaryExpr:
						valid = false
					}
					return true
				})
				if !valid {
					bad = fmt.Errorf("unrecognized WASI skip condition in %s::%s", path, fn.Name.Name)
				} else {
					gated = true
				}
				return true
			})
			if bad != nil {
				return bad
			}
			// A WASI getter feeding any skip-bearing function must have a recognized gate.
			// This also rejects gates hidden in helpers instead of silently losing their callers.
			if (len(calls) > 0 || variableMention) && !gated {
				return fmt.Errorf("unrecognized WASI skip dependency in %s::%s", path, fn.Name.Name)
			}
			if gated {
				if !strings.HasPrefix(fn.Name.Name, "Test") || fn.Recv != nil {
					return fmt.Errorf("WASI gate in helper %s::%s requires audit", path, fn.Name.Name)
				}
				relative, _ := filepath.Rel(root, filepath.Dir(path))
				pkg := module
				if relative != "." {
					pkg += "/" + filepath.ToSlash(relative)
				}
				gates[pkg+"::"+fn.Name.Name] = true
			}
		}
		return nil
	})
	return gates, err
}

func requireWASI(p *plan) error {
	gates, err := wasiGates(".")
	if err != nil {
		return err
	}
	if len(gates) == 0 {
		return nil
	}
	requirement := &wasiRequirement{Shard: p.Count - 1, Name: "required WASI", Command: `bash cloud/setup.sh --wasi-sdk && source "${ADAMIC_TOOLS:-/opt/adamic-tools}/env.sh" && ADAMIC_TEST_WASI=1 ADAMIC_ORACLE_WASI=1 adamic-gate shard -index ` + strconv.Itoa(p.Count-1) + " -count " + strconv.Itoa(p.Count) + " -out <dir>"}
	found := map[string]bool{}
	for i := range p.Units {
		u := &p.Units[i]
		parent, _, _ := strings.Cut(u.Test, "/")
		key := u.Package + "::" + parent
		if gates[key] {
			u.WASI = true
			found[key] = true
		}
	}
	for key := range gates {
		if !found[key] {
			return fmt.Errorf("WASI gate not enumerated by go test -list: %s", key)
		}
		requirement.Gates = append(requirement.Gates, key)
	}
	sort.Strings(requirement.Gates)
	p.WASI = requirement
	return nil
}

func wasiReady() error {
	for _, name := range []string{"ADAMIC_TEST_WASI", "ADAMIC_ORACLE_WASI"} {
		if os.Getenv(name) != "1" {
			return fmt.Errorf("required WASI shard refuses to start: %s=1 required; run bash cloud/setup.sh --wasi-sdk and source its env.sh", name)
		}
	}
	sysroot := os.Getenv("WASI_SYSROOT")
	if sysroot == "" {
		return fmt.Errorf("required WASI shard refuses to start: WASI_SYSROOT missing from setup env.sh")
	}
	headerFound := false
	for _, header := range []string{filepath.Join(sysroot, "include", "stdlib.h"), filepath.Join(sysroot, "include", "wasm32-wasi", "stdlib.h")} {
		if info, err := os.Stat(header); err == nil && info.Mode().IsRegular() {
			headerFound = true
		}
	}
	if !headerFound {
		return fmt.Errorf("required WASI shard refuses to start: SDK stdlib.h missing in %s", sysroot)
	}
	for _, path := range []string{filepath.Join(sysroot, "lib", "wasm32-wasi", "libc.a"), filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "clang"), filepath.Join(filepath.Dir(filepath.Dir(sysroot)), "bin", "wasm-ld")} {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			return fmt.Errorf("required WASI shard refuses to start: SDK missing %s", path)
		}
		if (filepath.Base(path) == "clang" || filepath.Base(path) == "wasm-ld") && info.Mode().Perm()&0111 == 0 {
			return fmt.Errorf("WASI SDK tool is not executable: %s", path)
		}
	}
	return nil
}

func wasiSkips(p plan, index int, results []result) []string {
	var failures []string
	for _, r := range results {
		if r.Action != "skip" || r.Test == "" {
			continue
		}
		for _, c := range p.Complements {
			if c.Package != r.Package || !strings.HasPrefix(r.Test, c.Parent+"/") {
				continue
			}
			for _, u := range p.Units {
				parent, _, _ := strings.Cut(u.Test, "/")
				if u.Package == r.Package && u.WASI && parent == c.Parent {
					failures = append(failures, fmt.Sprintf("shard %d required WASI unit %s skipped: %s", index, r.key(), r.Reason))
					break
				}
			}
		}
		for _, u := range p.Units {
			if !u.WASI || u.Package != r.Package {
				continue
			}
			if r.Test == u.Test || strings.HasPrefix(r.Test, u.Test+"/") || strings.HasPrefix(u.Test, r.Test+"/") {
				failures = append(failures, fmt.Sprintf("shard %d required WASI unit %s skipped: %s (%s)", index, u.key(), r.key(), r.Reason))
			}
		}
	}
	return failures
}
