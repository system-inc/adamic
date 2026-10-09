// Observe the disputed shapes using the same TypeScript version as the pinned census.
const assert = require('node:assert/strict');
const ts = require('../../stage3/api/node_modules/typescript');
assert.equal(ts.version, '6.0.3');
function assertTree(value, active = new Set()) {
    if (!value || typeof value !== 'object') return;
    assert(!active.has(value), 'owning cycle');
    active.add(value);
    for (const child of Object.values(value)) assertTree(child, active);
    active.delete(value);
}
const parsed = ts.parseJsonConfigFileContent(
    { files: ['main.ts'], references: [{ path: '../dependency', circular: true }] },
    { useCaseSensitiveFileNames: true, readDirectory: () => [], fileExists: () => false, readFile: () => undefined },
    '/project',
);
const references = parsed.projectReferences;
assert.deepEqual(references, [{ path: '/dependency', originalPath: '../dependency', prepend: undefined, circular: true }]);
const program = ts.createProgram({ rootNames: [], options: { noLib: true }, projectReferences: references });
assert.equal(program.getProjectReferences(), references);
assertTree(references);
console.log('K60 ' + JSON.stringify(references) + '; Program keeps input array identity; scalar-only entries; no owning cycle');
const options = { incremental: true, declaration: true, noLib: true, tsBuildInfoFile: '/project/build.tsbuildinfo', outDir: '/project/output' };
const host = ts.createIncrementalCompilerHost(options);
host.getCurrentDirectory = () => '/project';
host.fileExists = path => path === '/project/main.ts';
host.readFile = path => path === '/project/main.ts' ? 'export const x=1;' : undefined;
const builder = ts.createEmitAndSemanticDiagnosticsBuilderProgram(['/project/main.ts'], options, host);
assert(builder.state && builder.getProgram().getBuildInfo);
for (const kind of [ts.getBuilderFileEmit(options), ts.BuilderFileEmit.Dts, ts.BuilderFileEmit.Js]) {
    builder.state.affectedFilesPendingEmit = new Map([['/project/main.ts', kind]]);
    const info = builder.getProgram().getBuildInfo();
    const pending = info.affectedFilesPendingEmit;
    assert(pending && pending.length === 1);
    assertTree(pending);
    assert(typeof pending[0] === 'number' || (Array.isArray(pending[0]) && pending[0].every(item => typeof item === 'number')));
    const next = builder.getProgram().getBuildInfo().affectedFilesPendingEmit;
    assert.notEqual(next, pending);
    assert.deepEqual(next, pending);
    assert.deepEqual(JSON.parse(JSON.stringify(info)).affectedFilesPendingEmit, pending);
    console.log('K144 ' + JSON.stringify(pending) + '; fresh serialization array per call; scalar/number tuples; JSON round trip stable; no owning cycle');
}
