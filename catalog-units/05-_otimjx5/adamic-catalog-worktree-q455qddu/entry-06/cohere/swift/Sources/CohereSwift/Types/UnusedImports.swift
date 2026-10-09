import Foundation
import SwiftSyntax

/*
 `unused-import`: an `import` in a file of ours that nothing in the file uses, read from the build's index store.

 The question is narrow because `MemberImportVisibility` is on for every target we check: each file must
 import the modules whose members it uses, so a file's imports are its own and an import is used exactly when
 something written in that file resolves into the module or into a module it re-exports. The index answers
 both halves. Each file's unit lists the modules the file imports; each module's records declare its symbols,
 so a symbol a file refers to is placed in the module that declares it.

 An import counts as used when a reference in the file:
   - resolves to a declaration of that module, or of a module it might re-export (`ModuleReexports`, read
     with `Reach.assumed`: a system module is taken to re-export everything it imports);
   - or names that module anywhere in its signature. The index does not record every member reference: on
     Presence, `$0.rawValue` on a MediaPipe `BlendShape` reached through a closure parameter has no occurrence
     at all, yet the file needs MediaPipe's import for it. The closure's type is in the signature of the
     function the file does call, and crediting every module a referenced signature names closes that gap.
 A reference to a module that is also imported directly credits only that module, so `import AppKit` beside
 `import Foundation` is not kept alive by `URL`.

 Zero false positives is the bar, so a file is left unchecked, and counted, when the index cannot vouch for
 it: no unit newer than the file, a reference no module claims, or any `#if` in it, since the index describes
 the configuration the build compiled and an inactive branch may be the only user of an import.
 `@_exported` imports are API and never reported.

 Crediting alone keeps both halves of a pair such as `import AppKit` beside `import SwiftUI` in a file that
 uses `CGFloat`, since each import's closure covers the reference though either alone would do. `redundant`
 then removes what the others already bring, so the rule reports a used import too when the file can do
 without it.

 Measured before it shipped, every finding was removed on a `git archive` copy and `swift build --build-tests`
 stayed clean: 143 of 143 on ahraos-macos, 828 of 828 on Presence. Against SwiftLint 0.65.1's analyzer
 `unused_import` on ahraos-macos, 123 to this rule's 143: 112 the same and 31 this rule's alone. Of SwiftLint's
 other 11, ten break the build once removed with this rule's. Five are SwiftLint flagging every import that
 brings a module the file needs (KingdomFeedLog, KingdomFeedLogLine, RowRegistration and the two SidebarDrag
 geometries, where this rule removes the others and keeps one); five are imports the file needs outright
 (Pane, LocalPtyByteSource, ScrollHoldAssert, ServicesClientByteSource, AhraOsShellHelper). The eleventh is
 real and missed: `import Darwin` in AhraOsServices/main.swift, beside `import AppKit`, for `SIG_IGN`. That is
 the Darwin overlay's, and an overlay's top-level declarations become visible through re-exported parts of
 its Clang module in some cases (Darwin's, through ObjectiveC or Dispatch) and not in others (CoreGraphics's,
 through Foundation), so nothing here believes it.
 */
struct UnusedImports {
    static let ruleName = "cohere-swift/correctness-no-unused-import"

    /* Modules every file imports without saying so. */
    static let implicitModules: Set<String> = [
        "Swift", "_Concurrency", "_StringProcessing", "_SwiftConcurrencyShims", "SwiftOnoneSupport",
    ]

    struct Result {
        /* Each unused import, with the declaration as written for the report's line. */
        var findings: [(finding: FindingRecord, subject: String)] = []
        var filesChecked = 0
        /* Files the index could not vouch for, by path, with why. */
        var filesUnchecked: [String: String] = [:]
        /* Imports judged, and imports never reported by design (re-exported API), in the checked files. */
        var importsChecked = 0
        var importsExported = 0
    }

    static let notCompiled = "the build has not compiled it as it stands"
    static let conditional = "it has #if, and the index describes only the configuration the build compiled"
    static let unresolvedReference = "a reference in it resolves to no module the index declares"

    struct Source {
        let store: IndexStore
        let unit: IndexStore.Unit
    }

    let stores: [IndexStore]
    let demangler: SwiftDemangler?

    func run(files: [ParsedFile]) -> Result {
        var result = Result()
        let units = stores.flatMap { store in store.units().map { Source(store: store, unit: $0) } }
        let fresh = IndexStore.freshUnits(of: files, in: stores)
        var checked: [(file: ParsedFile, source: Source, references: [IndexStore.RecordOccurrence])] = []
        for file in files {
            guard let described = fresh[file.url.path] else {
                result.filesUnchecked[file.url.path] = Self.notCompiled
                continue
            }
            let source = Source(store: described.store, unit: described.unit)
            if Self.hasConditionalCompilation(file.tree) {
                result.filesUnchecked[file.url.path] = Self.conditional
                continue
            }
            let references = source.unit.ownRecords.flatMap { source.store.occurrences(inRecord: $0) ?? [] }
                .filter { $0.isReference && $0.kind != IndexStore.moduleKind }
            checked.append((file, source, references))
        }

        let referenced = Set(checked.flatMap { $0.references.map(\.symbol) })
        let resolver = ModuleResolver(units: units, referenced: referenced, demangler: demangler)
        let reexports = ModuleReexports(units: units.map(\.unit), files: files)

        for entry in checked {
            let imports = Self.imports(in: entry.file.tree)
            let imported = Set(imports.map(\.module))
            var used: Set<String> = []
            /* What each reference needs, the standard library left out: what the file's imports must keep in reach. */
            var needs: [Need] = []
            var unresolved = 0
            for reference in entry.references {
                guard let declaring = resolver.declaringModules(of: reference) else {
                    unresolved += 1
                    continue
                }
                let modules = declaring.union(resolver.modulesNamed(inSignatureOf: reference.symbol))
                /*
                 The standard library is every file's without an import, and every system module imports it, so
                 a reference into it credits no import through a re-export. Found by the fixture: `Int` kept
                 `import Foundation` alive in a file that used nothing of Foundation's.
                 */
                let reexported = modules.subtracting(Self.implicitModules)
                needs.append(
                    Need(modules: reexported, declaring: declaring, headers: resolver.declaringHeaders(of: reference))
                )
                let direct = modules.intersection(imported)
                if !direct.isEmpty {
                    used.formUnion(direct)
                    continue
                }
                for module in imported
                where !reexported.isDisjoint(with: reexports.closure(of: module, reach: .assumed)) {
                    used.insert(module)
                }
            }
            if unresolved > 0 {
                result.filesUnchecked[entry.file.url.path] = Self.unresolvedReference
                continue
            }
            result.filesChecked += 1
            var judged: [ImportDeclaration] = []
            for declaration in imports where !Self.implicitModules.contains(declaration.module) {
                if declaration.isExported {
                    result.importsExported += 1
                    continue
                }
                result.importsChecked += 1
                judged.append(declaration)
            }
            let unused = judged.filter { !used.contains($0.module) }
            for declaration in unused {
                let finding = entry.file.finding(
                    at: declaration.node,
                    rule: Self.ruleName,
                    messageId: "unusedImport",
                    message:
                        "Nothing in this file uses \(declaration.module): no name here resolves into it or into a module it re-exports. Remove the import.",
                    suggestions: [
                        FindingRecord.Suggestion(
                            message: "Remove `import \(declaration.module)`",
                            fixes: [Self.removal(of: declaration.node, in: entry.file)],
                        )
                    ],
                )
                result.findings.append((finding, declaration.node.trimmedDescription))
            }
            let kept = imported.subtracting(unused.map(\.module))
            for redundant in Self.redundant(
                judged.filter { used.contains($0.module) },
                kept: kept,
                needs: needs,
                reexports: reexports,
            ) {
                let others = redundant.coveredBy.map { "`import \($0)`" }.joined(separator: ", ")
                let verb = redundant.coveredBy.count == 1 ? "re-exports" : "re-export"
                let finding = entry.file.finding(
                    at: redundant.declaration.node,
                    rule: Self.ruleName,
                    messageId: "redundantImport",
                    message:
                        "Everything this file uses through \(redundant.declaration.module) also comes through \(others), which \(verb) it. Remove the import.",
                    suggestions: [
                        FindingRecord.Suggestion(
                            message: "Remove `import \(redundant.declaration.module)`",
                            fixes: [Self.removal(of: redundant.declaration.node, in: entry.file)],
                        )
                    ],
                )
                result.findings.append((finding, redundant.declaration.node.trimmedDescription))
            }
        }
        return result
    }

    /*
     What one reference needs from the file's imports: every module it names, the declaring ones among them, and
     the headers that declare it when it is a Clang declaration. A declaring module is in reach when the module is
     visible or, for a Clang declaration, when one of its headers is.
     */
    struct Need {
        var modules: Set<String>
        var declaring: Set<String>
        var headers: Set<String>
    }

    struct ImportDeclaration {
        var module: String
        var isExported: Bool
        var node: ImportDeclSyntax
    }

    /* The file's top-level imports. One inside `#if` is never reached here: such a file is left unchecked. */
    static func imports(in tree: SourceFileSyntax) -> [ImportDeclaration] {
        tree.statements.compactMap { statement in
            guard let declaration = statement.item.as(ImportDeclSyntax.self), let first = declaration.path.first else {
                return nil
            }
            let isExported = declaration.attributes.contains { attribute in
                attribute.as(AttributeSyntax.self)?.attributeName.trimmedDescription == "_exported"
            }
            return ImportDeclaration(module: first.name.text, isExported: isExported, node: declaration)
        }
    }

    static func hasConditionalCompilation(_ tree: SourceFileSyntax) -> Bool {
        final class Finder: SyntaxVisitor {
            var found = false
            override func visit(_ node: IfConfigDeclSyntax) -> SyntaxVisitorContinueKind {
                found = true
                return .skipChildren
            }
        }
        let finder = Finder(viewMode: .sourceAccurate)
        finder.walk(tree)
        return finder.found
    }

    /*
     The edit that removes the import's line: its indentation, the declaration, a trailing `//` comment, and the
     newline that ends it. It starts after the leading trivia, never in it: on a file's first import that trivia is
     the whole file header, and below a line that ends in a comment it begins with that line's newline, so a range
     reaching back from it took the line above. Both deleted what no finding named. Where code shares the line,
     only the declaration goes.
     */
    static func removal(of node: ImportDeclSyntax, in file: ParsedFile) -> FindingRecord.Edit {
        let utf8 = Array(file.source.utf8)
        let start = node.positionAfterSkippingLeadingTrivia.utf8Offset
        let end = node.endPositionBeforeTrailingTrivia.utf8Offset
        let blank: (UInt8) -> Bool = { $0 == UInt8(ascii: " ") || $0 == UInt8(ascii: "\t") }
        var lineStart = start
        while lineStart > 0 && blank(utf8[lineStart - 1]) {
            lineStart -= 1
        }
        var lineEnd = end
        while lineEnd < utf8.count && utf8[lineEnd] != UInt8(ascii: "\n") {
            lineEnd += 1
        }
        let rest = utf8[end..<lineEnd].drop(while: blank)
        guard lineStart == 0 || utf8[lineStart - 1] == UInt8(ascii: "\n"), rest.isEmpty || rest.starts(with: "//".utf8)
        else {
            return FindingRecord.Edit(start: start, end: end, text: "")
        }
        return FindingRecord.Edit(start: lineStart, end: lineEnd < utf8.count ? lineEnd + 1 : lineEnd, text: "")
    }

    /*
     Imports the file can do without because its other imports already bring everything it needs from them:
     `import AppKit` beside `import SwiftUI` in a file that uses `CGFloat` and `NSView`, where SwiftUI re-exports
     AppKit. An import can go only when every module it might have brought to a reference (`Reach.bounded`) is
     still reached through what stays, by edges read from interfaces and module maps alone (`Reach.known`). A
     module a reference needs counts whether it declares the reference or only appears in its signature, since
     the signature stands in for member references the index leaves out. Imports are tried least specific
     first, the widest known reach first and then the last written, so the narrower import that names what the
     file uses stays. Each removal is judged against the imports that remain after the ones before it, so the
     set reported is safe to remove together.
     */
    static func redundant(
        _ candidates: [ImportDeclaration],
        kept: Set<String>,
        needs: [Need],
        reexports: ModuleReexports,
    ) -> [(declaration: ImportDeclaration, coveredBy: [String])] {
        var kept = kept
        let ordered = candidates.enumerated().sorted { first, second in
            let firstReach = reexports.reach(of: first.element.module, reach: .known).count
            let secondReach = reexports.reach(of: second.element.module, reach: .known).count
            return firstReach != secondReach ? firstReach > secondReach : first.offset > second.offset
        }.map(\.element)
        var removed: [(declaration: ImportDeclaration, provided: Set<String>, headers: Set<String>)] = []
        for candidate in ordered where kept.contains(candidate.module) {
            let reach = reexports.reach(of: candidate.module, reach: .bounded)
            let remaining = kept.subtracting([candidate.module])
            let stillReached = remaining.reduce(into: implicitModules) { reached, module in
                reached.formUnion(reexports.reach(of: module, reach: .known))
            }
            var provided: Set<String> = []
            var headers: Set<String> = []
            /* An import another one re-exports whole brings nothing that one does not, whatever the file needs. */
            if !stillReached.contains(candidate.module) {
                let visible = reexports.visibleHeaders(through: stillReached)
                let covered = needs.allSatisfy { need in
                    need.modules.intersection(reach).allSatisfy { module in
                        provided.insert(module)
                        if stillReached.contains(module) {
                            return true
                        }
                        guard need.declaring.contains(module), !need.headers.isDisjoint(with: visible) else {
                            return false
                        }
                        headers.formUnion(need.headers)
                        return true
                    }
                }
                guard covered else { continue }
            }
            kept = remaining
            removed.append((candidate, provided, headers))
        }
        /* Named after every removal, so the imports a message points to are ones the file keeps. */
        return removed.map { entry in
            let coveredBy = kept.filter { module in
                let reached = reexports.reach(of: module, reach: .known)
                return reached.contains(entry.declaration.module) || !reached.isDisjoint(with: entry.provided)
                    || !reexports.visibleHeaders(through: reached).isDisjoint(with: entry.headers)
            }.sorted()
            return (entry.declaration, coveredBy.isEmpty ? kept.sorted() : coveredBy)
        }
    }

    /*
     Which module declares a symbol, from every record the store holds. Only the symbols the checked files refer to
     are kept, so the table stays the size of the question rather than the size of the SDK.
     */
    struct ModuleResolver {
        private var declaring: [String: Set<String>] = [:]
        /* Top-level names Clang modules declare, for symbols Swift spells by name (`s:So11sockaddr_unV`, `s:SC7AF_UNIX`). */
        private var clangNames: [String: Set<String>] = [:]
        /* The headers that declare a Clang symbol or name, for coverage finer than a module (`ModuleReexports.visibleHeaders`). */
        private var declaringHeaders: [String: Set<String>] = [:]
        private var clangNameHeaders: [String: Set<String>] = [:]
        private let knownModules: Set<String>
        private let demangler: SwiftDemangler?

        init(units: [Source], referenced: Set<String>, demangler: SwiftDemangler?) {
            self.demangler = demangler
            knownModules = Set(units.map(\.unit.module).filter { !$0.isEmpty })
            var wanted = referenced
            var wantedClangNames: Set<String> = []
            for symbol in referenced {
                if let base = Self.accessorBase(symbol) {
                    wanted.insert(base)
                }
                if let objectiveCType = Self.objectiveCContainer(symbol) {
                    wanted.insert(objectiveCType)
                }
                if let name = Self.clangImportedName(symbol) {
                    wantedClangNames.insert(name)
                }
            }
            /*
             A record can be named by more than one unit, and each unit's module declares what it holds: read once,
             credited to every module whose unit names it. Crediting only the first unit read put Accelerate's
             vDSP in another module on Presence, and `import Accelerate` read as unused where every call needs it.
             */
            var modulesOfRecord: [String: (store: IndexStore, modules: Set<String>, file: String)] = [:]
            for source in units {
                for record in source.unit.allRecords {
                    if modulesOfRecord[record] == nil {
                        modulesOfRecord[record] = (source.store, [], source.unit.recordFiles[record] ?? "")
                    }
                    modulesOfRecord[record]?.modules.insert(source.unit.module)
                }
            }
            for (record, holder) in modulesOfRecord {
                let header = holder.file.hasSuffix(".h") ? holder.file : nil
                for occurrence in holder.store.occurrences(inRecord: record) ?? [] where occurrence.isDeclaration {
                    if wanted.contains(occurrence.symbol) {
                        declaring[occurrence.symbol, default: []].formUnion(holder.modules)
                        if let header, occurrence.symbol.hasPrefix("c:") {
                            declaringHeaders[occurrence.symbol, default: []].insert(header)
                        }
                    }
                    if occurrence.symbol.hasPrefix("c:"), wantedClangNames.contains(occurrence.name) {
                        clangNames[occurrence.name, default: []].formUnion(holder.modules)
                        if let header {
                            clangNameHeaders[occurrence.name, default: []].insert(header)
                        }
                    }
                }
            }
        }

        /* The modules that declare what the occurrence refers to, or nil when nothing the store holds says. */
        func declaringModules(of occurrence: IndexStore.RecordOccurrence) -> Set<String>? {
            let symbol = occurrence.symbol
            if let modules = declaring[symbol] {
                return modules
            }
            if let base = Self.accessorBase(symbol), let modules = declaring[base] {
                return modules
            }
            if let module = Self.categoryModule(symbol) {
                return [module]
            }
            if let container = Self.objectiveCContainer(symbol), let modules = declaring[container] {
                return modules
            }
            if let module = FileSymbols.Occurrence(
                line: 0,
                column: 0,
                symbol: symbol,
                name: occurrence.name,
                isReference: true,
            ).declaringModule {
                return [module]
            }
            if symbol.hasPrefix("s:s")
                || (symbol.hasPrefix("s:S") && !symbol.hasPrefix("s:So") && !symbol.hasPrefix("s:SC"))
            {
                return ["Swift"]
            }
            if let name = Self.clangImportedName(symbol), let modules = clangNames[name] {
                return modules
            }
            return nil
        }

        /*
         The headers that declare what a Clang occurrence refers to, followed the way `declaringModules` follows it,
         or empty when it is no Clang declaration a header holds (a Swift one, a category member).
         */
        func declaringHeaders(of occurrence: IndexStore.RecordOccurrence) -> Set<String> {
            let symbol = occurrence.symbol
            if declaring[symbol] != nil {
                return declaringHeaders[symbol] ?? []
            }
            if let base = Self.accessorBase(symbol), declaring[base] != nil {
                return declaringHeaders[base] ?? []
            }
            if Self.categoryModule(symbol) != nil {
                return []
            }
            if let container = Self.objectiveCContainer(symbol), declaring[container] != nil {
                return declaringHeaders[container] ?? []
            }
            if FileSymbols.Occurrence(line: 0, column: 0, symbol: symbol, name: occurrence.name, isReference: true)
                .declaringModule != nil
            {
                return []
            }
            if symbol.hasPrefix("s:s")
                || (symbol.hasPrefix("s:S") && !symbol.hasPrefix("s:So") && !symbol.hasPrefix("s:SC"))
            {
                return []
            }
            if let name = Self.clangImportedName(symbol), clangNames[name] != nil {
                return clangNameHeaders[name] ?? []
            }
            return []
        }

        /* Every module of the build the declaration's demangled signature names, `CoreFoundation.CGFloat` and the like. */
        func modulesNamed(inSignatureOf symbol: String) -> Set<String> {
            guard let demangler, let declaration = demangler.declaration(ofSymbol: symbol) else { return [] }
            var modules: Set<String> = []
            var identifier = ""
            for character in declaration {
                if character.isLetter || character.isNumber || character == "_" {
                    identifier.append(character)
                    continue
                }
                if character == ".", knownModules.contains(identifier) {
                    modules.insert(identifier)
                }
                identifier = ""
            }
            return modules
        }

        /* An accessor's property or subscript: `...vg` (getter) to `...vp`, `...ig` to `...ip`. */
        static func accessorBase(_ symbol: String) -> String? {
            guard symbol.hasPrefix("s:"), symbol.count > 4 else { return nil }
            let characters = Array(symbol)
            let accessor = characters[characters.count - 1]
            let kind = characters[characters.count - 2]
            guard (kind == "v" || kind == "i"), "gsmrwWMaAlux".contains(accessor) else { return nil }
            return String(characters.dropLast()) + "p"
        }

        /* The module a category member carries in its own name: `c:@CM@Foundation@@objc(cs)NSString(im)...`. */
        static func categoryModule(_ symbol: String) -> String? {
            guard symbol.hasPrefix("c:@CM@") else { return nil }
            let rest = symbol.dropFirst("c:@CM@".count)
            guard let end = rest.firstIndex(of: "@") else { return nil }
            return String(rest[..<end])
        }

        /* The class or protocol an Objective-C member belongs to: `c:objc(cs)NSView(im)setIdentifier:` to `c:objc(cs)NSView`. */
        static func objectiveCContainer(_ symbol: String) -> String? {
            for prefix in ["c:objc(cs)", "c:objc(pl)"] where symbol.hasPrefix(prefix) {
                let rest = symbol.dropFirst(prefix.count)
                guard let end = rest.firstIndex(of: "(") else { return nil }
                return prefix + rest[..<end]
            }
            return nil
        }

        /* The C name Swift wraps in `s:So<length><name>` or `s:SC<length><name>`. */
        static func clangImportedName(_ symbol: String) -> String? {
            guard symbol.hasPrefix("s:So") || symbol.hasPrefix("s:SC") else { return nil }
            let rest = symbol.utf8.dropFirst(4)
            let digits = rest.prefix { $0 >= UInt8(ascii: "0") && $0 <= UInt8(ascii: "9") }
            guard let length = Int(String(decoding: digits, as: UTF8.self)), length > 0 else { return nil }
            let name = rest.dropFirst(digits.count).prefix(length)
            return name.count == length ? String(decoding: name, as: UTF8.self) : nil
        }
    }
}
