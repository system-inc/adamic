import Foundation
import Testing

@testable import CohereSwift

/*
 The session against the toolchain's real sourcekitd, so the hand-declared C layouts are proven against the
 library rather than against each other. Every direction a session could be quietly wrong in: a clean
 module must come back empty, an injected error and a warning must come back placed where `.dia` places
 them, a name declared in another file of the module must resolve (or the session is checking files
 alone), and expression types must be real types, or a typed rule asking them would judge nothing.

 Serialized, although the session locks each question itself, so a failure here reads as the session's own
 rather than as two tests' requests interleaving.
 */
@Suite(.serialized)
struct SourcekitdTests {
    /* A two-file module on disk: `Shape.swift` declares a type and `Use.swift` uses it. */
    static func module(use: String) throws -> (use: String, arguments: [String]) {
        let directory = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-sourcekitd-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        let shape = directory.appendingPathComponent("Shape.swift")
        try "struct Shape {\n    var sides: Int\n}\n".write(to: shape, atomically: true, encoding: .utf8)
        let useFile = directory.appendingPathComponent("Use.swift")
        try use.write(to: useFile, atomically: true, encoding: .utf8)
        let sdk = try ProcessRunner().run("xcrun", ["--show-sdk-path"], in: directory)
        let sdkPath = String(decoding: sdk.standardOutput, as: UTF8.self).trimmingCharacters(
            in: .whitespacesAndNewlines
        )
        let arguments = [
            "-module-name", "Subject", "-parse-as-library", "-swift-version", "6", "-sdk", sdkPath,
            shape.resolvingSymlinksInPath().path, useFile.resolvingSymlinksInPath().path,
        ]
        return (useFile.resolvingSymlinksInPath().path, arguments)
    }

    static func session() throws -> Sourcekitd {
        try Sourcekitd.shared()
    }

    @Test func aCleanFileThatUsesItsModuleIsEmpty() throws {
        let module = try Self.module(use: "func triangle() -> Shape {\n    Shape(sides: 3)\n}\n")
        #expect(try Self.session().diagnostics(file: module.use, arguments: module.arguments).isEmpty)
    }

    @Test func anInjectedErrorComesBackAtItsLine() throws {
        let module = try Self.module(use: "func triangle() -> Shape {\n    Shape(sides: 3)\n}\n")
        let injected = "func triangle() -> Shape {\n    Shape(sides: 3)\n}\nlet broken: Int = \"nope\"\n"
        let diagnostics = try Self.session().diagnostics(file: module.use, arguments: module.arguments, text: injected)
        #expect(diagnostics.map(\.findingSeverity) == [.error])
        #expect(diagnostics.first?.line == 4)
        #expect(diagnostics.first?.message.contains("cannot convert") == true)
    }

    /* The `.dia` test's warning, at the `.dia` test's position, so the two readers place a diagnostic the same way. */
    @Test func aWarningCarriesItsPositionAndGroupLink() throws {
        let module = try Self.module(
            use: "func subject() -> Int {\n    var neverMutated = 3\n    return neverMutated\n}\n"
        )
        let diagnostics = try Self.session().diagnostics(file: module.use, arguments: module.arguments)
        #expect(diagnostics.map(\.findingSeverity) == [.warning])
        #expect(diagnostics.first?.line == 2)
        #expect(diagnostics.first?.column == 9)
        #expect(diagnostics.first.map(TypesOracle.group(of:)) == "VariableNeverMutated")
    }

    /* Without the module's other file, `Shape` would not resolve: the session must check a file in its module. */
    @Test func aNameFromAnotherFileOfTheModuleResolves() throws {
        let module = try Self.module(use: "let square = Shape(sides: 4)\n")
        let alone = module.arguments.filter { !$0.hasSuffix("Shape.swift") }
        #expect(try Self.session().diagnostics(file: module.use, arguments: module.arguments).isEmpty)
        #expect(
            try Self.session().diagnostics(file: module.use, arguments: alone).contains { $0.message.contains("Shape") }
        )
    }

    @Test func expressionTypesAreTheCompilersTypes() throws {
        let source = "let names = [\"a\", \"b\"]\nlet total = names.count\n"
        let module = try Self.module(use: source)
        let types = try Self.session().expressionTypes(file: module.use, arguments: module.arguments)
        let utf8 = Array(source.utf8)
        func type(of spelled: String) -> String? {
            types.first { String(decoding: utf8[$0.offset..<$0.offset + $0.length], as: UTF8.self) == spelled }?.type
        }
        #expect(type(of: "[\"a\", \"b\"]") == "[String]")
        #expect(type(of: "names.count") == "Int")
    }
}
