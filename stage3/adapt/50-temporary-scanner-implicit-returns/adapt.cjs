const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const tree = path.resolve(process.argv[2]);
const file = path.join(tree, 'src/compiler/scanner.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const names = new Set(['getShebang', 'scanIdentifier']);
const found = new Set();
const edits = [];
const newline = text.includes('\r\n') ? '\r\n' : '\n';
function visit(node) {
    if (ts.isFunctionDeclaration(node) && node.name && names.has(node.name.text)) {
        if (found.has(node.name.text) || !node.body) throw new Error('unexpected scanner function: ' + node.name.text);
        found.add(node.name.text);
        const last = node.body.statements.at(-1);
        if (!(last && ts.isReturnStatement(last) && last.expression && ts.isIdentifier(last.expression) && last.expression.text === 'undefined')) {
            const close = node.body.end - 1;
            const lineStart = text.lastIndexOf('\n', close) + 1;
            const indentation = text.slice(lineStart, close);
            if (!/^ *$/.test(indentation)) throw new Error('unexpected scanner closing brace layout');
            edits.push({start: lineStart, text: indentation + '    return undefined;' + newline});
        }
    }
    ts.forEachChild(node, visit);
}
visit(source);
if (found.size !== names.size) throw new Error('missing scanner optional-return function');
let adapted = text;
for (const edit of edits.sort((a, b) => b.start - a.start)) adapted = adapted.slice(0, edit.start) + edit.text + adapted.slice(edit.start);
if (adapted !== text) fs.writeFileSync(file, adapted);
console.log(JSON.stringify({files: edits.length ? 1 : 0, explicitUndefinedReturns: edits.length, declined: []}));
