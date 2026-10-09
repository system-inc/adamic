import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 The naming rules. The vocabulary tests are the Go rule's own vectors (`abbreviation_vocabulary_test.go`),
 so the two engines are held to the same answers on the same file.
 */
struct NamingRuleTests {
    static func vocabulary() throws -> AbbreviationVocabulary {
        try AbbreviationVocabulary.compiledIn()
    }

    static func file(_ source: String, name: String = "Subject.swift") -> ParsedFile {
        ParsedFile(
            url: URL(fileURLWithPath: "/fixture/\(name)"),
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
    }

    /* `name:line` for each finding, so an expectation reads as what was flagged and where. */
    static func flagged(_ rule: some FileRule, _ source: String, name: String = "Subject.swift") -> [String] {
        rule.findings(in: file(source, name: name)).map { finding in
            /* swift-format quotes the name last in single quotes ("remove the leading '_' from the name '_cache'"); the vocabulary quotes it first in double quotes. */
            let singleQuoted = finding.message.split(separator: "'", omittingEmptySubsequences: false)
            let quoted =
                singleQuoted.count >= 3
                ? String(singleQuoted[singleQuoted.count - 2])
                : finding.message.split(separator: "\"").dropFirst().first.map(String.init) ?? ""
            return "\(quoted):\(finding.line)"
        }
    }

    @Test func theSharedFileCarriesEveryFormInOrder() throws {
        let vocabulary = try Self.vocabulary()
        #expect(vocabulary.earlyPrefixes.map(\.abbreviation) == ["ctx", "db", "tx", "opts", "cur", "pct", "prev"])
        #expect(
            vocabulary.segments.map(\.abbreviation) == [
                "cwd", "dir", "env", "cli", "len", "seq", "db", "tx", "vars", "var",
            ]
        )
        #expect(vocabulary.suffixes.map(\.abbreviation).prefix(5) == ["prop", "props", "param", "params", "ms"])
        #expect(vocabulary.allowedNames == ["URLSearchParams"])
        #expect(vocabulary.allowedSegments == ["InnoDb"])
    }

    /* The Go test's count: every form of every entry judged once, 92 in all. */
    @Test func everyFormIsJudged() throws {
        let vocabulary = try Self.vocabulary()
        var judged = 0
        for word in vocabulary.wholeByName.keys {
            #expect(
                vocabulary.find(word)?.form == "whole" && vocabulary.find(word)?.messageId == "abbreviatedIdentifier",
                "whole \(word)",
            )
            judged += 1
        }
        for entry in vocabulary.earlyPrefixes + vocabulary.latePrefixes {
            let name = entry.abbreviation + "Widget"
            #expect(
                vocabulary.find(name)?.form == "prefix" && vocabulary.find(name)?.messageId == "abbreviatedIdentifier",
                "prefix \(name)",
            )
            judged += 1
        }
        for entry in vocabulary.suffixes {
            let name = "widget" + AbbreviationVocabulary.capitalized(entry.abbreviation)
            #expect(vocabulary.find(name)?.form == "suffix", "suffix \(name)")
            judged += 1
        }
        for entry in vocabulary.segments {
            let name = "widget" + AbbreviationVocabulary.capitalized(entry.abbreviation) + "Name"
            #expect(
                vocabulary.find(name)?.form == "segment" && vocabulary.find(name)?.abbreviation == entry.abbreviation,
                "segment \(name)",
            )
            judged += 1
        }
        #expect(judged == 92)
    }

    @Test(arguments: [
        ("themeVarsVar", "abbreviatedWordSegment", #"Use "themeVariablesVar"."#),
        ("timeoutMsRef", "millisecondSuffix", #"Use "timeoutInMillisecondsRef""#),
        ("dbVal", "abbreviatedIdentifier", #"Use "databaseVal"."#),
    ])
    func orderIsObservableWhereItMatters(name: String, messageId: String, advice: String) throws {
        let finding = try #require(try Self.vocabulary().find(name))
        #expect(finding.messageId == messageId)
        #expect(finding.message.contains(advice), "\(finding.message)")
    }

    @Test func allowedNamesAndSegmentsAreNotJudged() throws {
        let vocabulary = try Self.vocabulary()
        #expect(vocabulary.find("URLSearchParams") == nil)
        #expect(vocabulary.find("InnoDbMaximumIndexKeyBytes") == nil)
        #expect(vocabulary.find("properties") == nil)
        #expect(vocabulary.find("durationInMilliseconds") == nil)
    }

    @Test(arguments: [
        #"{"abbreviations": [{"abbreviation": "val", "expansion": "value", "suffx": {}}]}"#,
        #"{"abbreviations": [{"abbreviation": "val", "expansion": "value"}]}"#,
        #"{"abbreviations": [{"abbreviation": "val", "expansion": "value", "whole": {"style": "loud"}}]}"#,
        #"{"abbreviations": [{"abbreviation": "val", "whole": {"style": "plain"}}]}"#,
        #"{"abbreviations": [{"abbreviation": "val", "expansion": "value", "prefix": {"phase": "middle"}}]}"#,
        #"{"abbreviations": [{"abbreviation": "Val", "expansion": "value", "segment": {}}]}"#,
        #"{"abbreviations": [{"abbreviation": "val", "expansion": "value", "segment": {}}, {"abbreviation": "val", "expansion": "value", "segment": {}}]}"#,
        #"{"abbreviations": [{"abbreviation": "ms", "expansion": "m", "suffix": {"matcher": "seconds"}}]}"#,
        #"{"abbreviations": [{"abbreviation": "val", "expansion": "value", "whole": {"messageId": "noVal", "style": "plain"}}]}"#,
    ])
    func aMalformedFileIsRefused(file: String) {
        #expect(throws: (any Error).self) { try AbbreviationVocabulary.load(data: Data(file.utf8)) }
    }

    /*
     The compiled-in words are cohere's `policy/Abbreviations.json`, byte for byte. A word added to the file
     and not regenerated (`go run ./policy/tools/generate`) fails here, so the two engines cannot drift.
     */
    @Test func theCompiledInVocabularyIsPolicysFile() throws {
        let policyFile = URL(fileURLWithPath: #filePath)
            .deletingLastPathComponent().deletingLastPathComponent().deletingLastPathComponent()
            .deletingLastPathComponent()
            .appendingPathComponent("policy/Abbreviations.json")
        let onDisk = try Data(contentsOf: policyFile)
        #expect(!onDisk.isEmpty)
        #expect(Data(PolicyAbbreviations.file.utf8) == onDisk, "run go run ./policy/tools/generate")
    }

    @Test func declaredNamesAreJudgedAndReferencesAreNot() throws {
        let source = """
            struct Config {
                let maxBytes: Int
                func load(params: [Int], from src: String, max maximum: Int) -> Int {
                    let idx = params.count + Int.max
                    return idx
                }
            }
            """
        #expect(
            Self.flagged(ConsistencyNoAbbreviatedIdentifier(vocabulary: try Self.vocabulary()), source) == [
                "Config:1", "maxBytes:2", "params:3", "idx:4",
            ]
        )
    }

    /* An override's name and single-name labels are the superclass's; a separate label is API and only the name after it is ours. */
    @Test func spellingsChosenElsewhereAreNotJudged() throws {
        let source = """
            class Pane: Base {
                override func setConfig(params: Int) {}
                override var maxWidth: Double { 1 }
                func requestOpenLink(source: String, params parameters: [String: String]) {}
                func update() {
                    if let config { use(config) }
                    let x = 1, i = 2
                }
            }
            """
        #expect(Self.flagged(ConsistencyNoAbbreviatedIdentifier(vocabulary: try Self.vocabulary()), source).isEmpty)
    }

    @Test func lowerCamelCaseIsSwiftFormatsOwnRule() {
        let source = """
            import XCTest
            let max_bytes = 1
            func LoadAll() {}
            final class ScrollTests: XCTestCase {
                func test_scrolls_to_bottom() {}
            }
            """
        #expect(Self.flagged(SwiftFormatRule.requireLowerCamelCase, source) == ["max_bytes:2", "LoadAll:3"])
    }

    /* swift-format's text, made a sentence, then the house's reason, filed under the catalog's id. */
    @Test func lowerCamelCaseSaysSwiftFormatsWordsAndTheReason() {
        let findings = SwiftFormatRule.requireLowerCamelCase.findings(in: Self.file("let max_bytes = 1"))
        #expect(findings.map(\.messageId) == ["alwaysUseLowerCamelCase"])
        #expect(
            findings.first?.message
                == "Rename the constant 'max_bytes' using lowerCamelCase. Swift spells values in lowerCamelCase and types in UpperCamelCase, so a reader tells which is which at a glance, and an underscore inside a name is a word boundary camel case already marks."
        )
    }

    @Test func leadingUnderscoresAreSwiftFormatsOwnRule() {
        #expect(
            Self.flagged(SwiftFormatRule.noLeadingUnderscores, "let _cache = 1\nlet _ = 2\nfunc _reset() {}\n") == [
                "_cache:1", "_reset:3",
            ]
        )
    }
}
