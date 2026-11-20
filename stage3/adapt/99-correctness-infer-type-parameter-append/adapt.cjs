'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

assert.equal(ts.version, '6.0.3');

const owners = [
    {
        file: 'checker.ts',
        parent: 'createTypeChecker',
        name: 'getInferTypeParameters',
        before: `function getInferTypeParameters(node: ConditionalTypeNode): TypeParameter[] | undefined {
        let result: TypeParameter[] | undefined;
        if (node.locals) {
            node.locals.forEach(symbol => {
                if (symbol.flags & SymbolFlags.TypeParameter) {
                    result = append(result, getDeclaredTypeOfSymbol(symbol));
                }
            });
        }
        return result;
    }`,
        old: 'append(result, getDeclaredTypeOfSymbol(symbol))',
        replacement: 'append(result, getDeclaredTypeOfTypeParameter(symbol))',
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
    if (actual === after) return { file, raw, output: raw };
    assert.equal(actual, pin.before, `${file}: ${pin.name} drift or partial adaptation`);
    const output = raw.slice(0, start) + ownerText.replace(pin.old, pin.replacement) + raw.slice(owner.end);
    return { file, raw, output };
}

function adapt(tree) {
    const plans = owners.map(pin => plan(tree, pin));
    const changed = plans.filter(({ raw, output }) => raw !== output);
    for (const { file, raw } of plans) assert.equal(fs.readFileSync(file, 'utf8'), raw, `${file}: changed while planning`);
    for (const { file, output } of changed) fs.writeFileSync(file, output);
    return { files: changed.length, javascriptChanging: changed.length !== 0 };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
