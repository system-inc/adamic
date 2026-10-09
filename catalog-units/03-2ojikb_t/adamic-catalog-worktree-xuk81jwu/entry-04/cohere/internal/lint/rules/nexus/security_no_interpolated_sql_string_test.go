package nexus

import (
	"strings"
	"testing"

	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
)

const securityNoInterpolatedSqlStringFile = "/repository/source/SecurityNoInterpolatedSqlString.ts"

func securityNoInterpolatedSqlStringSource(lines ...string) string {
	return strings.Join(lines, "\n") + "\n"
}

// securityNoInterpolatedSqlStringContactsApi models `modules/apple/contacts/ContactsApi.ts` in ahra on
// 2026-10-02, trimmed to the two lookups that build SQL: the name search, whose condition is a
// fragment that starts with `(LOWER(` and is keyed on `LIKE`, and the phone lookup, a full statement.
// The `sqlite3` command line is declared rather than imported, so the fixture needs no `@types/node`.
func securityNoInterpolatedSqlStringContactsApi() string {
	return securityNoInterpolatedSqlStringSource(
		"declare function execSync(command: string, options: { encoding: 'utf-8'; timeout: number }): string;",
		"declare function findAllDatabases(): string[];",
		"export function searchContacts(searchName: string): string[] {",
		"    const searchWords = searchName",
		"        .toLowerCase()",
		"        .split(/\\s+/)",
		"        .filter((word) => word.length > 0);",
		"    const results: string[] = [];",
		"    for(const databasePath of findAllDatabases()) {",
		"        try {",
		"            const wordConditions = searchWords",
		"                .map(",
		"                    (word) =>",
		"                        `(LOWER(r.ZFIRSTNAME) LIKE '%${word}%' OR LOWER(r.ZLASTNAME) LIKE '%${word}%' OR LOWER(r.ZORGANIZATION) LIKE '%${word}%')`,",
		"                )",
		"                .join(' AND ');",
		"            const query = `",
		"        SELECT",
		"          r.Z_PK,",
		"          r.ZFIRSTNAME,",
		"          r.ZLASTNAME,",
		"          r.ZORGANIZATION,",
		"          p.ZFULLNUMBER,",
		"          p.ZLABEL",
		"        FROM ZABCDRECORD r",
		"        LEFT JOIN ZABCDPHONENUMBER p ON p.ZOWNER = r.Z_PK",
		"        WHERE ${wordConditions}",
		"      `;",
		"            results.push(execSync(`sqlite3 \"${databasePath}\" \"${query}\"`, { encoding: 'utf-8', timeout: 5000 }));",
		"        }",
		"        catch {",
		"            continue;",
		"        }",
		"    }",
		"    return results;",
		"}",
		"export function findContactByPhone(phoneNumber: string): string | null {",
		"    const digits = phoneNumber.replace(/\\D/g, '').slice(-10);",
		"    if(digits.length < 10) return null;",
		"    for(const databasePath of findAllDatabases()) {",
		"        const query = `",
		"        SELECT",
		"          r.ZFIRSTNAME,",
		"          r.ZLASTNAME,",
		"          r.ZORGANIZATION",
		"        FROM ZABCDRECORD r",
		"        JOIN ZABCDPHONENUMBER p ON p.ZOWNER = r.Z_PK",
		"        WHERE REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(p.ZFULLNUMBER, ' ', ''), '-', ''), '(', ''), ')', ''), '+', '') LIKE '%${digits}'",
		"        LIMIT 1",
		"      `;",
		"        const result = execSync(`sqlite3 \"${databasePath}\" \"${query}\"`, { encoding: 'utf-8', timeout: 5000 });",
		"        if(result.trim()) return result.trim();",
		"    }",
		"    return null;",
		"}",
	)
}

// securityNoInterpolatedSqlStringBoundContactsApi is the same two lookups after the fix: every value
// is a bound parameter, the `%` wildcards travel with the value (`%${word}%`, a template that is not
// SQL), and the queries carry only `?`.
func securityNoInterpolatedSqlStringBoundContactsApi() string {
	return securityNoInterpolatedSqlStringSource(
		"declare class Database {",
		"    constructor(path: string, options: { readonly: boolean });",
		"    prepare(sql: string): { all(...parameters: string[]): string[]; get(...parameters: string[]): string | undefined };",
		"}",
		"declare function findAllDatabases(): string[];",
		"export function searchContacts(searchName: string): string[] {",
		"    const searchWords = searchName",
		"        .toLowerCase()",
		"        .split(/\\s+/)",
		"        .filter((word) => word.length > 0);",
		"    const results: string[] = [];",
		"    for(const databasePath of findAllDatabases()) {",
		"        const wordConditions = searchWords",
		"            .map(() => `(LOWER(r.ZFIRSTNAME) LIKE ? OR LOWER(r.ZLASTNAME) LIKE ? OR LOWER(r.ZORGANIZATION) LIKE ?)`)",
		"            .join(' AND ');",
		"        const patterns = searchWords.flatMap((word) => [`%${word}%`, `%${word}%`, `%${word}%`]);",
		"        const query = `",
		"        SELECT r.Z_PK, r.ZFIRSTNAME, r.ZLASTNAME, r.ZORGANIZATION, p.ZFULLNUMBER, p.ZLABEL",
		"        FROM ZABCDRECORD r",
		"        LEFT JOIN ZABCDPHONENUMBER p ON p.ZOWNER = r.Z_PK",
		"        WHERE ${wordConditions}",
		"      `;",
		"        results.push(...new Database(databasePath, { readonly: true }).prepare(query).all(...patterns));",
		"    }",
		"    return results;",
		"}",
		"export function findContactByPhone(phoneNumber: string): string | null {",
		"    const digits = phoneNumber.replace(/\\D/g, '').slice(-10);",
		"    if(digits.length < 10) return null;",
		"    for(const databasePath of findAllDatabases()) {",
		"        const query = `",
		"        SELECT r.ZFIRSTNAME, r.ZLASTNAME, r.ZORGANIZATION",
		"        FROM ZABCDRECORD r",
		"        JOIN ZABCDPHONENUMBER p ON p.ZOWNER = r.Z_PK",
		"        WHERE REPLACE(REPLACE(REPLACE(REPLACE(REPLACE(p.ZFULLNUMBER, ' ', ''), '-', ''), '(', ''), ')', ''), '+', '') LIKE '%' || ?",
		"        LIMIT 1",
		"      `;",
		"        const result = new Database(databasePath, { readonly: true }).prepare(query).get(digits);",
		"        if(result) return result;",
		"    }",
		"    return null;",
		"}",
	)
}

// securityNoInterpolatedSqlStringAhraOsTriggers models `phiCountThreshold` in
// `modules/os/sensation/AhraOsTriggers.ts` in ahra on 2026-10-02 (`:1588`), which counts rows of Phi's
// production `EngagementEvent` table over MySQL. `escapeLine` is the line that escapes the identifier:
// quotes only before the fix, backslashes and quotes after it. The error message
// `unknown metric '${metric}'` is a quoted interpolation too, in prose, and stays silent either way.
func securityNoInterpolatedSqlStringAhraOsTriggers(escapeLine string) string {
	return securityNoInterpolatedSqlStringSource(
		"declare function readProductionCount(sql: string): Promise<number>;",
		"export async function phiCountThreshold(parameters: Record<string, unknown>): Promise<number> {",
		"    const metric = String(parameters.metric ?? 'EngagementEvents');",
		"    let sql: string;",
		"    if(metric === 'EngagementEvents') {",
		escapeLine,
		"        if(!viewIdentifier) throw new Error('phiCountThreshold: EngagementEvents needs a viewIdentifier');",
		"        // Exact page + its query-string variants (utm-tagged visits land as",
		"        // '/stack?utm_...'), without swallowing subpages like '/stack/reviews'.",
		"        sql =",
		"            `SELECT COUNT(*) AS totalCount FROM EngagementEvent ` +",
		"            `WHERE viewIdentifier = '${viewIdentifier}' OR viewIdentifier LIKE '${viewIdentifier}?%'`;",
		"    }",
		"    else if(metric === 'Orders') {",
		"        sql = `SELECT COUNT(*) AS totalCount FROM CommerceOrder`;",
		"    }",
		"    else {",
		"        throw new Error(`phiCountThreshold: unknown metric '${metric}' (EngagementEvents | Orders)`);",
		"    }",
		"    return readProductionCount(sql);",
		"}",
	)
}

// securityNoInterpolatedSqlStringWording is the quoted-value message, written out here rather than read
// from the rule, so this compares the finding against the text it had before the wording moved to
// policy/messages.
const securityNoInterpolatedSqlStringWording = "This value is written between the single quotes of a SQL string, so a quote in it ends the string early: the query breaks on ordinary input like `O'Brien`, and a crafted value runs as SQL. Pass it as a bound parameter (`WHERE name LIKE ?` with the value beside the query). Where the query language has no parameters, give the value a type that cannot hold a quote (a number, a union of literals, a template literal type such as `${number}-${number}-${number}`), or escape it in place with `.replace(/\\\\/g, '\\\\\\\\').replace(/'/g, \"''\")`, which doubles backslashes as well as quotes."

// securityNoInterpolatedSqlStringReported is the source text each finding points at, in order.
func securityNoInterpolatedSqlStringReported(result rule_testing.Result, sourceText string) []string {
	reported := make([]string, 0, len(result.Diagnostics))
	for _, diagnostic := range result.Diagnostics {
		reported = append(reported, sourceText[diagnostic.Range.Pos():diagnostic.Range.End()])
	}
	return reported
}

func securityNoInterpolatedSqlStringExpectSpans(t *testing.T, result rule_testing.Result, sourceText string, wantSpans []string) {
	t.Helper()
	wantIds := make([]string, 0, len(wantSpans))
	for range wantSpans {
		wantIds = append(wantIds, securityNoInterpolatedSqlStringId)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	reported := securityNoInterpolatedSqlStringReported(result, sourceText)
	for index := range wantSpans {
		if reported[index] != wantSpans[index] {
			t.Fatalf("finding %d points at\n%s\nwant\n%s", index, reported[index], wantSpans[index])
		}
	}
	for _, diagnostic := range result.Diagnostics {
		if diagnostic.Message.Description != securityNoInterpolatedSqlStringWording {
			t.Fatalf("message is %q", diagnostic.Message.Description)
		}
		if len(diagnostic.Fixes) != 0 || len(diagnostic.Suggestions) != 0 {
			t.Fatalf("expected no fix and no suggestion, got %d and %d", len(diagnostic.Fixes), len(diagnostic.Suggestions))
		}
	}
}

// The real contact search, before the fix: the three search-word conditions in the fragment, and the
// phone digits in the full statement. The `sqlite3 "${databasePath}" "${query}"` command beside them
// is not SQL and stays silent.
func TestSecurityNoInterpolatedSqlStringFiresOnContactsApi(t *testing.T) {
	t.Parallel()

	sourceText := securityNoInterpolatedSqlStringContactsApi()
	result := rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, sourceText)
	securityNoInterpolatedSqlStringExpectSpans(t, result, sourceText, []string{"word", "word", "word", "digits"})
}

func TestSecurityNoInterpolatedSqlStringStaysSilentOnBoundContactsApi(t *testing.T) {
	t.Parallel()

	result := rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, securityNoInterpolatedSqlStringBoundContactsApi())
	rule_testing.ExpectClean(t, result)
}

// The real counter nerve, before the fix: quotes are doubled, backslashes are not, and the query runs
// on MySQL. Both interpolations of the identifier fire.
func TestSecurityNoInterpolatedSqlStringFiresOnQuoteOnlyEscape(t *testing.T) {
	t.Parallel()

	sourceText := securityNoInterpolatedSqlStringAhraOsTriggers("        const viewIdentifier = String(parameters.viewIdentifier ?? '').replace(/'/g, \"''\");")
	result := rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, sourceText)
	securityNoInterpolatedSqlStringExpectSpans(t, result, sourceText, []string{"viewIdentifier", "viewIdentifier"})
}

// The same nerve with backslashes doubled before quotes: no input can close the string on any engine,
// so the finding's claim would be false, and nothing fires.
func TestSecurityNoInterpolatedSqlStringStaysSilentOnCompleteEscape(t *testing.T) {
	t.Parallel()

	sourceText := securityNoInterpolatedSqlStringAhraOsTriggers("        const viewIdentifier = String(parameters.viewIdentifier ?? '').replace(/\\\\/g, '\\\\\\\\').replace(/'/g, \"''\");")
	result := rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, sourceText)
	rule_testing.ExpectClean(t, result)
}

func TestSecurityNoInterpolatedSqlStringFires(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name      string
		lines     []string
		wantSpans []string
	}{
		{"a LIKE pattern in a SELECT", []string{
			"export function find(name: string) {",
			"    return `SELECT name FROM ZABCDRECORD WHERE name LIKE '%${name}%'`;",
			"}",
		}, []string{"name"}},
		// Both values sit in a `VALUES` list, after `(` and after `,`.
		{"values in an INSERT", []string{
			"export function insert(name: string, email: string) {",
			"    return `INSERT INTO Contact (name, email) VALUES ('${name}', '${email}')`;",
			"}",
		}, []string{"name", "email"}},
		{"a BETWEEN range", []string{
			"export function range(start: string, end: string) {",
			"    return `SELECT id FROM campaign WHERE segments.date BETWEEN '${start}' AND '${end}'`;",
			"}",
		}, []string{"start", "end"}},
		{"an UPDATE assignment", []string{
			"export function rename(name: string, identifier: number) {",
			"    return `UPDATE Contact SET name = '${name}' WHERE id = ${identifier}`;",
			"}",
		}, []string{"name"}},
		// `$.` inside the string is text, and the string opens after a `,`.
		{"a JSON path argument", []string{
			"export function extract(field: string) {",
			"    return `SELECT JSON_EXTRACT(data, '$.${field}') FROM EngagementEvent`;",
			"}",
		}, []string{"field"}},
		// `AdsRegistry.ts:186` in ahra: `WHERE` needs no second keyword.
		{"a WHERE clause on its own", []string{
			"declare function mapUniversalToGoogleStatus(status: 'active' | 'paused'): string;",
			"export function filter(status: 'active' | 'paused') {",
			"    return `WHERE campaign.status = '${mapUniversalToGoogleStatus(status)}'`;",
			"}",
		}, []string{"mapUniversalToGoogleStatus(status)"}},
		{"an AND clause continuing a query", []string{
			"export function clause(accountId: string) {",
			"    return ` AND accountId = '${accountId}' ORDER BY createdAt`;",
			"}",
		}, []string{"accountId"}},
		{"two interpolations in one string", []string{
			"export function find(first: string, last: string) {",
			"    return `SELECT id FROM Contact WHERE name LIKE '%${first}${last}%'`;",
			"}",
		}, []string{"first", "last"}},
		// The `''` is one quote inside the string, so `${suffix}` is still inside it.
		{"a string holding a doubled quote", []string{
			"export function find(suffix: string) {",
			"    return `SELECT id FROM Contact WHERE name = 'O''${suffix}'`;",
			"}",
		}, []string{"suffix"}},
		// The apostrophe in the comment does not open a string.
		{"a comment before the value", []string{
			"export function find(name: string) {",
			"    return `SELECT id FROM Contact -- the user's search",
			"        WHERE name = '${name}'`;",
			"}",
		}, []string{"name"}},
		{"a value of type any", []string{
			"export function find(row: any) {",
			"    return `SELECT id FROM Contact WHERE name = '${row.name}'`;",
			"}",
		}, []string{"row.name"}},
		{"a value of type unknown", []string{
			"export function find(name: unknown) {",
			"    return `SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}, []string{"name"}},
		{"a branded string", []string{
			"type Email = string & { readonly brand: 'Email' };",
			"export function find(email: Email) {",
			"    return `SELECT id FROM AccountEmail WHERE emailAddress = '${email}'`;",
			"}",
		}, []string{"email"}},
		{"a template literal type with a string hole", []string{
			"export function find(path: `/${string}`) {",
			"    return `SELECT id FROM EngagementEvent WHERE viewIdentifier = '${path}'`;",
			"}",
		}, []string{"path"}},
		// A constant that breaks the query is still a broken query.
		{"a string literal that holds a quote", []string{
			"export function find(name: 'plain' | \"O'Brien\") {",
			"    return `SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}, []string{"name"}},
		{"a type parameter constrained to string", []string{
			"export function find<Name extends string>(name: Name) {",
			"    return `SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}, []string{"name"}},
		{"a string mapping", []string{
			"export function find(name: Uppercase<string>) {",
			"    return `SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}, []string{"name"}},
		{"a nullable string", []string{
			"export function find(name: string | undefined) {",
			"    return `SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}, []string{"name"}},
		// Correct on MySQL, open on SQLite.
		{"quotes escaped with a backslash", []string{
			"export function find(name: string) {",
			"    return `SELECT id FROM Contact WHERE name = '${name.replace(/'/g, \"\\\\'\")}'`;",
			"}",
		}, []string{"name.replace(/'/g, \"\\\\'\")"}},
		{"backslashes doubled without quotes", []string{
			"export function find(name: string) {",
			"    return `SELECT id FROM Contact WHERE name = '${name.replace(/\\\\/g, '\\\\\\\\')}'`;",
			"}",
		}, []string{"name.replace(/\\\\/g, '\\\\\\\\')"}},
		{"an escaping function the rule does not open", []string{
			"declare function escapeSqlString(value: string): string;",
			"export function find(name: string) {",
			"    return `SELECT id FROM Contact WHERE name = '${escapeSqlString(name)}'`;",
			"}",
		}, []string{"escapeSqlString(name)"}},
		// Only a `const` is followed to its initializer.
		{"an escape held by a let", []string{
			"export function find(input: string) {",
			"    let name = input.replace(/\\\\/g, '\\\\\\\\').replace(/'/g, \"''\");",
			"    name = input;",
			"    return `SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}, []string{"name"}},
		// `/'/` without `g` replaces the first quote only.
		{"a non-global escape", []string{
			"export function find(name: string) {",
			"    return `SELECT id FROM Contact WHERE name = '${name.replace(/\\\\/g, '\\\\\\\\').replace(/'/, \"''\")}'`;",
			"}",
		}, []string{"name.replace(/\\\\/g, '\\\\\\\\').replace(/'/, \"''\")"}},
		// `iMessageApi.ts` in ahra: Python's triple quotes are not single quotes, so the `LIKE` string inside them is found.
		{"SQL inside a Python script", []string{
			"export function script(normalized: string, count: number) {",
			"    return `",
			"import sqlite3",
			"conn = sqlite3.connect(db_path)",
			"query = \"\"\"",
			"SELECT datetime(m.date, 'unixepoch', 'localtime') as date, m.text",
			"FROM message m",
			"WHERE c.chat_identifier LIKE '%${normalized}%'",
			"LIMIT ${count}",
			"\"\"\"",
			"`;",
			"}",
		}, []string{"normalized"}},
		{"a NOT LIKE fragment", []string{
			"export function clause(word: string) {",
			"    return `(r.ZFIRSTNAME NOT LIKE '%${word}%' OR r.ZLASTNAME IS NULL)`;",
			"}",
		}, []string{"word"}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			sourceText := securityNoInterpolatedSqlStringSource(testCase.lines...)
			result := rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, sourceText)
			securityNoInterpolatedSqlStringExpectSpans(t, result, sourceText, testCase.wantSpans)
		})
	}
}

func TestSecurityNoInterpolatedSqlStringStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a number", []string{
			"export function find(identifier: number, limit: bigint, active: boolean) {",
			"    return `SELECT id FROM Contact WHERE id = '${identifier}' AND rank < '${limit}' AND active = '${active}'`;",
			"}",
		}},
		// `QuickBooksAdapter.ts:330` in ahra, with the type it should have.
		{"a union of string literals", []string{
			"export function accounts(accountType: 'Bank' | 'CreditCard') {",
			"    return `SELECT * FROM Account WHERE AccountType = '${accountType}' STARTPOSITION 1 MAXRESULTS 100`;",
			"}",
		}},
		{"an enum member", []string{
			"enum EnergyRecordKind { Consume = 'Consume', Grant = 'Grant' }",
			"export function energy() {",
			"    return `SELECT id FROM EnergyRecord WHERE type = '${EnergyRecordKind.Consume}'`;",
			"}",
		}},
		{"a date-shaped template literal type", []string{
			"type DayString = `${number}-${number}-${number}`;",
			"export function report(date: DayString) {",
			"    return `SELECT campaign.name FROM campaign WHERE segments.date = '${date}'`;",
			"}",
		}},
		{"null and undefined", []string{
			"export function find(nothing: null, missing: undefined) {",
			"    return `SELECT id FROM Contact WHERE a = '${nothing}' AND b = '${missing}'`;",
			"}",
		}},
		// Identifiers built from constants are literal types.
		{"a constant identifier", []string{
			"const table = 'Contact';",
			"export function exists() {",
			"    return `SELECT name FROM sqlite_master WHERE type = 'table' AND name = '${table}'`;",
			"}",
		}},
		// Object types are not reported; their text is the runtime's.
		{"a Date", []string{
			"export function since(cutoff: Date) {",
			"    return `SELECT id FROM Contact WHERE createdAt > '${cutoff}'`;",
			"}",
		}},
		{"placeholders", []string{
			"export function find(table: 'Contact', placeholders: string) {",
			"    return `SELECT id FROM ${table} WHERE name LIKE ? AND id IN (${placeholders}) AND email = $1`;",
			"}",
		}},
		{"an identifier in backquotes and double quotes", []string{
			"export function count(table: string, column: string) {",
			"    return `SELECT COUNT(*) AS c FROM \\`${table}\\` WHERE \"${column}\" IS NOT NULL`;",
			"}",
		}},
		{"a value outside quotes", []string{
			"export function find(where: string) {",
			"    return `SELECT id FROM Contact WHERE ${where} AND name = 'x'`;",
			"}",
		}},
		// The quote before it closed `'it''s'`, doubled quote and all.
		{"a value after a closed string", []string{
			"export function find(clause: string) {",
			"    return `SELECT id FROM Contact WHERE a = 'it''s' AND ${clause} AND c = 'd'`;",
			"}",
		}},
		{"a value inside a line comment", []string{
			"export function find(note: string) {",
			"    return `SELECT id FROM Contact -- note = '${note}'",
			"        WHERE name = 'x'`;",
			"}",
		}},
		{"a value inside a block comment", []string{
			"export function find(note: string) {",
			"    return `SELECT id FROM Contact /* note = '${note}' */ WHERE name = 'x'`;",
			"}",
		}},
		// Read as code, the apostrophe would open a string around `${clause}`.
		{"an apostrophe in a line comment before an unquoted value", []string{
			"export function find(clause: string) {",
			"    return `SELECT id FROM Contact -- don't widen",
			"        WHERE ${clause} AND c = 'd'`;",
			"}",
		}},
		// Read as code, `= '` inside the double quotes would open a string around `${clause}`.
		{"double-quoted text holding quotes around an unquoted value", []string{
			"export function find(clause: string) {",
			"    return `SELECT \"a = 'b\" FROM Contact WHERE ${clause} AND c = \"'\" ORDER BY id`;",
			"}",
		}},
		{"backquoted names holding quotes around an unquoted value", []string{
			"export function find(clause: string) {",
			"    return `SELECT \\`a = 'b\\` FROM Contact WHERE ${clause} AND \\`'\\` = 1 ORDER BY id`;",
			"}",
		}},
		// The tag decides what an interpolation becomes.
		{"a tagged template", []string{
			"declare function sql(strings: TemplateStringsArray, ...values: unknown[]): string;",
			"export function find(name: string) {",
			"    return sql`SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}},
		// Missed by decision; see the doc comment.
		{"lowercase SQL", []string{
			"export function find(name: string) {",
			"    return `select id from Contact where name = '${name}'`;",
			"}",
		}},
		// `AhraOsLifecycleCommandLineInterface.ts` in ahra.
		{"English prose with a lowercase keyword", []string{
			"export function refuse(name: string, count: number) {",
			"    return `delete refused: '${name}' has ${count} child position(s)`;",
			"}",
		}},
		// Not a statement: it opens with `Write`.
		{"a prompt that names SQL keywords", []string{
			"export function prompt(table: string) {",
			"    return `Write one query. Use ORDER BY and LIMIT. The table name = '${table}'.`;",
			"}",
		}},
		// Keywords count in uppercase only, the first one included.
		{"a prompt opening with a capitalized keyword", []string{
			"export function prompt(topic: string) {",
			"    return `Update the filter. Join terms with AND or OR. The topic = '${topic}'.`;",
			"}",
		}},
		// `from` is lowercase, so `SELECT` stands alone.
		{"a shouted sentence with one keyword", []string{
			"export function prompt(option: string) {",
			"    return `SELECT '${option}' from the menu`;",
			"}",
		}},
		// Two keywords make it a statement, but `wrote` does not take a value.
		{"a prompt whose quote is not in a value position", []string{
			"export function prompt(input: string) {",
			"    return `SELECT one answer FROM the list below. The user wrote '${input}'.`;",
			"}",
		}},
		// A fragment whose other half this rule cannot see.
		{"a string that closes in another template", []string{
			"export function open(name: string) {",
			"    return `WHERE name = '${name}` + `'`;",
			"}",
		}},
		// Where the string ends depends on the engine.
		{"a backslash inside a string", []string{
			"export function find(name: string) {",
			"    return `SELECT id FROM Contact WHERE a = 'x\\\\' AND name = '${name}'`;",
			"}",
		}},
		// A comment on MySQL only, so where the next string starts depends on the engine.
		{"a hash outside a string", []string{
			"export function find(clause: string) {",
			"    return `SELECT id FROM Contact # a = 'b",
			"        WHERE ${clause} AND c = 'd`;",
			"}",
		}},
		{"dollar-quoted strings holding quotes around an unquoted value", []string{
			"export function find(clause: string) {",
			"    return `SELECT $$a = '$$ FROM Contact WHERE ${clause} AND b = $$'$$ ORDER BY id`;",
			"}",
		}},
		// MySQL reads `a--1` as arithmetic, so `'x` opens a string and `${name}` is outside it.
		{"two dashes with no space", []string{
			"export function find(name: string) {",
			"    return `SELECT a--1, 'x",
			"        WHERE name = '${name}'`;",
			"}",
		}},
		// A fragment needs a second keyword.
		{"a LIKE in a log line", []string{
			"export function log(pattern: string) {",
			"    return `Matching names LIKE '${pattern}'`;",
			"}",
		}},
		// An odd number of quotes stands before the `LIKE`.
		{"a LIKE inside another quoted text", []string{
			"export function command(pattern: string) {",
			"    return `echo 'WHERE name LIKE '${pattern}' OR all'`;",
			"}",
		}},
		{"a LIKE pattern that never closes", []string{
			"export function clause(word: string) {",
			"    return `(r.ZFIRSTNAME LIKE '%${word}% OR r.ZLASTNAME IS NULL`;",
			"}",
		}},
		{"a LIKE pattern holding a backslash", []string{
			"export function clause(word: string) {",
			"    return `(r.ZFIRSTNAME LIKE '%\\\\_${word}%' OR r.ZLASTNAME IS NULL)`;",
			"}",
		}},
		// Not a statement: the leading `import` keeps Python's quotes from being read as SQL.
		{"Python code assigning a quoted value", []string{
			"export function script(name: string) {",
			"    return `import sqlite3",
			"name = '${name}'",
			"rows = conn.execute(\"SELECT * FROM handle WHERE id = ?\", (name,))",
			"`;",
			"}",
		}},
		{"a complete escape inline", []string{
			"export function find(name: string) {",
			"    return `SELECT id FROM Contact WHERE name = '${name.replace(/\\\\/g, '\\\\\\\\').replace(/'/g, \"''\")}'`;",
			"}",
		}},
		{"a complete escape with quotes first, through replaceAll", []string{
			"export function find(input: string) {",
			"    const name = input.replaceAll(\"'\", \"''\").replaceAll('\\\\', '\\\\\\\\');",
			"    return `SELECT id FROM Contact WHERE name = '${name}'`;",
			"}",
		}},
		// `DriveApi.ts` in ahra: not SQL, no statement keyword.
		{"Google Drive's query language", []string{
			"export function query(searchTerm: string) {",
			"    return `name contains '${searchTerm}'`;",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			result := rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, securityNoInterpolatedSqlStringSource(testCase.lines...))
			rule_testing.ExpectClean(t, result)
		})
	}
}

// The rule declares the checker and declines a file without one, rather than reporting every number
// and literal union.
func TestSecurityNoInterpolatedSqlStringDeclinesWithoutAChecker(t *testing.T) {
	t.Parallel()
	if !SecurityNoInterpolatedSqlString.NeedsTypeChecker {
		t.Fatal("the rule must declare NeedsTypeChecker")
	}
	if listeners := SecurityNoInterpolatedSqlString.Run(rule.Context{}, nil); listeners != nil {
		t.Errorf("with no checker the rule must register no listeners, got %d", len(listeners))
	}
}

// securityNoInterpolatedSqlValueWording is the bare-value message, written out here.
const securityNoInterpolatedSqlValueWording = "This value is written into a SQL statement without quotes, right after a comparison, so it becomes " +
	"part of the SQL rather than a value: a space, a keyword or a quote in it changes what the query does, " +
	"and a crafted value runs as SQL. Quoting it does not help, since the value can close the quote. Pass it " +
	"as a bound parameter (`WHERE id = ?` with the value beside the query). Where the query language has no " +
	"parameters, as with GAQL, give the value a type that holds only fixed text: a number, a union of " +
	"literals, or a template literal type such as `${number}`."

// securityNoInterpolatedSqlValueFinding is one expected finding: its id and the source it points at.
type securityNoInterpolatedSqlValueFinding struct {
	id   string
	span string
}

func securityNoInterpolatedSqlValueExpect(t *testing.T, result rule_testing.Result, sourceText string, want []securityNoInterpolatedSqlValueFinding) {
	t.Helper()
	wantIds := make([]string, 0, len(want))
	for _, finding := range want {
		wantIds = append(wantIds, finding.id)
	}
	rule_testing.ExpectFindings(t, result, wantIds...)
	reported := securityNoInterpolatedSqlStringReported(result, sourceText)
	for index, finding := range want {
		if reported[index] != finding.span {
			t.Fatalf("finding %d points at %q, want %q", index, reported[index], finding.span)
		}
		wording := securityNoInterpolatedSqlStringWording
		if finding.id == securityNoInterpolatedSqlValueId {
			wording = securityNoInterpolatedSqlValueWording
		}
		if got := result.Diagnostics[index].Message.Description; got != wording {
			t.Fatalf("finding %d reads %q", index, got)
		}
	}
}

// The bare-value branch: an unquoted interpolation right after a comparison, in a confirmed statement,
// whose type can hold any text. The first case is the GAQL lookup in ahra's
// `modules/google/ads/GoogleAdsCampaignApi.ts` before #kx29bm3 typed its ids as numbers.
func TestSecurityNoInterpolatedSqlValueFires(t *testing.T) {
	t.Parallel()

	value := func(span string) securityNoInterpolatedSqlValueFinding {
		return securityNoInterpolatedSqlValueFinding{securityNoInterpolatedSqlValueId, span}
	}
	cases := []struct {
		name  string
		lines []string
		want  []securityNoInterpolatedSqlValueFinding
	}{
		{"a GAQL id typed string, from the command line", []string{
			"export function campaignDetails(campaignId: string) {",
			"    return `SELECT campaign.id, campaign.name FROM campaign WHERE campaign.id = ${campaignId}`;",
			"}",
		}, []securityNoInterpolatedSqlValueFinding{value("campaignId")}},
		{"every comparison", []string{
			"export function filter(a: string, b: string, c: string, d: string, e: string, f: string, g: string) {",
			"    return `SELECT id FROM t WHERE a = ${a} AND b <> ${b} AND c != ${c} AND d < ${d} AND e > ${e} AND f <= ${f} AND g >= ${g}`;",
			"}",
		}, []securityNoInterpolatedSqlValueFinding{value("a"), value("b"), value("c"), value("d"), value("e"), value("f"), value("g")}},
		{"a literal built by a helper that doubles only the quote", []string{
			"declare function sqlLiteral(value: string): string;",
			"export function count(table: string, watermark: string) {",
			"    return `SELECT COUNT(*) AS c FROM ${table} WHERE updated_at <= ${sqlLiteral(watermark)}`;",
			"}",
		}, []securityNoInterpolatedSqlValueFinding{value("sqlLiteral(watermark)")}},
		{"any, unknown, a branded string, a template with a text hole, an open type parameter", []string{
			"type EmailType = string & { brand: 'Email' };",
			"export function lookup<T>(a: any, b: unknown, c: EmailType, d: `user-${string}`, e: T) {",
			"    return `SELECT id FROM t WHERE a = ${a} AND b = ${b} AND c = ${c} AND d = ${d} AND e = ${e}`;",
			"}",
		}, []securityNoInterpolatedSqlValueFinding{value("a"), value("b"), value("c"), value("d"), value("e")}},
		{"a quoted value and a bare one, each under its own id", []string{
			"export function find(name: string, identifier: string) {",
			"    return `SELECT id FROM Contact WHERE name = '${name}' AND id = ${identifier}`;",
			"}",
		}, []securityNoInterpolatedSqlValueFinding{{securityNoInterpolatedSqlStringId, "name"}, value("identifier")}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sourceText := securityNoInterpolatedSqlStringSource(testCase.lines...)
			result := rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, sourceText)
			securityNoInterpolatedSqlValueExpect(t, result, sourceText, testCase.want)
		})
	}
}

// What the bare-value branch leaves alone: values that hold only fixed text, unquoted interpolations
// that build SQL from code rather than compare a value, and templates that are not a statement.
func TestSecurityNoInterpolatedSqlValueStaysSilent(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name  string
		lines []string
	}{
		{"a GAQL id typed as a number, after #kx29bm3", []string{
			"export function campaignDetails(campaignId: number) {",
			"    return `SELECT campaign.id FROM campaign WHERE campaign.id = ${campaignId}`;",
			"}",
		}},
		{"a GAQL enum from a union of literals, which GAQL writes unquoted", []string{
			"export function byStatus(status: 'ENABLED' | 'PAUSED') {",
			"    return `SELECT campaign.id FROM campaign WHERE campaign.status = ${status}`;",
			"}",
		}},
		{"a numeric template type, a bigint and a boolean", []string{
			"export function byDate(date: `${number}-${number}-${number}`, big: bigint, flag: boolean) {",
			"    return `SELECT id FROM t WHERE day = ${date} AND total > ${big} AND active = ${flag}`;",
			"}",
		}},
		{"placeholders built for an IN list", []string{
			"export function byIds(placeholders: string) {",
			"    return `SELECT id FROM t WHERE id IN (${placeholders})`;",
			"}",
		}},
		{"a condition, a table, a column and a limit built from code", []string{
			"export function query(condition: string, table: string, column: string, limit: string) {",
			"    return `SELECT id FROM ${table} WHERE id > 0 AND ${condition} ORDER BY ${column} LIMIT ${limit}`;",
			"}",
		}},
		{"a log line that is not a statement", []string{
			"export function note(identifier: string) {",
			"    return `Updated campaign id = ${identifier}`;",
			"}",
		}},
		{"lowercase SQL", []string{
			"export function find(identifier: string) {",
			"    return `select id from t where id = ${identifier}`;",
			"}",
		}},
		{"a tagged template, whose tag decides what the value becomes", []string{
			"declare function sql(strings: TemplateStringsArray, ...values: unknown[]): string;",
			"export function find(identifier: string) {",
			"    return sql`SELECT id FROM t WHERE id = ${identifier}`;",
			"}",
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			sourceText := securityNoInterpolatedSqlStringSource(testCase.lines...)
			rule_testing.ExpectClean(t, rule_testing.RunTyped(t, SecurityNoInterpolatedSqlString, securityNoInterpolatedSqlStringFile, sourceText))
		})
	}
}
