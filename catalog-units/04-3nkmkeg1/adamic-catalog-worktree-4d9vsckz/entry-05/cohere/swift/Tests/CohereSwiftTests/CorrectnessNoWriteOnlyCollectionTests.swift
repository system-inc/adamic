import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `correctness-no-write-only-collection` both ways. The unit cases parse a source string and hand the rule symbols
 built by position, one occurrence at every token a case names, so each case says what the compiler would have
 resolved; a name no case resolves is a name the index did not record. A key `"<line>:<name>"` resolves that name on
 that line only. The survey found no write-only collection on either proving ground, so the real sites are modeled as
 they stand (silent, each for the reason the rule must respect) and with the one change that would make each fire.
 Then every other shape the rule flags and every shape it leaves alone, the TypeScript rule's cases carried over
 where Swift has them. The end-to-end case runs a real package, so the symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct CorrectnessNoWriteOnlyCollectionTests {
    /* The standard library's declarations, as the index names them (read from a build on 2026-10-03). */
    static let standardLibrary: [String: String] = [
        "Array": "s:Sa",
        "Set": "s:Sh",
        "Dictionary": "s:SD",
        "append": "s:Sa6appendyyxnF",
        "insert": "s:Sh6insertySb8inserted_x17memberAfterInserttxnF",
        "update": "s:Sh6update4withxSgxn_tF",
        "remove": "s:Sa6remove2atxSi_tF",
        "removeFirst": "s:SmsE11removeFirst7ElementQzyF",
        "removeLast": "s:SmsSKRzrlE10removeLastyySiF",
        "popLast": "s:SmsSKRzrlE7popLast7ElementSTQzSgyF",
        "popFirst": "s:Sh8popFirstxSgyF",
        "removeAll": "s:Sa9removeAll15keepingCapacityySb_tF",
        "removeSubrange": "s:Sa14removeSubrangeyySnySiGF",
        "replaceSubrange": "s:Sa15replaceSubrange_4withySnySiG_qd__nt7ElementQyd__RszSlRd__lF",
        "reserveCapacity": "s:Sa15reserveCapacityyySiF",
        "formUnion": "s:Sh9formUnionyyqd__n7ElementQyd__RszSTRd__lF",
        "formIntersection": "s:Sh16formIntersectionyyqd__7ElementQyd__RszSTRd__lF",
        "formSymmetricDifference": "s:Sh23formSymmetricDifferenceyyqd__n7ElementQyd__RszSTRd__lF",
        "subtract": "s:Sh8subtractyyqd__7ElementQyd__RszSTRd__lF",
        "updateValue": "s:SD11updateValue_6forKeyq_Sgq_n_xtF",
        "removeValue": "s:SD11removeValue6forKeyq_Sgx_tF",
        "merge": "s:SD5merge_16uniquingKeysWithySDyxq_G_q_q__q_tKXEtKF",
        "sort": "s:SMsSL7ElementRpzrlE4sortyyF",
        "[": "s:SayxSicip",
        "+=": "s:Sa2peoiyySayxGz_ABtFZ",
    ]

    /* Every finding as the name it spans, after checking its message and that it carries no fix. */
    static func findings(_ source: String, resolving: [String: String] = standardLibrary) -> [String] {
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
            guard let symbol = resolving["\(location.line):\(token.text)"] ?? resolving[token.text] else { continue }
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
        let rule = CorrectnessNoWriteOnlyCollection()
        let found = rule.findings(in: file, symbols: FileSymbols(occurrences, ownedModules: ["Control"]))
        #expect(found.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        let lines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return found.map { finding in
            #expect(finding.messageId == "writeOnlyCollection")
            #expect(
                finding.message
                    == "This collection is only ever added to: every use of it puts something in or takes something out, and nothing reads what it holds. The work that fills it is thrown away. Either the code that was meant to read it is missing (a value kept for the result and never copied in), or the collection is dead and should be deleted along with its writes."
            )
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
            guard let endLine = finding.endLine, let endColumn = finding.endColumn, endLine == finding.line else {
                return "spans lines"
            }
            return String(decoding: lines[finding.line - 1][(finding.column - 1)..<(endColumn - 1)], as: UTF8.self)
        }
    }

    // MARK: The real sites

    /*
     `ahraos-macos` `Sources/AhraOsServices/ProcessStatisticsSampling.swift:142` on 2026-10-03,
     `processStatisticsForPane`. A depth-first walk of the process table: `results` is appended and returned, `stack`
     is popped by the loop, and `visited` is Swift's dedupe idiom, whose insert's answer is read. With the guard
     dropped, the insert is a bare statement, nothing ever asks the set, and a cycle in the table walks forever.
     */
    static func processStatistics(visit: String) -> String {
        """
        enum ProcessStatisticsSampling {
            static func processStatisticsForPane(
                shellPid: Int32,
                rowsByPid: [Int32: PsRow],
                childrenByPid: [Int32: [Int32]]
            ) -> [ProcessStatistics] {
                guard shellPid > 0 else { return [] }
                var results: [ProcessStatistics] = []
                var visited = Set<Int32>()
                var stack: [Int32] = [shellPid]
                while let pid = stack.popLast() {
        \(visit)
                    guard let row = rowsByPid[pid] else { continue }
                    if let children = childrenByPid[pid] {
                        stack.append(contentsOf: children)
                    }
                    guard !row.isZombie else { continue }
                    let isShell = pid == shellPid || shellCommandNames.contains(row.command)
                    results.append(
                        ProcessStatistics(
                            pid: row.pid,
                            command: row.command,
                            cpuPercent: row.cpuPercent,
                            residentMemoryBytes: Int64(row.residentKilobytes) * 1024,
                            isShell: isShell
                        ))
                }
                return results
            }
        }

        """
    }

    @Test func processStatisticsAsItStandsReadsTheInsertsAnswer() {
        #expect(
            Self.findings(
                Self.processStatistics(visit: "            guard visited.insert(pid).inserted else { continue }")
            ) == []
        )
    }

    @Test func processStatisticsWithTheGuardDroppedIsFlagged() {
        #expect(Self.findings(Self.processStatistics(visit: "            visited.insert(pid)")) == ["visited"])
    }

    /*
     `ahraos-macos` `Sources/AhraOs/AhraOsWindowModel.swift:1229` on 2026-10-03, `childProcessSummary(from:)`. The
     index of each command's tally is kept so a repeat bumps its count. As it stands the lookup reads the dictionary;
     with the lookup gone, every command is appended afresh and the index nothing asks is the rank-map bug the
     TypeScript rule was found on.
     */
    static func tallies(lookup: Bool) -> String {
        let body =
            lookup
            ? """
                    if let index = tallyIndexByName[command] {
                        tallies[index].count += 1
                    }
                    else {
                        tallyIndexByName[command] = tallies.count
                        tallies.append((name: command, count: 1))
                    }
            """
            : """
                    tallyIndexByName[command] = tallies.count
                    tallies.append((name: command, count: 1))
            """
        return """
            extension AhraOsWindowModel {
                private static func childProcessSummary(from processes: [ProcessStatistics]) -> String? {
                    let childCommands =
                        processes
                        .filter { !$0.isShell }
                        .map { $0.command }
                    guard !childCommands.isEmpty else { return nil }

                    var tallies: [(name: String, count: Int)] = []
                    var tallyIndexByName: [String: Int] = [:]
                    for command in childCommands {
            \(body)
                    }
                    return
                        tallies
                        .map { tally in tally.count > 1 ? "\\(tally.name) (\\(tally.count))" : tally.name }
                        .joined(separator: ", ")
                }
            }

            """
    }

    @Test func talliesAsTheyStandReadTheIndex() {
        #expect(Self.findings(Self.tallies(lookup: true)) == [])
    }

    @Test func talliesWithoutTheLookupFlagTheIndex() {
        #expect(Self.findings(Self.tallies(lookup: false)) == ["tallyIndexByName"])
    }

    /*
     `ahraos-macos` `Sources/AhraOs/KingdomNode.swift:62` on 2026-10-03, `KingdomNode.roots(from:)`. The children are
     grouped through `[parent, default: []].append`, a write through an element that the rule reads as a read (as
     TypeScript reads `get(key).push`), and read again below. `visited` is written by a nested function and read by
     a closure inside it. Without that read it is only written, from inside the nested function.
     */
    static func kingdom(filter: String) -> String {
        """
        extension KingdomNode {
            static func roots(from positions: [KingdomPosition]) -> [KingdomNode] {
                guard !positions.isEmpty else { return [] }
                let byProfileId = Dictionary(positions.map { ($0.profileId, $0) }) { first, _ in first }
                var childrenByParent: [String: [KingdomPosition]] = [:]
                var rootPositions: [KingdomPosition] = []
                for position in positions {
                    if let parent = position.parent, byProfileId[parent] != nil {
                        childrenByParent[parent, default: []].append(position)
                    }
                    else {
                        rootPositions.append(position)
                    }
                }

                var visited: Set<String> = []
                func node(for position: KingdomPosition) -> KingdomNode {
                    visited.insert(position.profileId)
                    let directReports = (childrenByParent[position.profileId] ?? [])
        \(filter)
                        .sorted(by: ordered)
                    let childNodes = directReports.map(node(for:))
                    return KingdomNode(position: position, children: childNodes.isEmpty ? nil : childNodes)
                }

                return rootPositions.sorted(by: ordered).map(node(for:))
            }
        }

        """
    }

    @Test func kingdomAsItStandsReadsVisitedInAClosure() {
        #expect(
            Self.findings(Self.kingdom(filter: "                .filter { !visited.contains($0.profileId) }")) == []
        )
    }

    @Test func kingdomWithoutTheFilterFlagsVisited() {
        #expect(Self.findings(Self.kingdom(filter: "                .filter { $0.isActive }")) == ["visited"])
    }

    /*
     `ahraos-presence` `Libraries/VRMKit/Sources/VRMRealityKit/CustomType/GLTFEntity.swift:137` on 2026-10-03,
     `deformedVertices(movedBy:minimumShare:)`: the dedupe idiom inside a `guard`'s conditions, and `vertices` filled
     and returned. Silent as it stands; with the condition moved to a bare insert after the guard, `seen` is flagged.
     */
    static func deformedVertices(guarded: Bool) -> String {
        let condition = guarded ? ",\n                seen.insert(ObjectIdentifier(mesh.source)).inserted" : ""
        let bare = guarded ? "" : "\n            seen.insert(ObjectIdentifier(mesh.source))"
        return """
            extension GLTFEntity {
                public func deformedVertices(movedBy joints: [Entity], minimumShare: Float = 0.5) -> [(rest: SIMD3<Float>, deformed: SIMD3<Float>)] {
                    flushSkinPoseIfNeeded()
                    var vertices: [(rest: SIMD3<Float>, deformed: SIMD3<Float>)] = []
                    var seen = Set<ObjectIdentifier>()
                    for binding in skinBindings {
                        guard let mesh = binding.deformedMesh, mesh.geometry.isSkinned\(condition)
                        else { continue }\(bare)
                        vertices.append((rest: mesh.rest, deformed: mesh.deformed))
                    }
                    return vertices
                }
            }

            """
    }

    @Test func deformedVerticesAsTheyStandReadTheInsertsAnswer() {
        #expect(Self.findings(Self.deformedVertices(guarded: true)) == [])
    }

    @Test func deformedVerticesWithABareInsertFlagSeen() {
        #expect(Self.findings(Self.deformedVertices(guarded: false)) == ["seen"])
    }

    /*
     `ahraos-presence` `Sources/AhraOSPresence/Studio/Panels/PresenceApplicationDelegate+MotionSearchProbe.swift:72` on
     2026-10-03, `runMotionSearchProbe`, trimmed to the two timing arrays: the one finding on either proving ground.
     Each keystroke appends its flush time to `flushed` and its settle time to `settled`, from a closure inside the
     nested `press`. The summary reads `settled` (its median and maximum, typing and deleting) and never `flushed`, so
     every flush time is collected and thrown away. Fixed, the summary reads both, and the rule is silent.
     */
    static func motionSearchProbe(summarizesFlushed: Bool) -> String {
        let flushedSummary =
            summarizesFlushed
            ? "\n                presenceLog(\"presence: MOTION SEARCH flushed median \\(flushed.prefix(typed).median()) ms\")"
            : ""
        return """
            extension PresenceApplicationDelegate {
                func runMotionSearchProbe(query: String, editor: NSTextView, contentView: NSView, window: NSWindow) {
                    let keystrokes = query.map { Keystroke.type(String($0)) } + query.map { _ in Keystroke.delete }
                    var flushed: [Double] = []
                    var settled: [Double] = []
                    @MainActor func press(_ index: Int) {
                        guard index < keystrokes.count else {
                            let typed = query.count
                            presenceLog("presence: MOTION SEARCH typing median \\(settled.prefix(typed).median()) ms max \\(settled.prefix(typed).max() ?? 0) ms")\(flushedSummary)
                            return
                        }
                        let start = CACurrentMediaTime()
                        switch keystrokes[index] {
                        case .type(let character): editor.insertText(character, replacementRange: editor.selectedRange())
                        case .delete: editor.deleteBackward(nil)
                        }
                        contentView.layoutSubtreeIfNeeded()
                        window.displayIfNeeded()
                        let flush = (CACurrentMediaTime() - start) * 1000
                        Self.atEndOfRunLoopTurn {
                            let settle = (CACurrentMediaTime() - start) * 1000
                            flushed.append(flush)
                            settled.append(settle)
                            presenceLog("presence: MOTION SEARCH flushed \\(flush) ms settled \\(settle) ms")
                            DispatchQueue.main.asyncAfter(deadline: .now() + 0.12) { press(index + 1) }
                        }
                    }
                    press(0)
                }
            }

            """
    }

    @Test func motionSearchProbeAsItStandsFlagsFlushed() {
        #expect(Self.findings(Self.motionSearchProbe(summarizesFlushed: false)) == ["flushed"])
    }

    @Test func motionSearchProbeSummarizingBothIsClean() {
        #expect(Self.findings(Self.motionSearchProbe(summarizesFlushed: true)) == [])
    }

    // MARK: Every flagged shape

    @Test func everyFlaggedShapeIsFound() {
        let cases: [(String, String, [String])] = [
            (
                "an array appended in a loop, TypeScript's first case",
                """
                func collect(rows: [Row]) {
                    var seen: [String] = []
                    for row in rows {
                        seen.append(row.id)
                    }
                }
                """,
                ["seen"],
            ),
            (
                "a set inserted as a loop's only statement",
                """
                func collect(rows: [Row]) -> Int {
                    var seen = Set<String>()
                    for row in rows { seen.insert(row.id) }
                    return rows.count
                }
                """,
                ["seen"],
            ),
            (
                "an array written through its subscript from a closure",
                """
                func collect(rows: [Row]) {
                    var byIndex = Array(repeating: "", count: rows.count)
                    rows.enumerated().forEach { byIndex[$0.offset] = $0.element.id }
                }
                """,
                ["byIndex"],
            ),
            (
                "a dictionary set, updated, removed from and cleared",
                """
                func collect(rows: [Row]) {
                    var pending: [Int: String] = [:]
                    for row in rows {
                        pending[row.rowid] = row.id
                        pending.updateValue(row.id, forKey: row.rowid + 1)
                        pending.removeValue(forKey: row.rowid - 1)
                    }
                    pending.removeAll()
                }
                """,
                ["pending"],
            ),
            (
                "an array inserted into, removed from, popped and shifted",
                """
                func collect() {
                    var queue = ["a", "b", "c"]
                    queue.insert("z", at: 0)
                    queue.remove(at: 1)
                    queue.removeFirst()
                    queue.removeLast()
                    _ = queue.popLast()
                    queue.removeFirst(1)
                    queue.removeLast(1)
                    queue.removeSubrange(0..<1)
                    queue.replaceSubrange(0..<1, with: ["y"])
                    queue.insert(contentsOf: ["x"], at: 0)
                    queue.removeAll(keepingCapacity: true)
                }
                """,
                ["queue"],
            ),
            (
                "a set's other writers",
                """
                func collect(rows: [String], others: Set<String>) {
                    var names: Set = ["a"]
                    names.reserveCapacity(rows.count)
                    names.formUnion(rows)
                    names.formIntersection(others)
                    names.formSymmetricDifference(others)
                    names.subtract(others)
                    names.update(with: "b")
                    names.remove("a")
                    _ = names.popFirst()
                    names.removeFirst()
                }
                """,
                ["names"],
            ),
            (
                "in a closure stored as a property, TypeScript's property initializer",
                """
                final class Collector {
                    let collect = { (rows: [Row]) in
                        var seen = Set<String>()
                        for row in rows { seen.insert(row.id) }
                    }
                }
                """,
                ["seen"],
            ),
            (
                "appended in a single-expression closure: append returns nothing to hand on",
                """
                func collect(rows: [Row]) {
                    var ids: [String] = []
                    rows.forEach { ids.append($0.id) }
                }
                """,
                ["ids"],
            ),
            (
                "+= on an array, Swift's append(contentsOf:)",
                """
                func collect(rows: [Row]) {
                    var ids = [String]()
                    for row in rows {
                        ids += [row.id]
                    }
                }
                """,
                ["ids"],
            ),
            (
                "behind try and await",
                """
                func collect(source: Source) async throws {
                    var ids = Array<String>()
                    try ids.append(contentsOf: source.load())
                    await ids.append(source.next())
                    try await ids.append(contentsOf: source.rest())
                }
                """,
                ["ids"],
            ),
            (
                "a copy of another collection: a Swift collection is a value, so nothing else sees the writes",
                """
                func collect(original: [String]) -> [String] {
                    var copy: [String] = original
                    copy.append("extra")
                    return original
                }
                """,
                ["copy"],
            ),
            (
                "every spelling of the type",
                """
                func collect(rows: [Row]) {
                    var literal = [1, 2]
                    literal.append(3)
                    var keyed = ["a": 1]
                    keyed["b"] = 2
                    var sugar = [Int: String]()
                    sugar[1] = "one"
                    var named = Dictionary<String, Int>()
                    named["a"] = 1
                    var fromRows = Set(rows)
                    fromRows.removeAll()
                    var annotated: Array<Int> = []
                    annotated.append(1)
                    var bare: Dictionary = ["a": 1]
                    bare.removeAll()
                }
                """,
                ["literal", "keyed", "sugar", "named", "fromRows", "annotated", "bare"],
            ),
            (
                "a dictionary written with a default and an array past its end: a trap is not a read",
                """
                func collect(keys: [String]) {
                    var counts: [String: Int] = [:]
                    for key in keys {
                        counts[key, default: 0] = 1
                    }
                    var slots = [Int]()
                    slots[9] = 1
                }
                """,
                ["counts", "slots"],
            ),
            (
                "a value-returning write as the only statement of an if and a switch that are statements",
                """
                func collect(rows: [Row]) {
                    var seen = Set<String>()
                    for row in rows {
                        if row.isActive {
                            seen.insert(row.id)
                        }
                        else if row.isPending {
                            seen.remove(row.id)
                        }
                    }
                    var pending: [Int: String] = [:]
                    for row in rows {
                        switch row.kind {
                        case .open:
                            pending.updateValue(row.id, forKey: row.rowid)
                        default:
                            pending.removeValue(forKey: row.rowid)
                        }
                    }
                }
                """,
                ["seen", "pending"],
            ),
            (
                "written in a defer and a repeat loop",
                """
                func collect(rows: [Row]) {
                    var log: [String] = []
                    defer { log.removeAll() }
                    repeat { log.removeLast() } while rows.isEmpty
                    log.append("start")
                }
                """,
                ["log"],
            ),
            (
                "a backticked name",
                """
                func collect() {
                    var `default`: [Int] = []
                    `default`.append(1)
                }
                """,
                ["`default`"],
            ),
            (
                "inside an initializer, an accessor and a nested closure",
                """
                struct Holder {
                    init(rows: [Row]) {
                        var ids = Set<String>()
                        for row in rows { ids.insert(row.id) }
                    }
                    var total: Int {
                        var seen: [Int] = []
                        seen.append(1)
                        return 1
                    }
                    func run() {
                        Task {
                            var queued: [Int] = []
                            queued.append(2)
                        }
                    }
                }
                """,
                ["ids", "seen", "queued"],
            ),
        ]
        for (label, source, expected) in cases {
            #expect(Self.findings(source) == expected, "\(label)")
        }
    }

    /* Scope, not spelling: the outer collection is only written, and an inner binding of the same name is read. The outer is reported, at its own declaration. */
    @Test func shadowingIsResolvedByScope() {
        let source = """
            func collect(rows: [Row]) {
                var ids: [String] = []
                ids.append("a")
                do {
                    let ids = ["b"]
                    use(ids)
                }
                rows.forEach { ids in use(ids) }
                for ids in rows { use(ids) }
                if let ids = rows.first { use(ids) }
                func inner(ids: [Row]) { use(ids) }
                use(row.ids)
                use(ids: 1)
            }
            """
        #expect(Self.findings(source) == ["ids"])
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let found = CorrectnessNoWriteOnlyCollection().findings(in: file, symbols: FileSymbols([]))
        #expect(
            found.isEmpty,
            "with no symbols, append is not known to be the standard library's: \(found.map(\.line))",
        )
    }

    // MARK: Every safe shape

    @Test func everySafeShapeIsClean() {
        let cases: [(String, String)] = [
            (
                "a set that guards with contains",
                """
                func unique(rows: [Row]) {
                    var seen = Set<String>()
                    for row in rows {
                        if seen.contains(row.id) { continue }
                        seen.insert(row.id)
                        use(row)
                    }
                }
                """,
            ),
            (
                "the dedupe idiom, in an if, a guard and a filter",
                """
                func unique(rows: [Row]) -> [Row] {
                    var seen = Set<String>()
                    var firsts = Set<String>()
                    var kept = Set<String>()
                    for row in rows {
                        if seen.insert(row.id).inserted { use(row) }
                        guard firsts.insert(row.id).inserted else { continue }
                    }
                    return rows.filter { kept.insert($0.id).inserted }
                }
                """,
            ),
            (
                "returned",
                """
                func collect(rows: [Row]) -> [String] {
                    var ids: [String] = []
                    for row in rows { ids.append(row.id) }
                    return ids
                }
                """,
            ),
            (
                "stored in a value",
                """
                func collect(rows: [Row]) -> Summary {
                    var ids: [String] = []
                    for row in rows { ids.append(row.id) }
                    return Summary(ids: ids)
                }
                """,
            ),
            (
                "its count read",
                """
                func count(rows: [Row]) -> Int {
                    var ids: [String] = []
                    for row in rows { ids.append(row.id) }
                    return ids.count
                }
                """,
            ),
            (
                "copied into another, TypeScript's spread",
                """
                func collect() {
                    var ids: [String] = []
                    ids.append("a")
                    use(Array(ids))
                }
                """,
            ),
            (
                "a mutating call whose result is used",
                """
                func collect() -> String? {
                    var ids: [String] = ["a"]
                    let first = ids.removeFirst()
                    use(first)
                    return ids.popLast()
                }
                """,
            ),
            (
                "a value-returning write in a single-expression closure, which may hand it on: TypeScript's arrow body",
                """
                func collect(rows: [Row]) {
                    var seen = Set<String>()
                    rows.forEach { seen.insert($0.id) }
                }
                """,
            ),
            (
                "a value-returning write alone in a do, which may become an expression",
                """
                func collect() {
                    var seen = Set<String>()
                    do { seen.insert("a") }
                }
                """,
            ),
            (
                "a value-returning write as a nested function's single expression, which it returns",
                """
                func collect() -> Bool {
                    var seen = Set<String>()
                    func add() -> (inserted: Bool, memberAfterInsert: String) { seen.insert("b") }
                    return add().inserted
                }
                """,
            ),
            (
                "a value-returning write as a branch of an if expression, which is its value",
                """
                func collect(rows: [Row]) -> Bool {
                    var seen = Set<String>()
                    let added = if rows.isEmpty { seen.insert("c") } else { seen.insert("d") }
                    return added.inserted
                }
                """,
            ),
            (
                "a value-returning write as a branch of a switch expression returned implicitly",
                """
                func collect(row: Row) -> (inserted: Bool, memberAfterInsert: String) {
                    var seen = Set<String>()
                    return switch row.kind {
                    case .open: seen.insert("a")
                    default: seen.insert("b")
                    }
                }
                """,
            ),
            (
                "compared with ==, which reads, and a += of ours",
                """
                func collect(other: [String]) -> Bool {
                    var ids: [String] = []
                    ids.append("a")
                    var names: [String] = []
                    names.append("b")
                    names += other
                    return ids == other
                }
                """,
            ),
            (
                "a compound write through an element, which reads",
                """
                func count() {
                    var counts: [Int] = [0]
                    counts[0] += 1
                    var byName: [String: Int] = [:]
                    byName["a", default: 0] += 1
                }
                """,
            ),
            (
                "a write through an element, as TypeScript reads get(key).push",
                """
                func group(rows: [Row]) {
                    var groups: [String: [Row]] = [:]
                    for row in rows { groups[row.kind, default: []].append(row) }
                    var flags = [false]
                    flags[0].toggle()
                }
                """,
            ),
            (
                "at file scope, a global any file can read",
                """
                var registry: [String: Int] = [:]
                registry["a"] = 1
                func register(name: String) {
                    registry[name] = 1
                }
                """,
            ),
            (
                "a stored property, and a member of a local type",
                """
                final class Registry {
                    var names: [String] = []
                    func add(_ name: String) { names.append(name) }
                }
                func local() {
                    struct Box {
                        var items: [Int] = []
                        mutating func add() { items.append(1) }
                    }
                }
                """,
            ),
            (
                "never referenced, which swiftc reports",
                """
                func collect() {
                    var ids: [String] = []
                }
                """,
            ),
            (
                "reassigned whole, as TypeScript declines a let",
                """
                func collect() {
                    var ids: [String] = []
                    ids.append("a")
                    ids = []
                }
                """,
            ),
            (
                "a set type of our own named Set",
                """
                func collect() {
                    var ranks = Set<Int>()
                    ranks.insert(1)
                }
                """,
            ),
            (
                "a recorder of ours built from a literal",
                """
                func collect() {
                    var recorder: Recorder = []
                    recorder.append("a")
                }
                """,
            ),
            (
                "a type the source does not spell",
                """
                func collect(rows: [Row]) {
                    var ids = rows.map(\\.id)
                    ids.append("a")
                    var made = makeIds()
                    made.append("b")
                    var qualified: Swift.Array<Int> = []
                    qualified.append(1)
                }
                """,
            ),
            (
                "an append of ours, from an extension on Array",
                """
                func collect() {
                    var ids: [String] = []
                    ids.append("a")
                    ids.append("b", twice: true)
                }
                """,
            ),
            (
                "a subscript of ours on Array",
                """
                func collect() {
                    var slots: [Int] = [0]
                    slots[UInt(0)] = 1
                }
                """,
            ),
            (
                "iterated",
                """
                func collect() {
                    var ids: [String] = []
                    ids.append("a")
                    for id in ids { use(id) }
                }
                """,
            ),
            (
                "read in a closure",
                """
                func collect() -> () -> String {
                    var ids: [String] = []
                    ids.append("a")
                    return { ids.joined(separator: ",") }
                }
                """,
            ),
            (
                "captured by a capture list",
                """
                func collect() -> () -> Void {
                    var ids: [String] = []
                    ids.append("a")
                    return { [ids] in use(0) }
                }
                """,
            ),
            (
                "captured under another name",
                """
                func collect() -> () -> Void {
                    var ids: [String] = []
                    ids.append("a")
                    return { [copy = ids] in use(copy) }
                }
                """,
            ),
            (
                "passed inout, and interpolated",
                """
                func collect() {
                    var ids: [String] = []
                    ids.append("a")
                    fill(&ids)
                    var names: [String] = []
                    names.append("b")
                    log("\\(names)")
                }
                """,
            ),
            (
                "methods that hand the elements to our code, and reorder them",
                """
                func collect(rows: [Row]) {
                    var ids: [String] = []
                    ids.append("a")
                    ids.removeAll(where: { $0.isEmpty })
                    var sorted: [Int] = []
                    sorted.append(1)
                    sorted.sort()
                    var merged: [String: Int] = [:]
                    merged["a"] = 1
                    merged.merge(["b": 2]) { first, _ in first }
                }
                """,
            ),
            (
                "lazy, observed, and wrapped",
                """
                func collect() {
                    lazy var ids: [String] = []
                    ids.append("a")
                    var watched: [String] = [] {
                        didSet { use(watched) }
                    }
                    watched.append("b")
                    @Clamped var bounded: [Int] = []
                    bounded.append(1)
                }
                """,
            ),
            (
                "declared inside #if, which the code after #endif can name",
                """
                func collect() {
                    #if DEBUG
                    var ids: [String] = []
                    ids.append("a")
                    #endif
                }
                """,
            ),
            (
                "a name the scope reading cannot place: a local function of the same name",
                """
                func collect() {
                    var ids: [String] = []
                    ids.append("a")
                    func ids() {}
                    ids()
                }
                """,
            ),
            (
                "the shadowing binding read, the outer only written, TypeScript's case",
                """
                func collect() {
                    var ids: [String] = []
                    ids.append("a")
                    do {
                        let ids = ["b"]
                        use(ids)
                    }
                    use(ids.count)
                }
                """,
            ),
            (
                "a sibling binding in the same declaration reads it",
                """
                func collect() {
                    var ids: [String] = [], first = ids.first
                    ids.append("a")
                    use(first)
                }
                """,
            ),
        ]
        for (label, source) in cases {
            var resolving = Self.standardLibrary
            if label == "a set type of our own named Set" {
                resolving["Set"] = "s:7Control3SetV"
            }
            if label == "compared with ==, which reads, and a += of ours" {
                resolving["=="] = "s:SasSQRzlE2eeoiySbSayxG_ABtFZ"
                resolving["6:+="] = "s:7Control2peoiyySaySSGz_ACtF"
            }
            if label == "a subscript of ours on Array" {
                resolving["3:["] = "s:Sa7ControlEyxSucip"
            }
            if label == "an append of ours, from an extension on Array" {
                resolving["4:append"] = "s:Sa7ControlE6append_5twiceyx_SbtF"
            }
            #expect(Self.findings(source, resolving: resolving) == [], "\(label)")
        }
    }

    /* An overload of ours with the standard library's own labels: the index says whose it is, and the spelling cannot. */
    @Test func anOverloadOfOursWithTheSameLabelsIsNotAWrite() {
        let source = """
            func collect() {
                var ids: [String] = []
                ids.append("a")
            }
            """
        #expect(Self.findings(source, resolving: ["append": "s:Sa7ControlE6appendyyxF"]) == [])
        #expect(Self.findings(source) == ["ids"])
    }

    /* The prefilter needs a `var` and a bracket or one of the three type names. */
    @Test func thePrefilterNeedsAVarAndACollection() {
        let rule = CorrectnessNoWriteOnlyCollection()
        let parse = { (source: String) in
            ParsedFile(
                url: URL(fileURLWithPath: "/fixture/Subject.swift"),
                targetName: "Fixture",
                targetKind: "library",
                source: source,
                tree: Parser.parse(source: source),
                nodeCount: 0,
            )
        }
        #expect(!rule.applies(to: parse("let x = 1\n")))
        #expect(!rule.applies(to: parse("func f() { var x = 1; x += 1 }\n")))
        #expect(rule.applies(to: parse("func f() { var x = Set<Int>() }\n")))
        #expect(rule.applies(to: parse("func f() { var x: [Int] = [] }\n")))
    }

    // MARK: End to end

    static let packageSource = #"""
        extension Array {
            mutating func append(_ newElement: Element, twice: Bool) {
                append(newElement)
                if twice { append(newElement) }
            }
        }

        enum Collect {
            static func ranks(rows: [Int]) -> Int {
                var ranks: [Int: Int] = [:]
                for (index, row) in rows.enumerated() {
                    ranks[row] = index
                }
                return rows.count
            }

            static func seen(rows: [Int]) -> Int {
                var seen = Set<Int>()
                for row in rows { seen.insert(row) }
                return rows.count
            }

            static func unique(rows: [Int]) -> [Int] {
                var seen = Set<Int>()
                return rows.filter { seen.insert($0).inserted }
            }

            static func twice(rows: [Int]) -> Int {
                var doubled: [Int] = []
                for row in rows { doubled.append(row, twice: true) }
                return rows.count
            }

            static func named(rows: [Int]) -> Int {
                var named = Array<Int>()
                named += rows
                return rows.count
            }
        }

        """#

    @Test func theIndexResolvesTheCollectionsAndTheirWriters() async throws {
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
        let rule = CorrectnessNoWriteOnlyCollection()
        let candidates = parsed.files.filter { rule.applies(to: $0) }
        let symbols = SymbolProvider(
            scratchPaths: Pipeline.symbolScratchPaths(package: package, root: root),
            runner: ProcessRunner(),
        ).symbols(for: candidates)
        #expect(symbols.fromIndex == 1, "the build's index should describe the file: \(symbols.unavailable)")
        let found = candidates.flatMap { file in
            rule.findings(in: file, symbols: symbols.symbols[file.url.path] ?? FileSymbols([])).map {
                "\($0.line):\($0.column)"
            }
        }
        /* `ranks` (a dictionary's subscript), `seen` (Set's insert, through `Set` by name) and `named` (`Array` by name, `+=`); not `unique` (read) and not `doubled` (our append). */
        #expect(found == ["10:13", "18:13", "35:13"], "\(found)")
    }
}
