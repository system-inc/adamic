const fs = require('node:fs'), path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT || 'typescript');
const tree = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree,'slice.json'))) throw new Error('adaptation 56 requires a declaration slice');
let calls = 0, members = 0;
for (const name of ['core.ts','scanner.ts','debug.ts']) {
    const file = path.join(tree,'src/compiler',name);
    let text = fs.readFileSync(file,'utf8');
    const source = ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true), edits = [];
    function visit(node) {
        if (name === 'debug.ts' && ts.isFunctionDeclaration(node) && node.name?.text === 'assert') {
            edits.push({start:node.getFullStart(),end:node.end,text:''}); members++; return;
        }
        if (ts.isExpressionStatement(node) && ts.isCallExpression(node.expression)) {
            const call = node.expression;
            if (call.expression.getText(source) === 'Debug.assert') {
                if (call.arguments.length > 2 || (call.arguments[1] && !ts.isStringLiteral(call.arguments[1]))) throw new Error('unsupported assertion arguments');
                const message = call.arguments[1] ? 'False expression: '+call.arguments[1].text : 'False expression.';
                edits.push({start:node.getStart(source),end:node.end,text:`if (!(${call.arguments[0].getText(source)})) Debug.fail(${JSON.stringify(message)});`}); calls++;
            }
        }
        ts.forEachChild(node,visit);
    }
    visit(source);
    for (const edit of edits.sort((a,b)=>b.start-a.start)) text=text.slice(0,edit.start)+edit.text+text.slice(edit.end);
    if (edits.length) fs.writeFileSync(file,text);
}
console.log(JSON.stringify({calls,members}));
