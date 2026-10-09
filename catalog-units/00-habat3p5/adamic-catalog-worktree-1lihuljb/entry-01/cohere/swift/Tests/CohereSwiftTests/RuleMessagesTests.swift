import CryptoKit
import Foundation
import Testing

@testable import CohereSwift

/*
 The compiled-in house messages (`RuleMessages`, generated from cohere's policy/messages/ by
 `go run ./policy/tools/generate`) against the files they were made from, and against the rules that render them.
 */
struct RuleMessagesTests {
    /* `swift/Tests/CohereSwiftTests/` up to cohere's root. */
    static let repositoryRoot = URL(fileURLWithPath: #filePath)
        .deletingLastPathComponent().deletingLastPathComponent().deletingLastPathComponent()
        .deletingLastPathComponent()

    /*
     The digest the generator records (policy/swift_messages.go, swiftMessagesDigest), recomputed here from the files
     on disk by a resolver of its own: for each Swift rule and each of its messages, the rule, the id and the text with
     its Swift terms and string phrases in place, then each object phrase it picks from as `<<name>>` with each
     option's name and its Swift text. Every field ends with a zero byte, and everything is in name order. A message
     whose `languages` leave out Swift does not reach it, nor does an option with no Swift text, so a TypeScript-only
     edit changes nothing here.
     */
    static func digestOnDisk() throws -> String {
        let directory = repositoryRoot.appendingPathComponent("policy/messages")
        var entries: [(rule: String, id: String, fields: [String])] = []
        for name in try FileManager.default.contentsOfDirectory(atPath: directory.path) where name.hasSuffix(".json") {
            let data = try Data(contentsOf: directory.appendingPathComponent(name))
            guard let file = try JSONSerialization.jsonObject(with: data) as? [String: Any],
                let rule = (file["rules"] as? [String: String])?["Swift"]
            else { continue }
            let phrases = file["phrases"] as? [String: Any] ?? [:]
            for (id, value) in file["messages"] as? [String: [String: Any]] ?? [:] {
                if let languages = value["languages"] as? [String], !languages.contains("Swift") {
                    continue
                }
                var text = value["text"] as? String ?? ""
                for (term, words) in value["terms"] as? [String: [String: String]] ?? [:] {
                    text = text.replacingOccurrences(of: "[[\(term)]]", with: words["Swift"] ?? "")
                }
                var picked: [String: [String: String]] = [:]
                for (phrase, phraseValue) in phrases where text.contains("<<\(phrase)>>") {
                    if let shared = phraseValue as? String {
                        text = text.replacingOccurrences(of: "<<\(phrase)>>", with: shared)
                    }
                    else if let options = phraseValue as? [String: Any] {
                        /* An option is text for every language, or text by language; one with no Swift text does not reach Swift. */
                        picked[phrase] = options.compactMapValues { option in
                            option as? String ?? (option as? [String: String])?["Swift"]
                        }
                    }
                }
                var fields = [rule, id, text]
                for phrase in picked.keys.sorted(by: Self.byBytes) {
                    fields.append("<<\(phrase)>>")
                    let options = picked[phrase] ?? [:]
                    for option in options.keys.sorted(by: Self.byBytes) {
                        fields += [option, options[option] ?? ""]
                    }
                }
                entries.append((rule, id, fields))
            }
        }
        var hash = SHA256()
        let ordered = entries.sorted { left, right in
            left.rule == right.rule ? Self.byBytes(left.id, right.id) : Self.byBytes(left.rule, right.rule)
        }
        for entry in ordered {
            for field in entry.fields {
                hash.update(data: Data(field.utf8) + Data([0]))
            }
        }
        return hash.finalize()
            .map { byte in
                let digits = String(byte, radix: 16)
                return digits.count == 1 ? "0" + digits : digits
            }
            .joined()
    }

    /* Go's order: by UTF-8 bytes, not by Swift's Unicode-aware comparison. */
    static func byBytes(_ left: String, _ right: String) -> Bool {
        left.utf8.lexicographicallyPrecedes(right.utf8)
    }

    /* A message file edited and not regenerated fails here, so the Swift rules cannot print words the catalog no longer holds. */
    @Test func theMessagesWereGeneratedFromTheFilesOnDisk() throws {
        #expect(RuleMessages.sourceDigest == (try Self.digestOnDisk()), "run go run ./policy/tools/generate")
    }

    /* Every Swift source except the generated catalog, as one text, for the test below to search. */
    static func engineSources() throws -> String {
        let sources = repositoryRoot.appendingPathComponent("swift/Sources")
        guard let walker = FileManager.default.enumerator(at: sources, includingPropertiesForKeys: nil) else {
            return ""
        }
        var text = ""
        for case let file as URL in walker
        where file.pathExtension == "swift" && !file.lastPathComponent.hasSuffix(".generated.swift") {
            text += try String(contentsOf: file, encoding: .utf8)
        }
        return text
    }

    /* `RuleMessages.Rule.id(`, the call a rule makes. */
    static func isCalled(_ entry: String, in sources: String) -> Bool {
        sources.contains("RuleMessages.\(entry)(")
    }

    /*
     An entry no rule renders fails, as the Go registry test fails on one. The positive control comes first: the
     search finds bare-throw's call, which a rule makes, and does not find a name nothing calls, so a search that
     matches nothing, or everything, cannot pass the check below.
     */
    @Test func everyMessageIsRenderedByARule() throws {
        let sources = try Self.engineSources()
        #expect(Self.isCalled("ConsistencyNoBareThrow.bareThrow", in: sources))
        #expect(!Self.isCalled("ConsistencyNoBareThrow.noSuchMessage", in: sources))
        #expect(!RuleMessages.all.isEmpty)
        for entry in RuleMessages.all {
            #expect(Self.isCalled(entry, in: sources), "\(entry) is in the catalog and no rule renders it")
        }
    }
}
