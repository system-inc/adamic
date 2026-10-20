// Package waitdelay is adamiclint's rule against exec.Cmd.WaitDelay on a command whose output goes to memory.
//
// When a command's Stdout or Stderr is anything but an *os.File, exec copies it through a pipe of its own, and with
// WaitDelay set, Wait closes those reads a fixed time after the process exits and reports exec.ErrWaitDelay if the
// copying had not finished. So a run that is slow to drain its output, or whose descendant still holds a pipe, is red
// by chance (#n3gk22h), and on Loom one such red holds a whole block of other people's changes. An *os.File goes to
// the child as it is, with nothing to copy, so WaitDelay there only bounds the wait. The cure is childguard.Capture,
// which reads each stream to its end with no timer.
//
// The rule reads one function at a time. A WaitDelay is flagged when the same function gives that command an output
// that isn't an *os.File, calls its Output or CombinedOutput, hands it to childguard.Run or childguard.CombinedOutput,
// or returns it, since a caller may buffer what it is handed: stage1/cohere/markdownblocks' leafCompositionCommand set
// WaitDelay and returned the command, and its caller gave it bytes.Buffer outputs. childguard.Run wraps every output
// in a writer of its own, never an *os.File, and childguard.CombinedOutput gives the command a bytes.Buffer, so exec
// copies through pipes under both; childguard.Capture gives it *os.File pipes, so a command handed to it is not. A
// line carrying // WaitDelay: <reason>, on it or in the comment just above, is a use someone has thought about, and
// the reason is what the next reader needs.
package waitdelay

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/types/typeutil"
)

// Message names the cure, so the author can act on the finding without looking anything up.
const Message = "exec.Cmd.WaitDelay with buffered output reports ErrWaitDelay by chance when a slow run drains late; " +
	"run it with childguard.Capture, or say why in // WaitDelay: <reason> (#n3gk22h)"

var Analyzer = &analysis.Analyzer{
	Name: "waitdelay",
	Doc:  "flags exec.Cmd.WaitDelay on a command whose output is buffered or whose callers may buffer it (#n3gk22h)",
	Run:  run,
}

// childguardPath is the package that holds the cure; how it bounds its own reads is its business.
const childguardPath = "github.com/system-inc/adamic/internal/childguard"

const escapePrefix = "WaitDelay:"

func run(pass *analysis.Pass) (any, error) {
	// An external test package's path carries _test, and it belongs to childguard as much as the package does.
	if strings.TrimSuffix(pass.Pkg.Path(), "_test") == childguardPath {
		return nil, nil
	}
	for _, file := range pass.Files {
		escaped := escapedLines(pass.Fset, file)
		for _, declaration := range file.Decls {
			check(pass, declaration, escaped)
		}
	}
	return nil, nil
}

// command names an exec.Cmd the way a function's code reaches it: the variable it starts from, and the fields walked
// from there, so command.WaitDelay and command.Stdout are the same command, and job.command.WaitDelay another one.
type command struct {
	root types.Object
	path string
}

// delay is one place WaitDelay is set. literal is the exec.Cmd composite literal that sets it, or nil for an
// assignment; reached says whether the code names the command, which a literal handed straight to a call doesn't.
type delay struct {
	position token.Pos
	command  command
	reached  bool
	literal  *ast.CompositeLit
}

// check reads one top-level declaration whole, function literals inside it included, since a closure that buffers
// the command its function set WaitDelay on is the same mistake.
func check(pass *analysis.Pass, declaration ast.Decl, escaped map[int]bool) {
	var delays []delay
	buffered := map[command]bool{}
	returned := map[command]bool{}
	literalCommands := map[*ast.CompositeLit]command{}
	literalBuffered := map[*ast.CompositeLit]bool{}
	literalReturned := map[*ast.CompositeLit]bool{}
	// A literal is visited after the statement that holds it, so the statement says which command it is first.
	nameLiteral := func(target ast.Expr, value ast.Expr) {
		literal := commandLiteral(pass, value)
		if literal == nil {
			return
		}
		if name, ok := commandOf(pass, target); ok {
			literalCommands[literal] = name
		}
	}
	ast.Inspect(declaration, func(node ast.Node) bool {
		switch node := node.(type) {
		case *ast.AssignStmt:
			for index, target := range node.Lhs {
				if len(node.Lhs) == len(node.Rhs) {
					nameLiteral(target, node.Rhs[index])
				}
				selector, ok := ast.Unparen(target).(*ast.SelectorExpr)
				if !ok {
					continue
				}
				switch cmdField(pass, selector) {
				case "WaitDelay":
					name, reached := commandOf(pass, selector.X)
					delays = append(delays, delay{position: selector.Sel.Pos(), command: name, reached: reached})
				case "Stdout", "Stderr":
					if !buffers(pass, assignedType(pass, node, index)) {
						continue
					}
					if name, ok := commandOf(pass, selector.X); ok {
						buffered[name] = true
					}
				}
			}
		case *ast.ValueSpec:
			for index, target := range node.Names {
				if index < len(node.Values) {
					nameLiteral(target, node.Values[index])
				}
			}
		case *ast.ReturnStmt:
			for _, result := range node.Results {
				if literal := commandLiteral(pass, result); literal != nil {
					literalReturned[literal] = true
				}
				if name, ok := commandOf(pass, result); ok && isCmd(pass.TypesInfo.TypeOf(result)) {
					returned[name] = true
				}
			}
		case *ast.CompositeLit:
			if !isCmd(pass.TypesInfo.TypeOf(node)) {
				return true
			}
			for _, element := range node.Elts {
				pair, ok := element.(*ast.KeyValueExpr)
				if !ok {
					continue
				}
				key, ok := pair.Key.(*ast.Ident)
				if !ok {
					continue
				}
				field, ok := pass.TypesInfo.Uses[key].(*types.Var)
				if !ok || !isCmdField(field) {
					continue
				}
				switch field.Name() {
				case "WaitDelay":
					name, reached := literalCommands[node]
					delays = append(delays, delay{position: key.Pos(), command: name, reached: reached, literal: node})
				case "Stdout", "Stderr":
					if buffers(pass, pass.TypesInfo.Types[pair.Value]) {
						literalBuffered[node] = true
					}
				}
			}
		case *ast.CallExpr:
			if guardBuffers(typeutil.StaticCallee(pass.TypesInfo, node)) {
				for _, argument := range node.Args {
					if literal := commandLiteral(pass, argument); literal != nil {
						literalBuffered[literal] = true
					}
					if name, ok := commandOf(pass, argument); ok && isCmd(pass.TypesInfo.TypeOf(argument)) {
						buffered[name] = true
					}
				}
			}
			selector, ok := ast.Unparen(node.Fun).(*ast.SelectorExpr)
			if !ok {
				return true
			}
			switch cmdMethod(pass, selector) {
			case "Output", "CombinedOutput":
				if name, ok := commandOf(pass, selector.X); ok {
					buffered[name] = true
				}
			}
		}
		return true
	})
	for _, delay := range delays {
		flagged := delay.literal != nil && (literalBuffered[delay.literal] || literalReturned[delay.literal])
		if delay.reached && (buffered[delay.command] || returned[delay.command]) {
			flagged = true
		}
		if flagged && !escaped[pass.Fset.Position(delay.position).Line] {
			pass.Reportf(delay.position, "%s", Message)
		}
	}
}

// escapedLines are the lines a // WaitDelay: <reason> comment covers: its own, for a comment at the end of a line,
// and the line just after its comment group, so a reason may wrap onto lines of its own above the code.
func escapedLines(fileSet *token.FileSet, file *ast.File) map[int]bool {
	lines := map[int]bool{}
	for _, group := range file.Comments {
		for _, comment := range group.List {
			text, ok := strings.CutPrefix(comment.Text, "//")
			if !ok {
				continue
			}
			reason, ok := strings.CutPrefix(strings.TrimSpace(text), escapePrefix)
			// An escape with no reason says nothing to the next reader, so it escapes nothing.
			if !ok || strings.TrimSpace(reason) == "" {
				continue
			}
			lines[fileSet.Position(comment.Pos()).Line] = true
			lines[fileSet.Position(group.End()).Line+1] = true
		}
	}
	return lines
}

// commandOf names the command an expression reaches, through parentheses, & and *, or reports false when the
// expression is something no later line could name again, such as a call's result or an element of a slice.
func commandOf(pass *analysis.Pass, expression ast.Expr) (command, bool) {
	switch expression := ast.Unparen(expression).(type) {
	case *ast.Ident:
		object := pass.TypesInfo.ObjectOf(expression)
		if object == nil {
			return command{}, false
		}
		return command{root: object}, true
	case *ast.SelectorExpr:
		selection, ok := pass.TypesInfo.Selections[expression]
		if !ok {
			// A qualified identifier: a package-level variable of another package.
			object := pass.TypesInfo.ObjectOf(expression.Sel)
			if object == nil {
				return command{}, false
			}
			return command{root: object}, true
		}
		if selection.Kind() != types.FieldVal {
			return command{}, false
		}
		inner, ok := commandOf(pass, expression.X)
		if !ok {
			return command{}, false
		}
		return command{root: inner.root, path: inner.path + "." + expression.Sel.Name}, true
	case *ast.StarExpr:
		return commandOf(pass, expression.X)
	case *ast.UnaryExpr:
		if expression.Op == token.AND {
			return commandOf(pass, expression.X)
		}
	}
	return command{}, false
}

// commandLiteral is the exec.Cmd composite literal an expression is, through parentheses and &, or nil.
func commandLiteral(pass *analysis.Pass, expression ast.Expr) *ast.CompositeLit {
	expression = ast.Unparen(expression)
	if unary, ok := expression.(*ast.UnaryExpr); ok && unary.Op == token.AND {
		expression = ast.Unparen(unary.X)
	}
	literal, ok := expression.(*ast.CompositeLit)
	if !ok || !isCmd(pass.TypesInfo.TypeOf(literal)) {
		return nil
	}
	return literal
}

// assignedType is the type of the value an assignment gives its index'th target, a call's result included.
func assignedType(pass *analysis.Pass, assignment *ast.AssignStmt, index int) types.TypeAndValue {
	if len(assignment.Lhs) == len(assignment.Rhs) {
		return pass.TypesInfo.Types[assignment.Rhs[index]]
	}
	if len(assignment.Rhs) == 1 {
		if tuple, ok := pass.TypesInfo.TypeOf(assignment.Rhs[0]).(*types.Tuple); ok && index < tuple.Len() {
			return types.TypeAndValue{Type: tuple.At(index).Type()}
		}
	}
	return types.TypeAndValue{}
}

// buffers says a value given as Stdout or Stderr makes exec copy the output itself: anything but nil or an *os.File.
// A value whose type is unknown is not counted, since the rule would be guessing.
func buffers(pass *analysis.Pass, value types.TypeAndValue) bool {
	if value.Type == nil || value.IsNil() {
		return false
	}
	return !isOSFile(value.Type)
}

func isOSFile(t types.Type) bool {
	pointer, ok := types.Unalias(t).(*types.Pointer)
	if !ok {
		return false
	}
	return isNamed(pointer.Elem(), "os", "File")
}

// isCmd says t is exec.Cmd or a pointer to it.
func isCmd(t types.Type) bool {
	if t == nil {
		return false
	}
	if pointer, ok := types.Unalias(t).(*types.Pointer); ok {
		t = pointer.Elem()
	}
	return isNamed(t, "os/exec", "Cmd")
}

func isNamed(t types.Type, packagePath, name string) bool {
	named, ok := types.Unalias(t).(*types.Named)
	if !ok {
		return false
	}
	object := named.Obj()
	return object.Pkg() != nil && object.Pkg().Path() == packagePath && object.Name() == name
}

// isCmdField says field is one of exec.Cmd's own fields, by identity, so a struct of some other type with a field
// named WaitDelay is never mistaken for it.
func isCmdField(field *types.Var) bool {
	if !field.IsField() || field.Pkg() == nil || field.Pkg().Path() != "os/exec" {
		return false
	}
	object := field.Pkg().Scope().Lookup("Cmd")
	if object == nil {
		return false
	}
	structure, ok := object.Type().Underlying().(*types.Struct)
	if !ok {
		return false
	}
	for index := range structure.NumFields() {
		if structure.Field(index) == field {
			return true
		}
	}
	return false
}

// cmdField is the name of the exec.Cmd field a selector reaches, an embedded exec.Cmd's included, or "".
func cmdField(pass *analysis.Pass, selector *ast.SelectorExpr) string {
	selection, ok := pass.TypesInfo.Selections[selector]
	if !ok || selection.Kind() != types.FieldVal {
		return ""
	}
	field, ok := selection.Obj().(*types.Var)
	if !ok || !isCmdField(field) {
		return ""
	}
	return field.Name()
}

// guardBuffers says callee is childguard's Run or CombinedOutput, by the function object, so a function of the same
// name anywhere else is never mistaken for it. Capture is not among them: its outputs are *os.File pipes.
func guardBuffers(callee *types.Func) bool {
	if callee == nil || callee.Pkg() == nil || callee.Pkg().Path() != childguardPath {
		return false
	}
	if signature, ok := callee.Type().(*types.Signature); !ok || signature.Recv() != nil {
		return false
	}
	return callee.Name() == "Run" || callee.Name() == "CombinedOutput"
}

// cmdMethod is the name of the exec.Cmd method a selector calls, or "".
func cmdMethod(pass *analysis.Pass, selector *ast.SelectorExpr) string {
	selection, ok := pass.TypesInfo.Selections[selector]
	if !ok || selection.Kind() != types.MethodVal {
		return ""
	}
	method, ok := selection.Obj().(*types.Func)
	if !ok {
		return ""
	}
	signature, ok := method.Type().(*types.Signature)
	if !ok || signature.Recv() == nil || !isCmd(signature.Recv().Type()) {
		return ""
	}
	return method.Name()
}
