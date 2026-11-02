'use strict';
const fs = require('node:fs');
const path = require('node:path');
const assert = require('node:assert/strict');
const ts = require('typescript');

// Existing optional data properties only. Keep the explicit undefined stores.
const sites = [
    ...['showConfig', 'help', 'init', 'listFiles', 'listEmittedFiles', 'build', 'version']
        .map(name => ({owner: 'CompilerOptions', before: `${name}?: boolean;`})),
    {owner: 'CompilerOptions', before: 'readonly configFile?: TsConfigSourceFile;'},
    {owner: 'CompilerOptions', before: 'project?: string;'},
    {owner: 'ProjectReference', before: 'originalPath?: string;'},
].map(site => ({file: 'types.ts', ...site})).concat([
    {file: 'commandLineParser.ts', owner: 'TSConfig', before: 'exclude?: readonly string[];'},
    {file: 'commandLineParser.ts', owner: 'TSConfig', before: 'include?: readonly string[];'},
]);
function nodes(root, predicate) {
    const result = [];
    (function visit(node) { if (predicate(node)) result.push(node); ts.forEachChild(node, visit); })(root);
    return result;
}
function adapt(tree) {
    assert.equal(ts.version, '6.0.3');
    const plans = [];
    let adapted = 0;
    for (const file of new Set(sites.map(site => site.file))) {
        const fileName = path.join(tree, 'src/compiler', file);
        const text = fs.readFileSync(fileName, 'utf8');
        const source = ts.createSourceFile(fileName, text, ts.ScriptTarget.Latest, true);
        assert.equal(source.parseDiagnostics.length, 0, `cannot parse ${file}`);
        const edits = [];
        for (const site of sites.filter(site => site.file === file)) {
            const owners = nodes(source, node => ts.isInterfaceDeclaration(node) && node.name.text === site.owner);
            assert.equal(owners.length, 1, `owner drift: ${site.owner}`);
            const after = site.before.slice(0, -1) + ' | undefined;';
            const matches = owners[0].members.filter(node => node.getText(source) === site.before || node.getText(source) === after);
            assert.equal(matches.length, 1, `member drift: ${site.owner}.${site.before}`);
            const member = matches[0];
            assert(ts.isPropertySignature(member) && member.questionToken && member.type, 'expected optional property');
            if (member.getText(source) === after) adapted++;
            else edits.push({at: member.type.end, text: ' | undefined'});
        }
        let result = text;
        for (const edit of edits.sort((a, b) => b.at - a.at)) result = result.slice(0, edit.at) + edit.text + result.slice(edit.at);
        assert.equal(ts.createSourceFile(fileName, result, ts.ScriptTarget.Latest, true).parseDiagnostics.length, 0, `cannot parse adapted ${file}`);
        plans.push({fileName, text, result, declarations: edits.length});
    }
    assert(adapted === 0 || adapted === sites.length, `tree is partly adapted (${adapted} of ${sites.length})`);
    // Validate every file and every unchanged input before the first write.
    for (const plan of plans) assert.equal(fs.readFileSync(plan.fileName, 'utf8'), plan.text, 'source changed while planning');
    for (const plan of plans) if (plan.declarations) fs.writeFileSync(plan.fileName, plan.result);
    return {files: plans.filter(plan => plan.declarations).length,
        declarations: plans.reduce((sum, plan) => sum + plan.declarations, 0), javascriptChanging: false};
}
module.exports = {adapt, sites};
if (require.main === module) {
    assert.equal(process.argv.length, 3, 'usage: node adapt.cjs <tree>');
    console.log(JSON.stringify(adapt(path.resolve(process.argv[2])), null, 2));
}
