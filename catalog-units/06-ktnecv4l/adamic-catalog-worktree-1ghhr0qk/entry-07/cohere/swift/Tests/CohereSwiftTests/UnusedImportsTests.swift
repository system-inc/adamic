import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `unused-import` end to end on a real package of three targets, through the `--unused` report. `Base` declares
 the types, `Bridge` re-exports `Base`, and `Control`'s files each hold one case: an import that is used, one
 used only through a re-export, two that nothing uses, one kept alive by nothing but a sibling's re-export of a
 module the file imports directly, a file with `#if` that must be left unchecked, an `@_exported` import that
 is API and never reported, and pairs of system imports, two where the other import covers one and two
 where it does not. The records are a report, so the summary counts no findings.
 */
@Suite(.serialized)
struct UnusedImportsTests {
    static let manifest = """
        // swift-tools-version:6.0
        import PackageDescription

        let features: [SwiftSetting] = [.enableUpcomingFeature("ExistentialAny"), .enableUpcomingFeature("MemberImportVisibility")]

        let package = Package(
            name: "Control",
            targets: [
                .target(name: "Base", swiftSettings: features),
                .target(name: "Bridge", dependencies: ["Base"], swiftSettings: features),
                .target(name: "Control", dependencies: ["Base", "Bridge"], swiftSettings: features),
            ]
        )

        """

    static let files: [String: String] = [
        "Base/Item.swift": """
        public struct Item {
            public init() {}
        }

        """,
        "Bridge/Exports.swift": """
        @_exported import Base

        """,
        "Control/UsesBase.swift": """
        import Base

        func make() -> Item { Item() }

        """,
        "Control/ViaBridge.swift": """
        import Bridge

        func makeThroughTheBridge() -> Item { Item() }

        """,
        "Control/Unused.swift": """
        import Foundation
        import Base

        func plain() -> Int { 1 }

        """,
        "Control/Direct.swift": """
        import Base
        import Bridge

        func direct() -> Item { Item() }

        """,
        "Control/Conditional.swift": """
        import Foundation

        #if DEBUG
        func stamp() -> Date { Date() }
        #endif

        """,
        "Control/Exported.swift": """
        @_exported import Base

        """,
        "Control/Paired.swift": """
        import AppKit
        import SwiftUI

        func width() -> CGFloat { 1 }

        """,
        "Control/Received.swift": """
        import AppKit
        import SwiftUI

        @available(macOS 10.15, *)
        struct Received: View {
            var body: some View {
                Color.clear.onReceive(NotificationCenter.default.publisher(for: NSApplication.didBecomeActiveNotification)) { _ in }
            }
        }

        """,
        "Control/Imaged.swift": """
        import CoreGraphics
        import Foundation

        func picture(at url: URL) -> CGImage? { nil }

        """,
        "Control/Combined.swift": """
        import Combine
        import Foundation

        func link() -> URL? { URL(string: "https://example.com") }
        func hold() -> AnyCancellable { AnyCancellable {} }

        """,
    ]

    /* One `--no-fix --unused` run: each unused record as `file:line subject`, its message by that key, the coverage record, the phase, and how many findings carried the rule. */
    static func run() async throws -> (
        items: [String], messages: [String: String], coverage: [String: Any], phase: String, findings: Int
    ) {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-unused-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: root, withIntermediateDirectories: true)
        try manifest.write(to: root.appendingPathComponent("Package.swift"), atomically: true, encoding: .utf8)
        for (path, source) in files {
            let url = root.appendingPathComponent("Sources/\(path)")
            try FileManager.default.createDirectory(
                at: url.deletingLastPathComponent(),
                withIntermediateDirectories: true,
            )
            try source.write(to: url, atomically: true, encoding: .utf8)
        }
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix", "--unused"],
            workingDirectory: root,
        )
        var stream = Data()
        let writer = ContractWriter { stream.append($0) }
        _ = try await Pipeline(options: options, writer: writer, workingDirectory: root).run()
        var items: [String] = []
        var messages: [String: String] = [:]
        var coverage: [String: Any] = [:]
        var phase = ""
        var findings = 0
        for line in stream.split(separator: UInt8(ascii: "\n")) {
            let record = try #require(try JSONSerialization.jsonObject(with: Data(line)) as? [String: Any])
            switch record["kind"] as? String {
                case "unused" where record["rule"] as? String == UnusedImports.ruleName:
                    let file = (record["file"] as? String ?? "").components(separatedBy: "/Sources/").last ?? ""
                    let item = "\(file):\(record["line"] as? Int ?? 0) \(record["subject"] as? String ?? "")"
                    items.append(item)
                    messages[item] = record["message"] as? String ?? ""
                case "unusedCoverage" where record["rule"] as? String == UnusedImports.ruleName:
                    coverage = record
                case "phase" where record["name"] as? String == "unused":
                    phase = record["outcome"] as? String ?? ""
                case "finding" where record["rule"] as? String == UnusedImports.ruleName:
                    findings += 1
                default:
                    break
            }
        }
        return (items, messages, coverage, phase, findings)
    }

    @Test func reportsExactlyTheImportsNothingUses() async throws {
        let run = try await Self.run()
        #expect(run.phase == "ran")
        #expect(
            run.items.sorted() == [
                "Control/Direct.swift:2 import Bridge",
                "Control/Paired.swift:2 import SwiftUI",
                "Control/Received.swift:1 import AppKit",
                "Control/Unused.swift:1 import Foundation",
                "Control/Unused.swift:2 import Base",
            ],
            "the used, re-exported, conditional and @_exported imports must not be reported: \(run.items)",
        )
        /*
         Paired: `CGFloat` comes through either import, so one goes and the other names it. SwiftUI re-exports
         AppKit, so SwiftUI is the wider and goes first. Received: SwiftUI is needed for the view, and SwiftUI
         re-exports AppKit whole (its umbrella header imports AppKit's), so AppKit goes whatever the file needs
         from it, `onReceive`'s `Combine.Publisher` included. Combined: Foundation imports Combine
         without re-exporting it, so `AnyCancellable` needs `import Combine` and nothing goes. A check that
         believed every import a re-export would remove Combine there and break the build. Imaged: `CGImage`
         comes through no header Foundation re-exports, so `import CoreGraphics` stays. On Presence, crediting a
         Clang module's `export *` with every module its headers import read CoreGraphics as Foundation's and
         removed this import from a file that needs it.
         */
        #expect(
            run.messages["Control/Paired.swift:2 import SwiftUI"]?.contains(
                "comes through `import AppKit`, which re-exports it"
            ) == true,
            "\(run.messages)",
        )
        #expect(
            !run.items.contains { $0.hasPrefix("Control/Combined.swift") },
            "Foundation does not re-export Combine: \(run.items)",
        )
        #expect(
            !run.items.contains { $0.hasPrefix("Control/Imaged.swift") },
            "Foundation re-exports CoreGraphics's geometry headers, not CGImage: \(run.items)",
        )
        #expect(run.coverage["found"] as? Int == 5)
        #expect(run.coverage["rule"] as? String == UnusedImports.ruleName)
        let notChecked = run.coverage["filesNotChecked"] as? [String: Int] ?? [:]
        #expect(notChecked == [UnusedImports.conditional: 1], "only the #if file is left unchecked: \(notChecked)")
        let skipped = run.coverage["skipped"] as? [String: Int] ?? [:]
        #expect(
            skipped.values.reduce(0, +) == 2,
            "Bridge's and Control's @_exported imports are counted as never reported: \(skipped)",
        )
        /* A report, not a gate: nothing it found is a finding. */
        #expect(run.findings == 0, "unused-import findings must be unused records, never finding records")
    }

    /* The source with the removal of the import naming `module` applied. */
    static func removing(_ module: String, from source: String) throws -> String {
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/fixture/Subject.swift"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let node = try #require(
            file.tree.statements.compactMap { $0.item.as(ImportDeclSyntax.self) }
                .first { $0.path.trimmedDescription == module }
        )
        let edit = UnusedImports.removal(of: node, in: file)
        var bytes = Array(source.utf8)
        bytes.replaceSubrange(edit.start..<edit.end, with: Array(edit.text.utf8))
        return String(decoding: bytes, as: UTF8.self)
    }

    /*
     The removal takes the import's own line and nothing else. On ahraos-macos it started in the leading trivia,
     so a file's first import took the file header with it (AhraOsMain.swift, 18 lines), and an import below
     `import AhraOsCore // DevFlags` took that line too (MetalTerminalPane.swift).
     */
    @Test func removalTakesOnlyTheImportsLine() throws {
        #expect(
            try Self.removing("AhraOsCore", from: "//\n//  Main.swift\n//\n\nimport AhraOsCore\nimport AppKit\n")
                == "//\n//  Main.swift\n//\n\nimport AppKit\n"
        )
        #expect(
            try Self.removing("AppKit", from: "import AhraOsCore // DevFlags\nimport AppKit\nimport SwiftUI\n")
                == "import AhraOsCore // DevFlags\nimport SwiftUI\n"
        )
        #expect(
            try Self.removing("AhraOsCore", from: "import AhraOsCore // DevFlags\nimport AppKit\n")
                == "import AppKit\n"
        )
        #expect(try Self.removing("AppKit", from: "import Foundation\nimport AppKit") == "import Foundation\n")
        #expect(try Self.removing("AppKit", from: "import AppKit; let x = 1\n") == "; let x = 1\n")
    }
}
