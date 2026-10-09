package nexus

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/checking"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/policy"
)

const securityNoInterpolatedSqlStringId = "interpolatedSqlString"

const securityNoInterpolatedSqlValueId = "interpolatedSqlValue"

// The rule's two messages, whose wording lives in `policy/messages/security-no-interpolated-sql-string.json`,
// the quoted one shared with the Swift twin.
var (
	securityNoInterpolatedSqlStringText = policy.MessageOf("nexus/security-no-interpolated-sql-string", securityNoInterpolatedSqlStringId)
	securityNoInterpolatedSqlValueText  = policy.MessageOf("nexus/security-no-interpolated-sql-string", securityNoInterpolatedSqlValueId)
)

// SecurityNoInterpolatedSqlString reports a value interpolated inside the single quotes of a SQL string
// written as a template literal, when the value's type can hold a quote.
//
//	invalid: `SELECT name FROM ZABCDRECORD WHERE name LIKE '%${word}%'`          (word: string)
//	invalid: `(LOWER(r.ZFIRSTNAME) LIKE '%${word}%' OR ...)`                     (a fragment, keyed on LIKE)
//	invalid: `WHERE viewIdentifier = '${viewIdentifier}'`                         (quotes doubled, backslash not)
//	valid:   `SELECT name FROM ZABCDRECORD WHERE name LIKE ?`                    (a bound parameter)
//	valid:   `SELECT * FROM Account WHERE AccountType = '${accountType}'`         (accountType: 'Bank' | 'Expense')
//	valid:   `SELECT * FROM t WHERE id = '${identifier}'`                         (identifier: number)
//	valid:   `SELECT * FROM \`${table}\` WHERE id = ?`                            (an identifier, not a string)
//	valid:   `name contains '${searchTerm}'`                                     (not SQL: no statement keyword)
//
// # Where it came from
//
// Two sites in ahra, found by the new-rules sweep (`#tevhg3f`), built as task `#3nvwzjs`:
//
//   - `modules/apple/contacts/ContactsApi.ts:62`, the contact search: each search word goes into
//     `LIKE '%${word}%'`, so `ahra contacts search "O'Brien"` ends the SQL string at the apostrophe, the
//     query fails, the `try` around it swallows the failure, and the search returns nothing.
//   - `modules/os/sensation/AhraOsTriggers.ts:1589`, a counter nerve over Phi's production
//     `EngagementEvent` table. It doubles every `'` before interpolating, which closes the quote hole
//     for SQLite, but the query runs on MySQL, where a backslash also escapes: a `viewIdentifier`
//     ending in `\` escapes the closing quote, and the string runs on into the rest of the query.
//
// # What makes a template SQL
//
// The shape of its own text, never the driver it is handed to. eslint-plugin-unicorn's version only
// recognizes `node:sqlite` and would find nothing in ahra, which uses `better-sqlite3`, the `sqlite3`
// and `mysql` command lines, PlanetScale, Google Ads and QuickBooks. Two shapes count, and a template
// matching neither is left alone:
//
//   - **A statement.** The template's text begins, after whitespace and `(`, with an uppercase SQL
//     statement or clause keyword (`SELECT`, `INSERT`, `UPDATE`, `DELETE`, `REPLACE`, `WITH`, `SHOW`,
//     `EXPLAIN`, `PRAGMA`, `WHERE`, `AND`, `OR`, `SET`, `VALUES`, `HAVING`), and, unless that keyword
//     is `WHERE` or `HAVING`, at least one more uppercase SQL keyword stands outside its strings and
//     comments, so a shouted `SELECT '${option}' to continue` is not a statement. The whole template
//     is then read as SQL from its first character: single-quoted strings, where a doubled quote is one
//     quote, double-quoted and backquoted names, `--` and `/* */` comments. An interpolation is
//     reported when
//     it sits inside a single-quoted string that closes later in the same template and whose opening
//     quote stands where a value goes: after a comparison (`=`, `<>`, `!=`, `<`, `>`, `<=`, `>=`),
//     `||`, `+`, `(`, `,`, or the keywords `LIKE`, `GLOB`, `REGEXP`, `RLIKE`, `BETWEEN`, `AND`,
//     `SELECT`, `WHEN`, `THEN`, `ELSE`, `ESCAPE`, `DEFAULT` or `INTERVAL`.
//   - **A `LIKE` pattern in a fragment.** The contact search builds its condition as a fragment that
//     starts with `(LOWER(`, so it has no statement keyword to key on. A template that is not a
//     statement is still read where an uppercase `LIKE` is followed by a single quote: that quote
//     opens a SQL string, and every interpolation up to its closing quote is inside it. The template
//     must also carry another uppercase `AND`, `OR`, `NOT`, `WHERE`, `SELECT`, `FROM` or `ESCAPE`, so
//     a log line (`Matching names LIKE '${pattern}'`) is not taken for SQL, and the count of single
//     quotes before the `LIKE` must be even, so a `LIKE` written inside some other quoted text (`echo
//     'WHERE name LIKE '${pattern}' OR all'`) is not either. This is how the SQL inside the Python
//     scripts of `iMessageApi.ts` is read: Python's double and triple quotes around the query are not
//     single quotes, so the `LIKE '%...%'` inside them is found.
//
// Keywords are matched in uppercase only. Lowercase SQL is missed rather than risk reading English
// as a statement: ahra's lowercase matches for a keyword before `'${` are all prose, such as
// `delete refused: '${name}' has ...` in `modules/os/lifecycle/AhraOsLifecycleCommandLineInterface.ts`
// and `Could not update floor designation for '${accountName}'` in the finance command line.
//
// Every way the reading can go wrong makes the rule silent on the whole template rather than wrong
// about one span. It gives up on a template whose quoted text holds a backslash (an escape in MySQL,
// an ordinary character in SQLite and standard SQL, so where the string ends depends on a dialect the
// text does not name), on `#` outside a string (a comment in MySQL only), on `--` not followed by
// whitespace (MySQL reads `a--1` as arithmetic), on a `$$` or `$tag$` dollar quote, and on a template
// that ends inside a string, a name or a block comment, which is a fragment whose other half this
// rule cannot see. A tagged template (a `sql` tag before the backquote) is never read: its tag
// decides what an interpolation becomes, and the usual SQL tags turn it into a bound parameter.
//
// # A value with no quotes around it
//
// In a confirmed statement, an interpolation outside any string, right after a comparison (`=`, `<>`,
// `!=`, `<`, `>`, `<=`, `>=`), is a value written straight into the SQL, and it is reported under its
// own id, `interpolatedSqlValue`, when its type can hold any text: `string`, `any`, `unknown`, a branded
// string, a `string` mapping, a template literal type with a text hole, or a type parameter whose
// constraint allows one. A number, a bigint, a boolean, a string literal and a union of them are fixed
// text and stay silent, which is how GAQL writes an enum (`campaign.status = ${status}`). Quoting and
// escaping are no answer here, so the escape chain above does not apply.
//
// Only comparisons, on purpose. After `(`, `,`, `AND`, `FROM` or `ORDER BY`, an unquoted interpolation
// builds SQL from code (an `IN (${placeholders})` list, an `AND ${condition}` fragment, a table or a
// column), and reading those would be false findings.
//
// Where it came from: `#3ng4vkk`. On ahra in 2026-10 it found 25 sites. 23 were GAQL ids typed `string`
// from the command line across `modules/google/ads` (fixed in `#kx29bm3` by typing them as numbers),
// and 2 were `sqlLiteral(...)` in `modules/data/DataMirrorReconcile.ts`, a helper that doubled only the
// quote while one query ran on MySQL, where a backslash still escapes (fixed in ahra 7d88414). www and
// api had none.
//
// # Which values can hold a quote
//
// The interpolated expression's type, through its constraint, decides. A value is reported when some
// member of its union is `string`, `any` or `unknown`, a branded `string` intersection, an
// `Uppercase<string>`-style mapping, a template literal type with a hole that can hold a quote, or a
// string literal whose text itself contains `'` or `\` (a constant that breaks the query is still a
// broken query). `null` and `undefined` print as words, and numbers, bigints, booleans, enums and
// quote-free string literals cannot hold a quote, so they are silent: `'${limit}'` with `limit:
// number` and `'${status}'` with `status: 'Active' | 'Paused'` are not findings. Object types (a
// `Date`, an array) are not reported either; their `toString` is the runtime's, and calling them
// quote-capable would be a guess. That is a missed finding for a `string[]` interpolated whole, which
// joins its members with commas.
//
// # Escaping, decided
//
// An escaped value is still a string, so the types alone would report it. The rule accepts exactly one
// escape as making a value safe, written where the reader can see it: a `replace` or `replaceAll`
// chain on a string whose last two steps double every backslash (`/\\/g` to `\\\\`) and double every
// single quote (`/'/g` to two quotes), in either order, either on the interpolated expression itself
// or in the initializer of the `const` it names. With both doubled, no input can close the string in SQLite,
// in standard SQL, or in MySQL with or without `NO_BACKSLASH_ESCAPES`, so the finding's claim, that a
// quote in the value ends the string, would be false. (In SQLite the doubled backslashes then reach
// the data as two characters; that is a wrong value, not an open string, and a bound parameter is
// still the better fix.)
//
// Every other escape is reported, on purpose:
//
//   - `'` doubled alone is the `AhraOsTriggers.ts` bug: correct for SQLite, open on MySQL, and the
//     text does not say which engine runs it.
//   - `'` turned into `\'` is the reverse: correct on MySQL, and on SQLite the backslash is an
//     ordinary character and the quote still closes the string.
//   - A call to an escaping function (`escapeSqlString(value)`) is a `string` whose body this rule
//     does not open, and naming functions that are trusted to escape would be a name heuristic. A
//     function that really escapes both characters is a finding that reads true-but-harmless; its fix
//     is a bound parameter, or the visible chain above.
//
// # What it does not see
//
// A query assembled with `+` from plain strings, a quote that opens in one template and closes in
// another, and a value interpolated inside a double-quoted string (a string in MySQL, a name in
// standard SQL) are all missed. So is SQL that starts with a lowercase keyword or with an
// interpolation. LIKE wildcards (`%`, `_`) in a value are not a quote problem and are not reported.
//
// # No fix
//
// The fix is a bound parameter, which moves the value out of the query text and into the driver call
// beside it, an edit across two expressions the rule cannot write safely. For a query language with
// no parameters (GAQL, QuickBooks), the fix is a narrower type or the visible escape above, and which
// one is right is the author's call.
var SecurityNoInterpolatedSqlString = rule.Rule{
	Name:             "nexus/security-no-interpolated-sql-string",
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.TypeChecker == nil {
			// Which values can hold a quote is a question about types. Without a checker every
			// number and literal union would be reported, so the rule declines the file.
			return nil
		}
		return rule.Listeners{
			ast.KindTemplateExpression: func(node *ast.Node) {
				if node.Parent != nil && node.Parent.Kind == ast.KindTaggedTemplateExpression {
					return
				}
				template := node.AsTemplateExpression()
				texts := []string{template.Head.Text()}
				for _, span := range template.TemplateSpans.Nodes {
					texts = append(texts, span.AsTemplateSpan().Literal.Text())
				}
				var quotedSpans []int
				var bareSpans []int
				if securityNoInterpolatedSqlStringOpensWithStatement(texts[0]) {
					quotedSpans, bareSpans = securityNoInterpolatedSqlStringReadStatement(texts)
				} else {
					quotedSpans = securityNoInterpolatedSqlStringReadLikePatterns(texts)
				}
				for _, spanIndex := range quotedSpans {
					expression := template.TemplateSpans.Nodes[spanIndex].AsTemplateSpan().Expression
					if !securityNoInterpolatedSqlStringCanHoldQuote(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(expression), 0) {
						continue
					}
					if securityNoInterpolatedSqlStringIsEscaped(ctx, expression) {
						continue
					}
					ctx.ReportNode(expression, rule.Message{Id: securityNoInterpolatedSqlStringId, Description: securityNoInterpolatedSqlStringText.Render(nil)})
				}
				for _, spanIndex := range bareSpans {
					expression := template.TemplateSpans.Nodes[spanIndex].AsTemplateSpan().Expression
					if !securityNoInterpolatedSqlStringCanHoldText(ctx.TypeChecker, ctx.TypeChecker.GetTypeAtLocation(expression), 0) {
						continue
					}
					ctx.ReportNode(expression, rule.Message{Id: securityNoInterpolatedSqlValueId, Description: securityNoInterpolatedSqlValueText.Render(nil)})
				}
			},
		}
	},
}

// securityNoInterpolatedSqlStringLeadingKeywords are the uppercase words a template's text may open
// with to be read as a SQL statement, or a clause of one.
var securityNoInterpolatedSqlStringLeadingKeywords = map[string]bool{
	"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true, "REPLACE": true, "WITH": true,
	"SHOW": true, "EXPLAIN": true, "PRAGMA": true, "WHERE": true, "AND": true, "OR": true, "SET": true,
	"VALUES": true, "HAVING": true,
}

// securityNoInterpolatedSqlStringSelfConfirmingKeywords are the leading keywords that need no second
// one: an uppercase `WHERE` or `HAVING` opening a template is a clause, where `SELECT`, `UPDATE`,
// `DELETE`, `WITH` or `AND` could open a shouted English sentence.
var securityNoInterpolatedSqlStringSelfConfirmingKeywords = map[string]bool{"WHERE": true, "HAVING": true}

// securityNoInterpolatedSqlStringKeywords are the uppercase words that count as SQL when a statement
// is confirmed by a second keyword outside its strings and comments.
var securityNoInterpolatedSqlStringKeywords = map[string]bool{
	"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true, "REPLACE": true, "WITH": true,
	"SHOW": true, "EXPLAIN": true, "PRAGMA": true, "WHERE": true, "AND": true, "OR": true, "SET": true,
	"VALUES": true, "HAVING": true, "FROM": true, "INTO": true, "LIKE": true, "GLOB": true,
	"BETWEEN": true, "JOIN": true, "ON": true, "IN": true, "IS": true, "NOT": true, "NULL": true,
	"ORDER": true, "GROUP": true, "BY": true, "LIMIT": true, "OFFSET": true, "AS": true,
	"DISTINCT": true, "COUNT": true, "TABLE": true, "STATUS": true, "CASE": true, "WHEN": true,
	"THEN": true, "ELSE": true, "END": true, "UNION": true, "ESCAPE": true, "ASC": true, "DESC": true,
}

// securityNoInterpolatedSqlStringValueTokens are the tokens after which an opening quote begins a
// value: comparisons, concatenation, an argument or list position, and the keywords that take a value.
var securityNoInterpolatedSqlStringValueTokens = map[string]bool{
	"=": true, "==": true, "<": true, ">": true, "<=": true, ">=": true, "<>": true, "!=": true,
	"||": true, "+": true, "(": true, ",": true,
	"LIKE": true, "GLOB": true, "REGEXP": true, "RLIKE": true, "BETWEEN": true, "AND": true,
	"SELECT": true, "WHEN": true, "THEN": true, "ELSE": true, "ESCAPE": true, "DEFAULT": true,
	"INTERVAL": true,
}

// securityNoInterpolatedSqlStringFragmentKeywords are the second keyword a `LIKE` fragment must carry.
var securityNoInterpolatedSqlStringFragmentKeywords = map[string]bool{
	"AND": true, "OR": true, "NOT": true, "WHERE": true, "SELECT": true, "FROM": true, "ESCAPE": true,
}

// securityNoInterpolatedSqlStringOpensWithStatement says whether a template's first text opens, after
// whitespace and `(`, with an uppercase statement or clause keyword.
func securityNoInterpolatedSqlStringOpensWithStatement(head string) bool {
	return securityNoInterpolatedSqlStringLeadingKeywords[securityNoInterpolatedSqlStringLeadingWord(head)]
}

// securityNoInterpolatedSqlStringLeadingWord is the first word of a template's text after whitespace
// and `(`.
func securityNoInterpolatedSqlStringLeadingWord(head string) string {
	trimmed := strings.TrimLeft(head, " \t\r\n(")
	return trimmed[:securityNoInterpolatedSqlStringWordLength(trimmed, 0)]
}

// securityNoInterpolatedSqlStringIsWordCharacter is a character of a SQL name or keyword.
func securityNoInterpolatedSqlStringIsWordCharacter(character byte) bool {
	return character == '_' || (character >= 'a' && character <= 'z') || (character >= 'A' && character <= 'Z') ||
		(character >= '0' && character <= '9')
}

// securityNoInterpolatedSqlStringWordLength is the length of the word starting at `start`.
func securityNoInterpolatedSqlStringWordLength(text string, start int) int {
	end := start
	for end < len(text) && securityNoInterpolatedSqlStringIsWordCharacter(text[end]) {
		end++
	}
	return end - start
}

func securityNoInterpolatedSqlStringIsWhitespace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\r' || character == '\n'
}

// securityNoInterpolatedSqlStringState is where a SQL reading stands between two characters.
type securityNoInterpolatedSqlStringState int

const (
	securityNoInterpolatedSqlStringInCode securityNoInterpolatedSqlStringState = iota
	securityNoInterpolatedSqlStringInString
	securityNoInterpolatedSqlStringInDoubleQuotes
	securityNoInterpolatedSqlStringInBackquotes
	securityNoInterpolatedSqlStringInLineComment
	securityNoInterpolatedSqlStringInBlockComment
)

// securityNoInterpolatedSqlStringReadStatement reads a template's texts as one SQL statement, with each
// interpolation an opaque value between two texts, and returns the indexes of the interpolations that
// sit inside a single-quoted string in a value position. It returns nothing when the text is not
// confirmed as SQL or when its reading is uncertain anywhere (see the doc comment).
func securityNoInterpolatedSqlStringReadStatement(texts []string) ([]int, []int) {
	state := securityNoInterpolatedSqlStringInCode
	previousToken := ""
	keywordCount := 0
	openedAtValue := false
	var quotedSpans []int
	var bareSpans []int
	for textIndex, text := range texts {
		for index := 0; index < len(text); {
			character := text[index]
			switch state {
			case securityNoInterpolatedSqlStringInCode:
				switch {
				case securityNoInterpolatedSqlStringIsWhitespace(character):
					index++
				case character == '\'':
					state = securityNoInterpolatedSqlStringInString
					openedAtValue = securityNoInterpolatedSqlStringValueTokens[previousToken]
					index++
				case character == '"':
					state = securityNoInterpolatedSqlStringInDoubleQuotes
					index++
				case character == '`':
					state = securityNoInterpolatedSqlStringInBackquotes
					index++
				case character == '-' && index+1 < len(text) && text[index+1] == '-':
					if index+2 >= len(text) || !securityNoInterpolatedSqlStringIsWhitespace(text[index+2]) {
						return nil, nil
					}
					state = securityNoInterpolatedSqlStringInLineComment
					index += 3
				case character == '/' && index+1 < len(text) && text[index+1] == '*':
					state = securityNoInterpolatedSqlStringInBlockComment
					index += 2
				case character == '#':
					return nil, nil
				case character == '$':
					// `$1` and `$name` are placeholders; `$$` and `$tag$` open a dollar-quoted string.
					length := securityNoInterpolatedSqlStringWordLength(text, index+1)
					if index+1+length < len(text) && text[index+1+length] == '$' {
						return nil, nil
					}
					previousToken = "$"
					index += 1 + length
				case securityNoInterpolatedSqlStringIsWordCharacter(character):
					length := securityNoInterpolatedSqlStringWordLength(text, index)
					word := text[index : index+length]
					if securityNoInterpolatedSqlStringKeywords[word] {
						keywordCount++
					}
					previousToken = strings.ToUpper(word)
					index += length
				case strings.IndexByte("=<>!|", character) >= 0:
					end := index
					for end < len(text) && strings.IndexByte("=<>!|", text[end]) >= 0 {
						end++
					}
					previousToken = text[index:end]
					index = end
				default:
					previousToken = string(character)
					index++
				}
			case securityNoInterpolatedSqlStringInString:
				switch character {
				case '\\':
					return nil, nil
				case '\'':
					if index+1 < len(text) && text[index+1] == '\'' {
						index += 2
						continue
					}
					state = securityNoInterpolatedSqlStringInCode
					previousToken = "'"
				}
				index++
			case securityNoInterpolatedSqlStringInDoubleQuotes, securityNoInterpolatedSqlStringInBackquotes:
				closing := byte('"')
				if state == securityNoInterpolatedSqlStringInBackquotes {
					closing = '`'
				}
				switch character {
				case '\\':
					return nil, nil
				case closing:
					if index+1 < len(text) && text[index+1] == closing {
						index += 2
						continue
					}
					state = securityNoInterpolatedSqlStringInCode
					previousToken = string(closing)
				}
				index++
			case securityNoInterpolatedSqlStringInLineComment:
				if character == '\n' {
					state = securityNoInterpolatedSqlStringInCode
				}
				index++
			case securityNoInterpolatedSqlStringInBlockComment:
				if character == '*' && index+1 < len(text) && text[index+1] == '/' {
					state = securityNoInterpolatedSqlStringInCode
					index += 2
					continue
				}
				index++
			}
		}
		if textIndex == len(texts)-1 {
			break
		}
		// The interpolation after this text.
		switch state {
		case securityNoInterpolatedSqlStringInString:
			if openedAtValue {
				quotedSpans = append(quotedSpans, textIndex)
			}
		case securityNoInterpolatedSqlStringInCode:
			if securityNoInterpolatedSqlStringComparisons[previousToken] {
				bareSpans = append(bareSpans, textIndex)
			}
			previousToken = "${}"
		}
	}
	if state != securityNoInterpolatedSqlStringInCode && state != securityNoInterpolatedSqlStringInLineComment {
		return nil, nil
	}
	if keywordCount < 2 && !securityNoInterpolatedSqlStringSelfConfirmingKeywords[securityNoInterpolatedSqlStringLeadingWord(texts[0])] {
		return nil, nil
	}
	return quotedSpans, bareSpans
}

// securityNoInterpolatedSqlStringComparisons are the operators after which an unquoted interpolation is
// a value compared, not a name or a fragment of SQL.
var securityNoInterpolatedSqlStringComparisons = map[string]bool{
	"=": true, "==": true, "<": true, ">": true, "<=": true, ">=": true, "<>": true, "!=": true,
}

// securityNoInterpolatedSqlStringCanHoldText says whether a value of this type can print arbitrary text,
// rather than a number, a boolean or a fixed literal.
func securityNoInterpolatedSqlStringCanHoldText(typeChecker *checker.Checker, valueType *checker.Type, depth int) bool {
	if valueType == nil || depth > 8 {
		return false
	}
	for part := range type_checking.UnionTypePartsSeq(valueType) {
		flags := part.Flags()
		switch {
		case type_checking.IsIntrinsicErrorType(part):
		case flags&checker.TypeFlagsAnyOrUnknown != 0:
			return true
		case flags&(checker.TypeFlagsString|checker.TypeFlagsStringMapping) != 0:
			return true
		case flags&checker.TypeFlagsTemplateLiteral != 0:
			for _, hole := range part.AsTemplateLiteralType().Types() {
				if securityNoInterpolatedSqlStringCanHoldText(typeChecker, hole, depth+1) {
					return true
				}
			}
		case flags&checker.TypeFlagsIntersection != 0:
			for _, member := range type_checking.IntersectionTypeParts(part) {
				if member.Flags()&(checker.TypeFlagsString|checker.TypeFlagsStringMapping) != 0 {
					return true
				}
			}
		case flags&checker.TypeFlagsInstantiable != 0:
			constraint := checker.Checker_getBaseConstraintOfType(typeChecker, part)
			if constraint == nil || constraint == part || securityNoInterpolatedSqlStringCanHoldText(typeChecker, constraint, depth+1) {
				return true
			}
		}
	}
	return false
}

// securityNoInterpolatedSqlStringReadLikePatterns finds, in a template that is not a statement, every
// uppercase `LIKE` followed by a single quote, and returns the indexes of the interpolations between
// that quote and its closing one.
func securityNoInterpolatedSqlStringReadLikePatterns(texts []string) []int {
	if !securityNoInterpolatedSqlStringHasFragmentKeyword(texts) {
		return nil
	}
	var quotedSpans []int
	quotesBefore := 0
	for textIndex := 0; textIndex < len(texts); textIndex++ {
		text := texts[textIndex]
		for index := 0; index < len(text); index++ {
			if text[index] == '\'' {
				quotesBefore++
				continue
			}
			if !strings.HasPrefix(text[index:], "LIKE") ||
				(index > 0 && securityNoInterpolatedSqlStringIsWordCharacter(text[index-1])) {
				continue
			}
			quote := index + len("LIKE")
			if quote >= len(text) || !securityNoInterpolatedSqlStringIsWhitespace(text[quote]) {
				continue
			}
			for quote < len(text) && securityNoInterpolatedSqlStringIsWhitespace(text[quote]) {
				quote++
			}
			if quote >= len(text) || text[quote] != '\'' {
				continue
			}
			if quotesBefore%2 != 0 {
				continue
			}
			spans, closingText, closingIndex := securityNoInterpolatedSqlStringReadString(texts, textIndex, quote+1)
			if closingText < 0 {
				return quotedSpans
			}
			quotedSpans = append(quotedSpans, spans...)
			// Resume after the closing quote; the opening and closing quotes are a pair, so the
			// count before what follows stays even.
			textIndex = closingText
			text = texts[textIndex]
			index = closingIndex
		}
	}
	return quotedSpans
}

// securityNoInterpolatedSqlStringHasFragmentKeyword says whether a fragment carries an uppercase
// `AND`, `OR`, `NOT`, `WHERE`, `SELECT`, `FROM` or `ESCAPE` as a whole word anywhere in its text.
func securityNoInterpolatedSqlStringHasFragmentKeyword(texts []string) bool {
	for _, text := range texts {
		for index := 0; index < len(text); {
			length := securityNoInterpolatedSqlStringWordLength(text, index)
			if length == 0 {
				index++
				continue
			}
			if securityNoInterpolatedSqlStringFragmentKeywords[text[index:index+length]] {
				return true
			}
			index += length
		}
	}
	return false
}

// securityNoInterpolatedSqlStringReadString reads a single-quoted SQL string from just after its
// opening quote, across interpolations, and returns the interpolations inside it with the text and
// index of its closing quote. It returns a closing text of -1 when the string never closes or holds a
// backslash.
func securityNoInterpolatedSqlStringReadString(texts []string, textIndex int, start int) ([]int, int, int) {
	var spans []int
	index := start
	for ; textIndex < len(texts); textIndex++ {
		text := texts[textIndex]
		for ; index < len(text); index++ {
			switch text[index] {
			case '\\':
				return nil, -1, 0
			case '\'':
				if index+1 < len(text) && text[index+1] == '\'' {
					index++
					continue
				}
				return spans, textIndex, index
			}
		}
		if textIndex < len(texts)-1 {
			spans = append(spans, textIndex)
		}
		index = 0
	}
	return nil, -1, 0
}

// securityNoInterpolatedSqlStringCanHoldQuote says whether a value of this type can print a single
// quote or a backslash into the text around it. `depth` bounds the walk through template literal holes.
func securityNoInterpolatedSqlStringCanHoldQuote(typeChecker *checker.Checker, valueType *checker.Type, depth int) bool {
	if valueType == nil || depth > 8 {
		return false
	}
	for part := range type_checking.UnionTypePartsSeq(valueType) {
		if securityNoInterpolatedSqlStringPartCanHoldQuote(typeChecker, part, depth) {
			return true
		}
	}
	return false
}

func securityNoInterpolatedSqlStringPartCanHoldQuote(typeChecker *checker.Checker, part *checker.Type, depth int) bool {
	flags := part.Flags()
	switch {
	case type_checking.IsIntrinsicErrorType(part):
		return false
	case flags&checker.TypeFlagsAnyOrUnknown != 0:
		return true
	case flags&checker.TypeFlagsStringLiteral != 0:
		text, _ := part.AsLiteralType().Value().(string)
		return strings.ContainsAny(text, "'\\")
	case flags&(checker.TypeFlagsString|checker.TypeFlagsStringMapping) != 0:
		return true
	case flags&checker.TypeFlagsTemplateLiteral != 0:
		templateType := part.AsTemplateLiteralType()
		for _, text := range templateType.Texts() {
			if strings.ContainsAny(text, "'\\") {
				return true
			}
		}
		for _, hole := range templateType.Types() {
			if securityNoInterpolatedSqlStringCanHoldQuote(typeChecker, hole, depth+1) {
				return true
			}
		}
		return false
	case flags&checker.TypeFlagsIntersection != 0:
		// A branded string (`string & { brand: 'Email' }`) is a string; an intersection of objects
		// is an object.
		for _, member := range type_checking.IntersectionTypeParts(part) {
			if member.Flags()&checker.TypeFlagsStringLike != 0 {
				return securityNoInterpolatedSqlStringPartCanHoldQuote(typeChecker, member, depth)
			}
		}
	case flags&checker.TypeFlagsInstantiable != 0:
		// A type parameter, an indexed access or a conditional prints whatever its constraint
		// allows, and with no constraint it is as open as `unknown`.
		constraint := checker.Checker_getBaseConstraintOfType(typeChecker, part)
		if constraint == nil || constraint == part {
			return true
		}
		return securityNoInterpolatedSqlStringCanHoldQuote(typeChecker, constraint, depth+1)
	}
	return false
}

// securityNoInterpolatedSqlStringIsEscaped says whether an interpolated expression, or the
// initializer of the `const` it names, ends in the one escape the doc comment accepts.
func securityNoInterpolatedSqlStringIsEscaped(ctx rule.Context, expression *ast.Node) bool {
	expression = ast.SkipParentheses(expression)
	if securityNoInterpolatedSqlStringIsEscapeChain(ctx, expression) {
		return true
	}
	if expression.Kind != ast.KindIdentifier {
		return false
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(expression)
	if symbol == nil || symbol.ValueDeclaration == nil || symbol.ValueDeclaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	declaration := symbol.ValueDeclaration
	if declaration.Parent == nil || declaration.Parent.Kind != ast.KindVariableDeclarationList ||
		declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	return initializer != nil && securityNoInterpolatedSqlStringIsEscapeChain(ctx, ast.SkipParentheses(initializer))
}

// securityNoInterpolatedSqlStringIsEscapeChain says whether an expression's last two calls are a
// string `replace` or `replaceAll` doubling every backslash and one doubling every single quote, in
// either order.
func securityNoInterpolatedSqlStringIsEscapeChain(ctx rule.Context, expression *ast.Node) bool {
	outer, outerReceiver := securityNoInterpolatedSqlStringReplaceStep(ctx, expression)
	if outer == "" {
		return false
	}
	inner, _ := securityNoInterpolatedSqlStringReplaceStep(ctx, ast.SkipParentheses(outerReceiver))
	return (outer == "'" && inner == "\\") || (outer == "\\" && inner == "'")
}

// securityNoInterpolatedSqlStringReplaceStep reads a `replace` with a global `/'/g`, or a
// `replaceAll` with the string `"'"`, whose replacement is two quotes, and their backslash twins
// (`/\\/g` or `'\\'` to `'\\\\'`), on a string receiver. It returns which character the step doubles,
// with the receiver, and an empty character for anything else.
func securityNoInterpolatedSqlStringReplaceStep(ctx rule.Context, expression *ast.Node) (string, *ast.Node) {
	if expression.Kind != ast.KindCallExpression {
		return "", nil
	}
	call := expression.AsCallExpression()
	if call.QuestionDotToken != nil || call.Arguments == nil || len(call.Arguments.Nodes) != 2 ||
		call.Expression.Kind != ast.KindPropertyAccessExpression {
		return "", nil
	}
	access := call.Expression.AsPropertyAccessExpression()
	method := access.Name().Text()
	if method != "replace" && method != "replaceAll" {
		return "", nil
	}
	receiverType := ctx.TypeChecker.GetTypeAtLocation(access.Expression)
	if receiverType == nil || receiverType.Flags()&checker.TypeFlagsStringLike == 0 {
		return "", nil
	}
	pattern := call.Arguments.Nodes[0]
	replacement := call.Arguments.Nodes[1]
	if !ast.IsStringLiteralLike(replacement) {
		return "", nil
	}
	var character string
	switch pattern.Kind {
	case ast.KindRegularExpressionLiteral:
		text := pattern.Text()
		closing := strings.LastIndexByte(text, '/')
		if closing <= 0 || !strings.Contains(text[closing+1:], "g") {
			return "", nil
		}
		switch text[1:closing] {
		case "'":
			character = "'"
		case "\\\\":
			character = "\\"
		}
	case ast.KindStringLiteral, ast.KindNoSubstitutionTemplateLiteral:
		if method != "replaceAll" {
			return "", nil
		}
		character = pattern.Text()
	}
	if (character != "'" && character != "\\") || replacement.Text() != character+character {
		return "", nil
	}
	return character, access.Expression
}
