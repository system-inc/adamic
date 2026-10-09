import Foundation
import IndexStoreShim

/*
 The index store every debug `swift build` writes beside its products (`<scratch>/out/v5`), read through the
 toolchain's own `libIndexStore.dylib`: for each file the build compiled, every name in it and the
 declaration it resolves to.

 It is the same compiler's verdict the `.dia` records hold, kept for names instead of diagnostics, and it
 costs nothing extra: the build already wrote it. A unit describes one compile of one file and names the
 record holding that file's occurrences. A unit is trusted for a file only when it is newer than the file as
 it stands, the rule the types phase applies to `.dia` records, because an older one describes text that is
 no longer there.

 Measured on Presence: 2,531 units, listed in 9ms.

 `@safe` though it holds the store and the library's C functions: they are private, bound in `init` and called
 only in the readers at the bottom of this class, and every string the library lends there is copied into Swift
 before the reader that lent it is disposed.
 */
@safe final class IndexStore {
    /* The library or the store could not be opened. */
    struct Failure: Error, CustomStringConvertible {
        var description: String
    }

    /* `INDEXSTORE_UNIT_DEPENDENCY_UNIT`: a module the unit's file imports, named by its module name. */
    private static let unitDependency: Int32 = 1
    /* `INDEXSTORE_UNIT_DEPENDENCY_RECORD`: the dependency that names the record of the unit's own file. */
    private static let recordDependency: Int32 = 2
    /* `INDEXSTORE_UNIT_DEPENDENCY_FILE`: a file the compile read, such as a Clang module's headers and the module maps it consulted. */
    private static let fileDependency: Int32 = 3
    /* `INDEXSTORE_SYMBOL_ROLE_DECLARATION` and `INDEXSTORE_SYMBOL_ROLE_DEFINITION`. */
    static let declarationRoles: UInt64 = 1 << 0 | 1 << 1
    /* `INDEXSTORE_SYMBOL_ROLE_REFERENCE`. */
    static let referenceRole: UInt64 = 1 << 2
    /* `INDEXSTORE_SYMBOL_ROLE_IMPLICIT`. */
    static let implicitRole: UInt64 = 1 << 8
    /* `INDEXSTORE_SYMBOL_ROLE_REL_CHILDOF`: a declaration, related to the type or extension that holds it. */
    static let childOfRole: UInt64 = 1 << 9
    /* `INDEXSTORE_SYMBOL_ROLE_REL_BASEOF`: a protocol or superclass named in an inheritance clause, related to the type, protocol or extension that names it. */
    static let baseOfRole: UInt64 = 1 << 10
    /* `INDEXSTORE_SYMBOL_ROLE_REL_OVERRIDEOF`: a declaration that overrides a superclass member or witnesses a protocol requirement, related to what it overrides. */
    static let overrideOfRole: UInt64 = 1 << 11
    /* `INDEXSTORE_SYMBOL_ROLE_REL_EXTENDEDBY`: the type an extension extends, related to the extension. */
    static let extendedByRole: UInt64 = 1 << 14
    /* The relations `occurrences(inRecord:relations:)` reads. */
    static let readRelations = childOfRole | baseOfRole | overrideOfRole | extendedByRole
    /* `INDEXSTORE_SYMBOL_KIND_CLASS`, `_PROTOCOL` and `_TYPEALIAS`. */
    static let classKind: Int32 = 7
    static let protocolKind: Int32 = 8
    static let typeAliasKind: Int32 = 11
    /* `INDEXSTORE_SYMBOL_KIND_CONSTRUCTOR`. */
    static let constructorKind: Int32 = 22
    /* `INDEXSTORE_SYMBOL_KIND_MODULE`: a module's own name, as an `import` line or a qualified name spells it. */
    static let moduleKind: Int32 = 1

    /*
     One unit: one compile of one file of ours, or one module the build imported (a `.swiftinterface` or a
     Clang `.pcm`, marked system). Its unit dependencies are what the file imports, the implicit standard
     library modules among them; for a system module they are the modules it imports in turn.
     */
    struct Unit: Sendable {
        var name: String
        /* The compiled file, resolved. Empty for a module the build imported. */
        var mainFile: String
        var module: String
        /* What the compile wrote, as the compiler names it: for a Swift module the build imported, the `.swiftinterface` it read; for a Clang module, its `.pcm`. */
        var outputFile: String
        var isSystem: Bool
        var written: Date
        var importedModules: [String]
        /* The records holding the unit's own file's occurrences. */
        var ownRecords: [String]
        /* Every record the unit names: for a system module, the records declaring what it exports. */
        var allRecords: [String]
        /* The file each record describes: for a Clang module, the header that declares what the record holds. */
        var recordFiles: [String: String]
        /* The files the compile read without recording them: for a Clang module, its headers and the module maps it consulted. */
        var files: [String]
    }

    /* One relation of an occurrence: its roles, and the symbol it relates the occurrence to. */
    struct Relation: Sendable {
        var roles: UInt64
        var symbol: String
    }

    /* One occurrence as the record holds it, roles and kind kept whole for questions `FileSymbols` does not ask. */
    struct RecordOccurrence: Sendable {
        var line: Int
        var column: Int
        var symbol: String
        var name: String
        var roles: UInt64
        var kind: Int32
        /* The occurrence's child-of, base-of, override-of and extended-by relations, when the record was read with them. */
        var relations: [Relation] = []

        /* What the occurrence overrides or witnesses. */
        var overridden: [String] { relations.filter { $0.roles & IndexStore.overrideOfRole != 0 }.map(\.symbol) }

        var isReference: Bool { roles & IndexStore.referenceRole != 0 }
        var isDeclaration: Bool { roles & IndexStore.declarationRoles != 0 }
        var isImplicit: Bool { roles & IndexStore.implicitRole != 0 }
    }

    /* One unit as its reader reports it, copied whole before the reader is disposed, since every string it lends dies with it. */
    private struct UnitReading {
        var mainFile: String
        var module: String
        var outputFile: String
        var isSystem: Bool
        var dependencies: [Dependency]
    }

    /* One dependency of a unit, of the three kinds `units` reads. */
    private enum Dependency {
        /* A module the unit's file imports, empty when the store names none. */
        case module(String)
        /* A record the unit names, and the file it describes. */
        case record(name: String, file: String)
        /* A file the compile read without recording it. */
        case file(String)
    }

    /* A Swift closure boxed so a C iterator can carry it as its context pointer. */
    private final class Visitor<Element> {
        let visit: (Element) -> Void

        init(_ visit: @escaping (Element) -> Void) {
            self.visit = visit
        }
    }

    private let store: CohereIndexStore
    private let storePath: URL
    private let dispose: CohereIndexStoreDispose
    private let unitsApply: CohereIndexStoreUnitsApply
    private let unitReaderCreate: CohereIndexUnitReaderCreate
    private let unitReaderDispose: CohereIndexUnitReaderDispose
    private let unitMainFile: CohereIndexUnitReaderGetMainFile
    private let unitModuleName: CohereIndexUnitReaderGetModuleName
    private let unitOutputFile: CohereIndexUnitReaderGetOutputFile
    private let unitIsSystem: CohereIndexUnitReaderIsSystemUnit
    private let dependenciesApply: CohereIndexUnitReaderDependenciesApply
    private let dependencyKind: CohereIndexUnitDependencyGetKind
    private let dependencyName: CohereIndexUnitDependencyGetName
    private let dependencyFilePath: CohereIndexUnitDependencyGetFilePath
    private let dependencyModuleName: CohereIndexUnitDependencyGetModuleName
    private let recordReaderCreate: CohereIndexRecordReaderCreate
    private let recordReaderDispose: CohereIndexRecordReaderDispose
    private let occurrencesApply: CohereIndexRecordReaderOccurrencesApply
    private let occurrenceSymbol: CohereIndexOccurrenceGetSymbol
    private let occurrenceRoles: CohereIndexOccurrenceGetRoles
    private let occurrenceLineColumn: CohereIndexOccurrenceGetLineColumn
    private let relationsApply: CohereIndexOccurrenceRelationsApply
    private let relationRoles: CohereIndexSymbolRelationGetRoles
    private let relationSymbol: CohereIndexSymbolRelationGetSymbol
    private let symbolName: CohereIndexSymbolGetString
    private let symbolIdentifier: CohereIndexSymbolGetString
    private let symbolKind: CohereIndexSymbolGetKind

    /* Every unit in the store, read once, on first use. */
    private var cachedUnits: [Unit]?

    init(libraryPath: String, storePath: URL) throws {
        guard let handle = unsafe dlopen(libraryPath, RTLD_NOW | RTLD_LOCAL) else {
            let reason = unsafe dlerror().map { unsafe String(cString: $0) } ?? "no reason given"
            throw Failure(description: "could not load the index store library at \(libraryPath): \(reason)")
        }
        /*
         One entry point, bound to the type the shim declares for it. Unsafe because the cast trusts the symbol to
         have that C signature, which holds because the shim's types follow indexstore.h, the header IndexStoreDB
         compiles against. Each binding below is unsafe for the same reason.
         */
        func symbol<Function>(_ name: String, as type: Function.Type) throws -> Function {
            guard let address = unsafe dlsym(handle, name) else {
                throw Failure(description: "the index store library at \(libraryPath) has no \(name)")
            }
            return unsafe unsafeBitCast(address, to: type)
        }
        let create = try unsafe symbol("indexstore_store_create", as: CohereIndexStoreCreate.self)
        unsafe dispose = try symbol("indexstore_store_dispose", as: CohereIndexStoreDispose.self)
        unsafe unitsApply = try symbol("indexstore_store_units_apply_f", as: CohereIndexStoreUnitsApply.self)
        unsafe unitReaderCreate = try symbol("indexstore_unit_reader_create", as: CohereIndexUnitReaderCreate.self)
        unsafe unitReaderDispose = try symbol("indexstore_unit_reader_dispose", as: CohereIndexUnitReaderDispose.self)
        unsafe unitMainFile = try symbol(
            "indexstore_unit_reader_get_main_file",
            as: CohereIndexUnitReaderGetMainFile.self,
        )
        unsafe unitModuleName = try symbol(
            "indexstore_unit_reader_get_module_name",
            as: CohereIndexUnitReaderGetModuleName.self,
        )
        unsafe unitOutputFile = try symbol(
            "indexstore_unit_reader_get_output_file",
            as: CohereIndexUnitReaderGetOutputFile.self,
        )
        unsafe unitIsSystem = try symbol(
            "indexstore_unit_reader_is_system_unit",
            as: CohereIndexUnitReaderIsSystemUnit.self,
        )
        unsafe dependenciesApply = try symbol(
            "indexstore_unit_reader_dependencies_apply_f",
            as: CohereIndexUnitReaderDependenciesApply.self,
        )
        unsafe dependencyKind = try symbol(
            "indexstore_unit_dependency_get_kind",
            as: CohereIndexUnitDependencyGetKind.self,
        )
        unsafe dependencyName = try symbol(
            "indexstore_unit_dependency_get_name",
            as: CohereIndexUnitDependencyGetName.self,
        )
        unsafe dependencyFilePath = try symbol(
            "indexstore_unit_dependency_get_filepath",
            as: CohereIndexUnitDependencyGetFilePath.self,
        )
        unsafe dependencyModuleName = try symbol(
            "indexstore_unit_dependency_get_modulename",
            as: CohereIndexUnitDependencyGetModuleName.self,
        )
        unsafe recordReaderCreate = try symbol(
            "indexstore_record_reader_create",
            as: CohereIndexRecordReaderCreate.self,
        )
        unsafe recordReaderDispose = try symbol(
            "indexstore_record_reader_dispose",
            as: CohereIndexRecordReaderDispose.self,
        )
        unsafe occurrencesApply = try symbol(
            "indexstore_record_reader_occurrences_apply_f",
            as: CohereIndexRecordReaderOccurrencesApply.self,
        )
        unsafe occurrenceSymbol = try symbol(
            "indexstore_occurrence_get_symbol",
            as: CohereIndexOccurrenceGetSymbol.self,
        )
        unsafe occurrenceRoles = try symbol("indexstore_occurrence_get_roles", as: CohereIndexOccurrenceGetRoles.self)
        unsafe occurrenceLineColumn = try symbol(
            "indexstore_occurrence_get_line_col",
            as: CohereIndexOccurrenceGetLineColumn.self,
        )
        unsafe relationsApply = try symbol(
            "indexstore_occurrence_relations_apply_f",
            as: CohereIndexOccurrenceRelationsApply.self,
        )
        unsafe relationRoles = try symbol(
            "indexstore_symbol_relation_get_roles",
            as: CohereIndexSymbolRelationGetRoles.self,
        )
        unsafe relationSymbol = try symbol(
            "indexstore_symbol_relation_get_symbol",
            as: CohereIndexSymbolRelationGetSymbol.self,
        )
        unsafe symbolName = try symbol("indexstore_symbol_get_name", as: CohereIndexSymbolGetString.self)
        unsafe symbolIdentifier = try symbol("indexstore_symbol_get_usr", as: CohereIndexSymbolGetString.self)
        unsafe symbolKind = try symbol("indexstore_symbol_get_kind", as: CohereIndexSymbolGetKind.self)
        guard FileManager.default.fileExists(atPath: storePath.appendingPathComponent("v5/units").path),
            let store = unsafe create(storePath.path, nil)
        else {
            throw Failure(
                description:
                    "no index store at \(storePath.path): the build wrote none, or wrote it where this engine does not look"
            )
        }
        unsafe self.store = store
        self.storePath = storePath
    }

    deinit {
        unsafe dispose(store)
    }

    /* The library beside the `swift` that `xcrun` resolves. */
    static func toolchainLibraryPath(runner: ProcessRunner = ProcessRunner()) throws -> String {
        let library = try SerializedDiagnosticsReader.toolchainLibraryPath(runner: runner)
        return URL(fileURLWithPath: library).deletingLastPathComponent().appendingPathComponent("libIndexStore.dylib")
            .path
    }

    /* The store a build into this scratch path writes. */
    static func storePath(scratchPath: URL) -> URL {
        scratchPath.appendingPathComponent("out", isDirectory: true)
    }

    /* Every name in the file and what it resolves to, or nil when no unit newer than the file describes it. */
    func symbols(of file: URL) -> FileSymbols? {
        let path = file.resolvingSymlinksInPath().path
        let sourceModified =
            (try? file.resourceValues(forKeys: [.contentModificationDateKey]).contentModificationDate) ?? .distantFuture
        guard
            let newest = records()[path]?.filter({ $0.written >= sourceModified }).max(by: { $0.written < $1.written })
        else { return nil }
        guard let recorded = occurrences(inRecord: newest.record) else { return nil }
        return FileSymbols(
            recorded.map { occurrence in
                FileSymbols.Occurrence(
                    line: occurrence.line,
                    column: occurrence.column,
                    symbol: occurrence.symbol,
                    name: occurrence.name,
                    isReference: occurrence.isReference,
                    isImplicit: occurrence.isImplicit,
                )
            }
        )
    }

    /*
     The unit that describes each file as it stands, by the file's path as given: the newest unit of ours, across
     the stores, that names the file and was written after the file was. A file with none is absent: the build has
     not compiled it as it stands, so its record describes text that is no longer there.
     */
    static func freshUnits(
        of files: [ParsedFile],
        in stores: [IndexStore],
    ) -> [String: (store: IndexStore, unit: Unit)] {
        let newest = newestUnits(in: stores)
        var fresh: [String: (store: IndexStore, unit: Unit)] = [:]
        for file in files {
            let modified =
                (try? file.url.resourceValues(forKeys: [.contentModificationDateKey]).contentModificationDate)
                ?? .distantFuture
            if let described = newest[file.url.resolvingSymlinksInPath().path], described.unit.written >= modified {
                fresh[file.url.path] = described
            }
        }
        return fresh
    }

    /* The newest unit of ours that names each compiled file, across the stores, by the file's resolved path. */
    static func newestUnits(in stores: [IndexStore]) -> [String: (store: IndexStore, unit: Unit)] {
        var newest: [String: (store: IndexStore, unit: Unit)] = [:]
        for store in stores {
            for unit in store.units() where !unit.isSystem && !unit.mainFile.isEmpty && !unit.ownRecords.isEmpty {
                if let held = newest[unit.mainFile], held.unit.written >= unit.written {
                    continue
                }
                newest[unit.mainFile] = (store, unit)
            }
        }
        return newest
    }

    /* Each file's records, by resolved path, with when the unit naming them was written. */
    private func records() -> [String: [(record: String, written: Date)]] {
        var byFile: [String: [(record: String, written: Date)]] = [:]
        for unit in units() where !unit.mainFile.isEmpty {
            byFile[unit.mainFile, default: []].append(contentsOf: unit.ownRecords.map { ($0, unit.written) })
        }
        return byFile
    }

    /*
     One spelling for a path under an SDK. The index names the same header through `MacOSX.sdk` and through the
     versioned SDK it links to, so each SDK root is resolved once and every path under it is written through that.
     */
    final class SDKRoots {
        private var resolved: [String: String] = [:]

        func canonical(_ path: String) -> String {
            guard let range = path.range(of: ".sdk/") else { return path }
            let root = String(path[..<range.upperBound])
            if resolved[root] == nil {
                resolved[root] = URL(fileURLWithPath: root).resolvingSymlinksInPath().path + "/"
            }
            return (resolved[root] ?? root) + path[range.upperBound...]
        }
    }

    /* Every unit in the store, read once. Measured on Presence: 2,531 units, listed in 9ms. */
    func units() -> [Unit] {
        if let cachedUnits {
            return cachedUnits
        }
        let unitsDirectory = storePath.appendingPathComponent("v5/units", isDirectory: true)
        var units: [Unit] = []
        let sdkRoots = SDKRoots()
        for name in unitNames() {
            guard let reading = readUnit(named: name) else { continue }
            let written =
                (try? unitsDirectory.appendingPathComponent(name).resourceValues(forKeys: [.contentModificationDateKey])
                    .contentModificationDate) ?? .distantPast
            var imported: [String] = []
            var own: [String] = []
            var all: [String] = []
            var recordFiles: [String: String] = [:]
            var files: [String] = []
            for dependency in reading.dependencies {
                switch dependency {
                    case .module(let module):
                        if !module.isEmpty {
                            imported.append(module)
                        }
                    case .record(let record, let file):
                        all.append(record)
                        recordFiles[record] = sdkRoots.canonical(file)
                        /* The record of the unit's own file, matched by path: a unit's other records belong to files it only read. */
                        if file == reading.mainFile {
                            own.append(record)
                        }
                    case .file(let file):
                        files.append(sdkRoots.canonical(file))
                }
            }
            units.append(
                Unit(
                    name: name,
                    mainFile: reading.mainFile.isEmpty
                        ? "" : URL(fileURLWithPath: reading.mainFile).resolvingSymlinksInPath().path,
                    module: reading.module,
                    outputFile: reading.outputFile,
                    isSystem: reading.isSystem,
                    written: written,
                    importedModules: imported,
                    ownRecords: own,
                    allRecords: all,
                    recordFiles: recordFiles,
                    files: files,
                )
            )
        }
        cachedUnits = units
        return units
    }

    /*
     Every occurrence one record holds, or nil when the record cannot be read. A reader: each call into the library
     is handed the open reader or an occurrence it yielded, and `text` copies every string before the `defer`
     disposes the reader. With `relations`, each occurrence also carries its relations of the kinds
     `readRelations` names, read through the same iterator `forEach` drives, whose relation handles live as long
     as the occurrence that yields them.
     */
    func occurrences(inRecord record: String, relations: Bool = false) -> [RecordOccurrence]? {
        guard let reader = unsafe recordReaderCreate(store, record, nil) else { return nil }
        defer { unsafe recordReaderDispose(reader) }
        var occurrences: [RecordOccurrence] = []
        unsafe Self.forEach(in: reader, occurrencesApply) { occurrence in
            var line: UInt32 = 0
            var column: UInt32 = 0
            unsafe occurrenceLineColumn(occurrence, &line, &column)
            let symbol = unsafe occurrenceSymbol(occurrence)
            let roles = unsafe occurrenceRoles(occurrence)
            var related: [Relation] = []
            if relations && roles & Self.readRelations != 0 {
                unsafe Self.forEach(in: occurrence, relationsApply) { relation in
                    let relationRoles = unsafe relationRoles(relation)
                    if relationRoles & Self.readRelations != 0 {
                        related.append(
                            Relation(
                                roles: relationRoles,
                                symbol: unsafe Self.text(symbolIdentifier(relationSymbol(relation))),
                            )
                        )
                    }
                }
            }
            unsafe occurrences.append(
                RecordOccurrence(
                    line: Int(line),
                    column: Int(column),
                    symbol: Self.text(symbolIdentifier(symbol)),
                    name: Self.text(symbolName(symbol)),
                    roles: roles,
                    kind: Int32(symbolKind(symbol)),
                    relations: related,
                )
            )
        }
        return occurrences
    }

    /*
     One unit's reader, read whole into Swift values. A reader like `occurrences(inRecord:)`, unsafe and correct for
     the same reasons: every getter is handed the open reader or a dependency it yielded, and every string is copied
     before the `defer` disposes the reader.
     */
    private func readUnit(named name: String) -> UnitReading? {
        guard let reader = unsafe unitReaderCreate(store, name, nil) else { return nil }
        defer { unsafe unitReaderDispose(reader) }
        let mainFile = unsafe Self.text(unitMainFile(reader))
        var dependencies: [Dependency] = []
        unsafe Self.forEach(in: reader, dependenciesApply) { dependency in
            switch unsafe dependencyKind(dependency) {
                case Self.unitDependency:
                    dependencies.append(.module(unsafe Self.text(dependencyModuleName(dependency))))
                case Self.recordDependency:
                    dependencies.append(
                        unsafe .record(
                            name: Self.text(dependencyName(dependency)),
                            file: Self.text(dependencyFilePath(dependency)),
                        )
                    )
                case Self.fileDependency:
                    dependencies.append(.file(unsafe Self.text(dependencyFilePath(dependency))))
                default:
                    break
            }
        }
        return unsafe UnitReading(
            mainFile: mainFile,
            module: Self.text(unitModuleName(reader)),
            outputFile: Self.text(unitOutputFile(reader)),
            isSystem: unitIsSystem(reader),
            dependencies: dependencies,
        )
    }

    /*
     The name of every unit in the store. With `forEach`, one of the two places a Swift closure crosses into the
     library, here for the iterator that yields names rather than handles. Unsafe and correct for the reasons
     `forEach` gives; each name is copied inside the applier, since the store lends it only for the call.
     */
    private func unitNames() -> [String] {
        var names: [String] = []
        let visitor = Visitor<String> { names.append($0) }
        withExtendedLifetime(visitor) {
            _ = unsafe unitsApply(store, 0, Unmanaged.passUnretained(visitor).toOpaque()) { context, name in
                guard let context = unsafe context else { return false }
                unsafe Unmanaged<Visitor<String>>.fromOpaque(context).takeUnretainedValue().visit(IndexStore.text(name))
                return true
            }
        }
        return names
    }

    /*
     Hands `visit` each handle an iterator yields: a unit's dependencies, a record's occurrences or an occurrence's
     relations, whose iterators share one C type. A C applier can capture nothing, so `visit` crosses as the context pointer, boxed
     and unretained. Unsafe because nothing in that pointer says what it points at or keeps it alive; correct
     because only this function's applier reads it back, as the type it boxed, and the iterator is synchronous,
     calling the applier only before it returns, while `withExtendedLifetime` holds the box.
     */
    private static func forEach(
        in reader: UnsafeMutableRawPointer,
        _ iterate: CohereIndexRecordReaderOccurrencesApply,
        _ visit: (UnsafeMutableRawPointer) -> Void,
    ) {
        unsafe withoutActuallyEscaping(visit) { visit in
            let visitor = unsafe Visitor(visit)
            unsafe withExtendedLifetime(visitor) {
                _ = unsafe iterate(reader, Unmanaged.passUnretained(visitor).toOpaque()) { context, handle in
                    guard let context = unsafe context, let handle = unsafe handle else { return true }
                    unsafe Unmanaged<Visitor<UnsafeMutableRawPointer>>.fromOpaque(context).takeUnretainedValue().visit(
                        handle
                    )
                    return true
                }
            }
        }
    }

    /*
     A string the store lends, copied into a Swift one. Unsafe because the store lends a bare pointer and a length,
     not NUL-terminated; correct because the length bounds the read and every caller copies before the reader that
     lent the string is disposed.
     */
    private static func text(_ string: CohereIndexString) -> String {
        guard let data = unsafe string.data else { return "" }
        return unsafe String(decoding: UnsafeRawBufferPointer(start: data, count: string.length), as: UTF8.self)
    }
}
