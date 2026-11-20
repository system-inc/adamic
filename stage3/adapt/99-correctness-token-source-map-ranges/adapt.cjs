'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

assert.equal(ts.version, '6.0.3');

const owners = [
    {
        file: 'factory/nodeFactory.ts',
        name: 'mergeTokenSourceMapRanges',
        before: `function mergeTokenSourceMapRanges(sourceRanges: (TextRange | undefined)[], destRanges: (TextRange | undefined)[]) {
    if (!destRanges) destRanges = [];
    for (const key in sourceRanges) {
        destRanges[key] = sourceRanges[key];
    }
    return destRanges;
}`,
        old: '(sourceRanges: (TextRange | undefined)[], destRanges: (TextRange | undefined)[])',
        replacement: '(sourceRanges: (SourceMapRange | undefined)[], destRanges: (SourceMapRange | undefined)[]): (SourceMapRange | undefined)[]',
    },
];

function plan(tree, pin) {
    const file = path.join(tree, 'src/compiler', pin.file);
    const raw = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, raw, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `${file}: parse error`);
    const matches = [];
    function visit(node) {
        if (ts.isFunctionDeclaration(node) && node.name?.text === pin.name) matches.push(node);
        ts.forEachChild(node, visit);
    }
    visit(source);
    assert.equal(matches.length, 1, `${file}: unexpected ${pin.name} count`);
    const [owner] = matches;
    if (pin.parent) {
        assert(ts.isBlock(owner.parent) && ts.isFunctionDeclaration(owner.parent.parent)
            && owner.parent.parent.name?.text === pin.parent, `${file}: ${pin.name} moved`);
    }
    else {
        assert.equal(owner.parent, source, `${file}: ${pin.name} moved`);
    }
    assert.equal(pin.before.split(pin.old).length, 2, `${file}: ambiguous edit`);
    const after = pin.before.replace(pin.old, pin.replacement);
    const start = owner.getStart(source);
    const ownerText = raw.slice(start, owner.end);
    const actual = ownerText.replaceAll('\r\n', '\n');
    const imports = source.statements.filter(node => ts.isImportDeclaration(node)
        && node.moduleSpecifier.text === '../_namespaces/ts.js');
    assert.equal(imports.length, 1, `${file}: namespace import drift`);
    const bindings = imports[0].importClause?.namedBindings;
    assert(bindings && ts.isNamedImports(bindings), `${file}: expected a named import`);
    const elements = [...bindings.elements];
    const sourceIndex = elements.findIndex(element => element.getText(source) === 'type SourceMapSource');
    assert(sourceIndex > 0, `${file}: SourceMapSource import drift`);
    const ranges = elements.filter(element => element.name.text === 'SourceMapRange');
    if (actual === after) {
        assert.equal(ranges.length, 1, `${file}: partly adapted import`);
        assert.equal(elements[sourceIndex - 1], ranges[0], `${file}: misplaced SourceMapRange import`);
        assert.equal(ranges[0].getText(source), 'type SourceMapRange', `${file}: SourceMapRange import drift`);
        return { file, raw, output: raw };
    }
    assert.equal(ranges.length, 0, `${file}: partly adapted import`);
    const at = elements[sourceIndex].getStart(source);
    const leading = /^(\r?\n[ \t]*)type SourceMapSource/.exec(raw.slice(elements[sourceIndex].getFullStart(), elements[sourceIndex].end));
    assert(leading, `${file}: expected SourceMapSource on its own line`);

    assert.equal(actual, pin.before, `${file}: ${pin.name} drift or partial adaptation`);
    const output = raw.slice(0, at) + 'type SourceMapRange,' + leading[1] + raw.slice(at, start) + ownerText.replace(pin.old, pin.replacement) + raw.slice(owner.end);
    return { file, raw, output };
}

function adapt(tree) {
    const plans = owners.map(pin => plan(tree, pin));
    const changed = plans.filter(({ raw, output }) => raw !== output);
    for (const { file, raw } of plans) assert.equal(fs.readFileSync(file, 'utf8'), raw, `${file}: changed while planning`);
    for (const { file, output } of changed) fs.writeFileSync(file, output);
    return { files: changed.length, javascriptChanging: false };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
