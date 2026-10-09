package guard

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

// A rule declares exactly what it reads through ctx.Program, and this is what makes that structural
// rather than remembered.
//
// The findings cache decides from the declaration (rule.ProgramReads) whether a rule's findings can be
// replayed and on what key: compiler options and the default library are in the key already, this
// file's module resolution is in the type fingerprint, and anything else keeps the rule out of the
// cache (#b8k3bp6). An undeclared read serves a stale result when what it read changes: zero findings,
// forever, indistinguishable from a clean tree. An overdeclared read keeps a rule out of the cache it
// could have joined. So both directions fail here.
//
// ctx.Program refuses an undeclared read at runtime too, but only on the paths a run happens to take.
// This reads each rule file's syntax, so it covers every path:
//
//   - every method a rule calls on its context's Program, through any chain (`state.ctx.Program`);
//   - every function a rule hands its Program to, by what that function reads, followed through the
//     helpers in the rule packages and in `checking` to a fixed point;
//   - a Program handed to a function this cannot see is refused, since its reads cannot be counted.
//
// Text was what this checked until #ym4v8bc found two holes in it, both in the Tailwind rules. A
// helper that reads the program through a rule.Context hides the read from its callers, so a file that
// reads the program through a context and defines no rule still fails here: such a helper takes a
// rule.Program, and every call site then names ctx.Program. And a declaration in a comment is not a
// declaration: only a ProgramReads key in a composite literal counts.
func TestRulesDeclareExactlyWhatTheyReadOfTheProgram(t *testing.T) {
	t.Parallel()

	ruleFiles := ruleSourceFiles(t)
	if len(ruleFiles) == 0 {
		// A check that found nothing to check passes for the wrong reason, which is the same shape
		// as the defect it guards against.
		t.Fatal("found no rule source files, so this test proved nothing")
	}
	helpers := programHelperReads(t, append(append(append([]string{}, ruleFiles...), checkingSourceFiles(t)...),
		ruleInterfaceSourceFiles(t)...))

	sawProgramReader := false
	for _, path := range ruleFiles {
		facts := readProgramFacts(t, path, helpers)
		if !facts.readsProgram {
			continue
		}
		sawProgramReader = true

		if !facts.definesRule {
			t.Errorf("%s reads the program through a rule.Context and defines no rule, so the rules "+
				"calling it read the program without naming it and nothing makes them declare it. "+
				"Take a rule.Program instead, so each call site names ctx.Program", path)
			continue
		}
		for _, unseen := range facts.unseenCallees {
			t.Errorf("%s hands its Program to %s, which this cannot read, so its reads cannot be counted",
				path, unseen)
		}
		if missing := facts.reads &^ facts.declared; missing != 0 {
			t.Errorf("%s reads %s through ctx.Program and does not declare it (declares %s), so a findings "+
				"cache would replay its findings after what it read had changed", path, missing, facts.declared)
		}
		if extra := facts.declared &^ facts.reads; extra != 0 {
			t.Errorf("%s declares %s and reads none of it, which keeps it out of a cache it could join; "+
				"declare %s", path, extra, facts.reads)
		}
	}

	if !sawProgramReader {
		// Dozens of rules read the program today. Zero would mean the search is looking in the wrong
		// place rather than that the tree got cleaner.
		t.Fatal("no rule source reads ctx.Program, which is implausible: this check is " +
			"looking at the wrong files")
	}
}

// The check sees code and not text, and counts reads by kind, in every direction it could be wrong.
//
// Each source is a shape this guard could misjudge, written to a file and read the way the guard
// reads the tree, so the guard is shown able to fail rather than only shown passing.
func TestReadProgramFactsSeesCodeNotComments(t *testing.T) {
	t.Parallel()

	helpers := map[string]programRead{"type_checking.IsPromiseLike": readsCompilerOptions | readsDefaultLibrary,
		"probe.fingerprint": readsCompilerOptions}
	testCases := []struct {
		name   string
		source string
		want   programFacts
	}{
		{
			// The shape DesignSystemForProgram had: a helper reading the program for its callers.
			name: "a helper reading the program through a context",
			source: `package probe
func helper(ctx rule.Context) bool { return ctx.Program.Options() != nil }`,
			want: programFacts{readsProgram: true, reads: readsCompilerOptions},
		},
		{
			name: "a declaration in a comment",
			source: `package probe
// Every caller must declare ` + "`ProgramReads: rule.ReadsOtherFiles`" + `.
var Probe = rule.Rule{Run: func(ctx rule.Context, options any) rule.Listeners { _ = ctx.Program.SourceFiles(); return nil }}`,
			want: programFacts{readsProgram: true, definesRule: true, reads: readsOtherFiles},
		},
		{
			name: "a read in a comment only",
			source: `package probe
// This rule never touches ctx.Program.
var Probe = rule.Rule{Run: func(ctx rule.Context, options any) rule.Listeners { return nil }}`,
			want: programFacts{definesRule: true},
		},
		{
			// A context under another name, and reached through a struct, is still the context.
			name: "a context under another name, through a chain",
			source: `package probe
var Probe = rule.Rule{ProgramReads: rule.ReadsModuleResolution, Run: func(c rule.Context, options any) rule.Listeners { _ = state{c}.c.Program.ResolveModule(nil, nil); return nil }}`,
			want: programFacts{readsProgram: true, definesRule: true, reads: readsModuleResolution, declared: readsModuleResolution},
		},
		{
			// A read through a helper counts as the helper's reads.
			name: "a read through a helper",
			source: `package probe
var Probe = rule.Rule{ProgramReads: rule.ReadsCompilerOptions, Run: func(ctx rule.Context, options any) rule.Listeners { _ = type_checking.IsPromiseLike(ctx.Program, nil, nil); return nil }}`,
			want: programFacts{readsProgram: true, definesRule: true, reads: readsCompilerOptions | readsDefaultLibrary, declared: readsCompilerOptions},
		},
		{
			// The walk calls a rule's ProgramFingerprint with the rule's own view of the program, so its
			// reads count as the rule's.
			name: "a read through the program fingerprint",
			source: `package probe
var Probe = rule.Rule{ProgramReads: rule.ReadsCompilerOptions, ProgramFingerprint: fingerprint, Run: func(ctx rule.Context, options any) rule.Listeners { return nil }}`,
			want: programFacts{readsProgram: true, definesRule: true, reads: readsCompilerOptions, declared: readsCompilerOptions},
		},
		{
			name: "a program fingerprint this cannot see",
			source: `package probe
var Probe = rule.Rule{ProgramFingerprint: elsewhere, Run: func(ctx rule.Context, options any) rule.Listeners { return nil }}`,
			want: programFacts{readsProgram: true, definesRule: true, unseenCallees: []string{"probe.elsewhere"}},
		},
		{
			name: "a Program handed to a function this cannot see",
			source: `package probe
var Probe = rule.Rule{Run: func(ctx rule.Context, options any) rule.Listeners { somewhere.Else(ctx.Program); return nil }}`,
			want: programFacts{readsProgram: true, definesRule: true, unseenCallees: []string{"somewhere.Else"}},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			path := filepath.Join(t.TempDir(), "probe.go")
			if err := os.WriteFile(path, []byte(testCase.source), 0o644); err != nil {
				t.Fatal(err)
			}
			got := readProgramFacts(t, path, helpers)
			if got.readsProgram != testCase.want.readsProgram || got.definesRule != testCase.want.definesRule ||
				got.reads != testCase.want.reads || got.declared != testCase.want.declared ||
				strings.Join(got.unseenCallees, ",") != strings.Join(testCase.want.unseenCallees, ",") {
				t.Errorf("got %+v, want %+v", got, testCase.want)
			}
		})
	}
}

// programRead mirrors rule.ProgramRead. This package reads rule sources as syntax and does not link
// the rules, so it names the kinds the way the sources spell them.
type programRead uint8

const (
	readsCompilerOptions programRead = 1 << iota
	readsDefaultLibrary
	readsModuleResolution
	readsOtherFiles
	readsDesignSystem
)

func (reads programRead) String() string {
	names := []string{}
	for read, name := range map[programRead]string{
		readsCompilerOptions:  "ReadsCompilerOptions",
		readsDefaultLibrary:   "ReadsDefaultLibrary",
		readsModuleResolution: "ReadsModuleResolution",
		readsOtherFiles:       "ReadsOtherFiles",
		readsDesignSystem:     "ReadsDesignSystem",
	} {
		if reads&read != 0 {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	if len(names) == 0 {
		return "nothing"
	}
	return strings.Join(names, " | ")
}

// programReadsByConstant is each declaration constant as a source spells it.
var programReadsByConstant = map[string]programRead{
	"ReadsCompilerOptions":  readsCompilerOptions,
	"ReadsDefaultLibrary":   readsDefaultLibrary,
	"ReadsModuleResolution": readsModuleResolution,
	"ReadsOtherFiles":       readsOtherFiles,
	"ReadsDesignSystem":     readsDesignSystem,
}

// programReadsByMethod is what each rule.Program method reads. It must name every method the
// interface has, which the build cannot check from here, so a method missing from this table is
// refused as unseen rather than counted as reading nothing.
var programReadsByMethod = map[string]programRead{
	"Options":                        readsCompilerOptions,
	"GetCurrentDirectory":            readsCompilerOptions,
	"UseCaseSensitiveFileNames":      readsCompilerOptions,
	"IsSourceFileDefaultLibrary":     readsDefaultLibrary,
	"DefaultLibraryPath":             readsDefaultLibrary,
	"ResolveModule":                  readsModuleResolution,
	"GetSourceFileForResolvedModule": readsOtherFiles,
	"SourceFiles":                    readsOtherFiles,
	"GetSourceFile":                  readsOtherFiles,
	"FS":                             readsOtherFiles,
	"DesignSystemFS":                 readsDesignSystem,
	"Identity":                       0,
}

// programFacts is what one rule file does with the program, read from its syntax.
type programFacts struct {
	// readsProgram is a `.Program` selector on a rule.Context, under any name and through any chain.
	readsProgram bool
	// definesRule is a rule.Rule composite literal.
	definesRule bool
	// reads is what the file reads through its context's Program, directly and through helpers.
	reads programRead
	// declared is the file's ProgramReads key.
	declared programRead
	// unseenCallees are functions handed the Program whose reads this cannot count.
	unseenCallees []string
}

// programUse is one place a Program value is used: a method called on it, or a function it is passed
// to, named `package.Function` in the caller's terms.
type programUse struct {
	methods []string
	callees []string
}

// usesOf collects how the Program values isProgram picks out are used inside node.
func usesOf(node ast.Node, packageName string, isProgram func(ast.Expr) bool) programUse {
	uses := programUse{}
	ast.Inspect(node, func(inner ast.Node) bool {
		call, isCall := inner.(*ast.CallExpr)
		if !isCall {
			return true
		}
		if selector, isSelector := call.Fun.(*ast.SelectorExpr); isSelector && isProgram(selector.X) {
			uses.methods = append(uses.methods, selector.Sel.Name)
		}
		for _, argument := range call.Args {
			if !isProgram(argument) {
				continue
			}
			switch callee := call.Fun.(type) {
			case *ast.Ident:
				uses.callees = append(uses.callees, packageName+"."+callee.Name)
			case *ast.SelectorExpr:
				if qualifier, isIdentifier := callee.X.(*ast.Ident); isIdentifier {
					uses.callees = append(uses.callees, qualifier.Name+"."+callee.Sel.Name)
				} else {
					uses.callees = append(uses.callees, "(method)."+callee.Sel.Name)
				}
			default:
				uses.callees = append(uses.callees, "(an expression)")
			}
		}
		return true
	})
	return uses
}

// programHelperReads is what every function taking a rule.Program reads of it, followed through the
// functions it hands the Program on to, to a fixed point. Keyed `package.Function`.
func programHelperReads(t *testing.T, paths []string) map[string]programRead {
	t.Helper()
	uses := map[string]programUse{}
	for _, path := range paths {
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatalf("parsing %s: %v", path, err)
		}
		for _, declaration := range file.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil || function.Recv != nil {
				continue
			}
			names := map[string]bool{}
			for _, field := range function.Type.Params.List {
				// Inside the rule package itself the type is spelled bare (rule.FingerprintPath).
				bare, isBare := field.Type.(*ast.Ident)
				if isRuleSelector(field.Type, "Program") || (file.Name.Name == "rule" && isBare && bare.Name == "Program") {
					for _, name := range field.Names {
						names[name.Name] = true
					}
				}
			}
			if len(names) == 0 {
				continue
			}
			isProgram := func(expression ast.Expr) bool {
				identifier, isIdentifier := expression.(*ast.Ident)
				return isIdentifier && names[identifier.Name]
			}
			uses[file.Name.Name+"."+function.Name.Name] = usesOf(function.Body, file.Name.Name, isProgram)
		}
	}

	reads := map[string]programRead{}
	for changed := true; changed; {
		changed = false
		for name, use := range uses {
			total := reads[name]
			for _, method := range use.methods {
				total |= programReadsByMethod[method]
			}
			for _, callee := range use.callees {
				total |= reads[callee]
			}
			if total != reads[name] {
				reads[name] = total
				changed = true
			}
		}
	}
	for name := range uses {
		if _, present := reads[name]; !present {
			reads[name] = 0
		}
	}
	return reads
}

// readProgramFacts parses one file and reports what it does with the program.
//
// Names of rule.Context parameters are collected first, since a rule may call its context anything;
// ctx and context are added for the rare function literal whose parameter types are elided.
func readProgramFacts(t *testing.T, path string, helpers map[string]programRead) programFacts {
	t.Helper()

	file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
	if err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	contextNames := map[string]bool{"ctx": true, "context": true}
	ast.Inspect(file, func(node ast.Node) bool {
		if function, isFunction := node.(*ast.FuncType); isFunction && function.Params != nil {
			for _, field := range function.Params.List {
				if isRuleSelector(field.Type, "Context") {
					for _, name := range field.Names {
						contextNames[name.Name] = true
					}
				}
			}
		}
		return true
	})

	// isContextProgram is `<context>.Program`, where the context is a name or the last link of a
	// chain (`state.ctx.Program`).
	isContextProgram := func(expression ast.Expr) bool {
		selector, isSelector := expression.(*ast.SelectorExpr)
		if !isSelector || selector.Sel.Name != "Program" {
			return false
		}
		switch receiver := selector.X.(type) {
		case *ast.Ident:
			return contextNames[receiver.Name]
		case *ast.SelectorExpr:
			return contextNames[receiver.Sel.Name]
		}
		return false
	}

	facts := programFacts{}
	// The function a rule names as its ProgramFingerprint, which the walk calls with the rule's own view of
	// the program, so what it reads is the rule's to declare as much as anything Run reads.
	var fingerprint string
	ast.Inspect(file, func(node ast.Node) bool {
		switch typed := node.(type) {
		case *ast.SelectorExpr:
			if isContextProgram(typed) {
				facts.readsProgram = true
			}
		case *ast.KeyValueExpr:
			if key, isIdentifier := typed.Key.(*ast.Ident); isIdentifier && key.Name == "ProgramFingerprint" {
				if function, isNamed := typed.Value.(*ast.Ident); isNamed {
					fingerprint = function.Name
				} else {
					facts.unseenCallees = append(facts.unseenCallees, "a ProgramFingerprint that is not a named function")
				}
				facts.readsProgram = true
			}
			if key, isIdentifier := typed.Key.(*ast.Ident); isIdentifier && key.Name == "ProgramReads" {
				ast.Inspect(typed.Value, func(inner ast.Node) bool {
					if selector, isSelector := inner.(*ast.SelectorExpr); isSelector {
						facts.declared |= programReadsByConstant[selector.Sel.Name]
					}
					return true
				})
			}
		case *ast.CompositeLit:
			if isRuleSelector(typed.Type, "Rule") {
				facts.definesRule = true
			}
		}
		return true
	})

	if fingerprint != "" {
		read, known := helpers[file.Name.Name+"."+fingerprint]
		if !known {
			facts.unseenCallees = append(facts.unseenCallees, file.Name.Name+"."+fingerprint)
		}
		facts.reads |= read
	}

	uses := usesOf(file, file.Name.Name, isContextProgram)
	for _, method := range uses.methods {
		read, known := programReadsByMethod[method]
		if !known {
			facts.unseenCallees = append(facts.unseenCallees, "Program."+method)
			continue
		}
		facts.reads |= read
	}
	for _, callee := range uses.callees {
		read, known := helpers[callee]
		if !known {
			facts.unseenCallees = append(facts.unseenCallees, callee)
			continue
		}
		facts.reads |= read
	}
	return facts
}

// isRuleSelector reports whether an expression is `rule.<name>`.
func isRuleSelector(expression ast.Expr, name string) bool {
	selector, isSelector := expression.(*ast.SelectorExpr)
	if !isSelector {
		return false
	}
	packageName, isIdentifier := selector.X.(*ast.Ident)
	return isIdentifier && packageName.Name == "rule" && selector.Sel.Name == name
}

// ruleSourceFiles lists the non-test Go files under internal/lint/rules.
func ruleSourceFiles(t *testing.T) []string {
	t.Helper()
	return goSourceFilesUnder(t, "../lint/rules")
}

// ruleInterfaceSourceFiles lists the non-test Go files of the rule package, whose helpers taking a
// Program (rule.FingerprintPath) read it for the rules that call them.
func ruleInterfaceSourceFiles(t *testing.T) []string {
	t.Helper()
	return goSourceFilesUnder(t, "../lint/rule")
}

// checkingSourceFiles lists the non-test Go files of the shared type-checking helpers, which take a
// rule.Program from the rules that call them.
func checkingSourceFiles(t *testing.T) []string {
	t.Helper()
	return goSourceFilesUnder(t, "../lint/checking")
}

func goSourceFilesUnder(t *testing.T, root string) []string {
	t.Helper()

	var found []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		found = append(found, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walking %s: %v", root, err)
	}
	return found
}
