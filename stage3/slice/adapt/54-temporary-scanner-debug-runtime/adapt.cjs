const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.SLICE_TYPESCRIPT || 'typescript');
const tree = path.resolve(process.argv[2]);
if (!fs.existsSync(path.join(tree,'slice.json'))) throw new Error('adaptation 54 requires a declaration slice');
const file = path.join(tree,'src/compiler/debug.ts');
let text = fs.readFileSync(file,'utf8');
const source = ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true);
const edits = [];
function visit(node, inFail) {
    if (ts.isFunctionDeclaration(node)) inFail = node.name?.text === 'fail';
    if (inFail && ts.isDebuggerStatement(node)) edits.push({start:node.getStart(source),end:node.end});
    ts.forEachChild(node, child => visit(child,inFail));
}
visit(source,false);
for (const edit of edits.sort((a,b)=>b.start-a.start)) text = text.slice(0,edit.start)+text.slice(edit.end);
if (edits.length) fs.writeFileSync(file,text);
console.log(JSON.stringify({removedDebugger:edits.length}));
