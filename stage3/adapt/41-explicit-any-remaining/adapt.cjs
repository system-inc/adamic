'use strict';
const fs = require('node:fs'), path = require('node:path'), ts = require('typescript');
if (ts.version !== '6.0.3') throw Error('requires stock TypeScript 6.0.3');
const tree = path.resolve(process.argv[2]);
const rules = require('./rules.json');
const plans = [];
for (const file of [...new Set(rules.map(r => r.file))]) {
    const name = path.join(tree, file), before = fs.readFileSync(name, 'utf8');
    let text = before, removed = 0;
    // Empty replacements delete complete declaration lines, atomically by group.
    // Never use split("") to recognize an already deleted owner.
    const deletionStates = new Map();
    const deletions = rules.filter(r => r.file === file && r.after === '');
    if (deletions.length) {
        const source = ts.createSourceFile(file, before, ts.ScriptTarget.Latest, true);
        if (source.parseDiagnostics.length) throw Error('deletion source syntax rejected: ' + file);
        for (const r of deletions) {
            if (!r.before || !r.deleteGroup || !r.declaration) throw Error('unbound deletion rule: ' + r.id);
            const bindings = source.statements.filter(n => n.name?.text === r.declaration ||
                (ts.isVariableStatement(n) && n.declarationList.declarations.some(d => d.name?.text === r.declaration)) ||
                (ts.isImportDeclaration(n) && (n.importClause?.name?.text === r.declaration ||
                    n.importClause?.namedBindings?.elements?.some(e => e.name.text === r.declaration))));
            const count = before.split(r.before).length - 1;
            if (bindings.length === 0 && count === 0) deletionStates.set(r.id, true);
            else if (bindings.length === 1 && count === 1 && ts.isFunctionDeclaration(bindings[0]) &&
                bindings[0].getText(source) === r.before) deletionStates.set(r.id, false);
            else throw Error('changed or duplicate deletion owner: ' + r.id);
        }
        for (const group of new Set(deletions.map(r => r.deleteGroup))) {
            const rows = deletions.filter(r => r.deleteGroup === group);
            if (rows.length < 2 || !rows.every(r => deletionStates.get(r.id) === deletionStates.get(rows[0].id)))
                throw Error('half-applied deletion group: ' + group);
        }
    }

    for (const r of rules.filter(r => r.file === file)) {
        const newline = text.includes('\r\n') ? '\r\n' : '\n';
        const before = r.before.replace(/\r?\n/g, newline), after = r.after.replace(/\r?\n/g, newline);
        if (r.after === '') {
            if (deletionStates.get(r.id)) continue;
            const line = before + newline;
            if (text.split(line).length !== 2) throw Error('missing complete deletion line: ' + r.id);
            text = text.replace(line, ''); removed++;
            continue;
        }

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
