package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const securityNoInterpolatedShellCommandId = "interpolatedShellCommand"

// securityNoInterpolatedShellCommandText is the rule's message, whose wording lives in
// `policy/messages/security-no-interpolated-shell-command.json`.
var securityNoInterpolatedShellCommandText = policy.MessageOf("nexus/security-no-interpolated-shell-command", securityNoInterpolatedShellCommandId)

func securityNoInterpolatedShellCommandMessage() rule.Message {
	return rule.Message{Id: securityNoInterpolatedShellCommandId, Description: securityNoInterpolatedShellCommandText.Render(nil)}
}

// SecurityNoInterpolatedShellCommand reports a shell command string built from a value whose type
// does not prove it is a literal, handed to a `node:child_process` function that runs it through a
// shell.
//
//	invalid: execSync(`sqlite3 "${databasePath}" "${query}"`)                     (both are `string`)
//	invalid: execSync('open "' + url + '"')
//	invalid: const command = `ffmpeg -i "${input}"`; execSync(command);
//	invalid: await promisify(exec)(`git log ${branch}`)
//	invalid: spawn('git', ['log', branch], { shell: true })
//	valid:   execFileSync('sqlite3', [databasePath, query])
//	valid:   execSync(`ps -Eww -o command= -p ${process.pid}`)                    (a `number`)
//	valid:   execSync(`git ${mode}`)                       (`mode: 'status' | 'diff'`, a literal union)
//	valid:   execSync('ps -Ao pid=,command=')
//	valid:   spawn('npx', ['-y', 'github:owner/package'], { shell: true })      (all literal)
//	valid:   execSync(command)                         (an opaque `string` parameter: not built here)
//
// # Where it came from
//
// `modules/apple/contacts/ContactsApi.ts:79` in ahra: the contact search builds
// `LIKE '%${word}%'` from the words an agent searches for, often a name read from an email, wraps
// the query in shell double quotes and hands it to `execSync`. `O'Brien` breaks the SQL and the
// catch around it turns that into "no contacts found", and a `$(...)` in the name runs as a
// command. Filed as `#5na28yw`; found by both research passes of the new-rules sweep (`#tevhg3f`),
// built as task `#ednac8q`.
//
// # Conditions, every one of them required
//
//   - **The callee is `node:child_process`'s, by the checker.** The call's resolved signature is
//     declared as a function directly inside `declare module "child_process"` or `"node:child_process"`
//     (the two names `@types/node` has used for the module), or as `__promisify__` in that
//     function's namespace, which is the signature `util.promisify(exec)` resolves to. So a
//     namespace import (`NodeChildProcess.execSync`), a named or renamed import, `import x =
//     require(...)` and a promisified wrapper are all seen, and a local function that happens to be
//     called `exec` never is. A `require` the checker types as `any` resolves no signature and is
//     not seen.
//   - **The function runs its command through a shell.** `exec` and `execSync` always do. `spawn`,
//     `spawnSync`, `execFile` and `execFileSync` do only when an object literal argument carries a
//     `shell` property whose type is `true` or a non-empty string literal (the shell to use);
//     node joins the command and every argument with spaces and gives the line to the shell.
//     Without that, the arguments reach the program as they are and nothing here is reported.
//   - **The command is built here.** The command argument (after parentheses, and through `const`
//     bindings to their initializer) is a template literal with substitutions, a `+`
//     concatenation, or a conditional with such a branch. An opaque command (a `string`
//     parameter, a property, a call's result) is not reported: it may be a whole script its caller
//     wrote as a literal, and nothing at this call says a value was spliced into it. Under a truthy
//     `shell`, node itself splices each element of an array literal of arguments into the line, so
//     an element is judged even when it is a bare reference.
//   - **A value in it is not provably literal.** The command is flattened into the text the author
//     wrote and the values spliced between it, through substitutions, `+` operands and `const`
//     bindings, because the checker types a nested template as `string` even when every part of it
//     is literal. A value is safe when its type, through a type parameter's constraint, is made
//     only of string literals, numbers, bigints, booleans, enum members, `null` and `undefined`:
//     text the author wrote, or text with no shell syntax in it. A conditional is judged branch by
//     branch, each as a command of its own. Anything else, a `string`, `any`, a template literal
//     type, a branded string, is a value the shell will parse.
//
// # A quoted heredoc's body is data
//
// ahra feeds SQL to `pscale shell` as `cat <<'EOSQL' | pscale shell ...\n${sql}\nEOSQL`. The shell
// expands nothing in the body of a heredoc whose delimiter is quoted, so `${sql}` there is never
// parsed as code, and reporting it would be false (`i18n-conversion.ts:243`, the first draft's one
// false finding). So the flattened text is scanned for heredocs: a value in a quoted heredoc's body
// is skipped, and a value on the opener line, after the closing line, or in an unquoted heredoc's
// body is judged as usual (`PlanetScaleApi.ts:380` still reports its database and branch).
// `securityNoInterpolatedShellCommandQuotedHeredocValues` gives the grammar it reads. The scan errs
// one way only: a `<<` it cannot read declines the whole command, and so does a value outside any
// body that might hold a `<<` (a union of string literals one of which does, or a conditional
// branch that opens one), since its body would run on over text judged without it. A value that
// itself holds the delimiter line and ends the body early is a miss the rule accepts.
//
// # What it deliberately does not see
//
//   - A command held in a `let`, a parameter, a property or returned from a function: see "built
//     here" above. These are the opaque commands the research counted separately.
//   - Arguments passed as anything but an array literal under a truthy `shell`, and spread
//     elements inside one: their elements are not visible at the call.
//   - A `shell` option given through a variable, a spread or a shorthand property, or typed as a
//     plain `string` or `boolean` that may be empty or false at run time.
//   - `execFile('sh', ['-c', script])` and its kin: an explicit shell invoked with a script
//     argument. The shell is the program there, chosen on purpose, and naming every shell and every
//     flag that takes a script is a list rather than a type.
//
// # Hand-escaped values are still reported
//
// The 11 `python3 -c`, `osascript -e` and `swift -e` sites in ahra wrap a script in single quotes
// after replacing each `'` in it with quote, backslash, quote, quote: an escape POSIX shells honour
// inside single quotes, so they quote correctly. They are reported anyway, because the escape is a
// convention every call site must repeat exactly (`MacOsApi.ts:71` spells it differently, also
// correctly), the command still reaches a shell, and the fix makes the escaping unnecessary rather
// than merely right. `JSON.stringify` is not such an escape: it leaves `$` and backticks live
// inside the double quotes it adds (`ConversationsBackup.ts:78`).
//
// # No fix
//
// Splitting a shell string into a program and an argument list changes what runs whenever the
// string used shell syntax on purpose (a pipe, `2>/dev/null`, a glob), so the rewrite is the
// author's.
var SecurityNoInterpolatedShellCommand = rule.Rule{
	Name:             "nexus/security-no-interpolated-shell-command",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			// Both the callee and the literal-ness of every piece are type questions. A name match
			// alone would report every function called `exec`, so with no checker the rule declines.
			return nil
		}
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				functionName := securityNoInterpolatedShellCommandFunction(ctx, node)
				arguments := call.Arguments.Nodes
				switch functionName {
				case "exec", "execSync":
				case "spawn", "spawnSync", "execFile", "execFileSync":
					if !securityNoInterpolatedShellCommandHasShell(ctx, arguments[1:]) {
						return
					}
				default:
					return
				}
				built := securityNoInterpolatedShellCommandBuilt(ctx, arguments[0], 0)
				if built != nil && securityNoInterpolatedShellCommandJudge(ctx, built, 0) == securityNoInterpolatedShellCommandUnsafe {
					ctx.ReportNode(arguments[0], securityNoInterpolatedShellCommandMessage())
					return
				}
				if functionName == "exec" || functionName == "execSync" || len(arguments) < 2 {
					return
				}
				// Under a truthy `shell`, node joins the command and each argument with spaces, so every
				// element of the argument list is spliced into the line the shell reads.
				argumentList := ast.SkipParentheses(arguments[1])
				if argumentList.Kind != ast.KindArrayLiteralExpression {
					return
				}
				for _, element := range argumentList.AsArrayLiteralExpression().Elements.Nodes {
					if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
						continue
					}
					if securityNoInterpolatedShellCommandJudge(ctx, element, 0) == securityNoInterpolatedShellCommandUnsafe {
						ctx.ReportNode(element, securityNoInterpolatedShellCommandMessage())
						return
					}
				}
			},
		}
	},
}

// securityNoInterpolatedShellCommandDepthLimit bounds the walk through substitutions,
// concatenations and `const` bindings. A chain this long is never written by hand; past it a command
// is not found to be built, or is declined, and either way a finding is missed, never invented.
const securityNoInterpolatedShellCommandDepthLimit = 16

// securityNoInterpolatedShellCommandFunction names the `node:child_process` function a call
// resolves to, or returns "" when the call is to anything else. A promisified function reports the
// name of the function it wraps.
func securityNoInterpolatedShellCommandFunction(ctx rule.Context, call *ast.Node) string {
	signature := ctx.TypeChecker.GetResolvedSignature(call)
	if signature == nil {
		return ""
	}
	declaration := signature.Declaration()
	if declaration == nil || declaration.Kind != ast.KindFunctionDeclaration || declaration.Name() == nil {
		return ""
	}
	functionName := declaration.Name().Text()
	module := securityNoInterpolatedShellCommandEnclosingModule(declaration)
	if module == nil {
		return ""
	}
	if functionName == "__promisify__" && module.Name().Kind == ast.KindIdentifier {
		functionName = module.Name().Text()
		module = securityNoInterpolatedShellCommandEnclosingModule(module)
		if module == nil {
			return ""
		}
	}
	if module.Name().Kind != ast.KindStringLiteral {
		return ""
	}
	switch module.Name().Text() {
	case "child_process", "node:child_process":
		return functionName
	}
	return ""
}

// securityNoInterpolatedShellCommandEnclosingModule is the module or namespace declaration whose
// body directly holds a declaration, or nil when it sits anywhere else.
func securityNoInterpolatedShellCommandEnclosingModule(declaration *ast.Node) *ast.Node {
	block := declaration.Parent
	if block == nil || block.Kind != ast.KindModuleBlock {
		return nil
	}
	module := block.Parent
	if module == nil || module.Kind != ast.KindModuleDeclaration {
		return nil
	}
	return module
}

// securityNoInterpolatedShellCommandHasShell says whether an object literal among the arguments
// sets `shell` to a value that is truthy by its type: `true`, or a non-empty string literal naming
// the shell. node tests the option for truthiness, so a plain `boolean` or `string` might still mean
// no shell and is not counted.
func securityNoInterpolatedShellCommandHasShell(ctx rule.Context, arguments []*ast.Node) bool {
	for _, argument := range arguments {
		argument = ast.SkipParentheses(argument)
		if argument.Kind != ast.KindObjectLiteralExpression {
			continue
		}
		for _, property := range argument.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind != ast.KindPropertyAssignment || property.Name() == nil {
				continue
			}
			if property.Name().Kind != ast.KindIdentifier && property.Name().Kind != ast.KindStringLiteral {
				continue
			}
			if property.Name().Text() != "shell" {
				continue
			}
			valueType := ctx.TypeChecker.GetTypeAtLocation(property.AsPropertyAssignment().Initializer)
			if valueType == nil {
				return false
			}
			return type_checking.Every(type_checking.UnionTypeParts(valueType), securityNoInterpolatedShellCommandIsTruthyShell)
		}
	}
	return false
}

// securityNoInterpolatedShellCommandIsTruthyShell says whether one member of a `shell` option's
// type always turns the shell on.
func securityNoInterpolatedShellCommandIsTruthyShell(part *checker.Type) bool {
	if type_checking.IsTypeFlagSet(part, checker.TypeFlagsBooleanLiteral) {
		value, _ := part.AsLiteralType().Value().(bool)
		return value
	}
	if type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral) {
		value, _ := part.AsLiteralType().Value().(string)
		return value != ""
	}
	return false
}

// securityNoInterpolatedShellCommandBuilt finds the expression a command is built by: a template
// with substitutions, a `+` concatenation, or a conditional with such a branch, looking through
// parentheses and `const` bindings. It returns nil for anything else, an opaque command this rule
// does not judge.
func securityNoInterpolatedShellCommandBuilt(ctx rule.Context, expression *ast.Node, depth int) *ast.Node {
	if depth > securityNoInterpolatedShellCommandDepthLimit {
		return nil
	}
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindTemplateExpression:
		return expression
	case ast.KindBinaryExpression:
		if expression.AsBinaryExpression().OperatorToken.Kind == ast.KindPlusToken {
			return expression
		}
	case ast.KindConditionalExpression:
		conditional := expression.AsConditionalExpression()
		if securityNoInterpolatedShellCommandBuilt(ctx, conditional.WhenTrue, depth+1) != nil ||
			securityNoInterpolatedShellCommandBuilt(ctx, conditional.WhenFalse, depth+1) != nil {
			return expression
		}
	case ast.KindIdentifier:
		if initializer := securityNoInterpolatedShellCommandConstInitializer(ctx, expression); initializer != nil {
			return securityNoInterpolatedShellCommandBuilt(ctx, initializer, depth+1)
		}
	}
	return nil
}

// securityNoInterpolatedShellCommandVerdict is what the rule concludes about a command or a piece
// of one.
type securityNoInterpolatedShellCommandVerdict int

const (
	// securityNoInterpolatedShellCommandSafe: every piece the shell parses is provably literal.
	securityNoInterpolatedShellCommandSafe securityNoInterpolatedShellCommandVerdict = iota
	// securityNoInterpolatedShellCommandUnsafe: some piece the shell parses is not.
	securityNoInterpolatedShellCommandUnsafe
	// securityNoInterpolatedShellCommandDeclined: the rule cannot tell which pieces the shell parses,
	// so it says nothing. Every decline is a finding missed, never one invented.
	securityNoInterpolatedShellCommandDeclined
)

// securityNoInterpolatedShellCommandPiece is one stretch of a flattened command: text the author
// wrote, or a value whose text is not known here.
type securityNoInterpolatedShellCommandPiece struct {
	text  string
	value *ast.Node
}

// securityNoInterpolatedShellCommandHole stands for a value's unknown text in a flattened command.
// It is a NUL, which no command line can carry, so it never matches anything the scanner looks for.
const securityNoInterpolatedShellCommandHole = "\x00"

// securityNoInterpolatedShellCommandJudge decides whether the shell parses a value in a command
// that is not provably literal. It flattens the command into text and values, finds the values the
// shell reads as data rather than code (inside a quoted heredoc's body), and judges the rest.
func securityNoInterpolatedShellCommandJudge(ctx rule.Context, expression *ast.Node, depth int) securityNoInterpolatedShellCommandVerdict {
	if depth > securityNoInterpolatedShellCommandDepthLimit {
		return securityNoInterpolatedShellCommandDeclined
	}
	var pieces []securityNoInterpolatedShellCommandPiece
	if !securityNoInterpolatedShellCommandFlatten(ctx, expression, depth, &pieces) {
		return securityNoInterpolatedShellCommandDeclined
	}
	quoted, understood := securityNoInterpolatedShellCommandQuotedHeredocValues(pieces)
	if !understood {
		return securityNoInterpolatedShellCommandDeclined
	}
	verdict := securityNoInterpolatedShellCommandSafe
	for index, piece := range pieces {
		if piece.value == nil || quoted[index] {
			continue
		}
		switch securityNoInterpolatedShellCommandJudgeValue(ctx, piece.value, depth) {
		case securityNoInterpolatedShellCommandDeclined:
			return securityNoInterpolatedShellCommandDeclined
		case securityNoInterpolatedShellCommandUnsafe:
			verdict = securityNoInterpolatedShellCommandUnsafe
		}
	}
	return verdict
}

// securityNoInterpolatedShellCommandFlatten appends an expression's pieces in order: the text of
// string literals and template parts, and the pieces of every substitution, `+` concatenation of
// strings and `const` binding it is built from. Anything else is one value. A value whose type is
// a single string literal is its text. It returns false when the walk runs too deep.
func securityNoInterpolatedShellCommandFlatten(ctx rule.Context, expression *ast.Node, depth int, pieces *[]securityNoInterpolatedShellCommandPiece) bool {
	if depth > securityNoInterpolatedShellCommandDepthLimit {
		return false
	}
	expression = ast.SkipParentheses(expression)
	switch expression.Kind {
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		*pieces = append(*pieces, securityNoInterpolatedShellCommandPiece{text: expression.Text()})
		return true
	case ast.KindTemplateExpression:
		template := expression.AsTemplateExpression()
		*pieces = append(*pieces, securityNoInterpolatedShellCommandPiece{text: template.Head.Text()})
		for _, span := range template.TemplateSpans.Nodes {
			if !securityNoInterpolatedShellCommandFlatten(ctx, span.AsTemplateSpan().Expression, depth+1, pieces) {
				return false
			}
			*pieces = append(*pieces, securityNoInterpolatedShellCommandPiece{text: span.AsTemplateSpan().Literal.Text()})
		}
		return true
	case ast.KindBinaryExpression:
		// A numeric `count + 1` flattens too; its operands are numbers and judged safe one by one.
		binary := expression.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindPlusToken {
			return securityNoInterpolatedShellCommandFlatten(ctx, binary.Left, depth+1, pieces) &&
				securityNoInterpolatedShellCommandFlatten(ctx, binary.Right, depth+1, pieces)
		}
	case ast.KindIdentifier:
		if initializer := securityNoInterpolatedShellCommandConstInitializer(ctx, expression); initializer != nil {
			switch ast.SkipParentheses(initializer).Kind {
			case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral, ast.KindTemplateExpression,
				ast.KindBinaryExpression, ast.KindConditionalExpression, ast.KindIdentifier:
				return securityNoInterpolatedShellCommandFlatten(ctx, initializer, depth+1, pieces)
			}
		}
	}
	if text, isSingleLiteral := securityNoInterpolatedShellCommandSingleStringLiteral(ctx, expression); isSingleLiteral {
		*pieces = append(*pieces, securityNoInterpolatedShellCommandPiece{text: text})
		return true
	}
	*pieces = append(*pieces, securityNoInterpolatedShellCommandPiece{value: expression})
	return true
}

// securityNoInterpolatedShellCommandJudgeValue judges one value the shell parses. A literal type
// is safe. A conditional is judged branch by branch, each as a command of its own, because the
// checker types a branch built from literals as `string`; a branch that opens a heredoc is
// declined, since its body would run on past the conditional into text judged without it. Any
// other value is unsafe.
func securityNoInterpolatedShellCommandJudgeValue(ctx rule.Context, value *ast.Node, depth int) securityNoInterpolatedShellCommandVerdict {
	if securityNoInterpolatedShellCommandLiteralType(ctx, value) {
		if securityNoInterpolatedShellCommandMayOpenHeredoc(ctx, value) {
			return securityNoInterpolatedShellCommandDeclined
		}
		return securityNoInterpolatedShellCommandSafe
	}
	if value.Kind != ast.KindConditionalExpression {
		return securityNoInterpolatedShellCommandUnsafe
	}
	conditional := value.AsConditionalExpression()
	verdict := securityNoInterpolatedShellCommandSafe
	for _, branch := range []*ast.Node{conditional.WhenTrue, conditional.WhenFalse} {
		var pieces []securityNoInterpolatedShellCommandPiece
		if !securityNoInterpolatedShellCommandFlatten(ctx, branch, depth+1, &pieces) {
			return securityNoInterpolatedShellCommandDeclined
		}
		for _, piece := range pieces {
			if strings.Contains(piece.text, "<<") {
				return securityNoInterpolatedShellCommandDeclined
			}
		}
		switch securityNoInterpolatedShellCommandJudge(ctx, branch, depth+1) {
		case securityNoInterpolatedShellCommandDeclined:
			return securityNoInterpolatedShellCommandDeclined
		case securityNoInterpolatedShellCommandUnsafe:
			verdict = securityNoInterpolatedShellCommandUnsafe
		}
	}
	return verdict
}

// securityNoInterpolatedShellCommandMayOpenHeredoc says whether a literal-typed value outside any
// heredoc body could hold a `<<`, which would open a body the scan of its command never saw: one of
// several string literals its type allows (a single one is spliced in as text and scanned).
func securityNoInterpolatedShellCommandMayOpenHeredoc(ctx rule.Context, value *ast.Node) bool {
	valueType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, value)
	return type_checking.Some(type_checking.UnionTypeParts(valueType), func(part *checker.Type) bool {
		if !type_checking.IsTypeFlagSet(part, checker.TypeFlagsStringLiteral) {
			return false
		}
		text, _ := part.AsLiteralType().Value().(string)
		return strings.Contains(text, "<<")
	})
}

// securityNoInterpolatedShellCommandQuotedHeredocValues marks the values that sit in the body of a
// heredoc whose delimiter is quoted (`<<'EOSQL'`, `<<"EOSQL"`, `<<\EOSQL`, each with an optional
// `-`). The shell expands nothing in such a body, so a value there is data, not code: this is how
// `pscale shell` is fed SQL in ahra. A body runs from the line after its opener to a line holding
// only the delimiter (after leading tabs, for `<<-`), and several openers on one line are read in
// order. An unquoted heredoc's body is expanded, so its values stay judged, but it is tracked so
// the bodies after it line up.
//
// It answers not understood when a `<<` is followed by something it cannot read as a delimiter,
// and the rule then declines the command. Reading a `<<` the shell would not treat as an opener
// (inside quotes, a comment, or another language's script) can only mark values safe that are not,
// which misses a finding; it never invents one. Neither can a value that itself holds the
// delimiter line and so ends the body early: that is a miss too, and a known one.
func securityNoInterpolatedShellCommandQuotedHeredocValues(pieces []securityNoInterpolatedShellCommandPiece) (map[int]bool, bool) {
	var builder strings.Builder
	valueAt := map[int]int{}
	for index, piece := range pieces {
		if piece.value != nil {
			valueAt[builder.Len()] = index
			builder.WriteString(securityNoInterpolatedShellCommandHole)
			continue
		}
		builder.WriteString(piece.text)
	}
	text := builder.String()
	quoted := map[int]bool{}

	type heredoc struct {
		delimiter     string
		stripsTabs    bool
		quotedOpening bool
	}
	var pending []heredoc
	var body *heredoc
	for lineStart := 0; lineStart <= len(text); {
		lineEnd := strings.IndexByte(text[lineStart:], '\n')
		if lineEnd < 0 {
			lineEnd = len(text)
		} else {
			lineEnd += lineStart
		}
		line := text[lineStart:lineEnd]
		if body != nil {
			candidate := line
			if body.stripsTabs {
				candidate = strings.TrimLeft(candidate, "\t")
			}
			if candidate == body.delimiter {
				body = nil
			} else if body.quotedOpening {
				for offset := lineStart; offset < lineEnd; offset++ {
					if index, isValue := valueAt[offset]; isValue {
						quoted[index] = true
					}
				}
			}
		} else {
			for offset := 0; offset < len(line); {
				opener := strings.Index(line[offset:], "<<")
				if opener < 0 {
					break
				}
				cursor := offset + opener + 2
				if cursor < len(line) && line[cursor] == '<' {
					// `<<<` is a here-string, not a heredoc.
					offset = cursor + 1
					continue
				}
				opened := heredoc{}
				if cursor < len(line) && line[cursor] == '-' {
					opened.stripsTabs = true
					cursor++
				}
				for cursor < len(line) && (line[cursor] == ' ' || line[cursor] == '\t') {
					cursor++
				}
				delimiter, length, quotedOpening := securityNoInterpolatedShellCommandHeredocDelimiter(line[cursor:])
				if length == 0 {
					return nil, false
				}
				opened.delimiter = delimiter
				opened.quotedOpening = quotedOpening
				pending = append(pending, opened)
				offset = cursor + length
			}
		}
		if body == nil && len(pending) > 0 && lineEnd < len(text) {
			body = &pending[0]
			pending = pending[1:]
		}
		lineStart = lineEnd + 1
	}
	return quoted, true
}

// securityNoInterpolatedShellCommandHeredocDelimiter reads the delimiter word at the start of the
// text after `<<`: `'word'`, `"word"`, `\word` or a bare word. It returns the word, how many bytes
// it spans, and whether it was quoted; a length of zero means it is none of those.
func securityNoInterpolatedShellCommandHeredocDelimiter(text string) (string, int, bool) {
	if text == "" {
		return "", 0, false
	}
	switch text[0] {
	case '\'', '"':
		closing := strings.IndexByte(text[1:], text[0])
		if closing <= 0 {
			return "", 0, false
		}
		word := text[1 : 1+closing]
		if strings.ContainsAny(word, securityNoInterpolatedShellCommandHole+"\n") {
			return "", 0, false
		}
		return word, closing + 2, true
	case '\\':
		length := securityNoInterpolatedShellCommandWordLength(text[1:])
		if length == 0 {
			return "", 0, false
		}
		return text[1 : 1+length], length + 1, true
	}
	length := securityNoInterpolatedShellCommandWordLength(text)
	if length == 0 {
		return "", 0, false
	}
	return text[:length], length, false
}

// securityNoInterpolatedShellCommandWordLength counts the letters, digits and underscores a text
// starts with.
func securityNoInterpolatedShellCommandWordLength(text string) int {
	length := 0
	for length < len(text) {
		character := text[length]
		if character != '_' && (character < '0' || character > '9') && (character < 'a' || character > 'z') && (character < 'A' || character > 'Z') {
			break
		}
		length++
	}
	return length
}

// securityNoInterpolatedShellCommandSingleStringLiteral is the text of a value whose type is one
// string literal, which is the text the value always holds.
func securityNoInterpolatedShellCommandSingleStringLiteral(ctx rule.Context, expression *ast.Node) (string, bool) {
	expressionType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)
	if expressionType == nil || !type_checking.IsTypeFlagSet(expressionType, checker.TypeFlagsStringLiteral) {
		return "", false
	}
	text, isString := expressionType.AsLiteralType().Value().(string)
	return text, isString
}

// securityNoInterpolatedShellCommandLiteralType says whether every member of an expression's type
// (a type parameter read through its constraint) is text the shell cannot misread as syntax the
// author did not write: a string literal, a number, a bigint, a boolean, an enum member, `null` or
// `undefined`.
func securityNoInterpolatedShellCommandLiteralType(ctx rule.Context, expression *ast.Node) bool {
	expressionType := type_checking.GetConstrainedTypeAtLocation(ctx.TypeChecker, expression)
	if expressionType == nil {
		return false
	}
	return type_checking.Every(type_checking.UnionTypeParts(expressionType), func(part *checker.Type) bool {
		return type_checking.IsTypeFlagSet(part, securityNoInterpolatedShellCommandLiteralFlags)
	})
}

// securityNoInterpolatedShellCommandLiteralFlags are the type flags of a piece whose text is either
// written by the author or free of shell syntax. `TypeFlagsNumberLike` covers `number`, number
// literals and numeric enums; a string enum member carries `TypeFlagsStringLiteral`.
const securityNoInterpolatedShellCommandLiteralFlags = checker.TypeFlagsStringLiteral | checker.TypeFlagsNumberLike |
	checker.TypeFlagsBigIntLike | checker.TypeFlagsBooleanLike | checker.TypeFlagsEnumLike |
	checker.TypeFlagsNull | checker.TypeFlagsUndefined

// securityNoInterpolatedShellCommandConstInitializer is the initializer of the `const` an
// identifier names, when it is a plain `const name = value` binding, or nil otherwise. A `const`
// cannot be reassigned, so its initializer is the only value the identifier ever holds.
func securityNoInterpolatedShellCommandConstInitializer(ctx rule.Context, identifier *ast.Node) *ast.Node {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || symbol.ValueDeclaration == nil {
		return nil
	}
	declaration := symbol.ValueDeclaration
	if declaration.Kind != ast.KindVariableDeclaration || !ast.IsVarConst(declaration) {
		return nil
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Name() == nil || variable.Name().Kind != ast.KindIdentifier {
		return nil
	}
	return variable.Initializer
}
