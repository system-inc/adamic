// declarations 1808's shape, lying: the second computed name's expression is no entity name, so the read of
// name.expression.kind through the view stops there, after the honest call has printed what Node prints.
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
const entity: Node = { kind: Kind.Name, parent: undefined };
const other: Node = { kind: Kind.Other, parent: undefined };
const honest: Name = { kind: Kind.Computed, parent: undefined, expression: entity };
const lyingName: Name = { kind: Kind.Computed, parent: undefined, expression: other };
const declaration: Declaration = { kind: Kind.Prop, parent: undefined, name: honest };
const lying: Declaration = { kind: Kind.Prop, parent: undefined, name: lyingName };
console.log(`${late(declaration)}`);
console.log(`${late(lying)}`);
