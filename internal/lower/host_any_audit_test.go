package lower

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/load"
)

// Host anys are reachable only along admitted member edges. Rule 800 gates a
// direct unregistered access; host-object coverage gates the same hop through a
// view. Library containers carry their arguments, not an exemption for a host
// object inside one. Library-owned any declarations have their separate audit.
func TestHostAnyAudit(test *testing.T) {
	test.Parallel()
	allowed := hostAnyModelled(test)
	audit := newHostAnyAudit(test)
	audit.run(test, allowed)
}

type hostAnyRoot struct {
	symbol      *ast.Symbol
	declaration *ast.Node
}

type hostAnyFinding struct {
	Root  string
	Path  string
	Where string
}

type hostAnyAudit struct {
	program  *load.Program
	checker  *checker.Checker
	members  map[string]bool
	roots    map[string][]hostAnyRoot
	types    map[*checker.Type]bool
	states   int
	findings []hostAnyFinding
}

func newHostAnyAudit(test *testing.T) *hostAnyAudit {
	test.Helper()
	file := filepath.Join(test.TempDir(), "host-any.a")
	// Loading the bundled Node index also loads the declarations of the registered
	// modules. Keep a module namespace to resolve compatibility registrations such
	// as node:process.cwd to the checker's actual Process.cwd declaration.
	source := `import * as auditProcess from 'node:process';
import 'node:fs'; import 'node:perf_hooks'; import 'node:console';
import 'node:buffer'; import 'node:tty'; import 'node:stream';
import 'node:net'; import 'node:path'; import 'node:os';`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		test.Fatal(err)
	}
	program, err := load.Load([]string{file})
	if err != nil {
		test.Fatal(err)
	}
	c, release := program.Checker(context.Background(), program.Files()[0])
	test.Cleanup(release)
	a := &hostAnyAudit{
		program: program,
		checker: c,
		members: map[string]bool{},
		roots:   map[string][]hostAnyRoot{},
		types:   map[*checker.Type]bool{},
	}
	for name, implemented := range implementedNodeMembers {
		if implemented {
			a.members[name] = true
		}
	}
	// These two reads are bound by name on process.stdout without registration.
	a.members["node:tty.WriteStream.isTTY"] = true
	a.members["node:tty.WriteStream.columns"] = true
	l := lowering{program: program, checker: c}
	seen := map[string]map[*ast.Node]bool{}
	add := func(name string, symbol *ast.Symbol, declaration *ast.Node) {
		if seen[name] == nil {
			seen[name] = map[*ast.Node]bool{}
		}
		if !seen[name][declaration] {
			seen[name][declaration] = true
			a.roots[name] = append(a.roots[name], hostAnyRoot{symbol, declaration})
		}
	}
	for _, file := range program.CompilerProgram().GetSourceFiles() {
		if !load.IsNodeLibrary(file) {
			continue
		}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if name := node.Name(); name != nil {
				if symbol := l.symbol(name); symbol != nil {
					for _, declaration := range symbol.Declarations {
						// Examine every merged declaration, using the production naming rule.
						one := &ast.Symbol{Name: symbol.Name, Declarations: []*ast.Node{declaration}}
						member := nodeLibraryName(one)
						if a.admitted(one) {
							add(member, symbol, declaration)
						}
					}
				}
			}
			node.ForEachChild(visit)
			return false
		}
		file.AsNode().ForEachChild(visit)
	}
	var processType *checker.Type
	var imports ast.Visitor
	imports = func(node *ast.Node) bool {
		if node.Kind == ast.KindNamespaceImport {
			processType = c.GetTypeAtLocation(node.Name())
		}
		node.ForEachChild(imports)
		return false
	}
	program.Files()[0].AsNode().ForEachChild(imports)
	for name := range a.members {
		if len(a.roots[name]) != 0 {
			continue
		}
		if strings.HasPrefix(name, "node:process.") && processType != nil {
			member := strings.TrimPrefix(name, "node:process.")
			if symbol := c.GetPropertyOfType(processType, member); symbol != nil {
				for _, declaration := range symbol.Declarations {
					if load.IsNodeLibrary(ast.GetSourceFileOfNode(declaration)) {
						add(name, symbol, declaration)
					}
				}
			}
		}
		if len(a.roots[name]) == 0 {
			test.Errorf("registered host member has no bundled declaration: %s", name)
		}
	}
	return a
}

// Modelled positions come only from the audit's Modelled section, never a
// parallel Go allow-list. Locations are declaration file:line, as in the table.
func hostAnyModelled(test *testing.T) map[string]bool {
	test.Helper()
	data, err := os.ReadFile("../../docs/axioms/host-any.md")
	if err != nil {
		test.Fatal(err)
	}
	found := false
	active := false
	allowed := map[string]bool{}
	location := regexp.MustCompile(`(?:[A-Za-z0-9_./-]+\.d\.ts):[1-9][0-9]*`)
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "## ") {
			active = strings.TrimSpace(line) == "## Modelled"
			found = found || active
		}
		if !active || !strings.HasPrefix(strings.TrimSpace(line), "|") {
			continue
		}
		for _, one := range location.FindAllString(line, -1) {
			allowed[one] = true
		}
	}
	if !found {
		test.Fatal("host-any.md has no Modelled section")
	}
	return allowed
}

func hostAnyLibraryFile(file *ast.SourceFile) bool {
	if file == nil {
		return false
	}
	name := file.FileName().AsString()
	return load.IsNodeLibrary(file) || strings.HasPrefix(name, "bundled:///node/node_modules/undici-types/")
}

func (a *hostAnyAudit) run(test *testing.T, allowed map[string]bool) {
	test.Helper()
	names := []string{}
	for name := range a.roots {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		type state struct {
			typ *checker.Type
			out bool
		}
		seen := map[state]bool{}
		findingSeen := map[string]bool{}
		var walk func(*checker.Type, bool, *ast.Node, string)
		walk = func(typ *checker.Type, out bool, site *ast.Node, path string) {
			if typ == nil {
				return
			}
			a.types[typ] = true
			if typ.Flags()&checker.TypeFlagsAny != 0 {
				if !out || site == nil || !hostAnyLibraryFile(ast.GetSourceFileOfNode(site)) {
					return
				}
				file := ast.GetSourceFileOfNode(site).FileName().AsString()
				file = strings.TrimPrefix(file, "bundled:///node/node_modules/@types/node/")
				file = strings.TrimPrefix(file, "bundled:///node/node_modules/")
				location := fmt.Sprintf("%s:%d", file, a.program.Line(site))
				if !allowed[location] && !findingSeen[location] {
					findingSeen[location] = true
					a.findings = append(a.findings, hostAnyFinding{name, path, a.program.Where(site)})
				}
				return
			}
			if typ.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
				for i, part := range typ.Types() {
					walk(part, out, site, fmt.Sprintf("%s.arm[%d]", path, i))
				}
				return
			}
			if alias := typ.Alias(); alias != nil {
				for i, arg := range alias.TypeArguments() {
					walk(arg, out, site, fmt.Sprintf("%s.alias-argument[%d]", path, i))
				}
			}
			if typ.Flags()&checker.TypeFlagsObject == 0 {
				return
			}
			if typ.ObjectFlags()&checker.ObjectFlagsReference != 0 {
				for i, arg := range a.checker.GetTypeArguments(typ) {
					walk(arg, out, site, fmt.Sprintf("%s.type-argument[%d]", path, i))
				}
			}
			if checker.IsTupleType(typ) {
				return
			}
			// The library's own declarations are audited separately. Their type
			// arguments were visited above, including host objects held in containers.
			if symbol := typ.Symbol(); symbol != nil {
				host := false
				for _, declaration := range symbol.Declarations {
					host = host || hostAnyLibraryFile(ast.GetSourceFileOfNode(declaration))
				}
				if !host {
					return
				}
			}
			canonical := typ
			if typ.ObjectFlags()&checker.ObjectFlagsReference != 0 && typ.Target() != nil {
				canonical = typ.Target()
			}
			key := state{canonical, out}
			if seen[key] {
				return
			}
			seen[key] = true
			a.states++
			for _, kind := range []checker.SignatureKind{checker.SignatureKindCall, checker.SignatureKindConstruct} {
				for i, signature := range a.checker.GetSignaturesOfType(typ, kind) {
					declaration := signature.Declaration()
					if declaration == nil {
						declaration = site
					}
					walk(
						a.checker.GetReturnTypeOfSignature(signature),
						out,
						declaration,
						fmt.Sprintf("%s.signature[%d:%d].return", path, kind, i),
					)
					for _, parameter := range signature.Parameters() {
						for _, pd := range parameter.Declarations {
							walk(
								a.checker.GetTypeOfSymbolAtLocation(parameter, pd),
								!out,
								pd,
								fmt.Sprintf("%s.signature[%d:%d].parameter(%s)", path, kind, i, parameter.Name),
							)
						}
					}
				}
			}
			properties := slices.Clone(a.checker.GetPropertiesOfType(typ))
			sort.Slice(properties, func(i, j int) bool { return properties[i].Name < properties[j].Name })
			for _, property := range properties {
				if !a.admitted(property) {
					continue
				} // Gated by rule 800 and host-object coverage.
				for _, declaration := range property.Declarations {
					walk(
						a.checker.GetTypeOfSymbolAtLocation(property, declaration),
						out,
						declaration,
						path+"."+property.Name,
					)
				}
			}
		}
		for _, root := range a.roots[name] {
			walk(a.checker.GetTypeOfSymbolAtLocation(root.symbol, root.declaration), true, root.declaration, name)
		}
	}
	for _, finding := range a.findings {
		test.Errorf("%s: value-carrying any at %s (%s)", finding.Root, finding.Path, finding.Where)
	}
	test.Logf(
		"audited %d registered/bound names; %d types; %d type/polarity states; %d reachable host anys",
		len(names),
		len(a.types),
		a.states,
		len(a.findings),
	)
}

// Bound stdout reads remain audit roots; registered hops use the production
// declaration check, including namespace-scoped structural members.
func (a *hostAnyAudit) admitted(symbol *ast.Symbol) bool {
	name := nodeLibraryName(symbol)
	return a.members[name] && (!implementedNodeMembers[name] || nodeLibraryMemberImplemented(symbol))
}
