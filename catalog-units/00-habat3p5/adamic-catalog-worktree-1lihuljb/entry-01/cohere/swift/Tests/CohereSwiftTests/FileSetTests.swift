import Foundation
import Testing

@testable import CohereSwift

/* An app is decided per target: one file importing a UI framework makes every file in the target app code. */
struct FileSetTests {
    static func target(kind: String, files: [String: String]) throws -> PackageModel.Target {
        let directory = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-target-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        var sources: [URL] = []
        for (name, text) in files {
            let url = directory.appendingPathComponent(name)
            try text.write(to: url, atomically: true, encoding: .utf8)
            sources.append(url)
        }
        return PackageModel.Target(
            name: "Subject",
            kind: kind,
            directory: directory,
            sources: sources,
            languageMode: "6",
        )
    }

    @Test func oneUserInterfaceImportMakesTheWholeTargetAnApp() throws {
        let app = try Self.target(
            kind: "executable",
            files: [
                "Model.swift": "import Foundation\n",
                "Window.swift": "import Foundation\n@preconcurrency import AppKit\n",
            ],
        )
        #expect(try FileSet.isApplication(app))
    }

    @Test func aToolWithoutUserInterfaceIsNotAnApp() throws {
        let tool = try Self.target(
            kind: "executable",
            files: ["main.swift": "import Foundation\n// import SwiftUI is only mentioned here\n"],
        )
        #expect(try !FileSet.isApplication(tool))
    }

    @Test func aLibraryIsNeverAnApp() throws {
        let library = try Self.target(kind: "library", files: ["View.swift": "import SwiftUI\n"])
        #expect(try !FileSet.isApplication(library))
    }
}
