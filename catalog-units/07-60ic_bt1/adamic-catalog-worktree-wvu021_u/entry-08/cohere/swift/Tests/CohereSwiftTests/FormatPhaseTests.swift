import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 The format phase formats every file the house way, with no `.swift-format` anywhere, and must agree with
 `xcrun swift-format` given the same settings byte for byte, because a formatter cohere cannot reproduce
 by hand is one nobody can check. The fixture is deliberately misformatted, and its expected text is
 written out: if the outputs equalled the input, the comparison would pass for a formatter that does
 nothing.
 */
struct FormatPhaseTests {
    static let dirtySource = #"""
        import Foundation
        struct  Example{
          let value:Int
            func doubled()->Int { value*2 }
        /* A prose block above a declaration, kept as written. */
        func describe(_ kind:Kind)->String{
        switch kind {
        case .small: return "small"
        case .large: return "large"
        }
        }
        func window(over items:[Int])->ArraySlice<Int>{ items[0 ..< min(3, items.count)] }
        func summary() -> String { Example.combine(firstArgumentWithALongName: value, secondArgumentWithALongName: value * 2, thirdArgumentWithALongName: value * 3) }
        static func combine(firstArgumentWithALongName: Int, secondArgumentWithALongName: Int, thirdArgumentWithALongName: Int) -> String { "\(firstArgumentWithALongName)" }
        let names = [
        "a",
        "b"
        ]
        enum Kind { case small, large }
        }

        """#

    /*
     The house way, by hand: case labels indented, no spaces around `..<`, a call and a signature that do not
     fit broken one argument per line with a trailing comma, the return type kept with the closing
     parenthesis, and a written-out array kept open with its trailing comma added.
     */
    static let houseSource = #"""
        import Foundation

        struct Example {
            let value: Int
            func doubled() -> Int { value * 2 }
            /* A prose block above a declaration, kept as written. */
            func describe(_ kind: Kind) -> String {
                switch kind {
                    case .small: return "small"
                    case .large: return "large"
                }
            }
            func window(over items: [Int]) -> ArraySlice<Int> { items[0..<min(3, items.count)] }
            func summary() -> String {
                Example.combine(
                    firstArgumentWithALongName: value,
                    secondArgumentWithALongName: value * 2,
                    thirdArgumentWithALongName: value * 3,
                )
            }
            static func combine(
                firstArgumentWithALongName: Int,
                secondArgumentWithALongName: Int,
                thirdArgumentWithALongName: Int,
            ) -> String { "\(firstArgumentWithALongName)" }
            let names = [
                "a",
                "b",
            ]
            enum Kind { case small, large }
        }

        """#

    static func workspace(source: String) throws -> URL {
        let directory = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-format-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        try source.write(to: directory.appendingPathComponent("Example.swift"), atomically: true, encoding: .utf8)
        return directory
    }

    static func parsed(_ url: URL, source: String) -> ParsedFile {
        ParsedFile(
            url: url,
            targetName: "Example",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
    }

    @Test func aFileIsFormattedTheHouseWayWithNoSwiftFormatAnywhere() throws {
        let directory = try Self.workspace(source: Self.dirtySource)
        let file = directory.appendingPathComponent("Example.swift")
        guard case let .changed(formatted) = FormatPhase.format(Self.parsed(file, source: Self.dirtySource)) else {
            Issue.record("the dirty fixture was not changed")
            return
        }
        #expect(formatted == Self.houseSource)
        guard case .unchanged = FormatPhase.format(Self.parsed(file, source: formatted)) else {
            Issue.record("a second pass moved the house's own output")
            return
        }
    }

    @Test func theLibraryMatchesTheToolchainFormatterGivenTheHouseSettings() throws {
        let directory = try Self.workspace(source: Self.dirtySource)
        let file = directory.appendingPathComponent("Example.swift")
        guard case let .changed(formatted) = FormatPhase.format(Self.parsed(file, source: Self.dirtySource)) else {
            Issue.record("the dirty fixture was not changed, so the comparison below would prove nothing")
            return
        }
        /* The house written out as swift-format's own JSON, outside the workspace, and handed to the toolchain by flag. */
        let settings = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-house-\(UUID().uuidString).json"
        )
        try JSONEncoder().encode(HouseSwiftFormat.configuration).write(to: settings)
        let toolchain = try ProcessRunner().run(
            "xcrun",
            ["swift-format", "format", "--configuration", settings.path, file.path],
            in: directory,
        )
        #expect(toolchain.succeeded, "xcrun swift-format failed: \(toolchain.standardError)")
        #expect(formatted == String(decoding: toolchain.standardOutput, as: UTF8.self))
    }

    @Test func theFirstDifferingLineIsReported() {
        #expect(FormatPhase.firstDifferingLine("a\nb\nc", "a\nB\nc") == 2)
        #expect(FormatPhase.firstDifferingLine("a\nb", "a\nb\n") == 3)
    }
}
