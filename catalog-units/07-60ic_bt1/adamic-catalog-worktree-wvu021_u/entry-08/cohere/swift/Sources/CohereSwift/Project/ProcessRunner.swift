import Foundation

/*
 Runs a tool (`swift`, `xcrun`) and hands back everything it said.

 Output goes to temporary files rather than pipes. A pipe holds about 64 KB, and a process that fills it
 blocks until someone reads; reading only after the process exits then waits forever. `swift package
 describe` on a large package passes that size, and a hang there would look like a slow build rather
 than a bug.
 */
public struct ProcessRunner: Sendable {
    /* What one invocation produced. */
    public struct Result: Sendable {
        public var exitCode: Int32
        public var standardOutput: Data
        public var standardError: String

        public var succeeded: Bool { exitCode == 0 }
    }

    /* Why a tool could not be run at all, as distinct from running and failing. */
    public struct LaunchFailure: Error, CustomStringConvertible {
        public var command: String
        public var underlying: String

        public var description: String { "could not run \(command): \(underlying)" }
    }

    public init() {}

    public func run(_ executable: String, _ arguments: [String], in directory: URL) throws -> Result {
        let temporaryDirectory = FileManager.default.temporaryDirectory
        let identifier = UUID().uuidString
        let outputFile = temporaryDirectory.appendingPathComponent("cohere-swift-\(identifier).out")
        let errorFile = temporaryDirectory.appendingPathComponent("cohere-swift-\(identifier).err")
        FileManager.default.createFile(atPath: outputFile.path, contents: nil)
        FileManager.default.createFile(atPath: errorFile.path, contents: nil)
        defer {
            removeTemporaryFile(outputFile)
            removeTemporaryFile(errorFile)
        }

        let process = Process()
        process.executableURL = URL(fileURLWithPath: "/usr/bin/env")
        process.arguments = [executable] + arguments
        process.currentDirectoryURL = directory
        do {
            process.standardOutput = try FileHandle(forWritingTo: outputFile)
            process.standardError = try FileHandle(forWritingTo: errorFile)
            try process.run()
        }
        catch {
            throw LaunchFailure(command: ([executable] + arguments).joined(separator: " "), underlying: "\(error)")
        }
        process.waitUntilExit()

        /*
         A file that cannot be read back is an error rather than an empty answer: empty output from `swift
         package describe` would read as a package with nothing in it, which would check nothing and say so cleanly.
         */
        let output = try Data(contentsOf: outputFile)
        let errorText = String(decoding: try Data(contentsOf: errorFile), as: UTF8.self)
        return Result(exitCode: process.terminationStatus, standardOutput: output, standardError: errorText)
    }

    private func removeTemporaryFile(_ url: URL) {
        do {
            try FileManager.default.removeItem(at: url)
        }
        catch {
            /*
             Ignored on purpose. A leftover file costs a few bytes in a directory the system sweeps, and
             failing a finished run over it would trade a real answer for tidiness.
             */
        }
    }
}
