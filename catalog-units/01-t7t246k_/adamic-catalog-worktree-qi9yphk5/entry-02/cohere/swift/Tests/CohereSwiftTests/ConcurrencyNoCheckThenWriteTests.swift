import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `concurrency-no-check-then-write` both ways. The unit cases parse a source string and hand the rule symbols built
 by position, one occurrence at every token a case names, so each case says what the compiler would have resolved;
 a name no case resolves is a name the index did not record. A key `"<line>:<name>"` resolves that name on that
 line only, for a file where one name means two declarations. The three sites the survey found in Presence are
 modeled first, as they stand and with the one change that would make each fire, then every other shape the rule
 flags and every shape it leaves alone, the TypeScript rule's cases carried over where Swift has them. The
 end-to-end case runs a real package, so the symbols come from the index the build wrote.
 */
@Suite(.serialized)
struct ConcurrencyNoCheckThenWriteTests {
    static let fileExists = "c:objc(cs)NSFileManager(im)fileExistsAtPath:"
    static let fileExistsIsDirectory = "c:objc(cs)NSFileManager(im)fileExistsAtPath:isDirectory:"
    static let urlPath = "s:10Foundation3URLV4pathSSvp"
    static let urlPathPercentEncoded = "s:10Foundation3URLV4path14percentEncodedSSSb_tF"
    static let dataWrite = "s:10Foundation4DataV5write2to7optionsyAA3URLV_So20NSDataWritingOptionsVtKF"
    static let stringWriteUrl = "s:Sy10FoundationE5write2to10atomically8encodingyAA3URLV_SbSSAAE8EncodingVtKF"
    static let stringWriteFile = "s:Sy10FoundationE5write6toFile10atomically8encodingyqd___SbSSAAE8EncodingVtKSyRd__lF"
    static let createFile = "c:objc(cs)NSFileManager(im)createFileAtPath:contents:attributes:"
    static let copyItem = "c:objc(cs)NSFileManager(im)copyItemAtURL:toURL:error:"
    static let moveItem = "c:objc(cs)NSFileManager(im)moveItemAtURL:toURL:error:"
    static let appendingPathComponent = "s:10Foundation3URLV22appendingPathComponentyACSSF"
    static let appendingPathExtension = "s:10Foundation3URLV22appendingPathExtensionyACSSF"
    static let appendingPath = "s:10Foundation3URLV9appending4path13directoryHintACx_AC09DirectoryF0OtSyRzlF"
    /* Ours, in Presence: `CharacterCard.write(_:portrait:to:)` and `read(from:)`, and `BodyImport.destination(for:)`. */
    static let cardWrite =
        "s:15AhraOSPresence13CharacterCardO5write_8portrait2toyAA10AppearanceV_So7NSImageCAA3URLVtKFZ"
    static let cardRead = "s:15AhraOSPresence13CharacterCardO4read4fromAA10AppearanceV10Foundation3URLV_tKFZ"
    static let bodyDestination = "s:15AhraOSPresence10BodyImportC11destination3for10Foundation3URLVSS_tFZ"
    /* A project's own `fileExists(atPath:)`, which is not FileManager's whatever its name. */
    static let ourFileExists = "s:7Control7StorageO10fileExists6atPathSbSS_tFZ"

    /* What a typical case resolves: the check, `.path`, `Data.write`, and the path builders. */
    static let foundation: [String: String] = [
        "fileExists": fileExists,
        "path": urlPath,
        "write": dataWrite,
        "createFile": createFile,
        "copyItem": copyItem,
        "moveItem": moveItem,
        "appendingPathComponent": appendingPathComponent,
        "appendingPathExtension": appendingPathExtension,
        "appending": appendingPath,
    ]

    /* Every finding the rule makes on a source, its names resolved as the suite's comment describes. */
    static func records(_ source: String, resolving: [String: String] = foundation) -> [FindingRecord] {
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
        let rule = ConcurrencyNoCheckThenWrite()
        let found = rule.findings(in: file, symbols: FileSymbols(occurrences, ownedModules: ["Control"]))
        #expect(found.isEmpty || rule.applies(to: file), "the prefilter must never hide a finding")
        return found
    }

    /* Every finding as the text of its span and the message's named write. */
    static func findings(_ source: String, resolving: [String: String] = foundation) -> [String] {
        let lines = source.split(separator: "\n", omittingEmptySubsequences: false).map { Array($0.utf8) }
        return records(source, resolving: resolving).map { finding in
            #expect(finding.messageId == "checkThenWrite")
            #expect(finding.fixes.isEmpty && finding.suggestions.isEmpty, "the rule never fixes")
            guard let endLine = finding.endLine, let endColumn = finding.endColumn, endLine == finding.line else {
                return "spans lines"
            }
            let span = String(decoding: lines[finding.line - 1][(finding.column - 1)..<(endColumn - 1)], as: UTF8.self)
            let named =
                [
                    "write(to:)", "write(to:options:)", "write(to:atomically:encoding:)",
                    "write(toFile:atomically:encoding:)", "createFile(atPath:contents:)",
                ]
                .first { finding.message == RuleMessages.ConcurrencyNoCheckThenWrite.checkThenWrite(write: $0).text }
                ?? "unnamed"
            return "\(span) | \(named)"
        }
    }

    // MARK: The real sites

    /*
     `ahraos-presence` `Sources/AhraOSPresence/Stage/StagePane+Summoning.swift:309` on 2026-10-03, `saveSummonedCard`.
     The loop's condition is the check and a second clause, which also stops on a card of the same body (replaced
     on purpose), and the card is written by `CharacterCard.write`, ours. Each alone keeps the rule silent.
     */
    static func summoning(condition: String, write: String) -> String {
        """
        import AppKit

        extension StagePane {
            private func saveSummonedCard(named name: String, bodyId: String) {
                Task { @MainActor in
                    guard let portrait = await self.portrait() else {
                        presenceLog("summon: nothing to photograph, card not saved")
                        return
                    }
                    do {
                        try FileManager.default.createDirectory(at: CharacterCard.folder, withIntermediateDirectories: true)
                    }
                    catch {
                        presenceLog("summon: could not make the cards folder: \\(error)")
                    }

                    let base = name.replacingOccurrences(of: "/", with: "-").trimmingCharacters(in: .whitespacesAndNewlines)
                    var url = CharacterCard.folder.appendingPathComponent(base.isEmpty ? "Summoned" : base)
                        .appendingPathExtension("png")
                    var number = 2
        \(condition)
                    {
                        url = CharacterCard.folder.appendingPathComponent("\\(base) \\(number)").appendingPathExtension("png")
                        number += 1
                    }

                    var card = Appearance()
                    card.name = name
                    card.bodyId = bodyId
                    do {
        \(write)
                        presenceLog("summon: saved card \\(url.lastPathComponent) for body \\(bodyId)")
                    }
                    catch {
                        presenceLog("summon: could not save card: \\(error)")
                    }
                }
            }
        }

        """
    }

    static let summoningCondition = """
                    while FileManager.default.fileExists(atPath: url.path),
                        (try? CharacterCard.read(from: url))?.bodyId != bodyId
        """
    static let summoningCheckAlone = """
                    while FileManager.default.fileExists(atPath: url.path)
        """
    static let summoningCardWrite = """
                        try CharacterCard.write(card, portrait: portrait, to: url)
        """
    /* `CharacterCard.write`'s body, `Studio/CharacterCard.swift:45`, inlined: the payload and its plain `Data.write`. */
    static let summoningInlineWrite = """
                        guard let tiff = portrait.tiffRepresentation, let bitmap = NSBitmapImageRep(data: tiff),
                            let png = bitmap.representation(using: .png, properties: [:]) else { return }
                        var payload = Data()
                        payload.append(png)
                        payload.append(try JSONEncoder().encode(card))
                        try payload.write(to: url)
        """

    /* Foundation's names, `CharacterCard.read`, and `CharacterCard.write` on the line that calls it (where `write` is ours, not `Data`'s). */
    static func summoningFindings(condition: String, write: String) -> [String] {
        let source = Self.summoning(condition: condition, write: write)
        var resolving = Self.foundation.merging(["read": Self.cardRead]) { _, new in new }
        if let index = source.split(separator: "\n", omittingEmptySubsequences: false).firstIndex(where: {
            $0.contains("CharacterCard.write(")
        }) {
            resolving["\(index + 1):write"] = Self.cardWrite
        }
        return Self.findings(source, resolving: resolving)
    }

    @Test func summoningAsItStandsIsAKnownMiss() {
        #expect(Self.summoningFindings(condition: Self.summoningCondition, write: Self.summoningCardWrite).isEmpty)
    }

    @Test func summoningWithTheCheckAloneStillWritesThroughAFunctionOfOurs() {
        #expect(Self.summoningFindings(condition: Self.summoningCheckAlone, write: Self.summoningCardWrite).isEmpty)
    }

    @Test func summoningWithTheCardWrittenInlineButTheSecondClauseKeptIsNotTheShape() {
        #expect(Self.summoningFindings(condition: Self.summoningCondition, write: Self.summoningInlineWrite).isEmpty)
    }

    @Test func summoningWithTheCheckAloneAndTheCardWrittenInlineIsFlagged() {
        let found = Self.summoningFindings(condition: Self.summoningCheckAlone, write: Self.summoningInlineWrite)
        #expect(found == ["FileManager.default.fileExists(atPath: url.path) | write(to:)"], "\(found)")
    }

    /* The repair: the write claims the name with `.withoutOverwriting` and the loop tries the next name on the error. No check is left to race. */
    @Test func summoningsRepairIsClean() {
        let source = """
            import AppKit

            func saveSummonedCard(_ payload: Data, base: String) throws {
                var url = CharacterCard.folder.appendingPathComponent(base).appendingPathExtension("png")
                var number = 2
                while true {
                    do {
                        try payload.write(to: url, options: .withoutOverwriting)
                        break
                    }
                    catch let error as CocoaError where error.code == .fileWriteFileExists {
                        url = CharacterCard.folder.appendingPathComponent("\\(base) \\(number)").appendingPathExtension("png")
                        number += 1
                    }
                }
            }

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* `Sources/AhraOSPresence/Studio/FaceSets.swift:126`: the free name is returned, and `StudioModel.swift:926` writes it later. One function is all the rule reads. */
    static let faceSets = """
        import AppKit

        enum FaceSets {
            /// A filename not yet taken in the folder: the name, then the name and a number.
            static func unusedUrl(for name: String) -> URL {
                let base = name.replacingOccurrences(of: "/", with: "-").trimmingCharacters(in: .whitespacesAndNewlines)
                let stem = base.isEmpty || base == "Unnamed" ? "Face" : base
                var candidate = Self.folder.appendingPathComponent(stem).appendingPathExtension("png")
                var number = 2
                while FileManager.default.fileExists(atPath: candidate.path) {
                    candidate = Self.folder.appendingPathComponent("\\(stem) \\(number)").appendingPathExtension("png")
                    number += 1
                }
                return candidate
            }
        }

        """

    @Test func faceSetsHelperReturnsTheNameAndIsAKnownMiss() {
        #expect(Self.findings(Self.faceSets).isEmpty)
    }

    /* The helper and its caller's write in one function: the shape the rule reads, flagged. */
    @Test func faceSetsHelperWithItsCallersWriteInlineIsFlagged() {
        let source = Self.faceSets.replacingOccurrences(
            of: "static func unusedUrl(for name: String) -> URL {",
            with: "static func save(_ payload: Data, for name: String) throws {",
        )
        .replacingOccurrences(of: "        return candidate\n", with: "        try payload.write(to: candidate)\n")
        let found = Self.findings(source)
        #expect(found == ["FileManager.default.fileExists(atPath: candidate.path) | write(to:)"], "\(found)")
    }

    /* `Sources/AhraOSPresence/Studio/Bodies/BodyImport.swift:302`: the path is a call to a function of ours, never proven pure, and the caller lands the file with `moveItem`, which refuses a destination already there. */
    static let bodyImport = """
        import Foundation

        final class BodyImport {
            private static func destination(for stem: String) -> URL {
                BodyLibrary.rootUrl.appendingPathComponent("\\(stem).vrm")
            }

            /// The first "<stem> 2", "<stem> 3" not taken.
            private static func free(_ stem: String) -> String {
                var number = 2
                while FileManager.default.fileExists(atPath: Self.destination(for: "\\(stem) \\(number)").path) { number += 1 }
                return "\\(stem) \\(number)"
            }

            func file(_ output: URL, stem: String, payload: Data) throws {
                var number = 2
                while FileManager.default.fileExists(atPath: Self.destination(for: "\\(stem) \\(number)").path) { number += 1 }
                try payload.write(to: Self.destination(for: "\\(stem) \\(number)"))
                let destination = Self.destination(for: "\\(stem) \\(number)")
                try FileManager.default.moveItem(at: output, to: destination)
            }
        }

        """

    @Test func bodyImportsHelperAndAnInlinedWriteThroughOurFunctionAreNotProven() {
        #expect(
            Self.findings(
                Self.bodyImport,
                resolving: Self.foundation.merging(["destination": Self.bodyDestination]) { _, new in new },
            ).isEmpty
        )
    }

    // MARK: Every flagged shape

    /* One function over a folder, a directory string, a base name, bytes and an artifact's input, with the given lines as its body. */
    static func shape(_ body: String) -> String {
        """
        import Foundation

        func save(_ data: Data, text: String, base: String, folder: URL, directory: String, items: [Data], input: ArtifactInput) async throws {
            var number = 2
            var url = folder.appendingPathComponent("\\(base).png")
        \(body)
        }

        """
    }

    static let freeNameLoop = """
            while FileManager.default.fileExists(atPath: url.path) {
                url = folder.appendingPathComponent("\\(base) \\(number).png")
                number += 1
            }
        """
    static let freeNameCheck = "FileManager.default.fileExists(atPath: url.path)"

    @Test func everyFlaggedShapeIsFound() {
        let cases: [(name: String, body: String, resolving: [String: String], expected: String)] = [
            (
                "a while, then Data.write with no options", Self.freeNameLoop + "\n    try data.write(to: url)",
                Self.foundation, "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "options that replace: .atomic", Self.freeNameLoop + "\n    try data.write(to: url, options: .atomic)",
                Self.foundation, "\(Self.freeNameCheck) | write(to:options:)",
            ),
            (
                "options that replace: [.atomic]",
                Self.freeNameLoop + "\n    try data.write(to: url, options: [.atomic])", Self.foundation,
                "\(Self.freeNameCheck) | write(to:options:)",
            ),
            (
                "options that replace: []", Self.freeNameLoop + "\n    try data.write(to: url, options: [])",
                Self.foundation, "\(Self.freeNameCheck) | write(to:options:)",
            ),
            (
                "a repeat-while",
                """
                    repeat {
                        url = folder.appendingPathComponent("\\(base) \\(number).png")
                        number += 1
                    } while FileManager.default.fileExists(atPath: url.path)
                    try data.write(to: url)
                """,
                Self.foundation, "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "a String written to the URL",
                Self.freeNameLoop + "\n    try text.write(to: url, atomically: true, encoding: .utf8)",
                Self.foundation.merging(["write": Self.stringWriteUrl]) { _, new in new },
                "\(Self.freeNameCheck) | write(to:atomically:encoding:)",
            ),
            (
                "a path string, a String written to the file",
                """
                    var path = directory + "/" + base + ".txt"
                    while FileManager.default.fileExists(atPath: path) {
                        path = directory + "/" + base + " \\(number).txt"
                        number += 1
                    }
                    try text.write(toFile: path, atomically: true, encoding: .utf8)
                """,
                Self.foundation.merging(["write": Self.stringWriteFile]) { _, new in new },
                "FileManager.default.fileExists(atPath: path) | write(toFile:atomically:encoding:)",
            ),
            (
                "createFile over the same concatenated path, spelled out on both sides",
                """
                    while FileManager.default.fileExists(atPath: directory + "/\\(number).png") { number += 1 }
                    _ = FileManager.default.createFile(atPath: directory + "/\\(number).png", contents: data)
                """,
                Self.foundation,
                "FileManager.default.fileExists(atPath: directory + \"/\\(number).png\") | createFile(atPath:contents:)",
            ),
            (
                "path builders spelled out on both sides",
                """
                    while FileManager.default.fileExists(atPath: folder.appendingPathComponent("\\(number)").appendingPathExtension("png").path) { number += 1 }
                    try data.write(to: folder.appendingPathComponent("\\(number)").appendingPathExtension("png"))
                """,
                Self.foundation,
                "FileManager.default.fileExists(atPath: folder.appendingPathComponent(\"\\(number)\").appendingPathExtension(\"png\").path) | write(to:)",
            ),
            (
                "a let after the loop, read through",
                """
                    while FileManager.default.fileExists(atPath: folder.appending(path: "\\(number).png").path) { number += 1 }
                    let target = folder.appending(path: "\\(number).png")
                    try data.write(to: target)
                """,
                Self.foundation,
                "FileManager.default.fileExists(atPath: folder.appending(path: \"\\(number).png\").path) | write(to:)",
            ),
            (
                "the URL's path read into a let after the loop, written as a file path",
                Self.freeNameLoop
                    + "\n    let target = url.path\n    try text.write(toFile: target, atomically: true, encoding: .utf8)",
                Self.foundation.merging(["write": Self.stringWriteFile]) { _, new in new },
                "\(Self.freeNameCheck) | write(toFile:atomically:encoding:)",
            ),
            (
                "the write in a branch after the loop",
                Self.freeNameLoop + "\n    if !data.isEmpty {\n        try data.write(to: url)\n    }", Self.foundation,
                "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "the loop in a branch, the write after it",
                "    if !data.isEmpty {\n" + Self.freeNameLoop + "\n    }\n    try data.write(to: url)",
                Self.foundation, "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "the loop in a do, the write after it",
                "    do {\n" + Self.freeNameLoop + "\n    }\n    try data.write(to: url)", Self.foundation,
                "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "an await between the loop and the write, a write in a loop after it",
                Self.freeNameLoop
                    + "\n    try await Task.sleep(nanoseconds: 10)\n    for item in items {\n        try item.write(to: url)\n    }",
                Self.foundation, "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "the loop and the write inside a task, as Presence's summoning writes",
                "    Task { @MainActor in\n        var url = url\n        var number = number\n" + Self.freeNameLoop
                    + "\n        try? data.write(to: url)\n    }",
                Self.foundation, "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "a FileManager held in a let",
                """
                    let manager = FileManager.default
                    while manager.fileExists(atPath: url.path) {
                        url = folder.appendingPathComponent("\\(base) \\(number).png")
                        number += 1
                    }
                    try data.write(to: url)
                """,
                Self.foundation, "manager.fileExists(atPath: url.path) | write(to:)",
            ),
            (
                "the counter reaching the path through a let in the loop",
                """
                    while FileManager.default.fileExists(atPath: url.path) {
                        let name = "\\(base) \\(number).png"
                        url = folder.appendingPathComponent(name)
                        number -= 1
                    }
                    try data.write(to: url)
                """,
                Self.foundation, "\(Self.freeNameCheck) | write(to:)",
            ),
            (
                "the TypeScript rule's incident with the fields read into lets before the loop",
                """
                    let outputDirectory = input.outputDirectory
                    let fileExtension = input.fileExtension
                    while FileManager.default.fileExists(atPath: outputDirectory.appendingPathComponent("\\(number).\\(fileExtension)").path) { number += 1 }
                    try data.write(to: outputDirectory.appendingPathComponent("\\(number).\\(fileExtension)"))
                """,
                Self.foundation,
                "FileManager.default.fileExists(atPath: outputDirectory.appendingPathComponent(\"\\(number).\\(fileExtension)\").path) | write(to:)",
            ),
            (
                "the write in a guard's else after the loop in it",
                "    guard data.isEmpty else {\n" + Self.freeNameLoop
                    + "\n        try data.write(to: url)\n        return\n    }",
                Self.foundation, "\(Self.freeNameCheck) | write(to:)",
            ),
        ]
        for testCase in cases {
            let found = Self.findings(Self.shape(testCase.body), resolving: testCase.resolving)
            #expect(found == [testCase.expected], "\(testCase.name): \(found)")
        }
    }

    // MARK: Every shape left alone

    @Test func everySafeShapeIsClean() {
        let cases: [(name: String, body: String, resolving: [String: String])] = [
            (
                "a wait for a lock to go away, no counter",
                """
                    let lock = folder.appendingPathComponent("lock")
                    while FileManager.default.fileExists(atPath: lock.path) {
                        try await Task.sleep(nanoseconds: 100)
                    }
                    try data.write(to: lock)
                """,
                Self.foundation,
            ),
            (
                "the write claims with .withoutOverwriting",
                Self.freeNameLoop + "\n    try data.write(to: url, options: .withoutOverwriting)", Self.foundation,
            ),
            (
                "a list holding .withoutOverwriting",
                Self.freeNameLoop + "\n    try data.write(to: url, options: [.atomic, .withoutOverwriting])",
                Self.foundation,
            ),
            (
                "options the source does not show",
                Self.freeNameLoop
                    + "\n    let options: Data.WritingOptions = []\n    try data.write(to: url, options: options)",
                Self.foundation,
            ),
            (
                "options qualified by their type",
                Self.freeNameLoop + "\n    try data.write(to: url, options: Data.WritingOptions.atomic)",
                Self.foundation,
            ),
            (
                "a copy, which refuses a destination already there",
                Self.freeNameLoop + "\n    try FileManager.default.copyItem(at: folder, to: url)", Self.foundation,
            ),
            (
                "a move, which refuses a destination already there",
                Self.freeNameLoop + "\n    try FileManager.default.moveItem(at: folder, to: url)", Self.foundation,
            ),
            (
                "a write to a different path",
                Self.freeNameLoop
                    + "\n    try data.write(to: url.appendingPathExtension(\"json\"))\n    try data.write(to: folder.appendingPathComponent(\"\\(base) \\(number).png\"))",
                Self.foundation,
            ),
            (
                "the URL changes between the loop and the write",
                Self.freeNameLoop + "\n    url = url.appendingPathExtension(\"tmp\")\n    try data.write(to: url)",
                Self.foundation,
            ),
            (
                "a method called on the URL between the loop and the write",
                Self.freeNameLoop + "\n    url.standardize()\n    try data.write(to: url)", Self.foundation,
            ),
            (
                "the counter moves between the loop and the write",
                """
                    while FileManager.default.fileExists(atPath: folder.appendingPathComponent("\\(number).png").path) { number += 1 }
                    number += 1
                    try data.write(to: folder.appendingPathComponent("\\(number).png"))
                """,
                Self.foundation,
            ),
            (
                "the counter written from a closure",
                """
                    let reset = { number = 0 }
                    while FileManager.default.fileExists(atPath: folder.appendingPathComponent("\\(number).png").path) { number += 1 }
                    reset()
                    try data.write(to: folder.appendingPathComponent("\\(number).png"))
                """,
                Self.foundation,
            ),
            (
                "the write in a closure after the loop",
                Self.freeNameLoop + "\n    let later = { try? data.write(to: url) }\n    later()", Self.foundation,
            ),
            (
                "the write in a closure capturing the URL",
                Self.freeNameLoop + "\n    DispatchQueue.main.async { [url] in try? data.write(to: url) }",
                Self.foundation,
            ),
            (
                "the TypeScript rule's incident as it stood: the path read through the input's properties",
                """
                    while FileManager.default.fileExists(atPath: input.outputDirectory.appendingPathComponent("\\(number).\\(input.fileExtension)").path) { number += 1 }
                    try data.write(to: input.outputDirectory.appendingPathComponent("\\(number).\\(input.fileExtension)"))
                """,
                Self.foundation,
            ),
            ("the write before the loop", "    try data.write(to: url)\n" + Self.freeNameLoop, Self.foundation),
            (
                "the write in the other branch",
                "    if data.isEmpty {\n" + Self.freeNameLoop
                    + "\n    } else {\n        try data.write(to: url)\n    }", Self.foundation,
            ),
            (
                "the write past an enclosing loop",
                "    for item in items {\n" + Self.freeNameLoop
                    + "\n        _ = item\n    }\n    try data.write(to: url)", Self.foundation,
            ),
            (
                "the write after a guard whose else holds the loop",
                "    guard data.isEmpty else {\n" + Self.freeNameLoop
                    + "\n        return\n    }\n    try data.write(to: url)", Self.foundation,
            ),
            (
                "the loop stops on a name that exists",
                """
                    while !FileManager.default.fileExists(atPath: url.path) {
                        url = folder.appendingPathComponent("\\(base) \\(number).png")
                        number += 1
                    }
                    try data.write(to: url)
                """,
                Self.foundation,
            ),
            (
                "a check with a second argument",
                """
                    while FileManager.default.fileExists(atPath: url.path, isDirectory: nil) {
                        url = folder.appendingPathComponent("\\(base) \\(number).png")
                        number += 1
                    }
                    try data.write(to: url)
                """,
                Self.foundation.merging(["fileExists": Self.fileExistsIsDirectory]) { _, new in new },
            ),
            (
                "a fileExists of ours", Self.freeNameLoop + "\n    try data.write(to: url)",
                Self.foundation.merging(["fileExists": Self.ourFileExists]) { _, new in new },
            ),
            (
                "a fileExists the index did not record", Self.freeNameLoop + "\n    try data.write(to: url)",
                Self.foundation.filter { $0.key != "fileExists" },
            ),
            (
                "a write of ours", Self.freeNameLoop + "\n    try data.write(to: url)",
                Self.foundation.merging(["write": Self.cardWrite]) { _, new in new },
            ),
            (
                "a write the index did not record", Self.freeNameLoop + "\n    try data.write(to: url)",
                Self.foundation.filter { $0.key != "write" },
            ),
            (
                "a let from before the loop, spelled out on the other side",
                """
                    let first = folder.appendingPathComponent("\\(number).png")
                    while FileManager.default.fileExists(atPath: folder.appendingPathComponent("\\(number).png").path) { number += 1 }
                    try data.write(to: first)
                """,
                Self.foundation,
            ),
            (
                "a path read through a property",
                """
                    while FileManager.default.fileExists(atPath: Storage.folder.appendingPathComponent("\\(number).png").path) { number += 1 }
                    try data.write(to: Storage.folder.appendingPathComponent("\\(number).png"))
                """,
                Self.foundation,
            ),
            (
                "a URL made from a path string, which resolves a relative path against the working directory",
                """
                    while FileManager.default.fileExists(atPath: directory + "/\\(number).png") { number += 1 }
                    try data.write(to: URL(fileURLWithPath: directory + "/\\(number).png"))
                """,
                Self.foundation,
            ),
            (
                "a percent-encoded path in the check",
                """
                    while FileManager.default.fileExists(atPath: url.path()) {
                        url = folder.appendingPathComponent("\\(base) \\(number).png")
                        number += 1
                    }
                    try data.write(to: url)
                """,
                Self.foundation.merging(["path": Self.urlPathPercentEncoded]) { _, new in new },
            ),
            (
                "a write through optional chaining",
                Self.freeNameLoop + "\n    let maybe: Data? = data\n    try maybe?.write(to: url)", Self.foundation,
            ),
            (
                "a different url shadowing the checked one",
                Self.freeNameLoop + "\n    if let url = URL(string: base) {\n        try data.write(to: url)\n    }",
                Self.foundation,
            ),
            (
                "a path builder given a directory hint",
                """
                    while FileManager.default.fileExists(atPath: folder.appending(path: "\\(number).png", directoryHint: .notDirectory).path) { number += 1 }
                    try data.write(to: folder.appending(path: "\\(number).png", directoryHint: .notDirectory))
                """,
                Self.foundation,
            ),
        ]
        for testCase in cases {
            let found = Self.findings(Self.shape(testCase.body), resolving: testCase.resolving)
            #expect(found.isEmpty, "\(testCase.name): \(found)")
        }
    }

    /* A loop at file scope, as in a `main.swift`: its bindings are globals, which another file can change. */
    @Test func aLoopAtFileScopeIsNotRead() {
        let source = """
            import Foundation

            let folder = URL(fileURLWithPath: "/tmp")
            var number = 2
            var url = folder.appendingPathComponent("1.png")
            while FileManager.default.fileExists(atPath: url.path) {
                url = folder.appendingPathComponent("\\(number).png")
                number += 1
            }
            try Data().write(to: url)

            """
        #expect(Self.findings(source).isEmpty)
    }

    /* The message names the write and the repair; checked against a literal rather than the rule's own text. */
    @Test func theMessageNamesTheWriteAndTheRepair() {
        let found = Self.records(Self.shape(Self.freeNameLoop + "\n    try data.write(to: url)"))
        #expect(
            found.map(\.message) == [
                "This loop looks for a file name that is free, and the `write(to:)` after it creates that file in a separate step, so two saves running at once (in this process or in two) can both find the same name free and the second silently overwrites the first. Claim the name in the step that creates the file: `data.write(to: url, options: .withoutOverwriting)` (or `FileManager.moveItem` from a temporary file) fails with `CocoaError.fileWriteFileExists` when the name is taken, so try the next name on that error instead of checking first."
            ]
        )
        #expect(found.map(\.rule) == ["cohere-swift/concurrency-no-check-then-write"])
    }

    /*
     The finding a user sees is the catalog's entry (policy/messages/concurrency-no-check-then-write.json), id and text,
     so an edited entry reaches the finding once regenerated. The literal above pins the words; this pins where they
     come from.
     */
    @Test func theFindingIsTheCatalogsMessage() {
        let found = Self.records(Self.shape(Self.freeNameLoop + "\n    try data.write(to: url)"))
        let message = RuleMessages.ConcurrencyNoCheckThenWrite.checkThenWrite(write: "write(to:)")
        #expect(found.map(\.messageId) == [message.id])
        #expect(found.map(\.message) == [message.text])
    }

    /* The prefilter holds every flagged file and declines a file with no check. */
    @Test func thePrefilterNeedsACheckAndAWrite() {
        let rule = ConcurrencyNoCheckThenWrite()
        let parsed = { (source: String) in
            ParsedFile(
                url: URL(fileURLWithPath: "/fixture/Subject.swift"),
                targetName: "Fixture",
                targetKind: "library",
                source: source,
                tree: Parser.parse(source: source),
                nodeCount: 0,
            )
        }
        #expect(rule.applies(to: parsed(Self.shape(Self.freeNameLoop + "\n    try data.write(to: url)"))))
        #expect(!rule.applies(to: parsed("func save(_ data: Data, to url: URL) throws { try data.write(to: url) }")))
    }

    // MARK: End to end

    static let packageSource = #"""
        import Foundation

        enum Storage {
            static func fileExists(atPath path: String) -> Bool {
                path.isEmpty
            }
        }

        enum Cards {
            static func save(_ data: Data, named base: String, in folder: URL) throws {
                var url = folder.appendingPathComponent(base).appendingPathExtension("png")
                var number = 2
                while FileManager.default.fileExists(atPath: url.path) {
                    url = folder.appendingPathComponent("\(base) \(number)").appendingPathExtension("png")
                    number += 1
                }
                try data.write(to: url)
            }

            static func claim(_ data: Data, named base: String, in folder: URL) throws {
                var url = folder.appendingPathComponent(base).appendingPathExtension("png")
                var number = 2
                while FileManager.default.fileExists(atPath: url.path) {
                    url = folder.appendingPathComponent("\(base) \(number)").appendingPathExtension("png")
                    number += 1
                }
                try data.write(to: url, options: .withoutOverwriting)
            }

            static func note(_ text: String, in directory: String) throws {
                var number = 1
                while FileManager.default.fileExists(atPath: directory + "/\(number).txt") {
                    number += 1
                }
                try text.write(toFile: directory + "/\(number).txt", atomically: true, encoding: .utf8)
            }

            static func lookalike(_ data: Data, in directory: String) {
                var number = 1
                while Storage.fileExists(atPath: directory + "/\(number).bin") {
                    number += 1
                }
                _ = FileManager.default.createFile(atPath: directory + "/\(number).bin", contents: data)
            }

            static func create(_ data: Data, in folder: URL) {
                var number = 1
                while FileManager.default.fileExists(atPath: folder.appending(path: "\(number).bin").path) {
                    number += 1
                }
                _ = FileManager.default.createFile(atPath: folder.appending(path: "\(number).bin").path, contents: data)
            }
        }

        """#

    @Test func theIndexResolvesFoundationsCheckWritesAndPathBuilders() async throws {
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
        let rule = ConcurrencyNoCheckThenWrite()
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
        /* `save` (Data.write), `note` (String's write(toFile:)) and `create` (createFile through `appending(path:)` and `.path`); not `claim` (exclusive) and not `lookalike` (our own check). */
        #expect(found == ["13:15", "32:15", "48:15"], "\(found)")
    }
}
