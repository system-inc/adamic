'use strict';
const crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const path = require('node:path');
// Hash runtime syntax, excluding annotations and generic binders. A changed body
// requires a new review; changing types must not disguise a newly added writer.
function fingerprint(node) {
    function visit(n) {
        if (ts.isTypeNode(n) || ts.isTypeParameterDeclaration(n) || ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n)) return undefined;
        const children = [];
        ts.forEachChild(n, c => { const v = visit(c); if (v !== undefined) children.push(v); });
        return [ts.SyntaxKind[n.kind], ts.isIdentifier(n) || ts.isPrivateIdentifier(n) || ts.isLiteralExpression(n) ? n.text : null, children];
    }
    return crypto.createHash('sha256').update(JSON.stringify(visit(node))).digest('hex');
}
function find(program, tree, owner) {
    const sf = program.getSourceFile(path.join(tree, owner.file));
    const matches = [];
    function visit(n) {
        if (ts.isFunctionDeclaration(n) && n.name?.text === owner.function && n.body) matches.push(n);
        ts.forEachChild(n, visit);
    }
    visit(sf);
    if (matches.length !== 1) throw new Error(`expected one body: ${owner.function}`);
    return matches[0];
}
module.exports = function plans(program, tree, specifications, edits) {
    const records = [];
    for (const spec of specifications) {
        for (const owner of spec.owners) {
            const fn = find(program, tree, owner);
            if (fingerprint(fn.body) !== owner.runtime_sha256) throw new Error(`reviewed runtime body changed: ${owner.function}`);
            records.push({ family: spec.family, ...owner });
        }
        for (const edit of spec.edits) {
            const file = program.getSourceFile(path.join(tree, edit.file));
            const owner = edit.owner ? find(program, tree, { file: edit.file, function: edit.owner }) : undefined;
            const begin = owner ? owner.getStart(file) : 0;
            const text = file.text.slice(begin, owner ? owner.end : file.text.length);
            const count = needle => text.split(needle).length - 1;
            // Some shorter original signatures are contained in their adapted form.
            if (count(edit.after) === 1) continue;
            if (count(edit.before) !== 1) throw new Error(`type edit shape changed: ${spec.family}: ${edit.before}`);
            if (!edits.has(file)) edits.set(file, []);
            const start = begin + text.indexOf(edit.before);
            edits.get(file).push({ start, end: start + edit.before.length, text: edit.after });
        }
    }
    return records;
};
module.exports.fingerprint = fingerprint;
module.exports.find = find;
