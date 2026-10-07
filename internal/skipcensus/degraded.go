package skipcensus

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
)

var inputVariable = regexp.MustCompile(`\bADAMIC_[A-Z0-9_]+\b`)

// MissingInputVariables uses the same vocabulary for source and runtime output.
// Each line is independent: a normal message elsewhere cannot supply a variable.
func MissingInputVariables(message string) []string {
	var variables []string
	for _, line := range strings.Split(message, "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(line, "set ADAMIC_") || strings.Contains(lower, "not checked") || strings.Contains(lower, "absent") {
			variables = append(variables, inputVariable.FindAllString(line, -1)...)
		}
	}
	if len(variables) == 0 {
		return nil
	}
	return unique(variables)
}

func identity(row Row, test string) string {
	text := row.Condition
	if row.Kind == "degraded-input" {
		text += "\n" + row.Kind + "\n" + row.Message
	}
	sum := sha256.Sum256([]byte(text))
	return fmt.Sprintf("%s:%x", test, sum[:])
}

func sourceString(expression ast.Expr, seen map[*ast.Object]bool) (string, bool) {
	if id, ok := expression.(*ast.Ident); ok {
		if id.Obj == nil || seen[id.Obj] {
			return "", false
		}
		seen[id.Obj] = true
		return sourceString(declarationValue(id), seen)
	}

	if binary, ok := expression.(*ast.BinaryExpr); ok && binary.Op == token.ADD {
		left, l := sourceString(binary.X, seen)
		right, r := sourceString(binary.Y, seen)
		return left + right, l && r
	}
	if paren, ok := expression.(*ast.ParenExpr); ok {
		return sourceString(paren.X, seen)
	}
	return stringValue(expression, seen)
}

func censusCall(set *token.FileSet, call *ast.CallExpr, testingAlias string, file *ast.File) (string, string, []string) {
	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", "", nil
	}
	message := ""
	if len(call.Args) > 0 {
		message = printed(set, call.Args[0])
	}
	if testingReceiver(set, selector.X, testingAlias, map[*ast.Object]bool{}) {
		if isSkipName(selector.Sel.Name) {
			return "skip", message, nil
		}
		if selector.Sel.Name != "Log" && selector.Sel.Name != "Logf" {
			return "", "", nil
		}
	} else {
		receiver, ok := selector.X.(*ast.Ident)
		if !ok || receiver.Obj != nil {
			return "", "", nil
		}
		alias := "fmt"
		for _, imp := range file.Imports {
			path, _ := strconv.Unquote(imp.Path.Value)
			if path == "fmt" && imp.Name != nil {
				alias = imp.Name.Name
			}
		}
		if receiver.Name != alias {
			return "", "", nil
		}
		switch selector.Sel.Name {
		case "Print", "Printf", "Println":
		case "Fprint", "Fprintf", "Fprintln":
			if len(call.Args) < 2 {
				return "", "", nil
			}
			destination := printed(set, call.Args[0])
			testOutput := false
			if writer, ok := call.Args[0].(*ast.CallExpr); ok {
				if method, ok := writer.Fun.(*ast.SelectorExpr); ok && method.Sel.Name == "Output" {
					testOutput = testingReceiver(set, method.X, testingAlias, map[*ast.Object]bool{})
				}
			}
			if destination != "os.Stdout" && destination != "os.Stderr" && !testOutput {
				return "", "", nil
			}
			call = &ast.CallExpr{Args: call.Args[1:]}
		default:
			return "", "", nil
		}
	}
	var pieces []string
	for _, arg := range call.Args {
		if value, known := sourceString(arg, map[*ast.Object]bool{}); known {
			pieces = append(pieces, value)
		}
	}
	separator := " "
	if selector.Sel.Name == "Print" || selector.Sel.Name == "Fprint" {
		separator = ""
	}
	text := strings.Join(pieces, separator)
	variables := MissingInputVariables(text)
	if len(variables) == 0 {
		return "", "", nil
	}
	if strings.HasSuffix(selector.Sel.Name, "f") && len(pieces) > 0 {
		// Preserve the format separately from constant arguments used to name
		// inputs; runtime matching accepts supported Go format substitutions.
		text = pieces[0]
	}
	return "degraded-input", strconv.Quote(text), variables
}

var diagnosticFormat = regexp.MustCompile(`%[-+ #0]*[0-9]*(?:\.[0-9]+)?[sdvq]`)

func diagnosticReasonMatches(literal, line string, variables []string) bool {
	actual := MissingInputVariables(line)
	if strings.Join(actual, ",") != strings.Join(variables, ",") {
		return false
	}
	if strings.Contains(line, literal) {
		return true
	}
	var pattern strings.Builder
	previous := 0
	for _, span := range diagnosticFormat.FindAllStringIndex(literal, -1) {
		pattern.WriteString(regexp.QuoteMeta(literal[previous:span[0]]))
		if literal[span[1]-1] == 'd' {
			pattern.WriteString(`[+-]?[0-9]+`)
		} else {
			pattern.WriteString(`[^\n]*`)
		}
		previous = span[1]
	}
	if previous == 0 {
		return false
	}
	pattern.WriteString(regexp.QuoteMeta(literal[previous:]))
	matched, _ := regexp.MatchString(pattern.String(), line)
	return matched
}
