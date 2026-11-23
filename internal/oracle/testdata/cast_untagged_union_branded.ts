// tsc's utilitiesPublic.ts:960 (#axqgdck head 2): Declaration and Expression add only a brand to Node, so each is a
// phantom extension, held as its base, and a cast from their union to Identifier is the untagged union's checked
// view. The values are made by brand mints. An Identifier reads its text; a Declaration reads what it shares.
interface NodeBase { readonly kind: number; readonly pos: number }
interface Declaration extends NodeBase { readonly _declarationBrand: undefined }
interface Expression extends NodeBase { readonly _expressionBrand: undefined }
interface IdentifierFields extends NodeBase { readonly escapedText: string }
interface Identifier extends IdentifierFields { readonly _expressionBrand: undefined }
function textOf(declaration: Declaration | Expression): string {
    return (declaration as Identifier).escapedText;
}
function positionOf(declaration: Declaration | Expression): number {
    return (declaration as Identifier).pos;
}
const fields: IdentifierFields = { kind: 80, pos: 4, escapedText: 'name' };
const identifier = fields as unknown as Identifier;
console.log(textOf(identifier));
const base: NodeBase = { kind: 262, pos: 9 };
const declaration = base as unknown as Declaration;
console.log(String(positionOf(declaration)));
