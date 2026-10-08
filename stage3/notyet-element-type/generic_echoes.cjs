// Independent stock-TypeScript audit of generic scopes at the census sites.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require('typescript');
const [input, adapted, output] = process.argv.slice(2);
if (!input || !adapted || !output) throw new Error('usage: generic_echoes.cjs INPUT_JSON ADAPTED OUTPUT_JSON');
const expected = new Map([['an array of T', 42], ['an array of U', 6], ['an array of NonNullable<T>', 3], ['an array of Child', 2], ['an array of V', 2], ['an array of TState', 1]]);
const selected = new Map();
for (const row of JSON.parse(fs.readFileSync(input, 'utf8'))) {
    if (expected.has(row.reason)) selected.set(row.reason + '\0' + row.where, row);
}
const counts = new Map();
const files = new Map();
const sites = [];
for (const row of selected.values()) {
    const match = /^(.*):(\d+):(\d+)$/.exec(row.where);
    if (!match) throw new Error('invalid location: ' + row.where);
    const filename = match[1];
    if (!files.has(filename)) {
        const source = fs.readFileSync(path.join(adapted, filename), 'utf8');
        files.set(filename, { file: ts.createSourceFile(filename, source, ts.ScriptTarget.Latest, true), sha256: crypto.createHash('sha256').update(source).digest('hex') });
    }
    const { file, sha256 } = files.get(filename);
    const position = file.getPositionOfLineAndCharacter(Number(match[2]) - 1, Number(match[3]) - 1);
    let node = file;
    const descend = current => {
        if (current.getStart(file) <= position && position < current.end) {
            node = current;
            ts.forEachChild(current, descend);
        }
    };
    descend(file);
    const scopes = [];
    for (let parent = node; parent; parent = parent.parent) {
        if (parent.typeParameters?.length) {
            scopes.push({ kind: ts.SyntaxKind[parent.kind], name: parent.name?.getText(file) || '<anonymous>', parameters: parent.typeParameters.map(parameter => parameter.name.text) });
        }
    }
    const wanted = row.reason === 'an array of NonNullable<T>' ? 'T' : row.reason.slice('an array of '.length);
    if (process.env.GENERIC_ECHO_MUTANT === 'drop-parameters') for (const scope of scopes) scope.parameters = [];
    if (!scopes.some(scope => scope.parameters.includes(wanted))) throw new Error('missing generic owner for ' + row.where + ': ' + wanted);
    counts.set(row.reason, (counts.get(row.reason) || 0) + 1);
    sites.push({ where: row.where, reason: row.reason, attempted_unit: row.unit, source_sha256: sha256, generic_scopes: scopes });
}
for (const [reason, count] of expected) if (counts.get(reason) !== count) throw new Error('site count mismatch for ' + reason);
fs.writeFileSync(output, JSON.stringify({ stock_typescript: ts.version, explanation: 'All selected sites refer to a declared type parameter in an enclosing generic scope. The census attempts isolated declarations without invented specializations. This cancels these elementType storage observations as generic census echoes; it does not certify the enclosing compiler functions or their other stops.', counts: Object.fromEntries(counts), sites }, null, 2) + '\n');
console.log('generic-scope audit passed: ' + sites.length + ' unique sites');
