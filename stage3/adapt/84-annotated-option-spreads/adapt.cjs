#!/usr/bin/env node
'use strict';
// Adaptation 84, #w743639's K: three unannotated spreads get the type they are used as, so each literal is made as
// its slot holds it instead of a fixed object later seen as one ("fixed objects and records have different storage;
// copy explicitly", and under it, at the third, a widening):
//   - checker.ts, getSpecifierForModuleSymbol: `const specifierCompilerOptions = isBundle ? { ...compilerOptions, baseUrl: ... } : compilerOptions;`
//     passed to moduleSpecifiers.getModuleSpecifiers as its CompilerOptions;
//   - moduleNameResolver.ts, nodeModuleNameResolverWorker: `const diagnosticsCompilerOptions = { ...state.compilerOptions, moduleResolution: ... };`
//     passed to getConditions and stored as diagnosticState's compilerOptions (ModuleResolutionState.compilerOptions);
//   - moduleNameResolver.ts, nodeModuleNameResolverWorker: `const diagnosticState = { ...state, compilerOptions: ..., conditions: ... };`
//     handed to tryResolve as its ModuleResolutionState, annotated so, so its conditions (getConditions' string[]) is
//     taken as the state's readonly string[] where it's built, never seen as one that can be written.
// Type-only: each annotation adds `: <type>` after its name and nothing else, and the emitted JavaScript does not
// change. Ruled by Adamic on #w743639 after the probes J-grhaqpf and J-wmsybfj. Each
// declaration must be found unannotated with upstream's initializer start; one already annotated keeps its bytes.
const fs = require('node:fs');
const path = require('node:path');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
if (ts.version !== '6.0.3') throw new Error(`expected stock TypeScript 6.0.3, got ${ts.version}`);
if (process.argv.length !== 3) throw new Error('usage: node adapt.cjs <tree>');
const tree = path.resolve(process.argv[2]);
const sites = [
    { file: 'src/compiler/checker.ts', owner: 'getSpecifierForModuleSymbol', name: 'specifierCompilerOptions', starts: 'isBundle ? { ...compilerOptions, baseUrl:' },
    { file: 'src/compiler/moduleNameResolver.ts', owner: 'nodeModuleNameResolverWorker', name: 'diagnosticsCompilerOptions', starts: '{ ...state.compilerOptions, moduleResolution:' },
    { file: 'src/compiler/moduleNameResolver.ts', owner: 'nodeModuleNameResolverWorker', name: 'diagnosticState', starts: '{ ...state, compilerOptions: diagnosticsCompilerOptions,', type: 'ModuleResolutionState' },
];
function enclosing(node) {
    for (let current = node.parent; current; current = current.parent) {
        if (ts.isFunctionDeclaration(current) && current.name) return current.name.text;
    }
    return undefined;
}
const records = [];
for (const site of sites) {
    const fileName = path.join(tree, site.file);
    const text = fs.readFileSync(fileName, 'utf8');
    const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
    const found = [];
    ts.forEachChild(source, function visit(node) {
        // By name, owner and the start of its initializer: nodeModuleNameResolverWorker holds two `diagnosticState`s, and
        // only the one that sets compilerOptions is this site's.
        if (ts.isVariableDeclaration(node) && ts.isIdentifier(node.name) && node.name.text === site.name && enclosing(node) === site.owner &&
            node.initializer && node.initializer.getText(source).replace(/\s+/g, ' ').startsWith(site.starts)) found.push(node);
        ts.forEachChild(node, visit);
    });
    if (found.length !== 1) throw new Error(`expected one ${site.name} in ${site.owner} starting ${JSON.stringify(site.starts)}, found ${found.length}`);
    const declaration = found[0];
    const annotation = declaration.type && declaration.type.getText(source);
    const wanted = site.type || 'CompilerOptions';
    if (annotation !== undefined && annotation !== wanted) throw new Error(`${site.name} is annotated ${annotation}`);
    const at = source.getLineAndCharacterOfPosition(declaration.getStart(source));
    records.push({ owner: site.owner, name: site.name, where: `${site.file}:${at.line + 1}:${at.character + 1}`, edited: annotation === undefined });
    if (annotation === undefined) {
        if (fs.readFileSync(fileName, 'utf8') !== text) throw new Error(`source changed during adaptation: ${fileName}`);
        fs.writeFileSync(fileName, text.slice(0, declaration.name.end) + `: ${wanted}` + text.slice(declaration.name.end));
    }
}
console.log(JSON.stringify({ typescript: ts.version, task: '#w743639', javascriptChanging: false, edits: records }, null, 2));
