package lower

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"
)

func programRegionCensus(t *testing.T, indexed bool) []byte {
	t.Helper()
	root := programRegionSource(t)
	input, err := os.Open("../oracle/testdata/program_region/census.csv")
	if err != nil {
		t.Fatal(err)
	}
	records, err := csv.NewReader(input).ReadAll()
	input.Close()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1686 {
		t.Fatalf("inventory records %d", len(records)-1)
	}
	roots := []tspath.RootedFilePath{}
	err = filepath.WalkDir(filepath.Join(root, "src/compiler"), func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, ".ts") {
			roots = append(roots, tspath.RootedFilePathFromAbsolute(filepath.ToSlash(path)))
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	fileSystem := bundled.WrapFS(osvfs.FS())
	options := &core.CompilerOptions{Strict: core.TSTrue, Module: core.ModuleKindESNext, ModuleResolution: core.ModuleResolutionKindBundler, Target: core.ScriptTargetESNext, NoEmit: core.TSTrue, Types: []string{}}
	config := tsoptions.NewParsedCommandLine(options, roots, nil, tspath.RootedDirectoryPathFromAbsolute(filepath.ToSlash(root)), fileSystem.CaseSensitivity())
	sourceProgram := compiler.NewProgram(compiler.ProgramOptions{Config: config, Host: compiler.NewCachedFSCompilerHost(fileSystem, bundled.LibPath(), nil, nil, nil), SingleThreaded: core.TSTrue})
	entry := sourceProgram.GetSourceFile(roots[0])
	if entry == nil {
		t.Fatal("entry missing")
	}
	typeChecker, release := sourceProgram.GetTypeCheckerForFile(context.Background(), entry)
	defer release()
	analysis := &lowering{checker: typeChecker, result: &ir.Program{}, this: -1, functionIndex: -1}
	finder := &cycleFinder{l: analysis, where: map[*checker.Type]*ast.Node{}}
	type property struct {
		node         *ast.Node
		holder, slot *checker.Type
	}
	properties := map[string][]property{}
	for _, file := range sourceProgram.GetSourceFiles() {
		relative, err := filepath.Rel(filepath.Join(root, "src/compiler"), file.FileName().AsString())
		if err != nil || strings.HasPrefix(relative, "..") {
			continue
		}
		relative = filepath.ToSlash(relative)
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			if node.Kind == ast.KindPropertySignature || node.Kind == ast.KindPropertyDeclaration || node.Kind == ast.KindMethodSignature || node.Kind == ast.KindMethodDeclaration || node.Kind == ast.KindGetAccessor {
				name := node.Name()
				if name != nil && (ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral || name.Kind == ast.KindNumericLiteral) {
					pos := scanner.GetTokenPosOfNode(name, file, false)
					line := strings.Count(file.Text()[:pos], "\n") + 1
					slot := typeChecker.GetTypeAtLocation(name)
					holder := typeChecker.GetTypeAtLocation(node.Parent)
					if node.Parent.Name() != nil && (node.Parent.Kind == ast.KindInterfaceDeclaration || node.Parent.Kind == ast.KindClassDeclaration) {
						holder = typeChecker.GetDeclaredTypeOfSymbol(typeChecker.GetSymbolAtLocation(node.Parent.Name()))
					}
					key := fmt.Sprintf("%s:%d:%s", relative, line, name.Text())
					properties[key] = append(properties[key], property{node, holder, slot})
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
	}
	site := regexp.MustCompile(`\[([^\]]+):(\d+)\]`)
	inputs := make([]property, len(records)-1)
	for i, row := range records[1:] {
		if strings.HasPrefix(row[0], "K") {
			continue
		}
		location := site.FindStringSubmatch(row[2])
		if location == nil {
			t.Fatalf("no site for %s", row[0])
		}
		name := strings.Trim(row[1][strings.LastIndex(row[1], ".")+1:], "\"'")
		candidates := properties[location[1]+":"+location[2]+":"+name]
		if len(candidates) == 0 {
			t.Fatalf("pinned declared slot missing: %s %s", row[0], row[1])
		}
		chosen := candidates[0]
		// A shared physical line can carry anonymous refinements. Their source holder
		// is used directly; we do not replace it with a named broad view from the witness.
		inputs[i] = chosen
		finder.use(chosen.holder, chosen.node)
		finder.use(chosen.slot, chosen.node)
	}
	// Resolve instantiated container schemas at actual source declaration/allocation
	// sites as well as through fields. Private/local types need not be exported.
	wanted := map[string]bool{}
	resolvedContainers := map[string]*checker.Type{}
	for _, row := range records[1:] {
		if strings.HasPrefix(row[0], "K") {
			wanted[programCensusTypeLabel(strings.ReplaceAll(row[1], "&#124;", "|"))] = true
		}
	}
	for _, file := range sourceProgram.GetSourceFiles() {
		relative, err := filepath.Rel(filepath.Join(root, "src/compiler"), file.FileName().AsString())
		if err != nil || strings.HasPrefix(relative, "..") {
			continue
		}
		var visit ast.Visitor
		visit = func(node *ast.Node) bool {
			var proven *checker.Type
			if node.Kind == ast.KindVariableDeclaration || node.Kind == ast.KindParameter || node.Kind == ast.KindTypeAliasDeclaration {
				if node.Name() != nil && ast.IsIdentifier(node.Name()) {
					proven = typeChecker.GetTypeAtLocation(node.Name())
				}
			} else if ast.IsTypeNode(node) || node.Kind == ast.KindNewExpression || node.Kind == ast.KindArrayLiteralExpression {
				proven = typeChecker.GetTypeAtLocation(node)
			}
			if proven != nil {
				label := typeChecker.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
				if wanted[programCensusTypeLabel(label)] {
					resolvedContainers[programCensusTypeLabel(label)] = proven
					finder.use(proven, node)
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
	}
	started := time.Now()
	selected := finder.programRegionSelectionMode(indexed)
	t.Logf("selection indexed=%t: %.6f seconds", indexed, time.Since(started).Seconds())
	labels := map[string]*checker.Type{}
	format := func(proven *checker.Type) string {
		return typeChecker.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
	}
	candidates := []*checker.Type{}
	for proven := range finder.where {
		candidates = append(candidates, proven)
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].Id() < candidates[j].Id() })
	for _, proven := range candidates {
		labels[programCensusTypeLabel(format(proven))] = proven
		expanded := typeChecker.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation|checker.TypeFormatFlagsInTypeAlias, nil)
		labels[programCensusTypeLabel(expanded)] = proven
	}
	for label, proven := range resolvedContainers {
		labels[label] = proven
	}
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	writer.Write([]string{"record_id", "declared_type", "in_region_or_counted", "why", "slot_type"})
	seen := map[string]bool{}
	in, unresolved, waiting := 0, 0, 0
	for i, row := range records[1:] {
		if seen[row[0]] {
			t.Fatalf("duplicate %s", row[0])
		}
		seen[row[0]] = true
		candidate := inputs[i].holder
		slotType := ""
		declared := row[1]
		why := "outside the owning-field SCC of the provisional declaration graph"
		if candidate != nil {
			declared = format(candidate)
			slotType = format(inputs[i].slot)
		} else {
			declared = strings.ReplaceAll(row[1], "&#124;", "|")
			candidate = labels[programCensusTypeLabel(declared)]
			// Census K131 alpha-renames PrivateEnvironment's TEntry while
			// substituting it through LexicalEnvironment<TPrivateEntry>. Both
			// identities have the same concrete Node key and unknown value;
			// use that checker schema without inventing a concrete value type.
			if candidate == nil && regexp.MustCompile(`^Map<Node, T[A-Za-z0-9_]+>$`).MatchString(declared) {
				for _, actual := range resolvedContainers {
					arguments := typeChecker.GetTypeArguments(actual)
					if analysis.isLibraryType(actual, "Map") && len(arguments) == 2 && format(arguments[0]) == "Node" && arguments[1].Flags()&checker.TypeFlagsTypeParameter != 0 {
						candidate = actual
						break
					}
				}
			}
			if candidate == nil {
				unresolved++
				why = "no concrete identity in declaration-only graph; keep counted pending lowered generic/capture/container identity"
			}
		}
		membership := "counted"
		if candidate != nil {
			switch {
			case finder.programContainer(candidate):
				why = "allocation-site container rule" + "; no selected element/key/value; stays counted"
				if finder.template(candidate) {
					why += "; waiting for concrete allocation-site type"
				}
				if finder.programContainerMember(candidate, selected) {
					membership = "in region"
					in++
					why = "allocation-site container rule" + "; selected element/key/value; region member"
				}
			case finder.template(candidate):
				waiting++
				why = "uninstantiated generic/any declaration; keep counted pending concrete allocation-site type"
			case finder.programScalarStorage(candidate):
				why = "scalar scratch storage; no owned graph elements"
			case selected[cycleNode{proven: candidate}] && !finder.weak(candidate) && !analysis.isLibraryType(candidate, "Promise"):
				membership = "in region"
				in++
				why = "owning-field SCC selected by programRegionSelection; declaration audit has no lowered capture cells"
			}
		}
		if expected, reviewed := programCompilerContainerReading[row[0]]; reviewed {
			if candidate == nil {
				t.Fatalf("reviewed container %s still unresolved", row[0])
			}
			if membership != expected {
				why += "; differs from compiler reading (" + expected + "); provisional structural SCC selection"
				t.Logf("compiler reading differs: %s got %s want %s (%s)", row[0], membership, expected, declared)
				if row[0] == "K60" || row[0] == "K144" {
					t.Fatal("ruled scalar container became a member")
				}
			}
		}
		if err := writer.Write([]string{row[0], declared, membership, why, slotType}); err != nil {
			t.Fatal(err)
		}
	}
	if waiting != 77 {
		t.Fatalf("uninstantiated declarations waiting: %d, want 77", waiting)
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1685 {
		t.Fatal("record omitted")
	}
	// Independent anchors make missing/changed input records visible.
	required := []string{"D11", "D633", "D634", "D708", "D938", "K15", "K69"}
	sort.Strings(required)
	for _, id := range required {
		if !seen[id] {
			t.Fatalf("required census record missing: %s", id)
		}
	}
	return output.Bytes()
}

// Compiler's reading is an independent comparison, never a membership override.
var programCompilerContainerReading = map[string]string{
	"K24": "in region", "K67": "in region", "K85": "in region", "K86": "in region", "K118": "in region",
	"K109": "in region", "K116": "in region", "K117": "in region", "K130": "in region", "K131": "in region",
	"K60": "counted", "K144": "counted", "K90": "counted", "K91": "counted",
}

// Checker output orders union constituents and spells readonly arrays differently
// from the stock census. Canonicalize display only; retain checker identities for
// selection, and do not synthesize missing types or substitute record-specific flags.
func programCensusTypeLabel(label string) string {
	label = strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(label), "readonly "))
	if strings.HasSuffix(label, "[]") {
		return "(" + programCensusTypeLabel(strings.TrimSuffix(label, "[]")) + ")[]"
	}
	if strings.HasPrefix(label, "(") && strings.HasSuffix(label, ")") {
		label = label[1 : len(label)-1]
	}
	depth, start := 0, 0
	parts := []string{}
	for i, r := range label {
		switch r {
		case '(', '[', '{', '<':
			depth++
		case ')', ']', '}', '>':
			if depth > 0 {
				depth--
			}
		case '|':
			if depth == 0 {
				parts = append(parts, programCensusTypeLabel(label[start:i]))
				start = i + 1
			}
		}
	}
	if len(parts) != 0 {
		parts = append(parts, programCensusTypeLabel(label[start:]))
		sort.Strings(parts)
		return strings.Join(parts, "|")
	}
	// Normalize unions within generic arguments too (e.g. SourceFile | Bundle).
	for start := 0; start < len(label); start++ {
		if label[start] != '<' {
			continue
		}
		depth, end := 1, start+1
		for ; end < len(label) && depth != 0; end++ {
			if label[end] == '<' {
				depth++
			}
			if label[end] == '>' {
				depth--
			}
		}
		if depth != 0 {
			break
		}
		inside := label[start+1 : end-1]
		// Normalize each argument; commas in nested generics must stay nested.
		args, begin, nesting := []string{}, 0, 0
		for i, r := range inside {
			switch r {
			case '<', '(', '[', '{':
				nesting++
			case '>', ')', ']', '}':
				if nesting > 0 {
					nesting--
				}
			case ',':
				if nesting == 0 {
					args = append(args, programCensusTypeLabel(inside[begin:i]))
					begin = i + 1
				}
			}
		}
		args = append(args, programCensusTypeLabel(inside[begin:]))
		replacement := "<" + strings.Join(args, ",") + ">"
		label = label[:start] + replacement + label[end:]
		start += len(replacement) - 1
	}
	return strings.Join(strings.Fields(label), "")
}
func programRegionSource(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "program-tsc-source", Files: []string{"internal/oracle/testdata/program_region/source.sh", "internal/oracle/testdata/program_region/parse_manifest.json"}}
	build := func(directory string) error {
		out, err := exec.Command("bash", "../oracle/testdata/program_region/source.sh", directory).CombinedOutput()
		if err != nil {
			return fmt.Errorf("pinned TypeScript source: %w: %s", err, out)
		}
		return nil
	}
	return buildcache.BuiltElsewhere(t, "the pinned tsc corpus", inputs, "internal/oracle.TestProduct_ProgramTscSource", build)
}

func programRegionCensusProduct(t *testing.T, indexed bool) string {
	t.Helper()
	return buildcache.Product(t, buildcache.Inputs{Name: fmt.Sprint("program-region-census-", indexed), Files: []string{"internal/lower", "internal/ir", "internal/oracle/testdata/program_region/census.csv", "internal/oracle/testdata/program_region/source.sh", "internal/oracle/testdata/program_region/parse_manifest.json"}}, func(directory string) error {
		return os.WriteFile(path.Join(directory, "membership.csv"), programRegionCensus(t, indexed), 0600)
	})
}
func TestProduct_ProgramRegionCensusIndexed(t *testing.T) {
	t.Parallel()
	programRegionCensusProduct(t, true)
}
func TestProduct_ProgramRegionCensusPairwise(t *testing.T) {
	t.Parallel()
	programRegionCensusProduct(t, false)
}
func TestProgramRegionCensusMembership(t *testing.T) {
	t.Parallel()
	a, err := os.ReadFile(path.Join(programRegionCensusProduct(t, true), "membership.csv"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path.Join(programRegionCensusProduct(t, false), "membership.csv"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("indexed and pairwise membership differ")
	}
}
