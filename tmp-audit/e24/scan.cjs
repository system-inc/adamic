"use strict";

// Every element access in the adapted compiler project whose key's type is a type parameter, with the receiver's type
// and the key's constraint, so the count of #e24qsqx's shape reads from the checker, not from spelling.
const path = require("node:path");
const ts = require("typescript");

const tree = path.resolve(process.argv[2]);
const config = path.join(tree, "src/compiler/tsconfig.json");
const parsed = ts.getParsedCommandLineOfConfigFile(config, {}, { ...ts.sys, onUnRecoverableConfigFileDiagnostic: d => { throw new Error(ts.flattenDiagnosticMessageText(d.messageText, "\n")); } });
const program = ts.createProgram({ rootNames: parsed.fileNames, options: parsed.options, projectReferences: parsed.projectReferences });
const checker = program.getTypeChecker();
const rows = [];
let accesses = 0;
for (const source of program.getSourceFiles()) {
    if (!source.fileName.startsWith(path.join(tree, "src/compiler/"))) continue;
    (function visit(node) {
        if (ts.isElementAccessExpression(node)) {
            accesses++;
            const keyType = checker.getTypeAtLocation(node.argumentExpression);
            if (keyType.flags & ts.TypeFlags.TypeParameter) {
                const receiver = checker.getTypeAtLocation(node.expression);
                const constraint = checker.getBaseConstraintOfType(keyType);
                const generic = !!(receiver.flags & (ts.TypeFlags.TypeParameter | ts.TypeFlags.IndexedAccess)) ||
                    checker.typeToString(receiver).includes(checker.typeToString(keyType));
                const owner = (() => { for (let p = node.parent; p; p = p.parent) if (ts.isFunctionLike(p) && p.name) return p.name.getText(source); })();
                const { line, character } = source.getLineAndCharacterOfPosition(node.getStart(source));
                rows.push({
                    where: `${path.relative(tree, source.fileName)}:${line + 1}:${character + 1}`,
                    owner, generic, write: ts.isBinaryExpression(node.parent) && node.parent.left === node && node.parent.operatorToken.kind === ts.SyntaxKind.EqualsToken,
                    receiver: checker.typeToString(receiver).slice(0, 60),
                    key: checker.typeToString(keyType), constraint: constraint ? checker.typeToString(constraint).slice(0, 60) : "(none)",
                    text: node.getText(source).slice(0, 70),
                });
            }
        }
        ts.forEachChild(node, visit);
    })(source);
}
const fixed = rows.filter(row => !row.generic);
console.log(`SCAN compiler project: ${accesses} element accesses, ${rows.length} keyed by a type parameter, ${fixed.length} of them on a receiver that isn't itself generic in the key`);
for (const row of rows) {
    console.log(["SCAN", row.generic ? "generic-receiver" : "fixed-receiver", row.write ? "write" : "read", row.where, row.owner ?? "-", `key ${row.key} extends ${row.constraint}`, `receiver ${row.receiver}`, row.text].join("\t"));
}
