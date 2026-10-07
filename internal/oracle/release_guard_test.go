package oracle

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// Trace the actual option expressions on the ordinary release path. This deliberately fails on an
// unfamiliar option-producing expression instead of guessing default options after a refactor.
// Flags and the runtime key are evaluated by native itself; neither is copied into this guard.
func TestReleaseBuildConfigurationAgrees(t *testing.T) {
	t.Parallel()
	optionType := reflect.TypeOf(native.Options{})
	for index := range optionType.NumField() {
		kind := optionType.Field(index).Type.Kind()
		if kind != reflect.Bool && kind != reflect.Int && kind != reflect.String {
			t.Fatalf("Options.%s needs a value-semantics proof in the release trace", optionType.Field(index).Name)
		}
	}
	requireReleaseCall(t, "cmd/adamic", "main", "run", 0, "os.Args[1:]")
	requireReleaseCall(t, "cmd/adamic", "run", "build", 2, "arguments[4:]")
	requireReleaseCall(t, "internal/oracle", "released", "cachedNative", 2, "true")
	requireReleaseCall(t, "internal/oracle", "released", "releasedUncached", 1, "program")
	requireReleaseCall(t, "internal/oracle", "cachedNative", "releasedUncached", 1, "program")
	shipping := traceReleaseOptions(t, releaseFunction(t, "cmd/adamic", "build"))
	oracle := traceReleaseOptions(t, releaseFunction(t, "internal/oracle", "releasedUncached"))
	shippingFlags, oracleFlags := native.Flags(shipping), native.Flags(oracle)
	shippingKey, oracleKey := native.RuntimeFlagsKey(shippingFlags), native.RuntimeFlagsKey(oracleFlags)
	t.Logf("shipping: options=%#v flags=%q runtime-key=%s", shipping, shippingFlags, shippingKey)
	t.Logf("oracle: options=%#v flags=%q runtime-key=%s", oracle, oracleFlags, oracleKey)
	if shipping != oracle || !reflect.DeepEqual(shippingFlags, oracleFlags) || shippingKey != oracleKey {
		t.Fatalf("release configurations differ\nshipping: options=%#v flags=%q runtime-key=%s\noracle: options=%#v flags=%q runtime-key=%s",
			shipping, shippingFlags, shippingKey, oracle, oracleFlags, oracleKey)
	}
	// Hold the connection to RuntimeLibrary's input, not just our accessor. Both Build paths must
	// pass their unchanged options to Flags and RuntimeLibrary, whose cache must receive Flags.
	requireReleaseCall(t, "internal/native", "Build", "Flags", 0, "options")
	requireReleaseCall(t, "internal/native", "Build", "RuntimeLibrary", 1, "options")
	requireReleaseCall(t, "internal/native", "RuntimeLibrary", "cachedRuntime", 1, "Flags(options)")
	requireReleaseCall(t, "internal/native", "cachedRuntime", "runtimeKey", 1, "flags")
	requireReleaseCall(t, "internal/native", "RuntimeFlagsKey", "runtimeKey", 1, "flags")
	for index, flag := range shippingFlags {
		changed := append([]string{}, shippingFlags[:index]...)
		changed = append(changed, shippingFlags[index+1:]...)
		if native.RuntimeFlagsKey(changed) == shippingKey {
			t.Errorf("runtime key omitted %s\nshipping: %q\nchanged: %q", flag, shippingFlags, changed)
		}
	}
	// Exercise an arbitrary future flag too, including the current LTO candidate.
	changed := append(append([]string{}, shippingFlags...), "-flto=thin")
	if native.RuntimeFlagsKey(changed) == shippingKey {
		t.Errorf("runtime key omitted an added flag\nshipping: %q\nchanged: %q", shippingFlags, changed)
	}
}

func releaseFunction(t *testing.T, directory, name string) *ast.FuncDecl {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(repository, directory, "*.go"))
	if err != nil {
		t.Fatal(err)
	}
	var found *ast.FuncDecl
	for _, path := range paths {
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Contains(source, []byte("func "+name+"(")) {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, source, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			if function, ok := declaration.(*ast.FuncDecl); ok && function.Recv == nil && function.Name.Name == name {
				if found != nil {
					t.Fatalf("ambiguous release function %s.%s", directory, name)
				}
				found = function
			}
		}
	}
	if found == nil {
		t.Fatalf("release function %s.%s changed; update the trace", directory, name)
	}
	return found
}

func releaseSyntax(node ast.Node) string {
	var text bytes.Buffer
	if err := format.Node(&text, token.NewFileSet(), node); err != nil {
		panic(err)
	}
	return text.String()
}

func requireReleaseCall(t *testing.T, directory, function, callee string, index int, want string) {
	t.Helper()
	body := releaseFunction(t, directory, function).Body
	if directory == "cmd/adamic" && function == "run" {
		body = releaseDispatchBody(t, body)
	}
	calls := 0
	ast.Inspect(body, func(node ast.Node) bool {
		if call, ok := node.(*ast.CallExpr); ok && releaseSyntax(call.Fun) == callee {
			calls++
			if len(call.Args) <= index || releaseSyntax(call.Args[index]) != want {
				t.Errorf("%s.%s must carry its release configuration into %s: got %s, want argument %d = %s", directory, function, callee, releaseSyntax(call), index, want)
			}
		}
		// A mutation of these value options would invalidate the passthrough proof.
		if assignment, ok := node.(*ast.AssignStmt); ok {
			for _, target := range assignment.Lhs {
				if identifier, ok := target.(*ast.Ident); ok && (identifier.Name == "options" || directory == "cmd/adamic" && (function == "run" || function == "main") && identifier.Name == "arguments") {
					t.Errorf("%s.%s changes options; update the release trace", directory, function)
				}
				if field, ok := target.(*ast.SelectorExpr); ok && releaseSyntax(field.X) == "options" {
					t.Errorf("%s.%s changes options; update the release trace", directory, function)
				}
			}
		}
		if address, ok := node.(*ast.UnaryExpr); ok && address.Op == token.AND && releaseSyntax(address.X) == "options" {
			t.Errorf("%s.%s exposes mutable options; update the release trace", directory, function)
		}
		if increment, ok := node.(*ast.IncDecStmt); ok {
			if field, ok := increment.X.(*ast.SelectorExpr); ok && releaseSyntax(field.X) == "options" {
				t.Errorf("%s.%s changes options; update the release trace", directory, function)
			}
		}
		return true
	})
	if calls != 1 {
		t.Errorf("%s.%s: want one %s call on the shared build path, got %d", directory, function, callee, calls)
	}
}

// Select the dispatch arm for the actual four-argument native release invocation. A WASI arm
// may call the same builder with additional options; it is a different command invocation.
func releaseDispatchBody(t *testing.T, body *ast.BlockStmt) *ast.BlockStmt {
	t.Helper()
	if len(body.List) == 0 {
		t.Fatal("empty command dispatcher")
	}
	dispatch, ok := body.List[0].(*ast.SwitchStmt)
	if !ok || dispatch.Tag != nil || dispatch.Init != nil {
		t.Fatal("command dispatcher changed; update the release trace")
	}
	arguments := []string{"build", "fixture.a", "-o", "binary"}
	var evaluate func(ast.Expr) any
	evaluate = func(expression ast.Expr) any {
		switch expression := expression.(type) {
		case *ast.BasicLit:
			if expression.Kind == token.INT {
				value, err := strconv.Atoi(expression.Value)
				if err == nil {
					return value
				}
			}
			if expression.Kind == token.STRING {
				value, err := strconv.Unquote(expression.Value)
				if err == nil {
					return value
				}
			}
		case *ast.IndexExpr:
			if releaseSyntax(expression.X) == "arguments" {
				index, ok := evaluate(expression.Index).(int)
				if ok && index >= 0 && index < len(arguments) {
					return arguments[index]
				}
			}
		case *ast.CallExpr:
			if releaseSyntax(expression) == "len(arguments)" {
				return len(arguments)
			}
		case *ast.BinaryExpr:
			left := evaluate(expression.X)
			if expression.Op == token.LAND {
				value, ok := left.(bool)
				if ok && !value {
					return false
				}
				if ok {
					return evaluate(expression.Y)
				}
			}
			right := evaluate(expression.Y)
			if expression.Op == token.EQL {
				return reflect.DeepEqual(left, right)
			}
			if expression.Op == token.GEQ {
				left, leftOK := left.(int)
				right, rightOK := right.(int)
				if leftOK && rightOK {
					return left >= right
				}
			}
		}
		t.Fatalf("cannot trace release dispatch condition %s", releaseSyntax(expression))
		return nil
	}
	for _, statement := range dispatch.Body.List {
		clause, ok := statement.(*ast.CaseClause)
		if !ok || len(clause.List) == 0 {
			t.Fatal("release command fell through to the default arm")
		}
		for _, condition := range clause.List {
			selected, ok := evaluate(condition).(bool)
			if !ok {
				t.Fatal("release dispatch condition is not boolean")
			}
			if selected {
				return &ast.BlockStmt{List: clause.Body}
			}
		}
	}
	t.Fatal("no release command dispatch arm")
	return nil
}

func traceReleaseOptions(t *testing.T, function *ast.FuncDecl) native.Options {
	t.Helper()
	// adamic build FILE -o OUTPUT has no optional arguments. The compiler-only tsgo arm is not
	// selected. Unrelated loading and error handling may be skipped only if it cannot touch options.
	values := map[string]any{"arguments": []string{}}
	var evaluate func(ast.Expr) any
	evaluate = func(expression ast.Expr) any {
		switch expression := expression.(type) {
		case *ast.Ident:
			if expression.Name == "true" {
				return true
			}
			if expression.Name == "false" {
				return false
			}
			if expression.Name == "nil" {
				return nil
			}
			if value, ok := values[expression.Name]; ok {
				return value
			}
		case *ast.SelectorExpr:
			if options, ok := evaluate(expression.X).(native.Options); ok {
				field := reflect.ValueOf(options).FieldByName(expression.Sel.Name)
				if field.IsValid() && field.CanInterface() {
					return field.Interface()
				}
			}
		case *ast.BasicLit:
			if expression.Kind == token.INT {
				value, err := strconv.Atoi(expression.Value)
				if err == nil {
					return value
				}
			}
			if expression.Kind == token.STRING {
				value, err := strconv.Unquote(expression.Value)
				if err == nil {
					return value
				}
			}
		case *ast.CompositeLit:
			if releaseSyntax(expression.Type) == "native.Options" {
				options := native.Options{}
				for _, element := range expression.Elts {
					entry, ok := element.(*ast.KeyValueExpr)
					if !ok {
						t.Fatalf("%s: unnamed option: %s", function.Name, releaseSyntax(element))
					}
					field := reflect.ValueOf(&options).Elem().FieldByName(releaseSyntax(entry.Key))
					value := reflect.ValueOf(evaluate(entry.Value))
					if !field.IsValid() || !field.CanSet() || field.Type() != value.Type() {
						t.Fatalf("%s: unsupported option %s", function.Name, releaseSyntax(entry))
					}
					field.Set(value)
				}
				return options
			}
		case *ast.CallExpr:
			if releaseSyntax(expression.Fun) == "len" && len(expression.Args) == 1 {
				if arguments, ok := evaluate(expression.Args[0]).([]string); ok {
					return len(arguments)
				}
			}
		case *ast.BinaryExpr:
			if expression.Op == token.LAND {
				left, ok := evaluate(expression.X).(bool)
				if ok && !left {
					return false
				}
				if ok {
					return evaluate(expression.Y)
				}
			}
			if expression.Op == token.EQL || expression.Op == token.NEQ {
				equal := reflect.DeepEqual(evaluate(expression.X), evaluate(expression.Y))
				return equal == (expression.Op == token.EQL)
			}
			if expression.Op == token.LSS {
				left, leftOK := evaluate(expression.X).(int)
				right, rightOK := evaluate(expression.Y).(int)
				if leftOK && rightOK {
					return left < right
				}
			}
		}
		t.Fatalf("%s: cannot trace release expression %s; update the guard", function.Name, releaseSyntax(expression))
		return nil
	}
	touchesOptions := func(node ast.Node) bool {
		found := false
		ast.Inspect(node, func(node ast.Node) bool {
			if identifier, ok := node.(*ast.Ident); ok {
				if _, ok := values[identifier.Name].(native.Options); ok {
					found = true
				}
			}
			return true
		})
		return found
	}
	assign := func(statement *ast.AssignStmt) {
		if statement.Tok != token.ASSIGN && statement.Tok != token.DEFINE {
			t.Fatalf("%s: compound option assignment needs a trace: %s", function.Name, releaseSyntax(statement))
		}
		if len(statement.Lhs) != len(statement.Rhs) {
			t.Fatalf("%s: cannot trace %s", function.Name, releaseSyntax(statement))
		}
		for index, target := range statement.Lhs {
			if identifier, ok := target.(*ast.Ident); ok {
				values[identifier.Name] = evaluate(statement.Rhs[index])
			} else if field, ok := target.(*ast.SelectorExpr); ok {
				name := releaseSyntax(field.X)
				options, ok := values[name].(native.Options)
				if !ok {
					t.Fatalf("%s: cannot trace %s", function.Name, releaseSyntax(statement))
				}
				value := reflect.ValueOf(evaluate(statement.Rhs[index]))
				destination := reflect.ValueOf(&options).Elem().FieldByName(field.Sel.Name)
				if !destination.IsValid() || !destination.CanSet() || destination.Type() != value.Type() {
					t.Fatalf("%s: cannot trace %s", function.Name, releaseSyntax(statement))
				}
				destination.Set(value)
				values[name] = options
			} else {
				t.Fatalf("%s: cannot trace %s", function.Name, releaseSyntax(statement))
			}
		}
	}
	var walk func([]ast.Stmt) (native.Options, bool)
	walk = func(statements []ast.Stmt) (native.Options, bool) {
		for _, statement := range statements {
			if declaration, ok := statement.(*ast.DeclStmt); ok {
				if general, ok := declaration.Decl.(*ast.GenDecl); ok {
					for _, specification := range general.Specs {
						if value, ok := specification.(*ast.ValueSpec); ok && value.Type != nil && releaseSyntax(value.Type) == "error" && len(value.Values) == 0 {
							for _, name := range value.Names {
								values[name.Name] = nil
							}
						}
					}
				}
			}
			if conditional, ok := statement.(*ast.IfStmt); ok {
				if releaseSyntax(conditional.Cond) == "native.UsesTSGo(program)" {
					otherwise, ok := conditional.Else.(*ast.BlockStmt)
					if !ok {
						t.Fatalf("%s: no ordinary build arm", function.Name)
					}
					if options, found := walk(otherwise.List); found {
						return options, true
					}
					continue
				}
				if conditional.Init != nil {
					// Validation takes these scalar options by value and cannot change the caller's
					// configuration. It may refuse the build, which is outside this agreement check.
					if initial, ok := conditional.Init.(*ast.AssignStmt); ok && len(initial.Rhs) == 1 {
						if call, ok := initial.Rhs[0].(*ast.CallExpr); ok && releaseSyntax(call.Fun) == "native.ValidateOptions" && len(call.Args) == 1 {
							if _, ok := evaluate(call.Args[0]).(native.Options); !ok {
								t.Fatal("validation options changed")
							}
							continue
						}
					}
					if options, found := walk([]ast.Stmt{conditional.Init}); found {
						return options, true
					}
				}
				if touchesOptions(conditional) {
					condition, ok := evaluate(conditional.Cond).(bool)
					if !ok {
						t.Fatalf("%s: conditional option use %s; update the guard", function.Name, releaseSyntax(conditional))
					}
					if condition {
						if options, found := walk(conditional.Body.List); found {
							return options, true
						}
					} else if conditional.Else != nil {
						otherwise, ok := conditional.Else.(*ast.BlockStmt)
						if !ok {
							t.Fatal("release conditional changed")
						}
						if options, found := walk(otherwise.List); found {
							return options, true
						}
					}
					continue
				}
			}
			if loop, ok := statement.(*ast.ForStmt); ok && touchesOptions(loop.Body) {
				initial, ok := loop.Init.(*ast.AssignStmt)
				if !ok {
					t.Fatalf("%s: cannot trace release loop", function.Name)
				}
				assign(initial)
				if condition, ok := evaluate(loop.Cond).(bool); !ok || condition {
					t.Fatalf("%s: release option loop unexpectedly executes", function.Name)
				}
				continue
			}
			var build *ast.CallExpr
			ast.Inspect(statement, func(node ast.Node) bool {
				if call, ok := node.(*ast.CallExpr); ok && releaseSyntax(call.Fun) == "native.Build" {
					if build != nil {
						t.Fatalf("%s: multiple build paths; update the guard", function.Name)
					}
					build = call
				}
				return true
			})
			if build != nil {
				if len(build.Args) != 3 {
					t.Fatalf("%s: build signature changed", function.Name)
				}
				options, ok := evaluate(build.Args[2]).(native.Options)
				if !ok {
					t.Fatalf("%s: build options changed", function.Name)
				}
				return options, true
			}
			if assignment, ok := statement.(*ast.AssignStmt); ok {
				initializes := false
				for _, expression := range assignment.Rhs {
					if literal, ok := expression.(*ast.CompositeLit); ok && releaseSyntax(literal.Type) == "native.Options" {
						initializes = true
					}
				}
				if initializes || touchesOptions(assignment) {
					assign(assignment)
					continue
				}
			}
			if touchesOptions(statement) {
				t.Fatalf("%s: cannot trace option use %s; update the guard", function.Name, releaseSyntax(statement))
			}
		}
		return native.Options{}, false
	}
	options, found := walk(function.Body.List)
	if !found {
		t.Fatalf("%s: no native.Build on the release path", function.Name)
	}
	return options
}
