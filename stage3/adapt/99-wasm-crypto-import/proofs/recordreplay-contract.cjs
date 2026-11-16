'use strict';
const fs = require('node:fs'), path = require('node:path'), os = require('node:os'), assert = require('node:assert/strict');
const ts = require('typescript');
const {compiler, nodes} = require('../../47-host-errors/adapt.cjs');
const {adapt, plannedProgram} = require('../adapt.cjs');
const recordreplay = require('../recordreplay.cjs');
const baseline = path.resolve(process.argv[2]), expected = fs.readFileSync(process.argv[3], 'utf8');
const scratch = fs.mkdtempSync(path.join(os.tmpdir(), 'recordreplay-contract-'));
function owner(file) {
    const found = nodes(ts, file, n => ts.isFunctionDeclaration(n) && n.name?.text === 'getNodeSystem');
    assert.equal(found.length, 1); return found[0];
}
function parse(text) { return ts.createSourceFile('sys.ts', text, ts.ScriptTarget.Latest, true); }
try {
    fs.cpSync(path.join(baseline, 'src'), path.join(scratch, 'src'), {recursive: true});
    assert(!fs.existsSync(path.join(scratch, 'node_modules')), 'contract must exercise official apply without upstream dependencies');
    const fileName = path.join(scratch, 'src/compiler/sys.ts');
    if (process.argv[4] === 'no-node-resolution-mutant') {
        const original = recordreplay.virtualHost;
        recordreplay.virtualHost = (host, file) => { original(host, file); delete host.resolveTypeReferenceDirectives; };
        adapt(scratch);
        throw Error('disabled Node type resolution survived');
    }
    assert.equal(adapt(scratch).files, 2);
    assert.equal(fs.readFileSync(fileName, 'utf8'), expected);
    const declarationName = path.join(scratch, 'src/recordreplay.d.ts');
    assert.equal(fs.readFileSync(declarationName, 'utf8'), recordreplay.declaration);
    const stamps = [fileName, declarationName].map(p => fs.statSync(p, {bigint: true}).mtimeNs);
    assert.equal(adapt(scratch).files, 0);
    assert.deepEqual([fileName, declarationName].map(p => fs.statSync(p, {bigint: true}).mtimeNs), stamps);
    const stock = require('./recordreplay/stock-check.cjs').check(scratch);
    assert.equal(stock.withDeclaration.length, 0);
    assert(stock.withoutDeclaration.some(d => d.code === 2339 && d.message.includes('recordreplay')));
    // Relocate the complete owner and add unrelated members. Selection follows
    // the nodeSystem.debugMode AST, without a line, offset or neighboring member.
    const relocated = parse(expected.replace('function getNodeSystem()', '\n\n/* relocated owner */\nfunction getNodeSystem()')
        .replace('debugMode: !!process.env.NODE_INSPECTOR_IPC', 'unrelatedProbe: 1,\n            debugMode: !!process.env.NODE_INSPECTOR_IPC'));
    assert.equal(recordreplay.accesses(relocated, owner(relocated)).getText(relocated), 'process.recordreplay');
    for (const [name, text] of [
        ['wrong store', expected.replace('const nodeSystem:', 'const otherSystem:')],
        ['wrong property', expected.replace('debugMode: !!process.env.NODE_INSPECTOR_IPC', 'otherFlag: !!process.env.NODE_INSPECTOR_IPC')],
        ['duplicate read', expected.replace('!!process.recordreplay,', '!!process.recordreplay || !!process.recordreplay,')],
    ]) {
        const file = parse(text); assert.notEqual(text, expected);
        assert.throws(() => recordreplay.accesses(file, owner(file)), undefined, name);
    }
    const p = compiler(scratch, ts), f = p.getSourceFile(fileName);
    const shadow = expected.replace('function getNodeSystem()', 'function getNodeSystem(process: { recordreplay: unknown })');
    const shadowProgram = plannedProgram(p, f, shadow), shadowFile = shadowProgram.getSourceFile(fileName);
    assert.throws(() => recordreplay.validate(shadowProgram, shadowFile, owner(shadowFile)), /pinned ambient NodeJS.Process/);
    console.log(JSON.stringify({noUpstreamNodeModules: true,exactComposedOutput: true,stockBuildDeclarationPresent: true,stockReadTypes: true,withoutDeclarationTS2339: true,zeroWriteReplay: true,relocatedOwnerAndUnrelatedMember: true,wrongStoreRejected: true,wrongPropertyRejected: true,duplicateReadRejected: true,localProcessShadowRejected: true}));
} finally { fs.rmSync(scratch, {recursive: true, force: true}); }
