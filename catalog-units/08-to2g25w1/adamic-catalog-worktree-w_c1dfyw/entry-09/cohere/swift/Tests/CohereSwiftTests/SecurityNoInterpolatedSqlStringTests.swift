import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `security-no-interpolated-sql-string` both ways. The reading cases port the TypeScript rule's own fixtures, each
 value a parameter annotated `String` (or a type that cannot hold a quote), with symbols built by position as the
 index records them: every standard library type name resolves to its real symbol, and `replacingOccurrences` to
 Foundation's. The type gate is judged end to end on one real package, built once, so every name resolves as the
 compiler decided: `Store.swift` models Presence's store as it stands (`PresenceStore.swift` and
 `PresenceStore+Undo.swift` and the conformance test, on 2026-10-03), `BoundStore.swift` is the same code repaired
 (each value bound, or escaped in full where it stays in the text), and `Shapes.swift` holds one line per type the
 gate reads. A line that should be reported ends
 in `// fires`; every other line must be silent.
 */
@Suite(.serialized)
struct SecurityNoInterpolatedSqlStringTests {
    static let foundationReplacingOccurrences =
        "s:Sy10FoundationE20replacingOccurrences2of4with7options5rangeSSqd___qd_0_So22NSStringCompareOptionsVSnySS5IndexVGSgtSyRd__SyRd_0_r0_lF"
    static let names = [
        "String": "s:SS", "Substring": "s:Ss", "Character": "s:SJ", "Int": "s:Si",
        "replacingOccurrences": foundationReplacingOccurrences,
    ]

    /* The text of each finding's span, in order, with names resolved as `names` says. */
    static func reported(_ lines: [String], names: [String: String] = names) -> [String] {
        let source = lines.joined(separator: "\n") + "\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        var occurrences: [FileSymbols.Occurrence] = []
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            if let symbol = names[token.text] {
                let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: symbol,
                        name: token.text,
                        isReference: true,
                    )
                )
            }
        }
        let sourceLines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return SecurityNoInterpolatedSqlString().findings(in: file, symbols: FileSymbols(occurrences)).map { finding in
            guard finding.line == finding.endLine, let endColumn = finding.endColumn else {
                return "spans more than line \(finding.line)"
            }
            return String(
                decoding: sourceLines[finding.line - 1][(finding.column - 1)..<(endColumn - 1)],
                as: UTF8.self,
            )
        }
    }

    @Test(
        arguments: [
            (
                "a LIKE pattern in a SELECT",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "SELECT name FROM ZABCDRECORD WHERE name LIKE '%\(name)%'""#,
                    "}",
                ], ["name"],
            ),
            /* Both values sit in a `VALUES` list, after `(` and after `,`. */
            (
                "values in an INSERT",
                [
                    #"func insert(name: String, email: String) -> String {"#,
                    #"    "INSERT INTO Contact (name, email) VALUES ('\(name)', '\(email)')""#,
                    "}",
                ], ["name", "email"],
            ),
            (
                "a BETWEEN range",
                [
                    #"func range(start: String, end: String) -> String {"#,
                    #"    "SELECT id FROM campaign WHERE segments.date BETWEEN '\(start)' AND '\(end)'""#,
                    "}",
                ], ["start", "end"],
            ),
            (
                "an UPDATE assignment",
                [
                    #"func rename(name: String, identifier: Int) -> String {"#,
                    #"    "UPDATE Contact SET name = '\(name)' WHERE id = \(identifier)""#,
                    "}",
                ], ["name"],
            ),
            /* `$.` inside the string is text, and the string opens after a `,`. */
            (
                "a JSON path argument",
                [
                    #"func extract(field: String) -> String {"#,
                    #"    "SELECT JSON_EXTRACT(data, '$.\(field)') FROM EngagementEvent""#,
                    "}",
                ], ["field"],
            ),
            /* `WHERE` needs no second keyword. */
            (
                "a WHERE clause on its own",
                [
                    #"func filter(status: String) -> String {"#,
                    #"    "WHERE campaign.status = '\(status)'""#,
                    "}",
                ], ["status"],
            ),
            (
                "an AND clause continuing a query",
                [
                    #"func clause(accountId: String) -> String {"#,
                    #"    " AND accountId = '\(accountId)' ORDER BY createdAt""#,
                    "}",
                ], ["accountId"],
            ),
            (
                "two interpolations in one string",
                [
                    #"func find(first: String, last: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name LIKE '%\(first)\(last)%'""#,
                    "}",
                ], ["first", "last"],
            ),
            /* The `''` is one quote inside the string, so the value is still inside it. */
            (
                "a string holding a doubled quote",
                [
                    #"func find(suffix: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name = 'O''\(suffix)'""#,
                    "}",
                ], ["suffix"],
            ),
            /* The apostrophe in the comment does not open a string. */
            (
                "a comment before the value, in a multi-line literal",
                [
                    #"func find(name: String) -> String {"#,
                    #"    """"#,
                    #"    SELECT id FROM Contact -- the user's search"#,
                    #"    WHERE name = '\(name)'"#,
                    #"    """"#,
                    "}",
                ], ["name"],
            ),
            /* `\'` is a quote once the literal is read as the program sees it. */
            (
                "quotes written as escapes",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name = \'\(name)\'""#,
                    "}",
                ], ["name"],
            ),
            (
                "a raw string",
                [
                    #"func find(name: String) -> String {"#,
                    ##"    #"SELECT id FROM Contact WHERE name = '\#(name)'"#"##,
                    "}",
                ], ["name"],
            ),
            (
                "the members of TypeScript's string, any and unknown",
                [
                    #"func find(part: Substring, letter: Character, value: Any, maybe: String?) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE a = '\(part)' AND b = '\(letter)' AND c = '\(value)' AND d = '\(maybe ?? "")'""#,
                    "}",
                ], ["part", "letter", "value", #"maybe ?? """#],
            ),
            /* Correct on MySQL, open on SQLite. */
            (
                "quotes escaped with a backslash",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name = '\(name.replacingOccurrences(of: "'", with: "\\'"))'""#,
                    "}",
                ], [#"name.replacingOccurrences(of: "'", with: "\\'")"#],
            ),
            /* Correct on SQLite, open on MySQL, and the text does not say which runs it. */
            (
                "quotes doubled alone",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name = '\(name.replacingOccurrences(of: "'", with: "''"))'""#,
                    "}",
                ], [#"name.replacingOccurrences(of: "'", with: "''")"#],
            ),
            /* Only a `let` is followed to its initializer. */
            (
                "an escape held by a var",
                [
                    #"func find(input: String) -> String {"#,
                    #"    var name: String = input.replacingOccurrences(of: "\\", with: "\\\\").replacingOccurrences(of: "'", with: "''")"#,
                    #"    name = input"#,
                    #"    return "SELECT id FROM Contact WHERE name = '\(name)'""#,
                    "}",
                ], ["name"],
            ),
            (
                "a NOT LIKE fragment",
                [
                    #"func clause(word: String) -> String {"#,
                    #"    "(r.ZFIRSTNAME NOT LIKE '%\(word)%' OR r.ZLASTNAME IS NULL)""#,
                    "}",
                ], ["word"],
            ),
            /* `ContactsApi.ts` in ahra: a fragment keyed on `LIKE`, three times. */
            (
                "the contact search's fragment",
                [
                    #"func clause(word: String) -> String {"#,
                    #"    "(LOWER(r.ZFIRSTNAME) LIKE '%\(word)%' OR LOWER(r.ZLASTNAME) LIKE '%\(word)%' OR LOWER(r.ZORGANIZATION) LIKE '%\(word)%')""#,
                    "}",
                ], ["word", "word", "word"],
            ),
        ] as [(String, [String], [String])]
    )
    func fires(name: String, lines: [String], spans: [String]) {
        #expect(Self.reported(lines) == spans, "\(name)")
    }

    @Test(
        arguments: [
            (
                "numbers and booleans",
                [
                    #"func find(identifier: Int, limit: Int64, ratio: Double, active: Bool) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE id = '\(identifier)' AND rank < '\(limit)' AND r = '\(ratio)' AND active = '\(active)'""#,
                    "}",
                ],
            ),
            /* `PresenceStore.swift:126`: `ATTACH` is not a value keyword in the TypeScript rule. */
            (
                "an ATTACH of a path",
                [
                    #"func attach(payload: String) -> String {"#,
                    #"    "PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000; PRAGMA synchronous = NORMAL; ATTACH '\(payload)' AS payload;""#,
                    "}",
                ],
            ),
            (
                "a compile-time name",
                [
                    #"func exists() -> String {"#,
                    #"    let table = "Contact""#,
                    #"    return "SELECT name FROM sqlite_master WHERE type = 'table' AND name = '\(table)'""#,
                    "}",
                ],
            ),
            (
                "placeholders and names outside quotes",
                [
                    #"func find(table: String, placeholders: String) -> String {"#,
                    #"    "SELECT id FROM \(table) WHERE name LIKE ? AND id IN (\(placeholders)) AND email = $1""#,
                    "}",
                ],
            ),
            (
                "an identifier in backquotes and double quotes",
                [
                    #"func count(table: String, column: String) -> String {"#,
                    #"    "SELECT COUNT(*) AS c FROM `\(table)` WHERE \"\(column)\" IS NOT NULL""#,
                    "}",
                ],
            ),
            (
                "a value after a closed string",
                [
                    #"func find(clause: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE a = 'it''s' AND \(clause) AND c = 'd'""#,
                    "}",
                ],
            ),
            (
                "a value inside a line comment",
                [
                    #"func find(note: String) -> String {"#,
                    #"    """"#,
                    #"    SELECT id FROM Contact -- note = '\(note)'"#,
                    #"    WHERE name = 'x'"#,
                    #"    """"#,
                    "}",
                ],
            ),
            (
                "a value inside a block comment",
                [
                    #"func find(note: String) -> String {"#,
                    #"    "SELECT id FROM Contact /* note = '\(note)' */ WHERE name = 'x'""#,
                    "}",
                ],
            ),
            /* Read as code, the apostrophe would open a string around the clause. */
            (
                "an apostrophe in a line comment before an unquoted value",
                [
                    #"func find(clause: String) -> String {"#,
                    #"    """"#,
                    #"    SELECT id FROM Contact -- don't widen"#,
                    #"    WHERE \(clause) AND c = 'd'"#,
                    #"    """"#,
                    "}",
                ],
            ),
            (
                "double-quoted text holding quotes around an unquoted value",
                [
                    #"func find(clause: String) -> String {"#,
                    #"    "SELECT \"a = 'b\" FROM Contact WHERE \(clause) AND c = \"'\" ORDER BY id""#,
                    "}",
                ],
            ),
            (
                "lowercase SQL",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "select id from Contact where name = '\(name)'""#,
                    "}",
                ],
            ),
            (
                "English prose with a lowercase keyword",
                [
                    #"func refuse(name: String, count: Int) -> String {"#,
                    #"    "delete refused: '\(name)' has \(count) child position(s)""#,
                    "}",
                ],
            ),
            (
                "a prompt that names SQL keywords",
                [
                    #"func prompt(table: String) -> String {"#,
                    #"    "Write one query. Use ORDER BY and LIMIT. The table name = '\(table)'.""#,
                    "}",
                ],
            ),
            (
                "a prompt opening with a capitalized keyword",
                [
                    #"func prompt(topic: String) -> String {"#,
                    #"    "Update the filter. Join terms with AND or OR. The topic = '\(topic)'.""#,
                    "}",
                ],
            ),
            (
                "a shouted sentence with one keyword",
                [
                    #"func prompt(option: String) -> String {"#,
                    #"    "SELECT '\(option)' from the menu""#,
                    "}",
                ],
            ),
            (
                "a prompt whose quote is not in a value position",
                [
                    #"func prompt(input: String) -> String {"#,
                    #"    "SELECT one answer FROM the list below. The user wrote '\(input)'.""#,
                    "}",
                ],
            ),
            (
                "a string that closes in another literal",
                [
                    #"func open(name: String) -> String {"#,
                    #"    "WHERE name = '\(name)" + "'""#,
                    "}",
                ],
            ),
            (
                "a backslash inside a string",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE a = 'x\\' AND name = '\(name)'""#,
                    "}",
                ],
            ),
            (
                "a hash outside a string",
                [
                    #"func find(clause: String) -> String {"#,
                    #"    """"#,
                    #"    SELECT id FROM Contact # a = 'b"#,
                    #"    WHERE \(clause) AND c = 'd"#,
                    #"    """"#,
                    "}",
                ],
            ),
            (
                "dollar-quoted strings holding quotes around an unquoted value",
                [
                    #"func find(clause: String) -> String {"#,
                    #"    "SELECT $$a = '$$ FROM Contact WHERE \(clause) AND b = $$'$$ ORDER BY id""#,
                    "}",
                ],
            ),
            (
                "two dashes with no space",
                [
                    #"func find(name: String) -> String {"#,
                    #"    """"#,
                    #"    SELECT a--1, 'x"#,
                    #"    WHERE name = '\(name)'"#,
                    #"    """"#,
                    "}",
                ],
            ),
            (
                "a LIKE in a log line",
                [
                    #"func log(pattern: String) -> String {"#,
                    #"    "Matching names LIKE '\(pattern)'""#,
                    "}",
                ],
            ),
            (
                "a LIKE inside another quoted text",
                [
                    #"func command(pattern: String) -> String {"#,
                    #"    "echo 'WHERE name LIKE '\(pattern)' OR all'""#,
                    "}",
                ],
            ),
            (
                "a LIKE pattern that never closes",
                [
                    #"func clause(word: String) -> String {"#,
                    #"    "(r.ZFIRSTNAME LIKE '%\(word)% OR r.ZLASTNAME IS NULL""#,
                    "}",
                ],
            ),
            (
                "a LIKE pattern holding a backslash",
                [
                    #"func clause(word: String) -> String {"#,
                    #"    "(r.ZFIRSTNAME LIKE '%\\_\(word)%' OR r.ZLASTNAME IS NULL)""#,
                    "}",
                ],
            ),
            (
                "a complete escape inline",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name = '\(name.replacingOccurrences(of: "\\", with: "\\\\").replacingOccurrences(of: "'", with: "''"))'""#,
                    "}",
                ],
            ),
            (
                "a complete escape with quotes first, through a let",
                [
                    #"func find(input: String) -> String {"#,
                    #"    let name = input.replacingOccurrences(of: "'", with: "''").replacingOccurrences(of: "\\", with: "\\\\")"#,
                    #"    return "SELECT id FROM Contact WHERE name = '\(name)'""#,
                    "}",
                ],
            ),
            /* A custom interpolation decides what a label means. */
            (
                "an interpolation with a label",
                [
                    #"func find(name: String) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name = '\(name, privacy: .public)'""#,
                    "}",
                ],
            ),
            (
                "a type the index did not name",
                [
                    #"func find(name: Label) -> String {"#,
                    #"    "SELECT id FROM Contact WHERE name = '\(name)'""#,
                    "}",
                ],
            ),
            (
                "a closure parameter with no annotation, shadowing a String",
                [
                    #"func find(name: String, names: [Int]) -> [String] {"#,
                    #"    names.map { name in "SELECT id FROM Contact WHERE name = '\(name)'" }"#,
                    "}",
                ],
            ),
            (
                "a case pattern shadowing a String",
                [
                    #"func find(name: String, value: Value) -> String {"#,
                    #"    switch value {"#,
                    #"    case .text(let name): return "SELECT id FROM Contact WHERE name = '\(name)'""#,
                    #"    default: return """#,
                    #"    }"#,
                    "}",
                ],
            ),
            (
                "Google Drive's query language",
                [
                    #"func query(searchTerm: String) -> String {"#,
                    #"    "name contains '\(searchTerm)'""#,
                    "}",
                ],
            ),
        ] as [(String, [String])]
    )
    func staysSilent(name: String, lines: [String]) {
        #expect(Self.reported(lines).isEmpty, "\(name): \(Self.reported(lines))")
    }

    /*
     An enum's own `rawValue` is one of its raw values, a closed set, as TypeScript's literal union; a struct's is any
     string, as TypeScript's branded one. Declared outside the package here, so neither is a property this file could
     hold.
     */
    @Test func anEnumsRawValueIsAClosedSetAndAStructsIsNot() {
        let lines = [
            #"func find(kind: Kind) -> String {"#,
            #"    "SELECT id FROM items WHERE kind = '\(kind.rawValue)'""#,
            "}",
        ]
        let enumeration = Self.names.merging(["rawValue": "s:8External4KindO8rawValueSSvp"]) { $1 }
        #expect(Self.reported(lines, names: enumeration).isEmpty)
        let structure = Self.names.merging(["rawValue": "s:8External4NameV8rawValueSSvp"]) { $1 }
        #expect(Self.reported(lines, names: structure) == ["kind.rawValue"])
    }

    /* The finding names its repair and carries no fix. */
    @Test func theFindingSaysWhatToDoAndFixesNothing() throws {
        let source = "func find(name: String) -> String { \"SELECT id FROM Contact WHERE name = '\\(name)'\" }\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let location = file.locations.location(
            for: try #require(file.tree.tokens(viewMode: .sourceAccurate).first { $0.text == "String" })
                .positionAfterSkippingLeadingTrivia
        )
        let symbols = FileSymbols([
            FileSymbols.Occurrence(
                line: location.line,
                column: location.column,
                symbol: "s:SS",
                name: "String",
                isReference: true,
            )
        ])
        let finding = try #require(SecurityNoInterpolatedSqlString().findings(in: file, symbols: symbols).first)
        #expect(finding.rule == "cohere-swift/security-no-interpolated-sql-string")
        #expect(finding.messageId == "interpolatedSqlString")
        #expect(finding.message.contains("parameter"))
        #expect(!finding.message.contains("\u{2014}"))
        #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty)
    }

    @Test func aFileWithNoQuotedInterpolationIsNotFetched() {
        let source = "let greeting = \"Hello, \\(name)\"\nlet sql = \"SELECT id FROM Contact WHERE name = ?\"\n"
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Plain.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        #expect(!SecurityNoInterpolatedSqlString().applies(to: file))
    }

    static let manifest = """
        // swift-tools-version:6.2
        import PackageDescription

        let package = Package(
            name: "Control",
            platforms: [.macOS(.v15)],
            targets: [
                .target(
                    name: "Control",
                    swiftSettings: [.enableUpcomingFeature("ExistentialAny"), .enableUpcomingFeature("MemberImportVisibility"), .strictMemorySafety()]
                )
            ]
        )

        """

    /* Presence's store as it stands: the connection's shape, the value it reads back, and the four statements that quote a value. */
    static let store = """
        import Foundation

        enum SQLValue: Equatable {
            case null
            case integer(Int64)
            case text(String)

            var text: String? {
                if case .text(let text) = self { return text }
                return nil
            }

            var integer: Int64? {
                if case .integer(let integer) = self { return integer }
                return nil
            }
        }

        struct SQLRow {
            var values: [String: SQLValue]

            subscript(_ column: String) -> SQLValue { values[column] ?? .null }
        }

        final class SQLiteConnection {
            func execute(_ sql: String) throws {}

            @discardableResult
            func rows(_ sql: String, _ parameters: [SQLValue] = []) throws -> [SQLRow] { [] }

            func value(_ sql: String, _ parameters: [SQLValue] = []) throws -> SQLValue { .null }
        }

        final class Store {
            static let kindTables = ["Body": "bodies", "Part": "parts", "Motion": "motions"]
            static let derivedTables = ["followers", "variants"]
            let connection = SQLiteConnection()
            let title: String
            var columns: [String: [String]] = [:]

            init(root: URL) throws {
                self.title = root.lastPathComponent
                let payload = root.appendingPathComponent("Library/motions.db").path.replacingOccurrences(of: "'", with: "''")
                try self.connection.execute(
                    "PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000; PRAGMA synchronous = NORMAL; ATTACH '\\(payload)' AS payload;"
                )
                for table in ["items"] + Array(Self.kindTables.values) + Self.derivedTables {
                    self.columns[table] = try self.connection.rows("SELECT name FROM pragma_table_info('\\(table)')").compactMap { $0["name"].text } // fires
                }
            }

            func verify() throws -> [String] {
                var problems: [String] = []
                let count = { (sql: String) throws -> Int64 in try self.connection.value(sql).integer ?? 0 }
                for (kind, table) in Self.kindTables.sorted(by: { $0.key < $1.key }) {
                    let missing = try count(
                        "SELECT count(*) FROM items WHERE kind = '\\(kind)' AND id NOT IN (SELECT id FROM \\(table))") // fires
                    if missing > 0 { problems.append("\\(missing) \\(kind) items have no \\(table) row") }
                }
                return problems
            }

            func writeBack(_ row: SQLRow) throws {
                guard let table = row["table_name"].text, let operation = row["op"].text else { return }
                let keys = try self.connection.rows("SELECT name FROM pragma_table_info('\\(table)') WHERE pk > 0 ORDER BY pk") // fires
                    .compactMap { $0["name"].text }
                if operation == "Insert" {
                    try self.connection.rows("DELETE FROM \\(table) WHERE \\(keys.map { "\\($0) = ?" }.joined(separator: " AND "))")
                }
            }

            func create(_ kind: String) throws -> String { kind.lowercased() }
        }

        func bothFacesFindAnAsset(store: Store) throws {
            let path = "Genshin/Motions/Clip.vrma"
            for location in ["Store", "Dataset", "Support"] {
                _ = try store.connection.rows("SELECT '\\(location)' AS location, '\\(path)' AS path") // fires
            }
            _ = try store.connection.rows("SELECT 'Dataset' AS location, '\\(path)' AS path")
            let id = try store.create("Body")
            try store.connection.execute("UPDATE items SET favorite = 1 WHERE id = '\\(id)'") // fires
        }

        """

    /* The same statements repaired as the message says: each value bound, and the last escaped in full where it stays in the text. */
    static let boundStore = """
        import Foundation

        extension Store {
            func columnsBound() throws {
                for table in ["items"] + Array(Self.kindTables.values) + Self.derivedTables {
                    self.columns[table] = try self.connection.rows("SELECT name FROM pragma_table_info(?)", [.text(table)]).compactMap { $0["name"].text }
                }
            }

            func verifyBound() throws -> Int64 {
                var total: Int64 = 0
                for (kind, table) in Self.kindTables.sorted(by: { $0.key < $1.key }) {
                    total += try self.connection.value("SELECT count(*) FROM items WHERE kind = ? AND id NOT IN (SELECT id FROM \\(table))", [.text(kind)]).integer ?? 0
                }
                return total
            }

            func writeBackBound(_ row: SQLRow) throws {
                guard let table = row["table_name"].text else { return }
                try self.connection.rows("SELECT name FROM pragma_table_info(?) WHERE pk > 0 ORDER BY pk", [.text(table)])
            }
        }

        func bothFacesFindAnAssetBound(store: Store) throws {
            for location in ["Store", "Dataset", "Support"] {
                _ = try store.connection.rows("SELECT ? AS location, ? AS path", [.text(location), .text("Genshin/Motions/Clip.vrma")])
            }
            let id = try store.create("Body")
            try store.connection.execute("UPDATE items SET favorite = 1 WHERE id = '\\(id.replacingOccurrences(of: "\\\\", with: "\\\\\\\\").replacingOccurrences(of: "'", with: "''"))'")
        }

        """

    /* One line per type the gate reads. */
    static let shapes = """
        import Foundation

        enum Kind: String {
            case body = "Body"
            case part = "Part"
        }

        struct Name: RawRepresentable {
            let rawValue: String
        }

        final class Shapes {
            static let table = "items"
            var mode = "Store"
            var nickname: String?
            let connection = SQLiteConnection()

            func label() -> String { "label" }
            func number() -> Int { 1 }

            func fires(name: String, part: Substring, letter: Character, value: Any, maybe: String?, url: URL, wrapped: Name, names: [String], counts: [String: Int]) throws {
                var mutable = "Body"
                mutable += "s"
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(name)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(part)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE initial = '\\(letter)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE data = '\\(value)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(maybe ?? "")'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE kind = '\\(mutable)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE label = '\\(self.label())'") // fires
                _ = try self.connection.rows("SELECT id FROM assets WHERE path = '\\(url.path)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE nickname = '\\(self.nickname ?? "")'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(wrapped.rawValue)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE location = '\\(self.mode)'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(name.replacingOccurrences(of: "'", with: "''"))'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(name + "s")'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE data = '\\(String(describing: value))'") // fires
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\("O'Brien")'") // fires
                if let nickname = self.nickname {
                    _ = try self.connection.rows("SELECT id FROM items WHERE nickname = '\\(nickname)'") // fires
                }
                for (index, each) in names.enumerated() {
                    _ = try self.connection.rows("SELECT id FROM items WHERE rank = '\\(index)' AND name = '\\(each)'") // fires
                }
                for key in counts.keys.sorted() {
                    _ = try self.connection.rows("SELECT id FROM items WHERE kind = '\\(key)'") // fires
                }
            }

            func silent(count: Int, size: Int64, ratio: Double, flag: Bool, kind: Kind, names: [String], generic: some CustomStringConvertible, row: SQLValue, store: Store) throws {
                let path = "Genshin/Motions/Clip.vrma"
                let escaped = path.replacingOccurrences(of: "\\\\", with: "\\\\\\\\").replacingOccurrences(of: "'", with: "''")
                _ = try self.connection.rows("SELECT id FROM items WHERE id = '\\(count)' AND size = '\\(size)' AND ratio = '\\(ratio)' AND flag = '\\(flag)'")
                _ = try self.connection.rows("SELECT id FROM items WHERE kind = '\\(kind)' AND raw = '\\(kind.rawValue)'")
                _ = try self.connection.rows("SELECT id FROM assets WHERE path = '\\(path)'")
                _ = try self.connection.rows("SELECT id FROM \\(Self.table) WHERE name = '\\(Self.table)'")
                _ = try self.connection.rows("SELECT id FROM assets WHERE path = '\\(escaped)'")
                _ = try self.connection.rows("SELECT id FROM assets WHERE path = '\\(path.replacingOccurrences(of: "'", with: "''").replacingOccurrences(of: "\\\\", with: "\\\\\\\\"))'")
                _ = try self.connection.rows("SELECT id FROM assets WHERE path = '\\(path.replacing(/\\\\/, with: "\\\\\\\\").replacing("'", with: "''"))'")
                _ = try self.connection.rows("SELECT name FROM items WHERE number = '\\(self.number())'")
                _ = try self.connection.rows("SELECT \\(names.joined(separator: ", ")) FROM items WHERE kind = ?", [.text(kind.rawValue)])
                _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(generic)'")
                _ = try self.connection.rows("SELECT id FROM items WHERE title = '\\(store.title)'")
                _ = names.map { name in "SELECT id FROM items WHERE name = '\\(name)'" }
                switch row {
                case .text(let text):
                    _ = try self.connection.rows("SELECT id FROM items WHERE name = '\\(text)'")
                default:
                    break
                }
                _ = "Could not update floor designation for '\\(path)'"
            }
        }

        """

    static let files = ["Store.swift": store, "BoundStore.swift": boundStore, "Shapes.swift": shapes]

    /* Built once, then each file read back with the symbols the build's index gives it. */
    static func reportedLines() async throws -> [String: [Int]] {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-sql-\(UUID().uuidString)",
            isDirectory: true,
        )
        let sources = root.appendingPathComponent("Sources/Control", isDirectory: true)
        try FileManager.default.createDirectory(at: sources, withIntermediateDirectories: true)
        try manifest.write(to: root.appendingPathComponent("Package.swift"), atomically: true, encoding: .utf8)
        for (name, source) in files {
            try source.write(to: sources.appendingPathComponent(name), atomically: true, encoding: .utf8)
        }
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix", "--types"],
            workingDirectory: root,
        )
        var stream = Data()
        _ = try await Pipeline(options: options, writer: ContractWriter { stream.append($0) }, workingDirectory: root)
            .run()
        let parsed = files.keys.sorted().map { name in
            ParsedFile(
                url: sources.appendingPathComponent(name),
                targetName: "Control",
                targetKind: "regular",
                source: files[name] ?? "",
                tree: Parser.parse(source: files[name] ?? ""),
                nodeCount: 0,
            )
        }
        var provider = SymbolProvider(scratchPaths: [Pipeline.scratchPath(for: root)], runner: ProcessRunner())
        provider.ownedModules = ["Control"]
        let symbols = provider.symbols(for: parsed)
        var lines: [String: [Int]] = [:]
        for file in parsed {
            let fileSymbols = try #require(
                symbols.symbols[file.url.path],
                "no symbols for \(file.url.lastPathComponent): \(symbols.unavailable); the build said \(String(decoding: stream, as: UTF8.self))",
            )
            lines[file.url.lastPathComponent] = SecurityNoInterpolatedSqlString().findings(
                in: file,
                symbols: fileSymbols,
            ).map(\.line)
        }
        return lines
    }

    /* The lines a fixture marks `// fires`, one finding each. */
    static func marked(_ source: String) -> [Int] {
        source.split(separator: "\n", omittingEmptySubsequences: false).enumerated().compactMap { index, line in
            line.hasSuffix("// fires") ? index + 1 : nil
        }
    }

    /*
     Presence's four quoting statements are found (the table names and kinds are `String`s from constants, true
     but harmless today; the conformance test's location and item id), its bound twins are silent, and so is the
     test's `path`, a `let` of a literal with no quote. Each shape in `Shapes.swift` answers as its comment says.
     */
    @Test func theTypeGateReadsWhatTheCompilerDecided() async throws {
        let lines = try await Self.reportedLines()
        #expect(lines["Store.swift"] == Self.marked(Self.store), "Store.swift: \(lines["Store.swift"] ?? [])")
        #expect(lines["BoundStore.swift"] == [], "BoundStore.swift: \(lines["BoundStore.swift"] ?? [])")
        #expect(lines["Shapes.swift"] == Self.marked(Self.shapes), "Shapes.swift: \(lines["Shapes.swift"] ?? [])")
    }
}
