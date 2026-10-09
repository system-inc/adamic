import CohereSwift
import Foundation

/*
 The process boundary: arguments in, records out on stdout, prose on stderr, and an exit code that
 matches the summary.

 Exit 2 means the engine could not run (a refused flag, a missing package, a toolchain that would not
 answer), and in that case no summary is written, so the front door can only say that nothing was
 checked. Nothing here prints to stdout except through the contract writer.
 */
@main
struct CohereSwiftCommand {
    static func main() async {
        let workingDirectory = URL(fileURLWithPath: FileManager.default.currentDirectoryPath, isDirectory: true)
        do {
            let options = try CommandOptions.parse(
                Array(CommandLine.arguments.dropFirst()),
                workingDirectory: workingDirectory,
            )
            let pipeline = Pipeline(options: options, writer: .standardOutput(), workingDirectory: workingDirectory)
            exit(try await pipeline.run())
        }
        catch {
            FileHandle.standardError.write(Data("cohere-swift: \(error)\n".utf8))
            exit(2)
        }
    }
}
