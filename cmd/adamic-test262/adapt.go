package main

import (
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

// adaptKinds are the in-memory rewrites --adapt counts. Nothing here is written back to the
// test262 checkout. A rewrite that would change meaning is not applied: the test stays as written,
// and the refusal stays a fact about Adamic rather than about test262's spelling.
const (
	adaptVarToLet      = "var-to-let"
	adaptCallbackParam = "callback-param"
	adaptStrictEq      = "strict-eq"
	adaptThrowError    = "throw-error"
)

// adapted is one test body after --adapt, plus how many times each rewrite fired.
type adapted struct {
	Source string
	Counts map[string]int
}

// adaptSource rewrites a test body. If the body uses syntax the scan does not understand, the body
// comes back unchanged: a partial edit from a confused scan would be a meaning change we did not prove.
func adaptSource(source string) adapted {
	parsed := parseBody(source)
	if parsed.failed {
		return adapted{Source: source}
	}
	parsed.finish()
	if parsed.failed || len(parsed.edits) == 0 {
		return adapted{Source: source}
	}
	return adapted{Source: applyEdits(source, parsed.edits), Counts: parsed.counts}
}

type edit struct {
	start int
	end   int
	text  string
}

func applyEdits(source string, edits []edit) string {
	sort.Slice(edits, func(i int, j int) bool { return edits[i].start < edits[j].start })
	for index := len(edits) - 1; index >= 0; index-- {
		change := edits[index]
		source = source[:change.start] + change.text + source[change.end:]
	}
	return source
}

type token struct {
	kind          string
	text          string
	start         int
	end           int
	newlineBefore bool
}

func tokenize(source string) []token {
	var tokens []token
	lineStart := true
	for index := 0; index < len(source); {
		character := source[index]
		if character == ' ' || character == '\t' || character == '\r' || character == '\n' {
			if character == '\n' {
				lineStart = true
			}
			index++
			continue
		}
		if character == '/' && index+1 < len(source) && source[index+1] == '/' {
			end := strings.IndexByte(source[index:], '\n')
			if end < 0 {
				break
			}
			index += end
			continue
		}
		if character == '/' && index+1 < len(source) && source[index+1] == '*' {
			end := strings.Index(source[index+2:], "*/")
			if end < 0 {
				return nil
			}
			if strings.Contains(source[index:index+2+end+2], "\n") {
				lineStart = true
			}
			index += 2 + end + 2
			continue
		}
		newline := lineStart
		lineStart = false
		if character == '\'' || character == '"' {
			end := endOfString(source, index)
			tokens = append(
				tokens,
				token{kind: "str", text: source[index:end], start: index, end: end, newlineBefore: newline},
			)
			index = end
			continue
		}
		if character == '`' {
			end := endOfTemplate(source, index)
			if end < 0 {
				return nil
			}
			tokens = append(
				tokens,
				token{kind: "tmpl", text: source[index:end], start: index, end: end, newlineBefore: newline},
			)
			index = end
			continue
		}
		if character == '/' && allowsRegex(tokens) {
			end := endOfRegex(source, index)
			if end < 0 {
				return nil
			}
			tokens = append(
				tokens,
				token{kind: "regexp", text: source[index:end], start: index, end: end, newlineBefore: newline},
			)
			index = end
			continue
		}
		if isDigit(character) || (character == '.' && index+1 < len(source) && isDigit(source[index+1])) {
			end := endOfNumber(source, index)
			tokens = append(
				tokens,
				token{kind: "num", text: source[index:end], start: index, end: end, newlineBefore: newline},
			)
			index = end
			continue
		}
		runeValue, size := utf8.DecodeRuneInString(source[index:])
		if isIdentifierStart(runeValue) || (runeValue >= 0x80 && unicode.IsLetter(runeValue)) {
			end := endOfIdent(source, index)
			tokens = append(
				tokens,
				token{kind: "ident", text: source[index:end], start: index, end: end, newlineBefore: newline},
			)
			index = end
			continue
		}
		_ = size
		end := index + 1
		if index+2 < len(source) {
			three := source[index : index+3]
			if three == "===" || three == "!==" || three == ">>>" || three == "..." {
				end = index + 3
			}
		}
		if end == index+1 && index+1 < len(source) {
			two := source[index : index+2]
			switch two {
			case "==", "!=", "<=", ">=", "=>", "++", "--", "&&", "||", "??", "+=", "-=", "*=", "/=",
				"%=", "&=", "|=", "^=", "**", "?.", "<<", ">>":
				end = index + 2
			}
		}
		tokens = append(
			tokens,
			token{kind: "punct", text: source[index:end], start: index, end: end, newlineBefore: newline},
		)
		index = end
	}
	return tokens
}

func allowsRegex(tokens []token) bool {
	if len(tokens) == 0 {
		return true
	}
	previous := tokens[len(tokens)-1]
	switch previous.kind {
	case "num", "str", "regexp", "tmpl":
		return false
	case "ident":
		switch previous.text {
		case "return", "throw", "case", "else", "do", "typeof", "void", "in", "of", "new",
			"delete", "yield", "await", "instanceof", "extends":
			return true
		default:
			return false
		}
	}
	switch previous.text {
	case ")", "]", "}", "++", "--":
		return false
	default:
		return true
	}
}

func endOfRegex(source string, index int) int {
	index++
	inClass := false
	for index < len(source) {
		character := source[index]
		if character == '\\' && index+1 < len(source) {
			index += 2
			continue
		}
		if character == '[' {
			inClass = true
		} else if character == ']' {
			inClass = false
		} else if character == '/' && !inClass {
			index++
			for index < len(source) && isIdentByte(source[index]) {
				index++
			}
			return index
		} else if character == '\n' {
			return -1
		}
		index++
	}
	return -1
}

func endOfTemplate(source string, index int) int {
	index++
	for index < len(source) {
		if source[index] == '\\' && index+1 < len(source) {
			index += 2
			continue
		}
		if source[index] == '`' {
			return index + 1
		}
		if source[index] == '$' && index+1 < len(source) && source[index+1] == '{' {
			_, next, ok := expressionInside(source, index+2, 1)
			if !ok {
				return -1
			}
			index = next
			continue
		}
		index++
	}
	return -1
}

func endOfNumber(source string, index int) int {
	if source[index] == '0' && index+1 < len(source) {
		switch source[index+1] {
		case 'x', 'X', 'b', 'B', 'o', 'O':
			index += 2
			for index < len(source) && (isHex(source[index]) || source[index] == '_') {
				index++
			}
			return index
		}
	}
	sawDigit := false
	for index < len(source) && (isDigit(source[index]) || source[index] == '_') {
		if isDigit(source[index]) {
			sawDigit = true
		}
		index++
	}
	if index < len(source) && source[index] == '.' {
		index++
		for index < len(source) && (isDigit(source[index]) || source[index] == '_') {
			if isDigit(source[index]) {
				sawDigit = true
			}
			index++
		}
	}
	if !sawDigit {
		return index
	}
	if index < len(source) && (source[index] == 'e' || source[index] == 'E') {
		next := index + 1
		if next < len(source) && (source[next] == '+' || source[next] == '-') {
			next++
		}
		if next < len(source) && isDigit(source[next]) {
			index = next + 1
			for index < len(source) && isDigit(source[index]) {
				index++
			}
		}
	}
	if index < len(source) && (source[index] == 'n') {
		index++
	}
	return index
}

func endOfIdent(source string, index int) int {
	for index < len(source) {
		runeValue, size := utf8.DecodeRuneInString(source[index:])
		if runeValue < 0x80 && !isIdentifierPart(runeValue) {
			break
		}
		if runeValue >= 0x80 &&
			!unicode.IsLetter(runeValue) &&
			!unicode.IsDigit(runeValue) &&
			runeValue != '_' &&
			runeValue != '$' {
			break
		}
		index += size
	}
	return index
}

func isDigit(character byte) bool { return character >= '0' && character <= '9' }
func isHex(character byte) bool {
	return isDigit(character) || (character >= 'a' && character <= 'f') || (character >= 'A' && character <= 'F')
}
func isIdentByte(character byte) bool { return isIdentifierPart(rune(character)) }

// jtype is a type the scan proved. A nil type is unknown, and unknown is never annotated and
// never used to justify rewriting ==.
type jtype struct {
	kind   string
	elem   *jtype
	fields []jfield
}

type jfield struct {
	name string
	typ  *jtype
}

func (kind *jtype) string() string {
	if kind == nil {
		return ""
	}
	switch kind.kind {
	case "number", "string", "boolean", "undefined", "null":
		return kind.kind
	case "array":
		inner := kind.elem.string()
		if inner == "" {
			return ""
		}
		if strings.Contains(inner, " ") {
			return "(" + inner + ")[]"
		}
		return inner + "[]"
	case "object":
		if len(kind.fields) == 0 {
			return ""
		}
		parts := make([]string, 0, len(kind.fields))
		for _, field := range kind.fields {
			rendered := field.typ.string()
			if rendered == "" || !identText(field.name) {
				return ""
			}
			parts = append(parts, field.name+": "+rendered)
		}
		return "{ " + strings.Join(parts, ", ") + " }"
	case "union":
		return ""
	default:
		return ""
	}
}

func identText(name string) bool {
	if name == "" {
		return false
	}
	first, _ := utf8.DecodeRuneInString(name)
	if !isIdentifierStart(first) {
		return false
	}
	for _, character := range name {
		if !isIdentifierPart(character) {
			return false
		}
	}
	return true
}

func sameType(left *jtype, right *jtype) bool {
	if left == nil || right == nil || left.kind != right.kind {
		return false
	}
	switch left.kind {
	case "number", "string", "boolean", "undefined", "null":
		return true
	case "array":
		return sameType(left.elem, right.elem)
	case "object":
		if len(left.fields) != len(right.fields) {
			return false
		}
		for index := range left.fields {
			if left.fields[index].name != right.fields[index].name ||
				!sameType(left.fields[index].typ, right.fields[index].typ) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// strictEqualOK reports whether == and === agree on every value of these two types.
// The same primitive does not coerce. Two objects compare by reference either way.
// null compared with undefined does coerce, and that pair is not the same kind, so it stays.
func strictEqualOK(left *jtype, right *jtype) bool {
	if left == nil || right == nil || left.kind != right.kind {
		return false
	}
	switch left.kind {
	case "number", "string", "boolean", "null", "undefined", "object", "array":
		return true
	default:
		return false
	}
}

var (
	typeNumber    = &jtype{kind: "number"}
	typeString    = &jtype{kind: "string"}
	typeBoolean   = &jtype{kind: "boolean"}
	typeUndefined = &jtype{kind: "undefined"}
	typeNull      = &jtype{kind: "null"}
)

func arrayOf(elem *jtype) *jtype {
	if elem == nil {
		return nil
	}
	return &jtype{kind: "array", elem: elem}
}

type texpr struct {
	op        string
	name      string
	prop      string
	scope     *scope
	fn        *fnInfo
	left      *texpr
	right     *texpr
	elems     []*texpr
	fields    []namedExpr
	arrayCtor bool
}

type namedExpr struct {
	name string
	expr *texpr
}

type binding struct {
	kind     string
	name     string
	fn       *fnInfo
	init     *texpr
	assigns  []assignNote
	typ      *jtype
	poison   bool
	visiting bool
}

type assignNote struct {
	rhs *texpr
	inc bool
}

type fnInfo struct {
	params []param
	scope  *scope
}

type param struct {
	name       string
	nameEnd    int
	typed      bool
	rest       bool
	pattern    bool
	hasDefault bool
}

type scope struct {
	kind   string
	parent *scope
	fn     *scope
	loop   *scope
	start  int
	end    int
	names  map[string]*binding
}

type reference struct {
	name   string
	scope  *scope
	at     int
	inInit *binding
}

type varDecl struct {
	name    string
	binding *binding
	nameAt  int
	block   *scope
	fn      *scope
	loop    *scope
	loopVar bool
	stmt    *varStmt
	unsafe  bool
}

type varStmt struct {
	keyword token
	decls   []*varDecl
}

type callNote struct {
	method  string
	recv    *texpr
	args    []*texpr
	viaCall bool
}

type cmpNote struct {
	op    string
	start int
	end   int
	left  *texpr
	right *texpr
}

type throwNote struct {
	start int
	end   int
}

type lateAssign struct {
	name  string
	scope *scope
	note  assignNote
}

type parser struct {
	toks              []token
	pos               int
	scope             *scope
	failed            bool
	unsafeVars        bool
	edits             []edit
	counts            map[string]int
	refs              []reference
	varStmts          []*varStmt
	calls             []callNote
	cmps              []cmpNote
	throws            []throwNote
	lateAssigns       []lateAssign
	initializing      *binding
	instanceOfTest262 bool
}

func parseBody(source string) *parser {
	parsed := &parser{toks: tokenize(source), counts: map[string]int{}}
	if parsed.toks == nil && strings.TrimSpace(source) != "" {
		parsed.failed = true
		return parsed
	}
	for index := 0; index+1 < len(parsed.toks); index++ {
		if parsed.toks[index].text == "instanceof" && parsed.toks[index+1].text == "Test262Error" {
			parsed.instanceOfTest262 = true
		}
	}
	parsed.scope = &scope{kind: "module", names: map[string]*binding{}, start: 0}
	parsed.scope.fn = parsed.scope
	parsed.parseStatements(true)
	if parsed.scope != nil && parsed.scope.kind == "module" {
		parsed.scope.end = len(parsed.toks)
	}
	if !parsed.failed && parsed.pos < len(parsed.toks) {
		parsed.failed = true
	}
	return parsed
}

func (parser *parser) peek() token {
	if parser.pos >= len(parser.toks) {
		return token{kind: "eof", text: "eof"}
	}
	return parser.toks[parser.pos]
}

func (parser *parser) eof() bool { return parser.pos >= len(parser.toks) }

func (parser *parser) next() token {
	tok := parser.peek()
	if parser.pos < len(parser.toks) {
		parser.pos++
	}
	return tok
}

func (parser *parser) is(text string) bool { return parser.peek().text == text }

func (parser *parser) fail() { parser.failed = true }

func (parser *parser) expect(text string) {
	if parser.failed {
		return
	}
	if !parser.is(text) {
		parser.fail()
		return
	}
	parser.next()
}

func (parser *parser) parseStatements(top bool) {
	for !parser.failed && !parser.eof() {
		if !top && parser.is("}") {
			return
		}
		before := parser.pos
		parser.parseStatement()
		if parser.pos == before {
			parser.fail()
			return
		}
	}
}

func (parser *parser) parseStatement() {
	if parser.failed || parser.eof() {
		return
	}
	switch parser.peek().text {
	case ";":
		parser.next()
		return
	case "{":
		parser.parseBlock()
		return
	case "var", "let", "const":
		parser.parseDeclaration(false)
		parser.semi()
		return
	case "function":
		parser.parseFunction(true)
		return
	case "class":
		parser.skipBraced()
		parser.unsafeVars = true
		return
	case "if":
		parser.next()
		parser.expect("(")
		parser.parseSequence(false)
		parser.expect(")")
		parser.parseControlled()
		if parser.is("else") {
			parser.next()
			parser.parseControlled()
		}
		return
	case "for":
		parser.parseFor()
		return
	case "while":
		parser.next()
		loop := parser.push("loop")
		parser.expect("(")
		parser.parseSequence(false)
		parser.expect(")")
		parser.parseControlled()
		parser.pop(loop)
		return
	case "do":
		parser.next()
		loop := parser.push("loop")
		parser.parseControlled()
		parser.expect("while")
		parser.expect("(")
		parser.parseSequence(false)
		parser.expect(")")
		parser.semi()
		parser.pop(loop)
		return
	case "switch":
		parser.parseSwitch()
		return
	case "try":
		parser.parseTry()
		return
	case "throw":
		parser.parseThrow()
		return
	case "return":
		parser.next()
		if !parser.eof() && !parser.is("}") && !parser.is(";") && !parser.peek().newlineBefore {
			parser.parseSequence(false)
		}
		parser.semi()
		return
	case "break", "continue", "debugger":
		parser.next()
		if parser.peek().kind == "ident" &&
			!isKeyword(parser.peek().text) &&
			!parser.peek().newlineBefore &&
			!parser.is(";") &&
			!parser.is("}") {
			parser.next()
		}
		parser.semi()
		return
	case "import", "export", "with":
		parser.fail()
		return
	}
	if parser.peek().kind == "ident" &&
		!isKeyword(parser.peek().text) &&
		parser.pos+1 < len(parser.toks) &&
		parser.toks[parser.pos+1].text == ":" {
		parser.next()
		parser.next()
		parser.parseStatement()
		return
	}
	if parser.peek().text == "async" &&
		parser.pos+1 < len(parser.toks) &&
		parser.toks[parser.pos+1].text == "function" {
		parser.fail()
		return
	}
	parser.parseSequence(false)
	parser.semi()
}

func (parser *parser) parseControlled() {
	block := parser.push("block")
	parser.parseStatement()
	parser.pop(block)
}

func (parser *parser) parseBlock() {
	parser.expect("{")
	block := parser.push("block")
	parser.parseStatements(false)
	parser.pop(block)
	parser.expect("}")
}

func (parser *parser) push(kind string) *scope {
	child := &scope{
		kind:   kind,
		parent: parser.scope,
		names:  map[string]*binding{},
		start:  parser.pos,
		fn:     parser.scope.fn,
		loop:   parser.scope.loop,
	}
	if kind == "func" {
		child.fn = child
		child.loop = nil
	}
	if kind == "loop" {
		child.loop = child
	}
	parser.scope = child
	return child
}

func (parser *parser) pop(child *scope) {
	child.end = parser.pos
	if parser.scope == child {
		parser.scope = child.parent
	}
}

func (parser *parser) semi() {
	if parser.failed {
		return
	}
	if parser.is(";") {
		parser.next()
		return
	}
	if parser.eof() || parser.is("}") || parser.peek().newlineBefore {
		return
	}
	parser.fail()
}

func (parser *parser) parseDeclaration(forHeader bool) {
	keyword := parser.next()
	stmt := &varStmt{keyword: keyword}
	for !parser.failed {
		if parser.peek().kind != "ident" || isKeyword(parser.peek().text) {
			parser.fail()
			return
		}
		nameTok := parser.next()
		binding := &binding{kind: keyword.text, name: nameTok.text}
		decl := &varDecl{
			name:    nameTok.text,
			binding: binding,
			nameAt:  parser.pos - 1,
			block:   parser.scope,
			fn:      parser.scope.fn,
			loop:    parser.scope.loop,
			loopVar: forHeader && keyword.text == "var",
			stmt:    stmt,
		}
		if keyword.text == "var" {
			existing := parser.scope.fn.names[nameTok.text]
			if existing == nil {
				parser.scope.fn.names[nameTok.text] = binding
			} else if existing.kind == "var" {
				existing.poison = true
				decl.binding = existing
				binding = existing
			} else {
				decl.unsafe = true
			}
		} else if _, exists := parser.scope.names[nameTok.text]; exists {
			binding.poison = true
			parser.scope.names[nameTok.text] = binding
		} else {
			parser.scope.names[nameTok.text] = binding
		}
		if parser.is("=") {
			parser.next()
			saved := parser.initializing
			parser.initializing = binding
			binding.init = parser.parseAssign(forHeader)
			parser.initializing = saved
			if binding.init != nil && binding.init.op == "fn" && binding.fn == nil {
				binding.fn = binding.init.fn
			}
		}
		stmt.decls = append(stmt.decls, decl)
		if parser.is(",") {
			parser.next()
			continue
		}
		break
	}
	if keyword.text == "var" {
		parser.varStmts = append(parser.varStmts, stmt)
	}
}

func (parser *parser) parseFor() {
	parser.next()
	loop := parser.push("loop")
	parser.expect("(")
	switch {
	case parser.is("var"), parser.is("let"), parser.is("const"):
		parser.parseDeclaration(true)
		if parser.is("in") || parser.is("of") {
			parser.next()
			parser.parseAssign(false)
		} else {
			parser.expect(";")
			if !parser.is(";") {
				parser.parseSequence(false)
			}
			parser.expect(";")
			if !parser.is(")") {
				parser.parseSequence(false)
			}
		}
	case parser.is(";"):
		parser.next()
		if !parser.is(";") {
			parser.parseSequence(false)
		}
		parser.expect(";")
		if !parser.is(")") {
			parser.parseSequence(false)
		}
	default:
		parser.parseSequence(true)
		if parser.is("in") || parser.is("of") {
			parser.next()
			parser.parseAssign(false)
		} else {
			parser.expect(";")
			if !parser.is(";") {
				parser.parseSequence(false)
			}
			parser.expect(";")
			if !parser.is(")") {
				parser.parseSequence(false)
			}
		}
	}
	parser.expect(")")
	parser.parseControlled()
	parser.pop(loop)
}

func (parser *parser) parseSwitch() {
	parser.next()
	parser.expect("(")
	parser.parseSequence(false)
	parser.expect(")")
	parser.expect("{")
	block := parser.push("switch")
	for !parser.failed && !parser.eof() && !parser.is("}") {
		if parser.is("case") {
			parser.next()
			parser.parseSequence(false)
			parser.expect(":")
			continue
		}
		if parser.is("default") {
			parser.next()
			parser.expect(":")
			continue
		}
		before := parser.pos
		parser.parseStatement()
		if parser.pos == before {
			parser.fail()
			return
		}
	}
	parser.pop(block)
	parser.expect("}")
}

func (parser *parser) parseTry() {
	parser.next()
	parser.parseBlock()
	if parser.is("catch") {
		parser.next()
		parser.expect("(")
		catchScope := parser.push("block")
		if parser.peek().kind == "ident" && !isKeyword(parser.peek().text) {
			name := parser.next()
			catchScope.names[name.text] = &binding{kind: "catch", name: name.text, poison: true}
			if parser.is("if") {
				parser.fail()
				return
			}
		} else {
			parser.skipBalanced(parser.peek().text, matchingClose(parser.peek().text))
		}
		parser.expect(")")
		parser.parseBlock()
		parser.pop(catchScope)
	}
	if parser.is("finally") {
		parser.next()
		parser.parseBlock()
	}
}

func (parser *parser) parseThrow() {
	parser.next()
	if parser.peek().newlineBefore {
		parser.fail()
		return
	}
	if parser.is("new") &&
		parser.pos+1 < len(parser.toks) &&
		parser.toks[parser.pos+1].text == "Test262Error" &&
		!parser.instanceOfTest262 {
		name := parser.toks[parser.pos+1]
		parser.throws = append(parser.throws, throwNote{start: name.start, end: name.end})
	}
	parser.parseSequence(false)
	parser.semi()
}

func (parser *parser) parseFunction(statement bool) *fnInfo {
	parser.next()
	var name string
	var nameTok token
	if parser.peek().kind == "ident" && !isKeyword(parser.peek().text) {
		nameTok = parser.next()
		name = nameTok.text
	} else if statement {
		parser.fail()
		return nil
	}
	fnScope := parser.push("func")
	info := &fnInfo{scope: fnScope}
	if statement && name != "" {
		binding := &binding{kind: "func", name: name, fn: info}
		parent := fnScope.parent
		if existing := parent.names[name]; existing != nil {
			existing.poison = true
			existing.fn = nil
		} else {
			parent.names[name] = binding
		}
	}
	parser.parseParamList(info, fnScope)
	if parser.is(":") {
		parser.next()
		parser.skipType()
	}
	parser.expect("{")
	if !statement && name != "" {
		// The expression's own name is visible inside it, and it is not the binding a call site sees.
		fnScope.names[name] = &binding{kind: "func", name: name, poison: true}
	}
	parser.parseStatements(false)
	parser.expect("}")
	parser.pop(fnScope)
	return info
}

func (parser *parser) parseParamList(info *fnInfo, fnScope *scope) {
	parser.expect("(")
	for !parser.failed && !parser.is(")") && !parser.eof() {
		param := param{}
		if parser.is("...") {
			parser.next()
			param.rest = true
		}
		if parser.is("{") || parser.is("[") {
			param.pattern = true
			parser.skipBalanced(parser.peek().text, matchingClose(parser.peek().text))
		} else if parser.peek().kind == "ident" && !isKeyword(parser.peek().text) {
			name := parser.next()
			param.name = name.text
			param.nameEnd = name.end
			if parser.is(":") {
				param.typed = true
				parser.next()
				parser.skipType()
			}
			if parser.is("=") {
				param.hasDefault = true
				parser.next()
				parser.parseAssign(false)
			}
			if param.name != "" {
				fnScope.names[param.name] = &binding{kind: "param", name: param.name}
			}
		} else {
			parser.fail()
			break
		}
		info.params = append(info.params, param)
		if parser.is(",") {
			parser.next()
			continue
		}
		break
	}
	parser.expect(")")
}

func (parser *parser) skipType() {
	depth := 0
	for !parser.failed && !parser.eof() {
		text := parser.peek().text
		if depth == 0 && (text == "," || text == ")" || text == "=" || text == "{" || text == ";" || text == "=>") {
			return
		}
		if text == "(" || text == "{" || text == "[" || text == "<" {
			depth++
		} else if text == ")" || text == "}" || text == "]" || text == ">" {
			depth--
			if depth < 0 {
				return
			}
		}
		parser.next()
	}
}

func matchingClose(open string) string {
	switch open {
	case "{":
		return "}"
	case "[":
		return "]"
	case "(":
		return ")"
	default:
		return ""
	}
}

func (parser *parser) skipBalanced(open string, close string) {
	if close == "" || !parser.is(open) {
		parser.fail()
		return
	}
	depth := 0
	for !parser.eof() {
		text := parser.next().text
		if text == open {
			depth++
		} else if text == close {
			depth--
			if depth == 0 {
				return
			}
		}
	}
	parser.fail()
}

func (parser *parser) skipBraced() {
	for !parser.failed && !parser.eof() && !parser.is("{") {
		switch parser.peek().text {
		case "(", "[":
			open := parser.peek().text
			parser.skipBalanced(open, matchingClose(open))
		default:
			parser.next()
		}
	}
	parser.skipBalanced("{", "}")
}

func (parser *parser) parseSequence(noIn bool) *texpr {
	expr := parser.parseAssign(noIn)
	for parser.is(",") && !parser.failed {
		parser.next()
		expr = parser.parseAssign(noIn)
	}
	return expr
}

func (parser *parser) parseAssign(noIn bool) *texpr {
	left := parser.parseConditional(noIn)
	if parser.failed || left == nil {
		return left
	}
	op := parser.peek().text
	if op == "=" || op == "+=" || op == "-=" || op == "*=" || op == "/=" || op == "%=" {
		parser.next()
		right := parser.parseAssign(noIn)
		if left.op == "name" {
			note := assignNote{rhs: right}
			if op != "=" {
				note.rhs = &texpr{op: "unknown"}
				note.inc = false
			}
			if binding := lookup(left.scope, left.name); binding != nil {
				binding.assigns = append(binding.assigns, note)
			} else {
				parser.lateAssigns = append(
					parser.lateAssigns,
					lateAssign{name: left.name, scope: left.scope, note: note},
				)
			}
		}
		if op == "=" {
			return right
		}
		return &texpr{op: "unknown"}
	}
	return left
}

func (parser *parser) parseConditional(noIn bool) *texpr {
	expr := parser.parseBinary(noIn, 1)
	if !parser.is("?") {
		return expr
	}
	parser.next()
	yes := parser.parseAssign(false)
	parser.expect(":")
	no := parser.parseAssign(noIn)
	return &texpr{op: "ternary", left: yes, right: no, elems: []*texpr{expr}}
}

func (parser *parser) parseBinary(noIn bool, min int) *texpr {
	left := parser.parseUnary()
	for !parser.failed {
		op := parser.peek()
		prec := precedence(op.text, noIn)
		if prec < min {
			break
		}
		parser.next()
		right := parser.parseBinary(noIn, prec+1)
		if (op.text == "==" || op.text == "!=") && left != nil && right != nil {
			parser.cmps = append(
				parser.cmps,
				cmpNote{op: op.text, start: op.start, end: op.end, left: left, right: right},
			)
		}
		kind := "unknown"
		switch op.text {
		case "==", "!=", "===", "!==", "<", ">", "<=", ">=", "instanceof", "in":
			kind = "bool"
		case "+":
			kind = "add"
		case "-":
			kind = "sub"
		case "*", "/", "%":
			kind = "mul"
		case "&&", "||", "??":
			kind = "logic"
		}
		left = &texpr{op: kind, left: left, right: right}
	}
	return left
}

func precedence(op string, noIn bool) int {
	switch op {
	case "??":
		return 1
	case "||":
		return 2
	case "&&":
		return 3
	case "==", "!=", "===", "!==":
		return 4
	case "<", ">", "<=", ">=", "instanceof":
		return 5
	case "in":
		if noIn {
			return 0
		}
		return 5
	case "<<", ">>", ">>>":
		return 6
	case "+", "-":
		return 7
	case "*", "/", "%":
		return 8
	default:
		return 0
	}
}

func (parser *parser) parseUnary() *texpr {
	switch parser.peek().text {
	case "++", "--":
		parser.next()
		operand := parser.parseUnary()
		parser.noteInc(operand)
		return &texpr{op: "inc", left: operand}
	case "+", "-", "!", "~", "typeof", "void", "delete":
		op := parser.next().text
		operand := parser.parseUnary()
		switch op {
		case "typeof":
			return &texpr{op: "typeof", left: operand}
		case "void":
			return &texpr{op: "void"}
		case "!":
			return &texpr{op: "bool"}
		case "+":
			return &texpr{op: "num"}
		case "-":
			if operand != nil && operand.op == "num" {
				return operand
			}
			return &texpr{op: "num", left: operand}
		default:
			return &texpr{op: "unknown"}
		}
	case "await", "yield":
		parser.fail()
		return nil
	}
	return parser.parsePostfix()
}

func (parser *parser) noteInc(operand *texpr) {
	if operand == nil || operand.op != "name" {
		return
	}
	note := assignNote{inc: true}
	if binding := lookup(operand.scope, operand.name); binding != nil {
		binding.assigns = append(binding.assigns, note)
	} else {
		parser.lateAssigns = append(
			parser.lateAssigns,
			lateAssign{name: operand.name, scope: operand.scope, note: note},
		)
	}
}

func (parser *parser) parsePostfix() *texpr {
	expr := parser.parsePrimary()
	for !parser.failed && expr != nil {
		switch {
		case parser.is(".") || parser.is("?."):
			optional := parser.next().text == "?."
			if optional && parser.is("(") {
				parser.parseCall(&texpr{op: "unknown"})
				expr = &texpr{op: "unknown"}
				continue
			}
			name := parser.peek()
			if name.kind != "ident" {
				parser.fail()
				return expr
			}
			parser.next()
			expr = &texpr{op: "member", prop: name.text, left: expr}
		case parser.is("("):
			if expr.op == "name" && expr.name == "Array" {
				call := parser.finishCall(expr)
				expr = &texpr{op: "array", elems: call, arrayCtor: true}
				continue
			}
			expr = parser.parseCall(expr)
		case parser.is("["):
			parser.next()
			parser.parseSequence(false)
			parser.expect("]")
			expr = &texpr{op: "index", left: expr}
		case (parser.is("++") || parser.is("--")) && !parser.peek().newlineBefore:
			parser.noteInc(expr)
			parser.next()
			expr = &texpr{op: "inc", left: expr}
		default:
			return expr
		}
	}
	return expr
}

func (parser *parser) parseCall(callee *texpr) *texpr {
	args := parser.finishCall(callee)
	parser.noteCall(callee, args)
	return &texpr{op: "call", left: callee, elems: args}
}

func (parser *parser) finishCall(callee *texpr) []*texpr {
	parser.expect("(")
	var args []*texpr
	for !parser.failed && !parser.is(")") && !parser.eof() {
		if parser.is("...") {
			parser.next()
			parser.parseAssign(false)
			args = append(args, &texpr{op: "unknown"})
		} else {
			args = append(args, parser.parseAssign(false))
		}
		if parser.is(",") {
			parser.next()
			continue
		}
		break
	}
	parser.expect(")")
	_ = callee
	return args
}

func (parser *parser) noteCall(callee *texpr, args []*texpr) {
	if callee == nil || callee.op != "member" {
		return
	}
	if callee.prop == "call" && callee.left != nil && callee.left.op == "member" && callbackMethod(callee.left.prop) {
		parser.calls = append(
			parser.calls,
			callNote{method: callee.left.prop, recv: callee.left.left, args: args, viaCall: true},
		)
		return
	}
	if callbackMethod(callee.prop) {
		parser.calls = append(parser.calls, callNote{method: callee.prop, recv: callee.left, args: args})
	}
}

func callbackMethod(name string) bool {
	switch name {
	case "forEach", "map", "filter", "every", "some", "find", "findIndex", "findLast", "findLastIndex",
		"flatMap", "sort", "reduce", "reduceRight", "replace", "replaceAll":
		return true
	default:
		return false
	}
}

func (parser *parser) parsePrimary() *texpr {
	tok := parser.peek()
	switch tok.kind {
	case "num":
		parser.next()
		if strings.HasSuffix(tok.text, "n") {
			return &texpr{op: "unknown"}
		}
		return &texpr{op: "num"}
	case "str":
		parser.next()
		return &texpr{op: "str"}
	case "regexp":
		parser.next()
		return &texpr{op: "unknown"}
	case "tmpl":
		parser.next()
		if strings.Contains(tok.text, "${") {
			// A template can close over a var. Missing that reference would make var to let a lie,
			// so the whole file keeps its vars. The template text itself is not rewritten.
			parser.unsafeVars = true
		}
		return &texpr{op: "str"}
	}
	switch tok.text {
	case "[":
		return parser.parseArray()
	case "{":
		return parser.parseObject()
	case "function":
		return &texpr{op: "fn", fn: parser.parseFunction(false)}
	case "new":
		return parser.parseNew()
	case "(":
		if parser.looksArrow() {
			return parser.parseArrow()
		}
		parser.next()
		expr := parser.parseSequence(false)
		parser.expect(")")
		return expr
	case "ident":
		return parser.parseIdent()
	}
	if tok.kind == "ident" {
		return parser.parseIdent()
	}
	parser.fail()
	return nil
}

func (parser *parser) parseNew() *texpr {
	parser.next()
	expr := parser.parsePostfix()
	if expr != nil && expr.arrayCtor {
		return expr
	}
	if expr != nil && expr.op == "call" && expr.left != nil && expr.left.op == "name" && expr.left.name == "Array" {
		return &texpr{op: "array", elems: expr.elems, arrayCtor: true}
	}
	return &texpr{op: "unknown"}
}

func (parser *parser) parseIdent() *texpr {
	tok := parser.next()
	switch tok.text {
	case "true", "false":
		return &texpr{op: "bool"}
	case "null":
		return &texpr{op: "null"}
	case "undefined":
		if lookup(parser.scope, "undefined") == nil {
			return &texpr{op: "undefined"}
		}
	case "NaN", "Infinity":
		if lookup(parser.scope, tok.text) == nil {
			return &texpr{op: "num"}
		}
	}
	if isKeyword(tok.text) {
		parser.fail()
		return nil
	}
	if parser.is("=>") {
		fnScope := parser.push("func")
		fnScope.names[tok.text] = &binding{kind: "param", name: tok.text}
		return parser.finishArrow([]param{{name: tok.text, nameEnd: tok.end}}, fnScope)
	}
	resolved := lookup(parser.scope, tok.text)
	inInit := parser.initializing
	if resolved == nil || resolved != parser.initializing {
		inInit = nil
	}
	parser.refs = append(
		parser.refs,
		reference{name: tok.text, scope: parser.scope, at: parser.pos - 1, inInit: inInit},
	)
	return &texpr{op: "name", name: tok.text, scope: parser.scope}
}

func (parser *parser) parseArray() *texpr {
	parser.expect("[")
	var elems []*texpr
	hole := false
	for !parser.failed && !parser.is("]") && !parser.eof() {
		if parser.is(",") {
			hole = true
			parser.next()
			continue
		}
		if parser.is("...") {
			parser.next()
			parser.parseAssign(false)
			hole = true
		} else {
			elems = append(elems, parser.parseAssign(false))
		}
		if parser.is(",") {
			parser.next()
			continue
		}
		break
	}
	parser.expect("]")
	if hole {
		return &texpr{op: "unknown"}
	}
	return &texpr{op: "array", elems: elems}
}

func (parser *parser) parseObject() *texpr {
	parser.expect("{")
	var fields []namedExpr
	pure := true
	for !parser.failed && !parser.is("}") && !parser.eof() {
		if parser.is("...") {
			parser.next()
			parser.parseAssign(false)
			pure = false
		} else if parser.is("[") {
			parser.skipBalanced("[", "]")
			parser.parseMethodTail(true)
			pure = false
		} else if parser.peek().kind == "ident" || parser.peek().kind == "str" || parser.peek().kind == "num" {
			key := parser.next()
			name := key.text
			if key.kind == "str" && len(name) >= 2 {
				name = name[1 : len(name)-1]
			}
			if isModifier(key.text) &&
				(parser.peek().kind == "ident" || parser.peek().kind == "str" || parser.is("[")) {
				pure = false
				if parser.is("[") {
					parser.skipBalanced("[", "]")
				} else {
					parser.next()
				}
				parser.parseMethodTail(true)
			} else if parser.is("(") {
				parser.parseMethodTail(false)
				pure = false
			} else if parser.is(":") {
				parser.next()
				fields = append(fields, namedExpr{name: name, expr: parser.parseAssign(false)})
			} else if (parser.is(",") || parser.is("}")) && key.kind == "ident" && !isKeyword(key.text) {
				fields = append(
					fields,
					namedExpr{name: name, expr: &texpr{op: "name", name: name, scope: parser.scope}},
				)
				resolved := lookup(parser.scope, name)
				inInit := parser.initializing
				if resolved != parser.initializing {
					inInit = nil
				}
				parser.refs = append(
					parser.refs,
					reference{name: name, scope: parser.scope, at: parser.pos - 1, inInit: inInit},
				)
			} else {
				parser.fail()
				return nil
			}
		} else {
			parser.fail()
			return nil
		}
		if parser.is(",") {
			parser.next()
			continue
		}
		break
	}
	parser.expect("}")
	if !pure {
		return &texpr{op: "unknown"}
	}
	return &texpr{op: "object", fields: fields}
}

func isModifier(text string) bool {
	return text == "get" || text == "set" || text == "async" || text == "static"
}

func (parser *parser) parseMethodTail(alreadyNamed bool) {
	if !alreadyNamed && !parser.is("(") {
		parser.fail()
		return
	}
	if !parser.is("(") {
		parser.fail()
		return
	}
	fnScope := parser.push("func")
	info := &fnInfo{scope: fnScope}
	parser.parseParamList(info, fnScope)
	parser.expect("{")
	parser.parseStatements(false)
	parser.expect("}")
	parser.pop(fnScope)
}

func (parser *parser) looksArrow() bool {
	if !parser.is("(") {
		return false
	}
	depth := 0
	for index := parser.pos; index < len(parser.toks); index++ {
		text := parser.toks[index].text
		if text == "(" {
			depth++
		} else if text == ")" {
			depth--
			if depth == 0 {
				return index+1 < len(parser.toks) && parser.toks[index+1].text == "=>"
			}
		}
	}
	return false
}

func (parser *parser) parseArrow() *texpr {
	parser.next()
	fnScope := parser.push("func")
	var params []param
	for !parser.failed && !parser.is(")") && !parser.eof() {
		param := param{}
		if parser.is("...") {
			parser.next()
			param.rest = true
		}
		if parser.is("[") || parser.is("{") {
			param.pattern = true
			open := parser.peek().text
			parser.skipBalanced(open, matchingClose(open))
		} else if parser.peek().kind == "ident" && !isKeyword(parser.peek().text) {
			name := parser.next()
			param.name = name.text
			param.nameEnd = name.end
			if parser.is(":") {
				param.typed = true
				parser.next()
				parser.skipType()
			}
			if parser.is("=") {
				param.hasDefault = true
				parser.next()
				parser.parseAssign(false)
			}
			fnScope.names[param.name] = &binding{kind: "param", name: param.name}
		} else {
			parser.fail()
			break
		}
		params = append(params, param)
		if parser.is(",") {
			parser.next()
			continue
		}
		break
	}
	parser.expect(")")
	return parser.finishArrow(params, fnScope)
}

func (parser *parser) finishArrow(params []param, fnScope *scope) *texpr {
	info := &fnInfo{params: params, scope: fnScope}
	parser.expect("=>")
	if parser.is("{") {
		parser.next()
		parser.parseStatements(false)
		parser.expect("}")
	} else {
		parser.parseAssign(false)
	}
	parser.pop(fnScope)
	return &texpr{op: "fn", fn: info}
}

func isKeyword(text string) bool {
	switch text {
	case "break", "case", "catch", "class", "const", "continue", "debugger", "default", "delete",
		"do", "else", "export", "extends", "finally", "for", "function", "if", "import", "in",
		"instanceof", "let", "new", "return", "super", "switch", "this", "throw", "try", "typeof",
		"var", "void", "while", "with", "yield", "await", "enum", "true", "false", "null", "static",
		"of":
		return true
	default:
		return false
	}
}

func lookup(scope *scope, name string) *binding {
	for scope != nil {
		if binding := scope.names[name]; binding != nil {
			return binding
		}
		scope = scope.parent
	}
	return nil
}

func (parser *parser) finish() {
	if parser.failed {
		return
	}
	for _, late := range parser.lateAssigns {
		if binding := lookup(late.scope, late.name); binding != nil {
			binding.assigns = append(binding.assigns, late.note)
		}
	}
	if !parser.unsafeVars {
		parser.rewriteVars()
	}
	parser.annotateCallbacks()
	parser.rewriteEquals()
	parser.rewriteThrows()
}

func (parser *parser) rewriteVars() {
	for _, stmt := range parser.varStmts {
		safe := len(stmt.decls) > 0
		for _, decl := range stmt.decls {
			if !parser.declSafe(decl) {
				safe = false
				break
			}
		}
		if !safe {
			continue
		}
		parser.edits = append(parser.edits, edit{start: stmt.keyword.start, end: stmt.keyword.end, text: "let"})
		parser.counts[adaptVarToLet] += len(stmt.decls)
	}
}

func (parser *parser) declSafe(decl *varDecl) bool {
	if decl.unsafe || decl.binding == nil || decl.binding.poison {
		return false
	}
	if decl.block != nil && decl.block.kind == "switch" {
		// `let` directly in a switch case is a syntax error. A brace inside the case is its own block.
		return false
	}
	decls := 0
	for _, stmt := range parser.varStmts {
		for _, other := range stmt.decls {
			if other.binding == decl.binding {
				decls++
			}
		}
	}
	if decls != 1 {
		return false
	}
	for _, ref := range parser.refs {
		binding := lookup(ref.scope, ref.name)
		if binding != decl.binding {
			continue
		}
		if ref.inInit == decl.binding {
			return false
		}
		// A use whose token sits before the declarator is the var/let split: var reads undefined,
		// let is in the temporal dead zone. That includes a closure written above the declarator.
		if ref.at >= 0 && ref.at < decl.nameAt {
			return false
		}
		// let is scoped to the block that holds the declarator. A read from outside that block,
		// including from a function that is not inside it, would stop seeing the binding.
		if decl.block != nil && !scopeWithin(ref.scope, decl.block) {
			return false
		}
		if parser.capturedAcrossLoop(decl, ref) {
			return false
		}
	}
	return true
}

func (parser *parser) capturedAcrossLoop(decl *varDecl, ref reference) bool {
	if ref.scope == nil || ref.scope.fn == decl.fn {
		return false
	}
	for scope := ref.scope; scope != nil && scope != decl.fn; scope = scope.parent {
		if scope.kind != "loop" {
			continue
		}
		if decl.loopVar && decl.loop == scope {
			return true
		}
		if decl.block != scope && scopeWithin(decl.block, scope) {
			return true
		}
	}
	return false
}

func scopeWithin(inner *scope, outer *scope) bool {
	for inner != nil {
		if inner == outer {
			return true
		}
		inner = inner.parent
	}
	return false
}

func (parser *parser) annotateCallbacks() {
	groups := map[*fnInfo][][]*jtype{}
	for _, call := range parser.calls {
		fn := parser.callbackFn(call)
		if fn == nil {
			continue
		}
		groups[fn] = append(groups[fn], parser.inferCall(call))
	}
	for fn, inferences := range groups {
		if len(inferences) == 0 {
			continue
		}
		for index, param := range fn.params {
			if param.name == "" || param.typed || param.rest || param.pattern || param.hasDefault {
				continue
			}
			var chosen *jtype
			agree := true
			for _, inference := range inferences {
				if inference == nil || index >= len(inference) || inference[index] == nil {
					agree = false
					break
				}
				if chosen == nil {
					chosen = inference[index]
					continue
				}
				if !sameType(chosen, inference[index]) {
					agree = false
					break
				}
			}
			if !agree || chosen == nil {
				continue
			}
			rendered := chosen.string()
			if rendered == "" {
				continue
			}
			parser.edits = append(parser.edits, edit{start: param.nameEnd, end: param.nameEnd, text: ": " + rendered})
			parser.counts[adaptCallbackParam]++
			if fn.scope != nil {
				if binding := fn.scope.names[param.name]; binding != nil && binding.kind == "param" {
					binding.typ = chosen
				}
			}
		}
	}
}

func (parser *parser) callbackFn(call callNote) *fnInfo {
	index := 0
	if call.method == "replace" || call.method == "replaceAll" {
		index = 1
	}
	if call.viaCall {
		index++
	}
	if index >= len(call.args) || call.args[index] == nil || call.args[index].op != "name" {
		return nil
	}
	arg := call.args[index]
	binding := lookup(arg.scope, arg.name)
	if binding == nil || binding.fn == nil || binding.poison {
		return nil
	}
	return binding.fn
}

func (parser *parser) inferCall(call callNote) []*jtype {
	if call.method == "replace" || call.method == "replaceAll" {
		return []*jtype{typeString}
	}
	elem := parser.receiverElem(call)
	if elem == nil {
		return nil
	}
	numberAndArray := []*jtype{typeNumber, arrayOf(elem)}
	switch call.method {
	case "sort":
		return []*jtype{elem, elem}
	case "reduce", "reduceRight":
		initIndex := 1
		if call.viaCall {
			initIndex = 2
		}
		if len(call.args) > initIndex {
			initType := parser.eval(call.args[initIndex])
			if initType == nil {
				return nil
			}
			return append([]*jtype{initType, elem}, numberAndArray...)
		}
		return append([]*jtype{elem, elem}, numberAndArray...)
	default:
		return append([]*jtype{elem}, numberAndArray...)
	}
}

func (parser *parser) receiverElem(call callNote) *jtype {
	var recv *texpr
	if call.viaCall {
		if len(call.args) == 0 {
			return nil
		}
		recv = call.args[0]
	} else {
		recv = call.recv
	}
	typ := parser.eval(recv)
	if typ == nil || typ.kind != "array" || typ.elem == nil {
		return nil
	}
	return typ.elem
}

func (parser *parser) rewriteEquals() {
	for _, cmp := range parser.cmps {
		if !strictEqualOK(parser.eval(cmp.left), parser.eval(cmp.right)) {
			continue
		}
		text := "==="
		if cmp.op == "!=" {
			text = "!=="
		}
		parser.edits = append(parser.edits, edit{start: cmp.start, end: cmp.end, text: text})
		parser.counts[adaptStrictEq]++
	}
}

func (parser *parser) rewriteThrows() {
	for _, throw := range parser.throws {
		parser.edits = append(parser.edits, edit{start: throw.start, end: throw.end, text: "Error"})
		parser.counts[adaptThrowError]++
	}
}

func (parser *parser) eval(expr *texpr) *jtype {
	if expr == nil {
		return nil
	}
	switch expr.op {
	case "num":
		return typeNumber
	case "str":
		return typeString
	case "bool":
		return typeBoolean
	case "null":
		return typeNull
	case "undefined", "void":
		return typeUndefined
	case "typeof":
		return typeString
	case "name":
		binding := lookup(expr.scope, expr.name)
		if binding == nil {
			if expr.name == "undefined" {
				return typeUndefined
			}
			return nil
		}
		return parser.bindingType(binding)
	case "array":
		return parser.evalArray(expr)
	case "object":
		return parser.evalObject(expr)
	case "member":
		if expr.prop == "length" {
			recv := parser.eval(expr.left)
			if recv != nil && (recv.kind == "array" || recv.kind == "string") {
				return typeNumber
			}
		}
		return nil
	case "add":
		return addType(parser.eval(expr.left), parser.eval(expr.right))
	case "sub", "mul":
		left := parser.eval(expr.left)
		right := parser.eval(expr.right)
		if left != nil && right != nil && left.kind == "number" && right.kind == "number" {
			return typeNumber
		}
		return nil
	case "inc":
		inner := parser.eval(expr.left)
		if inner != nil && inner.kind == "number" {
			return typeNumber
		}
		return nil
	case "ternary":
		left := parser.eval(expr.left)
		right := parser.eval(expr.right)
		if sameType(left, right) {
			return left
		}
		return nil
	case "logic":
		return nil
	default:
		return nil
	}
}

func (parser *parser) bindingType(binding *binding) *jtype {
	if binding == nil || binding.poison {
		return nil
	}
	if binding.kind == "param" || binding.kind == "func" || binding.kind == "catch" {
		return binding.typ
	}
	if binding.typ != nil || binding.visiting {
		return binding.typ
	}
	binding.visiting = true
	var typ *jtype
	if binding.init != nil {
		typ = parser.eval(binding.init)
	}
	for _, assign := range binding.assigns {
		if assign.inc {
			if typ == nil || typ.kind != "number" {
				typ = nil
				binding.poison = true
				break
			}
			continue
		}
		rhs := parser.eval(assign.rhs)
		if typ == nil {
			typ = rhs
			continue
		}
		if rhs == nil || !sameType(typ, rhs) {
			typ = nil
			binding.poison = true
			break
		}
	}
	binding.visiting = false
	if binding.poison {
		return nil
	}
	binding.typ = typ
	return typ
}

func (parser *parser) evalArray(expr *texpr) *jtype {
	if expr.arrayCtor && len(expr.elems) == 1 {
		only := parser.eval(expr.elems[0])
		if only != nil && only.kind == "number" {
			return nil
		}
	}
	if len(expr.elems) == 0 {
		return nil
	}
	var elem *jtype
	for _, item := range expr.elems {
		next := parser.eval(item)
		if next == nil {
			return nil
		}
		if elem == nil {
			elem = next
			continue
		}
		if sameType(elem, next) {
			continue
		}
		if elem.kind == "object" || next.kind == "object" || elem.kind == "array" || next.kind == "array" {
			return nil
		}
		merged := mergeUnion(elem, next)
		if merged == nil {
			return nil
		}
		elem = merged
	}
	if elem != nil && elem.kind == "union" {
		return nil
	}
	return arrayOf(elem)
}

func mergeUnion(left *jtype, right *jtype) *jtype {
	// Mixed primitive arrays are a real lowering gap (an array of a union). Leaving the element
	// unknown means the callback stays unannotated instead of being labeled with a type stage 0
	// cannot lower, which would just move the refusal.
	_ = left
	_ = right
	return nil
}

func (parser *parser) evalObject(expr *texpr) *jtype {
	if len(expr.fields) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var fields []jfield
	for _, field := range expr.fields {
		if !identText(field.name) || seen[field.name] {
			return nil
		}
		seen[field.name] = true
		typ := parser.eval(field.expr)
		if typ == nil || typ.string() == "" {
			return nil
		}
		fields = append(fields, jfield{name: field.name, typ: typ})
	}
	return &jtype{kind: "object", fields: fields}
}

func addType(left *jtype, right *jtype) *jtype {
	if left == nil || right == nil {
		return nil
	}
	if left.kind == "number" && right.kind == "number" {
		return typeNumber
	}
	if (left.kind == "string" || right.kind == "string") && additivePrimitive(left) && additivePrimitive(right) {
		return typeString
	}
	return nil
}

func additivePrimitive(kind *jtype) bool {
	switch kind.kind {
	case "number", "string", "boolean":
		return true
	default:
		return false
	}
}
