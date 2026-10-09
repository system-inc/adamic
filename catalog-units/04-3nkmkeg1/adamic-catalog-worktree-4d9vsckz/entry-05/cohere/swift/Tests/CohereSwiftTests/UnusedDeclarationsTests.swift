import Foundation
import SwiftParser
import SwiftSyntax
import Testing

@testable import CohereSwift

/*
 `unused-declaration` end to end on a real package, through the `--unused` report. Each file holds one case: a
 private function, a file-private constant and an extension's member nothing calls; a function only itself
 calls; an unused private type holding a member of its own, reported once as the type; a base method only
 its override reaches; a `@State` property used only as `$name`; a Codable type whose coding keys and stored
 properties the synthesized conformance reads; an instance property that holds an observer token beside one
 that holds a plain value; a struct of plain numbers whose padding is never named; stored properties of types
 whose conformances reach Codable through a protocol of ours, a type alias or an extension in another file, of
 a type inheriting from NSObject, of a Hashable type, whose synthesized hash reads them, and of its twin with
 no Hashable, the one that is reported; the per-case coding keys of a Codable enum, declared in it and in an
 extension, and of the same enum without Codable, which are reported; a protocol witness; an `@objc` method; a file with `#if`; an internal function nothing calls; and, judged across the package, public
 API of a library product, a function only a test calls, one named only in a `#if` branch the build never
 compiled, and a wrapped property read only through its projection. The records are a report, so the summary
 counts no findings.
 */
@Suite(.serialized)
struct UnusedDeclarationsTests {
    static let manifest = """
        // swift-tools-version:6.0
        import PackageDescription

        let package = Package(
            name: "Control",
            platforms: [.macOS(.v14)],
            products: [.library(name: "Control", targets: ["Control"])],
            targets: [
                .target(name: "Control", swiftSettings: [.enableUpcomingFeature("ExistentialAny"), .enableUpcomingFeature("MemberImportVisibility")]),
                .testTarget(name: "ControlTests", dependencies: ["Control"]),
            ]
        )

        """

    static let files: [String: String] = [
        "Control/Helpers.swift": """
        private func unusedHelper() -> Int { 1 }
        fileprivate let unusedConstant = 3
        private func countdown(_ count: Int) -> Int { count == 0 ? 0 : countdown(count - 1) }
        private func usedHelper() -> Int { 2 }

        func entry() -> Int { usedHelper() }
        func internalUnused() {}

        """,
        "Control/Types.swift": """
        /* Never built. */
        private struct Lonely {
            func inside() {}
        }

        private class Base {
            func speak() {}
        }

        private final class Derived: Base {
            override func speak() {}
        }

        func talk() {
            Derived().speak()
        }

        """,
        "Control/Panel.swift": """
        import SwiftUI

        struct Panel: View {
            @State private var onlyProjected = false

            var body: some View {
                Toggle("Flag", isOn: $onlyProjected)
            }
        }

        private extension Panel {
            func extensionHelper() {}
        }

        """,
        "Control/Payload.swift": """
        import Foundation

        private struct Payload: Codable {
            var kept: Int
            var neverRead = 0

            private enum CodingKeys: String, CodingKey {
                case kept
                case neverRead
            }
        }

        func encoded() -> Data? {
            try? JSONEncoder().encode(Payload(kept: 1))
        }

        """,
        "Control/Holder.swift": """
        import Foundation

        final class Holder {
            private let token = NotificationCenter.default.addObserver(forName: nil, object: nil, queue: nil) { _ in }
            private var plain = 0
        }

        func makeHolder() -> Holder { Holder() }

        """,
        "Control/Witness.swift": """
        import AppKit

        private struct Described: CustomStringConvertible {
            var description: String { "described" }
        }

        final class Target: NSObject {
            @objc private func clicked() {}
        }

        func describe() -> String { String(describing: Described()) }

        """,
        "Control/Constants.swift": """
        private struct Constants {
            var scale: Float
            var padding: Int32 = 0
        }

        func constantsSize() -> Int { MemoryLayout<Constants>.stride + Int(Constants(scale: 1).scale) }

        """,
        "Control/Conformances.swift": """
        import Foundation

        protocol Stored: Codable {}
        protocol Named {}
        typealias Wire = Codable & Sendable

        private struct ViaOurs: Stored {
            var kept = 0
            var neverRead = 0
        }

        private struct Wired: Wire {
            var kept = 0
            var neverRead = 0
        }

        private struct Tag: Named, Hashable {
            var label = "tag"
            var unusedField = 0
        }

        private struct Untagged: Named {
            var label = "tag"
            var unusedField = 0
        }

        final class Watcher: NSObject {
            private var count = 0
        }

        struct Later {
            private var hidden = 0
        }

        func conformances() -> [Any] {
            [ViaOurs().kept, Wired().kept, Tag().label, Untagged().label, Watcher(), Later()]
        }

        """,
        "Control/Requests.swift": """
        /* The wire keeps `maxBytes` and `file`: only the synthesized Codable reads the per-case keys that say so. */
        enum Request: Codable {
            case fetch(sessionId: Int, maximumBytes: Int)
            case store(path: String)

            enum FetchCodingKeys: String, CodingKey {
                case sessionId
                case maximumBytes = "maxBytes"
            }
        }

        extension Request {
            enum StoreCodingKeys: String, CodingKey {
                case path = "file"
            }
        }

        /* The same keys with no Codable to read them. */
        enum Plain {
            case fetch(sessionId: Int, maximumBytes: Int)

            enum FetchCodingKeys: String, CodingKey {
                case sessionId
                case maximumBytes = "maxBytes"
            }
        }

        func requests() -> [Any] {
            [Request.fetch(sessionId: 1, maximumBytes: 2), Request.store(path: "/"), Plain.fetch(sessionId: 1, maximumBytes: 2)]
        }

        """,
        "Control/LaterEncoding.swift": """
        extension Later: Encodable {
            func encode(to encoder: any Encoder) throws {}
        }

        """,
        "Control/Conditional.swift": """
        private func stamp() -> Int { 1 }

        #if DEBUG
        func debugStamp() -> Int { stamp() }
        #endif

        #if CONTROL_NEVER_SET
        func unbuilt() -> Int { onlyInConditional() }
        #endif

        """,
        "Control/Entry.swift": """
        /* The library's API: public in a library product, so whoever depends on it may call it. */
        public func run() -> Int {
            talk()
            _ = encoded()
            _ = makeHolder()
            _ = describe()
            _ = conformances()
            _ = requests()
            _ = Panel()
            _ = Target()
            return entry() + constantsSize() + Meter().report().count
        }

        public func neverCalledButPublic() {}

        func onlyTests() -> Int { 1 }

        func onlyInConditional() -> Int { 1 }

        @propertyWrapper
        struct Logged {
            var wrappedValue: Int
            var projectedValue: String { "logged" }
        }

        final class Meter {
            @Logged private var level = 1

            func report() -> String { $level }
        }

        """,
        "Tests/ControlTests/UsesTests.swift": """
        import Testing

        @testable import Control

        @Test func callsTheFunctionOnlyTestsUse() {
            #expect(onlyTests() == 1)
        }

        """,
    ]

    /* One `--no-fix --unused` run: each unused-declaration record as `file:line subject`, its coverage record, the phase, and how many findings carried the rule. */
    static func run() async throws -> (items: [String], coverage: [String: Any], phase: String, findings: Int) {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-unused-declarations-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: root, withIntermediateDirectories: true)
        try manifest.write(to: root.appendingPathComponent("Package.swift"), atomically: true, encoding: .utf8)
        for (path, source) in files {
            let url = root.appendingPathComponent(path.hasPrefix("Tests/") ? path : "Sources/\(path)")
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
        var coverage: [String: Any] = [:]
        var phase = ""
        var findings = 0
        for line in stream.split(separator: UInt8(ascii: "\n")) {
            let record = try #require(try JSONSerialization.jsonObject(with: Data(line)) as? [String: Any])
            switch record["kind"] as? String {
                case "unused" where record["rule"] as? String == UnusedDeclarations.ruleName:
                    let file = (record["file"] as? String ?? "").components(separatedBy: "/Sources/").last ?? ""
                    items.append("\(file):\(record["line"] as? Int ?? 0) \(record["subject"] as? String ?? "")")
                case "unusedCoverage" where record["rule"] as? String == UnusedDeclarations.ruleName:
                    coverage = record
                case "phase" where record["name"] as? String == "unused":
                    phase = record["outcome"] as? String ?? ""
                case "finding" where record["rule"] as? String == UnusedDeclarations.ruleName:
                    findings += 1
                default:
                    break
            }
        }
        return (items, coverage, phase, findings)
    }

    @Test func reportsExactlyTheDeclarationsNothingUses() async throws {
        let run = try await Self.run()
        #expect(run.phase == "ran")
        /*
         `countdown` calls only itself, which is no use. `Lonely` holds `inside`, and only `Lonely` is reported, so
         the two removals never overlap. `Base.speak` is reached only through the override that `talk` calls, and
         removing it breaks the override. `onlyProjected` is read only as `$onlyProjected`. `token` holds an
         observer for as long as a `Holder` lives, so it is never reported; `plain` holds a number nobody reads.
         `internalUnused` is internal, and nothing in the package or its tests calls it.
         `Constants.padding` is never named, but a struct of plain numbers is read by its bytes, and the padding
         holds the layout. `Tag.unusedField` is never reported: Tag is Hashable, and its synthesized `==` and
         `hash(into:)` read every stored property, so removing it would make two tags that differ by it equal.
         `Untagged.unusedField`, the same field with no Hashable, is reported. `Plain.FetchCodingKeys` is reported:
         Plain is not Codable, so nothing reads its keys, while Request's, in its body and in an extension, keep
         its wire keys.
         */
        #expect(
            run.items.sorted() == [
                "Control/Conformances.swift:24 var unusedField",
                "Control/Helpers.swift:1 func unusedHelper()",
                "Control/Helpers.swift:2 let unusedConstant",
                "Control/Helpers.swift:3 func countdown(_:)",
                "Control/Helpers.swift:7 func internalUnused()",
                "Control/Holder.swift:5 var plain",
                "Control/Panel.swift:12 func extensionHelper()",
                "Control/Requests.swift:22 enum FetchCodingKeys",
                "Control/Types.swift:2 struct Lonely",
            ],
            "\(run.items)",
        )
        #expect(run.coverage["found"] as? Int == 9)
        let notChecked = run.coverage["filesNotChecked"] as? [String: Int] ?? [:]
        #expect(notChecked == [UnusedImports.conditional: 1], "only the #if file is left unchecked: \(notChecked)")
        let skipped = run.coverage["skipped"] as? [String: Int] ?? [:]
        #expect(
            skipped[UnusedDeclarations.codingKeys] == 10,
            "Payload's CodingKeys and its two cases, Request's two per-case enums and their three cases, and Plain's two cases, which go with their enum: \(skipped)",
        )
        #expect(skipped[UnusedDeclarations.synthesizedEquality] == 2, "Tag's two fields: \(skipped)")
        /*
         Codable reaches every stored property: Payload's two directly, ViaOurs's two through `Stored`, a protocol of
         ours that refines Codable, Wired's two through `Wire`, a type alias the index spells as Encodable and
         Decodable, and Later's one through an extension in another file. Watcher inherits from NSObject, an
         Objective-C class the index cannot see into, so its field is never reported either.
         */
        #expect(skipped[UnusedDeclarations.reflectedStorage] == 7, "\(skipped)")
        #expect(skipped[UnusedDeclarations.unseenConformance] == 1, "Watcher's count: \(skipped)")
        #expect(skipped[UnusedDeclarations.lifetime] == 1, "Holder's token: \(skipped)")
        #expect(skipped[UnusedDeclarations.objectiveC] == 1, "Target's @objc method: \(skipped)")
        #expect(skipped[UnusedDeclarations.layout] == 2, "Constants' two fields: \(skipped)")
        #expect(
            skipped[UnusedDeclarations.overrides] == 4,
            "Derived's override and the witnesses Described.description, Panel.body and Later.encode(to:): \(skipped)",
        )
        /*
         The package-wide kinds. `run` and `neverCalledButPublic` are public in a library product. `onlyInConditional`
         is named only in a `#if` branch the build never compiled. Target and Watcher descend from NSObject, which the
         Objective-C runtime can find by name. Panel's `@State` is a view's, which SwiftUI compares. `onlyTests` is
         called only from the test target, and `Meter.level` is read only as `$level`: neither is reported.
         */
        #expect(skipped[UnusedDeclarations.publicAPI] == 2, "\(skipped)")
        #expect(skipped[UnusedDeclarations.hiddenName] == 1, "\(skipped)")
        #expect(skipped[UnusedDeclarations.objectiveCClass] == 2, "\(skipped)")
        #expect(skipped[UnusedDeclarations.viewStorage] == 1, "\(skipped)")
        /* A report, not a gate: nothing it found is a finding. */
        #expect(run.findings == 0, "unused-declaration findings must be unused records, never finding records")
    }

    /* The suggested removal takes the declaration's lines and the comment written above it, and cuts a declaration that shares its line to its own text. */
    @Test func removalTakesTheLinesAndTheCommentAbove() throws {
        let source = """
            struct Kept {}

            // MARK: - Helpers
            /* Explains the next line. */
            private func gone() {}
            private struct Inline { func first() {}; func second() {} }

            """
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Removal.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let function = try #require(file.tree.statements.compactMap { $0.item.as(FunctionDeclSyntax.self) }.first)
        let edit = UnusedDeclarations.removal(of: Syntax(function), in: file)
        var bytes = Array(source.utf8)
        bytes.replaceSubrange(edit.start..<edit.end, with: Array(edit.text.utf8))
        #expect(
            String(decoding: bytes, as: UTF8.self)
                == "struct Kept {}\n\n// MARK: - Helpers\nprivate struct Inline { func first() {}; func second() {} }\n"
        )

        let inline = try #require(file.tree.statements.compactMap { $0.item.as(StructDeclSyntax.self) }.last)
        let member = try #require(inline.memberBlock.members.first?.decl)
        let cut = UnusedDeclarations.removal(of: Syntax(member), in: file)
        #expect(String(decoding: Array(source.utf8)[cut.start..<cut.end], as: UTF8.self) == "func first() {}")
    }

    /* A comment trailing the declaration's line goes with it, and so does one of the two blank lines it stood between. Found on ahraos-macos, where `// unused` trailed the property and the cut kept it and orphaned the doc comment above. */
    @Test func removalTakesATrailingCommentAndOneBlankLine() throws {
        let source = """
            final class Holder {
                var kept = 0

                /// Says what the next line holds.
                private var gone: [Int] = [] // unused

                var alsoKept = 1
            }

            """
        let file = ParsedFile(
            url: URL(fileURLWithPath: "/Trailing.swift"),
            targetName: "Control",
            targetKind: "regular",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        let holder = try #require(file.tree.statements.compactMap { $0.item.as(ClassDeclSyntax.self) }.first)
        let member = try #require(holder.memberBlock.members.dropFirst().first?.decl)
        let edit = UnusedDeclarations.removal(of: Syntax(member), in: file)
        var bytes = Array(source.utf8)
        bytes.replaceSubrange(edit.start..<edit.end, with: Array(edit.text.utf8))
        #expect(
            String(decoding: bytes, as: UTF8.self)
                == "final class Holder {\n    var kept = 0\n\n    var alsoKept = 1\n}\n"
        )
    }
}
