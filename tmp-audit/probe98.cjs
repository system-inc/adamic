'use strict';

// Correctness's probes of the record guard (verifyRecordUses), run on a tree stage3/apply.sh wrote.
// Each probe edits the before and the after snapshot, runs the head's adapter on a scratch copy and prints
// whether it refused (exit 1, both hashes unchanged) or accepted (exit 0). A probe that observes an absent
// calculatedFlags or jsxFlags and is accepted is a fail-open.
const assert = require('node:assert/strict');
const { createHash } = require('node:crypto');
const { execFileSync, spawnSync } = require('node:child_process');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const ts = require('typescript');

const tree = path.resolve(process.argv[2]);
const adapter = process.argv[3] ? path.resolve(process.argv[3])
    : path.join(__dirname, '../stage3/adapt/99-seats-build22-98-links-classes/adapt.cjs');
const ref = 'refs/stage3/99-seats-build22-98-links-classes';
const snapshot = (side, file) => execFileSync('git', ['-C', tree, 'show', `${ref}-${side}:src/compiler/${file}`],
    { encoding: 'utf8', maxBuffer: 8 * 1024 * 1024 });

function inGetter(text, adapted, middle) {
    const eol = text.includes('\r\n') ? '\r\n' : '\n';
    const allocation = adapted ? 'new NodeLinks()' : 'new (NodeLinks as any)()';
    const expression = `nodeLinks[nodeId] || (nodeLinks[nodeId] = ${allocation})`;
    const anchor = `        return ${expression};`;
    assert.equal(text.split(anchor).length, 2, 'one getNodeLinks return anchor');
    return text.replace(anchor, [`const links: typeof nodeLinks[number] = ${expression};`, ...middle, 'return links;']
        .map(line => '        ' + line).join(eol));
}

function atFunctionStart(text, name, line) {
    const eol = text.includes('\r\n') ? '\r\n' : '\n';
    const source = ts.createSourceFile('checker.ts', text, ts.ScriptTarget.Latest, true);
    let found;
    (function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === name) found = node;
        ts.forEachChild(node, visit);
    })(source);
    assert(found, name);
    const at = found.body.statements[0].getStart(source);
    return text.slice(0, at) + line + eol + '        ' + text.slice(at);
}

const observe = 'if (Object.keys(leak()).length !== 1) throw new Error("unreviewed fresh links observation");';
const probes = [
    ['P1 arrow block body returns the record inside getNodeLinks', (t, a) => inGetter(t, a, ['const leak = () => { return links; };', observe])],
    ['P2 anonymous function expression returns the record inside getNodeLinks', (t, a) => inGetter(t, a, ['const leak = function () { return links; };', observe])],
    ['P3 arguments aliases the import helper links parameter', (t) => atFunctionStart(t, 'resolveImportSymbolType',
        'if (Object.keys(arguments[1]).length !== 1) throw new Error("unreviewed fresh links observation");')],
    ['C1 concise arrow body (control, expect refusal)', (t, a) => inGetter(t, a, ['const leak = () => links;', observe])],
    ['C2 default parameter (control, expect refusal)', (t, a) => inGetter(t, a, ['const leak = (x = links) => x;', observe])],
    ['C3 record stored in another object (control, expect refusal)', (t, a) => inGetter(t, a, ['const holder: any = {};', 'holder.x = links;', 'if (Object.keys(holder.x).length !== 1) throw new Error("unreviewed fresh links observation");'])],
    ['C4 reassignment into a let (control, expect refusal)', (t, a) => inGetter(t, a, ['let other: any;', 'other = links;', 'if (Object.keys(other).length !== 1) throw new Error("unreviewed fresh links observation");'])],
    ['C5 named nested function returns the record (control, expect refusal)', (t, a) => inGetter(t, a, ['function leak() { return links; }', observe])],
    ['C6 accessor alias (control, expect refusal)', (t, a) => inGetter(t, a, ['const leak = () => getLinks(node); const getLinks = getNodeLinks;', observe])],
];

const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'probe98-'));
const compiler = path.join(scratch, 'src/compiler');
fs.mkdirSync(compiler, { recursive: true });
const checker = path.join(compiler, 'checker.ts');
const types = path.join(compiler, 'types.ts');
const hashes = () => [checker, types].map(file => createHash('sha256').update(fs.readFileSync(file)).digest('hex').slice(0, 12));
try {
    for (const side of ['before', 'after']) {
        const text = snapshot(side, 'checker.ts');
        const typeText = snapshot(side, 'types.ts');
        for (const [label, edit] of probes) {
            fs.writeFileSync(checker, edit(text, side === 'after'));
            fs.writeFileSync(types, typeText);
            const before = hashes();
            const result = spawnSync(process.execPath, [adapter, scratch], { encoding: 'utf8' });
            const after = hashes();
            const why = result.status === 0 ? result.stdout.trim()
                : (result.stderr.split('\n').find(line => line.startsWith('AssertionError')) || result.stderr.trim().split('\n')[0]);
            const verdict = result.status === 0 ? 'ACCEPTED' : 'refused';
            console.log(`PROBE ${side} ${label}: ${verdict} exit ${result.status} ${before.join(' ')} -> ${after.join(' ')} ${why}`);
        }
    }
} finally {
    fs.rmSync(scratch, { recursive: true, force: true });
}
