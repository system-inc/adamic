// Package firstlinkphases holds the first-link harness's phase sets to the phases internal/lower records. It is a
// test-only package of its own, so its reads of the harness's report and of lower's source key this one unit and
// none of lower's products (#nm7z8fv, Keys' narrow shape).
package firstlinkphases

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"testing"
)

// firstLinkHarnessReport is the first-link harness's report, from this directory. Its BODY_PHASES and
// WHOLE_PROGRAM_PHASES name every phase a trap-mode record may carry, and a record whose phase isn't in either set
// fails the report, so the two sets and lower's calls have to agree.
const firstLinkHarnessReport = "../../../stage3/firstlink/report.py"

// firstLinkLowerDirectory is internal/lower, from this directory. Its *.go files but the tests are read.
const firstLinkLowerDirectory = ".."

// firstLinkHelpers are the five calls that carry a first-link phase, and the methods of those names whose own
// bodies pass a phase through rather than name one, so their bodies are left out by their declarations.
var firstLinkHelpers = []string{"firstLinkPass", "firstLinkProgramStop", "firstLinkCatch", "firstLinkRow", "firstLinkRecord"}

// firstLinkReportItem is one string in a set literal of report.py, which is read as text and never run.
var firstLinkReportItem = regexp.MustCompile(`"([^"]+)"`)

// firstLinkReportPhases is the named set in report's text, by literal parsing. A set that isn't there, or holds no
// string, fails the test, so the comparison can't pass empty.
func firstLinkReportPhases(t *testing.T, report, name string) []string {
	t.Helper()
	literal := regexp.MustCompile(`(?ms)^` + regexp.QuoteMeta(name) + `\s*=\s*\{(.*?)\}`).FindStringSubmatch(report)
	if literal == nil {
		t.Fatalf("%s holds no %s = {...} set", firstLinkHarnessReport, name)
	}
	var phases []string
	for _, item := range firstLinkReportItem.FindAllStringSubmatch(literal[1], -1) {
		phases = append(phases, item[1])
	}
	if len(phases) == 0 {
		t.Fatalf("%s's %s is empty", firstLinkHarnessReport, name)
	}
	slices.Sort(phases)
	return slices.Compact(phases)
}

// firstLinkStringLiteral is the value of a string literal, and whether the expression is one.
func firstLinkStringLiteral(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

// firstLinkHelperCall is the helper a call names through a selector (l.firstLinkRow(...)), or empty.
func firstLinkHelperCall(node ast.Node) (*ast.CallExpr, string) {
	call, ok := node.(*ast.CallExpr)
	if !ok {
		return nil, ""
	}
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || !slices.Contains(firstLinkHelpers, selector.Sel.Name) {
		return nil, ""
	}
	return call, selector.Sel.Name
}

// firstLinkRowPhase is a firstLinkRow call's literal phase, its third argument, and whether it has one.
func firstLinkRowPhase(call *ast.CallExpr) (string, bool) {
	if len(call.Args) < 3 {
		return "", false
	}
	return firstLinkStringLiteral(call.Args[2])
}

// firstLinkTrapPhase is the literal Phase of an ir.Trap composite literal, and whether the expression is one with
// a literal Phase.
func firstLinkTrapPhase(expression ast.Expr) (string, bool) {
	literal, ok := expression.(*ast.CompositeLit)
	if !ok {
		return "", false
	}
	if selector, ok := literal.Type.(*ast.SelectorExpr); !ok || selector.Sel.Name != "Trap" {
		return "", false
	}
	for _, element := range literal.Elts {
		pair, ok := element.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		if key, ok := pair.Key.(*ast.Ident); ok && key.Name == "Phase" {
			return firstLinkStringLiteral(pair.Value)
		}
	}
	return "", false
}

// firstLinkRowNames is, for one function declaration, each name that holds a row only from firstLinkRow calls
// with a literal phase: every assignment to it (its :=, every =) is such a call's first result, there is at least
// one, and the name is never a parameter, a result, a range binding or a var declaration anywhere in the
// declaration, closures included. One binding from elsewhere drops the name, so a later reassignment can't pass on
// the first assignment's word.
func firstLinkRowNames(declaration *ast.FuncDecl) map[string]bool {
	fromRow, elsewhere := map[string]bool{}, map[string]bool{}
	bindFields := func(fields *ast.FieldList) {
		if fields == nil {
			return
		}
		for _, field := range fields.List {
			for _, name := range field.Names {
				elsewhere[name.Name] = true
			}
		}
	}
	bindFields(declaration.Recv)
	ast.Inspect(declaration, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.FuncType:
			bindFields(node.Params)
			bindFields(node.Results)
		case *ast.RangeStmt:
			for _, bound := range []ast.Expr{node.Key, node.Value} {
				if name, ok := bound.(*ast.Ident); ok {
					elsewhere[name.Name] = true
				}
			}
		case *ast.ValueSpec:
			for _, name := range node.Names {
				elsewhere[name.Name] = true
			}
		case *ast.AssignStmt:
			call, helper := (*ast.CallExpr)(nil), ""
			if len(node.Rhs) == 1 {
				call, helper = firstLinkHelperCall(node.Rhs[0])
			}
			_, literal := "", false
			if helper == "firstLinkRow" {
				_, literal = firstLinkRowPhase(call)
			}
			for position, target := range node.Lhs {
				name, ok := target.(*ast.Ident)
				if !ok || name.Name == "_" {
					continue
				}
				if position == 0 && literal {
					fromRow[name.Name] = true
				} else {
					elsewhere[name.Name] = true
				}
			}
		}
		return true
	})
	names := map[string]bool{}
	for name := range fromRow {
		if !elsewhere[name] {
			names[name] = true
		}
	}
	return names
}

// firstLinkLoweringPhases is every phase lower's non-test files name, body and whole-program, each sorted, and
// every helper call outside the helpers' own bodies that names no phase the test can read, by file:line. A call
// reads when it is firstLinkPass, firstLinkProgramStop or firstLinkCatch with a literal first argument (the first
// two whole-program, the third a body phase), firstLinkRow with a literal third, or firstLinkRecord with an
// ir.Trap literal holding a literal Phase, or with a name firstLinkRowNames holds. Every *.go file there but the
// tests is parsed, so a phase moving to a new file is still seen, and a file that doesn't parse fails the test.
// override, by file name, is parsed in place of a file's text, so a mutant needs no file of its own. A package that
// names no phase on either side fails the test.
func firstLinkLoweringPhases(t *testing.T, override map[string]string) (body, whole, unread []string) {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(firstLinkLowerDirectory, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	files := token.NewFileSet()
	parsed := 0
	for _, path := range paths {
		name := filepath.Base(path)
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		text, overridden := override[name]
		if !overridden {
			contents, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			text = string(contents)
		}
		file, err := parser.ParseFile(files, name, text, parser.SkipObjectResolution)
		if err != nil {
			t.Fatalf("%s doesn't parse: %v", name, err)
		}
		parsed++
		for _, declaration := range file.Decls {
			rows := map[string]bool{}
			if function, ok := declaration.(*ast.FuncDecl); ok {
				if function.Recv != nil && slices.Contains(firstLinkHelpers, function.Name.Name) {
					continue
				}
				rows = firstLinkRowNames(function)
			}
			ast.Inspect(declaration, func(node ast.Node) bool {
				call, helper := firstLinkHelperCall(node)
				if call == nil {
					return true
				}
				phase, read := "", false
				switch helper {
				case "firstLinkPass", "firstLinkProgramStop", "firstLinkCatch":
					if len(call.Args) > 0 {
						phase, read = firstLinkStringLiteral(call.Args[0])
					}
				case "firstLinkRow":
					phase, read = firstLinkRowPhase(call)
				case "firstLinkRecord":
					if len(call.Args) > 0 {
						if phase, read = firstLinkTrapPhase(call.Args[0]); !read {
							if name, ok := call.Args[0].(*ast.Ident); ok && rows[name.Name] {
								return true
							}
						}
					}
				}
				switch {
				case !read:
					position := files.Position(call.Pos())
					unread = append(unread, position.Filename+":"+strconv.Itoa(position.Line)+" "+helper)
				case helper == "firstLinkPass" || helper == "firstLinkProgramStop":
					whole = append(whole, phase)
				default:
					body = append(body, phase)
				}
				return true
			})
		}
	}
	if parsed == 0 {
		t.Fatalf("no lowering source under %s", firstLinkLowerDirectory)
	}
	if len(body) == 0 || len(whole) == 0 {
		t.Fatalf("lower's %d non-test files name no first-link phase (body %d, whole-program %d)", parsed, len(body), len(whole))
	}
	slices.Sort(body)
	slices.Sort(whole)
	return slices.Compact(body), slices.Compact(whole), unread
}

// firstLinkPhaseGap says how the report's two sets differ from the phases lowering names, or is empty when they
// agree.
func firstLinkPhaseGap(reportBody, reportWhole, body, whole []string) string {
	var gaps []string
	for _, side := range []struct {
		name           string
		report, lowers []string
	}{{"BODY_PHASES", reportBody, body}, {"WHOLE_PROGRAM_PHASES", reportWhole, whole}} {
		for _, phase := range side.lowers {
			if !slices.Contains(side.report, phase) {
				gaps = append(gaps, side.name+" lacks "+phase+", which lowering names")
			}
		}
		for _, phase := range side.report {
			if !slices.Contains(side.lowers, phase) {
				gaps = append(gaps, side.name+" holds "+phase+", which lowering never names")
			}
		}
	}
	return strings.Join(gaps, "; ")
}

// firstLinkPlant is a lowering file's text with one change, failing when the anchor isn't there exactly once.
func firstLinkPlant(t *testing.T, file, before, after string) map[string]string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(firstLinkLowerDirectory, file))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(contents), before) != 1 {
		t.Fatalf("%s no longer holds %q once; the mutant can't be planted", file, before)
	}
	return map[string]string{file: strings.Replace(string(contents), before, after, 1)}
}

// The first-link harness's report names exactly the phases internal/lower records, and every helper call there
// outside the helpers' own bodies names its phase in a shape this test reads: BODY_PHASES the body shapes',
// WHOLE_PROGRAM_PHASES the whole-program ones'. Mutants, each named: report.py's BODY_PHASES without Statics; a
// firstLinkPass("probeOnly", ...) call in a lowering file; a firstLinkPass whose phase is a variable; a row
// reassigned from elsewhere before firstLinkRefusalGap records it.
func TestFirstLinkPhasesMatchHarnessReport(t *testing.T) {
	t.Parallel()
	// Reads: stage3/firstlink/report.py
	// Reads: internal/lower/*.go
	contents, err := os.ReadFile(firstLinkHarnessReport)
	if err != nil {
		t.Fatal(err)
	}
	report := string(contents)
	reportBody, reportWhole := firstLinkReportPhases(t, report, "BODY_PHASES"), firstLinkReportPhases(t, report, "WHOLE_PROGRAM_PHASES")
	body, whole, unread := firstLinkLoweringPhases(t, nil)
	if len(unread) != 0 {
		t.Fatalf("first-link calls whose phase this test can't read: %v", unread)
	}
	if gap := firstLinkPhaseGap(reportBody, reportWhole, body, whole); gap != "" {
		t.Fatalf("the harness's report and internal/lower differ: %s", gap)
	}

	dropped := `"Statics", `
	if strings.Count(report, dropped) != 1 {
		t.Fatalf("%s no longer holds %s once; the mutant can't be planted", firstLinkHarnessReport, dropped)
	}
	withoutStatics := firstLinkReportPhases(t, strings.Replace(report, dropped, "", 1), "BODY_PHASES")
	if gap := firstLinkPhaseGap(withoutStatics, reportWhole, body, whole); !strings.Contains(gap, "Statics") {
		t.Errorf("with Statics dropped from BODY_PHASES: got %q, want the gap to name Statics", gap)
	}

	passes := "\nfunc (l *lowering) firstLinkPass(name string,"
	probe := firstLinkPlant(t, "first_link_passes.go", passes, "\nfunc (l *lowering) firstLinkProbeOnly() error { return l.firstLinkPass(\"probeOnly\", false, nil, nil) }\n"+passes)
	probeBody, probeWhole, probeUnread := firstLinkLoweringPhases(t, probe)
	if gap := firstLinkPhaseGap(reportBody, reportWhole, probeBody, probeWhole); len(probeUnread) != 0 || !strings.Contains(gap, "probeOnly") {
		t.Errorf("with a firstLinkPass(\"probeOnly\", ...) call planted: got gap %q and unread %v, want the gap to name probeOnly", gap, probeUnread)
	}

	variable := firstLinkPlant(t, "first_link_passes.go", passes, "\nfunc (l *lowering) firstLinkProbeVariable(name string) error { return l.firstLinkPass(name, false, nil, nil) }\n"+passes)
	if _, _, variableUnread := firstLinkLoweringPhases(t, variable); len(variableUnread) != 1 || !strings.HasPrefix(variableUnread[0], "first_link_passes.go:") || !strings.HasSuffix(variableUnread[0], " firstLinkPass") {
		t.Errorf("with a firstLinkPass whose phase is a variable: got unread %v, want that one call named", variableUnread)
	}

	gap := "if row, stopped := l.firstLinkRow(nil, stop, \"Refusal\", nil, nil); stopped {\n"
	reassigned := firstLinkPlant(t, "first_link_refusals.go", gap, gap+"\t\trow = r\n")
	if _, _, reassignedUnread := firstLinkLoweringPhases(t, reassigned); len(reassignedUnread) != 1 || !strings.HasPrefix(reassignedUnread[0], "first_link_refusals.go:") || !strings.HasSuffix(reassignedUnread[0], " firstLinkRecord") {
		t.Errorf("with the row reassigned from elsewhere before it's recorded: got unread %v, want that record named", reassignedUnread)
	}
}
