import Foundation
import SwiftFormat

/*
 The one way Swift is formatted: swift-format 604.0.0's own `Configuration`, built here in code, so no
 repository carries a `.swift-format` and no project chooses its own layout (#mg4dgjm, ruled 2026-10-03).
 It is TypeScript's model, where the Prettier options live in one block of the Nexus tier and every other
 place is refused (internal/format/formatoptions/resolve.go).

 The values are the `.swift-format` both proving grounds carried, identical in each, plus the ruling's
 changes toward the TypeScript house: trailing commas always, one argument per line once a call breaks,
 the return type kept with the closing parenthesis, generic requirements breaking like arguments, and
 indented case labels. Existing line breaks stay respected, because turning that off joins short bodies
 onto one line, which Prettier never does, and still changed files on a second pass. Range operators take
 no spaces.

 Every setting is written out, so the house reads in one place and a swift-format upgrade that moves a
 default cannot move the house with it. A setting carries its reason where it differs from swift-format's
 default.

 A `.swift-format` left at or above a file the engine owns is refused by name, never merged: a file
 cohere does not read, which `Format.sh` or an editor still would, is two formatters pulling the same
 lines apart.
 */
enum HouseSwiftFormat {
    static var configuration: Configuration {
        var configuration = Configuration()
        /* TypeScript's printWidth. */
        configuration.lineLength = 120
        /* TypeScript's tabWidth. */
        configuration.indentation = .spaces(4)
        configuration.tabWidth = 8
        configuration.maximumBlankLines = 1
        /* One space before a trailing comment, as Prettier writes it. */
        configuration.spacesBeforeEndOfLineComments = 1
        /* Kept on by ruling: off joins short bodies onto one line and squashes written-out dictionaries, which Prettier never does. */
        configuration.respectsExistingLineBreaks = true
        /* `else` and `catch` begin their own line, as the TypeScript house prints them. */
        configuration.lineBreakBeforeControlFlowKeywords = true
        /* Prettier's fits-or-breaks-all: a call that does not fit puts each argument on its own line. */
        configuration.lineBreakBeforeEachArgument = true
        /* Generic requirements break the way arguments do. */
        configuration.lineBreakBeforeEachGenericRequirement = true
        configuration.lineBreakBetweenDeclarationAttributes = false
        /* The return type stays with the closing parenthesis, as Prettier keeps it. */
        configuration.prioritizeKeepingFunctionOutputTogether = true
        configuration.indentConditionalCompilationBlocks = true
        /* Kept off by ruling: on breaks a trailing callback away from its call, the opposite of Prettier. */
        configuration.lineBreakAroundMultilineExpressionChainComponents = false
        configuration.fileScopedDeclarationPrivacy.accessLevel = .private
        /* Prettier indents case labels inside a switch. */
        configuration.indentSwitchCaseLabels = true
        /* Kept off by ruling: `0..<count`, no spaces around `..<` and `...`. */
        configuration.spacesAroundRangeFormationOperators = false
        configuration.noAssignmentInExpressions = NoAssignmentInExpressionsConfiguration()
        /* TypeScript's trailingComma `all`: a list broken over lines ends every element with a comma. */
        configuration.multilineTrailingCommaBehavior = .alwaysUsed
        configuration.multiElementCollectionTrailingCommas = true
        configuration.reflowMultilineStringLiterals = .never
        configuration.indentBlankLines = false
        configuration.orderedImports = OrderedImportsConfiguration()
        configuration.rules = rules
        return configuration
    }

    /*
     Every rule swift-format 604.0.0 has, each named, so a rule a later swift-format adds arrives missing and a
     test says so, rather than arriving on or off by accident. Only the rules that rewrite code matter to the
     formatter; the lint-only ones are listed for completeness, since cohere runs its own lint.
     */
    static let rules: [String: Bool] = [
        "AllPublicDeclarationsHaveDocumentation": false,
        "AlwaysUseLiteralForEmptyCollectionInit": false,
        "AlwaysUseLowerCamelCase": true,
        "BeginDocumentationCommentWithOneLineSummary": false,
        "DoNotUseSemicolons": true,
        /* Off: cohere's own private-over-fileprivate rule owns this, and two rewriters of one line disagree. */
        "FileScopedDeclarationPrivacy": false,
        "NeverForceUnwrap": false,
        "NeverUseForceTry": false,
        "NeverUseImplicitlyUnwrappedOptionals": false,
        /* Off: an access level written on an extension stays where it was written rather than being pushed onto each member. */
        "NoAccessLevelOnExtensionDeclaration": false,
        /* Off: the house writes its why-comments as block comments. */
        "NoBlockComments": false,
        "NoEmptyLinesOpeningClosingBraces": false,
        "NoLeadingUnderscores": false,
        "OmitExplicitReturns": false,
        /* Off: `case a, b` stays on one line as written. */
        "OneCasePerLine": false,
        "OneVariableDeclarationPerLine": true,
        "OrderedImports": true,
        "ReturnVoidInsteadOfEmptyTuple": true,
        "UseEarlyExits": false,
        "UseShorthandTypeNames": true,
        "UseSingleLinePropertyGetter": true,
        "UseTripleSlashForDocumentationComments": true,
        "UseWhereClausesInForLoops": false,
        "ValidateDocumentationComments": false,

        /*
         Off, though on by swift-format's default: the `.swift-format` the house was ruled from listed its rules,
         and swift-format reads a rule missing from a listed table as off. They stay off until a ruling turns
         one on, because each that rewrites code (NoParensAroundConditions, GroupNumericLiterals and others)
         reformats files nobody measured.
         */
        "AmbiguousTrailingClosureOverload": false,
        "AvoidRetroactiveConformances": false,
        "DontRepeatTypeInStaticProperties": false,
        "FullyIndirectEnum": false,
        "GroupNumericLiterals": false,
        "IdentifiersMustBeASCII": false,
        "NoAssignmentInExpressions": false,
        "NoCasesWithOnlyFallthrough": false,
        "NoEmptyTrailingClosureParentheses": false,
        "NoLabelsInCasePatterns": false,
        "NoParensAroundConditions": false,
        "NoPlaygroundLiterals": false,
        "NoVoidReturnOnFunctionSignature": false,
        "OnlyOneTrailingClosureArgument": false,
        "ReplaceForEachWithForLoop": false,
        "TypeNamesShouldBeCapitalized": false,
        "UseExplicitNilCheckInConditions": false,
        "UseLetInEveryBoundCaseVariable": false,
        "UseSynthesizedInitializer": false,
    ]

    /*
     Every `.swift-format` at or above one of the files, up to and including the boundary (the repository
     root), sorted. Each directory is looked at once however many files sit below it. Nothing above the
     boundary is looked at, so a config belonging to some directory above the checkout is never ours to
     refuse.
     */
    static func leftoverConfigurationFiles(above files: [URL], boundary: URL) -> [URL] {
        let boundaryPath = boundary.resolvingSymlinksInPath().path
        var visited: Set<String> = []
        var found: [URL] = []
        for file in files {
            var directory = file.resolvingSymlinksInPath().deletingLastPathComponent()
            while directory.path == boundaryPath || directory.path.hasPrefix(boundaryPath + "/"),
                visited.insert(directory.path).inserted
            {
                let candidate = directory.appendingPathComponent(".swift-format")
                if FileManager.default.fileExists(atPath: candidate.path) {
                    found.append(candidate)
                }
                if directory.path == boundaryPath || directory.path == "/" {
                    break
                }
                directory.deleteLastPathComponent()
            }
        }
        return found.sorted { $0.path < $1.path }
    }

    /* The refusal for one leftover file: what it is, why it is refused, and what to do. */
    static func refusal(of leftover: URL) -> String {
        "\(leftover.path) is a .swift-format, and Swift has one house format built into cohere-swift, so this file is never read and anything that still reads it formats differently from cohere; delete it, and bring any setting it needs to the house format (HouseSwiftFormat.swift)"
    }
}
