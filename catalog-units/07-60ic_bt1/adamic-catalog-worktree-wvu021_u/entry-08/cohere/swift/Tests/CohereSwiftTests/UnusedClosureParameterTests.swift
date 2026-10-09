import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `unused-closure-parameter`, both directions, covering every example SwiftLint's
 `unused_closure_parameter` documents. The failing cases assert the exact line and column of the parameter's
 name. The passing cases are the near misses: a name read only inside a nested closure or a string
 interpolation, a backticked spelling, a `$name` projection, a shorthand `if let`, a bare capture, and a
 local that shadows the parameter. The fix cases check that only the name token becomes `_`, that a type
 annotation stays, and that a `$name` parameter gets no fix.
 */
struct UnusedClosureParameterTests {
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
        return UnusedClosureParameter().findings(in: file)
    }

    static func positions(_ source: String) -> [String] {
        findings(source).map { "\($0.line):\($0.column)" }
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

    /* SwiftLint's triggering examples, each in its own fixture because several redeclare the same names. */
    @Test func incumbentTriggeringExamplesAreFound() {
        #expect(Self.positions("[1, 2].map { number in\n return 3 }") == ["1:14"])
        #expect(Self.positions("[1, 2].map { number in\n return numberWithSuffix }") == ["1:14"])
        #expect(Self.positions("[1, 2].map { number in\n return 3 // number }\n}") == ["1:14"])
        #expect(Self.positions("[1, 2].map { number in\n return 3 \"number\" }") == ["1:14"])
        #expect(Self.positions("[1, 2].something { number, idx in\n return number }") == ["1:28"])
        #expect(Self.positions("genericsFunc { (number: TypeA, idx: TypeB) in return idx }") == ["1:17"])
        #expect(Self.positions("let c: (Int) -> Void = { foo in _ = .foo }") == ["1:26"])
        #expect(Self.positions("hoge(arg: num) { num in\n}") == ["1:18"])
        #expect(Self.positions("fooFunc { 아 in }") == ["1:11"])
        #expect(Self.positions("func foo () {\n bar { number in return 3 }\n}") == ["2:8"])
        #expect(
            Self.positions(
                """
                viewModel?.profileImage.didSet(weak: self) { (self, profileImage) in
                    profileImageView.image = profileImage
                }
                """
            ) == ["1:47"]
        )
        #expect(
            Self.positions("let failure: Failure = { task, error in\n    observer.sendFailed(error)\n}") == ["1:26"]
        )
        #expect(Self.positions("List($names) { $name in\n    Text(\"Foo\")\n}") == ["1:16"])
        #expect(
            Self.positions("let class1 = \"a\"\n_ = [\"a\"].filter { `class` in `class1`.hasPrefix(\"a\") }") == [
                "2:20"
            ]
        )
    }

    /* SwiftLint's non-triggering examples. */
    @Test func incumbentNonTriggeringExamplesAreNot() {
        let sources = [
            "[1, 2].map { $0 + 1 }",
            "[1, 2].map({ $0 + 1 })",
            "[1, 2].map { number in\n number + 1 \n}",
            "[1, 2].map { _ in\n 3 \n}",
            "[1, 2].something { number, idx in\n return number * idx\n}",
            "let isEmpty = [1, 2].isEmpty()",
            "violations.sorted(by: { lhs, rhs in \n return lhs.location > rhs.location\n})",
            """
            rlmConfiguration.migrationBlock.map { rlmMigration in
                return { migration, schemaVersion in
                    rlmMigration(migration.rlmMigration, schemaVersion)
                }
            }
            """,
            "genericsFunc { (a: Type, b) in\n    a + b\n}",
            "var label: UILabel = { (lbl: UILabel) -> UILabel in\n    lbl.backgroundColor = .red\n    return lbl\n}(UILabel())",
            "hoge(arg: num) { num in\n    return num\n}",
            "({ (manager: FileManager) in\n  print(manager)\n})(FileManager.default)",
            "withPostSideEffect { input in\n    if true { print(\"\\(input)\") }\n}",
            "viewModel?.profileImage.didSet(weak: self) { (self, profileImage) in\n    self.profileImageView.image = profileImage\n}",
            "let failure: Failure = { task, error in\n    observer.sendFailed(error, task)\n}",
            "List($names) { $name in\n    Text(name)\n}",
            "List($names) { $name in\n    TextField($name)\n}",
            "_ = [\"a\"].filter { `class` in `class`.hasPrefix(\"a\") }",
            "let closure: (Int) -> Void = { `foo` in _ = foo }",
            "let closure: (Int) -> Void = { foo in _ = `foo` }",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    /*
     Findings SwiftLint misses, each a parameter the body provably never reads: a nested closure that rebinds
     the name for its whole body, `$0` in a nested closure, and a key path component that names a property.
     */
    @Test func shadowedByNestedClosuresAndKeyPathsAreFound() {
        #expect(Self.positions("run { value in\n    other.map { value in value * 2 }\n}") == ["1:7"])
        #expect(Self.positions("run { value in\n    later { [value = 3] in print(value) }\n}") == ["1:7"])
        #expect(Self.positions("run { value in\n    items.map { $0 + 1 }\n}") == ["1:7"])
        #expect(Self.positions("run { name in\n    items.map(\\.name)\n}") == ["1:7"])
        #expect(Self.positions("run { first, second in 0 }") == ["1:7", "1:14"])
        #expect(Self.positions("run { outer in\n    later { inner in 1 }\n}") == ["1:7", "2:13"])
    }

    /* Reads SwiftLint also counts, in places the name hides: nested closures, interpolation, member bases, inout writes. */
    @Test func readsInNestedPlacesAreNot() {
        let sources = [
            "run { value in\n    later { print(value) }\n}",
            "run { value in\n    later { other in print(value, other) }\n}",
            "run { value in\n    print(\"\\(value.count)\")\n}",
            "run { (value: inout Int) in\n    value = 3\n}",
            "run { value in\n    switch other {\n    case value: break\n    default: break\n    }\n}",
            "run { value in\n    _ = items[keyPath: \\.[value]]\n}",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    /* Reads that spell no reference, where a rename to `_` would not compile: shorthand binding, bare captures, backing storage. */
    @Test func readsWithoutAReferenceAreNot() {
        let sources = [
            "run { value in\n    if let value {}\n}",
            "run { value in\n    guard let value else { return }\n}",
            "run { (self) in\n    guard let self else { return }\n}",
            "run { value in\n    later { [value] in }\n}",
            "run { value in\n    later { [weak value] in }\n}",
            "run { value in\n    later { [value] in print(value) }\n}",
            "run { value in\n    later { [copy = value] in print(copy) }\n}",
            "run { value in\n    later { value in print(value) }(value)\n}",
            "take { $item in\n    print(_item)\n}",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    /*
     Shadowing by a local declaration is left alone, on purpose: every reference to the name counts as a read,
     so the first fixture is a known miss rather than a scope resolution the syntax cannot back. The rest read
     the parameter for real: `var value = value` reads it, and a nested `$value` closure does not hide a `$value` parameter's spellings.
     */
    @Test func shadowingByLocalsIsNotFlagged() {
        let sources = [
            "run { value in\n    let value = 3\n    print(value)\n}",
            "run { value in\n    var value = value\n    value += 1\n}",
            "run { value in\n    func helper(value: Int) -> Int { value }\n    print(helper(value: 1))\n}",
            "List($names) { $name in\n    later { name in print($name) }\n}",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    /* The name token alone becomes `_`: a type annotation stays, a backticked name goes whole, a sibling is untouched. */
    @Test func fixRenamesTheNameOnly() {
        #expect(
            Self.fixed("genericsFunc { (number: TypeA, idx: TypeB) in return idx }")
                == "genericsFunc { (_: TypeA, idx: TypeB) in return idx }"
        )
        #expect(
            Self.fixed("[1, 2].something { number, idx in\n return number }")
                == "[1, 2].something { number, _ in\n return number }"
        )
        #expect(
            Self.fixed("let class1 = \"a\"\n_ = [\"a\"].filter { `class` in `class1`.hasPrefix(\"a\") }")
                == "let class1 = \"a\"\n_ = [\"a\"].filter { _ in `class1`.hasPrefix(\"a\") }"
        )
        #expect(Self.fixed("run { first, second in 0 }") == "run { _, _ in 0 }")
        #expect(Self.fixed("run { (value: inout Int) in }") == "run { (_: inout Int) in }")
        #expect(
            Self.fixed("viewModel?.didSet(weak: self) { (self, image) in\n    imageView.image = image\n}")
                == "viewModel?.didSet(weak: self) { (_, image) in\n    imageView.image = image\n}"
        )
    }

    /* A `$name` parameter is reported without a fix, because `_` drops the property wrapper the `$` asked for. */
    @Test func projectedParameterHasNoFix() {
        let found = Self.findings("List($names) { $name in\n    Text(\"Foo\")\n}")
        #expect(found.count == 1)
        #expect(found.first?.fixes.isEmpty == true)
    }
}
