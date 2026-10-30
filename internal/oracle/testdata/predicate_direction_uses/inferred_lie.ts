// inferred_some.ts with a lying inner predicate (#6fas8b6): isExportAssignment says yes to a V. The
// checker infers isScopeMarker's predicate from its calls, and its true answer is theirs, so the lie
// stops at isExportAssignment's own true check, with its located message, on both backends. Node
// says true.
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
const assignmentKinds: string[] = ["A", "V"];
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
console.log(String(hasScopeMarker([{ kind: "V" }])));
