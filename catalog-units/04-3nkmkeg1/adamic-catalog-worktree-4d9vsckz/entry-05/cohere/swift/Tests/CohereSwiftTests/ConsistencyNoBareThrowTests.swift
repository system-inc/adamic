import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for `cohere-swift/consistency-no-bare-throw`, both directions. The reporting fixtures are the real sites on
 the proving grounds, near verbatim: Presence's `DanceImport` (a guard's throw, and a `??` fallback after a
 framework's optional error), its `EffectBench` tool (a throw inside a `do`), and macOS's `scroll-probe` (an
 `NSError` spread over three lines). Each asserts the line and column of the `NSError` call, so a finding on the
 `throw` keyword or the wrong branch fails rather than passing on its count.

 The silent fixtures are the near misses the same files hold: an SDK domain with a real code (macOS's
 `ChildProcessReaperTests`, and SwiftTerm's `NSOSStatusErrorDomain` document errors), a typed `CocoaError`
 carrying a description (Presence's `RigNarration`), a helper that builds the error and is thrown by its call
 (Presence's `MToonLockerSky`), and the `NSError`s an Objective-C contract asks for without a throw.
 */
struct ConsistencyNoBareThrowTests {
    static func file(_ source: String, targetKind: String = "library") -> ParsedFile {
        let url = URL(fileURLWithPath: "/fixture/Subject.swift")
        return ParsedFile(
            url: url,
            targetName: "Fixture",
            targetKind: targetKind,
            source: source,
            tree: Parser.parse(source: source),
            nodeCount: 0,
        )
    }

    /* The findings for a file the rule agrees to read, and none for one it declines, as the pipeline runs it. */
    static func findings(_ source: String, targetKind: String = "library") -> [FindingRecord] {
        let file = Self.file(source, targetKind: targetKind)
        guard ConsistencyNoBareThrow().applies(to: file) else { return [] }
        return ConsistencyNoBareThrow().findings(in: file)
    }

    static func positions(_ findings: [FindingRecord]) -> [String] {
        findings.map { "\($0.line):\($0.column)" }
    }

    /* Presence's `DanceImport.swift:92`: a guard with no video track throws an error made up on the spot. */
    @Test func aGuardsAdHocErrorIsFound() {
        let source = """
            let asset = AVURLAsset(url: video)
            guard let track = try await asset.loadTracks(withMediaType: .video).first else {
                throw NSError(domain: "DanceImport", code: 1, userInfo: [NSLocalizedDescriptionKey: "no video track"])
            }
            """
        let found = Self.findings(source)
        #expect(Self.positions(found) == ["3:11"])
        #expect(found.map(\.endColumn) == [107])
        #expect(found.map(\.messageId) == ["bareThrow"])
    }

    /* The message names the domain, the one part that varies, and the repair; checked against a literal rather than the rule's own text. */
    @Test func theMessageNamesTheDomainAndTheRepair() {
        let found = Self.findings(#"func f() throws { throw NSError(domain: "bench", code: 1) }"#)
        #expect(
            found.map(\.message) == [
                "This throws an NSError made up here, with the domain \"bench\", which names no declared failure. Declare the failure as a case of an error type of our own (an enum conforming to Error, and LocalizedError for its text) and throw that case, so a caller catches it by case. Until then a caller can match it only by repeating the domain string and the code, and nothing keeps the two in step."
            ]
        )
        #expect(found.map(\.rule) == ["cohere-swift/consistency-no-bare-throw"])
    }

    /*
     The finding a user sees is the catalog's entry (policy/messages/consistency-no-bare-throw.json), id and text, so an
     edited entry reaches the finding once regenerated. The literal above pins the words; this pins where they come from.
     */
    @Test func theFindingIsTheCatalogsMessage() {
        let found = Self.findings(#"func f() throws { throw NSError(domain: "bench", code: 1) }"#)
        let message = RuleMessages.ConsistencyNoBareThrow.bareThrow(domain: #""bench""#)
        #expect(found.map(\.messageId) == [message.id])
        #expect(found.map(\.message) == [message.text])
    }

    /* Presence's `DanceImport.swift:149`: the fallback after a framework's optional error is the same ad-hoc error. */
    @Test func aFallbackAfterAnOptionalErrorIsFound() {
        let source = """
            if reader.status == .failed { throw reader.error ?? NSError(domain: "DanceImport", code: 2) }
            return index
            """
        #expect(Self.positions(Self.findings(source)) == ["1:53"])
    }

    /* A chain of fallbacks folds to the right, and the made-up error is the last one. */
    @Test func theLastOfSeveralFallbacksIsFound() {
        let source = """
            throw first ?? second ?? NSError(domain: "Chain", code: 3)
            """
        #expect(Self.positions(Self.findings(source)) == ["1:26"])
    }

    /* Presence's `EffectBench/main.swift:599` and macOS's `scroll-probe/main.swift:259`: a tool target is read like any other, and a call over three lines is one finding at its start. */
    @Test func aToolsAdHocErrorsAreFound() {
        let source = """
            do {
                let library = try device.makeLibrary(source: source, options: nil)
                guard let function = library.makeFunction(name: "effect") else {
                    throw NSError(domain: "bench", code: 1, userInfo: [NSLocalizedDescriptionKey: "no kernel named effect"])
                }
                pipeline = try device.makeComputePipelineState(function: function)
            }
            guard let tap = CGEvent.tapCreate(tap: .cgSessionEventTap, place: .headInsertEventTap, options: .listenOnly, eventsOfInterest: mask, callback: callback, userInfo: userInfo) else {
                throw NSError(domain: "ScrollProbe", code: 1,
                              userInfo: [NSLocalizedDescriptionKey:
                                "Failed to create event tap. Ensure Accessibility permission is granted."])
            }
            """
        let found = Self.findings(source, targetKind: "executable")
        #expect(Self.positions(found) == ["4:15", "9:11"])
        #expect(found.map(\.endLine) == [4, 11])
    }

    /* The other spellings of the constructor, a parenthesized throw, both branches of a ternary, and an interpolated domain. */
    @Test func everySpellingOfTheConstructorIsFound() {
        let source = """
            throw Foundation.NSError(domain: "Import", code: 1)
            throw NSError.init(domain: "Import", code: 2)
            throw Foundation.NSError.init(domain: "Import", code: 3)
            throw (NSError(domain: "Import", code: 4))
            throw missing ? NSError(domain: "Import", code: 5) : NSError(domain: "Import", code: 6)
            throw NSError(domain: "\\(Self.self)", code: 7)
            """
        #expect(Self.positions(Self.findings(source)) == ["1:7", "2:7", "3:7", "4:8", "5:17", "5:54", "6:7"])
    }

    /* A throw nested in a closure and one in a `Task` each report their own. */
    @Test func throwsInsideClosuresAreFound() {
        let source = """
            let load: () throws -> Data = {
                throw NSError(domain: "Loader", code: 1)
            }
            Task {
                throw NSError(domain: "Loader", code: 2)
            }
            """
        #expect(Self.positions(Self.findings(source)) == ["2:11", "5:11"])
    }

    /*
     macOS's `ChildProcessReaperTests.swift:104` and SwiftTerm's `Document.swift:33` name a domain the SDK declares,
     with a code from its catalog, and the literal spellings of SDK domains are the same thing.
     */
    @Test func platformDomainsAreNot() {
        let source = """
            guard result == 0 else {
                throw NSError(domain: NSPOSIXErrorDomain, code: Int(result))
            }
            throw NSError(domain: NSOSStatusErrorDomain, code: controlErr, userInfo: nil)
            throw NSError(domain: "NSCocoaErrorDomain", code: NSFileReadCorruptFileError)
            throw NSError(domain: "AVFoundationErrorDomain", code: -11800)
            throw NSError(domain: "com.apple.LocalAuthentication", code: -2)
            throw NSError(domain: "kCFErrorDomainCFNetwork", code: 2)
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Presence's `RigNarration.swift:264` and its neighbours: a typed error has a case to catch, whatever text it carries. */
    @Test func typedErrorsAreNot() {
        let source = """
            guard let start = record.studioStartTime else {
                throw CocoaError(.fileReadCorruptFile, userInfo: [NSLocalizedDescriptionKey: "narration has no start time"])
            }
            throw URLError(.badServerResponse)
            throw POSIXError(POSIXErrorCode(rawValue: errno) ?? .EIO)
            throw CancellationError()
            throw VRMError(kind: kind, message: "spring bone \\(name) must be finite and nonnegative")
            throw Errors.NSError(domain: "Ours", code: 1)
            let note = NSError.self
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* Presence's `MToonLockerSky.swift:48`: the error is built by a helper and thrown by its call, which the original does not read either. */
    @Test func errorsBuiltElsewhereAndThrownByNameAreNot() {
        let source = """
            func fail(_ reason: String) -> any Error {
                NSError(
                    domain: "MToonLockerSky", code: 1,
                    userInfo: [NSLocalizedDescriptionKey: "the locker sky \\(reason)"])
            }
            guard width >= Self.requiredWidth else { throw fail("needs rows of \\(Self.requiredWidth), not \\(width)") }
            let error = NSError(domain: "Import", code: 1)
            throw error
            throw reader.error ?? fallback
            throw NSError(domain: Self.errorDomain, code: 1)
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* The `NSError`s an Objective-C contract asks for are not thrown: a completion handler, an out-parameter, a continuation. */
    @Test func errorsHandedToObjectiveCAreNot() {
        let source = """
            completionHandler(nil, NSError(domain: "Export", code: 1))
            error?.pointee = NSError(domain: "Export", code: 2)
            delegate.session(self, didFailWithError: NSError(domain: "Export", code: 3))
            continuation.resume(throwing: NSError(domain: "Export", code: 4))
            return .failure(NSError(domain: "Export", code: 5))
            """
        #expect(Self.findings(source).isEmpty)
    }

    /* A test target's thrown message is the diagnosis: the same shape that reports in a library is not read there. */
    @Test func testTargetsAreNotRead() {
        let source = """
            guard result == 0 else {
                throw NSError(domain: "ChildProcessReaperTests", code: Int(result))
            }
            """
        #expect(!ConsistencyNoBareThrow().applies(to: Self.file(source, targetKind: "test")))
        #expect(Self.findings(source, targetKind: "test").isEmpty)
        #expect(Self.positions(Self.findings(source, targetKind: "library")) == ["2:11"])
    }

    /* The prefilter is never narrower than the rule: a file holding the shape is read, and one without either word is not. */
    @Test func thePrefilterReadsEveryFileThatCanHoldTheShape() {
        #expect(ConsistencyNoBareThrow().applies(to: Self.file(#"throw NSError(domain: "X", code: 1)"#)))
        #expect(!ConsistencyNoBareThrow().applies(to: Self.file(#"let error = NSError(domain: "X", code: 1)"#)))
        #expect(!ConsistencyNoBareThrow().applies(to: Self.file("throw CocoaError(.fileReadUnknown)")))
    }

    /* Half-typed source still parses into a tree the rule walks, and walks without a finding it cannot place. */
    @Test func malformedThrowsAreSurvived() {
        for source in [
            "throw", "throw NSError(", "throw NSError(domain: ", "throw ?? NSError(domain: \"X\", code: 1)", "throw (",
            "throw NSError",
        ] {
            _ = Self.findings(source)
        }
    }
}
