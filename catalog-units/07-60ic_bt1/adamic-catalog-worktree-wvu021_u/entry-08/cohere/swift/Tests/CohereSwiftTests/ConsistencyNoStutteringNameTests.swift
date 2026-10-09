import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `consistency-no-stuttering-name`, both directions. The firing cases carry over every case the Go rule's
 test fires on (`result`, `outcome`, `data`, `value`, `response`, optional chaining, inside a comparison) and
 add one for each Swift binding form the lookup resolves. The silent cases carry over the Go rule's near
 misses and add Swift's own: implicit `self`, re-binding shorthands, overrides, and bindings that are not in
 scope where the stutter is read.
 */
struct ConsistencyNoStutteringNameTests {
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
        return ConsistencyNoStutteringName().findings(in: file)
    }

    static func positions(_ source: String) -> [String] {
        findings(source).map { "\($0.line):\($0.column)" }
    }

    /* The Go test's firing cases, each a local so the base is a name the code chose. */
    @Test func theGenericWordsStutter() {
        let source = """
            func run() {
                let result = load()
                let outcome = load()
                let data = load()
                let value = load()
                let response = load()
                use(result.result, outcome.outcome, data.data, value.value, response.response)
            }
            """
        #expect(Self.positions(source) == ["7:16", "7:32", "7:46", "7:58", "7:74"])
    }

    @Test func optionalChainingAndForceUnwrappingStutterToo() {
        let source = """
            func run(outcome: Outcome?) {
                let readable = outcome?.outcome == .unreadable
                let forced = outcome!.outcome
            }
            """
        #expect(Self.positions(source) == ["2:29", "3:27"])
    }

    @Test func theMessageNamesTheWord() {
        let found = Self.findings("func run(outcome: Outcome) { _ = outcome.outcome }")
        #expect(found.count == 1)
        #expect(found.first?.messageId == "stutteringName")
        #expect(
            found.first?.message
                == "\"outcome.outcome\" stutters, which means the name is carrying nothing: it repeats the field instead of saying which outcome this is. Rename the value for what it holds, the type it came back as or whatever distinguishes it from another outcome in this scope, so a reader forty lines down does not have to find the declaration."
        )
    }

    /* Every binding form that chooses a name, each one the base of a stutter. */
    @Test func everyBindingFormIsResolved() {
        let source = """
            func run(for item: Item, state: State, _ thing: Thing) {
                _ = item.item
                _ = state.state
                _ = thing.thing
                entries.map { value in value.value }
                entries.map { (key, value) in value.value }
                entries.map { [output = source] in output.output }
                if let data = load() { _ = data.data }
                guard let response = load() else { return }
                _ = response.response
                while let result = next() { _ = result.result }
                if case .some(let outcome) = load() { _ = outcome.outcome }
                if case let .some(value) = load() { _ = value.value }
                for item in items where item.item.isReady { _ = item.item }
                switch load() {
                case .success(let result): _ = result.result
                default: break
                }
                do { try work() } catch let value { _ = value.value }
                func nested() { _ = state.state }
            }
            """
        #expect(
            Self.positions(source) == [
                "2:14", "3:15", "4:15", "5:34", "6:41", "7:47", "8:37", "10:18", "11:44", "12:55", "13:51",
                "14:34", "14:58", "16:43", "19:51", "20:31",
            ]
        )
    }

    /* A top-level variable is chosen here, and seen by a free function written before or after it. */
    @Test func aTopLevelVariableIsResolved() {
        let source = """
            func early() { _ = value.value }
            let value = Wrapper()
            """
        #expect(Self.positions(source) == ["1:26"])
    }

    @Test func aNonOverridingSingleNameParameterIsJudged() {
        let source = """
            struct Handler: Receiving {
                func receive(result: Outcome) { _ = result.result }
                init(value: Wrapper) { _ = value.value }
            }
            """
        #expect(Self.positions(source) == ["2:48", "3:38"])
    }

    /* The Go test's near misses: generic words doing their job, a real word that repeats, a deep access. */
    @Test func theGoNearMissesStaySilent() {
        let source = """
            func run(response: Response, result: [Int], user: User, session: Session, outcome: Outcome) {
                _ = response.json()
                _ = result.count
                _ = user.user
                _ = result.value
                _ = session.session
                _ = result["result"]
                _ = outcome.status.code
                _ = 1
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /*
     A bare name inside a type that is not a local may be a property, whose name may be a protocol requirement
     or an override declared anywhere, so it is not judged. A global seen from inside a type is not judged either,
     since a member of the same name would shadow it.
     */
    @Test func propertiesAndUnresolvedNamesStaySilent() {
        let source = """
            let value = Wrapper()
            final class Model: Store {
                var state: State
                override var result: Outcome { fatalError("") }
                func render() {
                    _ = state.state
                    _ = result.result
                    _ = value.value
                    _ = self.state.state
                    _ = data.data
                    entries.map { _ in state.state }
                }
            }
            func elsewhere() { _ = item.item }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* `if let value`, `guard let value` and `[value]` re-bind the outer name; the declaration that chose it decides. */
    @Test func rebindingShorthandsDeferToTheOuterName() {
        let source = """
            struct Model {
                var value: Wrapper?
                func render() {
                    if let value { _ = value.value }
                    guard let value else { return }
                    _ = value.value
                    entries.map { [value] in value.value }
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    @Test func anOverridesSingleNameParameterStaysSilent() {
        let source = """
            final class Child: Parent {
                override func receive(result: Outcome) { _ = result.result }
                override init(value: Wrapper) { _ = value.value }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* An override's second name is still ours, so a stutter on it is reported. */
    @Test func anOverridesSecondNameIsJudged() {
        let source = """
            final class Child: Parent {
                override func receive(_ result: Outcome) { _ = result.result }
            }
            """
        #expect(Self.positions(source) == ["2:59"])
    }

    /* A binding is only seen where Swift sees it: not in a guard's own else, an if's else, or after its block. */
    @Test func bindingsOutOfScopeStaySilent() {
        let source = """
            struct Model {
                var value: Wrapper
                func render() {
                    guard let value = load() else { _ = value.value; return }
                    if let result = load() { } else { _ = result.result }
                    do { let data = load() }
                    _ = data.data
                    entries.map { [state = state.state] in state }
                    func helper(item: Int = item.item) { }
                    _ = outcome.outcome
                    let outcome = load()
                }
            }
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Swift's other near misses: key paths, implicit members, type names, a different member, and words off the list. */
    @Test func swiftNearMissesStaySilent() {
        let source = """
            func run(value: Wrapper, camera: Camera, values: [Wrapper]) {
                _ = \\Wrapper.value.value
                _ = values.map(\\.value.value)
                _ = .value
                _ = Value.value
                _ = value.wrapped.value
                _ = camera.camera
                _ = values.values
                _ = value.Value
            }
            """
        #expect(Self.findings(source).isEmpty)
    }
}
