import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `private-over-fileprivate`, both directions, covering every example SwiftLint's
 `private_over_fileprivate` documents. The failing cases assert the exact line and column of the
 `fileprivate` keyword. The passing cases are the near misses: nested `fileprivate` members, `fileprivate(set)`
 inside a type, extensions, and every other access level at the top level. The fix cases check that the
 keyword alone is swapped, so `fileprivate(set)` keeps its `(set)`.
 */
struct PrivateOverFileprivateTests {
    static func findings(_ source: String) -> [FindingRecord] {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        let file = ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: "library",
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
        return PrivateOverFileprivate().findings(in: file)
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    /* Every fix applied, last first so earlier offsets stay true. */
    static func fixed(_ source: String) -> String {
        var bytes = Array(source.utf8)
        let edits = findings(source).flatMap(\.fixes).sorted { $0.start > $1.start }
        for edit in edits {
            bytes.replaceSubrange(edit.start..<edit.end, with: Array(edit.text.utf8))
        }
        return String(decoding: bytes, as: UTF8.self)
    }

    /* SwiftLint's four triggering examples, one per line where they share a fixture. */
    @Test func incumbentTriggeringExamplesAreFound() {
        let source = """
            fileprivate enum MyEnum {}
            fileprivate class MyClass {
              fileprivate(set) var myInt = 4
            }
            fileprivate actor MyActor {
              fileprivate let myInt = 4
            }
                fileprivate func f() {}
                fileprivate var x = 0
            """
        #expect(Self.positions(Self.findings(source)) == ["1:1", "2:1", "5:1", "8:5", "9:5"])
    }

    /* The remaining declaration kinds, after attributes and other modifiers, inside a top-level `#if`, and `fileprivate(set)` at file scope. */
    @Test func everyCheckedKindIsFound() {
        let source = """
            fileprivate struct Thing {}
            fileprivate protocol Marker {}
            fileprivate typealias Alias = Int
            fileprivate let (first, second) = (1, 2)
            @MainActor fileprivate final class Holder {}
            fileprivate(set) var counter = 0
            #if DEBUG
            fileprivate func debugOnly() {}
            #if os(macOS)
            fileprivate var nested = 0
            #endif
            #endif
            """
        #expect(Self.positions(Self.findings(source)) == ["1:1", "2:1", "3:1", "4:1", "5:12", "6:1", "8:1", "10:1"])
    }

    /* SwiftLint's eleven non-triggering examples. */
    @Test func incumbentNonTriggeringExamplesAreNot() {
        let sources = [
            "extension String {}",
            "private extension String {}",
            "public protocol P {}",
            "open extension \n String {}",
            "internal extension String {}",
            "package typealias P = Int",
            "extension String {\n  fileprivate func Something(){}\n}",
            "class MyClass {\n  fileprivate let myInt = 4\n}",
            "actor MyActor {\n  fileprivate let myInt = 4\n}",
            "class MyClass {\n  fileprivate(set) var myInt = 4\n}",
            "struct Outer {\n  struct Inter {\n    fileprivate struct Inner {}\n  }\n}",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    /* Near misses: an extension, members of a function or closure, a `#if` inside a type, and the word in a comment or string. */
    @Test func nearMissesAreNot() {
        let source = """
            fileprivate extension String {}
            private func helper() {}
            enum Namespace {
                #if DEBUG
                fileprivate static func debugOnly() {}
                #endif
            }
            let closure = { () -> Int in
                struct Local { fileprivate var inner = 0 }
                return 1
            }
            // fileprivate func commented() {}
            let text = "fileprivate var quoted = 0"
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func fixSwapsTheKeywordOnly() {
        let source = """
            fileprivate func helper() {}
            @MainActor fileprivate final class Holder {
                fileprivate var inner = 0
            }
            fileprivate(set) var counter = 0
            """
        let expected = """
            private func helper() {}
            @MainActor private final class Holder {
                fileprivate var inner = 0
            }
            private(set) var counter = 0
            """
        #expect(Self.fixed(source) == expected)
    }
}
