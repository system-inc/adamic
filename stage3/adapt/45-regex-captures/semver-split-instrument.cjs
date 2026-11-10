'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');
const directory = path.join(process.argv[2], 'built/local');
const changed = [];
for (const name of fs.readdirSync(directory).filter(name => name.endsWith('.js'))) {
    const file = path.join(directory, name);
    // No mutation through a link into the uninstrumented tree.
    assert(!fs.lstatSync(file).isSymbolicLink());
    const text = fs.readFileSync(file, 'utf8');
    if (!text.includes('function parseRange')) continue;
    const source = ts.createSourceFile(file,text,ts.ScriptTarget.Latest,true,ts.ScriptKind.JS);
    const edits = [];
    function walk(node, inRange = false) {
        if (ts.isFunctionDeclaration(node)) inRange = /^parseRange\d*$/.test(node.name?.text || '');
        if (inRange && ts.isCallExpression(node) && ts.isPropertyAccessExpression(node.expression) &&
            node.expression.name.text === 'split' && node.arguments.length === 1 &&
            ts.isRegularExpressionLiteral(node.arguments[0]) && node.arguments[0].getText(source) === '/\\s+/') {
            assert.match(node.expression.expression.getText(source), /^range\d*$/);
            edits.push({start:node.getStart(source), end:node.end,
                before:node.getText(source), after:`globalThis.__semverSplit(${node.expression.expression.getText(source)})`});
        }
        ts.forEachChild(node, child => walk(child, inRange));
    }
    walk(source);
    let result = text;
    for (const edit of edits.reverse()) result = result.slice(0,edit.start)+edit.after+result.slice(edit.end);
    if (edits.length) {fs.writeFileSync(file,result); changed.push({name,edits});}
}
assert(changed.some(row=>row.name==='run.js'), 'built harness must contain the counted split');
console.log(JSON.stringify(changed,null,2));
