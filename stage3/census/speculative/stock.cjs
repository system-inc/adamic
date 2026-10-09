// Independent TypeScript AST evidence for recovered boundary ancestry.
const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const ts = require('typescript');
if (ts.version !== '6.0.3') throw new Error(`expected TypeScript 6.0.3, got ${ts.version}`);
const root = path.resolve(process.argv[2]);
const rows = fs.readFileSync(process.argv[3], 'utf8').trim().split('\n').map(JSON.parse);
const files = [];
const syntaxKinds = {};
for (const [name, code] of Object.entries(ts.SyntaxKind)) {
    if (typeof code === 'number') (syntaxKinds[code] ??= []).push(name);
}
for (const row of rows.slice(1)) {
    const filename = path.resolve(row.file);
    if (!filename.startsWith(root + path.sep)) continue;
    const bytes = fs.readFileSync(filename);
    const text = bytes.toString('utf8');
    const source = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
    const offsets = new Uint32Array(text.length + 1);
    let units = 0;
    let byteOffset = 0;
    for (const character of text) {
        if (character.length === 2) offsets[units + 1] = byteOffset + 3;
        units += character.length;
        byteOffset += Buffer.byteLength(character, 'utf8');
        offsets[units] = byteOffset;
    }
    const offset = position => offsets[position];
    const nodes = [];
    function visit(node, ancestors, ancestorKinds = []) {
        const span = [offset(Math.max(0, node.pos)), offset(node.end)];
        nodes.push({start: span[0], end: span[1], kind: ts.SyntaxKind[node.kind], kind_code: node.kind, ancestors, ancestor_kinds: ancestorKinds});
        ts.forEachChild(node, child => visit(child, [...ancestors, span], [...ancestorKinds, node.kind]));
    }
    visit(source, []);
    files.push({file:path.relative(root,filename), bytes:bytes.length,
                sha256:crypto.createHash('sha256').update(bytes).digest('hex'), nodes});
}
fs.writeFileSync(process.argv[4], JSON.stringify({typescript:ts.version,syntax_kinds:syntaxKinds,files}) + '\n');
