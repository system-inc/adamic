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
			tokens = append(tokens, token{kind: "str", text: source[index:end], start: index, end: end, newlineBefore: newline})
			index = end
			continue
		}
		if character == '`' {
			end := endOfTemplate(source, index)
			if end < 0 {
				return nil
			}
			tokens = append(tokens, token{kind: "tmpl", text: source[index:end], start: index, end: end, newlineBefore: newline})
			index = end
			continue
		}
		if character == '/' && allowsRegex(tokens) {
			end := endOfRegex(source, index)
			if end < 0 {
				return nil
			}
			tokens = append(tokens, token{kind: "regexp", text: source[index:end], start: index, end: end, newlineBefore: newline})
			index = end
			continue
		}
		if isDigit(character) || (character == '.' && index+1 < len(source) && isDigit(source[index+1])) {
			end := endOfNumber(source, index)
			tokens = append(tokens, token{kind: "num", text: source[index:end], start: index, end: end, newlineBefore: newline})
			index = end
			continue
		}
		runeValue, size := utf8.DecodeRuneInString(source[index:])
		if isIdentifierStart(runeValue) || (runeValue >= 0x80 && unicode.IsLetter(runeValue)) {
			end := endOfIdent(source, index)
			tokens = append(tokens, token{kind: "ident", text: source[index:end], start: index, end: end, newlineBefore: newline})
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
		tokens = append(tokens, token{kind: "punct", text: source[index:end], start: index, end: end, newlineBefore: newline})
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
		if runeValue >= 0x80 && !unicode.IsLetter(runeValue) && !unicode.IsDigit(runeValue) && runeValue != '_' && runeValue != '$' {
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
			if left.fields[index].name != right.fields[index].name || !sameType(left.fields[index].typ, right.fields[index].typ) {
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

func (p *parser) peek() token {
	if p.pos >= len(p.toks) {
		return token{kind: "eof", text: "eof"}
	}
	return p.toks[p.pos]
}

func (p *parser) eof() bool { return p.pos >= len(p.toks) }

func (p *parser) next() token {
	tok := p.peek()
	if p.pos < len(p.toks) {
		p.pos++
	}
	return tok
}

func (p *parser) is(text string) bool { return p.peek().text == text }

func (p *parser) fail() { p.failed = true }

func (p *parser) expect(text string) {
	if p.failed {
		return
	}
	if !p.is(text) {
		p.fail()
		return
	}
	p.next()
}

func (p *parser) parseStatements(top bool) {
	for !p.failed && !p.eof() {
		if !top && p.is("}") {
			return
		}
		before := p.pos
		p.parseStatement()
		if p.pos == before {
			p.fail()
			return
		}
	}
}

func (p *parser) parseStatement() {
	if p.failed || p.eof() {
		return
	}
	switch p.peek().text {
	case ";":
		p.next()
		return
	case "{":
		p.parseBlock()
		return
	case "var", "let", "const":
		p.parseDeclaration(false)
		p.semi()
		return
	case "function":
		p.parseFunction(true)
		return
	case "class":
		p.skipBraced()
		p.unsafeVars = true
		return
	case "if":
		p.next()
		p.expect("(")
		p.parseSequence(false)
		p.expect(")")
		p.parseControlled()
		if p.is("else") {
			p.next()
			p.parseControlled()
		}
		return
	case "for":
		p.parseFor()
		return
	case "while":
		p.next()
		loop := p.push("loop")
		p.expect("(")
		p.parseSequence(false)
		p.expect(")")
		p.parseControlled()
		p.pop(loop)
		return
	case "do":
		p.next()
		loop := p.push("loop")
		p.parseControlled()
		p.expect("while")
		p.expect("(")
		p.parseSequence(false)
		p.expect(")")
		p.semi()
		p.pop(loop)
		return
	case "switch":
		p.parseSwitch()
		return
	case "try":
		p.parseTry()
		return
	case "throw":
		p.parseThrow()
		return
	case "return":
		p.next()
		if !p.eof() && !p.is("}") && !p.is(";") && !p.peek().newlineBefore {
			p.parseSequence(false)
		}
		p.semi()
		return
	case "break", "continue", "debugger":
		p.next()
		if p.peek().kind == "ident" && !isKeyword(p.peek().text) && !p.peek().newlineBefore && !p.is(";") && !p.is("}") {
			p.next()
		}
		p.semi()
		return
	case "import", "export", "with":
		p.fail()
		return
	}
	if p.peek().kind == "ident" && !isKeyword(p.peek().text) && p.pos+1 < len(p.toks) && p.toks[p.pos+1].text == ":" {
		p.next()
		p.next()
		p.parseStatement()
		return
	}
	if p.peek().text == "async" && p.pos+1 < len(p.toks) && p.toks[p.pos+1].text == "function" {
		p.fail()
		return
	}
	p.parseSequence(false)
	p.semi()
}

func (p *parser) parseControlled() {
	block := p.push("block")
	p.parseStatement()
	p.pop(block)
}

func (p *parser) parseBlock() {
	p.expect("{")
	block := p.push("block")
	p.parseStatements(false)
	p.pop(block)
	p.expect("}")
}

func (p *parser) push(kind string) *scope {
	child := &scope{
		kind:   kind,
		parent: p.scope,
		names:  map[string]*binding{},
		start:  p.pos,
		fn:     p.scope.fn,
		loop:   p.scope.loop,
	}
	if kind == "func" {
		child.fn = child
		child.loop = nil
	}
	if kind == "loop" {
		child.loop = child
	}
	p.scope = child
	return child
}

func (p *parser) pop(child *scope) {
	child.end = p.pos
	if p.scope == child {
		p.scope = child.parent
	}
}

func (p *parser) semi() {
	if p.failed {
		return
	}
	if p.is(";") {
		p.next()
		return
	}
	if p.eof() || p.is("}") || p.peek().newlineBefore {
		return
	}
	p.fail()
}

func (p *parser) parseDeclaration(forHeader bool) {
	keyword := p.next()
	stmt := &varStmt{keyword: keyword}
	for !p.failed {
		if p.peek().kind != "ident" || isKeyword(p.peek().text) {
			p.fail()
			return
		}
		nameTok := p.next()
		binding := &binding{kind: keyword.text, name: nameTok.text}
		decl := &varDecl{
			name:    nameTok.text,
			binding: binding,
			nameAt:  p.pos - 1,
			block:   p.scope,
			fn:      p.scope.fn,
			loop:    p.scope.loop,
			loopVar: forHeader && keyword.text == "var",
			stmt:    stmt,
		}
		if keyword.text == "var" {
			existing := p.scope.fn.names[nameTok.text]
			if existing == nil {
				p.scope.fn.names[nameTok.text] = binding
			} else if existing.kind == "var" {
				existing.poison = true
				decl.binding = existing
				binding = existing
			} else {
				decl.unsafe = true
			}
		} else if _, exists := p.scope.names[nameTok.text]; exists {
			binding.poison = true
			p.scope.names[nameTok.text] = binding
		} else {
			p.scope.names[nameTok.text] = binding
		}
		if p.is("=") {
			p.next()
			saved := p.initializing
			p.initializing = binding
			binding.init = p.parseAssign(forHeader)
			p.initializing = saved
			if binding.init != nil && binding.init.op == "fn" && binding.fn == nil {
				binding.fn = binding.init.fn
			}
		}
		stmt.decls = append(stmt.decls, decl)
		if p.is(",") {
			p.next()
			continue
		}
		break
	}
	if keyword.text == "var" {
		p.varStmts = append(p.varStmts, stmt)
	}
}

func (p *parser) parseFor() {
	p.next()
	loop := p.push("loop")
	p.expect("(")
	switch {
	case p.is("var"), p.is("let"), p.is("const"):
		p.parseDeclaration(true)
		if p.is("in") || p.is("of") {
			p.next()
			p.parseAssign(false)
		} else {
			p.expect(";")
			if !p.is(";") {
				p.parseSequence(false)
			}
			p.expect(";")
			if !p.is(")") {
				p.parseSequence(false)
			}
		}
	case p.is(";"):
		p.next()
		if !p.is(";") {
			p.parseSequence(false)
		}
		p.expect(";")
		if !p.is(")") {
			p.parseSequence(false)
		}
	default:
		p.parseSequence(true)
		if p.is("in") || p.is("of") {
			p.next()
			p.parseAssign(false)
		} else {
			p.expect(";")
			if !p.is(";") {
				p.parseSequence(false)
			}
			p.expect(";")
			if !p.is(")") {
				p.parseSequence(false)
			}
		}
	}
	p.expect(")")
	p.parseControlled()
	p.pop(loop)
}

func (p *parser) parseSwitch() {
	p.next()
	p.expect("(")
	p.parseSequence(false)
	p.expect(")")
	p.expect("{")
	block := p.push("switch")
	for !p.failed && !p.eof() && !p.is("}") {
		if p.is("case") {
			p.next()
			p.parseSequence(false)
			p.expect(":")
			continue
		}
		if p.is("default") {
			p.next()
			p.expect(":")
			continue
		}
		before := p.pos
		p.parseStatement()
		if p.pos == before {
			p.fail()
			return
		}
	}
	p.pop(block)
	p.expect("}")
}

func (p *parser) parseTry() {
	p.next()
	p.parseBlock()
	if p.is("catch") {
		p.next()
		p.expect("(")
		catchScope := p.push("block")
		if p.peek().kind == "ident" && !isKeyword(p.peek().text) {
			name := p.next()
			catchScope.names[name.text] = &binding{kind: "catch", name: name.text, poison: true}
			if p.is("if") {
				p.fail()
				return
			}
		} else {
			p.skipBalanced(p.peek().text, matchingClose(p.peek().text))
		}
		p.expect(")")
		p.parseBlock()
		p.pop(catchScope)
	}
	if p.is("finally") {
		p.next()
		p.parseBlock()
	}
}

func (p *parser) parseThrow() {
	p.next()
	if p.peek().newlineBefore {
		p.fail()
		return
	}
	if p.is("new") && p.pos+1 < len(p.toks) && p.toks[p.pos+1].text == "Test262Error" && !p.instanceOfTest262 {
		name := p.toks[p.pos+1]
		p.throws = append(p.throws, throwNote{start: name.start, end: name.end})
	}
	p.parseSequence(false)
	p.semi()
}

func (p *parser) parseFunction(statement bool) *fnInfo {
	p.next()
	var name string
	var nameTok token
	if p.peek().kind == "ident" && !isKeyword(p.peek().text) {
		nameTok = p.next()
		name = nameTok.text
	} else if statement {
		p.fail()
		return nil
	}
	fnScope := p.push("func")
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
	p.parseParamList(info, fnScope)
	if p.is(":") {
		p.next()
		p.skipType()
	}
	p.expect("{")
	if !statement && name != "" {
		// The expression's own name is visible inside it, and it is not the binding a call site sees.
		fnScope.names[name] = &binding{kind: "func", name: name, poison: true}
	}
	p.parseStatements(false)
	p.expect("}")
	p.pop(fnScope)
	return info
}

func (p *parser) parseParamList(info *fnInfo, fnScope *scope) {
	p.expect("(")
	for !p.failed && !p.is(")") && !p.eof() {
		param := param{}
		if p.is("...") {
			p.next()
			param.rest = true
		}
		if p.is("{") || p.is("[") {
			param.pattern = true
			p.skipBalanced(p.peek().text, matchingClose(p.peek().text))
		} else if p.peek().kind == "ident" && !isKeyword(p.peek().text) {
			name := p.next()
			param.name = name.text
			param.nameEnd = name.end
			if p.is(":") {
				param.typed = true
				p.next()
				p.skipType()
			}
			if p.is("=") {
				param.hasDefault = true
				p.next()
				p.parseAssign(false)
			}
			if param.name != "" {
				fnScope.names[param.name] = &binding{kind: "param", name: param.name}
			}
		} else {
			p.fail()
			break
		}
		info.params = append(info.params, param)
		if p.is(",") {
			p.next()
			continue
		}
		break
	}
	p.expect(")")
}

func (p *parser) skipType() {
	depth := 0
	for !p.failed && !p.eof() {
		text := p.peek().text
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
		p.next()
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

func (p *parser) skipBalanced(open string, close string) {
	if close == "" || !p.is(open) {
		p.fail()
		return
	}
	depth := 0
	for !p.eof() {
		text := p.next().text
		if text == open {
			depth++
		} else if text == close {
			depth--
			if depth == 0 {
				return
			}
		}
	}
	p.fail()
}

func (p *parser) skipBraced() {
	for !p.failed && !p.eof() && !p.is("{") {
		switch p.peek().text {
		case "(", "[":
			open := p.peek().text
			p.skipBalanced(open, matchingClose(open))
		default:
			p.next()
		}
	}
	p.skipBalanced("{", "}")
}

func (p *parser) parseSequence(noIn bool) *texpr {
	expr := p.parseAssign(noIn)
	for p.is(",") && !p.failed {
		p.next()
		expr = p.parseAssign(noIn)
	}
	return expr
}

func (p *parser) parseAssign(noIn bool) *texpr {
	left := p.parseConditional(noIn)
	if p.failed || left == nil {
		return left
	}
	op := p.peek().text
	if op == "=" || op == "+=" || op == "-=" || op == "*=" || op == "/=" || op == "%=" {
		p.next()
		right := p.parseAssign(noIn)
		if left.op == "name" {
			note := assignNote{rhs: right}
			if op != "=" {
				note.rhs = &texpr{op: "unknown"}
				note.inc = false
			}
			if binding := lookup(left.scope, left.name); binding != nil {
				binding.assigns = append(binding.assigns, note)
			} else {
				p.lateAssigns = append(p.lateAssigns, lateAssign{name: left.name, scope: left.scope, note: note})
			}
		}
		if op == "=" {
			return right
		}
		return &texpr{op: "unknown"}
	}
	return left
}

func (p *parser) parseConditional(noIn bool) *texpr {
	expr := p.parseBinary(noIn, 1)
	if !p.is("?") {
		return expr
	}
	p.next()
	yes := p.parseAssign(false)
	p.expect(":")
	no := p.parseAssign(noIn)
	return &texpr{op: "ternary", left: yes, right: no, elems: []*texpr{expr}}
}

func (p *parser) parseBinary(noIn bool, min int) *texpr {
	left := p.parseUnary()
	for !p.failed {
		op := p.peek()
		prec := precedence(op.text, noIn)
		if prec < min {
			break
		}
		p.next()
		right := p.parseBinary(noIn, prec+1)
		if (op.text == "==" || op.text == "!=") && left != nil && right != nil {
			p.cmps = append(p.cmps, cmpNote{op: op.text, start: op.start, end: op.end, left: left, right: right})
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

func (p *parser) parseUnary() *texpr {
	switch p.peek().text {
	case "++", "--":
		p.next()
		operand := p.parseUnary()
		p.noteInc(operand)
		return &texpr{op: "inc", left: operand}
	case "+", "-", "!", "~", "typeof", "void", "delete":
		op := p.next().text
		operand := p.parseUnary()
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
		p.fail()
		return nil
	}
	return p.parsePostfix()
}

func (p *parser) noteInc(operand *texpr) {
	if operand == nil || operand.op != "name" {
		return
	}
	note := assignNote{inc: true}
	if binding := lookup(operand.scope, operand.name); binding != nil {
		binding.assigns = append(binding.assigns, note)
	} else {
		p.lateAssigns = append(p.lateAssigns, lateAssign{name: operand.name, scope: operand.scope, note: note})
	}
}

func (p *parser) parsePostfix() *texpr {
	expr := p.parsePrimary()
	for !p.failed && expr != nil {
		switch {
		case p.is(".") || p.is("?."):
			optional := p.next().text == "?."
			if optional && p.is("(") {
				p.parseCall(&texpr{op: "unknown"})
				expr = &texpr{op: "unknown"}
				continue
			}
			name := p.peek()
			if name.kind != "ident" {
				p.fail()
				return expr
			}
			p.next()
			expr = &texpr{op: "member", prop: name.text, left: expr}
		case p.is("("):
			if expr.op == "name" && expr.name == "Array" {
				call := p.finishCall(expr)
				expr = &texpr{op: "array", elems: call, arrayCtor: true}
				continue
			}
			expr = p.parseCall(expr)
		case p.is("["):
			p.next()
			p.parseSequence(false)
			p.expect("]")
			expr = &texpr{op: "index", left: expr}
		case (p.is("++") || p.is("--")) && !p.peek().newlineBefore:
			p.noteInc(expr)
			p.next()
			expr = &texpr{op: "inc", left: expr}
		default:
			return expr
		}
	}
	return expr
}

func (p *parser) parseCall(callee *texpr) *texpr {
	args := p.finishCall(callee)
	p.noteCall(callee, args)
	return &texpr{op: "call", left: callee, elems: args}
}

func (p *parser) finishCall(callee *texpr) []*texpr {
	p.expect("(")
	var args []*texpr
	for !p.failed && !p.is(")") && !p.eof() {
		if p.is("...") {
			p.next()
			p.parseAssign(false)
			args = append(args, &texpr{op: "unknown"})
		} else {
			args = append(args, p.parseAssign(false))
		}
		if p.is(",") {
			p.next()
			continue
		}
		break
	}
	p.expect(")")
	_ = callee
	return args
}

func (p *parser) noteCall(callee *texpr, args []*texpr) {
	if callee == nil || callee.op != "member" {
		return
	}
	if callee.prop == "call" && callee.left != nil && callee.left.op == "member" && callbackMethod(callee.left.prop) {
		p.calls = append(p.calls, callNote{method: callee.left.prop, recv: callee.left.left, args: args, viaCall: true})
		return
	}
	if callbackMethod(callee.prop) {
		p.calls = append(p.calls, callNote{method: callee.prop, recv: callee.left, args: args})
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

func (p *parser) parsePrimary() *texpr {
	tok := p.peek()
	switch tok.kind {
	case "num":
		p.next()
		if strings.HasSuffix(tok.text, "n") {
			return &texpr{op: "unknown"}
		}
		return &texpr{op: "num"}
	case "str":
		p.next()
		return &texpr{op: "str"}
	case "regexp":
		p.next()
		return &texpr{op: "unknown"}
	case "tmpl":
		p.next()
		if strings.Contains(tok.text, "${") {
			// A template can close over a var. Missing that reference would make var to let a lie,
			// so the whole file keeps its vars. The template text itself is not rewritten.
			p.unsafeVars = true
		}
		return &texpr{op: "str"}
	}
	switch tok.text {
	case "[":
		return p.parseArray()
	case "{":
		return p.parseObject()
	case "function":
		return &texpr{op: "fn", fn: p.parseFunction(false)}
	case "new":
		return p.parseNew()
	case "(":
		if p.looksArrow() {
			return p.parseArrow()
		}
		p.next()
		expr := p.parseSequence(false)
		p.expect(")")
		return expr
	case "ident":
		return p.parseIdent()
	}
	if tok.kind == "ident" {
		return p.parseIdent()
	}
	p.fail()
	return nil
}

func (p *parser) parseNew() *texpr {
	p.next()
	expr := p.parsePostfix()
	if expr != nil && expr.arrayCtor {
		return expr
	}
	if expr != nil && expr.op == "call" && expr.left != nil && expr.left.op == "name" && expr.left.name == "Array" {
		return &texpr{op: "array", elems: expr.elems, arrayCtor: true}
	}
	return &texpr{op: "unknown"}
}

func (p *parser) parseIdent() *texpr {
	tok := p.next()
	switch tok.text {
	case "true", "false":
		return &texpr{op: "bool"}
	case "null":
		return &texpr{op: "null"}
	case "undefined":
		if lookup(p.scope, "undefined") == nil {
			return &texpr{op: "undefined"}
		}
	case "NaN", "Infinity":
		if lookup(p.scope, tok.text) == nil {
			return &texpr{op: "num"}
		}
	}
	if isKeyword(tok.text) {
		p.fail()
		return nil
	}
	if p.is("=>") {
		fnScope := p.push("func")
		fnScope.names[tok.text] = &binding{kind: "param", name: tok.text}
		return p.finishArrow([]param{{name: tok.text, nameEnd: tok.end}}, fnScope)
	}
	resolved := lookup(p.scope, tok.text)
	inInit := p.initializing
	if resolved == nil || resolved != p.initializing {
		inInit = nil
	}
	p.refs = append(p.refs, reference{name: tok.text, scope: p.scope, at: p.pos - 1, inInit: inInit})
	return &texpr{op: "name", name: tok.text, scope: p.scope}
}

func (p *parser) parseArray() *texpr {
	p.expect("[")
	var elems []*texpr
	hole := false
	for !p.failed && !p.is("]") && !p.eof() {
		if p.is(",") {
			hole = true
			p.next()
			continue
		}
		if p.is("...") {
			p.next()
			p.parseAssign(false)
			hole = true
		} else {
			elems = append(elems, p.parseAssign(false))
		}
		if p.is(",") {
			p.next()
			continue
		}
		break
	}
	p.expect("]")
	if hole {
		return &texpr{op: "unknown"}
	}
	return &texpr{op: "array", elems: elems}
}

func (p *parser) parseObject() *texpr {
	p.expect("{")
	var fields []namedExpr
	pure := true
	for !p.failed && !p.is("}") && !p.eof() {
		if p.is("...") {
			p.next()
			p.parseAssign(false)
			pure = false
		} else if p.is("[") {
			p.skipBalanced("[", "]")
			p.parseMethodTail(true)
			pure = false
		} else if p.peek().kind == "ident" || p.peek().kind == "str" || p.peek().kind == "num" {
			key := p.next()
			name := key.text
			if key.kind == "str" && len(name) >= 2 {
				name = name[1 : len(name)-1]
			}
			if isModifier(key.text) && (p.peek().kind == "ident" || p.peek().kind == "str" || p.is("[")) {
				pure = false
				if p.is("[") {
					p.skipBalanced("[", "]")
				} else {
					p.next()
				}
				p.parseMethodTail(true)
			} else if p.is("(") {
				p.parseMethodTail(false)
				pure = false
			} else if p.is(":") {
				p.next()
				fields = append(fields, namedExpr{name: name, expr: p.parseAssign(false)})
			} else if (p.is(",") || p.is("}")) && key.kind == "ident" && !isKeyword(key.text) {
				fields = append(fields, namedExpr{name: name, expr: &texpr{op: "name", name: name, scope: p.scope}})
				resolved := lookup(p.scope, name)
				inInit := p.initializing
				if resolved != p.initializing {
					inInit = nil
				}
				p.refs = append(p.refs, reference{name: name, scope: p.scope, at: p.pos - 1, inInit: inInit})
			} else {
				p.fail()
				return nil
			}
		} else {
			p.fail()
			return nil
		}
		if p.is(",") {
			p.next()
			continue
		}
		break
	}
	p.expect("}")
	if !pure {
		return &texpr{op: "unknown"}
	}
	return &texpr{op: "object", fields: fields}
}

func isModifier(text string) bool {
	return text == "get" || text == "set" || text == "async" || text == "static"
}

func (p *parser) parseMethodTail(alreadyNamed bool) {
	if !alreadyNamed && !p.is("(") {
		p.fail()
		return
	}
	if !p.is("(") {
		p.fail()
		return
	}
	fnScope := p.push("func")
	info := &fnInfo{scope: fnScope}
	p.parseParamList(info, fnScope)
	p.expect("{")
	p.parseStatements(false)
	p.expect("}")
	p.pop(fnScope)
}

func (p *parser) looksArrow() bool {
	if !p.is("(") {
		return false
	}
	depth := 0
	for index := p.pos; index < len(p.toks); index++ {
		text := p.toks[index].text
		if text == "(" {
			depth++
		} else if text == ")" {
			depth--
			if depth == 0 {
				return index+1 < len(p.toks) && p.toks[index+1].text == "=>"
			}
		}
	}
	return false
}

func (p *parser) parseArrow() *texpr {
	p.next()
	fnScope := p.push("func")
	var params []param
	for !p.failed && !p.is(")") && !p.eof() {
		param := param{}
		if p.is("...") {
			p.next()
			param.rest = true
		}
		if p.is("[") || p.is("{") {
			param.pattern = true
			open := p.peek().text
			p.skipBalanced(open, matchingClose(open))
		} else if p.peek().kind == "ident" && !isKeyword(p.peek().text) {
			name := p.next()
			param.name = name.text
			param.nameEnd = name.end
			if p.is(":") {
				param.typed = true
				p.next()
				p.skipType()
			}
			if p.is("=") {
				param.hasDefault = true
				p.next()
				p.parseAssign(false)
			}
			fnScope.names[param.name] = &binding{kind: "param", name: param.name}
		} else {
			p.fail()
			break
		}
		params = append(params, param)
		if p.is(",") {
			p.next()
			continue
		}
		break
	}
	p.expect(")")
	return p.finishArrow(params, fnScope)
}

func (p *parser) finishArrow(params []param, fnScope *scope) *texpr {
	info := &fnInfo{params: params, scope: fnScope}
	p.expect("=>")
	if p.is("{") {
		p.next()
		p.parseStatements(false)
		p.expect("}")
	} else {
		p.parseAssign(false)
	}
	p.pop(fnScope)
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

func (p *parser) finish() {
	if p.failed {
		return
	}
	for _, late := range p.lateAssigns {
		if binding := lookup(late.scope, late.name); binding != nil {
			binding.assigns = append(binding.assigns, late.note)
		}
	}
	if !p.unsafeVars {
		p.rewriteVars()
	}
	p.annotateCallbacks()
	p.rewriteEquals()
	p.rewriteThrows()
}

func (p *parser) rewriteVars() {
	for _, stmt := range p.varStmts {
		safe := len(stmt.decls) > 0
		for _, decl := range stmt.decls {
			if !p.declSafe(decl) {
				safe = false
				break
			}
		}
		if !safe {
			continue
		}
		p.edits = append(p.edits, edit{start: stmt.keyword.start, end: stmt.keyword.end, text: "let"})
		p.counts[adaptVarToLet] += len(stmt.decls)
	}
}

func (p *parser) declSafe(decl *varDecl) bool {
	if decl.unsafe || decl.binding == nil || decl.binding.poison {
		return false
	}
	if decl.block != nil && decl.block.kind == "switch" {
		// `let` directly in a switch case is a syntax error. A brace inside the case is its own block.
		return false
	}
	decls := 0
	for _, stmt := range p.varStmts {
		for _, other := range stmt.decls {
			if other.binding == decl.binding {
				decls++
			}
		}
	}
	if decls != 1 {
		return false
	}
	for _, ref := range p.refs {
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
		if p.capturedAcrossLoop(decl, ref) {
			return false
		}
	}
	return true
}

func (p *parser) capturedAcrossLoop(decl *varDecl, ref reference) bool {
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

func (p *parser) annotateCallbacks() {
	groups := map[*fnInfo][][]*jtype{}
	for _, call := range p.calls {
		fn := p.callbackFn(call)
		if fn == nil {
			continue
		}
		groups[fn] = append(groups[fn], p.inferCall(call))
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
			p.edits = append(p.edits, edit{start: param.nameEnd, end: param.nameEnd, text: ": " + rendered})
			p.counts[adaptCallbackParam]++
			if fn.scope != nil {
				if binding := fn.scope.names[param.name]; binding != nil && binding.kind == "param" {
					binding.typ = chosen
				}
			}
		}
	}
}

func (p *parser) callbackFn(call callNote) *fnInfo {
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

func (p *parser) inferCall(call callNote) []*jtype {
	if call.method == "replace" || call.method == "replaceAll" {
		return []*jtype{typeString}
	}
	elem := p.receiverElem(call)
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
			initType := p.eval(call.args[initIndex])
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

func (p *parser) receiverElem(call callNote) *jtype {
	var recv *texpr
	if call.viaCall {
		if len(call.args) == 0 {
			return nil
		}
		recv = call.args[0]
	} else {
		recv = call.recv
	}
	typ := p.eval(recv)
	if typ == nil || typ.kind != "array" || typ.elem == nil {
		return nil
	}
	return typ.elem
}

func (p *parser) rewriteEquals() {
	for _, cmp := range p.cmps {
		if !strictEqualOK(p.eval(cmp.left), p.eval(cmp.right)) {
			continue
		}
		text := "==="
		if cmp.op == "!=" {
			text = "!=="
		}
		p.edits = append(p.edits, edit{start: cmp.start, end: cmp.end, text: text})
		p.counts[adaptStrictEq]++
	}
}

func (p *parser) rewriteThrows() {
	for _, throw := range p.throws {
		p.edits = append(p.edits, edit{start: throw.start, end: throw.end, text: "Error"})
		p.counts[adaptThrowError]++
	}
}

func (p *parser) eval(expr *texpr) *jtype {
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
		return p.bindingType(binding)
	case "array":
		return p.evalArray(expr)
	case "object":
		return p.evalObject(expr)
	case "member":
		if expr.prop == "length" {
			recv := p.eval(expr.left)
			if recv != nil && (recv.kind == "array" || recv.kind == "string") {
				return typeNumber
			}
		}
		return nil
	case "add":
		return addType(p.eval(expr.left), p.eval(expr.right))
	case "sub", "mul":
		left := p.eval(expr.left)
		right := p.eval(expr.right)
		if left != nil && right != nil && left.kind == "number" && right.kind == "number" {
			return typeNumber
		}
		return nil
	case "inc":
		inner := p.eval(expr.left)
		if inner != nil && inner.kind == "number" {
			return typeNumber
		}
		return nil
	case "ternary":
		left := p.eval(expr.left)
		right := p.eval(expr.right)
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

func (p *parser) bindingType(binding *binding) *jtype {
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
		typ = p.eval(binding.init)
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
		rhs := p.eval(assign.rhs)
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

func (p *parser) evalArray(expr *texpr) *jtype {
	if expr.arrayCtor && len(expr.elems) == 1 {
		only := p.eval(expr.elems[0])
		if only != nil && only.kind == "number" {
			return nil
		}
	}
	if len(expr.elems) == 0 {
		return nil
	}
	var elem *jtype
	for _, item := range expr.elems {
		next := p.eval(item)
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

func (p *parser) evalObject(expr *texpr) *jtype {
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
		typ := p.eval(field.expr)
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
