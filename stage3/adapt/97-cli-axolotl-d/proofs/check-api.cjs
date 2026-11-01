'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const {canonical, changes} = require('../../../lane/normalize-api.cjs');

// Proves api-sanctions.json against stock 6.0.3's typescript.d.ts and against the emitted declarations of the adapted tree (README.md).
// usage: node check-api.cjs <stock typescript.d.ts> <tree before 97d> <tree after 97d> <output directory>
assert.equal(ts.version, '6.0.3');
const sanctions = JSON.parse(fs.readFileSync(path.join(__dirname, '../api-sanctions.json'), 'utf8'));
const {targets} = require('../adapt.cjs');
assert.equal(sanctions.length, 19);
assert.deepEqual(sanctions.map(s => s.path), targets.filter(t => t.scope === 'Public').map(t => [t.owner, t.member]), 'the sanctions are not the 19 public targets');

function nodes(source, test) {
    const found = [];
    (function visit(node) {
        if (test(node)) found.push(node);
        ts.forEachChild(node, visit);
    })(source);
    return found;
}

// generate-api.cjs's ownerPath: every named ancestor except the ts namespace.
function ownerPath(node) {
    const names = [];
    for (let current = node; current; current = current.parent) {
        if (current.name && !(ts.isModuleDeclaration(current) && current.name.getText(node.getSourceFile()) === 'ts')) names.unshift(current.name.getText(node.getSourceFile()));
    }
    return names;
}

function member(text) {
    const model = ts.createSourceFile('sanction.d.ts', `interface Sanction { ${text} }`, ts.ScriptTarget.Latest, true);
    assert.equal(model.parseDiagnostics.length, 0, `cannot parse ${text}`);
    assert.equal(model.statements[0].members.length, 1);
    return model.statements[0].members[0];
}

// after must be exactly the idiom over before: `x?: { x(…): R; }["x"] | undefined`, where the inner method is before without its question token.
function idiom(before, after, name) {
    const method = member(before);
    assert(ts.isMethodSignature(method) && method.questionToken && method.name.getText() === name, `${name}: before is not an optional method`);
    const property = member(after);
    assert(ts.isPropertySignature(property) && property.questionToken && property.name.getText() === name, `${name}: after is not an optional property`);
    const union = property.type;
    assert(union && ts.isUnionTypeNode(union) && union.types.length === 2 && union.types[1].kind === ts.SyntaxKind.UndefinedKeyword, `${name}: after is not a union with undefined`);
    const indexed = union.types[0];
    assert(ts.isIndexedAccessTypeNode(indexed) && ts.isTypeLiteralNode(indexed.objectType) && indexed.objectType.members.length === 1, `${name}: after is not an indexed type literal`);
    assert(ts.isLiteralTypeNode(indexed.indexType) && ts.isStringLiteral(indexed.indexType.literal) && indexed.indexType.literal.text === name, `${name}: after indexes another name`);
    const inner = indexed.objectType.members[0];
    assert(ts.isMethodSignature(inner) && !inner.questionToken, `${name}: the indexed member is not a plain method`);
    const strip = node => JSON.stringify(canonical(node).filter(child => !(Array.isArray(child) && child[0] === 'QuestionToken')));
    assert.equal(strip(inner), strip(method), `${name}: the indexed method is not before's signature`);
}

// 1. Each sanction names one optional method in stock typescript.d.ts by generate-api.cjs's owner path, spelled exactly as before.
// Returns stock typescript.d.ts with every sanction applied, the API this head declares.
function expectedApi(stockText) {
    const stock = ts.createSourceFile('typescript.d.ts', stockText, ts.ScriptTarget.Latest, true);
    const lines = [];
    const edits = [];
    for (const sanction of sanctions) {
        assert.equal(sanction.kind, 'PropertySignature');
        assert(sanction.why && !/[–—]/.test(sanction.why), `${sanction.path}: why is missing or uses a dash`);
        const found = nodes(stock, n => (ts.isMethodSignature(n) || ts.isPropertySignature(n)) && JSON.stringify(ownerPath(n)) === JSON.stringify(sanction.path));
        assert.equal(found.length, 1, `${sanction.path}: ${found.length} stock owners, not one`);
        assert(ts.isMethodSignature(found[0]), `${sanction.path}: stock owner is not a method`);
        assert.equal(found[0].getText(stock), sanction.before, `${sanction.path}: before is not stock's text`);
        idiom(sanction.before, sanction.after, sanction.path[1]);
        const line = stock.getLineAndCharacterOfPosition(found[0].getStart(stock)).line + 1;
        lines.push({path: sanction.path, line, before: sanction.before, after: sanction.after});
        edits.push({start: found[0].getStart(stock), end: found[0].end, text: sanction.after});
    }
    let expected = stockText;
    for (const edit of edits.sort((a, b) => b.start - a.start)) expected = expected.slice(0, edit.start) + edit.text + expected.slice(edit.end);
    return {expected, lines};
}

function main(stockFile, beforeTree, afterTree, output) {
    const stockText = fs.readFileSync(stockFile, 'utf8');
    const {expected, lines} = expectedApi(stockText);

    // 2. The lane's own normalizer sees exactly these 19 members change, each from a method to a property, and no comment change.
    const stockChanges = changes(stockText, expected);
    assert.deepEqual(stockChanges.map(c => c.declaration), sanctions.map(s => `member:${JSON.stringify(['ts', ...s.path])}#0`).sort(), 'typescript.d.ts changes other declarations');
    for (const change of stockChanges) assert.deepEqual([JSON.parse(change.before)[0], JSON.parse(change.after)[0]], ['MethodSignature', 'PropertySignature']);

    // 3. The adapted tree's own declarations, stripInternal as the public bundle is, change in the same 19 members and in nothing else,
    // and each emitted member normalizes to its sanction's after.
    const files = [...new Set(targets.map(t => t.file))];
    const emitted = {before: {}, after: {}};
    for (const [side, tree] of [['before', beforeTree], ['after', afterTree]]) {
        for (const file of files) {
            const text = fs.readFileSync(path.join(tree, 'src/compiler', file), 'utf8');
            emitted[side][file] = ts.transpileDeclaration(text, {fileName: file, compilerOptions: {stripInternal: true}}).outputText;
        }
    }
    const emittedChanges = files.flatMap(file => changes(emitted.before[file], emitted.after[file]));
    assert.deepEqual(emittedChanges.map(c => c.declaration).sort(), sanctions.map(s => `member:${JSON.stringify(s.path)}#0`).sort(), 'the emitted public declarations change other members');
    for (const sanction of sanctions) {
        const change = emittedChanges.find(c => c.declaration === `member:${JSON.stringify(sanction.path)}#0`);
        assert.equal(change.after, JSON.stringify(canonical(member(sanction.after))), `${sanction.path}: the emitted member is not the sanctioned after`);
        assert.equal(change.before, JSON.stringify(canonical(member(sanction.before))), `${sanction.path}: the emitted member was not stock's before`);
    }

    // The posted forms: the typescript.d.ts lines, and both diffs.
    fs.mkdirSync(output, {recursive: true});
    const scratch = fs.mkdtempSync(path.join(output, '.work-'));
    const write = (relative, text) => {
        fs.mkdirSync(path.dirname(path.join(scratch, relative)), {recursive: true});
        fs.writeFileSync(path.join(scratch, relative), text);
    };
    write('stock/typescript.d.ts', stockText);
    write('expected/typescript.d.ts', expected);
    for (const side of ['before', 'after']) for (const file of files) write(`${side}/${file.replace(/\.ts$/, '.d.ts')}`, emitted[side][file]);
    function diff(a, b) {
        try {
            execFileSync('git', ['diff', '--no-index', '--no-color', '-U1', a, b], {cwd: scratch, encoding: 'utf8'});
            return '';
        } catch (error) {
            assert.equal(error.status, 1, error.stderr);
            return error.stdout;
        }
    }
    fs.writeFileSync(path.join(output, 'typescript.d.ts.diff'), diff('stock', 'expected'));
    fs.writeFileSync(path.join(output, 'emitted-public.d.ts.diff'), diff('before', 'after'));
    for (const relative of ['stock/typescript.d.ts', 'expected/typescript.d.ts', ...['before', 'after'].flatMap(side => files.map(file => `${side}/${file.replace(/\.ts$/, '.d.ts')}`))]) fs.unlinkSync(path.join(scratch, relative));
    for (const directory of ['stock', 'expected', 'before', 'after']) fs.rmdirSync(path.join(scratch, directory));
    fs.rmdirSync(scratch);
    fs.writeFileSync(path.join(output, 'typescript.d.ts.lines.json'), JSON.stringify(lines, null, 2) + '\n');
    console.log(JSON.stringify({sanctions: sanctions.length, typescriptDtsChanges: stockChanges.length, emittedChanges: emittedChanges.length, stockLines: lines.map(l => l.line)}));
}
module.exports = {expectedApi};
if (require.main === module) {
    assert.equal(process.argv.length, 6, 'usage: node check-api.cjs <stock typescript.d.ts> <tree before> <tree after> <output directory>');
    main(...process.argv.slice(2));
}
