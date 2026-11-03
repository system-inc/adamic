'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

const before = 'const resolutionsInFile = createModeAwareCache<ResolutionWithFailedLookupLocations>();';
const after = 'const resolutionsInFile = createModeAwareCache<ResolvedModuleWithFailedLookupLocations>();';
function nodes(root, predicate) {
    const found = [];
    (function visit(node) { if (predicate(node)) found.push(node); ts.forEachChild(node, visit); })(root);
    return found;
}
function locate(fileName, text) {
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    assert.equal(source.parseDiagnostics.length, 0, 'cannot parse program.ts');
    function named(name) {
        const found = nodes(source, n => ts.isFunctionDeclaration(n) && n.body && n.name?.text === name);
        assert.equal(found.length, 1, `${name} count drift`);
        return found[0];
    }
    const program = named('createProgram'), fn = named('processImportedModules');
    assert.equal(fn.parent, program.body, 'processImportedModules left createProgram');
    const declarations = nodes(fn, n => ts.isVariableDeclaration(n) && n.name.getText(source) === 'resolutionsInFile');
    assert.equal(declarations.length, 1, 'resolutionsInFile count drift');
    const declaration = declarations[0], statement = declaration.parent.parent;
    assert(ts.isVariableStatement(statement), 'cache declaration moved');
    const spelling = statement.getText(source);
    assert(spelling === before || spelling === after, 'cache statement is not an inventoried spelling');
    const block = statement.parent;
    assert(ts.isBlock(block) && ts.isIfStatement(block.parent) && block.parent.parent === fn.body, 'cache block moved');
    assert.equal(block.parent.expression.getText(source), 'file.imports.length || file.moduleAugmentations.length', 'cache guard drift');
    const statements = block.statements;
    assert.equal(statements.length, 7, 'cache block statement count drift');
    assert.equal(statements[0].getText(source), 'const moduleNames = getModuleNames(file);');
    assert.equal(statements[1].getText(source).replace(/\r\n/g, '\n'), 'const resolutions = resolvedModulesProcessing?.get(file.path) ||\n                resolveModuleNamesReusingOldState(moduleNames, file);');
    assert.equal(statements[2].getText(source), 'Debug.assert(resolutions.length === moduleNames.length);');
    assert.equal(statements[3].getText(source), 'const optionsForFile = getCompilerOptionsForFile(file);');
    assert.equal(statements[4], statement, 'cache declaration moved');
    assert.equal(statements[5].getText(source), '(resolvedModules ??= new Map()).set(file.path, resolutionsInFile);');
    const references = nodes(fn, n => ts.isIdentifier(n) && n.text === 'resolutionsInFile');
    assert.equal(references.length, 3, 'cache alias use count drift');
    assert.equal(references[1].parent.getText(source), '(resolvedModules ??= new Map()).set(file.path, resolutionsInFile)');
    assert.equal(references[2].parent.parent.parent.getText(source), 'resolutionsInFile.set(moduleName, mode, resolutions[index]!);');
    const loop = statements[6];
    assert(ts.isForStatement(loop), 'cache population loop moved');
    assert.equal(loop.initializer.getText(source), 'let index = 0');
    assert.equal(loop.condition.getText(source), 'index < moduleNames.length');
    assert.equal(loop.incrementor.getText(source), 'index++');
    assert.equal(references[2].parent.parent.parent.parent, loop.statement, 'cache write left population loop');
    return {source, fn, statement, adapted: spelling === after};
}
function validate(fileName, text) {
    const site = locate(fileName, text);
    return site;
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const fileName = path.join(tree, 'src/compiler/program.ts');
    const text = fs.readFileSync(fileName, 'utf8');
    const site = validate(fileName, text);
    const result = site.adapted ? text : text.slice(0, site.statement.getStart(site.source)) + after + text.slice(site.statement.end);
    assert(validate(fileName, result).adapted, 'planned result is not fully adapted');
    assert.equal(fs.readFileSync(fileName, 'utf8'), text, 'source changed while planning');
    if (!site.adapted) fs.writeFileSync(fileName, result);
    return {files: site.adapted ? 0 : 1, edits: site.adapted ? 0 : 1, javascriptChanging: false, behaviorChanging: false};
}
module.exports = {adapt, before, after};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
