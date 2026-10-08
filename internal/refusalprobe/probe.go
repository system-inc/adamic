package refusalprobe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Program carries both inputs so every finding is reproducible without the writer.
type Program struct {
	Index                                          int
	Seed                                           uint64
	Entry, Surrounding, Expected, Source, Neighbor string
}

// Generate cycles through all executable entries before repeating one. The seed changes the
// placement, constants and padding; sequence coverage does not depend on random luck.
func Generate(seed uint64, index int) Program {
	entries := executable()
	entry := entries[index%len(entries)]
	return generate(seed, index, entry)
}

// GenerateEntries restricts regression measurements to the entries changed by a patch.
func GenerateEntries(seed uint64, index int, names []string) (Program, error) {
	if len(names) == 0 {
		return Generate(seed, index), nil
	}
	name := names[index%len(names)]
	for _, entry := range executable() {
		if entry.Name == name {
			return generate(seed, index, entry), nil
		}
	}
	return Program{}, fmt.Errorf("unknown or non-executable entry %q", name)
}

func generate(seed uint64, index int, entry Entry) Program {
	random := rand.New(rand.NewPCG(seed, uint64(index)))
	context := int(random.Uint64N(7))
	wrappers := [][2]string{
		{"", ""},
		{"function run(): void {\n", "\n}\nrun();\n"},
		{"class Context { run(): void {\n", "\n} }\nnew Context().run();\n"},
		{"const run = (): void => {\n", "\n};\nrun();\n"},
		{"function run<T>(item: T): T {\n", "\nreturn item;\n}\nrun(1);\n"},
		{"if (true) {\n{\n", "\n}\n}\n"},
		{"class Context { readonly run: () => void = (): void => {\n", "\n}; }\nnew Context().run();\n"},
	}
	names := []string{"top-level", "function", "method", "closure", "generic", "nested-block", "class-field"}
	prefix := fmt.Sprintf("const padding = %d;\nconsole.log(`${padding}`);\n", random.Uint64N(10000))
	// Runtime imports are accepted at all placements and exercise refusal traversal after imports.
	if random.Uint64N(2) == 0 {
		prefix = "import { panic } from 'adamic';\n" + prefix
	}
	wrap := func(source string) string {
		return prefix + wrappers[context][0] + source + "\n" + wrappers[context][1]
	}
	if entry.Placement == "module" {
		context = 0
		names[0] = "module"
	}
	// Runtime namespaces must initialize before the generated padding call.
	if entry.Placement == "module-first" {
		context = 0
		names[0] = "module"
		wrap = func(source string) string { return source + "\n" + prefix }
	}
	if entry.Placement == "prefix" {
		wrap = func(source string) string {
			return source + "\n" + prefix + wrappers[context][0] + "console.log('neighbor');\n" + wrappers[context][1]
		}
	}
	return Program{index, seed, entry.Name, names[context], entry.Diagnostic, wrap(entry.Bad), wrap(entry.Good)}
}

func executable() []Entry {
	var entries []Entry
	for _, entry := range Catalog() {
		if entry.Boundary == "" && !entry.Accepted {
			entries = append(entries, entry)
		}
	}
	return entries
}

// ValidateCatalog fails when a syntax/operator map refusal has no catalog diagnostic. Parse Go,
// rather than matching source formatting. Boundary entries count, but are never called successes.
func ValidateCatalog(root string) error {
	path := filepath.Join(root, "internal/lower/refusals.go")
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		return err
	}
	covered := func(text string) bool {
		for _, entry := range Catalog() {
			if entry.Diagnostic == text {
				return true
			}
		}
		return false
	}
	var missing []string
	dynamic := map[string][]string{
		`name + " suppression directive"`:        {"ts-ignore", "ts-expect-error"},
		`"@" + pragma.Name + " checking pragma"`: {"ts-nocheck", "ts-check"},
		`refused.what`:                           {"non-null"},
		`"a method read as a value (" + symbol.Name + " would lose its object, and this with it)"`: {"unbound-method"},
	}
	entryExists := func(name string) bool {
		for _, entry := range Catalog() {
			if entry.Name == name {
				return true
			}
		}
		return false
	}
	// Error-returning helpers called by the pass must have an explicit catalog owner too.
	helpers := map[string][]string{
		"nodeLibraryRefusal":     {"node-library"},
		"typedArrayUnsupported":  {"typed-array-unsupported"},
		"predicateArguments":     {"predicate-argument"},
		"namespaceRefusal":       {"namespace"},
		"checkOverrides":         {"method-override"},
		"refuseStringWidening":   {"mutable-variance"},
		"classViewRefusal":       {"nominal-class"},
		"prototypeRead":          {"prototype-read"},
		"provePredicate":         {"unproven-predicate"},
		"enumRefusal":            {"enum-tag", "enum-object-view", "enum-string-object-view", "enum-object-write", "enum-prototype-name", "enum-nonfinite-name", "enum-nested", "enum-merged", "enum-ambient", "parameter-properties"},
		"refuseOptionalWidening": {"optional-widening"},
	}
	ast.Inspect(file, func(node ast.Node) bool {
		conditional, ok := node.(*ast.IfStmt)
		if !ok || conditional.Init == nil {
			return true
		}
		assignment, ok := conditional.Init.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for _, expression := range assignment.Rhs {
			call, ok := expression.(*ast.CallExpr)
			if !ok {
				continue
			}
			selector, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				continue
			}
			receiver, ok := selector.X.(*ast.Ident)
			if !ok || receiver.Name != "l" {
				continue
			}
			names, known := helpers[selector.Sel.Name]
			if !known {
				missing = append(missing, "helper "+selector.Sel.Name)
			}
			for _, name := range names {
				if !entryExists(name) {
					missing = append(missing, "helper "+selector.Sel.Name+" entry "+name)
				}
			}
		}
		return true
	})
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.CompositeLit)
		if !ok {
			return true
		}
		kind, ok := literal.Type.(*ast.Ident)
		if !ok || kind.Name != "Refused" {
			return true
		}
		for _, element := range literal.Elts {
			pair, ok := element.(*ast.KeyValueExpr)
			if !ok {
				continue
			}
			key, ok := pair.Key.(*ast.Ident)
			if !ok || key.Name != "What" {
				continue
			}
			if text, ok := pair.Value.(*ast.BasicLit); ok {
				diagnostic, err := strconv.Unquote(text.Value)
				if err != nil || !covered(diagnostic) {
					missing = append(missing, diagnostic)
				}
			} else {
				var expression bytes.Buffer
				if err := format.Node(&expression, token.NewFileSet(), pair.Value); err != nil {
					missing = append(missing, err.Error())
					continue
				}
				entries, known := dynamic[expression.String()]
				if !known {
					missing = append(missing, expression.String())
				}
				for _, name := range entries {
					if !entryExists(name) {
						missing = append(missing, name)
					}
				}
			}
		}
		return true
	})
	ast.Inspect(file, func(node ast.Node) bool {
		declaration, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for _, name := range declaration.Names {
			if name.Name != "refusals" && name.Name != "refusedOperators" {
				continue
			}
			for _, value := range declaration.Values {
				literal, ok := value.(*ast.CompositeLit)
				if !ok {
					continue
				}
				for _, element := range literal.Elts {
					pair, ok := element.(*ast.KeyValueExpr)
					if !ok {
						continue
					}
					refusal, ok := pair.Value.(*ast.CompositeLit)
					if !ok || len(refusal.Elts) == 0 {
						continue
					}
					text, ok := refusal.Elts[0].(*ast.BasicLit)
					if !ok {
						missing = append(missing, "nonliteral refusal")
						continue
					}
					diagnostic, err := strconv.Unquote(text.Value)
					if err != nil || !covered(diagnostic) {
						missing = append(missing, diagnostic)
					}
				}
			}
		}
		return true
	})
	if len(missing) > 0 {
		return fmt.Errorf("refusal catalog is incomplete: %s", strings.Join(missing, ", "))
	}
	return nil
}

// Compile follows adamic c's load/lower path, then exercises C generation on accepted programs.
// Inputs are fresh for every call. There is no probe result cache.
func Compile(ctx context.Context, source string) error {
	directory, err := os.MkdirTemp("", "adamic-refusal-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(directory)
	path := filepath.Join(directory, "main.a")
	if err := os.WriteFile(path, []byte(source), 0644); err != nil {
		return err
	}
	program, err := load.Load([]string{path})
	if err != nil {
		return normalizeError(directory, err)
	}
	lowered, err := lower.Lower(ctx, program)
	if err != nil {
		return normalizeError(directory, err)
	}
	if native.UsesTSGo(lowered) {
		return errors.New("tsgo requires a native build")
	}
	_ = native.C(lowered)
	return nil
}

// Temporary paths must not affect reproducible output or uncached comparisons.
func normalizeError(directory string, err error) error {
	var refused *lower.Refused
	if errors.As(err, &refused) {
		refused.Where = strings.TrimPrefix(refused.Where, directory+"/")
		return refused
	}
	var notYet *lower.NotYet
	if errors.As(err, &notYet) {
		notYet.Where = strings.TrimPrefix(notYet.Where, directory+"/")
		return notYet
	}
	var check *load.CheckError
	if errors.As(err, &check) {
		for index := range check.Diagnostics {
			check.Diagnostics[index] = strings.ReplaceAll(check.Diagnostics[index], directory+"/", "")
		}
		return check
	}
	return err
}

type Finding struct {
	Kind       string
	Program    Program
	Diagnostic string
}

// Diagnose requires the specific refusal, including its What/Fix identity, not any error.
func Diagnose(program Program, err error) *Finding {
	if err == nil {
		return &Finding{"accepted", program, "compiled without refusal"}
	}
	var refused *lower.Refused
	if errors.As(err, &refused) {
		// A short name such as any or in must still be a whole diagnostic construct.
		text := refused.What + "; " + refused.Fix
		match := strings.Contains(text, program.Expected)
		if len(program.Expected) <= 4 {
			match = refused.What == program.Expected
		}
		if match {
			return nil
		}
		return &Finding{"wrong-refusal", program, err.Error()}
	}
	var notYet *lower.NotYet
	if errors.As(err, &notYet) {
		return &Finding{"not-yet", program, err.Error()}
	}
	return &Finding{"load-error", program, err.Error()}
}

// Check rejects broken surroundings before judging the refusal. A neighbor failure is an
// invalid probe, never evidence of a soundness hole in the compiler.
func Check(ctx context.Context, program Program) *Finding {
	if err := Compile(ctx, program.Neighbor); err != nil {
		return &Finding{"invalid-neighbor", program, err.Error()}
	}
	return Diagnose(program, Compile(ctx, program.Source))
}

// WriteFinding saves only finding inputs, each with its accepted control.
func WriteFinding(directory string, finding *Finding) error {
	if err := os.MkdirAll(directory, 0755); err != nil {
		return err
	}
	name := fmt.Sprintf("%06d-%s", finding.Program.Index, finding.Program.Entry)
	for suffix, text := range map[string]string{".a": finding.Program.Source, "-neighbor.a": finding.Program.Neighbor, ".txt": finding.Kind + "\n" + finding.Diagnostic + "\n"} {
		if err := os.WriteFile(filepath.Join(directory, name+suffix), []byte(text), 0644); err != nil {
			return err
		}
	}
	return nil
}
