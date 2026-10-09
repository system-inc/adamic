import Foundation
import SwiftFormat
import Testing

@testable import CohereSwift

/*
 The house Swift format holds exactly what was ruled (#mg4dgjm, 2026-10-03), and nothing configures it
 anywhere else: a `.swift-format` at or above a file of ours, or a format key in the project's settings, is
 refused by name. The values are written out here rather than read from the house, so changing a setting
 fails this test before it reformats every repository.
 */
struct HouseSwiftFormatTests {
    @Test func theHouseHoldsTheRuledSettings() {
        let house = HouseSwiftFormat.configuration
        /* The `.swift-format` both proving grounds carried. */
        #expect(house.lineLength == 120)
        #expect(house.indentation == .spaces(4))
        #expect(house.tabWidth == 8)
        #expect(house.maximumBlankLines == 1)
        #expect(house.spacesBeforeEndOfLineComments == 1)
        #expect(house.lineBreakBeforeControlFlowKeywords)
        #expect(house.indentConditionalCompilationBlocks)
        #expect(!house.lineBreakBetweenDeclarationAttributes)
        #expect(!house.lineBreakAroundMultilineExpressionChainComponents)
        #expect(house.fileScopedDeclarationPrivacy.accessLevel == .private)
        #expect(house.multiElementCollectionTrailingCommas)
        #expect(house.reflowMultilineStringLiterals == .never)
        #expect(!house.indentBlankLines)
        /* The ruling's changes. */
        #expect(house.multilineTrailingCommaBehavior == .alwaysUsed)
        #expect(house.lineBreakBeforeEachArgument)
        #expect(house.prioritizeKeepingFunctionOutputTogether)
        #expect(house.lineBreakBeforeEachGenericRequirement)
        #expect(house.indentSwitchCaseLabels)
        /* Kept as they were, by ruling. */
        #expect(house.respectsExistingLineBreaks)
        #expect(!house.spacesAroundRangeFormationOperators)
    }

    @Test func theHouseRuleTableIsTheOneTheProvingGroundsCarried() {
        let rules = HouseSwiftFormat.configuration.rules
        /* Every rule swift-format has is named, so a rule a later swift-format adds fails here rather than arriving by accident. */
        #expect(Set(rules.keys) == Set(Configuration.defaultRuleEnablements.keys))
        let on: Set<String> = [
            "AlwaysUseLowerCamelCase",
            "DoNotUseSemicolons",
            "OneVariableDeclarationPerLine",
            "OrderedImports",
            "ReturnVoidInsteadOfEmptyTuple",
            "UseShorthandTypeNames",
            "UseSingleLinePropertyGetter",
            "UseTripleSlashForDocumentationComments",
        ]
        #expect(Set(rules.filter(\.value).keys) == on)
    }

    /* A throwaway directory, with `.git` at its top when it stands for a repository. */
    static func directory(repository: Bool) throws -> URL {
        let directory = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-house-\(UUID().uuidString)",
            isDirectory: true,
        )
        try FileManager.default.createDirectory(at: directory, withIntermediateDirectories: true)
        if repository {
            try FileManager.default.createDirectory(
                at: directory.appendingPathComponent(".git", isDirectory: true),
                withIntermediateDirectories: true,
            )
        }
        return directory.resolvingSymlinksInPath()
    }

    static func write(_ text: String, to file: URL) throws {
        try FileManager.default.createDirectory(
            at: file.deletingLastPathComponent(),
            withIntermediateDirectories: true,
        )
        try text.write(to: file, atomically: true, encoding: .utf8)
    }

    @Test func aSwiftFormatAtOrAboveAFileUpToTheBoundaryIsFound() throws {
        let repository = try Self.directory(repository: true)
        let package = repository.appendingPathComponent("Apps/Tool", isDirectory: true)
        let files = [
            package.appendingPathComponent("Sources/Tool/A.swift"),
            package.appendingPathComponent("Sources/Tool/B.swift"),
        ]
        #expect(HouseSwiftFormat.leftoverConfigurationFiles(above: files, boundary: repository).isEmpty)
        try Self.write("{}", to: repository.appendingPathComponent(".swift-format"))
        try Self.write("{}", to: package.appendingPathComponent("Sources/.swift-format"))
        #expect(
            HouseSwiftFormat.leftoverConfigurationFiles(above: files, boundary: repository).map(\.path)
                == [
                    package.appendingPathComponent("Sources/.swift-format").path,
                    repository.appendingPathComponent(".swift-format").path,
                ].sorted()
        )
    }

    @Test func aSwiftFormatAboveTheBoundaryIsNotOursToRefuse() throws {
        let outer = try Self.directory(repository: false)
        try Self.write("{}", to: outer.appendingPathComponent(".swift-format"))
        let repository = outer.appendingPathComponent("Repository", isDirectory: true)
        let file = repository.appendingPathComponent("Sources/A.swift")
        #expect(HouseSwiftFormat.leftoverConfigurationFiles(above: [file], boundary: repository).isEmpty)
        /* A sibling whose name extends the boundary's is not inside it. */
        try Self.write("{}", to: outer.appendingPathComponent("RepositoryOther/.swift-format"))
        #expect(
            HouseSwiftFormat.leftoverConfigurationFiles(
                above: [outer.appendingPathComponent("RepositoryOther/A.swift")],
                boundary: repository,
            ).isEmpty
        )
    }

    @Test func aRunOverAPackageWithASwiftFormatIsRefusedByName() async throws {
        let root = try Self.directory(repository: true)
        try Self.write(PipelineControlTests.manifest, to: root.appendingPathComponent("Package.swift"))
        try Self.write(
            PipelineControlTests.cleanSource,
            to: root.appendingPathComponent("Sources/Control/Control.swift"),
        )
        let leftover = root.appendingPathComponent(".swift-format")
        try Self.write(#"{"version":1,"lineLength":120}"#, to: leftover)
        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path, "--no-fix"],
            workingDirectory: root,
        )
        var lines = Data()
        let writer = ContractWriter { lines.append($0) }
        do {
            _ = try await Pipeline(options: options, writer: writer, workingDirectory: root).run()
            Issue.record("a package with a .swift-format was checked as if the file were not there")
        }
        catch let failure as Pipeline.RunFailure {
            #expect(failure.description.hasPrefix(leftover.path + " is a .swift-format"))
            #expect(failure.description.contains("delete it"))
        }
        /* Refused before anything is checked: no phase and no summary, so the front door says nothing was. */
        let kinds = String(decoding: lines, as: UTF8.self).split(separator: "\n").compactMap { line in
            (try? JSONSerialization.jsonObject(with: Data(line.utf8)) as? [String: Any])?["kind"] as? String
        }
        #expect(!kinds.contains("phase"))
        #expect(!kinds.contains("summary"))
    }

    @Test func aFormatKeyInTheProjectSettingsIsRefusedByName() {
        for (settings, key) in [
            (#"{"format":{"printWidth":100},"swift":{"rules":{}}}"#, "\"format\" configures formatting"),
            (
                #"{"swift":{"format":{"lineLength":100},"rules":{}}}"#,
                "the swift block's \"format\" configures formatting",
            ),
        ] {
            do {
                _ = try RuleConfiguration.parse(Data(settings.utf8), path: "/project/CohereSettings.json")
                Issue.record("\(settings) was read as if its format key were not there")
            }
            catch let failure as RuleConfiguration.ReadFailure {
                #expect(failure.path == "/project/CohereSettings.json")
                #expect(failure.reason.hasPrefix(key))
                #expect(failure.reason.contains("remove the key"))
            }
            catch {
                Issue.record("refused, but not as a settings failure: \(error)")
            }
        }
    }

    @Test func settingsWithoutAFormatKeyStillRead() throws {
        let configuration = try RuleConfiguration.parse(
            Data(#"{"swift":{"rules":{"cohere-swift/consistency-no-print":"off"}}}"#.utf8),
            path: "/project/CohereSettings.json",
        )
        #expect(configuration.severity(of: "cohere-swift/consistency-no-print") == .off)
    }
}
