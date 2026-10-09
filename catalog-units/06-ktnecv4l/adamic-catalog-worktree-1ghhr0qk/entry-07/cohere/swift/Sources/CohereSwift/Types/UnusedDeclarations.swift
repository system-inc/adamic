import Foundation
import SwiftSyntax

/*
 `unused-declaration`: a declaration of ours that nothing refers to, read from the build's index store.

 A declaration `private` or `fileprivate`, or inside a type or extension that is, can be named only in its own
 file, so the file's own index record holds every reference it will ever have. Any other declaration is judged
 against every record of the package, its tests included, and one public in a library product (or in a target
 a product re-exports) is API that someone this build never sees may use. A declaration is used when a record
 holds a reference to it from outside its own text (a function calling itself, or a type naming itself in its
 body, is not a use), or when another declaration overrides or witnesses it, which removing it would break. A
 property wrapped by an attribute is also used when its `$name` or `_name` is, since `$flag` passed as a binding
 names the projection and never the property. A stored property is used when a call of its struct's
 synthesized memberwise initializer passes it: on ahraos-macos the index records `KingdomSidebar(...)` as a
 call of that initializer and records no reference to the `livePaneProfileIds:` it passes.

 The package's records do not hold everything: an inactive `#if` branch was never compiled, and a file the
 build has not compiled as it stands has no record that describes it. A declaration whose name appears in either
 is never reported, whatever the records say.

 What the index cannot see is treated as used, and counted by reason rather than reported, after Periphery's
 list: protocol witnesses and overrides; anything the Objective-C runtime reaches (`@objc`, `@IBAction`,
 `@IBOutlet`, `@NSManaged`, `dynamic`); entry points and previews; members the compiler calls by name
 (a property wrapper's `wrappedValue`, a result builder's `build` methods, `callAsFunction`, dynamic member
 lookup); a declaration carrying an attribute that may register it, a macro such as `@Test`. Some are left
 alone because removing them changes what the program does without breaking the build: an initializer, which
 may exist to keep a type from being built; a stored property of a type whose conformances reach `Encodable` or
 `Decodable` (`ConformanceGraph`, through protocols of ours and extensions anywhere in the module), or reach what
 the index cannot see into, a C or Objective-C type or a protocol with no record; the coding keys such a type's
 synthesized conformance reads by name, `CodingKeys` and, for an enum case with associated values, the case's
 own `<Case>CodingKeys`, which name its labels on the wire (macos's `FetchHistoryCodingKeys` keeps
 `maximumBytes` as `maxBytes`: removed, the wire key changed and the build stayed green); a stored property of a
 type whose conformances reach `Equatable` (`Hashable` and `Comparable` refine it), which synthesized `==` and
 `hash(into:)` compare, so removing one nothing names makes two values that differed by it equal; a stored
 property of a type a macro may expand; an instance's stored property whose initializer runs
 code, which may be held for what it does or keeps alive (a subscription, an observer token); a field of a
 struct made only of numbers and SIMD vectors, whose bytes a shader or C may read whole, so an unread field is
 padding that holds the layout (found on Presence, where LockerBloom's `unused: Int32` pads the constants its
 Metal shader reads); a case of an enum that may be reached through its conformances or raw values
 (`CaseIterable`, `init(rawValue:)`).

 Reading every finding on both proving grounds added three more the build would never catch. A SwiftUI view's
 stored properties are compared by SwiftUI to decide when to draw again, so one nothing reads by name still
 does work (KingdomSidebar's `livePaneProfileIds` exists only for that). Swift Testing runs a type's `@Test`
 functions without a `@Suite`, so a type holding a registered declaration is registered too. And a stored
 property of a type that descends from another package's (MLX's `Module`) may be read by reflection. A class
 that descends from an Objective-C class can be found by name, from a nib, a plist or `NSClassFromString`, and an
 XCTest method is run by name, so neither is reported.

 A file is left unchecked, and counted, when the index cannot vouch for it: no unit newer than the file, or any
 `#if` in it, since an inactive branch may be the only user of a declaration. Only the outermost declaration
 is reported when an unused one holds others, so every declaration reported can go together.
 */
struct UnusedDeclarations {
    static let ruleName = "cohere-swift/correctness-no-unused-declaration"

    struct Result {
        /* Each unused declaration, with its keyword and name for the report's line. */
        var findings: [(finding: FindingRecord, subject: String)] = []
        var filesChecked = 0
        /* Files the index could not vouch for, by path, with why. */
        var filesUnchecked: [String: String] = [:]
        /* Declarations judged, and declarations never reported, by why. */
        var declarationsChecked = 0
        var skipped: [String: Int] = [:]
    }

    static let overrides = "an override or a protocol witness, reached through what it overrides"
    static let objectiveC = "reached by the Objective-C runtime (@objc, @IBAction, @IBOutlet, @NSManaged, dynamic)"
    static let entryPoint = "an entry point or a preview (@main, PreviewProvider), reached from outside the program"
    static let registered = "an attribute that may register it, a macro such as @Test"
    static let compilerCalled =
        "called by the compiler by name: a property wrapper's or result builder's members, callAsFunction, dynamic member lookup"
    static let codingKeys =
        "coding keys (CodingKeys, or a case's <Case>CodingKeys) of a type whose conformances reach Encodable or Decodable, read by the synthesized conformance"
    static let initializer = "an initializer, which may exist to keep its type from being built another way"
    static let reflectedStorage =
        "a stored property of a type whose conformances reach Encodable or Decodable, which read every stored property"
    static let synthesizedEquality =
        "a stored property of a type whose conformances reach Equatable or Hashable, which synthesized == and hash(into:) compare"
    static let unseenConformance =
        "a stored property of a type that conforms to or inherits from what the index cannot see into (a C or Objective-C type, a type alias, a protocol with no record)"
    static let lifetime =
        "an instance's stored property whose initializer runs code, which may be kept for what it does or keeps alive"
    static let layout =
        "a field of a struct of plain numbers, whose bytes may be read whole (a shader's constants, a C struct), padding included"
    static let reachableCase =
        "a case of an enum with raw values or conformances that may reach it (CaseIterable, Codable, init(rawValue:))"
    static let sharedDeclaration =
        "declared together with others (`let a = 1, b = 2`, `case a, b`), which the report cannot remove apart"
    static let unplaced = "the index records no definition at its name"
    static let publicAPI = "public API of a library product, used by whoever depends on the package"
    static let hiddenName =
        "its name appears where the index cannot see (inside #if, or in a file the build has not compiled as it stands)"
    static let objectiveCClass = "a class the Objective-C runtime can find by name (a nib, a plist, NSClassFromString)"
    static let testMethod = "an XCTest method, which XCTest finds and runs by name"
    static let viewStorage =
        "a stored property of a SwiftUI view or one of its kin, which SwiftUI compares to decide when to draw again"
    static let foreignStorage =
        "a stored property of a type that inherits from or conforms to another package's type, which may read it by reflection (MLX's Module does)"

    let stores: [IndexStore]
    /* The library products' targets, as `packageRoot|target`: their public declarations are API. */
    let apiTargets: Set<String>
    /* The modules this package owns, vendored packages left out. */
    let ownedModules: Set<String>

    init(stores: [IndexStore], package: PackageModel? = nil) {
        self.stores = stores
        self.apiTargets = Set(
            (package?.allPackages ?? []).flatMap { member in
                member.libraryProductTargets.map { Self.targetKey(root: member.root, target: $0) }
            }
        )
        self.ownedModules = package.map(Pipeline.ownedModules) ?? []
    }

    /* A target's name as the compiler spells its module: anything that is not an identifier character made `_`. */
    static func moduleName(of target: String) -> String {
        String(target.map { $0.isLetter || $0.isNumber || $0 == "_" ? $0 : "_" })
    }

    static func targetKey(root: URL?, target: String) -> String {
        "\(root?.standardizedFileURL.path ?? "")|\(target)"
    }

    /* Every reference in the package's records, by what it refers to and by name, and every declaration something overrides. */
    struct References {
        var bySymbol: [String: [(file: String, place: Place)]] = [:]
        var byName: [String: [(file: String, place: Place)]] = [:]
        var overridden: Set<String> = []
        /* Each struct's synthesized initializers, by the struct: their symbols and their labels (`init(kept:neverRead:)`). */
        var synthesizedInitializers: [String: [(symbol: String, labels: [String])]] = [:]

        mutating func add(_ occurrences: [IndexStore.RecordOccurrence], file: String) {
            for occurrence in occurrences {
                let place = Place(line: occurrence.line, column: occurrence.column)
                if occurrence.isDeclaration, occurrence.isImplicit, occurrence.kind == IndexStore.constructorKind,
                    let holder = occurrence.relations.first(where: { $0.roles & IndexStore.childOfRole != 0 })?.symbol
                {
                    let labels = occurrence.name.dropFirst("init(".count).dropLast().split(separator: ":").map(
                        String.init
                    )
                    synthesizedInitializers[holder, default: []].append((occurrence.symbol, labels))
                }
                if occurrence.isReference {
                    bySymbol[occurrence.symbol, default: []].append((file, place))
                    /* Reading a property records its getter too; either names the property. */
                    if let base = UnusedImports.ModuleResolver.accessorBase(occurrence.symbol) {
                        bySymbol[base, default: []].append((file, place))
                    }
                    byName[occurrence.name, default: []].append((file, place))
                }
                overridden.formUnion(occurrence.overridden)
            }
        }
    }

    func run(files: [ParsedFile]) -> Result {
        var result = Result()
        /*
         A library product's API is also every target its targets re-export with `@_exported import`: VRMKit's
         VRMRealityKit product hands on VRMKitRuntime whole, so what VRMKitRuntime declares public is the product's.
         */
        var targetsByModule: [String: String] = [:]
        for file in files {
            targetsByModule[Self.targetKey(root: file.packageRoot, target: Self.moduleName(of: file.targetName))] =
                file.targetName
        }
        var apiTargets = self.apiTargets
        var grew = true
        while grew {
            grew = false
            for file in files
            where apiTargets.contains(Self.targetKey(root: file.packageRoot, target: file.targetName)) {
                for declaration in UnusedImports.imports(in: file.tree) where declaration.isExported {
                    guard
                        let target = targetsByModule[
                            Self.targetKey(root: file.packageRoot, target: declaration.module)
                        ]
                    else { continue }
                    grew = apiTargets.insert(Self.targetKey(root: file.packageRoot, target: target)).inserted || grew
                }
            }
        }
        let fresh = IndexStore.freshUnits(of: files, in: stores)
        /* Every record of ours, read once with its relations: the files judged, and every extension a conformance may come from. */
        var graph = ConformanceGraph(stores: stores)
        var occurrencesByFile: [String: [IndexStore.RecordOccurrence]] = [:]
        var references = References()
        for (path, described) in IndexStore.newestUnits(in: stores) {
            let occurrences = described.unit.ownRecords.flatMap {
                described.store.occurrences(inRecord: $0, relations: true) ?? []
            }
            occurrencesByFile[path] = occurrences
            graph.add(occurrences)
            references.add(occurrences, file: path)
        }
        /*
         Names used where the index cannot look: every name inside `#if`, whose inactive branches it never saw, and
         every name in a file the build has not compiled as it stands. A declaration the package can name is never
         reported under one of them.
         */
        var hiddenNames: Set<String> = []
        for file in files {
            if fresh[file.url.path] == nil {
                hiddenNames.formUnion(Self.names(in: Syntax(file.tree)))
            }
            else if UnusedImports.hasConditionalCompilation(file.tree) {
                for conditional in Self.conditionals(in: file.tree) {
                    hiddenNames.formUnion(Self.names(in: Syntax(conditional)))
                }
            }
        }
        for file in files {
            guard let described = fresh[file.url.path] else {
                result.filesUnchecked[file.url.path] = UnusedImports.notCompiled
                continue
            }
            if UnusedImports.hasConditionalCompilation(file.tree) {
                result.filesUnchecked[file.url.path] = UnusedImports.conditional
                continue
            }
            let occurrences = occurrencesByFile[described.unit.mainFile] ?? []
            result.filesChecked += 1
            judge(
                file,
                path: described.unit.mainFile,
                occurrences: occurrences,
                references: references,
                hiddenNames: hiddenNames,
                apiTargets: apiTargets,
                graph: &graph,
                into: &result,
            )
        }
        return result
    }

    /*
     One file's declarations against the package's records. `path` is the file as the index names it, which is
     how a reference is told to be in this file.
     */
    private func judge(
        _ file: ParsedFile,
        path: String,
        occurrences: [IndexStore.RecordOccurrence],
        references: References,
        hiddenNames: Set<String>,
        apiTargets: Set<String>,
        graph: inout ConformanceGraph,
        into result: inout Result,
    ) {
        var definitions: [Place: [IndexStore.RecordOccurrence]] = [:]
        for occurrence in occurrences where occurrence.isDeclaration && !occurrence.isImplicit {
            definitions[Place(line: occurrence.line, column: occurrence.column), default: []].append(occurrence)
        }
        let isAPITarget = apiTargets.contains(Self.targetKey(root: file.packageRoot, target: file.targetName))

        var collector = Collector(tree: file.tree)
        collector.collect()
        var unused: [(candidate: Candidate, symbol: IndexStore.RecordOccurrence, range: ClosedRange<Place>)] = []
        for candidate in collector.candidates {
            if let exemption = candidate.exemption {
                result.skipped[exemption, default: 0] += 1
                continue
            }
            let location = file.locations.location(for: candidate.name.positionAfterSkippingLeadingTrivia)
            let placed = definitions[Place(line: location.line, column: location.column)] ?? []
            guard let definition = placed.first else {
                result.skipped[Self.unplaced, default: 0] += 1
                continue
            }
            if candidate.reach == .api && isAPITarget {
                result.skipped[Self.publicAPI, default: 0] += 1
                continue
            }
            if placed.contains(where: { $0.roles & IndexStore.overrideOfRole != 0 }) || candidate.isOverride {
                result.skipped[Self.overrides, default: 0] += 1
                continue
            }
            /* The type holding it, by the index's child-of relation, and what that type descends from. */
            let holder = definition.relations.first { $0.roles & IndexStore.childOfRole != 0 }?.symbol
            let holderAncestors = holder.map { graph.ancestors(of: $0) } ?? []
            /* Coding keys are read by the synthesized conformance, so they are used wherever one may be synthesized: only a holder that clearly is not Codable leaves them to the references. */
            if candidate.isCodingKeys, (holder.map { graph.reach(of: $0) } ?? .unseen) != .clear {
                result.skipped[Self.codingKeys, default: 0] += 1
                continue
            }
            if candidate.isType, graph.ancestors(of: definition.symbol).contains(where: { $0.hasPrefix("c:objc(cs)") })
            {
                result.skipped[Self.objectiveCClass, default: 0] += 1
                continue
            }
            if candidate.keyword == "func", candidate.name.text.hasPrefix("test"),
                holderAncestors.contains("c:objc(cs)XCTestCase")
            {
                result.skipped[Self.testMethod, default: 0] += 1
                continue
            }
            if candidate.isInstanceStorage {
                /* What the holding type's conformances reach decides whether a conformance reads it. */
                switch holder.map({ graph.reach(of: $0) }) ?? .unseen {
                    case .reached:
                        result.skipped[Self.reflectedStorage, default: 0] += 1
                        continue
                    case .unseen:
                        result.skipped[Self.unseenConformance, default: 0] += 1
                        continue
                    case .clear:
                        if let holder, graph.reach(of: holder, toward: ConformanceGraph.equatable) == .reached {
                            result.skipped[Self.synthesizedEquality, default: 0] += 1
                            continue
                        }
                        if holderAncestors.contains(where: {
                            ["SwiftUI", "SwiftUICore"].contains(ConformanceGraph.swiftModule(of: $0) ?? "")
                        }) {
                            result.skipped[Self.viewStorage, default: 0] += 1
                            continue
                        }
                        let foreign = holderAncestors.contains { ancestor in
                            ConformanceGraph.swiftModule(of: ancestor).map {
                                graph.compiledModules.contains($0) && !ownedModules.contains($0)
                            } ?? false
                        }
                        if foreign {
                            result.skipped[Self.foreignStorage, default: 0] += 1
                            continue
                        }
                        if let exemption = candidate.storageExemption {
                            result.skipped[exemption, default: 0] += 1
                            continue
                        }
                }
            }
            let name = candidate.name.text.trimmingCharacters(in: CharacterSet(charactersIn: "`"))
            if candidate.reach != .file && hiddenNames.contains(name) {
                result.skipped[Self.hiddenName, default: 0] += 1
                continue
            }
            result.declarationsChecked += 1
            let start = file.locations.location(for: candidate.node.positionAfterSkippingLeadingTrivia)
            let end = file.locations.location(for: candidate.node.endPositionBeforeTrailingTrivia)
            let range = Place(line: start.line, column: start.column)...Place(line: end.line, column: end.column)
            let symbols = Set(placed.map(\.symbol))
            if !symbols.isDisjoint(with: references.overridden) {
                continue
            }
            var places = symbols.flatMap { references.bySymbol[$0] ?? [] }
            if candidate.isInstanceStorage, let holder {
                /*
                 A memberwise initializer names the property by its label, and the index does not always record that as a
                 reference to the property: KingdomSidebar(livePaneProfileIds:) on ahraos-macos has none. A call of the
                 synthesized initializer that takes it is a use.
                 */
                for initializer in references.synthesizedInitializers[holder] ?? []
                where initializer.labels.contains(name) || initializer.labels.contains("_" + name) {
                    places += references.bySymbol[initializer.symbol] ?? []
                }
            }
            if candidate.isWrapped {
                /* A private property's projection can be named only in its file; an internal one's anywhere in the package. */
                let projections = (references.byName["$" + name] ?? []) + (references.byName["_" + name] ?? [])
                places += candidate.reach == .file ? projections.filter { $0.file == path } : projections
            }
            if places.contains(where: { $0.file != path || !range.contains($0.place) }) {
                continue
            }
            unused.append((candidate, definition, range))
        }

        /* The outermost only: a declaration inside another one reported goes with it. */
        let outermost = unused.filter { inner in
            !unused.contains { outer in
                outer.range != inner.range && outer.range.contains(inner.range.lowerBound)
                    && outer.range.contains(inner.range.upperBound)
            }
        }
        for entry in outermost {
            let name =
                entry.candidate.keyword == "func" || entry.candidate.keyword == "subscript"
                ? entry.symbol.name : entry.candidate.name.text
            let subject = "\(entry.candidate.keyword) \(name)"
            let message =
                entry.candidate.reach == .file
                ? "Nothing refers to \(name), and it is private to this file, so nothing outside the file can. Remove it."
                : "Nothing in the package refers to \(name), its tests included. Remove it."
            let finding = file.finding(
                at: entry.candidate.name,
                rule: Self.ruleName,
                messageId: "unusedDeclaration",
                message: message,
                suggestions: [
                    FindingRecord.Suggestion(
                        message: "Remove `\(subject)`",
                        fixes: [Self.removal(of: entry.candidate.node, in: file)],
                    )
                ],
            )
            result.findings.append((finding, subject))
        }
    }

    /* A place in a file as the index gives it: 1-based line, 1-based UTF-8 column. */
    struct Place: Hashable, Comparable {
        var line: Int
        var column: Int

        static func < (first: Place, second: Place) -> Bool {
            (first.line, first.column) < (second.line, second.column)
        }
    }

    /* Who can name a declaration: only its own file, the package, or, public in a library product, anyone who depends on it. */
    enum Reach {
        case file
        case package
        case api
    }

    /* One declaration to judge, or one never reported, with why. */
    struct Candidate {
        /* The whole declaration, attributes and body, which a removal takes. */
        var node: Syntax
        /* The name the index places the declaration at. */
        var name: TokenSyntax
        var keyword: String
        var exemption: String?
        var reach = Reach.file
        var isType = false
        /* A property with an attribute, whose projection (`$name`) and storage (`_name`) are its uses too. */
        var isWrapped = false
        var isOverride = false
        /* `CodingKeys`, or the `<Case>CodingKeys` of one of its enum's cases: used when its type's conformances reach Codable. */
        var isCodingKeys = false
        /* An instance's stored property, which a conformance of its type may read whole. */
        var isInstanceStorage = false
        /* Why the stored property is never reported if no conformance reads it: what its initializer keeps alive, or a layout. */
        var storageExemption: String?
    }

    /*
     Every declaration in the file and what encloses it. Types come first, in a pass of their own, so an extension
     knows whether the type it extends is private to the file and what conformances the type has, wherever in the
     file each is written.
     */
    struct Collector {
        /* What a declaration sits inside: whether that keeps it in this file, and the type it belongs to. */
        struct Scope {
            var fileScoped = false
            /* Whether anything declared here can be public: false inside a type that is not, or an extension marked lower. */
            var canBePublic = true
            /* Members public without saying so: a public protocol's requirements, a `public extension`'s members. */
            var membersPublic = false
            /* A public enum, whose cases are public without saying so. */
            var casesPublic = false
            var typeName: String?
            /* A struct whose stored properties are all plain numbers or SIMD vectors: its bytes may be its meaning. */
            var isLayout = false
            var conformances: Set<String> = []
            var attributes: Set<String> = []
        }

        let tree: SourceFileSyntax
        private(set) var candidates: [Candidate] = []
        /* Types declared in the file by qualified name (`Outer.Inner`), and whether each is private to it. */
        private var fileScopedTypes: [String: Bool] = [:]
        /* What each type conforms to or inherits from, by qualified name: its declaration's clause and every extension's in the file. */
        private var conformances: [String: Set<String>] = [:]
        /* The keys synthesized Codable reads for each enum case with associated values, by the enum's qualified name: `case fetchHistory(...)` reads `FetchHistoryCodingKeys`. */
        private var perCaseKeys: [String: Set<String>] = [:]

        init(tree: SourceFileSyntax) {
            self.tree = tree
        }

        mutating func collect() {
            for statement in tree.statements {
                if let declaration = statement.item.as(DeclSyntax.self) {
                    survey(declaration, enclosing: nil, fileScoped: false)
                }
            }
            for statement in tree.statements {
                if let declaration = statement.item.as(DeclSyntax.self) {
                    visit(declaration, in: Scope())
                }
            }
        }

        /* The first pass: each type's qualified name, whether it is private to the file, and its conformances. */
        private mutating func survey(_ declaration: DeclSyntax, enclosing: String?, fileScoped: Bool) {
            if let extended = declaration.as(ExtensionDeclSyntax.self) {
                let name = extended.extendedType.trimmedDescription
                conformances[name, default: []].formUnion(Self.inherited(extended.inheritanceClause))
                for member in extended.memberBlock.members {
                    survey(
                        member.decl,
                        enclosing: name,
                        fileScoped: fileScoped || Self.isFileScoped(extended.modifiers),
                    )
                }
                return
            }
            guard let group = declaration.asProtocol((any DeclGroupSyntax).self),
                let named = declaration.asProtocol((any NamedDeclSyntax).self)
            else { return }
            let name = enclosing.map { "\($0).\(named.name.text)" } ?? named.name.text
            let scoped = fileScoped || Self.isFileScoped(group.modifiers)
            fileScopedTypes[name] = (fileScopedTypes[name] ?? true) && scoped
            conformances[name, default: []].formUnion(Self.inherited(group.inheritanceClause))
            for member in group.memberBlock.members {
                if let cases = member.decl.as(EnumCaseDeclSyntax.self) {
                    for element in cases.elements where element.parameterClause != nil {
                        perCaseKeys[name, default: []].insert(Self.perCaseKeysName(of: element.name.text))
                    }
                }
                survey(member.decl, enclosing: name, fileScoped: scoped)
            }
        }

        private mutating func visit(_ declaration: DeclSyntax, in scope: Scope) {
            let attributes = Self.attributeNames(declaration.asProtocol((any WithAttributesSyntax).self)?.attributes)
            let modifiers = declaration.asProtocol((any WithModifiersSyntax).self)?.modifiers ?? []
            let fileScoped = scope.fileScoped || Self.isFileScoped(modifiers)
            let runtime =
                !attributes.isDisjoint(with: Self.objectiveCAttributes)
                || modifiers.contains { ["dynamic", "optional"].contains($0.name.text) }
                || scope.attributes.contains("objcMembers")
            let entry = scope.attributes.contains("main") || scope.conformances.contains("PreviewProvider")
            let isOverride = modifiers.contains { $0.name.text == "override" }
            let saysPublic =
                modifiers.contains { ["public", "open"].contains($0.name.text) } || attributes.contains("_spi")
            let saysLower = modifiers.contains {
                ["internal", "package", "fileprivate", "private"].contains($0.name.text) && $0.detail == nil
            }
            let isPublic =
                scope.canBePublic
                && (saysPublic
                    || (!saysLower
                        && (scope.membersPublic || (scope.casesPublic && declaration.is(EnumCaseDeclSyntax.self)))))
            let reach: Reach = fileScoped ? .file : isPublic ? .api : .package

            if let extended = declaration.as(ExtensionDeclSyntax.self) {
                let name = extended.extendedType.trimmedDescription
                let inner = Scope(
                    fileScoped: Self.isFileScoped(extended.modifiers) || fileScopedTypes[name] == true,
                    canBePublic: !saysLower,
                    membersPublic: saysPublic,
                    typeName: name,
                    conformances: conformances[name] ?? [],
                    attributes: [],
                )
                for member in extended.memberBlock.members {
                    visit(member.decl, in: inner)
                }
                return
            }

            if let group = declaration.asProtocol((any DeclGroupSyntax).self),
                let named = declaration.asProtocol((any NamedDeclSyntax).self)
            {
                let name = scope.typeName.map { "\($0).\(named.name.text)" } ?? named.name.text
                let typeConformances = conformances[name] ?? []
                do {
                    var exemption: String?
                    if runtime {
                        exemption = UnusedDeclarations.objectiveC
                    }
                    else if entry || attributes.contains("main") || typeConformances.contains("PreviewProvider") {
                        exemption = UnusedDeclarations.entryPoint
                    }
                    else if !attributes.subtracting(Self.knownAttributes).isEmpty {
                        exemption = UnusedDeclarations.registered
                    }
                    let keysName = named.name.text
                    candidates.append(
                        Candidate(
                            node: Syntax(declaration),
                            name: named.name,
                            keyword: group.introducer.text,
                            exemption: exemption,
                            reach: reach,
                            isType: true,
                            isCodingKeys: keysName == "CodingKeys"
                                || scope.typeName.map { perCaseKeys[$0]?.contains(keysName) == true } == true,
                        )
                    )
                }
                let typeIndex = candidates.count - 1
                let inner = Scope(
                    fileScoped: fileScoped,
                    canBePublic: isPublic,
                    membersPublic: isPublic && declaration.is(ProtocolDeclSyntax.self),
                    casesPublic: isPublic && declaration.is(EnumDeclSyntax.self),
                    typeName: name,
                    isLayout: declaration.as(StructDeclSyntax.self).map(Self.isLayout) ?? false,
                    conformances: typeConformances,
                    attributes: attributes.union(entry ? ["main"] : []),
                )
                let membersStart = candidates.count
                for member in group.memberBlock.members {
                    visit(member.decl, in: inner)
                }
                /* A type that holds a declaration something registers is found through it: Swift Testing runs a type's `@Test` functions without a `@Suite`. */
                if candidates[typeIndex].exemption == nil,
                    candidates[membersStart...].contains(where: { $0.exemption == UnusedDeclarations.registered })
                {
                    candidates[typeIndex].exemption = UnusedDeclarations.registered
                }
                return
            }

            var exemption: String?
            if runtime {
                exemption = UnusedDeclarations.objectiveC
            }
            else if entry {
                exemption = UnusedDeclarations.entryPoint
            }

            if let function = declaration.as(FunctionDeclSyntax.self) {
                let name = function.name.text
                if exemption == nil {
                    if !attributes.subtracting(Self.knownAttributes).isEmpty {
                        exemption = UnusedDeclarations.registered
                    }
                    else if name == "callAsFunction"
                        || (scope.attributes.contains("resultBuilder") && name.hasPrefix("build"))
                        || (scope.attributes.contains("dynamicCallable") && name.hasPrefix("dynamicallyCall"))
                    {
                        exemption = UnusedDeclarations.compilerCalled
                    }
                }
                candidates.append(
                    Candidate(
                        node: Syntax(function),
                        name: function.name,
                        keyword: "func",
                        exemption: exemption,
                        reach: reach,
                        isOverride: isOverride,
                    )
                )
            }
            else if let variable = declaration.as(VariableDeclSyntax.self) {
                guard variable.bindings.count == 1, let binding = variable.bindings.first,
                    let pattern = binding.pattern.as(IdentifierPatternSyntax.self)
                else {
                    if let first = variable.bindings.first?.pattern.firstToken(viewMode: .sourceAccurate) {
                        candidates.append(
                            Candidate(
                                node: Syntax(variable),
                                name: first,
                                keyword: variable.bindingSpecifier.text,
                                exemption: UnusedDeclarations.sharedDeclaration,
                                reach: reach,
                            )
                        )
                    }
                    return
                }
                let name = pattern.identifier.text
                let isStored = Self.isStored(binding)
                let isStatic = modifiers.contains { ["static", "class"].contains($0.name.text) }
                let isInstanceStorage = isStored && scope.typeName != nil && !isStatic
                if exemption == nil {
                    if scope.attributes.contains("propertyWrapper")
                        && ["wrappedValue", "projectedValue"].contains(name)
                    {
                        exemption = UnusedDeclarations.compilerCalled
                    }
                    else if isInstanceStorage,
                        !scope.attributes.subtracting(Self.knownAttributes).subtracting(["main"]).isEmpty
                    {
                        /* A type a macro may expand (`@Model`), whose stored properties it may persist or read by name. */
                        exemption = UnusedDeclarations.registered
                    }
                }
                /* Asked after the type's conformances, which come first because a conformance reads the property whatever its value. */
                var storageExemption: String?
                if isInstanceStorage, let value = binding.initializer?.value, !Self.isPlainValue(value) {
                    storageExemption = UnusedDeclarations.lifetime
                }
                else if isInstanceStorage, scope.isLayout {
                    storageExemption = UnusedDeclarations.layout
                }
                candidates.append(
                    Candidate(
                        node: Syntax(variable),
                        name: pattern.identifier,
                        keyword: variable.bindingSpecifier.text,
                        exemption: exemption,
                        reach: reach,
                        isWrapped: !attributes.subtracting(Self.knownAttributes).isEmpty,
                        isOverride: isOverride,
                        isInstanceStorage: isInstanceStorage,
                        storageExemption: storageExemption,
                    )
                )
            }
            else if let alias = declaration.as(TypeAliasDeclSyntax.self) {
                candidates.append(
                    Candidate(
                        node: Syntax(alias),
                        name: alias.name,
                        keyword: "typealias",
                        exemption: exemption,
                        reach: reach,
                    )
                )
            }
            else if let cases = declaration.as(EnumCaseDeclSyntax.self) {
                guard let element = cases.elements.first else { return }
                if exemption == nil {
                    if cases.elements.count > 1 {
                        exemption = UnusedDeclarations.sharedDeclaration
                    }
                    else if scope.typeName?.hasSuffix("CodingKeys") == true {
                        exemption = UnusedDeclarations.codingKeys
                    }
                    else if !scope.conformances.isSubset(of: Self.caseBlindConformances) {
                        exemption = UnusedDeclarations.reachableCase
                    }
                }
                candidates.append(
                    Candidate(
                        node: Syntax(cases),
                        name: element.name,
                        keyword: "case",
                        exemption: exemption,
                        reach: reach,
                    )
                )
            }
            else if let subscriptDeclaration = declaration.as(SubscriptDeclSyntax.self) {
                let dynamicMember =
                    subscriptDeclaration.parameterClause.parameters.first?.firstName.text == "dynamicMember"
                if exemption == nil && dynamicMember {
                    exemption = UnusedDeclarations.compilerCalled
                }
                candidates.append(
                    Candidate(
                        node: Syntax(subscriptDeclaration),
                        name: subscriptDeclaration.subscriptKeyword,
                        keyword: "subscript",
                        exemption: exemption,
                        reach: reach,
                        isOverride: isOverride,
                    )
                )
            }
            else if let initializer = declaration.as(InitializerDeclSyntax.self) {
                candidates.append(
                    Candidate(
                        node: Syntax(initializer),
                        name: initializer.initKeyword,
                        keyword: "init",
                        exemption: exemption ?? UnusedDeclarations.initializer,
                        reach: reach,
                    )
                )
            }
        }

        /* The keys enum Swift's synthesized Codable reads for a case with associated values: the case's name, first letter capitalized, then `CodingKeys`. */
        static func perCaseKeysName(of caseName: String) -> String {
            caseName.prefix(1).uppercased() + caseName.dropFirst() + "CodingKeys"
        }

        /* `private` or `fileprivate` on the declaration itself, and not on its setter alone (`private(set)`). */
        static func isFileScoped(_ modifiers: DeclModifierListSyntax) -> Bool {
            modifiers.contains { ["private", "fileprivate"].contains($0.name.text) && $0.detail == nil }
        }

        /* Each name an inheritance clause lists, as its last component: `Swift.Codable` and `@unchecked Sendable` give `Codable` and `Sendable`. */
        static func inherited(_ clause: InheritanceClauseSyntax?) -> Set<String> {
            var names: Set<String> = []
            func add(_ type: TypeSyntax) {
                if let attributed = type.as(AttributedTypeSyntax.self) {
                    add(attributed.baseType)
                }
                else if let composition = type.as(CompositionTypeSyntax.self) {
                    composition.elements.forEach { add($0.type) }
                }
                else if let member = type.as(MemberTypeSyntax.self) {
                    names.insert(member.name.text)
                }
                else if let identifier = type.as(IdentifierTypeSyntax.self) {
                    names.insert(identifier.name.text)
                }
                else {
                    names.insert(type.trimmedDescription)
                }
            }
            clause?.inheritedTypes.forEach { add($0.type) }
            return names
        }

        static func attributeNames(_ attributes: AttributeListSyntax?) -> Set<String> {
            Set((attributes ?? []).compactMap { $0.as(AttributeSyntax.self)?.attributeName.trimmedDescription })
        }

        /* A stored property: no accessors, or only observers. */
        static func isStored(_ binding: PatternBindingSyntax) -> Bool {
            guard let accessors = binding.accessorBlock?.accessors else { return true }
            guard let list = accessors.as(AccessorDeclListSyntax.self) else { return false }
            return list.allSatisfy { ["willSet", "didSet"].contains($0.accessorSpecifier.text) }
        }

        /*
         A struct made only of numbers: every instance stored property is typed as an integer, a float, a `Bool` or a
         SIMD vector or matrix, or untyped and given a number. Such a struct is how Swift spells a C struct or a
         shader's constants, and its fields are read by offset, not by name.
         */
        static func isLayout(_ structure: StructDeclSyntax) -> Bool {
            var fields = 0
            for member in structure.memberBlock.members {
                guard let variable = member.decl.as(VariableDeclSyntax.self),
                    !variable.modifiers.contains(where: { $0.name.text == "static" })
                else { continue }
                for binding in variable.bindings where isStored(binding) {
                    fields += 1
                    if let type = binding.typeAnnotation?.type.trimmedDescription {
                        guard numericTypes.contains(type) || numericGenerics.contains(where: { type.hasPrefix($0) })
                        else { return false }
                    }
                    else if let value = binding.initializer?.value {
                        guard value.is(IntegerLiteralExprSyntax.self) || value.is(FloatLiteralExprSyntax.self) else {
                            return false
                        }
                    }
                    else {
                        return false
                    }
                }
            }
            return fields > 0
        }

        static let numericTypes: Set<String> = [
            "Int", "Int8", "Int16", "Int32", "Int64", "UInt", "UInt8", "UInt16", "UInt32", "UInt64", "Float", "Float16",
            "Float32", "Float64", "Double",
            "CGFloat", "Bool",
        ]

        /* SIMD vectors and the simd module's vector and matrix types, by how their names begin: `SIMD4<Float>`, `simd_float4x4`. */
        static let numericGenerics = ["SIMD2<", "SIMD3<", "SIMD4<", "SIMD8<", "SIMD16<", "simd_", "matrix_", "packed_"]

        /* A value that runs no code of anyone's to make: a literal, a collection of them, a member like `.zero`. */
        static func isPlainValue(_ expression: ExprSyntax) -> Bool {
            if expression.is(IntegerLiteralExprSyntax.self) || expression.is(FloatLiteralExprSyntax.self)
                || expression.is(BooleanLiteralExprSyntax.self)
                || expression.is(NilLiteralExprSyntax.self)
            {
                return true
            }
            if let string = expression.as(StringLiteralExprSyntax.self) {
                return string.segments.allSatisfy { $0.is(StringSegmentSyntax.self) }
            }
            if let array = expression.as(ArrayExprSyntax.self) {
                return array.elements.allSatisfy { isPlainValue($0.expression) }
            }
            if let dictionary = expression.as(DictionaryExprSyntax.self) {
                guard case .elements(let elements) = dictionary.content else { return true }
                return elements.allSatisfy { isPlainValue($0.key) && isPlainValue($0.value) }
            }
            if let prefix = expression.as(PrefixOperatorExprSyntax.self) {
                return prefix.operator.text == "-" && isPlainValue(prefix.expression)
            }
            if let member = expression.as(MemberAccessExprSyntax.self) {
                return member.base.map { $0.is(DeclReferenceExprSyntax.self) } ?? true
            }
            return false
        }

        /* Attributes that register nothing and that the Objective-C runtime does not read. Any other names something that may reach the declaration, a macro or a framework, and keeps it. */
        static let knownAttributes: Set<String> = [
            "available", "discardableResult", "inline", "inlinable", "usableFromInline", "frozen", "MainActor",
            "Sendable", "preconcurrency",
            "nonobjc", "ViewBuilder", "ToolbarContentBuilder", "SceneBuilder", "CommandsBuilder",
            "warn_unqualified_access", "_disfavoredOverload",
            "backDeployed", "specialize", "_specialize", "_effects", "_optimize", "_semantics", "resultBuilder",
            "propertyWrapper", "dynamicCallable",
            "dynamicMemberLookup", "unchecked", "retroactive", "Observable", "ObservationIgnored", "ObservationTracked",
            "concurrent", "safe", "unsafe",
        ]

        static let objectiveCAttributes: Set<String> = [
            "objc", "objcMembers", "IBAction", "IBOutlet", "IBInspectable", "IBDesignable", "IBSegueAction",
            "NSManaged", "GKInspectable", "_cdecl",
            "_silgen_name", "_dynamicReplacement", "NSApplicationMain", "UIApplicationMain",
        ]

        /* Conformances that never reach a case the program does not name. */
        static let caseBlindConformances: Set<String> = [
            "Equatable", "Hashable", "Comparable", "Sendable", "Error", "LocalizedError", "CustomStringConvertible",
            "CustomDebugStringConvertible",
        ]
    }

    /* Every name a piece of syntax spells, backticks and a projection's `$` taken off. */
    static func names(in syntax: Syntax) -> Set<String> {
        var names: Set<String> = []
        for token in syntax.tokens(viewMode: .sourceAccurate) {
            switch token.tokenKind {
                case .identifier(let text):
                    names.insert(text.trimmingCharacters(in: CharacterSet(charactersIn: "`")))
                case .dollarIdentifier(let text):
                    names.insert(String(text.dropFirst()))
                default:
                    break
            }
        }
        return names
    }

    /* Every `#if` in the file, outermost ones only, each with all its branches. */
    static func conditionals(in tree: SourceFileSyntax) -> [IfConfigDeclSyntax] {
        final class Finder: SyntaxVisitor {
            var found: [IfConfigDeclSyntax] = []
            override func visit(_ node: IfConfigDeclSyntax) -> SyntaxVisitorContinueKind {
                found.append(node)
                return .skipChildren
            }
        }
        let finder = Finder(viewMode: .sourceAccurate)
        finder.walk(tree)
        return finder.found
    }

    /*
     The edit that removes the declaration: its whole lines when it has them to itself, with the comments written
     directly above it and a comment trailing its last line, and otherwise only its own text. A `// MARK:` line
     above it stays, since it heads what follows rather than this one declaration. Removed from between two blank
     lines, it takes one of them, so no double gap is left where it was.
     */
    static func removal(of node: Syntax, in file: ParsedFile) -> FindingRecord.Edit {
        let utf8 = Array(file.source.utf8)
        let start = node.positionAfterSkippingLeadingTrivia.utf8Offset
        let end = node.endPositionBeforeTrailingTrivia.utf8Offset
        var lineStart = start
        while lineStart > 0 && utf8[lineStart - 1] != UInt8(ascii: "\n") {
            lineStart -= 1
        }
        var lineEnd = end
        while lineEnd < utf8.count && utf8[lineEnd] != UInt8(ascii: "\n") {
            lineEnd += 1
        }
        let blank: (UInt8) -> Bool = { $0 == UInt8(ascii: " ") || $0 == UInt8(ascii: "\t") }
        let rest = utf8[end..<lineEnd].drop(while: blank)
        guard utf8[lineStart..<start].allSatisfy(blank), rest.isEmpty || rest.starts(with: "//".utf8) else {
            return FindingRecord.Edit(start: start, end: end, text: "")
        }
        var removalStart = lineStart
        var cursor = start
        scan: for piece in node.leadingTrivia.pieces.reversed() {
            switch piece {
                case .spaces, .tabs:
                    cursor -= piece.sourceLength.utf8Length
                case .newlines(let count), .carriageReturnLineFeeds(let count):
                    guard count == 1 else { break scan }
                    cursor -= piece.sourceLength.utf8Length
                case .lineComment(let text), .blockComment(let text), .docLineComment(let text),
                    .docBlockComment(let text):
                    guard !text.hasPrefix("// MARK") else { break scan }
                    cursor -= piece.sourceLength.utf8Length
                    removalStart = cursor
                    while removalStart > 0 && utf8[removalStart - 1] != UInt8(ascii: "\n") {
                        removalStart -= 1
                    }
                default:
                    break scan
            }
        }
        var removalEnd = min(lineEnd + 1, utf8.count)
        if Self.isBlankLine(endingAt: removalStart, in: utf8), let next = Self.blankLineEnd(from: removalEnd, in: utf8)
        {
            removalEnd = next
        }
        return FindingRecord.Edit(start: removalStart, end: removalEnd, text: "")
    }

    /* Whether the line ending just before `offset` holds only spaces and tabs. */
    private static func isBlankLine(endingAt offset: Int, in utf8: [UInt8]) -> Bool {
        guard offset > 0 else { return false }
        var cursor = offset - 1
        while cursor > 0 && (utf8[cursor - 1] == UInt8(ascii: " ") || utf8[cursor - 1] == UInt8(ascii: "\t")) {
            cursor -= 1
        }
        return cursor == 0 || utf8[cursor - 1] == UInt8(ascii: "\n")
    }

    /* The offset past the line starting at `offset` when it holds only spaces and tabs, or nil. */
    private static func blankLineEnd(from offset: Int, in utf8: [UInt8]) -> Int? {
        var cursor = offset
        while cursor < utf8.count && (utf8[cursor] == UInt8(ascii: " ") || utf8[cursor] == UInt8(ascii: "\t")) {
            cursor += 1
        }
        guard cursor < utf8.count, utf8[cursor] == UInt8(ascii: "\n") else { return nil }
        return cursor + 1
    }
}
