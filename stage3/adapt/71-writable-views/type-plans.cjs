'use strict';
const crypto = require('node:crypto');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
const path = require('node:path');
// Hash runtime syntax, excluding annotations and generic binders. adapt.cjs
// compares proposed files with the input left by earlier numbered adaptations.
function fingerprint(node) {
    function visit(n) {
        if ((ts.isImportSpecifier(n) || ts.isExportSpecifier(n)) && n.isTypeOnly) return undefined;
        if (ts.isImportDeclaration(n) && n.importClause?.isTypeOnly || ts.isExportDeclaration(n) && n.isTypeOnly) return undefined;
        if (ts.isAsExpression(n) || ts.isTypeAssertionExpression(n) || ts.isSatisfiesExpression(n) || ts.isNonNullExpression(n)) return visit(n.expression);
        if (ts.isTypeNode(n) || ts.isTypeParameterDeclaration(n) || ts.isInterfaceDeclaration(n) || ts.isTypeAliasDeclaration(n)) return undefined;
        const children = [];
        ts.forEachChild(n, c => { const v = visit(c); if (v !== undefined) children.push(v); });
        return [ts.SyntaxKind[n.kind], ts.isIdentifier(n) || ts.isPrivateIdentifier(n) || ts.isLiteralExpression(n) ? n.text : null, children];
    }
    return crypto.createHash('sha256').update(JSON.stringify(visit(node))).digest('hex');
}
function canonicalType(text) {
    // Stock TypeScript may print union members in symbol-discovery order.
    return text.replace(/\b[A-Za-z0-9_]+(?: \| [A-Za-z0-9_]+)+/g,
        union => union.split(' | ').sort().join(' | '));
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
            // Earlier numbered adaptations may legitimately change runtime syntax.
            // adapt.cjs compares every proposed file with this invocation's input
            // before writing anything, so 71 must still preserve runtime syntax.
            records.push({ family: spec.family, ...owner, runtime_sha256: fingerprint(fn.body) });
        }
        for (const edit of spec.edits) {
            const file = program.getSourceFile(path.join(tree, edit.file));
            const owner = edit.owner ? find(program, tree, { file: edit.file, function: edit.owner }) : undefined;
            const begin = owner ? owner.getStart(file) : 0;
            const text = file.text.slice(begin, owner ? owner.end : file.text.length);
            const count = needle => text.split(needle).length - 1;
            if (edit.fresh_array_type) {
                let matches = 0;
                const checker = program.getTypeChecker();
                function visit(n, parent) {
                    const selected = edit.parenthesized ? parent && ts.isParenthesizedExpression(parent) && [edit.before, edit.after].includes(parent.getText(file)) : [edit.before, edit.after].includes(n.getText(file));
                    if (ts.isBinaryExpression(n) && n.operatorToken.kind === ts.SyntaxKind.EqualsToken &&
                        selected) {
                        const literal = ts.isAsExpression(n.right) ? n.right.expression : n.right;
                        if (!ts.isArrayLiteralExpression(literal) || literal.elements.length) throw new Error('expected fresh empty allocation');
                        const declared = checker.getContextualType(n.right);
                        const array = declared?.isUnion() ? declared.types.find(t => checker.isArrayType(t)) : declared;
                        if (!array || !checker.isArrayType(array) || canonicalType(checker.typeToString(array, n, ts.TypeFormatFlags.NoTruncation)) !== canonicalType(edit.fresh_array_type))
                            throw new Error('fresh array holder type changed: ' + edit.before + ': ' + (array && checker.typeToString(array, n, ts.TypeFormatFlags.NoTruncation)));
                        matches++;
                    }
                    ts.forEachChild(n, child => visit(child, n));
                }
                visit(owner || file);
                if (matches !== (edit.count || 1)) throw new Error('fresh array assignment shape changed: ' + edit.before);
            }
            // Some shorter original signatures are contained in their adapted form.
            const expected = edit.count || 1;
            if (count(edit.after) === expected) continue;
            if (count(edit.before) !== expected) throw new Error(`type edit shape changed: ${spec.family}: ${edit.before}`);
            if (!edits.has(file)) edits.set(file, []);
            let cursor = 0;
            for (let i = 0; i < expected; i++) {
                const offset = text.indexOf(edit.before, cursor), start = begin + offset;
                edits.get(file).push({ start, end: start + edit.before.length, text: edit.after });
                cursor = offset + edit.before.length;
            }
        }
    }
    return records;
};
module.exports.fingerprint = fingerprint;
module.exports.find = find;
