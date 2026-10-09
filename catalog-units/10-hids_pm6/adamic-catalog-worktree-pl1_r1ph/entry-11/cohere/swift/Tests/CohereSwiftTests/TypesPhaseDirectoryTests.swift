import Testing

@testable import CohereSwift

/*
 Which target a build directory's records belong to. The names are the ones measured in real scratches:
 ahraos-macos for the testable variant, this engine's own package for an executable named for its product.
 A prefix match read `cohere-swift-parity-p.build` as `cohere-swift`'s and left the parity target's files
 without a record.
 */
struct TypesPhaseDirectoryTests {
    @Test func eachDirectoryNamesItsOwnTargetOrProduct() {
        #expect(TypesPhase.directoryStem("AhraOs-p.build") == "AhraOs")
        #expect(TypesPhase.directoryStem("CohereSwift-t.build") == "CohereSwift")
        #expect(TypesPhase.directoryStem("AhraOsServicesTests-p.build") == "AhraOsServicesTests")
        #expect(TypesPhase.directoryStem("AhraOsServices--36C56AD10F55DF70-testable-t.build") == "AhraOsServices")
        #expect(TypesPhase.directoryStem("cohere-swift-p.build") == "cohere-swift")
        #expect(TypesPhase.directoryStem("cohere-swift-parity-p.build") == "cohere-swift-parity")
    }

    /*
     The types phase names swiftbuild, the build system whose records it reads. Swift 6.4's help lists it as the
     default and 6.3's as an option; a toolchain whose help lists no swiftbuild is refused by name.
     */
    @Test func swiftbuildIsNamedOnlyWhenTheToolchainOffersIt() {
        let current = """
              --build-system <build-system>
                                      Specify the build system to use. (default: swiftbuild)
                    native            - Native Build System (deprecated)
                    swiftbuild        - Swift Build build engine (default)
                    xcode             - Xcode build system integration (deprecated)
            """
        let older = """
              --build-system <build-system>
                                      (values: native, xcode; default: native)
            """
        #expect(TypesPhase.buildSystemArguments(help: current) == ["--build-system", "swiftbuild"])
        let inline = "  --build-system <build-system>  (values: native, swiftbuild, xcode; default: native)"
        #expect(TypesPhase.buildSystemArguments(help: inline) == ["--build-system", "swiftbuild"])
        #expect(TypesPhase.buildSystemArguments(help: older) == nil)
        #expect(TypesPhase.buildSystemArguments(help: "") == nil)
    }
}
