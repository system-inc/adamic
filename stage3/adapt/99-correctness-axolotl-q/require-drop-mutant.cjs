#!/usr/bin/env node
'use strict';
// Restoring mutant: alter only H05's after text in a temporary adapter and tree.
// Recorded on Node v24.19.0: isNodeLikeSystem=false; requireType=undefined.
// The temporary adapter/tree are removed after observing the oracle result.
const fs = require('node:fs'), path = require('node:path'), os = require('node:os');
const assert = require('node:assert/strict'), crypto = require('node:crypto');
const {execFileSync} = require('node:child_process');
const ts = require('typescript');
const [tree, output] = process.argv.slice(2);
assert(tree && output, 'usage: require-drop-mutant.cjs APPLIED_TREE OUTPUT.json');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), '99q-require-mutant-'));
try {
    const adapterFile = path.join(__dirname, 'adapt.cjs');
    const pin = require('./adapt.cjs').owners.find(owner => owner.owner === 'isNodeLikeSystem');
    const restored = pin.after.replace('&& !process.browser;', '&& !process.browser\n        && typeof require !== "undefined";');
    assert.notEqual(restored, pin.after);
    const digest = text => crypto.createHash('sha256').update(text).digest('hex');
    const adapterText = fs.readFileSync(adapterFile, 'utf8')
        .replace(JSON.stringify(pin.after), JSON.stringify(restored))
        .replace(pin.ownerAfterSha256, digest(restored));
    const mutant = path.join(scratch, 'mutant-adapt.cjs');
    fs.writeFileSync(mutant, adapterText);
    const raw = execFileSync('git', ['-C', tree, 'show', 'refs/stage3/99-correctness-axolotl-q-before:src/compiler/core.ts']);
    fs.mkdirSync(path.join(scratch, 'src/compiler'), {recursive:true});
    const coreFile = path.join(scratch, 'src/compiler/core.ts');
    fs.writeFileSync(coreFile, raw);
    const applied = require(mutant).adapt(scratch);
    const source = ts.createSourceFile('core.ts', fs.readFileSync(coreFile, 'utf8'), ts.ScriptTarget.Latest, true);
    const fn = source.statements.filter(node => ts.isFunctionDeclaration(node) && node.name?.text === 'isNodeLikeSystem');
    assert.equal(fn.length, 1);
    const declarations = source.statements.filter(node => ts.isVariableStatement(node) && node.modifiers?.some(m => m.kind === ts.SyntaxKind.DeclareKeyword));
    const probe = path.join(scratch, 'probe.a');
    fs.writeFileSync(probe, [...declarations, fn[0]].map(n => n.getText(source)).join('\n') + '\nconsole.log(JSON.stringify({isNodeLikeSystem:isNodeLikeSystem(),requireType:typeof require}));\n');
    const runner = path.resolve(__dirname, '../../../oracle/node.mjs');
    const result = JSON.parse(execFileSync(process.execPath, ['--disable-warning=ExperimentalWarning', runner, probe], {encoding:'utf8'}));
    assert.deepEqual(result, {isNodeLikeSystem:false,requireType:'undefined'});
    fs.writeFileSync(output, JSON.stringify({node:process.version, mutation:'Restore the require test in H05 after text; recompute its temporary after hash', applied, result, temporaryMutationRemoved:true}, null, 2)+'\n');
    console.log(JSON.stringify(result));
} finally {
    fs.rmSync(scratch, {recursive:true,force:true});
}
