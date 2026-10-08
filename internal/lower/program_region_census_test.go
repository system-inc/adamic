package lower

import (
	"bytes"
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/bundled"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/compiler"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/microsoft/TypeScript/tsc/shim/tsoptions"
	"github.com/microsoft/TypeScript/tsc/shim/tspath"
	"github.com/microsoft/TypeScript/tsc/shim/vfs/osvfs"
	"github.com/system-inc/adamic/internal/ir"
)

var updateProgramMembership = flag.Bool("update-program-membership", false, "regenerate docs/cycles-decision/membership.csv from the pinned declaration graph")

// This is an audit of declared allocation types on the original source, using
// the actual provisional graph builder and selector. It does not pretend the
// original compiler lowers, nor invent lowered capture cells or generic instances.
func TestProgramRegionCensusMembership(t *testing.T) {
	root := os.Getenv("ADAMIC_PROGRAM_CENSUS_ROOT")
	if root == "" {
		t.Skip("set ADAMIC_PROGRAM_CENSUS_ROOT to pristine TypeScript 6.0.3 with generated diagnostics")
	}
	pin, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil || strings.TrimSpace(string(pin)) != "050880ce59e30b356b686bd3144efe24f875ebc8" {
		t.Fatalf("source pin %s %v", pin, err)
	}
	// Verify that the input inventory still belongs to the explicitly fetched census.
	censusPin, err := exec.Command("git", "rev-parse", "fcb7451a").Output()
	if err != nil || strings.TrimSpace(string(censusPin)) != "fcb7451a8239723fd903ae5998f54d2a86715e1f" {
		t.Fatalf("census pin %s %v", censusPin, err)
	}
	input, err := os.Open("../../docs/cycles-decision/census-groups.csv")
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
	ledger, err := exec.Command("git", "show", "fcb7451a:docs/stage3-tsc-cycles.md").Output()
	if err != nil {
		t.Fatal(err)
	}
	inventory := map[string]string{}
	for _, match := range regexp.MustCompile("(?m)^\\| ([DK][0-9]+) `([^`]+)` \\|").FindAllStringSubmatch(string(ledger), -1) {
		inventory[match[1]] = match[2]
	}
	if len(inventory) != 1685 {
		t.Fatalf("pinned inventory size %d", len(inventory))
	}
	for _, row := range records[1:] {
		if inventory[row[0]] != row[1] {
			t.Fatalf("record %s differs from pinned census", row[0])
		}
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
		location := site.FindStringSubmatch(row[5])
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
	for _, row := range records[1:] {
		if strings.HasPrefix(row[0], "K") {
			wanted[strings.ReplaceAll(row[1], "&#124;", "|")] = true
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
			} else if node.Kind == ast.KindNewExpression || node.Kind == ast.KindArrayLiteralExpression {
				proven = typeChecker.GetTypeAtLocation(node)
			}
			if proven != nil {
				label := typeChecker.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation, nil)
				if wanted[label] {
					finder.use(proven, node)
				}
			}
			return node.ForEachChild(visit)
		}
		file.AsNode().ForEachChild(visit)
	}
	selected := finder.programRegionSelection()
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
		labels[format(proven)] = proven
		expanded := typeChecker.TypeToStringEx(proven, nil, checker.TypeFormatFlagsNoTruncation|checker.TypeFormatFlagsInTypeAlias, nil)
		labels[expanded] = proven
	}
	var output bytes.Buffer
	writer := csv.NewWriter(&output)
	writer.Write([]string{"record_id", "declared_type", "in_region_or_counted", "why", "slot_type"})
	seen := map[string]bool{}
	in, unresolved := 0, 0
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
			candidate = labels[declared]
			if candidate == nil {
				unresolved++
				why = "no concrete identity in declaration-only graph; keep counted pending lowered generic/capture/container identity"
			}
		}
		membership := "counted"
		if candidate != nil {
			switch {
			case finder.template(candidate):
				why = "uninstantiated generic/any declaration; keep counted pending concrete allocation-site type"
			case finder.programScalarStorage(candidate):
				why = "scalar scratch storage; no owned graph elements"
			case selected[cycleNode{proven: candidate}] && !finder.weak(candidate) && !analysis.isLibraryType(candidate, "Promise"):
				membership = "in region"
				in++
				why = "owning-field SCC selected by programRegionSelection; declaration audit has no lowered capture cells"
			}
		}
		if err := writer.Write([]string{row[0], declared, membership, why, slotType}); err != nil {
			t.Fatal(err)
		}
	}
	writer.Flush()
	if err := writer.Error(); err != nil {
		t.Fatal(err)
	}
	if len(seen) != 1685 {
		t.Fatal("record omitted")
	}
	path := "../../docs/cycles-decision/membership.csv"
	if *updateProgramMembership {
		if err := os.WriteFile(path, output.Bytes(), 0644); err != nil {
			t.Fatal(err)
		}
	} else {
		recorded, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(recorded, output.Bytes()) {
			t.Fatal("membership changed; regenerate with -update-program-membership and review")
		}
	}
	t.Logf("1685 records: in region %d counted %d unresolved container identities %d; source files %d", in, 1685-in, unresolved, len(roots))
	// Independent anchors make missing/changed input records visible.
	required := []string{"D11", "D633", "D634", "D708", "D938", "K15", "K69"}
	sort.Strings(required)
	for _, id := range required {
		if !seen[id] {
			t.Fatalf("required census record missing: %s", id)
		}
	}
}
