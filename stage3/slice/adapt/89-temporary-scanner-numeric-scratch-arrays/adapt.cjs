const fs = require('node:fs'), path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT || 'typescript');
const root = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(root, 'slice.json')) && process.argv[3] !== '--full-baseline') throw new Error('adaptation 89 requires a declaration slice');
const file = path.join(root, 'src/compiler/core.ts');
const text = fs.readFileSync(file, 'utf8');
const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
const functions = source.statements.filter(n => ts.isFunctionDeclaration(n) && n.name?.text === 'levenshteinWithMax');
if (functions.length !== 1) throw new Error('expected one levenshteinWithMax');
const edits = [];
let constructors = 0, reads = 0;
function visit(node) {
    if (ts.isVariableDeclaration(node) && ['previous','current'].includes(node.name.getText(source))) {
        const init = node.initializer;
        if (!init || !ts.isNewExpression(init) || init.expression.getText(source) !== 'Array' || init.arguments?.length !== 1 || init.arguments[0].getText(source) !== 's2.length + 1') throw new Error('scratch constructor drift');
        if (!init.typeArguments?.length) {
            edits.push({start: init.expression.end, end: init.expression.end, text: '<number>'}); constructors++;
        } else if (init.typeArguments.length !== 1 || init.typeArguments[0].getText(source) !== 'number') throw new Error('scratch element drift');
    }
    if (ts.isElementAccessExpression(node) && ['previous','current'].includes(node.expression.getText(source))) {
        const parent = node.parent;
        const write = ts.isBinaryExpression(parent) && parent.left === node && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken;
        if (!write && !ts.isNonNullExpression(parent)) { edits.push({start: node.end,end: node.end,text:'!'}); reads++; }
    }
    ts.forEachChild(node, visit);
}
visit(functions[0]);
let changed = text;
for (const edit of edits.sort((a,b) => b.start-a.start)) changed = changed.slice(0,edit.start)+edit.text+changed.slice(edit.end);
if (edits.length) fs.writeFileSync(file,changed);
console.log(JSON.stringify({constructors,reads}));
