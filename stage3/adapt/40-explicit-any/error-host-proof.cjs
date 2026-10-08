'use strict';
const fs = require('node:fs'), path = require('node:path'), crypto = require('node:crypto');
const assert = require('node:assert/strict'), {spawnSync} = require('node:child_process');
const ts = require('typescript'), adapter = require('./error-host.cjs');
assert.equal(ts.version, '6.0.3'); assert.equal(process.version, 'v24.19.0');
const [treeArg, proofArg] = process.argv.slice(2);
assert(treeArg && proofArg, 'usage: error-host-proof.cjs TREE NEW_PROOF');
const tree = path.resolve(treeArg), proof = path.resolve(proofArg);
fs.mkdirSync(proof, {recursive:true});
const hash = bytes => crypto.createHash('sha256').update(bytes).digest('hex');
const sources = [...new Set(adapter.rules.map(r => r[0]))].map(file => {
    const name = path.join(tree,file), after = fs.readFileSync(name,'utf8');
    let before = after;
    for (const [f,,line] of adapter.rules.filter(r => r[0] === file)) {
        const adapted = line.replace('(Error as any)', 'Error');
        assert.equal(before.split(adapted).length,2);
        before = before.replace(adapted,line);
    }
    return {file,name,before,after};
});
const walk = dir => fs.readdirSync(dir,{withFileTypes:true}).flatMap(e => e.isDirectory()?walk(path.join(dir,e.name)):[path.join(dir,e.name)]).sort();
function outputs() {
    return Object.fromEntries(walk(path.join(tree,'built')).filter(f => /\.(js|mjs|cjs|d\.ts)$/.test(f)).map(f => [path.relative(tree,f),hash(fs.readFileSync(f))]));
}
function build(stem) {
    const fd = fs.openSync(path.join(proof,stem+'.log'),'w');
    try {
        assert.equal(spawnSync('npm',['run','clean'],{cwd:tree,stdio:['ignore',fd,fd]}).status,0);
        return spawnSync('npm',['run','build'],{cwd:tree,stdio:['ignore',fd,fd]}).status;
    } finally {fs.closeSync(fd);}
}
const mutants = [], emissions = {};
try {
    for(const s of sources) fs.writeFileSync(s.name,s.before);
    assert.equal(build('before-build'),0);
    const before = outputs();
    adapter.apply(tree);
    assert.equal(build('after-build'),0);
    const after = outputs(); assert.deepEqual(after,before,'all JavaScript and declarations identical');
    for(const s of sources) {
        assert.equal(fs.readFileSync(s.name,'utf8'),s.after);
        const emit = text => ts.transpileModule(text,{fileName:s.file,compilerOptions:{target:ts.ScriptTarget.ESNext,module:ts.ModuleKind.ESNext}}).outputText;
        const oldJS = emit(s.before), newJS = emit(s.after);
        assert.equal(newJS,oldJS);
        fs.writeFileSync(path.join(proof,path.basename(s.file)+'.before.js'),oldJS);
        fs.writeFileSync(path.join(proof,path.basename(s.file)+'.after.js'),newJS);
        emissions[s.file] = {bytes:Buffer.byteLength(newJS),sha256:hash(newJS)};
        const second = adapter.plan(s.after,s.file);assert.equal(second.text,s.after);assert.equal(second.removed,0);
        const line = adapter.rules.find(r => r[0] === s.file)[2];
        for (const [name, changed] of [
            ['unknown receiver', s.before.replace('(Error as any)', '(Error as unknown)')],
            ['duplicate line', s.before.replace(line, line + '\n' + line)],
            ['missing line', s.before.replace(line, '')],
            ['changed owner', s.before.replace(s.file.endsWith('debug.ts') ? 'function fail(' : 'function setStackTraceLimit(', 'function wrongOwner(')],
        ]) {
            assert.throws(() => adapter.plan(changed, s.file));
            mutants.push(s.file + ': adapter rejected ' + name);
        }
        const changed = s.after.replace(s.file.endsWith('debug.ts')?'captureStackTrace(e,':'stackTraceLimit = 100',s.file.endsWith('debug.ts')?'captureStackTrace(new Error(),':'stackTraceLimit = 101');
        assert.notEqual(emit(changed),newJS);mutants.push(s.file+': runtime statement mutant caught by emission comparison');
    }
    const declaration = Object.keys(after).find(f => f.endsWith('typescript.d.ts'));assert(declaration);
    const mutantAPI = path.join(proof, 'api-byte-mutant.d.ts');
    fs.writeFileSync(mutantAPI, Buffer.concat([fs.readFileSync(path.join(tree, declaration)), Buffer.from('\n// changed API byte input\n')]));
    const apiMutant = {...after,[declaration]:hash(fs.readFileSync(mutantAPI))};assert.throws(() => assert.deepEqual(apiMutant,before));mutants.push('public API byte mutant caught by build artifact comparison');
    for(const s of sources) {
        const member = s.file.endsWith('debug.ts')?'captureStackTrace':'stackTraceLimit';
        fs.writeFileSync(s.name,s.after.replaceAll('Error.'+member,'Error.'+member+'Wrong'));
        const status = build(member+'-mutant');assert.notEqual(status,0);
        assert.match(fs.readFileSync(path.join(proof,member+'-mutant.log'),'utf8'),/TS(?:2339|2551)/);
        mutants.push(member+'Wrong rejected by upstream build');fs.writeFileSync(s.name,s.after);
    }
    const report = {node:process.version,typescript:ts.version,removed:4,newDiagnostics:0,emissions,outputs:after,changedDeclarations:[],mutants,idempotent:true};
    fs.writeFileSync(path.join(proof,'proof.json'),JSON.stringify(report,null,2)+'\n');console.log(JSON.stringify(report,null,2));
} finally {
    for(const s of sources) fs.writeFileSync(s.name,s.after);
    assert.equal(build('restored-build'),0);
}
