// declarations 1808: a late-bound declaration, whose name is a computed name, whose expression is an entity
// name, each a member slot. name.expression.kind is checked at the read, two levels down.
enum Kind { Root, Doc, Tag, Param, Prop, Name, Computed, Literal, Other }
interface Node { readonly kind: Kind; readonly parent: Node | undefined; }
interface Name extends Node { readonly expression: Node | undefined; }
interface Declaration extends Node { readonly name: Name | undefined; }
interface EntityName extends Node { readonly kind: Kind.Name; }
interface ComputedName extends Name { readonly kind: Kind.Computed; readonly expression: EntityName; }
interface LateBound extends Declaration { readonly name: ComputedName; }
function late(node: Declaration): Kind {
    const declaration = node as LateBound;
    return declaration.name.expression.kind;
}
const expression: Node = { kind: Kind.Name, parent: undefined };
const computed: Name = { kind: Kind.Computed, parent: undefined, expression };
const declaration: Declaration = { kind: Kind.Prop, parent: undefined, name: computed };
console.log(`${late(declaration)}`);
