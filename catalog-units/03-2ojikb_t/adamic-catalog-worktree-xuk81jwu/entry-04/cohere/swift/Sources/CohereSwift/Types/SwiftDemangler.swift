import Foundation
import Synchronization

/*
 The toolchain's own demangler, `libswiftDemangle.dylib` beside `libIndexStore.dylib`: it turns a Swift
 declaration's mangled name into the declaration it spells, signature and all. `$sSTsE7forEachyyy7ElementQzKXEKF`
 is `(extension in Swift):Swift.Sequence.forEach((A.Element) throws -> ()) throws -> ()`.

 The index names the declaration a name resolves to by its unified symbol name, which for a Swift declaration
 is its mangled name with `s:` in place of `$s`. A typed rule that needs more than which declaration (the type
 of a parameter, say) reads it here, in the compiler's own words, rather than in a reading of the mangling of
 ours. Loaded at run time, as the index store library is, so the binary never carries one Xcode's path.

 `@safe` though it holds a C function: the function is private, and its one call is in `demangle`, which hands it
 a buffer `demangle` sized and owns.
 */
@safe final class SwiftDemangler: Sendable {
    /* The library could not be found or opened. */
    struct Failure: Error, CustomStringConvertible {
        var description: String
    }

    /* `swift_demangle_getDemangledName(const char *, char *, size_t) -> size_t`: the full length, written up to the buffer's size. */
    private typealias GetDemangledName =
        @convention(c) (UnsafePointer<CChar>?, UnsafeMutablePointer<CChar>?, Int) -> Int

    private let getDemangledName: GetDemangledName

    /* The process's demangler, loaded the first time anyone asks. A load that failed is retried, since nothing was kept. */
    private static let loaded = Mutex<SwiftDemangler?>(nil)

    static func shared(runner: ProcessRunner = ProcessRunner()) throws -> SwiftDemangler {
        try loaded.withLock { demangler in
            if let demangler {
                return demangler
            }
            let created = try SwiftDemangler(libraryPath: toolchainLibraryPath(runner: runner))
            demangler = created
            return created
        }
    }

    init(libraryPath: String) throws {
        guard let handle = unsafe dlopen(libraryPath, RTLD_NOW | RTLD_LOCAL) else {
            let reason = unsafe dlerror().map { unsafe String(cString: $0) } ?? "no reason given"
            throw Failure(description: "could not load the demangler at \(libraryPath): \(reason)")
        }
        guard let address = unsafe dlsym(handle, "swift_demangle_getDemangledName") else {
            throw Failure(description: "the demangler at \(libraryPath) has no swift_demangle_getDemangledName")
        }
        /* Unsafe because the cast trusts the symbol to have the signature `GetDemangledName` spells, the one SwiftDemangle.h declares. */
        unsafe getDemangledName = unsafeBitCast(address, to: GetDemangledName.self)
    }

    /* The library beside the `swift` that `xcrun` resolves. */
    static func toolchainLibraryPath(runner: ProcessRunner = ProcessRunner()) throws -> String {
        let library = try SerializedDiagnosticsReader.toolchainLibraryPath(runner: runner)
        return URL(fileURLWithPath: library).deletingLastPathComponent().appendingPathComponent(
            "libswiftDemangle.dylib"
        ).path
    }

    /*
     The declaration a unified symbol name spells, or nil for a name that is not a Swift declaration's (`c:` is
     C's and Objective-C's) or that does not demangle.
     */
    func declaration(ofSymbol symbol: String) -> String? {
        guard symbol.hasPrefix("s:") else { return nil }
        return demangle("$s" + symbol.dropFirst(2))
    }

    /* A mangled name, `$s` and all, as the demangler prints it. The first call says how long the answer is when the buffer was short. */
    func demangle(_ mangled: String) -> String? {
        var capacity = 512
        while true {
            var buffer = [CChar](repeating: 0, count: capacity)
            /*
             The one call into the library, unsafe because a C function is trusted with a raw buffer. It stays in
             bounds because it is told the buffer's own count and writes no further, and the buffer is lent only
             for the call.
             */
            let length = mangled.withCString { name in
                buffer.withUnsafeMutableBufferPointer { output in
                    unsafe getDemangledName(name, output.baseAddress, output.count)
                }
            }
            guard length > 0 else { return nil }
            if length < capacity {
                return String(decoding: buffer.prefix(length).map { UInt8(bitPattern: $0) }, as: UTF8.self)
            }
            capacity = length + 1
        }
    }
}
