"use strict";

const assert = require("node:assert/strict");
const fs = require("node:fs");
const ts = require("typescript");

assert.equal(ts.version, "6.0.3");

const [beforeFile, afterFile] = process.argv.slice(2);
assert(beforeFile && afterFile, "usage: node make-barrel.cjs <before-file> <after-file>");
const file = "src/compiler/_namespaces/ts.ts";
const before = fs.readFileSync(beforeFile);
const after = fs.readFileSync(afterFile);
console.error("Before/after barrel bytes equal:", before.equals(after));
assert(before.equals(after));

const source = ts.createSourceFile(file, before.toString(), ts.ScriptTarget.Latest, true);
assert.equal(source.parseDiagnostics.length, 0);
const exportDeclarations = source.statements.filter(ts.isExportDeclaration);
const stars = exportDeclarations.filter(node => !node.exportClause && !node.isTypeOnly && node.moduleSpecifier);
console.error(`Statements: ${source.statements.length}; export declarations: ${exportDeclarations.length}; unqualified export stars: ${stars.length}`);
const tail = [
    'import * as moduleSpecifiers from "./ts.moduleSpecifiers.js";',
    'export { moduleSpecifiers };',
    'import * as performance from "./ts.performance.js";',
    'export { performance };',
];
const unexpected = [];
for (const [index, node] of source.statements.entries()) {
    const line = source.getLineAndCharacterOfPosition(node.getStart(source)).line + 1;
    const statement = node.getText(source);
    const star = stars.includes(node);
    console.error(`${file}:${line} ${star ? "star" : ts.SyntaxKind[node.kind]} ${statement}`);
    if (index < 74 ? !star || line !== index + 3 : statement !== tail[index - 74] || line !== index + 3) {
        unexpected.push(line);
    }
}
assert.equal(source.statements.length, 78);
assert.equal(exportDeclarations.length, 76);
assert.equal(stars.length, 74);
assert.deepEqual(unexpected, [], "unreviewed statements");
console.log(JSON.stringify(source.statements.map(node => node.getText(source)), null, 2));
