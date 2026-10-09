import CryptoKit
import Foundation
import SwiftSyntax

/*
 What another file of the module can see of this one: every token outside the bodies of functions,
 initializers, deinitializers, accessors and subscripts, hashed without trivia.

 This is the premise Swift's own incremental build stakes its correctness on. The frontend type-checks the
 bodies of the file it is compiling and only the declarations of every other file, so a change confined to
 bodies cannot change what the compiler says about any other file, and the driver recompiles nothing else
 for one. When two texts of a file share a fingerprint, every other file's compiler record still holds.

 Deliberately wider than the driver's own interface hash wherever the two might differ, because too wide
 only means building when it was not needed, and too narrow means a stale verdict. A stored property's
 initializer is kept, since an unannotated property's type is inferred from it; so are default argument
 values, top-level code, attributes, and every comment-free token of a declaration. A body is replaced by a
 marker rather than dropped, so gaining or losing one (a protocol requirement given a default) still changes
 the fingerprint.
 */
enum InterfaceFingerprint {
    static func of(_ tree: SourceFileSyntax) -> String {
        let visitor = Visitor(viewMode: .sourceAccurate)
        visitor.walk(tree)
        return visitor.hasher.finalize().map { String($0 >> 4, radix: 16) + String($0 & 0x0f, radix: 16) }.joined()
    }

    private final class Visitor: SyntaxVisitor {
        var hasher = SHA256()

        private func add(_ text: String) {
            hasher.update(data: Data(text.utf8))
            /* A separator no token contains, so `ab c` and `a bc` hash apart. */
            hasher.update(data: Data([0]))
        }

        override func visit(_ token: TokenSyntax) -> SyntaxVisitorContinueKind {
            add(token.text)
            return .skipChildren
        }

        override func visit(_ node: FunctionDeclSyntax) -> SyntaxVisitorContinueKind {
            walkExcept(node, body: node.body.map(Syntax.init))
        }

        override func visit(_ node: InitializerDeclSyntax) -> SyntaxVisitorContinueKind {
            walkExcept(node, body: node.body.map(Syntax.init))
        }

        override func visit(_ node: DeinitializerDeclSyntax) -> SyntaxVisitorContinueKind {
            walkExcept(node, body: node.body.map(Syntax.init))
        }

        override func visit(_ node: AccessorDeclSyntax) -> SyntaxVisitorContinueKind {
            walkExcept(node, body: node.body.map(Syntax.init))
        }

        /* A subscript's accessors are bodies: its element type is always written. */
        override func visit(_ node: SubscriptDeclSyntax) -> SyntaxVisitorContinueKind {
            walkExcept(node, body: node.accessorBlock.map(Syntax.init))
        }

        /*
         A computed property's getter written without `get` is a body too, and the property's type is always
         written, since Swift cannot infer a computed property's type. `willSet` and `didSet` arrive as
         accessor declarations and are handled above.
         */
        override func visit(_ node: PatternBindingSyntax) -> SyntaxVisitorContinueKind {
            guard let block = node.accessorBlock, node.typeAnnotation != nil, case .getter = block.accessors else {
                return .visitChildren
            }
            return walkExcept(node, body: Syntax(block))
        }

        /* Every child of `node` but `body`, which counts only as present. */
        private func walkExcept(_ node: some SyntaxProtocol, body excluded: Syntax?) -> SyntaxVisitorContinueKind {
            for child in node.children(viewMode: .sourceAccurate) {
                if let excluded, child.id == excluded.id {
                    add("{body}")
                }
                else {
                    walk(child)
                }
            }
            return .skipChildren
        }
    }
}
