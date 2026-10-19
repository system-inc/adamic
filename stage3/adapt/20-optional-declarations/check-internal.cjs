'use strict';
const assert = require('node:assert/strict');
const fs = require('node:fs'), os = require('node:os'), path = require('node:path'), cp = require('node:child_process');
const ts = require('../../api/node_modules/typescript');
const root = fs.mkdtempSync(path.join(os.tmpdir(), 'optional-internal-check-'));
fs.mkdirSync(path.join(root, 'src/compiler'), { recursive: true });
const original = fs.readFileSync(path.join(__dirname, 'testdata/internal-resolution.a'), 'utf8');
const file = path.join(root, 'src/compiler/fixture.ts');
fs.writeFileSync(file, original);
const report = JSON.parse(cp.execFileSync(process.execPath, [path.join(__dirname, 'adapt.cjs'), root], {
    env: { ...process.env, TSC_ADAPT_TYPESCRIPT: require.resolve('../../api/node_modules/typescript') },
    timeout: 90000, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'],
}));
assert.deepEqual(report.declarations.map(owner => owner.owner_path.join('.')).sort(), [
    'ArrayStorage.value', 'AssignmentStorage.value', 'CallableStorage.callback',
    'GenericData.value', 'InitializerData.value', 'MaskedStorage.optional',
]);
const adapted = fs.readFileSync(file, 'utf8');
for (const unchanged of ['interface BareStorage<T> { value?: T; }', 'interface ProjectReference { originalPath?: string; }',
    'interface ResolvedModuleFull { originalPath?: string; packageId?: PackageId; }',
    'interface MethodStorage { callback?(): number; }', 'interface IndexedStorage { value?: number; }',
    'interface ReceiverStorage { value?: number; }', 'interface PresentStorage { value: number; }',
    'interface MaybeStorage { value?: number; }']) assert(adapted.includes(unchanged), unchanged);
const emit = text => ts.transpileModule(text, { fileName: 'fixture.ts', compilerOptions: {
    target: ts.ScriptTarget.ES2024, module: ts.ModuleKind.ESNext,
} }).outputText;
assert.equal(emit(original), emit(adapted));
assert.throws(() => assert.equal(emit(original), emit(adapted + '\nconsole.log("mutant");\n')));
for (const [name, text] of [['before', original], ['after', adapted]]) {
    const script = path.join(root, name + '.mjs'); fs.writeFileSync(script, emit(text));
    assert.equal(cp.execFileSync(process.execPath, [script], { timeout: 10000, encoding: 'utf8' }), '[["value"],true]\n');
}
console.log(JSON.stringify({ status: 'pass', declarations: report.declarations, controls: 'bare T, public slot and blocked internal peer, optional method, indexed read, generic receiver, required view', root }));
