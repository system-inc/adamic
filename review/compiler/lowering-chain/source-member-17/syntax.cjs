// Count executable compiler-source AST, excluding comments and helper template text.
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const source = process.argv[2];
if (!source) throw new Error('provide pinned TypeScript src/compiler');
const rows = [];
function scan(directory) {
    for (const entry of fs.readdirSync(directory, {withFileTypes: true}).sort((a, b) => a.name.localeCompare(b.name))) {
        const file = path.join(directory, entry.name);
        if (entry.isDirectory()) { scan(file); continue; }
        if (!entry.name.endsWith('.ts')) continue;
        const parsed = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
        function visit(node) {
            let shape;
            if (ts.isTryStatement(node)) {
                shape = node.catchClause ? node.finallyBlock ? 'try/catch/finally' : 'try/catch' : 'try/finally';
            } else if (ts.isThrowStatement(node)) {
                shape = ts.isNewExpression(node.expression) && node.expression.expression.getText(parsed) === 'Error' ? 'throw new Error' : ts.isIdentifier(node.expression) ? 'throw identifier' : 'throw other';
            } else if (ts.isClassDeclaration(node) && node.heritageClauses?.some(clause => clause.types.some(type => type.expression.getText(parsed) === 'Error'))) {
                shape = 'Error subclass';
            }
            if (shape) rows.push({shape, file: path.relative(source, file), line: parsed.getLineAndCharacterOfPosition(node.getStart(parsed)).line + 1});
            ts.forEachChild(node, visit);
        }
        visit(parsed);
    }
}
scan(source);
const counts = {};
for (const row of rows) counts[row.shape] = (counts[row.shape] || 0) + 1;
const expected = {'try/catch': 26, 'try/finally': 8, 'throw new Error': 9, 'throw identifier': 3};
for (const [shape, count] of Object.entries(expected)) {
    if (counts[shape] !== count) throw new Error(`${shape}: expected ${count}, got ${counts[shape]}`);
}
if (Object.keys(counts).length !== Object.keys(expected).length) throw new Error('unexpected shape');
console.log(JSON.stringify({pin: '050880ce59e30b356b686bd3144efe24f875ebc8', parser: ts.version, counts, rows}, null, 2));
