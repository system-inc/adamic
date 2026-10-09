import Foundation

/*
 The compiler arguments the last build passed for each Swift file, read back from what the build recorded,
 so sourcekitd checks a file exactly as `swift build` compiled it.

 Swift 6.4's build system offers no supported way to ask for them: SwiftPM has no compile-commands output,
 and the build manifest names each compile without its arguments. The arguments are kept in the build
 description's `task-store.msgpack`, one MessagePack array per compile, shaped
 `["builtin-SwiftDriver", "--", <swiftc>, <arguments>...]`. That file is undocumented, so this reader
 trusts only that shape: it finds each array by its first string, decodes nothing but strings, and keeps a
 command only when it names a module and lists its files. Anything else is a loud failure, never an empty
 answer, because a command table that silently came back empty would leave every file unchecked.

 Measured on ahraos-macos: 24 arrays for 13 modules. The store repeats a command once per task that
 carries it, and an executable compiles twice (once normally, once testable for its tests, the testable
 one marked by `-entry-point-function-name`). A file takes the normal one when both list it.
 */
struct CompileCommands {
    /* One module's compile, reduced to what sourcekitd needs to type-check a file of it. */
    struct Command: Equatable {
        var moduleName: String
        /* Every argument for sourcekitd, the module's files included, driver-only flags removed. */
        var arguments: [String]
        var files: [String]
        var testable: Bool
    }

    /* The build left no record this reader recognises. */
    struct ReadFailure: Error, CustomStringConvertible {
        var description: String
    }

    let commands: [Command]
    private let commandForFile: [String: Int]

    init(commands: [Command]) {
        self.commands = commands
        var commandForFile: [String: Int] = [:]
        /* The normal compile first, so a file an executable and its testable twin both list takes the normal one. */
        for (index, command) in commands.enumerated().sorted(by: { !$0.element.testable && $1.element.testable }) {
            for file in command.files where commandForFile[file] == nil {
                commandForFile[file] = index
            }
        }
        self.commandForFile = commandForFile
    }

    /* The compile that built this file, or nil when the last build did not compile it. */
    func command(for file: URL) -> Command? {
        commandForFile[file.resolvingSymlinksInPath().path].map { commands[$0] }
    }

    /* The commands of the newest build description under this scratch path. */
    static func read(scratchPath: URL) throws -> CompileCommands {
        let buildData = scratchPath.appendingPathComponent("out/Intermediates.noindex/XCBuildData", isDirectory: true)
        let priors = buildData.appendingPathComponent("prior-build-descriptions.txt")
        let listing: String
        do {
            listing = try String(contentsOf: priors, encoding: .utf8)
        }
        catch {
            throw ReadFailure(
                description:
                    "the build left no \(priors.path), so this engine cannot tell which build description holds the compile arguments"
            )
        }
        /* The list is written oldest first; the newest is the build that just ran. */
        guard let newest = listing.split(whereSeparator: \.isNewline).last else {
            throw ReadFailure(description: "\(priors.path) names no build description")
        }
        let store = buildData.appendingPathComponent("\(newest).xcbuilddata/task-store.msgpack")
        let bytes: Data
        do {
            bytes = try Data(contentsOf: store)
        }
        catch {
            throw ReadFailure(
                description:
                    "the newest build description has no task store at \(store.path), so the build system's layout has moved"
            )
        }
        let commands = try driverCommands(in: [UInt8](bytes)).map(command(from:))
        guard !commands.isEmpty else {
            throw ReadFailure(
                description:
                    "\(store.path) holds no Swift compile this engine recognises, so the build system's record format has moved"
            )
        }
        return CompileCommands(commands: deduplicated(commands))
    }

    /* Every `builtin-SwiftDriver` array in the store, as its strings, the three-element prefix removed. */
    static func driverCommands(in bytes: [UInt8]) throws -> [[String]] {
        let marker = [UInt8(0xA0 | 19)] + Array("builtin-SwiftDriver".utf8)
        var commands: [[String]] = []
        var searchFrom = 0
        while let found = firstIndex(of: marker, in: bytes, from: searchFrom) {
            searchFrom = found + marker.count
            guard let count = arrayCount(endingAt: found, in: bytes) else { continue }
            var reader = MessagePackStrings(bytes: bytes, position: found)
            var strings: [String] = []
            for _ in 0..<count {
                guard let string = reader.next() else { break }
                strings.append(string)
            }
            /* A marker that is not the head of a whole command array is an occurrence of the words in some other record. */
            guard strings.count == count, count > 3, strings[1] == "--" else { continue }
            commands.append(Array(strings.dropFirst(3)))
            searchFrom = reader.position
        }
        return commands
    }

    /*
     The length of the MessagePack array whose header ends just before `index`. array16 and array32 are tried
     before fixarray, because the last byte of a longer header can itself look like a fixarray header.
     */
    private static func arrayCount(endingAt index: Int, in bytes: [UInt8]) -> Int? {
        if index >= 3, bytes[index - 3] == 0xDC {
            return Int(bytes[index - 2]) << 8 | Int(bytes[index - 1])
        }
        if index >= 5, bytes[index - 5] == 0xDD {
            return (index - 4..<index).reduce(0) { $0 << 8 | Int(bytes[$1]) }
        }
        if index >= 1, (0x90...0x9F).contains(bytes[index - 1]) {
            return Int(bytes[index - 1] & 0x0F)
        }
        return nil
    }

    private static func firstIndex(of pattern: [UInt8], in bytes: [UInt8], from start: Int) -> Int? {
        guard let first = pattern.first, bytes.count >= pattern.count else { return nil }
        var index = start
        while index <= bytes.count - pattern.count {
            if bytes[index] == first, bytes[index..<index + pattern.count].elementsEqual(pattern) {
                return index
            }
            index += 1
        }
        return nil
    }

    /*
     Flags only the driver understands, or that make sourcekitd write files: batching, incremental state,
     outputs, the explicit-module build and its caches. sourcekitd builds its modules implicitly, so the
     explicit module cache (the second `-module-cache-path`) is dropped and the first, implicit one kept.
     */
    private static let flagsDropped: Set<String> = [
        "-c", "-enable-batch-mode", "-incremental", "-save-temps", "-color-diagnostics", "-explicit-module-build",
        "-emit-dependencies", "-emit-module", "-serialize-diagnostics", "-emit-const-values", "-emit-objc-header",
        "-experimental-emit-module-separately", "-disable-cmo", "-validate-clang-modules-once",
    ]
    private static let flagsDroppedWithValue: Set<String> = [
        "-output-file-map", "-clang-scanner-module-cache-path", "-sdk-module-cache-path", "-emit-module-path",
        "-dependency-scan-serialize-diagnostics-path", "-const-gather-protocols-list", "-emit-objc-header-path",
        "-index-store-path", "-clang-build-session-file",
    ]

    /* One driver command, as sourcekitd should receive it: its file list read in, the driver's flags removed. */
    static func command(from driver: [String]) throws -> Command {
        var arguments: [String] = []
        var files: [String] = []
        var moduleName = ""
        var keptModuleCache = false
        var remaining = driver[...]
        while let argument = remaining.popFirst() {
            if flagsDropped.contains(argument)
                || (argument.hasPrefix("-j") && argument.dropFirst(2).allSatisfy(\.isNumber))
            {
                continue
            }
            if flagsDroppedWithValue.contains(argument) {
                remaining = remaining.dropFirst()
                continue
            }
            if argument == "-module-cache-path" {
                if keptModuleCache {
                    remaining = remaining.dropFirst()
                    continue
                }
                keptModuleCache = true
            }
            /* The build's stat cache of the SDK speeds its own clang; a stale one would make sourcekitd miss headers. */
            if argument == "-Xcc", remaining.first == "-ivfsstatcache" {
                remaining = remaining.dropFirst(3)
                continue
            }
            if argument == "-module-name", let name = remaining.first {
                moduleName = name
            }
            if argument.hasPrefix("@") {
                let list = String(argument.dropFirst())
                let listed: String
                do {
                    listed = try String(contentsOfFile: list, encoding: .utf8)
                }
                catch {
                    throw ReadFailure(
                        description:
                            "the compile of a module reads its files from \(list), and that file could not be read: \(error)"
                    )
                }
                let paths = listed.split(whereSeparator: \.isNewline).map {
                    URL(fileURLWithPath: String($0)).resolvingSymlinksInPath().path
                }
                files.append(contentsOf: paths)
                arguments.append(contentsOf: paths)
                continue
            }
            arguments.append(argument)
        }
        guard !moduleName.isEmpty else {
            throw ReadFailure(
                description: "a Swift compile in the build's task store names no module, so its record format has moved"
            )
        }
        return Command(
            moduleName: moduleName,
            arguments: arguments,
            files: files,
            testable: driver.contains("-entry-point-function-name"),
        )
    }

    private static func deduplicated(_ commands: [Command]) -> [Command] {
        var seen: [Command] = []
        for command in commands where !seen.contains(command) {
            seen.append(command)
        }
        return seen
    }

    /* Reads MessagePack strings one after another, and stops at the first value that is not one. */
    struct MessagePackStrings {
        let bytes: [UInt8]
        var position: Int

        mutating func next() -> String? {
            guard position < bytes.count else { return nil }
            let head = bytes[position]
            let length: Int
            let headerSize: Int
            switch head {
                case 0xA0...0xBF:
                    length = Int(head & 0x1F)
                    headerSize = 1
                case 0xD9:
                    guard position + 1 < bytes.count else { return nil }
                    length = Int(bytes[position + 1])
                    headerSize = 2
                case 0xDA:
                    guard position + 2 < bytes.count else { return nil }
                    length = Int(bytes[position + 1]) << 8 | Int(bytes[position + 2])
                    headerSize = 3
                case 0xDB:
                    guard position + 4 < bytes.count else { return nil }
                    length = (position + 1...position + 4).reduce(0) { $0 << 8 | Int(bytes[$1]) }
                    headerSize = 5
                default:
                    return nil
            }
            let start = position + headerSize
            guard start + length <= bytes.count else { return nil }
            position = start + length
            return String(decoding: bytes[start..<start + length], as: UTF8.self)
        }
    }
}
