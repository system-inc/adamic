'use strict';
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const ts = require(process.env.CENSUS_TYPESCRIPT || 'typescript');
assert.equal(ts.version, '6.0.3');
function nodes(root, predicate) {
    const result = [];
    function visit(node) { if (predicate(node)) result.push(node); ts.forEachChild(node, visit); }
    visit(root);
    return result;
}
function census(tree) {
    const generators = [], delegations = [];
    function walk(directory) {
        for (const entry of fs.readdirSync(directory, {withFileTypes: true}).sort((a,b) => a.name.localeCompare(b.name))) {
            const file = path.join(directory, entry.name);
            if (entry.isDirectory()) { walk(file); continue; }
            if (!file.endsWith('.ts')) continue;
            const source = ts.createSourceFile(file, fs.readFileSync(file, 'utf8'), ts.ScriptTarget.Latest, true);
            assert.equal(source.parseDiagnostics.length, 0, file);
            function location(node) {
                const point = source.getLineAndCharacterOfPosition(node.getStart(source));
                return {file: path.relative(tree, file), line: point.line+1, column: point.character+1};
            }
            for (const node of nodes(source, n => ts.isFunctionLike(n) && !!n.asteriskToken)) generators.push({...location(node), name: node.name?.getText(source) || '<anonymous>', body: node.body.getText(source)});
            for (const node of nodes(source, n => ts.isYieldExpression(n) && !!n.asteriskToken)) delegations.push(location(node));
        }
    }
    walk(path.join(tree, 'src/compiler'));
    return {typescript: ts.version, generatorCount: generators.length, delegationCount: delegations.length, generators, delegations};
}
if (require.main === module) {
    const report = census(path.resolve(process.argv[2]));
    if (process.argv[3] && process.argv[3] !== '--check') fs.writeFileSync(process.argv[3], JSON.stringify(report,null,2)+'\n');
    console.log(JSON.stringify(report));
    if (process.argv.includes('--check') && (report.generatorCount || report.delegationCount)) { console.error('census failed: generators or yield delegation remain'); process.exitCode = 1; }
}
module.exports = {census, nodes, ts};
