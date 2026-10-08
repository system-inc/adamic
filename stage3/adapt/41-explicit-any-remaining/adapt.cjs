'use strict';
const fs = require('node:fs'), path = require('node:path'), ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('requires stock TypeScript 6.0.3');
const tree = path.resolve(process.argv[2]);
const rules = require('./rules.json');
const plans = [];
for (const file of [...new Set(rules.map(r => r.file))]) {
    const name = path.join(tree, file), before = fs.readFileSync(name, 'utf8');
    let text = before, removed = 0;
    for (const r of rules.filter(r => r.file === file)) {
        const newline = text.includes('\r\n') ? '\r\n' : '\n';
        const before = r.before.replace(/\r?\n/g, newline), after = r.after.replace(/\r?\n/g, newline);
        const oldCount = text.split(before).length - 1;
        const newCount = text.split(after).length - 1;
        if (newCount === 1 && text.replace(after, '').split(before).length === 1) continue;
        if (oldCount !== 1 || newCount !== 0) throw Error('missing or duplicate reviewed site: ' + r.id);
        text = text.replace(before, after); removed++;
    }
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    if (source.parseDiagnostics.length) throw Error('adapted syntax rejected: ' + file);
    plans.push({name, before, text, removed});
}
for (const p of plans) if (fs.readFileSync(p.name, 'utf8') !== p.before) throw Error('concurrent source edit');
for (const p of plans) if (p.text !== p.before) fs.writeFileSync(p.name, p.text);
console.log(JSON.stringify({adaptation: 41, sites: plans.reduce((n,p) => n + p.removed, 0)}));
require('./void.cjs').apply(tree);
