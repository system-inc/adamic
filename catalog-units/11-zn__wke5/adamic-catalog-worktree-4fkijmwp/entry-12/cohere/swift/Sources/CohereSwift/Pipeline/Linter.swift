import Foundation

/*
 Runs every rule over the files in scope and accounts for what each rule did.

 The accounting is half the job. A rule that was offered no file, a rule that declined every file it was
 offered, and a rule that read every file and found nothing all produce the same empty finding list, and
 a reader deciding whether a clean run means anything needs to know which of the three it was looking at.
 */
struct Linter {
    /* The rules' findings and the record that says how they were produced. */
    struct Result {
        var findings: [FindingRecord]
        var record: LintRecord
    }

    let configuration: RuleConfiguration
    let fileRules: [any FileRule]
    var typedRules: [any TypedFileRule] = []
    /* What names resolve to in the files some enabled typed rule applies to, fetched before the walk (`SymbolProvider`). */
    var symbols = SymbolProvider.Result(symbols: [:], unavailable: [:], fromIndex: 0, fromSourcekitd: 0)

    /* `manifests` holds each owned package's parsed `Package.swift`, keyed by the package root's path; vendored packages are absent and never checked. */
    /* `reusable` holds, by path, what the fix phase's last walk found per rule in exactly the text being linted; those files are not walked again. */
    func run(
        package: PackageModel,
        manifests: [String: ParsedFile],
        files: [ParsedFile],
        reusable: [String: [String: [FindingRecord]]] = [:],
    ) async -> Result {
        let ownedPackages = package.allPackages.filter { $0.root == package.root || !package.isVendored($0) }
        var findings: [FindingRecord] = []
        var listening: [String: Int] = [:]
        var reporting: [String: Int] = [:]
        var scopedOff: [String: Int] = [:]
        var rulesRun = 0

        for rule in RuleRegistry.packageRules {
            let severity = configuration.severity(of: rule.name)
            guard severity != .off else {
                scopedOff[rule.name] = 1
                continue
            }
            rulesRun += 1
            for member in ownedPackages {
                listening[rule.name, default: 0] += 1
                let found = rule.findings(in: member, manifest: manifests[member.root.path]).map {
                    Self.applying(severity, to: $0)
                }
                reporting[rule.name, default: 0] += found.count
                findings.append(contentsOf: found)
            }
        }

        let enabled = fileRules.filter { configuration.severity(of: $0.name) != .off }
        for rule in fileRules where configuration.severity(of: rule.name) == .off {
            scopedOff[rule.name] = files.count
        }
        rulesRun += enabled.count
        let enabledTyped = typedRules.filter { configuration.severity(of: $0.name) != .off }
        for rule in typedRules where configuration.severity(of: rule.name) == .off {
            scopedOff[rule.name] = files.count
        }
        rulesRun += enabledTyped.count
        let symbols = symbols

        /*
         Every file is walked on its own, so files are walked side by side. Each answers with what every enabled
         rule found in it (nil where the rule does not apply), and the results are laid out rule by rule, file by
         file, the order a sequential walk printed, so the stream is the same however the work was scheduled.
         */
        let perFile = await withTaskGroup(of: (Int, [[FindingRecord]?]).self) { group in
            for (index, file) in files.enumerated() {
                group.addTask {
                    let byRule = enabled.map { rule -> [FindingRecord]? in
                        guard rule.applies(to: file) else { return nil }
                        return reusable[file.url.path]?[rule.name] ?? rule.findings(in: file)
                    }
                    /* A typed rule with no symbols for a file it applies to judged nothing there; the crash list below says so. */
                    let byTypedRule = enabledTyped.map { rule -> [FindingRecord]? in
                        guard rule.applies(to: file), let found = symbols.symbols[file.url.path] else { return nil }
                        return rule.findings(in: file, symbols: found)
                    }
                    return (index, byRule + byTypedRule)
                }
            }
            var collected = [[[FindingRecord]?]](repeating: [], count: files.count)
            for await (index, byRule) in group {
                collected[index] = byRule
            }
            return collected
        }
        for (ruleIndex, name) in (enabled.map(\.name) + enabledTyped.map(\.name)).enumerated() {
            let severity = configuration.severity(of: name)
            for fileIndex in files.indices {
                guard let raw = perFile[fileIndex][ruleIndex] else { continue }
                listening[name, default: 0] += 1
                let found = raw.map { Self.applying(severity, to: $0) }
                reporting[name, default: 0] += found.count
                findings.append(contentsOf: found)
            }
        }

        let ranNames = (RuleRegistry.packageRules.map(\.name) + fileRules.map(\.name) + typedRules.map(\.name)).filter {
            scopedOff[$0] == nil
        }
        /* A file a typed rule applies to and no source of symbols described: nothing that rule would say about it was said. */
        let unchecked = files.filter { file in
            symbols.unavailable[file.url.path] != nil && enabledTyped.contains { $0.applies(to: file) }
        }
        let crashes = unchecked.map { file in
            LintRecord.Crash(
                file: file.url.path,
                error:
                    "the typed rules could not read what its names resolve to: \(symbols.unavailable[file.url.path] ?? "")",
            )
        }
        let silent = ranNames.filter { (listening[$0] ?? 0) == 0 }.sorted()
        let watchedAndQuiet = ranNames.filter { (listening[$0] ?? 0) > 0 && (reporting[$0] ?? 0) == 0 }.count

        let record = LintRecord(
            findings: findings.count,
            rulesRun: rulesRun,
            filesWalked: files.count,
            nodesVisited: files.reduce(0) { $0 + $1.nodeCount },
            elapsedMilliseconds: 0,
            reusedFrom: "",
            rulesSilent: silent,
            rulesWatchedAndQuiet: watchedAndQuiet,
            crashes: crashes,
            rulesScopedOff: scopedOff,
            rulesNotConfigured: [],
            configurationNote: configuration.note,
        )
        return Result(findings: findings, record: record)
    }

    /* A rule reports at error; the config can lower it to warning. Both count as findings. */
    private static func applying(_ severity: RuleConfiguration.Severity, to finding: FindingRecord) -> FindingRecord {
        var adjusted = finding
        adjusted.severity = severity == .warning ? .warning : .error
        return adjusted
    }
}
