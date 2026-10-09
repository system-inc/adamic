import Foundation
import SwiftParser
import Testing

@testable import CohereSwift

/*
 Fixture pairs for duplicate-imports, both directions, built from SwiftLint's documented `duplicate_imports`
 cases. Where SwiftLint flags a pair this rule leaves alone (a submodule beside its module, a scoped import
 beside the whole module, imports that differ by attribute or modifier, a top-level import after a guarded
 one), the case sits with the passing fixtures and says why in a comment.
 */
struct DuplicateImportsTests {
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
        return DuplicateImports().findings(in: file)
    }

    static func positions(_ source: String) -> [String] {
        findings(source).map { "\($0.line):\($0.column)" }
    }

    /* The source with every fix applied, last edit first so earlier offsets stay valid. */
    static func fixed(_ source: String) -> String {
        var bytes = Array(source.utf8)
        let edits = findings(source).flatMap(\.fixes).sorted { $0.start > $1.start }
        for edit in edits {
            bytes.replaceSubrange(edit.start..<edit.end, with: Array(edit.text.utf8))
        }
        return String(decoding: bytes, as: UTF8.self)
    }

    @Test func exactRepeatsAreFound() {
        #expect(Self.positions("import Foundation\nimport Dispatch\nimport Foundation\n") == ["3:1"])
        #expect(Self.positions("import Foundation\nimport Foundation\nimport Foundation\n") == ["2:1", "3:1"])
        #expect(Self.positions("import CoreImage.CIFilterBuiltins\nimport CoreImage.CIFilterBuiltins\n") == ["2:1"])
        #expect(Self.positions("import A.B.C\nimport A.B.C\n") == ["2:1"])
        #expect(Self.positions("import A; import A\n") == ["1:11"])
    }

    @Test func attributedRepeatsAreFound() {
        #expect(Self.positions("@_implementationOnly import A\n@_implementationOnly import A\n") == ["2:1"])
        #expect(Self.positions("@testable import A\n@testable import A\n") == ["2:1"])
        #expect(Self.positions("@testable @preconcurrency import A\n@preconcurrency @testable import A\n") == ["2:1"])
        #expect(Self.positions("@_spi(Internal) import A\n@_spi(Internal) import A\n") == ["2:1"])
        #expect(Self.positions("public import A\npublic import A\n") == ["2:1"])
        #expect(Self.positions("import struct A.Foo\nimport struct A.Foo\n") == ["2:1"])
    }

    /* Spacing, comments, and backticks do not make two imports different. */
    @Test func spellingNoiseDoesNotHideARepeat() {
        #expect(Self.positions("import A // first\nimport  A // second\n") == ["2:1"])
        #expect(Self.positions("import A\nimport `A`\n") == ["2:1"])
    }

    /* SwiftLint's own case: a top-level import repeated after an `#if` block. */
    @Test func repeatsAroundAnIfBlockAreFound() {
        let source = """
            import A
            #if DEBUG
                @testable import KsApi
            #else
                import KsApi
            #endif
            import A
            """
        #expect(Self.positions(source) == ["7:1"])
    }

    @Test func repeatsInsideOneBranchAreFound() {
        #expect(Self.positions("#if DEBUG\nimport A\nimport A\n#endif\n") == ["3:1"])
    }

    /* A guarded import under a top-level one adds nothing whenever it compiles, wherever it sits in the file. SwiftLint misses both. */
    @Test func guardedRepeatsOfATopLevelImportAreFound() {
        #expect(Self.positions("import A\n#if DEBUG\n    import A\n#endif\n") == ["3:5"])
        #expect(Self.positions("#if os(macOS)\n#if DEBUG\nimport A\n#endif\nimport A\n#endif\n") == ["3:1"])
    }

    @Test func distinctImportsAreNot() {
        #expect(Self.findings("import A\nimport B\nimport C\n").isEmpty)
        #expect(Self.findings("import A.B\nimport A.C\n").isEmpty)
        #expect(Self.findings("@_implementationOnly import A\n@_implementationOnly import B\n").isEmpty)
        #expect(Self.findings("@testable import A\n@testable import B\n").isEmpty)
        #expect(Self.findings("import A // module\nimport B // module\n").isEmpty)
        #expect(Self.findings("#if TEST\nfunc test() {\n}\n").isEmpty)
        #expect(Self.findings("import Foo\n@testable import struct Foo.Bar\n").isEmpty)
        #expect(Self.findings("import CoreImage\nimport CoreImage.CIFilterBuiltins\n").isEmpty)
    }

    /* Only one branch compiles, so the same import in two branches is said once per build. */
    @Test func separateBranchesAreNot() {
        let source = """
            #if DEBUG
                @testable import KsApi
            #else
                import KsApi
            #endif
            #if canImport(UIKit)
            import UIKit
            #elseif canImport(AppKit)
            import AppKit
            #else
            import UIKit
            #endif
            """
        #expect(Self.findings(source).isEmpty)
    }

    /*
     Deleting a top-level import because a guarded one exists would break every build where the guard is off,
     so the top-level line 4 is never reported, though it comes second. SwiftLint reports line 4. The guarded
     line 2 is the one that adds nothing, and is reported instead.
     */
    @Test func topLevelImportIsKeptOverAGuardedOne() {
        #expect(Self.positions("#if DEBUG\nimport A\n#endif\nimport A\n") == ["2:1"])
        #expect(Self.fixed("#if DEBUG\nimport A\n#endif\nimport A\n") == "#if DEBUG\n#endif\nimport A\n")
    }

    /*
     SwiftLint flags every one of these. A submodule may or may not come with its module (the module map
     decides, which syntax cannot read), and a scoped import also decides which name wins when two modules
     declare it, so none can be deleted on syntax alone.
     */
    @Test func submodulesAndScopedImportsAreNot() {
        let sources = [
            "import A\nimport A.B.C\nimport A.B\n",
            "import A.B\nimport A.B.C\n",
            "import A.B.C\nimport A.B.C.D\nimport A.B.C.E\n",
            "import Foundation\nimport Foundation.NSString\n",
            "import Foundation.NSString\nimport Foundation\n",
            "import A.B.C\nimport A.B\nimport A\n",
            "import A\nimport class A.Foo\n",
            "import A\nimport enum A.B.Foo\n",
            "import A\nimport func A.Foo\n",
            "import A\nimport let A.Foo\n",
            "import A\nimport protocol A.Foo\n",
            "import A\nimport struct A.Foo\n",
            "import A\nimport typealias A.Foo\n",
            "import A\nimport var A.Foo\n",
            "import A.B\nimport struct A.B.Foo\n",
            "@testable import Foo\nimport struct Foo.Bar\n",
            "import struct A.Foo\nimport class A.Foo\n",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    /* Each of these means something a plain import does not, so neither line is a repeat of the other. SwiftLint flags all of them. */
    @Test func attributesAndModifiersMakeADifferentImport() {
        let sources = [
            "import A\n@testable import A\n",
            "@testable import A\nimport A\n",
            "@_exported import A\nimport A\n",
            "import A\n@preconcurrency import A\n",
            "@_spi(First) import A\n@_spi(Second) import A\n",
            "public import A\nimport A\n",
            "@_implementationOnly import A\nimport A\n",
        ]
        for source in sources {
            #expect(Self.findings(source).isEmpty, "\(source)")
        }
    }

    @Test func fixDeletesTheRepeatedLine() {
        #expect(
            Self.fixed("import Foundation\nimport Dispatch\nimport Foundation\n\nlet value = 1\n")
                == "import Foundation\nimport Dispatch\n\nlet value = 1\n"
        )
        #expect(Self.fixed("import A\nimport A\nimport A\n") == "import A\n")
        #expect(Self.fixed("import A\n#if DEBUG\n    import A\n#endif\n") == "import A\n#if DEBUG\n#endif\n")
        #expect(Self.fixed("import A\r\nimport A\r\nimport B\r\n") == "import A\r\nimport B\r\n")
    }

    @Test func fixAtTheEndOfAFileLeavesNoBlankLine() {
        #expect(Self.fixed("import A\nimport A") == "import A")
    }

    /* A comment on, in, or just above the repeat, or code sharing its line, would go with it or be stranded, so the finding stands without a fix. */
    @Test func noFixWhenTheLineHoldsMoreThanTheImport() {
        let sources = [
            "import A\nimport A // second\n",
            "import A\n// Needed for the bridge.\nimport A\n",
            "import A\nimport /* again */ A\n",
            "import A; import A\n",
            "import A\nimport A; let value = 1\n",
        ]
        for source in sources {
            let findings = Self.findings(source)
            #expect(findings.count == 1, "\(source)")
            #expect(findings.flatMap(\.fixes).isEmpty, "\(source)")
        }
    }

    @Test func messageNamesTheEarlierLine() {
        let findings = Self.findings("import A\nimport B\nimport A\n")
        let namesLine = findings.first?.message.range(of: "on line 1,") != nil
        #expect(namesLine)
    }
}
