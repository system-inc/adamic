import Foundation
import SwiftSyntax

/*
 A file is named for what it holds: `Type.swift` for the type it declares, and `Type+Purpose.swift` for
 extensions of a type it does not declare.

 Kirk's ruling. A reader looking for `ScrollbackBuffer` opens `ScrollbackBuffer.swift`, and a reader
 looking for what was added to `String` finds every `String+…` file together. The Swift form of our
 TypeScript `require-matching-file-name`. How many types a file holds is not judged (Kirk, 2026-10-03:
 TypeScript has no such rule, and file-length catches the file that is actually too big); what is
 judged is that the name leads to the main one.

 - A file declaring a type is named for it. Holding several, it is named for the main one, so any type
   may be the file's name as long as one is: the one its name matches, else the first is the main type,
   and the file is told to take that name. Extensions of the main type may sit beside it.
 - An extension of any other type belongs in that type's `+Purpose` file. A nested type can be named
   either way: `Outer.Inner+Purpose.swift` or `Outer+Purpose.swift`.
 - `main.swift` keeps its name, which SwiftPM gives meaning to (top-level code), so naming is not judged
   there. One-type-per-file still is.
 - Two extensions of another type stay beside the file's type, because moving them would cost what the
   ruling is for. Both were raised by @system_cohere_swift_ahraos_macos on real code:
   - A `private` or `fileprivate` extension. It is used only in this file, and moving it out would force
     it to `internal`, widening access to satisfy a naming rule. This holds in a `Type+Purpose.swift` file
     too (the Circle's second question, on `Terminal+AuthoritativeCheckpoint.swift`'s `VtColor` helper), as
     long as the file has a visible face of its own: a declared type or a non-private extension it is
     named for. A file of nothing but private extensions answers to the plain naming rule.
   - An extension whose every member names the file's type. `extension View { func delayWidthUntilIdle()
     }` beside `struct DelayWidthUntilIdle: ViewModifier` is the modifier's public face, and putting it in
     `View+Something.swift` separates the modifier from its only entry point, the opposite of findability.
 - A `Type+Purpose.swift` file may hold helper types beside the visible extension it is named for: private or
   fileprivate of any size, or a struct or enum of at most 30 lines (Kirk's ruling of 2026-10-03,
   #r3hfpe8). A larger internal type there is a type the file is not named for, so it is reported. Found by
   @system_cohere_swift_ahraos_presence un-nesting a probe file's helpers and Prune's `BinarySlice`: the
   ruling said they need not nest, and this rule then told the file to be renamed after the helper. A file
   whose only face is a helper still answers to the plain naming rule.
 - A file with no types and no extensions (free functions, a script) has no name to match.
 - `Name.generated.swift` is named `Name`. The marker says a generator wrote the file (Kirk's convention, shared
   with TypeScript), not what it holds, and a generated file answers to this rule like any other.
 */
public struct ConsistencyRequireMatchingFileName: FileRule {
    public let name = "cohere-swift/consistency-require-matching-file-name"
    public let origin = RuleOrigin.house
    public let upstreamName: String? = nil

    public init() {}

    public func applies(to file: ParsedFile) -> Bool {
        file.url.lastPathComponent != "main.swift"
    }

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let fileName = file.url.lastPathComponent
        var stem = file.url.deletingPathExtension().lastPathComponent
        if stem.hasSuffix(Self.generatedMarker) {
            stem.removeLast(Self.generatedMarker.count)
        }
        let declarations = TopLevelDeclarations(file.tree)
        var findings: [FindingRecord] = []

        let primary = declarations.primary(stem: stem)
        /* A visible extension the file is named for: the face of a `Type+Purpose.swift` file. */
        let hasPurposeFace = declarations.extensions.contains { other in
            Self.isPurposeFile(stem: stem, for: other.name)
                && !(other.extensionDeclaration.map(Self.isFilePrivate) ?? false)
        }
        /* Every type here is a helper to that face, so the file is named for the face, not for a helper. */
        let helpersOnly =
            hasPurposeFace && !declarations.types.isEmpty
            && declarations.types.allSatisfy { Self.isHelper($0, declarations: declarations, in: file) }
        if let declared = primary, declared.name != stem, !helpersOnly {
            findings.append(
                file.finding(
                    at: declared.token,
                    rule: name,
                    message: RuleMessages.ConsistencyRequireMatchingFileName.requireMatchingFileName(
                        typeName: declared.name,
                        fileName: fileName,
                        alternative: .none,
                    ),
                )
            )
        }

        let declaredName = helpersOnly ? nil : primary?.name
        let helperNames = helpersOnly ? Set(declarations.types.map(\.name)) : []
        /* The file has a public face: a type it declares, or a visible extension it is named for. A file-private helper may sit beside either. */
        let anchored = declaredName != nil || hasPurposeFace
        for extended in declarations.extensions
        where extended.name != declaredName && !helperNames.contains(extended.name) {
            if Self.isPurposeFile(stem: stem, for: extended.name) {
                continue
            }
            if let declaration = extended.extensionDeclaration, anchored, Self.isFilePrivate(declaration) {
                continue
            }
            if let declaredName, let declaration = extended.extensionDeclaration,
                Self.everyMemberNames(declaredName, in: declaration)
            {
                continue
            }
            let suggestion =
                "\(extended.name.split(separator: ".").first.map(String.init) ?? extended.name)+Purpose.swift"
            let placement =
                declaredName.map {
                    RuleMessages.ConsistencyRequireMatchingFileName.Placement.besideType(
                        extendedName: extended.name,
                        typeName: $0,
                        suggestion: suggestion,
                    )
                }
                ?? .withoutType(fileName: fileName, extendedName: extended.name, suggestion: suggestion)
            findings.append(
                file.finding(
                    at: extended.token,
                    rule: name,
                    message: RuleMessages.ConsistencyRequireMatchingFileName.extensionOutsideItsFile(
                        placement: placement
                    ),
                )
            )
        }
        return findings
    }

    /* What Kirk's convention puts between a generated file's name and `.swift`. */
    static let generatedMarker = ".generated"

    /* The most lines a struct or enum may have and still count as a helper in a purpose file. */
    static let smallValueTypeLines = 30

    /* A type that may sit beside the extension a `Type+Purpose.swift` file is named for: private or fileprivate of any size, or a struct or enum of at most `smallValueTypeLines`. */
    static func isHelper(
        _ type: TopLevelDeclarations.Entry,
        declarations: TopLevelDeclarations,
        in file: ParsedFile,
    ) -> Bool {
        guard let declaration = type.token.parent.flatMap({ DeclSyntax($0) }) else { return false }
        if let modifiers = declaration.asProtocol((any WithModifiersSyntax).self)?.modifiers,
            modifiers.contains(where: {
                $0.name.tokenKind == .keyword(.private) || $0.name.tokenKind == .keyword(.fileprivate)
            })
        {
            return true
        }
        guard declaration.is(StructDeclSyntax.self) || declaration.is(EnumDeclSyntax.self) else { return false }
        let lines =
            lineCount(of: declaration, in: file)
            + declarations.extensions
            .filter { $0.name == type.name }
            .compactMap(\.extensionDeclaration)
            .map { lineCount(of: DeclSyntax($0), in: file) }
            .reduce(0, +)
        return lines <= smallValueTypeLines
    }

    /*
     The lines a declaration occupies, from the comment directly above it (or its first attribute or keyword)
     to its closing brace. A comment counts when no blank line separates it from the declaration, so a file
     header above the first type is never billed to it.
     */
    static func lineCount(of declaration: DeclSyntax, in file: ParsedFile) -> Int {
        guard let firstToken = declaration.firstToken(viewMode: .sourceAccurate) else { return 0 }
        var start = firstToken.positionAfterSkippingLeadingTrivia
        var position = firstToken.position
        var attachedComment: AbsolutePosition?
        /* Line breaks since the last comment; two of them, with only spaces between, make a blank line that detaches every comment above it. */
        var lineBreaks = 0
        for piece in firstToken.leadingTrivia {
            switch piece {
                case .newlines(let count), .carriageReturns(let count), .carriageReturnLineFeeds(let count):
                    lineBreaks += count
                    if lineBreaks > 1 {
                        attachedComment = nil
                    }
                default:
                    if piece.isComment {
                        attachedComment = attachedComment ?? position
                        lineBreaks = 0
                    }
            }
            position = position.advanced(by: piece.sourceLength.utf8Length)
        }
        if let attachedComment {
            start = attachedComment
        }
        let startLine = file.locations.location(for: start).line
        let endLine = declaration.endLocation(converter: file.locations).line
        return endLine - startLine + 1
    }

    static func isFilePrivate(_ declaration: ExtensionDeclSyntax) -> Bool {
        declaration.modifiers.contains {
            $0.name.tokenKind == .keyword(.private) || $0.name.tokenKind == .keyword(.fileprivate)
        }
    }

    /* Every member mentions the type by name, in its signature or its body: the extension exists to expose that type. */
    static func everyMemberNames(_ typeName: String, in declaration: ExtensionDeclSyntax) -> Bool {
        let members = declaration.memberBlock.members
        guard !members.isEmpty else { return false }
        return members.allSatisfy { member in
            member.decl.tokens(viewMode: .sourceAccurate).contains { $0.tokenKind == .identifier(typeName) }
        }
    }

    /* `Type+Purpose`, or `Outer+Purpose` for a nested `Outer.Inner`, or exactly the extended name. */
    static func isPurposeFile(stem: String, for extended: String) -> Bool {
        if stem == extended || stem.hasPrefix(extended + "+") {
            return true
        }
        if let outer = extended.split(separator: ".").first, extended.contains(".") {
            return stem == String(outer) || stem.hasPrefix(outer + "+")
        }
        return false
    }
}
