import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `consistency-no-iso-string-date-cut` both ways. The unit cases parse a source string and hand the rule symbols built
 by position, one occurrence for each symbol a case names at every token spelled that way, so each case says what
 the compiler would have resolved; a name no case resolves is a name the index did not record. Locals get no
 symbol, as the build's index records none. A key `"<line>:<name>"` resolves that name on that line only. The one
 site the survey found, in Presence, is modeled first as it stands and repaired, then the ISO 8601 sites of both
 proving grounds that render whole timestamps, each left alone as it stands and flagged once cut, then every
 other shape the rule flags and every shape it leaves alone. The end-to-end case builds a real package, so the
 symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct ConsistencyNoIsoStringDateCutTests {
    static let style = ConsistencyNoIsoStringDateCut.styleMemberPrefix
    /* Read from a real build's index: Foundation's `iso8601`, the style's initializer, and the standard library's cuts. */
    static let iso8601 = "s:10Foundation11FormatStylePA2A4DateV07ISO8601bC0VRszrlE7iso8601AGvpZ"
    static let styleInitializer =
        "s:10Foundation4DateV18ISO8601FormatStyleV13dateSeparator0f4TimeG004timeG00i4ZoneG026includingFractionalSeconds0iJ0A2E0bG0O_AE0bhG0OAE0hG0OAE0hjG0OSbAA0hJ0Vtcfc"
    static let prefix = "s:SlsE6prefixy11SubSequenceQzSiF"
    static let split =
        "s:SlsSQ7ElementRpzrlE5split9separator9maxSplits25omittingEmptySubsequencesSay11SubSequenceQzGAB_SiSbtF"
    static let first = "s:SlsE5first7ElementQzSgvp"
    static let components = "s:Sy10FoundationE10components11separatedBySaySSGqd___tSyRd__lF"
    /* `DateComponents.ISO8601FormatStyle`'s chain, spelled the same as `Date`'s. */
    static let componentsYear = "s:10Foundation14DateComponentsV18ISO8601FormatStyleV4yearAEyF"
    static let componentsMonth = "s:10Foundation14DateComponentsV18ISO8601FormatStyleV5monthAEyF"
    static let componentsDay = "s:10Foundation14DateComponentsV18ISO8601FormatStyleV3dayAEyF"
    /* An `iso8601` and a `prefix` of ours. */
    static let ourIso8601 = "s:15AhraOSPresence4DateV0aB0E7iso8601AEvpZ"
    static let ourPrefix = "s:15AhraOSPresence6StringV0aB0E6prefixySSSiF"

    /* What a typical case resolves: Foundation's style and formatter, and the standard library's cuts. */
    static let foundation: [String: [String]] = [
        "formatted": [ConsistencyNoIsoStringDateCut.dateFormatted],
        "ISO8601Format": [ConsistencyNoIsoStringDateCut.dateIso8601Format],
        "format": [ConsistencyNoIsoStringDateCut.styleFormat],
        "iso8601": [iso8601],
        "year": [style + "4yearAEyF"],
        "month": [style + "5monthAEyF"],
        "day": [style + "3dayAEyF"],
        "weekOfYear": [style + "10weekOfYearAEyF"],
        "time": [style + "4time26includingFractionalSecondsAESb_tF"],
        "timeZone": [style + "8timeZone9separatorA2E04TimeG9SeparatorO_tF"],
        "dateSeparator": [style + "13dateSeparatoryA2E0bG0OF"],
        "timeSeparator": [style + "13timeSeparatoryA2E04TimeG0OF"],
        "ISO8601FormatStyle": [style, styleInitializer],
        "ISO8601DateFormatter": [
            "c:objc(cs)NSISO8601DateFormatter", ConsistencyNoIsoStringDateCut.formatterInitializer,
        ],
        "string": [ConsistencyNoIsoStringDateCut.formatterString],
        "formatOptions": [ConsistencyNoIsoStringDateCut.formatterOptions],
        "prefix": [prefix],
        "split": [split],
        "first": [first],
        "components": [components],
    ]

    /* Every finding the rule makes on a source, its names resolved as the suite's comment describes. */
    static func records(_ source: String, resolving: [String: [String]] = foundation) -> [FindingRecord] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        var occurrences: [FileSymbols.Occurrence] = []
        for token in file.tree.tokens(viewMode: .sourceAccurate) {
            let location = file.locations.location(for: token.positionAfterSkippingLeadingTrivia)
            /* A name a pattern declares is the declaration, as the index records it; anywhere else it is a reference. */
            let isReference = !(token.parent?.is(IdentifierPatternSyntax.self) ?? false)
            for symbol in resolving["\(location.line):\(token.text)"] ?? resolving[token.text] ?? [] {
                occurrences.append(
                    FileSymbols.Occurrence(
                        line: location.line,
                        column: location.column,
                        symbol: symbol,
                        name: token.text,
                        isReference: isReference,
                    )
                )
            }
        }
        let rule = ConsistencyNoIsoStringDateCut()
        let found = rule.findings(in: file, symbols: FileSymbols(occurrences, ownedModules: ["AhraOSPresence"]))
        #expect(found.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        return found
    }

    /* Every finding as the text of its span and its message id. */
    static func findings(_ source: String, resolving: [String: [String]] = foundation) -> [String] {
        let lines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return records(source, resolving: resolving).map { finding in
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
            guard let endLine = finding.endLine, let endColumn = finding.endColumn, endLine == finding.line else {
                return "spans lines"
            }
            let span = String(decoding: lines[finding.line - 1][(finding.column - 1)..<(endColumn - 1)], as: UTF8.self)
            return "\(span) | \(finding.messageId)"
        }
    }

    // MARK: The real sites

    /*
     `ahraos-presence` `Sources/AhraOSPresence/Studio/Designer/DesignerBodiesProbe.swift:125` on 2026-10-03, `order(model:)`:
     each body's added date in the BODY ORDER diagnostic line, the UTC day of a file's local creation time.
     */
    static func bodiesProbe(added: String) -> String {
        """
        import Foundation

        extension DesignerBodiesProbe {
            private static func order(model: StudioModel) {
                let catalog = model.bodies.catalog
                catalog.readIfNeeded()
                let bodies = model.bodies.library.allBodies
                let starred = Set(bodies.map(\\.id).filter { catalog.entry(for: $0).isFavorite })
                func line(_ name: String, _ arranged: [BodyLibrary.Body]) {
                    let first = arranged.prefix(5).map { body -> String in
                        let entry = catalog.entry(for: body.id)
        \(added)
                        return
                            "\\(body.title) [\\(entry.kind ?? "?") added \\(added)"
                            + "\\(starred.contains(body.id) ? " *" : "")]"
                    }
                    presenceLog("presence: BODY ORDER \\(name), \\(arranged.count) shown: \\(first.joined(separator: "; "))")
                }
            }
        }

        """
    }

    @Test func theBodiesProbesAddedDateIsFlagged() {
        let found = Self.findings(
            Self.bodiesProbe(
                added:
                    "                let added = BodyDates.added(body.fileUrl).formatted(.iso8601.year().month().day())"
            )
        )
        #expect(
            found == [
                "BodyDates.added(body.fileUrl).formatted(.iso8601.year().month().day()) | isoDateWithoutTimeZone"
            ],
            "\(found)",
        )
    }

    /* The repair names the zone; `arranged.prefix(5)` on the same file is an array's prefix, never a timestamp's. */
    @Test func theBodiesProbesAddedDateWithItsZoneIsClean() {
        let local =
            "                let added = BodyDates.added(body.fileUrl).formatted(Date.ISO8601FormatStyle(timeZone: .current).year().month().day())"
        #expect(Self.findings(Self.bodiesProbe(added: local)).isEmpty)
        let utc =
            "                let added = BodyDates.added(body.fileUrl).formatted(Date.ISO8601FormatStyle(timeZone: .gmt).year().month().day())"
        #expect(Self.findings(Self.bodiesProbe(added: utc)).isEmpty)
    }

    /* `ahraos-macos` `Sources/AhraOs/MainThreadStallMeter.swift:159`, `isoNow()`: a confined formatter rendering a whole timestamp. */
    static func stallMeter(_ body: String) -> String {
        """
        import Foundation

        extension MainThreadStallMeter {
            private static func isoNow() -> String {
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = [.withInternetDateTime, .withFractionalSeconds]
        \(body)
            }
        }

        """
    }

    @Test func theStallMetersWholeTimestampIsClean() {
        #expect(Self.findings(Self.stallMeter("        return formatter.string(from: Date())")).isEmpty)
    }

    @Test func theStallMetersTimestampCutToItsDayIsFlagged() {
        let found = Self.findings(Self.stallMeter("        return String(formatter.string(from: Date()).prefix(10))"))
        #expect(found == ["formatter.string(from: Date()).prefix(10) | isoStringCutToDate"], "\(found)")
    }

    @Test func theStallMetersCutWithTheFormattersZoneSetIsClean() {
        #expect(
            Self.findings(
                Self.stallMeter(
                    "        formatter.timeZone = .current\n        return String(formatter.string(from: Date()).prefix(10))"
                )
            ).isEmpty
        )
    }

    /* `ahraos-presence` `Tests/LibraryTests/PresenceStore+ParityFreeze.swift:18`: the freeze instant as a whole timestamp, compared with stored ones. */
    static func parityFreeze(_ at: String) -> String {
        """
        import Foundation

        extension PresenceStore {
            func itemsChanged(after freeze: Date) throws -> Set<String> {
        \(at)
                return Set(try self.locked { try self.connection.rows("SELECT row_id FROM acts WHERE at > ?", [at]) })
            }
        }

        """
    }

    @Test func theParityFreezeTimestampIsClean() {
        #expect(
            Self.findings(Self.parityFreeze("        let at = ISO8601DateFormatter().string(from: freeze)")).isEmpty
        )
    }

    @Test func theParityFreezeTimestampCutToItsDayIsFlagged() {
        let found = Self.findings(
            Self.parityFreeze("        let at = ISO8601DateFormatter().string(from: freeze).prefix(10)")
        )
        #expect(found == ["ISO8601DateFormatter().string(from: freeze).prefix(10) | isoStringCutToDate"], "\(found)")
    }

    /* `ahraos-presence` `Hearing/PresenceApplicationDelegate+SongIdentityProbe.swift:74`: an hour ago, written whole into a JSON field. */
    static func songProbe(_ hourAgo: String) -> String {
        """
        import Foundation

        extension PresenceApplicationDelegate {
            func backdate(_ text: inout String, range: Range<String.Index>) {
        \(hourAgo)
                text.replaceSubrange(range, with: "\\"lastHeard\\" : \\"\\(hourAgo)\\"")
            }
        }

        """
    }

    @Test func theSongProbesHourAgoIsClean() {
        #expect(
            Self.findings(
                Self.songProbe(
                    "        let hourAgo = ISO8601DateFormatter().string(from: Date().addingTimeInterval(-3600))"
                )
            ).isEmpty
        )
    }

    @Test func theSongProbesHourAgoSplitToItsDayIsFlagged() {
        let found = Self.findings(
            Self.songProbe(
                "        let hourAgo = ISO8601DateFormatter().string(from: Date().addingTimeInterval(-3600)).split(separator: \"T\").first ?? \"\""
            )
        )
        #expect(
            found == [
                "ISO8601DateFormatter().string(from: Date().addingTimeInterval(-3600)).split(separator: \"T\").first | isoStringCutToDate"
            ],
            "\(found)",
        )
    }

    /* `ahraos-macos` `Sources/AhraOsServices/TraceLog.swift:37`: a whole timestamp style held in a static and formatted by `format(_:)`. */
    @Test func theTraceLogsTimestampIsClean() {
        let source = """
            import Foundation

            enum TraceLog {
                private static let timestamp = Date.ISO8601FormatStyle(includingFractionalSeconds: true)

                static func line(_ resolved: String, now: Date) -> String {
                    "\\(timestamp.format(now)) [services] \\(resolved)\\n"
                }
            }

            """
        #expect(Self.findings(source).isEmpty)
    }

    // MARK: Flagged

    @Test(arguments: [
        (
            "date.formatted(.iso8601.year().month().day())",
            "date.formatted(.iso8601.year().month().day()) | isoDateWithoutTimeZone",
        ),
        (
            "date.formatted(.iso8601.year().month())",
            "date.formatted(.iso8601.year().month()) | isoDateWithoutTimeZone",
        ),
        (
            "date.formatted(.iso8601.year().weekOfYear().day())",
            "date.formatted(.iso8601.year().weekOfYear().day()) | isoDateWithoutTimeZone",
        ),
        (
            "date.formatted(.iso8601.day().year().month())",
            "date.formatted(.iso8601.day().year().month()) | isoDateWithoutTimeZone",
        ),
        (
            "date.formatted(.iso8601.dateSeparator(.omitted).year().month().day())",
            "date.formatted(.iso8601.dateSeparator(.omitted).year().month().day()) | isoDateWithoutTimeZone",
        ),
        (
            "date.formatted(.iso8601.year().month().day().timeSeparator(.omitted))",
            "date.formatted(.iso8601.year().month().day().timeSeparator(.omitted)) | isoDateWithoutTimeZone",
        ),
        ("date.formatted((.iso8601.year()))", "date.formatted((.iso8601.year())) | isoDateWithoutTimeZone"),
        (
            "date.formatted(Date.ISO8601FormatStyle.iso8601.year().month().day())",
            "date.formatted(Date.ISO8601FormatStyle.iso8601.year().month().day()) | isoDateWithoutTimeZone",
        ),
        (
            "date.formatted(Date.ISO8601FormatStyle().year().month().day())",
            "date.formatted(Date.ISO8601FormatStyle().year().month().day()) | isoDateWithoutTimeZone",
        ),
        (
            "date.formatted(Date.ISO8601FormatStyle(dateSeparator: .omitted, includingFractionalSeconds: true).year().month().day())",
            "date.formatted(Date.ISO8601FormatStyle(dateSeparator: .omitted, includingFractionalSeconds: true).year().month().day()) | isoDateWithoutTimeZone",
        ),
        (
            "date.ISO8601Format(.iso8601.year().month().day())",
            "date.ISO8601Format(.iso8601.year().month().day()) | isoDateWithoutTimeZone",
        ),
        (
            "Date.ISO8601FormatStyle().year().month().day().format(date)",
            "Date.ISO8601FormatStyle().year().month().day().format(date) | isoDateWithoutTimeZone",
        ),
        (
            "ISO8601DateFormatter().string(from: date).prefix(10)",
            "ISO8601DateFormatter().string(from: date).prefix(10) | isoStringCutToDate",
        ),
        (
            "ISO8601DateFormatter().string(from: date).prefix(7)",
            "ISO8601DateFormatter().string(from: date).prefix(7) | isoStringCutToDate",
        ),
        (
            "ISO8601DateFormatter().string(from: date).prefix(1)",
            "ISO8601DateFormatter().string(from: date).prefix(1) | isoStringCutToDate",
        ),
        (
            "(ISO8601DateFormatter().string(from: date)).prefix(10)",
            "(ISO8601DateFormatter().string(from: date)).prefix(10) | isoStringCutToDate",
        ),
        (
            "Foundation.ISO8601DateFormatter().string(from: date).prefix(10)",
            "Foundation.ISO8601DateFormatter().string(from: date).prefix(10) | isoStringCutToDate",
        ),
        ("date.ISO8601Format().prefix(10)", "date.ISO8601Format().prefix(10) | isoStringCutToDate"),
        ("date.formatted(.iso8601).prefix(10)", "date.formatted(.iso8601).prefix(10) | isoStringCutToDate"),
        (
            "ISO8601DateFormatter().string(from: date).split(separator: \"T\")[0]",
            "ISO8601DateFormatter().string(from: date).split(separator: \"T\")[0] | isoStringCutToDate",
        ),
        (
            "ISO8601DateFormatter().string(from: date).components(separatedBy: \"T\").first",
            "ISO8601DateFormatter().string(from: date).components(separatedBy: \"T\").first | isoStringCutToDate",
        ),
        (
            "ISO8601DateFormatter().string(from: date).components(separatedBy: \"T\")[0]",
            "ISO8601DateFormatter().string(from: date).components(separatedBy: \"T\")[0] | isoStringCutToDate",
        ),
        (
            "date.ISO8601Format().split(separator: \"T\").first",
            "date.ISO8601Format().split(separator: \"T\").first | isoStringCutToDate",
        ),
    ])
    func flagsTheShape(expression: String, expected: String) {
        let source = """
            import Foundation

            func render(_ date: Date) -> Any {
                \(expression)
            }

            """
        let found = Self.findings(source)
        #expect(found == [expected], "\(found)")
    }

    /* A source for each shape, from the cases below: a cut, a date-only style, a confined date-only formatter. */
    static let everyMessageSources = [
        """
        import Foundation

        func render(_ date: Date) -> Any {
            ISO8601DateFormatter().string(from: date).prefix(10)
        }

        """,
        """
        import Foundation

        func render(_ date: Date) -> Any {
            date.formatted(.iso8601.year().month().day())
        }

        """,
        """
        import Foundation

        func dayStamp(_ date: Date) -> String {
            let formatter = ISO8601DateFormatter()
            formatter.formatOptions = [.withFullDate]
            return formatter.string(from: date)
        }

        """,
    ]

    /* Each shape's message, checked against a literal rather than the rule's own text. */
    @Test func eachShapeNamesItsRenderingAndTheRepair() {
        let found = Self.everyMessageSources.flatMap { Self.records($0) }
        #expect(
            found.map(\.message) == [
                "This cuts an ISO 8601 timestamp down to its date part. ISO 8601 formatting is UTC unless it is given a time zone, so the cut is the UTC calendar day whether or not that was the day meant, and in Utah the UTC day turns over at 5 pm in winter and 6 pm in summer: an evening run gets tomorrow. Format the day itself and say the zone: Date.ISO8601FormatStyle(timeZone: .current).year().month().day() when the local day is meant, or timeZone: .gmt when the UTC day is.",
                "This formats a date as its ISO 8601 calendar day without naming a time zone. Date.ISO8601FormatStyle is UTC unless it is given one, so this is the UTC day whether or not that was the day meant, and in Utah the UTC day turns over at 5 pm in winter and 6 pm in summer: an evening run gets tomorrow. Say the zone: Date.ISO8601FormatStyle(timeZone: .current) when the local day is meant, or timeZone: .gmt when the UTC day is.",
                "This formats a date as its ISO 8601 calendar day without naming a time zone. ISO8601DateFormatter is UTC unless its timeZone is set, so this is the UTC day whether or not that was the day meant, and in Utah the UTC day turns over at 5 pm in winter and 6 pm in summer: an evening run gets tomorrow. Say the zone: set the formatter's timeZone to .current when the local day is meant, or to .gmt when the UTC day is.",
            ]
        )
        #expect(found.map(\.rule) == Array(repeating: "cohere-swift/consistency-no-iso-string-date-cut", count: 3))
    }

    /*
     The finding a user sees is the catalog's entry (policy/messages/consistency-no-iso-string-date-cut.json), id and
     text, so an edited entry reaches the finding once regenerated. The literals above pin the words; this pins where
     they come from.
     */
    @Test func eachFindingIsTheCatalogsMessage() {
        let found = Self.everyMessageSources.flatMap { Self.records($0) }
        let messages = [
            RuleMessages.ConsistencyNoIsoStringDateCut.isoStringCutToDate(),
            RuleMessages.ConsistencyNoIsoStringDateCut.isoDateWithoutTimeZone(renderer: .formatStyle),
            RuleMessages.ConsistencyNoIsoStringDateCut.isoDateWithoutTimeZone(renderer: .dateFormatter),
        ]
        #expect(found.map(\.messageId) == messages.map(\.id))
        #expect(found.map(\.message) == messages.map(\.text))
    }

    /* A timestamp held in a local `let`, cut later, as the TypeScript rule follows a `const` (its migration stamp). */
    @Test func aTimestampHeldInALocalLetIsFollowed() {
        let source = """
            import Foundation

            func migrationStamp() -> String {
                let stamp = ISO8601DateFormatter().string(from: Date())
                let datePart = stamp.prefix(10)
                let timePart = stamp.dropFirst(11).prefix(8)
                return "\\(datePart)-\\(timePart)"
            }

            """
        let found = Self.findings(source)
        #expect(found == ["stamp.prefix(10) | isoStringCutToDate"], "\(found)")
    }

    /* Two `let`s deep, and through a closure that reads the outer one. */
    @Test func aTimestampHeldTwiceAndReadInAClosureIsFollowed() {
        let source = """
            import Foundation

            func days(_ dates: [Date]) -> [Substring] {
                let now = Date().ISO8601Format()
                let stamp = now
                return dates.map { _ in stamp.prefix(10) }
            }

            """
        let found = Self.findings(source)
        #expect(found == ["stamp.prefix(10) | isoStringCutToDate"], "\(found)")
    }

    /* A property `let` of this file, found through the index's declaration of its name. */
    @Test func aTimestampHeldInAPropertyLetIsFollowed() {
        let source = """
            import Foundation

            struct Snapshot {
                let takenAt = Date().ISO8601Format()

                var day: Substring {
                    takenAt.prefix(10)
                }
            }

            """
        var resolving = Self.foundation
        resolving["takenAt"] = ["s:15AhraOSPresence8SnapshotV7takenAtSSvp"]
        let found = Self.findings(source, resolving: resolving)
        #expect(found == ["takenAt.prefix(10) | isoStringCutToDate"], "\(found)")
    }

    /* A confined formatter whose options are a date alone renders the UTC day without any cut. */
    @Test(arguments: [
        "[.withFullDate]",
        ".withFullDate",
        "[.withYear, .withMonth, .withDay, .withDashSeparatorInDate]",
        "[.withYear, .withMonth]",
        "[.withFullDate, .withFractionalSeconds]",
    ])
    func aDateOnlyFormatterIsFlagged(options: String) {
        let source = """
            import Foundation

            func dayStamp(_ date: Date) -> String {
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = \(options)
                return formatter.string(from: date)
            }

            """
        let found = Self.findings(source)
        #expect(found == ["formatter.string(from: date) | isoDateWithoutTimeZone"], "\(found)")
    }

    /* A confined formatter with its default options, cut, and used twice. */
    @Test func aConfinedFormatterWithDefaultOptionsIsReadWhenCut() {
        let source = """
            import Foundation

            func window(from start: Date, to end: Date) -> String {
                let formatter = ISO8601DateFormatter()
                let first = formatter.string(from: start).prefix(10)
                let last = formatter.string(from: end)
                return "\\(first)..\\(last)"
            }

            """
        let found = Self.findings(source)
        #expect(found == ["formatter.string(from: start).prefix(10) | isoStringCutToDate"], "\(found)")
    }

    /* A space between date and time: the date still opens the text, but there is no `T` to split on. */
    @Test func aSpaceSeparatedTimestampIsCutByPrefixButNotSplit() {
        let source = """
            import Foundation

            func stamps(_ date: Date) -> [Any] {
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = [.withInternetDateTime, .withSpaceBetweenDateAndTime]
                return [formatter.string(from: date).prefix(10), formatter.string(from: date).split(separator: "T").first as Any]
            }

            """
        let found = Self.findings(source)
        #expect(found == ["formatter.string(from: date).prefix(10) | isoStringCutToDate"], "\(found)")
    }

    // MARK: Left alone

    @Test(arguments: [
        /* The zone said. */
        "date.formatted(Date.ISO8601FormatStyle(timeZone: .current).year().month().day())",
        "date.formatted(Date.ISO8601FormatStyle(dateSeparator: .dash, timeZone: .gmt).year().month().day())",
        "date.formatted(.iso8601Date(timeZone: .current))",
        "ISO8601DateFormatter.string(from: date, timeZone: .current, formatOptions: [.withFullDate])",
        /* A decoded style carries whatever zone it was encoded with. */
        "try? Date.ISO8601FormatStyle(from: decoder).year().format(date)",
        /* Whole timestamps, and renderings that name their zone. */
        "date.formatted(.iso8601)",
        "date.ISO8601Format()",
        "ISO8601DateFormatter().string(from: date)",
        "date.formatted(.iso8601.dateSeparator(.omitted))",
        "date.formatted(.iso8601.year().month().day().time(includingFractionalSeconds: false))",
        "date.formatted(.iso8601.year().month().day().timeZone(separator: .colon))",
        "date.ISO8601Format(.iso8601.time(includingFractionalSeconds: true))",
        /* Cuts past the date part, empty, computed, of the time half, or after the string was touched. */
        "ISO8601DateFormatter().string(from: date).prefix(11)",
        "ISO8601DateFormatter().string(from: date).prefix(16)",
        "ISO8601DateFormatter().string(from: date).prefix(0)",
        "ISO8601DateFormatter().string(from: date).prefix(length)",
        "ISO8601DateFormatter().string(from: date).suffix(10)",
        "ISO8601DateFormatter().string(from: date).split(separator: \"T\")[1]",
        "ISO8601DateFormatter().string(from: date).split(separator: \"T\").last",
        "ISO8601DateFormatter().string(from: date).split(separator: \".\").first",
        "ISO8601DateFormatter().string(from: date).split(separator: \"T\", maxSplits: 1).first",
        "ISO8601DateFormatter().string(from: date).lowercased().prefix(10)",
        "ISO8601DateFormatter().string(from: date).replacingOccurrences(of: \"T\", with: \" \").prefix(19)",
        "date.formatted(.iso8601.year().month().day().time(includingFractionalSeconds: false)).prefix(10)",
        /* Not an ISO timestamp, and not Foundation's style. */
        "\"2026-10-03T00:00:00Z\".prefix(10)",
        "date.description.prefix(10)",
        "components.formatted(.iso8601.year().month().day())",
        "try? Date(text, strategy: .iso8601.year().month().day())",
    ])
    func leavesAlone(expression: String) {
        let source = """
            import Foundation

            func render(_ date: Date, components: DateComponents, decoder: any Decoder, text: String, length: Int) -> Any {
                \(expression)
            }

            """
        var resolving = Self.foundation
        /* On this fixture `components` is the parameter; the `DateComponents` case's chain is that type's. */
        if expression.hasPrefix("components.") {
            resolving["year"] = [Self.componentsYear]
            resolving["month"] = [Self.componentsMonth]
            resolving["day"] = [Self.componentsDay]
            resolving["formatted"] = [
                "s:10Foundation14DateComponentsV9formattedy12FormatOutputQzxAA0E5StyleRzAC0E5InputRtzlF"
            ]
        }
        #expect(Self.findings(source, resolving: resolving).isEmpty)
    }

    /* `DateComponents.ISO8601FormatStyle`'s chain with `Date`'s `formatted`: the style is still not `Date`'s. */
    @Test func aDateComponentsStyleIsNeverDates() {
        var resolving = Self.foundation
        resolving["year"] = [Self.componentsYear]
        resolving["month"] = [Self.componentsMonth]
        resolving["day"] = [Self.componentsDay]
        #expect(
            Self.findings(
                "import Foundation\nlet day = components.formatted(.iso8601.year().month().day())\n",
                resolving: resolving,
            ).isEmpty
        )
    }

    /* Names the index did not record, or resolved to declarations of ours. */
    @Test(arguments: [
        ("formatted", [String]()),
        ("iso8601", [String]()),
        ("iso8601", [ourIso8601]),
        ("year", [String]()),
    ])
    func aStyleTheIndexDoesNotPlaceIsLeftAlone(name: String, symbols: [String]) {
        var resolving = Self.foundation
        resolving[name] = symbols
        #expect(
            Self.findings(
                "import Foundation\nlet day = date.formatted(.iso8601.year().month().day())\n",
                resolving: resolving,
            ).isEmpty
        )
    }

    @Test(arguments: [
        ("ISO8601DateFormatter", ["c:objc(cs)NSISO8601DateFormatter"]),
        ("string", [String]()),
        ("prefix", [String]()),
        ("prefix", [ourPrefix]),
        ("ISO8601Format", [String]()),
    ])
    func aCutTheIndexDoesNotPlaceIsLeftAlone(name: String, symbols: [String]) {
        var resolving = Self.foundation
        resolving[name] = symbols
        let source =
            "import Foundation\nlet first = ISO8601DateFormatter().string(from: date).prefix(10)\nlet second = date.ISO8601Format().prefix(10)\n"
        let found = Self.findings(source, resolving: resolving)
        /* Removing one name silences the line that needs it and only that line. */
        let expected =
            [
                ("ISO8601DateFormatter", ["date.ISO8601Format().prefix(10) | isoStringCutToDate"]),
                ("string", ["date.ISO8601Format().prefix(10) | isoStringCutToDate"]),
                ("prefix", []),
                ("ISO8601Format", ["ISO8601DateFormatter().string(from: date).prefix(10) | isoStringCutToDate"]),
            ].first { $0.0 == name }?.1 ?? []
        #expect(found == expected, "\(found)")
    }

    /* A style held before it formats is a known miss. */
    @Test func aStyleHeldInABindingIsAKnownMiss() {
        let source = """
            import Foundation

            func day(_ date: Date) -> String {
                let style = Date.ISO8601FormatStyle().year().month().day()
                return date.formatted(style)
            }

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Every way a formatter's zone or options can change out of the rule's sight. */
    @Test(arguments: [
        /* The zone set. */
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions = [.withFullDate]\nformatter.timeZone = .current\nreturn formatter.string(from: date)",
        /* Handed to another function, which may set it. */
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions = [.withFullDate]\nconfigure(formatter)\nreturn formatter.string(from: date)",
        /* Read in an interpolation, captured by name. */
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions = [.withFullDate]\nlet run = { [formatter] in formatter.string(from: date) }\nreturn run()",
        /* Two assignments. */
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions = [.withFullDate]\nformatter.formatOptions = [.withInternetDateTime]\nreturn formatter.string(from: date)",
        /* Assigned in a branch. */
        "let formatter = ISO8601DateFormatter()\nif short {\n    formatter.formatOptions = [.withFullDate]\n}\nreturn formatter.string(from: date)",
        /* Used before its options are set. */
        "let formatter = ISO8601DateFormatter()\nlet before = formatter.string(from: date)\nformatter.formatOptions = [.withFullDate]\nreturn before",
        /* Options the rule does not read. */
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions = options\nreturn formatter.string(from: date)",
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions.insert(.withFullDate)\nreturn formatter.string(from: date)",
        /* A zone field: the rendering names its zone. */
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions = [.withFullDate, .withTimeZone]\nreturn formatter.string(from: date)",
        /* A `var`. */
        "var formatter = ISO8601DateFormatter()\nformatter.formatOptions = [.withFullDate]\nreturn formatter.string(from: date)",
        /* Shadowed by a closure parameter of the same name. */
        "let formatter = ISO8601DateFormatter()\nformatter.formatOptions = [.withFullDate]\nreturn [other].map { formatter in formatter.string(from: date) }.joined()",
    ])
    func aFormatterTheFileDoesNotFullyShowIsLeftAlone(body: String) {
        let indented = body.split(separator: "\n").map { "    \($0)" }.joined(separator: "\n")
        let source = """
            import Foundation

            func dayStamp(_ date: Date, short: Bool, options: ISO8601DateFormatter.Options, other: ISO8601DateFormatter) -> String {
            \(indented)
            }

            """
        #expect(Self.findings(source).isEmpty, "\(Self.findings(source))")
    }

    /* The closure-initialized static formatter: another file may set its zone, so it is a known miss. */
    @Test func aClosureInitializedStaticFormatterIsAKnownMiss() {
        let source = """
            import Foundation

            enum Days {
                static let formatter: ISO8601DateFormatter = {
                    let formatter = ISO8601DateFormatter()
                    formatter.formatOptions = [.withFullDate]
                    return formatter
                }()

                static func day(_ date: Date) -> String {
                    formatter.string(from: date)
                }
            }

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Bindings the rule does not follow: a `var`, an `if let`, a parameter that shadows, and a top-level formatter, which is a global. */
    @Test func bindingsThatAreNotALetOfThisScopeAreLeftAlone() {
        let source = """
            import Foundation

            let globalFormatter = ISO8601DateFormatter()

            func stamps(_ date: Date, optional: String?) -> [Any] {
                var stamp = ISO8601DateFormatter().string(from: date)
                stamp += "!"
                let cut = stamp.prefix(10)
                var results: [Any] = [cut]
                if let held = optional {
                    results.append(held.prefix(10))
                }
                let shadowed = ISO8601DateFormatter().string(from: date)
                func inner(shadowed: String) -> Substring { shadowed.prefix(10) }
                results.append(inner(shadowed: "x"))
                globalFormatter.formatOptions = [.withFullDate]
                results.append(globalFormatter.string(from: date))
                return results
            }

            """
        #expect(Self.findings(source).isEmpty, "\(Self.findings(source))")
    }

    /* A formatter declared and configured in `main.swift`'s top-level code is a global, which another file may give a zone. */
    @Test func aTopLevelFormatterIsAGlobalAndLeftAlone() {
        let source = """
            import Foundation

            let dayFormatter = ISO8601DateFormatter()
            dayFormatter.formatOptions = [.withFullDate]
            print(dayFormatter.string(from: Date()))

            """
        #expect(Self.findings(source).isEmpty, "\(Self.findings(source))")
    }

    /* A property `let` the index places but that holds no timestamp, and a property of another file the index names but this file does not declare. */
    @Test func aPropertyThatIsNotATimestampOrNotHereIsLeftAlone() {
        let source = """
            import Foundation

            struct Snapshot {
                let label = "2026-10-03T00:00:00Z"

                var day: Substring { label.prefix(10) }
                var other: Substring { elsewhere.prefix(10) }
            }

            """
        var resolving = Self.foundation
        resolving["label"] = ["s:15AhraOSPresence8SnapshotV5labelSSvp"]
        resolving["elsewhere"] = ["s:15AhraOSPresence5OtherV9elsewhereSSvp"]
        #expect(Self.findings(source, resolving: resolving).isEmpty)
    }

    /* `DateFormatter` is local time unless told otherwise: not this trap. */
    @Test func aDateFormatterIsNotIso8601() {
        let source = """
            import Foundation

            func day(_ date: Date) -> String {
                let formatter = DateFormatter()
                formatter.dateFormat = "yyyy-MM-dd"
                return formatter.string(from: date)
            }

            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func thePrefilterNeedsIso8601() {
        let rule = ConsistencyNoIsoStringDateCut()
        func parsed(_ source: String) -> ParsedFile {
            ParsedFile(
                url: URL(fileURLWithPath: "/fixture/Subject.swift"),
                targetName: "Fixture",
                targetKind: "library",
                source: source,
                tree: Parser.parse(source: source),
                nodeCount: 0,
            )
        }
        #expect(rule.applies(to: parsed("let day = date.formatted(.iso8601.year())")))
        #expect(rule.applies(to: parsed("let stamp = ISO8601DateFormatter()")))
        #expect(!rule.applies(to: parsed("let day = date.formatted(.dateTime.year())")))
    }

    // MARK: End to end

    static let packageSource = #"""
        import Foundation

        public enum Days {
            public static func probe(_ date: Date) -> String {
                date.formatted(.iso8601.year().month().day())
            }

            public static func local(_ date: Date) -> String {
                date.formatted(Date.ISO8601FormatStyle(timeZone: .current).year().month().day())
            }

            public static func cut(_ date: Date) -> String {
                String(ISO8601DateFormatter().string(from: date).prefix(10))
            }

            public static func held(_ date: Date) -> String {
                let stamp = date.ISO8601Format()
                return String(stamp.split(separator: "T").first ?? "")
            }

            public static func dayOnly(_ date: Date) -> String {
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = [.withFullDate]
                return formatter.string(from: date)
            }

            public static func zoned(_ date: Date) -> String {
                let formatter = ISO8601DateFormatter()
                formatter.formatOptions = [.withFullDate]
                formatter.timeZone = .current
                return formatter.string(from: date)
            }

            public static func components(_ components: DateComponents) -> String {
                components.formatted(.iso8601.year().month().day())
            }

            public static func parsed(_ text: String) -> Date? {
                try? Date(text, strategy: .iso8601.year().month().day())
            }
        }

        """#

    @Test func theIndexResolvesFoundationsStyleFormatterAndCuts() async throws {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-typed-\(UUID().uuidString)",
            isDirectory: true,
        )
        let sources = root.appendingPathComponent("Sources/Control", isDirectory: true)
        try FileManager.default.createDirectory(at: sources, withIntermediateDirectories: true)
        try PipelineControlTests.manifest.write(
            to: root.appendingPathComponent("Package.swift"),
            atomically: true,
            encoding: .utf8,
        )
        try Self.packageSource.write(
            to: sources.appendingPathComponent("Control.swift"),
            atomically: true,
            encoding: .utf8,
        )

        /* One run builds the package and writes its index. The rule is run by hand on what the run left. */
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix"],
            workingDirectory: root,
        )
        _ = try await Pipeline(options: options, writer: ContractWriter { _ in }, workingDirectory: root).run()
        let package = try PackageModel.load(
            root: root,
            scratchPath: Pipeline.scratchPath(for: root),
            runner: ProcessRunner(),
        )
        let parsed = await SourceParser().parse(try FileSet.build(package: package).owned)
        let rule = ConsistencyNoIsoStringDateCut()
        let candidates = parsed.files.filter { rule.applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            rule.findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.line):\($0.column) \($0.messageId)"
            }
        }
        /* `probe`, `cut`, `held` and `dayOnly`; not `local` or `zoned` (the zone said), `components` (not Date's style) or `parsed` (a parse). */
        #expect(
            found == [
                "5:9 isoDateWithoutTimeZone", "13:16 isoStringCutToDate", "18:23 isoStringCutToDate",
                "24:16 isoDateWithoutTimeZone",
            ],
            "\(found)",
        )
    }
}
