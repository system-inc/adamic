const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT || 'typescript');
const tree = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree, 'slice.json'))) throw new Error('adaptation 52 requires a declaration slice');
const targets = new Map([
    ['core.ts', new Map([['forEach', new Set(['array'])], ['levenshteinWithMax', new Set(['s1', 's2'])]])],
    ['scanner.ts', new Map([['lookupInUnicodeMap', new Set(['map'])]])],
    ['utilities.ts', new Map([['parsePseudoBigInt', new Set(['segments'])]])],
]);
let changed = 0;
for (const [name, functions] of targets) {
    const file = path.join(tree, 'src/compiler', name);
    let text = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const edits = [];
    function visit(node, active) {
        if (ts.isFunctionDeclaration(node)) {
            active = functions.get(node.name?.text);
            if (name === 'core.ts' && node.name?.text === 'forEach') {
                const original = node.getText(source);
                const revised = original.replace('for (let i = 0; i < array.length; i++)', 'for (const [i, element] of array.entries())')
                    .replace('callback(array[i], i)', 'callback(element, i)');
                if (revised !== original) edits.push({start:node.getStart(source),end:node.end,text:revised});
                return;
            }
        }
        if (active && ts.isBinaryExpression(node) && node.operatorToken.kind === ts.SyntaxKind.BarEqualsToken
            && ts.isElementAccessExpression(ts.isNonNullExpression(node.left) ? node.left.expression : node.left)
            && active.has((ts.isNonNullExpression(node.left) ? node.left.expression : node.left).expression.getText(source))) {
            const left = node.left.getText(source), right = node.right.getText(source);
            edits.push({start: node.getStart(source), end: node.end, text: `${left} = (${left} ?? Debug.fail("stage3: missing dense element")) | ${right}`});
            return;
        }
        if (active && ts.isElementAccessExpression(node) && active.has(node.expression.getText(source))) {
            const parent = node.parent;
            const write = ts.isBinaryExpression(parent) && parent.left === node && parent.operatorToken.kind === ts.SyntaxKind.EqualsToken;
            const checked = ts.isBinaryExpression(parent) && parent.left === node && parent.operatorToken.kind === ts.SyntaxKind.QuestionQuestionToken;
            if (!write && !checked) edits.push({start: node.getStart(source), end: node.end,
                text: `(${node.getText(source)} ?? Debug.fail("stage3: missing dense element"))`});
            return;
        }
        ts.forEachChild(node, child => visit(child, active));
    }
    visit(source);
    for (const edit of edits.sort((a,b) => b.start-a.start)) text = text.slice(0,edit.start)+edit.text+text.slice(edit.end);
    if (edits.length && name === 'utilities.ts' && !text.includes('import { Debug }')) text = 'import { Debug } from "./debug.ts";\r\n'+text;
    if (edits.length) { fs.writeFileSync(file,text); changed += edits.length; }
}
console.log(JSON.stringify({checkedReads: changed}));
