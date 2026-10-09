import Foundation
import Testing

@testable import CohereSwift

/*
 The reader against a real `.dia` the toolchain's own compiler writes, so the declared libclang layouts
 are proven against the library rather than against each other. Both directions: a file with a warning
 and an error must come back with both, and a clean file must come back empty, or the reader could be
 returning nothing for every file and pass.
 */
struct SerializedDiagnosticsReaderTests {
    static func serializedDiagnostics(for source: String) throws -> URL {
        let directory = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-dia-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        let file = directory.appendingPathComponent("Subject.swift")
        try source.write(to: file, atomically: true, encoding: .utf8)
        let record = directory.appendingPathComponent("Subject.dia")
        /*
         The frontend rather than the `swiftc` driver: measured, the driver accepts `-serialize-diagnostics-path`
         with `-typecheck` and writes nothing, which would make every test here fail on a missing file instead
         of on the reader.
         */
        let sdk = try ProcessRunner().run("xcrun", ["--show-sdk-path"], in: directory)
        let sdkPath = String(decoding: sdk.standardOutput, as: UTF8.self).trimmingCharacters(
            in: .whitespacesAndNewlines
        )
        _ = try ProcessRunner().run(
            "xcrun",
            [
                "swift-frontend", "-typecheck", "-primary-file", file.path, "-swift-version", "6", "-sdk", sdkPath,
                "-serialize-diagnostics-path", record.path,
            ],
            in: directory,
        )
        try #require(
            FileManager.default.fileExists(atPath: record.path),
            "the compiler wrote no serialized diagnostics, so this test would check nothing",
        )
        return record
    }

    static func reader() throws -> SerializedDiagnosticsReader {
        try SerializedDiagnosticsReader(libraryPath: SerializedDiagnosticsReader.toolchainLibraryPath())
    }

    @Test func aWarningCarriesItsGroupAndPosition() throws {
        let record = try Self.serializedDiagnostics(
            for: "func subject() -> Int {\n    var neverMutated = 3\n    return neverMutated\n}\n"
        )
        let diagnostics = try Self.reader().read(record).filter { $0.severity != nil }
        #expect(diagnostics.count == 1)
        #expect(diagnostics.first?.severity == .warning)
        #expect(diagnostics.first?.group == "VariableNeverMutated")
        #expect(diagnostics.first?.line == 2)
        #expect(diagnostics.first?.column == 9)
        #expect(diagnostics.first?.file.hasSuffix("Subject.swift") == true)
    }

    @Test func anErrorIsAnError() throws {
        let record = try Self.serializedDiagnostics(for: "let broken: Int = \"nope\"\n")
        let diagnostics = try Self.reader().read(record).filter { $0.severity != nil }
        #expect(diagnostics.map(\.severity) == [.error])
        #expect(diagnostics.first?.message.contains("cannot convert") == true)
    }

    @Test func aCleanFileIsEmpty() throws {
        let record = try Self.serializedDiagnostics(for: "let fine = 1\n")
        #expect(try !Self.reader().read(record).contains { $0.severity != nil })
    }

    @Test func notesAreNotFindings() {
        #expect(SerializedDiagnosticsReader.severity(1) == nil)
        #expect(SerializedDiagnosticsReader.severity(2) == .warning)
        #expect(SerializedDiagnosticsReader.severity(3) == .error)
        #expect(SerializedDiagnosticsReader.severity(4) == .error)
    }
}
