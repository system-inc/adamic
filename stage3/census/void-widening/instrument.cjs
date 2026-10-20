'use strict';
// Disposable source tree only. Preserve each observed expression's value.
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require('../../api/node_modules/typescript');
const [tree, table] = process.argv.slice(2), data = JSON.parse(fs.readFileSync(table));
const sites = data.sites || data;
const selected = sites.map((s, id) => ({...s, id})).filter(s =>
    s.file === 'src/compiler/factory/nodeFactory.ts' && s.line === 1164 ||
    s.file === 'src/compiler/utilities.ts' && s.line === 8582);
assert(selected.length >= 2);
for (const name of [...new Set(selected.map(s => s.file))]) {
    const file = path.join(tree, name), text = fs.readFileSync(file, 'utf8');
    const sf = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    const events = new Map();
    for (const s of selected.filter(s => s.file === name)) {
        const at = sf.getPositionOfLineAndCharacter(s.line - 1, s.column - 1);
        let target;
        function visit(n) { if (!target && n.getStart(sf) === at && ts.isExpressionNode(n)) target = n; ts.forEachChild(n, visit); }
        visit(sf); assert(target);
        // A callable relation's concise body is the runtime value being read.
        if (ts.isArrowFunction(target)) target = target.body;
        assert(ts.isCallExpression(target));
        const start = target.getStart(sf), end = target.end;
        // Multiple compiler refusal forms may name the same actual read.
        const key = start + ':' + end;
        if (!events.has(key)) events.set(key, {start, end, ids: []});
        events.get(key).ids.push(s.id);
    }
    let changed = text;
    for (const e of [...events.values()].sort((a,b) => b.start - a.start)) {
        let expression = text.slice(e.start, e.end);
        for (const id of e.ids) expression = `globalThis.__adamicVoidRead(${id}, ${expression})`;
        changed = changed.slice(0, e.start) + expression + changed.slice(e.end);
    }
    fs.writeFileSync(file, 'declare global { var __adamicVoidRead: <T>(id: number, value: T) => T; }\n' + changed);
}
console.log(JSON.stringify({instrumentedSites: selected.map(s => ({id:s.id,file:s.file,line:s.line,column:s.column}))}));
