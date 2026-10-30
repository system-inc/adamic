// tsc's isScopeMarker (utilitiesPublic.ts:2133, #6fas8b6): the checker infers node is ExportAssignment
// | ExportDeclaration from checked calls joined by ||, and its one use passes it to some, whose
// predicate parameter is (value: T) => boolean, so nothing narrows by its answer. Both predicates are
// unproven but honest, and the program agrees with Node.
interface Statement {
    readonly kind: string;
}
interface ExportAssignment extends Statement {
    readonly kind: "A";
    readonly expression: string;
}
interface ExportDeclaration extends Statement {
    readonly kind: "D";
    readonly names: string;
}
const assignmentKinds: string[] = ["A"];
const declarationKinds: string[] = ["D"];
// Unproven: the prover can't read a list, so each call is checked.
function isExportAssignment(node: Statement): node is ExportAssignment {
    return assignmentKinds.includes(node.kind);
}
function isExportDeclaration(node: Statement): node is ExportDeclaration {
    return declarationKinds.includes(node.kind);
}
function some<T>(array: readonly T[] | undefined, predicate?: (value: T) => boolean): boolean {
    if (array) {
        if (predicate) {
            for (const v of array) {
                if (predicate(v)) {
                    return true;
                }
            }
        } else {
            return array.length > 0;
        }
    }
    return false;
}
function isScopeMarker(node: Statement) {
    return isExportAssignment(node) || isExportDeclaration(node);
}
function hasScopeMarker(statements: readonly Statement[]): boolean {
    return some(statements, isScopeMarker);
}
console.log(String(hasScopeMarker([{ kind: "V" }, { kind: "F" }])));
console.log(String(hasScopeMarker([{ kind: "V" }, { kind: "D" }])));
console.log(String(hasScopeMarker([{ kind: "A" }])));
