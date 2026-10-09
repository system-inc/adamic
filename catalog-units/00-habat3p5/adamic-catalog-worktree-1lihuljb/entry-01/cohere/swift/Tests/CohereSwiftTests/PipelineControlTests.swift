import CohereSwift
import Foundation
import Testing

/*
 The known-dirty controls: a real package, built by the real toolchain, run through the whole pipeline
 with one fault injected per phase. Each phase must catch its own fault at the line it was put on, and a
 type error must stop lint. The clean package comes first and must pass with no findings, or the dirty
 runs below would prove only that the pipeline always complains.

 These are slow on purpose (each run is a `swift build`), and they are the only tests that would notice a
 phase quietly wired to nothing: the per-rule and per-phase tests each check a piece, not that the pieces
 are connected. Every run is `--no-fix`, so the fixture is read and never written.
 */
@Suite(.serialized)
struct PipelineControlTests {
    static let manifest = """
        // swift-tools-version:6.2
        import PackageDescription

        let package = Package(
            name: "Control",
            targets: [
                .target(
                    name: "Control",
                    swiftSettings: [.enableUpcomingFeature("ExistentialAny"), .enableUpcomingFeature("MemberImportVisibility"), .strictMemorySafety()]
                )
            ]
        )

        """

    static let cleanSource = """
        struct Control {
            let value: Int

            func doubled() -> Int {
                value * 2
            }
        }

        """

    /* What one run said: the stream's records by kind, in order. */
    struct Run {
        var records: [[String: Any]]

        func of(_ kind: String) -> [[String: Any]] {
            records.filter { $0["kind"] as? String == kind }
        }

        func phase(_ name: String) -> String? {
            of("phase").first { $0["name"] as? String == name }?["outcome"] as? String
        }

        /* `rule:line` for every finding, so an assertion reads as what was caught and where. */
        var findings: [String] {
            of("finding").map { "\($0["rule"] as? String ?? ""):\($0["line"] as? Int ?? 0)" }
        }
    }

    /* A fresh package per run, because the engine's build cache is keyed by the package's path. */
    static func run(
        source: String,
        manifest: String = manifest,
        otherFiles: [String: Data] = [:],
        arguments: [String] = [],
        mutating: Bool = false,
        sink: ((Data) -> Void)? = nil,
    ) async throws -> Run {
        let root = FileManager.default.temporaryDirectory.appendingPathComponent(
            "cohere-swift-control-\(UUID().uuidString)",
            isDirectory: true,
        )
        let sources = root.appendingPathComponent("Sources/Control", isDirectory: true)
        try FileManager.default.createDirectory(at: sources, withIntermediateDirectories: true)
        try manifest.write(to: root.appendingPathComponent("Package.swift"), atomically: true, encoding: .utf8)
        try source.write(to: sources.appendingPathComponent("Control.swift"), atomically: true, encoding: .utf8)
        for (name, contents) in otherFiles {
            try contents.write(to: sources.appendingPathComponent(name))
        }
        defer {
            do {
                try FileManager.default.removeItem(at: root)
            }
            catch {
                /* Ignored on purpose: a leftover fixture in the temporary directory costs nothing, and failing a test that already answered would hide its answer. */
            }
        }

        let options = try CommandOptions.parse(
            ["--contract", "\(EngineVersion.contract)", "--root", root.path] + (mutating ? [] : ["--no-fix"])
                + arguments,
            workingDirectory: root,
        )
        var lines = Data()
        let writer = ContractWriter { data in
            lines.append(data)
            sink?(data)
        }
        _ = try await Pipeline(options: options, writer: writer, workingDirectory: root).run()
        let records = try lines.split(separator: UInt8(ascii: "\n")).map { line in
            try #require(try JSONSerialization.jsonObject(with: Data(line)) as? [String: Any])
        }
        if !mutating {
            #expect(
                String(decoding: try Data(contentsOf: sources.appendingPathComponent("Control.swift")), as: UTF8.self)
                    == source,
                "a --no-fix run wrote to the package",
            )
        }
        return Run(records: records)
    }

    @Test func theCleanPackagePasses() async throws {
        let run = try await Self.run(source: Self.cleanSource)
        #expect(run.findings.isEmpty)
        #expect(run.phase("types") == "ran")
        #expect(run.phase("lint") == "ran")
        #expect(run.of("summary").first?["exitCode"] as? Int == 0)
    }

    /*
     The front door reads a skip whose detail is exactly `not requested` (for unused, that phrase, the dash and how to
     ask) as no gap, and any other skip as one (f9ec90c, Contract.md). The words are written out here rather than read
     from the engine's constants, so rewording them fails this test before it breaks every narrowed run.
     */
    @Test func aPhaseNobodyAskedForIsSkippedInTheContractsWords() async throws {
        let run = try await Self.run(source: Self.cleanSource, arguments: ["--lint"])
        let details = Dictionary(
            run.of("phase").compactMap { phase in
                (phase["name"] as? String).map { ($0, "\(phase["outcome"] ?? "")|\(phase["detail"] ?? "")") }
            },
            uniquingKeysWith: { first, _ in first },
        )
        #expect(details["fix"] == "skipped|not requested")
        #expect(details["types"] == "skipped|not requested")
        #expect(details["unused"] == "skipped|not requested — this is a report, ask for it with --unused")
        #expect(run.phase("lint") == "ran")
        #expect(run.of("summary").first?["complete"] as? Bool == true)
    }

    @Test func formatCatchesAMisindentedLine() async throws {
        let source = Self.cleanSource.replacingOccurrences(of: "    let value: Int", with: "let value: Int")
        let run = try await Self.run(source: source)
        #expect(run.findings == ["cohere-swift/consistency-require-formatting:2"])
        #expect(run.phase("lint") == "ran", "a formatting finding does not stop the phases after it")
    }

    @Test func lintCatchesAForceUnwrap() async throws {
        let source = Self.cleanSource.replacingOccurrences(
            of: "        value * 2",
            with: "        Int(\"2\")! * value",
        )
        let run = try await Self.run(source: source)
        #expect(run.findings == ["cohere-swift/force-unwrapping:5"])
        #expect(run.phase("types") == "ran")
    }

    /*
     When the fix phase rewrote nothing, lint reuses its last walk instead of walking again, and must find exactly
     what a fresh walk finds. The --no-fix run has no fixer, so its lint walks; the two are compared.
     */
    @Test func lintReusesTheFixWalkAndFindsTheSame() async throws {
        let source = Self.cleanSource.replacingOccurrences(
            of: "        value * 2",
            with: "        Int(\"2\")! * value",
        )
        let reused = try await Self.run(source: source, mutating: true)
        let walked = try await Self.run(source: source)
        #expect(reused.phase("lint") == "reused")
        #expect(reused.of("lint").first?["reusedFrom"] as? String == "fix")
        #expect(walked.phase("lint") == "ran")
        #expect(reused.findings == walked.findings)
        #expect(reused.findings == ["cohere-swift/force-unwrapping:5"])
    }

    @Test func typesCatchesATypeErrorAndStopsLint() async throws {
        let source = Self.cleanSource.replacingOccurrences(
            of: "        value * 2",
            with: "        let text: Int = \"two\"\n        return value * text",
        )
        let run = try await Self.run(source: source)
        #expect(run.findings == [":5"], "the compiler's error, which carries no rule name")
        #expect(run.phase("lint") == "notReached")
        #expect(run.of("summary").first?["exitCode"] as? Int == 1)
    }

    /* The naming rule judges with the compiled-in words: no argument names a file, and a word from policy's file is still flagged. */
    @Test func theCompiledInVocabularyIsTheOneJudgedWith() async throws {
        let source = Self.cleanSource.replacingOccurrences(of: "    let value: Int", with: "    let val: Int")
        let run = try await Self.run(source: source.replacingOccurrences(of: "value * 2", with: "val * 2"))
        #expect(run.findings == ["cohere-swift/consistency-no-abbreviated-identifier:2"])
    }

    /*
     The fix record names every file it wrote, once each, with the fixes by rule and whether the formatter changed it, so
     the front door can list what a run cohered and the count equals the list. A fix, a reformat, and a run that writes
     nothing.
     */
    @Test func theFixRecordNamesEachRewrittenFile() async throws {
        let fixed = try await Self.run(
            source: Self.cleanSource + "\nfileprivate func helper() {}\n",
            mutating: true,
        )
        let fixRecord = try #require(fixed.of("fix").first)
        let changed = try #require(fixRecord["changedFiles"] as? [[String: Any]])
        #expect(changed.count == fixRecord["filesRewritten"] as? Int)
        #expect(changed.count == 1)
        #expect((changed.first?["file"] as? String)?.hasSuffix("Sources/Control/Control.swift") == true)
        #expect(changed.first?["fixedBy"] as? [String: Int] == ["cohere-swift/private-over-fileprivate": 1])
        #expect(changed.first?["formatted"] as? Bool == false)

        let reformatted = try await Self.run(
            source: Self.cleanSource.replacingOccurrences(of: "value * 2", with: "value  *  2"),
            mutating: true,
        )
        let reformattedRecord = try #require(reformatted.of("fix").first)
        let reformattedFiles = try #require(reformattedRecord["changedFiles"] as? [[String: Any]])
        #expect(reformattedFiles.count == reformattedRecord["filesRewritten"] as? Int)
        #expect(reformattedFiles.first?["fixedBy"] as? [String: Int] == [:])
        #expect(reformattedFiles.first?["formatted"] as? Bool == true)

        let untouched = try await Self.run(source: Self.cleanSource)
        let untouchedRecord = try #require(untouched.of("fix").first)
        #expect(
            untouchedRecord["changedFiles"] as? [[String: Any]] != nil,
            "an engine that writes the list writes it empty",
        )
        #expect((untouchedRecord["changedFiles"] as? [[String: Any]])?.isEmpty == true)
        #expect(untouchedRecord["filesRewritten"] as? Int == 0)
    }

    /* Contract 2: a file the engine cannot read is a record, between `project` and the first `phase`, and the run is incomplete. */
    @Test func anUnreadableFileIsNamedInARecord() async throws {
        let latin1 = Data("// caf".utf8) + Data([0xE9, 0x0A])
        let run = try await Self.run(source: Self.cleanSource, otherFiles: ["Latin1.swift": latin1])
        let kinds = run.records.compactMap { $0["kind"] as? String }
        let unreadable = try #require(run.of("unreadable").first)
        #expect((unreadable["file"] as? String)?.hasSuffix("/Sources/Control/Latin1.swift") == true)
        #expect((unreadable["error"] as? String)?.isEmpty == false)
        let at = try #require(kinds.firstIndex(of: "unreadable"))
        #expect(kinds.firstIndex(of: "project").map { $0 < at } == true)
        #expect(kinds.firstIndex(of: "phase").map { at < $0 } == true)
        #expect(run.of("summary").first?["complete"] as? Bool == false)
    }

    /*
     `--no-fix` forbids resolving dependencies, so a package whose Package.resolved is missing (ahraos-macos
     gitignores its own) fails before the compiler runs. The run must say that, not blame the build layout.
     */
    @Test func aBuildThatFailsBeforeCompilingSaysSo() async throws {
        let manifest = Self.manifest.replacingOccurrences(
            of: "name: \"Control\",\n    targets:",
            with:
                "name: \"Control\",\n    dependencies: [.package(url: \"https://example.invalid/Missing.git\", from: \"1.0.0\")],\n    targets:",
        )
        await #expect {
            _ = try await Self.run(source: Self.cleanSource, manifest: manifest)
        } throws: { error in
            String(describing: error).contains("failed before compiling anything")
        }
    }
}
