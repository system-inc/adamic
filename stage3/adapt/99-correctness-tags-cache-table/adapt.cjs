'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');
const pins = require('./owners.json');

const moduleFile = 'src/compiler/jsDocCache.ts';
const moduleText = fs.readFileSync(path.join(__dirname, 'jsDocCache.ts'), 'utf8');

function locate(source, pin) {
    const found = [];
    function visit(node, owners) {
        const name = ts.isImportDeclaration(node) && ts.isStringLiteral(node.moduleSpecifier)
            ? node.moduleSpecifier.text : node.name?.text;
        const next = name && (ts.isModuleDeclaration(node) || ts.isFunctionDeclaration(node) ||
            ts.isInterfaceDeclaration(node) || ts.isImportDeclaration(node)) ? [...owners, name] : owners;
        if (ts.SyntaxKind[node.kind] === pin.kind && JSON.stringify(next) === JSON.stringify(pin.owner)) {
            found.push(node);
        }
        ts.forEachChild(node, child => visit(child, next));
    }
    visit(source, []);
    return found;
}

function sourceFile(file, text) {
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, file + ': parse diagnostics');
    return source;
}

function deletionSpan(source, pin, nodes, context) {
    const anchors = pin.between.map(name =>
        locate(source, { kind: pin.kind, owner: [name] }));
    assert(anchors.every(matches => matches.length === 1), context + ': missing or duplicate deletion anchor');
    const [[previous], [next]] = anchors;
    const start = previous.end;
    const end = next.getStart(source);
    assert(start <= end, context + ': deletion anchors out of order');
    assert(nodes.length <= 1 && nodes.every(node => start <= node.getStart(source) && node.end <= end),
        context + ': deleted owner moved or duplicated');

    const current = source.text.slice(start, end);
    if (nodes.length === 1 && current === pin.before) {
        return { after: false, edit: { start, end, value: pin.after } };
    }
    assert(nodes.length === 0 && current === pin.after, context + ': drift or half-adapted deletion span');
    return { after: true };
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = [];
    const states = [];
    for (const file of new Set(pins.map(pin => pin.file))) {
        const filename = path.join(tree, file);
        const raw = fs.readFileSync(filename, 'utf8');
        const crlf = raw.includes('\r\n');
        assert(!/\r(?!\n)/.test(raw) && !(crlf && /(?<!\r)\n/.test(raw)), file + ': mixed endings');
        const text = raw.replace(/\r\n/g, '\n');
        const source = sourceFile(file, text);
        const edits = [];
        for (const pin of pins.filter(pin => pin.file === file)) {
            const nodes = locate(source, pin);
            const added = pin.addImport ? locate(source, { kind: 'ImportDeclaration', owner: ['./jsDocCache.js'] }) : [];
            const context = file + ': ' + pin.owner.join('.');
            if (pin.between) {
                const deletion = deletionSpan(source, pin, nodes, context);
                states.push(deletion.after);
                if (deletion.edit) edits.push(deletion.edit);
                continue;
            }
            assert.equal(nodes.length, 1, context + ': missing or duplicate owner');
            const node = nodes[0];
            const current = node?.getText(source) ?? '';
            const before = current === pin.before && added.length === 0;
            const after = current === pin.after &&
                (!pin.addImport || (added.length === 1 && added[0].getText(source) === pin.addImport));
            assert(before || after, context + ': drift or half-adapted owner');
            states.push(after);
            if (before) {
                edits.push({
                    start: node.getStart(source),
                    end: node.end,
                    value: pin.after + (pin.addImport ? '\n' + pin.addImport : ''),
                });
            }
        }
        let output = text;
        let boundary = text.length;
        for (const edit of edits.sort((a, b) => b.start - a.start)) {
            assert(edit.end <= boundary, file + ': overlapping edits');
            boundary = edit.start;
            output = output.slice(0, edit.start) + edit.value + output.slice(edit.end);
        }
        sourceFile(file, output);
        plans.push({ filename, raw, output: crlf ? output.replace(/\n/g, '\r\n') : output });
    }

    const modulePath = path.join(tree, moduleFile);
    sourceFile(moduleFile, moduleText);
    const moduleExists = fs.existsSync(modulePath);
    assert(!moduleExists || fs.readFileSync(modulePath, 'utf8') === moduleText, moduleFile + ': drift');
    states.push(moduleExists);
    assert(states.every(state => state === states[0]), 'half-adapted tree');
    for (const plan of plans) assert.equal(fs.readFileSync(plan.filename, 'utf8'), plan.raw);
    if (!moduleExists) {
        for (const plan of plans) fs.writeFileSync(plan.filename, plan.output);
        fs.writeFileSync(modulePath, moduleText, { flag: 'wx' });
    }
    return { files: moduleExists ? 0 : plans.length + 1, javascriptChanging: true };
}

module.exports = { adapt, locate };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs TREE');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
