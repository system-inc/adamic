'use strict';

const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const ts = require('typescript');

function one(nodes, predicate, label) {
    const matches = nodes.filter(predicate);
    assert.equal(matches.length, 1, `expected one ${label}`);
    return matches[0];
}

function functionNamed(statements, name) {
    return one(statements, node => ts.isFunctionDeclaration(node) && node.name?.text === name && node.body, name);
}

function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const file = path.join(tree, 'src/compiler/moduleNameResolver.ts');
    const text = fs.readFileSync(file, 'utf8');
    const source = ts.createSourceFile(file, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0);

    const factory = functionNamed(source.statements, 'getLoadModuleFromTargetExportOrImport');
    const load = functionNamed(factory.body.statements, 'loadModuleFromTargetExportOrImport');
    const tryLoad = functionNamed(load.body.statements, 'tryLoadInputFileForPath');
    const directories = functionNamed(tryLoad.body.statements, 'getOutputDirectoriesForBaseDirectory');
    const candidates = one(directories.body.statements, node =>
        ts.isVariableStatement(node) && node.declarationList.declarations.some(declaration =>
            ts.isIdentifier(declaration.name) && declaration.name.text === 'candidateDirectories'),
        'candidateDirectories statement');
    const before = 'const candidateDirectories = [];';
    const after = 'const candidateDirectories: string[] = [];';
    const snapshot = candidates.getText(source);

    if (snapshot === after) return { files: 0, edits: 0 };
    assert.equal(snapshot, before, 'unexpected candidateDirectories declaration');
    fs.writeFileSync(file, text.slice(0, candidates.getStart(source)) + after + text.slice(candidates.end));
    return { files: 1, edits: 1 };
}

module.exports = { adapt };
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2]))));
}
