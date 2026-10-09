import SwiftParser
import Testing

@testable import CohereSwift

/*
 The fingerprint both ways, edge by edge: an edit another file cannot see must keep it, and an edit another
 file can see must change it. The second half is the one that matters, because a fingerprint that never
 changed would pass every "keeps" test and let a stale compiler record stand.
 */
struct InterfaceFingerprintTests {
    static func fingerprint(_ source: String) -> String {
        InterfaceFingerprint.of(Parser.parse(source: source))
    }

    static func same(_ before: String, _ after: String) -> Bool {
        fingerprint(before) == fingerprint(after)
    }

    @Test func aFunctionBodyIsInvisible() {
        #expect(
            Self.same(
                "func count() -> Int {\n    1\n}\n",
                "func count() -> Int {\n    let two = 2\n    return two\n}\n",
            )
        )
    }

    @Test func commentsAndLayoutAreInvisible() {
        #expect(
            Self.same("struct Shape { var sides: Int }\n", "/* A shape. */\nstruct Shape {\n    var sides: Int\n}\n")
        )
    }

    @Test func initializerDeinitializerAndObserverBodiesAreInvisible() {
        let before =
            "final class Box {\n    var size = 0 { didSet { print(1) } }\n    init() { size = 1 }\n    deinit { print(2) }\n}\n"
        let after =
            "final class Box {\n    var size = 0 { didSet { print(3) } }\n    init() { size = 4 }\n    deinit { print(5) }\n}\n"
        #expect(Self.same(before, after))
    }

    @Test func anAnnotatedComputedPropertysGetterIsInvisible() {
        #expect(Self.same("var total: Int { 1 }\n", "var total: Int { 2 + 3 }\n"))
        #expect(Self.same("var total: Int { get { 1 } }\n", "var total: Int { get { 2 } }\n"))
        #expect(Self.same("subscript(index: Int) -> Int { index }\n", "subscript(index: Int) -> Int { index + 1 }\n"))
    }

    @Test func aSignatureIsVisible() {
        #expect(!Self.same("func count() -> Int { 1 }\n", "func count() -> Double { 1 }\n"))
        #expect(!Self.same("func count(of items: [Int]) -> Int { 1 }\n", "func count(in items: [Int]) -> Int { 1 }\n"))
    }

    /* An unannotated property's type is inferred from its initializer, so the initializer is interface. */
    @Test func aStoredPropertysInitializerIsVisible() {
        #expect(!Self.same("let limit = 1\n", "let limit = 1.0\n"))
        #expect(!Self.same("struct Box { var handler = { 1 } }\n", "struct Box { var handler = { \"one\" } }\n"))
    }

    @Test func aDefaultArgumentIsVisible() {
        #expect(!Self.same("func wait(seconds: Int = 1) {}\n", "func wait(seconds: Int = 2) {}\n"))
    }

    @Test func topLevelCodeIsVisible() {
        #expect(!Self.same("print(1)\n", "print(2)\n"))
    }

    @Test func gainingABodyIsVisible() {
        #expect(
            !Self.same(
                "protocol Shape {\n    func area() -> Double\n}\n",
                "protocol Shape {\n    func area() -> Double {}\n}\n",
            )
        )
    }

    @Test func anAttributeIsVisible() {
        #expect(!Self.same("func run() {}\n", "@MainActor func run() {}\n"))
    }
}
