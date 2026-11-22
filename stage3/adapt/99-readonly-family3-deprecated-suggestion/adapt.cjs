'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');

function one(nodes, test, label) {
    const matches = nodes.filter(test);
    assert.equal(matches.length, 1, `expected one ${label}`);
    return matches[0];
}

function parse(tree, file) {
    const filename = path.join(tree, 'src/compiler', file);
    const text = fs.readFileSync(filename, 'utf8');
    const source = ts.createSourceFile(filename, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
    return { filename, source, text };
}

function functionNamed(nodes, name) {
    return one(nodes, node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body, name);
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3', 'requires stock TypeScript 6.0.3');
    const core = parse(tree, 'core.ts').source;
    const guard = functionNamed(core.statements, 'isArray');
    assert.equal(guard.type?.getText(core), 'value is readonly unknown[]');
    const statement = one(guard.body.statements, ts.isReturnStatement, 'isArray return');
    assert.equal(statement.expression?.getText(core), 'Array.isArray(value)');

    const { filename, source, text } = parse(tree, 'checker.ts');
    const namespace = one(source.statements, node => ts.isImportDeclaration(node) &&
        node.moduleSpecifier.getText(source) === '"./_namespaces/ts.js"', 'checker ts import');
    const imports = namespace.importClause?.namedBindings;
    assert(imports && ts.isNamedImports(imports));
    assert.equal(one(imports.elements, node => node.name.text === 'isArray', 'isArray import').getText(source), 'isArray');

    const checker = functionNamed(source.statements, 'createTypeChecker');
    const entry = functionNamed(checker.body.statements, 'addDeprecatedSuggestion');
    const worker = functionNamed(checker.body.statements, 'addDeprecatedSuggestionWorker');
    const declarationType = node => {
        const parameter = one(node.parameters, p => p.name.getText(source) === 'declarations',
            `${node.name.text} declarations parameter`);
        assert(parameter.type && !parameter.questionToken && !parameter.initializer && !parameter.dotDotDotToken);
        return parameter.type;
    };
    const variables = worker.body.statements.filter(ts.isVariableStatement)
        .flatMap(node => [...node.declarationList.declarations]);
    const tag = one(variables, node => node.name.getText(source) === 'deprecatedTag', 'deprecatedTag binding');
    const selector = tag.initializer;
    assert(selector && ts.isConditionalExpression(selector));
    assert.equal(selector.whenTrue.getText(source), 'forEach(declarations, getJSDocDeprecatedTag)');
    assert.equal(selector.whenFalse.getText(source), 'getJSDocDeprecatedTag(declarations)');

    const slots = [
        [declarationType(entry), 'Node[]', 'readonly Node[]'],
        [declarationType(worker), 'Node | Node[]', 'Node | readonly Node[]'],
        [selector.condition, 'Array.isArray(declarations)', 'isArray(declarations)'],
    ];
    const positions = slots.map(([node, before, after]) => {
        const found = node.getText(source);
        assert(found === before || found === after, `unreviewed checker text: ${found}`);
        return { start: node.getStart(source), end: node.end, after, changed: found === before };
    });
    assert(positions.every(edit => edit.changed === positions[0].changed), 'partly adapted checker');
    if (!positions[0].changed) return { files: 0, edits: 0, javascriptChanging: true };

    let output = text;
    for (const { start, end, after } of positions.sort((a, b) => b.start - a.start)) {
        output = output.slice(0, start) + after + output.slice(end);
    }
    assert.equal(fs.readFileSync(filename, 'utf8'), text, 'checker.ts changed during planning');
    fs.writeFileSync(filename, output);
    return { files: 1, edits: positions.length, javascriptChanging: true };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
