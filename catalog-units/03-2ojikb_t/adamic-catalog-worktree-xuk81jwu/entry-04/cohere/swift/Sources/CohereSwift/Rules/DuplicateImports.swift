import SwiftSyntax

/*
 No import written twice: `import Foundation` and, a few lines later, `import Foundation` again. The second
 says nothing the first did not, and it hides which one was meant, so a reader editing one of them cannot
 tell whether the other still matters. The repair is to delete the repeat, and a fix does that when the
 import sits alone on its line with no comment of its own.

 Two imports are the same only when everything written on them matches: attributes, modifiers, the
 declaration kind, and the full module path. Each difference below is a different import, not a repeat:
 - `import Foo` and `import Foo.Bar`. Whether the umbrella module already brings in a submodule is decided by
   the module map (Foundation re-exports its submodules, CoreImage keeps `CIFilterBuiltins` apart), and
   syntax cannot read it. SwiftLint flags these as duplicates and special-cases CoreImage; we do not flag
   them, since deleting one may break the build.
 - `import Foo` and `import struct Foo.Bar`. A scoped import is not redundant beside the whole module: it
   also decides which `Bar` an unqualified name means when two modules declare one. SwiftLint flags these.
 - `@testable import Foo`, `@_exported import Foo`, `@preconcurrency import Foo`, `@_spi(Name) import Foo`,
   `public import Foo` and a plain `import Foo` each mean something different, so a plain import beside an
   attributed one is not a repeat. SwiftLint ignores attributes and modifiers entirely and flags all of these.
 - Imports in different `#if` branches are not duplicates; only one branch is compiled.

 An import counts as a repeat when another identical import is active wherever it is: one earlier in the
 same `#if` branch, or one in an enclosing branch or at top level. So an `import Foo` inside `#if DEBUG` is
 reported when a top-level `import Foo` already exists, whichever comes first in the file. The reverse is
 not reported: a top-level import is not a repeat of one inside `#if DEBUG`, because deleting it breaks
 every build where `DEBUG` is off. SwiftLint gets this backwards, flagging the top-level import that follows
 a guarded one and missing the guarded one under a top-level import.

 An import whose attribute list itself holds an `#if` is skipped, because its spelling depends on the build.
 */
public struct DuplicateImports: FileRule {
    public let name = "cohere-swift/duplicate-imports"
    public let origin = RuleOrigin.swiftLint
    public let upstreamName: String? = "duplicate_imports"

    public init() {}

    public func findings(in file: ParsedFile) -> [FindingRecord] {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(file.tree)
        let imports = visitor.found
        var findings: [FindingRecord] = []
        for (position, candidate) in imports.enumerated() {
            let covering = imports.enumerated().first { otherPosition, other in
                guard otherPosition != position, other.key == candidate.key else { return false }
                if other.clauses == candidate.clauses {
                    return otherPosition < position
                }
                return other.clauses.count < candidate.clauses.count
                    && Array(candidate.clauses.prefix(other.clauses.count)) == other.clauses
            }
            guard let covering else { continue }
            let coveringLine = covering.element.declaration.startLocation(converter: file.locations).line
            findings.append(
                file.finding(
                    at: candidate.declaration,
                    rule: name,
                    messageId: "duplicateImport",
                    message:
                        "This import repeats the one on line \(coveringLine), which is already in effect wherever this one is, so it adds nothing and hides which of the two was meant. Delete this line.",
                    fixes: Self.deletion(of: candidate.declaration, in: file).map { [$0] } ?? [],
                )
            )
        }
        return findings
    }

    /*
     The edit that removes the import's whole line, or nil when deleting the line could take something else
     with it: code sharing the line (`import A; import B`), a comment inside or after the import, or a comment
     above it that would be left describing the wrong line.
     */
    static func deletion(of declaration: ImportDeclSyntax, in file: ParsedFile) -> FindingRecord.Edit? {
        if declaration.leadingTrivia.contains(where: \.isComment)
            || declaration.trailingTrivia.contains(where: \.isComment)
        {
            return nil
        }
        let inner = declaration.trimmed
        let tokens = Array(inner.tokens(viewMode: .sourceAccurate))
        if tokens.contains(where: {
            $0.leadingTrivia.contains(where: \.isComment) || $0.trailingTrivia.contains(where: \.isComment)
        }) {
            return nil
        }
        let bytes = Array(file.source.utf8)
        let space = UInt8(ascii: " ")
        let tab = UInt8(ascii: "\t")
        let lineFeed = UInt8(ascii: "\n")
        let carriageReturn = UInt8(ascii: "\r")

        var lineStart = declaration.positionAfterSkippingLeadingTrivia.utf8Offset
        while lineStart > 0, bytes[lineStart - 1] == space || bytes[lineStart - 1] == tab {
            lineStart -= 1
        }
        if lineStart > 0, bytes[lineStart - 1] != lineFeed, bytes[lineStart - 1] != carriageReturn {
            return nil
        }
        var lineEnd = declaration.endPositionBeforeTrailingTrivia.utf8Offset
        while lineEnd < bytes.count, bytes[lineEnd] == space || bytes[lineEnd] == tab {
            lineEnd += 1
        }
        if lineEnd == bytes.count {
            /* The last line, with no newline after it: take the newline before it instead, so no blank line is left. */
            var previousLineEnd = lineStart
            if previousLineEnd > 0, bytes[previousLineEnd - 1] == lineFeed {
                previousLineEnd -= 1
            }
            if previousLineEnd > 0, bytes[previousLineEnd - 1] == carriageReturn {
                previousLineEnd -= 1
            }
            return FindingRecord.Edit(start: previousLineEnd, end: lineEnd, text: "")
        }
        if bytes[lineEnd] == carriageReturn, lineEnd + 1 < bytes.count, bytes[lineEnd + 1] == lineFeed {
            return FindingRecord.Edit(start: lineStart, end: lineEnd + 2, text: "")
        }
        if bytes[lineEnd] == lineFeed || bytes[lineEnd] == carriageReturn {
            return FindingRecord.Edit(start: lineStart, end: lineEnd + 1, text: "")
        }
        return nil
    }

    /* One file-level import: what it says, and the `#if` clauses it sits inside, outermost first. */
    struct ImportRecord {
        var declaration: ImportDeclSyntax
        var key: String
        var clauses: [SyntaxIdentifier]
    }

    /* Collects every import that sits at file level, directly or inside `#if` clauses, in file order. */
    final class Visitor: SyntaxVisitor {
        private(set) var found: [ImportRecord] = []

        override func visit(_ node: ImportDeclSyntax) -> SyntaxVisitorContinueKind {
            if let key = Self.key(of: node), let clauses = Self.clauses(enclosing: node) {
                found.append(ImportRecord(declaration: node, key: key, clauses: clauses))
            }
            return .skipChildren
        }

        /* Everything written on the import, without trivia, so spacing and comments do not make two imports differ. */
        static func key(of node: ImportDeclSyntax) -> String? {
            var attributes: [String] = []
            for element in node.attributes {
                guard case let .attribute(attribute) = element else { return nil }
                attributes.append(spelling(of: attribute))
            }
            let modifiers = node.modifiers.map { spelling(of: $0) }
            let kind = node.importKindSpecifier?.text ?? ""
            let path = node.path.map { unquoted($0.name.text) }.joined(separator: ".")
            return [attributes.sorted().joined(separator: " "), modifiers.sorted().joined(separator: " "), kind, path]
                .joined(separator: "|")
        }

        /* An identifier written in backticks names the same module as one written without. */
        static func unquoted(_ identifier: String) -> String {
            identifier.count > 1 && identifier.hasPrefix("`") && identifier.hasSuffix("`")
                ? String(identifier.dropFirst().dropLast()) : identifier
        }

        static func spelling(of node: some SyntaxProtocol) -> String {
            node.tokens(viewMode: .sourceAccurate).map(\.text).joined()
        }

        /* The `#if` clauses between the import and the file, or nil if the import sits anywhere else, which only a broken file does. */
        static func clauses(enclosing node: ImportDeclSyntax) -> [SyntaxIdentifier]? {
            var clauses: [SyntaxIdentifier] = []
            var current = node.parent
            while let ancestor = current {
                if ancestor.is(SourceFileSyntax.self) {
                    return clauses.reversed()
                }
                if ancestor.is(IfConfigClauseSyntax.self) {
                    clauses.append(ancestor.id)
                }
                else if !(ancestor.is(CodeBlockItemSyntax.self) || ancestor.is(CodeBlockItemListSyntax.self)
                    || ancestor.is(IfConfigClauseListSyntax.self) || ancestor.is(IfConfigDeclSyntax.self))
                {
                    return nil
                }
                current = ancestor.parent
            }
            return nil
        }
    }
}
