package release

// SwiftEngineFileName is the Swift engine's executable in a platform package, in `bin/` beside cohere's
// own. The release puts it there and a released cohere looks for it there and nowhere else, so both read
// this one spelling.
const SwiftEngineFileName = "cohere-swift"

// SwiftEngineStampFileName is the generated file that carries the release's commit into the Swift engine,
// beside EngineVersion.swift, which reads it under COHERE_RELEASE_STAMP. swift/.gitignore names it.
const SwiftEngineStampFileName = "EngineReleaseStamp.generated.swift"

// ShipsSwiftEngine reports whether a platform's package carries the Swift engine.
//
// macOS only. The engine links the toolchain's libraries and reads its compiler's serialized
// diagnostics, so it ships where Xcode does. On every other platform cohere refuses a Swift package by
// name rather than checking nothing, and the release stages no engine there.
func ShipsSwiftEngine(operatingSystem string) bool {
	return operatingSystem == "darwin"
}
