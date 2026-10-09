import Foundation
import SwiftSyntax

/*
 One file, parsed once, with everything later phases need to report against it.

 Every phase that reads syntax reads this same tree. That is what lets the fix phase's walk be reused by
 lint when nothing was rewritten, and it is why a rule and the formatter can never disagree about what a
 file says.
 */
public struct ParsedFile: Sendable {
    public var url: URL
    public var targetName: String
    public var targetKind: String
    public var source: String
    public var tree: SourceFileSyntax
    public var locations: SourceLocationConverter
    /* Every node in the tree, counted once at parse time for the coverage line's `nodes visited`. */
    public var nodeCount: Int
    /* The root of the package whose target compiles this file (a local package's own root, not the checked root's); nil where nothing set it, such as a test fixture. */
    public var packageRoot: URL?

    public init(
        url: URL,
        targetName: String,
        targetKind: String,
        source: String,
        tree: SourceFileSyntax,
        nodeCount: Int,
        packageRoot: URL? = nil,
    ) {
        self.packageRoot = packageRoot
        self.url = url
        self.targetName = targetName
        self.targetKind = targetKind
        self.source = source
        self.tree = tree
        self.locations = SourceLocationConverter(fileName: url.path, tree: tree)
        self.nodeCount = nodeCount
    }

    /*
     A finding at a node, positioned the way the contract says: 1-based lines, 1-based columns counted in
     UTF-8 bytes, from the first character of the node rather than the trivia before it.
     */
    public func finding(
        at node: some SyntaxProtocol,
        rule: String,
        messageId: String,
        message: String,
        severity: FindingRecord.Severity = .error,
        fixes: [FindingRecord.Edit] = [],
        suggestions: [FindingRecord.Suggestion] = [],
    ) -> FindingRecord {
        let start = node.startLocation(converter: locations)
        let end = node.endLocation(converter: locations)
        return FindingRecord(
            source: .rule,
            file: url.path,
            line: start.line,
            column: start.column,
            endLine: end.line,
            endColumn: end.column,
            severity: severity,
            rule: rule,
            messageId: messageId,
            message: message,
            fixes: fixes,
            suggestions: suggestions,
        )
    }

    /* A finding whose id and text come from the house catalog (`RuleMessages`), so the two cannot be paired wrong. */
    func finding(
        at node: some SyntaxProtocol,
        rule: String,
        message: RuleMessages.Message,
        severity: FindingRecord.Severity = .error,
        fixes: [FindingRecord.Edit] = [],
        suggestions: [FindingRecord.Suggestion] = [],
    ) -> FindingRecord {
        finding(
            at: node,
            rule: rule,
            messageId: message.id,
            message: message.text,
            severity: severity,
            fixes: fixes,
            suggestions: suggestions,
        )
    }
}
