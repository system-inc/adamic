'use strict';
// Compare full upstream builds with just the reviewed local annotation changed.
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const assert = require('node:assert/strict'), {spawnSync} = require('node:child_process');
const ts = require('typescript');
assert.equal(ts.version, '6.0.3');
const [treeArg, proofArg] = process.argv.slice(2);
assert(treeArg && proofArg, 'usage: node filesystem-proof.cjs <adapted-tree> <proof-directory>');
const tree = path.resolve(treeArg), proof = path.resolve(proofArg);
fs.mkdirSync(proof, {recursive: true});
const file = path.join(tree, 'src/compiler/sys.ts');
const adapted = fs.readFileSync(file, 'utf8');
const annotation = 'let stat: import("fs").Stats | import("fs").Dirent | undefined;';
assert.equal(adapted.split(annotation).length, 2, 'one reviewed stat declaration');
const before = adapted.replace(annotation, 'let stat: any;');
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const walk = dir => fs.readdirSync(dir, {withFileTypes: true}).flatMap(e => e.isDirectory() ? walk(path.join(dir, e.name)) : [path.join(dir, e.name)]).sort();
function outputs() {
    const result = {};
    for (const f of walk(path.join(tree, 'built')).filter(f => /\.(?:js|mjs|cjs|d\.ts)$/.test(f))) result[path.relative(tree, f)] = hash(fs.readFileSync(f));
    return result;
}
function build(name) {
    const fd = fs.openSync(path.join(proof, name + '.log'), 'w');
    try {
        assert.equal(spawnSync('npm', ['run', 'clean'], {cwd: tree, stdio: ['ignore', fd, fd]}).status, 0, 'clean upstream build');
        return spawnSync('npm', ['run', 'build'], {cwd: tree, stdio: ['ignore', fd, fd]}).status;
    }
    finally { fs.closeSync(fd); }
}
const mutants = [];
try {
    fs.writeFileSync(file, before);
    assert.equal(build('before-build'), 0, 'before upstream build');
    const original = outputs();
    fs.writeFileSync(file, adapted);
    assert.equal(build('after-build'), 0, 'after upstream build');
    const actual = outputs();
    assert.deepEqual(actual, original, 'full emitted JavaScript and declaration file set and bytes');
    // sys.ts is bundled by upstream; also compare its own stock JavaScript emission.
    const emit = text => ts.transpileModule(text, {fileName: 'sys.ts', compilerOptions: {target: ts.ScriptTarget.ES2020, module: ts.ModuleKind.ESNext, removeComments: false}}).outputText;
    const oldJS = emit(before), newJS = emit(adapted);
    assert.equal(newJS, oldJS, 'sys.ts emitted JavaScript bytes');
    fs.writeFileSync(path.join(proof, 'before-sys.js'), oldJS);
    fs.writeFileSync(path.join(proof, 'after-sys.js'), newJS);
    const changed = {...actual};
    const first = Object.keys(changed).find(f => f.endsWith('.js'));
    changed[first] = hash('mutated JavaScript');
    assert.throws(() => assert.deepEqual(changed, original));
    mutants.push('emitted JavaScript byte mutant caught by hash comparison');
    assert.throws(() => assert.equal(newJS + '\n', oldJS));
    mutants.push('sys.ts emitted byte mutant caught by exact comparison');
    const planned = require('./classes.cjs').planEnum(before, 'src/compiler/sys.ts', 'filesystem');
    assert.equal(planned.text, adapted, 'independent single-token source reconstruction');
    assert.equal(planned.removed, 1);
    const second = require('./classes.cjs').planEnum(adapted, 'src/compiler/sys.ts', 'filesystem');
    assert.equal(second.text, adapted); assert.equal(second.removed, 0);
    assert.throws(() => require('./classes.cjs').planEnum(before.replace('let stat: any;', 'let stat: string;'), 'src/compiler/sys.ts', 'filesystem'));
    mutants.push('unreviewed stat annotation rejected by adapter');
    for (const [name, mutate] of [
        ['renamed owner', s => s.replace('function getAccessibleFileSystemEntries(', 'function changedOwner(')],
        ['missing declaration', s => s.replace('let stat: any;', '')],
        ['duplicate declaration', s => s.replace('let stat: any;', 'let stat: any;\r\n                    let stat: any;')],
        ['initialized declaration', s => s.replace('let stat: any;', 'let stat: any = undefined;')],
        ['missing type', s => s.replace('let stat: any;', 'let stat;')],
        ['changed declaration line', s => s.replace('let stat: any;', 'let stat: any; // drift')],
    ]) {
        assert.throws(() => require('./classes.cjs').planEnum(mutate(before), 'src/compiler/sys.ts', 'filesystem'));
        mutants.push('filesystem guard: ' + name);
    }

    fs.writeFileSync(file, adapted.replace(annotation, 'let stat: string;'));
    const mutantExit = build('string-mutant-build');
    assert.notEqual(mutantExit, 0, 'wrong stat type must fail upstream build');
    const log = fs.readFileSync(path.join(proof, 'string-mutant-build.log'), 'utf8');
    assert.match(log, /sys\.ts.*TS2322/, 'wrong assignment caught by upstream checker');
    assert.match(log, /Stats|Dirent/, 'filesystem assignment diagnostic');
    mutants.push('stat: string rejected by tsc own build with TS2322');
    const report = {node: process.version, typescript: ts.version, removed: 1, jsFiles: Object.keys(actual).filter(f => /\.(js|mjs|cjs)$/.test(f)).length, declarationFiles: Object.keys(actual).filter(f => f.endsWith('.d.ts')).length, sysJavaScriptSHA256: hash(newJS), changedDeclarations: [], outputs: actual, idempotent: true, mutantExit, mutants};
    fs.writeFileSync(path.join(proof, 'proof.json'), JSON.stringify(report, null, 2) + '\n');
    console.log(JSON.stringify(report, null, 2));
} finally {
    fs.writeFileSync(file, adapted);
    assert.equal(build('restored-build'), 0, 'restore successful upstream build after mutant');
}
