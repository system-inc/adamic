import ClangDiagnosticsShim
import Foundation

/*
 Reads the compiler's serialized diagnostics (`.dia` files) through the toolchain's own libclang.

 This is how the types phase knows what the compiler said about a file even when this run's build did not
 recompile it. An incremental `swift build` prints only the diagnostics of files it recompiled; measured on
 Swift 6.4, a cold build printed a warning and a no-op rebuild printed nothing while the warning was still
 in the code. Every compiled file's `.dia` survives until that file compiles again, so reading them
 answers for every file, not only the ones that happened to rebuild.

 libclang is loaded with dlopen from the toolchain `xcrun` names, not linked by path. See the shim header
 for why.

 `@safe` though it holds libclang's C functions: they are private, called only in `read` and `text`, and every
 pointer libclang lends there is copied into a Swift value before the set or string that lent it is disposed.
 */
@safe final class SerializedDiagnosticsReader {
    /* One diagnostic as the compiler serialized it. Columns are 1-based and count bytes, which is the contract's unit already. */
    struct Diagnostic: Equatable {
        var file: String
        var line: Int
        var column: Int
        var severity: FindingRecord.Severity?
        var message: String
        /* The warning group, `VariableNeverMutated`, carried in libclang's category text. Empty for errors that belong to no group. */
        var group: String
    }

    /* libclang could not be found or loaded, so no compiler diagnostic can be read and the types phase cannot vouch for anything. */
    struct LoadFailure: Error, CustomStringConvertible {
        var description: String
    }

    private let loadDiagnostics: CohereClangLoadDiagnostics
    private let numberOfDiagnostics: CohereClangGetNumDiagnosticsInSet
    private let diagnosticAt: CohereClangGetDiagnosticInSet
    private let disposeSet: CohereClangDisposeDiagnosticSet
    private let severityOf: CohereClangGetDiagnosticSeverity
    private let spellingOf: CohereClangGetDiagnosticSpelling
    private let categoryOf: CohereClangGetDiagnosticCategoryText
    private let locationOf: CohereClangGetDiagnosticLocation
    private let fileLocation: CohereClangGetFileLocation
    private let fileName: CohereClangGetFileName
    private let cString: CohereClangGetCString
    private let disposeString: CohereClangDisposeString

    init(libraryPath: String) throws {
        guard let handle = unsafe dlopen(libraryPath, RTLD_NOW | RTLD_LOCAL) else {
            let reason = unsafe dlerror().map { unsafe String(cString: $0) } ?? "no reason given"
            throw LoadFailure(description: "could not load libclang at \(libraryPath): \(reason)")
        }
        /*
         One entry point, bound to the type the shim declares for it. Unsafe because the cast trusts the symbol to
         have that C signature, which holds because the shim's types follow libclang's own headers, stable since its
         first release. Each binding below is unsafe for the same reason.
         */
        func symbol<Function>(_ name: String, as type: Function.Type) throws -> Function {
            guard let address = unsafe dlsym(handle, name) else {
                throw LoadFailure(description: "libclang at \(libraryPath) has no \(name)")
            }
            return unsafe unsafeBitCast(address, to: type)
        }
        unsafe loadDiagnostics = try symbol("clang_loadDiagnostics", as: CohereClangLoadDiagnostics.self)
        unsafe numberOfDiagnostics = try symbol(
            "clang_getNumDiagnosticsInSet",
            as: CohereClangGetNumDiagnosticsInSet.self,
        )
        unsafe diagnosticAt = try symbol("clang_getDiagnosticInSet", as: CohereClangGetDiagnosticInSet.self)
        unsafe disposeSet = try symbol("clang_disposeDiagnosticSet", as: CohereClangDisposeDiagnosticSet.self)
        unsafe severityOf = try symbol("clang_getDiagnosticSeverity", as: CohereClangGetDiagnosticSeverity.self)
        unsafe spellingOf = try symbol("clang_getDiagnosticSpelling", as: CohereClangGetDiagnosticSpelling.self)
        unsafe categoryOf = try symbol(
            "clang_getDiagnosticCategoryText",
            as: CohereClangGetDiagnosticCategoryText.self,
        )
        unsafe locationOf = try symbol("clang_getDiagnosticLocation", as: CohereClangGetDiagnosticLocation.self)
        unsafe fileLocation = try symbol("clang_getFileLocation", as: CohereClangGetFileLocation.self)
        unsafe fileName = try symbol("clang_getFileName", as: CohereClangGetFileName.self)
        unsafe cString = try symbol("clang_getCString", as: CohereClangGetCString.self)
        unsafe disposeString = try symbol("clang_disposeString", as: CohereClangDisposeString.self)
    }

    /* The libclang beside the `swift` that `xcrun` resolves: the toolchain that also builds the checked package. */
    static func toolchainLibraryPath(runner: ProcessRunner = ProcessRunner()) throws -> String {
        let result = try runner.run("xcrun", ["--find", "swift"], in: URL(fileURLWithPath: "/"))
        let swift = String(decoding: result.standardOutput, as: UTF8.self).trimmingCharacters(
            in: .whitespacesAndNewlines
        )
        guard result.succeeded, !swift.isEmpty else {
            throw LoadFailure(
                description:
                    "xcrun could not find swift, so there is no toolchain libclang to read diagnostics with: \(result.standardError)"
            )
        }
        return URL(fileURLWithPath: swift)
            .deletingLastPathComponent()
            .deletingLastPathComponent()
            .appendingPathComponent("lib/libclang.dylib")
            .path
    }

    /*
     Every diagnostic in one `.dia` file. The one function that walks a set, so its calls into libclang are marked
     here: they pass back the handles libclang lent, and every one of them runs before the set is disposed.
     */
    func read(_ url: URL) throws -> [Diagnostic] {
        var error: Int32 = 0
        guard let set = unsafe loadDiagnostics(url.path, &error, nil) else {
            throw LoadFailure(
                description: "\(url.path) could not be read as serialized diagnostics (libclang error \(error))"
            )
        }
        defer { unsafe disposeSet(set) }
        var diagnostics: [Diagnostic] = []
        for index in unsafe 0..<numberOfDiagnostics(set) {
            guard let diagnostic = unsafe diagnosticAt(set, index) else { continue }
            var file: CohereClangFile?
            var line: UInt32 = 0
            var column: UInt32 = 0
            var offset: UInt32 = 0
            unsafe fileLocation(locationOf(diagnostic), &file, &line, &column, &offset)
            unsafe diagnostics.append(
                Diagnostic(
                    file: file.map { unsafe text(fileName($0)) } ?? "",
                    line: Int(line),
                    column: Int(column),
                    severity: Self.severity(severityOf(diagnostic)),
                    message: text(spellingOf(diagnostic)),
                    group: text(categoryOf(diagnostic)),
                )
            )
        }
        return diagnostics
    }

    /* libclang's CXDiagnosticSeverity: 0 ignored, 1 note, 2 warning, 3 error, 4 fatal. Notes and ignored diagnostics are not findings. */
    static func severity(_ raw: Int32) -> FindingRecord.Severity? {
        switch raw {
            case 2: .warning
            case 3, 4: .error
            default: nil
        }
    }

    /*
     A libclang string copied into a Swift one and disposed. Unsafe because libclang lends a bare C string; correct
     because the copy is made before the dispose, which this function owns, so no caller ever holds the pointer.
     */
    private func text(_ string: CohereClangString) -> String {
        defer { unsafe disposeString(string) }
        return unsafe cString(string).map { unsafe String(cString: $0) } ?? ""
    }
}
