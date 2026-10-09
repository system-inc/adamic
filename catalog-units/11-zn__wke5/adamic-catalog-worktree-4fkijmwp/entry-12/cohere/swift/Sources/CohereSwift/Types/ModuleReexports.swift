import Foundation

/*
 What each module makes visible beyond itself, for `unused-import`: an `import AppKit` brings CoreGraphics with
 it, so a file that uses `CGFloat` and imports both AppKit and CoreGraphics needs only one of them.

 Three readings of each module, because the rule asks questions with opposite safe directions:
   - `known` holds the edges the module's own text states, each a whole module. A Swift module's are the
     `@_exported import` lines of its `.swiftinterface`. A Clang module's are the modules a header it re-exports
     from imports whole: through the other's umbrella header, a header the other's map declares at its top, or
     `@import`, on a line the build certainly compiled (`Conditions`). A module of ours re-exports what its
     files `@_exported import`.
   - `assumed` holds, for each system module, everything it imports, whether or not it re-exports it.
   - `bounded` holds the most each module can re-export. A Swift module whose `.swiftinterface` was read
     re-exports nothing its `@_exported` lines leave out, so those lines bound it, a submodule or a single
     declaration counted as its whole module. Any other module, a Clang one, one whose interface cannot be
     read, or a dependency whose sources are not ours to read, is bounded by everything it imports.
 Whether an import is used at all reads both (`Reach.assumed`): crediting an import with too much only keeps
 it, so a module whose interface or map cannot be read keeps its imports alive rather than going unread.
 Whether an import can go because the others already bring what it brought asks two things. What the others
 bring reads `known` alone (`Reach.known`): there, an edge too many would remove an import the file needs, so
 nothing is believed that was not read. What the import itself might have brought reads `Reach.bounded`: an
 edge too few there would miss something it brought, and the bound is never smaller than the truth. It is
 tighter than `assumed` where that matters: AppKit imports SwiftUICore, which imports Combine, yet AppKit's
 interface does not `@_exported` either, so `import AppKit` never brought Combine, and a file that hands a
 `Combine.Publisher` to SwiftUI's `onReceive` can lose AppKit when it keeps SwiftUI.

 A Clang module's `export *` re-exports a header's worth of another module as often as a whole one, and that
 part is no module: CoreVideo's headers bring CoreGraphics's for C, not for Swift, and no Swift file sees
 `CGImage` through Foundation. `visibleHeaders` answers that finer question, one header at a time, for a
 reference a Clang header declares.

 Swift and Clang modules of one name share a node. `import AppKit` loads the Swift overlay, which re-exports the
 Clang module of the same name, and a Clang module imported whole brings its overlay with it.
 */
struct ModuleReexports {
    enum Reach {
        /* The edges read from interfaces, module maps and our own `@_exported` lines. */
        case known
        /* Those, and every module a system module imports. */
        case assumed
        /* The most each module can re-export. */
        case bounded
    }

    private(set) var known: [String: Set<String>] = [:]
    private(set) var assumed: [String: Set<String>] = [:]
    private(set) var bounded: [String: Set<String>] = [:]
    private let headers: Headers

    init(units: [IndexStore.Unit], files: [ParsedFile]) {
        var interfaces: [String: InterfaceExports?] = [:]
        var moduleMaps: [String: ModuleMap?] = [:]
        var clangModules: [Headers.ClangModule] = []
        let moduleOfFile = Dictionary(
            units.filter { !$0.isSystem }.map { ($0.mainFile, $0.module) },
            uniquingKeysWith: { first, _ in first },
        )
        /* A module of the build is bounded by its files' `@_exported` lines only when every one of its files was read. */
        let read = Set(files.map { $0.url.resolvingSymlinksInPath().path })
        let unread = Set(units.filter { !$0.isSystem && !read.contains($0.mainFile) }.map(\.module))
        for unit in units where !unit.isSystem && !unit.module.isEmpty && unread.contains(unit.module) {
            bounded[unit.module, default: []].formUnion(unit.importedModules)
        }
        for unit in units where unit.isSystem && !unit.module.isEmpty {
            assumed[unit.module, default: []].formUnion(unit.importedModules)
            if unit.outputFile.hasSuffix(".swiftinterface") {
                if interfaces[unit.outputFile] == nil {
                    interfaces[unit.outputFile] = Self.exportedImports(ofInterface: unit.outputFile)
                }
                let exports = interfaces[unit.outputFile].flatMap { $0 }
                known[unit.module, default: []].formUnion(exports?.whole ?? [])
                bounded[unit.module, default: []].formUnion(exports?.atMost ?? Set(unit.importedModules))
                continue
            }
            bounded[unit.module, default: []].formUnion(unit.importedModules)
            func map(at path: String) -> ModuleMap? {
                if moduleMaps[path] == nil {
                    moduleMaps[path] = (try? String(contentsOfFile: path, encoding: .utf8)).map(ModuleMap.init(text:))
                }
                return moduleMaps[path] ?? nil
            }
            for candidate in Self.moduleMapCandidates(of: unit) {
                var path = candidate
                if let extern = map(at: path)?.externs[unit.module] {
                    path =
                        URL(fileURLWithPath: path).deletingLastPathComponent().appendingPathComponent(extern)
                        .standardizedFileURL.path
                }
                guard let exports = map(at: path)?.exports(ofModule: unit.module) else { continue }
                clangModules.append(Headers.ClangModule(unit: unit, exports: exports, moduleMap: path))
                break
            }
        }
        /*
         `export *` re-exports what the module's headers import, and a header imports a whole module only through its
         umbrella header (`#import <Foundation/Foundation.h>`), a header its map declares at the top
         (`<dispatch/dispatch.h>`), or `@import Name;`. Any other header of it (`<CoreGraphics/CGGeometry.h>`)
         brings that header's declarations and no more. So `known` takes the whole-module imports alone, and
         `visibleHeaders` answers for the rest one header at a time.
         */
        headers = Headers(clangModules: clangModules)
        for clangModule in clangModules {
            let imported = Set(clangModule.unit.importedModules)
            known[clangModule.unit.module, default: []].formUnion(
                headers.wholeImports(of: clangModule.unit.module, named: clangModule.exports.named).intersection(
                    imported
                )
            )
        }
        for file in files {
            guard let module = moduleOfFile[file.url.resolvingSymlinksInPath().path] else { continue }
            for declaration in UnusedImports.imports(in: file.tree) where declaration.isExported {
                known[module, default: []].insert(declaration.module)
            }
        }
        for (module, edges) in known {
            assumed[module, default: []].formUnion(edges)
            bounded[module, default: []].formUnion(edges)
        }
    }

    /* Every module visible through an import of `module`, itself not included unless a cycle leads back to it. */
    func closure(of module: String, reach: Reach) -> Set<String> {
        let edges: [String: Set<String>]
        switch reach {
            case .known:
                edges = known
            case .assumed:
                edges = assumed
            case .bounded:
                edges = bounded
        }
        var seen: Set<String> = []
        var pending = [module]
        while let current = pending.popLast() {
            for next in edges[current, default: []] where seen.insert(next).inserted {
                pending.append(next)
            }
        }
        return seen
    }

    /* The module and everything visible through it. */
    func reach(of module: String, reach: Reach) -> Set<String> {
        closure(of: module, reach: reach).union([module])
    }

    /*
     The headers whose declarations are visible through modules imported whole, `known` reach in hand. Finer than
     a module: Foundation's umbrella imports CoreFoundation's, which includes `<stdlib.h>`, which brings Darwin's
     `_stdlib.h`, so `getenv` comes through Foundation without `import Darwin`.
     */
    func visibleHeaders(through modules: Set<String>) -> Set<String> {
        modules.reduce(into: Set<String>()) { visible, module in visible.formUnion(headers.visible(through: module)) }
    }

    /*
     The Clang headers of the build and what each makes visible. A module imported whole shows its own headers,
     those of `explicit` submodules aside. A header's `#import` and `#include` lines, on lines the build
     certainly compiled, make the header they name visible too when the submodule the header forms re-exports
     what it imports: its own block's `export *` when the map declares the header, the top-level `export *` for
     the umbrella, `module * { export * }` for a header the umbrella brings in. The walk goes on from a header
     only when it re-exports the same way, and passes straight through a `textual header`, which is pasted
     where it is included. A header is found by the name an include spells (`CoreGraphics/CGGeometry.h`,
     `sys/types.h`) among the headers this build compiled; one it cannot find is left out, which can only make
     less visible.
     */
    final class Headers {
        struct ClangModule {
            var unit: IndexStore.Unit
            var exports: ModuleMap.Exports
            /* The map the exports were read from, which places the umbrella header. */
            var moduleMap: String
        }

        private var modulesOfHeader: [String: Set<String>] = [:]
        private var headersOfModule: [String: Set<String>] = [:]
        private var pathOfName: [String: String] = [:]
        /* Modules whose maps say `export *` at the top, and those that say it for every inferred submodule. */
        private var topReexporting: Set<String> = []
        private var everySubmoduleReexporting: Set<String> = []
        /* Headers a map declares, with whether their block says `export *`. */
        private var declaredReexports: [String: Bool] = [:]
        private var textualHeaders: Set<String> = []
        /* The umbrella and top-level headers, each importing its module whole. */
        private var wholeModuleOfHeader: [String: String] = [:]
        private var umbrellaOfModule: [String: String] = [:]
        private var parsed: [String: (headers: [String], modules: [String])] = [:]
        private var visibleThrough: [String: Set<String>] = [:]

        init(clangModules: [ClangModule]) {
            for clangModule in clangModules {
                let module = clangModule.unit.module
                let exports = clangModule.exports
                let place = { (header: String) in Self.memberPath(header, moduleMap: clangModule.moduleMap) }
                /* An umbrella header that only imports declares nothing, so no record names it; its map places it. */
                var declared = Set(exports.headerExports.keys.map(place))
                if let umbrella = exports.umbrella {
                    declared.insert(place(umbrella))
                    umbrellaOfModule[module] = place(umbrella)
                    wholeModuleOfHeader[place(umbrella)] = module
                }
                for header in exports.topHeaders {
                    wholeModuleOfHeader[place(header)] = module
                }
                for (header, reexports) in exports.headerExports {
                    declaredReexports[place(header)] = reexports
                }
                for header in exports.textualHeaders.map(place) where FileManager.default.fileExists(atPath: header) {
                    textualHeaders.insert(header)
                    if let name = Self.includeName(of: header), pathOfName[name] == nil {
                        pathOfName[name] = header
                    }
                }
                let explicit = Set(exports.explicitHeaders.map(place))
                let own = Self.ownHeaders(of: clangModule.unit).union(
                    declared.filter { FileManager.default.fileExists(atPath: $0) }
                )
                headersOfModule[module, default: []].formUnion(own.subtracting(explicit))
                for header in own {
                    modulesOfHeader[header, default: []].insert(module)
                    if let name = Self.includeName(of: header), pathOfName[name] == nil {
                        pathOfName[name] = header
                    }
                }
                if exports.all {
                    topReexporting.insert(module)
                }
                if exports.everySubmodule {
                    everySubmoduleReexporting.insert(module)
                }
            }
        }

        /* A Clang module's own headers: those its records describe, and those it read from its framework. */
        static func ownHeaders(of unit: IndexStore.Unit) -> Set<String> {
            var own = Set(unit.recordFiles.values.filter { $0.hasSuffix(".h") })
            own.formUnion(unit.files.filter { $0.hasSuffix(".h") && $0.contains("/\(unit.module).framework/Headers/") })
            return own
        }

        /* A framework map names headers under `Headers/`; any other names them beside itself. */
        static func memberPath(_ header: String, moduleMap: String) -> String {
            if let range = moduleMap.range(of: ".framework/Modules/", options: .backwards) {
                return String(moduleMap[..<range.lowerBound]) + ".framework/Headers/" + header
            }
            return URL(fileURLWithPath: moduleMap).deletingLastPathComponent().appendingPathComponent(header)
                .standardizedFileURL.path
        }

        /* How an include spells a header: `Name/Header.h` for a framework's, the path under `usr/include` otherwise. */
        static func includeName(of path: String) -> String? {
            if let range = path.range(of: ".framework/Headers/", options: .backwards) {
                let framework = path[..<range.lowerBound].split(separator: "/").last.map(String.init) ?? ""
                return framework.isEmpty ? nil : "\(framework)/\(path[range.upperBound...])"
            }
            if let range = path.range(of: "/usr/include/", options: .backwards) {
                return String(path[range.upperBound...])
            }
            return nil
        }

        /*
         The modules `module` re-exports whole: those a header it re-exports from imports through the other's
         umbrella header or `@import Name;`, and those its map names in `export Name` that its umbrella header
         imports whole.
         */
        func wholeImports(of module: String, named: Set<String>) -> Set<String> {
            var whole: Set<String> = []
            for header in headersOfModule[module] ?? [] {
                let imports = parse(header)
                var found = Set(imports.modules)
                found.formUnion(imports.headers.compactMap { wholeModuleOfHeader[$0] })
                found.remove(module)
                if reexports(header) {
                    whole.formUnion(found)
                }
                else if header == umbrellaOfModule[module] {
                    whole.formUnion(found.intersection(named))
                }
            }
            return whole
        }

        func visible(through module: String) -> Set<String> {
            if let cached = visibleThrough[module] {
                return cached
            }
            var visible = headersOfModule[module] ?? []
            var pending = Array(visible)
            while let header = pending.popLast() {
                guard reexports(header) else { continue }
                for included in parse(header).headers where visible.insert(included).inserted {
                    pending.append(included)
                }
            }
            visibleThrough[module] = visible
            return visible
        }

        /*
         Whether the submodule the header forms re-exports what the header imports: its own block's `export *` when
         the map declares it, the top-level `export *` for the umbrella, and `module * { export * }` for a header
         the umbrella brings in. Every module the header belongs to must say so.
         */
        private func reexports(_ header: String) -> Bool {
            guard let modules = modulesOfHeader[header], !modules.isEmpty else { return false }
            return modules.allSatisfy { module in
                if umbrellaOfModule[module] == header {
                    return topReexporting.contains(module)
                }
                if let declared = declaredReexports[header] {
                    return declared
                }
                return everySubmoduleReexporting.contains(module)
            }
        }

        /*
         The headers a header includes or imports, found among the build's, and the modules it `@import`s whole.
         Only a line the build certainly compiled counts: every `#if` around it must hold for Swift on macOS
         (`Conditions`), so CoreVideo's `#include <ApplicationServices/ApplicationServices.h>`, which sits in
         the branch `defined(__swift__)` leaves out, brings nothing.
         */
        private func parse(_ header: String) -> (headers: [String], modules: [String]) {
            if let cached = parsed[header] {
                return cached
            }
            /* Set first, so a textual header that includes itself back adds nothing more. */
            parsed[header] = ([], [])
            var included: [String] = []
            var modules: [String] = []
            if let text = try? String(contentsOfFile: header, encoding: .utf8) {
                var conditions = Conditions(exists: { [pathOfName] name in pathOfName[name] != nil })
                for line in Conditions.logicalLines(of: text) {
                    if line.hasPrefix("@import ") {
                        guard conditions.compiled else { continue }
                        let name = line.dropFirst("@import ".count).prefix { $0.isLetter || $0.isNumber || $0 == "_" }
                        if !name.isEmpty,
                            line.dropFirst("@import ".count + name.count).trimmingCharacters(in: .whitespaces)
                                .hasPrefix(";")
                        {
                            modules.append(String(name))
                        }
                        continue
                    }
                    guard line.hasPrefix("#") else { continue }
                    let directive = line.dropFirst().trimmingCharacters(in: .whitespaces)
                    if conditions.read(directive) {
                        continue
                    }
                    guard conditions.compiled, directive.hasPrefix("import") || directive.hasPrefix("include") else {
                        continue
                    }
                    if let open = directive.firstIndex(of: "<"), let close = directive[open...].firstIndex(of: ">") {
                        if let path = pathOfName[String(directive[directive.index(after: open)..<close])] {
                            included.append(path)
                        }
                    }
                    else if let open = directive.firstIndex(of: "\""),
                        let close = directive[directive.index(after: open)...].firstIndex(of: "\"")
                    {
                        let name = String(directive[directive.index(after: open)..<close])
                        let sibling = URL(fileURLWithPath: header).deletingLastPathComponent().appendingPathComponent(
                            name
                        ).standardizedFileURL.path
                        if modulesOfHeader[sibling] != nil || textualHeaders.contains(sibling) {
                            included.append(sibling)
                        }
                    }
                }
            }
            /* A textual header is pasted where it is included, so what it includes, the includer includes. */
            var headers: [String] = []
            for path in included {
                if textualHeaders.contains(path) {
                    let pasted = parse(path)
                    headers.append(contentsOf: pasted.headers)
                    modules.append(contentsOf: pasted.modules)
                }
                else {
                    headers.append(path)
                }
            }
            parsed[header] = (headers, modules)
            return (headers, modules)
        }
    }

    /*
     One header's `#if` stack, read for Swift on macOS, to say which lines the build certainly compiled. A
     condition is true, false or not known, and a line counts only when every condition around it is known true.
     Known: macOS's target macros, the macros Swift's Clang importer defines or leaves out, `__has_include` of a
     header the build has, and an include guard. Anything else is not known, which can only leave a line out.
     */
    struct Conditions {
        static let values: [String: Int] = [
            "TARGET_OS_OSX": 1, "TARGET_OS_MAC": 1, "TARGET_OS_IPHONE": 0, "TARGET_OS_IOS": 0, "TARGET_OS_WATCH": 0,
            "TARGET_OS_TV": 0,
            "TARGET_OS_VISION": 0, "TARGET_OS_XR": 0, "TARGET_OS_MACCATALYST": 0, "TARGET_OS_SIMULATOR": 0,
            "TARGET_OS_EMBEDDED": 0,
            "TARGET_OS_DRIVERKIT": 0, "TARGET_OS_BRIDGE": 0, "TARGET_OS_WIN32": 0, "TARGET_OS_WINDOWS": 0,
            "TARGET_OS_LINUX": 0,
            "__METAL_VERSION__": 0,
        ]
        static let defined: [String: Bool] = values.mapValues { _ in true }.merging(
            [
                "__swift__": true, "__OBJC__": true, "__OBJC2__": true, "__APPLE__": true, "__MACH__": true,
                "__cplusplus": false,
                "__METAL_VERSION__": false, "CF_EXCLUDE_CSTD_HEADERS": false,
            ],
            uniquingKeysWith: { _, new in new },
        )

        /* Each open `#if`: whether its live branch holds, and whether an earlier branch did. */
        private var stack: [(live: Bool?, taken: Bool?)] = []
        /* The macro an `#ifndef` just tested, which a `#define` of it right after makes an include guard. */
        private var guardMacro: String?
        private let exists: (String) -> Bool

        init(exists: @escaping (String) -> Bool) {
            self.exists = exists
        }

        var compiled: Bool { stack.allSatisfy { $0.live == true } }

        /* Reads a conditional directive (or an include guard's `#define`) and says whether it was one. */
        mutating func read(_ directive: String) -> Bool {
            let word = directive.prefix { $0.isLetter }
            let rest = directive.dropFirst(word.count).trimmingCharacters(in: .whitespaces)
            let guarded = guardMacro
            guardMacro = nil
            switch word {
                case "ifdef":
                    stack.append((Self.defined[rest], Self.defined[rest]))
                case "ifndef":
                    let value = Self.defined[rest].map { !$0 }
                    stack.append((value, value))
                    guardMacro = rest
                case "if":
                    let value = evaluate(rest)
                    stack.append((value, value))
                    if rest.hasPrefix("!defined") {
                        guardMacro = rest.dropFirst("!defined".count).trimmingCharacters(
                            in: CharacterSet(charactersIn: " ()")
                        )
                    }
                case "elif":
                    guard let top = stack.popLast() else { return true }
                    let value = evaluate(rest)
                    let live: Bool? =
                        top.taken == true ? false : (top.taken == false ? value : (value == false ? false : nil))
                    stack.append((live, Self.or(top.taken, value)))
                case "else":
                    guard let top = stack.popLast() else { return true }
                    stack.append((top.taken.map { !$0 }, true))
                case "endif":
                    _ = stack.popLast()
                case "define":
                    /* `#ifndef X` then `#define X`: an include guard, true the first time the header is read. */
                    if let guarded, rest.prefix(while: { !$0.isWhitespace }) == guarded, !stack.isEmpty {
                        stack[stack.count - 1] = (true, true)
                    }
                default:
                    return false
            }
            return true
        }

        static func or(_ first: Bool?, _ second: Bool?) -> Bool? {
            if first == true || second == true {
                return true
            }
            return first == false && second == false ? false : nil
        }

        static func and(_ first: Bool?, _ second: Bool?) -> Bool? {
            if first == false || second == false {
                return false
            }
            return first == true && second == true ? true : nil
        }

        /* `!`, `&&`, `||`, parentheses, `defined`, `__has_include`, known macros and integers; anything else is not known. */
        func evaluate(_ expression: String) -> Bool? {
            var tokens: [String] = []
            var index = expression.startIndex
            while index < expression.endIndex {
                let character = expression[index]
                if character.isWhitespace {
                    index = expression.index(after: index)
                }
                else if character.isLetter || character == "_" || character.isNumber {
                    let word = expression[index...].prefix { $0.isLetter || $0.isNumber || $0 == "_" }
                    index = expression.index(index, offsetBy: word.count)
                    if word == "__has_include" {
                        guard let open = expression[index...].firstIndex(where: { $0 == "<" || $0 == "\"" }),
                            let close = expression[expression.index(after: open)...].firstIndex(where: {
                                $0 == ">" || $0 == "\""
                            }),
                            let end = expression[close...].firstIndex(of: ")")
                        else { return nil }
                        tokens.append(exists(String(expression[expression.index(after: open)..<close])) ? "1" : "?")
                        index = expression.index(after: end)
                    }
                    else {
                        tokens.append(String(word))
                    }
                }
                else if expression[index...].hasPrefix("&&") || expression[index...].hasPrefix("||") {
                    tokens.append(String(expression[index...].prefix(2)))
                    index = expression.index(index, offsetBy: 2)
                }
                else if "!()".contains(character), !expression[index...].hasPrefix("!=") {
                    tokens.append(String(character))
                    index = expression.index(after: index)
                }
                else {
                    return nil
                }
            }
            var position = 0
            func primary() -> Bool?? {
                guard position < tokens.count else { return .none }
                let token = tokens[position]
                position += 1
                switch token {
                    case "!":
                        guard let value = primary() else { return .none }
                        return .some(value.map { !$0 })
                    case "(":
                        guard let value = disjunction(), position < tokens.count, tokens[position] == ")" else {
                            return .none
                        }
                        position += 1
                        return .some(value)
                    case "defined":
                        let parenthesized = position < tokens.count && tokens[position] == "("
                        position += parenthesized ? 1 : 0
                        guard position < tokens.count else { return .none }
                        let name = tokens[position]
                        position += 1
                        if parenthesized {
                            guard position < tokens.count, tokens[position] == ")" else { return .none }
                            position += 1
                        }
                        return .some(Self.defined[name])
                    case "?":
                        return .some(nil)
                    default:
                        if let first = token.first, first.isNumber {
                            let digits = token.prefix { $0.isNumber }
                            return .some(Int(digits).map { $0 != 0 })
                        }
                        return .some(Self.values[token].map { $0 != 0 })
                }
            }
            func conjunction() -> Bool?? {
                guard let first = primary() else { return .none }
                var value: Bool? = first
                while position < tokens.count, tokens[position] == "&&" {
                    position += 1
                    guard let next = primary() else { return .none }
                    value = Self.and(value, next)
                }
                return .some(value)
            }
            func disjunction() -> Bool?? {
                guard let first = conjunction() else { return .none }
                var value: Bool? = first
                while position < tokens.count, tokens[position] == "||" {
                    position += 1
                    guard let next = conjunction() else { return .none }
                    value = Self.or(value, next)
                }
                return .some(value)
            }
            guard let value = disjunction(), position == tokens.count else { return nil }
            return value
        }

        /* The header's lines with comments dropped and backslash continuations joined, trimmed. */
        static func logicalLines(of text: String) -> [String] {
            var lines: [String] = []
            var pending = ""
            var inComment = false
            for rawLine in text.split(separator: "\n", omittingEmptySubsequences: false) {
                var line = ""
                var index = rawLine.startIndex
                while index < rawLine.endIndex {
                    if inComment {
                        if rawLine[index...].hasPrefix("*/") {
                            inComment = false
                            index = rawLine.index(index, offsetBy: 2)
                        }
                        else {
                            index = rawLine.index(after: index)
                        }
                        continue
                    }
                    if rawLine[index...].hasPrefix("/*") {
                        inComment = true
                        index = rawLine.index(index, offsetBy: 2)
                        continue
                    }
                    if rawLine[index...].hasPrefix("//") {
                        break
                    }
                    line.append(rawLine[index])
                    index = rawLine.index(after: index)
                }
                if line.hasSuffix("\\") {
                    pending += line.dropLast() + " "
                    continue
                }
                lines.append((pending + line).trimmingCharacters(in: .whitespaces))
                pending = ""
            }
            return lines
        }
    }

    struct InterfaceExports {
        /* Modules re-exported whole: what `known` may believe. */
        var whole: Set<String> = []
        /* Every module any `@_exported` line touches, in part or whole: what `bounded` must allow. */
        var atMost: Set<String> = []
    }

    /*
     What a `.swiftinterface` re-exports, or nil when it cannot be read. Only a whole module is `whole`: a
     submodule (`@_exported import Darwin.C`) or one declaration (`@_exported import struct Foundation.URL`)
     re-exports less than the module's name, and `known` never holds more than was read. Both still count
     toward `atMost`, which must never hold less.
     */
    static func exportedImports(ofInterface path: String) -> InterfaceExports? {
        guard let text = try? String(contentsOfFile: path, encoding: .utf8) else { return nil }
        var exports = InterfaceExports()
        for line in text.split(separator: "\n", omittingEmptySubsequences: true) {
            let words = line.split(whereSeparator: \.isWhitespace)
            guard words.first?.hasPrefix("@") == true, let importIndex = words.firstIndex(of: "import"),
                importIndex + 1 < words.count
            else { continue }
            guard words[..<importIndex].contains("@_exported") else { continue }
            let path = words[words.count - 1]
            guard let module = path.split(separator: ".").first else { continue }
            exports.atMost.insert(String(module))
            if words[..<importIndex].allSatisfy({ $0.hasPrefix("@") }), importIndex + 2 == words.count,
                !path.contains(".")
            {
                exports.whole.insert(String(path))
            }
        }
        return exports
    }

    /*
     Where a Clang module's map may be, most likely first: a framework's `Modules/module.modulemap`, found from
     the headers the unit read, then every module map the compile consulted (`usr/include`'s are named for the
     modules they declare, `Darwin.modulemap`).
     */
    static func moduleMapCandidates(of unit: IndexStore.Unit) -> [String] {
        var candidates: [String] = []
        let framework = "/\(unit.module).framework/"
        if let header = unit.files.first(where: { $0.contains(framework) })
            ?? unit.recordFiles.values.first(where: { $0.contains(framework) }),
            let range = header.range(of: framework)
        {
            candidates.append(String(header[..<range.upperBound]) + "Modules/module.modulemap")
        }
        let maps = unit.files.filter { $0.hasSuffix(".modulemap") }
        candidates.append(
            contentsOf: maps.filter { URL(fileURLWithPath: $0).lastPathComponent == "\(unit.module).modulemap" }
        )
        candidates.append(
            contentsOf: maps.filter { URL(fileURLWithPath: $0).lastPathComponent != "\(unit.module).modulemap" }
        )
        var seen: Set<String> = []
        return candidates.filter { seen.insert($0).inserted }
    }

    /*
     A Clang module map, read for one thing: what each top-level module exports. A module's own `export`
     declarations are the ones directly inside its braces; a submodule's (`module * { export * }`) are not the
     top-level module's, and are left out, because `known` never holds more than was read.
     */
    struct ModuleMap {
        struct Exports {
            /* `export *`: every module the headers import. */
            var all = false
            /* `module * { export * }`: each header's own submodule re-exports what that header imports. */
            var everySubmodule = false
            var named: Set<String> = []
            /* The header that imports the whole module, as the map names it. */
            var umbrella: String?
            /* Headers the top-level module declares itself: importing one imports the module, as the umbrella does. */
            var topHeaders: [String] = []
            /* Every header a module or submodule declares, with whether its block says `export *`. */
            var headerExports: [String: Bool] = [:]
            /* Headers of `explicit` submodules, which an import of the module leaves out. */
            var explicitHeaders: Set<String> = []
            /* `textual header`s: pasted into whatever includes them, so what they include is the includer's. */
            var textualHeaders: Set<String> = []
        }

        private var modules: [String: Exports] = [:]
        /* `extern module Name "file"`: the map that declares Name, relative to this one. */
        private(set) var externs: [String: String] = [:]

        init(text: String) {
            let tokens = Self.tokens(of: text)
            /* One open module block: the top-level module it belongs to, the headers it declares, and its `export *`. */
            final class Block {
                let top: String
                let isTop: Bool
                let isWildcard: Bool
                let isExplicit: Bool
                var headers: [String] = []
                var exportsAll = false
                init(top: String, isTop: Bool, isWildcard: Bool, isExplicit: Bool) {
                    self.top = top
                    self.isTop = isTop
                    self.isWildcard = isWildcard
                    self.isExplicit = isExplicit
                }
            }
            /* Braces that open no module block hold nil. */
            var stack: [Block?] = []
            var pending: (name: String, isWildcard: Bool, isExplicit: Bool)?
            var index = 0
            while index < tokens.count {
                let token = tokens[index]
                switch token {
                    case "module" where index + 1 < tokens.count:
                        /* `[framework] [explicit] module Name [attributes] {`, or `extern module Name "path"`, which declares nothing here. */
                        if index > 0, tokens[index - 1] == "extern" {
                            if index + 2 < tokens.count, tokens[index + 2].hasPrefix("\"") {
                                externs[tokens[index + 1]] = String(tokens[index + 2].dropFirst().dropLast())
                                index += 3
                                continue
                            }
                            index += 2
                            continue
                        }
                        let isExplicit = tokens[max(0, index - 2)..<index].contains("explicit")
                        pending = (tokens[index + 1], tokens[index + 1] == "*", isExplicit)
                        index += 2
                        continue
                    case "{":
                        if let opened = pending {
                            let top = stack.compactMap { $0 }.first?.top ?? opened.name
                            let isExplicit = opened.isExplicit || stack.contains { $0?.isExplicit == true }
                            stack.append(
                                Block(
                                    top: top,
                                    isTop: stack.isEmpty,
                                    isWildcard: opened.isWildcard,
                                    isExplicit: isExplicit,
                                )
                            )
                            if stack.count == 1, modules[top] == nil {
                                modules[top] = Exports()
                            }
                            pending = nil
                        }
                        else {
                            stack.append(nil)
                        }
                    case "}":
                        guard let closed = stack.popLast(), let block = closed else { break }
                        if block.isTop {
                            modules[block.top]?.all = block.exportsAll
                            modules[block.top]?.topHeaders.append(contentsOf: block.headers)
                        }
                        else if block.isWildcard, block.exportsAll {
                            modules[block.top]?.everySubmodule = true
                        }
                        for header in block.headers {
                            modules[block.top]?.headerExports[header] = block.exportsAll
                            if block.isExplicit {
                                modules[block.top]?.explicitHeaders.insert(header)
                            }
                        }
                    case "export" where index + 1 < tokens.count:
                        guard let open = stack.last, let block = open else { break }
                        let exported = tokens[index + 1]
                        if exported == "*" {
                            block.exportsAll = true
                        }
                        else if block.isTop, !exported.contains(".") {
                            modules[block.top]?.named.insert(exported)
                        }
                        index += 2
                        continue
                    case "umbrella" where index + 2 < tokens.count:
                        /* `umbrella header "Foundation.h"`; an umbrella directory names no header and is left out. */
                        guard let open = stack.last, let block = open, tokens[index + 1] == "header",
                            tokens[index + 2].hasPrefix("\"")
                        else { break }
                        let header = String(tokens[index + 2].dropFirst().dropLast())
                        if block.isTop {
                            modules[block.top]?.umbrella = header
                        }
                        else {
                            block.headers.append(header)
                        }
                        index += 3
                        continue
                    case "header" where index + 1 < tokens.count:
                        /* A member header. `exclude` and `private` ones are not the module's to re-export; `textual` ones belong to no module. */
                        guard let open = stack.last, let block = open, tokens[index + 1].hasPrefix("\"") else { break }
                        let header = String(tokens[index + 1].dropFirst().dropLast())
                        let qualifier = index > 0 ? tokens[index - 1] : ""
                        if qualifier == "textual" {
                            modules[block.top]?.textualHeaders.insert(header)
                        }
                        else if !["exclude", "private", "umbrella"].contains(qualifier) {
                            block.headers.append(header)
                        }
                        index += 2
                        continue
                    default:
                        break
                }
                index += 1
            }
        }

        func exports(ofModule name: String) -> Exports? {
            modules[name]
        }

        /* Words, `{`, `}`, `[`, `]`, `*` and quoted strings, with `//` and `/* */` comments dropped. */
        static func tokens(of text: String) -> [String] {
            var tokens: [String] = []
            var word = ""
            let characters = Array(text)
            var index = 0
            func flush() {
                if !word.isEmpty {
                    tokens.append(word)
                    word = ""
                }
            }
            while index < characters.count {
                let character = characters[index]
                let next: Character? = index + 1 < characters.count ? characters[index + 1] : nil
                if character == "/", next == "/" {
                    flush()
                    while index < characters.count, characters[index] != "\n" {
                        index += 1
                    }
                    continue
                }
                if character == "/", next == "*" {
                    flush()
                    index += 2
                    while index + 1 < characters.count, !(characters[index] == "*" && characters[index + 1] == "/") {
                        index += 1
                    }
                    index += 2
                    continue
                }
                if character == "\"" {
                    flush()
                    var literal = "\""
                    index += 1
                    while index < characters.count, characters[index] != "\"" {
                        literal.append(characters[index])
                        index += 1
                    }
                    tokens.append(literal + "\"")
                    index += 1
                    continue
                }
                if "{}[]*,".contains(character) {
                    flush()
                    if character != "," {
                        tokens.append(String(character))
                    }
                }
                else if character.isWhitespace {
                    flush()
                }
                else {
                    word.append(character)
                }
                index += 1
            }
            flush()
            return tokens
        }
    }
}
