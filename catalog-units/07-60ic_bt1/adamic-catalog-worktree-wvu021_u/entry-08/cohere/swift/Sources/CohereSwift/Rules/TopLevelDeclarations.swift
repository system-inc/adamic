import SwiftSyntax

/*
 The types a file declares at top level and the types it extends, in order: what the file-shape rules
 judge a file by.

 Declarations inside a top-level `#if` count, because a type behind a platform condition is still a type
 in this file. A name declared in two branches (`#if os(macOS) struct Pane … #else struct Pane …`) is
 one type, not two, so it is counted once.

 A `typealias` is not a type here. It names a type that is defined somewhere else, so it never makes a
 second definition in a file.
 */
struct TopLevelDeclarations {
    /* One declaration, with the token to report at. */
    struct Entry {
        var name: String
        var token: TokenSyntax
        /* Set for extensions, so a rule can read the extension's access level and members. */
        var extensionDeclaration: ExtensionDeclSyntax? = nil
    }

    var types: [Entry] = []
    var extensions: [Entry] = []

    /*
     The type the file is for: the one its name matches, else the first. Choosing by name keeps the advice
     pointed the right way. In `SidebarPositionContextMenu.swift`, a helper declared above
     `SidebarPositionContextMenu` must be the type told to move, not the file told to be renamed after it.
     */
    func primary(stem: String) -> Entry? {
        types.first { $0.name == stem } ?? types.first
    }

    init(_ tree: SourceFileSyntax) {
        var seenTypes = Set<String>()
        for statement in tree.statements {
            collect(statement.item, seenTypes: &seenTypes)
        }
    }

    private mutating func collect(_ item: CodeBlockItemSyntax.Item, seenTypes: inout Set<String>) {
        if let conditionalCompilation = item.as(IfConfigDeclSyntax.self) {
            for clause in conditionalCompilation.clauses {
                guard let elements = clause.elements?.as(CodeBlockItemListSyntax.self) else { continue }
                for element in elements {
                    collect(element.item, seenTypes: &seenTypes)
                }
            }
            return
        }
        if let named = Self.typeName(of: item) {
            if seenTypes.insert(named.text).inserted {
                types.append(Entry(name: named.text, token: named))
            }
            return
        }
        if let extensionDeclaration = item.as(ExtensionDeclSyntax.self) {
            extensions.append(
                Entry(
                    name: Self.baseName(of: extensionDeclaration.extendedType),
                    token: extensionDeclaration.extensionKeyword,
                    extensionDeclaration: extensionDeclaration,
                )
            )
        }
    }

    private static func typeName(of item: CodeBlockItemSyntax.Item) -> TokenSyntax? {
        if let declaration = item.as(StructDeclSyntax.self) { return declaration.name }
        if let declaration = item.as(ClassDeclSyntax.self) { return declaration.name }
        if let declaration = item.as(EnumDeclSyntax.self) { return declaration.name }
        if let declaration = item.as(ActorDeclSyntax.self) { return declaration.name }
        if let declaration = item.as(ProtocolDeclSyntax.self) { return declaration.name }
        return nil
    }

    /* The extended type's own name, without generic arguments: `Array<Int>` and `Array where …` both extend `Array`; `Outer.Inner` keeps its dots. */
    static func baseName(of type: TypeSyntax) -> String {
        if let identifier = type.as(IdentifierTypeSyntax.self) {
            return identifier.name.text
        }
        if let member = type.as(MemberTypeSyntax.self) {
            return baseName(of: member.baseType) + "." + member.name.text
        }
        return type.trimmedDescription
    }
}
