// tsc's getDeclaredExpandoInitializer (utilities.ts:3950, #6fas8b6): a checked predicate's answer is
// an argument to a call under &&. The checker narrows nothing through an argument, so the call is read
// wherever it stands. isPrototypeAccess is unproven but honest, and the program agrees with Node.
interface Identifier {
    readonly kind: "I";
    readonly text: string;
}
interface PrototypeAccess {
    readonly kind: "P";
    readonly property: string;
}
type Name = Identifier | PrototypeAccess;
interface Expression {
    readonly value: number;
}
interface Declaration {
    readonly name: Name;
    readonly initializer: Expression | undefined;
}
const prototypeKinds: string[] = ["P"];
// Unproven: the prover can't read a list, so its call is checked.
function isPrototypeAccess(node: Name): node is PrototypeAccess {
    return prototypeKinds.includes(node.kind);
}
function getExpandoInitializer(initializer: Expression, isPrototypeAssignment: boolean): Expression | undefined {
    return isPrototypeAssignment ? initializer : undefined;
}
function getDeclaredExpandoInitializer(node: Declaration): Expression | undefined {
    const init = node.initializer;
    return init && getExpandoInitializer(init, isPrototypeAccess(node.name));
}
const declarations: Declaration[] = [
    { name: { kind: "I", text: "a" }, initializer: { value: 1 } },
    { name: { kind: "P", property: "prototype" }, initializer: { value: 2 } },
    { name: { kind: "P", property: "prototype" }, initializer: undefined },
];
for (const declaration of declarations) {
    const found = getDeclaredExpandoInitializer(declaration);
    console.log(found === undefined ? "none" : String(found.value));
}
