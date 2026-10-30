// tsc's isEffectiveModuleDeclaration (utilities.ts:2098, #6fas8b6): the checker infers node is
// ModuleDeclaration | Identifier from checked calls joined by ||, and its one call, setValueDeclaration's
// (utilities.ts:4313), reads no false side: valueDeclaration is a Declaration, which the false side
// leaves whole. Both predicates are unproven but honest, and the program agrees with Node.
interface Node {
    readonly kind: string;
}
interface Declaration extends Node {
    readonly name: string;
}
interface ModuleDeclaration extends Declaration {
    readonly kind: "M";
    readonly body: string;
}
interface Identifier extends Declaration {
    readonly kind: "I";
    readonly text: string;
}
interface Symbol {
    valueDeclaration: Declaration | undefined;
}
const moduleKinds: string[] = ["M"];
const identifierKinds: string[] = ["I"];
// Unproven: the prover can't read a list, so each call is checked.
function isModuleDeclaration(node: Node): node is ModuleDeclaration {
    return moduleKinds.includes(node.kind);
}
function isIdentifier(node: Node): node is Identifier {
    return identifierKinds.includes(node.kind);
}
function isEffectiveModuleDeclaration(node: Node) {
    return isModuleDeclaration(node) || isIdentifier(node);
}
function setValueDeclaration(symbol: Symbol, node: Declaration): void {
    const valueDeclaration = symbol.valueDeclaration;
    if (!valueDeclaration || (valueDeclaration.kind !== node.kind && isEffectiveModuleDeclaration(valueDeclaration))) {
        symbol.valueDeclaration = node;
    }
}
const module: Declaration = { kind: "M", name: "m" };
const variable: Declaration = { kind: "V", name: "v" };
const other: Declaration = { kind: "F", name: "f" };
const symbol: Symbol = { valueDeclaration: undefined };
for (const declaration of [module, variable, other]) {
    setValueDeclaration(symbol, declaration);
    console.log(symbol.valueDeclaration === undefined ? "none" : symbol.valueDeclaration.name);
}
