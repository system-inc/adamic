// Package metamorphic rewrites the oracle's hand-written fixtures into programs that mean the same
// thing by another road: the module's statements wrapped in a function, a value read through an
// alias, a block of statements moved into a method on a class of its own, a value passed through an
// identity function or a closure, an expression split into temporaries and a single-use temporary
// put back where it was used. Every edit is made at the boundaries of nodes TypeScript's own parser
// found, with the checker naming what each identifier refers to, never by a pattern over the text.
//
// A transform proposes its edits as sites, and a site is kept only when the program with it still
// passes the checker and stage 0 lowers it; a site either refuses is counted and dropped. The
// fixtures already reach nearly all of the emitter, so the same constructs, moved like this, travel
// through other paths of the compiler: locals instead of globals, captures and parameters instead of
// direct reads, a borrowed value that is suddenly a returned one. What a variant prints on Node must
// be exactly what the original printed (the caller checks that), and then every way the oracle runs
// a program has to agree with Node on it.
package metamorphic

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"

	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

// Transform names one way of rewriting a program.
type Transform string

const (
	// Wrap moves the module's statements into a named function called once: its variables become
	// locals, its functions closures. Imports, types and classes (which stage 0 lowers only at the top
	// level) stay where they are, with whatever they reach.
	Wrap Transform = "wrap"
	// Alias reads a constant used twice through an alias declared beside it, from its second use on.
	Alias Transform = "alias"
	// Method moves a run of statements into a method on a class of its own, called once, with the
	// enclosing function's constants as its parameters.
	Method Transform = "method"
	// Identity passes values through an identity function or an immediately called closure.
	Identity Transform = "identity"
	// Temporaries splits an expression's operands and arguments into temporaries declared before its
	// statement, and puts a temporary used once, in the next statement, back where it was used.
	Temporaries Transform = "temporaries"
)

// Transforms are every transform, numbered from one in this order in the report.
var Transforms = []Transform{Wrap, Alias, Method, Identity, Temporaries}

// Combined is every transform, in the order they compose: the module wrapped first, so what follows
// works on locals; blocks moved out before anything adds names they would have to carry; values
// passed through identity last, so it wraps what the others made too.
var Combined = []Transform{Wrap, Method, Alias, Temporaries, Identity}

// ParseTransform reads a transform's name.
func ParseTransform(name string) (Transform, error) {
	for _, transform := range Transforms {
		if string(transform) == name {
			return transform, nil
		}
	}
	names := make([]string, len(Transforms))
	for index, transform := range Transforms {
		names[index] = string(transform)
	}
	return "", fmt.Errorf("no transform %q; the transforms are %s", name, strings.Join(names, ", "))
}

// Limits on one transform's work on one program: how many sites are tried with the checker, and how
// many are kept. Sites beyond what's tried are chosen by a stride, so they spread over the program.
const (
	tries = 24
	kept  = 12
)

// Variant is what one or more transforms made of a program.
type Variant struct {
	Source string
	// Sites is how many edits were kept, and Refused how many the checker or stage 0 refused.
	Sites   int
	Refused int
	// Refusals is the first refusal of each kind, for the report.
	Refusals []string
	// Skipped is why no variant came of it ("" when one did): nothing to rewrite, or a program the
	// transform leaves alone by design (one with exports, for Wrap).
	Skipped string
}

// Apply runs transforms in order over the program at path (absolute: its imports resolve beside it),
// each over what the one before made.
func Apply(path string, source string, transforms []Transform) Variant {
	result := Variant{Source: source}
	var skipped []string
	for _, transform := range transforms {
		step := apply(path, result.Source, transform)
		result.Refused += step.Refused
		result.Refusals = append(result.Refusals, step.Refusals...)
		if step.Skipped != "" {
			skipped = append(skipped, string(transform)+": "+step.Skipped)
			continue
		}
		result.Source = step.Source
		result.Sites += step.Sites
	}
	if result.Sites == 0 {
		result.Skipped = strings.Join(skipped, "; ")
		if result.Skipped == "" {
			result.Skipped = "no site was kept"
		}
	}
	return result
}

func apply(path string, source string, transform Transform) Variant {
	program, err := analyze(path, source)
	if err != nil {
		return Variant{Skipped: "the program doesn't check: " + firstLine(err.Error())}
	}
	defer program.release()
	if transform == Wrap {
		result := Variant{}
		for _, functions := range []functionsAs{asFunctionExpressions, asArrows, atTopLevel} {
			wrapped, skipped := wrap(program, functions)
			if skipped != "" {
				result.Skipped = skipped
				return result
			}
			if refusal := accepts(path, wrapped); refusal != "" {
				result.Refused++
				result.Refusals = append(result.Refusals, refusal)
				result.Skipped = "refused: " + refusal
				continue
			}
			result.Source, result.Sites, result.Skipped = wrapped, 1, ""
			return result
		}
		return result
	}
	var sites []site
	switch transform {
	case Alias:
		sites = aliasSites(program)
	case Method:
		sites = methodSites(program)
	case Identity:
		sites = identitySites(program)
	case Temporaries:
		sites = temporarySites(program)
	}
	if len(sites) == 0 {
		return Variant{Skipped: "no site"}
	}
	return greedy(path, program, sites)
}

// program is a source as the checker sees it.
type program struct {
	text    string
	file    *ast.SourceFile
	checker *checker.Checker
	release func()
	// names is every identifier the text spells, so a new name is never one already there.
	names map[string]bool
}

func analyze(path string, text string) (*program, error) {
	loaded, err := load.LoadOverlay([]string{path}, map[string]string{path: text})
	if err != nil {
		return nil, err
	}
	if len(loaded.Files()) == 0 {
		return nil, errors.New("no file")
	}
	file := loaded.Files()[0]
	if file.Text() != text {
		return nil, errors.New("the checker read another text")
	}
	typeChecker, release := loaded.Checker(context.Background(), file)
	result := &program{text: text, file: file, checker: typeChecker, release: release, names: map[string]bool{}}
	visit(file.AsNode(), func(node *ast.Node) bool {
		if node.Kind == ast.KindIdentifier {
			result.names[node.Text()] = true
		}
		return true
	})
	return result, nil
}

// fresh is a name the program doesn't spell, made from base, and from then on taken.
func (p *program) fresh(base string) string {
	name := base
	for index := 2; p.names[name]; index++ {
		name = fmt.Sprintf("%s%d", base, index)
	}
	p.names[name] = true
	return name
}

// start is where a node's own text begins, after its leading trivia.
func (p *program) start(node *ast.Node) int {
	return scanner.SkipTrivia(p.text, node.Pos())
}

func (p *program) source(node *ast.Node) string {
	return p.text[p.start(node):node.End()]
}

// indentation is the white space a statement's line begins with.
func (p *program) indentation(node *ast.Node) string {
	start := p.start(node)
	line := strings.LastIndexByte(p.text[:start], '\n') + 1
	end := line
	for end < start && (p.text[end] == ' ' || p.text[end] == '\t') {
		end++
	}
	return p.text[line:end]
}

// headerPosition is where declarations a transform adds go: after the last import, or at the top.
func (p *program) headerPosition() int {
	position := 0
	for _, statement := range p.file.Statements.Nodes {
		if statement.Kind == ast.KindImportDeclaration || statement.Kind == ast.KindImportEqualsDeclaration {
			position = statement.End()
		}
	}
	return position
}

// visit walks a node and its descendants, parents first; enter returns false to skip a node's
// children.
func visit(node *ast.Node, enter func(*ast.Node) bool) {
	if !enter(node) {
		return
	}
	node.ForEachChild(func(child *ast.Node) bool {
		visit(child, enter)
		return false
	})
}

// edit replaces the text from start to end; an insertion has start == end.
type edit struct {
	start, end int
	text       string
}

// site is one place a transform rewrites: the edits it makes, and the declarations it needs at the
// top (header), each named so two sites needing the same one share it.
type site struct {
	edits  []edit
	header []declaration
}

type declaration struct {
	name string
	text string
}

// conflicts says whether two sites' edits touch the same text.
func conflicts(left site, right site) bool {
	for _, a := range left.edits {
		for _, b := range right.edits {
			if a.start < b.end && b.start < a.end {
				return true
			}
			// An insertion inside the other's replacement.
			if (a.start == a.end && a.start > b.start && a.start < b.end) || (b.start == b.end && b.start > a.start && b.start < a.end) {
				return true
			}
		}
	}
	return false
}

// assemble is the text with sites' edits made and their declarations added at position.
func assemble(text string, position int, sites []site) string {
	type ordered struct {
		edit
		sequence int
	}
	var edits []ordered
	var header strings.Builder
	declared := map[string]bool{}
	for _, chosen := range sites {
		for _, needed := range chosen.header {
			if !declared[needed.name] {
				declared[needed.name] = true
				header.WriteString("\n")
				header.WriteString(needed.text)
				header.WriteString("\n")
			}
		}
	}
	if header.Len() > 0 {
		edits = append(edits, ordered{edit{position, position, header.String()}, -1})
	}
	for _, chosen := range sites {
		for _, change := range chosen.edits {
			edits = append(edits, ordered{change, len(edits)})
		}
	}
	// From the end back, so every edit's offsets are still the original's. At one offset, a
	// replacement goes first and then the insertions, the last first, so they read in their order.
	sort.SliceStable(edits, func(i, j int) bool {
		if edits[i].start != edits[j].start {
			return edits[i].start > edits[j].start
		}
		iReplaces, jReplaces := edits[i].end > edits[i].start, edits[j].end > edits[j].start
		if iReplaces != jReplaces {
			return iReplaces
		}
		return edits[i].sequence > edits[j].sequence
	})
	result := text
	for _, change := range edits {
		result = result[:change.start] + change.text + result[change.end:]
	}
	return result
}

// greedy keeps each site, in order, that the checker and stage 0 accept beside the ones already kept.
func greedy(path string, p *program, sites []site) Variant {
	if len(sites) > tries {
		strided := make([]site, 0, tries)
		for index := range tries {
			strided = append(strided, sites[index*len(sites)/tries])
		}
		sites = strided
	}
	position := p.headerPosition()
	var chosen []site
	result := Variant{}
	seen := map[string]bool{}
	for _, candidate := range sites {
		if len(chosen) >= kept {
			break
		}
		clash := false
		for _, existing := range chosen {
			if conflicts(existing, candidate) {
				clash = true
				break
			}
		}
		if clash {
			continue
		}
		text := assemble(p.text, position, append(append([]site{}, chosen...), candidate))
		if refusal := accepts(path, text); refusal != "" {
			result.Refused++
			kind := refusalKind(refusal)
			if !seen[kind] {
				seen[kind] = true
				result.Refusals = append(result.Refusals, refusal)
			}
			continue
		}
		chosen = append(chosen, candidate)
	}
	if len(chosen) == 0 {
		return Variant{Refused: result.Refused, Refusals: result.Refusals, Skipped: fmt.Sprintf("every site refused (%d)", result.Refused)}
	}
	result.Source = assemble(p.text, position, chosen)
	result.Sites = len(chosen)
	return result
}

// accepts checks and lowers a text as the file at path, and says why not, or "". A panic while
// lowering is the compiler's own crash, which is worth running, so it accepts the text.
func accepts(path string, text string) (refusal string) {
	loaded, err := load.LoadOverlay([]string{path}, map[string]string{path: text})
	if err != nil {
		var checkError *load.CheckError
		if errors.As(err, &checkError) && len(checkError.Diagnostics) > 0 {
			return "checker: " + firstLine(withoutLocation(checkError.Diagnostics[0]))
		}
		return "checker: " + firstLine(err.Error())
	}
	defer func() {
		if recover() != nil {
			refusal = ""
		}
	}()
	if _, err := lower.Lower(context.Background(), loaded); err != nil {
		var notYet *lower.NotYet
		var refused *lower.Refused
		switch {
		case errors.As(err, &notYet):
			return "not yet: " + notYet.What
		case errors.As(err, &refused):
			return "refused: " + refused.What
		}
		return "lowering: " + firstLine(err.Error())
	}
	return ""
}

// refusalKind is a refusal without what's particular to one site: the checker's code, or stage 0's
// words.
func refusalKind(refusal string) string {
	if strings.HasPrefix(refusal, "checker: ") {
		if index := strings.Index(refusal, "TS"); index >= 0 {
			end := index + 2
			for end < len(refusal) && refusal[end] >= '0' && refusal[end] <= '9' {
				end++
			}
			return refusal[index:end]
		}
	}
	return refusal
}

// withoutLocation is a diagnostic without the file and position it begins with.
func withoutLocation(diagnostic string) string {
	if index := strings.Index(diagnostic, ": error "); index >= 0 {
		return diagnostic[index+2:]
	}
	return diagnostic
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

// Imports says whether a program imports anything, and whether anything it imports is a file beside
// it (a relative path), which has to travel with it.
func Imports(path string, source string) (imports bool, relative bool) {
	file := parser.ParseSourceFile(ast.SourceFileParseOptions{FileName: parseName(path), Path: tspath.Path(parseName(path))}, source, core.ScriptKindTS)
	for _, statement := range file.Statements.Nodes {
		if statement.Kind != ast.KindImportDeclaration {
			continue
		}
		imports = true
		if specifier := statement.AsImportDeclaration().ModuleSpecifier; specifier != nil && strings.HasPrefix(specifier.Text(), ".") {
			relative = true
		}
	}
	return imports, relative
}

// parseName is the name the parser is given: absolute, as it requires, and TypeScript, which a .a
// file is to the parser.
func parseName(path string) string {
	name := "/" + filepath.Base(path)
	if strings.HasSuffix(name, ".a") {
		return name + ".ts"
	}
	return name
}
